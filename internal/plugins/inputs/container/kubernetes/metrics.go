// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package kubernetes

import (
	"github.com/GuanceCloud/cliutils/metrics"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	fetchErrorVec           *prometheus.GaugeVec
	collectCostVec          *prometheus.SummaryVec
	collectPtsVec           *prometheus.CounterVec
	collectResourcePtsVec   *prometheus.SummaryVec
	podMetricsQueryCountVec *prometheus.CounterVec
)

func setupMetrics() {
	fetchErrorVec = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "datakit",
			Subsystem: "kubernetes",
			Name:      "fetch_error",
			Help:      "Kubernetes resource fetch error",
		},
		[]string{
			"namespace",
			"resource",
			"error",
		},
	)

	collectCostVec = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace: "datakit",
			Subsystem: "kubernetes",
			Name:      "collect_cost_seconds",
			Help:      "Kubernetes resource collect cost",
		},
		[]string{
			"category",
		},
	)

	collectPtsVec = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "datakit",
			Subsystem: "kubernetes",
			Name:      "collect_pts_total",
			Help:      "Kubernetes collect point total",
		},
		[]string{
			"category",
		},
	)

	collectResourcePtsVec = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace: "datakit",
			Subsystem: "kubernetes",
			Name:      "collect_resource_pts_num",
			Help:      "Kubernetes resource collect point count",
		},
		[]string{
			"category",
			"kind",
			"fieldselector",
		},
	)

	podMetricsQueryCountVec = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "datakit",
			Subsystem: "kubernetes",
			Name:      "pod_metrics_query_total",
			Help:      "Kubernetes query pod metrics count",
		},
		[]string{
			"target",
		},
	)

	metrics.MustRegister(
		fetchErrorVec,
		collectCostVec,
		collectPtsVec,
		collectResourcePtsVec,
		podMetricsQueryCountVec,
	)
}
