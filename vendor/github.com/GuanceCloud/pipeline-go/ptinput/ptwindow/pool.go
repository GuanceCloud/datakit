package ptwindow

import (
	"sync"

	"github.com/GuanceCloud/cliutils/pkg/hash"
)

type WindowPool struct {
	sync.RWMutex
	pool map[[3]uint64]*PtWindow
}

func (m *WindowPool) Register(before, after int, k, v []string) {
	m.Lock()
	defer m.Unlock()

	key := [3]uint64{
		hash.Fnv1aHash(k),
		hash.Fnv1aHash(v),
		uint64(len(k)),
	}

	if m.pool == nil {
		m.pool = make(map[[3]uint64]*PtWindow)
	}

	if _, ok := m.pool[key]; !ok {
		m.pool[key] = NewWindow(before, after)
	}
}

func (m *WindowPool) Get(k, v []string) (*PtWindow, bool) {
	m.RLock()
	defer m.RUnlock()

	key := [3]uint64{
		hash.Fnv1aHash(k),
		hash.Fnv1aHash(v),
		uint64(len(k)),
	}
	w, ok := m.pool[key]
	return w, ok
}

func (m *WindowPool) Deprecated() {
	m.Lock()
	defer m.Unlock()
	for _, v := range m.pool {
		v.deprecated()
	}
}

func NewManager() *WindowPool {
	return &WindowPool{
		pool: make(map[[3]uint64]*PtWindow),
	}
}
