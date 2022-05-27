// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package bufpool wraps internal buffer pool functions
package bufpool

import (
	"bytes"
	"sync"
)

var pool = sync.Pool{
	New: func() interface{} {
		return bytes.NewBuffer(nil)
	},
}

func GetBuffer() *bytes.Buffer {
	buf, ok := pool.Get().(*bytes.Buffer)
	if !ok {
		return nil
	}
	buf.Reset()

	return buf
}

func PutBuffer(buf *bytes.Buffer) {
	pool.Put(buf)
}
