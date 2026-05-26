// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package gitlab

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func getMeasurementNameFromMetric(metricName string) string {
	measurementName, _ := getMeasurementAndFieldNameFromMetric(metricName)
	return measurementName
}

func TestMeasurementNameExamples(t *testing.T) {
	// Test examples from the promtheus.data file
	examples := []struct {
		metricName          string
		expectedMeasurement string
		description         string
	}{
		// From promtheus.data
		{"gitlab_banzai_cacheless_render_real_duration_seconds_bucket", "gitlab", "gitlab_ prefix histogram bucket"},
		{"gitlab_banzai_cacheless_render_real_duration_seconds_count", "gitlab", "gitlab_ prefix histogram count"},
		{"gitlab_banzai_cacheless_render_real_duration_seconds_sum", "gitlab", "gitlab_ prefix histogram sum"},
		{"gitlab_cache_misses_total", "gitlab", "gitlab_ prefix counter"},
		{"gitlab_cache_operation_duration_seconds_bucket", "gitlab", "gitlab_ prefix histogram bucket"},
		{"gitlab_cache_operation_duration_seconds_count", "gitlab", "gitlab_ prefix histogram count"},
		{"gitlab_cache_operation_duration_seconds_sum", "gitlab", "gitlab_ prefix histogram sum"},
		{"gitlab_cache_operations_total", "gitlab", "gitlab_ prefix counter"},
		{"gitlab_database_connection_pool_busy", "gitlab", "gitlab_ prefix gauge"},

		// Examples of http_ metrics (though not in promtheus.data)
		{"http_request_duration_seconds_bucket", "gitlab_http", "http_ prefix histogram bucket"},
		{"http_request_duration_seconds_count", "gitlab_http", "http_ prefix histogram count"},
		{"http_request_duration_seconds_sum", "gitlab_http", "http_ prefix histogram sum"},
		{"http_requests_total", "gitlab_http", "http_ prefix counter"},

		// Examples of other metrics
		{"ruby_sampler_duration_seconds_total", "gitlab_base", "ruby_ prefix counter"},
		{"process_cpu_seconds_total", "gitlab_base", "process_ prefix counter"},
		{"deployments", "gitlab_base", "no prefix metric"},
		{"go_goroutines", "gitlab_base", "go_ prefix gauge"},
		{"redis_operations_total", "gitlab_base", "redis_ prefix counter"},
	}

	for _, example := range examples {
		t.Run(fmt.Sprintf("%s -> %s", example.metricName, example.expectedMeasurement), func(t *testing.T) {
			measurementName := getMeasurementNameFromMetric(example.metricName)
			assert.Equal(t, example.expectedMeasurement, measurementName,
				"Metric %s should map to measurement %s: %s",
				example.metricName, example.expectedMeasurement, example.description)
		})
	}
}

func TestEdgeCases(t *testing.T) {
	testCases := []struct {
		metricName          string
		expectedMeasurement string
		description         string
	}{
		{"gitlab", "gitlab_base", "exact match 'gitlab'"},
		{"gitlab_", "gitlab", "just 'gitlab_'"},
		{"http", "gitlab_base", "exact match 'http'"},
		{"http_", "gitlab_http", "just 'http_'"},
		{"", "gitlab_base", "empty string"},
		{"_gitlab_cache", "gitlab_base", "underscore prefix"},
		{"gitlabhttp", "gitlab_base", "no underscore"},
		{"gitlab_http_mixed", "gitlab", "gitlab_ prefix even with http in name"},
		{"http_gitlab_mixed", "gitlab_http", "http_ prefix even with gitlab in name"},
	}

	for _, tc := range testCases {
		t.Run(tc.metricName, func(t *testing.T) {
			measurementName := getMeasurementNameFromMetric(tc.metricName)
			assert.Equal(t, tc.expectedMeasurement, measurementName)
		})
	}
}
