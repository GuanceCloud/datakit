// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2024-present Guance, Inc.

package io

import (
	"sync"
	"sync/atomic"
)

type SafeSlice[S ~[]T, T any] struct {
	endSize atomic.Int64
	trySize atomic.Int64
	slice   S
	lock    sync.RWMutex
}

func NewSafeSlice[S ~[]T, T any](bufSize int64) *SafeSlice[S, T] {
	a := &SafeSlice[S, T]{}
	a.slice = make(S, bufSize)
	a.endSize.Store(-1)
	return a
}

func (a *SafeSlice[S, T]) Reset(isClear bool) S {
	a.lock.Lock()
	defer a.lock.Unlock()
	l := a.endSize.Load()
	var data S
	if l == -1 {
		l = a.trySize.Load()
	}
	data = a.slice[:l]

	res := make(S, len(data))
	copy(res, data)
	a.endSize.Store(-1)
	a.trySize.Store(0)
	if isClear {
		// If it's a ptr, it's useful to set it to nil.
		var zero T
		for i := range data {
			data[i] = zero
		}
	}

	return res
}

func (a *SafeSlice[S, T]) Append(addData S) {
	if ok := a.fastAppend(addData); !ok {
		if a.endSize.Load() == -1 {
			a.Append(addData)
			return
		}
		if ok := a.slowAppend(addData); ok {
			return
		}
		a.Append(addData)
	}
}

func (a *SafeSlice[S, T]) slowAppend(addData S) bool {
	a.lock.Lock()
	defer a.lock.Unlock()
	if a.endSize.Load() != -1 {
		// endSize is the size of the first trySize overflow, i.e. any value less than endSize is valid.
		a.slice = a.slice[:a.endSize.Load()]

		l := int64(len(a.slice) + len(addData))
		trySize := a.trySize.Load()
		// All attempted sizes are stacked on top of trySize, so trySize is the new length we expect,
		// but reset may reset trySize, so take the maximum.
		if trySize < l {
			trySize = l
		}
		newSlice := make(S, nextSliceCap(int(trySize), cap(a.slice)))

		copy(newSlice, a.slice)
		copy(newSlice[len(a.slice):], addData)

		a.trySize.Store(l)
		a.slice = newSlice
		a.endSize.Store(-1)

		return true
	}
	return false
}

func (a *SafeSlice[S, T]) fastAppend(addData S) bool {
	a.lock.RLock()
	defer a.lock.RUnlock()
	l := int64(len(addData))
	if end, ok := a.mallocRLocked(l); ok {
		start := end - l
		copy(a.slice[start:end], addData)
		return true
	}

	return false
}

func (a *SafeSlice[S, T]) mallocRLocked(size int64) (end int64, ok bool) {
	end = a.trySize.Add(size)
	l := int64(cap(a.slice))
	if end > l {
		if end-size <= l {
			a.endSize.Store(end - size)
		}
		return end, false
	}
	return end, true
}

// nextSliceCap computes the next appropriate slice length.
func nextSliceCap(newLen, oldCap int) int {
	newCap := oldCap
	doubleCap := newCap + newCap
	if newLen > doubleCap {
		return newLen
	}

	const threshold = 256
	if oldCap < threshold {
		return doubleCap
	}
	for {
		// Transition from growing 2x for small slices
		// to growing 1.25x for large slices. This formula
		// gives a smooth-ish transition between the two.
		newCap += (newCap + 3*threshold) >> 2

		// We need to check `newCap >= newLen` and whether `newCap` overflowed.
		// newLen is guaranteed to be larger than zero, hence
		// when newCap overflows then `uint(newCap) > uint(newLen)`.
		// This allows to check for both with the same comparison.
		if uint(newCap) >= uint(newLen) {
			break
		}
	}

	// Set newCap to the requested cap when
	// the newCap calculation overflowed.
	if newCap <= 0 {
		return newLen
	}
	return newCap
}
