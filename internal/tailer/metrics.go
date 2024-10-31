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
	rotateVec        *prometheus.CounterVec
	parseFailVec     *prometheus.CounterVec
	openfileVec      *prometheus.GaugeVec
	socketLogConnect *prometheus.CounterVec
	socketLogCount   *prometheus.CounterVec
	socketLogLength  *prometheus.SummaryVec

	pendingBlockLength *prometheus.GaugeVec
	pendingByteSize    *prometheus.GaugeVec
)

func setupMetrics() {
	rotateVec = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "datakit",
			Subsystem: "tailer",
			Name:      "file_rotate_total",
			Help:      "Tailer rotate total",
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
			Help:      "Tailer parse fail total",
		},
		[]string{
			"source",
			"filepath",
			"mode",
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
	socketLogConnect = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "datakit",
			Subsystem: "input_logging_socket",
			Name:      "connect_status_total",
			Help:      "Connect and close count for net.conn",
		},
		[]string{"network", "status"})

	socketLogCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "datakit",
			Subsystem: "input_logging_socket",
			Name:      "feed_message_count_total",
			Help:      "Socket feed to IO message count",
		},
		[]string{
			"network",
		})

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
		[]string{"network"})

	pendingBlockLength = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "datakit",
			Subsystem: "input_logging",
			Name:      "pending_block_length",
			Help:      "The length of blocks that are pending processing",
		},
		[]string{"source", "filepath"})

	pendingByteSize = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "datakit",
			Subsystem: "input_logging",
			Name:      "pending_byte_size",
			Help:      "The size of bytes that are pending processing",
		},
		[]string{"source", "filepath"})

	metrics.MustRegister(
		openfileVec,
		parseFailVec,
		rotateVec,
		socketLogLength,
		socketLogCount,
		socketLogConnect,
		pendingBlockLength,
		pendingByteSize,
	)
}

//nolint:gochecknoinits
func init() {
	setupMetrics()
}
