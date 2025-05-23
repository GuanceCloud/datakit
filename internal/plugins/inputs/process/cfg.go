// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package process

import (
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

const (
	inputName      = "host_processes"
	objectFeedName = inputName + "/O"
	category       = "host"

	sampleConfig = `
[[inputs.host_processes]]
  # Only collect these matched process' metrics. For process objects
  # these white list not applied. Process name support regexp.
  # process_name = [".*nginx.*", ".*mysql.*"]

  # Process minimal run time(default 10m)
  # If process running time less than the setting, we ignore it(both for metric and object)
  min_run_time = "10m"

  ## Enable process metric collecting
  open_metric = false

  ## Enable listen ports tag, default is false
  enable_listen_ports = false

  ## Enable open files field, default is false
  enable_open_files = false

  ## only collect container-based process(object and metric)
  only_container_processes = false

  # Extra tags
  [inputs.host_processes.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
  # ...
`
)

type ProcessMetric struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
	ts     int64
}

// Point implement MeasurementV2.
func (m *ProcessMetric) Point() *point.Point {
	opts := point.DefaultMetricOptions()
	opts = append(opts, point.WithTimestamp(m.ts))

	return point.NewPointV2(m.name,
		append(point.NewTags(m.tags), point.NewKVs(m.fields)...),
		opts...)
}

//nolint:lll
func (m *ProcessMetric) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: inputName,
		Desc: "Collect process metrics, including CPU/memory usage, etc.",
		Type: "metric",
		Fields: map[string]interface{}{
			"cpu_usage":        newOtherFieldInfo(inputs.Float, inputs.Gauge, inputs.Percent, "CPU usage, the percentage of CPU occupied by the process since it was started. This value will be more stable (different from the instantaneous percentage of `top`)"),
			"cpu_usage_top":    newOtherFieldInfo(inputs.Float, inputs.Gauge, inputs.Percent, "CPU usage, the average CPU usage of the process within a collection cycle"),
			"mem_used_percent": newOtherFieldInfo(inputs.Float, inputs.Gauge, inputs.Percent, "Memory usage percentage"),
			"open_files":       newOtherFieldInfo(inputs.Int, inputs.Count, inputs.NCount, "Number of open files (only supports Linux)"),
			"rss":              newOtherFieldInfo(inputs.Int, inputs.Gauge, inputs.SizeByte, "Resident Set Size (resident memory size)"),
			"threads":          newOtherFieldInfo(inputs.Int, inputs.Count, inputs.NCount, "Total number of threads"),

			"voluntary_ctxt_switches":    newOtherFieldInfo(inputs.Int, inputs.Count, inputs.NCount, "From /proc/[PID]/status. Context switches that voluntary drop the CPU, such as `sleep()/read()/sched_yield()`. Linux only"),
			"nonvoluntary_ctxt_switches": newOtherFieldInfo(inputs.Int, inputs.Count, inputs.NCount, "From /proc/[PID]/status. Context switches that nonvoluntary drop the CPU. Linux only"),
			"proc_syscr":                 newOtherFieldInfo(inputs.Int, inputs.Count, inputs.NCount, "Linux from */proc/[PID]/io*, Windows from `GetProcessIoCounters()`. Count of `read()` like syscall`. Linux&Windows only"),
			"proc_syscw":                 newOtherFieldInfo(inputs.Int, inputs.Count, inputs.NCount, "Linux from */proc/[PID]/io*, Windows from `GetProcessIoCounters()`. Count of `write()` like syscall`. Linux&Windows only"),
			"proc_read_bytes":            newOtherFieldInfo(inputs.Int, inputs.Count, inputs.SizeByte, "Linux from */proc/[PID]/io*, Windows from `GetProcessIoCounters()`. Read bytes from disk"),
			"proc_write_bytes":           newOtherFieldInfo(inputs.Int, inputs.Count, inputs.SizeByte, "Linux from */proc/[PID]/io*, Windows from `GetProcessIoCounters()`. Written bytes to disk"),
			"page_minor_faults":          newOtherFieldInfo(inputs.Int, inputs.Count, inputs.SizeByte, "Linux from */proc/[PID]/stat*. The number of minor page faults. Linux only"),
			"page_major_faults":          newOtherFieldInfo(inputs.Int, inputs.Count, inputs.SizeByte, "Linux from */proc/[PID]/stat*. The number of major page faults. Linux only"),
			"page_children_minor_faults": newOtherFieldInfo(inputs.Int, inputs.Count, inputs.SizeByte, "Linux from */proc/[PID]/stat*. The number of minor page faults for this process. Linux only"),
			"page_children_major_faults": newOtherFieldInfo(inputs.Int, inputs.Count, inputs.SizeByte, "Linux from */proc/[PID]/stat*. The number of major page faults for this process. Linux only"),
		},
		Tags: map[string]interface{}{
			"container_id": inputs.NewTagInfo("Container ID of the process, only supported Linux"),
			"host":         inputs.NewTagInfo("Host name"),
			"pid":          inputs.NewTagInfo("Process ID"),
			"process_name": inputs.NewTagInfo("Process name"),
			"username":     inputs.NewTagInfo("Username"),
		},
	}
}

type ProcessObject struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
	ts     time.Time
}

// Point implement MeasurementV2.
func (m *ProcessObject) Point() *point.Point {
	opts := point.DefaultObjectOptions()
	opts = append(opts, point.WithTime(m.ts))

	return point.NewPointV2(m.name,
		append(point.NewTags(m.tags), point.NewKVs(m.fields)...),
		opts...)
}

//nolint:lll
func (m *ProcessObject) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: inputName,
		Desc: "Collect data on process objects, including process names, process commands, etc.",
		Type: "object",
		Fields: map[string]interface{}{
			"cmdline":          newOtherFieldInfo(inputs.String, inputs.UnknownType, inputs.UnknownUnit, "Command line parameters for the process"),
			"cpu_usage":        newOtherFieldInfo(inputs.Float, inputs.Gauge, inputs.Percent, "CPU usage, the percentage of CPU occupied by the process since it was started. This value will be more stable (different from the instantaneous percentage of `top`)"),
			"cpu_usage_top":    newOtherFieldInfo(inputs.Float, inputs.Gauge, inputs.Percent, "CPU usage, the average CPU usage of the process within a collection cycle"),
			"listen_ports":     newOtherFieldInfo(inputs.String, inputs.UnknownType, inputs.UnknownUnit, "The port the process is listening on"),
			"mem_used_percent": newOtherFieldInfo(inputs.Float, inputs.Gauge, inputs.Percent, "Memory usage percentage"),
			"message":          newOtherFieldInfo(inputs.String, inputs.UnknownType, inputs.UnknownUnit, "Process details"),
			"open_files":       newOtherFieldInfo(inputs.Int, inputs.Count, inputs.NCount, "Number of open files (only supports Linux, and the `enable_open_files` option needs to be turned on)"),
			"pid":              newOtherFieldInfo(inputs.Int, inputs.UnknownType, inputs.UnknownUnit, "Process ID"),
			"rss":              newOtherFieldInfo(inputs.Int, inputs.Gauge, inputs.SizeByte, "Resident Set Size (resident memory size)"),
			"start_time":       newOtherFieldInfo(inputs.Int, inputs.Gauge, inputs.TimestampMS, "process start time"),
			"started_duration": newOtherFieldInfo(inputs.Int, inputs.Gauge, inputs.TimestampSec, "Process startup time"),
			"state_zombie":     newOtherFieldInfo(inputs.Bool, inputs.Gauge, inputs.UnknownUnit, "Whether it is a zombie process"),
			"threads":          newOtherFieldInfo(inputs.Int, inputs.Count, inputs.NCount, "Total number of threads"),
			"work_directory":   newOtherFieldInfo(inputs.String, inputs.UnknownType, inputs.UnknownUnit, "Working directory (Linux only)"),

			"voluntary_ctxt_switches":    newOtherFieldInfo(inputs.Int, inputs.Count, inputs.NCount, "From /proc/[PID]/status. Context switches that voluntary drop the CPU, such as `sleep()/read()/sched_yield()`. Linux only"),
			"nonvoluntary_ctxt_switches": newOtherFieldInfo(inputs.Int, inputs.Count, inputs.NCount, "From /proc/[PID]/status. Context switches that nonvoluntary drop the CPU. Linux only"),
			"proc_syscr":                 newOtherFieldInfo(inputs.Int, inputs.Count, inputs.NCount, "Linux from */proc/[PID]/io*, Windows from `GetProcessIoCounters()`. Count of `read()` like syscall`. Linux&Windows only"),
			"proc_syscw":                 newOtherFieldInfo(inputs.Int, inputs.Count, inputs.NCount, "Linux from */proc/[PID]/io*, Windows from `GetProcessIoCounters()`. Count of `write()` like syscall`. Linux&Windows only"),
			"proc_read_bytes":            newOtherFieldInfo(inputs.Int, inputs.Count, inputs.SizeByte, "Linux from */proc/[PID]/io*, Windows from `GetProcessIoCounters()`. Read bytes from disk"),
			"proc_write_bytes":           newOtherFieldInfo(inputs.Int, inputs.Count, inputs.SizeByte, "Linux from */proc/[PID]/io*, Windows from `GetProcessIoCounters()`. Written bytes to disk"),
			"page_minor_faults":          newOtherFieldInfo(inputs.Int, inputs.Count, inputs.SizeByte, "Linux from */proc/[PID]/stat*. The number of minor page faults. Linux only"),
			"page_major_faults":          newOtherFieldInfo(inputs.Int, inputs.Count, inputs.SizeByte, "Linux from */proc/[PID]/stat*. The number of major page faults. Linux only"),
			"page_children_minor_faults": newOtherFieldInfo(inputs.Int, inputs.Count, inputs.SizeByte, "Linux from */proc/[PID]/stat*. The number of minor page faults of it's child processes. Linux only"),
			"page_children_major_faults": newOtherFieldInfo(inputs.Int, inputs.Count, inputs.SizeByte, "Linux from */proc/[PID]/stat*. The number of major page faults of it's child processes. Linux only"),
		},
		Tags: map[string]interface{}{
			"container_id": inputs.NewTagInfo("Container ID of the process, only supported Linux"),
			"host":         inputs.NewTagInfo("Host name"),
			"name":         inputs.NewTagInfo("Name field, consisting of `[host-name]_[pid]`"),
			"process_name": inputs.NewTagInfo("Process name"),
			"state":        inputs.NewTagInfo("Process status, currently not supported on Windows"),
			"username":     inputs.NewTagInfo("Username"),
		},
	}
}

func newOtherFieldInfo(datatype, ftype, unit, desc string) *inputs.FieldInfo {
	return &inputs.FieldInfo{
		DataType: datatype,
		Type:     ftype,
		Unit:     unit,
		Desc:     desc,
	}
}
