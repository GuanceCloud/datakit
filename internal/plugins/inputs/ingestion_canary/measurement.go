// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//nolint:lll,goconst,funlen // Measurement metadata contains intentionally long repeated literals.
package ingestioncanary

import (
	"github.com/GuanceCloud/cliutils/point"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

type IngestionCanaryResultMetric struct{}

func (m *IngestionCanaryResultMetric) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   "ingestion_canary_result",
		Cat:    point.Metric,
		Desc:   "Result metrics of each ingestion canary probe round, including category, status, storage destination and latency.",
		DescZh: "每次采集探测任务的结果指标，包含采集分类、状态、存储目标与耗时。",
		Fields: map[string]interface{}{
			"latency_ms": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				DataType: inputs.Int,
				Unit:     inputs.DurationMS,
				Desc:     "Latency from feeding a canary point until it becomes queryable, in milliseconds.",
				Taggedby: []string{"category", "status", "storage_index", "test_type"},
			},
		},
		Tags: map[string]interface{}{
			"category":      inputs.NewTagInfo("Data category: M (metric), L (logging), T (tracing)"),
			"status":        inputs.NewTagInfo("Test status: ok, timeout, error"),
			"storage_index": inputs.NewTagInfo("Storage index for logging data (optional)"),
		},
	}
}

type IngestionCanaryMetric struct{}

func (m *IngestionCanaryMetric) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   "ingestion_canary",
		Cat:    point.Metric,
		Desc:   "Synthetic metric points emitted by the collector per probe round to verify metric ingestion path.",
		DescZh: "采集器按轮次输出的合成指标点，用于验证指标链路可用性。",
		Fields: map[string]interface{}{
			"round": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				DataType: inputs.Int,
				Unit:     inputs.NCount,
				Desc:     "Round number of the ingestion canary probe.",
				Taggedby: []string{"test_type"},
			},
		},
		Tags: map[string]interface{}{
			"test_type": inputs.NewTagInfo("Test type: collect (collector) or cmd (CLI tool)"),
		},
	}
}

type IngestionCanaryLogging struct{}

func (m *IngestionCanaryLogging) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   "ingestion_canary",
		Cat:    point.Logging,
		Desc:   "Synthetic logging payload emitted by the canary to verify logging ingestion path.",
		DescZh: "采集器生成的合成日志样本，用于验证日志链路可用性。",
		Fields: map[string]interface{}{
			"round": &inputs.FieldInfo{
				DataType: inputs.Int,
				Unit:     inputs.NCount,
				Desc:     "Round number of the ingestion canary probe.",
				Taggedby: []string{"test_type"},
			},
			"message": &inputs.FieldInfo{
				DataType: inputs.String,
				Unit:     inputs.NoUnit,
				Desc:     "Synthetic log message emitted by the ingestion canary probe.",
				Taggedby: []string{"test_type"},
			},
		},
		Tags: map[string]interface{}{
			"test_type": inputs.NewTagInfo("Test type: collect (collector) or cmd (CLI tool)"),
		},
	}
}

type IngestionCanaryTracing struct{}

func (m *IngestionCanaryTracing) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   "ingestion_canary",
		Cat:    point.Tracing,
		Desc:   "Synthetic tracing payload emitted by the canary to verify trace collection and forwarding.",
		DescZh: "采集器生成的合成链路追踪样本，用于验证链路采集与上报链路。",
		Fields: map[string]interface{}{
			"trace_id": &inputs.FieldInfo{
				DataType: inputs.String,
				Unit:     inputs.NoUnit,
				Desc:     "Trace identifier emitted by the ingestion canary probe.",
				Taggedby: []string{"test_type", "span_type", "source", "service"},
			},
			"span_id": &inputs.FieldInfo{
				DataType: inputs.String,
				Unit:     inputs.NoUnit,
				Desc:     "Span identifier emitted by the ingestion canary probe.",
				Taggedby: []string{"test_type", "span_type", "source", "service"},
			},
			"parent_id": &inputs.FieldInfo{
				DataType: inputs.String,
				Unit:     inputs.NoUnit,
				Desc:     "Parent span identifier emitted by the ingestion canary probe.",
				Taggedby: []string{"test_type", "span_type", "source", "service"},
			},
			"resource": &inputs.FieldInfo{
				DataType: inputs.String,
				Unit:     inputs.NoUnit,
				Desc:     "Synthetic resource name emitted by the ingestion canary probe.",
				Taggedby: []string{"test_type", "span_type", "source", "service"},
			},
			"status": &inputs.FieldInfo{
				DataType: inputs.String,
				Unit:     inputs.NoUnit,
				Desc:     "Synthetic span status emitted by the ingestion canary probe.",
				Taggedby: []string{"test_type", "span_type", "source", "service"},
			},
			"start": &inputs.FieldInfo{
				DataType: inputs.Int,
				Unit:     inputs.TimestampUS,
				Desc:     "Trace span start time as a Unix timestamp in microseconds.",
				Taggedby: []string{"test_type", "span_type", "source", "service"},
			},
			"duration": &inputs.FieldInfo{
				DataType: inputs.Int,
				Unit:     inputs.DurationUS,
				Desc:     "Trace span duration in microseconds.",
				Taggedby: []string{"test_type", "span_type", "source", "service"},
			},
			"round": &inputs.FieldInfo{
				DataType: inputs.Int,
				Unit:     inputs.NCount,
				Desc:     "Round number of the ingestion canary probe.",
				Taggedby: []string{"test_type", "span_type", "source", "service"},
			},
		},
		Tags: map[string]interface{}{
			"span_type": inputs.NewTagInfo("Span type"),
			"source":    inputs.NewTagInfo("Source name"),
			"service":   inputs.NewTagInfo("Service name"),
			"test_type": inputs.NewTagInfo("Test type: collect (collector) or cmd (CLI tool)"),
		},
	}
}
