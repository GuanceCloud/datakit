// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package system

import (
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

type measurement struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
	ts     time.Time
}

//nolint:lll
func (m *measurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{}
}

func (m *measurement) LineProto() (*point.Point, error) {
	return point.NewPoint(m.name, m.tags, m.fields, point.MOpt())
}

type conntrackMeasurement measurement

func (m *conntrackMeasurement) LineProto() (*point.Point, error) {
	return point.NewPoint(m.name, m.tags, m.fields, point.MOpt())
}

//nolint:lll
func (m *conntrackMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: metricNameConntrack,
		Desc: "Connection track metrics (Linux only).",
		Fields: map[string]interface{}{
			"entries":             newFieldInfoCount("Current number of connections."),
			"entries_limit":       newFieldInfoCount("The size of the connection tracking table."),
			"stat_found":          newFieldInfoCount("The number of successful search entries."),
			"stat_invalid":        newFieldInfoCount("The number of packets that cannot be tracked."),
			"stat_ignore":         newFieldInfoCount("The number of reports that have been tracked."),
			"stat_insert":         newFieldInfoCount("The number of packets inserted."),
			"stat_insert_failed":  newFieldInfoCount("The number of packages that failed to insert."),
			"stat_drop":           newFieldInfoCount("The number of packets dropped due to connection tracking failure."),
			"stat_early_drop":     newFieldInfoCount("The number of partially tracked packet entries dropped due to connection tracking table full."),
			"stat_search_restart": newFieldInfoCount("The number of connection tracking table query restarts due to hash table size modification."),
		},
		Tags: map[string]interface{}{
			"host": &inputs.TagInfo{Desc: "hostname"},
		},
	}
}

type filefdMeasurement measurement

func (m *filefdMeasurement) LineProto() (*point.Point, error) {
	return point.NewPoint(m.name, m.tags, m.fields, point.MOpt())
}

//nolint:lll
func (m *filefdMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: metricNameFilefd,
		Desc: "System file handle metrics (Linux only).",
		Fields: map[string]interface{}{
			"allocated":    newFieldInfoCount("The number of allocated file handles."),
			"maximum_mega": &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.NCount, Desc: "The maximum number of file handles, unit M(10^6)."},
		},
		Tags: map[string]interface{}{
			"host": &inputs.TagInfo{Desc: "hostname"},
		},
	}
}

type systemMeasurement measurement

//nolint:lll
func (m *systemMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: metricNameSystem,
		Desc: "Basic information about system operation.",
		Fields: map[string]interface{}{
			"load1":           &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: "CPU load average over the past 1 minute."},
			"load5":           &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: "CPU load average over the past 5 minutes."},
			"load15":          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: "CPU load average over the past 15 minutes."},
			"load1_per_core":  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: "CPU single core load average over the past 1 minute."},
			"load5_per_core":  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: "CPU single core load average over the last 5 minutes."},
			"load15_per_core": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: "CPU single core load average over the past 15 minutes."},
			"n_cpus":          &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "CPU logical core count."},
			"n_users":         &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "User number."},
			"uptime":          &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.DurationSecond, Desc: "System uptime."},
			"memory_usage":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.Percent, Desc: "The percentage of used memory."},
			"cpu_total_usage": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.Percent, Desc: "The percentage of used CPU."},
		},
		Tags: map[string]interface{}{
			"host": &inputs.TagInfo{Desc: "hostname"},
		},
	}
}

func (m *systemMeasurement) LineProto() (*point.Point, error) {
	return point.NewPoint(m.name, m.tags, m.fields, point.MOpt())
}

func newFieldInfoCount(desc string) *inputs.FieldInfo {
	return &inputs.FieldInfo{
		Type:     inputs.Gauge,
		DataType: inputs.Int,
		Unit:     inputs.NCount,
		Desc:     desc,
	}
}
