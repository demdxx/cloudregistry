package cloudregistry

import (
	"reflect"
	"testing"
)

func TestServicePrefix_String(t *testing.T) {
	tests := []struct {
		name   string
		prefix ServicePrefix
		want   string
	}{
		{
			name: "only service name",
			prefix: ServicePrefix{
				Name: "my-service",
			},
			want: "services/my-service/",
		},
		{
			name: "with namespace",
			prefix: ServicePrefix{
				Name:      "my-service",
				Namespace: "production",
			},
			want: "services/production/my-service/",
		},
		{
			name: "with partition",
			prefix: ServicePrefix{
				Name:      "my-service",
				Partition: "region-1",
			},
			want: "services/my-service/region-1/",
		},
		{
			name: "with namespace and partition",
			prefix: ServicePrefix{
				Name:      "my-service",
				Namespace: "production",
				Partition: "region-1",
			},
			want: "services/production/my-service/region-1/",
		},
		{
			name: "empty service name",
			prefix: ServicePrefix{
				Name: "",
			},
			want: "services//",
		},
		{
			name:   "all empty",
			prefix: ServicePrefix{},
			want:   "services//",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.prefix.String(); got != tt.want {
				t.Errorf("ServicePrefix.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestServiceID_String(t *testing.T) {
	tests := []struct {
		name string
		id   ServiceID
		want string
	}{
		{
			name: "basic service ID",
			id: ServiceID{
				Name:       "my-service",
				InstanceID: "instance-123",
			},
			want: "services/my-service/",
		},
		{
			name: "with namespace",
			id: ServiceID{
				Name:       "my-service",
				Namespace:  "production",
				InstanceID: "instance-123",
			},
			want: "services/production/my-service/",
		},
		{
			name: "with partition",
			id: ServiceID{
				Name:       "my-service",
				Partition:  "region-1",
				InstanceID: "instance-123",
			},
			want: "services/my-service/region-1/",
		},
		{
			name: "with namespace and partition",
			id: ServiceID{
				Name:       "my-service",
				Namespace:  "production",
				Partition:  "region-1",
				InstanceID: "instance-123",
			},
			want: "services/production/my-service/region-1/",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.id.String(); got != tt.want {
				t.Errorf("ServiceID.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestServiceID_Prefix(t *testing.T) {
	tests := []struct {
		name string
		id   ServiceID
		want *ServicePrefix
	}{
		{
			name: "basic service ID",
			id: ServiceID{
				Name:       "my-service",
				InstanceID: "instance-123",
			},
			want: &ServicePrefix{
				Name: "my-service",
			},
		},
		{
			name: "with namespace and partition",
			id: ServiceID{
				Name:       "my-service",
				Namespace:  "production",
				Partition:  "region-1",
				InstanceID: "instance-123",
			},
			want: &ServicePrefix{
				Name:      "my-service",
				Namespace: "production",
				Partition: "region-1",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.id.Prefix(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ServiceID.Prefix() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestService_ID(t *testing.T) {
	tests := []struct {
		name    string
		service Service
		want    *ServiceID
	}{
		{
			name: "basic service",
			service: Service{
				Name:       "my-service",
				InstanceID: "instance-123",
				Hostname:   "localhost",
				Port:       8080,
			},
			want: &ServiceID{
				Name:       "my-service",
				InstanceID: "instance-123",
			},
		},
		{
			name: "service with namespace and partition",
			service: Service{
				Name:       "my-service",
				Namespace:  "production",
				Partition:  "region-1",
				InstanceID: "instance-123",
				Hostname:   "localhost",
				Port:       8080,
			},
			want: &ServiceID{
				Name:       "my-service",
				Namespace:  "production",
				Partition:  "region-1",
				InstanceID: "instance-123",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.service.ID(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Service.ID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestService_Prefix(t *testing.T) {
	tests := []struct {
		name    string
		service Service
		want    *ServicePrefix
	}{
		{
			name: "basic service",
			service: Service{
				Name:       "my-service",
				InstanceID: "instance-123",
				Hostname:   "localhost",
				Port:       8080,
			},
			want: &ServicePrefix{
				Name: "my-service",
			},
		},
		{
			name: "service with namespace and partition",
			service: Service{
				Name:       "my-service",
				Namespace:  "production",
				Partition:  "region-1",
				InstanceID: "instance-123",
				Hostname:   "localhost",
				Port:       8080,
			},
			want: &ServicePrefix{
				Name:      "my-service",
				Namespace: "production",
				Partition: "region-1",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.service.Prefix(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Service.Prefix() = %v, want %v", got, tt.want)
			}
		})
	}
}
