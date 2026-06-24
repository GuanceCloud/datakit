// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package kingbase

import (
	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

type ConnectionsMeasurement struct{}

func (m *ConnectionsMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   "kingbase_connections",
		Cat:    point.Metric,
		Desc:   "Kingbase database connection utilization metrics for active, idle, and configured maximum connections.",
		DescZh: "Kingbase 数据库连接使用情况指标，包括活跃连接、空闲连接和最大连接数。",
		Fields: map[string]interface{}{
			"active_connections": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of active connections to the database.",
				Taggedby: []string{"database", "db_version"},
			},
			"max_connections": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Maximum number of connections allowed.",
				Taggedby: []string{"database", "db_version"},
			},
			"idle_connections": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of idle connections in the database.",
				Taggedby: []string{"database", "db_version"},
			},
		},
		Tags: map[string]interface{}{
			"host":       &inputs.TagInfo{Desc: "The server address or the host name"},
			"server":     &inputs.TagInfo{Desc: "The address of the server. The value is `host:port`"},
			"db_version": &inputs.TagInfo{Desc: "The version of the database"},
			"database":   &inputs.TagInfo{Desc: "The name of the database"},
		},
	}
}

type TransactionsMeasurement struct{}

func (m *TransactionsMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   "kingbase_transactions",
		Cat:    point.Metric,
		Desc:   "Kingbase database transaction counters for committed and rolled-back transactions.",
		DescZh: "Kingbase 数据库事务计数指标，包括已提交事务和已回滚事务。",
		Fields: map[string]interface{}{
			"commits": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of transactions committed.",
				Taggedby: []string{"database", "db_version"},
			},
			"rollbacks": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of transactions rolled back.",
				Taggedby: []string{"database", "db_version"},
			},
		},
		Tags: map[string]interface{}{
			"host":       &inputs.TagInfo{Desc: "The server address or the host name"},
			"server":     &inputs.TagInfo{Desc: "The address of the server. The value is `host:port`"},
			"db_version": &inputs.TagInfo{Desc: "The version of the database"},
			"database":   &inputs.TagInfo{Desc: "The name of the database"},
		},
	}
}

type QueryPerformanceMeasurement struct{}

func (m *QueryPerformanceMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   "kingbase_query_performance",
		Cat:    point.Metric,
		Desc:   "Kingbase query execution performance metrics reported per query identifier and query text.",
		DescZh: "按查询 ID 和 SQL 文本采集的 Kingbase 查询执行性能指标。",
		Fields: map[string]interface{}{
			"mean_exec_time": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				DataType: inputs.Float,
				Unit:     inputs.DurationMS,
				Desc:     "Mean query execution time in milliseconds.",
				Taggedby: []string{"queryid", "query"},
			},
		},
		Tags: map[string]interface{}{
			"host":       &inputs.TagInfo{Desc: "The server address or the host name"},
			"server":     &inputs.TagInfo{Desc: "The address of the server. The value is `host:port`"},
			"db_version": &inputs.TagInfo{Desc: "The version of the database"},
			"database":   &inputs.TagInfo{Desc: "The name of the database"},
			"queryid":    &inputs.TagInfo{Desc: "Unique identifier of the query"},
			"query":      &inputs.TagInfo{Desc: "SQL query text"},
		},
	}
}

type LocksMeasurement struct{}

func (m *LocksMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   "kingbase_locks",
		Cat:    point.Metric,
		Desc:   "Kingbase lock wait metrics for locks that have not been granted.",
		DescZh: "Kingbase 未授予锁的等待锁指标。",
		Fields: map[string]interface{}{
			"waiting_locks": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of waiting locks in the database.",
				Taggedby: []string{"database", "db_version"},
			},
		},
		Tags: map[string]interface{}{
			"host":       &inputs.TagInfo{Desc: "The server address or the host name"},
			"server":     &inputs.TagInfo{Desc: "The address of the server. The value is `host:port`"},
			"db_version": &inputs.TagInfo{Desc: "The version of the database"},
			"database":   &inputs.TagInfo{Desc: "The name of the database"},
		},
	}
}

type QueryStatsMeasurement struct{}

func (m *QueryStatsMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   "kingbase_query_stats",
		Cat:    point.Metric,
		Desc:   "Kingbase statement statistics reported per query identifier, including execution time, calls, rows, and buffer access.",
		DescZh: "按查询 ID 采集的 Kingbase 语句统计指标，包括执行耗时、调用次数、返回行数和缓冲区访问。",
		Fields: map[string]interface{}{
			"total_time": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.DurationMS,
				Desc:     "Total execution time of the query in milliseconds.",
				Taggedby: []string{"queryid"},
			},
			"calls": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of times the query was executed.",
				Taggedby: []string{"queryid"},
			},
			"rows": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of rows returned by the query.",
				Taggedby: []string{"queryid"},
			},
			"shared_blks_hit": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of shared buffer blocks hit.",
				Taggedby: []string{"queryid"},
			},
			"shared_blks_read": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of shared buffer blocks read.",
				Taggedby: []string{"queryid"},
			},
		},
		Tags: map[string]interface{}{
			"host":       &inputs.TagInfo{Desc: "The server address or the host name"},
			"server":     &inputs.TagInfo{Desc: "The address of the server. The value is `host:port`"},
			"db_version": &inputs.TagInfo{Desc: "The version of the database"},
			"database":   &inputs.TagInfo{Desc: "The name of the database"},
			"queryid":    &inputs.TagInfo{Desc: "Unique identifier of the query"},
		},
	}
}

type BufferCacheMeasurement struct{}

func (m *BufferCacheMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   "kingbase_buffer_cache",
		Cat:    point.Metric,
		Desc:   "Kingbase database-level buffer cache hit and read metrics.",
		DescZh: "Kingbase 数据库级缓冲区缓存命中和读取指标。",
		Fields: map[string]interface{}{
			"buffer_hit_ratio": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.Percent,
				Desc:     "Buffer cache hit ratio as a percentage.",
				Taggedby: []string{"database", "db_version"},
			},
			"shared_blks_hit": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of shared buffer blocks hit.",
				Taggedby: []string{"database", "db_version"},
			},
			"shared_blks_read": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of shared buffer blocks read.",
				Taggedby: []string{"database", "db_version"},
			},
		},
		Tags: map[string]interface{}{
			"host":       &inputs.TagInfo{Desc: "The server address or the host name"},
			"server":     &inputs.TagInfo{Desc: "The address of the server. The value is `host:port`"},
			"db_version": &inputs.TagInfo{Desc: "The version of the database"},
			"database":   &inputs.TagInfo{Desc: "The name of the database"},
		},
	}
}

type DatabaseStatusMeasurement struct{}

func (m *DatabaseStatusMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   "kingbase_database_status",
		Cat:    point.Metric,
		Desc:   "Kingbase database status metrics for backend sessions, block IO, tuple changes, and conflicts.",
		DescZh: "Kingbase 数据库状态指标，包括后端会话、块 IO、元组变更和冲突数。",
		Fields: map[string]interface{}{
			"numbackends": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				DataType: inputs.Int,
				Unit:     inputs.NCount,
				Desc:     "Number of backends",
				Taggedby: []string{"database", "db_version"},
			},
			"blks_hit": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				DataType: inputs.Int,
				Unit:     inputs.NCount,
				Desc:     "Blocks hit",
				Taggedby: []string{"database", "db_version"},
			},
			"blks_read": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				DataType: inputs.Int,
				Unit:     inputs.NCount,
				Desc:     "Blocks read",
				Taggedby: []string{"database", "db_version"},
			},
			"tup_inserted": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				DataType: inputs.Int,
				Unit:     inputs.NCount,
				Desc:     "Tuples inserted",
				Taggedby: []string{"database", "db_version"},
			},
			"tup_updated": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				DataType: inputs.Int,
				Unit:     inputs.NCount,
				Desc:     "Tuples updated",
				Taggedby: []string{"database", "db_version"},
			},
			"tup_deleted": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				DataType: inputs.Int,
				Unit:     inputs.NCount,
				Desc:     "Tuples deleted",
				Taggedby: []string{"database", "db_version"},
			},
			"conflicts": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				DataType: inputs.Int,
				Unit:     inputs.NCount,
				Desc:     "The number of conflicts occurred.",
				Taggedby: []string{"database", "db_version"},
			},
		},
		Tags: map[string]interface{}{
			"host":       &inputs.TagInfo{Desc: "The server address or the host name"},
			"server":     &inputs.TagInfo{Desc: "The address of the server. The value is `host:port`"},
			"db_version": &inputs.TagInfo{Desc: "The version of the database"},
			"database":   &inputs.TagInfo{Desc: "The name of the database"},
		},
	}
}

type TablespaceMeasurement struct{}

func (m *TablespaceMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   "kingbase_tablespace",
		Cat:    point.Metric,
		Desc:   "Kingbase tablespace size metrics reported per tablespace.",
		DescZh: "按表空间采集的 Kingbase 表空间大小指标。",
		Fields: map[string]interface{}{
			"size_bytes": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				DataType: inputs.Int,
				Unit:     inputs.SizeByte,
				Desc:     "Tablespace size in bytes",
				Taggedby: []string{"spcname"},
			},
		},
		Tags: map[string]interface{}{
			"host":       &inputs.TagInfo{Desc: "The server address or the host name"},
			"server":     &inputs.TagInfo{Desc: "The address of the server. The value is `host:port`"},
			"db_version": &inputs.TagInfo{Desc: "The version of the database"},
			"database":   &inputs.TagInfo{Desc: "The name of the database"},
			"spcname":    &inputs.TagInfo{Desc: "Tablespace name"},
		},
	}
}

type LockDetailsMeasurement struct{}

func (m *LockDetailsMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   "kingbase_lock_details",
		Cat:    point.Metric,
		Desc:   "Kingbase lock count metrics grouped by lock type.",
		DescZh: "按锁类型分组采集的 Kingbase 锁数量指标。",
		Fields: map[string]interface{}{
			"lock_count": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of locks of a specific type.",
				Taggedby: []string{"lock_type"},
			},
		},
		Tags: map[string]interface{}{
			"host":       &inputs.TagInfo{Desc: "The server address or the host name"},
			"server":     &inputs.TagInfo{Desc: "The address of the server. The value is `host:port`"},
			"db_version": &inputs.TagInfo{Desc: "The version of the database"},
			"database":   &inputs.TagInfo{Desc: "The name of the database"},
			"lock_type":  &inputs.TagInfo{Desc: "Type of the lock (e.g., relation, tuple)"},
		},
	}
}

type IndexUsageMeasurement struct{}

func (m *IndexUsageMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   "kingbase_index_usage",
		Cat:    point.Metric,
		Desc:   "Kingbase index and sequential scan usage metrics for user tables.",
		DescZh: "Kingbase 用户表索引扫描和顺序扫描使用情况指标。",
		Fields: map[string]interface{}{
			"idx_scan": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of index scans.",
				Taggedby: []string{"database", "db_version"},
			},
			"seq_scan": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of sequential scans.",
				Taggedby: []string{"database", "db_version"},
			},
			"index_hit_ratio": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.Percent,
				Desc:     "Index hit ratio as a percentage.",
				Taggedby: []string{"database", "db_version"},
			},
		},
		Tags: map[string]interface{}{
			"host":       &inputs.TagInfo{Desc: "The server address or the host name"},
			"server":     &inputs.TagInfo{Desc: "The address of the server. The value is `host:port`"},
			"db_version": &inputs.TagInfo{Desc: "The version of the database"},
			"database":   &inputs.TagInfo{Desc: "The name of the database"},
		},
	}
}

type BackgroundWriterMeasurement struct{}

func (m *BackgroundWriterMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   "kingbase_bgwriter",
		Cat:    point.Metric,
		Desc:   "Kingbase background writer and checkpoint buffer write metrics.",
		DescZh: "Kingbase 后台写进程和检查点缓冲区写入指标。",
		Fields: map[string]interface{}{
			"buffers_clean": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of buffers written by the background writer.",
				Taggedby: []string{"database", "db_version"},
			},
			"buffers_backend": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of buffers written by backends.",
				Taggedby: []string{"database", "db_version"},
			},
			"checkpoints_timed": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of timed checkpoints.",
				Taggedby: []string{"database", "db_version"},
			},
			"checkpoints_req": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of requested checkpoints.",
				Taggedby: []string{"database", "db_version"},
			},
		},
		Tags: map[string]interface{}{
			"host":       &inputs.TagInfo{Desc: "The server address or the host name"},
			"server":     &inputs.TagInfo{Desc: "The address of the server. The value is `host:port`"},
			"db_version": &inputs.TagInfo{Desc: "The version of the database"},
			"database":   &inputs.TagInfo{Desc: "The name of the database"},
		},
	}
}

type SessionActivityMeasurement struct{}

func (m *SessionActivityMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   "kingbase_session_activity",
		Cat:    point.Metric,
		Desc:   "Kingbase session activity counts grouped by session state and wait event.",
		DescZh: "按会话状态和等待事件分组采集的 Kingbase 会话活动数量指标。",
		Fields: map[string]interface{}{
			"session_count": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of sessions in a specific state or wait event.",
				Taggedby: []string{"state", "wait_event"},
			},
		},
		Tags: map[string]interface{}{
			"host":       &inputs.TagInfo{Desc: "The server address or the host name"},
			"server":     &inputs.TagInfo{Desc: "The address of the server. The value is `host:port`"},
			"db_version": &inputs.TagInfo{Desc: "The version of the database"},
			"database":   &inputs.TagInfo{Desc: "The name of the database"},
			"state":      &inputs.TagInfo{Desc: "Session state (e.g., active, idle)"},
			"wait_event": &inputs.TagInfo{Desc: "Wait event (e.g., LWLock, IO)"},
		},
	}
}

type QueryCancellationMeasurement struct{}

func (m *QueryCancellationMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   "kingbase_query_cancellation",
		Cat:    point.Metric,
		Desc:   "Kingbase query cancellation risk indicators for temporary files and deadlocks.",
		DescZh: "Kingbase 查询取消相关风险指标，包括临时文件数和死锁数。",
		Fields: map[string]interface{}{
			"temp_files": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of temporary files created by queries.",
				Taggedby: []string{"database", "db_version"},
			},
			"deadlocks": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of deadlocks detected in the database.",
				Taggedby: []string{"database", "db_version"},
			},
		},
		Tags: map[string]interface{}{
			"host":       &inputs.TagInfo{Desc: "The server address or the host name"},
			"server":     &inputs.TagInfo{Desc: "The address of the server. The value is `host:port`"},
			"db_version": &inputs.TagInfo{Desc: "The version of the database"},
			"database":   &inputs.TagInfo{Desc: "The name of the database"},
		},
	}
}

type FunctionStatsMeasurement struct{}

func (m *FunctionStatsMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   "kingbase_function_stats",
		Cat:    point.Metric,
		Desc:   "Kingbase user function execution statistics reported per schema and function name.",
		DescZh: "按模式和函数名采集的 Kingbase 用户函数执行统计指标。",
		Fields: map[string]interface{}{
			"calls": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of times the function has been called.",
				Taggedby: []string{"schemaname", "funcname"},
			},
			"total_time": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.DurationMS,
				Desc:     "Total time spent in the function, including sub-functions (milliseconds).",
				Taggedby: []string{"schemaname", "funcname"},
			},
			"self_time": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.DurationMS,
				Desc:     "Time spent in the function itself, excluding sub-functions (milliseconds).",
				Taggedby: []string{"schemaname", "funcname"},
			},
		},
		Tags: map[string]interface{}{
			"host":       &inputs.TagInfo{Desc: "The server address or the host name"},
			"server":     &inputs.TagInfo{Desc: "The address of the server. The value is `host:port`"},
			"db_version": &inputs.TagInfo{Desc: "The version of the database"},
			"database":   &inputs.TagInfo{Desc: "The name of the database"},
			"schemaname": &inputs.TagInfo{Desc: "The schema name of the function"},
			"funcname":   &inputs.TagInfo{Desc: "The name of the function"},
		},
	}
}

type SlowQueriesMeasurement struct{}

func (m *SlowQueriesMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   "kingbase_slow_query",
		Cat:    point.Metric,
		Desc:   "Kingbase slow query metrics reported per query identifier and query text when mean execution time exceeds the configured threshold.",
		DescZh: "按查询 ID 和 SQL 文本采集的 Kingbase 慢查询指标，包含超过配置阈值的查询。",
		Fields: map[string]interface{}{
			"total_exec_time": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.DurationMS,
				Desc:     "Total execution time of the query (milliseconds).",
				Taggedby: []string{"queryid", "query"},
			},
			"calls": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of times the query has been executed.",
				Taggedby: []string{"queryid", "query"},
			},
			"mean_exec_time": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.DurationMS,
				Desc:     "Average execution time per query call (milliseconds).",
				Taggedby: []string{"queryid", "query"},
			},
		},
		Tags: map[string]interface{}{
			"host":       &inputs.TagInfo{Desc: "The server address or the host name"},
			"server":     &inputs.TagInfo{Desc: "The address of the server. The value is `host:port`"},
			"db_version": &inputs.TagInfo{Desc: "The version of the database"},
			"database":   &inputs.TagInfo{Desc: "The name of the database"},
			"queryid":    &inputs.TagInfo{Desc: "Unique identifier of the query"},
			"query":      &inputs.TagInfo{Desc: "Truncated SQL query text"},
		},
	}
}
