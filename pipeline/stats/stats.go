// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package stats used to record pl metrics
package stats

import (
	"sync"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
)

type EventOP string

const (
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
	// // debug only
	// go func() {
	// 	tk := time.NewTicker(time.Second * 10)
	// 	for {
	// 		<-tk.C
	// 		for _, v := range ReadStats() {
	// 			if v != nil {
	// 				l.Info(*v)
	// 			}
	// 		}
	// 	}
	// }()
}

type Stats struct {
	stats sync.Map
	event ScriptChangeEvent
}

func WriteEvent(event *ChangeEvent) {
	_plstats.event.Write(event)
}

func ReadEvent() []ChangeEvent {
	return _plstats.event.Read()
}

func UpdateScriptStatsMeta(category, ns, name string, scriptUpdate, enable bool, err ...error) {
	ts := time.Now().UnixNano()

	var compileErr error
	if len(err) > 0 {
		compileErr = err[0]
	}

	if stats, loaded := _plstats.stats.LoadOrStore(StatsKey(category, ns, name), &ScriptStats{
		meta: struct {
			startTS           int64
			scriptUpdateTS    int64
			scriptUpdateTimes uint64

			category, ns, name string
			enable             bool
			err                error

			metaUpdateTS int64

			sync.RWMutex
		}{
			startTS:           ts,
			scriptUpdateTS:    ts,
			scriptUpdateTimes: 1,
			category:          category,
			ns:                ns,
			name:              name,
			enable:            enable,
			metaUpdateTS:      ts,
			err:               compileErr,
		},
	}); loaded {
		if stats, ok := stats.(*ScriptStats); ok {
			if len(err) > 0 {
				stats.UpdateMeta(scriptUpdate, enable, err...)
			} else {
				stats.UpdateMeta(scriptUpdate, enable)
			}
		}
	}
}

func WriteScriptStats(category, ns, name string, pt, ptDrop, ptError uint64, err error) {
	v, ok := _plstats.stats.Load(StatsKey(category, ns, name))
	if !ok {
		return
	}
	if v, ok := v.(*ScriptStats); ok {
		v.WritePtCount(pt, ptDrop, ptDrop)
		if err != nil {
			v.WriteErr(err.Error())
		}
	}
}

func ReadStats() []*ScriptStatsROnly {
	ret := []*ScriptStatsROnly{}
	_plstats.stats.Range(func(key, value interface{}) bool {
		if value, ok := value.(*ScriptStats); ok && value != nil {
			ret = append(ret, value.Read())
		}
		return true
	})
	return ret
}

func StatsKey(category, ns, name string) string {
	return category + "::" + ns + "::" + name
}
