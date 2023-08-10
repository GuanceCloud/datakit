// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package typed

import "sync/atomic"

// Copy from go1.20.6:src/sync/atomic/type.go
// This type will be removed in the next version.

// A AtomicBool is an atomic boolean value.
// The zero value is false.
type AtomicBool struct {
	_ noCopy
	v uint32
}

// Load atomically loads and returns the value stored in x.
func (x *AtomicBool) Load() bool { return atomic.LoadUint32(&x.v) != 0 }

// Store atomically stores val into x.
func (x *AtomicBool) Store(val bool) { atomic.StoreUint32(&x.v, b32(val)) }

// Swap atomically stores new into x and returns the previous value.
func (x *AtomicBool) Swap(newBool bool) (oldBool bool) {
	return atomic.SwapUint32(&x.v, b32(newBool)) != 0
}

// CompareAndSwap executes the compare-and-swap operation for the boolean value x.
func (x *AtomicBool) CompareAndSwap(oldBool, newBool bool) (swapped bool) {
	return atomic.CompareAndSwapUint32(&x.v, b32(oldBool), b32(newBool))
}

// b32 returns a uint32 0 or 1 representing b.
func b32(b bool) uint32 {
	if b {
		return 1
	}
	return 0
}

// noCopy may be added to structs which must not be copied
// after the first use.
//
// See https://golang.org/issues/8005#issuecomment-190753527
// for details.
//
// Note that it must not be embedded, due to the Lock and Unlock methods.
type noCopy struct{}

// Lock is a no-op used by -copylocks checker from `go vet`.
func (*noCopy) Lock()   {}
func (*noCopy) Unlock() {}
