// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package promremote

import (
	"io"
	"sync"
)

var _bufferPool = sync.Pool{
	New: func() any {
		return &bytesWrap{
			buf: make([]byte, 0, 64*4096),
		}
	},
}

type bytesWrap struct {
	buf []byte
}

func (buf *bytesWrap) ReadFrom(r io.Reader) (int64, error) {
	b := buf.buf // 64 * 4k

	for {
		// copy from io.ReadAll()
		if len(b) == cap(b) {
			// Add more capacity (let append pick how much).
			b = append(b, 0)[:len(b)]
		}
		n, err := r.Read(b[len(b):cap(b)])
		b = b[:len(b)+n]

		if err != nil {
			if err == io.EOF {
				err = nil
			}
			buf.buf = b
			return int64(len(b)), err
		}
	}
}

func (buf *bytesWrap) Cap() int {
	return cap(buf.buf)
}

func (buf *bytesWrap) Bytes() []byte {
	return buf.buf
}

func (buf *bytesWrap) SetBytes(b []byte) {
	buf.buf = b
}

func getBuffer() *bytesWrap {
	return _bufferPool.Get().(*bytesWrap)
}

func putBuffer(buf *bytesWrap) {
	if buf != nil {
		buf.buf = buf.buf[:0]
		_bufferPool.Put(buf)
	}
}
