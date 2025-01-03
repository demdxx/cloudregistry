package etcd

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"time"

	errorsw "github.com/pkg/errors"
	"go.etcd.io/etcd/api/v3/v3rpc/rpctypes"
	clientv3 "go.etcd.io/etcd/client/v3"

	"github.com/demdxx/cloudregistry"
)

type valueWatcherWrapper struct {
	value   cloudregistry.ValueSetter
	watcher clientv3.WatchChan
}

// Registry is the etcd registry implementation.
type Registry struct {
	watcherWg   sync.WaitGroup
	watcherOnce sync.Once
	done        chan struct{}

	cli      *clientv3.Client
	prefix   string
	watchers chan *valueWatcherWrapper
	parent   *Registry
}

// Connect connects to the cloud registry.
func Connect(ctx context.Context, options ...Option) (*Registry, error) {
	conf := clientv3.Config{}
	for _, option := range options {
		option(&conf)
	}
	if conf.DialTimeout == 0 {
		conf.DialTimeout = 5 * time.Second
	}
	cli, err := clientv3.New(conf)
	if err != nil {
		return nil, err
	}
	return NewRegistry(cli), nil
}

// NewRegistry creates a new etcd registry.
func NewRegistry(cli *clientv3.Client) *Registry {
	return &Registry{
		cli:      cli,
		done:     make(chan struct{}, 1),
		watchers: make(chan *valueWatcherWrapper, 100),
	}
}

// Register registers a service in the cloud registry.
func (r *Registry) Register(ctx context.Context, service *cloudregistry.Service) error {
	// Create a lease with TTL equal to the service's health check TTL
	leaseTTL := int64(service.Check.TTL.Seconds())
	leaseResp, err := r.cli.Grant(ctx, leaseTTL)
	if err != nil {
		return err
	}

	// Prepare the service information
	serviceInfo := &cloudregistry.ServiceInfo{
		Name:       service.Name,
		InstanceID: service.InstanceID,
		Hostname:   service.Hostname,
		Port:       service.Port,
		Public:     service.Public,
		Private:    service.Private,
		LastUpdate: time.Now(),
	}

	// Serialize the serviceInfo object to JSON
	data, err := json.Marshal(serviceInfo)
	if err != nil {
		return err
	}

	// Put the service data into etcd under the key with the lease
	_, err = r.cli.Put(ctx, service.ID().String(),
		string(data), clientv3.WithLease(leaseResp.ID))
	if err != nil {
		return err
	}

	// Keep the lease alive in a background goroutine
	ch, err := r.cli.KeepAlive(ctx, leaseResp.ID)
	if err != nil {
		return err
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-r.done:
				return
			case _, ok := <-ch:
				if !ok {
					// KeepAlive channel closed
					return
				}
			}
		}
	}()

	return nil
}

// Deregister deregisters a service from the cloud registry.
func (r *Registry) Deregister(ctx context.Context, id *cloudregistry.ServiceID) error {
	_, err := r.cli.Delete(ctx, id.String())
	return err
}

// Discover discovers a service in the cloud registry.
func (r *Registry) Discover(ctx context.Context, prefix *cloudregistry.ServicePrefix, TTL time.Duration) ([]*cloudregistry.ServiceInfo, error) {
	// Get all keys under the service name
	resp, err := r.cli.Get(ctx, prefix.String(), clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}

	services := make([]*cloudregistry.ServiceInfo, 0, len(resp.Kvs))

	for _, kv := range resp.Kvs {
		service := new(cloudregistry.ServiceInfo)
		err := json.Unmarshal(kv.Value, service)
		if err != nil {
			continue // Skip invalid entries
		}
		services = append(services, service)
	}

	if len(services) == 0 {
		return nil, cloudregistry.ErrNotFound
	}

	return services, nil
}

// HealthCheck checks the health of a service in the cloud registry.
func (r *Registry) HealthCheck(ctx context.Context, id *cloudregistry.ServiceID, TTL time.Duration) error {
	// Retrieve the lease ID associated with the service key
	resp, err := r.cli.Get(ctx, id.String(), clientv3.WithKeysOnly())
	if err != nil {
		return err
	}

	if len(resp.Kvs) == 0 {
		return cloudregistry.ErrNotFound
	}

	leaseID := clientv3.LeaseID(resp.Kvs[0].Lease)
	if leaseID == 0 {
		return cloudregistry.ErrNotReady
	}

	// Keep the lease alive once
	_, err = r.cli.KeepAliveOnce(ctx, leaseID)
	switch err.(type) {
	case clientv3.ErrKeepAliveHalted:
		return errorsw.Wrap(cloudregistry.ErrNotReady, err.Error())
	default:
		if errors.Is(err, rpctypes.ErrLeaseNotFound) {
			return cloudregistry.ErrNotReady
		}
	}
	return err
}

// Values returns a ValueClient to interact with the cloud registry.
func (r *Registry) Values(ctx context.Context, prefix ...string) cloudregistry.ValueClient {
	if len(prefix) > 0 {
		return &Registry{
			done:     r.done,
			cli:      r.cli,
			prefix:   r.prefix + prefix[0],
			watchers: r.watchers,
			parent:   r,
		}
	}
	return r
}

// Value returns a value from the cloud registry.
func (r *Registry) Value(ctx context.Context, name string) (string, error) {
	resp, err := r.cli.Get(ctx, r.prefix+name)
	if err != nil {
		return "", err
	}
	if len(resp.Kvs) == 0 {
		return "", cloudregistry.ErrNotFound
	}
	return string(resp.Kvs[0].Value), nil
}

// SetValue sets a value in the cloud registry.
func (r *Registry) SetValue(ctx context.Context, name, value string) error {
	_, err := r.cli.Put(ctx, r.prefix+name, value)
	return err
}

// SubscribeValue subscribes to a value in the cloud registry.
func (r *Registry) SubscribeValue(ctx context.Context, name string, val cloudregistry.ValueSetter) error {
	ch := r.cli.Watch(ctx, r.prefix+name)
	return r.subscriveValue(ch, val)
}

// SubscribeValueWithPrefix subscribes to a value in the cloud registry.
func (r *Registry) SubscribeValueWithPrefix(ctx context.Context, prefix string, val cloudregistry.ValueSetter) error {
	ch := r.cli.Watch(ctx, r.prefix+prefix, clientv3.WithPrefix())
	return r.subscriveValue(ch, val)
}

func (r *Registry) subscriveValue(watcher clientv3.WatchChan, val cloudregistry.ValueSetter) error {
	if r.parent != nil {
		return r.parent.subscriveValue(watcher, val)
	}
	r.watchers <- &valueWatcherWrapper{value: val, watcher: watcher}
	r.watcherOnce.Do(func() {
		r.watcherWg.Add(1)
		go r.valueWatcher(r.cli.Ctx())
	})
	return nil
}

func (r *Registry) valueWatcher(ctx context.Context) {
	defer r.watcherWg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		case <-r.done:
			return
		case wr, ok := <-r.watchers:
			if !ok {
				return
			}
			if wresp, ok := <-wr.watcher; ok {
				for _, ev := range wresp.Events {
					var value any
					if err := json.Unmarshal(ev.Kv.Value, &value); err != nil {
						value = string(ev.Kv.Value)
					}
					if err := wr.value.SetValue(string(ev.Kv.Key), value); err != nil {
						continue
					}
				}
			}
			select {
			case <-ctx.Done():
				return
			case <-r.done:
				return
			default:
				r.watchers <- wr
			}
		}
	}
}

// Close closes the cloud registry connection.
func (r *Registry) Close() (err error) {
	// Signal the watcher to stop
	r.done <- struct{}{}

	// Close the watchers channel to signal watchers to stop
	close(r.watchers)

	// Wait for all watchers to finish
	r.watcherWg.Wait()

	// Close the etcd client
	return r.cli.Close()
}
