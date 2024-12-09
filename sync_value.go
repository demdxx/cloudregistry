package cloudregistry

import (
	"sync"

	"github.com/demdxx/gocast/v2"
)

// SyncValue is a thread-safe value holder.
type SyncValue[T any] struct {
	mx  sync.RWMutex
	val T
}

// NewSyncValue creates a new SyncValue with the given value.
func NewSyncValue[T any](val T) *SyncValue[T] {
	return &SyncValue[T]{val: val}
}

// Value returns the value.
func (v *SyncValue[T]) Value() T {
	v.mx.RLock()
	defer v.mx.RUnlock()
	return v.val
}

// SetValue sets the value to the given value.
func (v *SyncValue[T]) SetValue(_ string, val any) (err error) {
	v.mx.Lock()
	defer v.mx.Unlock()
	v.val, err = gocast.TryCast[T](val)
	return err
}

var _ Valuer[any] = (*SyncValue[any])(nil)
