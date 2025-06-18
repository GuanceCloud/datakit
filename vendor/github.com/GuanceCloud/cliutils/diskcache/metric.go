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
	rotateVec,
	removeVec,
	wakeupVec,
	posUpdatedVec,
	seekBackVec *prometheus.CounterVec

	sizeVec,
	openTimeVec,
	lastCloseTimeVec,
	capVec,
	maxDataVec,
	batchSizeVec,
	datafilesVec *prometheus.GaugeVec

	droppedDataVec,
	putBytesVec,
	getBytesVec,
	streamPutVec,
	getLatencyVec,
	putLatencyVec *prometheus.SummaryVec

	ns = "diskcache"
)

func setupMetrics() {
	streamPutVec = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace: ns,
			Name:      "stream_put",
			Help:      "Stream put times",
			Objectives: map[float64]float64{
				0.5:  0.05,
				0.9:  0.01,
				0.99: 0.001,
			},
		},
		[]string{"path"},
	)

	getLatencyVec = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace: ns,
			Name:      "get_latency",
			Help:      "Get() cost seconds",
			Objectives: map[float64]float64{
				0.5:  0.05,
				0.9:  0.01,
				0.99: 0.001,
			},
		},
		[]string{"path"},
	)

	putLatencyVec = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace: ns,
			Name:      "put_latency",
			Help:      "Put() cost seconds",
			Objectives: map[float64]float64{
				0.5:  0.05,
				0.9:  0.01,
				0.99: 0.001,
			},
		},
		[]string{"path"},
	)

	putBytesVec = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace: ns,
			Name:      "put_bytes",
			Help:      "Cache Put() bytes",
			Objectives: map[float64]float64{
				0.5:  0.05,
				0.9:  0.01,
				0.99: 0.001,
			},
		},
		[]string{"path"},
	)

	getBytesVec = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace: ns,
			Name:      "get_bytes",
			Help:      "Cache Get() bytes",
			Objectives: map[float64]float64{
				0.5:  0.05,
				0.9:  0.01,
				0.99: 0.001,
			},
		},
		[]string{"path"},
	)

	droppedDataVec = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace: ns,
			Name:      "dropped_data",
			Help:      "Dropped data during Put() when capacity reached.",
			Objectives: map[float64]float64{
				0.5:  0.05,
				0.9:  0.01,
				0.99: 0.001,
			},
		},
		[]string{"path", "reason"},
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

	wakeupVec = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: ns,
			Name:      "wakeup_total",
			Help:      "Wakeup count on sleeping write file",
		},
		[]string{"path"},
	)

	posUpdatedVec = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: ns,
			Name:      "pos_updated_total",
			Help:      ".pos file updated count",
		},
		[]string{"op", "path"},
	)

	seekBackVec = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: ns,
			Name:      "seek_back_total",
			Help:      "Seek back when Get() got any error",
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

	metrics.MustRegister(Metrics()...)
}

// ResetMetrics used to cleanup exist metrics of diskcache.
func ResetMetrics() {
	streamPutVec.Reset()
	droppedDataVec.Reset()
	rotateVec.Reset()
	wakeupVec.Reset()
	posUpdatedVec.Reset()
	seekBackVec.Reset()
	capVec.Reset()
	batchSizeVec.Reset()
	maxDataVec.Reset()
	sizeVec.Reset()
	datafilesVec.Reset()
	getLatencyVec.Reset()
	putLatencyVec.Reset()
	putBytesVec.Reset()
	getBytesVec.Reset()
}

func Metrics() []prometheus.Collector {
	return []prometheus.Collector{
		droppedDataVec,
		rotateVec,
		removeVec,
		wakeupVec,
		posUpdatedVec,
		seekBackVec,

		sizeVec,
		openTimeVec,
		lastCloseTimeVec,
		capVec,
		maxDataVec,
		batchSizeVec,
		datafilesVec,

		getLatencyVec,
		putLatencyVec,
		getBytesVec,
		putBytesVec,
	}
}

// nolint: gochecknoinits
func init() {
	setupMetrics()
}
