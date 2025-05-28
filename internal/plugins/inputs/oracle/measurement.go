// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package oracle

import (
	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

const (
	measurementOracleProcess    = "oracle_process"
	measurementOracleTablespace = "oracle_tablespace"
	measurementOracleSystem     = "oracle_system"
	measurementOracleLog        = "oracle_log"
	measurementLockedSession    = "oracle_locked_session"
	measurementWaitingEvent     = "oracle_waiting_event"
)

type lockMeasurement struct{}

func (*lockMeasurement) Point() *point.Point {
	return nil
}

func (*lockMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: measurementLockedSession,
		Desc: `[:octicons-tag-24: Version-1.74.0](../datakit/changelog.md#cl-1.74.0)`,
		Cat:  point.Metric,
		Tags: map[string]any{
			"event":          &inputs.TagInfo{Desc: "Locked session that waiting the specified event name"},
			"host":           &inputs.TagInfo{Desc: "Host name"},
			"oracle_server":  &inputs.TagInfo{Desc: "Server addr"},
			"oracle_service": &inputs.TagInfo{Desc: "Server service"},
		},
		Fields: map[string]any{
			"waiting_session_count": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Locked session count"},
		},
	}
}

type waitingEventMeasurement struct{}

func (*waitingEventMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: measurementWaitingEvent,
		Desc: `[:octicons-tag-24: Version-1.74.0](../datakit/changelog.md#cl-1.74.0)`,
		Cat:  point.Metric,
		Tags: map[string]any{
			"event":      &inputs.TagInfo{Desc: "Event name"},
			"event_type": &inputs.TagInfo{Desc: "Event type, such as `USER/BACKGROUND`"},
			"program":    &inputs.TagInfo{Desc: "Program(process) name that waiting the event"},
			"username":   &inputs.TagInfo{Desc: "Oracle username that waiting the event"},
		},
		Fields: map[string]any{
			"count": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Waiting event count"},
		},
	}
}

type processMeasurement struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
	ts     int64
}

// Point implement MeasurementV2.
func (m *processMeasurement) Point() *point.Point {
	opts := point.DefaultMetricOptions()
	opts = append(opts, point.WithTimestamp(m.ts))

	return point.NewPointV2(m.name,
		append(point.NewTags(m.tags), point.NewKVs(m.fields)...),
		opts...)
}

// https://docs.oracle.com/en/database/oracle/oracle-database/19/refrn/V-PROCESS.html#GUID-BBE32620-1043-4345-9448-51DB21547FEB
//
//nolint:lll
func (m *processMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: measurementOracleProcess,
		Cat:  point.Metric,
		Fields: map[string]interface{}{
			"pga_alloc_mem":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "PGA memory allocated by process"},
			"pga_freeable_mem": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "PGA memory freeable by process"},
			"pga_max_mem":      &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "PGA maximum memory ever allocated by process"},
			"pga_used_mem":     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "PGA memory used by process"},
			"pid":              &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NoUnit, Desc: "Oracle process identifier"},
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
	ts     int64
}

// Point implement MeasurementV2.
func (m *tablespaceMeasurement) Point() *point.Point {
	opts := point.DefaultMetricOptions()
	opts = append(opts, point.WithTimestamp(m.ts))

	return point.NewPointV2(m.name,
		append(point.NewTags(m.tags), point.NewKVs(m.fields)...),
		opts...)
}

//nolint:lll
func (m *tablespaceMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: measurementOracleTablespace,
		Cat:  point.Metric,
		Fields: map[string]interface{}{
			"in_use":     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.Percent, Desc: "Percentage of used space,as a function of the maximum possible Tablespace size"},
			"off_use":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "Total space consumed by the Tablespace, in database blocks"},
			"ts_size":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "Table space size"},
			"used_space": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "Used space"},
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
	ts     int64
}

// Point implement MeasurementV2.
func (m *systemMeasurement) Point() *point.Point {
	opts := point.DefaultMetricOptions()
	opts = append(opts, point.WithTimestamp(m.ts))

	return point.NewPointV2(m.name,
		append(point.NewTags(m.tags), point.NewKVs(m.fields)...),
		opts...)
}

//nolint:lll
func (m *systemMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: measurementOracleSystem,
		Cat:  point.Metric,
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
			"service_response_time":     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: "Service response time"},
			"session_count":             &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Session count"},
			"session_limit_usage":       &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.Percent, Desc: "Session limit usage"},
			"shared_pool_free":          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.Percent, Desc: "Shared pool free memory %"},
			"soft_parse_ratio":          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.Percent, Desc: "Soft parse ratio"},
			"sorts_per_user_call":       &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Sorts per user call"},
			"temp_space_used":           &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "Temp space used"},
			"user_rollbacks":            &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of user rollbacks"},
			"uptime":                    &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.DurationSecond, Desc: "Instance uptime"},
		},
		Tags: map[string]interface{}{
			"host":           &inputs.TagInfo{Desc: "Host name"},
			"oracle_server":  &inputs.TagInfo{Desc: "Server addr"},
			"oracle_service": &inputs.TagInfo{Desc: "Server service"},
			"pdb_name":       &inputs.TagInfo{Desc: "PDB name"},
		},
	}
}

type slowQueryMeasurement struct{}

//nolint:lll
func (x *slowQueryMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: measurementOracleLog,
		Cat:  point.Logging,
		Desc: `For full and detailed field into, see [here](https://docs.oracle.com/en/database/oracle/oracle-database/19/refrn/V-SQLAREA.html){:target="_blank"}`,
		Fields: map[string]any{
			// core performance metrics
			`sql_fulltext`:   &inputs.FieldInfo{Unit: inputs.NoUnit, DataType: inputs.String, Desc: "All characters of the SQL text for the current cursor"},
			`elapsed_time`:   &inputs.FieldInfo{Unit: inputs.DurationUS, DataType: inputs.Int, Desc: "Elapsed time (in microseconds) used by this cursor for parsing, executing, and fetching. If the cursor uses parallel execution, then `ELAPSED_TIME` is the cumulative time..."},
			`cpu_time`:       &inputs.FieldInfo{Unit: inputs.DurationUS, DataType: inputs.Int, Desc: "CPU time (in microseconds) used by this cursor for parsing, executing, and fetching"},
			`executions`:     &inputs.FieldInfo{Unit: inputs.NCount, DataType: inputs.Int, Desc: "Total number of executions, totalled over all the child cursors"},
			`disk_reads`:     &inputs.FieldInfo{Unit: inputs.NCount, DataType: inputs.Int, Desc: "Sum of the number of disk reads over all child cursors"},
			`buffer_gets`:    &inputs.FieldInfo{Unit: inputs.NCount, DataType: inputs.Int, Desc: "Sum of buffer gets over all child cursors"},
			`rows_processed`: &inputs.FieldInfo{Unit: inputs.NCount, DataType: inputs.Int, Desc: "Total number of rows processed on behalf of this SQL statement"},

			// wait metrics
			`user_io_wait_time`:     &inputs.FieldInfo{Unit: inputs.DurationUS, DataType: inputs.Int, Desc: "User I/O Wait Time (in microseconds)"},
			`concurrency_wait_time`: &inputs.FieldInfo{Unit: inputs.DurationUS, DataType: inputs.Int, Desc: "Concurrency wait time (in microseconds)"},
			`application_wait_time`: &inputs.FieldInfo{Unit: inputs.DurationUS, DataType: inputs.Int, Desc: "Application wait time (in microseconds)"},
			`cluster_wait_time`:     &inputs.FieldInfo{Unit: inputs.DurationUS, DataType: inputs.Int, Desc: "Cluster wait time (in microseconds)"},

			// execution plan metrics
			`plan_hash_value`: &inputs.FieldInfo{Unit: inputs.NoUnit, DataType: inputs.Int, Desc: "Numeric representation of the current SQL plan for this cursor. Comparing one `PLAN_HASH_VALUE` to another easily identifies whether or not two plans are the same (rather than comparing the two plans line by line)."},
			`parse_calls`:     &inputs.FieldInfo{Unit: inputs.NCount, DataType: inputs.Int, Desc: "Sum of all parse calls to all the child cursors under this parent"},
			`sorts`:           &inputs.FieldInfo{Unit: inputs.NCount, DataType: inputs.Int, Desc: "Sum of the number of sorts that were done for all the child cursors"},

			// context metrics
			`parsing_schema_name`: &inputs.FieldInfo{Unit: inputs.NoUnit, DataType: inputs.String, Desc: "Schema name that was used to parse this child cursor"},
			`last_active_time`:    &inputs.FieldInfo{Unit: inputs.NoUnit, DataType: inputs.String, Desc: "Time at which the query plan was last active"},

			// other fields
			`username`:    &inputs.FieldInfo{Unit: inputs.NoUnit, DataType: inputs.String, Desc: "Name of the user"},
			`avg_elapsed`: &inputs.FieldInfo{Unit: inputs.DurationUS, DataType: inputs.Int, Desc: "Average elapsed time of executions(`elapsed_time/executions`)"},
			`message`:     &inputs.FieldInfo{Unit: inputs.NoUnit, DataType: inputs.String, Desc: "JSON dump of all queried fields of table `V$SQLAREA`"},
		},

		Tags: map[string]any{
			`sql_id`:       &inputs.TagInfo{Desc: "SQL identifier of the parent cursor in the library cache"},
			`module`:       &inputs.TagInfo{Desc: "Contains the name of the module that was executing when the SQL statement was first parsed as set by calling `DBMS_APPLICATION_INFO.SET_MODULE`"},
			`action`:       &inputs.TagInfo{Desc: "Contains the name of the action that was executing when the SQL statement was first parsed as set by calling `DBMS_APPLICATION_INFO.SET_ACTION`"},
			`command_type`: &inputs.TagInfo{Desc: "Oracle command type definition"},
			`status`:       &inputs.TagInfo{Desc: "Log level, always `warning` here"},
		},
	}
}
