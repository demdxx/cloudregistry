package cloudregistry

import (
	"strings"
	"testing"
)

func TestGenerateInstanceID(t *testing.T) {
	tests := []struct {
		name        string
		serviceName string
	}{
		{
			name:        "basic service name",
			serviceName: "my-service",
		},
		{
			name:        "empty service name",
			serviceName: "",
		},
		{
			name:        "service name with special characters",
			serviceName: "my-service_v2.0",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GenerateInstanceID(tt.serviceName)

			// Check that it starts with the service name
			if !strings.HasPrefix(got, tt.serviceName) {
				t.Errorf("GenerateInstanceID() = %v, should start with %v", got, tt.serviceName)
			}

			// Check that it has the dash separator
			if tt.serviceName != "" && !strings.Contains(got, "-") {
				t.Errorf("GenerateInstanceID() = %v, should contain dash separator", got)
			}

			// Check that it's not empty
			if got == "" {
				t.Errorf("GenerateInstanceID() should not be empty")
			}

			// Check that multiple calls produce different IDs (with high probability)
			got2 := GenerateInstanceID(tt.serviceName)
			if got == got2 {
				// This could happen due to randomness, but let's try a few more times
				allSame := true
				for i := 0; i < 10; i++ {
					if GenerateInstanceID(tt.serviceName) != got {
						allSame = false
						break
					}
				}
				if allSame {
					t.Errorf("GenerateInstanceID() appears to not be random - got same value %v multiple times", got)
				}
			}
		})
	}
}

func TestGenerateInstanceID_Format(t *testing.T) {
	serviceName := "test-service"
	instanceID := GenerateInstanceID(serviceName)

	// Should be in format "service-name-number"
	parts := strings.Split(instanceID, "-")
	if len(parts) < 2 {
		t.Errorf("GenerateInstanceID() = %v, should have at least 2 parts separated by dash", instanceID)
	}

	// Last part should be a number
	lastPart := parts[len(parts)-1]
	if len(lastPart) == 0 {
		t.Errorf("GenerateInstanceID() = %v, last part should not be empty", instanceID)
	}

	// First parts should match the service name
	expectedPrefix := strings.Join(parts[:len(parts)-1], "-")
	if expectedPrefix != serviceName {
		t.Errorf("GenerateInstanceID() = %v, prefix should be %v but got %v", instanceID, serviceName, expectedPrefix)
	}
}
