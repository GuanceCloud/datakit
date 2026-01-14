// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package cmds

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	ingestioncanary "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/ingestion_canary"
)

func TestIngestionCanaryTool_Initialization(t *testing.T) {
	tool := &ingestionCanaryTool{
		canary: ingestioncanary.New(
			ingestioncanary.WithName("ingestion_canary"),
			ingestioncanary.WithTestType("cmd"),
		),
	}

	assert.NotNil(t, tool.canary)
	assert.Equal(t, "ingestion_canary", tool.canary.Name)
	assert.Equal(t, "cmd", tool.canary.TestType)
}

func TestIngestionCanaryTool_GenerateDataPoints(t *testing.T) {
	tool := &ingestionCanaryTool{
		canary: ingestioncanary.New(
			ingestioncanary.WithName("ingestion_canary"),
			ingestioncanary.WithTestType("cmd"),
		),
		round: 1,
	}

	ts := time.Now().Truncate(time.Millisecond)
	metricPt := tool.canary.Metric(ts, tool.round, nil)
	loggingPt := tool.canary.Logging(ts, tool.round, nil)
	tracingPt := tool.canary.Tracing(ts, tool.round, nil)

	assert.NotNil(t, metricPt)
	assert.Equal(t, "ingestion_canary", metricPt.Name())
	assert.Equal(t, int64(1), metricPt.InfluxFields()["round"])

	assert.NotNil(t, loggingPt)
	assert.Equal(t, "ingestion_canary", loggingPt.Name())

	assert.NotNil(t, tracingPt)
	assert.Equal(t, "ingestion_canary", tracingPt.Name())
}

func TestIngestionCanaryTool_QueryData_ContextCancellation(t *testing.T) {
	tool := &ingestionCanaryTool{
		canary: ingestioncanary.New(
			ingestioncanary.WithName("ingestion_canary"),
			ingestioncanary.WithTestType("cmd"),
		),
		round:        1,
		storageIndex: "default",
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	feedTime := time.Now()
	latency, err := tool.queryData(ctx, ingestioncanary.MetricCategory, feedTime)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "exited")
	assert.Equal(t, time.Duration(0), latency)
}
