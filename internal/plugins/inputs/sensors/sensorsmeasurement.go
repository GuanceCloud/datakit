// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package sensors

import (
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

//nolint:unused
type sensorsMeasurement struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
}

//nolint:lll
func (m *sensorsMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "sensors",
		Type: "metric",
		Tags: map[string]interface{}{
			"hostname": &inputs.TagInfo{Desc: "Host name"},
			"adapter":  &inputs.TagInfo{Desc: "Device adapter"},
			"chip":     &inputs.TagInfo{Desc: "Chip id"},
			"feature":  &inputs.TagInfo{Desc: "Gathering target"},
		},
		Fields: map[string]interface{}{
			"tmep*_crit":       &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.Celsius, Desc: `Critical temperature of this chip, '*' is the order number in the chip list.`},
			"temp*_crit_alarm": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.Celsius, Desc: `Alarm count, '*' is the order number in the chip list.`},
			"temp*_input":      &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.Celsius, Desc: `Current input temperature of this chip, '*' is the order number in the chip list.`},
			"tmep*_max":        &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.Celsius, Desc: `Max temperature of this chip, '*' is the order number in the chip list.`},
		},
	}
}
