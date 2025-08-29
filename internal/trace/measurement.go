// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//nolint:lll
package trace

import (
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

type TraceMeasurement struct {
	Name              string
	Tags              map[string]string
	Fields            map[string]interface{}
	TS                time.Time
	BuildPointOptions []point.Option
}

// Point implement MeasurementV2.
func (m *TraceMeasurement) Point() *point.Point {
	opts := append(point.CommonLoggingOptions(), point.WithTime(m.TS))
	opts = append(opts, m.BuildPointOptions...)

	return point.NewPoint(m.Name, append(point.NewTags(m.Tags), point.NewKVs(m.Fields)...), opts...)
}

func (m *TraceMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: m.Name,
		Desc: "This is the field description for the trace.",
		Cat:  point.Tracing,
		Tags: map[string]interface{}{
			TagHost:             &inputs.TagInfo{Desc: "Hostname."},
			TagContainerHost:    &inputs.TagInfo{Desc: "Container hostname. Available in OpenTelemetry. Optional."},
			TagEndpoint:         &inputs.TagInfo{Desc: "Endpoint info. Available in SkyWalking, Zipkin. Optional."},
			TagEnv:              &inputs.TagInfo{Desc: "Application environment info. Available in Jaeger. Optional."},
			TagHttpStatusCode:   &inputs.TagInfo{Desc: "HTTP response code. Available in DDTrace, OpenTelemetry. Optional."},
			TagHttpMethod:       &inputs.TagInfo{Desc: "HTTP request method name. Available in DDTrace, OpenTelemetry. Optional."},
			TagOperation:        &inputs.TagInfo{Desc: "Span name"},
			TagProject:          &inputs.TagInfo{Desc: "Project name. Available in Jaeger. Optional."},
			TagService:          &inputs.TagInfo{Desc: "Service name. Optional."},
			TagSourceType:       &inputs.TagInfo{Desc: "Tracing source type"},
			TagSpanStatus:       &inputs.TagInfo{Desc: "Span status"},
			TagSpanType:         &inputs.TagInfo{Desc: "Span type"},
			TagVersion:          &inputs.TagInfo{Desc: "Application version info. Available in Jaeger. Optional."},
			TagHttpRoute:        &inputs.TagInfo{Desc: "HTTP route. Optional."},
			TagHttpUrl:          &inputs.TagInfo{Desc: "HTTP URL. Optional."},
			TagDKFingerprintKey: &inputs.TagInfo{Desc: "DataKit fingerprint is DataKit hostname"},
			TagBaseService:      &inputs.TagInfo{Desc: "Span Base service name"},
			"db_host":           &inputs.TagInfo{Desc: "DB host name: ip or domain name. Optional."},
			"db_system":         &inputs.TagInfo{Desc: "Database system name:mysql,oracle...  Optional."},
			"db_name":           &inputs.TagInfo{Desc: "Database name. Optional."},
			"out_host":          &inputs.TagInfo{Desc: "This is the database host, equivalent to db_host,only DDTrace-go. Optional."},
		},
		Fields: map[string]interface{}{
			FieldDuration: &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.DurationUS, Desc: "Duration of span"},
			FieldMessage:  &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.NoUnit, Desc: "Origin content of span"},
			FieldParentID: &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.NoUnit, Desc: "Parent span ID of current span"},
			FieldResource: &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.NoUnit, Desc: "Resource name produce current span"},
			FieldSpanid:   &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.NoUnit, Desc: "Span id"},
			FieldStart:    &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.TimestampUS, Desc: "start time of span."},
			FieldTraceID:  &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.NoUnit, Desc: "Trace id"},
		},
	}
}

type TracingMetricMeasurement struct{}

func (m TracingMetricMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: TracingMetricName,
		Desc: "This is the field description for the Tracing_metrics.You can select which tags to delete by configuring a blacklist.",
		Cat:  point.Metric,
		Tags: map[string]interface{}{
			TagService:         &inputs.TagInfo{Desc: "Service name. Optional."},
			TagSource:          &inputs.TagInfo{Desc: "Source: DDTrace or OpenTelemetry. Optional."},
			TagOperation:       &inputs.TagInfo{Desc: "Span name"},
			TagEnv:             &inputs.TagInfo{Desc: "Application environment info. Available in Jaeger. Optional."},
			TagSpanStatus:      &inputs.TagInfo{Desc: "Span status"},
			TagVersion:         &inputs.TagInfo{Desc: "Application version info. Available in Jaeger. Optional."},
			FieldResource:      &inputs.TagInfo{Desc: "Application resource name. Optional."},
			TagHost:            &inputs.TagInfo{Desc: "Hostname."},
			TagHttpStatusCode:  &inputs.TagInfo{Desc: "HTTP response code. Available in DDTrace, OpenTelemetry. Optional."},
			TagHttpStatusClass: &inputs.TagInfo{Desc: "HTTP response code class. Available in 2xx 3xx 4xx 5xx. Optional."},
			TagPodName:         &inputs.TagInfo{Desc: "Pod name. Optional."},
			TagPodNamespace:    &inputs.TagInfo{Desc: "Pod namespace. Optional."},
			TagProject:         &inputs.TagInfo{Desc: "Project name. Optional."},
		},
		Fields: map[string]interface{}{
			"hits": &inputs.FieldInfo{
				Type: inputs.NCount, DataType: inputs.Int,
				Unit: inputs.NCount, Desc: "count of spans.",
			},
			"hits_by_http_status": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Int,
				Unit: inputs.NCount, Desc: "Represent the count of hits for a given span break down by HTTP status code.",
			},
			"latency": &inputs.FieldInfo{
				Type: inputs.Histogram, DataType: inputs.Int,
				Unit: inputs.DurationUS,
				Desc: "Represent the latency distribution for all services, resources, and versions across different environments and additional primary tags." +
					" Recommended for all latency measurement use cases.",
			},
			"errors": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Int,
				Unit: inputs.NCount, Desc: "Represent the count of errors for a given span.",
			},
			"errors_by_http_status": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Int,
				Unit: inputs.NCount, Desc: "Represent the count of errors for a given span by HTTP status code.",
			},
			"apdex": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float,
				Unit: inputs.NoUnit, Desc: "Measures the Apdex score for each web service. The currently set satisfaction threshold is 2 seconds." +
					"The tags for this metric are fixed: service, env, version, resource, source. The value range is 0~1.",
			},
		},
	}
}
