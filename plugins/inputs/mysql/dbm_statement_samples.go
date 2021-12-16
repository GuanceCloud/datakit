package mysql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

type dbmSample struct {
	Enabled bool `toml:"enabled"`
}

type dbmSampleMeasurement struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
	ts     time.Time
}

type eventStrategy struct {
	table    string
	interval int
}

type dbmSampleCache struct {
	checkPoint        int64
	version           mysqlVersion
	globalStatusTable string
	strategy          eventStrategy
	explainCache      cacheLimit
}

type eventRow struct {
	currentSchema       sql.NullString
	sqlText             sql.NullString
	digest              sql.NullString
	digestText          sql.NullString
	timerStart          sql.NullInt64
	timerEndTimeS       sql.NullString
	timerWaitNs         sql.NullString
	lockTimeNs          sql.NullString
	rowsAffected        sql.NullInt64
	rowsSent            sql.NullInt64
	rowsExamined        sql.NullInt64
	selectFullJoin      sql.NullInt64
	selectFullRangeJoin sql.NullInt64
	selectRange         sql.NullInt64
	selectRangeCheck    sql.NullInt64
	selectScan          sql.NullInt64
	sortMergePasses     sql.NullInt64
	sortRange           sql.NullInt64
	sortRows            sql.NullInt64
	sortScan            sql.NullInt64
	noIndexUsed         sql.NullInt64
	noGoodIndexUsed     sql.NullInt64
	processlistUser     sql.NullString
	processlistHost     sql.NullString
	processlistDB       sql.NullString
}

type planObj struct {
	timestamp           float64 // millisecond
	duration            float64 // nanosecond
	networkClientIP     string
	currentSchema       string
	planDefinition      string
	planSignature       string
	querySignature      string
	resourceHash        string
	statement           string
	digestText          string
	queryTruncated      string
	digest              string
	lockTimeNs          int64
	noGoodIndexUsed     int64
	noIndexUsed         int64
	processlistDB       string
	processlistUser     string
	rowsAffected        int64
	rowsExamined        int64
	rowsSent            int64
	selectFullJoin      int64
	selectFullRangeJoin int64
	selectRange         int64
	selectRangeCheck    int64
	selectScan          int64
	sortMergePasses     int64
	sortRange           int64
	sortRows            int64
	sortScan            int64
}

func (m *dbmSampleMeasurement) LineProto() (*io.Point, error) {
	return io.MakePoint(m.name, m.tags, m.fields, m.ts)
}

func (m *dbmSampleMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Desc: "选取部分执行耗时较高的 SQL 语句，获取其执行计划，并采集实际执行过程中的各种性能指标。",
		Name: "mysql_dbm_sample",
		Fields: map[string]interface{}{
			"timestamp": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The timestamp(millisecond) when then the event ends.",
			},
			"duration": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Value in nanoseconds of the event's duration.",
			},
			"lock_time_ns": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Time in nanoseconds spent waiting for locks. ",
			},
			"no_good_index_used": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.Int,
				Desc:     "0 if a good index was found for the statement, 1 if no good index was found.",
			},
			"no_index_used": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.Int,
				Desc:     "0 if the statement performed a table scan with an index, 1 if without an index.",
			},
			"rows_affected": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of rows the statement affected.",
			},
			"rows_examined": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of rows read during the statement's execution.",
			},
			"rows_sent": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of rows returned. ",
			},
			"select_full_join": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of joins performed by the statement which did not use an index.",
			},
			"select_full_range_join": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of joins performed by the statement which used a range search of the int first table. ",
			},
			"select_range": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of joins performed by the statement which used a range of the first table. ",
			},
			"select_range_check": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of joins without keys performed by the statement that check for key usage after int each row. ",
			},
			"select_scan": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of joins performed by the statement which used a full scan of the first table.",
			},
			"sort_merge_passes": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of merge passes by the sort algorithm performed by the statement. ",
			},
			"sort_range": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of sorts performed by the statement which used a range.",
			},
			"sort_rows": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of rows sorted by the statement. ",
			},
			"sort_scan": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of sorts performed by the statement which used a full table scan.",
			},
			"timer_wait_ns": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.Int,
				Desc:     "Value in nanoseconds of the event's duration ",
			},
			"message": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.String,
				Unit:     inputs.UnknownType,
				Desc:     "The text of the normalized statement digest.",
			},
		},
		Tags: map[string]interface{}{
			"current_schema":    &inputs.TagInfo{Desc: "The name of the current schema."},
			"plan_definition":   &inputs.TagInfo{Desc: "The plan definition of JSON format."},
			"plan_signature":    &inputs.TagInfo{Desc: "The hash value computed from plan definition."},
			"query_signature":   &inputs.TagInfo{Desc: "The hash value computed from digest_text."},
			"resource_hash":     &inputs.TagInfo{Desc: "The hash value computed from sql text."},
			"query_truncated":   &inputs.TagInfo{Desc: "It indicates whether the query is truncated."},
			"network_client_ip": &inputs.TagInfo{Desc: "The ip address of the client"},
			"digest": &inputs.TagInfo{
				Desc: "The digest hash value computed from the original normalized statement. ",
			},
			"processlist_db":   &inputs.TagInfo{Desc: "The name of the database."},
			"processlist_user": &inputs.TagInfo{Desc: "The user name of the client."},
		},
	}
}

var eventsStatementsCollectionInterval = map[string]int{
	"events_statements_history_long": 10,
	"events_statements_history":      10,
	"events_statements_current":      1,
}

// get the table from which samples should be collected.
func getSampleCollectionStrategy(i *Input) (eventStrategy, error) {
	var strategy eventStrategy
	if len(i.dbmSampleCache.strategy.table) > 0 {
		return i.dbmSampleCache.strategy, nil
	}

	var eventsStatementsTable string

	enabledSQL := `SELECT name
	FROM performance_schema.setup_consumers
	WHERE enabled = 'YES' AND name LIKE 'events_statements_%'`
	enabledConsumers := getCleanEnabledPerformanceSchemaConsumers(i.q(enabledSQL))

	if len(enabledConsumers) < 3 {
		err := enablePerformanceSchemaConsumers(i)
		if err != nil {
			l.Warn(err)
		} else {
			enabledConsumers = getCleanEnabledPerformanceSchemaConsumers(i.q(enabledSQL))
		}
	}

	if len(enabledConsumers) == 0 {
		return strategy, errors.New("no events_statements consumer")
	}

	l.Debugf("enabled performance_schema statements consumers: %s", enabledConsumers)

	tables := []string{
		"events_statements_history_long",
		"events_statements_current",
		"events_statements_history",
	}

	for _, table := range tables {
		if !isListHasStr(enabledConsumers, table) {
			continue
		}

		rows, err := getNewEventsStatements(i, table, 1)
		if err != nil || (len(rows) == 0) {
			continue
		}

		eventsStatementsTable = table
		break
	}

	if len(eventsStatementsTable) == 0 {
		return strategy, fmt.Errorf("all enabled events_statements_consumers %v are empty", enabledConsumers)
	}

	currentStrategy := eventStrategy{
		table:    eventsStatementsTable,
		interval: eventsStatementsCollectionInterval[eventsStatementsTable],
	}

	i.dbmSampleCache.strategy = currentStrategy

	return currentStrategy, nil
}

// enable consumers at runtime.
func enablePerformanceSchemaConsumers(i *Input) error {
	sqlStr := "CALL datakit.enable_events_statements_consumers()"
	if _, err := i.db.Exec(sqlStr); err != nil {
		return err
	}
	return nil
}

// collect events.
func getNewEventsStatements(i *Input, eventTable string, rowLimit int) ([]eventRow, error) {
	ctx := context.Background()
	var rows []eventRow
	conn, err := i.db.Conn(ctx)
	if err != nil {
		return rows, err
	}
	defer conn.Close() //nolint:errcheck

	// silence warnings
	if _, err := conn.ExecContext(ctx, "SET @@SESSION.sql_notes = 0"); err != nil {
		return rows, err
	}

	// drop temp table
	dropTempTableQuerySQL := "DROP TEMPORARY TABLE IF EXISTS datakit.temp_events"
	if _, err := conn.ExecContext(ctx, dropTempTableQuerySQL); err != nil {
		return rows, err
	}

	// create temp table
	createTempTableSQLTemplate := `
	CREATE TEMPORARY TABLE datakit.temp_events SELECT
        current_schema,
        sql_text,
        digest,
        digest_text,
        timer_start,
        timer_end,
        timer_wait,
        lock_time,
        rows_affected,
        rows_sent,
        rows_examined,
        select_full_join,
        select_full_range_join,
        select_range,
        select_range_check,
        select_scan,
        sort_merge_passes,
        sort_range,
        sort_rows,
        sort_scan,
        no_index_used,
        no_good_index_used,
        event_name,
        thread_id
     FROM performance_schema.%v
        WHERE sql_text IS NOT NULL
        AND event_name like 'statement/%%'
        AND digest_text is NOT NULL
        AND digest_text NOT LIKE 'EXPLAIN %%'
        AND timer_start > %v
    LIMIT %v
	`

	createTempTableSQL := fmt.Sprintf(createTempTableSQLTemplate, eventTable, i.dbmSampleCache.checkPoint, rowLimit)

	if _, err := conn.ExecContext(ctx, createTempTableSQL); err != nil {
		return rows, err
	}

	var subSelect string
	if i.dbmSampleCache.version.versionCompatible([]int{8, 0, 0}) {
		subSelect = "(SELECT *,row_number() over (partition by digest order by timer_wait desc) as row_num FROM %s)"
	} else {
		if _, err := conn.ExecContext(ctx, "set @row_num = 0"); err != nil {
			return rows, err
		}
		if _, err := conn.ExecContext(ctx, "set @current_digest = ''"); err != nil {
			return rows, err
		}
		subSelect = `(SELECT *,
			@row_num := IF(@current_digest = digest, @row_num + 1, 1) AS row_num,
			@current_digest := digest
			FROM %s ORDER BY digest, timer_wait)`
	}

	startupSQL := "(SELECT UNIX_TIMESTAMP()-VARIABLE_VALUE FROM %s WHERE VARIABLE_NAME='UPTIME')"
	startupTimeSubquery := fmt.Sprintf(startupSQL, i.dbmSampleCache.globalStatusTable)
	if _, err := conn.ExecContext(ctx, fmt.Sprintf("set @startup_time_s=%s", startupTimeSubquery)); err != nil {
		return rows, err
	}

	eventsStatementsQuery := `
		SELECT
		current_schema,
		sql_text,
		digest,
		digest_text,
		timer_start,
		@startup_time_s+timer_end*1e-12 as timer_end_time_s,
		timer_wait / 1000 AS timer_wait_ns,
		lock_time / 1000 AS lock_time_ns,
		rows_affected,
		rows_sent,
		rows_examined,
		select_full_join,
		select_full_range_join,
		select_range,
		select_range_check,
		select_scan,
		sort_merge_passes,
		sort_range,
		sort_rows,
		sort_scan,
		no_index_used,
		no_good_index_used,
		processlist_user,
		processlist_host,
		processlist_db
	FROM %v as E
	LEFT JOIN performance_schema.threads as T
		ON E.thread_id = T.thread_id
	WHERE sql_text IS NOT NULL
		AND timer_start > %v
		AND row_num = 1
	ORDER BY timer_wait DESC
	LIMIT %v
`
	subSelectSQL := fmt.Sprintf(subSelect, "datakit.temp_events")
	eventsStatementsQuerySQL := fmt.Sprintf(eventsStatementsQuery, subSelectSQL, i.dbmSampleCache.checkPoint, rowLimit)
	rawRows, err := conn.QueryContext(ctx, eventsStatementsQuerySQL)
	if err != nil {
		return rows, err
	}

	if rawRows.Err() != nil {
		return rows, rawRows.Err()
	}

	defer rawRows.Close() // nolint: errcheck

	for rawRows.Next() {
		row := eventRow{}
		if err := rawRows.Scan(
			&row.currentSchema,
			&row.sqlText,
			&row.digest,
			&row.digestText,
			&row.timerStart,
			&row.timerEndTimeS,
			&row.timerWaitNs,
			&row.lockTimeNs,
			&row.rowsAffected,
			&row.rowsSent,
			&row.rowsExamined,
			&row.selectFullJoin,
			&row.selectFullRangeJoin,
			&row.selectRange,
			&row.selectRangeCheck,
			&row.selectScan,
			&row.sortMergePasses,
			&row.sortRange,
			&row.sortRows,
			&row.sortScan,
			&row.noIndexUsed,
			&row.noGoodIndexUsed,
			&row.processlistUser,
			&row.processlistHost,
			&row.processlistDB); err != nil {
			l.Warn(err)
			continue
		}
		rows = append(rows, row)
	}

	if _, err := conn.ExecContext(ctx, dropTempTableQuerySQL); err != nil {
		return rows, err
	}

	return rows, nil
}

func filterValidStatementRows(i *Input, rows []eventRow) []eventRow {
	var filterRows []eventRow

	for _, row := range rows {
		if row.sqlText.Valid && len(row.sqlText.String) > 0 {
			filterRows = append(filterRows, row)
		}

		if row.timerStart.Valid {
			if row.timerStart.Int64 > i.dbmSampleCache.checkPoint {
				i.dbmSampleCache.checkPoint = row.timerStart.Int64
			}
		}
	}

	return filterRows
}

func collectPlanForStatements(i *Input, rows []eventRow) []planObj {
	var plans []planObj
	for _, row := range rows {
		plan, err := collectPlanForStatement(i, row)
		if err != nil {
			l.Warnf("collect plan error: %s", err.Error())
			continue
		}
		if len(plan.planDefinition) == 0 {
			continue
		}
		plans = append(plans, plan)
	}

	return plans
}

func collectPlanForStatement(i *Input, row eventRow) (planObj, error) {
	var plan planObj
	obfuscatedStatement := obfuscateSQL(row.sqlText.String)
	obfuscatedDigestText := obfuscateSQL(row.digestText.String)

	apmResourceHash := computeSQLSignature(obfuscatedStatement)
	querySignature := computeSQLSignature(obfuscatedDigestText)

	queryCacheKey := getRowKey(row.currentSchema.String, querySignature)

	if !checkLimitRate(i, queryCacheKey) {
		l.Debugf("ingore check: %s", queryCacheKey)
		return plan, nil
	}

	truncated := isTruncated(row.sqlText.String)

	// ignore truncated sql
	if truncated {
		return plan, nil
	}

	planStr, err := explainStatement(i, row.sqlText.String, row.currentSchema.String, obfuscatedStatement)
	if err != nil {
		return plan, err
	}

	if len(planStr) > 0 {
		normalizedPlan := obfuscatePlan(planStr)
		planSignature := computeSQLSignature(normalizedPlan)
		plan = planObj{
			planDefinition: normalizedPlan,
			planSignature:  planSignature,
			querySignature: querySignature,
			digestText:     obfuscatedDigestText,
			resourceHash:   apmResourceHash,
			statement:      obfuscatedStatement,
		}

		if truncated {
			plan.queryTruncated = "truncated"
		} else {
			plan.queryTruncated = "not_truncated"
		}

		if row.timerEndTimeS.Valid {
			if timerEndTimeS, err := strconv.ParseFloat(row.timerEndTimeS.String, 64); err == nil {
				plan.timestamp = timerEndTimeS * 1000
			}
		}
		if row.timerWaitNs.Valid {
			if timerWaitNs, err := strconv.ParseFloat(row.timerWaitNs.String, 64); err == nil {
				plan.duration = timerWaitNs
			}
		}

		if row.processlistHost.Valid {
			plan.networkClientIP = row.processlistHost.String
		}

		if row.currentSchema.Valid {
			plan.currentSchema = row.currentSchema.String
		}

		if row.digest.Valid {
			plan.digest = row.digest.String
		}

		if row.lockTimeNs.Valid {
			if lockTimeNs, err := strconv.Atoi(row.lockTimeNs.String); err == nil {
				plan.lockTimeNs = int64(lockTimeNs)
			}
		}
		if row.noGoodIndexUsed.Valid {
			plan.noGoodIndexUsed = row.noGoodIndexUsed.Int64
		}
		if row.noIndexUsed.Valid {
			plan.noIndexUsed = row.noIndexUsed.Int64
		}
		if row.processlistDB.Valid {
			plan.processlistDB = row.processlistDB.String
		}
		if row.processlistUser.Valid {
			plan.processlistUser = row.processlistUser.String
		}
		if row.rowsAffected.Valid {
			plan.rowsAffected = row.rowsAffected.Int64
		}
		if row.rowsExamined.Valid {
			plan.rowsExamined = row.rowsExamined.Int64
		}
		if row.rowsSent.Valid {
			plan.rowsSent = row.rowsSent.Int64
		}
		if row.selectFullJoin.Valid {
			plan.selectFullJoin = row.selectFullJoin.Int64
		}
		if row.selectFullRangeJoin.Valid {
			plan.selectFullRangeJoin = row.selectFullRangeJoin.Int64
		}
		if row.selectRange.Valid {
			plan.selectRange = row.selectRange.Int64
		}
		if row.selectRangeCheck.Valid {
			plan.selectRangeCheck = row.selectRangeCheck.Int64
		}
		if row.selectScan.Valid {
			plan.selectScan = row.selectScan.Int64
		}
		if row.sortMergePasses.Valid {
			plan.sortMergePasses = row.sortMergePasses.Int64
		}
		if row.sortRange.Valid {
			plan.sortRange = row.sortRange.Int64
		}
		if row.sortRows.Valid {
			plan.sortRows = row.sortRows.Int64
		}
		if row.sortScan.Valid {
			plan.sortScan = row.sortScan.Int64
		}
	}

	return plan, nil
}

// limit the explain rate of the same statement.
func checkLimitRate(i *Input, key string) bool {
	return i.dbmSampleCache.explainCache.Acquire(key)
}

func isTruncated(statement string) bool {
	return strings.HasSuffix(statement, "...")
}

func explainStatement(i *Input, statement string, schema string, obfuscatedStatement string) (string, error) {
	var plan string
	ctx := context.Background()
	conn, err := i.db.Conn(ctx)
	if err != nil {
		return plan, err
	}
	defer conn.Close() //nolint:errcheck
	startTime := time.Now()
	strategyCacheKey := fmt.Sprintf("explain_strategy:%s", schema)

	l.Debugf("explaining statement. schema=%s, statement='%s'", schema, statement)

	if !canExplain(obfuscatedStatement) {
		l.Debugf("ignore explain statement: %s", obfuscatedStatement)
		return plan, nil
	}

	// TODO cached strategy
	if len(schema) > 0 {
		if _, err := conn.ExecContext(ctx, fmt.Sprintf("USE %s", schema)); err != nil {
			return plan, err
		}
	}

	strategies := []string{"PROCEDURE", "FQ_PROCEDURE", "STATEMENT"}

	for _, strategy := range strategies {
		if len(schema) == 0 && (strategy == "PROCEDURE") {
			l.Debug("skipping procedure strategy: no default schema")
			continue
		}
		if strategy == "PROCEDURE" {
			plan, err = runExplainProcedure(conn, statement)
		}
		if strategy == "FQ_PROCEDURE" {
			plan, err = runFullyQualifiedExplainProcedure(conn, statement)
		}
		if strategy == "STATEMENT" {
			plan, err = runExplain(conn, statement)
		}
		if err != nil {
			l.Debug(err)
			continue
		}
		if len(plan) > 0 {
			return plan, nil
		}
	}
	fmt.Println(startTime, strategyCacheKey)
	return plan, nil
}

func runExplainProcedure(conn *sql.Conn, statement string) (string, error) {
	ctx := context.Background()
	row := conn.QueryRowContext(ctx, fmt.Sprintf("CALL explain_statement('%s')", statement))
	var plan string
	if row.Err() != nil {
		return plan, row.Err()
	}
	if err := row.Scan(&plan); err != nil {
		return plan, err
	}

	return plan, nil
}

func runFullyQualifiedExplainProcedure(conn *sql.Conn, statement string) (string, error) {
	ctx := context.Background()
	row := conn.QueryRowContext(ctx, fmt.Sprintf("CALL datakit.explain_statement('%s')", statement))
	var plan string
	if row.Err() != nil {
		return plan, row.Err()
	}
	if err := row.Scan(&plan); err != nil {
		return plan, err
	}

	return plan, nil
}

func runExplain(conn *sql.Conn, statement string) (string, error) {
	ctx := context.Background()
	row := conn.QueryRowContext(ctx, fmt.Sprintf("EXPLAIN FORMAT=json %s", statement))
	var plan string
	if row.Err() != nil {
		return plan, row.Err()
	}
	if err := row.Scan(&plan); err != nil {
		return plan, err
	}

	return plan, nil
}
