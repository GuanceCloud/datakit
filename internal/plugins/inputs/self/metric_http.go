// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package self

import (
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

var measurementHTTPName = "datakit_http"

type datakitHTTPMeasurement struct {
	inputs.CommonMeasurement
}

func (m *datakitHTTPMeasurement) LineProto() (*point.Point, error) {
	return point.NewPoint(m.CommonMeasurement.Name,
		m.CommonMeasurement.Tags,
		m.CommonMeasurement.Fields, point.MOpt())
}

func (m *datakitHTTPMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: measurementHTTPName,
		Type: "metric",
		Tags: map[string]interface{}{
			"api": &inputs.TagInfo{Desc: "API router of the DataKit HTTP"},
		},

		Fields: map[string]interface{}{
			"total_request_count": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				DataType: inputs.Int,
				Unit:     inputs.NCount,
				Desc:     "HTTP total request count",
			},
			"max_latency": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				DataType: inputs.Int,
				Unit:     inputs.DurationNS,
				Desc:     "HTTP max latency",
			},
			"avg_latency": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				DataType: inputs.Int,
				Unit:     inputs.DurationNS,
				Desc:     "HTTP average latency",
			},
			"2XX": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				DataType: inputs.Int,
				Unit:     inputs.NCount,
				Desc:     "HTTP status code 2xx count",
			},
			"3XX": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				DataType: inputs.Int,
				Unit:     inputs.NCount,
				Desc:     "HTTP status code 3xx count",
			},
			"4XX": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				DataType: inputs.Int,
				Unit:     inputs.NCount,
				Desc:     "HTTP status code 4xx count",
			},
			"5XX": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				DataType: inputs.Int,
				Unit:     inputs.NCount,
				Desc:     "HTTP status code 5xx count",
			},
			"limited": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				DataType: inputs.Int,
				Unit:     inputs.NCount,
				Desc:     "HTTP limited",
			},
		},
	}
}
