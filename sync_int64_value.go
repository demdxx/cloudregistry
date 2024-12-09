package cloudregistry

import (
	"sync/atomic"

	"github.com/demdxx/gocast/v2"
)

// SyncInt64Value is a thread-safe int64 value holder.
type SyncInt64Value struct {
	val int64
}

// NewSyncInt64Value creates a new SyncInt64Value with the given value.
func NewSyncInt64Value(val int64) *SyncInt64Value {
	return &SyncInt64Value{val: val}
}

// Value returns the value.
func (v *SyncInt64Value) Value() int64 {
	return atomic.LoadInt64(&v.val)
}

// SetValue sets the value to the given value.
func (v *SyncInt64Value) SetValue(_ string, val any) error {
	nval, err := gocast.TryNumber[int64](val)
	if err == nil {
		atomic.StoreInt64(&v.val, nval)
	}
	return err
}

var _ Valuer[int64] = (*SyncInt64Value)(nil)
