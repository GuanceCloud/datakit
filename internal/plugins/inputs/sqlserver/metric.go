// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package sqlserver

import (
	"time"

	"github.com/GuanceCloud/cliutils/point"
	dkpt "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

type Performance struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
	ts     time.Time
	ipt    *Input
}

// Point implement MeasurementV2.
func (m *Performance) Point() *point.Point {
	opts := point.DefaultMetricOptions()
	opts = append(opts, point.WithTime(m.ts), m.ipt.opt)

	return point.NewPointV2([]byte(m.name),
		append(point.NewTags(m.tags), point.NewKVs(m.fields)...),
		opts...)
}

func (m *Performance) LineProto() (*dkpt.Point, error) {
	return dkpt.NewPoint(m.name, m.tags, m.fields, dkpt.MOptElection())
}

//nolint:lll
func (m *Performance) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "sqlserver_performance",
		Type: "metric",
		Desc: "performance counter maintained by the server,[detail](https://docs.microsoft.com/en-us/sql/relational-databases/system-dynamic-management-views/sys-dm-os-performance-counters-transact-sql?view=sql-server-ver15)",
		Fields: map[string]interface{}{
			"cntr_value": newCountFieldInfo("Current value of the counter."),
		},
		Tags: map[string]interface{}{
			"object_name":    inputs.NewTagInfo("Category to which this counter belongs."),
			"counter_name":   inputs.NewTagInfo("Name of the counter. To get more information about a counter, this is the name of the topic to select from the list of counters in Use SQL Server Objects."),
			"sqlserver_host": inputs.NewTagInfo("host name which installed SQLServer"),
		},
	}
}

type WaitStatsCategorized struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
}

func (m *WaitStatsCategorized) LineProto() (*dkpt.Point, error) {
	return dkpt.NewPoint(m.name, m.tags, m.fields, dkpt.MOptElection())
}

//nolint:lll
func (m *WaitStatsCategorized) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "sqlserver_waitstats",
		Type: "metric",
		Desc: "information about all the waits encountered by threads that executed,[detail](https://docs.microsoft.com/en-us/sql/relational-databases/system-dynamic-management-views/sys-dm-os-wait-stats-transact-sql?view=sql-server-ver15)",
		Fields: map[string]interface{}{
			"max_wait_time_ms":    newTimeFieldInfo("Maximum wait time on this wait type."),
			"wait_time_ms":        newTimeFieldInfo("Total wait time for this wait type in milliseconds. This time is inclusive of signal_wait_time_ms"),
			"signal_wait_time_ms": newTimeFieldInfo("Difference between the time that the waiting thread was signaled and when it started running"),
			"resource_wait_ms":    newTimeFieldInfo("wait_time_ms-signal_wait_time_ms"),
			"waiting_tasks_count": newCountFieldInfo("Number of waits on this wait type. This counter is incremented at the start of each wait."),
		},
		Tags: map[string]interface{}{
			"sqlserver_host": inputs.NewTagInfo("host name which installed SQLServer"),
			"wait_type":      inputs.NewTagInfo("Name of the wait type. For more information, see Types of Waits, later in this topic"),
			"wait_category":  inputs.NewTagInfo("wait category info"),
		},
	}
}

type DatabaseIO struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
}

func (m *DatabaseIO) LineProto() (*dkpt.Point, error) {
	return dkpt.NewPoint(m.name, m.tags, m.fields, dkpt.MOptElection())
}

//nolint:lll
func (m *DatabaseIO) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "sqlserver_database_io",
		Type: "metric",
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
			"file_type":         inputs.NewTagInfo("Description of the file type, `ROWS/LOG/FILESTREAM/FULLTEXT` (Full-text catalogs earlier than SQL Server 2008.)"),
			"logical_filename":  inputs.NewTagInfo("Logical name of the file in the database"),
			"physical_filename": inputs.NewTagInfo("Operating-system file name."),
			"sqlserver_host":    inputs.NewTagInfo("host name which installed SQLServer"),
		},
	}
}

type ServerProperties struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
}

func (m *ServerProperties) LineProto() (*dkpt.Point, error) {
	return dkpt.NewPoint(m.name, m.tags, m.fields, dkpt.MOptElection())
}

//nolint:lll
func (m *ServerProperties) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "sqlserver",
		Type: "metric",
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
			"sqlserver_host": inputs.NewTagInfo("host name which installed SQLServer"),
		},
	}
}

type Schedulers struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
}

func (m *Schedulers) LineProto() (*dkpt.Point, error) {
	return dkpt.NewPoint(m.name, m.tags, m.fields, dkpt.MOptElection())
}

//nolint:lll
func (m *Schedulers) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "sqlserver_schedulers",
		Type: "metric",
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
			"sqlserver_host": inputs.NewTagInfo("host name which installed SQLServer"),
			"scheduler_id":   inputs.NewTagInfo("ID of the scheduler. All schedulers that are used to run regular queries have ID numbers less than 1048576. Those schedulers that have IDs greater than or equal to 1048576 are used internally by SQL Server, such as the dedicated administrator connection scheduler. Is not nullable."),
		},
	}
}

type VolumeSpace struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
}

func (m *VolumeSpace) LineProto() (*dkpt.Point, error) {
	return dkpt.NewPoint(m.name, m.tags, m.fields, dkpt.MOptElection())
}

//nolint:lll
func (m *VolumeSpace) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "sqlserver_volumespace",
		Type: "metric",
		Fields: map[string]interface{}{
			"volume_available_space_bytes": newByteFieldInfo("Available free space on the volume"),
			"volume_total_space_bytes":     newByteFieldInfo("Total size in bytes of the volume"),
			"volume_used_space_bytes":      newByteFieldInfo("Used size in bytes of the volume"),
		},
		Tags: map[string]interface{}{
			"sqlserver_host":     inputs.NewTagInfo("host name which installed SQLServer"),
			"volume_mount_point": inputs.NewTagInfo("Mount point at which the volume is rooted. Can return an empty string. Returns null on Linux operating system."),
		},
	}
}

type LockRow struct {
	name     string
	tags     map[string]string
	fields   map[string]interface{}
	election bool
}

func (m *LockRow) LineProto() (*dkpt.Point, error) {
	return dkpt.NewPoint(m.name, m.tags, m.fields, dkpt.LOptElectionV2(m.election))
}

//nolint:lll
func (m *LockRow) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "sqlserver_lock_row",
		Type: "logging",
		Fields: map[string]interface{}{
			"blocking_session_id":     newCountFieldInfo("ID of the session that is blocking the request"),
			"session_id":              newCountFieldInfo("ID of the session to which this request is related"),
			"cpu_time":                newTimeFieldInfo("CPU time in milliseconds that is used by the request"),
			"logical_reads":           newCountFieldInfo("Number of logical reads that have been performed by the request"),
			"row_count":               newCountFieldInfo("Number of rows returned on the session up to this point"),
			"memory_usage":            newCountFieldInfo("Number of 8-KB pages of memory used by this session"),
			"last_request_start_time": newTimeFieldInfo("Time at which the last request on the session began, in second"),
			"last_request_end_time":   newTimeFieldInfo("Time of the last completion of a request on the session, in second"),
		},
		Tags: map[string]interface{}{
			"host_name":      inputs.NewTagInfo("Name of the client workstation that is specific to a session"),
			"login_name":     inputs.NewTagInfo("SQL Server login name under which the session is currently executing"),
			"session_status": inputs.NewTagInfo("Status of the session"),
			"text":           inputs.NewTagInfo("Text of the SQL query"),
		},
	}
}

type LockTable struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
}

func (m *LockTable) LineProto() (*dkpt.Point, error) {
	return dkpt.NewPoint(m.name, m.tags, m.fields, dkpt.LOptElection())
}

//nolint:lll
func (m *LockTable) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "sqlserver_lock_table",
		Type: "logging",
		Fields: map[string]interface{}{
			"resource_session_id": newCountFieldInfo("Session ID that currently owns this request"),
		},
		Tags: map[string]interface{}{
			"object_name":    inputs.NewTagInfo("Name of the entity in a database with which a resource is associated"),
			"db_name":        inputs.NewTagInfo("Name of the database under which this resource is scoped"),
			"resource_type":  inputs.NewTagInfo("Represents the resource type"),
			"request_mode":   inputs.NewTagInfo("Mode of the request"),
			"request_status": inputs.NewTagInfo("Current status of this request"),
		},
	}
}

type LockDead struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
}

func (m *LockDead) LineProto() (*dkpt.Point, error) {
	return dkpt.NewPoint(m.name, m.tags, m.fields, dkpt.LOptElection())
}

//nolint:lll
func (m *LockDead) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "sqlserver_lock_dead",
		Type: "logging",
		Fields: map[string]interface{}{
			"request_session_id":  newCountFieldInfo("Session ID that currently owns this request"),
			"blocking_session_id": newCountFieldInfo("ID of the session that is blocking the request"),
		},
		Tags: map[string]interface{}{
			"blocking_object_name": inputs.NewTagInfo("Indicates the name of the object to which this partition belongs"),
			"db_name":              inputs.NewTagInfo("Name of the database under which this resource is scoped"),
			"resource_type":        inputs.NewTagInfo("Represents the resource type"),
			"request_mode":         inputs.NewTagInfo("Mode of the request"),
			"requesting_text":      inputs.NewTagInfo("Text of the SQL query which is requesting"),
			"blocking_text":        inputs.NewTagInfo("Text of the SQL query which is blocking"),
		},
	}
}

type LogicalIO struct {
	name     string
	tags     map[string]string
	fields   map[string]interface{}
	election bool
}

func (m *LogicalIO) LineProto() (*dkpt.Point, error) {
	return dkpt.NewPoint(m.name, m.tags, m.fields, dkpt.LOptElectionV2(m.election))
}

//nolint:lll
func (m *LogicalIO) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "sqlserver_logical_io",
		Type: "logging",
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
		},
	}
}

type WorkerTime struct {
	name     string
	tags     map[string]string
	fields   map[string]interface{}
	election bool
}

func (m *WorkerTime) LineProto() (*dkpt.Point, error) {
	return dkpt.NewPoint(m.name, m.tags, m.fields, dkpt.LOptElectionV2(m.election))
}

//nolint:lll
func (m *WorkerTime) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "sqlserver_worker_time",
		Type: "logging",
		Fields: map[string]interface{}{
			"creation_time":       newCountFieldInfo("The Unix time at which the plan was compiled, in millisecond"),
			"execution_count":     newCountFieldInfo("Number of times that the plan has been executed since it was last compiled"),
			"last_execution_time": newCountFieldInfo("Last time at which the plan started executing, unix time in millisecond"),
			"total_worker_time":   newCountFieldInfo("Total amount of CPU time, reported in milliseconds"),
			"avg_worker_time":     newCountFieldInfo("Average amount of CPU time, reported in milliseconds"),
		},
		Tags: map[string]interface{}{
			"message": inputs.NewTagInfo("Text of the SQL query"),
		},
	}
}

type DatabaseSize struct {
	name     string
	tags     map[string]string
	fields   map[string]interface{}
	election bool
}

func (m *DatabaseSize) LineProto() (*dkpt.Point, error) {
	return dkpt.NewPoint(m.name, m.tags, m.fields, dkpt.MOptElectionV2(m.election))
}

//nolint:lll
func (m *DatabaseSize) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "sqlserver_database_size",
		Type: "metric",
		Fields: map[string]interface{}{
			"data_size": newKByteFieldInfo("The size of file of Rows"),
			"log_size":  newKByteFieldInfo("The size of file of Log"),
		},
		Tags: map[string]interface{}{
			"name": inputs.NewTagInfo("Name of the database"),
		},
	}
}
