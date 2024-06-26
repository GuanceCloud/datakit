// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package oracle

import (
	"crypto/md5" //nolint:gosec
	"database/sql"
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"github.com/araddon/dateparse"
	"github.com/spf13/cast"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/obfuscate"
)

// SQLProcess for Oracle 11g+.
const SQLProcess = `SELECT 
	name pdb_name, 
	pid, 
	program, 
	nvl(pga_used_mem,0) pga_used_mem, 
	nvl(pga_alloc_mem,0) pga_alloc_mem, 
	nvl(pga_freeable_mem,0) pga_freeable_mem, 
	nvl(pga_max_mem,0) pga_max_mem
  FROM v$process p, v$containers c
  WHERE
  	c.con_id(+) = p.con_id`

// SQLProcessOld for Oracle 11g and 11g-.
const SQLProcessOld = `SELECT 
    PROGRAM, 
    PGA_USED_MEM, 
    PGA_ALLOC_MEM, 
    PGA_FREEABLE_MEM, 
    PGA_MAX_MEM
  FROM GV$PROCESS`

type processesRowDB struct {
	PdbName        sql.NullString `db:"PDB_NAME"`
	PID            uint64         `db:"PID"`
	Program        sql.NullString `db:"PROGRAM"`
	PGAUsedMem     float64        `db:"PGA_USED_MEM"`
	PGAAllocMem    float64        `db:"PGA_ALLOC_MEM"`
	PGAFreeableMem float64        `db:"PGA_FREEABLE_MEM"`
	PGAMaxMem      float64        `db:"PGA_MAX_MEM"`
}

func (ipt *Input) collectOracleProcess() (category point.Category, pts []*point.Point, err error) {
	category = point.Metric
	metricName := "oracle_process"
	rows := []processesRowDB{}

	if sql, ok := ipt.cacheSQL[metricName]; ok {
		if err = selectWrapper(ipt, &rows, sql); err != nil {
			err = fmt.Errorf("failed to collect processes info: %w", err)
			return
		}
	} else {
		err = selectWrapper(ipt, &rows, SQLProcess)
		ipt.cacheSQL[metricName] = SQLProcess
		l.Infof("Query for metric [%s], sql: %s", metricName, SQLProcess)
		if err != nil {
			if strings.Contains(err.Error(), "ORA-00942: table or view does not exist") {
				ipt.cacheSQL[metricName] = SQLProcessOld
				l.Infof("Query for metric [%s], sql: %s", metricName, SQLProcessOld)
				// oracle old version. 11g
				if err = selectWrapper(ipt, &rows, SQLProcessOld); err != nil {
					err = fmt.Errorf("failed to collect old processes info: %w", err)
					return
				}
			} else {
				err = fmt.Errorf("failed to collect processes info: %w", err)
				return
			}
		}
	}

	pts = make([]*point.Point, 0)
	for _, row := range rows {
		tags := map[string]string{}
		if row.PdbName.Valid {
			tags["pdb_name"] = row.PdbName.String
		}

		if row.Program.Valid {
			tags["program"] = row.Program.String
		}
		fields := map[string]interface{}{
			"pga_alloc_mem":    row.PGAAllocMem,
			"pga_freeable_mem": row.PGAFreeableMem,
			"pga_max_mem":      row.PGAMaxMem,
			"pga_used_mem":     row.PGAUsedMem,
		}
		if row.PID > 0 {
			fields["pid"] = row.PID
		}

		pts = append(pts, ipt.buildPoint(metricName, tags, fields, false))
	}

	return category, pts, nil
}

// SQLTableSpace for Oracle 11g+.
const SQLTableSpace = `SELECT
  c.name pdb_name,
  t.tablespace_name tablespace_name,
  NVL(m.used_space * t.block_size, 0) used,
  NVL(m.tablespace_size * t.block_size, 0) size_,
  NVL(m.used_percent, 0) in_use,
  NVL2(m.used_space, 0, 1) offline_
FROM
  cdb_tablespace_usage_metrics m, cdb_tablespaces t, v$containers c
WHERE
  m.tablespace_name(+) = t.tablespace_name and c.con_id(+) = t.con_id`

// SQLTableSpaceOld for Oracle 11g and 11g-.
const SQLTableSpaceOld = `SELECT
  m.tablespace_name,
  NVL(m.used_space * t.block_size, 0) as used,
  m.tablespace_size * t.block_size as size_,
  NVL(m.used_percent, 0) as in_use,
  NVL2(m.used_space, 0, 1) as offline_
FROM
  dba_tablespace_usage_metrics m
  join dba_tablespaces t on m.tablespace_name = t.tablespace_name`

type tableSpaceRowDB struct {
	PdbName        sql.NullString `db:"PDB_NAME"`
	TablespaceName string         `db:"TABLESPACE_NAME"`
	Used           float64        `db:"USED"`
	Size           float64        `db:"SIZE_"`
	InUse          float64        `db:"IN_USE"`
	Offline        float64        `db:"OFFLINE_"`
}

func (ipt *Input) collectOracleTableSpace() (category point.Category, pts []*point.Point, err error) {
	category = point.Metric
	metricName := "oracle_tablespace"
	rows := []tableSpaceRowDB{}

	if sql, ok := ipt.cacheSQL[metricName]; ok {
		if err = selectWrapper(ipt, &rows, sql); err != nil {
			err = fmt.Errorf("failed to collect table space: %w", err)
			return
		}
	} else {
		err = selectWrapper(ipt, &rows, SQLTableSpace)
		ipt.cacheSQL[metricName] = SQLTableSpace
		l.Infof("Query for metric [%s], sql: %s", metricName, SQLTableSpace)
		if err != nil {
			if strings.Contains(err.Error(), "ORA-00942: table or view does not exist") {
				ipt.cacheSQL[metricName] = SQLTableSpaceOld
				l.Infof("Query for metric [%s], sql: %s", metricName, SQLTableSpaceOld)
				// oracle old version. 11g
				if err = selectWrapper(ipt, &rows, SQLTableSpaceOld); err != nil {
					err = fmt.Errorf("failed to collect old table space info: %w", err)
					return
				}
			} else {
				err = fmt.Errorf("failed to collect table space info: %w", err)
				return
			}
		}
	}

	pts = make([]*point.Point, 0)
	for _, row := range rows {
		tags := map[string]string{
			"tablespace_name": row.TablespaceName,
		}
		if row.PdbName.Valid {
			tags["pdb_name"] = row.PdbName.String
		}

		fields := map[string]interface{}{
			"in_use":     row.InUse,
			"off_use":    row.Offline,
			"ts_size":    row.Size,
			"used_space": row.Used,
		}

		pts = append(pts, ipt.buildPoint(metricName, tags, fields, false))
	}

	return category, pts, nil
}

var systemCols = map[string]sysMetricsDefinition{
	"Average Active Sessions":                       {DDmetric: "active_sessions"},
	"Average Synchronous Single-Block Read Latency": {DDmetric: "avg_synchronous_single_block_read_latency", DBM: true},
	"Background CPU Usage Per Sec":                  {DDmetric: "active_background_on_cpu", DBM: true},
	"Background Time Per Sec":                       {DDmetric: "active_background", DBM: true},
	"Branch Node Splits Per Sec":                    {DDmetric: "branch_node_splits", DBM: true},
	"Buffer Cache Hit Ratio":                        {DDmetric: "buffer_cachehit_ratio"},
	"Consistent Read Changes Per Sec":               {DDmetric: "consistent_read_changes", DBM: true},
	"Consistent Read Gets Per Sec":                  {DDmetric: "consistent_read_gets", DBM: true},
	"CPU Usage Per Sec":                             {DDmetric: "active_sessions_on_cpu", DBM: true},
	"Current OS Load":                               {DDmetric: "os_load", DBM: true},
	"Database CPU Time Ratio":                       {DDmetric: "database_cpu_time_ratio", DBM: true},
	"Database Wait Time Ratio":                      {DDmetric: "database_wait_time_ratio"},
	"DB Block Changes Per Sec":                      {DDmetric: "db_block_changes", DBM: true},
	"DB Block Gets Per Sec":                         {DDmetric: "db_block_gets", DBM: true},
	"DBWR Checkpoints Per Sec":                      {DDmetric: "dbwr_checkpoints", DBM: true},
	"Disk Sort Per Sec":                             {DDmetric: "disk_sorts"},
	"Enqueue Deadlocks Per Sec":                     {DDmetric: "enqueue_deadlocks", DBM: true},
	"Enqueue Timeouts Per Sec":                      {DDmetric: "enqueue_timeouts"},
	"Execute Without Parse Ratio":                   {DDmetric: "execute_without_parse", DBM: true},
	"GC CR Block Received Per Second":               {DDmetric: "gc_cr_block_received"},
	"GC Current Block Received Per Second":          {DDmetric: "gc_current_block_received", DBM: true},
	"Global Cache Average CR Get Time":              {DDmetric: "gc_average_cr_get_time", DBM: true},
	"Global Cache Average Current Get Time":         {DDmetric: "gc_average_current_get_time", DBM: true},
	"Global Cache Blocks Corrupted":                 {DDmetric: "cache_blocks_corrupt"},
	"Global Cache Blocks Lost":                      {DDmetric: "cache_blocks_lost"},
	"Hard Parse Count Per Sec":                      {DDmetric: "hard_parses", DBM: true},
	"Host CPU Utilization (%)":                      {DDmetric: "host_cpu_utilization", DBM: true},
	"Leaf Node Splits Per Sec":                      {DDmetric: "leaf_nodes_splits", DBM: true},
	"Library Cache Hit Ratio":                       {DDmetric: "library_cachehit_ratio"},
	"Logical Reads Per Sec":                         {DDmetric: "logical_reads", DBM: true},
	"Logons Per Sec":                                {DDmetric: "logons"},
	"Long Table Scans Per Sec":                      {DDmetric: "long_table_scans"},
	"Memory Sorts Ratio":                            {DDmetric: "memory_sorts_ratio"},
	"Network Traffic Volume Per Sec":                {DDmetric: "network_traffic_volume", DBM: true},
	"PGA Cache Hit %":                               {DDmetric: "pga_cache_hit", DBM: true},
	"Parse Failure Count Per Sec":                   {DDmetric: "parse_failures", DBM: true},
	"Physical Read Bytes Per Sec":                   {DDmetric: "physical_read_bytes", DBM: true},
	"Physical Read IO Requests Per Sec":             {DDmetric: "physical_read_io_requests", DBM: true},
	"Physical Read Total IO Requests Per Sec":       {DDmetric: "physical_read_total_io_requests", DBM: true},
	"Physical Reads Direct Lobs Per Sec":            {DDmetric: "physical_reads_direct_lobs", DBM: true},
	"Physical Read Total Bytes Per Sec":             {DDmetric: "physical_read_total_bytes", DBM: true},
	"Physical Reads Direct Per Sec":                 {DDmetric: "physical_reads_direct", DBM: true},
	"Physical Reads Per Sec":                        {DDmetric: "physical_reads"},
	"Physical Write Bytes Per Sec":                  {DDmetric: "physical_write_bytes", DBM: true},
	"Physical Write IO Requests Per Sec":            {DDmetric: "physical_write_io_requests", DBM: true},
	"Physical Write Total Bytes Per Sec":            {DDmetric: "physical_write_total_bytes", DBM: true},
	"Physical Write Total IO Requests Per Sec":      {DDmetric: "physical_write_total_io_requests", DBM: true},
	"Physical Writes Direct Lobs Per Sec":           {DDmetric: "physical_writes_direct_lobs", DBM: true},
	"Physical Writes Direct Per Sec":                {DDmetric: "physical_writes_direct", DBM: true},
	"Physical Writes Per Sec":                       {DDmetric: "physical_writes"},
	"Process Limit %":                               {DDmetric: "process_limit", DBM: true},
	"Redo Allocation Hit Ratio":                     {DDmetric: "redo_allocation_hit_ratio", DBM: true},
	"Redo Generated Per Sec":                        {DDmetric: "redo_generated", DBM: true},
	"Redo Writes Per Sec":                           {DDmetric: "redo_writes", DBM: true},
	"Row Cache Hit Ratio":                           {DDmetric: "row_cache_hit_ratio", DBM: true},
	"Rows Per Sort":                                 {DDmetric: "rows_per_sort"},
	"SQL Service Response Time":                     {DDmetric: "service_response_time"},
	"Session Count":                                 {DDmetric: "session_count"},
	"Session Limit %":                               {DDmetric: "session_limit_usage"},
	"Shared Pool Free %":                            {DDmetric: "shared_pool_free"},
	"Soft Parse Ratio":                              {DDmetric: "soft_parse_ratio", DBM: true},
	"Temp Space Used":                               {DDmetric: "temp_space_used"},
	"Total Parse Count Per Sec":                     {DDmetric: "total_parse_count", DBM: true},
	"Total Sorts Per User Call":                     {DDmetric: "sorts_per_user_call"},
	"User Commits Per Sec":                          {DDmetric: "user_commits", DBM: true},
	"User Rollbacks Per Sec":                        {DDmetric: "user_rollbacks"},
}

var dic = map[string]string{
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

// SQLSystem is the get system info SQL query for Oracle for 12c2 and 12c2+.
// https://docs.oracle.com/en/database/oracle/oracle-database/12.2/refrn/V-CON_SYSMETRIC.html#GUID-B7A51ACA-952B-47FE-9BC5-B265A120A9F7
const SQLSystem = `SELECT 
	metric_name,
	value, 
	metric_unit, 
	--(end_time - begin_time)*24*3600 interval_length,
	name pdb_name 
  FROM %s s, v$containers c 
  WHERE s.con_id = c.con_id(+)`

// SQLSystemOld is the get system info SQL query for Oracle for for 12c1 and 12c1-.
const SQLSystemOld = `SELECT 
    VALUE, 
    METRIC_NAME 
  FROM GV$SYSMETRIC ORDER BY BEGIN_TIME`

type sysmetricsRowDB struct {
	MetricName sql.NullString  `db:"METRIC_NAME"`
	Value      sql.NullFloat64 `db:"VALUE"`
	MetricUnit sql.NullString  `db:"METRIC_UNIT,omitempty"`
	PdbName    sql.NullString  `db:"PDB_NAME,omitempty"`
}

type sysMetricsDefinition struct {
	DDmetric string
	DBM      bool
}

func (ipt *Input) collectOracleSystem() (category point.Category, pts []*point.Point, err error) {
	category = point.Metric
	metricName := "oracle_system"

	rows := []sysmetricsRowDB{}
	containerMetricName := fmt.Sprintf("%s-container", metricName)
	if sql, ok := ipt.cacheSQL[containerMetricName]; ok {
		if err = selectWrapper(ipt, &rows, sql); err != nil {
			err = fmt.Errorf("failed to collect system metrics: %w", err)
			return
		}
	} else {
		containerSQL := fmt.Sprintf(SQLSystem, "v$con_sysmetric")
		err = selectWrapper(ipt, &rows, containerSQL)
		ipt.cacheSQL[containerMetricName] = containerSQL
		l.Infof("Query for metric [%s], sql: %s", metricName, containerSQL)
		if err != nil {
			if strings.Contains(err.Error(), "ORA-00942: table or view does not exist") {
				ipt.cacheSQL[containerMetricName] = SQLSystemOld
				l.Infof("Query for metric [%s], sql: %s", metricName, SQLSystemOld)
				if err = selectWrapper(ipt, &rows, SQLSystemOld); err != nil {
					err = fmt.Errorf("failed to collect old system metrics: %w", err)
					return
				}
			} else {
				err = fmt.Errorf("failed to collect system metrics: %w", err)
				return
			}
		}
	}

	pts = make([]*point.Point, 0)
	tags := map[string]string{}
	fields := map[string]interface{}{}
	makeTagsAndFields := func(row sysmetricsRowDB, isExistedMap map[string]bool) {
		if metric, ok := systemCols[row.MetricName.String]; ok {
			value := row.Value.Float64
			if row.MetricUnit.String == "CentiSeconds Per Second" {
				value /= 100
			}
			if row.PdbName.Valid {
				tags["pdb_name"] = row.PdbName.String
			}

			alias, ok := dic[metric.DDmetric]
			if ok {
				fields[alias] = value
			} else {
				fields[metric.DDmetric] = value
			}

			if isExistedMap != nil {
				isExistedMap[row.MetricName.String] = true
			}
		}
	}

	isExistedContainerMetric := map[string]bool{}
	for _, row := range rows {
		makeTagsAndFields(row, isExistedContainerMetric)
	}

	if len(fields) > 0 {
		pts = append(pts, ipt.buildPoint(metricName, tags, fields, false))
	}

	// if old version, return
	if ipt.cacheSQL[containerMetricName] == SQLSystemOld {
		return
	}

	rows = rows[:0]
	tags = map[string]string{}
	fields = map[string]interface{}{}
	systemMetricName := fmt.Sprintf("%s-system", metricName)

	if sql, ok := ipt.cacheSQL[systemMetricName]; ok {
		if err = selectWrapper(ipt, &rows, sql); err != nil {
			err = fmt.Errorf("failed to collect system metrics: %w", err)
			return
		}
	} else {
		systemSQL := fmt.Sprintf(SQLSystem, "v$sysmetric") + " ORDER BY begin_time ASC, metric_name ASC"
		err = selectWrapper(ipt, &rows, systemSQL)
		ipt.cacheSQL[systemMetricName] = systemSQL
		l.Infof("Query for metric [%s], sql: %s", metricName, systemSQL)
		if err != nil {
			err = fmt.Errorf("failed to collect system metrics: %w", err)
			return
		}
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

	return category, pts, nil
}

func (ipt *Input) q(s string) rows {
	rows, err := ipt.db.Query(s)
	if err != nil {
		l.Errorf(`query failed, sql (%q), error: %s, ignored`, s, err.Error())
		return nil
	}

	if err := rows.Err(); err != nil {
		closeRows(rows)
		l.Errorf(`query row failed, sql (%q), error: %s, ignored`, s, err.Error())
		return nil
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

func GetMD5String32(bt []byte) string {
	return fmt.Sprintf("%X", md5.Sum(bt)) // nolint:gosec
}

const SQLSlow = `SELECT 
	sa.FIRST_LOAD_TIME,
	sa.INVALIDATIONS,
	sa.PARSE_CALLS,
	sa.DISK_READS,
	sa.DIRECT_WRITES,
	sa.BUFFER_GETS,
	sa.APPLICATION_WAIT_TIME,
	sa.CONCURRENCY_WAIT_TIME,
	sa.CLUSTER_WAIT_TIME,
	sa.USER_IO_WAIT_TIME,
	sa.PLSQL_EXEC_TIME,
	sa.JAVA_EXEC_TIME,
	sa.ROWS_PROCESSED,
	sa.COMMAND_TYPE,
	sa.OPTIMIZER_MODE,
	sa.OPTIMIZER_COST,
	sa.PARSING_USER_ID,
	sa.PARSING_SCHEMA_ID,
	sa.PARSING_SCHEMA_NAME,
	sa.KEPT_VERSIONS,
	sa.HASH_VALUE,
	sa.OLD_HASH_VALUE,
	sa.PLAN_HASH_VALUE,
	sa.MODULE,
	sa.MODULE_HASH,
	sa.ACTION,
	sa.ACTION_HASH,
	sa.SERIALIZABLE_ABORTS,
	sa.OUTLINE_CATEGORY,
	sa.CPU_TIME,
	sa.ELAPSED_TIME,
	sa.OUTLINE_SID,
	sa.REMOTE,
	sa.OBJECT_STATUS,
	sa.LITERAL_HASH_VALUE,
	sa.LAST_LOAD_TIME,
	sa.IS_OBSOLETE,
	sa.IS_BIND_SENSITIVE,
	sa.IS_BIND_AWARE,
	sa.CHILD_LATCH,
	sa.SQL_PROFILE,
	sa.SQL_PATCH,
	sa.SQL_PLAN_BASELINE,
	sa.PROGRAM_ID,
	sa.PROGRAM_LINE#,
	sa.EXACT_MATCHING_SIGNATURE,
	sa.FORCE_MATCHING_SIGNATURE,
	sa.LAST_ACTIVE_TIME,
	sa.TYPECHECK_MEM,
	sa.IO_CELL_OFFLOAD_ELIGIBLE_BYTES,
	sa.IO_INTERCONNECT_BYTES,
	sa.PHYSICAL_READ_REQUESTS,
	sa.PHYSICAL_READ_BYTES,
	sa.PHYSICAL_WRITE_REQUESTS,
	sa.PHYSICAL_WRITE_BYTES,
	sa.OPTIMIZED_PHY_READ_REQUESTS,
	sa.LOCKED_TOTAL,
	sa.PINNED_TOTAL,
	sa.IO_CELL_UNCOMPRESSED_BYTES,
	sa.IO_CELL_OFFLOAD_RETURNED_BYTES,
	sa.SQL_FULLTEXT,
	sa.SQL_ID,
	sa.SHARABLE_MEM,
	sa.PERSISTENT_MEM,
	sa.RUNTIME_MEM,
	sa.SORTS,
	sa.VERSION_COUNT,
	sa.LOADED_VERSIONS,
	sa.OPEN_VERSIONS,
	sa.USERS_OPENING,
	sa.FETCHES,
	sa.EXECUTIONS,
	sa.PX_SERVERS_EXECUTIONS,
	sa.END_OF_FETCH_COUNT,
	sa.USERS_EXECUTING,
	sa.LOADS,
	sa.ELAPSED_TIME/sa.EXECUTIONS AVG_ELAPSED,
	u.USERNAME
  FROM V$SQLAREA sa
  LEFT JOIN ALL_USERS u
  ON sa.PARSING_USER_ID = u.USER_ID
  WHERE
  	EXECUTIONS > 0 
		AND sa.ELAPSED_TIME/sa.EXECUTIONS > %d 
		AND LAST_ACTIVE_TIME > to_date('%s','yyyy-mm-dd hh24:mi:ss')
  `

//nolint:stylecheck
type slowQueryRowDB struct {
	FIRST_LOAD_TIME                sql.NullString  `db:"FIRST_LOAD_TIME"`
	INVALIDATIONS                  sql.NullFloat64 `db:"INVALIDATIONS"`
	PARSE_CALLS                    sql.NullFloat64 `db:"PARSE_CALLS"`
	DISK_READS                     sql.NullFloat64 `db:"DISK_READS"`
	DIRECT_WRITES                  sql.NullFloat64 `db:"DIRECT_WRITES"`
	BUFFER_GETS                    sql.NullFloat64 `db:"BUFFER_GETS"`
	APPLICATION_WAIT_TIME          sql.NullFloat64 `db:"APPLICATION_WAIT_TIME"`
	CONCURRENCY_WAIT_TIME          sql.NullFloat64 `db:"CONCURRENCY_WAIT_TIME"`
	CLUSTER_WAIT_TIME              sql.NullFloat64 `db:"CLUSTER_WAIT_TIME"`
	USER_IO_WAIT_TIME              sql.NullFloat64 `db:"USER_IO_WAIT_TIME"`
	PLSQL_EXEC_TIME                sql.NullFloat64 `db:"PLSQL_EXEC_TIME"`
	JAVA_EXEC_TIME                 sql.NullFloat64 `db:"JAVA_EXEC_TIME"`
	ROWS_PROCESSED                 sql.NullFloat64 `db:"ROWS_PROCESSED"`
	COMMAND_TYPE                   sql.NullFloat64 `db:"COMMAND_TYPE"`
	OPTIMIZER_MODE                 sql.NullString  `db:"OPTIMIZER_MODE"`
	OPTIMIZER_COST                 sql.NullFloat64 `db:"OPTIMIZER_COST"`
	PARSING_USER_ID                sql.NullFloat64 `db:"PARSING_USER_ID"`
	PARSING_SCHEMA_ID              sql.NullFloat64 `db:"PARSING_SCHEMA_ID"`
	PARSING_SCHEMA_NAME            sql.NullString  `db:"PARSING_SCHEMA_NAME"`
	KEPT_VERSIONS                  sql.NullString  `db:"KEPT_VERSIONS"`
	HASH_VALUE                     sql.NullString  `db:"HASH_VALUE"`
	OLD_HASH_VALUE                 sql.NullString  `db:"OLD_HASH_VALUE"`
	PLAN_HASH_VALUE                sql.NullString  `db:"PLAN_HASH_VALUE"`
	MODULE                         sql.NullString  `db:"MODULE"`
	MODULE_HASH                    sql.NullString  `db:"MODULE_HASH"`
	ACTION                         sql.NullString  `db:"ACTION"`
	ACTION_HASH                    sql.NullString  `db:"ACTION_HASH"`
	SERIALIZABLE_ABORTS            sql.NullString  `db:"SERIALIZABLE_ABORTS"`
	OUTLINE_CATEGORY               sql.NullString  `db:"OUTLINE_CATEGORY"`
	CPU_TIME                       sql.NullFloat64 `db:"CPU_TIME"`
	ELAPSED_TIME                   sql.NullFloat64 `db:"ELAPSED_TIME"`
	OUTLINE_SID                    sql.NullString  `db:"OUTLINE_SID"`
	REMOTE                         sql.NullString  `db:"REMOTE"`
	OBJECT_STATUS                  sql.NullString  `db:"OBJECT_STATUS"`
	LITERAL_HASH_VALUE             sql.NullString  `db:"LITERAL_HASH_VALUE"`
	LAST_LOAD_TIME                 sql.NullString  `db:"LAST_LOAD_TIME"`
	IS_OBSOLETE                    sql.NullString  `db:"IS_OBSOLETE"`
	IS_BIND_SENSITIVE              sql.NullString  `db:"IS_BIND_SENSITIVE"`
	IS_BIND_AWARE                  sql.NullString  `db:"IS_BIND_AWARE"`
	CHILD_LATCH                    sql.NullString  `db:"CHILD_LATCH"`
	SQL_PROFILE                    sql.NullString  `db:"SQL_PROFILE"`
	SQL_PATCH                      sql.NullString  `db:"SQL_PATCH"`
	SQL_PLAN_BASELINE              sql.NullString  `db:"SQL_PLAN_BASELINE"`
	PROGRAM_ID                     sql.NullString  `db:"PROGRAM_ID"`
	PROGRAM_LINE                   sql.NullString  `db:"PROGRAM_LINE#"`
	EXACT_MATCHING_SIGNATURE       sql.NullString  `db:"EXACT_MATCHING_SIGNATURE"`
	FORCE_MATCHING_SIGNATURE       sql.NullString  `db:"FORCE_MATCHING_SIGNATURE"`
	LAST_ACTIVE_TIME               sql.NullString  `db:"LAST_ACTIVE_TIME"`
	TYPECHECK_MEM                  sql.NullString  `db:"TYPECHECK_MEM"`
	IO_CELL_OFFLOAD_ELIGIBLE_BYTES sql.NullString  `db:"IO_CELL_OFFLOAD_ELIGIBLE_BYTES"`
	IO_INTERCONNECT_BYTES          sql.NullString  `db:"IO_INTERCONNECT_BYTES"`
	PHYSICAL_READ_REQUESTS         sql.NullString  `db:"PHYSICAL_READ_REQUESTS"`
	PHYSICAL_READ_BYTES            sql.NullString  `db:"PHYSICAL_READ_BYTES"`
	PHYSICAL_WRITE_REQUESTS        sql.NullString  `db:"PHYSICAL_WRITE_REQUESTS"`
	PHYSICAL_WRITE_BYTES           sql.NullString  `db:"PHYSICAL_WRITE_BYTES"`
	OPTIMIZED_PHY_READ_REQUESTS    sql.NullString  `db:"OPTIMIZED_PHY_READ_REQUESTS"`
	LOCKED_TOTAL                   sql.NullString  `db:"LOCKED_TOTAL"`
	PINNED_TOTAL                   sql.NullString  `db:"PINNED_TOTAL"`
	IO_CELL_UNCOMPRESSED_BYTES     sql.NullString  `db:"IO_CELL_UNCOMPRESSED_BYTES"`
	IO_CELL_OFFLOAD_RETURNED_BYTES sql.NullString  `db:"IO_CELL_OFFLOAD_RETURNED_BYTES"`
	SQL_FULLTEXT                   sql.NullString  `db:"SQL_FULLTEXT"`
	SQL_ID                         sql.NullString  `db:"SQL_ID"`
	SHARABLE_MEM                   sql.NullString  `db:"SHARABLE_MEM"`
	PERSISTENT_MEM                 sql.NullString  `db:"PERSISTENT_MEM"`
	RUNTIME_MEM                    sql.NullString  `db:"RUNTIME_MEM"`
	SORTS                          sql.NullString  `db:"SORTS"`
	VERSION_COUNT                  sql.NullString  `db:"VERSION_COUNT"`
	LOADED_VERSIONS                sql.NullString  `db:"LOADED_VERSIONS"`
	OPEN_VERSIONS                  sql.NullString  `db:"OPEN_VERSIONS"`
	USERS_OPENING                  sql.NullString  `db:"USERS_OPENING"`
	FETCHES                        sql.NullString  `db:"FETCHES"`
	EXECUTIONS                     sql.NullString  `db:"EXECUTIONS"`
	PX_SERVERS_EXECUTIONS          sql.NullString  `db:"PX_SERVERS_EXECUTIONS"`
	END_OF_FETCH_COUNT             sql.NullString  `db:"END_OF_FETCH_COUNT"`
	USERS_EXECUTING                sql.NullString  `db:"USERS_EXECUTING"`
	LOADS                          sql.NullString  `db:"LOADS"`
	AVG_ELAPSED                    sql.NullFloat64 `db:"AVG_ELAPSED"`
	USERNAME                       sql.NullString  `db:"USERNAME"`
}

const SQLQueryMaxActive = `SELECT MAX(LAST_ACTIVE_TIME) FROM V$SQLAREA`

//nolint:stylecheck
type maxQueryRowDB struct {
	MAX_LAST_ACTIVE_TIME sql.NullString `db:"MAX(LAST_ACTIVE_TIME)"`
}

func (ipt *Input) collectSlowQuery() (category point.Category, pts []*point.Point, err error) {
	category = point.Logging

	if len(ipt.lastActiveTime) == 0 {
		rows := []maxQueryRowDB{}
		err = selectWrapper(ipt, &rows, SQLQueryMaxActive)
		if err != nil {
			err = fmt.Errorf("failed to collect max query info: %w", err)
			return
		}

		for _, r := range rows {
			ipt.lastActiveTime = getOracleTimeString(r.MAX_LAST_ACTIVE_TIME.String)
		}

		l.Debugf("m.lastActiveTime =%s", ipt.lastActiveTime)

		return
	}

	rows := []slowQueryRowDB{}

	query := fmt.Sprintf(SQLSlow, ipt.slowQueryTime.Microseconds(), ipt.lastActiveTime)

	err = selectWrapper(ipt, &rows, query)
	if err != nil {
		err = fmt.Errorf("failed to collect slow query info: %w", err)
		return
	}

	if len(rows) == 0 {
		return
	}

	mResults := make([]map[string]interface{}, 0)

	for _, r := range rows {
		gotlastActiveTime, err := dateparse.ParseAny(r.LAST_ACTIVE_TIME.String)
		if err != nil {
			l.Warnf("parse LAST_ACTIVE_TIME(%s) failed: %s, ignored", r.LAST_ACTIVE_TIME.String, err.Error())
			continue
		}
		savedLastActiveTime, err := dateparse.ParseAny(ipt.lastActiveTime)
		if err != nil {
			l.Warnf("parse lastActiveTime(%s) failed: %s, ingored", ipt.lastActiveTime, err.Error())
			continue
		}
		if gotlastActiveTime.After(savedLastActiveTime) {
			ipt.lastActiveTime = getOracleTimeString(r.LAST_ACTIVE_TIME.String) // update saved.
		}

		mRes := make(map[string]interface{}, 78)

		mRes["first_load_time"] = r.FIRST_LOAD_TIME.String
		mRes["invalidations"] = r.INVALIDATIONS.Float64
		mRes["parse_calls"] = r.PARSE_CALLS.Float64
		mRes["disk_reads"] = r.DISK_READS.Float64
		mRes["direct_writes"] = r.DIRECT_WRITES.Float64
		mRes["buffer_gets"] = r.BUFFER_GETS.Float64
		mRes["application_wait_time"] = r.APPLICATION_WAIT_TIME.Float64
		mRes["concurrency_wait_time"] = r.CONCURRENCY_WAIT_TIME.Float64
		mRes["cluster_wait_time"] = r.CLUSTER_WAIT_TIME.Float64
		mRes["user_io_wait_time"] = r.USER_IO_WAIT_TIME.Float64
		mRes["plsql_exec_time"] = r.PLSQL_EXEC_TIME.Float64
		mRes["java_exec_time"] = r.JAVA_EXEC_TIME.Float64
		mRes["rows_processed"] = r.ROWS_PROCESSED.Float64
		mRes["command_type"] = r.COMMAND_TYPE.Float64
		mRes["optimizer_mode"] = r.OPTIMIZER_MODE.String
		mRes["optimizer_cost"] = r.OPTIMIZER_COST.Float64
		mRes["parsing_user_id"] = r.PARSING_USER_ID.Float64
		mRes["parsing_schema_id"] = r.PARSING_SCHEMA_ID.Float64
		mRes["parsing_schema_name"] = r.PARSING_SCHEMA_NAME.String
		mRes["kept_versions"] = r.KEPT_VERSIONS.String
		mRes["hash_value"] = r.HASH_VALUE.String
		mRes["old_hash_value"] = r.OLD_HASH_VALUE.String
		mRes["plan_hash_value"] = r.PLAN_HASH_VALUE.String
		mRes["module"] = r.MODULE.String
		mRes["module_hash"] = r.MODULE_HASH.String
		mRes["action"] = r.ACTION.String
		mRes["action_hash"] = r.ACTION_HASH.String
		mRes["serializable_aborts"] = r.SERIALIZABLE_ABORTS.String
		mRes["outline_category"] = r.OUTLINE_CATEGORY.String
		mRes["cpu_time"] = r.CPU_TIME.Float64
		mRes["elapsed_time"] = r.ELAPSED_TIME.Float64
		mRes["outline_sid"] = r.OUTLINE_SID.String
		mRes["remote"] = r.REMOTE.String
		mRes["object_status"] = r.OBJECT_STATUS.String
		mRes["literal_hash_value"] = r.LITERAL_HASH_VALUE.String
		mRes["last_load_time"] = r.LAST_LOAD_TIME.String
		mRes["is_obsolete"] = r.IS_OBSOLETE.String
		mRes["is_bind_sensitive"] = r.IS_BIND_SENSITIVE.String
		mRes["is_bind_aware"] = r.IS_BIND_AWARE.String
		mRes["child_latch"] = r.CHILD_LATCH.String
		mRes["sql_profile"] = r.SQL_PROFILE.String
		mRes["sql_patch"] = r.SQL_PATCH.String
		mRes["sql_plan_baseline"] = r.SQL_PLAN_BASELINE.String
		mRes["program_id"] = r.PROGRAM_ID.String
		mRes["program_line"] = r.PROGRAM_LINE.String
		mRes["exact_matching_signature"] = r.EXACT_MATCHING_SIGNATURE.String
		mRes["force_matching_signature"] = r.FORCE_MATCHING_SIGNATURE.String
		mRes["last_active_time"] = r.LAST_ACTIVE_TIME.String
		mRes["typecheck_mem"] = r.TYPECHECK_MEM.String
		mRes["io_cell_offload_eligible_bytes"] = r.IO_CELL_OFFLOAD_ELIGIBLE_BYTES.String
		mRes["io_interconnect_bytes"] = r.IO_INTERCONNECT_BYTES.String
		mRes["physical_read_requests"] = r.PHYSICAL_READ_REQUESTS.String
		mRes["physical_read_bytes"] = r.PHYSICAL_READ_BYTES.String
		mRes["physical_write_requests"] = r.PHYSICAL_WRITE_REQUESTS.String
		mRes["physical_write_bytes"] = r.PHYSICAL_WRITE_BYTES.String
		mRes["optimized_phy_read_requests"] = r.OPTIMIZED_PHY_READ_REQUESTS.String
		mRes["locked_total"] = r.LOCKED_TOTAL.String
		mRes["pinned_total"] = r.PINNED_TOTAL.String
		mRes["io_cell_uncompressed_bytes"] = r.IO_CELL_UNCOMPRESSED_BYTES.String
		mRes["io_cell_offload_returned_bytes"] = r.IO_CELL_OFFLOAD_RETURNED_BYTES.String

		fullText, err := obfuscateSQL(r.SQL_FULLTEXT.String)
		if err != nil {
			mRes["failed_obfuscate"] = err.Error()
		}
		mRes["sql_fulltext"] = fullText

		mRes["sql_id"] = r.SQL_ID.String
		mRes["sharable_mem"] = r.SHARABLE_MEM.String
		mRes["persistent_mem"] = r.PERSISTENT_MEM.String
		mRes["runtime_mem"] = r.RUNTIME_MEM.String
		mRes["sorts"] = r.SORTS.String
		mRes["version_count"] = r.VERSION_COUNT.String
		mRes["loaded_versions"] = r.LOADED_VERSIONS.String
		mRes["open_versions"] = r.OPEN_VERSIONS.String
		mRes["users_opening"] = r.USERS_OPENING.String
		mRes["fetches"] = r.FETCHES.String
		mRes["executions"] = r.EXECUTIONS.String
		mRes["px_servers_executions"] = r.PX_SERVERS_EXECUTIONS.String
		mRes["end_of_fetch_count"] = r.END_OF_FETCH_COUNT.String
		mRes["users_executing"] = r.USERS_EXECUTING.String
		mRes["loads"] = r.LOADS.String
		mRes["avg_elapsed"] = r.AVG_ELAPSED.Float64
		mRes["username"] = r.USERNAME.String

		mResults = append(mResults, mRes)
	}

	if len(mResults) == 0 {
		return
	}

	pts = make([]*point.Point, 0)
	for _, v := range mResults {
		jsn, err := json.Marshal(v)
		if err != nil {
			l.Warnf("Marshal json failed: %s, ignore this result", err.Error())
			continue
		}
		fields := map[string]interface{}{
			"status":  "warning",
			"message": string(jsn),
		}
		pts = append(pts, ipt.buildPoint("oracle_log", nil, fields, true))
	}

	return category, pts, nil
}

func (ipt *Input) collectCustomQuery() (category point.Category, pts []*point.Point, err error) {
	category = point.Metric

	pts = make([]*point.Point, 0)
	for _, item := range ipt.Query {
		arr := getCleanCustomQueries(ipt.q(item.SQL))
		if len(arr) == 0 {
			continue
		}

		for _, row := range arr {
			fields := make(map[string]interface{})
			tags := make(map[string]string)

			for _, tgKey := range item.Tags {
				if value, ok := row[tgKey]; ok {
					tags[tgKey] = cast.ToString(value)
					delete(row, tgKey)
				}
			}

			for _, fdKey := range item.Fields {
				if value, ok := row[fdKey]; ok {
					// transform all fields to float64
					fields[fdKey] = cast.ToFloat64(value)
				}
			}

			if len(fields) > 0 {
				pts = append(pts, ipt.buildPoint(item.Metric, tags, fields, false))
			}
		}
	}

	return category, pts, nil
}

func (ipt *Input) buildPoint(name string, tags map[string]string, fields map[string]interface{}, isLogging bool) *point.Point {
	var opts []point.Option

	if isLogging {
		opts = point.DefaultLoggingOptions()
	} else {
		opts = point.DefaultMetricOptions()
	}

	opts = append(opts, point.WithExtraTags(datakit.GlobalElectionTags()))
	opts = append(opts, point.WithTime(time.Now()))

	kvs := point.NewTags(tags)

	// add extended tags
	for k, v := range ipt.mergedTags {
		kvs = kvs.AddTag(k, v)
	}

	for k, v := range fields {
		kvs = kvs.Add(k, v, false, true)
	}

	return point.NewPointV2(name, kvs, opts...)
}

func selectWrapper[T any](ipt *Input, s T, sql string) error {
	now := time.Now()

	err := ipt.db.Select(s, sql)
	if err != nil && (strings.Contains(err.Error(), "ORA-01012") || strings.Contains(err.Error(), "database is closed")) {
		if err := ipt.initDBConnect(); err != nil {
			_ = ipt.db.Close()
		}
	}

	if err != nil {
		l.Errorf("executed sql: %s, cost: %v, err: %v\n", sql, time.Since(now), err)
	} else {
		l.Debugf("executed sql: %s, cost: %v, err: %v\n", sql, time.Since(now), err)
	}

	return err
}

func getOracleTimeString(in string) string {
	t, err := dateparse.ParseAny(in)
	out := ""
	if err != nil {
		l.Warnf("parse date(%s) error: %s", in, err.Error())
		out = strings.ReplaceAll(in, "T", " ")
		out = strings.ReplaceAll(out, "Z", "")
	} else {
		out = t.Format(("2006-01-02 15:04:05"))
	}
	return out
}

func obfuscateSQL(text string) (string, error) {
	reg, err := regexp.Compile(`\n|\s+`) //nolint:gocritic
	if err != nil {
		l.Debugf("Failed to obfuscate, err: %s \n", err.Error())
		return text, err
	}

	sql := strings.TrimSpace(reg.ReplaceAllString(text, " "))

	if out, err := obfuscate.NewObfuscator(nil).Obfuscate("sql", sql); err != nil {
		l.Debugf("Failed to obfuscate, err: %s \n", err.Error())
		return text, err
	} else {
		return out.Query, nil
	}
}
