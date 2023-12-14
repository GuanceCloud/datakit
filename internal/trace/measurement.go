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

	return point.NewPointV2(m.Name, append(point.NewTags(m.Tags), point.NewKVs(m.Fields)...), opts...)
}

func (m *TraceMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: m.Name,
		Type: "tracing",
		Tags: map[string]interface{}{
			TagHost:           &inputs.TagInfo{Desc: "Hostname."},
			TagContainerHost:  &inputs.TagInfo{Desc: "Container hostname. Available in OpenTelemetry. Optional."},
			TagEndpoint:       &inputs.TagInfo{Desc: "Endpoint info. Available in SkyWalking, Zipkin. Optional."},
			TagEnv:            &inputs.TagInfo{Desc: "Application environment info. Available in Jaeger. Optional."},
			TagHttpStatusCode: &inputs.TagInfo{Desc: "HTTP response code. Available in DDTrace, OpenTelemetry. Optional."},
			TagHttpMethod:     &inputs.TagInfo{Desc: "HTTP request method name. Available in DDTrace, OpenTelemetry. Optional."},
			TagOperation:      &inputs.TagInfo{Desc: "Span name"},
			TagProject:        &inputs.TagInfo{Desc: "Project name. Available in Jaeger. Optional."},
			TagService:        &inputs.TagInfo{Desc: "Service name. Optional."},
			TagSourceType:     &inputs.TagInfo{Desc: "Tracing source type"},
			TagSpanStatus:     &inputs.TagInfo{Desc: "Span status"},
			TagSpanType:       &inputs.TagInfo{Desc: "Span type"},
			TagVersion:        &inputs.TagInfo{Desc: "Application version info. Available in Jaeger. Optional."},
			TagHttpRoute:      &inputs.TagInfo{Desc: "HTTP route. Optional."},
			TagHttpUrl:        &inputs.TagInfo{Desc: "HTTP URL. Optional."},
		},
		Fields: map[string]interface{}{
			FieldDuration: &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.DurationUS, Desc: "Duration of span"},
			FieldMessage:  &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "Origin content of span"},
			FieldParentID: &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "Parent span ID of current span"},
			TagPid:        &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "Application process id. Available in DDTrace, OpenTelemetry. Optional."},
			FieldResource: &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "Resource name produce current span"},
			FieldSpanid:   &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "Span id"},
			FieldStart:    &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.TimestampUS, Desc: "start time of span."},
			FieldTraceID:  &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "Trace id"},
		},
	}
}
