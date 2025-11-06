// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package server is DCA's HTTP server
package server

import (
	"time"

	"github.com/GuanceCloud/cliutils/metrics"

	"github.com/prometheus/client_golang/prometheus"
)

var startTime = time.Now()

var (
	datakitTotalGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "datakit_number",
		Help: "The total number of datakits.",
	}, []string{"host_name", "os"})

	uptimeGauge = prometheus.NewGaugeFunc(prometheus.GaugeOpts{
		Name: "uptime_seconds",
		Help: "The uptime of the server.",
	}, func() float64 {
		return time.Since(startTime).Seconds()
	})

	apiElapsedVec = prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Name: "api_elapsed_seconds",
		Help: "API request cost",

		Objectives: map[float64]float64{
			0.5:  0.05,
			0.9:  0.01,
			0.99: 0.001,
		},

		MaxAge:     prometheus.DefMaxAge,
		AgeBuckets: prometheus.DefAgeBuckets,
	}, []string{
		"api",
		"method",
		"status",
	})

	websocketElapsedVec = prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Name: "websocket_elapsed_seconds",
		Help: "Websocket request cost",

		Objectives: map[float64]float64{
			0.5:  0.05,
			0.9:  0.01,
			0.99: 0.001,
		},

		MaxAge:     prometheus.DefMaxAge,
		AgeBuckets: prometheus.DefAgeBuckets,
	}, []string{
		"host_name",
		"action",
		"code",
	})
)

//nolint:gochecknoinits
func init() {
	metrics.MustRegister([]prometheus.Collector{
		datakitTotalGauge,
		uptimeGauge,
		apiElapsedVec,
		websocketElapsedVec,
	}...)
}
