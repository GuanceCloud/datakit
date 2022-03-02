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
