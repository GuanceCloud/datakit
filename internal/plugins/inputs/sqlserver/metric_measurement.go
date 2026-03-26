// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//nolint:lll // Metric descriptions are intentionally long for clarity
package sqlserver

import (
	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

const (
	TagGroupCommon         = "common"
	TagGroupServer         = "server"
	TagGroupPerformance    = "performance"
	TagGroupWaitStats      = "waitstats"
	TagGroupDatabaseIO     = "database_io"
	TagGroupSchedulers     = "schedulers"
	TagGroupVolumeSpace    = "volumespace"
	TagGroupDatabaseSize   = "database_size"
	TagGroupDatabaseFiles  = "database_files"
	TagGroupDatabaseBackup = "database_backup"
	TagGroupDbmMetric      = "dbm_metric"
	TagGroupDbmSession     = "dbm_session"
	TagGroupDbmConnection  = "dbm_connection"
)

type sqlserverMeasurement struct{}

//nolint:funlen // Info function contains all metric definitions
func (m *sqlserverMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   measurementSQLServer,
		Desc:   "Metric set including SQL Server server, performance, wait stats, database IO, schedulers, volume space, database size/files/backup, and DBM (metric/session/connection) statistics, unified in v2",
		DescZh: "指标集包含 SQL Server server、performance、wait stats、database IO、schedulers、volume space、database size/files/backup 和 DBM (metric/session/connection) 相关指标，v2 版本统一",
		Cat:    point.Metric,
		Tags:   m.getTags(),
		Fields: m.getFields(),
	}
}

func (m *sqlserverMeasurement) getTags() map[string]interface{} {
	return mergeMaps(
		m.getCommonTags(),
		m.getServerTags(),
		m.getPerformanceTags(),
		m.getWaitStatsTags(),
		m.getDatabaseIOTags(),
		m.getSchedulersTags(),
		m.getVolumeSpaceTags(),
		m.getDatabaseSizeTags(),
		m.getDatabaseFilesTags(),
		m.getDatabaseBackupTags(),
		m.getDbmMetricTags(),
		m.getDbmSessionTags(),
		m.getDbmConnectionTags(),
	)
}

func (m *sqlserverMeasurement) getCommonTags() map[string]interface{} {
	tags := make(map[string]interface{})
	tags["sqlserver_host"] = &inputs.TagInfo{Desc: "Host name which installed SQLServer"}
	tags["database_instance"] = &inputs.TagInfo{Desc: "SQL Server instance identifier derived from server name"}
	tags["server"] = &inputs.TagInfo{Desc: "The address of the server. The value is `host:port`"}
	return tags
}

func (m *sqlserverMeasurement) getServerTags() map[string]interface{} {
	return make(map[string]interface{})
}

func (m *sqlserverMeasurement) getPerformanceTags() map[string]interface{} {
	tags := make(map[string]interface{})
	tags["object_name"] = &inputs.TagInfo{Desc: "Category to which this counter belongs."}
	tags["counter_name"] = &inputs.TagInfo{Desc: "Name of the counter. To get more information about a counter, this is the name of the topic to select from the list of counters in Use SQL Server Objects."}
	tags["counter_type"] = &inputs.TagInfo{Desc: "Type of the counter"}
	tags["counter_instance"] = &inputs.TagInfo{Desc: "Name of the specific instance of the counter, for example a database, process, wait type, or resource pool."}
	return tags
}

func (m *sqlserverMeasurement) getWaitStatsTags() map[string]interface{} {
	tags := make(map[string]interface{})
	tags["wait_type"] = &inputs.TagInfo{Desc: "Name of the wait type. For more information, see Types of Waits, later in this topic"}
	tags["wait_category"] = &inputs.TagInfo{Desc: "Wait category info (e.g., Other Disk IO, Network IO, Parallelism, SQL CLR, Service Broker, etc.)"}
	return tags
}

func (m *sqlserverMeasurement) getDatabaseIOTags() map[string]interface{} {
	tags := make(map[string]interface{})
	tags["database_name"] = &inputs.TagInfo{Desc: "Database name"}
	tags["file_type"] = &inputs.TagInfo{Desc: "Description of the file type, `ROWS/LOG/FILESTREAM/FULLTEXT` (Full-text catalogs earlier than SQL Server 2008.)"}
	tags["logical_filename"] = &inputs.TagInfo{Desc: "Logical name of the file in the database"}
	tags["physical_filename"] = &inputs.TagInfo{Desc: "Operating-system file name."}
	return tags
}

func (m *sqlserverMeasurement) getSchedulersTags() map[string]interface{} {
	tags := make(map[string]interface{})
	tags["cpu_id"] = &inputs.TagInfo{Desc: "CPU ID assigned to the scheduler."}
	tags["scheduler_id"] = &inputs.TagInfo{Desc: "ID of the scheduler. All schedulers that are used to run regular queries have ID numbers less than 1048576. Those schedulers that have IDs greater than or equal to 1048576 are used internally by SQL Server, such as the dedicated administrator connection scheduler. Is not nullable."}
	return tags
}

func (m *sqlserverMeasurement) getVolumeSpaceTags() map[string]interface{} {
	tags := make(map[string]interface{})
	tags["volume_mount_point"] = &inputs.TagInfo{Desc: "Mount point at which the volume is rooted. Can return an empty string. Returns null on Linux operating system."}
	return tags
}

func (m *sqlserverMeasurement) getDatabaseSizeTags() map[string]interface{} {
	tags := make(map[string]interface{})
	tags["database_name"] = &inputs.TagInfo{Desc: "Name of the database"}
	return tags
}

func (m *sqlserverMeasurement) getDatabaseFilesTags() map[string]interface{} {
	tags := make(map[string]interface{})
	tags["database_name"] = &inputs.TagInfo{Desc: "Database name"}
	tags["state"] = &inputs.TagInfo{Desc: "Database file state: 0 = Online, 1 = Restoring, 2 = Recovering, 3 = Recovery_Pending, 4 = Suspect, 5 = Unknown, 6 = Offline, 7 = Defunct"}
	tags["physical_name"] = &inputs.TagInfo{Desc: "Operating-system file name"}
	tags["state_desc"] = &inputs.TagInfo{Desc: "Description of the file state"}
	tags["file_id"] = &inputs.TagInfo{Desc: "ID of the file within database"}
	tags["file_type_code"] = &inputs.TagInfo{Desc: "File type code: 0 = Rows, 1 = Log, 2 = File-Stream, 3 = Identified for informational purposes only, 4 = Full-text"}
	return tags
}

func (m *sqlserverMeasurement) getDatabaseBackupTags() map[string]interface{} {
	tags := make(map[string]interface{})
	tags["database_name"] = &inputs.TagInfo{Desc: "Database name"}
	return tags
}

func (m *sqlserverMeasurement) getDbmMetricTags() map[string]interface{} {
	tags := make(map[string]interface{})
	tags["database_name"] = &inputs.TagInfo{Desc: "The name of the database"}
	tags["schema_name"] = &inputs.TagInfo{Desc: "The schema name of the stored procedure (if applicable)"}
	tags["procedure_name"] = &inputs.TagInfo{Desc: "The name of the stored procedure in the format 'schema_name.procedure_name' (if applicable)"}
	tags["query_hash"] = &inputs.TagInfo{Desc: "The binary hash value of the query generated by SQL Server"}
	tags["query_plan_hash"] = &inputs.TagInfo{Desc: "The binary hash value of the query execution plan generated by SQL Server"}
	tags["query_signature"] = &inputs.TagInfo{Desc: "Hash signature generated from database_name:procedure_name:query_hash to link metrics and objects"}
	return tags
}

func (m *sqlserverMeasurement) getDbmSessionTags() map[string]interface{} {
	tags := make(map[string]interface{})
	tags["database_name"] = &inputs.TagInfo{Desc: "The name of the database"}
	tags["user_name"] = &inputs.TagInfo{Desc: "The name of the database user"}
	tags["session_status"] = &inputs.TagInfo{Desc: "Session status: active (has active request), idle (sleeping session), blocked (being blocked)"}
	tags["wait_group"] = &inputs.TagInfo{Desc: "Datakit unified wait group: Lock, I/O, Concurrency, Memory, Network, CPU, Commit/Log, Other."}
	return tags
}

func (m *sqlserverMeasurement) getDbmConnectionTags() map[string]interface{} {
	tags := make(map[string]interface{})
	tags["database_name"] = &inputs.TagInfo{Desc: "The name of the database"}
	tags["user_name"] = &inputs.TagInfo{Desc: "The name of the database user"}
	tags["connection_status"] = &inputs.TagInfo{Desc: "Connection status from SQL Server sys.dm_exec_sessions: running, sleeping, etc."}
	return tags
}

func (m *sqlserverMeasurement) getFields() map[string]interface{} {
	return mergeMaps(
		m.getServerFields(),
		m.getPerformanceFields(),
		m.getWaitStatsFields(),
		m.getDatabaseIOFields(),
		m.getSchedulersFields(),
		m.getVolumeSpaceFields(),
		m.getDatabaseSizeFields(),
		m.getDatabaseFilesFields(),
		m.getDatabaseBackupFields(),
		m.getDbmMetricFields(),
		m.getDbmSessionFields(),
		m.getDbmConnectionFields(),
	)
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

//nolint:funlen
func (m *sqlserverMeasurement) getServerFields() map[string]interface{} {
	fields := make(map[string]interface{})
	fields["cpu_count"] = newCountFieldInfo("Specifies the number of logical CPUs on the system. Not nullable")
	fields["uptime"] = newTimeFieldInfo("Total time elapsed since the last computer restart")
	fields["committed_memory"] = newByteFieldInfo("The amount of memory committed to the memory manager. Version > 2008")
	fields["physical_memory"] = newByteFieldInfo("Total physical memory on the machine. Version > 2008")
	fields["virtual_memory"] = newByteFieldInfo("Amount of virtual memory available to the process in user mode. Version > 2008")
	fields["target_memory"] = newByteFieldInfo("Amount of memory that can be consumed by the memory manager. When this value is larger than the committed memory, then the memory manager will try to obtain more memory. When it is smaller, the memory manager will try to shrink the amount of memory committed. Version > 2008")
	fields["db_online"] = newCountFieldInfo("Num of database state in online")
	fields["db_offline"] = newCountFieldInfo("Num of database state in offline")
	fields["db_recovering"] = newCountFieldInfo("Num of database state in recovering")
	fields["db_recovery_pending"] = newCountFieldInfo("Num of database state in recovery_pending")
	fields["db_restoring"] = newCountFieldInfo("Num of database state in restoring")
	fields["db_suspect"] = newCountFieldInfo("Num of database state in suspect")
	fields["server_memory"] = newByteFieldInfo("Memory used")

	m.addTaggedbyToFields(fields, TagGroupServer)
	return fields
}

//nolint:funlen
func (m *sqlserverMeasurement) getPerformanceFields() map[string]interface{} {
	fields := make(map[string]interface{})
	fields["cntr_value"] = newCountFieldInfo("Current value of the counter")
	fields["processes_blocked"] = newCountFieldInfo("The number of processes blocked.")
	fields["page_splits"] = newCountFieldInfo("The number of page splits per second.")
	fields["full_scans"] = newCountFieldInfo("Number of unrestricted full scans per second. These can be either base-table or full-index scans.")
	fields["memory_grants_pending"] = newCountFieldInfo("Specifies the total number of processes waiting for a workspace memory grant.")
	fields["total_server_memory"] = newIntKByteFieldInfo("Specifies the amount of memory the server has committed using the memory manager.")
	fields["sql_cache_memory"] = newIntKByteFieldInfo("Specifies the amount of memory the server is using for the dynamic SQL cache.")
	fields["memory_grants_outstanding"] = newCountFieldInfo("Specifies the total number of processes that have successfully acquired a workspace memory grant.")
	fields["database_cache_memory"] = newIntKByteFieldInfo("Specifies the amount of memory the server is currently using for the database pages cache.")
	fields["connection_memory"] = newIntKByteFieldInfo("Specifies the total amount of dynamic memory the server is using for maintaining connections.")
	fields["optimizer_memory"] = newIntKByteFieldInfo("Specifies the total amount of dynamic memory the server is using for query optimization.")
	fields["granted_workspace_memory"] = newIntKByteFieldInfo("Specifies the total amount of memory currently granted to executing processes, such as hash, sort, bulk copy, and index creation operations.")
	fields["lock_memory"] = newIntKByteFieldInfo("Specifies the total amount of dynamic memory the server is using for locks.")
	fields["stolen_server_memory"] = newIntKByteFieldInfo("Specifies the amount of memory the server is using for purposes other than database pages.")
	fields["log_pool_memory"] = newIntKByteFieldInfo("Total amount of dynamic memory the server is using for Log Pool.")
	fields["buffer_cache_hit_ratio"] = newPercentFieldInfo("The ratio of data pages found and read from the buffer cache over all data page requests.")
	fields["page_life_expectancy"] = newTimeFieldInfo("Duration that a page resides in the buffer pool.")
	fields["page_reads"] = newCountFieldInfo("Indicates the number of physical database page reads that are issued per second. This statistic displays the total number of physical page reads across all databases.")
	fields["page_writes"] = newCountFieldInfo("Indicates the number of physical database page writes that are issued per second.")
	fields["checkpoint_pages"] = newCountFieldInfo("The number of pages flushed to disk per second by a checkpoint or other operation that require all dirty pages to be flushed.")
	fields["auto_param_attempts"] = newCountFieldInfo("Number of auto-parameterization attempts per second.")
	fields["failed_auto_params"] = newCountFieldInfo("Number of failed auto-parameterization attempts per second.")
	fields["safe_auto_params"] = newCountFieldInfo("Number of safe auto-parameterization attempts per second.")
	fields["batch_requests"] = newCountFieldInfo("The number of batch requests per second.")
	fields["sql_compilations"] = newCountFieldInfo("The number of SQL compilations per second.")
	fields["sql_re_compilations"] = newCountFieldInfo("The number of SQL re-compilations per second.")
	fields["lock_waits"] = newCountFieldInfo("The number of times per second that SQL Server is unable to retain a lock right away for a resource.")
	fields["latch_waits"] = newCountFieldInfo("Number of latch requests that could not be granted immediately.")
	fields["deadlocks"] = newCountFieldInfo("Number of lock requests per second that resulted in a deadlock.")
	fields["cache_object_counts"] = newCountFieldInfo("Number of cache objects in the cache.")
	fields["cache_pages"] = newCountFieldInfo("Number of 8-kilobyte (KB) pages used by cache objects.")
	fields["transaction_delay"] = newCountFieldInfo("Total delay in waiting for unterminated commit acknowledgment for all the current transactions, in milliseconds.")
	fields["flow_control"] = newCountFieldInfo("Number of times flow-control initiated in the last second. Flow Control Time (ms/sec) divided by Flow Control/sec is the average time per wait.")
	fields["version_store_size"] = newIntKByteFieldInfo("The size of the version store in tempdb.")
	fields["version_cleanup_rate"] = newIntKByteFieldInfo("The cleanup rate of the version store in tempdb.")
	fields["version_generation_rate"] = newIntKByteFieldInfo("The generation rate of the version store in tempdb.")
	fields["longest_transaction_running_time"] = newTimeFieldInfo("The time (in seconds) that the oldest active transaction has been running. Only works if database is under read committed snapshot isolation level.")
	fields["backup_restore_throughput"] = newCountFieldInfo("Read/write throughput for backup and restore operations of a database per second.")
	fields["log_bytes_flushed"] = newByteFieldInfo("Total number of log bytes flushed.")
	fields["log_flushes"] = newCountFieldInfo("Number of log flushes per second.")
	fields["log_flush_wait_time"] = newTimeFieldInfo("Total wait time (in milliseconds) to flush the log. On an Always On secondary database, this value indicates the wait time for log records to be hardened to disk.")
	fields["transactions"] = newCountFieldInfo("Number of transactions started for the SQL Server instance per second.")
	fields["write_transactions"] = newCountFieldInfo("Number of transactions that wrote to all databases on the SQL Server instance and committed, in the last second.")
	fields["active_transactions"] = newCountFieldInfo("Number of active transactions across all databases on the SQL Server instance.")
	fields["user_connections"] = newCountFieldInfo("Number of user connections.")

	m.addTaggedbyToFields(fields, TagGroupPerformance)
	return fields
}

func (m *sqlserverMeasurement) getWaitStatsFields() map[string]interface{} {
	fields := make(map[string]interface{})
	fields["max_wait_time_ms"] = newTimeFieldInfo("Maximum wait time on this wait type.")
	fields["wait_time_ms"] = newTimeFieldInfo("Total wait time for this wait type in milliseconds. This time is inclusive of signal_wait_time_ms")
	fields["signal_wait_time_ms"] = newTimeFieldInfo("Difference between the time that the waiting thread was signaled and when it started running")
	fields["resource_wait_ms"] = newTimeFieldInfo("wait_time_ms-signal_wait_time_ms")
	fields["waiting_tasks_count"] = newCountFieldInfo("Number of waits on this wait type. This counter is incremented at the start of each wait.")

	m.addTaggedbyToFields(fields, TagGroupWaitStats)
	return fields
}

func (m *sqlserverMeasurement) getDatabaseIOFields() map[string]interface{} {
	fields := make(map[string]interface{})
	fields["read_bytes"] = newByteFieldInfo("Total number of bytes read on this file")
	fields["write_bytes"] = newByteFieldInfo("Total number of bytes written to the file")
	fields["read_latency_ms"] = newTimeFieldInfo("Total time, in milliseconds, that the users waited for reads issued on the file.")
	fields["write_latency_ms"] = newTimeFieldInfo("Total time, in milliseconds, that users waited for writes to be completed on the file")
	fields["reads"] = newCountFieldInfo("Number of reads issued on the file.")
	fields["writes"] = newCountFieldInfo("Number of writes issued on the file.")
	fields["rg_read_stall_ms"] = newTimeFieldInfo("Does not apply to:: SQL Server 2008 through SQL Server 2012 (11.x).Total IO latency introduced by IO resource governance for reads")
	fields["rg_write_stall_ms"] = newTimeFieldInfo("Does not apply to:: SQL Server 2008 through SQL Server 2012 (11.x).Total IO latency introduced by IO resource governance for writes. Is not nullable.")

	m.addTaggedbyToFields(fields, TagGroupDatabaseIO)
	return fields
}

func (m *sqlserverMeasurement) getSchedulersFields() map[string]interface{} {
	fields := make(map[string]interface{})
	fields["active_workers_count"] = newCountFieldInfo("Number of workers that are active. An active worker is never preemptive, must have an associated task, and is either running, runnable, or suspended. Is not nullable.")
	fields["context_switches_count"] = newCountFieldInfo("Number of context switches that have occurred on this scheduler")
	fields["current_tasks_count"] = newCountFieldInfo("Number of current tasks that are associated with this scheduler.")
	fields["current_workers_count"] = newCountFieldInfo("Number of workers that are associated with this scheduler. This count includes workers that are not assigned any task. Is not nullable.")
	fields["is_idle"] = newBoolFieldInfo("Scheduler is idle. No workers are currently running")
	fields["is_online"] = newBoolFieldInfo("If SQL Server is configured to use only some of the available processors on the server, this configuration can mean that some schedulers are mapped to processors that are not in the affinity mask. If that is the case, this column returns 0. This value means that the scheduler is not being used to process queries or batches.")
	fields["load_factor"] = newCountFieldInfo("Internal value that indicates the perceived load on this scheduler")
	fields["pending_disk_io_count"] = newCountFieldInfo("Number of pending I/Os that are waiting to be completed.")
	fields["preemptive_switches_count"] = newCountFieldInfo("Number of times that workers on this scheduler have switched to the preemptive mode")
	fields["runnable_tasks_count"] = newCountFieldInfo("Number of workers, with tasks assigned to them, that are waiting to be scheduled on the runnable queue.")
	fields["total_cpu_usage_ms"] = newTimeFieldInfo("Applies to: SQL Server 2016 (13.x) and laterTotal CPU consumed by this scheduler as reported by non-preemptive workers.")
	fields["total_scheduler_delay_ms"] = newTimeFieldInfo("Applies to: SQL Server 2016 (13.x) and laterThe time between one worker switching out and another one switching in")
	fields["work_queue_count"] = newCountFieldInfo("Number of tasks in the pending queue. These tasks are waiting for a worker to pick them up")
	fields["yield_count"] = newCountFieldInfo("Internal value that is used to indicate progress on this scheduler. This value is used by the Scheduler Monitor to determine whether a worker on the scheduler is not yielding to other workers on time.")

	m.addTaggedbyToFields(fields, TagGroupSchedulers)
	return fields
}

func (m *sqlserverMeasurement) getVolumeSpaceFields() map[string]interface{} {
	fields := make(map[string]interface{})
	fields["volume_available_space_bytes"] = newByteFieldInfo("Available free space on the volume")
	fields["volume_total_space_bytes"] = newByteFieldInfo("Total size in bytes of the volume")
	fields["volume_used_space_bytes"] = newByteFieldInfo("Used size in bytes of the volume")

	m.addTaggedbyToFields(fields, TagGroupVolumeSpace)
	return fields
}

func (m *sqlserverMeasurement) getDatabaseSizeFields() map[string]interface{} {
	fields := make(map[string]interface{})
	fields["data_size"] = newMByteFieldInfo("The size of file of Rows")
	fields["log_size"] = newMByteFieldInfo("The size of file of Log")

	m.addTaggedbyToFields(fields, TagGroupDatabaseSize)
	return fields
}

func (m *sqlserverMeasurement) getDatabaseFilesFields() map[string]interface{} {
	fields := make(map[string]interface{})
	fields["size"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.SizeKB,
		Desc:     "Current size of the database file",
	}

	m.addTaggedbyToFields(fields, TagGroupDatabaseFiles)
	return fields
}

func (m *sqlserverMeasurement) getDatabaseBackupFields() map[string]interface{} {
	fields := make(map[string]interface{})
	fields["backup_count"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.Count,
		Desc:     "The total count of successful backups made for a database",
	}

	m.addTaggedbyToFields(fields, TagGroupDatabaseBackup)
	return fields
}

//nolint:funlen
func (m *sqlserverMeasurement) getDbmMetricFields() map[string]interface{} {
	fields := make(map[string]interface{})
	// Total fields (cumulative values from SQL Server)
	fields["total_execution_count"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.NCount,
		Desc:     "The total number of query executions (cumulative value from SQL Server).",
	}
	fields["total_worker_time"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.DurationUS,
		Desc:     "The total CPU time (microseconds) consumed by query executions (cumulative value from SQL Server).",
	}
	fields["total_physical_reads"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.NCount,
		Desc:     "The total number of physical reads (cumulative value from SQL Server).",
	}
	fields["total_logical_writes"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.NCount,
		Desc:     "The total number of logical writes (cumulative value from SQL Server).",
	}
	fields["total_logical_reads"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.NCount,
		Desc:     "The total number of logical reads (cumulative value from SQL Server).",
	}
	fields["total_elapsed_time"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.DurationUS,
		Desc:     "The total elapsed time (microseconds) for query executions (cumulative value from SQL Server).",
	}
	fields["total_rows"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.NCount,
		Desc:     "The total number of rows returned (cumulative value from SQL Server).",
	}
	fields["total_clr_time"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.DurationUS,
		Desc:     "The total time (microseconds) spent inside CLR (cumulative value from SQL Server).",
	}
	fields["total_dop"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.NCount,
		Desc:     "The total degree of parallelism used (cumulative value from SQL Server).",
	}
	fields["total_grant_kb"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.SizeKB,
		Desc:     "The total amount of memory granted (KB) (cumulative value from SQL Server).",
	}
	fields["total_used_grant_kb"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.SizeKB,
		Desc:     "The total amount of granted memory actually used (KB) (cumulative value from SQL Server).",
	}
	fields["total_ideal_grant_kb"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.SizeKB,
		Desc:     "The total ideal amount of memory (KB) that should have been granted (cumulative value from SQL Server).",
	}
	fields["total_reserved_threads"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.NCount,
		Desc:     "The total number of reserved threads used (cumulative value from SQL Server).",
	}
	fields["total_used_threads"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.NCount,
		Desc:     "The total number of used threads (cumulative value from SQL Server).",
	}
	fields["total_columnstore_segment_reads"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.NCount,
		Desc:     "The total number of columnstore segments read (cumulative value from SQL Server).",
	}
	fields["total_columnstore_segment_skips"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.NCount,
		Desc:     "The total number of columnstore segments skipped (cumulative value from SQL Server).",
	}
	fields["total_spills"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.NCount,
		Desc:     "The total number of pages spilled to disk (cumulative value from SQL Server).",
	}
	// Delta fields (change between collection intervals)
	fields["delta_execution_count"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Count,
		Unit:     inputs.NCount,
		Desc:     "The number of times the query was executed during the collection interval (delta value).",
	}
	fields["delta_worker_time"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Count,
		Unit:     inputs.DurationUS,
		Desc:     "The CPU time (microseconds) consumed by query executions during the collection interval (delta value).",
	}
	fields["delta_physical_reads"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Count,
		Unit:     inputs.NCount,
		Desc:     "The number of physical reads performed during the collection interval (delta value).",
	}
	fields["delta_logical_writes"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Count,
		Unit:     inputs.NCount,
		Desc:     "The number of logical writes performed during the collection interval (delta value).",
	}
	fields["delta_query_logical_reads"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Count,
		Unit:     inputs.NCount,
		Desc:     "The number of logical reads performed during the collection interval (delta value).",
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
		Desc:     "The average elapsed time (microseconds) per query execution during the collection interval (calculated from delta_elapsed_time / delta_execution_count).",
	}
	fields["delta_rows"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Count,
		Unit:     inputs.NCount,
		Desc:     "The number of rows returned during the collection interval (delta value).",
	}
	fields["delta_clr_time"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Count,
		Unit:     inputs.DurationUS,
		Desc:     "The time (microseconds) spent inside CLR during the collection interval (delta value).",
	}
	fields["delta_dop"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Count,
		Unit:     inputs.NCount,
		Desc:     "The degree of parallelism used during the collection interval (delta value).",
	}
	fields["delta_grant_kb"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Count,
		Unit:     inputs.SizeKB,
		Desc:     "The amount of memory granted (KB) during the collection interval (delta value).",
	}
	fields["delta_used_grant_kb"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Count,
		Unit:     inputs.SizeKB,
		Desc:     "The amount of granted memory actually used (KB) during the collection interval (delta value).",
	}
	fields["delta_ideal_grant_kb"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Count,
		Unit:     inputs.SizeKB,
		Desc:     "The ideal amount of memory (KB) that should have been granted during the collection interval (delta value).",
	}
	fields["delta_reserved_threads"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Count,
		Unit:     inputs.NCount,
		Desc:     "The number of reserved threads used during the collection interval (delta value).",
	}
	fields["delta_used_threads"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Count,
		Unit:     inputs.NCount,
		Desc:     "The number of used threads during the collection interval (delta value).",
	}
	fields["delta_columnstore_segment_reads"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Count,
		Unit:     inputs.NCount,
		Desc:     "The number of columnstore segments read during the collection interval (delta value).",
	}
	fields["delta_columnstore_segment_skips"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Count,
		Unit:     inputs.NCount,
		Desc:     "The number of columnstore segments skipped during the collection interval (delta value).",
	}
	fields["delta_spills"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Count,
		Unit:     inputs.NCount,
		Desc:     "The number of pages spilled to disk during the collection interval (delta value).",
	}

	m.addTaggedbyToFields(fields, TagGroupDbmMetric)
	return fields
}

func (m *sqlserverMeasurement) getDbmSessionFields() map[string]interface{} {
	fields := make(map[string]interface{})
	fields["session_group_count"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.NCount,
		Desc:     "Number of sessions in this dimension group",
	}
	fields["session_total_cpu_time"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.DurationMS,
		Desc:     "Total CPU time consumed by active sessions in this group (milliseconds). Note: Only active sessions have this data.",
	}
	fields["session_total_wait_time"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.DurationMS,
		Desc:     "Total wait time for active sessions in this group (milliseconds). Note: Only active sessions have this data.",
	}
	fields["session_total_elapsed_time"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.DurationMS,
		Desc:     "Total elapsed time for active sessions in this group (milliseconds). Note: Only active sessions have this data.",
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
	fields["session_total_reads"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.NCount,
		Desc:     "Total number of reads by active sessions in this group. Note: Only active sessions have this data.",
	}
	fields["session_total_writes"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.NCount,
		Desc:     "Total number of writes by active sessions in this group. Note: Only active sessions have this data.",
	}
	fields["session_total_logical_reads"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.NCount,
		Desc:     "Total number of logical reads by active sessions in this group. Note: Only active sessions have this data.",
	}
	fields["session_total_open_transaction_count"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.NCount,
		Desc:     "Total number of open transactions across active sessions in this group. Note: Only active sessions have this data.",
	}
	fields["session_max_elapsed_time"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.DurationMS,
		Desc:     "Maximum elapsed time among active sessions in this group (milliseconds). Note: Only active sessions have this data.",
	}

	m.addTaggedbyToFields(fields, TagGroupDbmSession)
	return fields
}

func (m *sqlserverMeasurement) getDbmConnectionFields() map[string]interface{} {
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

func (m *sqlserverMeasurement) addTaggedbyToFields(fields map[string]interface{}, tagGroup string) {
	var tags map[string]interface{}

	// Select corresponding getTags function based on tagGroup
	switch tagGroup {
	case TagGroupServer:
		tags = m.getServerTags()
	case TagGroupPerformance:
		tags = m.getPerformanceTags()
	case TagGroupWaitStats:
		tags = m.getWaitStatsTags()
	case TagGroupDatabaseIO:
		tags = m.getDatabaseIOTags()
	case TagGroupSchedulers:
		tags = m.getSchedulersTags()
	case TagGroupVolumeSpace:
		tags = m.getVolumeSpaceTags()
	case TagGroupDatabaseSize:
		tags = m.getDatabaseSizeTags()
	case TagGroupDatabaseFiles:
		tags = m.getDatabaseFilesTags()
	case TagGroupDatabaseBackup:
		tags = m.getDatabaseBackupTags()
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
