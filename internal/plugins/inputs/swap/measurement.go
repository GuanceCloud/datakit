// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package swap

import "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"

type swapMetric struct{}

var swapMeasurement = &inputs.MeasurementInfo{
	Name: metricName,
	Type: "metric",
	Fields: map[string]interface{}{
		"total": &inputs.FieldInfo{
			Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.SizeByte,
			Desc: "Host swap memory free.",
		},
		"used": &inputs.FieldInfo{
			Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.SizeByte,
			Desc: "Host swap memory used.",
		},
		"free": &inputs.FieldInfo{
			Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.SizeByte,
			Desc: "Host swap memory total.",
		},
		"used_percent": &inputs.FieldInfo{
			Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.Percent,
			Desc: "Host swap memory percentage used.",
		},
		"in": &inputs.FieldInfo{
			Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.SizeByte,
			Desc: "Moving data from swap space to main memory of the machine.",
		},
		"out": &inputs.FieldInfo{
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
