//go:build linux
// +build linux

package stats

import (
	"runtime"
	"sync"

	p8s "github.com/prometheus/client_golang/prometheus"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/resourcelimit"
)

var runtimeMetricsOnce sync.Once

type runtimeInfoCollector struct{}

var (
	runtimeMemDesc = p8s.NewDesc(
		"dkebpf_runtime_memory_bytes",
		"datakit-ebpf memory usage stats in bytes",
		[]string{"type"}, nil,
	)

	runtimeMemObjectsDesc = p8s.NewDesc(
		"dkebpf_runtime_memory_objects",
		"datakit-ebpf Go heap object stats",
		[]string{"type"}, nil,
	)

	runtimeMemCounterDesc = p8s.NewDesc(
		"dkebpf_runtime_memory_ops_total",
		"datakit-ebpf Go memory operation counters",
		[]string{"type"}, nil,
	)

	runtimeMemRatioDesc = p8s.NewDesc(
		"dkebpf_runtime_memory_ratio",
		"datakit-ebpf memory ratios",
		[]string{"type"}, nil,
	)

	runtimeGCDesc = p8s.NewDesc(
		"dkebpf_runtime_gc",
		"datakit-ebpf Go GC stats",
		[]string{"type"}, nil,
	)

	runtimeGoroutinesDesc = p8s.NewDesc(
		"dkebpf_runtime_goroutines",
		"datakit-ebpf goroutine count",
		nil, nil,
	)
)

func RegisterRuntimeMetrics() {
	runtimeMetricsOnce.Do(func() {
		MustRegister(runtimeInfoCollector{})
	})
}

//nolint:gochecknoinits
func init() {
	RegisterRuntimeMetrics()
}

func (c runtimeInfoCollector) Describe(ch chan<- *p8s.Desc) {
	p8s.DescribeByCollect(c, ch)
}

func (c runtimeInfoCollector) Collect(ch chan<- p8s.Metric) {
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)

	ch <- p8s.MustNewConstMetric(runtimeGoroutinesDesc, p8s.GaugeValue, float64(runtime.NumGoroutine()))

	goTotal := saturatingSub(ms.Sys, ms.HeapReleased)
	goHeapUnused := saturatingSub(ms.HeapInuse, ms.HeapAlloc)
	goHeapFree := saturatingSub(ms.HeapIdle, ms.HeapReleased)

	emitBytes := func(metricType string, value uint64) {
		ch <- p8s.MustNewConstMetric(runtimeMemDesc, p8s.GaugeValue, float64(value), metricType)
	}

	emitBytes("go_total", goTotal)
	emitBytes("heap_alloc", ms.HeapAlloc)
	emitBytes("heap_inuse", ms.HeapInuse)
	emitBytes("heap_idle", ms.HeapIdle)
	emitBytes("heap_released", ms.HeapReleased)
	emitBytes("heap_unused", goHeapUnused)
	emitBytes("heap_free", goHeapFree)
	emitBytes("stack_inuse", ms.StackInuse)
	emitBytes("stack_sys", ms.StackSys)
	emitBytes("mspan_inuse", ms.MSpanInuse)
	emitBytes("mspan_sys", ms.MSpanSys)
	emitBytes("mcache_inuse", ms.MCacheInuse)
	emitBytes("mcache_sys", ms.MCacheSys)
	emitBytes("buck_hash_sys", ms.BuckHashSys)
	emitBytes("gc_sys", ms.GCSys)
	emitBytes("other_sys", ms.OtherSys)
	emitBytes("next_gc", ms.NextGC)
	emitBytes("total_alloc", ms.TotalAlloc)

	ch <- p8s.MustNewConstMetric(runtimeMemObjectsDesc, p8s.GaugeValue, float64(ms.HeapObjects), "heap")

	ch <- p8s.MustNewConstMetric(runtimeMemCounterDesc, p8s.CounterValue, float64(ms.Mallocs), "mallocs")
	ch <- p8s.MustNewConstMetric(runtimeMemCounterDesc, p8s.CounterValue, float64(ms.Frees), "frees")

	ch <- p8s.MustNewConstMetric(runtimeGCDesc, p8s.GaugeValue, float64(ms.NumGC), "cycles")
	ch <- p8s.MustNewConstMetric(runtimeGCDesc, p8s.GaugeValue, float64(ms.LastGC)/1e9, "last_gc_unix")
	ch <- p8s.MustNewConstMetric(runtimeGCDesc, p8s.GaugeValue, float64(ms.PauseTotalNs)/1e9, "pause_total_seconds")
	ch <- p8s.MustNewConstMetric(runtimeGCDesc, p8s.GaugeValue, ms.GCCPUFraction, "cpu_fraction")

	if procMem, err := resourcelimit.MyMemStat(); err == nil && procMem != nil {
		emitBytes("rss", procMem.RSS)
		emitBytes("vms", procMem.VMS)
		emitBytes("hwm", procMem.HWM)
		emitBytes("data", procMem.Data)
		emitBytes("stack", procMem.Stack)
		emitBytes("locked", procMem.Locked)
		emitBytes("swap", procMem.Swap)

		nonGoEstimate := saturatingSub(procMem.RSS, goTotal)
		nonHeapEstimate := saturatingSub(procMem.RSS, ms.HeapAlloc)
		emitBytes("non_go_estimate", nonGoEstimate)
		emitBytes("non_heap_estimate", nonHeapEstimate)

		if procMem.RSS > 0 {
			ch <- p8s.MustNewConstMetric(runtimeMemRatioDesc, p8s.GaugeValue, float64(goTotal)/float64(procMem.RSS), "go_total_vs_rss")
			ch <- p8s.MustNewConstMetric(runtimeMemRatioDesc, p8s.GaugeValue, float64(ms.HeapAlloc)/float64(procMem.RSS), "heap_alloc_vs_rss")
			ch <- p8s.MustNewConstMetric(runtimeMemRatioDesc, p8s.GaugeValue, float64(nonGoEstimate)/float64(procMem.RSS), "non_go_vs_rss")
		}
	}
}

func saturatingSub(a, b uint64) uint64 {
	if a < b {
		return 0
	}
	return a - b
}
