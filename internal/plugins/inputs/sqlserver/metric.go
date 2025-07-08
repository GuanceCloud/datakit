// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package sqlserver

import (
	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

type Measurement struct {
	name     string
	tags     map[string]string
	fields   map[string]interface{}
	election bool
}

func (m *Measurement) Point() *point.Point {
	return nil
}

func (m *Measurement) Info() *inputs.MeasurementInfo {
	return nil
}

type MetricMeasurment struct {
	Measurement
}

// Point implement MeasurementV2.
func (m *MetricMeasurment) Point() *point.Point {
	opts := point.DefaultMetricOptions()

	if m.election {
		opts = append(opts, point.WithExtraTags(datakit.GlobalElectionTags()))
	}

	return point.NewPointV2(m.name,
		append(point.NewTags(m.tags), point.NewKVs(m.fields)...),
		opts...)
}

type LoggingMeasurment struct {
	Measurement
}

// Point implement MeasurementV2.
func (m *LoggingMeasurment) Point() *point.Point {
	opts := point.DefaultLoggingOptions()

	if m.election {
		opts = append(opts, point.WithExtraTags(datakit.GlobalElectionTags()))
	}

	return point.NewPointV2(m.name,
		append(point.NewTags(m.tags), point.NewKVs(m.fields)...),
		opts...)
}

type SqlserverMeasurment struct {
	MetricMeasurment
}

//nolint:lll
func (m *SqlserverMeasurment) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "sqlserver",
		Cat:  point.Metric,
		Fields: map[string]interface{}{
			"cpu_count":           newCountFieldInfo("Specifies the number of logical CPUs on the system. Not nullable"),
			"uptime":              newTimeFieldInfo("Total time elapsed since the last computer restart"),
			"committed_memory":    newByteFieldInfo("The amount of memory committed to the memory manager. Version > 2008"),
			"physical_memory":     newByteFieldInfo("Total physical memory on the machine. Version > 2008"),
			"virtual_memory":      newByteFieldInfo("Amount of virtual memory available to the process in user mode. Version > 2008"),
			"target_memory":       newByteFieldInfo("Amount of memory that can be consumed by the memory manager. When this value is larger than the committed memory, then the memory manager will try to obtain more memory. When it is smaller, the memory manager will try to shrink the amount of memory committed. Version > 2008"),
			"db_online":           newCountFieldInfo("Num of database state in online"),
			"db_offline":          newCountFieldInfo("Num of database state in offline"),
			"db_recovering":       newCountFieldInfo("Num of database state in recovering"),
			"db_recovery_pending": newCountFieldInfo("Num of database state in recovery_pending"),
			"db_restoring":        newCountFieldInfo("Num of database state in restoring"),
			"db_suspect":          newCountFieldInfo("Num of database state in suspect"),
			"server_memory":       newByteFieldInfo("Memory used"),
		},
		Tags: map[string]interface{}{
			"sqlserver_host": inputs.NewTagInfo("Host name which installed SQLServer"),
			"server":         inputs.NewTagInfo("The address of the server. The value is `host:port`"),
		},
	}
}

type Performance struct {
	MetricMeasurment
}

//nolint:lll
var performanceMeasurementInfo = &inputs.MeasurementInfo{
	Name: "sqlserver_performance",
	Cat:  point.Metric,
	Desc: "performance counter maintained by the server,[detail](https://docs.microsoft.com/en-us/sql/relational-databases/system-dynamic-management-views/sys-dm-os-performance-counters-transact-sql?view=sql-server-ver15)",
	Fields: map[string]interface{}{
		"cntr_value":                       newCountFieldInfo("Current value of the counter"),
		"processes_blocked":                newCountFieldInfo("The number of processes blocked."),
		"page_splits":                      newCountFieldInfo("The number of page splits per second."),
		"full_scans":                       newCountFieldInfo("Number of unrestricted full scans per second. These can be either base-table or full-index scans."),
		"memory_grants_pending":            newCountFieldInfo("Specifies the total number of processes waiting for a workspace memory grant."),
		"total_server_memory":              newIntKByteFieldInfo("Specifies the amount of memory the server has committed using the memory manager."),
		"sql_cache_memory":                 newIntKByteFieldInfo("Specifies the amount of memory the server is using for the dynamic SQL cache."),
		"memory_grants_outstanding":        newCountFieldInfo("Specifies the total number of processes that have successfully acquired a workspace memory grant."),
		"database_cache_memory":            newIntKByteFieldInfo("Specifies the amount of memory the server is currently using for the database pages cache."),
		"connection_memory":                newIntKByteFieldInfo("Specifies the total amount of dynamic memory the server is using for maintaining connections."),
		"optimizer_memory":                 newIntKByteFieldInfo("Specifies the total amount of dynamic memory the server is using for query optimization."),
		"granted_workspace_memory":         newIntKByteFieldInfo("Specifies the total amount of memory currently granted to executing processes, such as hash, sort, bulk copy, and index creation operations."),
		"lock_memory":                      newIntKByteFieldInfo("Specifies the total amount of dynamic memory the server is using for locks."),
		"stolen_server_memory":             newIntKByteFieldInfo("Specifies the amount of memory the server is using for purposes other than database pages."),
		"log_pool_memory":                  newIntKByteFieldInfo("Total amount of dynamic memory the server is using for Log Pool."),
		"buffer_cache_hit_ratio":           newPercentFieldInfo("The ratio of data pages found and read from the buffer cache over all data page requests."),
		"page_life_expectancy":             newTimeFieldInfo("Duration that a page resides in the buffer pool."),
		"page_reads":                       newCountFieldInfo("Indicates the number of physical database page reads that are issued per second. This statistic displays the total number of physical page reads across all databases."),
		"page_writes":                      newCountFieldInfo("Indicates the number of physical database page writes that are issued per second."),
		"checkpoint_pages":                 newCountFieldInfo("The number of pages flushed to disk per second by a checkpoint or other operation that require all dirty pages to be flushed."),
		"auto_param_attempts":              newCountFieldInfo("Number of auto-parameterization attempts per second."),
		"failed_auto_params":               newCountFieldInfo("Number of failed auto-parameterization attempts per second."),
		"safe_auto_params":                 newCountFieldInfo("Number of safe auto-parameterization attempts per second."),
		"batch_requests":                   newCountFieldInfo("The number of batch requests per second."),
		"sql_compilations":                 newCountFieldInfo("The number of SQL compilations per second."),
		"sql_re_compilations":              newCountFieldInfo("The number of SQL re-compilations per second."),
		"lock_waits":                       newCountFieldInfo("The number of times per second that SQL Server is unable to retain a lock right away for a resource."),
		"latch_waits":                      newCountFieldInfo("Number of latch requests that could not be granted immediately."),
		"deadlocks":                        newCountFieldInfo("Number of lock requests per second that resulted in a deadlock."),
		"cache_object_counts":              newCountFieldInfo("Number of cache objects in the cache."),
		"cache_pages":                      newCountFieldInfo("Number of 8-kilobyte (KB) pages used by cache objects."),
		"transaction_delay":                newCountFieldInfo("Total delay in waiting for unterminated commit acknowledgment for all the current transactions, in milliseconds."),
		"flow_control":                     newCountFieldInfo("Number of times flow-control initiated in the last second. Flow Control Time (ms/sec) divided by Flow Control/sec is the average time per wait."),
		"version_store_size":               newIntKByteFieldInfo("The size of the version store in tempdb."),
		"version_cleanup_rate":             newIntKByteFieldInfo("The cleanup rate of the version store in tempdb."),
		"version_generation_rate":          newIntKByteFieldInfo("The generation rate of the version store in tempdb."),
		"longest_transaction_running_time": newTimeFieldInfo("The time (in seconds) that the oldest active transaction has been running. Only works if database is under read committed snapshot isolation level."),
		"backup_restore_throughput":        newCountFieldInfo("Read/write throughput for backup and restore operations of a database per second."),
		"log_bytes_flushed":                newByteFieldInfo("Total number of log bytes flushed."),
		"log_flushes":                      newCountFieldInfo("Number of log flushes per second."),
		"log_flush_wait_time":              newTimeFieldInfo("Total wait time (in milliseconds) to flush the log. On an Always On secondary database, this value indicates the wait time for log records to be hardened to disk."),
		"transactions":                     newCountFieldInfo("Number of transactions started for the SQL Server instance per second."),
		"write_transactions":               newCountFieldInfo("Number of transactions that wrote to all databases on the SQL Server instance and committed, in the last second."),
		"active_transactions":              newCountFieldInfo("Number of active transactions across all databases on the SQL Server instance."),
		"user_connections":                 newCountFieldInfo("Number of user connections."),
	},
	Tags: map[string]interface{}{
		"object_name":    inputs.NewTagInfo("Category to which this counter belongs."),
		"counter_name":   inputs.NewTagInfo("Name of the counter. To get more information about a counter, this is the name of the topic to select from the list of counters in Use SQL Server Objects."),
		"counter_type":   inputs.NewTagInfo("Type of the counter"),
		"instance":       inputs.NewTagInfo("Name of the specific instance of the counter"),
		"sqlserver_host": inputs.NewTagInfo("Host name which installed SQLServer"),
		"server":         inputs.NewTagInfo("The address of the server. The value is `host:port`"),
	},
}

//nolint:lll
func (m *Performance) Info() *inputs.MeasurementInfo {
	return performanceMeasurementInfo
}

type WaitStatsCategorized struct {
	MetricMeasurment
}

//nolint:lll
func (m *WaitStatsCategorized) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "sqlserver_waitstats",
		Cat:  point.Metric,
		Desc: "information about all the waits encountered by threads that executed,[detail](https://docs.microsoft.com/en-us/sql/relational-databases/system-dynamic-management-views/sys-dm-os-wait-stats-transact-sql?view=sql-server-ver15)",
		Fields: map[string]interface{}{
			"max_wait_time_ms":    newTimeFieldInfo("Maximum wait time on this wait type."),
			"wait_time_ms":        newTimeFieldInfo("Total wait time for this wait type in milliseconds. This time is inclusive of signal_wait_time_ms"),
			"signal_wait_time_ms": newTimeFieldInfo("Difference between the time that the waiting thread was signaled and when it started running"),
			"resource_wait_ms":    newTimeFieldInfo("wait_time_ms-signal_wait_time_ms"),
			"waiting_tasks_count": newCountFieldInfo("Number of waits on this wait type. This counter is incremented at the start of each wait."),
		},
		Tags: map[string]interface{}{
			"sqlserver_host": inputs.NewTagInfo("Host name which installed SQLServer"),
			"wait_type":      inputs.NewTagInfo("Name of the wait type. For more information, see Types of Waits, later in this topic"),
			"wait_category":  inputs.NewTagInfo("Wait category info"),
			"server":         inputs.NewTagInfo("The address of the server. The value is `host:port`"),
		},
	}
}

type DatabaseIO struct {
	MetricMeasurment
}

//nolint:lll
func (m *DatabaseIO) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "sqlserver_database_io",
		Cat:  point.Metric,
		Desc: "I/O statistics for data and log files,[detail](https://docs.microsoft.com/en-us/sql/relational-databases/system-dynamic-management-views/sys-dm-io-virtual-file-stats-transact-sql?view=sql-server-ver15)",
		Fields: map[string]interface{}{
			"read_bytes":        newByteFieldInfo("Total number of bytes read on this file"),
			"write_bytes":       newByteFieldInfo("Total number of bytes written to the file"),
			"read_latency_ms":   newTimeFieldInfo("Total time, in milliseconds, that the users waited for reads issued on the file."),
			"write_latency_ms":  newTimeFieldInfo("Total time, in milliseconds, that users waited for writes to be completed on the file"),
			"reads":             newCountFieldInfo("Number of reads issued on the file."),
			"writes":            newCountFieldInfo("Number of writes issued on the file."),
			"rg_read_stall_ms":  newTimeFieldInfo("Does not apply to:: SQL Server 2008 through SQL Server 2012 (11.x).Total IO latency introduced by IO resource governance for reads"),
			"rg_write_stall_ms": newTimeFieldInfo("Does not apply to:: SQL Server 2008 through SQL Server 2012 (11.x).Total IO latency introduced by IO resource governance for writes. Is not nullable."),
		},
		Tags: map[string]interface{}{
			"database_name":     inputs.NewTagInfo("Database name"),
			"file_type":         inputs.NewTagInfo("Description of the file type, `ROWS/LOG/FILESTREAM/FULLTEXT` (Full-text catalogs earlier than SQL Server 2008.)"),
			"logical_filename":  inputs.NewTagInfo("Logical name of the file in the database"),
			"physical_filename": inputs.NewTagInfo("Operating-system file name."),
			"sqlserver_host":    inputs.NewTagInfo("Host name which installed SQLServer"),
			"server":            inputs.NewTagInfo("The address of the server. The value is `host:port`"),
		},
	}
}

type Schedulers struct {
	MetricMeasurment
}

//nolint:lll
func (m *Schedulers) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "sqlserver_schedulers",
		Cat:  point.Metric,
		Desc: "One row per scheduler in SQL Server where each scheduler is mapped to an individual processor,[detail](https://docs.microsoft.com/en-us/sql/relational-databases/system-dynamic-management-views/sys-dm-os-schedulers-transact-sql?view=sql-server-ver15)",
		Fields: map[string]interface{}{
			"active_workers_count":      newCountFieldInfo("Number of workers that are active. An active worker is never preemptive, must have an associated task, and is either running, runnable, or suspended. Is not nullable."),
			"context_switches_count":    newCountFieldInfo("Number of context switches that have occurred on this scheduler"),
			"current_tasks_count":       newCountFieldInfo("Number of current tasks that are associated with this scheduler."),
			"current_workers_count":     newCountFieldInfo("Number of workers that are associated with this scheduler. This count includes workers that are not assigned any task. Is not nullable."),
			"is_idle":                   newBoolFieldInfo("Scheduler is idle. No workers are currently running"),
			"is_online":                 newBoolFieldInfo("If SQL Server is configured to use only some of the available processors on the server, this configuration can mean that some schedulers are mapped to processors that are not in the affinity mask. If that is the case, this column returns 0. This value means that the scheduler is not being used to process queries or batches."),
			"load_factor":               newCountFieldInfo("Internal value that indicates the perceived load on this scheduler"),
			"pending_disk_io_count":     newCountFieldInfo("Number of pending I/Os that are waiting to be completed."),
			"preemptive_switches_count": newCountFieldInfo("Number of times that workers on this scheduler have switched to the preemptive mode"),
			"runnable_tasks_count":      newCountFieldInfo("Number of workers, with tasks assigned to them, that are waiting to be scheduled on the runnable queue."),
			"total_cpu_usage_ms":        newTimeFieldInfo("Applies to: SQL Server 2016 (13.x) and laterTotal CPU consumed by this scheduler as reported by non-preemptive workers."),
			"total_scheduler_delay_ms":  newTimeFieldInfo("Applies to: SQL Server 2016 (13.x) and laterThe time between one worker switching out and another one switching in"),
			"work_queue_count":          newCountFieldInfo("Number of tasks in the pending queue. These tasks are waiting for a worker to pick them up"),
			"yield_count":               newCountFieldInfo("Internal value that is used to indicate progress on this scheduler. This value is used by the Scheduler Monitor to determine whether a worker on the scheduler is not yielding to other workers on time."),
		},
		Tags: map[string]interface{}{
			"cpu_id":         inputs.NewTagInfo("CPU ID assigned to the scheduler."),
			"sqlserver_host": inputs.NewTagInfo("Host name which installed SQLServer"),
			"scheduler_id":   inputs.NewTagInfo("ID of the scheduler. All schedulers that are used to run regular queries have ID numbers less than 1048576. Those schedulers that have IDs greater than or equal to 1048576 are used internally by SQL Server, such as the dedicated administrator connection scheduler. Is not nullable."),
			"server":         inputs.NewTagInfo("The address of the server. The value is `host:port`"),
		},
	}
}

type VolumeSpace struct {
	MetricMeasurment
}

//nolint:lll
func (m *VolumeSpace) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "sqlserver_volumespace",
		Desc: "The version should be greater than SQL Server 2008.",
		Cat:  point.Metric,
		Fields: map[string]interface{}{
			"volume_available_space_bytes": newByteFieldInfo("Available free space on the volume"),
			"volume_total_space_bytes":     newByteFieldInfo("Total size in bytes of the volume"),
			"volume_used_space_bytes":      newByteFieldInfo("Used size in bytes of the volume"),
		},
		Tags: map[string]interface{}{
			"sqlserver_host":     inputs.NewTagInfo("Host name which installed SQLServer"),
			"volume_mount_point": inputs.NewTagInfo("Mount point at which the volume is rooted. Can return an empty string. Returns null on Linux operating system."),
			"server":             inputs.NewTagInfo("The address of the server. The value is `host:port`"),
		},
	}
}

type LockRow struct {
	LoggingMeasurment
}

//nolint:lll
func (m *LockRow) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "sqlserver_lock_row",
		Cat:  point.Logging,
		Fields: map[string]interface{}{
			"blocking_session_id":     newCountFieldInfo("ID of the session that is blocking the request"),
			"session_id":              newCountFieldInfo("ID of the session to which this request is related"),
			"cpu_time":                newTimeFieldInfo("CPU time in milliseconds that is used by the request"),
			"logical_reads":           newCountFieldInfo("Number of logical reads that have been performed by the request"),
			"row_count":               newCountFieldInfo("Number of rows returned on the session up to this point"),
			"memory_usage":            newCountFieldInfo("Number of 8-KB pages of memory used by this session"),
			"last_request_start_time": newTimeFieldInfo("Time at which the last request on the session began, in second"),
			"last_request_end_time":   newTimeFieldInfo("Time of the last completion of a request on the session, in second"),
			"host_name":               newStringFieldInfo("Name of the client workstation that is specific to a session"),
			"login_name":              newStringFieldInfo("SQL Server login name under which the session is currently executing"),
			"session_status":          newStringFieldInfo("Status of the session"),
			"message":                 newStringFieldInfo("Text of the SQL query"),
		},
		Tags: map[string]interface{}{
			"server": inputs.NewTagInfo("The address of the server. The value is `host:port`"),
		},
	}
}

type LockTable struct {
	LoggingMeasurment
}

//nolint:lll
func (m *LockTable) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "sqlserver_lock_table",
		Cat:  point.Logging,
		Fields: map[string]interface{}{
			"request_session_id": newCountFieldInfo("Session ID that currently owns this request"),
			"object_name":        newStringFieldInfo("Name of the entity in a database with which a resource is associated"),
			"db_name":            newStringFieldInfo("Name of the database under which this resource is scoped"),
			"resource_type":      newStringFieldInfo("Represents the resource type"),
			"request_mode":       newStringFieldInfo("Mode of the request"),
			"request_status":     newStringFieldInfo("Current status of this request"),
		},
		Tags: map[string]interface{}{
			"server": inputs.NewTagInfo("The address of the server. The value is `host:port`"),
		},
	}
}

type LockDead struct {
	LoggingMeasurment
}

//nolint:lll
func (m *LockDead) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "sqlserver_lock_dead",
		Cat:  point.Logging,
		Fields: map[string]interface{}{
			"request_session_id":   newCountFieldInfo("Session ID that currently owns this request"),
			"blocking_session_id":  newCountFieldInfo("ID of the session that is blocking the request"),
			"blocking_object_name": newStringFieldInfo("Indicates the name of the object to which this partition belongs"),
			"db_name":              newStringFieldInfo("Name of the database under which this resource is scoped"),
			"resource_type":        newStringFieldInfo("Represents the resource type"),
			"request_mode":         newStringFieldInfo("Mode of the request"),
			"requesting_text":      newStringFieldInfo("Text of the SQL query which is requesting"),
			"blocking_text":        newStringFieldInfo("Text of the SQL query which is blocking"),
			"message":              newStringFieldInfo("Text of the SQL query which is blocking"),
		},
		Tags: map[string]interface{}{
			"server": inputs.NewTagInfo("The address of the server. The value is `host:port`"),
		},
	}
}

type LogicalIO struct {
	LoggingMeasurment
}

//nolint:lll
func (m *LogicalIO) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "sqlserver_logical_io",
		Cat:  point.Logging,
		Fields: map[string]interface{}{
			"avg_logical_io":       newCountFieldInfo("Average number of logical writes and logical reads"),
			"total_logical_io":     newCountFieldInfo("Total number of logical writes and logical reads"),
			"total_logical_reads":  newCountFieldInfo("Total amount of logical reads"),
			"total_logical_writes": newCountFieldInfo("Total amount of logical writes"),
			"creation_time":        newCountFieldInfo("The Unix time at which the plan was compiled, in millisecond"),
			"execution_count":      newCountFieldInfo("Number of times that the plan has been executed since it was last compiled"),
			"last_execution_time":  newCountFieldInfo("Last time at which the plan started executing, unix time in millisecond"),
		},
		Tags: map[string]interface{}{
			"message": inputs.NewTagInfo("Text of the SQL query"),
			"server":  inputs.NewTagInfo("The address of the server. The value is `host:port`"),
		},
	}
}

type WorkerTime struct {
	LoggingMeasurment
}

//nolint:lll
func (m *WorkerTime) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "sqlserver_worker_time",
		Cat:  point.Logging,
		Fields: map[string]interface{}{
			"creation_time":       newCountFieldInfo("The Unix time at which the plan was compiled, in millisecond"),
			"execution_count":     newCountFieldInfo("Number of times that the plan has been executed since it was last compiled"),
			"last_execution_time": newCountFieldInfo("Last time at which the plan started executing, unix time in millisecond"),
			"total_worker_time":   newCountFieldInfo("Total amount of CPU time, reported in milliseconds"),
			"avg_worker_time":     newCountFieldInfo("Average amount of CPU time, reported in milliseconds"),
		},
		Tags: map[string]interface{}{
			"message": inputs.NewTagInfo("Text of the SQL query"),
			"server":  inputs.NewTagInfo("The address of the server. The value is `host:port`"),
		},
	}
}

type DatabaseSize struct {
	MetricMeasurment
}

//nolint:lll
func (m *DatabaseSize) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "sqlserver_database_size",
		Cat:  point.Metric,
		Fields: map[string]interface{}{
			"data_size": newKByteFieldInfo("The size of file of Rows"),
			"log_size":  newKByteFieldInfo("The size of file of Log"),
		},
		Tags: map[string]interface{}{
			"database_name": inputs.NewTagInfo("Name of the database"),
			"server":        inputs.NewTagInfo("The address of the server. The value is `host:port`"),
		},
	}
}

type DatabaseFilesMeasurement struct {
	MetricMeasurment
}

//nolint:lll
var DatabaseFilesMeasurementInfo = &inputs.MeasurementInfo{
	Name: "sqlserver_database_files",
	Cat:  point.Metric,
	Fields: map[string]interface{}{
		"size": &inputs.FieldInfo{
			DataType: inputs.Int,
			Type:     inputs.Gauge,
			Unit:     inputs.SizeKB,
			Desc:     "Current size of the database file",
		},
	},
	Tags: map[string]interface{}{
		"database":      inputs.NewTagInfo("Database name"),
		"state":         inputs.NewTagInfo("Database file state: 0 = Online, 1 = Restoring, 2 = Recovering, 3 = Recovery_Pending, 4 = Suspect, 5 = Unknown, 6 = Offline, 7 = Defunct"),
		"physical_name": inputs.NewTagInfo("Operating-system file name"),
		"state_desc":    inputs.NewTagInfo("Description of the file state"),
		"file_id":       inputs.NewTagInfo("ID of the file within database"),
		"file_type":     inputs.NewTagInfo("File type: 0 = Rows, 1 = Log, 2 = File-Stream, 3 = Identified for informational purposes only, 4 = Full-text"),
		"server":        inputs.NewTagInfo("The address of the server. The value is `host:port`"),
	},
}

func (m *DatabaseFilesMeasurement) Info() *inputs.MeasurementInfo {
	return DatabaseFilesMeasurementInfo
}

type DatabaseBackupMeasurement struct {
	MetricMeasurment
}

//nolint:lll
func (m *DatabaseBackupMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "sqlserver_database_backup",
		Cat:  point.Metric,
		Fields: map[string]interface{}{
			"backup_count": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.Count,
				Desc:     "The total count of successful backups made for a database",
			},
		},
		Tags: map[string]interface{}{
			"database": inputs.NewTagInfo("Database name"),
			"server":   inputs.NewTagInfo("The address of the server. The value is `host:port`"),
		},
	}
}
