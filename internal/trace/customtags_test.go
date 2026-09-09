// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2026-present Guance, Inc.

package trace

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCustomTagsDDTraceMetricFields(t *testing.T) {
	publicTags := map[string]string{"builtin.metric": "builtin_metric"}
	customTags := NewCustomTags(
		[]string{"custom.metric", `reg:^business\.`},
		publicTags,
	)

	metrics, kvs := customTags.DDTraceMetricFields(map[string]float64{
		"custom.metric":    1,
		"business.latency": 2,
		"builtin.metric":   3,
		"other.metric":     4,
	}, nil)

	customMetric := kvs.Get("custom_metric")
	require.NotNil(t, customMetric)
	assert.Equal(t, float64(1), customMetric.Raw())
	businessLatency := kvs.Get("business_latency")
	require.NotNil(t, businessLatency)
	assert.Equal(t, float64(2), businessLatency.Raw())
	assert.Equal(t, map[string]float64{
		"builtin.metric": 3,
		"other.metric":   4,
	}, metrics)
}
