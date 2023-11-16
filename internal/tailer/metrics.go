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
	multilineVec  *prometheus.CounterVec
	rotateVec     *prometheus.CounterVec
	forceFlushVec *prometheus.CounterVec
	parseFailVec  *prometheus.CounterVec
	openfileVec   *prometheus.GaugeVec
)

func setupMetrics() {
	multilineVec = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "datakit",
			Subsystem: "tailer",
			Name:      "collect_multiline_state_total",
			Help:      "Tailer multiline state total",
		},
		[]string{
			"source",
			"filepath",
			"multilinestate",
		},
	)

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

	forceFlushVec = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "datakit",
			Subsystem: "tailer",
			Name:      "buffer_force_flush_total",
			Help:      "Tailer force flush total",
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

	metrics.MustRegister(
		multilineVec,
		openfileVec,
	)
}
