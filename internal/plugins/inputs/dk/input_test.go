// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package dk

import (
	T "testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSetup(t *T.T) {
	t.Run("replace", func(t *T.T) {
		i := def()
		i.setup("1.2.3.4:4321")
		assert.Equal(t, "http://1.2.3.4:4321/metrics", i.url)
	})
}

func TestReadenv(t *T.T) {
	t.Run("-", func(t *T.T) {
		i := def()

		i.ReadEnv(map[string]string{
			"ENV_INPUT_DK_ENABLE_ALL_METRICS": "on",
		})

		assert.Nil(t, i.MetricFilter)

		i = def()

		i.ReadEnv(map[string]string{
			"ENV_INPUT_DK_ADD_METRICS": `["a", "b"]`,
		})

		assert.Contains(t, i.MetricFilter, "a")
		assert.Contains(t, i.MetricFilter, "b")

		i = def()

		i.ReadEnv(map[string]string{
			"ENV_INPUT_DK_ONLY_METRICS": `["a", "b"]`,
		})

		assert.Len(t, i.MetricFilter, 2)
		assert.Contains(t, i.MetricFilter, "a")
		assert.Contains(t, i.MetricFilter, "b")

		i = def()

		i.ReadEnv(map[string]string{
			"ENV_INPUT_DK_INTERVAL": "10s",
		})

		assert.Equal(t, 10*time.Second, i.Interval)

		i = def()

		i.ReadEnv(map[string]string{
			"ENV_INPUT_DK_INTERVAL": "1m",
		})

		assert.Equal(t, time.Minute, i.Interval)
	})

	t.Run("invalid-json", func(t *T.T) {
		// Test invalid JSON for ADD_METRICS
		i := def()
		i.MetricFilter = []string{"default"}
		i.ReadEnv(map[string]string{
			"ENV_INPUT_DK_ADD_METRICS": `invalid json`,
		})
		// Should keep the default filter when JSON is invalid
		assert.Equal(t, []string{"default"}, i.MetricFilter)

		// Test invalid JSON for ONLY_METRICS
		i = def()
		i.MetricFilter = []string{"default"}
		i.ReadEnv(map[string]string{
			"ENV_INPUT_DK_ONLY_METRICS": `{"key": "value"}`,
		})
		// Should keep the default filter when JSON is invalid
		assert.Equal(t, []string{"default"}, i.MetricFilter)
	})

	t.Run("invalid-interval", func(t *T.T) {
		// Test invalid interval format
		i := def()
		defaultInterval := i.Interval
		i.ReadEnv(map[string]string{
			"ENV_INPUT_DK_INTERVAL": "invalid",
		})
		// Should keep the default interval when format is invalid
		assert.Equal(t, defaultInterval, i.Interval)
	})

	t.Run("default-interval", func(t *T.T) {
		// Test default interval is 30s
		i := def()
		assert.Equal(t, 30*time.Second, i.Interval)
	})

	t.Run("k8s-env-interval", func(t *T.T) {
		// Test various interval values for k8s deployment
		testCases := []struct {
			name     string
			value    string
			expected time.Duration
		}{
			{"10s", "10s", 10 * time.Second},
			{"30s", "30s", 30 * time.Second},
			{"1m", "1m", time.Minute},
			{"5m", "5m", 5 * time.Minute},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *T.T) {
				i := def()
				i.ReadEnv(map[string]string{
					"ENV_INPUT_DK_INTERVAL": tc.value,
				})
				assert.Equal(t, tc.expected, i.Interval)
			})
		}
	})
}
