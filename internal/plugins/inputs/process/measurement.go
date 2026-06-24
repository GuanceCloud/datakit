// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//nolint:lll,goconst,funlen // Measurement metadata contains intentionally long repeated literals.
package process

import (
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

type processMetric struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
	ts     int64
}

// Point implement MeasurementV2.
func (m *processMetric) Point() *point.Point {
	opts := point.DefaultMetricOptions()
	opts = append(opts, point.WithTimestamp(m.ts))

	return point.NewPoint(m.name,
		append(point.NewTags(m.tags), point.NewKVs(m.fields)...),
		opts...)
}

//nolint:lll
func (m *processMetric) Info() *inputs.MeasurementInfo {
	metricTags := []string{"pid", "cmdline", "process_name", "username", "container_id"}

	return &inputs.MeasurementInfo{
		Name:   inputName,
		Desc:   "Per-process runtime metrics collected from operating-system process APIs.",
		DescZh: "从操作系统进程 API 采集的单进程运行时指标。",
		Cat:    point.Metric,
		Fields: map[string]interface{}{
			"cpu_usage":        &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.Percent, Desc: "CPU usage, the percentage of CPU occupied by the process since it was started. This value will be more stable (different from the instantaneous percentage of `top`)", Taggedby: metricTags},
			"cpu_usage_top":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.Percent, Desc: "CPU usage, the average CPU usage of the process within a collection cycle", Taggedby: metricTags},
			"mem_used_percent": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.Percent, Desc: "Memory usage percentage", Taggedby: metricTags},
			"open_files":       &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Current number of open file descriptors for the process. Linux only.", Taggedby: metricTags},
			"rss":              &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "Resident set size in bytes.", Taggedby: metricTags},
			"vms":              &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "Virtual memory size in bytes.", Taggedby: metricTags},
			"threads":          &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Current number of threads in the process.", Taggedby: metricTags},

			"voluntary_ctxt_switches":    &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "From /proc/[PID]/status. Context switches that voluntary drop the CPU, such as `sleep()/read()/sched_yield()`. Linux only", Taggedby: metricTags},
			"nonvoluntary_ctxt_switches": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "From /proc/[PID]/status. Context switches that nonvoluntary drop the CPU. Linux only", Taggedby: metricTags},
			"proc_syscr":                 &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Linux from */proc/[PID]/io*, Windows from `GetProcessIoCounters()`. Count of `read()` like syscall`. Linux&Windows only", Taggedby: metricTags},
			"proc_syscw":                 &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Linux from */proc/[PID]/io*, Windows from `GetProcessIoCounters()`. Count of `write()` like syscall`. Linux&Windows only", Taggedby: metricTags},
			"proc_read_bytes":            &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.SizeByte, Desc: "Linux from */proc/[PID]/io*, Windows from `GetProcessIoCounters()`. Read bytes from disk", Taggedby: metricTags},
			"proc_write_bytes":           &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.SizeByte, Desc: "Linux from */proc/[PID]/io*, Windows from `GetProcessIoCounters()`. Written bytes to disk", Taggedby: metricTags},
			"page_minor_faults":          &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Linux from */proc/[PID]/stat*. The number of minor page faults. Linux only", Taggedby: metricTags},
			"page_major_faults":          &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Linux from */proc/[PID]/stat*. The number of major page faults. Linux only", Taggedby: metricTags},
			"page_children_minor_faults": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Linux from */proc/[PID]/stat*. The number of minor page faults for this process. Linux only", Taggedby: metricTags},
			"page_children_major_faults": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Linux from */proc/[PID]/stat*. The number of major page faults for this process. Linux only", Taggedby: metricTags},
		},
		Tags: map[string]interface{}{
			"container_id": inputs.NewTagInfo("Container ID of the process, only supported Linux"),
			"host":         inputs.NewTagInfo("Host name"),
			"pid":          inputs.NewTagInfo("Process ID"),
			"process_name": inputs.NewTagInfo("Process name"),
			"username":     inputs.NewTagInfo("Username"),
			"cmdline":      inputs.NewTagInfo("Command line parameters for the process"),
		},
	}
}

type processObject struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
	ts     time.Time
}

// Point implement MeasurementV2.
func (m *processObject) Point() *point.Point {
	opts := point.DefaultObjectOptions()
	opts = append(opts, point.WithTime(m.ts))

	return point.NewPoint(m.name,
		append(point.NewTags(m.tags), point.NewKVs(m.fields)...),
		opts...)
}

//nolint:lll
func (m *processObject) Info() *inputs.MeasurementInfo {
	objectTags := []string{"name", "process_name", "username", "state", "container_id"}

	return &inputs.MeasurementInfo{
		Name:   inputName,
		Desc:   "Per-process object records with identity, command, state, resource usage, and OS counter fields.",
		DescZh: "包含进程身份、命令、状态、资源使用量和操作系统计数字段的单进程对象数据。",
		Cat:    point.Object,
		Fields: map[string]interface{}{
			"cmdline":          &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.NoUnit, Desc: "Command line parameters for the process.", Taggedby: objectTags},
			"cpu_usage":        &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.Percent, Desc: "CPU usage, the percentage of CPU occupied by the process since it was started. This value will be more stable (different from the instantaneous percentage of `top`)", Taggedby: objectTags},
			"cpu_usage_top":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.Percent, Desc: "CPU usage, the average CPU usage of the process within a collection cycle", Taggedby: objectTags},
			"listen_ports":     &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.NoUnit, Desc: "JSON array of listening ports for the process.", Taggedby: objectTags},
			"mem_used_percent": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.Percent, Desc: "Memory usage percentage", Taggedby: objectTags},
			"message":          &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.NoUnit, Desc: "JSON summary of collected process tags, fields, memory info, and CPU time details.", Taggedby: objectTags},
			"open_files":       &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Current number of open file descriptors for the process. Linux only.", Taggedby: objectTags},
			"pid":              &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NoUnit, Desc: "Process ID.", Taggedby: objectTags},
			"rss":              &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "Resident set size in bytes.", Taggedby: objectTags},
			"vms":              &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "Virtual memory size in bytes.", Taggedby: objectTags},
			"start_time":       &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.TimestampMS, Desc: "Process start time as a Unix epoch timestamp in milliseconds.", Taggedby: objectTags},
			"started_duration": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.DurationSecond, Desc: "Elapsed time since the process started, in seconds.", Taggedby: objectTags},
			"state_zombie":     &inputs.FieldInfo{DataType: inputs.Bool, Type: inputs.Bool, Unit: inputs.Bool, Desc: "Whether the process is a zombie.", Taggedby: objectTags},
			"threads":          &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Current number of threads in the process.", Taggedby: objectTags},
			"work_directory":   &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.NoUnit, Desc: "Working directory of the process. Linux only.", Taggedby: objectTags},

			"voluntary_ctxt_switches":    &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "From /proc/[PID]/status. Context switches that voluntary drop the CPU, such as `sleep()/read()/sched_yield()`. Linux only", Taggedby: objectTags},
			"nonvoluntary_ctxt_switches": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "From /proc/[PID]/status. Context switches that nonvoluntary drop the CPU. Linux only", Taggedby: objectTags},
			"proc_syscr":                 &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Linux from */proc/[PID]/io*, Windows from `GetProcessIoCounters()`. Count of `read()` like syscall`. Linux&Windows only", Taggedby: objectTags},
			"proc_syscw":                 &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Linux from */proc/[PID]/io*, Windows from `GetProcessIoCounters()`. Count of `write()` like syscall`. Linux&Windows only", Taggedby: objectTags},
			"proc_read_bytes":            &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.SizeByte, Desc: "Linux from */proc/[PID]/io*, Windows from `GetProcessIoCounters()`. Read bytes from disk", Taggedby: objectTags},
			"proc_write_bytes":           &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.SizeByte, Desc: "Linux from */proc/[PID]/io*, Windows from `GetProcessIoCounters()`. Written bytes to disk", Taggedby: objectTags},
			"page_minor_faults":          &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Linux from */proc/[PID]/stat*. The number of minor page faults. Linux only", Taggedby: objectTags},
			"page_major_faults":          &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Linux from */proc/[PID]/stat*. The number of major page faults. Linux only", Taggedby: objectTags},
			"page_children_minor_faults": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Linux from */proc/[PID]/stat*. The number of minor page faults of its child processes. Linux only", Taggedby: objectTags},
			"page_children_major_faults": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Linux from */proc/[PID]/stat*. The number of major page faults of its child processes. Linux only", Taggedby: objectTags},
		},
		Tags: map[string]interface{}{
			"container_id": inputs.NewTagInfo("Container ID of the process if the process is running in container, Linux only"),
			"host":         inputs.NewTagInfo("Host name"),
			"name":         inputs.NewTagInfo("Process object name field, consisting of `[host-name]_[pid]`"),
			"process_name": inputs.NewTagInfo("Process name"),
			"state":        inputs.NewTagInfo("Process status. Supported on Linux and Windows"),
			"username":     inputs.NewTagInfo("Username"),
		},
	}
}
