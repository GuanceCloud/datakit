// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package self collect datakit self metrics.
package self

import (
	"runtime"
	"time"

	"github.com/GuanceCloud/cliutils/metrics"
	dto "github.com/prometheus/client_model/go"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/cgroup"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/election"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/goroutine"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/point"
)

var StartTime time.Time

type ClientStat struct {
	HostName string

	PID      int
	Uptime   int64
	OS       string
	OSDetail string
	Arch     string
	Proxy    string

	NumGoroutines int64
	HeapAlloc     int64
	HeapSys       int64
	HeapObjects   int64

	MinNumGoroutines int64
	MinHeapAlloc     int64
	MinHeapSys       int64
	MinHeapObjects   int64

	MaxNumGoroutines int64
	MaxHeapAlloc     int64
	MaxHeapSys       int64
	MaxHeapObjects   int64

	CPUUsage    float64
	CPUUsageTop float64

	DroppedPointsTotal uint64
	DroppedPoints      uint64

	mfs []*dto.MetricFamily

	electionInfo *election.ElectionInfo

	GoroutineStats *goroutine.Summary
}

func setMax(prev, cur int64) int64 {
	if prev == 0 || prev < cur {
		return cur
	} else {
		return prev
	}
}

func setMin(prev, cur int64) int64 {
	if prev == 0 || prev > cur {
		return cur
	} else {
		return prev
	}
}

func (s *ClientStat) Update() {
	mfs, err := metrics.Gather()
	if err == nil {
		s.mfs = mfs
	}

	s.HostName = config.Cfg.Hostname
	if config.Cfg.Dataway != nil && config.Cfg.Dataway.HTTPProxy != "" {
		s.Proxy = config.Cfg.Dataway.HTTPProxy
	}

	var memStatus runtime.MemStats
	runtime.ReadMemStats(&memStatus)

	s.NumGoroutines = int64(runtime.NumGoroutine())
	s.HeapAlloc = int64(memStatus.HeapAlloc)
	s.HeapSys = int64(memStatus.HeapSys)
	s.HeapObjects = int64(memStatus.HeapObjects)

	s.MaxNumGoroutines = setMax(s.MaxNumGoroutines, s.NumGoroutines)
	s.MinNumGoroutines = setMin(s.MinNumGoroutines, s.NumGoroutines)

	s.MaxHeapAlloc = setMax(s.MaxHeapAlloc, s.HeapAlloc)
	s.MinHeapAlloc = setMin(s.MinHeapAlloc, s.HeapAlloc)

	s.MaxHeapSys = setMax(s.MaxHeapSys, s.HeapSys)
	s.MinHeapSys = setMin(s.MinHeapSys, s.HeapSys)

	s.MaxHeapObjects = setMax(s.MaxHeapObjects, s.HeapObjects)
	s.MinHeapObjects = setMin(s.MinHeapObjects, s.HeapObjects)

	if x, err := cgroup.MyCPUPercent(0); err != nil {
		l.Warnf("get CPU usage failed: %s, ignored", err.Error())
	} else {
		s.CPUUsage = x
		s.CPUUsageTop = x
	}

	s.getElectedInfo()

	s.GoroutineStats = goroutine.GetStat()
}

var measurementName = "datakit"

func (s *ClientStat) ToMetric() *point.Point {
	s.Uptime = int64(time.Since(StartTime) / time.Second)

	tags := map[string]string{
		"uuid":              config.Cfg.UUID,
		"version":           datakit.Version,
		"os":                s.OS,
		"os_version_detail": s.OSDetail,
		"arch":              s.Arch,
		"host":              s.HostName,
	}

	var elected int64
	if s.electionInfo != nil {
		tags["namespace"] = s.electionInfo.Namespace
		elected = int64(s.electionInfo.ElectedTime / time.Second)
	}

	if s.Proxy != "" {
		tags["proxy"] = s.Proxy
	}

	fields := map[string]interface{}{
		"pid":    s.PID,
		"uptime": s.Uptime,

		"num_goroutines": s.NumGoroutines,
		"heap_alloc":     s.HeapAlloc,
		"heap_sys":       s.HeapSys,
		"heap_objects":   s.HeapObjects,
		"cpu_usage":      s.CPUUsage,
		"cpu_usage_top":  s.CPUUsageTop,

		"min_num_goroutines": s.MinNumGoroutines,
		"min_heap_alloc":     s.MinHeapAlloc,
		"min_heap_sys":       s.MinHeapSys,
		"min_heap_objects":   s.MinHeapObjects,

		"max_num_goroutines": s.MaxNumGoroutines,
		"max_heap_alloc":     s.MaxHeapAlloc,
		"max_heap_sys":       s.MaxHeapSys,
		"max_heap_objects":   s.MaxHeapObjects,

		"dropped_points_total": s.DroppedPointsTotal,
		"dropped_points":       s.DroppedPoints,
		"open_files":           datakit.OpenFiles(),

		"elected": elected,
		// Deprecated: used elected
		"incumbency": elected,
	}

	pt, err := point.NewPoint(measurementName, tags, fields, point.MOpt())
	if err != nil {
		l.Error(err)
	}

	return pt
}

func (s *ClientStat) ToGoroutineMetric() []*point.Point {
	var pts []*point.Point

	electNS := ""
	if s.electionInfo != nil {
		electNS = s.electionInfo.Namespace
	}

	if s.GoroutineStats != nil {
		// groutine that belongs to no group
		if s.NumGoroutines >= s.GoroutineStats.RunningTotal {
			tags := map[string]string{
				"host":      s.HostName,
				"group":     "unknown",
				"namespace": electNS,
			}

			fields := map[string]interface{}{
				"running_goroutine_num": s.NumGoroutines - s.GoroutineStats.RunningTotal,
			}
			if pt, err := point.NewPoint(measurementGoroutineName, tags, fields, point.MOpt()); err != nil {
				l.Errorf("ToGoroutineMetric make point error: %s, ignore", err.Error())
			} else {
				pts = append(pts, pt)
			}
		} else {
			l.Warnf("The total goroutine number is less than the number of the group goroutine. There may be some unexpected problem.")
		}

		// goroutine group stat
		for groupName, info := range s.GoroutineStats.Items {
			tags := map[string]string{
				"host":      s.HostName,
				"namespace": electNS,
				"group":     groupName,
			}

			fields := map[string]interface{}{
				"running_goroutine_num":  info.RunningTotal,
				"finished_goroutine_num": info.Total,
				"failed_num":             info.ErrCount,
				"total_cost_time":        info.CostTimeDuration.Nanoseconds(),
				"min_cost_time":          info.MinCostTimeDuration.Nanoseconds(),
				"max_cost_time":          info.MaxCostTimeDuration.Nanoseconds(),
			}

			if pt, err := point.NewPoint(measurementGoroutineName, tags, fields, point.MOpt()); err != nil {
				l.Errorf("ToGoroutineMetric make point error: %s, ignore", err.Error())
				continue
			} else {
				pts = append(pts, pt)
			}
		}
	}

	return pts
}

func (s *ClientStat) getElectedInfo() {
	s.electionInfo = election.GetElectionInfo(s.mfs)
}
