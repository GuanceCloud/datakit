package mysql

import (
	"database/sql"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

type dbmMetric struct {
	Enabled bool `toml:"enabled"`
}

type dbmStateMeasurement struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
	ts     time.Time
}

func (m *dbmStateMeasurement) LineProto() (*io.Point, error) {
	return io.MakePoint(m.name, m.tags, m.fields, m.ts)
}

func (m *dbmStateMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Desc: "记录查询语句的执行次数、等待耗时、锁定时间和查询的记录行数等。",
		Name: "mysql_dbm_metric",
		Fields: map[string]interface{}{
			"sum_timer_wait": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The total query execution time(nanosecond) per normalized query and schema.",
			},
			"count_star": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The total count of executed queries per normalized query and schema.",
			},
			"sum_errors": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The total count of queries run with an error per normalized query and schema.",
			},
			"sum_lock_time": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The total time(nanosecond) spent waiting on locks per normalized query and schema.",
			},
			"sum_rows_sent": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The number of rows sent per normalized query and schema.",
			},
			"sum_select_scan": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The total count of full table scans on the first table per normalized query and schema.",
			},
			"sum_no_index_used": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The total count of queries which do not use an index per normalized query and schema.",
			},
			"sum_rows_affected": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The number of rows mutated per normalized query and schema.",
			},
			"sum_rows_examined": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The number of rows examined per normalized query and schema.",
			},
			"sum_select_full_join": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The total count of full table scans on a joined table per normalized query and schema.",
			},
			"sum_no_good_index_used": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The total count of queries which used a sub-optimal index per normalized query and schema.",
			},
			"message": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.String,
				Unit:     inputs.UnknownType,
				Desc:     "The text of the normalized statement digest.",
			},
		},
		Tags: map[string]interface{}{
			"digest":          &inputs.TagInfo{Desc: " The digest hash value computed from the original normalized statement. "},
			"query_signature": &inputs.TagInfo{Desc: " The hash value computed from digest_text"},
			"schema_name":     &inputs.TagInfo{Desc: "The schema name"},
			"server":          &inputs.TagInfo{Desc: " The server address"},
		},
	}
}

type dbmRow struct {
	schemaName     string
	digest         string
	digestText     string
	querySignature string

	countStar          uint64
	sumTimerWait       uint64
	sumLockTime        uint64
	sumErrors          uint64
	sumRowsAffected    uint64
	sumRowsSent        uint64
	sumRowsExamined    uint64
	sumSelectScan      uint64
	sumSelectFullJoin  uint64
	sumNoIndexUsed     uint64
	sumNoGoodIndexUsed uint64
}

func (i *Input) getDbmMetric() ([]inputs.Measurement, error) {
	var (
		err     error
		dbmRows []dbmRow
	)

	if dbmRows, err = getSummaryRows(i.db); err != nil {
		return nil, err
	}

	metricRows, newDbmCache := getMetricRows(dbmRows, &i.dbmCache)

	// save dbm rows
	i.dbmCache = newDbmCache

	var collectCache []inputs.Measurement
	now := time.Now()
	for _, row := range metricRows {
		m := &dbmStateMeasurement{
			name: dbmMetricName,
			tags: map[string]string{
				"service": "mysql_dbm_metric",
			},
			fields: make(map[string]interface{}),
			ts:     now,
		}

		if len(row.digestText) > 0 {
			m.fields["message"] = row.digestText
		}

		if len(row.digest) > 0 {
			m.tags["digest"] = row.digest
		}

		if len(row.schemaName) > 0 {
			m.tags["schema_name"] = row.schemaName
		}

		if len(row.querySignature) > 0 {
			m.tags["query_signature"] = row.querySignature
		}

		m.fields["count_star"] = row.countStar
		m.fields["sum_timer_wait"] = row.sumTimerWait
		m.fields["sum_lock_time"] = row.sumLockTime
		m.fields["sum_errors"] = row.sumErrors
		m.fields["sum_rows_affected"] = row.sumRowsAffected
		m.fields["sum_rows_sent"] = row.sumRowsSent
		m.fields["sum_rows_examined"] = row.sumRowsExamined
		m.fields["sum_select_scan"] = row.sumSelectScan
		m.fields["sum_select_full_join"] = row.sumSelectFullJoin
		m.fields["sum_no_index_used"] = row.sumNoIndexUsed
		m.fields["sum_no_good_index_used"] = row.sumNoGoodIndexUsed

		for key, value := range i.Tags {
			m.tags[key] = value
		}

		collectCache = append(collectCache, m)
	}

	return collectCache, nil
}

// generate row key by shcemaName querySignature.
func getRowKey(schemaName, querySignature string) string {
	return schemaName + querySignature
}

// merge duplicate rows.
func mergeDuplicateRows(rows []dbmRow) []dbmRow {
	keyRows := make(map[string]dbmRow)
	for _, row := range rows {
		keyStr := getRowKey(row.schemaName, row.querySignature)
		if keyRow, ok := keyRows[keyStr]; ok {
			keyRow.countStar += row.countStar
			keyRow.sumTimerWait += row.sumTimerWait
			keyRow.sumLockTime += row.sumLockTime
			keyRow.sumErrors += row.sumErrors
			keyRow.sumRowsAffected += row.sumRowsAffected
			keyRow.sumRowsSent += row.sumRowsSent
			keyRow.sumRowsExamined += row.sumRowsExamined
			keyRow.sumSelectScan += row.sumSelectScan
			keyRow.sumSelectFullJoin += row.sumSelectFullJoin
			keyRow.sumNoIndexUsed += row.sumNoIndexUsed
			keyRow.sumNoGoodIndexUsed += row.sumNoGoodIndexUsed
		} else {
			keyRows[keyStr] = row
		}
	}

	dbmRows := []dbmRow{}

	for _, row := range keyRows {
		dbmRows = append(dbmRows, row)
	}

	return dbmRows
}

// query summary info from "performance_schema.events_statements_summary_by_digest".
func getSummaryRows(db *sql.DB) ([]dbmRow, error) {
	var (
		schemaName sql.NullString
		digest     sql.NullString
		digestText sql.NullString

		countStar          uint64
		sumTimerWait       uint64
		sumLockTime        uint64
		sumErrors          uint64
		sumRowsAffected    uint64
		sumRowsSent        uint64
		sumRowsExamined    uint64
		sumSelectScan      uint64
		sumSelectFullJoin  uint64
		sumNoIndexUsed     uint64
		sumNoGoodIndexUsed uint64
	)

	// query sql
	statementSummarySQL := `
		SELECT schema_name,digest,digest_text,count_star,
		sum_timer_wait,sum_lock_time,sum_errors,sum_rows_affected,
		sum_rows_sent,sum_rows_examined,sum_select_scan,sum_select_full_join,
		sum_no_index_used,sum_no_good_index_used 
		FROM performance_schema.events_statements_summary_by_digest 
		WHERE digest_text NOT LIKE 'EXPLAIN %' OR digest_text IS NULL 
		ORDER BY count_star DESC LIMIT 10000`

	// run query
	rows, err := db.Query(statementSummarySQL)
	if err != nil {
		return nil, err
	}
	defer rows.Close() //nolint:errcheck

	if rows.Err() != nil {
		return nil, rows.Err()
	}

	dbmRows := []dbmRow{}

	for rows.Next() {
		if err = rows.Scan(
			&schemaName,
			&digest,
			&digestText,
			&countStar,
			&sumTimerWait,
			&sumLockTime,
			&sumErrors,
			&sumRowsAffected,
			&sumRowsSent,
			&sumRowsExamined,
			&sumSelectScan,
			&sumSelectFullJoin,
			&sumNoIndexUsed,
			&sumNoGoodIndexUsed); err != nil {
			continue
		}
		var digestStr,
			digestTextStr,
			schemaNameStr string
		if digest.Valid {
			digestStr = digest.String
		}
		if digestText.Valid {
			digestTextStr = digestText.String
		}
		if schemaName.Valid {
			schemaNameStr = schemaName.String
		}

		digestTextStr = obfuscateSQL(digestTextStr)

		querySignature := computeSQLSignature(digestTextStr)

		dbmRowItem := dbmRow{
			digest:             digestStr,
			digestText:         digestTextStr,
			schemaName:         schemaNameStr,
			querySignature:     querySignature,
			countStar:          countStar,
			sumTimerWait:       sumTimerWait,
			sumLockTime:        sumLockTime,
			sumErrors:          sumErrors,
			sumRowsAffected:    sumRowsAffected,
			sumRowsSent:        sumRowsSent,
			sumRowsExamined:    sumRowsExamined,
			sumSelectScan:      sumSelectScan,
			sumSelectFullJoin:  sumSelectFullJoin,
			sumNoIndexUsed:     sumNoIndexUsed,
			sumNoGoodIndexUsed: sumNoGoodIndexUsed,
		}
		dbmRows = append(dbmRows, dbmRowItem)
	}

	dbmRows = mergeDuplicateRows(dbmRows)

	return dbmRows, nil
}

// calculate metric based on previous row identified by row key.
func getMetricRows(dbmRows []dbmRow, dbmCache *map[string]dbmRow) ([]dbmRow, map[string]dbmRow) {
	newDbmCache := make(map[string]dbmRow)
	metricRows := []dbmRow{}
	for _, row := range dbmRows {
		rowKey := getRowKey(row.schemaName, row.querySignature)
		if _, ok := newDbmCache[rowKey]; ok {
			l.Warnf("Duplicate querySignature: %s, using the new one", row.querySignature)
		}
		newDbmCache[rowKey] = row

		if oldRow, ok := (*dbmCache)[rowKey]; ok {
			diffRow := dbmRow{
				digest:         row.digest,
				digestText:     row.digestText,
				schemaName:     row.schemaName,
				querySignature: row.querySignature,
			}
			isChange := false
			if row.countStar >= oldRow.countStar {
				value := row.countStar - oldRow.countStar
				diffRow.countStar = value
				if value > 0 {
					isChange = true
				}
			} else {
				continue
			}
			if row.sumTimerWait >= oldRow.sumTimerWait {
				value := row.sumTimerWait - oldRow.sumTimerWait
				diffRow.sumTimerWait = value / 1000 // nanosecond
				if value > 0 {
					isChange = true
				}
			} else {
				continue
			}
			if row.sumLockTime >= oldRow.sumLockTime {
				value := row.sumLockTime - oldRow.sumLockTime
				diffRow.sumLockTime = value / 1000 // nanosecond
				if value > 0 {
					isChange = true
				}
			} else {
				continue
			}
			if row.sumErrors >= oldRow.sumErrors {
				value := row.sumErrors - oldRow.sumErrors
				diffRow.sumErrors = value
				if value > 0 {
					isChange = true
				}
			} else {
				continue
			}
			if row.sumRowsAffected >= oldRow.sumRowsAffected {
				value := row.sumRowsAffected - oldRow.sumRowsAffected
				diffRow.sumRowsAffected = value
				if value > 0 {
					isChange = true
				}
			} else {
				continue
			}
			if row.sumRowsSent >= oldRow.sumRowsSent {
				value := row.sumRowsSent - oldRow.sumRowsSent
				diffRow.sumRowsSent = value
				if value > 0 {
					isChange = true
				}
			} else {
				continue
			}
			if row.sumRowsExamined >= oldRow.sumRowsExamined {
				value := row.sumRowsExamined - oldRow.sumRowsExamined
				diffRow.sumRowsExamined = value
				if value > 0 {
					isChange = true
				}
			} else {
				continue
			}
			if row.sumSelectScan >= oldRow.sumSelectScan {
				value := row.sumSelectScan - oldRow.sumSelectScan
				diffRow.sumSelectScan = value
				if value > 0 {
					isChange = true
				}
			} else {
				continue
			}
			if row.sumSelectFullJoin >= oldRow.sumSelectFullJoin {
				value := row.sumSelectFullJoin - oldRow.sumSelectFullJoin
				diffRow.sumSelectFullJoin = value
				if value > 0 {
					isChange = true
				}
			} else {
				continue
			}
			if row.sumNoIndexUsed >= oldRow.sumNoIndexUsed {
				value := row.sumNoIndexUsed - oldRow.sumNoIndexUsed
				diffRow.sumNoIndexUsed = value
				if value > 0 {
					isChange = true
				}
			} else {
				continue
			}
			if row.sumNoGoodIndexUsed >= oldRow.sumNoGoodIndexUsed {
				value := row.sumNoGoodIndexUsed - oldRow.sumNoGoodIndexUsed
				diffRow.sumNoGoodIndexUsed = value
				if value > 0 {
					isChange = true
				}
			} else {
				continue
			}

			// No changes, no metric collected
			if !isChange {
				continue
			}

			metricRows = append(metricRows, diffRow)
		} else {
			continue
		}
	}

	return metricRows, newDbmCache
}
