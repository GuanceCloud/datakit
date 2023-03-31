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
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/git"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/cgroup"
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

type RuntimeInfo struct {
	Goroutines int     `json:"goroutines"`
	HeapAlloc  uint64  `json:"heap_alloc"`
	Sys        uint64  `json:"total_sys"`
	CPUUsage   float64 `json:"cpu_usage"`

	GCPauseTotal uint64        `json:"gc_pause_total"`
	GCNum        uint32        `json:"gc_num"`
	GCAvgCost    time.Duration `json:"gc_avg_bytes"`
}

func GetRuntimeInfo() *RuntimeInfo {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	var usage float64
	if u, err := cgroup.GetCPUPercent(3 * time.Millisecond); err == nil {
		usage = u
	}

	return &RuntimeInfo{
		Goroutines: runtime.NumGoroutine(),
		HeapAlloc:  m.HeapAlloc,
		Sys:        m.Sys,
		CPUUsage:   usage,

		GCPauseTotal: m.PauseTotalNs,
		GCNum:        m.NumGC,
	}
}

// collector for basic runtime info.
// These metrics are self generated, not triggered by outter
// operation(such as HTTP request, io Feed point).
type runtimeInfoCollector struct{}

var (
	riGoroutineDesc = p8s.NewDesc(
		"datakit_goroutines",
		"goroutine count within Datakit",
		nil, nil,
	)

	riHeapAllocDesc = p8s.NewDesc(
		"datakit_heap_alloc",
		"Datakit memory heap bytes",
		nil, nil,
	)

	riSysAllocDesc = p8s.NewDesc(
		"datakit_sys_alloc",
		"Datakit memory system bytes",
		nil, nil,
	)

	riCPUUsageDesc = p8s.NewDesc(
		"datakit_cpu_usage",
		"Datakit CPU usage(%)",
		nil, nil,
	)

	riGCPauseDesc = p8s.NewDesc(
		"datakit_gc_summary",
		"Datakit golang GC paused(nano-second)",
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
		"datakit_uptime",
		"Datakit uptime(second)",

		// hostname and cgroup set after init(), so make it a non-const-label.
		[]string{
			"hostname",
			"cgroup",
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

	riBeyondUsage = p8s.NewDesc(
		"datakit_data_overuse",
		"Does current workspace's data(metric/logging) usaguse(if 0 not beyond, or with a unix timestamp when overuse occurred)",
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
	ri := GetRuntimeInfo()

	ch <- p8s.MustNewConstSummary(riGCPauseDesc, uint64(ri.GCNum), float64(ri.GCPauseTotal), nil)
	ch <- p8s.MustNewConstMetric(riGoroutineDesc, p8s.GaugeValue, float64(ri.Goroutines))
	ch <- p8s.MustNewConstMetric(riHeapAllocDesc, p8s.GaugeValue, float64(ri.HeapAlloc))
	ch <- p8s.MustNewConstMetric(riSysAllocDesc, p8s.GaugeValue, float64(ri.Sys))
	ch <- p8s.MustNewConstMetric(riCPUUsageDesc, p8s.GaugeValue, ri.CPUUsage*100.0)
	ch <- p8s.MustNewConstMetric(riOpenFilesDesc, p8s.GaugeValue, float64(datakit.OpenFiles()))

	ch <- p8s.MustNewConstMetric(riCPUCores, p8s.GaugeValue, float64(runtime.NumCPU()))

	ch <- p8s.MustNewConstMetric(riUptimeDesc,
		p8s.GaugeValue,
		float64(time.Since(Uptime)/time.Second),
		datakit.DatakitHostName, cgroup.Info())
	ch <- p8s.MustNewConstMetric(riBeyondUsage, p8s.GaugeValue, float64(BeyondUsage))
}
