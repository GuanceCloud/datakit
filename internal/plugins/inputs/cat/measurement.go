// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package cat measurements
package cat

import (
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

type Measurement struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
}

func (m *Measurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: inputName,
		Fields: map[string]interface{}{
			"runtime_up-time": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.DurationMS,
				Desc: "运行时长",
			},

			"runtime_start-time": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.DurationSecond,
				Desc: "启动时间",
			},

			"os_available-processors": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.NCount,
				Desc: "主机 CPU 核心数",
			},

			"os_system-load-average": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.Percent,
				Desc: "系统负载",
			},

			"os_total-physical-memory": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.SizeByte,
				Desc: "总物理内存",
			},

			"os_free-physical-memory": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.SizeByte,
				Desc: "空闲内存",
			},

			"os_committed-virtual-memory": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.SizeByte,
				Desc: "使用内存",
			},

			"os_total-swap-space": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.SizeByte,
				Desc: "交换区总大小",
			},

			"os_free-swap-space": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.SizeByte,
				Desc: "交换区空闲大小",
			},

			"disk_total": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.SizeByte,
				Desc: "数据节点磁盘总量",
			},

			"disk_free": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.SizeByte,
				Desc: "磁盘空闲大小",
			},

			"disk_usable": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.SizeByte,
				Desc: "使用大小",
			},

			"memory_max": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.NCount,
				Desc: "最大使用内存",
			},

			"memory_total": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.NCount,
				Desc: "总使用内存",
			},

			"memory_free": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.NCount,
				Desc: "释放内存",
			},

			"memory_heap-usage": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.NCount,
				Desc: "当前使用内存",
			},

			"memory_non-heap-usage": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.NCount,
				Desc: "非堆内存",
			},

			"thread_count": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.NCount,
				Desc: "总线程数量",
			},

			"thread_daemon_count": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.NCount,
				Desc: "活跃线程数量",
			},

			"thread_peek_count": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.NCount,
				Desc: "线程峰值",
			},

			"thread_total_started_count": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.NCount,
				Desc: "总初始化过的线程",
			},

			"thread_cat_thread_count": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.NCount,
				Desc: "cat 使用线程数量",
			},

			"thread_pigeon_thread_count": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.NCount,
				Desc: "pigeon 线程数量",
			},

			"thread_http_thread_count": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.NCount,
				Desc: "http 线程数量",
			},
		},
		Tags: map[string]interface{}{
			"domain":               &inputs.TagInfo{Desc: "IP 地址"},
			"hostName":             &inputs.TagInfo{Desc: "主机名"},
			"runtime_java-version": &inputs.TagInfo{Desc: "Java version"},
			"runtime_user-name":    &inputs.TagInfo{Desc: "用户名"},
			"runtime_user-dir":     &inputs.TagInfo{Desc: "启动程序的 jar 包位置"},
			"os_name":              &inputs.TagInfo{Desc: "操作系统名称：`windows/linux/mac` 等"},
			"os_arch":              &inputs.TagInfo{Desc: "CPU 架构：amd/arm"},
			"os_version":           &inputs.TagInfo{Desc: "操作系统的内核版本"},
		},
	}
}

func (m *Measurement) LineProto() (*point.Point, error) {
	return point.NewPoint(m.name, m.tags, m.fields, nil)
}
