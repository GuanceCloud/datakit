// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package postgresql

import (
	// nolint:gosec
	"fmt"
	"strings"

	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/util"
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

type dbmMetric struct {
	Enabled bool `toml:"enabled"`
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

func (ipt *Input) getDbmMetric() error {
	cache, ok := ipt.metricQueryCache[DBMMetric]
	if !ok {
		availabeColumns, err := ipt.getPgStatStatementsColumns()
		if err != nil {
			return fmt.Errorf("get pg_stat_statements columns failed: %w", err)
		}

		for column := range statStatementsRequiredColumns {
			if !availabeColumns[column] {
				return fmt.Errorf("column %s not found", column)
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
		ipt.metricQueryCache[DBMMetric] = cache
		l.Infof("Query for metric [%s]: %s", cache.measurementInfo.Name, sql)
	}

	cache.ptsTime = ipt.ptsTime
	var info interface{} = "info"
	if points, err := ipt.getQueryPoints(cache, func(m map[string]*interface{}) {
		query := m["query"]
		if query != nil {
			text, ok := (*query).(string)
			if ok {
				text = util.ObfuscateSQL(text)
				var message interface{} = text
				m["message"] = &message
				delete(m, "query")
				var signature interface{} = util.ComputeSQLSignature(text)
				m["query_signature"] = &signature
				m["status"] = &info
				m["db"] = m["datname"]
				delete(m, "datname")
			}
		}
	}); err != nil {
		return fmt.Errorf("getQueryPoints error: %w", err)
	} else {
		ipt.collectCache[point.Logging] = append(ipt.collectCache[point.Logging], points...)
	}

	return nil
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
