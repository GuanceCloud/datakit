// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build windows
// +build windows

package winevent

import (
	"github.com/GuanceCloud/cliutils/metrics"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	eventCacheNumber prometheus.Gauge
	eventTimeDiff    prometheus.Summary
)

func metricsSetup() {
	eventCacheNumber = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "datakit",
			Subsystem: inputName,
			Name:      "cache_number",
			Help:      "The number of items in cache",
		},
	)

	eventTimeDiff = prometheus.NewSummary(
		prometheus.SummaryOpts{
			Namespace: "datakit",
			Subsystem: inputName,
			Name:      "event_time_diff_seconds",
			Help:      "The time difference between event time and current time",
		},
	)
}

//nolint:gochecknoinits
func init() {
	metricsSetup()
	metrics.MustRegister([]prometheus.Collector{
		eventCacheNumber,
		eventTimeDiff,
	}...)
}
