// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package mysql

import (
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

type dbmMetric struct {
	Enabled  bool             `toml:"enabled"`
	Interval datakit.Duration `toml:"interval"`
	Limit    int              `toml:"limit"`
}

type dbmStateMeasurement struct {
	name     string
	tags     map[string]string
	fields   map[string]interface{}
	election bool
}

// Point implement MeasurementV2.
func (m *dbmStateMeasurement) Point() *point.Point {
	opts := point.DefaultMetricOptions()

	if m.election {
		opts = append(opts, point.WithExtraTags(datakit.GlobalElectionTags()))
	}

	return point.NewPoint(m.name,
		append(point.NewTags(m.tags), point.NewKVs(m.fields)...),
		opts...)
}

func (m *dbmStateMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Desc: "Record the number of executions of the query statement, wait time, lock time, and the number of rows queried.",
		Name: metricNameMySQLDbmMetric,
		Cat:  point.Metric,
		Fields: map[string]interface{}{
			// Total values (cumulative values from MySQL)
			"count_star": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The total count of executed queries per normalized query and schema (cumulative).",
			},
			"sum_timer_wait": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The total query execution time(nanosecond) per normalized query and schema (cumulative).",
			},
			"sum_lock_time": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The total time(nanosecond) spent waiting on locks per normalized query and schema (cumulative).",
			},
			"sum_errors": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The total count of queries run with an error per normalized query and schema (cumulative).",
			},
			"sum_rows_affected": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The number of rows mutated per normalized query and schema (cumulative).",
			},
			"sum_rows_sent": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The number of rows sent per normalized query and schema (cumulative).",
			},
			"sum_rows_examined": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The number of rows examined per normalized query and schema (cumulative).",
			},
			"sum_select_scan": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The total count of full table scans on the first table per normalized query and schema (cumulative).",
			},
			"sum_select_full_join": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The total count of full table scans on a joined table per normalized query and schema (cumulative).",
			},
			"sum_no_index_used": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The total count of queries which do not use an index per normalized query and schema (cumulative).",
			},
			"sum_no_good_index_used": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The total count of queries which used a sub-optimal index per normalized query and schema (cumulative).",
			},
			// Delta values (change between collection intervals)
			"delta_count_star": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The change in count of executed queries per normalized query and schema between collection intervals.",
			},
			"delta_timer_wait": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The change in query execution time(nanosecond) per normalized query and schema between collection intervals.",
			},
			"delta_lock_time": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The change in time(nanosecond) spent waiting on locks per normalized query and schema between collection intervals.",
			},
			"delta_errors": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The change in count of queries run with an error per normalized query and schema between collection intervals.",
			},
			"delta_rows_affected": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The change in number of rows mutated per normalized query and schema between collection intervals.",
			},
			"delta_rows_sent": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The change in number of rows sent per normalized query and schema between collection intervals.",
			},
			"delta_rows_examined": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The change in number of rows examined per normalized query and schema between collection intervals.",
			},
			"delta_select_scan": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The change in count of full table scans on the first table per normalized query and schema between collection intervals.",
			},
			"delta_select_full_join": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The change in count of full table scans on a joined table per normalized query and schema between collection intervals.",
			},
			"delta_no_index_used": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The change in count of queries which do not use an index per normalized query and schema between collection intervals.",
			},
			"delta_no_good_index_used": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The change in count of queries which used a sub-optimal index per normalized query and schema between collection intervals.",
			},
			// Average values (calculated from delta values)
			"avg_timer_wait": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.DurationNS,
				//nolint:lll
				Desc: "The average query execution time (nanoseconds) per query execution during the collection interval (calculated from delta_timer_wait / delta_count_star).",
			},
		},
		Tags: map[string]interface{}{
			"server":            &inputs.TagInfo{Desc: "The address of the server. The value is `host:port`"},
			"host":              &inputs.TagInfo{Desc: "The server host address"},
			"database_instance": &inputs.TagInfo{Desc: "MySQL instance identifier from configured tag or @@server_uuid."},
			"digest":            &inputs.TagInfo{Desc: "The digest hash value computed from the original normalized statement. "},
			"query_signature":   &inputs.TagInfo{Desc: "The hash value computed from digest_text"},
			"schema_name":       &inputs.TagInfo{Desc: "The schema name."},
		},
	}
}

type dbmRow struct {
	schemaName     string
	digest         string
	digestText     string
	querySignature string

	// Total values (cumulative values from MySQL)
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

	// Delta values (change between collection intervals)
	deltaCountStar       uint64
	deltaTimerWait       uint64
	deltaLockTime        uint64
	deltaErrors          uint64
	deltaRowsAffected    uint64
	deltaRowsSent        uint64
	deltaRowsExamined    uint64
	deltaSelectScan      uint64
	deltaSelectFullJoin  uint64
	deltaNoIndexUsed     uint64
	deltaNoGoodIndexUsed uint64
}

// dbmMetricCache stores cumulative values needed to compute derivatives.
type dbmMetricCache struct {
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
			keyRows[keyStr] = keyRow
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

// getMetricRows computes deltas against the previous full snapshot and returns the next snapshot.
func getMetricRows(dbmRows []dbmRow, prevSnapshot map[string]dbmMetricCache) ([]dbmRow, map[string]dbmMetricCache) {
	metricRows := []dbmRow{}
	nextSnapshot := make(map[string]dbmMetricCache, len(dbmRows))

	for _, row := range dbmRows {
		rowKey := getRowKey(row.schemaName, row.querySignature)
		nextSnapshot[rowKey] = dbmMetricCache{
			countStar:          row.countStar,
			sumTimerWait:       row.sumTimerWait,
			sumLockTime:        row.sumLockTime,
			sumErrors:          row.sumErrors,
			sumRowsAffected:    row.sumRowsAffected,
			sumRowsSent:        row.sumRowsSent,
			sumRowsExamined:    row.sumRowsExamined,
			sumSelectScan:      row.sumSelectScan,
			sumSelectFullJoin:  row.sumSelectFullJoin,
			sumNoIndexUsed:     row.sumNoIndexUsed,
			sumNoGoodIndexUsed: row.sumNoGoodIndexUsed,
		}

		old, ok := prevSnapshot[rowKey]
		if !ok {
			// First time seeing this query in the previous snapshot: baseline only.
			continue
		}

		// Calculate derivatives for all fields.
		// Guard against counter reset/wraparound to avoid uint64 underflow.
		if row.countStar < old.countStar {
			continue
		}
		diffCountStar := row.countStar - old.countStar
		if diffCountStar <= 0 {
			continue
		}

		// Calculate and store delta values directly in the row
		row.deltaCountStar = diffCountStar
		if row.sumTimerWait >= old.sumTimerWait {
			row.deltaTimerWait = (row.sumTimerWait - old.sumTimerWait) / 1000 // nanosecond
		} else {
			continue
		}
		if row.sumLockTime >= old.sumLockTime {
			row.deltaLockTime = (row.sumLockTime - old.sumLockTime) / 1000 // nanosecond
		} else {
			continue
		}
		if row.sumErrors >= old.sumErrors {
			row.deltaErrors = row.sumErrors - old.sumErrors
		} else {
			continue
		}
		if row.sumRowsAffected >= old.sumRowsAffected {
			row.deltaRowsAffected = row.sumRowsAffected - old.sumRowsAffected
		} else {
			continue
		}
		if row.sumRowsSent >= old.sumRowsSent {
			row.deltaRowsSent = row.sumRowsSent - old.sumRowsSent
		} else {
			continue
		}
		if row.sumRowsExamined >= old.sumRowsExamined {
			row.deltaRowsExamined = row.sumRowsExamined - old.sumRowsExamined
		} else {
			continue
		}
		if row.sumSelectScan >= old.sumSelectScan {
			row.deltaSelectScan = row.sumSelectScan - old.sumSelectScan
		} else {
			continue
		}
		if row.sumSelectFullJoin >= old.sumSelectFullJoin {
			row.deltaSelectFullJoin = row.sumSelectFullJoin - old.sumSelectFullJoin
		} else {
			continue
		}
		if row.sumNoIndexUsed >= old.sumNoIndexUsed {
			row.deltaNoIndexUsed = row.sumNoIndexUsed - old.sumNoIndexUsed
		} else {
			continue
		}
		if row.sumNoGoodIndexUsed >= old.sumNoGoodIndexUsed {
			row.deltaNoGoodIndexUsed = row.sumNoGoodIndexUsed - old.sumNoGoodIndexUsed
		} else {
			continue
		}

		// Report both total (from query) and delta (calculated) values
		metricRows = append(metricRows, row)
	}

	return metricRows, nextSnapshot
}

func (ipt *Input) buildMysqlDbmMetric(rows []dbmRow, ptsTime time.Time) ([]*point.Point, error) {
	var pts []*point.Point
	opts := append(point.DefaultMetricOptions(), point.WithTime(ptsTime))

	for _, row := range rows {
		kvs := ipt.getKVs()

		// tags
		if len(row.schemaName) > 0 {
			kvs = kvs.AddTag("schema_name", row.schemaName)
		}
		if len(row.digest) > 0 {
			kvs = kvs.AddTag("digest", row.digest)
		}
		if len(row.querySignature) > 0 {
			kvs = kvs.AddTag("query_signature", row.querySignature)
		}

		// Fields - report both total and delta values
		// Total values (cumulative values from MySQL)
		kvs = kvs.Set("count_star", row.countStar)
		kvs = kvs.Set("sum_timer_wait", row.sumTimerWait)
		kvs = kvs.Set("sum_lock_time", row.sumLockTime)
		kvs = kvs.Set("sum_errors", row.sumErrors)
		kvs = kvs.Set("sum_rows_affected", row.sumRowsAffected)
		kvs = kvs.Set("sum_rows_sent", row.sumRowsSent)
		kvs = kvs.Set("sum_rows_examined", row.sumRowsExamined)
		kvs = kvs.Set("sum_select_scan", row.sumSelectScan)
		kvs = kvs.Set("sum_select_full_join", row.sumSelectFullJoin)
		kvs = kvs.Set("sum_no_index_used", row.sumNoIndexUsed)
		kvs = kvs.Set("sum_no_good_index_used", row.sumNoGoodIndexUsed)

		// Delta values (change between collection intervals)
		kvs = kvs.Set("delta_count_star", row.deltaCountStar)
		kvs = kvs.Set("delta_timer_wait", row.deltaTimerWait)
		kvs = kvs.Set("delta_lock_time", row.deltaLockTime)
		kvs = kvs.Set("delta_errors", row.deltaErrors)
		kvs = kvs.Set("delta_rows_affected", row.deltaRowsAffected)
		kvs = kvs.Set("delta_rows_sent", row.deltaRowsSent)
		kvs = kvs.Set("delta_rows_examined", row.deltaRowsExamined)
		kvs = kvs.Set("delta_select_scan", row.deltaSelectScan)
		kvs = kvs.Set("delta_select_full_join", row.deltaSelectFullJoin)
		kvs = kvs.Set("delta_no_index_used", row.deltaNoIndexUsed)
		kvs = kvs.Set("delta_no_good_index_used", row.deltaNoGoodIndexUsed)

		// Calculate average timer wait per execution
		if row.deltaCountStar > 0 {
			kvs = kvs.Set("avg_timer_wait", row.deltaTimerWait/row.deltaCountStar)
		}

		pts = append(pts, point.NewPoint(metricNameMySQLDbmMetric, kvs, opts...))
	}

	return pts, nil
}
