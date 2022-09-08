// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package calcutil contains calculate util
package calcutil

import "sync/atomic"

// AtomicMinusUint64 returns the new changed value, the origin value will be changed.
func AtomicMinusUint64(addr *uint64, minus int64) uint64 {
	if minus < 0 {
		minus = -minus // Even though subtraction, the value should be positive.
	}

	// https://pkg.go.dev/sync/atomic#AddUint64
	// To subtract a signed positive constant value c from x, do AddUint64(&x, ^uint64(c-1)).
	// In particular, to decrement x, do AddUint64(&x, ^uint64(0)).
	return atomic.AddUint64(addr, ^uint64(minus-1))
}
