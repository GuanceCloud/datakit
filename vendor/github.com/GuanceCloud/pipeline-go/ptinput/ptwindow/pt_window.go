// Package ptwindow impl pt context window
package ptwindow

import (
	"fmt"
	"sync"

	"github.com/GuanceCloud/cliutils/point"
)

func PutbackPoints(pts ...*point.Point) {
	if ptPool := point.GetPointPool(); ptPool != nil {
		for _, pt := range pts {
			if pt.HasFlag(point.Ppooled) {
				ptPool.Put(pt)
			}
		}
	}
}

type PtRing struct {
	ring []*point.Point
	pos  int

	elemLimit int
}

func (w *PtRing) put(pt *point.Point) {
	if w.pos >= len(w.ring) {
		w.pos = 0
	}

	if w.ring[w.pos] != nil {
		PutbackPoints(w.ring[w.pos])
	}
	w.ring[w.pos] = pt

	w.pos++
}

func (w *PtRing) clean() []*point.Point {
	var r []*point.Point

	for i := 0; i < len(w.ring); i++ {
		if w.ring[i] != nil {
			r = append(r, w.ring[i])
			w.ring[i] = nil
		}
	}
	return r
}

func NewRing(elem int) (*PtRing, error) {
	if elem <= 0 {
		return nil, fmt.Errorf("invalid ring size: %d", elem)
	}

	return &PtRing{
		ring:      make([]*point.Point, elem),
		elemLimit: elem,
	}, nil
}

type PtWindow struct {
	ringBefore *PtRing

	hit int // default 0

	before int
	after  int

	disableInsert bool
	sync.Mutex
}

func (w *PtWindow) deprecated() {
	w.Lock()
	defer w.Unlock()
	w.disableInsert = true

	pts := w.ringBefore.clean()
	PutbackPoints(pts...)
}

func (w *PtWindow) Move(pt *point.Point) []*point.Point {
	w.Lock()
	defer w.Unlock()

	if w.hit > 0 {
		w.hit--
		var rst []*point.Point
		if w.ringBefore != nil {
			if v := w.ringBefore.clean(); len(v) > 0 {
				rst = append(rst, v...)
			}
		}
		if pt != nil {
			rst = append(rst, pt)
		}
		return rst
	} else if w.ringBefore != nil && !w.disableInsert {
		w.ringBefore.put(pt)
	}

	return nil
}

func (w *PtWindow) Hit() {
	w.Lock()
	w.hit = w.after
	w.Unlock()
}

func NewWindow(before int, after int) *PtWindow {
	w := &PtWindow{
		ringBefore: nil,
		before:     before,
		after:      after,
	}

	if before > 0 {
		w.ringBefore, _ = NewRing(before)
	}

	return w
}
