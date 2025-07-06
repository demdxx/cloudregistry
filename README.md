# CloudRegistry

[![Build Status](https://github.com/demdxx/cloudregistry/workflows/Tests/badge.svg)](https://github.com/demdxx/cloudregistry/actions?workflow=Tests)
[![Go Report Card](https://goreportcard.com/badge/github.com/demdxx/cloudregistry)](https://goreportcard.com/report/github.com/demdxx/cloudregistry)
[![GoDoc](https://godoc.org/github.com/demdxx/cloudregistry?status.svg)](https://godoc.org/github.com/demdxx/cloudregistry)
[![Coverage Status](https://coveralls.io/repos/github/demdxx/cloudregistry/badge.svg)](https://coveralls.io/github/demdxx/cloudregistry)

CloudRegistry is a versatile Go package designed to facilitate interaction with various cloud service registries. Currently supporting **etcd**, with plans to integrate **Consul** and **ZooKeeper**, CloudRegistry provides a unified interface for service registration, discovery, and health monitoring.

## Features

- **Service Registration & Deregistration**: Easily register and deregister services in your chosen cloud registry.
- **Service Discovery**: Discover available services with support for TTL-based caching.
- **Health Checks**: Implement health checks to ensure service reliability.
- **Flexible Backend Support**: Currently supports etcd with upcoming support for Consul and ZooKeeper.
- **Subscription Mechanism**: Subscribe to value changes with or without prefixes.

## Supported Registries

- **etcd** *(Supported)*
- **Consul** *(Supported)*
- **ZooKeeper** *(Supported)*

## Installation

To install CloudRegistry, use `go get`:

```bash
go get github.com/demdxx/cloudregistry
```

## Usage

### Importing the Package

```go
import "github.com/demdxx/cloudregistry"
```

### Initializing the Registry

```go
import (
    "context"
    "log"
    "time"

    "github.com/demdxx/cloudregistry"
    "github.com/demdxx/cloudregistry/etcd"
    "github.com/demdxx/cloudregistry/consul"
    "github.com/demdxx/cloudregistry/zookeeper"
)

func main() {
    ctx := context.Background()

    // Initialize etcd registry
    etcdRegistry, err := etcd.Connect(ctx, etcd.WithURI("etcd://localhost:2379"))
    if err != nil {
        log.Fatalf("Failed to initialize etcd registry: %v", err)
    }
    defer etcdRegistry.Close()

    // Initialize Consul registry
    consulRegistry, err := consul.Connect(ctx, consul.WithURI("consul://localhost:8500"))
    if err != nil {
        log.Fatalf("Failed to initialize consul registry: %v", err)
    }
    defer consulRegistry.Close()

    // Initialize ZooKeeper registry
    zkRegistry, err := zookeeper.Connect(ctx, zookeeper.WithURI("zk://localhost:2181"))
    if err != nil {
        log.Fatalf("Failed to initialize zookeeper registry: %v", err)
    }
    defer zkRegistry.Close()

    // Use any registry (they all implement the same interface)
    registry := etcdRegistry // or consulRegistry or zkRegistry

    // Example service registration
    service := &cloudregistry.Service{
        Name:       "example-service",
        InstanceID: cloudregistry.GenerateInstanceID("example-service"),
        Hostname:   "localhost",
        Port:       8080,
        Public: []cloudregistry.Host{
            {
                Hostname: "localhost",
                Ports: cloudregistry.Ports{
                    "http": "80",
                },
            },
        },
        Check: cloudregistry.Check{
            ID:  "service-check",
            TTL: 30 * time.Second,
        },
    }

    if err := registry.Register(ctx, service); err != nil {
        log.Fatalf("Failed to register service: %v", err)
    }

    // Discover services
    services, err := registry.Discover(ctx, "example-service", 60*time.Second)
    if err != nil {
        log.Fatalf("Failed to discover services: %v", err)
    }

    for _, svc := range services {
        log.Printf("Discovered service: %s at %s:%d", svc.Name, svc.Hostname, svc.Port)
    }

    // Perform health check
    if err := registry.HealthCheck(ctx, service.Name, service.InstanceID, 30*time.Second); err != nil {
        log.Fatalf("Health check failed: %v", err)
    }
}
```

### Interfaces and Types

#### `Registry` Interface

Provides methods for service registration, deregistration, discovery, and health checks.

```go
type Registry interface {
    io.Closer
    ValueClient
    Register(ctx context.Context, service *Service) error
    Deregister(ctx context.Context, name, id string) error
    Discover(ctx context.Context, name string, TTL time.Duration) ([]*ServiceInfo, error)
    HealthCheck(ctx context.Context, name, id string, TTL time.Duration) error
}
```

#### `ValueClient` Interface

Handles key-value interactions with the registry.

```go
type ValueClient interface {
    Values(ctx context.Context, prefix ...string) ValueClient
    Value(ctx context.Context, name string) (string, error)
    SetValue(ctx context.Context, name, value string) error
    SubscribeValue(ctx context.Context, name string, val ValueSetter) error
    SubscribeValueWithPrefix(ctx context.Context, prefix string, val ValueSetter) error
}
```

#### Helper Functions

- `GenerateInstanceID(serviceName string) string`: Generates a pseudo-random service instance identifier.

```go
func GenerateInstanceID(serviceName string) string {
    return fmt.Sprintf("%s-%d", serviceName, rand.New(rand.NewSource(time.Now().UnixNano())).Int())
}
```

## Contributing

Contributions are welcome! Please open an issue or submit a pull request for any enhancements or bug fixes.

### TODO

- [ ] **Additional Features**: Expand subscription mechanisms and enhance error handling.
- [ ] **Performance Optimizations**: Implement connection pooling and caching strategies.
- [ ] **Monitoring**: Add metrics and health check endpoints.

## License

[Licensed under the Apache License, Version 2.0](LICENSE)

  ```sh
   Copyright [2024] Dmitry Ponomarev <demdxx@gmail.com>

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
  ```
