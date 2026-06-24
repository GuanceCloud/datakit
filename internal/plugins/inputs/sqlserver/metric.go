// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package sqlserver

import (
	"github.com/GuanceCloud/cliutils/point"
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

	return point.NewPoint(m.name,
		append(point.NewTags(m.tags), point.NewKVs(m.fields)...),
		opts...)
}

type LoggingMeasurment struct {
	Measurement
}

// Point implement MeasurementV2.
func (m *LoggingMeasurment) Point() *point.Point {
	opts := point.DefaultLoggingOptions()

	return point.NewPoint(m.name,
		append(point.NewTags(m.tags), point.NewKVs(m.fields)...),
		opts...)
}

type SqlserverMeasurment struct {
	MetricMeasurment
}

//nolint:lll
func (m *SqlserverMeasurment) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   "sqlserver",
		Desc:   "SQL Server instance metrics collected from system views and server properties.",
		DescZh: "从系统视图和服务器属性采集的 SQL Server 实例指标。",
		Cat:    point.Metric,
		Fields: map[string]interface{}{
			"cpu_count":           &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Specifies the number of logical CPUs on the system. Not nullable"},
			"uptime":              &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: "Total time elapsed since the last computer restart"},
			"committed_memory":    &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "The amount of memory committed to the memory manager. Version > 2008"},
			"physical_memory":     &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "Total physical memory on the machine. Version > 2008"},
			"virtual_memory":      &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "Amount of virtual memory available to the process in user mode. Version > 2008"},
			"target_memory":       &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "Amount of memory that can be consumed by the memory manager. When this value is larger than the committed memory, then the memory manager will try to obtain more memory. When it is smaller, the memory manager will try to shrink the amount of memory committed. Version > 2008"},
			"db_online":           &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Num of database state in online"},
			"db_offline":          &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Num of database state in offline"},
			"db_recovering":       &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Num of database state in recovering"},
			"db_recovery_pending": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Num of database state in recovery_pending"},
			"db_restoring":        &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Num of database state in restoring"},
			"db_suspect":          &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Num of database state in suspect"},
			"server_memory":       &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "Memory used"},
		},
		Tags: map[string]interface{}{
			"sqlserver_host":    inputs.NewTagInfo("Host name which installed SQLServer"),
			"database_instance": inputs.NewTagInfo("SQL Server instance identifier from configured tag or SQL Server server name."),
			"server":            inputs.NewTagInfo("The address of the server. The value is `host:port`"),
		},
	}
}

type Performance struct {
	MetricMeasurment
}

//nolint:lll
var performanceMeasurementInfo = &inputs.MeasurementInfo{
	Name:   "sqlserver_performance",
	Cat:    point.Metric,
	Desc:   "performance counter maintained by the server,[detail](https://docs.microsoft.com/en-us/sql/relational-databases/system-dynamic-management-views/sys-dm-os-performance-counters-transact-sql?view=sql-server-ver15){:target=\"_blank\"}",
	DescZh: "SQL Server 性能计数器指标。",
	Fields: map[string]interface{}{
		"cntr_value":                       &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Current value of the counter"},
		"processes_blocked":                &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "The number of processes blocked."},
		"page_splits":                      &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "The number of page splits per second."},
		"full_scans":                       &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of unrestricted full scans per second. These can be either base-table or full-index scans."},
		"memory_grants_pending":            &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Specifies the total number of processes waiting for a workspace memory grant."},
		"total_server_memory":              &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeKB, Desc: "Specifies the amount of memory the server has committed using the memory manager."},
		"sql_cache_memory":                 &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeKB, Desc: "Specifies the amount of memory the server is using for the dynamic SQL cache."},
		"memory_grants_outstanding":        &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Specifies the total number of processes that have successfully acquired a workspace memory grant."},
		"database_cache_memory":            &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeKB, Desc: "Specifies the amount of memory the server is currently using for the database pages cache."},
		"connection_memory":                &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeKB, Desc: "Specifies the total amount of dynamic memory the server is using for maintaining connections."},
		"optimizer_memory":                 &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeKB, Desc: "Specifies the total amount of dynamic memory the server is using for query optimization."},
		"granted_workspace_memory":         &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeKB, Desc: "Specifies the total amount of memory currently granted to executing processes, such as hash, sort, bulk copy, and index creation operations."},
		"lock_memory":                      &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeKB, Desc: "Specifies the total amount of dynamic memory the server is using for locks."},
		"stolen_server_memory":             &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeKB, Desc: "Specifies the amount of memory the server is using for purposes other than database pages."},
		"log_pool_memory":                  &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeKB, Desc: "Total amount of dynamic memory the server is using for Log Pool."},
		"buffer_cache_hit_ratio":           &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.Percent, Desc: "The ratio of data pages found and read from the buffer cache over all data page requests."},
		"page_life_expectancy":             &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: "Duration that a page resides in the buffer pool."},
		"page_reads":                       &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Indicates the number of physical database page reads that are issued per second. This statistic displays the total number of physical page reads across all databases."},
		"page_writes":                      &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Indicates the number of physical database page writes that are issued per second."},
		"checkpoint_pages":                 &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "The number of pages flushed to disk per second by a checkpoint or other operation that require all dirty pages to be flushed."},
		"auto_param_attempts":              &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of auto-parameterization attempts per second."},
		"failed_auto_params":               &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of failed auto-parameterization attempts per second."},
		"safe_auto_params":                 &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of safe auto-parameterization attempts per second."},
		"batch_requests":                   &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "The number of batch requests per second."},
		"sql_compilations":                 &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "The number of SQL compilations per second."},
		"sql_re_compilations":              &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "The number of SQL re-compilations per second."},
		"lock_waits":                       &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "The number of times per second that SQL Server is unable to retain a lock right away for a resource."},
		"latch_waits":                      &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of latch requests that could not be granted immediately."},
		"deadlocks":                        &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of lock requests per second that resulted in a deadlock."},
		"cache_object_counts":              &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of cache objects in the cache."},
		"cache_pages":                      &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of 8-kilobyte (KB) pages used by cache objects."},
		"transaction_delay":                &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Total delay in waiting for unterminated commit acknowledgment for all the current transactions, in milliseconds."},
		"flow_control":                     &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of times flow-control initiated in the last second. Flow Control Time (ms/sec) divided by Flow Control/sec is the average time per wait."},
		"version_store_size":               &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeKB, Desc: "The size of the version store in tempdb."},
		"version_cleanup_rate":             &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeKB, Desc: "The cleanup rate of the version store in tempdb."},
		"version_generation_rate":          &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeKB, Desc: "The generation rate of the version store in tempdb."},
		"longest_transaction_running_time": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: "The time (in seconds) that the oldest active transaction has been running. Only works if database is under read committed snapshot isolation level."},
		"backup_restore_throughput":        &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Read/write throughput for backup and restore operations of a database per second."},
		"log_bytes_flushed":                &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "Total number of log bytes flushed."},
		"log_flushes":                      &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of log flushes per second."},
		"log_flush_wait_time":              &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: "Total wait time (in milliseconds) to flush the log. On an Always On secondary database, this value indicates the wait time for log records to be hardened to disk."},
		"transactions":                     &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of transactions started for the SQL Server instance per second."},
		"write_transactions":               &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of transactions that wrote to all databases on the SQL Server instance and committed, in the last second."},
		"active_transactions":              &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of active transactions across all databases on the SQL Server instance."},
		"user_connections":                 &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of user connections."},
	},
	Tags: map[string]interface{}{
		"object_name":      inputs.NewTagInfo("Category to which this counter belongs."),
		"counter_name":     inputs.NewTagInfo("Name of the counter. To get more information about a counter, this is the name of the topic to select from the list of counters in Use SQL Server Objects."),
		"counter_type":     inputs.NewTagInfo("Type of the counter"),
		"counter_instance": inputs.NewTagInfo("Name of the specific instance of the counter, for example a database, process, wait type, or resource pool."),
		"sqlserver_host":   inputs.NewTagInfo("Host name which installed SQLServer"),
		"server":           inputs.NewTagInfo("The address of the server. The value is `host:port`"),
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
		Name:   "sqlserver_waitstats",
		Cat:    point.Metric,
		Desc:   "information about all the waits encountered by threads that executed,[detail](https://docs.microsoft.com/en-us/sql/relational-databases/system-dynamic-management-views/sys-dm-os-wait-stats-transact-sql?view=sql-server-ver15){:target=\"_blank\"}",
		DescZh: "SQL Server 等待统计指标。",
		Fields: map[string]interface{}{
			"max_wait_time_ms":    &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: "Maximum wait time on this wait type."},
			"wait_time_ms":        &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: "Total wait time for this wait type in milliseconds. This time is inclusive of signal_wait_time_ms"},
			"signal_wait_time_ms": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: "Difference between the time that the waiting thread was signaled and when it started running"},
			"resource_wait_ms":    &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: "wait_time_ms-signal_wait_time_ms"},
			"waiting_tasks_count": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of waits on this wait type. This counter is incremented at the start of each wait."},
		},
		Tags: map[string]interface{}{
			"sqlserver_host": inputs.NewTagInfo("Host name which installed SQLServer"),
			"wait_type":      inputs.NewTagInfo("Name of the wait type. For more information, see Types of Waits, later in this topic"),
			"wait_category":  inputs.NewTagInfo("Wait category info (e.g., Other Disk IO, Network IO, Parallelism, SQL CLR, Service Broker, etc.)"),
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
		Name:   "sqlserver_database_io",
		Cat:    point.Metric,
		Desc:   "I/O statistics for data and log files,[detail](https://docs.microsoft.com/en-us/sql/relational-databases/system-dynamic-management-views/sys-dm-io-virtual-file-stats-transact-sql?view=sql-server-ver15){:target=\"_blank\"}",
		DescZh: "SQL Server 数据库数据文件与日志文件 I/O 统计指标。",
		Fields: map[string]interface{}{
			"read_bytes":        &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "Total number of bytes read on this file"},
			"write_bytes":       &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "Total number of bytes written to the file"},
			"read_latency_ms":   &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: "Total time, in milliseconds, that the users waited for reads issued on the file."},
			"write_latency_ms":  &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: "Total time, in milliseconds, that users waited for writes to be completed on the file"},
			"reads":             &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of reads issued on the file."},
			"writes":            &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of writes issued on the file."},
			"rg_read_stall_ms":  &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: "Does not apply to:: SQL Server 2008 through SQL Server 2012 (11.x).Total IO latency introduced by IO resource governance for reads"},
			"rg_write_stall_ms": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: "Does not apply to:: SQL Server 2008 through SQL Server 2012 (11.x).Total IO latency introduced by IO resource governance for writes. Is not nullable."},
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
		Name:   "sqlserver_schedulers",
		Cat:    point.Metric,
		Desc:   "One row per scheduler in SQL Server where each scheduler is mapped to an individual processor,[detail](https://docs.microsoft.com/en-us/sql/relational-databases/system-dynamic-management-views/sys-dm-os-schedulers-transact-sql?view=sql-server-ver15){:target=\"_blank\"}",
		DescZh: "SQL Server 调度器指标。",
		Fields: map[string]interface{}{
			"active_workers_count":      &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of workers that are active. An active worker is never preemptive, must have an associated task, and is either running, runnable, or suspended. Is not nullable."},
			"context_switches_count":    &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of context switches that have occurred on this scheduler"},
			"current_tasks_count":       &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of current tasks that are associated with this scheduler."},
			"current_workers_count":     &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of workers that are associated with this scheduler. This count includes workers that are not assigned any task. Is not nullable."},
			"is_idle":                   &inputs.FieldInfo{DataType: inputs.Bool, Type: inputs.Gauge, Unit: inputs.NoUnit, Desc: "Scheduler is idle. No workers are currently running"},
			"is_online":                 &inputs.FieldInfo{DataType: inputs.Bool, Type: inputs.Gauge, Unit: inputs.NoUnit, Desc: "If SQL Server is configured to use only some of the available processors on the server, this configuration can mean that some schedulers are mapped to processors that are not in the affinity mask. If that is the case, this column returns 0. This value means that the scheduler is not being used to process queries or batches."},
			"load_factor":               &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Internal value that indicates the perceived load on this scheduler"},
			"pending_disk_io_count":     &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of pending I/Os that are waiting to be completed."},
			"preemptive_switches_count": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of times that workers on this scheduler have switched to the preemptive mode"},
			"runnable_tasks_count":      &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of workers, with tasks assigned to them, that are waiting to be scheduled on the runnable queue."},
			"total_cpu_usage_ms":        &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: "Applies to: SQL Server 2016 (13.x) and laterTotal CPU consumed by this scheduler as reported by non-preemptive workers."},
			"total_scheduler_delay_ms":  &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: "Applies to: SQL Server 2016 (13.x) and laterThe time between one worker switching out and another one switching in"},
			"work_queue_count":          &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of tasks in the pending queue. These tasks are waiting for a worker to pick them up"},
			"yield_count":               &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Internal value that is used to indicate progress on this scheduler. This value is used by the Scheduler Monitor to determine whether a worker on the scheduler is not yielding to other workers on time."},
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
		Name:   "sqlserver_volumespace",
		Desc:   "The version should be greater than SQL Server 2008.",
		DescZh: "SQL Server 卷空间指标，要求版本高于 SQL Server 2008。",
		Cat:    point.Metric,
		Fields: map[string]interface{}{
			"volume_available_space_bytes": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "Available free space on the volume"},
			"volume_total_space_bytes":     &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "Total size in bytes of the volume"},
			"volume_used_space_bytes":      &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "Used size in bytes of the volume"},
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
		Name:   "sqlserver_lock_row",
		Desc:   "SQL Server row lock blocking information collected from dynamic management views.",
		DescZh: "从动态管理视图采集的 SQL Server 行锁阻塞信息。",
		Cat:    point.Logging,
		Fields: map[string]interface{}{
			"blocking_session_id":     &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "ID of the session that is blocking the request"},
			"session_id":              &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "ID of the session to which this request is related"},
			"cpu_time":                &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: "CPU time in milliseconds that is used by the request"},
			"logical_reads":           &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of logical reads that have been performed by the request"},
			"row_count":               &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of rows returned on the session up to this point"},
			"memory_usage":            &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of 8-KB pages of memory used by this session"},
			"last_request_start_time": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: "Time at which the last request on the session began, in second"},
			"last_request_end_time":   &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: "Time of the last completion of a request on the session, in second"},
			"host_name":               &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.TODO, Desc: "Name of the client workstation that is specific to a session"},
			"login_name":              &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.TODO, Desc: "SQL Server login name under which the session is currently executing"},
			"session_status":          &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.TODO, Desc: "Status of the session"},
			"message":                 &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.TODO, Desc: "Text of the SQL query"},
		},
		Tags: map[string]interface{}{
			"sqlserver_host":    inputs.NewTagInfo("Host name which installed SQLServer"),
			"database_instance": inputs.NewTagInfo("SQL Server instance identifier from configured tag or SQL Server server name."),
			"server":            inputs.NewTagInfo("The address of the server. The value is `host:port`"),
		},
	}
}

type LockTable struct {
	LoggingMeasurment
}

//nolint:lll
func (m *LockTable) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   "sqlserver_lock_table",
		Desc:   "SQL Server table lock information collected from dynamic management views.",
		DescZh: "从动态管理视图采集的 SQL Server 表锁信息。",
		Cat:    point.Logging,
		Fields: map[string]interface{}{
			"request_session_id": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Session ID that currently owns this request"},
			"object_name":        &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.TODO, Desc: "Name of the entity in a database with which a resource is associated"},
			"db_name":            &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.TODO, Desc: "Name of the database under which this resource is scoped"},
			"resource_type":      &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.TODO, Desc: "Represents the resource type"},
			"request_mode":       &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.TODO, Desc: "Mode of the request"},
			"request_status":     &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.TODO, Desc: "Current status of this request"},
		},
		Tags: map[string]interface{}{
			"sqlserver_host":    inputs.NewTagInfo("Host name which installed SQLServer"),
			"database_instance": inputs.NewTagInfo("SQL Server instance identifier from configured tag or SQL Server server name."),
			"server":            inputs.NewTagInfo("The address of the server. The value is `host:port`"),
		},
	}
}

type LockDead struct {
	LoggingMeasurment
}

//nolint:lll
func (m *LockDead) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   "sqlserver_lock_dead",
		Desc:   "SQL Server deadlock and blocking information collected from dynamic management views.",
		DescZh: "从动态管理视图采集的 SQL Server 死锁与阻塞信息。",
		Cat:    point.Logging,
		Fields: map[string]interface{}{
			"request_session_id":   &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Session ID that currently owns this request"},
			"blocking_session_id":  &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "ID of the session that is blocking the request"},
			"blocking_object_name": &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.TODO, Desc: "Indicates the name of the object to which this partition belongs"},
			"db_name":              &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.TODO, Desc: "Name of the database under which this resource is scoped"},
			"resource_type":        &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.TODO, Desc: "Represents the resource type"},
			"request_mode":         &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.TODO, Desc: "Mode of the request"},
			"requesting_text":      &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.TODO, Desc: "Text of the SQL query which is requesting"},
			"blocking_text":        &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.TODO, Desc: "Text of the SQL query which is blocking"},
			"message":              &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.TODO, Desc: "Text of the SQL query which is blocking"},
		},
		Tags: map[string]interface{}{
			"sqlserver_host":    inputs.NewTagInfo("Host name which installed SQLServer"),
			"database_instance": inputs.NewTagInfo("SQL Server instance identifier from configured tag or SQL Server server name."),
			"server":            inputs.NewTagInfo("The address of the server. The value is `host:port`"),
		},
	}
}

type LogicalIO struct {
	LoggingMeasurment
}

//nolint:lll
func (m *LogicalIO) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   "sqlserver_logical_io",
		Desc:   "SQL Server logical I/O query statistics collected from dynamic management views.",
		DescZh: "从动态管理视图采集的 SQL Server 查询逻辑 I/O 统计信息。",
		Cat:    point.Logging,
		Fields: map[string]interface{}{
			"avg_logical_io":       &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Average number of logical writes and logical reads"},
			"total_logical_io":     &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Total number of logical writes and logical reads"},
			"total_logical_reads":  &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Total amount of logical reads"},
			"total_logical_writes": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Total amount of logical writes"},
			"creation_time":        &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "The Unix time at which the plan was compiled, in millisecond"},
			"execution_count":      &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of times that the plan has been executed since it was last compiled"},
			"last_execution_time":  &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Last time at which the plan started executing, unix time in millisecond"},
			"message":              &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.TODO, Desc: "Text of the SQL query"},
		},
		Tags: map[string]interface{}{
			"sqlserver_host":    inputs.NewTagInfo("Host name which installed SQLServer"),
			"database_instance": inputs.NewTagInfo("SQL Server instance identifier from configured tag or SQL Server server name."),
			"server":            inputs.NewTagInfo("The address of the server. The value is `host:port`"),
		},
	}
}

type WorkerTime struct {
	LoggingMeasurment
}

//nolint:lll
func (m *WorkerTime) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   "sqlserver_worker_time",
		Desc:   "SQL Server worker time query statistics collected from dynamic management views.",
		DescZh: "从动态管理视图采集的 SQL Server 查询工作线程时间统计信息。",
		Cat:    point.Logging,
		Fields: map[string]interface{}{
			"creation_time":       &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "The Unix time at which the plan was compiled, in millisecond"},
			"execution_count":     &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of times that the plan has been executed since it was last compiled"},
			"last_execution_time": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Last time at which the plan started executing, unix time in millisecond"},
			"total_worker_time":   &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Total amount of CPU time, reported in milliseconds"},
			"avg_worker_time":     &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Average amount of CPU time, reported in milliseconds"},
			"message":             &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.TODO, Desc: "Text of the SQL query"},
		},
		Tags: map[string]interface{}{
			"sqlserver_host": inputs.NewTagInfo("Host name which installed SQLServer"),
			"server":         inputs.NewTagInfo("The address of the server. The value is `host:port`"),
		},
	}
}

type DatabaseSize struct {
	MetricMeasurment
}

//nolint:lll
func (m *DatabaseSize) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   "sqlserver_database_size",
		Desc:   "SQL Server database size metrics collected from database file metadata.",
		DescZh: "从数据库文件元数据采集的 SQL Server 数据库大小指标。",
		Cat:    point.Metric,
		Fields: map[string]interface{}{
			"data_size": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeMB, Desc: "The size of file of Rows"},
			"log_size":  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeMB, Desc: "The size of file of Log"},
		},
		Tags: map[string]interface{}{
			"database_name":  inputs.NewTagInfo("Name of the database"),
			"sqlserver_host": inputs.NewTagInfo("Host name which installed SQLServer"),
			"server":         inputs.NewTagInfo("The address of the server. The value is `host:port`"),
		},
	}
}

type DatabaseFilesMeasurement struct {
	MetricMeasurment
}

//nolint:lll
var DatabaseFilesMeasurementInfo = &inputs.MeasurementInfo{
	Name:   "sqlserver_database_files",
	Desc:   "SQL Server database file metrics collected from database file metadata.",
	DescZh: "从数据库文件元数据采集的 SQL Server 数据库文件指标。",
	Cat:    point.Metric,
	Fields: map[string]interface{}{
		"size": &inputs.FieldInfo{
			DataType: inputs.Int,
			Type:     inputs.Gauge,
			Unit:     inputs.SizeKB,
			Desc:     "Current size of the database file",
		},
	},
	Tags: map[string]interface{}{
		"database_name":  inputs.NewTagInfo("Database name"),
		"state":          inputs.NewTagInfo("Database file state: 0 = Online, 1 = Restoring, 2 = Recovering, 3 = Recovery_Pending, 4 = Suspect, 5 = Unknown, 6 = Offline, 7 = Defunct"),
		"physical_name":  inputs.NewTagInfo("Operating-system file name"),
		"state_desc":     inputs.NewTagInfo("Description of the file state"),
		"file_id":        inputs.NewTagInfo("ID of the file within database"),
		"file_type_code": inputs.NewTagInfo("File type code: 0 = Rows, 1 = Log, 2 = File-Stream, 3 = Identified for informational purposes only, 4 = Full-text"),
		"sqlserver_host": inputs.NewTagInfo("Host name which installed SQLServer"),
		"server":         inputs.NewTagInfo("The address of the server. The value is `host:port`"),
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
		Name:   "sqlserver_database_backup",
		Desc:   "SQL Server database backup metrics collected from backup metadata.",
		DescZh: "从备份元数据采集的 SQL Server 数据库备份指标。",
		Cat:    point.Metric,
		Fields: map[string]interface{}{
			"backup_count": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.Count,
				Desc:     "The total count of successful backups made for a database",
			},
		},
		Tags: map[string]interface{}{
			"database_name":  inputs.NewTagInfo("Database name"),
			"server":         inputs.NewTagInfo("The address of the server. The value is `host:port`"),
			"sqlserver_host": inputs.NewTagInfo("Host name which installed SQLServer"),
		},
	}
}
