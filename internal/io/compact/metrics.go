// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package compact

import (
	"github.com/GuanceCloud/cliutils/metrics"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	BuildBodyCostVec = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace: "datakit",
			Subsystem: "io",
			Name:      "build_body_cost_seconds",
			Help:      "Build point HTTP body cost",

			Objectives: map[float64]float64{
				0.5:  0.05,
				0.9:  0.01,
				0.99: 0.001,
			},
		},
		[]string{"category", "encoding", "stage"},
	)

	buildBodyPointsVec = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace: "datakit",
			Subsystem: "io",
			Name:      "build_body_points",
			Help:      "Point count for single compact",

			Objectives: map[float64]float64{
				0.5:  0.05,
				0.9:  0.01,
				0.99: 0.001,
			},
		},
		[]string{"category", "encoding"},
	)

	buildBodyBatchPointsVec = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace: "datakit",
			Subsystem: "io",
			Name:      "build_body_batch_points",
			Help:      "Batch HTTP body points",

			Objectives: map[float64]float64{
				0.5:  0.05,
				0.9:  0.01,
				0.99: 0.001,
			},
		},
		[]string{"category", "encoding"},
	)

	buildBodyBatchBytesVec = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace: "datakit",
			Subsystem: "io",
			Name:      "build_body_batch_bytes",
			Help:      "Batch HTTP body size",

			Objectives: map[float64]float64{
				0.5:  0.05,
				0.9:  0.01,
				0.99: 0.001,
			},
		},
		[]string{"category", "encoding", "type"},
	)

	skippedPointVec = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "datakit",
			Subsystem: "io",
			Name:      "compact_skipped_point_total",
			Help:      "Skipped point count during encoding(Protobuf) point",
		},
		[]string{
			"category",
		},
	)

	buildBodyBatchCountVec = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace: "datakit",
			Subsystem: "io",
			Name:      "build_body_batches",
			Help:      "Batch HTTP body batches",

			Objectives: map[float64]float64{
				0.5:  0.05,
				0.9:  0.01,
				0.99: 0.001,
			},
		},
		[]string{"category", "encoding"},
	)

	bodyCounterVec = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "datakit",
			Subsystem: "io",
			Name:      "compact_body_total",
			Help:      "Compact total body",
		},
		[]string{
			"caller",
			"from",
			"op",
			"type",
		},
	)
)

func Metrics() []prometheus.Collector {
	return []prometheus.Collector{
		skippedPointVec,
		bodyCounterVec,
		BuildBodyCostVec,
		buildBodyBatchBytesVec,
		buildBodyBatchPointsVec,
		buildBodyBatchCountVec,
		buildBodyPointsVec,
	}
}

func doRegister() {
	metrics.MustRegister(
		buildBodyPointsVec,
		BuildBodyCostVec,
	)
}

func MetricsReset() {
	skippedPointVec.Reset()
	bodyCounterVec.Reset()

	BuildBodyCostVec.Reset()
	buildBodyBatchBytesVec.Reset()
	buildBodyBatchPointsVec.Reset()
	buildBodyBatchCountVec.Reset()
	buildBodyPointsVec.Reset()
}

// nolint:gochecknoinits
func init() {
	doRegister()
}
