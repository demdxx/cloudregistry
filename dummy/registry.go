package dummy

import (
	"context"
	"time"

	"github.com/demdxx/cloudregistry"
)

// Registry is a dummy implementation of the cloud registry.
type Registry struct{}

// Register registers a service in the cloud registry.
func (r *Registry) Register(ctx context.Context, service *cloudregistry.Service) error {
	return nil
}

// Deregister deregisters a service from the cloud registry.
func (r *Registry) Deregister(ctx context.Context, id *cloudregistry.ServiceID) error {
	return nil
}

// Discover discovers a service in the cloud registry.
func (r *Registry) Discover(ctx context.Context, prefix *cloudregistry.ServicePrefix, TTL time.Duration) ([]*cloudregistry.ServiceInfo, error) {
	return nil, nil
}

// HealthCheck checks the health of a service in the cloud registry.
func (r *Registry) HealthCheck(ctx context.Context, id *cloudregistry.ServiceID, TTL time.Duration) error {
	return nil
}

// Values returns a ValueClient to interact with the cloud registry.
func (r *Registry) Values(ctx context.Context, prefix ...string) cloudregistry.ValueClient {
	return r
}

// Value returns a value from the cloud registry.
func (r *Registry) Value(ctx context.Context, name string) (string, error) {
	return "", nil
}

// SetValue sets a value in the cloud registry.
func (r *Registry) SetValue(ctx context.Context, name, value string) error {
	return nil
}

// SubscribeValue subscribes to a value in the cloud registry.
func (r *Registry) SubscribeValue(ctx context.Context, name string, val cloudregistry.ValueSetter) error {
	return nil
}

// SubscribeValueWithPrefix subscribes to a value in the cloud registry.
func (r *Registry) SubscribeValueWithPrefix(ctx context.Context, prefix string, val cloudregistry.ValueSetter) error {
	return nil
}
