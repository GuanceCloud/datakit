package mysql

import (
	"database/sql"
	"fmt"
	"github.com/spf13/cast"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
	"strconv"
	"strings"
	"time"
)

type innodbMeasurement struct {
	i       *Input
	name    string
	tags    map[string]string
	fields  map[string]interface{}
	ts      time.Time
	resData map[string]interface{}
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
			"Innodb_mutex_spin_waits": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     inputs.TODO,
			},
			// status
			"Innodb_mutex_os_waits": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     inputs.TODO,
			},
			// status
			"Innodb_s_lock_os_waits": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     inputs.TODO,
			},
			// status
			"Innodb_x_lock_spin_waits": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     inputs.TODO,
			},
			// status
			"Innodb_x_lock_spin_rounds": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     inputs.TODO,
			},
			// status
			"Innodb_history_list_length": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     inputs.TODO,
			},
			// status
			"Innodb_row_lock_time": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.DurationMS,
				Desc:     inputs.TODO,
			},
			// status
			"Innodb_read_views": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     inputs.TODO,
			},
			"Innodb_tables_in_use": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     inputs.TODO,
			},
			"Innodb_locked_tables": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     inputs.TODO,
			},
			"Innodb_lock_structs": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     inputs.TODO,
			},
			"Innodb_locked_transactions": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     inputs.TODO,
			},

			"Innodb_os_file_reads": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     inputs.TODO,
			},
			"Innodb_os_file_writes": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     inputs.TODO,
			},
			"Innodb_os_file_fsyncs": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     inputs.TODO,
			},
			"Innodb_pending_normal_aio_reads": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     inputs.TODO,
			},
			"Innodb_pending_normal_aio_writes": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     inputs.TODO,
			},

			"Innodb_pending_ibuf_aio_reads": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     inputs.TODO,
			},
			"Innodb_pending_aio_log_ios": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     inputs.TODO,
			},
			"Innodb_pending_aio_sync_ios": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     inputs.TODO,
			},
			"Innodb_pending_log_flushes": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     inputs.TODO,
			},
			"Innodb_pending_buffer_pool_flushes": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     inputs.TODO,
			},
			"Innodb_ibuf_size": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     inputs.TODO,
			},
			"Innodb_ibuf_free_list": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     inputs.TODO,
			},
			"Innodb_ibuf_segment_size": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     inputs.TODO,
			},
			"Innodb_ibuf_merges": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     inputs.TODO,
			},
			"Innodb_ibuf_merged_delete_marks": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     inputs.TODO,
			},
			"Innodb_ibuf_merged_deletes": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     inputs.TODO,
			},
			"Innodb_ibuf_merged_inserts": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     inputs.TODO,
			},
			"Innodb_ibuf_merged": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     inputs.TODO,
			},
			"Innodb_hash_index_cells_total": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     inputs.TODO,
			},
			"Innodb_hash_index_cells_used": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     inputs.TODO,
			},
			"Innodb_log_writes": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     inputs.TODO,
			},
			"Innodb_pending_log_writes": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     inputs.TODO,
			},
			"Innodb_pending_checkpoint_writes": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     inputs.TODO,
			},
			"Innodb_lsn_current": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     inputs.TODO,
			},
			"Innodb_lsn_flushed": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     inputs.TODO,
			},

			"Innodb_lsn_last_checkpoint": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     inputs.TODO,
			},
			"Innodb_mem_total": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     inputs.TODO,
			},
			"Innodb_mem_additional_pool": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     inputs.TODO,
			},
			"Innodb_mem_adaptive_hash": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     inputs.TODO,
			},
			"Innodb_mem_page_hash": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     inputs.TODO,
			},
			"Innodb_mem_dictionary": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     inputs.TODO,
			},
			"Innodb_mem_file_system": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     inputs.TODO,
			},
			"Innodb_mem_lock_system": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     inputs.TODO,
			},
			"Innodb_mem_recovery_system": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     inputs.TODO,
			},
			"Innodb_mem_thread_hash": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     inputs.TODO,
			},
			"Innodb_buffer_pool_pages_free": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     inputs.TODO,
			},
			"Innodb_buffer_pool_pages_data": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     inputs.TODO,
			},
			"Innodb_buffer_pool_pages_total": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     inputs.TODO,
			},
			"Innodb_buffer_pool_pages_dirty": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     inputs.TODO,
			},
			"Innodb_buffer_pool_read_ahead": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     inputs.TODO,
			},
			"Innodb_buffer_pool_read_ahead_evicted": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     inputs.TODO,
			},
			"Innodb_buffer_pool_read_ahead_rnd": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     inputs.TODO,
			},
			"Innodb_pages_read": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     inputs.TODO,
			},
			"Innodb_pages_created": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     inputs.TODO,
			},
			"Innodb_pages_written": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     inputs.TODO,
			},
			"Innodb_rows_inserted": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     inputs.TODO,
			},
			"Innodb_rows_updated": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     inputs.TODO,
			},
			"Innodb_rows_deleted": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     inputs.TODO,
			},
			"Innodb_rows_read": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     inputs.TODO,
			},
			"Innodb_queries_inside": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     inputs.TODO,
			},
			"Innodb_queries_queued": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     inputs.TODO,
			},
			"Innodb_checkpoint_age": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     inputs.TODO,
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
func (m *innodbMeasurement) getInnodb() error {
	if err := m.i.db.Ping(); err != nil {
		l.Errorf("db connect error %v", err)
		return err
	}

	var globalInnodbSql = `SHOW /*!50000 ENGINE */ INNODB STATUS`

	// run query
	var t, name, stat string
	err := m.i.db.QueryRow(globalInnodbSql).Scan(&t, &name, &stat)
	if err != nil {
		l.Warnf("Privilege error or engine unavailable accessing the INNODB status tables (must grant PROCESS): %s", err)
		return err
	}

	if len(stat) > 0 {
		return m.parseInnodbStatus(stat)
	}

	return nil
}

func (m *innodbMeasurement) parseInnodbStatus(str string) error {
	isTransaction := false
	prevLine := ""

	for _, line := range strings.Split(str, "\n") {
		record := strings.Fields(line)

		// Innodb Semaphores
		if strings.Index(line, "Mutex spin waits") == 0 {
			// Mutex spin waits 79626940, rounds 157459864, OS waits 698719
			// Mutex spin waits 0, rounds 247280272495, OS waits 316513438
			increaseMap(m.resData, "Innodb_mutex_spin_waits", record[3])
			increaseMap(m.resData, "Innodb_mutex_spin_rounds", record[5])
			increaseMap(m.resData, "Innodb_mutex_os_waits", record[8])
			continue
		}
		if strings.Index(line, "RW-shared spins") == 0 && strings.Index(line, ";") > 0 {
			// RW-shared spins 3859028, OS waits 2100750; RW-excl spins
			// 4641946, OS waits 1530310
			increaseMap(m.resData, "Innodb_mutex_os_waits", record[2])
			increaseMap(m.resData, "Innodb_s_lock_os_waits", record[5])
			increaseMap(m.resData, "Innodb_x_lock_spin_waits", record[8])
			increaseMap(m.resData, "Innodb_x_lock_os_waits", record[11])
			continue
		}
		if strings.Index(line, "RW-shared spins") == 0 && strings.Index(line, "; RW-excl spins") < 0 {
			// Post 5.5.17 SHOW ENGINE INNODB STATUS syntax
			// RW-shared spins 604733, rounds 8107431, OS waits 241268
			increaseMap(m.resData, "Innodb_s_lock_spin_waits", record[2])
			increaseMap(m.resData, "Innodb_s_lock_spin_rounds", record[4])
			increaseMap(m.resData, "Innodb_s_lock_os_waits", record[7])
			continue
		}
		if strings.Index(line, "RW-excl spins") == 0 {
			// Post 5.5.17 SHOW ENGINE INNODB STATUS syntax
			// RW-excl spins 604733, rounds 8107431, OS waits 241268
			increaseMap(m.resData, "Innodb_x_lock_spin_waits", record[2])
			increaseMap(m.resData, "Innodb_x_lock_spin_rounds", record[4])
			increaseMap(m.resData, "Innodb_x_lock_os_waits", record[7])
			continue
		}
		if strings.Index(line, "seconds the semaphore:") > 0 {
			// --Thread 907205 has waited at handler/ha_innodb.cc line 7156 for 1.00 seconds the semaphore:
			increaseMap(m.resData, "Innodb_semaphore_waits", "1")
			wait := atof(record[9])
			wait = wait * 1000
			increaseMap(m.resData, "Innodb_mutex_spin_waits", fmt.Sprintf("%.f", wait))
			continue
		}

		// Innodb Transactions
		if strings.Index(line, "Trx id counter") == 0 {
			// The beginning of the TRANSACTIONS section: start counting
			// transactions
			// Trx id counter 0 1170664159
			// Trx id counter 861B144C
			isTransaction = true
			continue
		}
		if strings.Index(line, "History list length") == 0 {
			// History list length 132
			increaseMap(m.resData, "Innodb_history_list_length", record[3])
			continue
		}
		if isTransaction && strings.Index(line, "---TRANSACTION") == 0 {
			// ---TRANSACTION 0, not started, process no 13510, OS thread id 1170446656
			increaseMap(m.resData, "Innodb_current_transactions", "1")
			if strings.Index(line, "ACTIVE") > 0 {
				increaseMap(m.resData, "Innodb_active_transactions", "1")
			}
			continue
		}
		if isTransaction && strings.Index(line, "------- TRX HAS BEEN") == 0 {
			// ------- TRX HAS BEEN WAITING 32 SEC FOR THIS LOCK TO BE GRANTED:
			increaseMap(m.resData, "Innodb_row_lock_time", "1")
			continue
		}
		if strings.Index(line, "read views open inside InnoDB") > 0 {
			// 1 read views open inside InnoDB
			m.resData["Innodb_read_views"] = atof(record[0])
			continue
		}
		if strings.Index(line, "mysql tables in use") == 0 {
			// mysql tables in use 2, locked 2
			increaseMap(m.resData, "Innodb_tables_in_use", record[4])
			increaseMap(m.resData, "Innodb_locked_tables", record[6])
			continue
		}
		if isTransaction && strings.Index(line, "lock struct(s)") == 0 {
			// 23 lock struct(s), heap size 3024, undo log entries 27
			// LOCK WAIT 12 lock struct(s), heap size 3024, undo log entries 5
			// LOCK WAIT 2 lock struct(s), heap size 368
			if strings.Index(line, "LOCK WAIT") > 0 {
				increaseMap(m.resData, "Innodb_lock_structs", record[2])
				increaseMap(m.resData, "Innodb_locked_transactions", "1")
			} else if strings.Index(line, "ROLLING BACK") > 0 {
				// ROLLING BACK 127539 lock struct(s), heap size 15201832,
				// 4411492 row lock(s), undo log entries 1042488
				increaseMap(m.resData, "Innodb_lock_structs", record[0])
			} else {
				increaseMap(m.resData, "Innodb_lock_structs", record[0])
			}
			continue
		}

		// File I/O
		if strings.Index(line, " OS file reads, ") > 0 {
			// 8782182 OS file reads, 15635445 OS file writes, 947800 OS
			// fsyncs
			m.resData["Innodb_os_file_reads"] = atof(record[0])
			m.resData["Innodb_os_file_writes"] = atof(record[4])
			m.resData["Innodb_os_file_fsyncs"] = atof(record[8])
			continue
		}
		if strings.Index(line, "Pending normal aio reads:") == 0 {
			// Pending normal aio reads: 0, aio writes: 0,
			// or Pending normal aio reads: [0, 0, 0, 0] , aio writes: [0, 0, 0, 0] ,
			// or Pending normal aio reads: 0 [0, 0, 0, 0] , aio writes: 0 [0, 0, 0, 0] ,
			if len(record) == 16 {
				m.resData["Innodb_pending_normal_aio_reads"] = atof(record[4]) + atof(record[5]) + atof(record[6]) + atof(record[7])
				m.resData["Innodb_pending_normal_aio_writes"] = atof(record[11]) + atof(record[12]) + atof(record[13]) + atof(record[14])
			} else if len(record) == 18 {
				m.resData["Innodb_pending_normal_aio_reads"] = atof(record[4])
				m.resData["Innodb_pending_normal_aio_writes"] = atof(record[12])
			} else {
				m.resData["Innodb_pending_normal_aio_reads"] = atof(record[4])
				m.resData["Innodb_pending_normal_aio_writes"] = atof(record[7])
			}
			continue
		}
		if strings.Index(line, " ibuf aio reads") == 0 {
			// ibuf aio reads: 0, log i/o's: 0, sync i/o's: 0
			// or ibuf aio reads:, log i/o's:, sync i/o's:
			if len(record) == 10 {
				m.resData["Innodb_pending_ibuf_aio_reads"] = atof(record[3])
				m.resData["Innodb_pending_aio_log_ios"] = atof(record[6])
				m.resData["Innodb_pending_aio_sync_ios"] = atof(record[9])
			} else if len(record) == 7 {
				m.resData["Innodb_pending_ibuf_aio_reads"] = 0
				m.resData["Innodb_pending_aio_log_ios"] = 0
				m.resData["Innodb_pending_aio_sync_ios"] = 0
			}
			continue
		}
		if strings.Index(line, "Pending flushes (fsync)") == 0 {
			// Pending flushes (fsync) log: 0; buffer pool: 0
			m.resData["Innodb_pending_log_flushes"] = atof(record[4])
			m.resData["Innodb_pending_buffer_pool_flushes"] = atof(record[7])
			continue
		}

		// Insert Buffer and Adaptive Hash Index
		if strings.Index(line, "Ibuf for space 0: size ") == 0 {
			// Older InnoDB code seemed to be ready for an ibuf per tablespace.  It
			// had two lines in the output.  Newer has just one line, see below.
			// Ibuf for space 0: size 1, free list len 887, seg size 889, is not empty
			// Ibuf for space 0: size 1, free list len 887, seg size 889,
			m.resData["Innodb_ibuf_size"] = atof(record[5])
			m.resData["Innodb_ibuf_free_list"] = atof(record[9])
			m.resData["Innodb_ibuf_segment_size"] = atof(record[12])
			continue
		}
		if strings.Index(line, "Ibuf: size ") == 0 {
			// Ibuf: size 1, free list len 4634, seg size 4636,
			m.resData["Innodb_ibuf_size"] = atof(record[2])
			m.resData["Innodb_ibuf_free_list"] = atof(record[6])
			m.resData["Innodb_ibuf_segment_size"] = atof(record[9])
			if strings.Index(line, "merges") > 0 {
				m.resData["Innodb_ibuf_merges"] = atof(record[10])
			}
			continue
		}
		if strings.Index(line, ", delete mark ") > 0 && strings.Index(prevLine, "merged operations:") == 0 {
			// Output of show engine innodb status has changed in 5.5
			// merged operations:
			// insert 593983, delete mark 387006, delete 73092
			v1 := atof(record[1])
			v2 := atof(record[4])
			v3 := atof(record[6])
			m.resData["Innodb_ibuf_merged_inserts"] = v1
			m.resData["Innodb_ibuf_merged_delete_marks"] = v2
			m.resData["Innodb_ibuf_merged_deletes"] = v3
			m.resData["Innodb_ibuf_merged"] = v1 + v2 + v3
			continue
		}
		if strings.Index(line, " merged recs, ") > 0 {
			// 19817685 inserts, 19817684 merged recs, 3552620 merges
			m.resData["Innodb_ibuf_merged_inserts"] = atof(record[0])
			m.resData["Innodb_ibuf_merged"] = atof(record[2])
			m.resData["Innodb_ibuf_merges"] = atof(record[5])
			continue
		}
		if strings.Index(line, "Hash table size ") == 0 {
			// In some versions of InnoDB, the used cells is omitted.
			// Hash table size 4425293, used cells 4229064, ....
			// Hash table size 57374437, node heap has 72964 buffer(s) <--
			// no used cells
			m.resData["Innodb_hash_index_cells_total"] = atof(record[3])
			if strings.Index(line, "used cells") > 0 {
				m.resData["Innodb_hash_index_cells_used"] = atof(record[6])
			} else {
				m.resData["Innodb_hash_index_cells_used"] = 0
			}
			continue
		}

		// Log
		if strings.Index(line, " log i/o's done, ") > 0 {
			// 3430041 log i/o's done, 17.44 log i/o's/second
			// 520835887 log i/o's done, 17.28 log i/o's/second, 518724686
			// syncs, 2980893 checkpoints
			m.resData["Innodb_log_writes"] = atof(record[0])
			continue
		}
		if strings.Index(line, " pending log writes, ") > 0 {
			// 0 pending log writes, 0 pending chkp writes
			m.resData["Innodb_pending_log_writes"] = atof(record[0])
			m.resData["Innodb_pending_checkpoint_writes"] = atof(record[4])
			continue
		}
		if strings.Index(line, "Log sequence number") == 0 {
			// This number is NOT printed in hex in InnoDB plugin.
			// Log sequence number 272588624
			val := atof(record[3])
			if len(record) >= 5 {
				val = float64(makeBigint(record[3], record[4]))
			}
			m.resData["Innodb_lsn_current"] = val
			continue
		}
		if strings.Index(line, "Log flushed up to") == 0 {
			// This number is NOT printed in hex in InnoDB plugin.
			// Log flushed up to   272588624
			val := atof(record[4])
			if len(record) >= 6 {
				val = float64(makeBigint(record[4], record[5]))
			}
			m.resData["Innodb_lsn_flushed"] = val
			continue
		}
		if strings.Index(line, "Last checkpoint at") == 0 {
			// Last checkpoint at  272588624
			val := atof(record[3])
			if len(record) >= 5 {
				val = float64(makeBigint(record[3], record[4]))
			}
			m.resData["Innodb_lsn_last_checkpoint"] = val
			continue
		}

		// Buffer Pool and Memory
		// 5.6 or before
		if strings.Index(line, "Total memory allocated") == 0 && strings.Index(line, "in additional pool allocated") > 0 {
			// Total memory allocated 29642194944; in additional pool allocated 0
			// Total memory allocated by read views 96
			m.resData["Innodb_mem_total"] = atof(record[3])
			m.resData["Innodb_mem_additional_pool"] = atof(record[8])
			continue
		}
		if strings.Index(line, "Adaptive hash index ") == 0 {
			// Adaptive hash index 1538240664     (186998824 + 1351241840)
			v := atof(record[3])
			setIfEmpty(m.resData, "Innodb_mem_adaptive_hash", v)
			continue
		}
		if strings.Index(line, "Page hash           ") == 0 {
			// Page hash           11688584
			v := atof(record[2])
			setIfEmpty(m.resData, "Innodb_mem_page_hash", v)
			continue
		}
		if strings.Index(line, "Dictionary cache    ") == 0 {
			// Dictionary cache    145525560      (140250984 + 5274576)
			v := atof(record[2])
			setIfEmpty(m.resData, "Innodb_mem_dictionary", v)
			continue
		}
		if strings.Index(line, "File system         ") == 0 {
			// File system         313848         (82672 + 231176)
			v := atof(record[2])
			setIfEmpty(m.resData, "Innodb_mem_file_system", v)
			continue
		}
		if strings.Index(line, "Lock system         ") == 0 {
			// Lock system         29232616       (29219368 + 13248)
			v := atof(record[2])
			setIfEmpty(m.resData, "Innodb_mem_lock_system", v)
			continue
		}
		if strings.Index(line, "Recovery system     ") == 0 {
			// Recovery system     0      (0 + 0)
			v := atof(record[2])
			setIfEmpty(m.resData, "Innodb_mem_recovery_system", v)
			continue
		}
		if strings.Index(line, "Threads             ") == 0 {
			// Threads             409336         (406936 + 2400)
			v := atof(record[1])
			setIfEmpty(m.resData, "Innodb_mem_thread_hash", v)
			continue
		}
		if strings.Index(line, "Buffer pool size ") == 0 {
			// The " " after size is necessary to avoid matching the wrong line:
			// Buffer pool size        1769471
			// Buffer pool size, bytes 28991012864
			v := atof(record[3])
			setIfEmpty(m.resData, "Innodb_buffer_pool_pages_total", v)
			continue
		}
		if strings.Index(line, "Free buffers") == 0 {
			// Free buffers            0
			v := atof(record[2])
			setIfEmpty(m.resData, "Innodb_buffer_pool_pages_free", v)
			continue
		}
		if strings.Index(line, "Database pages") == 0 {
			// Database pages          1696503
			v := atof(record[2])
			setIfEmpty(m.resData, "Innodb_buffer_pool_pages_data", v)
			continue
		}
		if strings.Index(line, "Modified db pages") == 0 {
			// Modified db pages       160602
			v := atof(record[3])
			setIfEmpty(m.resData, "Innodb_buffer_pool_pages_dirty", v)
			continue
		}
		if strings.Index(line, "Pages read ahead") == 0 {
			v := atof(record[3])
			setIfEmpty(m.resData, "Innodb_buffer_pool_read_ahead", v)
			v = atof(record[7])
			setIfEmpty(m.resData, "Innodb_buffer_pool_read_ahead_evicted", v)
			v = atof(record[11])
			setIfEmpty(m.resData, "Innodb_buffer_pool_read_ahead_rnd", v)
			continue
		}
		if strings.Index(line, "Pages read") == 0 {
			// Pages read 15240822, created 1770238, written 21705836
			v := atof(record[2])
			setIfEmpty(m.resData, "Innodb_pages_read", v)
			v = atof(record[4])
			setIfEmpty(m.resData, "Innodb_pages_created", v)
			v = atof(record[6])
			setIfEmpty(m.resData, "Innodb_pages_written", v)
			continue
		}

		// Row Operations
		if strings.Index(line, "Number of rows inserted") == 0 {
			// Number of rows inserted 50678311, updated 66425915, deleted
			// 20605903, read 454561562
			m.resData["Innodb_rows_inserted"] = atof(record[4])
			m.resData["Innodb_rows_updated"] = atof(record[6])
			m.resData["Innodb_rows_deleted"] = atof(record[8])
			m.resData["Innodb_rows_read"] = atof(record[10])
			continue
		}
		if strings.Index(line, " queries inside InnoDB, ") > 0 {
			// 0 queries inside InnoDB, 0 queries in queue
			m.resData["Innodb_queries_inside"] = atof(record[0])
			m.resData["Innodb_queries_queued"] = atof(record[4])
			continue
		}

		// for next loop
		prevLine = line
	}

	// We need to calculate this metric separately
	m.resData["Innodb_checkpoint_age"] = cast.ToFloat64(m.resData["Innodb_lsn_current"]) - cast.ToFloat64(m.resData["Innodb_lsn_last_checkpoint"])

	return nil
}

func increaseMap(p map[string]interface{}, key string, src string) {
	val := atof(src)
	_, exists := p[key]
	if !exists {
		p[key] = val
		return
	}
	p[key] = cast.ToFloat64(p[key]) + val
}

func setIfEmpty(p map[string]interface{}, key string, val float64) {
	_, ok := p[key]
	if !ok {
		p[key] = val
	}
}

// parseValue can be used to convert values such as "ON","OFF","Yes","No" to 0,1
func parseValue(value sql.RawBytes) (float64, bool) {
	val := strings.ToLower(string(value))
	if val == "yes" || val == "on" {
		return 1, true
	}

	if val == "no" || val == "off" {
		return 0, true
	}
	n, err := strconv.ParseFloat(val, 64)
	return n, err == nil
}

// 提交数据
func (m *innodbMeasurement) submit() error {
	metricInfo := m.Info()
	for key, item := range metricInfo.Fields {
		if value, ok := m.resData[key]; ok {
			val, err := Conv(value, item.(*inputs.FieldInfo).DataType)
			if err != nil {
				m.i.err = err
				l.Errorf("innodbMeasurement metric %v value %v parse error %v", key, value, err)
			} else {
				m.fields[key] = val
			}
		}
	}

	return nil
}
