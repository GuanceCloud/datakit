// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package system

import (
	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

type measurement struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
	ts     int64
}

//nolint:lll
func (m *measurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Desc:   "System metrics emitted with the runtime measurement name.",
		DescZh: "使用运行时指标集名称上报的 system 指标。",
	}
}

type conntrackMeasurement measurement

// Point implement MeasurementV2.
func (m *conntrackMeasurement) Point() *point.Point {
	opts := point.DefaultMetricOptions()
	opts = append(opts, point.WithTimestamp(m.ts))

	return point.NewPoint(m.name,
		append(point.NewTags(m.tags), point.NewKVs(m.fields)...),
		opts...)
}

//nolint:lll
func (m *conntrackMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   metricNameConntrack,
		Cat:    point.Metric,
		Desc:   "Linux connection-tracking table usage and cumulative conntrack event counters.",
		DescZh: "Linux 连接跟踪表使用情况和累计 conntrack 事件计数。",
		Fields: map[string]interface{}{
			"entries":             &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Current number of connections."},
			"entries_limit":       &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The size of the connection tracking table."},
			"stat_found":          &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Cumulative number of successful connection-tracking lookups."},
			"stat_invalid":        &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Cumulative number of packets that could not be tracked."},
			"stat_ignore":         &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Cumulative number of packets ignored by connection tracking."},
			"stat_insert":         &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Cumulative number of new connection-tracking entries inserted."},
			"stat_insert_failed":  &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Cumulative number of connection-tracking entry insert failures."},
			"stat_drop":           &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Cumulative number of packets dropped because connection tracking failed."},
			"stat_early_drop":     &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Cumulative number of tracked entries dropped because the conntrack table was full."},
			"stat_search_restart": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Cumulative number of conntrack lookup restarts caused by hash-table changes."},
		},
		Tags: map[string]interface{}{
			"host": &inputs.TagInfo{Desc: "hostname"},
		},
	}
}

type filefdMeasurement measurement

// Point implement MeasurementV2.
func (m *filefdMeasurement) Point() *point.Point {
	opts := point.DefaultMetricOptions()
	opts = append(opts, point.WithTimestamp(m.ts))

	return point.NewPoint(m.name,
		append(point.NewTags(m.tags), point.NewKVs(m.fields)...),
		opts...)
}

//nolint:lll
func (m *filefdMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   metricNameFilefd,
		Cat:    point.Metric,
		Desc:   "Linux open file handle allocation and limit metrics.",
		DescZh: "Linux 已分配文件句柄数量和文件句柄上限指标。",
		Fields: map[string]interface{}{
			"allocated":    &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The number of allocated file handles."},
			"maximum_mega": &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.NCount, Desc: "Maximum open file handles, expressed in millions."},
		},
		Tags: map[string]interface{}{
			"host": &inputs.TagInfo{Desc: "hostname"},
		},
	}
}

type systemMeasurement measurement

// Point implement MeasurementV2.
func (m *systemMeasurement) Point() *point.Point {
	opts := point.DefaultMetricOptions()
	opts = append(opts, point.WithTimestamp(m.ts))

	return point.NewPoint(m.name,
		append(point.NewTags(m.tags), point.NewKVs(m.fields)...),
		opts...)
}

//nolint:lll
func (m *systemMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   metricNameSystem,
		Cat:    point.Metric,
		Desc:   "Host-wide CPU, load, memory, process, user, and uptime summary metrics.",
		DescZh: "主机 CPU、负载、内存、进程、登录用户和运行时长汇总指标。",
		Fields: map[string]interface{}{
			"cpu_total_usage":        &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.Percent, Desc: "Percentage of host CPU time currently in use."},
			"load1_per_core":         &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NoUnit, Desc: "One-minute load average normalized by logical CPU count."},
			"load1":                  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NoUnit, Desc: "One-minute system load average."},
			"load15_per_core":        &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NoUnit, Desc: "Fifteen-minute load average normalized by logical CPU count."},
			"load15":                 &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NoUnit, Desc: "Fifteen-minute system load average."},
			"load5_per_core":         &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NoUnit, Desc: "Five-minute load average normalized by logical CPU count."},
			"load5":                  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NoUnit, Desc: "Five-minute system load average."},
			"memory_usage":           &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.Percent, Desc: "Percentage of host memory currently in use."},
			"n_cpus":                 &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "CPU logical core count."},
			"n_users":                &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of logged-in users."},
			"processor_queue_length": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of ready threads waiting for processor time across the host (System\\Processor Queue Length). Windows amd64 only."},
			"process_count":          &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of processes currently running on the host."},
			"uptime":                 &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.DurationSecond, Desc: "Time since the host last booted."},
		},
	}
}
