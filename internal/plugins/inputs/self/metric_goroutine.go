// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package self

import (
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

var measurementGoroutineName = "datakit_goroutine"

type datakitGoroutineMeasurement struct {
	inputs.CommonMeasurement
}

func (m *datakitGoroutineMeasurement) LineProto() (*point.Point, error) {
	return point.NewPoint(m.CommonMeasurement.Name,
		m.CommonMeasurement.Tags,
		m.CommonMeasurement.Fields, point.MOpt())
}

func (m *datakitGoroutineMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: measurementGoroutineName,
		Type: "metric",
		Tags: map[string]interface{}{
			"group": &inputs.TagInfo{Desc: "The group name of the Goroutine."},
		},

		Fields: map[string]interface{}{
			"running_goroutine_num": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				DataType: inputs.Int,
				Unit:     inputs.NCount,
				Desc:     "The number of the running Goroutine",
			},
			"finished_goroutine_num": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				DataType: inputs.Int,
				Unit:     inputs.NCount,
				Desc:     "The number of the finished Goroutine",
			},
			"failed_num": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				DataType: inputs.Int,
				Unit:     inputs.NCount,
				Desc:     "The number of the Goroutine which has failed",
			},
			"total_cost_time": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				DataType: inputs.Int,
				Unit:     inputs.DurationNS,
				Desc:     "Total cost time in nanosecond",
			},
			"min_cost_time": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				DataType: inputs.Int,
				Unit:     inputs.DurationNS,
				Desc:     "Minimum cost time in nanosecond",
			},
			"max_cost_time": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				DataType: inputs.Int,
				Unit:     inputs.DurationNS,
				Desc:     "Maximum cost time in nanosecond",
			},
		},
	}
}
