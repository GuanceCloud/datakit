// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package swap

import (
	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

type swapMetric struct{}

var swapMeasurement = &inputs.MeasurementInfo{
	Name:   metricName,
	Desc:   "Host swap memory usage and cumulative swap I/O statistics.",
	DescZh: "主机交换分区内存使用情况和累计 swap I/O 统计。",
	Cat:    point.Metric,
	Fields: map[string]interface{}{
		"total": &inputs.FieldInfo{
			Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.SizeByte,
			Desc: "Total host swap memory.",
		},
		"used": &inputs.FieldInfo{
			Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.SizeByte,
			Desc: "Host swap memory used.",
		},
		"free": &inputs.FieldInfo{
			Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.SizeByte,
			Desc: "Host swap memory free.",
		},
		"used_percent": &inputs.FieldInfo{
			Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.Percent,
			Desc: "Host swap memory percentage used.",
		},
		"sin": &inputs.FieldInfo{
			Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.SizeByte,
			Desc: "Moving data from swap space to main memory of the machine.",
		},
		"sout": &inputs.FieldInfo{
			Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.SizeByte,
			Desc: "Moving main memory contents to swap disk when main memory space fills up.",
		},
	},
	Tags: map[string]interface{}{
		"host": &inputs.TagInfo{Desc: "hostname"},
	},
}

func (m *swapMetric) Info() *inputs.MeasurementInfo {
	return swapMeasurement
}
