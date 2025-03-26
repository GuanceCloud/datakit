// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package mysql

import (
	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

type dbmMetric struct {
	Enabled bool `toml:"enabled"`
}

type dbmStateMeasurement struct {
	name     string
	tags     map[string]string
	fields   map[string]interface{}
	election bool
}

// Point implement MeasurementV2.
func (m *dbmStateMeasurement) Point() *point.Point {
	opts := point.DefaultLoggingOptions()

	if m.election {
		opts = append(opts, point.WithExtraTags(datakit.GlobalElectionTags()))
	}

	return point.NewPointV2(m.name,
		append(point.NewTags(m.tags), point.NewKVs(m.fields)...),
		opts...)
}

func (m *dbmStateMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Desc: "Record the number of executions of the query statement, wait time, lock time, and the number of rows queried.",
		Name: metricNameMySQLDbmMetric,
		Type: "logging",
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
				Unit:     inputs.UnknownUnit,
				Desc:     "The text of the normalized statement digest.",
			},
		},
		Tags: map[string]interface{}{
			"host":            &inputs.TagInfo{Desc: "The server host address"},
			"service":         &inputs.TagInfo{Desc: "The service name and the value is 'mysql'"},
			"digest":          &inputs.TagInfo{Desc: "The digest hash value computed from the original normalized statement. "},
			"query_signature": &inputs.TagInfo{Desc: "The hash value computed from digest_text"},
			"schema_name":     &inputs.TagInfo{Desc: "The schema name"},
			"server": &inputs.TagInfo{
				Desc: "The server address containing both host and port",
			},
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
