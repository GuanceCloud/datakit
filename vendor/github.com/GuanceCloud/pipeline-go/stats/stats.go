// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package stats used to record pl metrics
package stats

import (
	"sync/atomic"
	"time"

	"github.com/GuanceCloud/cliutils/logger"
	"github.com/prometheus/client_golang/prometheus"
)

type EventOP string

const (
	StatsTimeFormat = "2006-01-02T15:04:05.999Z07:00"

	EventOpAdd                EventOP = "ADD"
	EventOpUpdate             EventOP = "UPDATE"
	EventOpDelete             EventOP = "DELETE"
	EventOpIndex              EventOP = "INDEX"
	EventOpIndexUpdate        EventOP = "INDEX_UPDATE"
	EventOpIndexDelete        EventOP = "INDEX_DELETE"
	EventOpIndexDeleteAndBack EventOP = "INDEX_DELETE_AND_BACK"
	EventOpCompileError       EventOP = "COMPILE_ERROR"

	DefaultSubSystem = "pipeline"
)

var (
	l = logger.DefaultSLogger("pl-stats")

	statsState atomic.Pointer[statsHolder]
	_          Stats = (*RecStats)(nil)
)

type statsHolder struct{ value Stats }

func SetStats(st Stats) {
	if st == nil {
		statsState.Store(nil)
		return
	}
	statsState.Store(&statsHolder{value: st})
}

func currentStats() Stats {
	if holder := statsState.Load(); holder != nil {
		return holder.value
	}
	return nil
}

func InitLog() {
	l = logger.SLogger("pl-stats")
}

type Stats interface {
	Metrics() []prometheus.Collector
	WriteMetric(tags map[string]string, pt, ptDrop, ptError float64, cost time.Duration)

	WriteEvent(event *ChangeEvent, tags map[string]string)
	ReadEvents(events []*ChangeEvent) []*ChangeEvent

	WriteUpdateTime(tags map[string]string)
}

func NewRecStats(ns, subsystem string, labelNames []string, eventSize int) *RecStats {
	var lbs []string
	mp := map[string]struct{}{}

	for _, name := range labelNames {
		if _, ok := mp[name]; !ok {
			mp[name] = struct{}{}
			lbs = append(lbs, name)
		}
	}
	for _, name := range defaultLabelNames {
		if _, ok := mp[name]; !ok {
			mp[name] = struct{}{}
			lbs = append(lbs, name)
		}
	}

	return &RecStats{
		metric: newRecMetric(ns, subsystem, lbs),
		event:  newRecEvent(eventSize, lbs),
	}
}

type RecStats struct {
	metric *RecMetric
	event  *RecEvent
}

func (stats *RecStats) Metrics() []prometheus.Collector {
	return stats.metric.Metrics()
}

func (stats *RecStats) WriteEvent(event *ChangeEvent, tags map[string]string) {
	stats.event.Write(event, tags)
}

func (stats *RecStats) WriteUpdateTime(tags map[string]string) {
	stats.metric.WriteUpdateTime(tags)
}

func (stats *RecStats) WriteMetric(tags map[string]string, pt, ptDrop, ptError float64, cost time.Duration) {
	stats.metric.WriteMetric(tags, pt, ptDrop, ptError, cost)
}

func (stats *RecStats) ReadEvents(events []*ChangeEvent) []*ChangeEvent {
	return stats.event.Read(events)
}

func WriteEvent(event *ChangeEvent, tags map[string]string) {
	stats := currentStats()
	if stats == nil {
		return
	}
	stats.WriteEvent(event, tags)
}

func WriteUpdateTime(tags map[string]string) {
	stats := currentStats()
	if stats == nil {
		return
	}
	stats.WriteUpdateTime(tags)
}

func WriteMetric(tags map[string]string, pt, ptDrop, ptError float64, cost time.Duration) {
	stats := currentStats()
	if stats == nil {
		return
	}
	stats.WriteMetric(tags, pt, ptDrop, ptError, cost)
}
