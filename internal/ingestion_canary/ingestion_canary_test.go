// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package ingestioncanary

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCanary_New(t *testing.T) {
	c := New(WithName("test"), WithTestType("collect"))
	assert.NotNil(t, c)
	assert.Equal(t, "test", c.Name)
	assert.Equal(t, "collect", c.TestType)
}

func TestCanary_GenerateDataPoints(t *testing.T) {
	c := New(WithName("ingestion_canary"), WithTestType("collect"))
	ts := time.Now().Truncate(time.Millisecond)
	round := int64(1)

	metricPt := c.Metric(ts, round, nil)
	loggingPt := c.Logging(ts, round, nil)
	tracingPt := c.Tracing(ts, round, nil)

	assert.NotNil(t, metricPt)
	assert.Equal(t, "ingestion_canary", metricPt.Name())
	assert.Equal(t, round, metricPt.InfluxFields()["round"])

	assert.NotNil(t, loggingPt)
	assert.Equal(t, "ingestion_canary", loggingPt.Name())
	assert.Contains(t, loggingPt.InfluxFields()["message"], "synthetic freshness probe")

	assert.NotNil(t, tracingPt)
	assert.Equal(t, "ingestion_canary", tracingPt.Name())
	assert.Equal(t, "0", tracingPt.InfluxFields()["parent_id"])
}

func TestCanary_BuildDQL(t *testing.T) {
	c := New(WithName("ingestion_canary"), WithTestType("collect"))

	dql := c.BuildDQL(MetricCategory, "")
	assert.Contains(t, dql, "M::ingestion_canary")
	assert.Contains(t, dql, "test_type = \"collect\"")

	dql = c.BuildDQL(LoggingCategory, "")
	assert.Contains(t, dql, "L('default')::ingestion_canary")

	dql = c.BuildDQL(LoggingCategory, "default")
	assert.Contains(t, dql, "L('default')::ingestion_canary")

	dql = c.BuildDQL(TracingCategory, "")
	assert.Contains(t, dql, "T::ingestion_canary")
}

func TestParseLast(t *testing.T) {
	resp := &DQLQueryResponse{
		Content: []DQLQueryContent{
			{
				Series: []DQLSeries{
					{
						Columns: []string{"time", "last(round)"},
						Values: [][]interface{}{
							{int64(1234567890), int64(42)},
						},
					},
				},
			},
		},
	}

	result, ok := ParseLast(resp)
	assert.True(t, ok)
	assert.NotNil(t, result)
	assert.Equal(t, int64(1234567890), result.TimeMs)
	assert.Equal(t, int64(42), result.Round)
}

func TestParseLast_EmptyResponse(t *testing.T) {
	result, ok := ParseLast(nil)
	assert.False(t, ok)
	assert.Nil(t, result)
}

func TestParseInt64(t *testing.T) {
	val, err := ParseInt64(int64(123))
	assert.NoError(t, err)
	assert.Equal(t, int64(123), val)

	val, err = ParseInt64(json.Number("456"))
	assert.NoError(t, err)
	assert.Equal(t, int64(456), val)

	_, err = ParseInt64("invalid")
	assert.Error(t, err)
}
