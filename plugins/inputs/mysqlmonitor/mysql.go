package mysqlmonitor

import (
	"database/sql"
	"time"
	"fmt"
	"strings"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"github.com/go-sql-driver/mysql"
	influxm "github.com/influxdata/influxdb1-client/models"

	"github.com/spf13/cast"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
)

const (
	defaultTimeout                             = 5 * time.Second
	defaultPerfEventsStatementsDigestTextLimit = 120
	defaultPerfEventsStatementsLimit           = 250
	defaultPerfEventsStatementsTimeLimit       = 86400
	defaultGatherGlobalVars                    = true
)

var (
	l             *logger.Logger
	name          = "mysqlMonitor"
)

func (_ *MysqlMonitor) Catalog() string {
	return "db"
}

func (_ *MysqlMonitor) SampleConfig() string {
	return configSample
}

func (m *MysqlMonitor) Run() {
	initLog()

	l.Info("mysqlMonitor input started...")

	m.initCfg()

	tick := time.NewTicker(m.IntervalDuration)
	defer tick.Stop()

	for {
		select {
		case <-tick.C:
			m.collectMetrics()
		case <-datakit.Exit.Wait():
			l.Info("exit")
			return
		}
	}
}

func initLog() {
	l = logger.SLogger("mysqlMonitor")
}

func (m *MysqlMonitor) initCfg() {
	// 采集频度
	m.IntervalDuration = 10 * time.Minute

	if m.Interval != "" {
		du, err := time.ParseDuration(m.Interval)
		if err != nil {
			l.Errorf("bad interval %s: %s, use default: 10m", m.Interval, err.Error())
		} else {
			m.IntervalDuration = du
		}
	}

	// 指标集名称
	if m.MetricName == "" {
		m.MetricName = name
	}

	if m.Timeout != "" {
		du, err := time.ParseDuration(m.Timeout)
		if err != nil {
			l.Errorf("config timeout value (%v)  error %v ", m.Timeout, err)
		} else {
			m.TimeoutDuration = du
		}
	}

	// build dsn string
	dsnStr := m.getDsnString()
	l.Infof("db build dsn connect str %s", dsnStr)
	db, err := sql.Open("mysql", dsnStr)
	if err != nil {
		l.Errorf("sql.Open(): %s", err.Error())
	} else {
		m.db = db
	}
}

func (m *MysqlMonitor) getDsnString() string {
	cfg := mysql.Config{
		AllowNativePasswords: true,
		CheckConnLiveness:    true,
	    User:                 m.User,
	    Passwd:               m.Pass,
	}

	// set addr
	if m.Sock != "" {
		cfg.Net = "unix"
		cfg.Addr = m.Sock
	} else {
		addr := fmt.Sprintf("%s:%d", m.Host, m.Port)
		cfg.Net = "tcp"
		cfg.Addr = addr
	}

	// set timeout
	if m.TimeoutDuration != 0 {
		cfg.Timeout = m.TimeoutDuration
	}

	// set Charset
	if m.Charset != "" {
		cfg.Params["charset"] = m.Charset
	}

	// ssl
	if m.Tls != nil {

	}

	// tls (todo)
	return cfg.FormatDSN()
}

func (m *MysqlMonitor) collectMetrics() error {
	defer m.db.Close()
	// ping
	if err := m.db.Ping(); err != nil {
		l.Errorf("db connect error %v", err)
		return err
	}

	m.resData = make(map[string]interface{})

	//STATUS data
	m.getStatus()

	// VARIABLES data
	m.getVariables()

	// innodb
	if m.options.DisableInnodbMetrics && m.innodbEngineEnabled()  {
		m.getInnodbStatus()
	}

	// Binary log statistics
    if _, ok := m.resData["log_bin"]; ok {
    	metric["INNODB_VARS"].disable = true
    	m.getLogStats()
    }

    // Compute key cache utilization metric
    m.computeCacheUtilization()

    if m.options.ExtraStatusMetrics {
    	// 额外的status metric 设置标志
    	metric["OPTIONAL_STATUS_VARS"].disable = true
    	if m.versionCompatible("5.6.6") {
    		metric["OPTIONAL_STATUS_VARS_5_6_6"].disable = true
    	}
    }

    if m.options.GaleraCluster {
    	metric["GALERA_VARS"].disable = true
    }

    if m.options.ExtraPerformanceMetrics && m.versionCompatible("5.6.0") {
    	// if _, ok := m.resData["performance_schema"] {
    	// 	metric["PERFORMANCE_VARS"].disable = true
    	// 	m.getQueryExecTime95thus()
     //        m.queryExecTimePerSchema()
    	// }
    }

    if m.options.SchemaSizeMetrics {
    	metric["SCHEMA_VARS"].disable = true
    	m.querySizePerschema()
    }

    // replication
    if m.options.Replication {
    	metric["SCHEMA_VARS"].disable = true
    	m.collectReplication()
    }

    m.submitMetrics()

	return nil
}

func (m *MysqlMonitor) submitMetrics() error {
    var (
    	tags   = make(map[string]string)
    	fields = make(map[string]interface{})
    )

    if m.Service != "" {
    	tags["service"] = m.Service
    }

    for tag, tagV := range m.Tags {
		tags[tag] = tagV
	}

	m.dupedMetrics()

    for _, kind := range metric {
   		if !kind.disable {
   			for k, item := range kind.metric {
   				if m.resData[k] != "" && !item.disable {
   					if item.parse != nil {
   						// error check (todo)
   						value, ok  := m.resData[k]
   						if !ok {
   							continue
   						}
   						var val interface{}
   						switch v := value.(type)  {
						case int64:
						    val = item.parse(v)
						    fields[item.name] = val
						case string:
						    val = item.parse(v)
						    fields[item.name] = val
						case map[string]float64:
							for kk, vv := range v {
								itemKey := fmt.Sprintf(item.name, kk)
								fields[itemKey] = vv
							}
						case map[string]int64:
							for kk, vv := range v {
								itemKey := fmt.Sprintf(item.name, kk)
								fields[itemKey] = vv
							}
						default:
						}
   					}
   				}
   			}
   		}
   	}

   	pt, err := io.MakeMetric(m.MetricName, tags, fields, time.Now())
	if err != nil {
		l.Errorf("make metric point error %v", err)
	}

	fmt.Println("pt=======>", string(pt))

	_, err = influxm.ParsePointsWithPrecision(pt, time.Now().UTC(), "")
	if err != nil {
		l.Errorf("[error] : %s", err.Error())
		return err
	}

	err = io.NamedFeed([]byte(pt), io.Metric, name)
	if err != nil {
		l.Errorf("push metric point error %v", err)
	}

	return nil
}

func (m *MysqlMonitor) getQueryExecTime95thus() error {

	return nil
}

func (m *MysqlMonitor) queryExecTimePerSchema() error {
	return nil
}

func (m *MysqlMonitor) versionCompatible(version string) bool {
	return false
}

// status data
func (m *MysqlMonitor) getStatus() error {
	if err := m.db.Ping(); err != nil {
		l.Errorf("db connect error %v", err)
		return err
	}

	globalStatusSql := "SHOW /*!50002 GLOBAL */ STATUS;"
	rows, err := m.db.Query(globalStatusSql)
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

func (m *MysqlMonitor) dupedMetrics() {
	dic := map[string]string{
		"Table_locks_waited": "Table_locks_waited_rate",
        "Table_locks_immediate": "Table_locks_immediate_rate",
	}

	for src, dst := range dic {
		if _, ok := m.resData[src]; ok {
			m.resData[dst] = m.resData[src]
		}
	}
}

// variables data
func (m *MysqlMonitor) getVariables() error {
	variablesSql := "SHOW GLOBAL VARIABLES;"
	rows, err := m.db.Query(variablesSql)
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

// innodb_engine_enabled
func (m *MysqlMonitor) innodbEngineEnabled() bool {
	innodbEnabledSql := `
	SELECT engine
	FROM information_schema.ENGINES
	WHERE engine='InnoDB' and support != 'no' and support != 'disabled';
	`
	result, err := m.db.Exec(innodbEnabledSql)
	if err != nil {
		l.Errorf("")
		return false
	}

	count, err := result.RowsAffected()
	if err != nil {
		l.Errorf("")
		return false
	}

	if count > 0 {
		return true
	}

	return false
}

// innodb status (todo)
func (m *MysqlMonitor) getInnodbStatus() error {
	innodbStatusSql := "SHOW /*!50000 ENGINE*/ INNODB STATUS;"
	var key, name, text string
	err := m.db.QueryRow(innodbStatusSql).Scan(&key, &name, &text)
	if err != nil {
		return err
	}

	if len(text) > 0 {
		return m.parseInnodbStatus(text)
	}

	return nil
}

func (m *MysqlMonitor) parseInnodbStatus(str string) error {
	// isTransaction := false
	prevLine := ""

	for _, line := range strings.Split(str, "\n") {
		record := strings.Fields(line)
		fmt.Println("line =====>", line)
		fmt.Println("record =====>", record)
		// Innodb Semaphores
		if strings.Index(line, "Mutex spin waits") == 0 {
			m.resData["Innodb_mutex_spin_waits"] = record[3]
			m.resData["Innodb_mutex_spin_rounds"] = record[5]
			m.resData["Innodb_mutex_os_waits"] = record[8]
			continue
		}

		if strings.Index(line, "RW-shared spins") == 0 && strings.Index(line, ";") > 0 {
			m.resData["Innodb_mutex_os_waits"] = record[2]
			m.resData["Innodb_s_lock_os_waits"] = record[5]
			m.resData["Innodb_x_lock_spin_waits"] = record[8]
			m.resData["Innodb_x_lock_os_waits"] = record[11]
			continue
		}

		if strings.Index(line, "RW-shared spins") == 0 && strings.Index(line, "; RW-excl spins") < 0 {
			m.resData["Innodb_s_lock_spin_waits"] = record[2]
			m.resData["Innodb_s_lock_spin_rounds"] = record[4]
			m.resData["Innodb_s_lock_os_waits"] = record[7]
			continue
		}

		if strings.Index(line, "RW-excl spins") == 0 {
			m.resData["Innodb_s_lock_spin_waits"] = record[2]
			m.resData["Innodb_x_lock_spin_rounds"] = record[4]
			m.resData["Innodb_x_lock_os_waits"] = record[7]
			continue
		}

		// if strings.Index(line, "seconds the semaphore:") > 0 {
		// 	increaseMap(p, "Innodb_semaphore_waits", "1")
		// 	wait := atof(record[9])
		// 	wait = wait * 1000
		// 	increaseMap(p, "Innodb_mutex_spin_waits", fmt.Sprintf("%.f", wait))
		// 	continue
		// }

		// // Innodb Transactions
		// if strings.Index(line, "Trx id counter") == 0 {
		// 	// The beginning of the TRANSACTIONS section: start counting
		// 	// transactions
		// 	// Trx id counter 0 1170664159
		// 	// Trx id counter 861B144C
		// 	isTransaction = true
		// 	continue
		// }
		// if strings.Index(line, "History list length") == 0 {
		// 	// History list length 132
		// 	increaseMap(p, "Innodb_history_list_length", record[3])
		// 	continue
		// }
		// if isTransaction && strings.Index(line, "---TRANSACTION") == 0 {
		// 	// ---TRANSACTION 0, not started, process no 13510, OS thread id 1170446656
		// 	increaseMap(p, "Innodb_current_transactions", "1")
		// 	if strings.Index(line, "ACTIVE") > 0 {
		// 		increaseMap(p, "Innodb_active_transactions", "1")
		// 	}
		// 	continue
		// }
		// if isTransaction && strings.Index(line, "------- TRX HAS BEEN") == 0 {
		// 	// ------- TRX HAS BEEN WAITING 32 SEC FOR THIS LOCK TO BE GRANTED:
		// 	increaseMap(p, "Innodb_row_lock_time", "1")
		// 	continue
		// }
		// if strings.Index(line, "read views open inside InnoDB") > 0 {
		// 	// 1 read views open inside InnoDB
		// 	m.resData["Innodb_read_views"] = atof(record[0])
		// 	continue
		// }
		// if strings.Index(line, "mysql tables in use") == 0 {
		// 	// mysql tables in use 2, locked 2
		// 	increaseMap(p, "Innodb_tables_in_use", record[4])
		// 	increaseMap(p, "Innodb_locked_tables", record[6])
		// 	continue
		// }
		// if isTransaction && strings.Index(line, "lock struct(s)") == 0 {
		// 	// 23 lock struct(s), heap size 3024, undo log entries 27
		// 	// LOCK WAIT 12 lock struct(s), heap size 3024, undo log entries 5
		// 	// LOCK WAIT 2 lock struct(s), heap size 368
		// 	if strings.Index(line, "LOCK WAIT") > 0 {
		// 		increaseMap(p, "Innodb_lock_structs", record[2])
		// 		increaseMap(p, "Innodb_locked_transactions", "1")
		// 	} else if strings.Index(line, "ROLLING BACK") > 0 {
		// 		// ROLLING BACK 127539 lock struct(s), heap size 15201832,
		// 		// 4411492 row lock(s), undo log entries 1042488
		// 		increaseMap(p, "Innodb_lock_structs", record[0])
		// 	} else {
		// 		increaseMap(p, "Innodb_lock_structs", record[0])
		// 	}
		// 	continue
		// }

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
		// if strings.Index(line, "Adaptive hash index ") == 0 {
		// 	// Adaptive hash index 1538240664     (186998824 + 1351241840)
		// 	v := atof(record[3])
		// 	setIfEmpty(p, "Innodb_mem_adaptive_hash", v)
		// 	continue
		// }
		// if strings.Index(line, "Page hash           ") == 0 {
		// 	// Page hash           11688584
		// 	v := atof(record[2])
		// 	setIfEmpty(p, "Innodb_mem_page_hash", v)
		// 	continue
		// }
		// if strings.Index(line, "Dictionary cache    ") == 0 {
		// 	// Dictionary cache    145525560      (140250984 + 5274576)
		// 	v := atof(record[2])
		// 	setIfEmpty(p, "Innodb_mem_dictionary", v)
		// 	continue
		// }
		// if strings.Index(line, "File system         ") == 0 {
		// 	// File system         313848         (82672 + 231176)
		// 	v := atof(record[2])
		// 	setIfEmpty(p, "Innodb_mem_file_system", v)
		// 	continue
		// }
		// if strings.Index(line, "Lock system         ") == 0 {
		// 	// Lock system         29232616       (29219368 + 13248)
		// 	v := atof(record[2])
		// 	setIfEmpty(p, "Innodb_mem_lock_system", v)
		// 	continue
		// }
		// if strings.Index(line, "Recovery system     ") == 0 {
		// 	// Recovery system     0      (0 + 0)
		// 	v := atof(record[2])
		// 	setIfEmpty(p, "Innodb_mem_recovery_system", v)
		// 	continue
		// }
		// if strings.Index(line, "Threads             ") == 0 {
		// 	// Threads             409336         (406936 + 2400)
		// 	v := atof(record[1])
		// 	setIfEmpty(p, "Innodb_mem_thread_hash", v)
		// 	continue
		// }
		// if strings.Index(line, "Buffer pool size ") == 0 {
		// 	// The " " after size is necessary to avoid matching the wrong line:
		// 	// Buffer pool size        1769471
		// 	// Buffer pool size, bytes 28991012864
		// 	v := atof(record[3])
		// 	setIfEmpty(p, "Innodb_buffer_pool_pages_total", v)
		// 	continue
		// }
		// if strings.Index(line, "Free buffers") == 0 {
		// 	// Free buffers            0
		// 	v := atof(record[2])
		// 	setIfEmpty(p, "Innodb_buffer_pool_pages_free", v)
		// 	continue
		// }
		// if strings.Index(line, "Database pages") == 0 {
		// 	// Database pages          1696503
		// 	v := atof(record[2])
		// 	setIfEmpty(p, "Innodb_buffer_pool_pages_data", v)
		// 	continue
		// }
		// if strings.Index(line, "Modified db pages") == 0 {
		// 	// Modified db pages       160602
		// 	v := atof(record[3])
		// 	setIfEmpty(p, "Innodb_buffer_pool_pages_dirty", v)
		// 	continue
		// }
		// if strings.Index(line, "Pages read ahead") == 0 {
		// 	v := atof(record[3])
		// 	setIfEmpty(p, "Innodb_buffer_pool_read_ahead", v)
		// 	v = atof(record[7])
		// 	setIfEmpty(p, "Innodb_buffer_pool_read_ahead_evicted", v)
		// 	v = atof(record[11])
		// 	setIfEmpty(p, "Innodb_buffer_pool_read_ahead_rnd", v)
		// 	continue
		// }
		// if strings.Index(line, "Pages read") == 0 {
		// 	// Pages read 15240822, created 1770238, written 21705836
		// 	v := atof(record[2])
		// 	setIfEmpty(p, "Innodb_pages_read", v)
		// 	v = atof(record[4])
		// 	setIfEmpty(p, "Innodb_pages_created", v)
		// 	v = atof(record[6])
		// 	setIfEmpty(p, "Innodb_pages_written", v)
		// 	continue
		// }

		// Row Operations
		if strings.Index(line, "Number of rows inserted") == 0 {
			m.resData["Innodb_rows_inserted"] = atof(record[4])
			m.resData["Innodb_rows_updated"] = atof(record[6])
			m.resData["Innodb_rows_deleted"] = atof(record[8])
			m.resData["Innodb_rows_read"] = atof(record[10])
			continue
		}
		if strings.Index(line, " queries inside InnoDB, ") > 0 {
			m.resData["Innodb_queries_inside"] = atof(record[0])
			m.resData["Innodb_queries_queued"] = atof(record[4])
			continue
		}

		// for next loop
		prevLine = line
	}

	// We need to calculate this metric separately
	// m.resData["Innodb_checkpoint_age"] = m.resData["Innodb_lsn_current"] - m.resData["Innodb_lsn_last_checkpoint"]

	return nil
}

// log stats
func (m *MysqlMonitor) getLogStats() error {
	logSql := "SHOW BINARY LOGS;"
	rows, err := m.db.Query(logSql)
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

// Compute key cache utilization metric (todo)
func (m *MysqlMonitor) computeCacheUtilization() error {
	return nil
}

func (m *MysqlMonitor) getQueryExecTime95th() error {
	queryExecTimeSql := `
	SELECT avg_us, ro as percentile FROM
	(SELECT avg_us, @rownum := @rownum + 1 as ro FROM
    (SELECT ROUND(avg_timer_wait / 1000000) as avg_us
        FROM performance_schema.events_statements_summary_by_digest
        ORDER BY avg_us ASC) p,
    (SELECT @rownum := 0) r) q
	WHERE q.ro > ROUND(.95*@rownum)
	ORDER BY percentile ASC
	LIMIT 1
	`
	rows, err := m.db.Query(queryExecTimeSql)
	if err != nil {
		return err
	}
	defer rows.Close()

	var avg int64
	for rows.Next() {
		var row1 *sql.RawBytes = new(sql.RawBytes)
		var row2 *sql.RawBytes = new(sql.RawBytes)

		if err = rows.Scan(row1, row2); err != nil {
			continue
		}

		avg = cast.ToInt64(string(*row1))
	}

	m.resData["perf_digest_95th_percentile_avg_us"] = avg
	return nil
}

func (m *MysqlMonitor) getQueryExecTimePerSchema() error {
	queryExecPerTimeSql := `
	SELECT schema_name, ROUND((SUM(sum_timer_wait) / SUM(count_star)) / 1000000) AS avg_us
	FROM performance_schema.events_statements_summary_by_digest
	WHERE schema_name IS NOT NULL
	GROUP BY schema_name;
	`
	rows, err := m.db.Query(queryExecPerTimeSql)
	if err != nil {
		return err
	}
	defer rows.Close()

	var schemaSize = map[string]int64{}
	for rows.Next() {
		var key string
		var val *sql.RawBytes = new(sql.RawBytes)

		if err = rows.Scan(&key, val); err != nil {
			continue
		}

		size := cast.ToInt64(string(*val))

		schemaSize[key] = size
	}

	m.resData["query_run_time_avg"] = schemaSize

	return nil
}

func (m *MysqlMonitor) querySizePerschema() error {
	querySizePerschemaSql := `
	SELECT   table_schema, IFNULL(SUM(data_length+index_length)/1024/1024,0) AS total_mb
	FROM     information_schema.tables
	GROUP BY table_schema;
	`
	rows, err := m.db.Query(querySizePerschemaSql)
	if err != nil {
		return err
	}
	defer rows.Close()

	var schemaSize = map[string]float64{}
	for rows.Next() {
		var key string
		var val *sql.RawBytes = new(sql.RawBytes)

		if err = rows.Scan(&key, val); err != nil {
			continue
		}

		size := cast.ToFloat64(string(*val))

		schemaSize[key] = size
	}

	m.resData["information_schema_size"] = schemaSize

	return nil
}

// replication (todo)
func (m *MysqlMonitor) collectReplication() error {
	return nil
}

// "synthetic" metrics
func (m *MysqlMonitor) computeSynthetic() error {
	return nil
}

func (m *MysqlMonitor) Test() (*inputs.TestResult, error) {
	m.test = true
	m.testData = nil

	m.collectMetrics()

	res := &inputs.TestResult{
		Result: m.testData,
		Desc:   "success!",
	}

	return res, nil
}

func init() {
	inputs.Add(name, func() inputs.Input {
		return &MysqlMonitor{}
	})
}
