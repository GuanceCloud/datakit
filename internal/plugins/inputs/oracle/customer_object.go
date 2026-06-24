// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package oracle

import (
	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

type customerObjectMeasurement struct {
	name     string
	tags     map[string]string
	fields   map[string]interface{}
	election bool
}

// Point implement MeasurementV2.
func (m *customerObjectMeasurement) Point() *point.Point {
	opts := point.DefaultObjectOptions()
	return point.NewPoint(m.name,
		append(point.NewTags(m.tags), point.NewKVs(m.fields)...),
		opts...)
}

//nolint:lll
func (m *customerObjectMeasurement) Info() *inputs.MeasurementInfo { //nolint:funlen
	return &inputs.MeasurementInfo{
		Name:           "database",
		Desc:           "Oracle database custom object metadata collected from database discovery.",
		DescZh:         "从数据库发现结果采集的 Oracle 数据库自定义对象元数据。",
		MetaDuplicated: true,
		Cat:            point.CustomObject,
		Fields: map[string]interface{}{
			"uptime": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.DurationSecond,
				Desc:     "Time in seconds since this instance last started.",
			},

			"display_name": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.Gauge,
				Unit:     inputs.NoUnit,
				Desc:     "Display name shown for this instance in the Datakit UI.",
			},

			"version": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.Gauge,
				Unit:     inputs.NoUnit,
				Desc:     "Server version reported by this instance.",
			},
		},
		Tags: map[string]interface{}{
			"name": &inputs.TagInfo{
				Desc: "Stable object identifier for this monitored instance.",
			},

			"col_co_status": &inputs.TagInfo{
				Desc: "Collector status for this instance, such as `OK` or `NotOK`.",
			},

			"ip": &inputs.TagInfo{
				Desc: "Configured connection IP address for this instance.",
			},

			"host": &inputs.TagInfo{
				Desc: "Hostname or address of the server that runs this instance.",
			},
			"reason": &inputs.TagInfo{
				Desc: "Reason reported when the collector status is not `OK`.",
			},
		},
	}
}
