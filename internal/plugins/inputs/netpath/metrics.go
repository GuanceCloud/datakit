// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package netpath

import (
	"strings"
	"sync"

	"github.com/GuanceCloud/cliutils/metrics"
	"github.com/prometheus/client_golang/prometheus"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
)

var (
	candidateCounter *prometheus.CounterVec
	dropCounter      *prometheus.CounterVec
	queueGauge       *prometheus.GaugeVec
	storeGauge       prometheus.Gauge
	storeBytesGauge  prometheus.Gauge
	taskCounter      *prometheus.CounterVec
	taskRunCost      *prometheus.SummaryVec
	feedErrorCounter prometheus.Counter
	observeMu        sync.Mutex
)

func metricsSetup() {
	candidateCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "datakit",
			Subsystem: inputName,
			Name:      "candidate_total",
			Help:      "The number of netpath dynamic candidates handled by status and reason.",
		},
		[]string{"status", "reason"},
	)

	dropCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "datakit",
			Subsystem: inputName,
			Name:      "drop_total",
			Help:      "The number of netpath tasks dropped by stage and reason.",
		},
		[]string{"stage", "reason"},
	)

	queueGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "datakit",
			Subsystem: inputName,
			Name:      "queue_size",
			Help:      "The current number of pending netpath tasks in internal queues.",
		},
		[]string{"queue"},
	)

	storeGauge = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "datakit",
			Subsystem: inputName,
			Name:      "store_size",
			Help:      "The current number of deduplicated netpath task contexts.",
		},
	)
	storeBytesGauge = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "datakit",
			Subsystem: inputName,
			Name:      "store_bytes",
			Help:      "The estimated retained bytes of stored and running dynamic netpath task contexts.",
		},
	)

	taskCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "datakit",
			Subsystem: inputName,
			Name:      "task_total",
			Help:      "The number of netpath probe tasks run by source, protocol and status.",
		},
		[]string{"source", "protocol", "status"},
	)

	taskRunCost = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace:  "datakit",
			Subsystem:  inputName,
			Name:       "task_run_cost_seconds",
			Help:       "Time cost to run netpath probe tasks.",
			Objectives: datakit.P8sStandardObjectives,
		},
		[]string{"source", "protocol", "status"},
	)

	feedErrorCounter = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: "datakit",
			Subsystem: inputName,
			Name:      "feed_error_total",
			Help:      "The number of netpath points failed to feed.",
		},
	)
}

func incCandidate(status, reason string) {
	candidateCounter.WithLabelValues(status, reason).Inc()
}

func incDrop(stage, reason string) {
	dropCounter.WithLabelValues(stage, reason).Inc()
}

func observeQueues(s *scheduler) {
	if s == nil {
		return
	}
	observeMu.Lock()
	defer observeMu.Unlock()
	queueGauge.WithLabelValues("input").Set(float64(len(s.inputCh)))
	queueGauge.WithLabelValues("process").Set(float64(len(s.processCh) + len(s.localCh)))
}

func observeStore(store *taskStore) {
	if store == nil {
		return
	}
	observeMu.Lock()
	defer observeMu.Unlock()
	n, retainedBytes := store.stats()
	storeGauge.Set(float64(n))
	storeBytesGauge.Set(float64(retainedBytes))
}

func observeRun(res probeResult) {
	status := probeStatus(res)
	taskCounter.WithLabelValues(res.task.Source, res.task.Protocol, status).Inc()
	taskRunCost.WithLabelValues(res.task.Source, res.task.Protocol, status).Observe(res.duration.Seconds())
}

func incFeedError() {
	feedErrorCounter.Inc()
}

func probeStatus(res probeResult) string {
	if res.err != nil {
		return "fail"
	}
	if status := res.tags["traceroute_status"]; status != "" {
		return strings.ToLower(status)
	}
	return "unknown"
}

func init() { //nolint:gochecknoinits
	metricsSetup()
}

func registerMetrics() {
	metrics.MustRegister([]prometheus.Collector{
		candidateCounter,
		dropCounter,
		queueGauge,
		storeGauge,
		storeBytesGauge,
		taskCounter,
		taskRunCost,
		feedErrorCounter,
	}...)
}
