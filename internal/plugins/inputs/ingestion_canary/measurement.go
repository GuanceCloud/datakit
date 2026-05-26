// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package ingestioncanary

import (
	"github.com/GuanceCloud/cliutils/point"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

type IngestionCanaryResultMetric struct{}

func (m *IngestionCanaryResultMetric) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "ingestion_canary_result",
		Cat:  point.Metric,
		Fields: map[string]interface{}{
			"latency_ms": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				DataType: inputs.Int,
				Unit:     inputs.DurationMS,
				Desc:     "Latency from feed to queryable in milliseconds",
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
		Name: "ingestion_canary",
		Cat:  point.Metric,
		Fields: map[string]interface{}{
			"round": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				DataType: inputs.Int,
				Unit:     inputs.UnknownUnit,
				Desc:     "Round number of the ingestion canary probe",
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
		Name: "ingestion_canary",
		Cat:  point.Logging,
		Fields: map[string]interface{}{
			"round": &inputs.FieldInfo{
				DataType: inputs.Int,
				Unit:     inputs.UnknownUnit,
				Desc:     "Round number of the ingestion canary probe",
			},
			"message": &inputs.FieldInfo{
				DataType: inputs.String,
				Unit:     inputs.UnknownUnit,
				Desc:     "Synthetic freshness probe message",
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
		Name: "ingestion_canary",
		Cat:  point.Tracing,
		Fields: map[string]interface{}{
			"trace_id": &inputs.FieldInfo{
				DataType: inputs.String,
				Unit:     inputs.UnknownUnit,
				Desc:     "Trace ID",
			},
			"span_id": &inputs.FieldInfo{
				DataType: inputs.String,
				Unit:     inputs.UnknownUnit,
				Desc:     "Span ID",
			},
			"parent_id": &inputs.FieldInfo{
				DataType: inputs.String,
				Unit:     inputs.UnknownUnit,
				Desc:     "Parent span ID",
			},
			"resource": &inputs.FieldInfo{
				DataType: inputs.String,
				Unit:     inputs.UnknownUnit,
				Desc:     "Resource name",
			},
			"status": &inputs.FieldInfo{
				DataType: inputs.String,
				Unit:     inputs.UnknownUnit,
				Desc:     "Span status",
			},
			"start": &inputs.FieldInfo{
				DataType: inputs.Int,
				Unit:     inputs.TimestampUS,
				Desc:     "Start time in microseconds",
			},
			"duration": &inputs.FieldInfo{
				DataType: inputs.Int,
				Unit:     inputs.DurationUS,
				Desc:     "Duration in microseconds",
			},
			"round": &inputs.FieldInfo{
				DataType: inputs.Int,
				Unit:     inputs.UnknownUnit,
				Desc:     "Round number of the ingestion canary probe",
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
