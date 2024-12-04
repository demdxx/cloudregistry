package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/demdxx/cloudregistry"
	"github.com/demdxx/cloudregistry/consul"
	"github.com/demdxx/cloudregistry/etcd"
)

func main() {
	registryConnect := flag.String("registry", "", "Registry connection string")
	flag.Parse()

	if *registryConnect == "" {
		flag.Usage()
		return
	}

	fmt.Println("############################################")
	fmt.Println("### Cloud Registry Example")
	fmt.Println("### Registry:", *registryConnect)
	fmt.Println("############################################")

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	// Connect to the registry
	registry, err := connectRegistry(ctx, *registryConnect)
	if err != nil {
		log.Fatal("Connect registry", err)
		return
	}
	defer registry.Close()

	// Register a service
	service := &cloudregistry.Service{
		Name:       "example",
		Namespace:  "",
		Partition:  "",
		InstanceID: cloudregistry.GenerateInstanceID("example"),
		Hostname:   "localhost",
		Port:       8080,
		Check: cloudregistry.Check{
			ID:  "example",
			TTL: 10 * time.Second,
		},
	}

	// Register the service
	fmt.Println("Register service:", service.Name)
	if err := registry.Register(ctx, service); err != nil {
		log.Fatal("Register service", err)
		return
	}

	// Deregister the service
	defer func() {
		fmt.Println("Deregister service:", service.Name)
		if err := registry.Deregister(context.Background(), service.ID()); err != nil {
			log.Println("Deregister service", err)
		}
	}()

	// Subscribe to the value changes
	fmt.Println("Subscribe to the value changes")
	registry.SubscribeValueWithPrefix(ctx, "example/",
		cloudregistry.ValueSetterFunc(func(key string, value any) error {
			services, _ := registry.Discover(ctx, service.Prefix(), 10*time.Second)
			fmt.Printf("Service discovered: %d\n", len(services))
			fmt.Printf("Value changed: %s = %s\n", key, value)
			return nil
		}))

	// Update the value timer
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			val := time.Now().String()
			fmt.Println("Set new value:", val)
			if err := registry.Values(ctx, "example/").SetValue(ctx, "key", val); err != nil {
				log.Println("Set value", err)
			}
		}
	}
}

// Connect to the registry
func connectRegistry(ctx context.Context, conn string) (cloudregistry.Registry, error) {
	switch {
	case strings.HasPrefix(conn, "etcd://"):
		return etcd.Connect(ctx, etcd.WithURI(conn))
	case strings.HasPrefix(conn, "consul://"):
		return consul.Connect(ctx, consul.WithURI(conn))
	default:
		return nil, errors.New("unsupported registry connection string")
	}
}
