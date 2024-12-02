package cloudregistry

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"time"
)

var (
	// ErrNotFound is returned when no service addresses are found.
	ErrNotFound = errors.New("no service addresses found")
	// ErrNotReady is returned when the service is not ready.
	ErrNotReady = errors.New("service is not ready")
)

// GenerateInstanceID generates a psuedo-random service instance identifier, using a service name. Suffixed by dash and number
func GenerateInstanceID(serviceName string) string {
	return fmt.Sprintf("%s-%d", serviceName, rand.New(rand.NewSource(time.Now().UnixNano())).Int())
}

// ValueClient is the interface that wraps the basic methods to interact with the cloud registry.
type ValueClient interface {
	// Values returns a ValueClient to interact with the cloud registry.
	Values(ctx context.Context, prefix ...string) ValueClient
	// Value returns a value from the cloud registry.
	Value(ctx context.Context, name string) (string, error)
	// SetValue sets a value in the cloud registry.
	SetValue(ctx context.Context, name, value string) error
	// SubscribeValue subscribes to a value in the cloud registry.
	SubscribeValue(ctx context.Context, name string, val ValueSetter) error
	// SubscribeValueWithPrefix subscribes to a value in the cloud registry.
	SubscribeValueWithPrefix(ctx context.Context, prefix string, val ValueSetter) error
}

// Registry is the interface that wraps the basic methods to interact with the cloud registry.
type Registry interface {
	io.Closer
	ValueClient
	// Register registers a service in the cloud registry.
	Register(ctx context.Context, service *Service) error
	// Deregister deregisters a service from the cloud registry.
	Deregister(ctx context.Context, name, id string) error
	// Discover discovers a service in the cloud registry.
	Discover(ctx context.Context, name string, TTL time.Duration) ([]*ServiceInfo, error)
	// HealthCheck checks the health of a service in the cloud registry.
	HealthCheck(ctx context.Context, name, id string, TTL time.Duration) error
}
