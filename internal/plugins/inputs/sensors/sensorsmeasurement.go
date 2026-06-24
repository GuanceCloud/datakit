// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package sensors

import (
	"github.com/GuanceCloud/cliutils/point"
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
		Name:   "sensors",
		Cat:    point.Metric,
		Desc:   "Hardware temperature sensor metrics collected from system sensor output, keyed by adapter, chip, and feature.",
		DescZh: "从系统传感器输出采集的硬件温度传感器指标，按 adapter、chip 和 feature 标识上报。",
		Tags: map[string]interface{}{
			"hostname": &inputs.TagInfo{Desc: "Host name"},
			"adapter":  &inputs.TagInfo{Desc: "Device adapter"},
			"chip":     &inputs.TagInfo{Desc: "Chip id"},
			"feature":  &inputs.TagInfo{Desc: "Gathering target"},
		},
		Fields: map[string]interface{}{
			"temp*_crit":       &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.Celsius, Desc: `Critical temperature threshold for this sensor, where '*' is the sensor index.`},
			"temp*_crit_alarm": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.Bool, Desc: `Whether the critical temperature alarm is active for this sensor, where '*' is the sensor index.`},
			"temp*_input":      &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.Celsius, Desc: `Current input temperature reported by this sensor, where '*' is the sensor index.`},
			"temp*_max":        &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.Celsius, Desc: `Maximum temperature threshold for this sensor, where '*' is the sensor index.`},
		},
	}
}
