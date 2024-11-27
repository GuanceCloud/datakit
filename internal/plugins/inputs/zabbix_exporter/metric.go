// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package exporter collect RealTime data.
package exporter

import (
	"github.com/GuanceCloud/cliutils/metrics"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	ZabbixCollectFiles   *prometheus.CounterVec
	ZabbixCollectMetrics *prometheus.CounterVec
	RequestAPIVec        *prometheus.SummaryVec
)

func metricsSetup() {
	ZabbixCollectMetrics = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "datakit",
			Subsystem: "input_zabbix_exporter",
			Name:      "collect_metric_total",
			Help:      "exporter metric count number from start",
		},
		[]string{
			"object",
		},
	)

	ZabbixCollectFiles = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "datakit",
			Subsystem: "input_zabbix_exporter",
			Name:      "collect_file_total",
			Help:      "The files number of exporter file",
		},
		[]string{
			"object",
		},
	)

	RequestAPIVec = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace: "datakit",
			Subsystem: "input_zabbix_exporter",
			Name:      "request_api",
			Help:      "The time of success or failed API requests",
			Objectives: map[float64]float64{
				0.5:  0.05,
				0.9:  0.01,
				0.99: 0.001,
			},
		}, []string{
			"status",
		})
}

func init() { //nolint:gochecknoinits
	metricsSetup()
	metrics.MustRegister(ZabbixCollectMetrics, ZabbixCollectFiles, RequestAPIVec)
}
