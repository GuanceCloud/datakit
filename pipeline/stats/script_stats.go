// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package stats

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

type ScriptStats struct {
	pt, ptDrop, ptError uint64

	totalCost int64

	lastRunErr struct {
		last100err [MaxErrorCount]string
		pos        int
		sync.RWMutex
	}

	meta struct {
		startTS           time.Time
		scriptUpdateTS    time.Time
		scriptUpdateTimes uint64

		category, ns, name string
		enable             bool
		err                string

		metaUpdateTS time.Time

		sync.RWMutex
	}
}

type ScriptStatsROnly struct {
	Pt, PtDrop, PtError uint64

	RunLastErrs []string

	TotalCost int64 // ns
	MetaTS    time.Time

	FirstTS           time.Time
	ScriptTS          time.Time
	ScriptUpdateTimes uint64

	Category, NS, Name string

	Enable       bool
	CompileError string
}

func (statsR ScriptStatsROnly) String() string {
	return fmt.Sprintf("%s enable: %v total_pt: %d dropped_pt: %d error_pt: %d "+
		"start: %s script: %s meta: %s compile_error: %v", StatsKey(statsR.Category, statsR.NS, statsR.Name),
		statsR.Enable, statsR.Pt, statsR.PtDrop, statsR.PtError, statsR.FirstTS.Format(StatsTimeFormat),
		statsR.ScriptTS.Format(StatsTimeFormat), statsR.MetaTS.Format(StatsTimeFormat), statsR.CompileError)
}

func (stats *ScriptStats) WritePtCount(pt, ptDrop, ptError uint64, cost int64) {
	if pt > 0 {
		atomic.AddUint64(&stats.pt, pt)
	}
	if ptDrop > 0 {
		atomic.AddUint64(&stats.ptDrop, ptDrop)
	}
	if ptError > 0 {
		atomic.AddUint64(&stats.ptError, ptError)
	}
	if cost > 0 {
		atomic.AddInt64(&stats.totalCost, cost)
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
		Pt:        atomic.LoadUint64(&stats.pt),
		PtDrop:    atomic.LoadUint64(&stats.ptDrop),
		PtError:   atomic.LoadUint64(&stats.ptError),
		TotalCost: atomic.LoadInt64(&stats.totalCost),
	}

	stats.meta.RLock()
	defer stats.meta.RUnlock()
	ret.Category = stats.meta.category
	ret.NS = stats.meta.ns
	ret.Name = stats.meta.name

	ret.FirstTS = stats.meta.startTS
	ret.ScriptTS = stats.meta.scriptUpdateTS
	ret.ScriptUpdateTimes = stats.meta.scriptUpdateTimes
	ret.Enable = stats.meta.enable
	ret.CompileError = stats.meta.err

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

	ret.RunLastErrs = last100

	return ret
}

func (stats *ScriptStats) UpdateMeta(scriptUpdate bool, enable bool, err ...string) {
	stats.meta.Lock()
	defer stats.meta.Unlock()

	t := time.Now()
	if scriptUpdate {
		stats.meta.scriptUpdateTS = t
		stats.meta.scriptUpdateTimes += 1
	}

	stats.meta.enable = enable
	if len(err) > 0 {
		stats.meta.err = err[0]
	}

	stats.meta.metaUpdateTS = t
}
