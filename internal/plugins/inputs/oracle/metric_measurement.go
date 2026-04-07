// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//nolint:lll // Metric descriptions are intentionally long for clarity
package oracle

import (
	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

const (
	TagGroupCommon        = "common"
	TagGroupProcess       = "process"
	TagGroupTablespace    = "tablespace"
	TagGroupSystem        = "system"
	TagGroupLockedSession = "locked_session"
	TagGroupWaitingEvent  = "waiting_event"
	TagGroupDbmMetric     = "dbm_metric"
	TagGroupDbmSession    = "dbm_session"
	TagGroupDbmConnection = "dbm_connection"
)

type oracleMeasurement struct{}

//nolint:funlen // Info function contains all metric definitions
func (m *oracleMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   measurementOracle,
		Desc:   "Metric set including Oracle process, tablespace, system, locked session, waiting event, and DBM (metric/session/connection) statistics, unified in v2",
		DescZh: "指标集包含 Oracle process、tablespace、system、locked session、waiting event 和 DBM (metric/session/connection) 相关指标，v2 版本统一",
		Cat:    point.Metric,
		Tags:   m.getTags(),
		Fields: m.getFields(),
	}
}

func (m *oracleMeasurement) getTags() map[string]interface{} {
	return mergeMaps(
		m.getCommonTags(),
		m.getProcessTags(),
		m.getTablespaceTags(),
		m.getSystemTags(),
		m.getLockedSessionTags(),
		m.getWaitingEventTags(),
		m.getDbmMetricTags(),
		m.getDbmSessionTags(),
		m.getDbmConnectionTags(),
	)
}

func (m *oracleMeasurement) getCommonTags() map[string]interface{} {
	tags := make(map[string]interface{})
	tags["host"] = &inputs.TagInfo{Desc: "Host name"}
	tags["database_instance"] = &inputs.TagInfo{Desc: "Oracle instance identifier from configured tag or v$instance.host_name."}
	tags["server"] = &inputs.TagInfo{Desc: "The address of the server. The value is `host:port`"}
	tags["oracle_server"] = &inputs.TagInfo{Desc: "Server addr. Deprecated. Please use `server`"}
	tags["oracle_service"] = &inputs.TagInfo{Desc: "Server service"}
	return tags
}

func (m *oracleMeasurement) getProcessTags() map[string]interface{} {
	tags := make(map[string]interface{})
	tags["pdb_name"] = &inputs.TagInfo{Desc: "PDB name"}
	tags["program"] = &inputs.TagInfo{Desc: "Program in progress"}
	return tags
}

func (m *oracleMeasurement) getTablespaceTags() map[string]interface{} {
	tags := make(map[string]interface{})
	tags["pdb_name"] = &inputs.TagInfo{Desc: "PDB name"}
	tags["tablespace_name"] = &inputs.TagInfo{Desc: "Table space name"}
	return tags
}

func (m *oracleMeasurement) getSystemTags() map[string]interface{} {
	tags := make(map[string]interface{})
	tags["pdb_name"] = &inputs.TagInfo{Desc: "PDB name"}
	return tags
}

func (m *oracleMeasurement) getLockedSessionTags() map[string]interface{} {
	tags := make(map[string]interface{})
	tags["event"] = &inputs.TagInfo{Desc: "Locked session that waiting the specified event name"}
	return tags
}

func (m *oracleMeasurement) getWaitingEventTags() map[string]interface{} {
	tags := make(map[string]interface{})
	tags["event"] = &inputs.TagInfo{Desc: "Event name"}
	tags["event_type"] = &inputs.TagInfo{Desc: "Event type, such as `USER/BACKGROUND`"}
	tags["program"] = &inputs.TagInfo{Desc: "Program(process) name that waiting the event"}
	tags["username"] = &inputs.TagInfo{Desc: "Oracle username that waiting the event"}
	return tags
}

func (m *oracleMeasurement) getDbmMetricTags() map[string]interface{} {
	tags := make(map[string]interface{})
	tags["cdb_name"] = &inputs.TagInfo{Desc: "The name of the CDB (Container Database)"}
	tags["con_id"] = &inputs.TagInfo{Desc: "The container ID (con_id) in Oracle multi tenant architecture"}
	tags["pdb_name"] = &inputs.TagInfo{Desc: "The name of the PDB (Pluggable Database)"}
	tags["force_matching_signature"] = &inputs.TagInfo{Desc: "The force matching signature of the query"}
	tags["plan_hash_value"] = &inputs.TagInfo{Desc: "The hash value of the query execution plan"}
	tags["query_signature"] = &inputs.TagInfo{Desc: "Hash signature generated to link metrics and objects"}
	return tags
}

func (m *oracleMeasurement) getDbmSessionTags() map[string]interface{} {
	tags := make(map[string]interface{})
	tags["cdb_name"] = &inputs.TagInfo{Desc: "The name of the CDB (Container Database)"}
	tags["pdb_name"] = &inputs.TagInfo{Desc: "The name of the PDB (Pluggable Database)"}
	tags["username"] = &inputs.TagInfo{Desc: "The name of the database user"}
	tags["session_status"] = &inputs.TagInfo{Desc: "Session status: active (ACTIVE status), idle (INACTIVE status), blocked (being blocked)"}
	tags["wait_group"] = &inputs.TagInfo{Desc: "Datakit unified wait group: Lock, I/O, Concurrency, Memory, Network, CPU, Commit/Log, Other."}
	return tags
}

func (m *oracleMeasurement) getDbmConnectionTags() map[string]interface{} {
	tags := make(map[string]interface{})
	tags["cdb_name"] = &inputs.TagInfo{Desc: "The name of the CDB (Container Database)"}
	tags["pdb_name"] = &inputs.TagInfo{Desc: "The name of the PDB (Pluggable Database)"}
	tags["username"] = &inputs.TagInfo{Desc: "The name of the database user"}
	tags["connection_status"] = &inputs.TagInfo{Desc: "Connection status: ACTIVE, INACTIVE, KILLED, etc."}
	return tags
}

func (m *oracleMeasurement) getFields() map[string]interface{} {
	return mergeMaps(
		m.getProcessFields(),
		m.getTablespaceFields(),
		m.getSystemFields(),
		m.getLockedSessionFields(),
		m.getWaitingEventFields(),
		m.getDbmMetricFields(),
		m.getDbmSessionFields(),
		m.getDbmConnectionFields(),
	)
}

func (m *oracleMeasurement) getProcessFields() map[string]interface{} {
	fields := make(map[string]interface{})
	fields["pga_alloc_mem"] = &inputs.FieldInfo{
		DataType: inputs.Float,
		Type:     inputs.Gauge,
		Unit:     inputs.SizeByte,
		Desc:     "PGA memory allocated by process",
	}
	fields["pga_freeable_mem"] = &inputs.FieldInfo{
		DataType: inputs.Float,
		Type:     inputs.Gauge,
		Unit:     inputs.SizeByte,
		Desc:     "PGA memory freeable by process",
	}
	fields["pga_max_mem"] = &inputs.FieldInfo{
		DataType: inputs.Float,
		Type:     inputs.Gauge,
		Unit:     inputs.SizeByte,
		Desc:     "PGA maximum memory ever allocated by process",
	}
	fields["pga_used_mem"] = &inputs.FieldInfo{
		DataType: inputs.Float,
		Type:     inputs.Gauge,
		Unit:     inputs.SizeByte,
		Desc:     "PGA memory used by process",
	}
	fields["pid"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.NoUnit,
		Desc:     "Oracle process identifier",
	}

	m.addTaggedbyToFields(fields, TagGroupProcess)
	return fields
}

func (m *oracleMeasurement) getTablespaceFields() map[string]interface{} {
	fields := make(map[string]interface{})
	fields["in_use"] = &inputs.FieldInfo{
		DataType: inputs.Float,
		Type:     inputs.Gauge,
		Unit:     inputs.Percent,
		Desc:     "Percentage of used space,as a function of the maximum possible Tablespace size",
	}
	fields["off_use"] = &inputs.FieldInfo{
		DataType: inputs.Float,
		Type:     inputs.Gauge,
		Unit:     inputs.SizeByte,
		Desc:     "Total space consumed by the Tablespace, in database blocks",
	}
	fields["ts_size"] = &inputs.FieldInfo{
		DataType: inputs.Float,
		Type:     inputs.Gauge,
		Unit:     inputs.SizeByte,
		Desc:     "Table space size",
	}
	fields["used_space"] = &inputs.FieldInfo{
		DataType: inputs.Float,
		Type:     inputs.Gauge,
		Unit:     inputs.SizeByte,
		Desc:     "Used space",
	}

	m.addTaggedbyToFields(fields, TagGroupTablespace)
	return fields
}

func (m *oracleMeasurement) getSystemFields() map[string]interface{} {
	fields := make(map[string]interface{})
	fields["active_sessions"] = &inputs.FieldInfo{
		DataType: inputs.Float,
		Type:     inputs.Gauge,
		Unit:     inputs.NCount,
		Desc:     "Number of active sessions",
	}
	fields["buffer_cachehit_ratio"] = &inputs.FieldInfo{
		DataType: inputs.Float,
		Type:     inputs.Gauge,
		Unit:     inputs.Percent,
		Desc:     "Ratio of buffer cache hits",
	}
	fields["cache_blocks_corrupt"] = &inputs.FieldInfo{
		DataType: inputs.Float,
		Type:     inputs.Gauge,
		Unit:     inputs.NCount,
		Desc:     "Corrupt cache blocks",
	}
	fields["cache_blocks_lost"] = &inputs.FieldInfo{
		DataType: inputs.Float,
		Type:     inputs.Gauge,
		Unit:     inputs.NCount,
		Desc:     "Lost cache blocks",
	}
	fields["consistent_read_changes"] = &inputs.FieldInfo{
		DataType: inputs.Float,
		Type:     inputs.Gauge,
		Unit:     inputs.NCount,
		Desc:     "Consistent read changes per second",
	}
	fields["consistent_read_gets"] = &inputs.FieldInfo{
		DataType: inputs.Float,
		Type:     inputs.Gauge,
		Unit:     inputs.NCount,
		Desc:     "Consistent read gets per second",
	}
	fields["cursor_cachehit_ratio"] = &inputs.FieldInfo{
		DataType: inputs.Float,
		Type:     inputs.Gauge,
		Unit:     inputs.Percent,
		Desc:     "Ratio of cursor cache hits",
	}
	fields["database_cpu_time_ratio"] = &inputs.FieldInfo{
		DataType: inputs.Float,
		Type:     inputs.Gauge,
		Unit:     inputs.Percent,
		Desc:     "Database CPU time ratio",
	}
	fields["database_wait_time_ratio"] = &inputs.FieldInfo{
		DataType: inputs.Float,
		Type:     inputs.Gauge,
		Unit:     inputs.Percent,
		Desc:     "Memory sorts per second",
	}
	fields["db_block_changes"] = &inputs.FieldInfo{
		DataType: inputs.Float,
		Type:     inputs.Gauge,
		Unit:     inputs.NCount,
		Desc:     "DB block changes per second",
	}
	fields["db_block_gets"] = &inputs.FieldInfo{
		DataType: inputs.Float,
		Type:     inputs.Gauge,
		Unit:     inputs.NCount,
		Desc:     "DB block gets per second",
	}
	fields["disk_sorts"] = &inputs.FieldInfo{
		DataType: inputs.Float,
		Type:     inputs.Gauge,
		Unit:     inputs.NCount,
		Desc:     "Disk sorts per second",
	}
	fields["enqueue_timeouts"] = &inputs.FieldInfo{
		DataType: inputs.Float,
		Type:     inputs.Gauge,
		Unit:     inputs.NCount,
		Desc:     "Enqueue timeouts per second",
	}
	fields["execute_without_parse"] = &inputs.FieldInfo{
		DataType: inputs.Float,
		Type:     inputs.Gauge,
		Unit:     inputs.NCount,
		Desc:     "Execute without parse ratio",
	}
	fields["gc_cr_block_received"] = &inputs.FieldInfo{
		DataType: inputs.Float,
		Type:     inputs.Gauge,
		Unit:     inputs.NCount,
		Desc:     "GC CR block received",
	}
	fields["host_cpu_utilization"] = &inputs.FieldInfo{
		DataType: inputs.Float,
		Type:     inputs.Gauge,
		Unit:     inputs.Percent,
		Desc:     "Host CPU utilization (%)",
	}
	fields["library_cachehit_ratio"] = &inputs.FieldInfo{
		DataType: inputs.Float,
		Type:     inputs.Gauge,
		Unit:     inputs.Percent,
		Desc:     "Ratio of library cache hits",
	}
	fields["logical_reads"] = &inputs.FieldInfo{
		DataType: inputs.Float,
		Type:     inputs.Gauge,
		Unit:     inputs.NCount,
		Desc:     "Logical reads per second",
	}
	fields["logons"] = &inputs.FieldInfo{
		DataType: inputs.Float,
		Type:     inputs.Gauge,
		Unit:     inputs.NCount,
		Desc:     "Number of logon attempts",
	}
	fields["memory_sorts_ratio"] = &inputs.FieldInfo{
		DataType: inputs.Float,
		Type:     inputs.Gauge,
		Unit:     inputs.Percent,
		Desc:     "Memory sorts ratio",
	}
	fields["pga_over_allocation_count"] = &inputs.FieldInfo{
		DataType: inputs.Float,
		Type:     inputs.Gauge,
		Unit:     inputs.NCount,
		Desc:     "Over-allocating PGA memory count",
	}
	fields["physical_reads_direct"] = &inputs.FieldInfo{
		DataType: inputs.Float,
		Type:     inputs.Gauge,
		Unit:     inputs.NCount,
		Desc:     "Physical reads direct per second",
	}
	fields["physical_reads"] = &inputs.FieldInfo{
		DataType: inputs.Float,
		Type:     inputs.Gauge,
		Unit:     inputs.NCount,
		Desc:     "Physical reads per second",
	}
	fields["physical_writes"] = &inputs.FieldInfo{
		DataType: inputs.Float,
		Type:     inputs.Gauge,
		Unit:     inputs.NCount,
		Desc:     "Physical writes per second",
	}
	fields["redo_generated"] = &inputs.FieldInfo{
		DataType: inputs.Float,
		Type:     inputs.Gauge,
		Unit:     inputs.NCount,
		Desc:     "Redo generated per second",
	}
	fields["redo_writes"] = &inputs.FieldInfo{
		DataType: inputs.Float,
		Type:     inputs.Gauge,
		Unit:     inputs.NCount,
		Desc:     "Redo writes per second",
	}
	fields["rows_per_sort"] = &inputs.FieldInfo{
		DataType: inputs.Float,
		Type:     inputs.Gauge,
		Unit:     inputs.NCount,
		Desc:     "Rows per sort",
	}
	fields["service_response_time"] = &inputs.FieldInfo{
		DataType: inputs.Float,
		Type:     inputs.Gauge,
		Unit:     inputs.DurationMS,
		Desc:     "Service response time",
	}
	fields["session_count"] = &inputs.FieldInfo{
		DataType: inputs.Float,
		Type:     inputs.Gauge,
		Unit:     inputs.NCount,
		Desc:     "Session count",
	}
	fields["session_limit_usage"] = &inputs.FieldInfo{
		DataType: inputs.Float,
		Type:     inputs.Gauge,
		Unit:     inputs.Percent,
		Desc:     "Session limit usage",
	}
	fields["shared_pool_free"] = &inputs.FieldInfo{
		DataType: inputs.Float,
		Type:     inputs.Gauge,
		Unit:     inputs.Percent,
		Desc:     "Shared pool free memory %",
	}
	fields["soft_parse_ratio"] = &inputs.FieldInfo{
		DataType: inputs.Float,
		Type:     inputs.Gauge,
		Unit:     inputs.Percent,
		Desc:     "Soft parse ratio",
	}
	fields["sorts_per_user_call"] = &inputs.FieldInfo{
		DataType: inputs.Float,
		Type:     inputs.Gauge,
		Unit:     inputs.NCount,
		Desc:     "Sorts per user call",
	}
	fields["temp_space_used"] = &inputs.FieldInfo{
		DataType: inputs.Float,
		Type:     inputs.Gauge,
		Unit:     inputs.SizeByte,
		Desc:     "Temp space used",
	}
	fields["user_rollbacks"] = &inputs.FieldInfo{
		DataType: inputs.Float,
		Type:     inputs.Gauge,
		Unit:     inputs.NCount,
		Desc:     "Number of user rollbacks",
	}
	fields["uptime"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.DurationSecond,
		Desc:     "Instance uptime",
	}

	m.addTaggedbyToFields(fields, TagGroupSystem)
	return fields
}

func (m *oracleMeasurement) getLockedSessionFields() map[string]interface{} {
	fields := make(map[string]interface{})
	fields["waiting_session_count"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.NCount,
		Desc:     "Locked session count",
	}

	m.addTaggedbyToFields(fields, TagGroupLockedSession)
	return fields
}

func (m *oracleMeasurement) getWaitingEventFields() map[string]interface{} {
	fields := make(map[string]interface{})
	fields["count"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.NCount,
		Desc:     "Waiting event count",
	}

	m.addTaggedbyToFields(fields, TagGroupWaitingEvent)
	return fields
}

//nolint:funlen
func (m *oracleMeasurement) getDbmMetricFields() map[string]interface{} {
	fields := make(map[string]interface{})
	// Total fields (cumulative values from Oracle)
	fields["total_executions"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.NCount,
		Desc:     "The total number of query executions (cumulative value from Oracle).",
	}
	fields["total_elapsed_time"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.DurationUS,
		Desc:     "The total elapsed time (microseconds) for query executions (cumulative value from Oracle).",
	}
	fields["total_cpu_time"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.DurationUS,
		Desc:     "The total CPU time (microseconds) consumed by query executions (cumulative value from Oracle).",
	}
	fields["total_buffer_gets"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.NCount,
		Desc:     "The total number of buffer gets (cumulative value from Oracle).",
	}
	fields["total_rows_processed"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.NCount,
		Desc:     "The total number of rows processed (cumulative value from Oracle).",
	}
	// Gauge fields
	fields["version_count"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.NCount,
		Desc:     "The number of versions of the cursor in the shared pool.",
	}
	fields["sharable_mem"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.SizeByte,
		Desc:     "The amount of sharable memory (bytes) used by the cursor.",
	}
	fields["typecheck_mem"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.SizeByte,
		Desc:     "The amount of typecheck memory (bytes) used by the cursor.",
	}
	// Delta fields (change between collection intervals)
	fields["delta_parse_calls"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Count,
		Unit:     inputs.NCount,
		Desc:     "The number of parse calls during the collection interval (delta value).",
	}
	fields["delta_disk_reads"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Count,
		Unit:     inputs.NCount,
		Desc:     "The number of disk reads during the collection interval (delta value).",
	}
	fields["delta_direct_writes"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Count,
		Unit:     inputs.NCount,
		Desc:     "The number of direct writes during the collection interval (delta value).",
	}
	fields["delta_direct_reads"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Count,
		Unit:     inputs.NCount,
		Desc:     "The number of direct reads during the collection interval (delta value).",
	}
	fields["delta_buffer_gets"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Count,
		Unit:     inputs.NCount,
		Desc:     "The number of buffer gets during the collection interval (delta value).",
	}
	fields["delta_rows_processed"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Count,
		Unit:     inputs.NCount,
		Desc:     "The number of rows processed during the collection interval (delta value).",
	}
	fields["delta_serializable_aborts"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Count,
		Unit:     inputs.NCount,
		Desc:     "The number of serializable aborts during the collection interval (delta value).",
	}
	fields["delta_fetches"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Count,
		Unit:     inputs.NCount,
		Desc:     "The number of fetches during the collection interval (delta value).",
	}
	fields["delta_executions"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Count,
		Unit:     inputs.NCount,
		Desc:     "The number of query executions during the collection interval (delta value).",
	}
	fields["delta_end_of_fetch_count"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Count,
		Unit:     inputs.NCount,
		Desc:     "The number of end of fetch operations during the collection interval (delta value).",
	}
	fields["delta_loads"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Count,
		Unit:     inputs.NCount,
		Desc:     "The number of loads during the collection interval (delta value).",
	}
	fields["delta_invalidations"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Count,
		Unit:     inputs.NCount,
		Desc:     "The number of invalidations during the collection interval (delta value).",
	}
	fields["delta_px_servers_executions"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Count,
		Unit:     inputs.NCount,
		Desc:     "The number of parallel server executions during the collection interval (delta value).",
	}
	fields["delta_cpu_time"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Count,
		Unit:     inputs.DurationUS,
		Desc:     "The CPU time (microseconds) consumed during the collection interval (delta value).",
	}
	fields["delta_elapsed_time"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Count,
		Unit:     inputs.DurationUS,
		Desc:     "The elapsed time (microseconds) for query executions during the collection interval (delta value).",
	}
	fields["avg_elapsed_time"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.DurationUS,
		Desc:     "The average elapsed time (microseconds) per query execution during the collection interval (calculated from delta_elapsed_time / delta_executions).",
	}
	fields["delta_application_wait_time"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Count,
		Unit:     inputs.DurationUS,
		Desc:     "The application wait time (microseconds) during the collection interval (delta value).",
	}
	fields["delta_concurrency_wait_time"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Count,
		Unit:     inputs.DurationUS,
		Desc:     "The concurrency wait time (microseconds) during the collection interval (delta value).",
	}
	fields["delta_cluster_wait_time"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Count,
		Unit:     inputs.DurationUS,
		Desc:     "The cluster wait time (microseconds) during the collection interval (delta value).",
	}
	fields["delta_user_io_wait_time"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Count,
		Unit:     inputs.DurationUS,
		Desc:     "The user I/O wait time (microseconds) during the collection interval (delta value).",
	}
	fields["delta_plsql_exec_time"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Count,
		Unit:     inputs.DurationUS,
		Desc:     "The PL/SQL execution time (microseconds) during the collection interval (delta value).",
	}
	fields["delta_java_exec_time"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Count,
		Unit:     inputs.DurationUS,
		Desc:     "The Java execution time (microseconds) during the collection interval (delta value).",
	}
	fields["delta_sorts"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Count,
		Unit:     inputs.NCount,
		Desc:     "The number of sorts during the collection interval (delta value).",
	}
	fields["delta_io_cell_offload_eligible_bytes"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Count,
		Unit:     inputs.SizeByte,
		Desc:     "The I/O cell offload eligible bytes during the collection interval (delta value).",
	}
	fields["delta_io_cell_uncompressed_bytes"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Count,
		Unit:     inputs.SizeByte,
		Desc:     "The I/O cell uncompressed bytes during the collection interval (delta value).",
	}
	fields["delta_io_cell_offload_returned_bytes"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Count,
		Unit:     inputs.SizeByte,
		Desc:     "The I/O cell offload returned bytes during the collection interval (delta value).",
	}
	fields["delta_io_interconnect_bytes"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Count,
		Unit:     inputs.SizeByte,
		Desc:     "The I/O interconnect bytes during the collection interval (delta value).",
	}
	fields["delta_physical_read_requests"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Count,
		Unit:     inputs.NCount,
		Desc:     "The number of physical read requests during the collection interval (delta value).",
	}
	fields["delta_physical_read_bytes"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Count,
		Unit:     inputs.SizeByte,
		Desc:     "The number of physical read bytes during the collection interval (delta value).",
	}
	fields["delta_physical_write_requests"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Count,
		Unit:     inputs.NCount,
		Desc:     "The number of physical write requests during the collection interval (delta value).",
	}
	fields["delta_physical_write_bytes"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Count,
		Unit:     inputs.SizeByte,
		Desc:     "The number of physical write bytes during the collection interval (delta value).",
	}
	fields["delta_obsolete_count"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Count,
		Unit:     inputs.NCount,
		Desc:     "The obsolete count during the collection interval (delta value).",
	}
	fields["delta_avoided_executions"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Count,
		Unit:     inputs.NCount,
		Desc:     "The avoided executions during the collection interval (delta value).",
	}

	m.addTaggedbyToFields(fields, TagGroupDbmMetric)
	return fields
}

func (m *oracleMeasurement) getDbmSessionFields() map[string]interface{} {
	fields := make(map[string]interface{})
	fields["session_group_count"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.NCount,
		Desc:     "Number of sessions in this dimension group",
	}
	fields["session_total_wait_time"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.DurationMS,
		Desc:     "Total wait time for sessions in this group (milliseconds). Note: Only active sessions have this data.",
	}
	fields["session_blocked_count"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.NCount,
		Desc:     "Number of sessions that are being blocked",
	}
	fields["session_blocking_count"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.NCount,
		Desc:     "Number of sessions that are blocking other sessions",
	}

	m.addTaggedbyToFields(fields, TagGroupDbmSession)
	return fields
}

func (m *oracleMeasurement) getDbmConnectionFields() map[string]interface{} {
	fields := make(map[string]interface{})
	fields["connection_count"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.NCount,
		Desc:     "Number of user connections in this dimension group",
	}

	m.addTaggedbyToFields(fields, TagGroupDbmConnection)
	return fields
}

func (m *oracleMeasurement) addTaggedbyToFields(fields map[string]interface{}, tagGroup string) {
	var tags map[string]interface{}

	// Select corresponding getTags function based on tagGroup
	switch tagGroup {
	case TagGroupProcess:
		tags = m.getProcessTags()
	case TagGroupTablespace:
		tags = m.getTablespaceTags()
	case TagGroupSystem:
		tags = m.getSystemTags()
	case TagGroupLockedSession:
		tags = m.getLockedSessionTags()
	case TagGroupWaitingEvent:
		tags = m.getWaitingEventTags()
	case TagGroupDbmMetric:
		tags = m.getDbmMetricTags()
	case TagGroupDbmSession:
		tags = m.getDbmSessionTags()
	case TagGroupDbmConnection:
		tags = m.getDbmConnectionTags()
	default:
		return
	}

	// Extract tag keys
	taggedBy := make([]string, 0, len(tags))
	for tag := range tags {
		taggedBy = append(taggedBy, tag)
	}

	// Add Taggedby to each field
	for _, field := range fields {
		if fieldInfo, ok := field.(*inputs.FieldInfo); ok {
			fieldInfo.Taggedby = taggedBy
		}
	}
}

func mergeMaps(fieldMaps ...map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for _, fields := range fieldMaps {
		for k, v := range fields {
			result[k] = v
		}
	}
	return result
}
