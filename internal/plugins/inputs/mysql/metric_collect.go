// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package mysql

import (
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/hashcode"
)

func hasKey(m map[string]interface{}, key string) bool {
	if m == nil {
		return false
	}

	if _, ok := m[key]; ok {
		return true
	}

	return false
}

// getStringValue transform int or float string number to int64.
func getIntFromString(textValue string) int64 {
	if v, err := strconv.ParseInt(textValue, 10, 64); err != nil {
		if v, err := strconv.ParseFloat(textValue, 64); err != nil {
			l.Warnf("transform value %s to int error: %s", textValue, err.Error())
		} else {
			return int64(v)
		}
	} else {
		return v
	}

	return 0
}

// allNumeric return true if all the elements in arr are number strings, or return false.
func allNumeric(arr []string) bool {
	if len(arr) == 0 {
		return false
	}
	for _, v := range arr {
		if _, err := strconv.ParseInt(v, 10, 64); err != nil {
			return false
		}
	}
	return true
}

// collect metric

func (ipt *Input) collectMysql() error {
	var err error

	ipt.globalStatus = map[string]interface{}{}
	ipt.globalVariables = map[string]interface{}{}
	ipt.binlog = map[string]interface{}{}

	// We should first collect global MySQL metrics
	if res := globalStatusMetrics(ipt.q("SHOW /*!50002 GLOBAL */ STATUS;")); res != nil {
		ipt.globalStatus = res
	} else {
		l.Warn("collect_show_status_failed")
	}

	if res := globalVariablesMetrics(ipt.q("SHOW GLOBAL VARIABLES;")); res != nil {
		ipt.globalVariables = res

		// Detect if binlog enabled
		switch v := ipt.globalVariables["log_bin"].(type) {
		case string:
			ipt.binLogOn = (v == "on" || v == "ON")
		default:
			ipt.binLogOn = false
		}
	} else {
		l.Warn("collect_show_variables_failed")
	}

	if ipt.binLogOn {
		if res := binlogMetrics(ipt.q("SHOW BINARY LOGS;")); res != nil {
			ipt.binlog = res
		} else {
			l.Warn("collect_show_binlog_failed")
		}
	}

	return err
}

func (ipt *Input) collectMysqlSchema() error {
	ipt.mSchemaSize = map[string]interface{}{}
	ipt.mSchemaQueryExecTime = map[string]interface{}{}

	querySizePerschemaSQL := `
		SELECT   table_schema, IFNULL(SUM(data_length+index_length)/1024/1024,0) AS total_mb
		FROM     information_schema.tables
		GROUP BY table_schema;
	`
	if res := getCleanSchemaData(ipt.q(querySizePerschemaSQL)); res != nil {
		ipt.mSchemaSize = res
	} else {
		l.Warn("collect_schema_size_failed")
	}

	queryExecPerTimeSQL := `
	SELECT schema_name, ROUND((SUM(sum_timer_wait) / SUM(count_star)) / 1000000) AS avg_us
	FROM performance_schema.events_statements_summary_by_digest
	WHERE schema_name IS NOT NULL
	GROUP BY schema_name;
	`
	if res := getCleanSchemaData(ipt.q(queryExecPerTimeSQL)); res != nil {
		ipt.mSchemaQueryExecTime = res
	} else {
		l.Warn("collect_schema_failed")
	}

	return nil
}

func (ipt *Input) collectMysqlInnodb() error {
	ipt.mInnodb = map[string]interface{}{}

	globalInnodbSQL := `SELECT NAME, COUNT FROM information_schema.INNODB_METRICS WHERE status='enabled'`

	if res := getCleanInnodb(ipt.q(globalInnodbSQL)); res != nil {
		ipt.mInnodb = res
	} else {
		l.Warnf("collect_innodb_failed")
	}

	if res, err := ipt.getInnodbStatus(); err != nil {
		l.Warnf("collect_innodb_status_failed: %s", err.Error())
	} else {
		for k, v := range res {
			ipt.mInnodb[k] = v
		}
	}

	// extract metric from i.globalStatus
	for k, v := range ipt.globalStatus {
		prefix := "Innodb_"
		if strings.HasPrefix(k, prefix) {
			newKey := strings.TrimPrefix(k, prefix)
			ipt.mInnodb[newKey] = v
		}
	}

	if hasKey(ipt.globalStatus, "Innodb_page_size") {
		var bufferPoolPagesUsed int64
		if innodbPageSize, ok := ipt.globalStatus["Innodb_page_size"].(int64); ok {
			if hasKey(ipt.mInnodb, "buffer_pool_pages_total") && hasKey(ipt.mInnodb, "buffer_pool_pages_free") {
				total, totalOK := ipt.mInnodb["buffer_pool_pages_total"].(int64)
				free, freeOK := ipt.mInnodb["buffer_pool_pages_free"].(int64)
				if totalOK && freeOK {
					bufferPoolPagesUsed = total - free
				}

				if !hasKey(ipt.mInnodb, "buffer_pool_pages_utilization") && total != 0 {
					ipt.mInnodb["buffer_pool_pages_utilization"] = float64(bufferPoolPagesUsed) / float64(total)
				}

				if !hasKey(ipt.mInnodb, "buffer_pool_bytes_data") {
					if hasKey(ipt.mInnodb, "buffer_pool_pages_data") {
						data, ok := ipt.mInnodb["buffer_pool_pages_data"].(int64)
						if ok {
							ipt.mInnodb["buffer_pool_bytes_data"] = data * innodbPageSize
						}
					}
				}

				if !hasKey(ipt.mInnodb, "buffer_pool_bytes_dirty") {
					if hasKey(ipt.mInnodb, "buffer_pool_pages_dirty") {
						data, ok := ipt.mInnodb["buffer_pool_pages_dirty"].(int64)
						if ok {
							ipt.mInnodb["buffer_pool_bytes_dirty"] = data * innodbPageSize
						}
					}
				}

				if !hasKey(ipt.mInnodb, "buffer_pool_bytes_free") {
					if hasKey(ipt.mInnodb, "buffer_pool_pages_free") {
						data, ok := ipt.mInnodb["buffer_pool_pages_free"].(int64)
						if ok {
							ipt.mInnodb["buffer_pool_bytes_free"] = data * innodbPageSize
						}
					}
				}

				if !hasKey(ipt.mInnodb, "buffer_pool_bytes_total") {
					if hasKey(ipt.mInnodb, "buffer_pool_pages_total") {
						data, ok := ipt.mInnodb["buffer_pool_pages_total"].(int64)
						if ok {
							ipt.mInnodb["buffer_pool_bytes_total"] = data * innodbPageSize
						}
					}
				}

				if !hasKey(ipt.mInnodb, "buffer_pool_bytes_used") {
					ipt.mInnodb["buffer_pool_bytes_used"] = bufferPoolPagesUsed * innodbPageSize
				}
			}
		}
	}

	return nil
}

var whiteSpaceReg = regexp.MustCompile(" +")

//nolint:funlen
func (ipt *Input) getInnodbStatus() (statusRes map[string]int64, err error) {
	innodbStatusSQL := "SHOW /*!50000 ENGINE*/ INNODB STATUS"
	if res, err := ipt.getQueryRows(innodbStatusSQL); err != nil {
		return statusRes, err
	} else {
		if len(res.rows) == 0 {
			return statusRes, fmt.Errorf("'SHOW ENGINE INNODB STATUS' returned no data")
		} else {
			statusRow := res.rows[0]
			if len(statusRow) != 3 {
				return statusRes, fmt.Errorf("get innodb status failed: invalid column length, expect 3 but got %d", len(statusRow))
			}
			statusTextRaw := statusRow[2]
			statusText := ""
			if textRaw, ok := (*statusTextRaw).([]uint8); !ok {
				return statusRes, fmt.Errorf("get innodb status failed: transform status text error")
			} else {
				statusText = string(textRaw)
			}

			txnSeen := false
			prevLine := ""
			bufferID := int64(-1)

			textList := strings.Split(statusText, "\n")
			results := map[string]int64{
				"semaphore_waits":      0,
				"semaphore_wait_time":  0,
				"current_transactions": 0,
				"active_transactions":  0,
				"tables_in_use":        0,
				"locked_tables":        0,
				"lock_structs":         0,
				"locked_transactions":  0,
			}

			for _, line := range textList {
				line = strings.TrimSpace(line)
				row := whiteSpaceReg.Split(line, -1)

				for i := range row {
					row[i] = strings.Trim(row[i], ",")
					row[i] = strings.Trim(row[i], ";")
					row[i] = strings.Trim(row[i], "[")
					row[i] = strings.Trim(row[i], "]")
				}

				if strings.HasPrefix(line, "---BUFFER POOL") {
					bufferID = getIntFromString(row[2])
				}

				switch {
				// SEMAPHORES
				case strings.HasPrefix(line, "Mutex spin waits"):
					// Mutex spin waits 79626940, rounds 157459864, OS waits 698719
					// Mutex spin waits 0, rounds 247280272495, OS waits 316513438
					results["mutex_spin_waits"] = getIntFromString(row[3])
					results["mutex_spin_rounds"] = getIntFromString(row[5])
					results["mutex_os_waits"] = getIntFromString(row[8])
				case strings.HasPrefix(line, "RW-shared spins") && strings.Index(line, ";") > 0:
					// RW-shared spins 3859028, OS waits 2100750; RW-excl spins
					// 4641946, OS waits 1530310
					results["s_lock_spin_waits"] = getIntFromString(row[2])
					results["x_lock_spin_waits"] = getIntFromString(row[8])
					results["s_lock_os_waits"] = getIntFromString(row[5])
					results["x_lock_os_waits"] = getIntFromString(row[11])
				case strings.HasPrefix(line, "RW-shared spins") && !strings.Contains(line, "; RW-excl spins"):
					// Post 5.5.17 SHOW ENGINE INNODB STATUS syntax
					// RW-shared spins 604733, rounds 8107431, OS waits 241268
					results["s_lock_spin_waits"] = getIntFromString(row[2])
					results["s_lock_spin_rounds"] = getIntFromString(row[4])
					results["s_lock_os_waits"] = getIntFromString(row[7])
				case strings.HasPrefix(line, "RW-excl spins"):
					// Post 5.5.17 SHOW ENGINE INNODB STATUS syntax
					// RW-excl spins 604733, rounds 8107431, OS waits 241268
					results["x_lock_spin_waits"] = getIntFromString(row[2])
					results["x_lock_spin_rounds"] = getIntFromString(row[4])
					results["x_lock_os_waits"] = getIntFromString(row[7])
				case strings.Index(line, "seconds the semaphore:") > 0:
					// --Thread 907205 has waited at handler/ha_innodb.cc line 7156 for 1.00 seconds the semaphore:
					results["semaphore_waits"] += 1
					v := getIntFromString(row[9])
					results["semaphore_wait_time"] += v * 1000
				// TRANSACTIONS
				case strings.HasPrefix(line, "Trx id counter"):
					// The beginning of the TRANSACTIONS section: start counting
					// transactions
					// Trx id counter 0 1170664159
					// Trx id counter 861B144C
					txnSeen = true
				case strings.HasPrefix(line, "History list length"):
					// History list length 132
					results["history_list_length"] = getIntFromString(row[3])
				case txnSeen && strings.HasPrefix(line, "---TRANSACTION"):
					// ---TRANSACTION 0, not started, process no 13510, OS thread id 1170446656
					results["current_transactions"] += 1
					if strings.Index(line, "ACTIVE") > 0 {
						results["active_transactions"] += 1
					}
				case strings.Index(line, "read views open inside InnoDB") > 0:
					// 1 read views open inside InnoDB
					results["read_views"] = getIntFromString(row[0])
				case strings.HasPrefix(line, "mysql tables in use"):
					// mysql tables in use 2, locked 2
					results["tables_in_use"] += getIntFromString(row[4])
					results["locked_tables"] += getIntFromString(row[6])

				case txnSeen && strings.Index(line, "lock struct(s)") > 0:
					// 23 lock struct(s), heap size 3024, undo log entries 27
					// LOCK WAIT 12 lock struct(s), heap size 3024, undo log entries 5
					// LOCK WAIT 2 lock struct(s), heap size 368
					switch {
					case strings.HasPrefix(line, "LOCK WAIT"):
						results["lock_structs"] += getIntFromString(row[2])
						results["locked_transactions"] += 1
					case strings.HasPrefix(line, "ROLLING BACK"):
						// ROLLING BACK 127539 lock struct(s), heap size 15201832,
						// 4411492 row lock(s), undo log entries 1042488
						results["lock_structs"] += getIntFromString(row[2])
					default:
						results["lock_structs"] += getIntFromString(row[0])
					}
					// FILE I/O
				case strings.Index(line, " OS file reads, ") > 0:
					// 8782182 OS file reads, 15635445 OS file writes, 947800 OS
					// fsyncs
					results["os_file_reads"] = getIntFromString(row[0])
					results["os_file_writes"] = getIntFromString(row[4])
					results["os_file_fsyncs"] = getIntFromString(row[8])
				case strings.HasPrefix(line, "Pending normal aio reads:"):
					switch {
					case len(row) == 8:
						// (len(row) == 8)  Pending normal aio reads: 0, aio writes: 0,
						results["pending_normal_aio_reads"] = getIntFromString(row[4])
						results["pending_normal_aio_writes"] = getIntFromString(row[7])
					case len(row) == 14:
						// (len(row) == 14) Pending normal aio reads: 0 [0, 0] , aio writes: 0 [0, 0] ,
						results["pending_normal_aio_reads"] = getIntFromString(row[4])
						results["pending_normal_aio_writes"] = getIntFromString(row[10])
					case len(row) == 16:
						// (len(row) == 16) Pending normal aio reads: [0, 0, 0, 0] , aio writes: [0, 0, 0, 0] ,
						switch {
						case allNumeric(row[4:8]) && allNumeric(row[11:15]):
							results["pending_normal_aio_reads"] = (getIntFromString(row[4]) +
								getIntFromString(row[5]) +
								getIntFromString(row[6]) +
								getIntFromString(row[7]))
							results["pending_normal_aio_writes"] = (getIntFromString(row[11]) +
								getIntFromString(row[12]) +
								getIntFromString(row[13]) +
								getIntFromString(row[14]))

						case allNumeric(row[4:9]) && allNumeric(row[12:15]): // (len(row) == 16) Pending normal aio reads: 0 [0, 0, 0, 0] , aio writes: 0 [0, 0] ,
							results["pending_normal_aio_reads"] = getIntFromString(row[4])
							results["pending_normal_aio_writes"] = getIntFromString(row[12])
						default:
							l.Warnf("can't parse line %s", line)
						}
					case len(row) == 18:
						// (len(row) == 18) Pending normal aio reads: 0 [0, 0, 0, 0] , aio writes: 0 [0, 0, 0, 0] ,
						results["pending_normal_aio_reads"] = getIntFromString(row[4])
						results["pending_normal_aio_writes"] = getIntFromString(row[12])
					case len(row) == 22:
						// (len(row) == 22)
						// Pending normal aio reads: 0 [0, 0, 0, 0, 0, 0, 0, 0] , aio writes: 0 [0, 0, 0, 0] ,
						results["pending_normal_aio_reads"] = getIntFromString(row[4])
						results["pending_normal_aio_writes"] = getIntFromString(row[16])
					}

				case strings.HasPrefix(line, "ibuf aio reads"):
					// ibuf aio reads: 0, log i/o's: 0, sync i/o's: 0
					// or ibuf aio reads:, log i/o's:, sync i/o's:
					if len(row) == 10 {
						results["pending_ibuf_aio_reads"] = getIntFromString(row[3])
						results["pending_aio_log_ios"] = getIntFromString(row[6])
						results["pending_aio_sync_ios"] = getIntFromString(row[9])
					} else if len(row) == 7 {
						results["pending_ibuf_aio_reads"] = 0
						results["pending_aio_log_ios"] = 0
						results["pending_aio_sync_ios"] = 0
					}
				case strings.HasPrefix(line, "Pending flushes (fsync)"):
					if len(row) == 4 {
						// Pending flushes (fsync): 0
						results["pending_buffer_pool_flushes"] = getIntFromString(row[3])
					} else {
						// Pending flushes (fsync) log: 0; buffer pool: 0
						results["pending_log_flushes"] = getIntFromString(row[4])
						results["pending_buffer_pool_flushes"] = getIntFromString(row[7])
					}
				case strings.HasPrefix(line, "Ibuf for space 0: size "):
					//  Older InnoDB code seemed to be ready for an ibuf per tablespace.  It
					//  had two lines in the output.  Newer has just one line, see below.
					//  Ibuf for space 0: size 1, free list len 887, seg size 889, is not empty
					//  Ibuf for space 0: size 1, free list len 887, seg size 889,
					results["ibuf_size"] = getIntFromString(row[5])
					results["ibuf_free_list"] = getIntFromString(row[9])
					results["ibuf_segment_size"] = getIntFromString(row[12])
				case strings.HasPrefix(line, "Ibuf: size "):
					// Ibuf: size 1, free list len 4634, seg size 4636,
					results["ibuf_size"] = getIntFromString(row[2])
					results["ibuf_free_list"] = getIntFromString(row[6])
					results["ibuf_segment_size"] = getIntFromString(row[9])
					if strings.Contains(line, "merges") {
						results["ibuf_merges"] = getIntFromString(row[10])
					}
				case strings.Index(line, ", delete mark ") > 0 && strings.HasPrefix(prevLine, "merged operations:"):
					// Output of show engine innodb status has changed in 5.5
					// merged operations:
					// insert 593983, delete mark 387006, delete 73092
					results["ibuf_merged_inserts"] = getIntFromString(row[1])
					results["ibuf_merged_delete_marks"] = getIntFromString(row[4])
					results["ibuf_merged_deletes"] = getIntFromString(row[6])
					results["ibuf_merged"] = (results["ibuf_merged_inserts"] + results["ibuf_merged_delete_marks"] + results["ibuf_merged_deletes"])
				case strings.Index(line, " merged recs, ") > 0:
					// 19817685 inserts, 19817684 merged recs, 3552620 merges
					results["ibuf_merged_inserts"] = getIntFromString(row[0])
					results["ibuf_merged"] = getIntFromString(row[2])
					results["ibuf_merges"] = getIntFromString(row[5])
				case strings.HasPrefix(line, "Hash table size "):
					// In some versions of InnoDB, the used cells is omitted.
					// Hash table size 4425293, used cells 4229064, ....
					// Hash table size 57374437, node heap has 72964 buffer(s) <--
					// no used cells
					results["hash_index_cells_total"] = getIntFromString(row[3])
					if strings.Index(line, "used cells") > 0 {
						results["hash_index_cells_used"] = getIntFromString(row[6])
					} else {
						results["hash_index_cells_used"] = 0
					}

				case strings.Index(line, " log i/o's done, ") > 0:
					// 3430041 log i/o's done, 17.44 log i/o's/second
					// 520835887 log i/o's done, 17.28 log i/o's/second, 518724686
					// syncs, 2980893 checkpoints
					results["log_writes"] = getIntFromString(row[0])

				case strings.Index(line, " pending log writes, ") > 0:
					// 0 pending log writes, 0 pending chkp writes
					results["pending_log_writes"] = getIntFromString(row[0])
					results["pending_checkpoint_writes"] = getIntFromString(row[4])
				case strings.HasPrefix(line, "Log sequence number"):
					// This number is NOT printed in hex in InnoDB plugin.
					// Log sequence number 272588624
					results["lsn_current"] = getIntFromString(row[3])
				case strings.HasPrefix(line, "Log flushed up to"):
					// This number is NOT printed in hex in InnoDB plugin.
					// Log flushed up to   272588624
					results["lsn_flushed"] = getIntFromString(row[4])
				case strings.HasPrefix(line, "Last checkpoint at"):
					// Last checkpoint at  272588624
					results["lsn_last_checkpoint"] = getIntFromString(row[3])

					// BUFFER POOL AND MEMORY
				case strings.HasPrefix(line, "Total memory allocated") && strings.Index(line, "in additional pool allocated") > 0:
					// Total memory allocated 29642194944; in additional pool allocated 0
					// Total memory allocated by read views 96
					results["mem_total"] = getIntFromString(row[3])
					results["mem_additional_pool"] = getIntFromString(row[8])
				case strings.HasPrefix(line, "Adaptive hash index "):
					// Adaptive hash index 1538240664     (186998824 + 1351241840)
					results["mem_adaptive_hash"] = getIntFromString(row[3])
				case strings.HasPrefix(line, "Page hash           "):
					//   Page hash           11688584
					results["mem_page_hash"] = getIntFromString(row[2])
				case strings.HasPrefix(line, "Dictionary cache    "):
					//   Dictionary cache    145525560      (140250984 + 5274576)
					results["mem_dictionary"] = getIntFromString(row[2])
				case strings.HasPrefix(line, "File system         "):
					//   File system         313848         (82672 + 231176)
					results["mem_file_system"] = getIntFromString(row[2])
				case strings.HasPrefix(line, "Lock system         "):
					//   Lock system         29232616       (29219368 + 13248)
					results["mem_lock_system"] = getIntFromString(row[2])
				case strings.HasPrefix(line, "Recovery system     "):
					//   Recovery system     0      (0 + 0)
					results["mem_recovery_system"] = getIntFromString(row[2])
				case strings.HasPrefix(line, "Threads             "):
					//   Threads             409336         (406936 + 2400)
					results["mem_thread_hash"] = getIntFromString(row[1])
				case strings.HasPrefix(line, "Buffer pool size "):
					// The " " after size is necessary to avoid matching the wrong line:
					// Buffer pool size        1769471
					// Buffer pool size, bytes 28991012864
					if bufferID == -1 {
						results["buffer_pool_pages_total"] = getIntFromString(row[3])
					}
				case strings.HasPrefix(line, "Free buffers"):
					// Free buffers            0
					if bufferID == -1 {
						results["buffer_pool_pages_free"] = getIntFromString(row[2])
					}
				case strings.HasPrefix(line, "Database pages"):
					// Database pages          1696503
					if bufferID == -1 {
						results["buffer_pool_pages_data"] = getIntFromString(row[2])
					}

				case strings.HasPrefix(line, "Modified db pages"):
					// Modified db pages       160602
					if bufferID == -1 {
						results["buffer_pool_pages_dirty"] = getIntFromString(row[3])
					}
				case strings.HasPrefix(line, "Pages read ahead"):
					// Must do this BEFORE the next test, otherwise it'll get fooled by this
					// line from the new plugin:
					// Pages read ahead 0.00/s, evicted without access 0.06/s
					continue
				case strings.HasPrefix(line, "Pages read"):
					// Pages read 15240822, created 1770238, written 21705836
					if bufferID == -1 {
						results["pages_read"] = getIntFromString(row[2])
						results["pages_created"] = getIntFromString(row[4])
						results["pages_written"] = getIntFromString(row[6])
					}
					// ROW OPERATIONS
				case strings.HasPrefix(line, "Number of rows inserted"):
					// Number of rows inserted 50678311, updated 66425915, deleted
					// 20605903, read 454561562
					results["rows_inserted"] = getIntFromString(row[4])
					results["rows_updated"] = getIntFromString(row[6])
					results["rows_deleted"] = getIntFromString(row[8])
					results["rows_read"] = getIntFromString(row[10])
				case strings.Index(line, " queries inside InnoDB, ") > 0:
					// 0 queries inside InnoDB, 0 queries in queue
					results["queries_inside"] = getIntFromString(row[0])
					results["queries_queued"] = getIntFromString(row[4])
				}
				prevLine = line
			}

			if lsnCurrent, ok := results["lsn_current"]; ok {
				if lsnLastCheckpoint, ok := results["lsn_last_checkpoint"]; ok {
					results["checkpoint_age"] = lsnCurrent - lsnLastCheckpoint
				}
			}

			if _, ok := results["checkpoint_age"]; !ok {
				l.Warn("unable to compute checkpoint_age, no InnoDB LSN metrics available")
			}

			statusRes = results
		}
	}

	return statusRes, err
}

type rowResponse struct {
	columnTypes []*sql.ColumnType
	rows        [][]*interface{}
}

func (ipt *Input) getQueryRows(sqlText string) (res rowResponse, err error) {
	rows, err := ipt.db.Query(sqlText)
	if err != nil {
		return
	}

	if err = rows.Err(); err != nil {
		closeRows(rows)
		return
	}

	columns, err := rows.Columns()
	if err != nil {
		return
	}

	columnTypes, err := rows.ColumnTypes()
	if err != nil {
		return
	}

	res.columnTypes = columnTypes

	for rows.Next() {
		rowResult := make([]interface{}, len(columns))
		for index := range columns {
			rowResult[index] = new(interface{})
		}

		if err = rows.Scan(rowResult...); err != nil {
			return
		}

		row := make([]*interface{}, len(columns))
		for i, v := range rowResult {
			vv, _ := v.(*interface{})
			row[i] = vv
		}
		res.rows = append(res.rows, row)
	}

	return res, err
}

func (ipt *Input) collectMysqlTableSchema() error {
	ipt.mTableSchema = []map[string]interface{}{}

	tableSchemaSQL := `
	SELECT
        TABLE_SCHEMA,
        TABLE_NAME,
        TABLE_TYPE,
        ifnull(ENGINE, 'NONE') as ENGINE,
        ifnull(VERSION, '0') as VERSION,
        ifnull(ROW_FORMAT, 'NONE') as ROW_FORMAT,
        ifnull(TABLE_ROWS, '0') as TABLE_ROWS,
        ifnull(DATA_LENGTH, '0') as DATA_LENGTH,
        ifnull(INDEX_LENGTH, '0') as INDEX_LENGTH,
        ifnull(DATA_FREE, '0') as DATA_FREE
    FROM information_schema.tables
    WHERE TABLE_SCHEMA NOT IN ('mysql', 'performance_schema', 'information_schema', 'sys')
	`

	if len(ipt.Tables) > 0 {
		var arr []string
		for _, table := range ipt.Tables {
			arr = append(arr, fmt.Sprintf("'%s'", table))
		}

		filterStr := strings.Join(arr, ",")
		tableSchemaSQL = fmt.Sprintf("%s and TABLE_NAME in (%s);", tableSchemaSQL, filterStr)
	}

	if res := getCleanTableSchema(ipt.q(tableSchemaSQL)); res != nil {
		ipt.mTableSchema = res
	} else {
		l.Warnf("collect_table_schema_failed")
		if len(ipt.mTableSchema) > 0 {
			ipt.mTableSchema = []map[string]interface{}{}
		}
	}

	return nil
}

func (ipt *Input) collectMysqlUserStatus() error {
	ipt.mUserStatusName = map[string]interface{}{}
	ipt.mUserStatusVariable = map[string]map[string]interface{}{}
	ipt.mUserStatusConnection = map[string]map[string]interface{}{}

	userSQL := `select DISTINCT(user) from mysql.user`

	if len(ipt.Users) > 0 {
		var arr []string
		for _, user := range ipt.Users {
			arr = append(arr, fmt.Sprintf("'%s'", user))
		}

		filterStr := strings.Join(arr, ",")
		userSQL = fmt.Sprintf("%s where user in (%s);", userSQL, filterStr)
	}

	if res := getCleanUserStatusName(ipt.q(userSQL)); res != nil {
		ipt.mUserStatusName = res
	} else {
		l.Warn("collect_user_name_failed")
	}

	userQuerySQL := `
	select VARIABLE_NAME, VARIABLE_VALUE
	from performance_schema.status_by_user
	where user='%s';
	`

	userConnSQL := `select USER, CURRENT_CONNECTIONS, TOTAL_CONNECTIONS
	from performance_schema.users
	where user = '%s';
    `

	for user := range ipt.mUserStatusName {
		if res := getCleanUserStatusVariable(ipt.q(fmt.Sprintf(userQuerySQL, user))); res != nil {
			ipt.mUserStatusVariable = make(map[string]map[string]interface{})
			ipt.mUserStatusVariable[user] = res
		}

		if res := getCleanUserStatusConnection(ipt.q(fmt.Sprintf(userConnSQL, user))); res != nil {
			ipt.mUserStatusConnection = make(map[string]map[string]interface{})
			ipt.mUserStatusConnection[user] = res
		}
	}

	if len(ipt.mUserStatusVariable) == 0 {
		l.Warnf("collect_user_variable_failed")
	}

	if len(ipt.mUserStatusConnection) == 0 {
		l.Warnf("collect_user_connection_failed")
	}

	return nil
}

func (ipt *Input) collectMysqlCustomQueries() error {
	ipt.mCustomQueries = map[string][]map[string]interface{}{}

	for _, item := range ipt.Query {
		arr := getCleanMysqlCustomQueries(ipt.q(item.SQL))
		if arr == nil {
			continue
		}
		if item.md5Hash == "" {
			hs := hashcode.GetMD5String32([]byte(item.SQL))
			item.md5Hash = hs
		}
		ipt.mCustomQueries[item.md5Hash] = make([]map[string]interface{}, 0)
		ipt.mCustomQueries[item.md5Hash] = arr
	}

	return nil
}

func (ipt *Input) collectMysqlDbmMetric() error {
	var err error
	defer func() {
		if err != nil {
			ipt.dbmMetricRows = []dbmRow{}

			// dbmCache cannot reset
		}
	}()

	statementSummarySQL := `
		SELECT schema_name,digest,digest_text,count_star,
		sum_timer_wait,sum_lock_time,sum_errors,sum_rows_affected,
		sum_rows_sent,sum_rows_examined,sum_select_scan,sum_select_full_join,
		sum_no_index_used,sum_no_good_index_used
		FROM performance_schema.events_statements_summary_by_digest
		WHERE digest_text NOT LIKE 'EXPLAIN %' OR digest_text IS NULL
		ORDER BY count_star DESC LIMIT 10000`

	dbmRows := getCleanSummaryRows(ipt.q(statementSummarySQL))
	if dbmRows == nil {
		err = fmt.Errorf("collect_summary_rows_failed")
		return err
	}

	metricRows, newDbmCache := getMetricRows(dbmRows, &ipt.dbmCache)
	ipt.dbmMetricRows = metricRows

	// save dbm rows
	ipt.dbmCache = newDbmCache

	return err
}

// mysql_dbm_sample
//----------------------------------------------------------------------

func (ipt *Input) collectMysqlDbmSample() error {
	var err error
	defer func() {
		if err != nil {
			ipt.dbmSamplePlans = []planObj{}
		}
	}()

	if len(ipt.dbmSampleCache.globalStatusTable) == 0 {
		if len(ipt.dbmSampleCache.version.version) == 0 {
			const sqlSelect = "SELECT VERSION();"
			version := getCleanMysqlVersion(ipt.q(sqlSelect))
			if version == nil {
				err = errors.New("version_nil")
				return err
			}
			ipt.dbmSampleCache.version = *version
		}

		if ipt.dbmSampleCache.version.flavor == strMariaDB || !(ipt.dbmSampleCache.version.versionCompatible([]int{5, 7, 0})) {
			ipt.dbmSampleCache.globalStatusTable = "information_schema.global_status"
		} else {
			ipt.dbmSampleCache.globalStatusTable = "performance_schema.global_status"
		}
	}

	strategy, err := getSampleCollectionStrategy(ipt)
	if err != nil {
		return err
	}

	rows, err := getNewEventsStatements(ipt, strategy.table, 5000)
	if err != nil {
		return err
	}

	rows = filterValidStatementRows(ipt, rows)

	plans := collectPlanForStatements(ipt, rows)
	if len(plans) > 0 {
		ipt.dbmSamplePlans = plans
	}

	return nil
}

//----------------------------------------------------------------------
