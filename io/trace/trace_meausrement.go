//nolint:lll
package trace

import (
	"time"

	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

type TraceMeasurement struct {
	Name   string
	Tags   map[string]string
	Fields map[string]interface{}
	TS     time.Time
}

func (tm *TraceMeasurement) LineProto() (*dkio.Point, error) {
	return dkio.MakePoint(tm.Name, tm.Tags, tm.Fields, tm.TS)
}

func (tm *TraceMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: tm.Name,
		Type: "tracing",
		Tags: map[string]interface{}{
			TAG_PROJECT:        &inputs.TagInfo{Desc: "project name"},
			TAG_OPERATION:      &inputs.TagInfo{Desc: "span name"},
			TAG_SERVICE:        &inputs.TagInfo{Desc: "service name"},
			TAG_VERSION:        &inputs.TagInfo{Desc: "application version info"},
			TAG_ENV:            &inputs.TagInfo{Desc: "application environment info"},
			TAG_HTTP_METHOD:    &inputs.TagInfo{Desc: "http request method name"},
			TAG_HTTP_CODE:      &inputs.TagInfo{Desc: "http response code"},
			TAG_TYPE:           &inputs.TagInfo{Desc: "span  service type"},
			TAG_ENDPOINT:       &inputs.TagInfo{Desc: "endpoint info"},
			TAG_SPAN_STATUS:    &inputs.TagInfo{Desc: "span status"},
			TAG_SPAN_TYPE:      &inputs.TagInfo{Desc: "span type"},
			TAG_CONTAINER_HOST: &inputs.TagInfo{Desc: "container hostname"},
		},
		Fields: map[string]interface{}{
			FIELD_TRACEID:  &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "Trace ID."},
			FIELD_PARENTID: &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "The parent span ID of current span."},
			FIELD_SPANID:   &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "Span ID."},
			FIELD_START:    &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.TimestampUS, Desc: "The point of start time of span."},
			FIELD_DURATION: &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.DurationUS, Desc: "Duration of span."},
			FIELD_MSG:      &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "The origin content of span."},
			FIELD_RESOURCE: &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "The resource name."},
			FIELD_PID:      &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "The application process id."},
		},
	}
}
