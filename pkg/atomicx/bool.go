package atomicx

import "sync/atomic"

// A Bool is an atomic implementation for boolean values.
type Bool uint32

// NewAtomicBool returns a Bool.
func NewAtomicBool() *Bool {
	return new(Bool)
}

// ForAtomicBool returns a Bool with given val.
func ForAtomicBool(val bool) *Bool {
	b := NewAtomicBool()
	b.Set(val)
	return b
}

// CompareAndSwap compares current value with given old, if equals, set to given val.
func (b *Bool) CompareAndSwap(old, val bool) bool {
	var ov, nv uint32
	if old {
		ov = 1
	}
	if val {
		nv = 1
	}
	return atomic.CompareAndSwapUint32((*uint32)(b), ov, nv)
}

// Set sets the value to v.
func (b *Bool) Set(v bool) {
	if v {
		atomic.StoreUint32((*uint32)(b), 1)
	} else {
		atomic.StoreUint32((*uint32)(b), 0)
	}
}

// True returns true if current value is true.
func (b *Bool) True() bool {
	return atomic.LoadUint32((*uint32)(b)) == 1
}
