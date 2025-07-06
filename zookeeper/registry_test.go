package zookeeper

import (
	"context"
	"testing"
	"time"

	"github.com/demdxx/cloudregistry"
)

func TestRegistry_Connect(t *testing.T) {
	// Test connection with default options
	_, err := Connect(context.Background())
	if err == nil {
		t.Error("Expected error when connecting to non-existent ZooKeeper, got nil")
	}

	// Test connection with custom hosts
	_, err = Connect(context.Background(), WithHosts([]string{"localhost:2181", "localhost:2182"}))
	if err == nil {
		t.Error("Expected error when connecting to non-existent ZooKeeper, got nil")
	}

	// Test connection with custom session timeout
	_, err = Connect(context.Background(), WithSessionTimeout(5*time.Second))
	if err == nil {
		t.Error("Expected error when connecting to non-existent ZooKeeper, got nil")
	}

	// Test connection with custom base path
	_, err = Connect(context.Background(), WithBasePath("/test"))
	if err == nil {
		t.Error("Expected error when connecting to non-existent ZooKeeper, got nil")
	}

	// Test connection with URI
	_, err = Connect(context.Background(), WithURI("zookeeper://localhost:2181/test?timeout=10s"))
	if err == nil {
		t.Error("Expected error when connecting to non-existent ZooKeeper, got nil")
	}
}

func TestRegistry_WithMockConn(t *testing.T) {
	// Since we can't easily test against a real ZooKeeper in unit tests,
	// we'll test the registry creation and basic structure

	// Test NewRegistry
	registry := NewRegistry(nil, "/test")
	if registry == nil {
		t.Fatal("NewRegistry should not return nil")
	}

	if registry.prefix != "/test" {
		t.Errorf("Expected prefix '/test', got '%s'", registry.prefix)
	}

	// Test with empty base path
	registry2 := NewRegistry(nil, "")
	if registry2 == nil {
		t.Fatal("NewRegistry should not return nil")
	}
	if registry2.prefix != defaultBasePath {
		t.Errorf("Expected default base path '%s', got '%s'", defaultBasePath, registry2.prefix)
	}
}

func TestRegistry_Interface(t *testing.T) {
	// Test that Registry implements cloudregistry.Registry interface
	var _ cloudregistry.Registry = (*Registry)(nil)

	// Test that Registry implements cloudregistry.ValueClient interface
	var _ cloudregistry.ValueClient = (*Registry)(nil)
}

func TestZkConfig(t *testing.T) {
	conf := &zkConfig{
		hosts:          []string{"localhost:2181"},
		sessionTimeout: 10 * time.Second,
		basePath:       "/services",
	}

	// Test WithHosts option
	WithHosts([]string{"host1:2181", "host2:2181"})(conf)
	if len(conf.hosts) != 2 || conf.hosts[0] != "host1:2181" || conf.hosts[1] != "host2:2181" {
		t.Errorf("WithHosts did not set hosts correctly: %v", conf.hosts)
	}

	// Test WithSessionTimeout option
	WithSessionTimeout(20 * time.Second)(conf)
	if conf.sessionTimeout != 20*time.Second {
		t.Errorf("WithSessionTimeout did not set timeout correctly: %v", conf.sessionTimeout)
	}

	// Test WithBasePath option
	WithBasePath("/custom")(conf)
	if conf.basePath != "/custom" {
		t.Errorf("WithBasePath did not set base path correctly: %s", conf.basePath)
	}
}

func TestRegistry_BuildPaths(t *testing.T) {
	registry := NewRegistry(nil, "/services")

	// Test buildServicePath
	serviceID := &cloudregistry.ServiceID{
		Name:       "test-service",
		Namespace:  "production",
		Partition:  "region-1",
		InstanceID: "instance-123",
	}

	servicePath := registry.buildServicePath(serviceID)
	expected := "/services/services/production/test-service/region-1/instance-123"
	if servicePath != expected {
		t.Errorf("buildServicePath() = %s, want %s", servicePath, expected)
	}

	// Test buildServicePrefixPath
	servicePrefix := &cloudregistry.ServicePrefix{
		Name:      "test-service",
		Namespace: "production",
		Partition: "region-1",
	}

	prefixPath := registry.buildServicePrefixPath(servicePrefix)
	expectedPrefix := "/services/services/production/test-service/region-1/"
	if prefixPath != expectedPrefix {
		t.Errorf("buildServicePrefixPath() = %s, want %s", prefixPath, expectedPrefix)
	}
}

func TestRegistry_Values(t *testing.T) {
	registry := NewRegistry(nil, "/services")

	// Test Values without prefix
	valueClient := registry.Values(context.Background())
	if valueClient == nil {
		t.Error("Values() should not return nil")
	}

	valueRegistry, ok := valueClient.(*Registry)
	if !ok {
		t.Error("Values() should return *Registry")
	}

	if valueRegistry.prefix != "/services" {
		t.Errorf("Values() prefix = %s, want /services", valueRegistry.prefix)
	}

	// Test Values with prefix
	valueClient2 := registry.Values(context.Background(), "config", "app")
	valueRegistry2, ok := valueClient2.(*Registry)
	if !ok {
		t.Error("Values() should return *Registry")
	}

	expectedPrefix := "/services/config/app"
	if valueRegistry2.prefix != expectedPrefix {
		t.Errorf("Values() with prefix = %s, want %s", valueRegistry2.prefix, expectedPrefix)
	}
}

// Test error scenarios and edge cases
func TestRegistry_ErrorCases(t *testing.T) {
	registry := NewRegistry(nil, "/services")
	ctx := context.Background()

	// Test operations with nil connection (should not panic)
	service := &cloudregistry.Service{
		Name:       "test-service",
		InstanceID: "test-instance",
		Hostname:   "localhost",
		Port:       8080,
	}

	// These should return errors due to nil connection, but not panic
	err := registry.Register(ctx, service)
	if err == nil {
		t.Error("Register with nil connection should return error")
	}

	serviceID := &cloudregistry.ServiceID{
		Name:       "test-service",
		InstanceID: "test-instance",
	}

	err = registry.Deregister(ctx, serviceID)
	if err == nil {
		t.Error("Deregister with nil connection should return error")
	}

	servicePrefix := &cloudregistry.ServicePrefix{
		Name: "test-service",
	}

	_, err = registry.Discover(ctx, servicePrefix, time.Minute)
	if err == nil {
		t.Error("Discover with nil connection should return error")
	}

	err = registry.HealthCheck(ctx, serviceID, time.Minute)
	if err == nil {
		t.Error("HealthCheck with nil connection should return error")
	}

	_, err = registry.Value(ctx, "test-key")
	if err == nil {
		t.Error("Value with nil connection should return error")
	}

	err = registry.SetValue(ctx, "test-key", "test-value")
	if err == nil {
		t.Error("SetValue with nil connection should return error")
	}

	setter := cloudregistry.ValueSetterFunc(func(key string, value any) error {
		return nil
	})

	err = registry.SubscribeValue(ctx, "test-key", setter)
	if err == nil {
		t.Error("SubscribeValue with nil connection should return error")
	}

	err = registry.SubscribeValueWithPrefix(ctx, "test-prefix", setter)
	if err == nil {
		t.Error("SubscribeValueWithPrefix with nil connection should return error")
	}

	// Test Close (should not panic with nil connection)
	err = registry.Close()
	if err != nil {
		t.Errorf("Close() error = %v, want nil", err)
	}
}

func TestValueWatcherWrapper(t *testing.T) {
	setter := cloudregistry.ValueSetterFunc(func(key string, value any) error {
		return nil
	})

	wrapper := &valueWatcherWrapper{
		value:    setter,
		path:     "/test/path",
		isPrefix: true,
	}

	if wrapper.value == nil {
		t.Error("valueWatcherWrapper.value should not be nil")
	}

	if wrapper.path != "/test/path" {
		t.Errorf("valueWatcherWrapper.path = %s, want /test/path", wrapper.path)
	}

	if !wrapper.isPrefix {
		t.Error("valueWatcherWrapper.isPrefix should be true")
	}
}

func TestWithURI(t *testing.T) {
	tests := []struct {
		name        string
		uri         string
		expectPanic bool
		expected    zkConfig
	}{
		{
			name: "basic zookeeper URI",
			uri:  "zookeeper://localhost:2181",
			expected: zkConfig{
				hosts:          []string{"localhost:2181"},
				sessionTimeout: defaultSessionTimeout,
				basePath:       defaultBasePath,
			},
		},
		{
			name: "multiple hosts",
			uri:  "zookeeper://host1:2181,host2:2181,host3:2181",
			expected: zkConfig{
				hosts:          []string{"host1:2181", "host2:2181", "host3:2181"},
				sessionTimeout: defaultSessionTimeout,
				basePath:       defaultBasePath,
			},
		},
		{
			name: "with custom path",
			uri:  "zookeeper://localhost:2181/myapp",
			expected: zkConfig{
				hosts:          []string{"localhost:2181"},
				sessionTimeout: defaultSessionTimeout,
				basePath:       "/myapp", // path with leading slash
			},
		},
		{
			name: "with timeout parameter",
			uri:  "zookeeper://localhost:2181?timeout=30s",
			expected: zkConfig{
				hosts:          []string{"localhost:2181"},
				sessionTimeout: 30 * time.Second,
				basePath:       defaultBasePath,
			},
		},
		{
			name: "with path and timeout",
			uri:  "zookeeper://localhost:2181/myapp?timeout=15s",
			expected: zkConfig{
				hosts:          []string{"localhost:2181"},
				sessionTimeout: 15 * time.Second,
				basePath:       "/myapp", // path with leading slash
			},
		},
		{
			name: "zk scheme",
			uri:  "zk://localhost:2181",
			expected: zkConfig{
				hosts:          []string{"localhost:2181"},
				sessionTimeout: defaultSessionTimeout,
				basePath:       defaultBasePath,
			},
		},
		{
			name:        "invalid scheme",
			uri:         "redis://localhost:6379",
			expectPanic: true,
		},
		{
			name:        "malformed URI",
			uri:         "://invalid",
			expectPanic: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.expectPanic {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("WithURI(%s) expected panic but didn't panic", tt.uri)
					}
				}()
			}

			conf := &zkConfig{
				hosts:          []string{},
				sessionTimeout: 0,
				basePath:       "",
			}

			WithURI(tt.uri)(conf)

			if !tt.expectPanic {
				if !equalSlices(conf.hosts, tt.expected.hosts) {
					t.Errorf("WithURI(%s) hosts = %v, want %v", tt.uri, conf.hosts, tt.expected.hosts)
				}
				if conf.sessionTimeout != tt.expected.sessionTimeout {
					t.Errorf("WithURI(%s) sessionTimeout = %v, want %v", tt.uri, conf.sessionTimeout, tt.expected.sessionTimeout)
				}
				if conf.basePath != tt.expected.basePath {
					t.Errorf("WithURI(%s) basePath = %s, want %s", tt.uri, conf.basePath, tt.expected.basePath)
				}
			}
		})
	}
}

func TestWithURI_EdgeCases(t *testing.T) {
	tests := []struct {
		name string
		uri  string
		want zkConfig
	}{
		{
			name: "empty host should be filtered",
			uri:  "zookeeper://host1:2181,,host2:2181",
			want: zkConfig{
				hosts:          []string{"host1:2181", "host2:2181"},
				sessionTimeout: defaultSessionTimeout,
				basePath:       defaultBasePath,
			},
		},
		{
			name: "invalid timeout should be ignored",
			uri:  "zookeeper://localhost:2181?timeout=invalid",
			want: zkConfig{
				hosts:          []string{"localhost:2181"},
				sessionTimeout: defaultSessionTimeout, // should use default
				basePath:       defaultBasePath,
			},
		},
		{
			name: "empty path should use default",
			uri:  "zookeeper://localhost:2181/",
			want: zkConfig{
				hosts:          []string{"localhost:2181"},
				sessionTimeout: defaultSessionTimeout,
				basePath:       defaultBasePath,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conf := &zkConfig{}
			WithURI(tt.uri)(conf)

			if !equalSlices(conf.hosts, tt.want.hosts) {
				t.Errorf("WithURI(%s) hosts = %v, want %v", tt.uri, conf.hosts, tt.want.hosts)
			}
			if conf.sessionTimeout != tt.want.sessionTimeout {
				t.Errorf("WithURI(%s) sessionTimeout = %v, want %v", tt.uri, conf.sessionTimeout, tt.want.sessionTimeout)
			}
			if conf.basePath != tt.want.basePath {
				t.Errorf("WithURI(%s) basePath = %s, want %s", tt.uri, conf.basePath, tt.want.basePath)
			}
		})
	}
}

// Helper function to compare slices
func equalSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
