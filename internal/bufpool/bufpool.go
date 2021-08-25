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
	buf := pool.Get().(*bytes.Buffer)
	buf.Reset()

	return buf
}

func PutBuffer(buf *bytes.Buffer) {
	pool.Put(buf)
}
