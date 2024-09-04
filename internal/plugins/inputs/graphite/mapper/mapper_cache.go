// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package mapper contains graphite mapper cache.
package mapper

type MetricMapperCacheResult struct {
	Mapping *MetricMapping
	Matched bool
	Labels  Labels
}

// MetricMapperCache MUST be thread-safe and should be instrumented with CacheMetrics.
type MetricMapperCache interface {
	// Get a cached result
	Get(metricKey string) (interface{}, bool)
	// Add a statsd MetricMapperResult to the cache
	Add(metricKey string, result interface{}) // Add an item to the cache
	// Reset clears the cache for config reloads
	Reset()
}

func formatKey(metricString string, metricType MetricType) string {
	return string(metricType) + "." + metricString
}
