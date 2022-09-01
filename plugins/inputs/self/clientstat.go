// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package self collect datakit self metrics.
package self

import (
	"fmt"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/shirou/gopsutil/host"
	pr "github.com/shirou/gopsutil/v3/process"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	dkhttp "gitlab.jiagouyun.com/cloudcare-tools/datakit/http"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/goroutine"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/election"
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

	// 选举
	Incumbency       int64
	ElectedNamespace string

	// HTTP
	HTTPStats map[string]*dkhttp.APIStat

	// goroutine
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
	s.HostName = config.Cfg.Hostname
	if config.Cfg.DataWayCfg != nil && config.Cfg.DataWayCfg.HTTPProxy != "" {
		s.Proxy = config.Cfg.DataWayCfg.HTTPProxy
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

	if cpuUsage, cpuUsageTop, err := getSelfCPUsage(); err != nil {
		l.Warnf("get CPU usage failed: %s, ignored", err.Error())
		s.CPUUsage = 0.0
		s.CPUUsageTop = 0.0
	} else {
		s.CPUUsage = cpuUsage
		s.CPUUsageTop = cpuUsageTop

		if runtime.GOOS == "windows" {
			s.CPUUsage /= float64(runtime.NumCPU())
			s.CPUUsageTop /= float64(runtime.NumCPU())
		}
	}

	du, ns := s.getElectedInfo()
	s.Incumbency = du
	s.ElectedNamespace = ns

	s.DroppedPoints = io.DroppedTotal() - s.DroppedPointsTotal
	s.DroppedPointsTotal = io.DroppedTotal()

	s.HTTPStats = dkhttp.GetMetrics()

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
		"namespace":         s.ElectedNamespace,
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

		"incumbency": s.Incumbency, // deprecated, used elected
		"elected":    s.Incumbency,
	}

	pt, err := point.NewPoint(measurementName, tags, fields, point.MOpt())
	if err != nil {
		l.Error(err)
	}

	return pt
}

func (s *ClientStat) ToHTTPMetric() []*point.Point {
	var rtPts []*point.Point
	for api, v := range s.HTTPStats {
		tags := map[string]string{
			"api": api,
		}

		fields := map[string]interface{}{
			"total_request_count": v.TotalCount,
			"max_latency":         int64(v.MaxLatency),
			"avg_latency":         int64(v.AvgLatency),
			"2XX":                 v.Status2xx,
			"3XX":                 v.Status3xx,
			"4XX":                 v.Status4xx,
			"5XX":                 v.Status5xx,
			"limited":             v.Limited,
		}

		pt, err := point.NewPoint(measurementHTTPName, tags, fields, point.MOpt())
		if err != nil {
			l.Error(err)
			continue
		}
		rtPts = append(rtPts, pt)
	}

	return rtPts
}

func (s *ClientStat) ToGoroutineMetric() []*point.Point {
	var pts []*point.Point

	if s.GoroutineStats != nil {
		// groutine that belongs to no group
		if s.NumGoroutines >= s.GoroutineStats.RunningTotal {
			tags := map[string]string{
				"host":      s.HostName,
				"namespace": s.ElectedNamespace,
				"group":     "unknown",
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
				"namespace": s.ElectedNamespace,
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

func (s *ClientStat) getElectedInfo() (int64, string) {
	if !config.Cfg.Election.Enable {
		return 0, ""
	}

	elected, ns, _ := election.Elected()
	if elected == "success" {
		return int64(time.Since(election.GetElectedTime()) / time.Second), ns
	} else {
		return 0, ""
	}
}

//------------------------------------------------------------------------------
// CPU Usage: copied and modified from plugins/inputs/process

var psRecorder = newProcRecorder()

// getSelfCPUsage returns cpu_usage, cpu_usage_top, error.
func getSelfCPUsage() (float64, float64, error) {
	tn := time.Now().UTC()
	selfProcess, err := getSelfProcess()
	if err != nil {
		return 0, 0, err
	}

	return parseSingleProcess(selfProcess, psRecorder, tn)
}

func parseSingleProcess(ps *pr.Process, procRec *procRecorder, tn time.Time) (float64, float64, error) {
	// you may get a null pointer here
	cputime, err := ps.Times()
	if err != nil {
		return 0, 0, err
	}

	defer func() {
		if cputime != nil {
			procRec.recorder[ps.Pid] = procRecStat{
				Pid:          ps.Pid,
				RecorderTime: tn,
				CPUUser:      cputime.User,
				CPUSystem:    cputime.System,
			}
		}
	}()

	return calculatePercent(ps, tn), procRec.calculatePercentTop(ps, tn), nil
}

func calculatePercent(ps *pr.Process, tn time.Time) float64 {
	if ps == nil {
		l.Error("nil point: param pr *Process")
		return 0.
	}
	cpuTime, err := ps.Times()
	if err != nil {
		l.Error("pid: %d, err: %w", ps.Pid, err)
		return 0.
	}
	created := time.Unix(0, getCreateTime(ps)*int64(time.Millisecond))
	totalTime := tn.Sub(created).Seconds()

	return 100 * ((cpuTime.User + cpuTime.System) / totalTime)
}

func getCreateTime(ps *pr.Process) int64 {
	start, err := ps.CreateTime()
	if err != nil {
		l.Warnf("get start time err:%s", err.Error())
		if bootTime, err := host.BootTime(); err != nil {
			return int64(bootTime) * 1000
		}
	}
	return start
}

type procRecStat struct {
	Pid int32

	RecorderTime time.Time

	CPUUser   float64
	CPUSystem float64
}

func newProcRecorder() *procRecorder {
	return &procRecorder{
		recorder: map[int32]procRecStat{},
	}
}

type procRecorder struct {
	recorder map[int32]procRecStat
	sync.RWMutex
}

// calculatePercentTop 计算一个周期内的 cpu 使用率，
// 若无上一次的记录或启动时间不一致的则返回自启动来的使用率.
func (p *procRecorder) calculatePercentTop(ps *pr.Process, pTime time.Time) float64 {
	p.RLock()
	defer p.RUnlock()

	if ps == nil {
		l.Error("nil point: param pr *Process")
		return 0.
	}

	if rec, ok := p.recorder[ps.Pid]; ok {
		return calculatePercentTop(ps, rec, pTime)
	} else {
		return calculatePercent(ps, pTime)
	}
}

func calculatePercentTop(ps *pr.Process, rec procRecStat, tn time.Time) float64 {
	if ps == nil {
		l.Error("nil point: param pr *Process")
		return 0.
	}
	cpuTime, err := ps.Times()
	if err != nil {
		l.Error("pid: %d, err: %w", ps.Pid, err)
		return 0.
	}
	timeDiff := tn.Sub(rec.RecorderTime).Seconds()
	if timeDiff == 0. || timeDiff == -0. ||
		timeDiff == 0 || timeDiff == -0 {
		l.Error("timeDiff == 0s")
		return 0
	}

	// TODO
	cpuDiff := cpuTime.User + cpuTime.System - rec.CPUUser - rec.CPUSystem
	switch {
	case cpuDiff <= 0.0 && cpuDiff >= -0.0000000001:
		return 0.0
	case cpuDiff < -0.0000000001:
	default:
		return 100 * (cpuDiff / timeDiff)
	}
	return calculatePercent(ps, tn)
}

func getSelfProcess() (*pr.Process, error) {
	pses, err := pr.Processes()
	if err != nil {
		return nil, err
	}

	selfPID := os.Getpid()
	for _, ps := range pses {
		if int(ps.Pid) == selfPID {
			return ps, nil
		}
	}

	return nil, fmt.Errorf("should not be here: not found process of the caller self")
}

//------------------------------------------------------------------------------
