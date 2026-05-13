// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package otlp

import (
	"time"

	"github.com/GuanceCloud/cliutils/point"
	metrics "github.com/GuanceCloud/tracing-protos/opentelemetry-gen-go/metrics/v1"
)

// NumberDataPointValue extracts the scalar value from an OTLP NumberDataPoint.
func NumberDataPointValue(pt *metrics.NumberDataPoint) (any, bool) {
	if isNil(pt) {
		return nil, false
	}

	msg := pt.ProtoReflect()
	oneof := msg.Descriptor().Oneofs().ByName("value")
	if oneof == nil {
		return nil, false
	}

	field := msg.WhichOneof(oneof)
	if field == nil {
		return nil, false
	}

	switch string(field.Name()) {
	case "as_double":
		return msg.Get(field).Float(), true
	case "as_int":
		return msg.Get(field).Int(), true
	default:
		return nil, false
	}
}

// NumberDataPointToPoint converts an OTLP NumberDataPoint into a metric point.
func NumberDataPointToPoint(measurement, field string, kvs point.KVs, pt *metrics.NumberDataPoint, kvOpts []point.KVOption) *point.Point {
	value, ok := NumberDataPointValue(pt)
	if !ok {
		return point.NewPoint(measurement, kvs, point.DefaultMetricOptions()...)
	}

	kvs = kvs.Add(field, value, kvOpts...)

	opts := point.DefaultMetricOptions()
	opts = append(opts, point.WithTime(time.Unix(0, int64(pt.GetTimeUnixNano()))))

	return point.NewPoint(measurement, kvs, opts...)
}

// SummaryDataPointToPoint converts an OTLP SummaryDataPoint into a metric point.
func SummaryDataPointToPoint(measurement, field string, kvs point.KVs, summary *metrics.SummaryDataPoint, kvOpts []point.KVOption) *point.Point {
	kvs = kvs.Add(field+"_count", summary.GetCount(), kvOpts...).
		Add(field+"_sum", summary.GetSum(), kvOpts...)

	opts := point.DefaultMetricOptions()
	opts = append(opts, point.WithTime(time.Unix(0, int64(summary.GetTimeUnixNano()))))

	return point.NewPoint(measurement, kvs, opts...)
}
