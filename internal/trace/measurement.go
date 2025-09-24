// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//nolint:lll
package trace

import (
	"fmt"
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
		Name:   m.Name,
		Desc:   "Following is tags/fields of tracing data",
		DescZh: "以下是采集上来的 tracing 字段说明",
		Cat:    point.Tracing,
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
			TagDKFingerprintKey: &inputs.TagInfo{Desc: "DataKit fingerprint(always DataKit's hostname)"},
			TagBaseService:      &inputs.TagInfo{Desc: "Span base service name"},
			TagDBHost:           &inputs.TagInfo{Desc: "DB host name: ip or domain name. Optional."},
			TagDBSystem:         &inputs.TagInfo{Desc: "Database system name:mysql,oracle...  Optional."},
			TagDBName:           &inputs.TagInfo{Desc: "Database name. Optional."},
			TagOutHost:          &inputs.TagInfo{Desc: "This is the database host, equivalent to db_host,only DDTrace-go. Optional."},
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

type TracingMetricMeasurement struct {
	Name,
	Source string
}

func (m TracingMetricMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   TracingMetricName,
		Desc:   fmt.Sprintf("Based on %s's span data, we count span count, span cost metrics", m.Name),
		DescZh: fmt.Sprintf("基于 %s 统计得到的指标数据，它记录了所产生的 span 计数、span 耗时等指标", m.Name),
		Cat:    point.Metric,
		Tags: map[string]interface{}{
			TagService:         &inputs.TagInfo{Desc: "Service name."},
			TagSource:          &inputs.TagInfo{Desc: fmt.Sprintf("Source, always `%s`", m.Source)},
			TagOperation:       &inputs.TagInfo{Desc: "Span name"},
			TagEnv:             &inputs.TagInfo{Desc: "Application environment info(if set in span)."},
			TagSpanStatus:      &inputs.TagInfo{Desc: "Span status(`ok/error`)"},
			TagVersion:         &inputs.TagInfo{Desc: "Application version info."},
			FieldResource:      &inputs.TagInfo{Desc: "Application resource name."},
			TagHost:            &inputs.TagInfo{Desc: "Hostname."},
			TagHttpStatusCode:  &inputs.TagInfo{Desc: "HTTP response code"},
			TagHttpStatusClass: &inputs.TagInfo{Desc: "HTTP response code class, such as `2xx/3xx/4xx/5xx`"},
			TagPodName:         &inputs.TagInfo{Desc: "Pod name(if set in span)."},
			TagPodNamespace:    &inputs.TagInfo{Desc: "Pod namespace(if set in span)."},
			TagProject:         &inputs.TagInfo{Desc: "Project name(if set in span)."},
			TagRemoteIP:        &inputs.TagInfo{Desc: "Remote IP."},
		},
		Fields: map[string]interface{}{
			"hits": &inputs.FieldInfo{
				Type: inputs.NCount, DataType: inputs.Int,
				Unit: inputs.NCount, Desc: "Count of spans.",
			},
			"hits_by_http_status": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Int,
				Unit: inputs.NCount, Desc: "Represent the count of hits for a given span group by HTTP status code.",
			},
			"latency_bucket": &inputs.FieldInfo{
				Type: inputs.Histogram, DataType: inputs.Int,
				Unit: inputs.NCount,
				Desc: "Represent the latency distribution for all services, resources, and versions across different environments and additional primary tags." +
					" Recommended for all latency measurement use cases. Use the 'le' tag for filtering",
			},
			"latency_sum": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Int,
				Unit: inputs.DurationUS,
				Desc: "The total latency of all web spans, corresponding to the 'latency_count'",
			},
			"latency_count": &inputs.FieldInfo{
				Type: inputs.NCount, DataType: inputs.Int,
				Unit: inputs.NCount,
				Desc: "The number of spans is equal to the number of web type spans.",
			},
			"errors": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Int,
				Unit: inputs.NCount, Desc: "Represent the count of errors for spans.",
			},
			"errors_by_http_status": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Int,
				Unit: inputs.NCount, Desc: "Represent the count of errors for a given span group by HTTP status code.",
			},
			"apdex": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float,
				Unit: inputs.NoUnit, Desc: "Measures the Apdex score for each web service. The currently set satisfaction threshold is 2 seconds." +
					"The tags for this metric are fixed: `service/env/version/resource/source`. The value range is 0~1.",
			},
		},
	}
}
