package cloudregistry

import (
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
