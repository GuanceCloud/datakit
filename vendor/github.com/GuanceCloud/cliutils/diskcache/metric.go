// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package diskcache

import (
	"github.com/GuanceCloud/cliutils/metrics"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	droppedBatchVec,
	droppedBytesVec,
	rotateVec,
	removeVec,
	putVec,
	getVec,
	putBytesVec,
	wakeupVec,
	getBytesVec *prometheus.CounterVec

	sizeVec,
	openTimeVec,
	lastCloseTimeVec,
	capVec,
	maxDataVec,
	batchSizeVec,
	datafilesVec *prometheus.GaugeVec

	getLatencyVec,
	putLatencyVec *prometheus.SummaryVec

	ns = "diskcache"
)

func setupMetrics() {
	getLatencyVec = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace: ns,
			Name:      "get_latency",
			Help:      "Get() time cost(micro-second)",
		},
		[]string{"path"},
	)

	putLatencyVec = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace: ns,
			Name:      "put_latency",
			Help:      "Put() time cost(micro-second)",
		},
		[]string{"path"},
	)

	droppedBytesVec = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: ns,
			Name:      "dropped_bytes_total",
			Help:      "Dropped bytes during Put() when capacity reached.",
		},
		[]string{"path"},
	)

	droppedBatchVec = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: ns,
			Name:      "dropped_total",
			Help:      "Dropped files during Put() when capacity reached.",
		},
		[]string{"path"},
	)

	rotateVec = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: ns,
			Name:      "rotate_total",
			Help:      "Cache rotate count, mean file rotate from data to data.0000xxx",
		},
		[]string{"path"},
	)

	removeVec = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: ns,
			Name:      "remove_total",
			Help:      "Removed file count, if some file read EOF, remove it from un-read list",
		},
		[]string{"path"},
	)

	putVec = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: ns,
			Name:      "put_total",
			Help:      "Cache Put() count",
		},
		[]string{"path"},
	)

	putBytesVec = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: ns,
			Name:      "put_bytes_total",
			Help:      "Cache Put() bytes count",
		},
		[]string{"path"},
	)

	getVec = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: ns,
			Name:      "get_total",
			Help:      "Cache Get() count",
		},
		[]string{"path"},
	)

	wakeupVec = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: ns,
			Name:      "wakeup_total",
			Help:      "Wakeup count on sleeping write file",
		},
		[]string{"path"},
	)

	getBytesVec = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: ns,
			Name:      "get_bytes_total",
			Help:      "Cache Get() bytes count",
		},
		[]string{"path"},
	)

	capVec = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: ns,
			Name:      "capacity",
			Help:      "Current capacity(in bytes)",
		},
		[]string{"path"},
	)

	maxDataVec = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: ns,
			Name:      "max_data",
			Help:      "Max data to Put(in bytes), default 0",
		},
		[]string{"path"},
	)

	batchSizeVec = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: ns,
			Name:      "batch_size",
			Help:      "Data file size(in bytes)",
		},
		[]string{"path"},
	)

	sizeVec = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: ns,
			Name:      "size",
			Help:      "Current cache size(in bytes)",
		},
		[]string{"path"},
	)

	openTimeVec = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: ns,
			Name:      "open_time",
			Help:      "Current cache Open time in unix timestamp(second)",
		},
		[]string{
			// NOTE: make them sorted.
			"no_fallback_on_error",
			"no_lock",
			"no_pos",
			"no_sync",
			"path",
		},
	)

	lastCloseTimeVec = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: ns,
			Name:      "last_close_time",
			Help:      "Current cache last Close time in unix timestamp(second)",
		},
		[]string{"path"},
	)

	datafilesVec = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: ns,
			Name:      "datafiles",
			Help:      "Current un-read data files",
		},
		[]string{"path"},
	)

	metrics.MustRegister(
		droppedBatchVec,
		droppedBytesVec,
		rotateVec,
		putVec,
		getVec,
		putBytesVec,
		wakeupVec,
		getBytesVec,

		openTimeVec,
		lastCloseTimeVec,
		capVec,
		batchSizeVec,
		maxDataVec,
		sizeVec,
		datafilesVec,

		getLatencyVec,
		putLatencyVec)
}

// register to specified registry for testing.
func register(reg *prometheus.Registry) {
	reg.MustRegister(
		droppedBatchVec,
		droppedBytesVec,
		rotateVec,
		putVec,
		getVec,
		putBytesVec,
		wakeupVec,
		getBytesVec,

		capVec,
		batchSizeVec,
		maxDataVec,
		sizeVec,
		datafilesVec,

		getLatencyVec,
		putLatencyVec)
}

// ResetMetrics used to cleanup exist metrics of diskcache.
func ResetMetrics() {
	droppedBatchVec.Reset()
	droppedBytesVec.Reset()
	rotateVec.Reset()
	putVec.Reset()
	getVec.Reset()
	putBytesVec.Reset()
	wakeupVec.Reset()
	getBytesVec.Reset()
	capVec.Reset()
	batchSizeVec.Reset()
	maxDataVec.Reset()
	sizeVec.Reset()
	datafilesVec.Reset()
	getLatencyVec.Reset()
	putLatencyVec.Reset()
}

// Labels export cache's labels used to query prometheus metrics.
//func (c *DiskCache) Labels() []string {
//	return c.labels
//}

func Metrics() []prometheus.Collector {
	return []prometheus.Collector{
		droppedBatchVec,
		droppedBytesVec,
		rotateVec,
		removeVec,
		putVec,
		getVec,
		putBytesVec,
		wakeupVec,
		getBytesVec,

		sizeVec,
		openTimeVec,
		lastCloseTimeVec,
		capVec,
		maxDataVec,
		batchSizeVec,
		datafilesVec,

		getLatencyVec,
		putLatencyVec,
	}
}

// nolint: gochecknoinits
func init() {
	setupMetrics()
}
