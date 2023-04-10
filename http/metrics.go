// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package http

import (
	"github.com/GuanceCloud/cliutils/metrics"
	p8s "github.com/prometheus/client_golang/prometheus"
)

var (
	apiCountVec *p8s.CounterVec

	apiElapsedVec,
	apiReqSizeVec *p8s.SummaryVec

	apiElapsedHistogram *p8s.HistogramVec
)

func metricsSetup() {
	apiElapsedVec = p8s.NewSummaryVec(
		p8s.SummaryOpts{
			Namespace: "datakit",
			Subsystem: "http",
			Name:      "api_elapsed",
			Help:      "API request cost(in ms)",
		},
		[]string{
			"api",
			"method",
			"status",
		},
	)

	//nolint:promlinter
	apiElapsedHistogram = p8s.NewHistogramVec(
		p8s.HistogramOpts{
			Namespace: "datakit",
			Subsystem: "http",
			Name:      "api_elapsed_histogram",
			Help:      "API request cost(in ms) histogram",
			Buckets: []float64{
				float64(10),    // 10ms
				float64(100),   // 100ms
				float64(1000),  // 1s
				float64(5000),  // 5s
				float64(30000), // 30s
			},
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
			Name:      "api_req_size",
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
		apiElapsedHistogram,
		apiReqSizeVec,
		apiCountVec,
	)
}

//nolint:gochecknoinits
func init() {
	metricsSetup()
}
