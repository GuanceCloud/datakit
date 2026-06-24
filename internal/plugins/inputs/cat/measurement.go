// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package cat measurements
package cat

import (
	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

type Measurement struct{}

var (
	catRuntimeTaggedby  = []string{"domain", "hostName", "runtime_java-version", "runtime_user-name", "runtime_user-dir"}
	catOSTaggedby       = []string{"domain", "hostName", "os_name", "os_arch", "os_version"}
	catDiskTaggedby     = []string{"domain", "hostName", "os_name"}
	catMemoryTaggedby   = []string{"domain", "hostName"}
	catMemoryGCTaggedby = []string{"domain", "hostName", "memory_gc_name"}
	catThreadTaggedby   = []string{"domain", "hostName"}
)

func (m *Measurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Cat:    point.Metric,
		Name:   inputName,
		Desc:   "CAT heartbeat runtime, operating system, disk, memory, and thread metrics emitted from application heartbeat XML payloads.",
		DescZh: "从 CAT 心跳 XML 负载中解析出的运行时、操作系统、磁盘、内存和线程指标。",
		Fields: map[string]interface{}{
			"runtime_up-time": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.DurationMS,
				Desc: "JVM uptime reported by CAT, in milliseconds.", Taggedby: catRuntimeTaggedby,
			},

			"runtime_start-time": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.DurationSecond,
				Desc: "JVM start time reported by CAT, as a Unix timestamp in seconds.", Taggedby: catRuntimeTaggedby,
			},

			"os_available-processors": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.NCount,
				Desc: "The number of available processors in the host.", Taggedby: catOSTaggedby,
			},

			"os_system-load-average": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.Percent,
				Desc: "Average system load.", Taggedby: catOSTaggedby,
			},

			"os_total-physical-memory": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.SizeByte,
				Desc: "Total physical memory size.", Taggedby: catOSTaggedby,
			},

			"os_free-physical-memory": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.SizeByte,
				Desc: "Free physical memory size.", Taggedby: catOSTaggedby,
			},

			"os_committed-virtual-memory": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.SizeByte,
				Desc: "Committed virtual memory size.", Taggedby: catOSTaggedby,
			},

			"os_total-swap-space": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.SizeByte,
				Desc: "Total swap space size.", Taggedby: catOSTaggedby,
			},

			"os_free-swap-space": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.SizeByte,
				Desc: "Free swap space size", Taggedby: catOSTaggedby,
			},

			"disk_total": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.SizeByte,
				Desc: "Total disk size of the reported disk volume.", Taggedby: catDiskTaggedby,
			},

			"disk_free": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.SizeByte,
				Desc: "Free disk size of the reported disk volume.", Taggedby: catDiskTaggedby,
			},

			"disk_usable": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.SizeByte,
				Desc: "Usable disk size of the reported disk volume.", Taggedby: catDiskTaggedby,
			},

			"memory_max": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.NCount,
				Desc: "Maximum JVM memory size.", Taggedby: catMemoryTaggedby,
			},

			"memory_total": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.NCount,
				Desc: "Current total JVM memory size.", Taggedby: catMemoryTaggedby,
			},

			"memory_free": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.NCount,
				Desc: "Current free JVM memory size.", Taggedby: catMemoryTaggedby,
			},

			"memory_heap-usage": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.NCount,
				Desc: "Used JVM heap memory.", Taggedby: catMemoryTaggedby,
			},

			"memory_non-heap-usage": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.NCount,
				Desc: "Used JVM non-heap memory.", Taggedby: catMemoryTaggedby,
			},

			"memory_gc_count": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.NCount,
				Desc: "Garbage collection count reported for the named GC.", Taggedby: catMemoryGCTaggedby,
			},

			"memory_gc_time": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.DurationMS,
				Desc: "Garbage collection time reported for the named GC, in milliseconds.", Taggedby: catMemoryGCTaggedby,
			},

			"thread_count": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.NCount,
				Desc: "Total number of threads.", Taggedby: catThreadTaggedby,
			},

			"thread_daemon_count": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.NCount,
				Desc: "The number of daemon threads.", Taggedby: catThreadTaggedby,
			},

			"thread_peek_count": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.NCount,
				Desc: "Peak thread count.", Taggedby: catThreadTaggedby,
			},

			"thread_total_started_count": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.NCount,
				Desc: "Total number of started threads.", Taggedby: catThreadTaggedby,
			},

			"thread_cat_thread_count": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.NCount,
				Desc: "The number of threads used by CAT.", Taggedby: catThreadTaggedby,
			},

			"thread_pigeon_thread_count": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.NCount,
				Desc: "The number of Pigeon threads.", Taggedby: catThreadTaggedby,
			},

			"thread_http_thread_count": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.NCount,
				Desc: "The number of HTTP threads.", Taggedby: catThreadTaggedby,
			},
		},
		Tags: map[string]interface{}{
			"domain":               &inputs.TagInfo{Desc: "CAT application domain."},
			"hostName":             &inputs.TagInfo{Desc: "Host name reported by CAT."},
			"runtime_java-version": &inputs.TagInfo{Desc: "Java version."},
			"runtime_user-name":    &inputs.TagInfo{Desc: "User name."},
			"runtime_user-dir":     &inputs.TagInfo{Desc: "User working directory reported in the CAT runtime block."},
			"os_name":              &inputs.TagInfo{Desc: "Operating system name on OS points; reused as disk volume ID on CAT disk points."},
			"os_arch":              &inputs.TagInfo{Desc: "CPU architecture such as AMD64 or ARM."},
			"os_version":           &inputs.TagInfo{Desc: "OS or kernel version."},
			"memory_gc_name":       &inputs.TagInfo{Desc: "Garbage collector name for CAT GC points."},
		},
	}
}
