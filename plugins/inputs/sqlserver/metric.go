package sqlserver

import (
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
	"time"
)

type Performance struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
	ts     time.Time
}

func (m *Performance) LineProto() (*io.Point, error) {
	return io.MakePoint(m.name, m.tags, m.fields, m.ts)
}

func (m *Performance) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "sqlserver_performance",
		Desc: "performance counter maintained by the server,[detail](https://docs.microsoft.com/en-us/sql/relational-databases/system-dynamic-management-views/sys-dm-os-performance-counters-transact-sql?view=sql-server-ver15)",
		Fields: map[string]interface{}{
			"cntr_value": newCountFieldInfo("Current value of the counter."),
		},
		Tags: map[string]interface{}{
			"object_name":    inputs.NewTagInfo("Category to which this counter belongs."),
			"counter_name":   inputs.NewTagInfo("Name of the counter. To get more information about a counter, this is the name of the topic to select from the list of counters in Use SQL Server Objects."),
			"sqlserver_host": inputs.NewTagInfo("host name which installed sqlserver"),
		},
	}
}

type WaitStatsCategorized struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
	ts     time.Time
}

func (m *WaitStatsCategorized) LineProto() (*io.Point, error) {
	return io.MakePoint(m.name, m.tags, m.fields, m.ts)
}

func (m *WaitStatsCategorized) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "sqlserver_waitstats",
		Desc: "information about all the waits encountered by threads that executed,[detail](https://docs.microsoft.com/en-us/sql/relational-databases/system-dynamic-management-views/sys-dm-os-wait-stats-transact-sql?view=sql-server-ver15)",
		Fields: map[string]interface{}{
			"max_wait_time_ms":    newTimeFieldInfo("Maximum wait time on this wait type."),
			"wait_time_ms":        newTimeFieldInfo("Total wait time for this wait type in milliseconds. This time is inclusive of signal_wait_time_ms"),
			"signal_wait_time_ms": newTimeFieldInfo("Difference between the time that the waiting thread was signaled and when it started running"),
			"resource_wait_ms":    newTimeFieldInfo("wait_time_ms-signal_wait_time_ms"),
			"waiting_tasks_count": newCountFieldInfo("Number of waits on this wait type. This counter is incremented at the start of each wait."),
		},
		Tags: map[string]interface{}{
			"sqlserver_host": inputs.NewTagInfo("host name which installed sqlserver"),
			"wait_type":      inputs.NewTagInfo("Name of the wait type. For more information, see Types of Waits, later in this topic"),
			"wait_category":  inputs.NewTagInfo("wait category info"),
		},
	}
}

type DatabaseIO struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
	ts     time.Time
}

func (m *DatabaseIO) LineProto() (*io.Point, error) {
	return io.MakePoint(m.name, m.tags, m.fields, m.ts)
}

func (m *DatabaseIO) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "sqlserver_database_io",
		Desc: "I/O statistics for data and log files,[detail](https://docs.microsoft.com/en-us/sql/relational-databases/system-dynamic-management-views/sys-dm-io-virtual-file-stats-transact-sql?view=sql-server-ver15)",
		Fields: map[string]interface{}{
			"read_bytes":        newByteFieldInfo("Total number of bytes read on this file"),
			"write_bytes":       newByteFieldInfo("Number of writes made on this file"),
			"read_latency_ms":   newTimeFieldInfo("Total time, in milliseconds, that the users waited for reads issued on the file."),
			"write_latency_ms":  newTimeFieldInfo("Total time, in milliseconds, that users waited for writes to be completed on the file"),
			"read":              newCountFieldInfo("Number of reads issued on the file."),
			"writes":            newCountFieldInfo("Number of writes issued on the file."),
			"rg_read_stall_ms":  newTimeFieldInfo("Does not apply to:: SQL Server 2008 through SQL Server 2012 (11.x).Total IO latency introduced by IO resource governance for reads"),
			"rg_write_stall_ms": newTimeFieldInfo("Does not apply to:: SQL Server 2008 through SQL Server 2012 (11.x).Total IO latency introduced by IO resource governance for writes. Is not nullable."),
		},
		Tags: map[string]interface{}{
			"database_name":     inputs.NewTagInfo("database name"),
			"file_type":         inputs.NewTagInfo("Description of the file type,ROWS、LOG、FILESTREAM、FULLTEXT (Full-text catalogs earlier than SQL Server 2008.)"),
			"logical_filename":  inputs.NewTagInfo("Logical name of the file in the database"),
			"physical_filename": inputs.NewTagInfo("Operating-system file name."),
			"sqlserver_host":    inputs.NewTagInfo("host name which installed sqlserver"),
		},
	}
}

type ServerProperties struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
	ts     time.Time
}

func (m *ServerProperties) LineProto() (*io.Point, error) {
	return io.MakePoint(m.name, m.tags, m.fields, m.ts)
}

func (m *ServerProperties) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "sqlserver",
		Fields: map[string]interface{}{
			"cpu_count":           newCountFieldInfo("Specifies the number of logical CPUs on the system. Not nullable."),
			"db_online":           newCountFieldInfo("num of database state in online"),
			"db_offline":          newCountFieldInfo("num of database state in offline"),
			"db_recovering":       newCountFieldInfo("num of database state in recovering"),
			"db_recovery_pending": newCountFieldInfo("num of database state in recovery_pending"),
			"db_restoring":        newCountFieldInfo("num of database state in restoring"),
			"db_suspect":          newCountFieldInfo("num of database state in suspect"),
			"server_memory":       newByteFieldInfo("memory used"),
		},
		Tags: map[string]interface{}{
			"sqlserver_host": inputs.NewTagInfo("host name which installed sqlserver"),
		},
	}
}

type Schedulers struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
	ts     time.Time
}

func (m *Schedulers) LineProto() (*io.Point, error) {
	return io.MakePoint(m.name, m.tags, m.fields, m.ts)
}

func (m *Schedulers) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "sqlserver_schedulers",
		Desc: "one row per scheduler in SQL Server where each scheduler is mapped to an individual processor,[detail](https://docs.microsoft.com/en-us/sql/relational-databases/system-dynamic-management-views/sys-dm-os-schedulers-transact-sql?view=sql-server-ver15)",
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
			"sqlserver_host": inputs.NewTagInfo("host name which installed sqlserver"),
			"scheduler_id":   inputs.NewTagInfo("ID of the scheduler. All schedulers that are used to run regular queries have ID numbers less than 1048576. Those schedulers that have IDs greater than or equal to 1048576 are used internally by SQL Server, such as the dedicated administrator connection scheduler. Is not nullable."),
		},
	}
}

type VolumeSpace struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
	ts     time.Time
}

func (m *VolumeSpace) LineProto() (*io.Point, error) {
	return io.MakePoint(m.name, m.tags, m.fields, m.ts)
}

func (m *VolumeSpace) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "sqlserver_volumespace",
		Fields: map[string]interface{}{
			"volume_available_space_bytes": newByteFieldInfo("Available free space on the volume"),
			"volume_total_space_bytes":     newByteFieldInfo("Total size in bytes of the volume"),
			"volume_used_space_bytes":      newByteFieldInfo("Used size in bytes of the volume"),
		},
		Tags: map[string]interface{}{
			"sqlserver_host":     inputs.NewTagInfo("host name which installed sqlserver"),
			"volume_mount_point": inputs.NewTagInfo("Mount point at which the volume is rooted. Can return an empty string. Returns null on Linux operating system."),
		},
	}
}
