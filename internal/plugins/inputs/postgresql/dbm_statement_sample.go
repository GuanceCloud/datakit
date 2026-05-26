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
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/util"

	"github.com/DataDog/datadog-agent/pkg/obfuscate"
)

type dbmSample struct {
	Enabled         bool             `toml:"enabled"`
	ExplainCacheTTL datakit.Duration `toml:"explain_cache_ttl"`
	PlanCacheTTL    datakit.Duration `toml:"plan_cache_ttl"`

	explainCache                 *util.CacheLimit
	explainErrorCache            *expirable.LRU[string, struct{}]
	planSignatureCache           *expirable.LRU[string, struct{}]
	explainParameterizedInstance *ExplainParameterizedQueries
}

type dbmActivity struct {
	Enabled  bool             `toml:"enabled"`
	Interval datakit.Duration `toml:"interval"`
}

const postgreSQLBackendTypeClient = "client backend"

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

func (ipt *Input) collectDbmActivityRows() ([]map[string]any, error) {
	activityColumns, err := ipt.getPGStatActivityColumns(PGStatActivityColumns)
	if err != nil {
		return nil, fmt.Errorf("get activity columns failed: %w", err)
	}

	if len(activityColumns) == 0 {
		return nil, fmt.Errorf("no activity columns")
	}

	rows, err := ipt.getNewPGStatActivityRows(activityColumns)
	if err != nil {
		return nil, fmt.Errorf("get activity rows failed: %w", err)
	}

	return rows, nil
}

func (ipt *Input) collectSampleActivity(rows []map[string]any, ptsTime time.Time) []*point.Point {
	pts := make([]*point.Point, 0, len(rows))
	sampleTimeUS := ptsTime.UnixMicro()
	for _, row := range rows {
		kvs := ipt.getKVs()
		opts := ipt.getKVsOpts(point.Logging)
		opts = append(opts, point.WithTimestamp(ptsTime.UnixNano()))
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
		kvs = kvs.AddTag("wait_group", cast.ToString(row["wait_group"]))
		kvs = kvs.AddTag("backend_type", cast.ToString(row["backend_type"]))
		kvs = kvs.AddTag("message", cast.ToString(row["statement"]))

		kvs = kvs.Set("backend_start", cast.ToInt64(row["backend_start"]))
		kvs = kvs.Set("query_start", cast.ToInt64(row["query_start"]))
		kvs = kvs.Set("xact_start", cast.ToInt64(row["xact_start"]))
		kvs = kvs.Set("state_change", cast.ToInt64(row["state_change"]))
		rowNow := cast.ToInt64(row["now"])
		if rowNow <= 0 {
			rowNow = sampleTimeUS
		}
		queryStart := cast.ToInt64(row["query_start"])
		xactStart := cast.ToInt64(row["xact_start"])
		stateChange := cast.ToInt64(row["state_change"])
		state := strings.TrimSpace(strings.ToLower(cast.ToString(row["state"])))

		if queryStart > 0 {
			switch state {
			case "idle", "idle in transaction":
				if stateChange >= queryStart {
					kvs = kvs.Set("duration", stateChange-queryStart)
				} else {
					kvs = kvs.Set("duration", int64(0))
				}
			default:
				if rowNow >= queryStart {
					kvs = kvs.Set("duration", rowNow-queryStart)
				} else {
					kvs = kvs.Set("duration", int64(0))
				}
			}
		} else {
			kvs = kvs.Set("duration", int64(0))
		}

		if xactStart > 0 && rowNow >= xactStart {
			kvs = kvs.Set("tx_duration", rowNow-xactStart)
		} else {
			kvs = kvs.Set("tx_duration", int64(0))
		}

		if stateChange > 0 && rowNow >= stateChange {
			kvs = kvs.Set("wait_duration", rowNow-stateChange)
		} else {
			kvs = kvs.Set("wait_duration", int64(0))
		}
		kvs = kvs.Set("blocking_pids", cast.ToString(row["blocking_pids"]))

		pts = append(pts, point.NewPoint(dbmActivityMeasurementInfo.Name, kvs, opts...))
	}
	return pts
}

func shouldCollectPGActivity(row map[string]any) bool {
	backendType := strings.TrimSpace(strings.ToLower(cast.ToString(row["backend_type"])))
	if backendType == "" {
		backendType = postgreSQLBackendTypeClient
	}
	state := strings.TrimSpace(strings.ToLower(cast.ToString(row["state"])))
	return backendType != postgreSQLBackendTypeClient || state != "idle"
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

	if V100.LessThan(*ipt.version) || V100.Equal(*ipt.version) {
		backendTypePredicate = fmt.Sprintf("backend_type != '%s' OR", postgreSQLBackendTypeClient)
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

	ctx, cancel := context.WithTimeout(context.Background(), ipt.Timeout.Duration)
	defer cancel()

	rows, err := ipt.service.Query(ctx, sql, args...)
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
	o := obfuscate.NewObfuscator(obfuscate.Config{
		SQL: obfuscate.SQLConfig{
			DBMS: obfuscate.DBMSPostgres,
		},
	})
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
			case []interface{}:
				if k == "blocking_pids" {
					newRow[k] = strings.Join(cast.ToStringSlice(trueVal), ",")
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
		usename := cast.ToString(newRow["usename"])
		backendType := cast.ToString(newRow["backend_type"])

		if (datname == "") && (backendType == postgreSQLBackendTypeClient) {
			continue
		}

		if backendType != postgreSQLBackendTypeClient {
			newRow["query_signature"] = generateQuerySignature(datname, usename, backendType)
		} else {
			obfResult, err := o.ObfuscateSQLString(query)
			if err != nil {
				l.Warnf("obfuscate dbm activity sql failed: %s, query: %s", err.Error(), query)
				continue
			}
			newRow["statement"] = obfResult.Query
			newRow["query_signature"] = generateQuerySignature(datname, usename, obfResult.Query)
		}
		newRow["wait_group"] = getPGActivityWaitGroup(newRow)
		if !shouldCollectPGActivity(newRow) {
			continue
		}

		newRows = append(newRows, newRow)
	}
	if insufficientPrivilegeCount > 0 {
		l.Warnf("insufficient privilege for %d of %d queries when collecting from pg_stat_activity", insufficientPrivilegeCount, totalCount)
	}

	return newRows, nil
}

func getPGActivityWaitGroup(row map[string]any) string {
	state := strings.ToLower(strings.TrimSpace(cast.ToString(row["state"])))
	blocked := strings.TrimSpace(cast.ToString(row["blocking_pids"])) != ""
	sessionStatus := getPGSessionStatus(state, blocked)
	return mapPostgreSQLWaitGroup(sessionStatus, cast.ToString(row["wait_event_type"]), cast.ToString(row["wait_event"]))
}

func mapPostgreSQLWaitGroup(sessionStatus, waitEventType, waitEvent string) string {
	waitType := strings.ToLower(strings.TrimSpace(waitEventType))
	waitName := strings.TrimSpace(waitEvent)
	waitNameUpper := strings.ToUpper(waitName)

	// Running on CPU usually has no wait event.
	if sessionStatus == "active" && waitType == "" && waitName == "" {
		return waitGroupCPU
	}

	switch waitType {
	case "lock":
		return waitGroupLock
	case "lwlock", "bufferpin":
		if strings.Contains(waitNameUpper, "WAL") {
			return waitGroupCommitLog
		}
		return waitGroupConcurrency
	case "io":
		if strings.Contains(waitNameUpper, "WAL") {
			return waitGroupCommitLog
		}
		return waitGroupIO
	case "client":
		return waitGroupNetwork
	case "ipc":
		if strings.Contains(waitNameUpper, "SYNCREP") {
			return waitGroupCommitLog
		}
		return waitGroupOther
	case "activity", "timeout", "extension":
		return waitGroupOther
	default:
		if strings.Contains(waitNameUpper, "WAL") {
			return waitGroupCommitLog
		}
		return waitGroupOther
	}
}

func getPGSessionStatus(state string, blocked bool) string {
	if blocked {
		return "blocked"
	}
	if state == "active" {
		return "active"
	}
	return "idle"
}
