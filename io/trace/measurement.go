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
	return dkio.NewPoint(tm.Name, tm.Tags, tm.Fields, &dkio.PointOption{
		Time:   tm.TS,
		Strict: false,
	})
}

func (tm *TraceMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: tm.Name,
		Type: "tracing",
		Tags: map[string]interface{}{
			TAG_CONTAINER_HOST: &inputs.TagInfo{Desc: "container hostname"},
			TAG_ENDPOINT:       &inputs.TagInfo{Desc: "endpoint info"},
			TAG_ENV:            &inputs.TagInfo{Desc: "application environment info"},
			TAG_HTTP_CODE:      &inputs.TagInfo{Desc: "http response code"},
			TAG_HTTP_METHOD:    &inputs.TagInfo{Desc: "http request method name"},
			TAG_OPERATION:      &inputs.TagInfo{Desc: "span name"},
			TAG_PROJECT:        &inputs.TagInfo{Desc: "project name"},
			TAG_SERVICE:        &inputs.TagInfo{Desc: "service name"},
			TAG_SOURCE_TYPE:    &inputs.TagInfo{Desc: "tracing source type"},
			TAG_SPAN_STATUS:    &inputs.TagInfo{Desc: "span status"},
			TAG_SPAN_TYPE:      &inputs.TagInfo{Desc: "span type"},
			TAG_VERSION:        &inputs.TagInfo{Desc: "application version info"},
		},
		Fields: map[string]interface{}{
			FIELD_DURATION:           &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.DurationUS, Desc: "duration of span"},
			FIELD_MSG:                &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "origin content of span"},
			FIELD_PARENTID:           &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "parent span ID of current span"},
			FIELD_PID:                &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "application process id."},
			FIELD_PRIORITY:           &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.UnknownUnit, Desc: ""},
			FIELD_RESOURCE:           &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "resource name produce current span"},
			FIELD_SAMPLE_RATE_GLOBAL: &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.UnknownUnit, Desc: "global sample ratio"},
			FIELD_SPANID:             &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "span id"},
			FIELD_START:              &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.TimestampUS, Desc: "start time of span."},
			FIELD_TRACEID:            &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "trace id"},
		},
	}
}
