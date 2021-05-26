package mysql

import (
	"database/sql"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
	"time"
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
				Desc:     "是否死锁",
			},
			// status
			"lock_timeouts": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "锁超时",
			},
			// status
			"lock_row_lock_current_waits": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "行锁等待",
			},
			// status
			"lock_row_lock_time": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "行锁时间",
			},
			// status
			"lock_row_lock_time_max": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "行锁最大时间",
			},
			// status
			"lock_row_lock_waits": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.DurationMS,
				Desc:     "行锁等待",
			},
			// status
			"lock_row_lock_time_avg": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "行锁平均时间",
			},
			"buffer_pool_size": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "buffer pool size",
			},
			"buffer_pool_reads": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "buffer pool reads",
			},
			"buffer_pool_read_requests": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "buffer pool read requests",
			},
			"buffer_pool_write_requests": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "buffer pool write requests",
			},
			"buffer_pool_wait_free": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "buffer pool wait free",
			},
			"buffer_pool_read_ahead": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "buffer pool read ahead",
			},
			"buffer_pool_read_ahead_evicted": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "buffer pool read ahead evicted",
			},
			"buffer_pool_pages_total": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "buffer pool pages total",
			},
			"buffer_pool_pages_misc": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "buffer pool pages misc",
			},
			"buffer_pool_pages_data": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "buffer pool pages data",
			},
			"buffer_pool_pages_dirty": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "buffer pool pages dirty",
			},
			"buffer_pool_bytes_dirty": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "buffer pool bytes dirty",
			},
			"buffer_pool_pages_free": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "buffer pool pages free",
			},
			"buffer_pages_created": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "buffer pages created",
			},
			"buffer_pages_written": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "buffer pages written",
			},
			"buffer_pages_read": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "buffer pages read",
			},
			"buffer_data_reads": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "buffer data reads",
			},
			"buffer_data_written": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "buffer data written",
			},
			"os_data_reads": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "os data reads",
			},
			"os_data_writes": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "os data writes",
			},
			"os_data_fsyncs": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "os data fsyncs",
			},
			"os_log_bytes_written": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "os log bytes written",
			},
			"os_log_fsyncs": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "os log fsyncs",
			},
			"os_log_pending_fsyncs": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "os log pending fsyncs",
			},
			"os_log_pending_writes": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "os log pending writes",
			},
			"trx_rseg_history_len": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "trx rseg history len",
			},
			"log_waits": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "log waits",
			},
			"log_write_requests": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "log write requests",
			},
			"log_writes": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "log writes",
			},

			"log_padded": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "log padded",
			},
			"adaptive_hash_searches": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "adaptive hash searches",
			},
			"adaptive_hash_searches_btree": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "adaptive hash searches btree",
			},
			"file_num_open_files": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "file num open files",
			},
			"ibuf_merges_insert": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "ibuf merges insert",
			},
			"ibuf_merges_delete_mark": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "ibuf merges delete mark",
			},
			"ibuf_merges_delete": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "ibuf merges delete",
			},
			"ibuf_merges_discard_insert": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "ibuf merges discard insert",
			},
			"ibuf_merges_discard_delete_mark": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "ibuf merges discard delete mark",
			},
			"ibuf_merges_discard_delete": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "ibuf merges discard delete",
			},
			"ibuf_merges": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "ibuf merges",
			},
			"ibuf_size": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "ibuf size",
			},
			"innodb_activity_count": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "innodb activity count",
			},
			"innodb_dblwr_writes": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "innodb dblwr writes",
			},
			"innodb_dblwr_pages_written": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "innodb dblwr pages written",
			},
			"innodb_page_size": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "innodb page size",
			},
			"innodb_rwlock_s_spin_waits": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "innodb rwlock_s spin waits",
			},
			"innodb_rwlock_x_spin_waits": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "innodb rwlock_x spin waits",
			},
			"innodb_rwlock_sx_spin_waits": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "innodb rwlock_sx spin waits",
			},
			"innodb_rwlock_s_spin_rounds": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "innodb rwlock_s spin rounds",
			},
			"innodb_rwlock_x_spin_rounds": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "innodb rwlock_x spin rounds",
			},
			"innodb_rwlock_sx_spin_rounds": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "innodb rwlock_sx spin rounds",
			},
			"innodb_rwlock_s_os_waits": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "innodb rwlock_s os waits",
			},
			"innodb_rwlock_x_os_waits": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "innodb rwlock_x os waits",
			},
			"innodb_rwlock_sx_os_waits": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "innodb rwlock_sx os waits",
			},
			"dml_inserts": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "dml inserts",
			},
			"dml_deletes": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "dml deletes",
			},
			"dml_updates": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "dml updates",
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
	if err := i.db.Ping(); err != nil {
		l.Errorf("db connect error %v", err)
		return nil, err
	}

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
