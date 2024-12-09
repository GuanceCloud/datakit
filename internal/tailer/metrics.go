// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package tailer

import (
	"github.com/GuanceCloud/cliutils/metrics"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	receiveCreateEventVec *prometheus.CounterVec
	discardVec            *prometheus.CounterVec
	openfileVec           *prometheus.GaugeVec
	rotateVec             *prometheus.CounterVec
	parseFailVec          *prometheus.CounterVec
	socketLogConnect      *prometheus.CounterVec
	socketLogCount        *prometheus.CounterVec
	socketLogLength       *prometheus.SummaryVec
)

func setupMetrics() {
	receiveCreateEventVec = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "datakit",
			Subsystem: "tailer",
			Name:      "receive_create_event_total",
			Help:      "Total number of 'CREATE' events received",
		},
		[]string{
			"source",
			"type",
		},
	)

	discardVec = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "datakit",
			Subsystem: "tailer",
			Name:      "discard_log_total",
			Help:      "Total logs discarded based on the whitelist",
		},
		[]string{
			"source",
			"filepath",
		},
	)

	openfileVec = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "datakit",
			Subsystem: "tailer",
			Name:      "open_file_num",
			Help:      "Tailer open file total",
		},
		[]string{
			"mode",
		},
	)

	rotateVec = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "datakit",
			Subsystem: "tailer",
			Name:      "file_rotate_total",
			Help:      "Total tailer rotated",
		},
		[]string{
			"source",
			"filepath",
		},
	)

	parseFailVec = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "datakit",
			Subsystem: "tailer",
			Name:      "parse_fail_total",
			Help:      "Total tailer parsing failed",
		},
		[]string{
			"source",
			"filepath",
			"mode",
		},
	)

	socketLogConnect = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "datakit",
			Subsystem: "input_logging_socket",
			Name:      "connect_status_total",
			Help:      "Connect and close count for net.conn",
		},
		[]string{"network", "status"},
	)

	socketLogCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "datakit",
			Subsystem: "input_logging_socket",
			Name:      "feed_message_count_total",
			Help:      "Socket feed to IO message count",
		},
		[]string{
			"network",
		},
	)

	socketLogLength = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace: "datakit",
			Subsystem: "input_logging_socket",
			Name:      "log_length",
			Help:      "Record the length of each log line",
			Objectives: map[float64]float64{
				0.5:  0.05,
				0.90: 0.01,
				0.99: 0.001,
			},
		},
		[]string{"network"},
	)

	metrics.MustRegister(
		receiveCreateEventVec,
		discardVec,
		openfileVec,
		parseFailVec,
		rotateVec,
		socketLogLength,
		socketLogCount,
		socketLogConnect,
	)
}

//nolint:gochecknoinits
func init() {
	setupMetrics()
}
