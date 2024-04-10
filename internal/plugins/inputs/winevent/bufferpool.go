// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.
// Some code modified from project Beats (https://github.com/elastic/beats).

//go:build windows
// +build windows

package winevent

import (
	"sync"
)

// bufferPool contains a pool of PooledByteBuffer objects.
var bufferPool = sync.Pool{
	New: func() interface{} { return &PooledByteBuffer{ByteBuffer: NewByteBuffer(1024)} },
}

// PooledByteBuffer is an expandable buffer backed by a byte slice.
type PooledByteBuffer struct {
	*ByteBuffer
}

// NewPooledByteBuffer return a PooledByteBuffer from the pool. The returned value must
// be released with Free().
func NewPooledByteBuffer() *PooledByteBuffer {
	b := bufferPool.Get().(*PooledByteBuffer)
	b.Reset()
	return b
}

// Free returns the PooledByteBuffer to the pool.
func (b *PooledByteBuffer) Free() {
	if b == nil {
		return
	}
	bufferPool.Put(b)
}
