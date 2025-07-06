package cloudregistry

import (
	"errors"
	"testing"
)

func TestValueSetterFunc_SetValue(t *testing.T) {
	tests := []struct {
		name    string
		fn      ValueSetterFunc
		key     string
		value   any
		wantErr bool
	}{
		{
			name: "successful set",
			fn: func(key string, value any) error {
				if key == "test-key" && value == "test-value" {
					return nil
				}
				return errors.New("unexpected arguments")
			},
			key:     "test-key",
			value:   "test-value",
			wantErr: false,
		},
		{
			name: "error case",
			fn: func(key string, value any) error {
				return errors.New("test error")
			},
			key:     "any-key",
			value:   "any-value",
			wantErr: true,
		},
		{
			name: "nil value",
			fn: func(key string, value any) error {
				if value == nil {
					return nil
				}
				return errors.New("expected nil")
			},
			key:     "test-key",
			value:   nil,
			wantErr: false,
		},
		{
			name: "different types",
			fn: func(key string, value any) error {
				switch v := value.(type) {
				case int:
					if v == 42 {
						return nil
					}
				case string:
					if v == "hello" {
						return nil
					}
				case bool:
					if v == true {
						return nil
					}
				}
				return errors.New("unexpected value")
			},
			key:     "test-key",
			value:   42,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.fn.SetValue(tt.key, tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValueSetterFunc.SetValue() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValueSetterFunc_Interface(t *testing.T) {
	// Test that ValueSetterFunc implements ValueSetter interface
	var setter ValueSetter = ValueSetterFunc(func(key string, value any) error {
		return nil
	})

	err := setter.SetValue("test", "value")
	if err != nil {
		t.Errorf("ValueSetterFunc as ValueSetter.SetValue() error = %v", err)
	}
}

func TestValueSetterFunc_Captures(t *testing.T) {
	// Test that the function can capture and modify external state
	captured := make(map[string]any)

	setter := ValueSetterFunc(func(key string, value any) error {
		captured[key] = value
		return nil
	})

	testCases := []struct {
		key   string
		value any
	}{
		{"key1", "value1"},
		{"key2", 42},
		{"key3", true},
		{"key4", nil},
	}

	for _, tc := range testCases {
		err := setter.SetValue(tc.key, tc.value)
		if err != nil {
			t.Errorf("ValueSetterFunc.SetValue(%v, %v) error = %v", tc.key, tc.value, err)
		}
	}

	// Verify all values were captured
	for _, tc := range testCases {
		if captured[tc.key] != tc.value {
			t.Errorf("Expected captured[%v] = %v, got %v", tc.key, tc.value, captured[tc.key])
		}
	}
}
