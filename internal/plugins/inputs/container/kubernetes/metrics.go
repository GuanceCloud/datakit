// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package kubernetes

import (
	"github.com/GuanceCloud/cliutils/metrics"
	"github.com/prometheus/client_golang/prometheus"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
)

var (
	collectCostVec          *prometheus.SummaryVec
	collectPtsVec           *prometheus.CounterVec
	collectResourceCostVec  *prometheus.SummaryVec
	podMetricsQueryCountVec *prometheus.CounterVec
	podAnnotationPromVec    *prometheus.SummaryVec
	objectChangeCountVec    *prometheus.CounterVec
)

func setupMetrics() {
	collectCostVec = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace: "datakit",
			Subsystem: "input_container",
			Name:      "kubernetes_collect_cost_seconds",
			Help:      "Total time (in seconds) spent collecting metrics from Kubernetes",

			Objectives: datakit.P8sStandardObjectives,
		},
		[]string{
			"category",
		},
	)

	collectPtsVec = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "datakit",
			Subsystem: "input_container",
			Name:      "kubernetes_collect_pts_total",
			Help:      "Total number of points collected from Kubernetes resources",
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
			Help:      "Total time (in seconds) spent collecting resource metrics from Kubernetes",

			Objectives: datakit.P8sStandardObjectives,
		},
		[]string{
			"category",
			"name",
		},
	)

	podMetricsQueryCountVec = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "datakit",
			Subsystem: "input_container",
			Name:      "kubernetes_pod_metrics_query_total",
			Help:      "Total number of metric queries made to Kubernetes Pods",
		},
		[]string{
			"target",
		},
	)

	podAnnotationPromVec = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace: "datakit",
			Subsystem: "input_container",
			Name:      "kubernetes_pod_annotation_prom_count",
			Help:      "The number of Prometheus-related annotations found in Kubernetes Pods",

			Objectives: datakit.P8sStandardObjectives,
		},
		[]string{
			"name",
		},
	)

	objectChangeCountVec = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "datakit",
			Subsystem: "input_container",
			Name:      "kubernetes_resources_change_total",
			Help:      "Total number of objects changed for Kubernetes resources.",
		},
		[]string{
			"resource",
			"type",
		},
	)

	metrics.MustRegister(
		collectCostVec,
		collectResourceCostVec,
		collectPtsVec,
		podMetricsQueryCountVec,
		podAnnotationPromVec,
		objectChangeCountVec,
	)
}
