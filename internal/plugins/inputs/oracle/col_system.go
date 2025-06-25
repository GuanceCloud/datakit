// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package oracle

import (
	"database/sql"
	"reflect"
	"strconv"
	"time"

	"github.com/GuanceCloud/cliutils/point"

	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
)

var systemCols = map[string]string{
	"Average Active Sessions":              "active_sessions",
	"Background Time Per Sec":              "active_background",
	"Branch Node Splits Per Sec":           "branch_node_splits",
	"Consistent Read Changes Per Sec":      "consistent_read_changes",
	"Consistent Read Gets Per Sec":         "consistent_read_gets",
	"Current OS Load":                      "os_load",
	"Database Wait Time Ratio":             "database_wait_time_ratio",
	"DB Block Gets Per Sec":                "db_block_gets",
	"DBWR Checkpoints Per Sec":             "dbwr_checkpoints",
	"Disk Sort Per Sec":                    "disk_sorts",
	"Execute Without Parse Ratio":          "execute_without_parse",
	"GC CR Block Received Per Second":      "gc_cr_block_received",
	"GC Current Block Received Per Second": "gc_current_block_received",
	"Global Cache Blocks Corrupted":        "cache_blocks_corrupt",
	"Global Cache Blocks Lost":             "cache_blocks_lost",
	"Leaf Node Splits Per Sec":             "leaf_nodes_splits",
	"Logons Per Sec":                       "logons",
	"Long Table Scans Per Sec":             "long_table_scans",
	"Memory Sorts Ratio":                   "memory_sorts_ratio",
	"PGA Cache Hit %":                      "pga_cache_hit",
	"Process Limit %":                      "process_limit",
	"Redo Allocation Hit Ratio":            "redo_allocation_hit_ratio",
	"Redo Writes Per Sec":                  "redo_writes",
	"Row Cache Hit Ratio":                  "row_cache_hit_ratio",
	"Rows Per Sort":                        "rows_per_sort",
	"Session Limit %":                      "session_limit_usage",
	"Shared Pool Free %":                   "shared_pool_free",
	"Soft Parse Ratio":                     "soft_parse_ratio",
	"Temp Space Used":                      "temp_space_used",
	"Total Parse Count Per Sec":            "total_parse_count",
	"Total Sorts Per User Call":            "sorts_per_user_call",

	// CPU usage
	"CPU Usage Per Sec":            "active_sessions_on_cpu",
	"Database CPU Time Ratio":      "database_cpu_time_ratio",
	"Host CPU Utilization (%)":     "host_cpu_utilization",
	"Background CPU Usage Per Sec": "active_background_on_cpu",
	"CPU Usage Per Txn":            "per_txn_cpu_usage",

	// IO performance
	"Physical Read Bytes Per Sec":                   "physical_read_bytes",
	"Physical Write Bytes Per Sec":                  "physical_write_bytes",
	"Physical Read Total IO Requests Per Sec":       "physical_read_total_io_requests",
	"Physical Write Total Bytes Per Sec":            "physical_write_total_bytes",
	"Physical Read IO Requests Per Sec":             "physical_read_io_requests",
	"Physical Write IO Requests Per Sec":            "physical_write_io_requests",
	"Physical Reads Direct Lobs Per Sec":            "physical_reads_direct_lobs",
	"Physical Read Total Bytes Per Sec":             "physical_read_total_bytes",
	"Physical Reads Direct Per Sec":                 "physical_reads_direct",
	"Physical Reads Per Sec":                        "physical_reads",
	"Physical Write Total IO Requests Per Sec":      "physical_write_total_io_requests",
	"Physical Writes Direct Lobs Per Sec":           "physical_writes_direct_lobs",
	"Physical Writes Direct Per Sec":                "physical_writes_direct",
	"Physical Writes Per Sec":                       "physical_writes",
	"DB Block Changes Per Sec":                      "db_block_changes",
	"Redo Generated Per Sec":                        "redo_generated",
	"Average Synchronous Single-Block Read Latency": "avg_synchronous_single_block_read_latency",
	"I/O Megabytes per Second":                      "io_mb_per_seconds", // for newer oracle version

	// DB throughput and efficiency
	"Executions Per Sec":          "executions_per_sec", // 每秒 SQL 执行次数
	"Executions Per Txn":          "executions_per_txn", // 每个事务的 SQL 执行次数
	"Parse Failure Count Per Sec": "parse_failures",     // 每秒解析失败次数
	"Hard Parse Count Per Sec":    "hard_parses",        // 每秒硬解析次数
	"Logical Reads Per Sec":       "logical_reads",      // 每秒逻辑读次数
	// 缓冲区缓存命中率（通常通过 1 - (Physical Reads / (DB Block Gets + Consistent Gets)) 计算，
	// 但 V$SYSMETRIC 可能有直接的或间接相关的度量）
	"Buffer Cache Hit Ratio":    "buffer_cachehit_ratio",
	"Library Cache Hit Ratio":   "library_cachehit_ratio", // 库缓存命中率
	"SQL Service Response Time": "service_response_time",  // SQL 服务响应时间 (厘秒/调用)

	// session and txn activity
	"Current Logons Count":   "current_logins_count", // 当前登录用户数
	"Session Count":          "session_count",        // 当前会话总数
	"Active Sessions":        "active_sessions",      // （可能通过其他方式计算，但有些度量可能相关）当前活动会话数
	"User Commits Per Sec":   "user_commits",         // 每秒用户提交次数
	"User Rollbacks Per Sec": "user_rollbacks",       // 每秒用户回滚次数
	"Transactions Per Sec":   "txn_per_second",       // 每秒事务数

	// memory related metrics
	"Total PGA Allocated (Bytes)":         "", // 当前总共分配的 PGA 内存
	"Total PGA Used by Workareas (Bytes)": "", // 工作区使用的 PGA 内存

	// network related metrics
	"Network Traffic Volume Per Sec": "network_traffic_volume", // 网络流量

	// concurrent and lock
	"Enqueue Deadlocks Per Sec": "enqueue_deadlocks", // 每秒死锁次数
	"Enqueue Timeouts Per Sec":  "enqueue_timeouts",  // 每秒队列超时次数
	"Enqueue Waits Per Sec":     "eneneue_waits",     // 每秒队列等待次数

	// RAC (Real Application Clusters)
	"Global Cache Average CR Get Time":      "gc_average_cr_get_time",      // 全局缓存 CR 块平均获取时间
	"Global Cache Average Current Get Time": "gc_average_current_get_time", // 全局缓存 Current 块平均获取时间
}

var aliasMapping = map[string]string{
	"buffer_cache_hit_ratio":       "buffer_cachehit_ratio",
	"cursor_cache_hit_ratio":       "cursor_cachehit_ratio",
	"library_cache_hit_ratio":      "library_cachehit_ratio",
	"shared_pool_free_%":           "shared_pool_free",
	"physical_read_bytes_per_sec":  "physical_reads",
	"physical_write_bytes_per_sec": "physical_writes",
	"enqueue_timeouts_per_sec":     "enqueue_timeouts",

	"gc_cr_block_received_per_second": "gc_cr_block_received",
	"global_cache_blocks_corrupted":   "cache_blocks_corrupt",
	"global_cache_blocks_lost":        "cache_blocks_lost",
	"average_active_sessions":         "active_sessions",
	"sql_service_response_time":       "service_response_time",
	"user_rollbacks_per_sec":          "user_rollbacks",
	"total_sorts_per_user_call":       "sorts_per_user_call",
	"rows_per_sort":                   "rows_per_sort",
	"disk_sort_per_sec":               "disk_sorts",
	"memory_sorts_ratio":              "memory_sorts_ratio",
	"database_wait_time_ratio":        "database_wait_time_ratio",
	"session_limit_%":                 "session_limit_usage",
	"session_count":                   "session_count",
	"temp_space_used":                 "temp_space_used",
}

var (
	sqlSystemMetric = map[string]string{
		// for oracle 11g
		"11": `SELECT VALUE, METRIC_NAME 
FROM GV$SYSMETRIC ORDER BY BEGIN_TIME`,

		"default": `SELECT metric_name, value, metric_unit, name pdb_name 
FROM v$sysmetric s, v$containers c 
WHERE s.con_id = c.con_id(+)`,
	}

	sqlConSystemMetric = map[string]string{
		"default": `SELECT metric_name, value, metric_unit, name pdb_name
FROM v$con_sysmetric s, v$containers c 
WHERE s.con_id = c.con_id(+)`,
	}
)

type sysmetricsRowDB struct {
	MetricName sql.NullString  `db:"METRIC_NAME"`
	Value      sql.NullFloat64 `db:"VALUE"`
	MetricUnit sql.NullString  `db:"METRIC_UNIT,omitempty"`
	PdbName    sql.NullString  `db:"PDB_NAME,omitempty"`
}

func (ipt *Input) collectOracleSystem() {
	var (
		metricName = "oracle_system"
		start      = time.Now()
		pts        []*point.Point
	)

	sql := sqlConSystemMetric["default"]
	if x, ok := sqlConSystemMetric[ipt.mainVersion]; ok { // use version specific SQL.
		sql = x
	}

	rows := []sysmetricsRowDB{}
	if err := selectWrapper(ipt, &rows, sql, getMetricName(metricName, "oracle_system")); err != nil {
		l.Warnf("failed to collect system metrics(%q): %s, oracle version %s", sql, err, ipt.mainVersion)
	}

	var (
		opts = ipt.getKVsOpts()
		kvs  = ipt.getKVs()
	)

	kvs = kvs.AddTag("version", ipt.fullVersion).
		AddV2("uptime", ipt.Uptime, true)

	makeTagsAndFields := func(row sysmetricsRowDB, isExistedMap map[string]bool) {
		if metric, ok := systemCols[row.MetricName.String]; ok {
			value := row.Value.Float64

			switch row.MetricUnit.String {
			case "CentiSeconds Per Second":
				value /= 100
			case "SQL Service Response Time":
				value *= 10.0 // ds -> ms
			}

			if row.PdbName.Valid {
				kvs = kvs.AddTag("pdb_name", row.PdbName.String)
			}

			alias, ok := aliasMapping[metric]
			if ok {
				kvs = kvs.Add(alias, value, false, true)
			} else {
				kvs = kvs.Add(metric, value, false, true)
			}

			if isExistedMap != nil {
				isExistedMap[row.MetricName.String] = true
			}
		} else {
			l.Debugf("skip metric %q", row.MetricName.String)
		}
	}

	isExistedContainerMetric := map[string]bool{}
	for _, row := range rows {
		makeTagsAndFields(row, isExistedContainerMetric)
	}

	if kvs.FieldCount() > 0 {
		pts = append(pts, point.NewPointV2(metricName, kvs, opts...))
	}

	rows = rows[:0] // reset rows
	kvs = ipt.getKVs()

	sql = sqlSystemMetric["default"]
	if x, ok := sqlSystemMetric[ipt.mainVersion]; ok { // use version specific SQL.
		sql = x
	}

	if err := selectWrapper(ipt, &rows, sql, getMetricName(metricName, "oracle_con_system")); err != nil {
		l.Warnf("failed to collect system metrics(%q): %s, oracle version %s", sql, err, ipt.mainVersion)
	}

	isExistedGlobalMetric := map[string]bool{}
	for _, row := range rows {
		if _, ok := isExistedContainerMetric[row.MetricName.String]; ok {
			continue
		}

		if _, ok := isExistedGlobalMetric[row.MetricName.String]; ok {
			continue
		}

		makeTagsAndFields(row, isExistedGlobalMetric)
	}

	if kvs.FieldCount() > 0 {
		pts = append(pts, point.NewPointV2(metricName, kvs, opts...))
	}

	l.Debugf("collect %d points from system(oracle version %s)", len(pts), ipt.mainVersion)

	if err := ipt.feeder.Feed(point.Metric,
		pts,
		dkio.WithCollectCost(time.Since(start)),
		dkio.WithElection(ipt.Election),
		dkio.WithSource(inputName)); err != nil {
		l.Warnf("feeder.Feed: %s, ignored", err)
	}
}

func (ipt *Input) q(s string, names ...string) rows {
	now := time.Now()
	rows, err := ipt.db.Query(s)
	if err != nil {
		l.Errorf(`query failed, sql (%q), error: %s, ignored`, s, err.Error())
		return nil
	}

	var name string
	if len(names) == 1 {
		name = names[0]
	}

	if err := rows.Err(); err != nil {
		closeRows(rows)
		l.Errorf(`query row failed, sql (%q), error: %s, ignored`, s, err.Error())
		return nil
	} else {
		metricName, sqlName := getMetricNames(name)
		if len(sqlName) > 0 {
			sqlQueryCostSummary.WithLabelValues(metricName, sqlName).Observe(float64(time.Since(now)) / float64(time.Second))
		}
	}

	return rows
}

type rows interface {
	Next() bool
	Scan(...interface{}) error
	Close() error
	Err() error
	Columns() ([]string, error)
}

func closeRows(r rows) {
	if err := r.Close(); err != nil {
		l.Warnf("Close: %s, ignored", err)
	}
}

func getCleanCustomQueries(r rows) []map[string]interface{} {
	l.Debugf("getCleanCustomQueries entry")

	if r == nil {
		l.Debug("r == nil")
		return nil
	}

	defer closeRows(r)

	var list []map[string]interface{}

	columns, err := r.Columns()
	if err != nil {
		l.Errorf("Columns() failed: %v", err)
	}
	l.Debugf("columns = %v", columns)
	columnLength := len(columns)
	l.Debugf("columnLength = %d", columnLength)

	cache := make([]interface{}, columnLength)
	for idx := range cache {
		var a interface{}
		cache[idx] = &a
	}

	for r.Next() {
		l.Debug("Next() entry")

		if err := r.Scan(cache...); err != nil {
			l.Errorf("Scan failed: %v", err)
		}

		l.Debugf("len(cache) = %d", len(cache))

		item := make(map[string]interface{})
		for i, data := range cache {
			key := columns[i]
			val := *data.(*interface{})

			if val != nil {
				vType := reflect.TypeOf(val)

				l.Debugf("key = %s, vType = %s, %s", key, vType.String(), vType.Name())

				switch vType.String() {
				case "int64":
					if v, ok := val.(int64); ok {
						item[key] = v
					} else {
						l.Warn("expect int64, ignored")
					}
				case "string":
					var data interface{}
					data, err := strconv.ParseFloat(val.(string), 64)
					if err != nil {
						data = val
					}
					item[key] = data
				case "time.Time":
					if v, ok := val.(time.Time); ok {
						item[key] = v
					} else {
						l.Warn("expect time.Time, ignored")
					}
				case "[]uint8":
					item[key] = string(val.([]uint8))
				default:
					l.Warn("unsupport data type '%s', ignored", vType)
				}
			}
		}

		list = append(list, item)
	}

	if err := r.Err(); err != nil {
		l.Errorf("Err() failed: %v", err)
	}

	l.Debugf("len(list) = %d", len(list))

	return list
}
