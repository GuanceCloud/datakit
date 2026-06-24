// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//nolint:lll,goconst,funlen // Measurement metadata contains intentionally long repeated literals.
package ipmi

import (
	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

type docMeasurement struct{}

// Info , reflected in the document
//
//nolint:lll
func (docMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   metricName,
		Cat:    point.Metric,
		Desc:   "IPMI hardware sensor metrics collected from the monitored host, including current, fan speed, power, temperature, utilization, voltage, and sensor status.",
		DescZh: "通过 IPMI 采集的硬件传感器指标，包含电流、风扇转速、功耗、温度、利用率、电压和传感器状态。",
		Fields: map[string]interface{}{
			"current":           &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.Ampere, Desc: "Electrical current reported by the IPMI sensor."},
			"fan_speed":         &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.RotationRete, Desc: "Fan rotation speed reported by the IPMI sensor."},
			"power_consumption": &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.Watt, Desc: "Power consumption reported by the IPMI sensor."},
			"temp":              &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.Celsius, Desc: "Temperature reported by the IPMI sensor."},
			"usage":             &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.Percent, Desc: "Utilization percentage reported by the IPMI sensor."},
			"voltage":           &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.Volt, Desc: "Voltage reported by the IPMI sensor."},
			"count":             &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.Count, Desc: "Generic count metric matched by the configured IPMI count regular expressions."},
			"status":            &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.NoUnit, Desc: "Numeric status code of the monitored unit."},
			"warning":           &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.Bool, Desc: "Whether the monitored unit is in a warning state: 1 means true and 0 means false."},
		},

		Tags: map[string]interface{}{
			"host": &inputs.TagInfo{Desc: "IPMI server host being monitored."},
			"unit": &inputs.TagInfo{Desc: "Normalized IPMI sensor name on the monitored host."},
		},
	}
}
