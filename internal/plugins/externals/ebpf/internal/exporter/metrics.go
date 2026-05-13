package exporter

import (
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	ePtsVec = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "dkebpf",
			Subsystem: "exporter",
			Name:      "points_total",
			Help:      "The number of data points processed by the exporter",
		},
		[]string{"name", "category"},
	)

	eBPFMapEntriesVec = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "dkebpf",
			Subsystem: "exporter",
			Name:      "bpf_map_entries",
			Help:      "Current number of entries observed in an eBPF map",
		},
		[]string{"component", "map"},
	)

	eBPFMapMaxEntriesVec = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "dkebpf",
			Subsystem: "exporter",
			Name:      "bpf_map_max_entries",
			Help:      "Configured maximum number of entries for an eBPF map",
		},
		[]string{"component", "map"},
	)

	eBPFMapFillRatioVec = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "dkebpf",
			Subsystem: "exporter",
			Name:      "bpf_map_fill_ratio",
			Help:      "Observed fill ratio of an eBPF map",
		},
		[]string{"component", "map"},
	)

	eBPFMapCleanupVec = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "dkebpf",
			Subsystem: "exporter",
			Name:      "bpf_map_cleanup_total",
			Help:      "The number of eBPF map cleanup attempts",
		},
		[]string{"component", "map", "result"},
	)

	eBPFEventDropVec = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "dkebpf",
			Subsystem: "exporter",
			Name:      "bpf_events_dropped_total",
			Help:      "The number of dropped eBPF events observed in userspace",
		},
		[]string{"component", "event", "reason"},
	)

	eBPFMapObserveErrVec = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "dkebpf",
			Subsystem: "exporter",
			Name:      "bpf_map_observe_errors_total",
			Help:      "The number of errors while observing eBPF map state",
		},
		[]string{"component", "map", "operation"},
	)

	eAggEntriesVec = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "dkebpf",
			Subsystem: "exporter",
			Name:      "agg_entries",
			Help:      "Current number of entries held by an in-memory aggregation pool",
		},
		[]string{"component"},
	)

	eAggFlushPointsVec = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "dkebpf",
			Subsystem: "exporter",
			Name:      "agg_flush_points_total",
			Help:      "The number of points emitted by aggregation flushes",
		},
		[]string{"component", "result"},
	)

	eAggFlushDurationVec = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "dkebpf",
			Subsystem: "exporter",
			Name:      "agg_flush_duration_seconds",
			Help:      "Aggregation flush duration in seconds",
			Buckets:   prometheus.ExponentialBuckets(0.0005, 2, 16),
		},
		[]string{"component", "result"},
	)

	eCacheEntriesVec = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "dkebpf",
			Subsystem: "exporter",
			Name:      "cache_entries",
			Help:      "Current number of entries held by an in-memory cache or queue",
		},
		[]string{"component", "cache"},
	)

	eCacheEvictionsTotalVec = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "dkebpf",
			Subsystem: "exporter",
			Name:      "cache_evictions_total",
			Help:      "Total number of entries evicted from in-memory caches",
		},
		[]string{"component", "cache", "reason"},
	)

	eSenderQueueLen = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "dkebpf",
			Subsystem: "exporter",
			Name:      "sender_queue_length",
			Help:      "Current number of pending send tasks in the exporter queue",
		},
	)

	eSenderBatchPoints = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Namespace: "dkebpf",
			Subsystem: "exporter",
			Name:      "sender_batch_points",
			Help:      "Number of points per exporter send batch",
			Buckets:   []float64{1, 4, 8, 16, 32, 64, 128, 256},
		},
	)

	eSenderBatchBytes = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Namespace: "dkebpf",
			Subsystem: "exporter",
			Name:      "sender_batch_bytes",
			Help:      "Encoded bytes per exporter send batch",
			Buckets:   prometheus.ExponentialBuckets(256, 2, 12),
		},
	)

	eSenderRequestTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "dkebpf",
			Subsystem: "exporter",
			Name:      "sender_requests_total",
			Help:      "Total number of exporter send requests",
		},
		[]string{"result"},
	)

	eSenderRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "dkebpf",
			Subsystem: "exporter",
			Name:      "sender_request_duration_seconds",
			Help:      "Exporter send request duration in seconds",
			Buckets:   prometheus.ExponentialBuckets(0.001, 2, 16),
		},
		[]string{"result"},
	)

	ePerfLostTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "dkebpf",
			Subsystem: "exporter",
			Name:      "perf_lost_samples_total",
			Help:      "Total number of lost perf samples observed in userspace",
		},
		[]string{"component", "stream"},
	)

	ePerfReadErrorsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "dkebpf",
			Subsystem: "exporter",
			Name:      "perf_read_errors_total",
			Help:      "Total number of perf reader errors observed in userspace",
		},
		[]string{"component", "stream"},
	)

	eTPacketStatsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "dkebpf",
			Subsystem: "exporter",
			Name:      "tpacket_stats_total",
			Help:      "Total number of packets, drops, and queue freezes reported by TPacket sockets",
		},
		[]string{"component", "type"},
	)

	//nolint:promlinter
	eNICGroupCountVec = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "dkebpf",
			Subsystem: "exporter",
			Name:      "nic_group_count",
			Help:      "Current number of logical NICs in each capture group",
		},
		[]string{"group", "virtual_nic"},
	)

	//nolint:promlinter
	eNICGroupRouteCountVec = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "dkebpf",
			Subsystem: "exporter",
			Name:      "nic_group_route_count",
			Help:      "Current number of active route bindings in each capture group",
		},
		[]string{"group", "virtual_nic"},
	)

	eAsyncQueueWaitDurationVec = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "dkebpf",
			Subsystem: "exporter",
			Name:      "async_queue_wait_duration_seconds",
			Help:      "Time spent waiting to enqueue async work in userspace",
			Buckets:   prometheus.ExponentialBuckets(0.00001, 2, 16),
		},
		[]string{"component"},
	)

	eAsyncProcessDurationVec = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "dkebpf",
			Subsystem: "exporter",
			Name:      "async_process_duration_seconds",
			Help:      "Time spent processing one async work item in userspace",
			Buckets:   prometheus.ExponentialBuckets(0.00001, 2, 18),
		},
		[]string{"component"},
	)
)

func ObserveBPFMap(component, mapName string, entries, maxEntries uint32) {
	eBPFMapEntriesVec.WithLabelValues(component, mapName).Set(float64(entries))
	eBPFMapMaxEntriesVec.WithLabelValues(component, mapName).Set(float64(maxEntries))

	fillRatio := 0.0
	if maxEntries > 0 {
		fillRatio = float64(entries) / float64(maxEntries)
	}
	eBPFMapFillRatioVec.WithLabelValues(component, mapName).Set(fillRatio)
}

func AddBPFMapCleanup(component, mapName, result string, n float64) {
	if n <= 0 {
		return
	}
	eBPFMapCleanupVec.WithLabelValues(component, mapName, result).Add(n)
}

func IncBPFEventDrop(component, event, reason string) {
	eBPFEventDropVec.WithLabelValues(component, event, reason).Inc()
}

func IncBPFMapObserveError(component, mapName, operation string) {
	eBPFMapObserveErrVec.WithLabelValues(component, mapName, operation).Inc()
}

func ObserveAggEntries(component string, entries int) {
	eAggEntriesVec.WithLabelValues(component).Set(float64(entries))
}

func ObserveAggFlush(component string, points int, dur time.Duration, result string) {
	eAggFlushPointsVec.WithLabelValues(component, result).Add(float64(points))
	eAggFlushDurationVec.WithLabelValues(component, result).Observe(dur.Seconds())
}

func ObserveCacheEntries(component, cache string, entries int) {
	eCacheEntriesVec.WithLabelValues(component, cache).Set(float64(entries))
}

func AddCacheEvictions(component, cache, reason string, entries int) {
	if entries <= 0 {
		return
	}
	eCacheEvictionsTotalVec.WithLabelValues(component, cache, reason).Add(float64(entries))
}

func ObserveSenderQueue(entries int) {
	eSenderQueueLen.Set(float64(entries))
}

func ObserveSenderBatch(points int, bytes int) {
	eSenderBatchPoints.Observe(float64(points))
	eSenderBatchBytes.Observe(float64(bytes))
}

func ObserveSenderRequest(result string, dur time.Duration) {
	eSenderRequestTotal.WithLabelValues(result).Inc()
	eSenderRequestDuration.WithLabelValues(result).Observe(dur.Seconds())
}

func AddPerfLost(component, stream string, count uint64) {
	ePerfLostTotal.WithLabelValues(component, stream).Add(float64(count))
}

func IncPerfReadError(component, stream string) {
	ePerfReadErrorsTotal.WithLabelValues(component, stream).Inc()
}

func AddTPacketStats(component string, packets, drops, freezes uint64) {
	if packets > 0 {
		eTPacketStatsTotal.WithLabelValues(component, "packets").Add(float64(packets))
	}
	if drops > 0 {
		eTPacketStatsTotal.WithLabelValues(component, "drops").Add(float64(drops))
	}
	if freezes > 0 {
		eTPacketStatsTotal.WithLabelValues(component, "queue_freezes").Add(float64(freezes))
	}
}

func ObserveNICGroupCount(group string, virtualNIC bool, count int) {
	eNICGroupCountVec.WithLabelValues(group, strconv.FormatBool(virtualNIC)).Set(float64(count))
}

func ObserveNICGroupRouteCount(group string, virtualNIC bool, count int) {
	eNICGroupRouteCountVec.WithLabelValues(group, strconv.FormatBool(virtualNIC)).Set(float64(count))
}

func ObserveAsyncQueueWait(component string, dur time.Duration) {
	eAsyncQueueWaitDurationVec.WithLabelValues(component).Observe(dur.Seconds())
}

func ObserveAsyncProcess(component string, dur time.Duration) {
	eAsyncProcessDurationVec.WithLabelValues(component).Observe(dur.Seconds())
}
