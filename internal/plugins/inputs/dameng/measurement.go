// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package dameng

import (
	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

var (
	damengHostDatabaseTaggedBy = []string{
		"host",
		"database",
	}
	damengCommonMemPoolTaggedBy = []string{
		"host",
		"database",
		"pool_name",
		"is_shared",
		"is_overflow",
	}
	damengTablespaceTaggedBy = []string{
		"host",
		"database",
		"tablespace_name",
	}
	damengDeadlockTaggedBy = []string{
		"host",
		"database",
		"trx_id",
		"sess_id",
	}
	damengBufferCacheTaggedBy = []string{
		"host",
		"database",
		"pool_name",
	}
	damengBlockSessionsTaggedBy = []string{
		"host",
		"database",
		"blocked_sess_id",
		"blocked_trx_id",
		"blocked_lock_type",
		"blocked_start_time",
		"blocking_sess_id",
		"blocking_ip",
		"blocking_trx_id",
	}
	damengSlowQueriesTaggedBy = []string{
		"host",
		"database",
		"sess_id",
		"sql_id",
		"sql_text",
	}
)

type MemoryMeasurement struct{}

func (m *MemoryMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   metricNameMemory,
		Desc:   "Aggregated memory pool metrics for Dameng.",
		DescZh: "Dameng 的内存池聚合指标。",
		Cat:    point.Metric,
		Fields: map[string]interface{}{
			"buffer_size_mb": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.SizeMB,
				Desc:     "Size of the buffer pool in MB.",
				Taggedby: damengHostDatabaseTaggedBy,
			},
			"mem_pool_size_mb": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.SizeMB,
				Desc:     "Size of the memory pool in MB.",
				Taggedby: damengHostDatabaseTaggedBy,
			},
			"total_size_mb": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.SizeMB,
				Desc:     "Total memory size (buffer pool + memory pool) in MB.",
				Taggedby: damengHostDatabaseTaggedBy,
			},
		},
		Tags: map[string]interface{}{
			"host":     &inputs.TagInfo{Desc: "The server address or the host name"},
			"database": &inputs.TagInfo{Desc: "The name of the database"},
		},
	}
}

type MemPoolMeasurement struct{}

func (m *MemPoolMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   metricNameMemPool,
		Desc:   "Per-pool memory allocation and extension metrics for Dameng.",
		DescZh: "Dameng 各内存池分配与扩展指标。",
		Cat:    point.Metric,
		Fields: map[string]interface{}{
			"org_size_mb": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.SizeMB,
				Desc:     "Original size of the memory pool in MB.",
				Taggedby: damengCommonMemPoolTaggedBy,
			},
			"total_size_mb": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.SizeMB,
				Desc:     "Total size of the memory pool in MB.",
				Taggedby: damengCommonMemPoolTaggedBy,
			},
			"reserved_size_mb": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.SizeMB,
				Desc:     "Reserved size of the memory pool in MB.",
				Taggedby: damengCommonMemPoolTaggedBy,
			},
			"data_size_mb": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.SizeMB,
				Desc:     "Data size in the memory pool in MB.",
				Taggedby: damengCommonMemPoolTaggedBy,
			},
			"extend_size_mb": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.SizeMB,
				Desc:     "Size of extended memory in the pool.",
				Taggedby: damengCommonMemPoolTaggedBy,
			},
			"target_size_mb": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.SizeMB,
				Desc:     "Target size of the memory pool.",
				Taggedby: damengCommonMemPoolTaggedBy,
			},
			"n_extend_normal": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of normal memory extensions.",
				Taggedby: damengCommonMemPoolTaggedBy,
			},
			"n_extend_exclusive": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of exclusive memory extensions.",
				Taggedby: damengCommonMemPoolTaggedBy,
			},
		},
		Tags: map[string]interface{}{
			"host":        &inputs.TagInfo{Desc: "The server address or the host name"},
			"database":    &inputs.TagInfo{Desc: "The name of the database"},
			"pool_name":   &inputs.TagInfo{Desc: "Name of the memory pool"},
			"is_shared":   &inputs.TagInfo{Desc: "Whether the memory pool is shared (Y/N)"},
			"is_overflow": &inputs.TagInfo{Desc: "Whether the memory pool is in overflow state (Y/N)"},
		},
	}
}

type TablespaceMeasurement struct{}

func (m *TablespaceMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   metricNameTablespace,
		Desc:   "Tablespace capacity and usage metrics for Dameng.",
		DescZh: "Dameng 表空间容量与使用率指标。",
		Cat:    point.Metric,
		Fields: map[string]interface{}{
			"total_size_mb": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.SizeMB,
				Desc:     "Total size of the table space in MB.",
				Taggedby: damengTablespaceTaggedBy,
			},
			"used_size_mb": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.SizeMB,
				Desc:     "Used size of the table space in MB.",
				Taggedby: damengTablespaceTaggedBy,
			},
			"free_size_mb": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.SizeMB,
				Desc:     "Free size of the table space in MB.",
				Taggedby: damengTablespaceTaggedBy,
			},
			"max_block_mb": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.SizeMB,
				Desc:     "Maximum block size in MB.",
				Taggedby: damengTablespaceTaggedBy,
			},
			"usage_ratio": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.Percent,
				Desc:     "Usage ratio of the table space as a percentage.",
				Taggedby: damengTablespaceTaggedBy,
			},
		},
		Tags: map[string]interface{}{
			"host":            &inputs.TagInfo{Desc: "The server address or the host name"},
			"database":        &inputs.TagInfo{Desc: "The name of the database"},
			"tablespace_name": &inputs.TagInfo{Desc: "Name of the table space"},
		},
	}
}

type ConnectionsMeasurement struct{}

func (m *ConnectionsMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   metricNameConnection,
		Desc:   "Database connection state counts for Dameng.",
		DescZh: "Dameng 数据库连接状态数量指标。",
		Cat:    point.Metric,
		Fields: map[string]interface{}{
			"active_connections": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of active connections to the database.",
				Taggedby: damengHostDatabaseTaggedBy,
			},
			"max_connections": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Maximum number of connections allowed.",
				Taggedby: damengHostDatabaseTaggedBy,
			},
			"idle_connections": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of idle connections in the database.",
				Taggedby: damengHostDatabaseTaggedBy,
			},
		},
		Tags: map[string]interface{}{
			"host":     &inputs.TagInfo{Desc: "The server address or the host name"},
			"database": &inputs.TagInfo{Desc: "The name of the database"},
		},
	}
}

type RatesMeasurement struct{}

func (m *RatesMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   metricNameRates,
		Desc:   "Query and transaction rates for Dameng.",
		DescZh: "Dameng 查询和事务速率指标。",
		Cat:    point.Metric,
		Fields: map[string]interface{}{
			"qps": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Queries per second executed in the database.",
				Taggedby: damengHostDatabaseTaggedBy,
			},
			"tps": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Transactions per second (commits + rollbacks) in the database.",
				Taggedby: damengHostDatabaseTaggedBy,
			},
		},
		Tags: map[string]interface{}{
			"host":     &inputs.TagInfo{Desc: "The server address or the host name"},
			"database": &inputs.TagInfo{Desc: "The name of the database"},
		},
	}
}

type SlowQueriesMeasurement struct{}

func (m *SlowQueriesMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   metricNameSlowQueries,
		Desc:   "Slow query execution details for Dameng.",
		DescZh: "Dameng 慢查询执行详情。",
		Cat:    point.Metric,
		Fields: map[string]interface{}{
			"exec_time": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.DurationMS,
				Desc:     "Execution time of the slow query in milliseconds.",
				Taggedby: damengSlowQueriesTaggedBy,
			},
			"n_runs": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of times the slow query has been executed.",
				Taggedby: damengSlowQueriesTaggedBy,
			},
		},
		Tags: map[string]interface{}{
			"host":     &inputs.TagInfo{Desc: "The server address or the host name"},
			"database": &inputs.TagInfo{Desc: "The name of the database"},
			"sess_id":  &inputs.TagInfo{Desc: "Session ID of the slow query"},
			"sql_id":   &inputs.TagInfo{Desc: "Unique identifier of the slow query"},
			"sql_text": &inputs.TagInfo{Desc: "Truncated SQL query text."},
		},
	}
}

type LocksMeasurement struct{}

func (m *LocksMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   metricNameLocks,
		Desc:   "Lock contention metrics for Dameng.",
		DescZh: "Dameng 锁竞争指标。",
		Cat:    point.Metric,
		Fields: map[string]interface{}{
			"waiting_locks": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of waiting locks in the database.",
				Taggedby: damengHostDatabaseTaggedBy,
			},
		},
		Tags: map[string]interface{}{
			"host":     &inputs.TagInfo{Desc: "The server address or the host name"},
			"database": &inputs.TagInfo{Desc: "The name of the database"},
		},
	}
}

type DeadlockMeasurement struct{}

func (m *DeadlockMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   metricNameDeadlock,
		Desc:   "Deadlock occurrence metrics for Dameng.",
		DescZh: "Dameng 死锁发生指标。",
		Cat:    point.Metric,
		Fields: map[string]interface{}{
			"deadlock_count": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of times the deadlock has occurred.",
				Taggedby: damengDeadlockTaggedBy,
			},
		},
		Tags: map[string]interface{}{
			"host":     &inputs.TagInfo{Desc: "The server address or the host name"},
			"database": &inputs.TagInfo{Desc: "The name of the database"},
			"trx_id":   &inputs.TagInfo{Desc: "Transaction ID of the deadlock"},
			"sess_id":  &inputs.TagInfo{Desc: "Session ID of the deadlock"},
		},
	}
}

type BufferCacheMeasurement struct{}

func (m *BufferCacheMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   metricNameBufferCache,
		Desc:   "Dameng buffer cache size and hit ratio metrics.",
		DescZh: "Dameng 缓冲池大小及命中率指标。",
		Cat:    point.Metric,
		Fields: map[string]interface{}{
			"total_size_bytes": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.SizeByte,
				Desc:     "Total size of the buffer pool in bytes.",
				Taggedby: damengBufferCacheTaggedBy,
			},
			"total_size_gb": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.SizeGB,
				Desc:     "Total size of the buffer pool in GB.",
				Taggedby: damengBufferCacheTaggedBy,
			},
			"buffer_hit_ratio": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.Percent,
				Desc:     "Buffer cache hit ratio as a percentage.",
				Taggedby: damengBufferCacheTaggedBy,
			},
		},
		Tags: map[string]interface{}{
			"host":      &inputs.TagInfo{Desc: "The server address or the host name"},
			"database":  &inputs.TagInfo{Desc: "The name of the database"},
			"pool_name": &inputs.TagInfo{Desc: "Name of the buffer pool"},
		},
	}
}

type BlockSessionsMeasurement struct{}

func (m *BlockSessionsMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   metricNameBlockSessions,
		Desc:   "Blocking session metrics for Dameng.",
		DescZh: "Dameng 锁阻塞会话指标。",
		Cat:    point.Metric,
		Fields: map[string]interface{}{
			"block_duration_min": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.DurationMinute,
				Desc:     "Duration of the block in minutes.",
				Taggedby: damengBlockSessionsTaggedBy,
			},
		},
		Tags: map[string]interface{}{
			"host":               &inputs.TagInfo{Desc: "The server address or the host name"},
			"database":           &inputs.TagInfo{Desc: "The name of the database"},
			"blocked_sess_id":    &inputs.TagInfo{Desc: "Session ID of the blocked session"},
			"blocked_trx_id":     &inputs.TagInfo{Desc: "Transaction ID of the blocked session"},
			"blocked_lock_type":  &inputs.TagInfo{Desc: "Type of the lock causing the block (e.g., object_lock, transaction_lock)"},
			"blocked_start_time": &inputs.TagInfo{Desc: "Start time of the blocked session"},
			"blocking_sess_id":   &inputs.TagInfo{Desc: "Session ID of the blocking session"},
			"blocking_ip":        &inputs.TagInfo{Desc: "Client IP of the blocking session"},
			"blocking_trx_id":    &inputs.TagInfo{Desc: "Transaction ID of the blocking session"},
		},
	}
}
