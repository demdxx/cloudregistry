package cloudregistry

import (
	"fmt"
	"reflect"
	"sync/atomic"
	"testing"
)

func TestSyncAtomicValue_Value(t *testing.T) {
	tests := []struct {
		name  string
		value any
		want  any
	}{
		{
			name:  "Test 1",
			value: "str",
			want:  "str",
		},
		{
			name:  "Test 2",
			value: 10,
			want:  10,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewSyncAtomicValue(&atomic.Value{})
			if err := v.SetValue("", tt.value); err != nil {
				t.Errorf("SyncAtomicValue.SetValue() error = %v", err)
				return
			}
			if got := v.Value(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SyncAtomicValue.Value() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSyncAtomicValue_Int64(t *testing.T) {
	tests := []struct {
		name  string
		value any
		want  int64
		err   bool
	}{
		{
			name:  "Test 1",
			value: "str",
			want:  -1,
			err:   true,
		},
		{
			name:  "Test 2",
			value: int64(10),
			want:  10,
			err:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewSyncInt64Value(-1)
			if err := v.SetValue("", tt.value); err != nil {
				if !tt.err {
					t.Errorf("SyncAtomicValue.SetValue() error = %v", err)
					return
				}
			} else if tt.err {
				t.Errorf("SyncAtomicValue.SetValue() expected error")
				return
			}
			if got := v.Value(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SyncAtomicValue.Value() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSyncAtomicValue_Uint64(t *testing.T) {
	tests := []struct {
		name  string
		value any
		want  uint64
		err   bool
	}{
		{
			name:  "Test 1",
			value: "str",
			want:  0,
			err:   true,
		},
		{
			name:  "Test 2",
			value: uint64(10),
			want:  10,
			err:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewSyncUInt64Value(0)
			if err := v.SetValue("", tt.value); err != nil {
				if !tt.err {
					t.Errorf("SyncAtomicValue.SetValue() error = %v", err)
					return
				}
			} else if tt.err {
				t.Errorf("SyncAtomicValue.SetValue() expected error")
				return
			}
			if got := v.Value(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SyncAtomicValue.Value() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSyncAtomicValue_String(t *testing.T) {
	tests := []struct {
		name  string
		value any
		want  string
		err   bool
	}{
		{
			name:  "Test 1",
			value: 10,
			want:  "10",
			err:   false,
		},
		{
			name:  "Test 2",
			value: "str",
			want:  "str",
			err:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewSyncValue("")
			if err := v.SetValue("", tt.value); err != nil {
				if !tt.err {
					t.Errorf("SyncAtomicValue.SetValue() error = %v", err)
					return
				}
			} else if tt.err {
				t.Errorf("SyncAtomicValue.SetValue() expected error")
				return
			}
			if got := v.Value(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SyncAtomicValue.Value() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSyncValue_Value(t *testing.T) {
	tests := []struct {
		name string
		init string
		want string
	}{
		{
			name: "string value",
			init: "hello",
			want: "hello",
		},
		{
			name: "empty string",
			init: "",
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewSyncValue(tt.init)
			if got := v.Value(); got != tt.want {
				t.Errorf("SyncValue.Value() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSyncValue_SetValue(t *testing.T) {
	tests := []struct {
		name     string
		init     string
		setValue any
		want     string
		wantErr  bool
	}{
		{
			name:     "set string",
			init:     "initial",
			setValue: "new_value",
			want:     "new_value",
			wantErr:  false,
		},
		{
			name:     "set int to string",
			init:     "initial",
			setValue: 42,
			want:     "42",
			wantErr:  false,
		},
		{
			name:     "set bool to string",
			init:     "initial",
			setValue: true,
			want:     "true",
			wantErr:  false,
		},
		{
			name:     "set nil",
			init:     "initial",
			setValue: nil,
			want:     "",
			wantErr:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewSyncValue(tt.init)
			err := v.SetValue("", tt.setValue)
			if (err != nil) != tt.wantErr {
				t.Errorf("SyncValue.SetValue() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got := v.Value(); got != tt.want {
				t.Errorf("SyncValue.Value() after SetValue = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSyncValue_Concurrent(t *testing.T) {
	v := NewSyncValue("initial")
	const numGoroutines = 50
	const numOperations = 100

	// Test concurrent reads and writes
	done := make(chan bool, numGoroutines*2)

	// Start writers
	for i := 0; i < numGoroutines; i++ {
		go func(value string) {
			for j := 0; j < numOperations; j++ {
				_ = v.SetValue("", value)
			}
			done <- true
		}(fmt.Sprintf("value-%d", i))
	}

	// Start readers
	for i := 0; i < numGoroutines; i++ {
		go func() {
			for j := 0; j < numOperations; j++ {
				_ = v.Value()
			}
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines*2; i++ {
		<-done
	}

	// The final value should be a valid string
	finalValue := v.Value()
	if finalValue == "" {
		t.Error("Final value should not be empty after concurrent operations")
	}
}

func TestSyncValue_Interface(t *testing.T) {
	// Test that SyncValue implements Valuer interface
	var v Valuer[string] = NewSyncValue("test")

	if got := v.Value(); got != "test" {
		t.Errorf("SyncValue as Valuer[string].Value() = %v, want %v", got, "test")
	}

	if err := v.SetValue("test", "new_value"); err != nil {
		t.Errorf("SyncValue as Valuer[string].SetValue() error = %v", err)
	}

	if got := v.Value(); got != "new_value" {
		t.Errorf("SyncValue as Valuer[string].Value() after SetValue = %v, want %v", got, "new_value")
	}
}
