// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package graphite metrics
package graphite

import (
	p8s "github.com/prometheus/client_golang/prometheus"

	"github.com/GuanceCloud/cliutils/metrics"
)

var (
	tagParseFailures   p8s.Counter
	lastProcessed      p8s.Gauge
	sampleExpiryMetric p8s.Gauge
)

func metricsSetup() {
	tagParseFailures = p8s.NewCounter(
		p8s.CounterOpts{
			Namespace: "datakit",
			Subsystem: "input_graphite",
			Name:      "tag_parse_failures_total",
			Help:      "Total count of samples with invalid tags",
		},
	)

	lastProcessed = p8s.NewGauge(
		p8s.GaugeOpts{
			Namespace: "datakit",
			Subsystem: "input_graphite",
			Name:      "last_processed_timestamp_seconds",
			Help:      "Unix timestamp of the last processed graphite metric.",
		},
	)

	sampleExpiryMetric = p8s.NewGauge(
		p8s.GaugeOpts{
			Namespace: "datakit",
			Subsystem: "input_graphite",
			Name:      "sample_expiry_seconds",
			Help:      "How long in seconds a metric sample is valid for.",
		},
	)
}

//nolint:gochecknoinits
func init() {
	metricsSetup()
	metrics.MustRegister(
		tagParseFailures,
		lastProcessed,
		sampleExpiryMetric,
	)
}
