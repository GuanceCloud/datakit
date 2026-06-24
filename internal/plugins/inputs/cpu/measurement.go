// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//nolint:lll,goconst,funlen // Measurement metadata contains intentionally long repeated literals.
package cpu

import (
	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

type docMeasurement struct{}

//nolint:lll
var cpuMeasurement = &inputs.MeasurementInfo{
	Name:   metricName,
	Cat:    point.Metric,
	Desc:   "CPU usage, load, and optional core temperature metrics derived from host CPU time deltas and sensor readings.",
	DescZh: "根据主机 CPU 时间差和传感器读数计算的 CPU 使用率、负载及可选的核心温度指标。",
	Fields: map[string]interface{}{
		"usage_user": &inputs.FieldInfo{
			Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.Percent,
			Desc: "Percentage of CPU time spent in user mode during the collection interval.",
		},

		"usage_nice": &inputs.FieldInfo{
			Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.Percent,
			Desc: "Percentage of CPU time spent running low-priority nice user processes during the collection interval. Linux only.",
		},

		"usage_system": &inputs.FieldInfo{
			Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.Percent,
			Desc: "Percentage of CPU time spent in system mode during the collection interval.",
		},

		"usage_idle": &inputs.FieldInfo{
			Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.Percent,
			Desc: "Percentage of CPU time spent idle during the collection interval.",
		},

		"usage_iowait": &inputs.FieldInfo{
			Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.Percent,
			Desc: "Percentage of CPU time spent waiting for I/O completion during the collection interval. Linux only.",
		},

		"usage_irq": &inputs.FieldInfo{
			Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.Percent,
			Desc: "Percentage of CPU time spent servicing hardware interrupts during the collection interval.",
		},

		"usage_softirq": &inputs.FieldInfo{
			Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.Percent,
			Desc: "Percentage of CPU time spent servicing software interrupts during the collection interval. Linux only.",
		},

		"usage_steal": &inputs.FieldInfo{
			Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.Percent,
			Desc: "Percentage of CPU time stolen by other operating systems when running in a virtualized environment. Linux only.",
		},

		"usage_guest": &inputs.FieldInfo{
			Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.Percent,
			Desc: "Percentage of CPU time spent running virtual CPUs for guest operating systems during the collection interval. Linux only.",
		},

		"usage_guest_nice": &inputs.FieldInfo{
			Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.Percent,
			Desc: "Percentage of CPU time spent running low-priority virtual CPUs for guest operating systems during the collection interval. Linux only, kernel 3.2.0 or later.",
		},

		"usage_total": &inputs.FieldInfo{
			Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.Percent,
			Desc: "Total active CPU usage percentage during the collection interval, computed from user, system, nice, iowait, irq, softirq, steal, guest, and guest_nice time.",
		},
		"core_temperature": &inputs.FieldInfo{
			Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.Celsius,
			Desc: "Average CPU core temperature across detected core temperature sensors, in degrees Celsius. Linux only.",
		},
		"load5s": &inputs.FieldInfo{
			Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.NoUnit,
			Desc: "Five-second CPU load estimate as a dimensionless host load value.",
		},
	},
	Tags: map[string]interface{}{
		"host": &inputs.TagInfo{Desc: "System hostname."},
		"cpu":  &inputs.TagInfo{Desc: "CPU identifier. `cpu-total` is the aggregate across all CPUs; per-core values are emitted only when `percpu` is enabled."},
	},
}

func (docMeasurement) Info() *inputs.MeasurementInfo {
	return cpuMeasurement
}
