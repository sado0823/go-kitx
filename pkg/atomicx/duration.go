package atomicx

import (
	"sync/atomic"
	"time"
)

// A Duration is an implementation of atomic duration.
type Duration int64

// NewAtomicDuration returns a Duration.
func NewAtomicDuration() *Duration {
	return new(Duration)
}

// ForAtomicDuration returns a Duration with given value.
func ForAtomicDuration(val time.Duration) *Duration {
	d := NewAtomicDuration()
	d.Set(val)
	return d
}

// CompareAndSwap compares current value with old, if equals, set the value to val.
func (d *Duration) CompareAndSwap(old, val time.Duration) bool {
	return atomic.CompareAndSwapInt64((*int64)(d), int64(old), int64(val))
}

// Load loads the current duration.
func (d *Duration) Load() time.Duration {
	return time.Duration(atomic.LoadInt64((*int64)(d)))
}

// Set sets the value to val.
func (d *Duration) Set(val time.Duration) {
	atomic.StoreInt64((*int64)(d), int64(val))
}
