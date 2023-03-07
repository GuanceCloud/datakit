package timeout

import (
	"bytes"
	"sync"
)

// BufferPool is Pool of *bytes.Buffer
type BufferPool struct {
	pool sync.Pool
}

// Get a bytes.Buffer pointer
func (p *BufferPool) Get() *bytes.Buffer {
	buf := p.pool.Get()
	if buf == nil {
		return &bytes.Buffer{}
	}
	return buf.(*bytes.Buffer)
}

// Put a bytes.Buffer pointer to BufferPool
func (p *BufferPool) Put(buf *bytes.Buffer) {
	p.pool.Put(buf)
}
