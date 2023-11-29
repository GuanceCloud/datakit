// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package oracle

import (
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

const (
	oracleProcess    = "oracle_process"
	oracleTablespace = "oracle_tablespace"
	oracleSystem     = "oracle_system"
)

type processMeasurement struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
	ts     time.Time
}

// Point implement MeasurementV2.
func (m *processMeasurement) Point() *point.Point {
	opts := point.DefaultMetricOptions()
	opts = append(opts, point.WithTime(m.ts))

	return point.NewPointV2(m.name,
		append(point.NewTags(m.tags), point.NewKVs(m.fields)...),
		opts...)
}

// https://docs.oracle.com/en/database/oracle/oracle-database/19/refrn/V-PROCESS.html#GUID-BBE32620-1043-4345-9448-51DB21547FEB
//
//nolint:lll
func (m *processMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: oracleProcess,
		Fields: map[string]interface{}{
			"pga_alloc_mem":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "PGA memory allocated by process"},
			"pga_freeable_mem": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "PGA memory freeable by process"},
			"pga_max_mem":      &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "PGA maximum memory ever allocated by process"},
			"pga_used_mem":     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "PGA memory used by process"},
			"pid":              &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: "Oracle process identifier"},
		},
		Tags: map[string]interface{}{
			"host":           &inputs.TagInfo{Desc: "Host name"},
			"oracle_server":  &inputs.TagInfo{Desc: "Server addr"},
			"oracle_service": &inputs.TagInfo{Desc: "Server service"},
			"pdb_name":       &inputs.TagInfo{Desc: "PDB name"},
			"program":        &inputs.TagInfo{Desc: "Program in progress"},
		},
	}
}

type tablespaceMeasurement struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
	ts     time.Time
}

// Point implement MeasurementV2.
func (m *tablespaceMeasurement) Point() *point.Point {
	opts := point.DefaultMetricOptions()
	opts = append(opts, point.WithTime(m.ts))

	return point.NewPointV2(m.name,
		append(point.NewTags(m.tags), point.NewKVs(m.fields)...),
		opts...)
}

//nolint:lll
func (m *tablespaceMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: oracleTablespace,
		Fields: map[string]interface{}{
			"in_use":     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Table space in-use"},
			"off_use":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Table space offline"},
			"ts_size":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "Table space size"},
			"used_space": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Used space"},
		},
		Tags: map[string]interface{}{
			"host":            &inputs.TagInfo{Desc: "Host name"},
			"oracle_server":   &inputs.TagInfo{Desc: "Server addr"},
			"oracle_service":  &inputs.TagInfo{Desc: "Server service"},
			"pdb_name":        &inputs.TagInfo{Desc: "PDB name"},
			"tablespace_name": &inputs.TagInfo{Desc: "Table space name"},
		},
	}
}

type systemMeasurement struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
	ts     time.Time
}

// Point implement MeasurementV2.
func (m *systemMeasurement) Point() *point.Point {
	opts := point.DefaultMetricOptions()
	opts = append(opts, point.WithTime(m.ts))

	return point.NewPointV2(m.name,
		append(point.NewTags(m.tags), point.NewKVs(m.fields)...),
		opts...)
}

//nolint:lll
func (m *systemMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: oracleSystem,
		Desc: "You have to wait for a few minutes to see these metrics when your running Oracle database's version is earlier than 12c.",
		Fields: map[string]interface{}{
			"active_sessions":           &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of active sessions"},
			"buffer_cachehit_ratio":     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.Percent, Desc: "Ratio of buffer cache hits"},
			"cache_blocks_corrupt":      &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Corrupt cache blocks"},
			"cache_blocks_lost":         &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Lost cache blocks"},
			"consistent_read_changes":   &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Consistent read changes per second"},
			"consistent_read_gets":      &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Consistent read gets per second"},
			"cursor_cachehit_ratio":     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.Percent, Desc: "Ratio of cursor cache hits"},
			"database_cpu_time_ratio":   &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.Percent, Desc: "Database CPU time ratio"},
			"database_wait_time_ratio":  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.Percent, Desc: "Memory sorts per second"},
			"db_block_changes":          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "DB block changes per second"},
			"db_block_gets":             &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "DB block gets per second"},
			"disk_sorts":                &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Disk sorts per second"},
			"enqueue_timeouts":          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Enqueue timeouts per second"},
			"execute_without_parse":     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Execute without parse ratio"},
			"gc_cr_block_received":      &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "GC CR block received"},
			"host_cpu_utilization":      &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.Percent, Desc: "Host CPU utilization (%)"},
			"library_cachehit_ratio":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.Percent, Desc: "Ratio of library cache hits"},
			"logical_reads":             &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Logical reads per second"},
			"logons":                    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of logon attempts"},
			"memory_sorts_ratio":        &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.Percent, Desc: "Memory sorts ratio"},
			"pga_over_allocation_count": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Over-allocating PGA memory count"},
			"physical_reads_direct":     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Physical reads direct per second"},
			"physical_reads":            &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Physical reads per second"},
			"physical_writes":           &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Physical writes per second"},
			"redo_generated":            &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Redo generated per second"},
			"redo_writes":               &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Redo writes per second"},
			"rows_per_sort":             &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Rows per sort"},
			"service_response_time":     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.TimestampSec, Desc: "Service response time"},
			"session_count":             &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Session count"},
			"session_limit_usage":       &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.Percent, Desc: "Session limit usage"},
			"shared_pool_free":          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.Percent, Desc: "Shared pool free memory %"},
			"soft_parse_ratio":          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.Percent, Desc: "Soft parse ratio"},
			"sorts_per_user_call":       &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Sorts per user call"},
			"temp_space_used":           &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "Temp space used"},
			"user_rollbacks":            &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of user rollbacks"},
		},
		Tags: map[string]interface{}{
			"host":           &inputs.TagInfo{Desc: "Host name"},
			"oracle_server":  &inputs.TagInfo{Desc: "Server addr"},
			"oracle_service": &inputs.TagInfo{Desc: "Server service"},
			"pdb_name":       &inputs.TagInfo{Desc: "PDB name"},
		},
	}
}
