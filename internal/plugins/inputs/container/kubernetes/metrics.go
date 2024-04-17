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
	collectResourceCostVec  *prometheus.SummaryVec
	collectPtsVec           *prometheus.CounterVec
	podMetricsQueryCountVec *prometheus.CounterVec
)

func setupMetrics() {
	fetchErrorVec = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "datakit",
			Subsystem: "input_container",
			Name:      "kubernetes_fetch_error",
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
			Subsystem: "input_container",
			Name:      "kubernetes_collect_cost_seconds",
			Help:      "Kubernetes collect cost",

			Objectives: map[float64]float64{
				0.5:  0.05,
				0.9:  0.01,
				0.99: 0.001,
			},
		},
		[]string{
			"category",
		},
	)

	collectResourceCostVec = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace: "datakit",
			Subsystem: "input_container",
			Name:      "kubernetes_collect_resource_cost_seconds",
			Help:      "Kubernetes collect resource cost",

			Objectives: map[float64]float64{
				0.5:  0.05,
				0.9:  0.01,
				0.99: 0.001,
			},
		},
		[]string{
			"category",
			"kind",
			"fieldselector",
		},
	)

	collectPtsVec = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "datakit",
			Subsystem: "input_container",
			Name:      "kubernetes_collect_pts_total",
			Help:      "Kubernetes collect point total",
		},
		[]string{
			"category",
		},
	)

	podMetricsQueryCountVec = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "datakit",
			Subsystem: "input_container",
			Name:      "kubernetes_pod_metrics_query_total",
			Help:      "Kubernetes query pod metrics count",
		},
		[]string{
			"target",
		},
	)

	metrics.MustRegister(
		fetchErrorVec,
		collectCostVec,
		collectResourceCostVec,
		collectPtsVec,
		podMetricsQueryCountVec,
	)
}
