// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package coracle

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/obfuscate"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/oracle/collect/ccommon"
)

type slowQueryLogging struct {
	x              collectParameters
	lastActiveTime string
}

var _ ccommon.DBMetricsCollector = (*slowQueryLogging)(nil)

func newSlowQueryLogging(opts ...collectOption) *slowQueryLogging {
	m := &slowQueryLogging{}

	for _, opt := range opts {
		if opt != nil {
			opt(&m.x)
		}
	}

	return m
}

func (m *slowQueryLogging) Collect() (*point.Point, error) {
	l.Debug("Collect entry")

	tf, err := m.slowQuery()
	if err != nil {
		return nil, err
	}

	if tf == nil || tf.IsEmpty() {
		return nil, nil
	}

	opt := &ccommon.BuildPointOpt{
		TF:         tf,
		MetricName: m.x.MetricName,
		Tags:       m.x.Ipt.tags,
		Host:       m.x.Ipt.host,
	}
	return ccommon.BuildPointLogging(l, opt), nil
}

// SLOW_QUERY is the get slow query info SQL query for Oracle.
//
// excluded fields SQL_TEXT, OPTIMIZER_ENV, ADDRESS, LAST_ACTIVE_CHILD_ADDRESS, BIND_DATA
//
//nolint:stylecheck
const SLOW_QUERY = `SELECT 
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
  	EXECUTIONS > 0 AND LAST_ACTIVE_TIME > to_date('%s','yyyy-mm-dd hh24:mi:ss')
  ORDER BY LAST_ACTIVE_TIME DESC`

//nolint:stylecheck
type slowQueryRowDB struct {
	FIRST_LOAD_TIME                sql.NullString `db:"FIRST_LOAD_TIME"`
	INVALIDATIONS                  sql.NullString `db:"INVALIDATIONS"`
	PARSE_CALLS                    sql.NullString `db:"PARSE_CALLS"`
	DISK_READS                     sql.NullString `db:"DISK_READS"`
	DIRECT_WRITES                  sql.NullString `db:"DIRECT_WRITES"`
	BUFFER_GETS                    sql.NullString `db:"BUFFER_GETS"`
	APPLICATION_WAIT_TIME          sql.NullString `db:"APPLICATION_WAIT_TIME"`
	CONCURRENCY_WAIT_TIME          sql.NullString `db:"CONCURRENCY_WAIT_TIME"`
	CLUSTER_WAIT_TIME              sql.NullString `db:"CLUSTER_WAIT_TIME"`
	USER_IO_WAIT_TIME              sql.NullString `db:"USER_IO_WAIT_TIME"`
	PLSQL_EXEC_TIME                sql.NullString `db:"PLSQL_EXEC_TIME"`
	JAVA_EXEC_TIME                 sql.NullString `db:"JAVA_EXEC_TIME"`
	ROWS_PROCESSED                 sql.NullString `db:"ROWS_PROCESSED"`
	COMMAND_TYPE                   sql.NullString `db:"COMMAND_TYPE"`
	OPTIMIZER_MODE                 sql.NullString `db:"OPTIMIZER_MODE"`
	OPTIMIZER_COST                 sql.NullString `db:"OPTIMIZER_COST"`
	PARSING_USER_ID                sql.NullString `db:"PARSING_USER_ID"`
	PARSING_SCHEMA_ID              sql.NullString `db:"PARSING_SCHEMA_ID"`
	PARSING_SCHEMA_NAME            sql.NullString `db:"PARSING_SCHEMA_NAME"`
	KEPT_VERSIONS                  sql.NullString `db:"KEPT_VERSIONS"`
	HASH_VALUE                     sql.NullString `db:"HASH_VALUE"`
	OLD_HASH_VALUE                 sql.NullString `db:"OLD_HASH_VALUE"`
	PLAN_HASH_VALUE                sql.NullString `db:"PLAN_HASH_VALUE"`
	MODULE                         sql.NullString `db:"MODULE"`
	MODULE_HASH                    sql.NullString `db:"MODULE_HASH"`
	ACTION                         sql.NullString `db:"ACTION"`
	ACTION_HASH                    sql.NullString `db:"ACTION_HASH"`
	SERIALIZABLE_ABORTS            sql.NullString `db:"SERIALIZABLE_ABORTS"`
	OUTLINE_CATEGORY               sql.NullString `db:"OUTLINE_CATEGORY"`
	CPU_TIME                       sql.NullString `db:"CPU_TIME"`
	ELAPSED_TIME                   sql.NullString `db:"ELAPSED_TIME"`
	OUTLINE_SID                    sql.NullString `db:"OUTLINE_SID"`
	REMOTE                         sql.NullString `db:"REMOTE"`
	OBJECT_STATUS                  sql.NullString `db:"OBJECT_STATUS"`
	LITERAL_HASH_VALUE             sql.NullString `db:"LITERAL_HASH_VALUE"`
	LAST_LOAD_TIME                 sql.NullString `db:"LAST_LOAD_TIME"`
	IS_OBSOLETE                    sql.NullString `db:"IS_OBSOLETE"`
	IS_BIND_SENSITIVE              sql.NullString `db:"IS_BIND_SENSITIVE"`
	IS_BIND_AWARE                  sql.NullString `db:"IS_BIND_AWARE"`
	CHILD_LATCH                    sql.NullString `db:"CHILD_LATCH"`
	SQL_PROFILE                    sql.NullString `db:"SQL_PROFILE"`
	SQL_PATCH                      sql.NullString `db:"SQL_PATCH"`
	SQL_PLAN_BASELINE              sql.NullString `db:"SQL_PLAN_BASELINE"`
	PROGRAM_ID                     sql.NullString `db:"PROGRAM_ID"`
	PROGRAM_LINE                   sql.NullString `db:"PROGRAM_LINE#"`
	EXACT_MATCHING_SIGNATURE       sql.NullString `db:"EXACT_MATCHING_SIGNATURE"`
	FORCE_MATCHING_SIGNATURE       sql.NullString `db:"FORCE_MATCHING_SIGNATURE"`
	LAST_ACTIVE_TIME               sql.NullString `db:"LAST_ACTIVE_TIME"`
	TYPECHECK_MEM                  sql.NullString `db:"TYPECHECK_MEM"`
	IO_CELL_OFFLOAD_ELIGIBLE_BYTES sql.NullString `db:"IO_CELL_OFFLOAD_ELIGIBLE_BYTES"`
	IO_INTERCONNECT_BYTES          sql.NullString `db:"IO_INTERCONNECT_BYTES"`
	PHYSICAL_READ_REQUESTS         sql.NullString `db:"PHYSICAL_READ_REQUESTS"`
	PHYSICAL_READ_BYTES            sql.NullString `db:"PHYSICAL_READ_BYTES"`
	PHYSICAL_WRITE_REQUESTS        sql.NullString `db:"PHYSICAL_WRITE_REQUESTS"`
	PHYSICAL_WRITE_BYTES           sql.NullString `db:"PHYSICAL_WRITE_BYTES"`
	OPTIMIZED_PHY_READ_REQUESTS    sql.NullString `db:"OPTIMIZED_PHY_READ_REQUESTS"`
	LOCKED_TOTAL                   sql.NullString `db:"LOCKED_TOTAL"`
	PINNED_TOTAL                   sql.NullString `db:"PINNED_TOTAL"`
	IO_CELL_UNCOMPRESSED_BYTES     sql.NullString `db:"IO_CELL_UNCOMPRESSED_BYTES"`
	IO_CELL_OFFLOAD_RETURNED_BYTES sql.NullString `db:"IO_CELL_OFFLOAD_RETURNED_BYTES"`
	SQL_FULLTEXT                   sql.NullString `db:"SQL_FULLTEXT"`
	SQL_ID                         sql.NullString `db:"SQL_ID"`
	SHARABLE_MEM                   sql.NullString `db:"SHARABLE_MEM"`
	PERSISTENT_MEM                 sql.NullString `db:"PERSISTENT_MEM"`
	RUNTIME_MEM                    sql.NullString `db:"RUNTIME_MEM"`
	SORTS                          sql.NullString `db:"SORTS"`
	VERSION_COUNT                  sql.NullString `db:"VERSION_COUNT"`
	LOADED_VERSIONS                sql.NullString `db:"LOADED_VERSIONS"`
	OPEN_VERSIONS                  sql.NullString `db:"OPEN_VERSIONS"`
	USERS_OPENING                  sql.NullString `db:"USERS_OPENING"`
	FETCHES                        sql.NullString `db:"FETCHES"`
	EXECUTIONS                     sql.NullString `db:"EXECUTIONS"`
	PX_SERVERS_EXECUTIONS          sql.NullString `db:"PX_SERVERS_EXECUTIONS"`
	END_OF_FETCH_COUNT             sql.NullString `db:"END_OF_FETCH_COUNT"`
	USERS_EXECUTING                sql.NullString `db:"USERS_EXECUTING"`
	LOADS                          sql.NullString `db:"LOADS"`
	AVG_ELAPSED                    sql.NullString `db:"AVG_ELAPSED"`
	USERNAME                       sql.NullString `db:"USERNAME"`
}

// SQL_QUERY_MAX_ACTIVE get the max LAST_ACTIVE_TIME in the table.
//
//nolint:stylecheck
const SQL_QUERY_MAX_ACTIVE = `SELECT MAX(LAST_ACTIVE_TIME) FROM V$SQLAREA`

//nolint:stylecheck
type maxQueryRowDB struct {
	MAX_LAST_ACTIVE_TIME sql.NullString `db:"MAX(LAST_ACTIVE_TIME)"`
}

func (m *slowQueryLogging) slowQuery() (*ccommon.TagField, error) {
	if len(m.lastActiveTime) == 0 {
		rows := []maxQueryRowDB{}
		err := selectWrapper(m.x.Ipt, &rows, SQL_QUERY_MAX_ACTIVE)
		if err != nil {
			return nil, fmt.Errorf("failed to collect max query info: %w", err)
		}

		for _, r := range rows {
			m.lastActiveTime = getOracleTimeString(r.MAX_LAST_ACTIVE_TIME.String)
		}

		fmt.Println("m.lastActiveTime =", m.lastActiveTime)

		return nil, nil
	}

	tf := ccommon.NewTagField()
	rows := []slowQueryRowDB{}

	query := fmt.Sprintf(SLOW_QUERY, m.lastActiveTime)

	err := selectWrapper(m.x.Ipt, &rows, query)
	if err != nil {
		return nil, fmt.Errorf("failed to collect slow query info: %w", err)
	}

	if len(rows) == 0 {
		return nil, nil
	}

	tf.AddField("status", "warning", nil)

	mResults := make([]map[string]string, 0)

	for _, r := range rows {
		gotlastActiveTime, err := time.Parse("2006-01-02T15:04:05Z", r.LAST_ACTIVE_TIME.String)
		if err != nil {
			l.Warnf(err.Error())
			continue
		}
		savedLastActiveTime, err := time.Parse("2006-01-02 15:04:05", m.lastActiveTime)
		if err != nil {
			l.Warnf(err.Error())
			continue
		}
		if gotlastActiveTime.After(savedLastActiveTime) {
			m.lastActiveTime = getOracleTimeString(r.LAST_ACTIVE_TIME.String) // update saved.
		}

		avgElapsed := r.AVG_ELAPSED.String
		du, err := time.ParseDuration(avgElapsed + "us")
		if err != nil {
			l.Warnf(err.Error())
			continue
		}
		if du < m.x.Ipt.SlowQueryTime {
			l.Debugf("%s < slow query time, skipped", du.String())
			continue
		}

		mRes := make(map[string]string, 78)

		mRes["first_load_time"] = r.FIRST_LOAD_TIME.String
		mRes["invalidations"] = r.INVALIDATIONS.String
		mRes["parse_calls"] = r.PARSE_CALLS.String
		mRes["disk_reads"] = r.DISK_READS.String
		mRes["direct_writes"] = r.DIRECT_WRITES.String
		mRes["buffer_gets"] = r.BUFFER_GETS.String
		mRes["application_wait_time"] = r.APPLICATION_WAIT_TIME.String
		mRes["concurrency_wait_time"] = r.CONCURRENCY_WAIT_TIME.String
		mRes["cluster_wait_time"] = r.CLUSTER_WAIT_TIME.String
		mRes["user_io_wait_time"] = r.USER_IO_WAIT_TIME.String
		mRes["plsql_exec_time"] = r.PLSQL_EXEC_TIME.String
		mRes["java_exec_time"] = r.JAVA_EXEC_TIME.String
		mRes["rows_processed"] = r.ROWS_PROCESSED.String
		mRes["command_type"] = r.COMMAND_TYPE.String
		mRes["optimizer_mode"] = r.OPTIMIZER_MODE.String
		mRes["optimizer_cost"] = r.OPTIMIZER_COST.String
		mRes["parsing_user_id"] = r.PARSING_USER_ID.String
		mRes["parsing_schema_id"] = r.PARSING_SCHEMA_ID.String
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
		mRes["cpu_time"] = r.CPU_TIME.String
		mRes["elapsed_time"] = r.ELAPSED_TIME.String
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
		mRes["avg_elapsed"] = r.AVG_ELAPSED.String
		mRes["username"] = r.USERNAME.String

		mResults = append(mResults, mRes)
	}

	if len(mResults) == 0 {
		return nil, nil
	}

	jsn, err := json.Marshal(mResults)
	if err != nil {
		return nil, err
	}

	tf.AddField("message", string(jsn), nil)

	return tf, nil
}

func getOracleTimeString(in string) string {
	out := strings.ReplaceAll(in, "T", " ")
	out = strings.ReplaceAll(out, "Z", "")
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
