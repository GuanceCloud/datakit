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

	expLabels = []string{
		// NOTE: make them sorted.
		"no_fallback_on_error",
		"no_lock",
		"no_pos",
		"no_sync",
		"path",
	}
)

func setupMetrics() {
	getLatencyVec = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace: ns,
			Name:      "get_latency",
			Help:      "Get() time cost(micro-second)",
		},
		expLabels,
	)

	putLatencyVec = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace: ns,
			Name:      "put_latency",
			Help:      "Put() time cost(micro-second)",
		},
		expLabels,
	)

	droppedBytesVec = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: ns,
			Name:      "dropped_bytes_total",
			Help:      "dropped bytes during Put() when capacity reached.",
		},
		expLabels,
	)

	droppedBatchVec = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: ns,
			Name:      "dropped_total",
			Help:      "dropped files during Put() when capacity reached.",
		},
		expLabels,
	)

	rotateVec = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: ns,
			Name:      "rotate_total",
			Help:      "cache rotate count, mean file rotate from data to data.0000xxx",
		},
		expLabels,
	)

	removeVec = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: ns,
			Name:      "remove_total",
			Help:      "removed file count, if some file read EOF, remove it from un-readed list",
		},
		expLabels,
	)

	putVec = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: ns,
			Name:      "put_total",
			Help:      "cache Put() count",
		},
		expLabels,
	)

	putBytesVec = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: ns,
			Name:      "put_bytes_total",
			Help:      "cache Put() bytes count",
		},
		expLabels,
	)

	getVec = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: ns,
			Name:      "get_total",
			Help:      "cache Get() count",
		},
		expLabels,
	)

	wakeupVec = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: ns,
			Name:      "wakeup_total",
			Help:      "wakeup count on sleeping write file",
		}, expLabels,
	)

	getBytesVec = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: ns,
			Name:      "get_bytes_total",
			Help:      "cache Get() bytes count",
		},
		expLabels,
	)

	capVec = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: ns,
			Name:      "capacity",
			Help:      "current capacity(in bytes)",
		},
		expLabels,
	)

	maxDataVec = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: ns,
			Name:      "max_data",
			Help:      "max data to Put(in bytes), default 0",
		},
		expLabels,
	)

	batchSizeVec = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: ns,
			Name:      "batch_size",
			Help:      "data file size(in bytes)",
		},
		expLabels,
	)

	sizeVec = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: ns,
			Name:      "size",
			Help:      "current cache size(in bytes)",
		},
		expLabels,
	)

	openTimeVec = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: ns,
			Name:      "open_time",
			Help:      "current cache Open time in unix timestamp(second)",
		},
		expLabels,
	)

	lastCloseTimeVec = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: ns,
			Name:      "last_close_time",
			Help:      "current cache last Close time in unix timestamp(second)",
		},
		expLabels,
	)

	datafilesVec = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: ns,
			Name:      "datafiles",
			Help:      "current un-readed data files",
		},
		expLabels,
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
func (c *DiskCache) Labels() []string {
	return c.labels
}

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
