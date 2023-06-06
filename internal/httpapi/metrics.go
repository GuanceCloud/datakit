// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package httpapi

import (
	"github.com/GuanceCloud/cliutils/metrics"
	p8s "github.com/prometheus/client_golang/prometheus"
)

var (
	apiCountVec *p8s.CounterVec

	apiElapsedVec,
	apiReqSizeVec *p8s.SummaryVec
)

func metricsSetup() {
	apiElapsedVec = p8s.NewSummaryVec(
		p8s.SummaryOpts{
			Namespace: "datakit",
			Subsystem: "http",
			Name:      "api_elapsed_seconds",
			Help:      "API request cost",
			Objectives: map[float64]float64{
				0.5:  0.05,
				0.75: 0.0075,
				0.95: 0.005,
			},

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

	metrics.MustRegister(
		apiElapsedVec,
		apiReqSizeVec,
		apiCountVec,
	)
}

//nolint:gochecknoinits
func init() {
	metricsSetup()
}
