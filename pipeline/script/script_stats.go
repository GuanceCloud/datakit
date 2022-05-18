// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package script

import (
	"sync"
	"sync/atomic"
)

const MaxErrorCount int = 100

type ScriptStats struct {
	PtCount     uint64
	PtDropCount uint64
	PtError     uint64
	Last100err  []string
}

type scriptStats struct {
	ptCount     uint64
	ptDropCount uint64
	ptError     uint64
	last100err  [MaxErrorCount]string
	pos         int
	sync.RWMutex
}

func (stats *scriptStats) WritePtCount(pt, ptDrop, ptError uint64) {
	if pt > 0 {
		atomic.AddUint64(&stats.ptCount, pt)
	}
	if ptDrop > 0 {
		atomic.AddUint64(&stats.ptDropCount, ptDrop)
	}
	if ptError > 0 {
		atomic.AddUint64(&stats.ptError, ptError)
	}
}

func (stats *scriptStats) WriteErr(err string) {
	if err == "" {
		return
	}
	stats.Lock()
	defer stats.Unlock()
	if stats.pos >= MaxErrorCount {
		stats.pos %= MaxErrorCount
	}
	stats.last100err[stats.pos] = err
	stats.pos += 1
}

func (stats *scriptStats) Read() *ScriptStats {
	last100 := []string{}
	stats.RLock()
	defer stats.RUnlock()
	curPos := stats.pos
	if curPos >= MaxErrorCount {
		curPos %= MaxErrorCount
	}

	if stats.last100err[curPos] == "" {
		for i := 0; i < curPos; i++ {
			last100 = append(last100, stats.last100err[i])
		}
	} else {
		for i := curPos; i < MaxErrorCount; i++ {
			last100 = append(last100, stats.last100err[i])
		}

		for i := 0; i < curPos; i++ {
			last100 = append(last100, stats.last100err[i])
		}
	}

	ret := &ScriptStats{
		PtCount:     atomic.LoadUint64(&stats.ptCount),
		PtDropCount: atomic.LoadUint64(&stats.ptDropCount),
		PtError:     atomic.LoadUint64(&stats.ptError),
		Last100err:  last100,
	}
	return ret
}
