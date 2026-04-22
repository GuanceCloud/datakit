// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package mysql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"github.com/cespare/xxhash/v2"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/util"

	"github.com/DataDog/datadog-agent/pkg/obfuscate"

	"github.com/hashicorp/golang-lru/v2/expirable"
)

const (
	dbmExecPlanObjectName   = "db_exec_plan"
	defaultExplainCacheSize = 10000
)

type dbmSample struct {
	Enabled               bool             `toml:"enabled"`
	Interval              datakit.Duration `toml:"interval"`
	ExplainCacheTTL       datakit.Duration `toml:"explain_cache_ttl"`
	PlanCacheTTL          datakit.Duration `toml:"plan_cache_ttl"`
	EventsStatementsLimit int              `toml:"events_statements_limit"`
}

type dbmSampleMeasurement struct {
	name     string
	tags     map[string]string
	fields   map[string]interface{}
	election bool
}

type dbmSampleCache struct {
	globalStatusTable     string
	eventsStatementsTable string
	explainCache          *util.CacheLimit
}

type eventRow struct {
	currentSchema       sql.NullString
	sqlText             sql.NullString
	digest              sql.NullString
	digestText          sql.NullString
	endEventID          sql.NullInt64
	timerStart          sql.NullInt64
	uptime              sql.NullString // @uptime (seconds since server start)
	now                 sql.NullInt64  // unix_timestamp()
	timerEnd            sql.NullInt64  // picoseconds
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
	statement           string
	digestText          string
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

func (m *dbmSampleMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Desc: "MySQL DBM execution plan objects. Each object represents a sampled query execution plan identified by query_signature and plan_signature.",
		Name: dbmExecPlanObjectName,
		Cat:  point.Object,
		Fields: map[string]interface{}{
			"timestamp": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.TimestampMS,
				Desc:     "The timestamp (millisecond) when the statement finished executing.",
			},
			"duration": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.DurationNS,
				Desc:     "Execution time of the statement (nanoseconds).",
			},
			"lock_time_ns": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.DurationNS,
				Desc:     "Time in nanoseconds spent waiting for locks.",
			},
			"no_good_index_used": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.EnumValue,
				Desc:     "0 if a good index was found for the statement, 1 if no good index was found.",
			},
			"no_index_used": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.EnumValue,
				Desc:     "0 if an index was used for the statement, 1 if no index was used.",
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
				Desc:     "Number of rows returned to the client.",
			},
			"message": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.String,
				Unit:     inputs.NoUnit,
				Desc:     "The obfuscated/normalized JSON execution plan definition.",
			},
			"statement": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.String,
				Unit:     inputs.NoUnit,
				Desc:     "The obfuscated/normalized SQL text corresponding to this execution plan.",
			},
		},
		Tags: map[string]interface{}{
			"name":              &inputs.TagInfo{Desc: "Object identity built from server, database_instance, and query_signature:plan_signature."},
			"server":            &inputs.TagInfo{Desc: "The server address (host:port)."},
			"database_instance": &inputs.TagInfo{Desc: "MySQL instance identifier from configured tag or @@server_uuid."},
			"database_type":     &inputs.TagInfo{Desc: "The type of the database. The value is `MySQL`."},
			"plan_type":         &inputs.TagInfo{Desc: "The format of the plan content. The value is `JSON`."},
			"schema_name":       &inputs.TagInfo{Desc: "The schema name."},
			"plan_signature":    &inputs.TagInfo{Desc: "Hash of the normalized execution plan to group identical plans."},
			"query_signature":   &inputs.TagInfo{Desc: "Hash from schema+digest_text, used to link with metrics and query objects."},
			"digest":            &inputs.TagInfo{Desc: "The digest hash from the original normalized statement (performance_schema)."},
		},
	}
}

// Point implement MeasurementV2.
func (m *dbmSampleMeasurement) Point() *point.Point {
	opts := point.DefaultObjectOptions()

	if m.election {
		opts = append(opts, point.WithExtraTags(datakit.GlobalElectionTags()))
	}

	return point.NewPoint(m.name,
		append(point.NewTags(m.tags), point.NewKVs(m.fields)...),
		opts...)
}

// get the table from which samples should be collected.
func getSampleCollectionStrategy(i *Input) (string, error) {
	if len(i.dbmSampleCache.eventsStatementsTable) > 0 {
		return i.dbmSampleCache.eventsStatementsTable, nil
	}

	var eventsStatementsTable string

	enabledSQL := `SELECT name
	FROM performance_schema.setup_consumers
	WHERE enabled = 'YES' AND name LIKE 'events_statements_%' AND name != 'events_statements_cpu'`
	ctx, cancel := context.WithTimeout(context.Background(), i.timeoutDuration)
	enabledConsumers := getCleanEnabledPerformanceSchemaConsumers(i.q(ctx, enabledSQL, getMetricName(metricNameMySQLDbmSample, "setup_consumers")))
	cancel()

	if len(enabledConsumers) < 3 {
		err := enablePerformanceSchemaConsumers(i)
		if err != nil {
			l.Warn(err)
		} else {
			ctx, cancel := context.WithTimeout(context.Background(), i.timeoutDuration)
			enabledConsumers = getCleanEnabledPerformanceSchemaConsumers(i.q(ctx, enabledSQL, getMetricName(metricNameMySQLDbmSample, "setup_consumers")))
			cancel()
		}
	}

	if len(enabledConsumers) == 0 {
		return "", errors.New("no events_statements consumer")
	}

	l.Debugf("enabled performance_schema statements consumers: %s", enabledConsumers)

	tables := []string{
		// "events_statements_history_long",
		"events_statements_current",
		// "events_statements_history",
	}

	for _, table := range tables {
		if !isListHasStr(enabledConsumers, table) {
			continue
		}

		eventsStatementsTable = table
		break
	}

	if len(eventsStatementsTable) == 0 {
		return "", fmt.Errorf("all enabled events_statements_consumers %v are empty", enabledConsumers)
	}

	i.dbmSampleCache.eventsStatementsTable = eventsStatementsTable

	return eventsStatementsTable, nil
}

// enable consumers at runtime.
func enablePerformanceSchemaConsumers(i *Input) error {
	ctx, cancel := context.WithTimeout(context.Background(), i.timeoutDuration)
	defer cancel()

	sqlStr := "CALL datakit.enable_events_statements_consumers()"
	if _, err := i.db.ExecContext(ctx, sqlStr); err != nil {
		return err
	}
	return nil
}

// collect events.
func getNewEventsStatements(i *Input, eventsStatementsTable string, rowLimit int) ([]eventRow, error) {
	ctx, cancel := context.WithTimeout(context.Background(), i.timeoutDuration)
	defer cancel()

	var rows []eventRow
	conn, err := i.db.Conn(ctx)
	if err != nil {
		return rows, err
	}
	defer conn.Close() //nolint:errcheck

	if len(i.dbmSampleCache.globalStatusTable) == 0 {
		return rows, errors.New("globalStatusTable not set")
	}

	// Datadog EVENTS_STATEMENTS_CURRENT_QUERY: set @uptime first
	uptimeSQL := fmt.Sprintf("SET @uptime = (SELECT VARIABLE_VALUE FROM %s WHERE VARIABLE_NAME='UPTIME')", i.dbmSampleCache.globalStatusTable)
	if _, err := conn.ExecContext(ctx, uptimeSQL); err != nil {
		return rows, err
	}

	eventsStatementsQuery := `
	SELECT
		current_schema,
		sql_text,
		digest,
		digest_text,
		end_event_id,
		timer_start,
		@uptime AS uptime,
		unix_timestamp() AS now,
		timer_end,
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
	FROM performance_schema.%s E
	LEFT JOIN performance_schema.threads AS T ON E.thread_id = T.thread_id
	WHERE sql_text IS NOT NULL
		AND event_name LIKE 'statement/%%'
		AND digest_text IS NOT NULL
		AND digest_text NOT LIKE 'EXPLAIN %%'
	ORDER BY timer_wait DESC
	LIMIT %d
`
	eventsStatementsQuerySQL := fmt.Sprintf(eventsStatementsQuery, eventsStatementsTable, rowLimit)
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
			&row.endEventID,
			&row.timerStart,
			&row.uptime,
			&row.now,
			&row.timerEnd,
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
	return rows, nil
}

func filterValidStatementRows(i *Input, rows []eventRow) []eventRow {
	var filterRows []eventRow
	eventTimestampMs := time.Now().UnixMilli()
	windowMs := i.DbmSample.PlanCacheTTL.Duration.Milliseconds()
	if windowMs <= 0 {
		windowMs = 3600 * 1000 // 1 hour
	}
	for _, row := range rows {
		if !row.sqlText.Valid || len(row.sqlText.String) == 0 {
			continue
		}
		// Skip completed queries past window
		if hasSampledSinceCompletion(row, eventTimestampMs, windowMs) {
			l.Debugf("skip completed query past window: %s", row.sqlText.String)
			continue
		}
		filterRows = append(filterRows, row)
	}

	return filterRows
}

func collectPlanForStatements(i *Input, rows []eventRow) []planObj {
	var plans []planObj
	sqlObfuscator := newMySQLSQLObfuscator()
	obfuscator := util.NewSQLPlanObfuscator()
	for _, row := range rows {
		plan, err := collectPlanForStatement(i, row, sqlObfuscator, obfuscator)
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

func collectPlanForStatement(i *Input, row eventRow, sqlObfuscator, planObfuscator *obfuscate.Obfuscator) (planObj, error) {
	var plan planObj
	obfSQLResult, err := sqlObfuscator.ObfuscateSQLString(row.sqlText.String)
	if err != nil {
		l.Warnf("obfuscate sql text failed: %s", err.Error())
		return plan, nil
	}
	obfuscatedStatement := obfSQLResult.Query

	obfDigestResult, err := sqlObfuscator.ObfuscateSQLString(row.digestText.String)
	if err != nil {
		l.Warnf("obfuscate digest text failed: %s", err.Error())
		return plan, nil
	}
	obfuscatedDigestText := obfDigestResult.Query

	// querySignature: keep consistent with other MySQL DBM places (schema + digest_text via xxhash)
	querySignature := generateQuerySignature(row.currentSchema.String, obfuscatedDigestText)

	if !checkLimitRate(i, querySignature) {
		l.Debugf("check limit rate failed, ignore plan: %s", obfuscatedStatement)
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
		normalizedPlan, err := planObfuscator.ObfuscateSQLExecPlan(planStr, true)
		if err != nil {
			return plan, fmt.Errorf("failed to normalize obfuscate: %w", err)
		}
		obfuscatedPlan, err := planObfuscator.ObfuscateSQLExecPlan(planStr, false)
		if err != nil {
			return plan, fmt.Errorf("failed to obfuscate: %w", err)
		}

		planSignature := ComputeSQLPlanSignature(normalizedPlan)
		planKey := generatePlanCacheKey(querySignature, planSignature)
		if !checkPlanSignatureRate(i, planKey) {
			return plan, nil
		}
		plan = planObj{
			planDefinition: obfuscatedPlan,
			planSignature:  planSignature,
			querySignature: querySignature,
			digestText:     obfuscatedDigestText,
			statement:      obfuscatedStatement,
		}

		// Reuse overflow-safe timer_end conversion.
		if tsMs, ok := calculateTimerEndMs(row); ok {
			plan.timestamp = float64(tsMs)
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
			if lockTimeNs, err := strconv.ParseInt(row.lockTimeNs.String, 10, 64); err == nil {
				plan.lockTimeNs = lockTimeNs
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
	if i.dbmSampleCache.explainCache == nil {
		ttl := i.DbmSample.ExplainCacheTTL.Duration
		if ttl <= 0 {
			ttl = 10 * time.Minute
		}
		i.dbmSampleCache.explainCache = &util.CacheLimit{
			Size: 1000,
			TTL:  int64(ttl.Seconds()),
		}
	}
	return i.dbmSampleCache.explainCache.Acquire(key)
}

// checkPlanSignatureRate returns true if this plan object key should be reported (not seen recently).
// Uses an expirable LRU to limit how often the same plan object is reported.
func checkPlanSignatureRate(i *Input, planKey string) bool {
	if planKey == "" {
		return false
	}
	if i.planSignatureCache == nil {
		ttl := i.DbmSample.PlanCacheTTL.Duration
		if ttl <= 0 {
			ttl = time.Hour
		}
		i.planSignatureCache = expirable.NewLRU[string, struct{}](defaultExplainCacheSize, nil, ttl)
	}
	if _, ok := i.planSignatureCache.Get(planKey); ok {
		l.Debugf("skip duplicate plan object (LRU): %s", planKey)
		return false
	}
	return true
}

// recordReportedPlanSignature records the plan object key in the LRU cache after the plan object is actually reported.
func (ipt *Input) recordReportedPlanSignature(planKey string) {
	if planKey == "" {
		return
	}
	ipt.planSignatureCache.Add(planKey, struct{}{})
}

func calculateTimerEndMs(row eventRow) (int64, bool) {
	if !row.now.Valid || !row.uptime.Valid || row.uptime.String == "" || !row.timerEnd.Valid {
		return 0, false
	}
	uptimeSec, err := strconv.ParseInt(row.uptime.String, 10, 64)
	if err != nil {
		return 0, false
	}
	// timer_end is picoseconds;
	bigintMaxInSeconds := float64(math.MaxUint64) * 1e-12
	secondsToAdd := math.Floor(float64(uptimeSec)/bigintMaxInSeconds) * bigintMaxInSeconds
	timerEndTimeS := float64(row.now.Int64) - float64(uptimeSec) + secondsToAdd + float64(row.timerEnd.Int64)*1e-12
	return int64(timerEndTimeS * 1000), true
}

func hasSampledSinceCompletion(row eventRow, eventTimestampMs int64, windowMs int64) bool {
	if !row.endEventID.Valid {
		return false
	}
	queryEndMs, ok := calculateTimerEndMs(row)
	if !ok {
		return false
	}
	timeDiff := eventTimestampMs - queryEndMs
	if timeDiff < 0 {
		timeDiff = -timeDiff
	}
	return timeDiff > windowMs
}

func isTruncated(statement string) bool {
	return strings.HasSuffix(statement, "...")
}

func explainStatement(i *Input, statement string, schema string, obfuscatedStatement string) (string, error) {
	var plan string
	ctx, cancel := context.WithTimeout(context.Background(), i.timeoutDuration)
	defer cancel()

	conn, err := i.db.Conn(ctx)
	if err != nil {
		return plan, err
	}
	defer conn.Close() //nolint:errcheck

	if !canExplain(obfuscatedStatement) {
		l.Debugf("ignore explain statement: %s", obfuscatedStatement)
		return plan, nil
	}

	// TODO cached strategy
	if len(schema) > 0 {
		if _, err := conn.ExecContext(ctx, fmt.Sprintf("USE `%s`", schema)); err != nil {
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
			plan, err = runExplainProcedure(ctx, conn, statement)
			if err != nil {
				l.Debugf("explain procedure error: %s", err.Error())
			}
		}
		if strategy == "FQ_PROCEDURE" {
			plan, err = runFullyQualifiedExplainProcedure(ctx, conn, statement)
			if err != nil {
				l.Debugf("explain fully qualified procedure error: %s", err.Error())
			}
		}
		if strategy == "STATEMENT" {
			plan, err = runExplain(ctx, conn, statement)
			if err != nil {
				l.Debugf("explain statement error: %s", err.Error())
			}
		}
		if err != nil {
			continue
		}
		if len(plan) > 0 {
			return plan, nil
		}
	}
	return plan, nil
}

// runExplainProcedure calls explain_statement(?) with statement as a bound parameter, so no manual.
func runExplainProcedure(ctx context.Context, conn *sql.Conn, statement string) (string, error) {
	//nolint:execinquery
	row := conn.QueryRowContext(ctx, "CALL explain_statement(?)", statement)
	var plan string
	if row.Err() != nil {
		return plan, row.Err()
	}
	if err := row.Scan(&plan); err != nil {
		return plan, err
	}

	return plan, nil
}

// runFullyQualifiedExplainProcedure calls datakit.explain_statement(?) with statement as a bound parameter.
func runFullyQualifiedExplainProcedure(ctx context.Context, conn *sql.Conn, statement string) (string, error) {
	//nolint:execinquery
	row := conn.QueryRowContext(ctx, "CALL datakit.explain_statement(?)", statement)
	var plan string
	if row.Err() != nil {
		return plan, row.Err()
	}
	if err := row.Scan(&plan); err != nil {
		return plan, err
	}

	return plan, nil
}

func runExplain(ctx context.Context, conn *sql.Conn, statement string) (string, error) {
	//nolint:execinquery
	row := conn.QueryRowContext(ctx, "EXPLAIN FORMAT=json "+statement)
	var plan string
	if row.Err() != nil {
		return plan, row.Err()
	}
	if err := row.Scan(&plan); err != nil {
		return plan, err
	}

	return plan, nil
}

func ComputeSQLPlanSignature(normalizedPlan string) string {
	h := xxhash.New()
	_, _ = h.WriteString(normalizedPlan)
	return fmt.Sprintf("%016x", h.Sum64())
}

func generatePlanCacheKey(querySignature, planSignature string) string {
	h := xxhash.New()
	_, _ = h.WriteString(querySignature)
	_, _ = h.WriteString(":")
	_, _ = h.WriteString(planSignature)
	return fmt.Sprintf("%016x", h.Sum64())
}
