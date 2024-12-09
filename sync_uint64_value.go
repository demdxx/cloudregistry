package cloudregistry

import (
	"sync/atomic"

	"github.com/demdxx/gocast/v2"
)

// SyncUInt64Value is a thread-safe uint64 value holder.
type SyncUInt64Value struct {
	val uint64
}

// NewSyncUInt64Value creates a new SyncUInt64Value with the given value.
func NewSyncUInt64Value(val uint64) *SyncUInt64Value {
	return &SyncUInt64Value{val: val}
}

// Value returns the value.
func (v *SyncUInt64Value) Value() uint64 {
	return atomic.LoadUint64(&v.val)
}

// SetValue sets the value to the given value.
func (v *SyncUInt64Value) SetValue(_ string, val any) error {
	nval, err := gocast.TryNumber[uint64](val)
	if err == nil {
		atomic.StoreUint64(&v.val, nval)
	}
	return err
}

var _ Valuer[uint64] = (*SyncUInt64Value)(nil)
