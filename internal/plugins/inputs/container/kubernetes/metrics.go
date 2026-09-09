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
	collectCostVec                   *prometheus.SummaryVec
	collectPtsVec                    *prometheus.CounterVec
	collectResourceCostVec           *prometheus.SummaryVec
	podMetricsQueryCountVec          *prometheus.CounterVec
	podAnnotationPromVec             *prometheus.SummaryVec
	podAnnotationPromActiveTasks     prometheus.Gauge
	podAnnotationPromInflightScrapes prometheus.Gauge
	podAnnotationPromScrapesTotal    *prometheus.CounterVec
	objectChangeCountVec             *prometheus.CounterVec
	informerStartsVec                *prometheus.CounterVec
	informerUnsyncedVec              *prometheus.GaugeVec
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

	podAnnotationPromActiveTasks = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "datakit",
			Subsystem: "input_container",
			Name:      "kubernetes_pod_annotation_prom_active_tasks",
			Help:      "Current number of active scrape tasks for the legacy datakit/prom.instances Kubernetes Pod annotation; excludes the kubernetesPrometheus collector.",
		},
	)

	podAnnotationPromInflightScrapes = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "datakit",
			Subsystem: "input_container",
			Name:      "kubernetes_pod_annotation_prom_inflight_scrapes",
			Help:      "Current number of in-flight HTTP scrapes for the legacy datakit/prom.instances Kubernetes Pod annotation; excludes the kubernetesPrometheus collector.",
		},
	)

	podAnnotationPromScrapesTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "datakit",
			Subsystem: "input_container",
			Name:      "kubernetes_pod_annotation_prom_scrapes_total",
			Help:      "Total number of completed HTTP scrapes for the legacy datakit/prom.instances Kubernetes Pod annotation, partitioned by result; excludes the kubernetesPrometheus collector.",
		},
		[]string{"result"},
	)
	for _, result := range []string{"success", "error", "timeout", "canceled"} {
		podAnnotationPromScrapesTotal.WithLabelValues(result).Add(0)
	}
	podAnnotationPromActiveTasks.Set(0)
	podAnnotationPromInflightScrapes.Set(0)

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

	informerStartsVec = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "datakit", Subsystem: "input_container", Name: "kubernetes_informer_starts_total",
		Help: "Number of informer starts by scope; excludes borrowed runtime Pod informers.",
	}, []string{"scope"})
	informerUnsyncedVec = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "datakit", Subsystem: "input_container", Name: "kubernetes_informer_unsynced",
		Help: "Number of informers still awaiting initial synchronization in the active scope; zero after retirement.",
	}, []string{"scope"})

	metrics.MustRegister(
		informerStartsVec,
		informerUnsyncedVec,
		collectCostVec,
		collectResourceCostVec,
		collectPtsVec,
		podMetricsQueryCountVec,
		podAnnotationPromVec,
		podAnnotationPromActiveTasks,
		podAnnotationPromInflightScrapes,
		podAnnotationPromScrapesTotal,
		objectChangeCountVec,
	)
}
