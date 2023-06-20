// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package stats used to record pl metrics
package stats

import (
	"strings"
	"sync"
	"time"

	"github.com/GuanceCloud/cliutils/logger"
	"github.com/GuanceCloud/cliutils/point"
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

func (stats *Stats) UpdateScriptStatsMeta(category point.Category, ns, name, script string, enable, deleted bool, err string) {
	ts := time.Now()

	defer func() {
		plUpdateVec.WithLabelValues(category.String(), name, ns).Set(float64(ts.Unix()))
	}()

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

func (stats *Stats) WriteScriptStats(category point.Category, ns, name string, pt, ptDrop, ptError uint64, cost int64, err error) {
	catStr := category.String()

	if pt > 0 {
		plPtsVec.WithLabelValues(catStr, name, ns).Add(float64(pt))
	}

	if ptDrop > 0 {
		plDropVec.WithLabelValues(catStr, name, ns).Add(float64(ptDrop))
	}

	if ptError > 0 {
		plErrPtsVec.WithLabelValues(catStr, name, ns).Add(float64(ptDrop))
	}

	if cost > 0 {
		plCostVec.WithLabelValues(catStr, name, ns).Observe(float64(cost) / float64(time.Second))
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

func UpdateScriptStatsMeta(category point.Category, ns, name, script string, enable, deleted bool, err string) {
	_plstats.UpdateScriptStatsMeta(category, ns, name, script, enable, deleted, err)
}

func WriteScriptStats(category point.Category, ns, name string, pt, ptDrop, ptError uint64, cost int64, err error) {
	_plstats.WriteScriptStats(category, ns, name, pt, ptDrop, ptError, cost, err)
}

func StatsKey(category point.Category, ns, name string) string {
	var b strings.Builder

	b.WriteString(ns)
	b.Write([]byte("::"))
	b.WriteString(category.String())
	b.Write([]byte("::"))
	b.WriteString(name)

	return b.String() // ns::cat::name
}
