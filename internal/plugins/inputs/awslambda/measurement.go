// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2024-present Guance, Inc.

package awslambda

import (
	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/awslambda/lambdaapi/metrics"
)

type metricMeasurement struct{}

//nolint:lll
func (*metricMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: inputName + "-metric",
		Cat:  point.Metric,
		Fields: map[string]interface{}{
			metrics.MaxMemoryUsedMetric: &inputs.FieldInfo{
				Type: inputs.Histogram, DataType: inputs.Int, Unit: inputs.SizeMBits,
				Desc: "Maximum memory used in MB.",
			},
			metrics.MemorySizeMetric: &inputs.FieldInfo{
				Type: inputs.Histogram, DataType: inputs.Int, Unit: inputs.SizeMBits,
				Desc: "Memory size configured for the Lambda function in MB.",
			},
			metrics.RuntimeDurationMetric: &inputs.FieldInfo{
				Type: inputs.Histogram, DataType: inputs.Int, Unit: inputs.DurationMS,
				Desc: "Duration of the runtime in milliseconds.",
			},
			metrics.BilledDurationMetric: &inputs.FieldInfo{
				Type: inputs.Histogram, DataType: inputs.Int, Unit: inputs.DurationMS,
				Desc: "Billed duration in milliseconds.",
			},
			metrics.DurationMetric: &inputs.FieldInfo{
				Type: inputs.Histogram, DataType: inputs.Int, Unit: inputs.DurationMS,
				Desc: "Total duration in milliseconds.",
			},
			metrics.PostRuntimeDurationMetric: &inputs.FieldInfo{
				Type: inputs.Histogram, DataType: inputs.Int, Unit: inputs.DurationMS,
				Desc: "Duration of the post-runtime phase in milliseconds.",
			},
			metrics.InitDurationMetric: &inputs.FieldInfo{
				Type: inputs.Histogram, DataType: inputs.Int, Unit: inputs.DurationMS,
				Desc: "Initialization duration in milliseconds.",
			},
			metrics.ResponseLatencyMetric: &inputs.FieldInfo{
				Type: inputs.Histogram, DataType: inputs.Int, Unit: inputs.DurationMS,
				Desc: "Response latency in milliseconds.",
			},
			metrics.ResponseDurationMetric: &inputs.FieldInfo{
				Type: inputs.Histogram, DataType: inputs.Int, Unit: inputs.DurationMS,
				Desc: "Response duration in milliseconds.",
			},
			metrics.ProducedBytesMetric: &inputs.FieldInfo{
				Type: inputs.Histogram, DataType: inputs.Int, Unit: inputs.SizeByte,
				Desc: "Bytes produced.",
			},
			metrics.OutOfMemoryMetric: &inputs.FieldInfo{
				Type: inputs.Histogram, DataType: inputs.Int, Unit: inputs.Count,
				Desc: "Out of memory errors count.",
			},
			metrics.TimeoutsMetric: &inputs.FieldInfo{
				Type: inputs.Histogram, DataType: inputs.Int, Unit: inputs.Count,
				Desc: "Timeouts count.",
			},
			metrics.ErrorsMetric: &inputs.FieldInfo{
				Type: inputs.Histogram, DataType: inputs.Int, Unit: inputs.Count,
				Desc: "Errors count.",
			},
			metrics.InvocationsMetric: &inputs.FieldInfo{
				Type: inputs.Histogram, DataType: inputs.Int, Unit: inputs.Count,
				Desc: "Invocation count.",
			},
		},
		Tags: measurementTags(),
	}
}

type logMeasurement struct{}

//nolint:lll
func (*logMeasurement) Info() *inputs.MeasurementInfo {
	tags := map[string]interface{}{}
	tags[AWSLogFrom] = &inputs.TagInfo{Desc: "log sources, currently only function are supported"}
	return &inputs.MeasurementInfo{
		Name: inputName + "-logging",
		Cat:  point.Logging,
		Fields: map[string]interface{}{
			"message": &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.NoUnit, Desc: "Log message."},
		},
		Tags: tags,
	}
}

func measurementTags() map[string]interface{} {
	return map[string]interface{}{
		LambdaFunctionName:       &inputs.TagInfo{Desc: "Lambda function name."},
		LambdaFunctionVersion:    &inputs.TagInfo{Desc: "Lambda function version."},
		AWSRegion:                &inputs.TagInfo{Desc: "AWS region where the function is executed."},
		LambdaFunctionMemorySize: &inputs.TagInfo{Desc: "Configured memory size for the Lambda function."},
		LambdaInitializationType: &inputs.TagInfo{Desc: "Initialization type of the Lambda function."},
		AccountID:                &inputs.TagInfo{Desc: "AWS Account ID."},
	}
}
