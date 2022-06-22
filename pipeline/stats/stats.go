// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package stats used to record pl metrics
package stats

import (
	"sort"
	"sync"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
)

type EventOP string

const (
	StatsTimeFormat = "2006-01-02T15:04:05.999Z07:00"

	MaxEventLen   int = 100
	MaxErrorCount int = 100

	EventOpAdd                EventOP = "ADD"
	EventOpUpdate             EventOP = "UPDATE"
	EventOpDelete             EventOP = "DELETE"
	EventOpIndex              EventOP = "INDEX"
	EventOpIndexUpdate        EventOP = "INDEX_UPDATE"
	EventOpIndexDelete        EventOP = "INDEX_DELETE"
	EventOpIndexDeleteAndBack EventOP = "INDEX_DELETE_AND_BACK"
	EventOpCompileError       EventOP = "COMPILE_ERROR"
)

var (
	l = logger.DefaultSLogger("pl-stats")

	_plstats = Stats{}
)

func InitStats() {
	l = logger.SLogger("pl-stats")
}

type Stats struct {
	stats sync.Map
	event ScriptChangeEvent
}

func (stats *Stats) WriteEvent(event *ChangeEvent) {
	stats.event.Write(event)
}

func (stats *Stats) ReadEvent() []ChangeEvent {
	return stats.event.Read()
}

func (stats *Stats) UpdateScriptStatsMeta(category, ns, name, script string, enable, deleted bool, err string) {
	ts := time.Now()

	if stats, loaded := stats.stats.LoadOrStore(StatsKey(category, ns, name), &ScriptStats{
		meta: ScriptMeta{
			script:            script,
			deleted:           deleted,
			enable:            enable,
			startTS:           ts,
			scriptUpdateTS:    ts,
			scriptUpdateTimes: 1,
			category:          category,
			ns:                ns,
			name:              name,
			metaUpdateTS:      ts,
			err:               err,
		},
	}); loaded {
		if stats, ok := stats.(*ScriptStats); ok {
			stats.UpdateMeta(script, enable, deleted, err)
		}
	}
}

func (stats *Stats) WriteScriptStats(category, ns, name string, pt, ptDrop, ptError uint64, cost int64, err error) {
	v, ok := stats.stats.Load(StatsKey(category, ns, name))
	if !ok {
		return
	}
	if v, ok := v.(*ScriptStats); ok {
		v.WritePtCount(pt, ptDrop, ptError, cost)
		if err != nil {
			v.WriteErr("time: " + time.Now().Format(StatsTimeFormat) + "error: " + err.Error())
		}
	}
}

func (stats *Stats) ReadStats() []ScriptStatsROnly {
	ret := []ScriptStatsROnly{}
	stats.stats.Range(func(key, value interface{}) bool {
		if value, ok := value.(*ScriptStats); ok && value != nil {
			if v := value.Read(); v != nil {
				ret = append(ret, *v)
			}
		}
		return true
	})
	return ret
}

func WriteEvent(event *ChangeEvent) {
	_plstats.WriteEvent(event)
}

func ReadEvent() []ChangeEvent {
	return _plstats.ReadEvent()
}

func UpdateScriptStatsMeta(category, ns, name, script string, enable, deleted bool, err string) {
	_plstats.UpdateScriptStatsMeta(category, ns, name, script, enable, deleted, err)
}

func WriteScriptStats(category, ns, name string, pt, ptDrop, ptError uint64, cost int64, err error) {
	_plstats.WriteScriptStats(category, ns, name, pt, ptDrop, ptError, cost, err)
}

func ReadStats() []ScriptStatsROnly {
	ret := SortStatsROnly(_plstats.ReadStats())
	sort.Sort(ret)
	return ret
}

func StatsKey(category, ns, name string) string {
	return category + "::" + ns + "::" + name
}

type SortStatsROnly []ScriptStatsROnly

func (s SortStatsROnly) Less(i, j int) bool {
	si := s[i]
	sj := s[j]
	if si.Deleted != sj.Deleted {
		if !si.Deleted {
			return true
		} else {
			return false
		}
	}

	if si.Enable != sj.Enable {
		if si.Enable {
			return true
		} else {
			return false
		}
	}

	if si.Pt != sj.Pt {
		if si.Pt > sj.Pt {
			return true
		} else {
			return false
		}
	}

	if si.Name != sj.Name {
		return sort.StringsAreSorted([]string{si.Name, sj.Name})
	}

	if si.Category != sj.Category {
		return sort.StringsAreSorted([]string{si.Category, sj.Category})
	}

	if si.NS != sj.NS {
		return sort.StringsAreSorted([]string{si.NS, sj.NS})
	}
	return false
}

func (s SortStatsROnly) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s SortStatsROnly) Len() int {
	return len(s)
}
