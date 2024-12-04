package cloudregistry

import (
	"strings"
	"time"
)

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
	// Health checks can be of different types and applicable in only some drivers.
	HTTP struct {
		URL     string
		Method  string
		Headers map[string][]string
	}
}

type ServicePrefix struct {
	Name      string
	Namespace string
	Partition string
}

// String returns the string representation of the service prefix.
func (prefix *ServicePrefix) String() string {
	nameBuff := strings.Builder{}
	_, _ = nameBuff.WriteString("services/")
	if prefix.Namespace != "" {
		_, _ = nameBuff.WriteString(prefix.Namespace)
		_ = nameBuff.WriteByte('/')
	}
	_, _ = nameBuff.WriteString(prefix.Name)
	_ = nameBuff.WriteByte('/')
	if prefix.Partition != "" {
		_, _ = nameBuff.WriteString(prefix.Partition)
		_ = nameBuff.WriteByte('/')
	}
	return nameBuff.String()
}

type ServiceID struct {
	Name       string
	Namespace  string
	Partition  string
	InstanceID string
}

// String returns the string representation of the service ID.
func (id *ServiceID) String() string {
	nameBuff := strings.Builder{}
	_, _ = nameBuff.WriteString("services/")
	if id.Namespace != "" {
		_, _ = nameBuff.WriteString(id.Namespace)
		_ = nameBuff.WriteByte('/')
	}
	_, _ = nameBuff.WriteString(id.Name)
	_ = nameBuff.WriteByte('/')
	if id.Partition != "" {
		_, _ = nameBuff.WriteString(id.Partition)
		_ = nameBuff.WriteByte('/')
	}
	return nameBuff.String()
}

// Prefix returns the service prefix of the service ID.
func (id *ServiceID) Prefix() *ServicePrefix {
	return &ServicePrefix{
		Name:      id.Name,
		Namespace: id.Namespace,
		Partition: id.Partition,
	}
}

// ServiceInfo represents a service in the cloud registry.
type ServiceInfo struct {
	Name       string            `json:"name"`
	Namespace  string            `json:"namespace,omitempty"`
	Partition  string            `json:"partition,omitempty"`
	InstanceID string            `json:"instance_id"`
	Hostname   string            `json:"hostname"`
	Port       int               `json:"port"`
	Public     []Host            `json:"public,omitempty"`
	Private    []Host            `json:"private,omitempty"`
	Tags       []string          `json:"tags,omitempty"`
	Meta       map[string]string `json:"meta,omitempty"`
	RawInfo    any               `json:"raw_info,omitempty"`
	LastUpdate time.Time         `json:"last_update"`
}

// Service represents a service in the cloud registry.
type Service struct {
	Name       string
	Namespace  string
	Partition  string
	InstanceID string
	Hostname   string
	Port       int
	Public     []Host
	Private    []Host
	Tags       []string
	Meta       map[string]string
	Check      Check
}

// ID returns the service ID of the service.
func (service *Service) ID() *ServiceID {
	return &ServiceID{
		Name:       service.Name,
		Namespace:  service.Namespace,
		Partition:  service.Partition,
		InstanceID: service.InstanceID,
	}
}

// Prefix returns the service prefix of the service.
func (service *Service) Prefix() *ServicePrefix {
	return &ServicePrefix{
		Name:      service.Name,
		Namespace: service.Namespace,
		Partition: service.Partition,
	}
}
