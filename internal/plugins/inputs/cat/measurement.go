// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package cat measurements
package cat

import (
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

type Measurement struct{}

func (m *Measurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Type: "metric",
		Name: inputName,
		Fields: map[string]interface{}{
			"runtime_up-time": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.DurationMS,
				Desc: "Runtime.",
			},

			"runtime_start-time": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.DurationSecond,
				Desc: "Start time.",
			},

			"os_available-processors": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.NCount,
				Desc: "The number of available processors in the host.",
			},

			"os_system-load-average": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.Percent,
				Desc: "Average system load.",
			},

			"os_total-physical-memory": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.SizeByte,
				Desc: "Total physical memory size.",
			},

			"os_free-physical-memory": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.SizeByte,
				Desc: "Free physical memory size.",
			},

			"os_committed-virtual-memory": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.SizeByte,
				Desc: "Committed virtual memory size.",
			},

			"os_total-swap-space": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.SizeByte,
				Desc: "Total swap space size.",
			},

			"os_free-swap-space": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.SizeByte,
				Desc: "Free swap space size",
			},

			"disk_total": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.SizeByte,
				Desc: "Total disk size of data nodes.",
			},

			"disk_free": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.SizeByte,
				Desc: "Free disk size.",
			},

			"disk_usable": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.SizeByte,
				Desc: "Used disk size.",
			},

			"memory_max": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.NCount,
				Desc: "Max memory usage.",
			},

			"memory_total": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.NCount,
				Desc: "Total memory size.",
			},

			"memory_free": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.NCount,
				Desc: "Free memory size.",
			},

			"memory_heap-usage": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.NCount,
				Desc: "The usage of heap memory.",
			},

			"memory_non-heap-usage": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.NCount,
				Desc: "The usage of non heap memory.",
			},

			"thread_count": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.NCount,
				Desc: "Total number of threads.",
			},

			"thread_daemon_count": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.NCount,
				Desc: "The number of daemon threads.",
			},

			"thread_peek_count": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.NCount,
				Desc: "Thread peek.",
			},

			"thread_total_started_count": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.NCount,
				Desc: "Total number of started threads.",
			},

			"thread_cat_thread_count": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.NCount,
				Desc: "The number of threads used by cat.",
			},

			"thread_pigeon_thread_count": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.NCount,
				Desc: "The number of pigeon threads.",
			},

			"thread_http_thread_count": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.NCount,
				Desc: "The number of http threads.",
			},
		},
		Tags: map[string]interface{}{
			"domain":               &inputs.TagInfo{Desc: "IP address."},
			"hostName":             &inputs.TagInfo{Desc: "Host name."},
			"runtime_java-version": &inputs.TagInfo{Desc: "Java version."},
			"runtime_user-name":    &inputs.TagInfo{Desc: "User name."},
			"runtime_user-dir":     &inputs.TagInfo{Desc: "The path of jar."},
			"os_name":              &inputs.TagInfo{Desc: "OS name:'Windows/Linux/Mac',etc."},
			"os_arch":              &inputs.TagInfo{Desc: "CPU architecture:AMD/ARM."},
			"os_version":           &inputs.TagInfo{Desc: "The kernel version of the OS."},
		},
	}
}
