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
	return point.NewPoint(m.name,
		append(point.NewTags(m.tags), point.NewKVs(m.fields)...),
		point.DefaultMetricOptions()...)
}

func (m *UDPMeasurement) Point() *point.Point {
	return point.NewPoint(m.name,
		append(point.NewTags(m.tags), point.NewKVs(m.fields)...),
		point.DefaultMetricOptions()...)
}

func (m *TCPMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   "tcp",
		Desc:   "TCP socket probe metrics for configured destination hosts and ports.",
		DescZh: "针对已配置目标主机和端口的 TCP socket 探测指标。",
		Cat:    point.Metric,
		Tags: map[string]interface{}{
			"dest_host": &inputs.TagInfo{Desc: "Configured TCP destination domain name or IP address, such as `www.google.com` or `1.2.3.4`."},
			"dest_port": &inputs.TagInfo{Desc: "Configured TCP destination port, such as `80`."},
			"proto":     &inputs.TagInfo{Desc: "Socket protocol. Always `tcp` for this measurement."},
		},
		Fields: map[string]interface{}{
			"response_time": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.DurationUS,
				Desc:     "TCP connection duration excluding DNS lookup time.",
				Taggedby: []string{"dest_host", "dest_port", "proto"},
			},
			"response_time_with_dns": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.DurationUS,
				Desc:     "TCP connection duration including DNS lookup time.",
				Taggedby: []string{"dest_host", "dest_port", "proto"},
			},
			"success": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NoUnit,
				Desc:     "Socket probe result: 1 for success and -1 for failure.",
				Taggedby: []string{"dest_host", "dest_port", "proto"},
			},
		},
	}
}

func (m *UDPMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   "udp",
		Desc:   "UDP socket probe result metrics for configured destination hosts and ports.",
		DescZh: "针对已配置目标主机和端口的 UDP socket 探测结果指标。",
		Cat:    point.Metric,
		Fields: map[string]interface{}{
			"success": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NoUnit,
				Desc:     "Socket probe result: 1 for success and -1 for failure.",
				Taggedby: []string{"dest_host", "dest_port", "proto"},
			},
		},
		Tags: map[string]interface{}{
			"dest_host": &inputs.TagInfo{Desc: "Configured UDP destination domain name or IP address."},
			"dest_port": &inputs.TagInfo{Desc: "Configured UDP destination port."},
			"proto":     &inputs.TagInfo{Desc: "Socket protocol. Always `udp` for this measurement."},
		},
	}
}
