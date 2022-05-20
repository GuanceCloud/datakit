// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package stats

import (
	"sync"
	"sync/atomic"
	"time"
)

type ScriptStats struct {
	pt, ptDrop, ptError uint64

	lastRunErr struct {
		last100err [MaxErrorCount]string
		pos        int
		sync.RWMutex
	}

	meta struct {
		startTS           int64
		scriptUpdateTS    int64
		scriptUpdateTimes uint64

		category, ns, name string
		enable             bool
		err                error

		metaUpdateTS int64

		sync.RWMutex
	}
}

type ScriptStatsROnly struct {
	Pt, PtDrop, PtError uint64

	LastRunErrs []string

	StartTS     int64
	UpdateTS    int64
	UpdateTimes uint64

	Category, NS, Name string

	Enable bool
	Error  error

	MetaTS int64
}

func (stats *ScriptStats) WritePtCount(pt, ptDrop, ptError uint64) {
	if pt > 0 {
		atomic.AddUint64(&stats.pt, pt)
	}
	if ptDrop > 0 {
		atomic.AddUint64(&stats.ptDrop, ptDrop)
	}
	if ptError > 0 {
		atomic.AddUint64(&stats.ptError, ptError)
	}
}

func (stats *ScriptStats) WriteErr(err string) {
	if err == "" {
		return
	}
	stats.lastRunErr.Lock()
	defer stats.lastRunErr.Unlock()
	if stats.lastRunErr.pos >= MaxErrorCount {
		stats.lastRunErr.pos %= MaxErrorCount
	}
	stats.lastRunErr.last100err[stats.lastRunErr.pos] = err
	stats.lastRunErr.pos += 1
}

func (stats *ScriptStats) Read() *ScriptStatsROnly {
	ret := &ScriptStatsROnly{
		Pt:      atomic.LoadUint64(&stats.pt),
		PtDrop:  atomic.LoadUint64(&stats.ptDrop),
		PtError: atomic.LoadUint64(&stats.ptError),
	}

	stats.meta.RLock()
	defer stats.meta.RUnlock()
	ret.Category = stats.meta.category
	ret.NS = stats.meta.ns
	ret.Name = stats.meta.name

	ret.StartTS = stats.meta.startTS
	ret.UpdateTS = stats.meta.scriptUpdateTS
	ret.UpdateTimes = stats.meta.scriptUpdateTimes
	ret.Enable = stats.meta.enable
	ret.Error = stats.meta.err

	ret.MetaTS = stats.meta.metaUpdateTS

	stats.lastRunErr.RLock()
	defer stats.lastRunErr.RUnlock()
	last100 := []string{}
	curPos := stats.lastRunErr.pos
	if curPos >= MaxErrorCount {
		curPos %= MaxErrorCount
	}

	if stats.lastRunErr.last100err[curPos] == "" {
		for i := 0; i < curPos; i++ {
			last100 = append(last100, stats.lastRunErr.last100err[i])
		}
	} else {
		for i := curPos; i < MaxErrorCount; i++ {
			last100 = append(last100, stats.lastRunErr.last100err[i])
		}

		for i := 0; i < curPos; i++ {
			last100 = append(last100, stats.lastRunErr.last100err[i])
		}
	}

	ret.LastRunErrs = last100

	return ret
}

func (stats *ScriptStats) UpdateMeta(scriptUpdate bool, enable bool, err ...error) {
	stats.meta.Lock()
	defer stats.meta.Unlock()

	ts := time.Now().UnixNano()
	if scriptUpdate {
		stats.meta.scriptUpdateTS = ts
		stats.meta.scriptUpdateTimes += 1
	}

	stats.meta.enable = enable
	if len(err) > 0 {
		stats.meta.err = err[0]
	}

	stats.meta.metaUpdateTS = ts
}
