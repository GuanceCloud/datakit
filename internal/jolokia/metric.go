// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package jolokia provides metrics collection for Jolokia client operations.
package jolokia

import (
	"github.com/GuanceCloud/cliutils/metrics"
	"github.com/prometheus/client_golang/prometheus"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
)

var (
	requestTotalVec   *prometheus.CounterVec
	requestLatencyVec *prometheus.SummaryVec
	requestErrorVec   *prometheus.CounterVec
)

func metricsSetup() {
	requestTotalVec = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "datakit",
			Subsystem: "jolokia",
			Name:      "request_total",
			Help:      "Total number of Jolokia requests",
		},
		[]string{"url", "input"},
	)

	requestLatencyVec = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace:  "datakit",
			Subsystem:  "jolokia",
			Name:       "request_latency_seconds",
			Help:       "Jolokia request latency in seconds",
			Objectives: datakit.P8sStandardObjectives,
		},
		[]string{"url", "input"},
	)

	requestErrorVec = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "datakit",
			Subsystem: "jolokia",
			Name:      "request_error_total",
			Help:      "Total number of Jolokia request errors",
		},
		[]string{"url", "input", "error_type"},
	)

	metrics.MustRegister(
		requestTotalVec,
		requestLatencyVec,
		requestErrorVec,
	)
}

//nolint:gochecknoinits
func init() {
	metricsSetup()
}
