package module

import (
	"crypto/md5"
	"fmt"
	"io"
	"sync"
)

// lscript global connection pool to databases
var connPool = pool{
	m:   make(map[string]connect),
	mux: sync.RWMutex{},
}

type connect interface {
	close()
}

type pool struct {
	m   map[string]connect
	mux sync.RWMutex
}

func (p *pool) load(key string) (connect, bool) {
	c, exist := p.m[key]
	return c, exist
}

func (p *pool) store(key string, conn connect) {
	p.mux.RLock()
	defer p.mux.RUnlock()
	p.m[key] = conn
}

func (p *pool) close() {
	p.mux.RLock()
	defer p.mux.RUnlock()

	for key, conn := range p.m {
		conn.close()
		p.m[key] = nil
	}
}

func joinKey(s string, arr ...string) string {
	h := md5.New()
	io.WriteString(h, s)
	for _, a := range arr {
		io.WriteString(h, a)
	}
	return fmt.Sprintf("%x", h.Sum(nil))
}
