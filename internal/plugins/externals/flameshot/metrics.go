// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package flameshot

import (
	"time"

	p8s "github.com/prometheus/client_golang/prometheus"
)

var (
	namespace = "flameshot"
	reg       = p8s.NewRegistry()

	uptimeGauge p8s.Gauge
	pidCount    *p8s.CounterVec
	uploadToDK  *p8s.CounterVec
)

func metricsSetup() {
	uptimeGauge = p8s.NewGauge(
		p8s.GaugeOpts{
			Namespace: namespace,
			Subsystem: "",
			Name:      "start_time_seconds",
			Help:      "The base start time of the application (Unix timestamp).",
		},
	)

	pidCount = p8s.NewCounterVec(
		p8s.CounterOpts{
			Namespace: namespace,
			Subsystem: "",
			Name:      "pid_total",
			Help:      "Number of eligible PIDs.",
		},
		[]string{"language"},
	)

	uploadToDK = p8s.NewCounterVec(
		p8s.CounterOpts{
			Namespace: namespace,
			Subsystem: "",
			Name:      "upload_to_datakit_total",
			Help:      "Count of upload to DataKit",
		},
		[]string{"service", "status_code"},
	)
}

// nolint: gochecknoinits
func init() {
	metricsSetup()

	reg.MustRegister(
		uptimeGauge,
		pidCount,
		uploadToDK,
	)
	uptimeGauge.Set(float64(time.Now().Unix()))
}
