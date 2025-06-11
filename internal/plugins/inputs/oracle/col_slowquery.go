// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package oracle

import (
	"database/sql"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"github.com/araddon/dateparse"

	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/obfuscate"
)

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
	to_char(sa.LAST_ACTIVE_TIME, 'yyyy-mm-dd hh24:mi:ss') LAST_ACTIVE_TIME_STR,
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
	COMMAND_TYPE                   sql.NullString  `db:"COMMAND_TYPE"`
	OPTIMIZER_MODE                 sql.NullString  `db:"OPTIMIZER_MODE"`
	OPTIMIZER_COST                 sql.NullFloat64 `db:"OPTIMIZER_COST"`
	PARSING_USER_ID                sql.NullString  `db:"PARSING_USER_ID"`
	PARSING_SCHEMA_ID              sql.NullFloat64 `db:"PARSING_SCHEMA_ID"`
	PARSING_SCHEMA_NAME            sql.NullString  `db:"PARSING_SCHEMA_NAME"`
	KEPT_VERSIONS                  sql.NullInt64   `db:"KEPT_VERSIONS"`
	HASH_VALUE                     sql.NullString  `db:"HASH_VALUE"`
	OLD_HASH_VALUE                 sql.NullString  `db:"OLD_HASH_VALUE"`
	PLAN_HASH_VALUE                sql.NullString  `db:"PLAN_HASH_VALUE"`
	MODULE                         sql.NullString  `db:"MODULE"`
	MODULE_HASH                    sql.NullString  `db:"MODULE_HASH"`
	ACTION                         sql.NullString  `db:"ACTION"`
	ACTION_HASH                    sql.NullString  `db:"ACTION_HASH"`
	SERIALIZABLE_ABORTS            sql.NullString  `db:"SERIALIZABLE_ABORTS"`
	OUTLINE_CATEGORY               sql.NullString  `db:"OUTLINE_CATEGORY"`
	CPU_TIME                       sql.NullInt64   `db:"CPU_TIME"`
	ELAPSED_TIME                   sql.NullInt64   `db:"ELAPSED_TIME"`
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
	LAST_ACTIVE_TIME_STR           sql.NullString  `db:"LAST_ACTIVE_TIME_STR"`
	TYPECHECK_MEM                  sql.NullString  `db:"TYPECHECK_MEM"`
	IO_CELL_OFFLOAD_ELIGIBLE_BYTES sql.NullInt64   `db:"IO_CELL_OFFLOAD_ELIGIBLE_BYTES"`
	IO_INTERCONNECT_BYTES          sql.NullString  `db:"IO_INTERCONNECT_BYTES"`
	PHYSICAL_READ_REQUESTS         sql.NullInt64   `db:"PHYSICAL_READ_REQUESTS"`
	PHYSICAL_READ_BYTES            sql.NullInt64   `db:"PHYSICAL_READ_BYTES"`
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
	RUNTIME_MEM                    sql.NullInt64   `db:"RUNTIME_MEM"`
	SORTS                          sql.NullInt64   `db:"SORTS"`
	VERSION_COUNT                  sql.NullString  `db:"VERSION_COUNT"`
	LOADED_VERSIONS                sql.NullString  `db:"LOADED_VERSIONS"`
	OPEN_VERSIONS                  sql.NullString  `db:"OPEN_VERSIONS"`
	USERS_OPENING                  sql.NullString  `db:"USERS_OPENING"`
	FETCHES                        sql.NullString  `db:"FETCHES"`
	EXECUTIONS                     sql.NullInt64   `db:"EXECUTIONS"`
	PX_SERVERS_EXECUTIONS          sql.NullString  `db:"PX_SERVERS_EXECUTIONS"`
	END_OF_FETCH_COUNT             sql.NullString  `db:"END_OF_FETCH_COUNT"`
	USERS_EXECUTING                sql.NullString  `db:"USERS_EXECUTING"`
	LOADS                          sql.NullString  `db:"LOADS"`
	AVG_ELAPSED                    sql.NullFloat64 `db:"AVG_ELAPSED"`
	USERNAME                       sql.NullString  `db:"USERNAME"`
}

const SQLQueryMaxActive = `SELECT to_char(MAX(LAST_ACTIVE_TIME), 'yyyy-mm-dd hh24:mi:ss') MAX_LAST_ACTIVE_TIME FROM V$SQLAREA`

//nolint:stylecheck
type maxQueryRowDB struct {
	MAX_LAST_ACTIVE_TIME sql.NullString `db:"MAX_LAST_ACTIVE_TIME"`
}

func (ipt *Input) collectSlowQuery() {
	if ipt.slowQueryTime == 0 {
		return
	}

	var (
		metricName = "oracle_log"
		rows       = []slowQueryRowDB{}
		start      = time.Now()
		pts        []*point.Point
	)

	if len(ipt.lastActiveTime) == 0 {
		rows := []maxQueryRowDB{}
		if err := selectWrapper(ipt, &rows, SQLQueryMaxActive, getMetricName(metricName, "slow_query_active")); err != nil {
			l.Errorf("failed to collect max query info: %s", err)
			return
		}

		for _, r := range rows {
			ipt.lastActiveTime = r.MAX_LAST_ACTIVE_TIME.String
		}

		l.Debugf("m.lastActiveTime =%s", ipt.lastActiveTime)
		return
	}

	query := fmt.Sprintf(SQLSlow, ipt.slowQueryTime.Microseconds(), ipt.lastActiveTime)

	if err := selectWrapper(ipt, &rows, query, getMetricName(metricName, "slow_query")); err != nil {
		l.Errorf("failed to collect slow query info: %s", err)
		return
	}

	if len(rows) == 0 {
		return
	}

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
			ipt.lastActiveTime = r.LAST_ACTIVE_TIME_STR.String // update saved.
		}

		kvs := ipt.getKVs()

		fullText, err := obfuscateSQL(r.SQL_FULLTEXT.String)
		if err != nil {
			kvs = kvs.AddV2("failed_obfuscate", err.Error(), true)
		}

		kvs = kvs.AddTag("sql_id", r.SQL_ID.String).
			AddTag("module", r.MODULE.String).
			AddTag("command_type", r.COMMAND_TYPE.String).
			AddV2("message", fullText, true).
			AddV2("elapsed_time", r.ELAPSED_TIME.Int64, true).
			AddV2("cpu_time", r.CPU_TIME.Int64, true).
			AddV2("executions", r.EXECUTIONS.Int64, true).
			AddV2("disk_reads", r.DISK_READS.Float64, true).
			AddV2("buffer_gets", r.BUFFER_GETS.Float64, true).
			AddV2("rows_processed", r.ROWS_PROCESSED.Float64, true).
			AddV2("user_io_wait_time", r.USER_IO_WAIT_TIME.Float64, true).
			AddV2("concurrency_wait_time", r.CONCURRENCY_WAIT_TIME.Float64, true).
			AddV2("application_wait_time", r.APPLICATION_WAIT_TIME.Float64, true).
			AddV2("cluster_wait_time", r.CLUSTER_WAIT_TIME.Float64, true).
			AddV2("plan_hash_value", r.PLAN_HASH_VALUE.String, true).
			AddV2("parse_calls", r.PARSE_CALLS.Float64, true).
			AddV2("sorts", r.SORTS.Int64, true).
			AddV2("parsing_schema_name", r.PARSING_SCHEMA_NAME.String, true).
			AddV2("action", r.ACTION.String, true).
			AddV2("last_active_time", r.LAST_ACTIVE_TIME.String, true).
			AddV2("username", r.USERNAME.String, true).
			AddV2("avg_elapsed", r.AVG_ELAPSED.Float64, true).
			AddV2("status", "warning", true) // add logging basic status and message field

		pts = append(pts, point.NewPointV2(metricName, kvs, ipt.getKVsOpts(point.Logging)...))
	}

	if err := ipt.feeder.FeedV2(point.Logging,
		pts,
		dkio.WithCollectCost(time.Since(start)),
		dkio.WithElection(ipt.Election),
		dkio.WithInputName(loggingFeedName)); err != nil {
		l.Warnf("feeder.FeedV2: %s, ignored", err)
	}
}

func (ipt *Input) getKVs() point.KVs {
	var kvs point.KVs

	// add extended tags
	for k, v := range ipt.mergedTags {
		kvs = kvs.AddTag(k, v)
	}

	return kvs
}

func getMetricNames(name string) (string, string) {
	names := strings.SplitN(name, ":", 2)
	metricName := ""
	sqlName := ""
	if len(names) == 1 {
		metricName = names[0]
		sqlName = names[0]
	} else if len(names) == 2 {
		metricName = names[0]
		sqlName = names[1]
	}

	return metricName, sqlName
}

func getMetricName(metricName, sqlName string) string {
	if sqlName == "" {
		return metricName
	} else {
		return metricName + ":" + sqlName
	}
}

var reg = regexp.MustCompile(`\n|\s+`) //nolint:gocritic

func obfuscateSQL(text string) (sql string, err error) {
	defer func() {
		sql = strings.TrimSpace(reg.ReplaceAllString(sql, " "))
	}()

	if out, err := obfuscate.NewObfuscator(nil).Obfuscate("sql", text); err != nil {
		return text, err
	} else {
		return out.Query, nil
	}
}
