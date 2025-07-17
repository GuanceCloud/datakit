// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package dameng

import (
	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

type MemoryMeasurement struct{}

func (m *MemoryMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: metricNameMemory,
		Cat:  point.Metric,
		Fields: map[string]interface{}{
			"buffer_size_mb": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.SizeMB,
				Desc:     "Size of the buffer pool in MB.",
			},
			"mem_pool_size_mb": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.SizeMB,
				Desc:     "Size of the memory pool in MB.",
			},
			"total_size_mb": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.SizeMB,
				Desc:     "Total memory size (buffer pool + memory pool) in MB.",
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
		Name: metricNameMemPool,
		Cat:  point.Metric,
		Fields: map[string]interface{}{
			"org_size_mb": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.SizeMB,
				Desc:     "Original size of the memory pool in MB.",
			},
			"total_size_mb": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.SizeMB,
				Desc:     "Total size of the memory pool in MB.",
			},
			"reserved_size_mb": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.SizeMB,
				Desc:     "Reserved size of the memory pool in MB.",
			},
			"data_size_mb": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.SizeMB,
				Desc:     "Data size in the memory pool in MB.",
			},
			"extend_size": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Size of extended memory in the pool.",
			},
			"target_size": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Target size of the memory pool.",
			},
			"n_extend_normal": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of normal memory extensions.",
			},
			"n_extend_exclusive": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of exclusive memory extensions.",
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
		Name: metricNameTablespace,
		Cat:  point.Metric,
		Fields: map[string]interface{}{
			"total_size_mb": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.SizeMB,
				Desc:     "Total size of the table space in MB.",
			},
			"used_size_mb": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.SizeMB,
				Desc:     "Used size of the table space in MB.",
			},
			"free_size_mb": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.SizeMB,
				Desc:     "Free size of the table space in MB.",
			},
			"max_block_mb": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.SizeMB,
				Desc:     "Maximum block size in MB.",
			},
			"usage_ratio": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.Percent,
				Desc:     "Usage ratio of the table space as a percentage.",
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
		Name: metricNameConnection,
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

type RatesMeasurement struct{}

func (m *RatesMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: metricNameRates,
		Cat:  point.Metric,
		Fields: map[string]interface{}{
			"qps": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Queries per second executed in the database.",
			},
			"tps": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Transactions per second (commits + rollbacks) in the database.",
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
		Name: metricNameSlowQueries,
		Cat:  point.Metric,
		Fields: map[string]interface{}{
			"exec_time": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.DurationMS,
				Desc:     "Execution time of the slow query in milliseconds.",
			},
			"n_runs": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of times the slow query has been executed.",
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
		Name: metricNameLocks,
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

type DeadlockMeasurement struct{}

func (m *DeadlockMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: metricNameDeadlock,
		Cat:  point.Metric,
		Fields: map[string]interface{}{
			"deadlock_count": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of times the deadlock has occurred.",
			},
		},
		Tags: map[string]interface{}{
			"host":             &inputs.TagInfo{Desc: "The server address or the host name"},
			"database":         &inputs.TagInfo{Desc: "The name of the database"},
			"deadlock_trx_id":  &inputs.TagInfo{Desc: "Transaction ID of the deadlock"},
			"deadlock_sess_id": &inputs.TagInfo{Desc: "Session ID of the deadlock"},
		},
	}
}

type BufferCacheMeasurement struct{}

func (m *BufferCacheMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: metricNameBufferCache,
		Cat:  point.Metric,
		Fields: map[string]interface{}{
			"total_size_bytes": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.SizeByte,
				Desc:     "Total size of the buffer pool in bytes.",
			},
			"total_size_gb": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.SizeGB,
				Desc:     "Total size of the buffer pool in GB.",
			},
			"buffer_hit_ratio": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.Percent,
				Desc:     "Buffer cache hit ratio as a percentage.",
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
		Name: metricNameBlockSessions,
		Cat:  point.Metric,
		Fields: map[string]interface{}{
			"block_duration_min": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.DurationMinute,
				Desc:     "Duration of the block in minutes.",
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
