// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package dialtesting

import (
	dt "github.com/GuanceCloud/cliutils/dialtesting"
	"github.com/GuanceCloud/cliutils/metrics"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	taskGauge                   *prometheus.GaugeVec
	taskDatawaySendFailedGauge  *prometheus.GaugeVec
	taskPullCostSummary         *prometheus.SummaryVec
	taskSynchronizedCounter     *prometheus.CounterVec
	taskCheckCostSummary        *prometheus.SummaryVec
	taskRunCostSummary          *prometheus.SummaryVec
	taskInvalidCounter          *prometheus.CounterVec
	taskExecTimeIntervalSummary *prometheus.SummaryVec
	taskMaxICMPConcurrency      prometheus.Gauge
	taskICMPConcurrency         prometheus.GaugeFunc

	workerJobChanGauge         *prometheus.GaugeVec
	workerJobGauge             prometheus.Gauge
	workerCachePointsGauge     *prometheus.GaugeVec
	workerCacheDropPointsGauge *prometheus.GaugeVec
	workerSendPointsGauge      *prometheus.GaugeVec
	workerSendCost             *prometheus.SummaryVec
)

func metricsSetup() {
	taskGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "datakit",
			Subsystem: "dialtesting",
			Name:      "task_number",
			Help:      "The number of tasks",
		},
		[]string{"region", "protocol"},
	)

	taskDatawaySendFailedGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "datakit",
			Subsystem: "dialtesting",
			Name:      "dataway_send_failed_number",
			Help:      "The number of failed sending",
		},
		[]string{"region", "protocol"},
	)

	taskPullCostSummary = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace: "datakit",
			Subsystem: "dialtesting",
			Name:      "pull_cost_seconds",
			Help:      "Time cost to pull tasks",

			Objectives: map[float64]float64{
				0.5:  0.05,
				0.9:  0.01,
				0.99: 0.001,
			},
		},
		[]string{"region", "is_first"},
	)

	taskSynchronizedCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "datakit",
			Subsystem: "dialtesting",
			Name:      "task_synchronized_total",
			Help:      "Task synchronized number",
		},
		[]string{"region", "protocol"},
	)

	taskInvalidCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "datakit",
			Subsystem: "dialtesting",
			Name:      "task_invalid_total",
			Help:      "Invalid task number",
		},
		[]string{"region", "protocol", "fail_reason"},
	)

	taskCheckCostSummary = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace: "datakit",
			Subsystem: "dialtesting",
			Name:      "task_check_cost_seconds",
			Help:      "Task check time",

			Objectives: map[float64]float64{
				0.5:  0.05,
				0.9:  0.01,
				0.99: 0.001,
			},
		},
		[]string{"region", "protocol", "status"},
	)

	taskRunCostSummary = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace: "datakit",
			Subsystem: "dialtesting",
			Name:      "task_run_cost_seconds",
			Help:      "Task run time",

			Objectives: map[float64]float64{
				0.5:  0.05,
				0.9:  0.01,
				0.99: 0.001,
			},
		},
		[]string{"region", "protocol"},
	)
	taskExecTimeIntervalSummary = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace: "datakit",
			Subsystem: "dialtesting",
			Name:      "task_exec_time_interval_seconds",
			Help:      "Task execution time interval",

			Objectives: map[float64]float64{
				0.5:  0.05,
				0.9:  0.01,
				0.99: 0.001,
			},
		},
		[]string{"region", "protocol"},
	)

	taskMaxICMPConcurrency = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "datakit",
			Subsystem: "dialtesting",
			Name:      "task_max_icmp_concurrency",
			Help:      "The max number of icmp packets sent at one time",
		},
	)

	taskICMPConcurrency = prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Namespace: "datakit",
			Subsystem: "dialtesting",
			Name:      "task_icmp_concurrency",
			Help:      "The current number of icmp packets sending",
		},
		func() float64 {
			return float64(len(dt.ICMPConcurrentCh))
		},
	)

	workerJobChanGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "datakit",
			Subsystem: "dialtesting",
			Name:      "worker_job_chan_number",
			Help:      "The number of the channel for the jobs",
		},
		[]string{"type"},
	)

	workerJobGauge = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "datakit",
			Subsystem: "dialtesting",
			Name:      "worker_job_number",
			Help:      "The number of the jobs to send data in parallel",
		},
	)

	workerCachePointsGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "datakit",
			Subsystem: "dialtesting",
			Name:      "worker_cached_points_number",
			Help:      "The number of cached points",
		},
		[]string{"region", "protocol"},
	)

	workerCacheDropPointsGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "datakit",
			Subsystem: "dialtesting",
			Name:      "worker_dropped_points_number",
			Help:      "The number of dropped points",
		},
		[]string{"region", "protocol"},
	)

	workerSendPointsGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "datakit",
			Subsystem: "dialtesting",
			Name:      "worker_send_points_number",
			Help:      "The number of the points which have been sent",
		},
		[]string{"region", "protocol", "status"},
	)

	workerSendCost = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace: "datakit",
			Subsystem: "dialtesting",
			Name:      "worker_send_cost_seconds",
			Help:      "Time cost to send points",

			Objectives: map[float64]float64{
				0.5:  0.05,
				0.9:  0.01,
				0.99: 0.001,
			},
		},
		[]string{"region", "protocol"},
	)
}

//nolint:gochecknoinits
func init() {
	metricsSetup()
	metrics.MustRegister([]prometheus.Collector{
		taskGauge,
		taskDatawaySendFailedGauge,
		taskSynchronizedCounter,
		taskPullCostSummary,
		taskCheckCostSummary,
		taskRunCostSummary,
		taskExecTimeIntervalSummary,
		taskInvalidCounter,
		workerCachePointsGauge,
		workerCacheDropPointsGauge,
		workerJobChanGauge,
		workerJobGauge,
		workerSendPointsGauge,
		workerSendCost,
		taskMaxICMPConcurrency,
		taskICMPConcurrency,
	}...)
}
