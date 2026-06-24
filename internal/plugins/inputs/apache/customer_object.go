// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package apache

import (
	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

type customerObjectMeasurement struct {
	name     string
	tags     map[string]string
	fields   map[string]interface{}
	election bool
}

var apacheWebServerTaggedby = []string{"name", "ip", "host", "col_co_status", "reason"}

// Point implement MeasurementV2.
func (m *customerObjectMeasurement) Point() *point.Point {
	opts := point.DefaultObjectOptions()
	if m.election {
		opts = append(opts,
			point.WithExtraTags(datakit.GlobalElectionTags()),
		)
		point.DefaultObjectOptions()
	}
	return point.NewPoint(m.name,
		append(point.NewTags(m.tags), point.NewKVs(m.fields)...),
		opts...)
}

//nolint:lll
func (m *customerObjectMeasurement) Info() *inputs.MeasurementInfo { //nolint:funlen
	return &inputs.MeasurementInfo{
		Name:           "web_server",
		Cat:            point.CustomObject,
		MetaDuplicated: true,
		Desc:           "Apache web server instance inventory and collector status information emitted as a custom object.",
		DescZh:         "以自定义对象形式上报的 Apache Web Server 实例清单与采集器状态信息。",
		Fields: map[string]interface{}{
			"uptime": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.DurationSecond,
				Desc:     "Time in seconds since this instance last started.",
				Taggedby: apacheWebServerTaggedby,
			},

			"display_name": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.String,
				Unit:     inputs.NoUnit,
				Desc:     "Display name shown for this instance in the Datakit UI.",
				Taggedby: apacheWebServerTaggedby,
			},

			"version": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.String,
				Unit:     inputs.NoUnit,
				Desc:     "Server version reported by this instance.",
				Taggedby: apacheWebServerTaggedby,
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
