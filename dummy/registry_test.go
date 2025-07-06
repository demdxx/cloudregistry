package dummy

import (
	"context"
	"testing"
	"time"

	"github.com/demdxx/cloudregistry"
)

func TestRegistry_Register(t *testing.T) {
	registry := &Registry{}
	service := &cloudregistry.Service{
		Name:       "test-service",
		InstanceID: "test-instance",
		Hostname:   "localhost",
		Port:       8080,
	}

	err := registry.Register(context.Background(), service)
	if err != nil {
		t.Errorf("Registry.Register() error = %v, want nil", err)
	}
}

func TestRegistry_Deregister(t *testing.T) {
	registry := &Registry{}
	serviceID := &cloudregistry.ServiceID{
		Name:       "test-service",
		InstanceID: "test-instance",
	}

	err := registry.Deregister(context.Background(), serviceID)
	if err != nil {
		t.Errorf("Registry.Deregister() error = %v, want nil", err)
	}
}

func TestRegistry_Discover(t *testing.T) {
	registry := &Registry{}
	prefix := &cloudregistry.ServicePrefix{
		Name: "test-service",
	}

	services, err := registry.Discover(context.Background(), prefix, 60*time.Second)
	if err != nil {
		t.Errorf("Registry.Discover() error = %v, want nil", err)
	}
	if services != nil {
		t.Errorf("Registry.Discover() = %v, want nil", services)
	}
}

func TestRegistry_HealthCheck(t *testing.T) {
	registry := &Registry{}
	serviceID := &cloudregistry.ServiceID{
		Name:       "test-service",
		InstanceID: "test-instance",
	}

	err := registry.HealthCheck(context.Background(), serviceID, 30*time.Second)
	if err != nil {
		t.Errorf("Registry.HealthCheck() error = %v, want nil", err)
	}
}

func TestRegistry_Values(t *testing.T) {
	registry := &Registry{}

	client := registry.Values(context.Background(), "test-prefix")
	if client == nil {
		t.Error("Registry.Values() should not return nil")
	}

	// Test that it returns the same registry instance
	if client != registry {
		t.Error("Registry.Values() should return the same registry instance")
	}
}

func TestRegistry_Value(t *testing.T) {
	registry := &Registry{}

	value, err := registry.Value(context.Background(), "test-key")
	if err != nil {
		t.Errorf("Registry.Value() error = %v, want nil", err)
	}
	if value != "" {
		t.Errorf("Registry.Value() = %v, want empty string", value)
	}
}

func TestRegistry_SetValue(t *testing.T) {
	registry := &Registry{}

	err := registry.SetValue(context.Background(), "test-key", "test-value")
	if err != nil {
		t.Errorf("Registry.SetValue() error = %v, want nil", err)
	}
}

func TestRegistry_SubscribeValue(t *testing.T) {
	registry := &Registry{}
	setter := cloudregistry.ValueSetterFunc(func(key string, value any) error {
		return nil
	})

	err := registry.SubscribeValue(context.Background(), "test-key", setter)
	if err != nil {
		t.Errorf("Registry.SubscribeValue() error = %v, want nil", err)
	}
}

func TestRegistry_SubscribeValueWithPrefix(t *testing.T) {
	registry := &Registry{}
	setter := cloudregistry.ValueSetterFunc(func(key string, value any) error {
		return nil
	})

	err := registry.SubscribeValueWithPrefix(context.Background(), "test-prefix", setter)
	if err != nil {
		t.Errorf("Registry.SubscribeValueWithPrefix() error = %v, want nil", err)
	}
}

func TestRegistry_Close(t *testing.T) {
	registry := &Registry{}

	err := registry.Close()
	if err != nil {
		t.Errorf("Registry.Close() error = %v, want nil", err)
	}
}

func TestRegistry_Interface(t *testing.T) {
	// Test that Registry implements cloudregistry.Registry interface
	var reg cloudregistry.Registry = &Registry{}

	// Test basic interface compliance
	ctx := context.Background()

	// Test service registration
	service := &cloudregistry.Service{
		Name:       "test",
		InstanceID: "test-instance",
		Hostname:   "localhost",
		Port:       8080,
	}
	if err := reg.Register(ctx, service); err != nil {
		t.Errorf("Registry as cloudregistry.Registry.Register() error = %v", err)
	}

	// Test service discovery
	prefix := &cloudregistry.ServicePrefix{Name: "test"}
	if _, err := reg.Discover(ctx, prefix, time.Minute); err != nil {
		t.Errorf("Registry as cloudregistry.Registry.Discover() error = %v", err)
	}

	// Test value operations
	if err := reg.SetValue(ctx, "key", "value"); err != nil {
		t.Errorf("Registry as cloudregistry.Registry.SetValue() error = %v", err)
	}

	if _, err := reg.Value(ctx, "key"); err != nil {
		t.Errorf("Registry as cloudregistry.Registry.Value() error = %v", err)
	}

	// Test close
	if err := reg.Close(); err != nil {
		t.Errorf("Registry as cloudregistry.Registry.Close() error = %v", err)
	}
}
