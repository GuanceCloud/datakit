package mysql

import (
	"database/sql"
	"time"

	"github.com/spf13/cast"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

type baseMeasurement struct {
	client  *sql.DB
	name    string
	tags    map[string]string
	fields  map[string]interface{}
	ts      time.Time
	resData map[string]interface{}
}

// 生成行协议
func (m *baseMeasurement) LineProto() (*io.Point, error) {
	return io.MakePoint(m.name, m.tags, m.fields, m.ts)
}

// 指定指标
func (m *baseMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "redis_info",
		Fields: map[string]*inputs.FieldInfo{
			// status
			"info_latency_ms": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.DurationMS,
				Desc:     "The latency of the redis INFO command.",
			},
			"Slow_queries": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Desc:     "",
			},
			"Questions": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Desc:     "",
			},
			"Queries": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Desc:     "",
			},
			"Com_select": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Desc:     "",
			},
			"Com_insert": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Desc:     "",
			},
			"Com_update": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Desc:     "",
			},
			"Com_delete": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Desc:     "",
			},
			"Com_replace": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Desc:     "",
			},
			"Com_load": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Desc:     "",
			},
			"Com_insert_select": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Desc:     "",
			},
			"Com_update_multi": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Desc:     "",
			},
			"Com_delete_multi": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Desc:     "",
			},
			"Com_replace_select": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Desc:     "",
			},
			// Connection Metrics
			"Connections": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Desc:     "",
			},
			"Max_used_connections": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Desc:     "",
			},
			"Aborted_clients": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Desc:     "",
			},
			"Aborted_connects": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Desc:     "",
			},
			// Table Cache Metrics
			"Open_files": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Desc:     "",
			},
			"Open_tables": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Desc:     "",
			},
			// Network Metrics
			"Bytes_sent": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Desc:     "",
			},
			"Bytes_received": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Desc:     "",
			},
			// Query Cache Metrics
			"Qcache_hits": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Desc:     "",
			},
			"Qcache_inserts": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Desc:     "",
			},
			"Qcache_lowmem_prunes": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Desc:     "",
			},
			// Table Lock Metrics
			"Table_locks_waited": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Desc:     "",
			},
			"Table_locks_waited_rate": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Desc:     "",
			},
			// Temporary Table Metrics
			"Created_tmp_tables": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Desc:     "",
			},
			"Created_tmp_disk_tables": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Desc:     "",
			},
			"Created_tmp_files": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Desc:     "",
			},
			// Thread Metrics
			"Threads_connected": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Desc:     "",
			},
			"Threads_running": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Desc:     "",
			},
			// MyISAM Metrics
			"Key_buffer_bytes_unflushed": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Desc:     "",
			},
			"Key_buffer_bytes_used": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Desc:     "",
			},
			"Key_read_requests": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Desc:     "",
			},
			"Key_reads": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Desc:     "",
			},
			"Key_write_requests": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Desc:     "",
			},
			"Key_writes": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Desc:     "",
			},

			//variables
			"Key_buffer_size": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Desc:     "",
			},
			"Key_cache_utilization": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Desc:     "",
			},
			"max_connections": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Desc:     "",
			},
			"query_cache_size": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Desc:     "",
			},
			"table_open_cache": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Desc:     "",
			},
			"thread_cache_size": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Desc:     "",
			},

			// binlog
			"Binlog_space_usage_bytes": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Desc:     "",
			},

			// OPTIONAL_STATUS
			"Binlog_cache_disk_use": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Desc:     "",
			},
			"Binlog_cache_use": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Desc:     "",
			},
			"Handler_commit": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Desc:     "",
			},
			"Handler_delete": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Desc:     "",
			},
			"Handler_prepare": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Desc:     "",
			},
			"Handler_read_first": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Desc:     "",
			},
			"Handler_read_key": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Desc:     "",
			},
			"Handler_read_next": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Desc:     "",
			},
			"Handler_read_prev": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Desc:     "",
			},
			"Handler_read_rnd": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Desc:     "",
			},
			"Handler_read_rnd_next": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Desc:     "",
			},
			"Handler_rollback": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Desc:     "",
			},
			"Handler_update": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Desc:     "",
			},
			"Handler_write": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Desc:     "",
			},
			"Opened_tables": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Desc:     "",
			},
			"Qcache_total_blocks": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Desc:     "",
			},
			"Qcache_free_blocks": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Desc:     "",
			},
			"Qcache_free_memory": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Desc:     "",
			},
			"Qcache_not_cached": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Desc:     "",
			},
			"Qcache_queries_in_cache": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Desc:     "",
			},
			"Select_full_join": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Desc:     "",
			},
			"Select_full_range_join": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Desc:     "",
			},
			"Select_range": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Desc:     "",
			},
			"Select_range_check": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Desc:     "",
			},
			"Select_scan": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Desc:     "",
			},
			"Sort_merge_passes": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Desc:     "",
			},
			"Sort_range": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Desc:     "",
			},
			"Sort_rows": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Desc:     "",
			},
			"Sort_scan": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Desc:     "",
			},
			"Table_locks_immediate": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Desc:     "",
			},
			"Table_locks_immediate_rate": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Desc:     "",
			},
			"Threads_cached": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Desc:     "",
			},
			"Threads_created": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Desc:     "",
			},
		},
	}
}

func CollectBaseMeasurement(cli *sql.DB, tags map[string]string) *baseMeasurement {
	m := &baseMeasurement{
		client:  cli,
		resData: make(map[string]interface{}),
		tags:    make(map[string]string),
		fields:  make(map[string]interface{}),
	}

	m.name = "mysql_base"
	m.tags = tags

	m.getStatus()
	m.getVariables()
	m.getLogStats()

	m.submit()

	return m
}

// 数据源获取数据
func (m *baseMeasurement) getStatus() error {
	if err := m.client.Ping(); err != nil {
		l.Errorf("db connect error %v", err)
		return err
	}

	globalStatusSql := "SHOW /*!50002 GLOBAL */ STATUS;"
	rows, err := m.client.Query(globalStatusSql)
	if err != nil {
		return err
	}

	defer rows.Close()

	for rows.Next() {
		var key string
		var val *sql.RawBytes = new(sql.RawBytes)

		if err = rows.Scan(&key, val); err != nil {
			// error (todo)
			continue
		}

		m.resData[key] = string(*val)
	}

	return nil
}

// variables data
func (m *baseMeasurement) getVariables() error {
	variablesSql := "SHOW GLOBAL VARIABLES;"
	rows, err := m.client.Query(variablesSql)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var key string
		var val *sql.RawBytes = new(sql.RawBytes)

		if err = rows.Scan(&key, val); err != nil {
			continue
		}
		m.resData[key] = string(*val)
	}

	return nil
}

// log stats
func (m *baseMeasurement) getLogStats() error {
	logSql := "SHOW BINARY LOGS;"
	rows, err := m.client.Query(logSql)
	if err != nil {
		return err
	}
	defer rows.Close()

	var binaryLogSpace int64
	for rows.Next() {
		var key string
		var val *sql.RawBytes = new(sql.RawBytes)

		if err = rows.Scan(&key, val); err != nil {
			continue
		}

		v := cast.ToInt64(string(*val))

		binaryLogSpace += v

		m.resData["Binlog_space_usage_bytes"] = binaryLogSpace
	}

	return nil
}

// 提交数据
func (m *baseMeasurement) submit() error {
	metricInfo := m.Info()
	for key, item := range metricInfo.Fields {
		if value, ok := m.resData[key]; ok {
			val, err := Conv(value, item.DataType)
			if err != nil {
				l.Errorf("baseMeasurement metric %v value %v parse error %v", key, value, err)
			} else {
				m.fields[key] = val
			}
		}
	}

	return nil
}