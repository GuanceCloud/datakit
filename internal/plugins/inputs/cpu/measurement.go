// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package cpu

import (
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

type docMeasurement struct{}

//nolint:lll
var cpuMeasurement = &inputs.MeasurementInfo{
	Name: metricName,
	Type: "metric",
	Fields: map[string]interface{}{
		"usage_user": &inputs.FieldInfo{
			Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.Percent,
			Desc: "% CPU in user mode.",
		},

		"usage_nice": &inputs.FieldInfo{
			Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.Percent,
			Desc: "% CPU in user mode with low priority (nice). Linux only",
		},

		"usage_system": &inputs.FieldInfo{
			Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.Percent,
			Desc: "% CPU in system mode.",
		},

		"usage_idle": &inputs.FieldInfo{
			Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.Percent,
			Desc: "% CPU in the idle task.",
		},

		"usage_iowait": &inputs.FieldInfo{
			Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.Percent,
			Desc: "% CPU waiting for I/O to complete. Linux only",
		},

		"usage_irq": &inputs.FieldInfo{
			Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.Percent,
			Desc: "% CPU servicing hardware interrupts.",
		},

		"usage_softirq": &inputs.FieldInfo{
			Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.Percent,
			Desc: "% CPU servicing soft interrupts. Linux only",
		},

		"usage_steal": &inputs.FieldInfo{
			Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.Percent,
			Desc: "% CPU spent in other operating systems when running in a virtualized environment. Linux only",
		},

		"usage_guest": &inputs.FieldInfo{
			Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.Percent,
			Desc: "% CPU spent running a virtual CPU for guest operating systems. Linux only",
		},

		"usage_guest_nice": &inputs.FieldInfo{
			Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.Percent,
			Desc: "% CPU spent running a nice guest(virtual CPU for guest operating systems). Linux only(>=3.2.0)",
		},

		"usage_total": &inputs.FieldInfo{
			Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.Percent,
			Desc: "% CPU in total active usage, as well as (100 - usage_idle).",
		},
		"core_temperature": &inputs.FieldInfo{
			Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.Celsius,
			Desc: "CPU core temperature. This is collected by default. Only collect the average temperature of all cores.",
		},
		"load5s": &inputs.FieldInfo{
			Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.UnknownUnit,
			Desc: "CPU average load in 5 seconds.",
		},
	},
	Tags: map[string]interface{}{
		"host": &inputs.TagInfo{Desc: "System hostname."},
		"cpu":  &inputs.TagInfo{Desc: "CPU core ID. For `cpu-total`, it means *all-CPUs-in-one-tag*. If you want every CPU's metric, please enable `percpu` option in *cpu.conf* or set `ENV_INPUT_CPU_PERCPU` under K8s"},
	},
}

func (docMeasurement) Info() *inputs.MeasurementInfo {
	return cpuMeasurement
}
