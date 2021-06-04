package mysql

import (
	"database/sql"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

type innodbMeasurement struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
	ts     time.Time
}

// 生成行协议
func (m *innodbMeasurement) LineProto() (*io.Point, error) {
	return io.MakePoint(m.name, m.tags, m.fields, m.ts)
}

// 指定指标
func (m *innodbMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "mysql_innodb",
		Fields: map[string]interface{}{
			// status
			"lock_deadlocks": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of deadlocks",
			},
			// status
			"lock_timeouts": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of lock timeouts",
			},
			// status
			"lock_row_lock_current_waits": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of row locks currently being waited for (innodb_row_lock_current_waits)",
			},
			// status
			"lock_row_lock_time": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Time spent in acquiring row locks, in milliseconds (innodb_row_lock_time)",
			},
			// status
			"lock_row_lock_time_max": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The maximum time to acquire a row lock, in milliseconds (innodb_row_lock_time_max)",
			},
			// status
			"lock_row_lock_waits": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.DurationMS,
				Desc:     "Number of times a row lock had to be waited for (innodb_row_lock_waits)",
			},
			// status
			"lock_row_lock_time_avg": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The average time to acquire a row lock, in milliseconds (innodb_row_lock_time_avg)",
			},
			"buffer_pool_size": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Server buffer pool size (all buffer pools) in bytes",
			},
			"buffer_pool_reads": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of reads directly from disk (innodb_buffer_pool_reads)",
			},
			"buffer_pool_read_requests": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of logical read requests (innodb_buffer_pool_read_requests)",
			},
			"buffer_pool_write_requests": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of write requests (innodb_buffer_pool_write_requests)",
			},
			"buffer_pool_wait_free": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of times waited for free buffer (innodb_buffer_pool_wait_free)",
			},
			"buffer_pool_read_ahead": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of pages read as read ahead (innodb_buffer_pool_read_ahead)",
			},
			"buffer_pool_read_ahead_evicted": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Read-ahead pages evicted without being accessed (innodb_buffer_pool_read_ahead_evicted)",
			},
			"buffer_pool_pages_total": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Total buffer pool size in pages (innodb_buffer_pool_pages_total)",
			},
			"buffer_pool_pages_misc": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Buffer pages for misc use such as row locks or the adaptive hash index (innodb_buffer_pool_pages_misc)",
			},
			"buffer_pool_pages_data": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Buffer pages containing data (innodb_buffer_pool_pages_data)",
			},
			"buffer_pool_pages_dirty": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Buffer pages currently dirty (innodb_buffer_pool_pages_dirty)",
			},
			"buffer_pool_bytes_dirty": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Buffer bytes containing data (innodb_buffer_pool_bytes_data)",
			},
			"buffer_pool_pages_free": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Buffer pages currently free (innodb_buffer_pool_pages_free)",
			},
			"buffer_pages_created": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of pages created (innodb_pages_created)",
			},
			"buffer_pages_written": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of pages written (innodb_pages_written)",
			},
			"buffer_pages_read": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of pages read (innodb_pages_read)",
			},
			"buffer_data_reads": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Amount of data read in bytes (innodb_data_reads)",
			},
			"buffer_data_written": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Amount of data written in bytes (innodb_data_written)",
			},
			"os_data_reads": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of reads initiated (innodb_data_reads)",
			},
			"os_data_writes": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of writes initiated (innodb_data_writes)",
			},
			"os_data_fsyncs": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of fsync() calls (innodb_data_fsyncs)",
			},
			"os_log_bytes_written": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Bytes of log written (innodb_os_log_written)",
			},
			"os_log_fsyncs": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of fsync log writes (innodb_os_log_fsyncs)",
			},
			"os_log_pending_fsyncs": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of pending fsync write (innodb_os_log_pending_fsyncs)",
			},
			"os_log_pending_writes": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of pending log file writes (innodb_os_log_pending_writes)",
			},
			"trx_rseg_history_len": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Length of the TRX_RSEG_HISTORY list",
			},
			"log_waits": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of log waits due to small log buffer (innodb_log_waits)",
			},
			"log_write_requests": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of log write requests (innodb_log_write_requests)",
			},
			"log_writes": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of log writes (innodb_log_writes)",
			},

			"log_padded": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Bytes of log padded for log write ahead",
			},
			"adaptive_hash_searches": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of successful searches using Adaptive Hash Index",
			},
			"adaptive_hash_searches_btree": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of searches using B-tree on an index search",
			},
			"file_num_open_files": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of files currently open (innodb_num_open_files)",
			},
			"ibuf_merges_insert": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of inserted records merged by change buffering",
			},
			"ibuf_merges_delete_mark": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of deleted records merged by change buffering",
			},
			"ibuf_merges_delete": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of purge records merged by change buffering",
			},
			"ibuf_merges_discard_insert": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of insert merged operations discarded",
			},
			"ibuf_merges_discard_delete_mark": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of deleted merged operations discarded",
			},
			"ibuf_merges_discard_delete": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of purge merged  operations discarded",
			},
			"ibuf_merges": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of change buffer merges",
			},
			"ibuf_size": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Change buffer size in pages",
			},
			"innodb_activity_count": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Current server activity count",
			},
			"innodb_dblwr_writes": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of doublewrite operations that have been performed (innodb_dblwr_writes)",
			},
			"innodb_dblwr_pages_written": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of pages that have been written for doublewrite operations (innodb_dblwr_pages_written)",
			},
			"innodb_page_size": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "InnoDB page size in bytes (innodb_page_size)",
			},
			"innodb_rwlock_s_spin_waits": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of rwlock spin waits due to shared latch request",
			},
			"innodb_rwlock_x_spin_waits": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of rwlock spin waits due to exclusive latch request",
			},
			"innodb_rwlock_sx_spin_waits": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of rwlock spin waits due to sx latch request",
			},
			"innodb_rwlock_s_spin_rounds": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of rwlock spin loop rounds due to shared latch request",
			},
			"innodb_rwlock_x_spin_rounds": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of rwlock spin loop rounds due to exclusive latch request",
			},
			"innodb_rwlock_sx_spin_rounds": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of rwlock spin loop rounds due to sx latch request",
			},
			"innodb_rwlock_s_os_waits": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of OS waits due to shared latch request",
			},
			"innodb_rwlock_x_os_waits": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of OS waits due to exclusive latch request",
			},
			"innodb_rwlock_sx_os_waits": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of OS waits due to sx latch request",
			},
			"dml_inserts": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of rows inserted",
			},
			"dml_deletes": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of rows deleted",
			},
			"dml_updates": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of rows updated",
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
func (i *Input) getInnodb() ([]inputs.Measurement, error) {
	var collectCache []inputs.Measurement

	var globalInnodbSql = `SELECT NAME, COUNT FROM information_schema.INNODB_METRICS WHERE status='enabled'`

	// run query
	rows, err := i.db.Query(globalInnodbSql)
	if err != nil {
		l.Errorf("query error %v", err)
		return nil, err
	}
	defer rows.Close()

	m := &innodbMeasurement{
		tags:   make(map[string]string),
		fields: make(map[string]interface{}),
	}

	m.name = "mysql_innodb"

	for key, value := range i.Tags {
		m.tags[key] = value
	}

	for rows.Next() {
		var key string
		var val *sql.RawBytes = new(sql.RawBytes)
		if err = rows.Scan(&key, val); err != nil {
			continue
		}

		value, err := Conv(string(*val), inputs.Int)
		if err != nil {
			l.Errorf("innodb get value conv error", err)
		} else {
			m.fields[key] = value
		}
	}

	collectCache = append(collectCache, m)

	return collectCache, nil
}
