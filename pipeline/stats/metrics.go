// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package stats

import (
	"github.com/GuanceCloud/cliutils/metrics"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	plPtsVec,
	plErrPtsVec,
	plDropVec *prometheus.CounterVec
	plUpdateVec *prometheus.GaugeVec
	plCostVec   *prometheus.SummaryVec
)

func Metrics() []prometheus.Collector {
	return []prometheus.Collector{
		plPtsVec,
		plDropVec,
		plUpdateVec,
		plCostVec,
	}
}

// nolint:gochecknoinits
func init() {
	plPtsVec = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "datakit",
			Subsystem: "pipeline",
			Name:      "point_total",
			Help:      "Pipeline processed total points",
		},
		[]string{
			"category",
			"name",
			"namespace",
		},
	)

	plDropVec = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "datakit",
			Subsystem: "pipeline",
			Name:      "drop_point_total",
			Help:      "Pipeline total dropped points",
		},
		[]string{
			"category",
			"name",
			"namespace",
		},
	)

	plErrPtsVec = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "datakit",
			Subsystem: "pipeline",
			Name:      "error_point_total",
			Help:      "Pipeline processed total error points",
		},
		[]string{
			"category",
			"name",
			"namespace",
		},
	)

	plCostVec = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace: "datakit",
			Subsystem: "pipeline",
			Name:      "cost",
			Help:      "Pipeline total running time(ms)",
		},
		[]string{
			"category",
			"name",
			"namespace",
		},
	)

	plUpdateVec = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "datakit",
			Subsystem: "pipeline",
			Name:      "update_time",
			Help:      "Pipeline last update time(unix timestamp)",
		},
		[]string{
			"category",
			"name",
			"namespace",
		},
	)

	metrics.MustRegister(Metrics()...)
}
