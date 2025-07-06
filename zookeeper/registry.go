package zookeeper

import (
	"context"
	"encoding/json"
	"fmt"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/go-zookeeper/zk"

	"github.com/demdxx/cloudregistry"
)

const (
	defaultSessionTimeout = 10 * time.Second
	defaultBasePath       = "/services"
)

// zkConfig holds ZooKeeper connection configuration.
type zkConfig struct {
	hosts          []string
	sessionTimeout time.Duration
	basePath       string
}

// valueWatcherWrapper wraps a ValueSetter with watch parameters.
type valueWatcherWrapper struct {
	value    cloudregistry.ValueSetter
	path     string
	isPrefix bool
}

// Registry is the ZooKeeper registry implementation.
type Registry struct {
	watcherWg   sync.WaitGroup
	watcherOnce sync.Once
	done        chan struct{}

	conn     *zk.Conn
	prefix   string
	watchers chan *valueWatcherWrapper
	parent   *Registry
}

// Connect connects to the ZooKeeper cloud registry.
func Connect(ctx context.Context, options ...Option) (*Registry, error) {
	conf := &zkConfig{
		hosts:          []string{"localhost:2181"},
		sessionTimeout: defaultSessionTimeout,
		basePath:       defaultBasePath,
	}

	for _, option := range options {
		option(conf)
	}

	conn, _, err := zk.Connect(conf.hosts, conf.sessionTimeout)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to ZooKeeper: %w", err)
	}

	// Ensure base path exists
	if err := ensurePath(conn, conf.basePath); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to create base path: %w", err)
	}

	return NewRegistry(conn, conf.basePath), nil
}

// NewRegistry creates a new ZooKeeper registry.
func NewRegistry(conn *zk.Conn, basePath string) *Registry {
	if basePath == "" {
		basePath = defaultBasePath
	}
	return &Registry{
		conn:     conn,
		prefix:   basePath,
		done:     make(chan struct{}, 1),
		watchers: make(chan *valueWatcherWrapper, 100),
	}
}

// Register registers a service in the ZooKeeper cloud registry.
func (r *Registry) Register(ctx context.Context, service *cloudregistry.Service) error {
	if r.conn == nil {
		return fmt.Errorf("ZooKeeper connection is nil")
	}
	// Create service path
	servicePath := r.buildServicePath(service.ID())

	// Prepare the service information
	serviceInfo := &cloudregistry.ServiceInfo{
		Name:       service.Name,
		Namespace:  service.Namespace,
		Partition:  service.Partition,
		InstanceID: service.InstanceID,
		Hostname:   service.Hostname,
		Port:       service.Port,
		Public:     service.Public,
		Private:    service.Private,
		Tags:       service.Tags,
		Meta:       service.Meta,
		LastUpdate: time.Now(),
	}

	// Serialize the serviceInfo object to JSON
	data, err := json.Marshal(serviceInfo)
	if err != nil {
		return fmt.Errorf("failed to marshal service info: %w", err)
	}

	// Ensure parent path exists
	parentPath := path.Dir(servicePath)
	if err := ensurePath(r.conn, parentPath); err != nil {
		return fmt.Errorf("failed to create parent path: %w", err)
	}

	// Create ephemeral sequential node for the service instance
	actualPath, err := r.conn.Create(servicePath, data, zk.FlagEphemeral, zk.WorldACL(zk.PermAll))
	if err != nil {
		return fmt.Errorf("failed to register service: %w", err)
	}

	// Start health check routine if TTL is specified
	if service.Check.TTL > 0 {
		go r.healthCheckRoutine(ctx, actualPath, service.Check.TTL)
	}

	return nil
}

// Deregister deregisters a service from the ZooKeeper cloud registry.
func (r *Registry) Deregister(ctx context.Context, id *cloudregistry.ServiceID) error {
	if r.conn == nil {
		return fmt.Errorf("ZooKeeper connection is nil")
	}
	servicePath := r.buildServicePath(id)

	// List all instances of this service
	children, _, err := r.conn.Children(servicePath)
	if err != nil {
		if err == zk.ErrNoNode {
			return nil // Already deregistered
		}
		return fmt.Errorf("failed to list service instances: %w", err)
	}

	// Find and delete the instance
	for _, child := range children {
		instancePath := path.Join(servicePath, child)
		data, _, err := r.conn.Get(instancePath)
		if err != nil {
			continue
		}

		var serviceInfo cloudregistry.ServiceInfo
		if err := json.Unmarshal(data, &serviceInfo); err != nil {
			continue
		}

		if serviceInfo.InstanceID == id.InstanceID {
			if err := r.conn.Delete(instancePath, -1); err != nil {
				return fmt.Errorf("failed to deregister service: %w", err)
			}
			return nil
		}
	}

	return nil
}

// Discover discovers services in the ZooKeeper cloud registry.
func (r *Registry) Discover(ctx context.Context, prefix *cloudregistry.ServicePrefix, TTL time.Duration) ([]*cloudregistry.ServiceInfo, error) {
	if r.conn == nil {
		return nil, fmt.Errorf("ZooKeeper connection is nil")
	}
	servicePath := r.buildServicePrefixPath(prefix)

	children, _, err := r.conn.Children(servicePath)
	if err != nil {
		if err == zk.ErrNoNode {
			return nil, cloudregistry.ErrNotFound
		}
		return nil, fmt.Errorf("failed to discover services: %w", err)
	}

	var services []*cloudregistry.ServiceInfo
	for _, child := range children {
		instancePath := path.Join(servicePath, child)
		data, _, err := r.conn.Get(instancePath)
		if err != nil {
			continue
		}

		var serviceInfo cloudregistry.ServiceInfo
		if err := json.Unmarshal(data, &serviceInfo); err != nil {
			continue
		}

		// Check if service is still alive (not older than TTL)
		if TTL > 0 && time.Since(serviceInfo.LastUpdate) > TTL {
			continue
		}

		services = append(services, &serviceInfo)
	}

	if len(services) == 0 {
		return nil, cloudregistry.ErrNotFound
	}

	return services, nil
}

// HealthCheck checks the health of a service in the ZooKeeper cloud registry.
func (r *Registry) HealthCheck(ctx context.Context, id *cloudregistry.ServiceID, TTL time.Duration) error {
	if r.conn == nil {
		return fmt.Errorf("ZooKeeper connection is nil")
	}
	servicePath := r.buildServicePath(id)

	children, _, err := r.conn.Children(servicePath)
	if err != nil {
		if err == zk.ErrNoNode {
			return cloudregistry.ErrNotFound
		}
		return fmt.Errorf("failed to check service health: %w", err)
	}

	// Find the instance and update its timestamp
	for _, child := range children {
		instancePath := path.Join(servicePath, child)
		data, stat, err := r.conn.Get(instancePath)
		if err != nil {
			continue
		}

		var serviceInfo cloudregistry.ServiceInfo
		if err := json.Unmarshal(data, &serviceInfo); err != nil {
			continue
		}

		if serviceInfo.InstanceID == id.InstanceID {
			// Update LastUpdate timestamp
			serviceInfo.LastUpdate = time.Now()
			newData, err := json.Marshal(serviceInfo)
			if err != nil {
				return fmt.Errorf("failed to marshal service info: %w", err)
			}

			_, err = r.conn.Set(instancePath, newData, stat.Version)
			if err != nil {
				return fmt.Errorf("failed to update service health: %w", err)
			}
			return nil
		}
	}

	return cloudregistry.ErrNotFound
}

// Values returns a ValueClient to interact with the cloud registry.
func (r *Registry) Values(ctx context.Context, prefix ...string) cloudregistry.ValueClient {
	newPrefix := r.prefix
	if len(prefix) > 0 {
		newPrefix = path.Join(r.prefix, strings.Join(prefix, "/"))
	}

	return &Registry{
		conn:     r.conn,
		prefix:   newPrefix,
		done:     r.done,
		watchers: r.watchers,
		parent:   r,
	}
}

// Value returns a value from the ZooKeeper cloud registry.
func (r *Registry) Value(ctx context.Context, name string) (string, error) {
	if r.conn == nil {
		return "", fmt.Errorf("ZooKeeper connection is nil")
	}
	fullPath := path.Join(r.prefix, name)
	data, _, err := r.conn.Get(fullPath)
	if err != nil {
		if err == zk.ErrNoNode {
			return "", cloudregistry.ErrNotFound
		}
		return "", fmt.Errorf("failed to get value: %w", err)
	}
	return string(data), nil
}

// SetValue sets a value in the ZooKeeper cloud registry.
func (r *Registry) SetValue(ctx context.Context, name, value string) error {
	if r.conn == nil {
		return fmt.Errorf("ZooKeeper connection is nil")
	}
	fullPath := path.Join(r.prefix, name)

	// Ensure parent path exists
	if err := ensurePath(r.conn, path.Dir(fullPath)); err != nil {
		return fmt.Errorf("failed to create parent path: %w", err)
	}

	// Try to update existing node first
	_, err := r.conn.Set(fullPath, []byte(value), -1)
	if err == zk.ErrNoNode {
		// Node doesn't exist, create it
		_, err = r.conn.Create(fullPath, []byte(value), 0, zk.WorldACL(zk.PermAll))
	}

	if err != nil {
		return fmt.Errorf("failed to set value: %w", err)
	}
	return nil
}

// SubscribeValue subscribes to a value in the ZooKeeper cloud registry.
func (r *Registry) SubscribeValue(ctx context.Context, name string, val cloudregistry.ValueSetter) error {
	if r.conn == nil {
		return fmt.Errorf("ZooKeeper connection is nil")
	}
	fullPath := path.Join(r.prefix, name)

	r.startWatcherOnce()

	wrapper := &valueWatcherWrapper{
		value:    val,
		path:     fullPath,
		isPrefix: false,
	}

	select {
	case r.watchers <- wrapper:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// SubscribeValueWithPrefix subscribes to values with a prefix in the ZooKeeper cloud registry.
func (r *Registry) SubscribeValueWithPrefix(ctx context.Context, prefix string, val cloudregistry.ValueSetter) error {
	if r.conn == nil {
		return fmt.Errorf("ZooKeeper connection is nil")
	}
	fullPath := path.Join(r.prefix, prefix)

	r.startWatcherOnce()

	wrapper := &valueWatcherWrapper{
		value:    val,
		path:     fullPath,
		isPrefix: true,
	}

	select {
	case r.watchers <- wrapper:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Close closes the ZooKeeper connection and stops all watchers.
func (r *Registry) Close() error {
	// Signal all watchers to stop
	close(r.done)

	// Wait for all watchers to finish
	r.watcherWg.Wait()

	// Close the connection
	if r.conn != nil {
		r.conn.Close()
	}

	return nil
}

// Helper methods

func (r *Registry) buildServicePath(id *cloudregistry.ServiceID) string {
	return path.Join(r.prefix, id.String(), id.InstanceID)
}

func (r *Registry) buildServicePrefixPath(prefix *cloudregistry.ServicePrefix) string {
	prefixStr := prefix.String()
	if !strings.HasSuffix(prefixStr, "/") {
		prefixStr += "/"
	}
	result := path.Join(r.prefix, prefixStr)
	// Ensure trailing slash is preserved
	if !strings.HasSuffix(result, "/") {
		result += "/"
	}
	return result
}

func (r *Registry) startWatcherOnce() {
	r.watcherOnce.Do(func() {
		r.watcherWg.Add(1)
		go r.watcherRoutine()
	})
}

func (r *Registry) watcherRoutine() {
	defer r.watcherWg.Done()

	watchers := make(map[string]*valueWatcherWrapper)

	for {
		select {
		case <-r.done:
			return
		case wrapper := <-r.watchers:
			if wrapper == nil {
				continue
			}

			watchers[wrapper.path] = wrapper
			go r.watchPath(wrapper)
		}
	}
}

func (r *Registry) watchPath(wrapper *valueWatcherWrapper) {
	for {
		select {
		case <-r.done:
			return
		default:
			if wrapper.isPrefix {
				r.watchPrefix(wrapper)
			} else {
				r.watchSingle(wrapper)
			}
			time.Sleep(time.Second) // Prevent busy waiting
		}
	}
}

func (r *Registry) watchSingle(wrapper *valueWatcherWrapper) {
	data, _, events, err := r.conn.GetW(wrapper.path)
	if err != nil {
		if err != zk.ErrNoNode {
			return
		}
		// Node doesn't exist, wait for creation
		_, _, events, err := r.conn.ExistsW(wrapper.path)
		if err != nil {
			return
		}
		select {
		case <-r.done:
			return
		case <-events:
			// Node was created, try again
			return
		}
	}

	// Notify about current value
	wrapper.value.SetValue(wrapper.path, string(data))

	// Wait for changes
	select {
	case <-r.done:
		return
	case <-events:
		// Value changed, will be picked up in next iteration
	}
}

func (r *Registry) watchPrefix(wrapper *valueWatcherWrapper) {
	children, _, events, err := r.conn.ChildrenW(wrapper.path)
	if err != nil {
		return
	}

	// Notify about all current children
	for _, child := range children {
		childPath := path.Join(wrapper.path, child)
		data, _, err := r.conn.Get(childPath)
		if err == nil {
			wrapper.value.SetValue(childPath, string(data))
		}
	}

	// Wait for changes
	select {
	case <-r.done:
		return
	case <-events:
		// Children changed, will be picked up in next iteration
	}
}

func (r *Registry) healthCheckRoutine(ctx context.Context, servicePath string, ttl time.Duration) {
	ticker := time.NewTicker(ttl / 3) // Update health every 1/3 of TTL
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-r.done:
			return
		case <-ticker.C:
			// Update the node to keep it alive
			_, stat, err := r.conn.Get(servicePath)
			if err != nil {
				return // Node was deleted
			}

			// Touch the node to update its modification time
			_, err = r.conn.Set(servicePath, nil, stat.Version)
			if err != nil {
				return // Failed to update
			}
		}
	}
}

// ensurePath creates all parent directories for the given path.
func ensurePath(conn *zk.Conn, zkPath string) error {
	if conn == nil {
		return fmt.Errorf("ZooKeeper connection is nil")
	}
	if zkPath == "/" {
		return nil
	}

	exists, _, err := conn.Exists(zkPath)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	// Ensure parent exists first
	if err := ensurePath(conn, path.Dir(zkPath)); err != nil {
		return err
	}

	// Create this directory
	_, err = conn.Create(zkPath, []byte{}, 0, zk.WorldACL(zk.PermAll))
	if err != nil && err != zk.ErrNodeExists {
		return err
	}

	return nil
}
