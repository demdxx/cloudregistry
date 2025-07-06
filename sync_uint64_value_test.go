package cloudregistry

import "testing"

func TestSyncUInt64Value_Value(t *testing.T) {
	tests := []struct {
		name string
		init uint64
		want uint64
	}{
		{
			name: "positive value",
			init: 42,
			want: 42,
		},
		{
			name: "zero value",
			init: 0,
			want: 0,
		},
		{
			name: "max uint64",
			init: 18446744073709551615, // 2^64 - 1
			want: 18446744073709551615,
		},
		{
			name: "large value",
			init: 9223372036854775808, // 2^63
			want: 9223372036854775808,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewSyncUInt64Value(tt.init)
			if got := v.Value(); got != tt.want {
				t.Errorf("SyncUInt64Value.Value() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSyncUInt64Value_SetValue(t *testing.T) {
	tests := []struct {
		name     string
		init     uint64
		setValue any
		want     uint64
		wantErr  bool
	}{
		{
			name:     "set uint64",
			init:     0,
			setValue: uint64(100),
			want:     100,
			wantErr:  false,
		},
		{
			name:     "set uint",
			init:     0,
			setValue: uint(200),
			want:     200,
			wantErr:  false,
		},
		{
			name:     "set uint32",
			init:     0,
			setValue: uint32(300),
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
			name:     "set positive int",
			init:     0,
			setValue: int(50),
			want:     50,
			wantErr:  false,
		},
		{
			name:     "set negative int (converts via overflow)",
			init:     42,
			setValue: int(-50),
			want:     18446744073709551566, // -50 as uint64 (two's complement)
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
			v := NewSyncUInt64Value(tt.init)
			err := v.SetValue("", tt.setValue)
			if (err != nil) != tt.wantErr {
				t.Errorf("SyncUInt64Value.SetValue() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got := v.Value(); got != tt.want {
				t.Errorf("SyncUInt64Value.Value() after SetValue = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSyncUInt64Value_Interface(t *testing.T) {
	// Test that SyncUInt64Value implements Valuer[uint64] interface
	var v Valuer[uint64] = NewSyncUInt64Value(42)

	if got := v.Value(); got != 42 {
		t.Errorf("SyncUInt64Value as Valuer[uint64].Value() = %v, want %v", got, 42)
	}

	if err := v.SetValue("test", uint64(100)); err != nil {
		t.Errorf("SyncUInt64Value as Valuer[uint64].SetValue() error = %v", err)
	}

	if got := v.Value(); got != 100 {
		t.Errorf("SyncUInt64Value as Valuer[uint64].Value() after SetValue = %v, want %v", got, 100)
	}
}
