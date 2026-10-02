package consul

import (
	"testing"
	"time"

	"github.com/demdxx/cloudregistry"
)

func TestAgentServiceCheckHTTPPath(t *testing.T) {
	check := agentServiceCheck(&cloudregistry.Service{
		Port: 8080,
		Check: cloudregistry.Check{
			ID:  "health",
			TTL: 20 * time.Second,
			HTTP: struct {
				URL     string
				Method  string
				Headers map[string][]string
			}{
				URL:    "/health",
				Method: "GET",
			},
		},
	})

	if check == nil {
		t.Fatal("expected HTTP check")
	}
	if check.CheckID != "health" {
		t.Fatalf("CheckID = %q", check.CheckID)
	}
	if check.HTTP != "http://127.0.0.1:8080/health" {
		t.Fatalf("HTTP = %q", check.HTTP)
	}
	if check.Method != "GET" {
		t.Fatalf("Method = %q", check.Method)
	}
	if check.Interval != "20s" {
		t.Fatalf("Interval = %q", check.Interval)
	}
	if check.Timeout != "5s" {
		t.Fatalf("Timeout = %q", check.Timeout)
	}
	if check.DeregisterCriticalServiceAfter != "60s" {
		t.Fatalf("DeregisterCriticalServiceAfter = %q", check.DeregisterCriticalServiceAfter)
	}
	if check.TTL != "" {
		t.Fatalf("TTL = %q, want empty", check.TTL)
	}
}

func TestAgentServiceCheckAbsoluteURL(t *testing.T) {
	check := agentServiceCheck(&cloudregistry.Service{
		Hostname: "revolvesyndicate.net",
		Port:     8080,
		Check: cloudregistry.Check{
			ID:  "health",
			TTL: 20 * time.Second,
			HTTP: struct {
				URL     string
				Method  string
				Headers map[string][]string
			}{
				URL: "https://example.test/health",
			},
		},
	})

	if check == nil {
		t.Fatal("expected HTTP check")
	}
	if check.HTTP != "https://example.test/health" {
		t.Fatalf("HTTP = %q", check.HTTP)
	}
	if check.Interval != "20s" {
		t.Fatalf("Interval = %q", check.Interval)
	}
	if check.TTL != "" {
		t.Fatalf("TTL = %q, want empty", check.TTL)
	}
}

func TestAgentServiceCheckTTLOnly(t *testing.T) {
	check := agentServiceCheck(&cloudregistry.Service{
		Check: cloudregistry.Check{
			ID:  "example",
			TTL: 10 * time.Second,
		},
	})

	if check == nil {
		t.Fatal("expected TTL check")
	}
	if check.CheckID != "example" {
		t.Fatalf("CheckID = %q", check.CheckID)
	}
	if check.TTL != "10s" {
		t.Fatalf("TTL = %q", check.TTL)
	}
	if check.DeregisterCriticalServiceAfter != "30s" {
		t.Fatalf("DeregisterCriticalServiceAfter = %q", check.DeregisterCriticalServiceAfter)
	}
	if check.HTTP != "" {
		t.Fatalf("HTTP = %q, want empty", check.HTTP)
	}
	if check.Interval != "" {
		t.Fatalf("Interval = %q, want empty", check.Interval)
	}
}

func TestAgentServiceCheckNone(t *testing.T) {
	if check := agentServiceCheck(&cloudregistry.Service{}); check != nil {
		t.Fatalf("expected nil check, got %+v", check)
	}
}

func TestAgentServiceCheckHTTPDefaultInterval(t *testing.T) {
	check := agentServiceCheck(&cloudregistry.Service{
		Port: 8080,
		Check: cloudregistry.Check{
			HTTP: struct {
				URL     string
				Method  string
				Headers map[string][]string
			}{
				URL: "health",
			},
		},
	})

	if check == nil {
		t.Fatal("expected HTTP check")
	}
	if check.HTTP != "http://127.0.0.1:8080/health" {
		t.Fatalf("HTTP = %q", check.HTTP)
	}
	if check.Interval != "10s" {
		t.Fatalf("Interval = %q", check.Interval)
	}
	if check.Timeout != "5s" {
		t.Fatalf("Timeout = %q", check.Timeout)
	}
	if check.TTL != "" {
		t.Fatalf("TTL = %q, want empty", check.TTL)
	}
}
