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
		Name: "kingbase_connections",
		Cat:  point.Metric,
		Fields: map[string]interface{}{
			"active_connections": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of active connections to the database.",
			},
			"max_connections": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Maximum number of connections allowed.",
			},
			"idle_connections": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of idle connections in the database.",
			},
		},
		Tags: map[string]interface{}{
			"host":     &inputs.TagInfo{Desc: "The server address or the host name"},
			"database": &inputs.TagInfo{Desc: "The name of the database"},
		},
	}
}

type TransactionsMeasurement struct{}

func (m *TransactionsMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "kingbase_transactions",
		Cat:  point.Metric,
		Fields: map[string]interface{}{
			"commits": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of transactions committed.",
			},
			"rollbacks": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of transactions rolled back.",
			},
		},
		Tags: map[string]interface{}{
			"host":     &inputs.TagInfo{Desc: "The server address or the host name"},
			"database": &inputs.TagInfo{Desc: "The name of the database"},
		},
	}
}

type QueryPerformanceMeasurement struct{}

func (m *QueryPerformanceMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "kingbase_query_performance",
		Cat:  point.Metric,
		Fields: map[string]interface{}{
			"mean_exec_time": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				DataType: inputs.Float,
				Unit:     inputs.DurationMS,
				Desc:     "Mean query execution time",
			},
		},
		Tags: map[string]interface{}{
			"host":     &inputs.TagInfo{Desc: "The server address or the host name"},
			"database": &inputs.TagInfo{Desc: "The name of the database"},
		},
	}
}

type LocksMeasurement struct{}

func (m *LocksMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "kingbase_locks",
		Cat:  point.Metric,
		Fields: map[string]interface{}{
			"waiting_locks": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of waiting locks in the database.",
			},
		},
		Tags: map[string]interface{}{
			"host":     &inputs.TagInfo{Desc: "The server address or the host name"},
			"database": &inputs.TagInfo{Desc: "The name of the database"},
		},
	}
}

type QueryStatsMeasurement struct{}

func (m *QueryStatsMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "kingbase_query_stats",
		Cat:  point.Metric,
		Fields: map[string]interface{}{
			"total_time": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.DurationMS,
				Desc:     "Total execution time of the query in milliseconds.",
			},
			"calls": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of times the query was executed.",
			},
			"rows": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of rows returned by the query.",
			},
			"shared_blks_hit": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of shared buffer blocks hit.",
			},
			"shared_blks_read": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of shared buffer blocks read.",
			},
		},
		Tags: map[string]interface{}{
			"host":     &inputs.TagInfo{Desc: "The server address or the host name"},
			"database": &inputs.TagInfo{Desc: "The name of the database"},
			"queryid":  &inputs.TagInfo{Desc: "Unique identifier of the query"},
		},
	}
}

type BufferCacheMeasurement struct{}

func (m *BufferCacheMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "kingbase_buffer_cache",
		Cat:  point.Metric,
		Fields: map[string]interface{}{
			"buffer_hit_ratio": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.Percent,
				Desc:     "Buffer cache hit ratio as a percentage.",
			},
			"shared_blks_hit": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of shared buffer blocks hit.",
			},
			"shared_blks_read": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of shared buffer blocks read.",
			},
		},
		Tags: map[string]interface{}{
			"host":     &inputs.TagInfo{Desc: "The server address or the host name"},
			"database": &inputs.TagInfo{Desc: "The name of the database"},
		},
	}
}

type DatabaseStatusMeasurement struct{}

func (m *DatabaseStatusMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "kingbase_database_status",
		Cat:  point.Metric,
		Fields: map[string]interface{}{
			"numbackends": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				DataType: inputs.Int,
				Unit:     inputs.NCount,
				Desc:     "Number of backends",
			},
			"blks_hit": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				DataType: inputs.Int,
				Unit:     inputs.NCount,
				Desc:     "Blocks hit",
			},
			"blks_read": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				DataType: inputs.Int,
				Unit:     inputs.NCount,
				Desc:     "Blocks read",
			},
			"tup_inserted": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				DataType: inputs.Int,
				Unit:     inputs.NCount,
				Desc:     "Tuples inserted",
			},
			"tup_updated": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				DataType: inputs.Int,
				Unit:     inputs.NCount,
				Desc:     "Tuples updated",
			},
			"tup_deleted": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				DataType: inputs.Int,
				Unit:     inputs.NCount,
				Desc:     "Tuples deleted",
			},
			"conflicts": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				DataType: inputs.Int,
				Unit:     inputs.NCount,
				Desc:     "The number of conflicts occurred.",
			},
		},
		Tags: map[string]interface{}{
			"host":     &inputs.TagInfo{Desc: "The server address or the host name"},
			"database": &inputs.TagInfo{Desc: "The name of the database"},
		},
	}
}

type TablespaceMeasurement struct{}

func (m *TablespaceMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "kingbase_tablespace",
		Cat:  point.Metric,
		Fields: map[string]interface{}{
			"size_bytes": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				DataType: inputs.Int,
				Unit:     inputs.SizeByte,
				Desc:     "Tablespace size in bytes",
			},
		},
		Tags: map[string]interface{}{
			"host":     &inputs.TagInfo{Desc: "The server address or the host name"},
			"database": &inputs.TagInfo{Desc: "The name of the database"},
			"spcname":  &inputs.TagInfo{Desc: "Tablespace name"},
		},
	}
}

type LockDetailsMeasurement struct{}

func (m *LockDetailsMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "kingbase_lock_details",
		Cat:  point.Metric,
		Fields: map[string]interface{}{
			"lock_count": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of locks of a specific type.",
			},
		},
		Tags: map[string]interface{}{
			"host":      &inputs.TagInfo{Desc: "The server address or the host name"},
			"database":  &inputs.TagInfo{Desc: "The name of the database"},
			"lock_type": &inputs.TagInfo{Desc: "Type of the lock (e.g., relation, tuple)"},
		},
	}
}

type IndexUsageMeasurement struct{}

func (m *IndexUsageMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "kingbase_index_usage",
		Cat:  point.Metric,
		Fields: map[string]interface{}{
			"idx_scan": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of index scans.",
			},
			"seq_scan": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of sequential scans.",
			},
			"index_hit_ratio": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.Percent,
				Desc:     "Index hit ratio as a percentage.",
			},
		},
		Tags: map[string]interface{}{
			"host":     &inputs.TagInfo{Desc: "The server address or the host name"},
			"database": &inputs.TagInfo{Desc: "The name of the database"},
		},
	}
}

type BackgroundWriterMeasurement struct{}

func (m *BackgroundWriterMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "kingbase_bgwriter",
		Cat:  point.Metric,
		Fields: map[string]interface{}{
			"buffers_clean": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of buffers written by the background writer.",
			},
			"buffers_backend": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of buffers written by backends.",
			},
			"checkpoints_timed": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of timed checkpoints.",
			},
			"checkpoints_req": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of requested checkpoints.",
			},
		},
		Tags: map[string]interface{}{
			"host":     &inputs.TagInfo{Desc: "The server address or the host name"},
			"database": &inputs.TagInfo{Desc: "The name of the database"},
		},
	}
}

type SessionActivityMeasurement struct{}

func (m *SessionActivityMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "kingbase_session_activity",
		Cat:  point.Metric,
		Fields: map[string]interface{}{
			"session_count": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of sessions in a specific state or wait event.",
			},
		},
		Tags: map[string]interface{}{
			"host":       &inputs.TagInfo{Desc: "The server address or the host name"},
			"database":   &inputs.TagInfo{Desc: "The name of the database"},
			"state":      &inputs.TagInfo{Desc: "Session state (e.g., active, idle)"},
			"wait_event": &inputs.TagInfo{Desc: "Wait event (e.g., LWLock, IO)"},
		},
	}
}

type QueryCancellationMeasurement struct{}

func (m *QueryCancellationMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "kingbase_query_cancellation",
		Cat:  point.Metric,
		Fields: map[string]interface{}{
			"temp_files": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of temporary files created by queries.",
			},
			"deadlocks": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of deadlocks detected in the database.",
			},
		},
		Tags: map[string]interface{}{
			"host":     &inputs.TagInfo{Desc: "The server address or the host name"},
			"database": &inputs.TagInfo{Desc: "The name of the database"},
		},
	}
}

type FunctionStatsMeasurement struct{}

func (m *FunctionStatsMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "kingbase_function_stats",
		Cat:  point.Metric,
		Fields: map[string]interface{}{
			"calls": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of times the function has been called.",
			},
			"total_time": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.DurationMS,
				Desc:     "Total time spent in the function, including sub-functions (milliseconds).",
			},
			"self_time": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.DurationMS,
				Desc:     "Time spent in the function itself, excluding sub-functions (milliseconds).",
			},
		},
		Tags: map[string]interface{}{
			"host":       &inputs.TagInfo{Desc: "The server address or the host name"},
			"database":   &inputs.TagInfo{Desc: "The name of the database"},
			"schemaname": &inputs.TagInfo{Desc: "The schema name of the function"},
			"funcname":   &inputs.TagInfo{Desc: "The name of the function"},
		},
	}
}

type SlowQueriesMeasurement struct{}

func (m *SlowQueriesMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "kingbase_slow_query",
		Cat:  point.Metric,
		Fields: map[string]interface{}{
			"total_exec_time": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.DurationMS,
				Desc:     "Total execution time of the query (milliseconds).",
			},
			"calls": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of times the query has been executed.",
			},
			"mean_exec_time": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.DurationMS,
				Desc:     "Average execution time per query call (milliseconds).",
			},
		},
		Tags: map[string]interface{}{
			"host":     &inputs.TagInfo{Desc: "The server address or the host name"},
			"database": &inputs.TagInfo{Desc: "The name of the database"},
			"queryid":  &inputs.TagInfo{Desc: "Unique identifier of the query"},
			"query":    &inputs.TagInfo{Desc: "Truncated SQL query text"},
		},
	}
}
