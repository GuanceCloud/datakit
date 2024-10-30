// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package dataway

import (
	sync "sync"

	"github.com/GuanceCloud/cliutils/point"
)

var (
	newBufferBodyPool, reuseBufferBodyPool sync.Pool

	wpool sync.Pool

	defaultBatchSize = (1 << 20) // 1MB
)

type bodyOpt func(*body)

func withNewBuffer(n int) bodyOpt {
	return func(b *body) {
		if n > 0 && b.sendBuf == nil && b.marshalBuf == nil {
			b.sendBuf = make([]byte, n)

			// +10% on marshal buffer: we need more bytes for meta-info about the body
			extra := int(float64(n) * .1)
			b.marshalBuf = make([]byte, n+extra)
			b.selfBuffer = 1
		}
	}
}

// withReusableBuffer assign outter buffer that not managed by body instance.
// if withNewBuffer() and withReusableBuffer() both passed, only 1 applied
// according to the order of bodyOpts.
func withReusableBuffer(send, marshal []byte) bodyOpt {
	return func(b *body) {
		if len(send) > 0 && len(marshal) > 0 { // sendBuf and marshalBuf should not nil
			b.sendBuf = send
			b.marshalBuf = marshal
			b.selfBuffer = 0 // buffer not comes from new buffer
		}
	}
}

func getNewBufferBody(opts ...bodyOpt) *body {
	var b *body
	if x := newBufferBodyPool.Get(); x == nil {
		b = &body{
			selfBuffer: 1,
		}
	} else {
		b = x.(*body)
	}

	for _, opt := range opts {
		opt(b)
	}

	if len(b.sendBuf) == 0 || len(b.marshalBuf) == 0 {
		panic("no buffer set for new-buffer-body")
	}

	return b
}

func getReuseBufferBody(opts ...bodyOpt) *body {
	var b *body
	if x := reuseBufferBodyPool.Get(); x == nil {
		b = &body{
			selfBuffer: 0,
		}
	} else {
		b = x.(*body)
	}

	for _, opt := range opts {
		opt(b)
	}

	if len(b.sendBuf) == 0 || len(b.marshalBuf) == 0 {
		panic("no buffer set for reuse-buffer-body")
	}

	return b
}

func putBody(b *body) {
	if b != nil {
		b.reset()

		if b.selfBuffer == 1 {
			newBufferBodyPool.Put(b)
		} else {
			reuseBufferBodyPool.Put(b)
		}
	}
}

func getWriter(opts ...WriteOption) *writer {
	var w *writer

	if x := wpool.Get(); x == nil {
		w = &writer{
			httpHeaders:    map[string]string{},
			batchBytesSize: defaultBatchSize,
		}
	} else {
		w = x.(*writer)
	}

	for _, opt := range opts {
		if opt != nil {
			opt(w)
		}
	}

	return w
}

func putWriter(w *writer) {
	w.category = point.UnknownCategory
	w.dynamicURL = ""
	w.points = w.points[:0]
	w.gzip = -1
	w.cacheClean = false
	w.cacheAll = false
	w.batchBytesSize = 1 << 20
	w.batchSize = 0
	w.bcb = nil

	for k := range w.httpHeaders {
		delete(w.httpHeaders, k)
	}
	w.httpEncoding = point.LineProtocol
	wpool.Put(w)
}
