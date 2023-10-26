// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package ipmi

import (
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

type docMeasurement struct{}

// Info , reflected in the document
//
//nolint:lll
func (docMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: metricName,
		Fields: map[string]interface{}{
			"current":           &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.Ampere, Desc: "Current."},
			"fan_speed":         &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.RotationRete, Desc: "Fan speed."},
			"power_consumption": &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.Watt, Desc: "Power consumption."},
			"temp":              &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.Celsius, Desc: "Temperature."},
			"usage":             &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.Percent, Desc: "Usage."},
			"voltage":           &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.Volt, Desc: "Voltage."},
			"count":             &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.Count, Desc: "Count."},
			"status":            &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.UnknownUnit, Desc: "Status of the unit."},
			"warning":           &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.UnknownUnit, Desc: "Warning on/off."},
		},

		Tags: map[string]interface{}{
			"host": &inputs.TagInfo{Desc: "Monitored host name"},
			"unit": &inputs.TagInfo{Desc: "Unit name in the host"},
		},
	}
}
