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
	openfilesVec          *prometheus.GaugeVec
	rotateVec             *prometheus.CounterVec
	parseFailVec          *prometheus.CounterVec
	multilineStateVec     *prometheus.CounterVec

	socketLogConnect *prometheus.CounterVec
	socketLogCount   *prometheus.CounterVec
	socketLogLength  *prometheus.SummaryVec

	decodeErrors *prometheus.CounterVec
)

func setupMetrics() {
	receiveCreateEventVec = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "datakit",
			Subsystem: "tailer",
			Name:      "receive_create_event_total",
			Help:      "Total number of received create events",
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
			Help:      "Total number of discarded based on the whitelist",
		},
		[]string{
			"source",
			"filepath",
		},
	)

	openfilesVec = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "datakit",
			Subsystem: "tailer",
			Name:      "open_files",
			Help:      "Total number of currently open files",
		},
		[]string{
			"source",
			"max",
		},
	)

	rotateVec = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "datakit",
			Subsystem: "tailer",
			Name:      "file_rotate_total",
			Help:      "Total number of file rotations performed",
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
			Help:      "Total number of failed parse attempts",
		},
		[]string{
			"source",
			"filepath",
			"mode",
		},
	)

	multilineStateVec = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "datakit",
			Subsystem: "tailer",
			Name:      "multiline_state_total",
			Help:      "Total number of multiline states encountered",
		},
		[]string{
			"source",
			"state",
		},
	)

	socketLogConnect = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "datakit",
			Subsystem: "tailer",
			Name:      "socket_connect_status_total",
			Help:      "Total number of socket connection status events",
		},
		[]string{"network", "status"},
	)

	socketLogCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "datakit",
			Subsystem: "tailer",
			Name:      "socket_feed_message_count_total",
			Help:      "Total number of messages fed to the socket",
		},
		[]string{
			"network",
		},
	)

	socketLogLength = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace: "datakit",
			Subsystem: "tailer",
			Name:      "socket_log_length",
			Help:      "Length of the log for socket communication",
			Objectives: map[float64]float64{
				0.5:  0.05,
				0.90: 0.01,
				0.99: 0.001,
			},
		},
		[]string{"network"},
	)

	decodeErrors = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "datakit",
			Subsystem: "tailer",
			Name:      "text_decode_errors_total",
			Help:      "Total count of text decoding failures",
		},
		[]string{
			"source",
			"character_encoding",
			"error_type",
		},
	)

	metrics.MustRegister(
		receiveCreateEventVec,
		discardVec,
		openfilesVec,
		parseFailVec,
		multilineStateVec,
		rotateVec,
		socketLogLength,
		socketLogCount,
		socketLogConnect,
		decodeErrors,
	)
}

//nolint:gochecknoinits
func init() {
	setupMetrics()
}
