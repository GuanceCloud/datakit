// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package cache contains graphite cache metrics
package cache

import (
	"github.com/GuanceCloud/cliutils/metrics"
	"github.com/prometheus/client_golang/prometheus"
)

type CacheMetrics struct {
	CacheLength    prometheus.Gauge
	CacheGetsTotal prometheus.Counter
	CacheHitsTotal prometheus.Counter
}

func NewCacheMetrics() *CacheMetrics {
	var m CacheMetrics

	m.CacheLength = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "datakit",
			Subsystem: "input_graphite",
			Name:      "metric_mapper_cache_length",
			Help:      "The count of unique metrics currently cached.",
		},
	)
	m.CacheGetsTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: "datakit",
			Subsystem: "input_graphite",
			Name:      "metric_cache_gets_total",
			Help:      "The count of total metric cache gets.",
		},
	)
	m.CacheHitsTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: "datakit",
			Subsystem: "input_graphite",
			Name:      "metric_mapper_cache_hits_total",
			Help:      "The count of total metric cache hits.",
		},
	)

	metrics.MustRegister(
		m.CacheLength,
		m.CacheGetsTotal,
		m.CacheHitsTotal,
	)

	return &m
}
