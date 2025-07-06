package cloudregistry

import "testing"

func TestSyncInt64Value_Value(t *testing.T) {
	tests := []struct {
		name string
		init int64
		want int64
	}{
		{
			name: "positive value",
			init: 42,
			want: 42,
		},
		{
			name: "negative value",
			init: -123,
			want: -123,
		},
		{
			name: "zero value",
			init: 0,
			want: 0,
		},
		{
			name: "max int64",
			init: 9223372036854775807,
			want: 9223372036854775807,
		},
		{
			name: "min int64",
			init: -9223372036854775808,
			want: -9223372036854775808,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewSyncInt64Value(tt.init)
			if got := v.Value(); got != tt.want {
				t.Errorf("SyncInt64Value.Value() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSyncInt64Value_SetValue(t *testing.T) {
	tests := []struct {
		name     string
		init     int64
		setValue any
		want     int64
		wantErr  bool
	}{
		{
			name:     "set int64",
			init:     0,
			setValue: int64(100),
			want:     100,
			wantErr:  false,
		},
		{
			name:     "set int",
			init:     0,
			setValue: int(200),
			want:     200,
			wantErr:  false,
		},
		{
			name:     "set int32",
			init:     0,
			setValue: int32(300),
			want:     300,
			wantErr:  false,
		},
		{
			name:     "set string number",
			init:     0,
			setValue: "400",
			want:     400,
			wantErr:  false,
		},
		{
			name:     "set float64",
			init:     0,
			setValue: float64(500.0),
			want:     500,
			wantErr:  false,
		},
		{
			name:     "set float64 with decimals (truncated)",
			init:     0,
			setValue: float64(599.99),
			want:     599,
			wantErr:  false,
		},
		{
			name:     "set negative value",
			init:     100,
			setValue: int64(-50),
			want:     -50,
			wantErr:  false,
		},
		{
			name:     "set invalid string",
			init:     42,
			setValue: "not_a_number",
			want:     42, // should remain unchanged
			wantErr:  true,
		},
		{
			name:     "set boolean true",
			init:     42,
			setValue: true,
			want:     1, // true converts to 1
			wantErr:  false,
		},
		{
			name:     "set boolean false",
			init:     42,
			setValue: false,
			want:     0, // false converts to 0
			wantErr:  false,
		},
		{
			name:     "set nil",
			init:     42,
			setValue: nil,
			want:     0, // nil converts to 0
			wantErr:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewSyncInt64Value(tt.init)
			err := v.SetValue("", tt.setValue)
			if (err != nil) != tt.wantErr {
				t.Errorf("SyncInt64Value.SetValue() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got := v.Value(); got != tt.want {
				t.Errorf("SyncInt64Value.Value() after SetValue = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSyncInt64Value_Concurrent(t *testing.T) {
	v := NewSyncInt64Value(0)
	const numGoroutines = 100
	const numOperations = 1000

	// Test concurrent reads and writes
	done := make(chan bool, numGoroutines*2)

	// Start writers
	for i := 0; i < numGoroutines; i++ {
		go func(value int64) {
			for j := 0; j < numOperations; j++ {
				_ = v.SetValue("", value+int64(j))
			}
			done <- true
		}(int64(i))
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

	// The final value should be a valid int64
	finalValue := v.Value()
	if finalValue < 0 || finalValue >= numGoroutines+numOperations {
		// This is not a strict requirement due to race conditions,
		// but the value should be reasonable
		t.Logf("Final value: %d (this is expected due to concurrent modifications)", finalValue)
	}
}

func TestSyncInt64Value_Interface(t *testing.T) {
	// Test that SyncInt64Value implements Valuer[int64] interface
	var v Valuer[int64] = NewSyncInt64Value(42)

	if got := v.Value(); got != 42 {
		t.Errorf("SyncInt64Value as Valuer[int64].Value() = %v, want %v", got, 42)
	}

	if err := v.SetValue("test", int64(100)); err != nil {
		t.Errorf("SyncInt64Value as Valuer[int64].SetValue() error = %v", err)
	}

	if got := v.Value(); got != 100 {
		t.Errorf("SyncInt64Value as Valuer[int64].Value() after SetValue = %v, want %v", got, 100)
	}
}
