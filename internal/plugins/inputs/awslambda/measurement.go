// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2024-present Guance, Inc.

package awslambda

import (
	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/awslambda/lambdaapi/metrics"
	itrace "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/trace"
)

var (
	awslambdaMetricTaggedby = []string{
		LambdaFunctionName,
		LambdaFunctionVersion,
		AWSRegion,
		LambdaFunctionMemorySize,
		LambdaInitializationType,
		AccountID,
		AWSRuntimeVersion,
		AWSRuntimeVersionARN,
	}
	awslambdaLogTaggedby = []string{
		LambdaFunctionName,
		LambdaFunctionVersion,
		AWSRegion,
		LambdaFunctionMemorySize,
		LambdaInitializationType,
		AccountID,
		AWSRuntimeVersion,
		AWSRuntimeVersionARN,
		AWSLogFrom,
	}
)

type metricMeasurement struct{}

//nolint:lll
func (*metricMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   inputName,
		Cat:    point.Metric,
		Desc:   "AWS Lambda metrics produced from Lambda Telemetry API platform events. Duration and memory fields are per-event values, while error, timeout, out-of-memory, and invocation fields are event count increments when emitted.",
		DescZh: "基于 AWS Lambda Telemetry API 平台事件生成的指标。时长和内存类字段是单次事件值，错误、超时、内存不足和调用次数字段在发出时表示一次增量计数。",
		Fields: map[string]interface{}{
			metrics.MaxMemoryUsedMetric: &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.SizeMB,
				Desc: "Maximum memory used by the invocation, in MB, from the Telemetry API platform.report metrics.maxMemoryUsedMB field.", Taggedby: awslambdaMetricTaggedby,
			},
			metrics.MemorySizeMetric: &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.SizeMB,
				Desc: "Configured memory size for the Lambda function, in MB, from the Telemetry API platform.report metrics.memorySizeMB field.", Taggedby: awslambdaMetricTaggedby,
			},
			metrics.RuntimeDurationMetric: &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.DurationMS,
				Desc: "Runtime phase duration for the invocation, in milliseconds, from the Telemetry API platform.runtimeDone metrics.durationMs field.", Taggedby: awslambdaMetricTaggedby,
			},
			metrics.BilledDurationMetric: &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.DurationMS,
				Desc: "Billed duration for the invocation, in milliseconds, from the Telemetry API platform.report metrics.billedDurationMs field.", Taggedby: awslambdaMetricTaggedby,
			},
			metrics.DurationMetric: &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.DurationMS,
				Desc: "Invocation duration, in milliseconds, from the Telemetry API platform.report metrics.durationMs field.", Taggedby: awslambdaMetricTaggedby,
			},
			metrics.PostRuntimeDurationMetric: &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.DurationMS,
				Desc: "Time between runtime completion and the platform report, in milliseconds, calculated from platform.runtimeDone and platform.report durationMs values.", Taggedby: awslambdaMetricTaggedby,
			},
			metrics.InitDurationMetric: &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.DurationMS,
				Desc: "Initialization phase duration, in milliseconds, from the Telemetry API platform.initReport metrics.durationMs field.", Taggedby: awslambdaMetricTaggedby,
			},
			metrics.ResponseLatencyMetric: &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.DurationMS,
				Desc: "Response latency span duration, in milliseconds, when a Telemetry API span named responseLatency is emitted.", Taggedby: awslambdaMetricTaggedby,
			},
			metrics.ResponseDurationMetric: &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.DurationMS,
				Desc: "Response duration span, in milliseconds, when a Telemetry API span named responseDuration is emitted.", Taggedby: awslambdaMetricTaggedby,
			},
			metrics.ProducedBytesMetric: &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.SizeByte,
				Desc: "Bytes produced by the invocation response, from the Telemetry API platform.runtimeDone metrics.producedBytes field when present.", Taggedby: awslambdaMetricTaggedby,
			},
			metrics.OutOfMemoryMetric: &inputs.FieldInfo{
				Type: inputs.Count, DataType: inputs.Int, Unit: inputs.NCount,
				Desc: "Count increment for Lambda out-of-memory failures when emitted.", Taggedby: awslambdaMetricTaggedby,
			},
			metrics.TimeoutsMetric: &inputs.FieldInfo{
				Type: inputs.Count, DataType: inputs.Int, Unit: inputs.NCount,
				Desc: "Count increment for invocations whose Telemetry API platform.runtimeDone status is timeout.", Taggedby: awslambdaMetricTaggedby,
			},
			metrics.ErrorsMetric: &inputs.FieldInfo{
				Type: inputs.Count, DataType: inputs.Int, Unit: inputs.NCount,
				Desc: "Count increment for invocations whose Telemetry API platform.runtimeDone status is not success.", Taggedby: awslambdaMetricTaggedby,
			},
			metrics.InvocationsMetric: &inputs.FieldInfo{
				Type: inputs.Count, DataType: inputs.Int, Unit: inputs.NCount,
				Desc: "Count increment for Lambda invocations when emitted.", Taggedby: awslambdaMetricTaggedby,
			},
		},
		Tags: measurementTags(),
	}
}

type logMeasurement struct{}

//nolint:lll
func (*logMeasurement) Info() *inputs.MeasurementInfo {
	tags := map[string]interface{}{}
	tags[AWSLogFrom] = &inputs.TagInfo{Desc: "Log source type, currently function or extension."}
	for k, v := range measurementTags() {
		tags[k] = v
	}
	return &inputs.MeasurementInfo{
		Name:   inputName,
		Cat:    point.Logging,
		Desc:   "AWS Lambda log records forwarded from the Lambda Telemetry API.",
		DescZh: "通过 AWS Lambda Telemetry API 转发的 Lambda 日志记录。",
		Fields: map[string]interface{}{
			"message": &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.NoUnit, Desc: "Log message.", Taggedby: awslambdaLogTaggedby},
		},
		Tags: tags,
	}
}

func measurementTags() map[string]interface{} {
	return map[string]interface{}{
		LambdaFunctionName:       &inputs.TagInfo{Desc: "Lambda function name."},
		LambdaFunctionVersion:    &inputs.TagInfo{Desc: "Lambda function version."},
		AWSRegion:                &inputs.TagInfo{Desc: "AWS region where the function is executed."},
		LambdaFunctionMemorySize: &inputs.TagInfo{Desc: "Configured memory size for the Lambda function, in MB."},
		LambdaInitializationType: &inputs.TagInfo{Desc: "Initialization type of the Lambda function."},
		AccountID:                &inputs.TagInfo{Desc: "AWS Account ID."},
		AWSRuntimeVersion: &inputs.TagInfo{
			Desc: "Lambda runtime version reported by Telemetry API init or restore start events when available.",
		},
		AWSRuntimeVersionARN: &inputs.TagInfo{
			Desc: "Lambda runtime version ARN reported by Telemetry API init or restore start events when available.",
		},
	}
}

type traceMeasurement struct{}

func (*traceMeasurement) Info() *inputs.MeasurementInfo {
	return (&itrace.TraceMeasurement{Name: inputName}).Info()
}
