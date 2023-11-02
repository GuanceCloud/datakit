// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package socket

import (
	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

type TCPMeasurement struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
}

type UDPMeasurement struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
}

func (m *TCPMeasurement) Point() *point.Point {
	return point.NewPointV2(m.name,
		append(point.NewTags(m.tags), point.NewKVs(m.fields)...),
		point.DefaultMetricOptions()...)
}

func (m *UDPMeasurement) Point() *point.Point {
	return point.NewPointV2(m.name,
		append(point.NewTags(m.tags), point.NewKVs(m.fields)...),
		point.DefaultMetricOptions()...)
}

func (m *TCPMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "tcp",
		Type: "metric",
		Tags: map[string]interface{}{
			"dest_host": &inputs.TagInfo{Desc: "TCP domain or host, such as `wwww.baidu.com`, `1.2.3.4`"},
			"dest_port": &inputs.TagInfo{Desc: "TCP port, such as `80`"},
			"proto":     &inputs.TagInfo{Desc: "Protocol, const to be `tcp`"},
		},
		Fields: map[string]interface{}{
			"response_time": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.DurationUS,
				Desc:     "TCP connection time(without DNS query time)",
			},
			"response_time_with_dns": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.DurationUS,
				Desc:     "TCP connection time(with DNS query time)",
			},
			"success": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.UnknownUnit,
				Desc:     "1: success/-1: failed",
			},
		},
	}
}

func (m *UDPMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "udp",
		Type: "metric",
		Fields: map[string]interface{}{
			"success": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.UnknownUnit,
				Desc:     "1: success/-1: failed",
			},
		},
		Tags: map[string]interface{}{
			"dest_host": &inputs.TagInfo{Desc: "UDP host"},
			"dest_port": &inputs.TagInfo{Desc: "UDP port"},
			"proto":     &inputs.TagInfo{Desc: "Protocol, const to be `udp`"},
		},
	}
}
