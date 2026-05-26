// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package ingestioncanary

import (
	"github.com/GuanceCloud/cliutils/metrics"
	"github.com/prometheus/client_golang/prometheus"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
)

var (
	dqlQueryDurationVec *prometheus.SummaryVec
	dqlQueryTotalVec    *prometheus.CounterVec
)

func setupMetrics() {
	dqlQueryDurationVec = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace:  "datakit",
			Subsystem:  "ingestion_canary",
			Name:       "dql_query_duration_seconds",
			Help:       "DQL query duration in seconds",
			Objectives: datakit.P8sStandardObjectives,
		},
		[]string{
			"category",      // M, L, T
			"storage_index", // storage index name (for logging)
			"status",        // success, error, not_found
		},
	)

	dqlQueryTotalVec = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "datakit",
			Subsystem: "ingestion_canary",
			Name:      "dql_query_total",
			Help:      "Total number of DQL queries",
		},
		[]string{
			"category",      // M, L, T
			"storage_index", // storage index name (for logging)
			"status",        // success, error, not_found
		},
	)

	metrics.MustRegister(
		dqlQueryDurationVec,
		dqlQueryTotalVec,
	)
}

//nolint:gochecknoinits
func init() {
	setupMetrics()
}
