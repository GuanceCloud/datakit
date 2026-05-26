// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package postgresql

import (
	"context"
	// nolint:gosec
	"fmt"
	"strings"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"github.com/cespare/xxhash/v2"
	"github.com/spf13/cast"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"

	"github.com/DataDog/datadog-agent/pkg/obfuscate"
)

var (
	statStatementsRequiredColumns = map[string]bool{
		"calls": true,
		"query": true,
		"rows":  true,
	}
	statStatementsTimingColumns = map[string]bool{
		"shared_blk_read_time":  true,
		"shared_blk_write_time": true,
	}
	statStatementsTimingColumnsLT17 = map[string]bool{
		"blk_read_time":  true,
		"blk_write_time": true,
	}
	statStatementsTagColumns = map[string]bool{
		"datname": true,
		"rolname": true,
		"query":   true,
	}
	statStatementsOptionalColumns = map[string]bool{
		"queryid": true,
	}
)

var statStatementsMetricsNoTimingColumns = map[string]bool{
	"calls":               true,
	"rows":                true,
	"total_time":          true,
	"total_exec_time":     true,
	"shared_blks_hit":     true,
	"shared_blks_read":    true,
	"shared_blks_dirtied": true,
	"shared_blks_written": true,
	"local_blks_hit":      true,
	"local_blks_read":     true,
	"local_blks_dirtied":  true,
	"local_blks_written":  true,
	"temp_blks_read":      true,
	"temp_blks_written":   true,
	"wal_records":         true,
	"wal_fpi":             true,
	"wal_bytes":           true,
	"total_plan_time":     true,
	"min_plan_time":       true,
	"max_plan_time":       true,
	"mean_plan_time":      true,
	"stddev_plan_time":    true,
}

var statStatementsDeltaColumns = map[string]string{
	"calls":                 "delta_calls",
	"rows":                  "delta_rows",
	"total_time":            "delta_total_exec_time",
	"total_exec_time":       "delta_total_exec_time",
	"shared_blks_hit":       "delta_shared_blks_hit",
	"shared_blks_read":      "delta_shared_blks_read",
	"shared_blks_dirtied":   "delta_shared_blks_dirtied",
	"shared_blks_written":   "delta_shared_blks_written",
	"local_blks_hit":        "delta_local_blks_hit",
	"local_blks_read":       "delta_local_blks_read",
	"local_blks_dirtied":    "delta_local_blks_dirtied",
	"local_blks_written":    "delta_local_blks_written",
	"temp_blks_read":        "delta_temp_blks_read",
	"temp_blks_written":     "delta_temp_blks_written",
	"wal_records":           "delta_wal_records",
	"wal_fpi":               "delta_wal_fpi",
	"wal_bytes":             "delta_wal_bytes",
	"total_plan_time":       "delta_total_plan_time",
	"shared_blk_read_time":  "delta_shared_blk_read_time",
	"shared_blk_write_time": "delta_shared_blk_write_time",
	"blk_read_time":         "delta_blk_read_time",
	"blk_write_time":        "delta_blk_write_time",
}

type dbmMetric struct {
	Enabled  bool             `toml:"enabled"`
	Interval datakit.Duration `toml:"interval"`
}

type dbmMetricRow struct {
	db             string
	rolname        string
	queryID        string
	querySignature string
	message        string
	metrics        map[string]float64
	deltas         map[string]float64
}

type dbmMetricValueCache struct {
	metrics map[string]float64
}

const sqlGetDbmMetric = `
	SELECT %s
  FROM pg_stat_statements as pg_stat_statements
  LEFT JOIN pg_roles
         ON pg_stat_statements.userid = pg_roles.oid
  LEFT JOIN pg_database
         ON pg_stat_statements.dbid = pg_database.oid
  WHERE query != '<insufficient privilege>'
  AND query NOT LIKE '/* DKIGNORE */%%'
`

const sqlDbmPgSettings = `
SELECT name, setting FROM pg_settings WHERE name IN (
	'pg_stat_statements.max',
	'track_activity_query_size',
	'track_io_timing'
)
`

func (ipt *Input) loadDbmPgSettings() error {
	ctx, cancel := context.WithTimeout(context.Background(), ipt.Timeout.Duration)
	defer cancel()

	rows, err := ipt.service.Query(ctx, sqlDbmPgSettings)
	if err != nil {
		return fmt.Errorf("query pg_settings for dbm: %w", err)
	}
	defer rows.Close()

	ipt.dbSetting = make(map[string]string)
	for rows.Next() {
		var name, setting string
		if err := rows.Scan(&name, &setting); err != nil {
			return fmt.Errorf("scan pg_settings row: %w", err)
		}
		if name != "" {
			ipt.dbSetting[name] = setting
		}
	}
	return nil
}

func (ipt *Input) collectDbmMetricWithRows(ptsTime time.Time) ([]*point.Point, []dbmMetricRow, error) {
	rows, err := ipt.collectDbmMetricRows()
	if err != nil {
		return nil, nil, err
	}

	metricPts, reportRows := ipt.buildDbmMetricPoints(rows, ptsTime)
	return metricPts, reportRows, nil
}

func (ipt *Input) collectDbmMetricRows() ([]dbmMetricRow, error) {
	cache := ipt.dbmMetricCache
	if cache == nil {
		availabeColumns, err := ipt.getPgStatStatementsColumns()
		if err != nil {
			return nil, fmt.Errorf("get pg_stat_statements columns failed: %w", err)
		}

		for column := range statStatementsRequiredColumns {
			if !availabeColumns[column] {
				return nil, fmt.Errorf("column %s not found", column)
			}
		}

		allColumns := []string{}
		for m := range statStatementsMetricsNoTimingColumns {
			allColumns = append(allColumns, m)
		}
		for m := range statStatementsTagColumns {
			allColumns = append(allColumns, m)
		}
		for m := range statStatementsOptionalColumns {
			allColumns = append(allColumns, m)
		}

		if v, ok := ipt.dbSetting["track_io_timing"]; ok && v == "on" {
			for m := range statStatementsTimingColumns {
				allColumns = append(allColumns, m)
			}
			for m := range statStatementsTimingColumnsLT17 {
				allColumns = append(allColumns, m)
			}
		}

		queryColumns := []string{}
		for _, column := range allColumns {
			if availabeColumns[column] {
				queryColumns = append(queryColumns, column)
			}
		}

		sql := fmt.Sprintf(sqlGetDbmMetric, strings.Join(queryColumns, ", "))

		if len(ipt.IgnoredDatabases) > 0 {
			sql += fmt.Sprintf(" AND pg_database.datname NOT IN ('%s')", strings.Join(ipt.IgnoredDatabases, "','"))
		} else if len(ipt.Databases) > 0 {
			sql += fmt.Sprintf(" AND pg_database.datname IN ('%s')", strings.Join(ipt.Databases, "','"))
		}

		cache = &queryCacheItem{
			q:               sql,
			measurementInfo: dbmMetricMeasurement{}.Info(),
		}
		ipt.dbmMetricCache = cache
		l.Infof("Query for metric [%s]: %s", cache.measurementInfo.Name, sql)
	}

	sqlStart := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), ipt.Timeout.Duration)
	defer cancel()

	rows, err := ipt.service.Query(ctx, cache.q)
	if err != nil {
		return nil, fmt.Errorf("query dbm metric failed: %w", err)
	}
	defer rows.Close() //nolint:errcheck

	sqlQueryCostSummary.WithLabelValues(dbmMetricMeasurementInfo.Name, dbmMetricMeasurementInfo.Name).Observe(time.Since(sqlStart).Seconds())

	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("get dbm metric columns failed: %w", err)
	}

	dbmRows := make([]dbmMetricRow, 0, 256)
	o := obfuscate.NewObfuscator(obfuscate.Config{
		SQL: obfuscate.SQLConfig{
			DBMS: obfuscate.DBMSPostgres,
		},
	})
	for rows.Next() {
		columnMap, err := ipt.service.GetColumnMap(rows, columns)
		if err != nil {
			return nil, fmt.Errorf("scan dbm metric row failed: %w", err)
		}

		row := dbmMetricRow{
			metrics: make(map[string]float64),
			deltas:  make(map[string]float64),
		}

		for col, v := range columnMap {
			if v == nil || *v == nil {
				continue
			}

			switch col {
			case "query":
				query := interfaceToString(*v)
				if query == "" {
					continue
				}
				obfResult, err := o.ObfuscateSQLString(query)
				if err != nil {
					l.Warnf("obfuscate dbm metric sql failed: %s, query: %s", err.Error(), query)
					continue
				}
				row.message = obfResult.Query
			case "datname":
				row.db = interfaceToString(*v)
			case "rolname":
				row.rolname = interfaceToString(*v)
			case "queryid":
				row.queryID = interfaceToString(*v)
			default:
				row.metrics[col] = cast.ToFloat64(*v)
			}
		}

		row.querySignature = generateQuerySignature(row.db, row.rolname, row.message)

		dbmRows = append(dbmRows, row)
	}

	dbmRows = mergeDuplicateDbmMetricRows(dbmRows)

	return dbmRows, nil
}

func generateQuerySignature(db, rolname, query string) string {
	h := xxhash.New()
	_, _ = h.WriteString(db)
	_, _ = h.WriteString(":")
	_, _ = h.WriteString(rolname)
	_, _ = h.WriteString(":")
	_, _ = h.WriteString(query)

	return fmt.Sprintf("%016x", h.Sum64())
}

func (ipt *Input) buildDbmMetricPoints(rows []dbmMetricRow, ptsTime time.Time) ([]*point.Point, []dbmMetricRow) {
	if ipt.dbmMetricValueCache == nil {
		ipt.dbmMetricValueCache = map[string]*dbmMetricValueCache{}
	}

	opts := append(point.DefaultMetricOptions(), point.WithTime(ptsTime))
	pts := make([]*point.Point, 0, len(rows)+1)
	reportRows := make([]dbmMetricRow, 0, len(rows))
	var totalCalls int64

	for _, row := range rows {
		totalCalls += int64(row.metrics["calls"])

		old, ok := ipt.dbmMetricValueCache[row.querySignature]
		ipt.dbmMetricValueCache[row.querySignature] = &dbmMetricValueCache{
			metrics: copyMetricMap(row.metrics),
		}
		if !ok {
			continue
		}

		valid := true
		for base, deltaName := range statStatementsDeltaColumns {
			curr, okCurr := row.metrics[base]
			prev, okPrev := old.metrics[base]
			if !okCurr || !okPrev {
				continue
			}
			delta := curr - prev
			if delta < 0 {
				valid = false
				break
			}
			row.deltas[deltaName] = delta
		}
		if !valid {
			continue
		}

		deltaCalls, ok := row.deltas["delta_calls"]
		if !ok || deltaCalls <= 0 {
			continue
		}

		kvs := ipt.getKVs()
		if row.db != "" {
			kvs = kvs.AddTag("db", row.db)
		}
		if row.rolname != "" {
			kvs = kvs.AddTag("rolname", row.rolname)
		}
		if row.queryID != "" {
			kvs = kvs.AddTag("queryid", row.queryID)
		}
		kvs = kvs.AddTag("query_signature", row.querySignature)

		for k, v := range row.metrics {
			kvs = kvs.Set(k, v)
		}
		for k, v := range row.deltas {
			kvs = kvs.Set(k, v)
		}

		if deltaExecTime, ok := row.deltas["delta_total_exec_time"]; ok && deltaCalls > 0 {
			kvs = kvs.Set("avg_total_exec_time", deltaExecTime/deltaCalls)
		}

		if deltaPlanTime, ok := row.deltas["delta_total_plan_time"]; ok && deltaCalls > 0 {
			kvs = kvs.Set("avg_total_plan_time", deltaPlanTime/deltaCalls)
		}

		pts = append(pts, point.NewPoint(dbmMetricMeasurementInfo.Name, kvs, opts...))
		reportRows = append(reportRows, row)
	}

	if totalCalls > 0 {
		kvs := ipt.getKVs()
		kvs = kvs.Set("total_calls", totalCalls)
		if !ipt.dbmMetricTotalTime.IsZero() && totalCalls >= ipt.dbmMetricTotalCalls {
			deltaTotalCalls := totalCalls - ipt.dbmMetricTotalCalls
			kvs = kvs.Set("delta_total_calls", deltaTotalCalls)

			if elapsed := ptsTime.Sub(ipt.dbmMetricTotalTime).Seconds(); elapsed > 0 {
				kvs = kvs.Set("dbm_qps", float64(deltaTotalCalls)/elapsed)
			}
		}
		pts = append(pts, point.NewPoint(dbmMetricMeasurementInfo.Name, kvs, opts...))
		ipt.dbmMetricTotalCalls = totalCalls
		ipt.dbmMetricTotalTime = ptsTime
	}

	return pts, reportRows
}

func interfaceToString(v interface{}) string {
	switch tv := v.(type) {
	case []uint8:
		return string(tv)
	case string:
		return tv
	default:
		return cast.ToString(tv)
	}
}

func copyMetricMap(src map[string]float64) map[string]float64 {
	dst := make(map[string]float64, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}

func mergeDuplicateDbmMetricRows(rows []dbmMetricRow) []dbmMetricRow {
	if len(rows) <= 1 {
		return rows
	}

	keyRows := make(map[string]dbmMetricRow, len(rows))
	for _, row := range rows {
		if keyRow, ok := keyRows[row.querySignature]; ok {
			for k, v := range row.metrics {
				keyRow.metrics[k] += v
			}
			keyRows[row.querySignature] = keyRow
		} else {
			row.metrics = copyMetricMap(row.metrics)
			keyRows[row.querySignature] = row
		}
	}

	merged := make([]dbmMetricRow, 0, len(keyRows))
	for _, row := range keyRows {
		merged = append(merged, row)
	}

	return merged
}

const sqlGetPgStatStatementsColumns = `
	SELECT * 
  FROM pg_stat_statements as pg_stat_statements
  LEFT JOIN pg_roles
         ON pg_stat_statements.userid = pg_roles.oid
  LEFT JOIN pg_database
         ON pg_stat_statements.dbid = pg_database.oid
  WHERE query != '<insufficient privilege>'
  AND query NOT LIKE '/* DKIGNORE */%'
	LIMIT 0
`

func (ipt *Input) getPgStatStatementsColumns() (map[string]bool, error) {
	if ipt.statColumnCache != nil {
		return ipt.statColumnCache, nil
	}

	columns, err := ipt.getSQLColumns(sqlGetPgStatStatementsColumns)
	if err != nil {
		return nil, fmt.Errorf("query pg_stat_statements failed: %w", err)
	}

	ipt.statColumnCache = make(map[string]bool)
	for _, column := range columns {
		ipt.statColumnCache[column] = true
	}

	return ipt.statColumnCache, nil
}
