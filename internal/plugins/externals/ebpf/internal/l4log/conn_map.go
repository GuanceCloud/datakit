//go:build linux
// +build linux

package l4log

import (
	"time"
)

const (

	// map create duration is 20s.
	mapCreateDuarion = time.Second * 20
)

type connMap struct {
	m           map[PMeta]*PValue
	insertCount int

	tn time.Time
}

func (cm *connMap) get(key PMeta) (*PValue, bool) {
	v, ok := cm.m[key]
	return v, ok
}

func (cm *connMap) delete(key PMeta) {
	delete(cm.m, key)

	// map 的估计容量低于 60% 时，将元素转移至新的 map，
	if float64(len(cm.m))/float64(cm.insertCount) <= 0.6 {
		tmp := make(map[PMeta]*PValue)
		for k, v := range cm.m {
			tmp[k] = v
		}
		// copy
		cm.m = tmp
		cm.insertCount = len(tmp)
	}
}

func (cm *connMap) insert(k PMeta, v *PValue) {
	if cm.m == nil {
		cm.m = make(map[PMeta]*PValue)
	}

	if float64(len(cm.m))/float64(cm.insertCount) <= 0.6 {
		tmp := make(map[PMeta]*PValue)
		for k, v := range cm.m {
			tmp[k] = v
		}
		// copy
		cm.m = tmp
		cm.insertCount = len(tmp)
	}

	if _, ok := cm.m[k]; ok {
		// just update
		cm.m[k] = v
	} else {
		cm.m[k] = v
		cm.insertCount++
	}
}

func newConnMap() *connMap {
	return &connMap{
		m:  make(map[PMeta]*PValue),
		tn: time.Now(),
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
