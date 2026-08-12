// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package httpapi

import (
	"github.com/GuanceCloud/cliutils/metrics"
	p8s "github.com/prometheus/client_golang/prometheus"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
)

var (
	apiCountVec *p8s.CounterVec

	apiElapsedVec,
	apiReqSizeVec *p8s.SummaryVec

	apiGlobalTagsUpdatedVec *p8s.GaugeVec

	dialtestingDebugRunsTotal       *p8s.CounterVec
	dialtestingDebugRunsActive      *p8s.GaugeVec
	dialtestingDebugPreparingSize   p8s.Gauge
	dialtestingDebugQueueSize       p8s.Gauge
	dialtestingDebugDurationSeconds *p8s.HistogramVec
	dialtestingDebugResultExpired   p8s.Counter
)

func metricsSetup() {
	apiElapsedVec = p8s.NewSummaryVec(
		p8s.SummaryOpts{
			Namespace: "datakit",
			Subsystem: "http",
			Name:      "api_elapsed_seconds",
			Help:      "API request cost",

			Objectives: datakit.P8sStandardObjectives,

			MaxAge:     p8s.DefMaxAge,
			AgeBuckets: p8s.DefAgeBuckets,
		},
		[]string{
			"api",
			"method",
			"status",
		},
	)

	apiReqSizeVec = p8s.NewSummaryVec(
		p8s.SummaryOpts{
			Namespace: "datakit",
			Subsystem: "http",
			Name:      "api_req_size_bytes",
			Help:      "API request body size",

			Objectives: datakit.P8sStandardObjectives,
		},
		[]string{
			"api",
			"method",
			"status",
		},
	)

	apiCountVec = p8s.NewCounterVec(
		p8s.CounterOpts{
			Namespace: "datakit",
			Subsystem: "http",
			Name:      "api_total",
			Help:      "API request counter",
		},
		[]string{
			"api",
			"method",
			"status",
		},
	)

	apiGlobalTagsUpdatedVec = p8s.NewGaugeVec(
		p8s.GaugeOpts{
			Namespace: "datakit",
			Subsystem: "http",
			Name:      "api_global_tags_last_updated",
			Help:      "Global tag updated timestamp, in second",
		},
		[]string{
			"api",
			"method",
			"status",
		},
	)

	dialtestingDebugRunsTotal = p8s.NewCounterVec(
		p8s.CounterOpts{Name: "dialing_debug_runs_total", Help: "Dialtesting debug run state transitions."},
		[]string{"type", "status"},
	)
	dialtestingDebugRunsActive = p8s.NewGaugeVec(
		p8s.GaugeOpts{Name: "dialing_debug_runs_active", Help: "Currently executing dialtesting debug runs."},
		[]string{"type"},
	)
	dialtestingDebugPreparingSize = p8s.NewGauge(
		p8s.GaugeOpts{Name: "dialing_debug_preparing_size", Help: "Dialtesting debug runs currently being prepared."},
	)
	dialtestingDebugQueueSize = p8s.NewGauge(
		p8s.GaugeOpts{Name: "dialing_debug_queue_size", Help: "Pending dialtesting debug runs."},
	)
	dialtestingDebugDurationSeconds = p8s.NewHistogramVec(
		p8s.HistogramOpts{Name: "dialing_debug_duration_seconds", Help: "Dialtesting debug execution duration."},
		[]string{"type"},
	)
	dialtestingDebugResultExpired = p8s.NewCounter(
		p8s.CounterOpts{Name: "dialing_debug_result_expired_total", Help: "Expired dialtesting debug results."},
	)

	metrics.MustRegister(
		apiElapsedVec,
		apiReqSizeVec,
		apiCountVec,
		apiGlobalTagsUpdatedVec,
		dialtestingDebugRunsTotal,
		dialtestingDebugRunsActive,
		dialtestingDebugPreparingSize,
		dialtestingDebugQueueSize,
		dialtestingDebugDurationSeconds,
		dialtestingDebugResultExpired,
	)
}

//nolint:gochecknoinits
func init() {
	metricsSetup()
}
