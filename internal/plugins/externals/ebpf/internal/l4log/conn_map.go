//go:build linux
// +build linux

package l4log

import (
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/ntp"
)

const (

	// map create duration is 20s.
	mapCreateDuarion = time.Second * 20

	connMapShrinkPercentThreshold = 60
	connMapShrinkMinEntries       = 256
	connMapShrinkMinDeletes       = 64
)

type connMap struct {
	m           map[PMeta]*PValue
	insertCount int
	deleteCount int

	dirty    []PMeta
	dirtySet map[PMeta]struct{}

	tn           time.Time
	lastFullScan time.Time
}

func (cm *connMap) get(key PMeta) (*PValue, bool) {
	v, ok := cm.m[key]
	return v, ok
}

func (cm *connMap) delete(key PMeta) {
	delete(cm.m, key)
	if cm.dirtySet != nil {
		delete(cm.dirtySet, key)
	}
	cm.deleteCount++
	cm.maybeShrink()
}

func (cm *connMap) insert(k PMeta, v *PValue) {
	if cm.m == nil {
		cm.m = make(map[PMeta]*PValue)
	}

	if _, ok := cm.m[k]; ok {
		// just update
		cm.m[k] = v
	} else {
		cm.m[k] = v
		cm.insertCount++
	}

	if cm.deleteCount >= connMapShrinkMinDeletes {
		cm.maybeShrink()
	}
}

func (cm *connMap) markDirty(key PMeta) {
	if cm.dirtySet == nil {
		cm.dirtySet = make(map[PMeta]struct{})
	}
	if _, ok := cm.dirtySet[key]; ok {
		return
	}
	cm.dirtySet[key] = struct{}{}
	cm.dirty = append(cm.dirty, key)
}

func (cm *connMap) drainDirty() []PMeta {
	if len(cm.dirty) == 0 {
		return nil
	}

	keys := cm.dirty
	cm.dirty = nil
	cm.dirtySet = nil
	return keys
}

func (cm *connMap) finishFullScan(ts time.Time) {
	cm.lastFullScan = ts
	cm.dirty = nil
	cm.dirtySet = nil
}

func (cm *connMap) maybeShrink() {
	if cm.insertCount < connMapShrinkMinEntries || cm.deleteCount < connMapShrinkMinDeletes {
		return
	}
	if len(cm.m)*100 > cm.insertCount*connMapShrinkPercentThreshold {
		return
	}

	tmp := make(map[PMeta]*PValue, len(cm.m))
	for k, v := range cm.m {
		tmp[k] = v
	}

	cm.m = tmp
	cm.insertCount = len(tmp)
	cm.deleteCount = 0
}

func newConnMap() *connMap {
	now := ntp.Now()
	return &connMap{
		m:            make(map[PMeta]*PValue),
		tn:           now,
		lastFullScan: now,
	}
}

type connsMaps struct {
	// default map create interval is 20s
	createInterval time.Duration

	maps []*connMap
}

func newConnsMaps(dur time.Duration) *connsMaps {
	if dur <= 0 {
		dur = mapCreateDuarion
	}
	return &connsMaps{
		maps: []*connMap{
			newConnMap(),
		},
		createInterval: dur,
	}
}

func (p *connsMaps) getMapAndV(k PMeta) (*connMap, *PValue, bool) {
	for _, mps := range p.maps {
		if mps == nil {
			continue
		}
		v, _ := mps.get(k)
		if v != nil {
			// 如果存在则返回
			return mps, v, true
		}
	}

	return nil, nil, false
}

func (p *connsMaps) insert2LastMap(k PMeta, v *PValue) *connMap {
	lenMaps := len(p.maps)

	lastMaps := p.maps[lenMaps-1]
	if lastMaps == nil || time.Since(lastMaps.tn) >= p.createInterval {
		lastMaps = newConnMap()
		p.maps = append(p.maps, lastMaps)
	}

	lastMaps.insert(k, v)
	return lastMaps
}

func (p *connsMaps) entries() int {
	if p == nil {
		return 0
	}

	total := 0
	for _, mps := range p.maps {
		if mps == nil {
			continue
		}
		total += len(mps.m)
	}
	return total
}
