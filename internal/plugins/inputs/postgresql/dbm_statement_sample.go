// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package postgresql

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"github.com/hashicorp/golang-lru/v2/expirable"
	"github.com/spf13/cast"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/util"
)

type dbmSample struct {
	Enabled bool `toml:"enabled"`

	explainCache                 *util.CacheLimit
	explainErrorCache            *expirable.LRU[string, struct{}]
	explainParameterizedInstance *ExplainParameterizedQueries
}

type dbmActivity struct {
	Enabled bool `toml:"enabled"`
}

var (
	PGStatActivityColumns = []string{
		"datid",
		"datname",
		"pid",
		"usesysid",
		"usename",
		"application_name",
		"client_addr",
		"client_hostname",
		"client_port",
		"backend_start",
		"xact_start",
		"query_start",
		"state_change",
		"wait_event_type",
		"wait_event",
		"state",
		"backend_xid",
		"backend_xmin",
		"query",
		"backend_type",
	}
	PGStatActivityColumnsMap = map[string]string{
		"backend_type": "backend_type::bytea as backend_type",
	}
	CurrentTimeFunc    = "clock_timestamp() as now,"
	PGBlockingPidsFunc = ",pg_blocking_pids(pid) as blocking_pids"
)

func (ipt *Input) getDbmSample() error {
	activityColumns, err := ipt.getPGStatActivityColumns(PGStatActivityColumns)
	if err != nil {
		return fmt.Errorf("get activity columns failed: %w", err)
	}

	if len(activityColumns) == 0 {
		return fmt.Errorf("no activity columns")
	}

	rows, err := ipt.getNewPGStatActivityRows(activityColumns)
	if err != nil {
		return fmt.Errorf("get activity rows failed: %w", err)
	}

	if ipt.DbmSample.Enabled {
		ipt.collectSamplePlans(rows)
	}

	if ipt.DbmActivity.Enabled {
		ipt.collectSampleActivity(rows)
	}

	return nil
}

func (ipt *Input) collectSamplePlans(rows []map[string]any) {
	for _, row := range rows {
		if (cast.ToString(row["statement"]) == "") || cast.ToString(row["backend_type"]) != "client backend" {
			continue
		}
		cacheKey := fmt.Sprintf("%s-%s", cast.ToString(row["datname"]), cast.ToString(row["query_signature"]))
		if !ipt.DbmSample.explainCache.Acquire(cacheKey) {
			continue
		}

		plan, err := ipt.getPlan(
			cast.ToString(row["datname"]),
			cast.ToString(row["query"]),
			cast.ToString(row["statement"]),
			cast.ToString(row["query_signature"]),
		)
		if err != nil {
			l.Warnf("get plan failed: %v", err.Error())
			continue
		}

		if plan == "" {
			continue
		}

		kvs := ipt.getKVs()
		opts := ipt.getKVsOpts(point.Logging)
		opts = append(opts, point.WithTimestamp(ipt.ptsTime.UnixNano()))
		kvs = kvs.AddTag("service", "postgresql")
		kvs = kvs.AddTag("status", "info")
		kvs = kvs.AddTag("query_signature", cast.ToString(row["query_signature"]))
		kvs = kvs.AddTag("client_hostname", cast.ToString(row["client_hostname"]))
		kvs = kvs.AddTag("client_port", cast.ToString(row["client_port"]))
		kvs = kvs.AddTag("client_addr", cast.ToString(row["client_addr"]))
		kvs = kvs.AddTag("application_name", cast.ToString(row["application_name"]))
		kvs = kvs.AddTag("usename", cast.ToString(row["usename"]))
		kvs = kvs.AddTag("datname", cast.ToString(row["datname"]))

		kvs = kvs.Set("message", cast.ToString(row["statement"]))
		kvs = kvs.Set("plan_definition", util.ObfuscateSQLExecPlan(plan, &util.ObfuscateLogger{Log: l}))

		ipt.collectCache[point.Logging] = append(ipt.collectCache[point.Logging], point.NewPoint(dbmSampleMeasurementInfo.Name, kvs, opts...))
	}
}

func (ipt *Input) getPlan(
	datname, statement, obfuscatedStatement, querySignature string,
) (string, error) {
	// originalStatement := statement
	if strings.ToLower(obfuscatedStatement[:3]) == "set" {
		statement = TrimLeadingSetStmts(statement)
		obfuscatedStatement = TrimLeadingSetStmts(obfuscatedStatement)
	}

	if !canExplainStatement(obfuscatedStatement) {
		l.Debugf("explain statement not supported: %s", obfuscatedStatement)
		return "", nil
	}

	if _, ok := ipt.DbmSample.explainErrorCache.Get(querySignature); ok {
		l.Debugf("explain statement error in cache: %s", obfuscatedStatement)
		return "", nil
	}

	if isParameterizedQuery(statement) {
		if ipt.DbmSample.explainParameterizedInstance == nil {
			return "", fmt.Errorf("explain parameterized instance is nil")
		}

		if plan, err := ipt.DbmSample.explainParameterizedInstance.ExplainStatement(datname, statement, obfuscatedStatement); err != nil {
			ipt.DbmSample.explainErrorCache.Add(querySignature, struct{}{})
			return "", fmt.Errorf("explain parameterized statement failed: %w", err)
		} else {
			return plan, nil
		}
	} else if plan, err := ipt.runExplain(datname, statement, obfuscatedStatement); err != nil {
		ipt.DbmSample.explainErrorCache.Add(querySignature, struct{}{})
		return "", fmt.Errorf("run explain failed: %w", err)
	} else {
		return plan, nil
	}
}

func (ipt *Input) runExplain(datname, statement, obfuscatedStatement string) (string, error) {
	conn, err := ipt.service.GetConn(datname)
	if err != nil {
		return "", fmt.Errorf("get conn failed: %w", err)
	}
	defer conn.Close()

	var encoding string
	rows, err := conn.Query(context.Background(), "SHOW client_encoding")
	if err != nil {
		return "", fmt.Errorf("query encoding failed: %w", err)
	}
	if rows.Next() {
		if err := rows.Scan(&encoding); err != nil {
			return "", fmt.Errorf("scan encoding failed: %w", err)
		}
	}
	rows.Close()

	l.Debugf("run explain query on database=%s, statement=%s", datname, obfuscatedStatement)
	if encoding == "SQLASCII" {
		l.Debugf("set client_encoding to utf-8, current encoding is %s", encoding)
		err := conn.Exec(context.Background(), "SET client_encoding = 'utf-8'")
		if err != nil {
			return "", fmt.Errorf("set client encoding failed: %w", err)
		}
	}

	query := fmt.Sprintf("SELECT %s($stmt$%s$stmt$)", "datakit.explain_statement", statement)

	var explainResult string
	rows, err = conn.Query(context.Background(), query)
	if err != nil {
		return "", fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	if rows.Next() {
		if err := rows.Scan(&explainResult); err != nil {
			return "", fmt.Errorf("scan explain result failed: %w", err)
		}
	}

	if len(explainResult) == 0 {
		return "", nil
	}

	return explainResult, nil
}

func (ipt *Input) collectSampleActivity(rows []map[string]any) {
	for _, row := range rows {
		kvs := ipt.getKVs()
		opts := ipt.getKVsOpts(point.Logging)
		opts = append(opts, point.WithTimestamp(ipt.ptsTime.UnixNano()))
		kvs = kvs.AddTag("service", "postgresql")
		kvs = kvs.AddTag("status", "info")
		kvs = kvs.AddTag("query_signature", cast.ToString(row["query_signature"]))
		kvs = kvs.AddTag("client_hostname", cast.ToString(row["client_hostname"]))
		kvs = kvs.AddTag("client_port", cast.ToString(row["client_port"]))
		kvs = kvs.AddTag("client_addr", cast.ToString(row["client_addr"]))
		kvs = kvs.AddTag("application_name", cast.ToString(row["application_name"]))
		kvs = kvs.AddTag("usename", cast.ToString(row["usename"]))
		kvs = kvs.AddTag("datname", cast.ToString(row["datname"]))
		kvs = kvs.AddTag("state", cast.ToString(row["state"]))
		kvs = kvs.AddTag("pid", cast.ToString(row["pid"]))
		kvs = kvs.AddTag("wait_event_type", cast.ToString(row["wait_event_type"]))
		kvs = kvs.AddTag("wait_event", cast.ToString(row["wait_event"]))
		kvs = kvs.AddTag("backend_type", cast.ToString(row["backend_type"]))
		kvs = kvs.AddTag("message", cast.ToString(row["statement"]))

		kvs = kvs.Set("backend_start", cast.ToInt64(row["backend_start"]))
		kvs = kvs.Set("query_start", cast.ToInt64(row["query_start"]))
		kvs = kvs.Set("xact_start", cast.ToInt64(row["xact_start"]))
		kvs = kvs.Set("state_change", cast.ToInt64(row["state_change"]))

		ipt.collectCache[point.Logging] = append(ipt.collectCache[point.Logging], point.NewPoint(dbmActivityMeasurementInfo.Name, kvs, opts...))
	}
}

func (ipt *Input) getPGStatActivityColumns(expected []string) ([]string, error) {
	if len(ipt.dbQueryCache.PGStatActivityColumns) > 0 {
		return ipt.dbQueryCache.PGStatActivityColumns, nil
	}

	columns, err := ipt.getSQLColumns("select * from pg_stat_activity limit 0")
	if err != nil {
		return nil, fmt.Errorf("get columns failed: %w", err)
	}

	columnMap := make(map[string]bool)
	for _, column := range columns {
		columnMap[column] = true
	}

	availableColumns := make([]string, 0)
	for _, column := range expected {
		if _, ok := columnMap[column]; ok {
			availableColumns = append(availableColumns, column)
		}
	}

	ipt.dbQueryCache.PGStatActivityColumns = availableColumns
	return ipt.dbQueryCache.PGStatActivityColumns, nil
}

const sqlGetPGStatActivity = `
    SELECT %s %s %s FROM pg_stat_activity
    WHERE %s
        (coalesce(TRIM(query), '') != '' AND pid != pg_backend_pid() AND query_start IS NOT NULL %s)
`

func (ipt *Input) getNewPGStatActivityRows(queryColumns []string) ([]map[string]any, error) {
	filters := ""
	args := []any{}
	if len(ipt.IgnoredDatabases) > 0 {
		filters += fmt.Sprintf(" AND datname NOT IN ('%s')", strings.Join(ipt.IgnoredDatabases, "','"))
	} else if len(ipt.Databases) > 0 {
		filters += fmt.Sprintf(" AND datname IN ('%s')", strings.Join(ipt.Databases, "','"))
	}

	if !ipt.dbQueryCache.ActivityLastQueryStart.IsZero() {
		filters += " AND NOT (query_start < $1 AND state = 'idle')"
		args = append(args, ipt.dbQueryCache.ActivityLastQueryStart)
	}

	timeFunc := CurrentTimeFunc
	blockingFunc := ""
	backendTypePredicate := ""

	if V100.LessThan(*ipt.version) {
		backendTypePredicate = "backend_type != 'client backend' OR"
	}

	if ipt.DbmActivity.Enabled && (V96.LessThan(*ipt.version) || V96.Equal(*ipt.version)) {
		blockingFunc = PGBlockingPidsFunc
	}

	activityColumns := []string{}
	for _, column := range queryColumns {
		if col, ok := PGStatActivityColumnsMap[column]; ok {
			activityColumns = append(activityColumns, col)
		} else {
			activityColumns = append(activityColumns, column)
		}
	}

	sql := fmt.Sprintf(sqlGetPGStatActivity, timeFunc, strings.Join(activityColumns, ","), blockingFunc, backendTypePredicate, filters)

	rows, err := ipt.service.Query(sql, args...)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("get columns failed: %w", err)
	}

	newRows := []map[string]any{}
	totalCount := 0
	insufficientPrivilegeCount := 0
	for rows.Next() {
		totalCount++
		columnMap, err := ipt.service.GetColumnMap(rows, columns)
		if err != nil {
			return nil, fmt.Errorf("get column map failed: %w", err)
		}
		newRow := map[string]any{}
		for k, v := range columnMap {
			if v == nil {
				continue
			}

			switch trueVal := (*v).(type) {
			case []uint8:
				newRow[k] = string(trueVal)
			case string:
				newRow[k] = trueVal
			case time.Time:
				newRow[k] = trueVal.UnixMicro()
				if k == "query_start" && trueVal.After(ipt.dbQueryCache.ActivityLastQueryStart) {
					ipt.dbQueryCache.ActivityLastQueryStart = trueVal
				}
			default:
				newRow[k] = cast.ToString(trueVal)
			}
		}
		query := cast.ToString(newRow["query"])
		if query == "" {
			continue
		}
		if query == "<insufficient privilege>" {
			insufficientPrivilegeCount++
			continue
		}

		datname := cast.ToString(newRow["datname"])
		backendType := cast.ToString(newRow["backend_type"])

		if (datname == "") && (backendType == "client backend") {
			continue
		}

		if backendType != "client backend" {
			newRow["query_signature"] = util.ComputeSQLSignature(backendType)
		} else {
			newRow["statement"] = util.ObfuscateSQL(query)
			newRow["query_signature"] = util.ComputeSQLSignature(query)
		}

		newRows = append(newRows, newRow)
	}
	if insufficientPrivilegeCount > 0 {
		l.Warnf("insufficient privilege for %d of %d queries when collecting from pg_stat_activity", insufficientPrivilegeCount, totalCount)
	}

	return newRows, nil
}

var supportedExplainStatements = map[string]struct{}{
	"select":  {},
	"table":   {},
	"delete":  {},
	"insert":  {},
	"replace": {},
	"update":  {},
	"with":    {},
}

func canExplainStatement(obfuscateSQL string) bool {
	obfuscateSQL = strings.TrimSpace(obfuscateSQL)

	// ignore explain statement
	if strings.HasPrefix(obfuscateSQL, "SELECT datakit.explain_statement") {
		return false
	}

	// ignore autovacuum statement
	if strings.HasPrefix(obfuscateSQL, "autovacuum:") {
		return false
	}

	// statement type
	parts := strings.SplitN(obfuscateSQL, " ", 2)
	if len(parts) == 0 {
		return false
	}

	stmtType := strings.ToLower(parts[0])
	_, ok := supportedExplainStatements[stmtType]

	return ok
}
