// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package compact

import (
	sync "sync"

	"github.com/GuanceCloud/cliutils/logger"
)

var (
	newBufferBodyPool, reuseBufferBodyPool sync.Pool

	wpool sync.Pool

	defaultBatchSize = (1 << 20) // 1MB

	l = logger.DefaultSLogger("compact")
)

type BodyOpt func(*Body)

func WithCaller(c string) BodyOpt {
	return func(b *Body) {
		b.caller = c
	}
}

func WithNewBuffer(n int) BodyOpt {
	return func(b *Body) {
		if n > 0 && b.SendBuf == nil && b.MarshalBuf == nil {
			b.SendBuf = make([]byte, n)

			// +10% on marshal buffer: we need more bytes for meta-info about the body
			extra := int(float64(n) * .1)
			b.MarshalBuf = make([]byte, n+extra)
			b.selfBuffer = bufOnwerSelf
		}
	}
}

// withReusableBuffer assign outter buffer that not managed by body instance.
// if withNewBuffer() and withReusableBuffer() both passed, only 1 applied
// according to the order of bodyOpts.
func withReusableBuffer(send, marshal []byte) BodyOpt {
	return func(b *Body) {
		if len(send) > 0 && len(marshal) > 0 { // sendBuf and marshalBuf should not nil
			b.SendBuf = send
			b.MarshalBuf = marshal
			b.selfBuffer = bufOnwerOthers // buffer not comes from new buffer
		}
	}
}

func GetNewBufferBody(opts ...BodyOpt) *Body {
	var (
		b      *Body
		malloc bool
	)

	if x := newBufferBodyPool.Get(); x == nil {
		malloc = true
		b = &Body{selfBuffer: bufOnwerSelf}
	} else {
		b = x.(*Body)
	}

	for _, opt := range opts {
		opt(b)
	}

	if malloc {
		bodyCounterVec.WithLabelValues(b.caller, "malloc", "get", "new").Inc()
	} else {
		bodyCounterVec.WithLabelValues(b.caller, "pool", "get", "new").Inc()
	}

	if len(b.SendBuf) == 0 || len(b.MarshalBuf) == 0 {
		panic("no buffer set for new-buffer-body")
	}

	return b
}

func getReuseBufferBody(opts ...BodyOpt) *Body {
	var (
		b      *Body
		malloc bool
	)

	if x := reuseBufferBodyPool.Get(); x == nil {
		b = &Body{
			selfBuffer: bufOnwerOthers,
		}
		malloc = true
	} else {
		b = x.(*Body)
	}

	for _, opt := range opts {
		opt(b)
	}

	if malloc {
		bodyCounterVec.WithLabelValues(b.caller, "malloc", "get", "reuse").Inc()
	} else {
		bodyCounterVec.WithLabelValues(b.caller, "pool", "get", "reuse").Inc()
	}

	if len(b.SendBuf) == 0 || len(b.MarshalBuf) == 0 {
		panic("no buffer set for reuse-buffer-body")
	}

	return b
}

func PutBody(b *Body) {
	if b != nil {
		caller := b.caller
		b.reset()

		if b.selfBuffer == bufOnwerSelf {
			newBufferBodyPool.Put(b)
			bodyCounterVec.WithLabelValues(caller, "pool", "put", "new").Inc()
		} else {
			reuseBufferBodyPool.Put(b)
			bodyCounterVec.WithLabelValues(caller, "pool", "put", "reuse").Inc()
		}
	}
}

func GetWriter(opts ...WriteOption) *Writer {
	var w *Writer

	if x := wpool.Get(); x == nil {
		w = &Writer{
			HTTPHeaders: map[string]string{},
		}
		w.reset()
	} else {
		w = x.(*Writer)
	}

	for _, opt := range opts {
		if opt != nil {
			opt(w)
		}
	}

	return w
}

func PutWriter(w *Writer) {
	w.reset()
	wpool.Put(w)
}
