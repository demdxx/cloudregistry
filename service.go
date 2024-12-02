package cloudregistry

import "time"

// Ports is a representation of the ports exposed by a host or a service.
// The key is the protocol name and the value is the port used for this protocol.
// Typical usage is:
//
//	Ports{
//		"http":"80",
//		"https": "443",
//	}
type Ports map[string]string

// Host represents a host in the cloud registry.
type Host struct {
	Hostname string `json:"hostname"`
	Ports    Ports  `json:"ports"`
	User     string `json:"user,omitempty"`
	Password string `json:"password,omitempty"`
}

// Check represents a health check for a service.
type Check struct {
	ID  string
	TTL time.Duration
}

// Service represents a service in the cloud registry.
type Service struct {
	Name       string
	InstanceID string
	Hostname   string
	Port       int
	Public     []Host
	Private    []Host
	Check      Check
}

type ServiceInfo struct {
	Name       string    `json:"name"`
	InstanceID string    `json:"instance_id"`
	Hostname   string    `json:"hostname"`
	Port       int       `json:"port"`
	Public     []Host    `json:"public,omitempty"`
	Private    []Host    `json:"private,omitempty"`
	LastUpdate time.Time `json:"last_update"`
}
