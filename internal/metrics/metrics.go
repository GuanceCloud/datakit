// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package metrics stores datakit runtime metrics.
package metrics

import (
	"fmt"
	"runtime"
	"time"

	"github.com/GuanceCloud/cliutils/metrics"
	p8s "github.com/prometheus/client_golang/prometheus"
	"github.com/shirou/gopsutil/v3/process"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/git"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/resourcelimit"
)

const (
	// DatakitLastError is the metric name that accept any error info
	// during datakit running.
	//
	// NOTE: we should not upload the metric to datakit self metric uploading.
	DatakitLastError = "datakit_last_err"
)

var (
	Uptime      = time.Now()
	BeyondUsage int64

	collector runtimeInfoCollector
)

//nolint:gochecknoinits
func init() {
	metrics.MustRegister(collector)
}

type runtimeInfo struct {
	goroutines int
	heapAlloc  uint64
	sys        uint64
	cpuUsage   float64

	gcPauseTotal uint64
	gcNum        uint32

	ioCountersstats *process.IOCountersStat
	numCtxSwitch    *process.NumCtxSwitchesStat
}

func getRuntimeInfo() *runtimeInfo {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	var usage float64
	if u, err := resourcelimit.MyCPUPercent(time.Second); err == nil {
		usage = u
	}

	return &runtimeInfo{
		goroutines: runtime.NumGoroutine(),
		heapAlloc:  m.HeapAlloc,
		sys:        m.Sys,
		cpuUsage:   usage,

		gcPauseTotal:    m.PauseTotalNs,
		gcNum:           m.NumGC,
		ioCountersstats: resourcelimit.MyIOCountersStat(),
		numCtxSwitch:    resourcelimit.MyCtxSwitch(),
	}
}

// collector for basic runtime info.
// These metrics are self generated, not triggered by outter
// operation(such as HTTP request, io Feed point).
type runtimeInfoCollector struct{}

var (
	riGoroutineDesc = p8s.NewDesc(
		"datakit_goroutines",
		"Goroutine count within Datakit",
		nil, nil,
	)

	riHeapAllocDesc = p8s.NewDesc(
		"datakit_heap_alloc_bytes",
		"Datakit memory heap bytes",
		nil, nil,
	)

	riSysAllocDesc = p8s.NewDesc(
		"datakit_sys_alloc_bytes",
		"Datakit memory system bytes",
		nil, nil,
	)

	riCPUUsageDesc = p8s.NewDesc(
		"datakit_cpu_usage",
		"Datakit CPU usage(%)",
		nil, nil,
	)

	riGCPauseDesc = p8s.NewDesc(
		"datakit_gc_summary_seconds",
		"Datakit golang GC paused",
		nil, nil,
	)

	riOpenFilesDesc = p8s.NewDesc(
		"datakit_open_files",
		"Datakit open files(only available on Linux)",
		nil, nil,
	)

	riCPUCores = p8s.NewDesc(
		"datakit_cpu_cores",
		"Datakit CPU cores",
		nil, nil,
	)

	riUptimeDesc = p8s.NewDesc(
		"datakit_uptime_seconds",
		"Datakit uptime",

		// hostname and cgroup set after init(), so make it a non-const-label.
		[]string{
			"hostname",
			"resource_limit",
			"lite",
		},

		// these are const labels.
		p8s.Labels{
			"version":     datakit.Version,
			"build_at":    git.BuildAt,
			"branch":      git.Branch,
			"os_arch":     fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
			"docker":      fmt.Sprintf("%v", datakit.Docker),
			"auto_update": fmt.Sprintf("%v", datakit.AutoUpdate),
		},
	)

	riCtxSwitch = p8s.NewDesc(
		"datakit_process_ctx_switch_total",
		"Datakit process context switch count(Linux only)",
		[]string{
			"type", // voluntary or involuntary, see https://courses.cs.duke.edu/spring01/cps110/slides/interleave/tsld008.htm
		},
		nil,
	)

	riIOCount = p8s.NewDesc(
		"datakit_process_io_count_total",
		"Datakit process IO count",
		[]string{
			"type", // r(read) or w(write)
		},
		nil,
	)

	riIOBytes = p8s.NewDesc(
		"datakit_process_io_bytes_total",
		"Datakit process IO bytes count",
		[]string{
			"type", // r(read) or w(write)
		},
		nil,
	)

	riBeyondUsage = p8s.NewDesc(
		"datakit_data_overuse",
		"Does current workspace's data(metric/logging) usage(if 0 not beyond, or with a unix timestamp when overuse occurred)",
		nil,
		nil,
	)
)

// Describe implements Collector Describe interface.
func (rc runtimeInfoCollector) Describe(ch chan<- *p8s.Desc) {
	p8s.DescribeByCollect(rc, ch)
}

// Collect implements Collector Collect interface.
func (rc runtimeInfoCollector) Collect(ch chan<- p8s.Metric) {
	ri := getRuntimeInfo()

	ch <- p8s.MustNewConstSummary(riGCPauseDesc, uint64(ri.gcNum), float64(ri.gcPauseTotal)/float64(time.Second), nil)
	ch <- p8s.MustNewConstMetric(riGoroutineDesc, p8s.GaugeValue, float64(ri.goroutines))
	ch <- p8s.MustNewConstMetric(riHeapAllocDesc, p8s.GaugeValue, float64(ri.heapAlloc))
	ch <- p8s.MustNewConstMetric(riSysAllocDesc, p8s.GaugeValue, float64(ri.sys))
	ch <- p8s.MustNewConstMetric(riCPUUsageDesc, p8s.GaugeValue, ri.cpuUsage)
	ch <- p8s.MustNewConstMetric(riOpenFilesDesc, p8s.GaugeValue, float64(datakit.OpenFiles()))

	ch <- p8s.MustNewConstMetric(riCPUCores, p8s.GaugeValue, float64(runtime.NumCPU()))

	ch <- p8s.MustNewConstMetric(riUptimeDesc,
		p8s.GaugeValue,
		float64(time.Since(Uptime)/time.Second),
		datakit.DatakitHostName, resourcelimit.Info(), fmt.Sprintf("%v", datakit.Lite))
	ch <- p8s.MustNewConstMetric(riBeyondUsage, p8s.GaugeValue, float64(BeyondUsage))

	if ri.numCtxSwitch != nil {
		ch <- p8s.MustNewConstMetric(riCtxSwitch, p8s.CounterValue, float64(ri.numCtxSwitch.Voluntary), "voluntary")
		ch <- p8s.MustNewConstMetric(riCtxSwitch, p8s.CounterValue, float64(ri.numCtxSwitch.Involuntary), "involuntary")
	}

	if ri.ioCountersstats != nil {
		ch <- p8s.MustNewConstMetric(riIOCount, p8s.CounterValue, float64(ri.ioCountersstats.ReadCount), "r")
		ch <- p8s.MustNewConstMetric(riIOCount, p8s.CounterValue, float64(ri.ioCountersstats.WriteCount), "w")

		ch <- p8s.MustNewConstMetric(riIOBytes, p8s.CounterValue, float64(ri.ioCountersstats.ReadBytes), "r")
		ch <- p8s.MustNewConstMetric(riIOBytes, p8s.CounterValue, float64(ri.ioCountersstats.WriteBytes), "w")
	}
}
