package cloudregistry

import (
	"sync/atomic"

	"github.com/demdxx/gocast/v2"
)

type syncValue[T any] interface {
	Load() T
	Store(T)
}

// SyncAtomicValue is a thread-safe value holder.
type SyncAtomicValue[T syncValue[V], V any] struct {
	val T
}

// NewSyncAtomicValue creates a new SyncAtomicValue with the given value.
func NewSyncAtomicValue[T syncValue[V], V any](vl T) *SyncAtomicValue[T, V] {
	return &SyncAtomicValue[T, V]{val: vl}
}

// Value returns the value.
func (v *SyncAtomicValue[T, V]) Value() V {
	return v.val.Load()
}

// SetValue sets the value to the given value.
func (v *SyncAtomicValue[T, V]) SetValue(_ string, val any) error {
	vl, err := gocast.TryCast[V](val)
	if err == nil {
		v.val.Store(vl)
	}
	return err
}

var _ Valuer[any] = (*SyncAtomicValue[*atomic.Value, any])(nil)
