package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/spf13/cast"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

type baseMeasurement struct {
	i       *Input
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
		Name: "mysql",
		Fields: map[string]interface{}{
			// status
			"Slow_queries": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The rate of slow queries.",
			},
			"Questions": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The rate of statements executed by the server.",
			},
			"Queries": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The rate of queries.",
			},
			"Com_select": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The rate of select statements.",
			},
			"Com_insert": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The rate of insert statements.",
			},
			"Com_update": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The rate of update statements.",
			},
			"Com_delete": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The rate of delete statements.",
			},
			"Com_replace": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The rate of replace statements.",
			},
			"Com_load": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The rate of load statements.",
			},
			"Com_insert_select": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The rate of insert-select statements.",
			},
			"Com_update_multi": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The rate of update-multi.",
			},
			"Com_delete_multi": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The rate of delete-multi statements.",
			},
			"Com_replace_select": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The rate of replace-select statements.",
			},
			// Connection Metrics
			"Connections": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The rate of connections to the server.",
			},
			"Max_used_connections": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The maximum number of connections that have been in use simultaneously since the server started.",
			},
			"Aborted_clients": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The number of connections that were aborted because the client died without closing the connection properly.",
			},
			"Aborted_connects": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The number of failed attempts to connect to the MySQL server.",
			},
			// Table Cache Metrics
			"Open_files": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The number of open files.",
			},
			"Open_tables": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The number of of tables that are open.",
			},
			// Network Metrics
			"Bytes_sent": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.SizeIByte,
				Desc:     "The number of bytes sent to all clients.",
			},
			"Bytes_received": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.SizeIByte,
				Desc:     "The number of bytes received from all clients.",
			},
			// Query Cache Metrics
			"Qcache_hits": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The rate of query cache hits.",
			},
			"Qcache_inserts": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The number of queries added to the query cache.",
			},
			"Qcache_lowmem_prunes": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The number of queries that were deleted from the query cache because of low memory.",
			},
			// Table Lock Metrics
			"Table_locks_waited": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The total number of times that a request for a table lock could not be granted immediately and a wait was needed.",
			},
			// Temporary Table Metrics
			"Created_tmp_tables": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The rate of internal temporary tables created by second by the server while executing statements.",
			},
			"Created_tmp_disk_tables": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The rate of internal on-disk temporary tables created by second by the server while executing statements.",
			},
			"Created_tmp_files": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The rate of temporary files created by second.",
			},
			// Thread Metrics
			"Threads_connected": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The number of currently open connections.",
			},
			"Threads_running": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The number of threads that are not sleeping.",
			},
			// MyISAM Metrics
			"Key_buffer_bytes_unflushed": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "MyISAM key buffer bytes unflushed.",
			},
			"Key_buffer_bytes_used": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "MyISAM key buffer bytes used.",
			},
			"Key_read_requests": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The number of requests to read a key block from the MyISAM key cache.",
			},
			"Key_reads": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The number of physical reads of a key block from disk into the MyISAM key cache. If Key_reads is large, then your key_buffer_size value is probably too small. The cache miss rate can be calculated as Key_reads/Key_read_requests.",
			},
			"Key_write_requests": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The number of requests to write a key block to the MyISAM key cache.",
			},
			"Key_writes": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The number of physical writes of a key block from the MyISAM key cache to disk.",
			},

			//variables
			"Key_buffer_size": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.SizeIByte,
				Desc:     "Size of the buffer used for index blocks.",
			},
			"Key_cache_utilization": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Desc:     "The key cache utilization ratio.",
			},
			"max_connections": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The maximum number of connections that have been in use simultaneously since the server started.",
			},
			"query_cache_size": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.SizeIByte,
				Desc:     "The amount of memory allocated for caching query results.",
			},
			"table_open_cache": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The number of open tables for all threads. Increasing this value increases the number of file descriptors that mysqld requires.",
			},
			"thread_cache_size": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.SizeIByte,
				Desc:     "How many threads the server should cache for reuse. When a client disconnects, the client's threads are put in the cache if there are fewer than thread_cache_size threads there.",
			},

			// binlog
			"Binlog_space_usage_bytes": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.SizeIByte,
				Desc:     inputs.TODO,
			},

			// OPTIONAL_STATUS
			"Binlog_cache_disk_use": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.SizeIByte,
				Desc:     "The number of transactions that used the temporary binary log cache but that exceeded the value of binlog_cache_size and used a temporary file to store statements from the transaction.",
			},
			"Binlog_cache_use": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.SizeIByte,
				Desc:     "The number of transactions that used the binary log cache.",
			},
			"Handler_commit": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The number of internal COMMIT statements.",
			},
			"Handler_delete": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The number of internal DELETE statements.",
			},
			"Handler_prepare": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The number of internal PREPARE statements.",
			},
			"Handler_read_first": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The number of internal READ_FIRST statements.",
			},
			"Handler_read_key": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The number of internal READ_KEY statements.",
			},
			"Handler_read_next": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The number of internal READ_NEXT statements.",
			},
			"Handler_read_prev": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The number of internal READ_PREV statements.",
			},
			"Handler_read_rnd": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The number of internal READ_RND statements.",
			},
			"Handler_read_rnd_next": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The number of internal READ_RND_NEXT statements.",
			},
			"Handler_rollback": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The number of internal ROLLBACK statements.",
			},
			"Handler_update": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The number of internal UPDATE statements.",
			},
			"Handler_write": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The number of internal WRITE statements.",
			},
			"Opened_tables": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The number of tables that have been opened. If Opened_tables is big, your table_open_cache value is probably too small.",
			},
			"Qcache_total_blocks": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The total number of blocks in the query cache.",
			},
			"Qcache_free_blocks": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.SizeIByte,
				Desc:     "The number of free memory blocks in the query cache.",
			},
			"Qcache_free_memory": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.SizeIByte,
				Desc:     "The amount of free memory for the query cache.",
			},
			"Qcache_not_cached": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The number of noncached queries (not cacheable, or not cached due to the query_cache_type setting).",
			},
			"Qcache_queries_in_cache": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The number of queries registered in the query cache.",
			},
			"Select_full_join": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The number of joins that perform table scans because they do not use indexes. If this value is not 0, you should carefully check the indexes of your tables.",
			},
			"Select_full_range_join": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The number of joins that used a range search on a reference table.",
			},
			"Select_range": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The number of joins that used ranges on the first table. This is normally not a critical issue even if the value is quite large.",
			},
			"Select_range_check": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The number of joins without keys that check for key usage after each row. If this is not 0, you should carefully check the indexes of your tables.",
			},
			"Select_scan": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The number of joins that did a full scan of the first table.",
			},
			"Sort_merge_passes": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The number of merge passes that the sort algorithm has had to do. If this value is large, you should consider increasing the value of the sort_buffer_size system variable.",
			},
			"Sort_range": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The number of sorts that were done using ranges.",
			},
			"Sort_rows": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The number of sorted rows.",
			},
			"Sort_scan": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The number of sorts that were done by scanning the table.",
			},
			"Table_locks_immediate": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The number of times that a request for a table lock could be granted immediately.",
			},
			"Threads_cached": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The number of threads in the thread cache.",
			},
			"Threads_created": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The number of threads created to handle connections. If Threads_created is big, you may want to increase the thread_cache_size value.",
			},
		},
		Tags: map[string]interface{}{
			"server": &inputs.TagInfo{
				Desc: "Server addr",
			},
		},
	}
}

// 数据源获取数据
func (m *baseMeasurement) getStatus() error {

	ctx, cancel := context.WithTimeout(context.Background(), m.i.timeoutDuration)
	defer cancel()

	if err := m.i.db.PingContext(ctx); err != nil {
		l.Errorf("connect error %v", err)
		return fmt.Errorf("connect error %v", err)
	}

	globalStatusSql := "SHOW /*!50002 GLOBAL */ STATUS;"
	rows, err := m.i.db.Query(globalStatusSql)
	if err != nil {
		l.Errorf("query error %v", err)
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
	rows, err := m.i.db.Query(variablesSql)
	if err != nil {
		l.Error(err)
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
	rows, err := m.i.db.Query(logSql)
	if err != nil {
		l.Error(err)
		return err
	}
	defer rows.Close()

	var binaryLogSpace int64
	for rows.Next() {
		var key string
		var val *sql.RawBytes = new(sql.RawBytes)

		if err = rows.Scan(&key, val); err != nil {
			l.Warnf("rows.Scan(): %s, ignored", err.Error())
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
			val, err := Conv(value, item.(*inputs.FieldInfo).DataType)
			if err != nil {
				m.i.err = err
				l.Errorf("baseMeasurement metric %v value %v parse error %v", key, value, err)
				return err
			} else {
				m.fields[key] = val
			}
		}
	}

	return nil
}
