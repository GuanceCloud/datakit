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
			TAG_HOST:             &inputs.TagInfo{Desc: "Hostname."},
			TAG_CONTAINER_HOST:   &inputs.TagInfo{Desc: "Container hostname. Available in OpenTelemetry. Optional."},
			TAG_ENDPOINT:         &inputs.TagInfo{Desc: "Endpoint info. Available in SkyWalking, Zipkin. Optional."},
			TAG_ENV:              &inputs.TagInfo{Desc: "Application environment info. Available in Jaeger. Optional."},
			TAG_HTTP_STATUS_CODE: &inputs.TagInfo{Desc: "HTTP response code. Available in DDTrace, OpenTelemetry. Optional."},
			TAG_HTTP_METHOD:      &inputs.TagInfo{Desc: "HTTP request method name. Available in DDTrace, OpenTelemetry. Optional."},
			TAG_OPERATION:        &inputs.TagInfo{Desc: "Span name"},
			TAG_PROJECT:          &inputs.TagInfo{Desc: "Project name. Available in Jaeger. Optional."},
			TAG_SERVICE:          &inputs.TagInfo{Desc: "Service name. Optional."},
			TAG_SOURCE_TYPE:      &inputs.TagInfo{Desc: "Tracing source type"},
			TAG_SPAN_STATUS:      &inputs.TagInfo{Desc: "Span status"},
			TAG_SPAN_TYPE:        &inputs.TagInfo{Desc: "Span type"},
			TAG_VERSION:          &inputs.TagInfo{Desc: "Application version info. Available in Jaeger. Optional."},
			TAG_HTTP_ROUTE:       &inputs.TagInfo{Desc: "HTTP route. Optional."},
			TAG_HTTP_URL:         &inputs.TagInfo{Desc: "HTTP URL. Optional."},
		},
		Fields: map[string]interface{}{
			FIELD_DURATION: &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.DurationUS, Desc: "Duration of span"},
			FIELD_MESSAGE:  &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "Origin content of span"},
			FIELD_PARENTID: &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "Parent span ID of current span"},
			TAG_PID:        &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "Application process id. Available in DDTrace, OpenTelemetry. Optional."},
			FIELD_PRIORITY: &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.UnknownUnit, Desc: "Optional."},
			FIELD_RESOURCE: &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "Resource name produce current span"},
			// FIELD_SAMPLE_RATE_GLOBAL: &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.UnknownUnit, Desc: "global sample ratio"},
			FIELD_SPANID:  &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "Span id"},
			FIELD_START:   &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.TimestampUS, Desc: "start time of span."},
			FIELD_TRACEID: &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "Trace id"},
		},
	}
}
