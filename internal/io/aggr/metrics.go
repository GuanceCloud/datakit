// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package aggr

import (
	"time"

	"github.com/GuanceCloud/cliutils/metrics"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	aggrSendSuccess *prometheus.CounterVec
	aggrSendFailed  *prometheus.CounterVec
	aggrSendPoints  *prometheus.CounterVec
	aggrLostPoints  *prometheus.CounterVec
	aggrSendLatency *prometheus.SummaryVec
)

func init() { //nolint:gochecknoinits
	aggrSendSuccess = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "datakit",
			Subsystem: "io",
			Name:      "aggr_send_success_total",
			Help:      "Aggregation data send success count",
		},
		[]string{"type", "category"},
	)

	aggrSendFailed = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "datakit",
			Subsystem: "io",
			Name:      "aggr_send_failed_total",
			Help:      "Aggregation data send failed count",
		},
		[]string{"type", "category", "reason"},
	)

	aggrSendPoints = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "datakit",
			Subsystem: "io",
			Name:      "aggr_send_points_total",
			Help:      "Aggregation data points sent count",
		},
		[]string{"type", "category"},
	)

	aggrLostPoints = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "datakit",
			Subsystem: "io",
			Name:      "aggr_lost_points_total",
			Help:      "Aggregation data points lost count",
		},
		[]string{"type", "category", "reason"},
	)

	aggrSendLatency = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace:  "datakit",
			Subsystem:  "io",
			Name:       "aggr_send_latency_seconds",
			Help:       "Aggregation data send latency in seconds",
			Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
		},
		[]string{"type", "category"},
	)

	// Register metrics
	metrics.MustRegister(
		aggrSendSuccess,
		aggrSendFailed,
		aggrSendPoints,
		aggrLostPoints,
		aggrSendLatency,
	)
}

func recordSendSuccess(sendType, category string) {
	if aggrSendSuccess != nil {
		aggrSendSuccess.WithLabelValues(sendType, category).Inc()
	}
}

func recordSendFailed(sendType, category, reason string) {
	if aggrSendFailed != nil {
		aggrSendFailed.WithLabelValues(sendType, category, reason).Inc()
	}
}

func recordSendPoints(sendType, category string, points int) {
	if aggrSendPoints != nil {
		aggrSendPoints.WithLabelValues(sendType, category).Add(float64(points))
	}
}

func recordLostPoints(sendType, category, reason string, points int) {
	if aggrLostPoints != nil {
		aggrLostPoints.WithLabelValues(sendType, category, reason).Add(float64(points))
	}
}

func recordSendLatency(sendType, category string, latency time.Duration) {
	if aggrSendLatency != nil {
		aggrSendLatency.WithLabelValues(sendType, category).Observe(latency.Seconds())
	}
}
