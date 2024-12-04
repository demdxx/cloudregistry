package consul

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/hashicorp/consul/api"

	"github.com/demdxx/cloudregistry"
)

const defaultWaitTime = 10 * time.Second

// valueWatcherWrapper wraps a ValueSetter with watch parameters.
type valueWatcherWrapper struct {
	value     cloudregistry.ValueSetter
	key       string
	waitIndex uint64
	isPrefix  bool
}

// Registry is the Consul registry implementation.
type Registry struct {
	watcherWg   sync.WaitGroup
	watcherOnce sync.Once
	done        chan struct{}

	client   *api.Client
	prefix   string
	watchers chan *valueWatcherWrapper
	parent   *Registry
}

// Connect connects to the Consul cloud registry.
func Connect(ctx context.Context, options ...Option) (*Registry, error) {
	conf := api.DefaultConfig()
	for _, option := range options {
		option(conf)
	}

	client, err := api.NewClient(conf)
	if err != nil {
		return nil, err
	}

	return NewRegistry(client), nil
}

// NewRegistry creates a new Consul registry.
func NewRegistry(client *api.Client) *Registry {
	return &Registry{
		client:   client,
		done:     make(chan struct{}, 1),
		watchers: make(chan *valueWatcherWrapper, 10),
	}
}

// Register registers a service in the Consul cloud registry.
func (r *Registry) Register(ctx context.Context, service *cloudregistry.Service) error {
	reg := &api.AgentServiceRegistration{
		ID:        service.InstanceID,
		Name:      service.Name,
		Namespace: service.Namespace,
		Partition: service.Partition,
		Address:   service.Hostname,
		Port:      service.Port,
		Tags:      service.Tags,
		Meta:      service.Meta,
		Check: &api.AgentServiceCheck{
			CheckID:                        service.Check.ID,
			TTL:                            fmt.Sprintf("%ds", int(service.Check.TTL.Seconds())),
			DeregisterCriticalServiceAfter: fmt.Sprintf("%ds", int(service.Check.TTL.Seconds()*3)),
			HTTP:                           service.Check.HTTP.URL,
			Method:                         service.Check.HTTP.Method,
			Header:                         service.Check.HTTP.Headers,
		},
	}
	return r.client.Agent().ServiceRegister(reg)
}

// Deregister deregisters a service from the Consul cloud registry.
func (r *Registry) Deregister(ctx context.Context, id *cloudregistry.ServiceID) error {
	return r.client.Agent().ServiceDeregister(id.InstanceID)
}

// Discover discovers a service in the Consul cloud registry.
func (r *Registry) Discover(ctx context.Context, prefix *cloudregistry.ServicePrefix, TTL time.Duration) ([]*cloudregistry.ServiceInfo, error) {
	services, _, err := r.client.Catalog().Service(prefix.Name, "", nil)
	if err != nil {
		return nil, err
	}

	if len(services) == 0 {
		return nil, cloudregistry.ErrNotFound
	}

	serviceInfos := make([]*cloudregistry.ServiceInfo, 0, len(services))
	for _, svc := range services {
		if prefix.Namespace != "" && prefix.Namespace != svc.Namespace {
			continue
		}
		if prefix.Partition != "" && prefix.Partition != svc.Partition {
			continue
		}
		info := &cloudregistry.ServiceInfo{
			Name:       svc.ServiceName,
			Namespace:  svc.Namespace,
			Partition:  svc.Partition,
			InstanceID: svc.ServiceID,
			Hostname:   svc.ServiceAddress,
			Port:       svc.ServicePort,
			Tags:       svc.ServiceTags,
			Meta:       svc.ServiceMeta,
			LastUpdate: time.Now(), // Consul does not provide last update time directly
			// Populate Public and Private hosts if applicable
			// This example assumes single host; adjust as needed
			Public: []cloudregistry.Host{
				{
					Hostname: svc.ServiceAddress,
					Ports: cloudregistry.Ports{
						"http": fmt.Sprintf("%d", svc.ServicePort),
					},
				},
			},
			RawInfo: svc, // Store the raw service info for future use
		}
		serviceInfos = append(serviceInfos, info)
	}

	return serviceInfos, nil
}

// HealthCheck performs a health check for a service in the Consul cloud registry.
func (r *Registry) HealthCheck(ctx context.Context, id *cloudregistry.ServiceID, TTL time.Duration) error {
	// In Consul, health is managed via TTL checks. Here, we can verify the service exists.
	services, _, err := r.client.Catalog().Service(id.Name, "", nil)
	if err != nil {
		return err
	}

	for _, svc := range services {
		if svc.ServiceID == id.InstanceID {
			return nil
		}
	}

	return cloudregistry.ErrNotReady
}

// Values returns a ValueClient to interact with the Consul key-value store.
func (r *Registry) Values(ctx context.Context, prefix ...string) cloudregistry.ValueClient {
	newPrefix := r.prefix
	if len(prefix) > 0 {
		newPrefix += prefix[0]
	}
	return &Registry{
		done:     r.done,
		client:   r.client,
		prefix:   newPrefix,
		watchers: r.watchers,
		parent:   r,
	}
}

// Value returns a value from the Consul key-value store.
func (r *Registry) Value(ctx context.Context, name string) (string, error) {
	kv := r.client.KV()
	pair, _, err := kv.Get(r.prefix+name, nil)
	if err != nil {
		return "", err
	}
	if pair == nil {
		return "", cloudregistry.ErrNotFound
	}
	return string(pair.Value), nil
}

// SetValue sets a value in the Consul key-value store.
func (r *Registry) SetValue(ctx context.Context, name, value string) error {
	kv := r.client.KV()
	p := &api.KVPair{
		Key:   r.prefix + name,
		Value: []byte(value),
	}
	_, err := kv.Put(p, nil)
	return err
}

// SubscribeValue subscribes to a value in the Consul key-value store.
func (r *Registry) SubscribeValue(ctx context.Context, name string, val cloudregistry.ValueSetter) error {
	return r.subscriveValue(ctx, r.prefix+name, false, val)
}

// SubscribeValueWithPrefix subscribes to values with a specific prefix in the Consul key-value store.
func (r *Registry) SubscribeValueWithPrefix(ctx context.Context, prefix string, val cloudregistry.ValueSetter) error {
	return r.subscriveValue(ctx, r.prefix+prefix, true, val)
}

// subscriveValue is a common internal method for handling subscriptions.
// keyOrPrefix: the key or prefix to watch
// isPrefix: true if watching a prefix, false if watching a single key
// val: the ValueSetter to invoke on updates
func (r *Registry) subscriveValue(ctx context.Context, keyOrPrefix string, isPrefix bool, val cloudregistry.ValueSetter) error {
	if r.parent != nil {
		return r.parent.subscriveValue(ctx, keyOrPrefix, isPrefix, val)
	}

	r.watcherOnce.Do(func() {
		r.watcherWg.Add(1)
		go r.valueWatcher(ctx)
	})

	wrapper := &valueWatcherWrapper{
		value:    val,
		key:      keyOrPrefix,
		isPrefix: isPrefix,
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-r.done:
		return nil
	case r.watchers <- wrapper:
		return nil
	}
}

// valueWatcher handles all subscription updates.
func (r *Registry) valueWatcher(ctx context.Context) {
	defer r.watcherWg.Done()
	kv := r.client.KV()
	for {
		select {
		case <-ctx.Done():
			return
		case <-r.done:
			return
		case wr := <-r.watchers:
			if wr == nil {
				continue
			}
			// TODO: Handle errors
			if wr.isPrefix {
				_ = r.handlePrefixWatch(ctx, kv, wr)
			} else {
				_ = r.handleKeyWatch(ctx, kv, wr)
			}
			select {
			case <-ctx.Done():
				return
			case <-r.done:
				return
			default:
				r.watchers <- wr
			}
			// time.Sleep(100 * time.Millisecond)
		}
	}
}

// handleKeyWatch handles subscription to a single key.
func (r *Registry) handleKeyWatch(ctx context.Context, kv *api.KV, wrapper *valueWatcherWrapper) error {
	opts := &api.QueryOptions{
		UseCache:   true,
		AllowStale: true,
		WaitTime:   defaultWaitTime,
		WaitIndex:  wrapper.waitIndex,
	}
	pair, meta, err := kv.Get(wrapper.key, opts.WithContext(ctx))
	if err != nil || pair == nil {
		if api.IsRetryableError(err) {
			wrapper.waitIndex = 0
		}
		return err
	}

	if meta.LastIndex != wrapper.waitIndex {
		wrapper.waitIndex = meta.LastIndex
	} else {
		return nil
	}

	var value any
	if err := json.Unmarshal(pair.Value, &value); err != nil {
		value = string(pair.Value)
	}
	if err := wrapper.value.SetValue(pair.Key, value); err != nil {
		return err
	}
	return nil
}

// handlePrefixWatch handles subscription to a key prefix.
func (r *Registry) handlePrefixWatch(ctx context.Context, kv *api.KV, wrapper *valueWatcherWrapper) error {
	opts := &api.QueryOptions{
		UseCache:   true,
		AllowStale: true,
		WaitTime:   defaultWaitTime,
		WaitIndex:  wrapper.waitIndex,
	}
	pairs, meta, err := kv.List(wrapper.key, opts.WithContext(ctx))
	if err != nil {
		if api.IsRetryableError(err) {
			wrapper.waitIndex = 0
		}
		return err
	}

	if meta.LastIndex != wrapper.waitIndex {
		wrapper.waitIndex = meta.LastIndex
	} else {
		return nil
	}

	for _, pair := range pairs {
		var value any
		if err := json.Unmarshal(pair.Value, &value); err != nil {
			value = string(pair.Value)
		}
		if err := wrapper.value.SetValue(pair.Key, value); err != nil {
			return err
		}
	}
	return nil
}

// Close closes the Consul client and waits for all watchers to finish.
func (r *Registry) Close() error {
	// Signal the watcher to stop
	r.done <- struct{}{}

	// Close the watchers channel to signal watchers to stop
	close(r.watchers)

	// Wait for all watcher goroutines to finish
	r.watcherWg.Wait()

	return nil
}
