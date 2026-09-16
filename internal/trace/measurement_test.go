// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package trace

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

func TestTracingMetricMeasurementQPSMetadata(t *testing.T) {
	ddInfo := (TracingMetricMeasurement{Source: "ddtrace", Name: "DDTrace", EnableQPS: true}).Info()
	qps, ok := ddInfo.Fields["qps"]
	require.True(t, ok)
	field, ok := qps.(*inputs.FieldInfo)
	require.True(t, ok)
	assert.Equal(t, inputs.Gauge, field.Type)
	assert.Equal(t, inputs.Int, field.DataType)
	assert.Equal(t, inputs.RequestsPerSec, field.Unit)
	_, ok = ddInfo.Tags["qps_overflow"]
	assert.True(t, ok)

	otelInfo := (TracingMetricMeasurement{Source: "opentelemetry", Name: "OpenTelemetry"}).Info()
	_, ok = otelInfo.Fields["qps"]
	assert.False(t, ok)
	_, ok = otelInfo.Tags["qps_overflow"]
	assert.False(t, ok)
}
