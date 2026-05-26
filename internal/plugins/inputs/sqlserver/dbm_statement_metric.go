// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package sqlserver

import (
	"context"
	"encoding/hex"
	"fmt"
	"sort"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"github.com/cespare/xxhash/v2"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/metrics"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

const (
	metricNameSQLServerDbmMetric = "sqlserver_dbm_metric"
)

type dbmStatementRow struct {
	databaseName         string
	planHandle           string
	queryHash            string
	queryPlanHash        string
	schemaName           string
	procedureName        string
	statementText        string
	isEncrypted          bool
	sqlserverHost        string
	statementStartOffset int64
	statementEndOffset   int64

	executionCount               int64
	totalWorkerTime              int64
	totalPhysicalReads           int64
	totalLogicalWrites           int64
	totalLogicalReads            int64
	totalClrTime                 int64
	totalElapsedTime             int64
	totalRows                    int64
	totalDop                     int64
	totalGrantKb                 int64
	totalUsedGrantKb             int64
	totalIdealGrantKb            int64
	totalReservedThreads         int64
	totalUsedThreads             int64
	totalColumnstoreSegmentReads int64
	totalColumnstoreSegmentSkips int64
	totalSpills                  int64

	querySignature string // Signature to link object and metric

	// Delta numeric fields (change between collection intervals)
	deltaElapsedTime             int64
	deltaExecutionCount          int64
	deltaWorkerTime              int64
	deltaPhysicalReads           int64
	deltaLogicalWrites           int64
	deltaLogicalReads            int64
	deltaClrTime                 int64
	deltaRows                    int64
	deltaDop                     int64
	deltaGrantKb                 int64
	deltaUsedGrantKb             int64
	deltaIdealGrantKb            int64
	deltaReservedThreads         int64
	deltaUsedThreads             int64
	deltaColumnstoreSegmentReads int64
	deltaColumnstoreSegmentSkips int64
	deltaSpills                  int64
}

func (ipt *Input) collectDbmMetric(ctx context.Context, ptsTime time.Time) ([]*dbmStatementRow, error) {
	if ipt.Dbm == nil || ipt.Dbm.Metric == nil || !ipt.Dbm.Metric.Enabled {
		return nil, nil
	}

	start := time.Now()

	// Calculate lookback window: use configured value if set, otherwise use 300 seconds
	lookbackWindowSeconds := ipt.Dbm.Metric.LookbackWindow
	if lookbackWindowSeconds <= 0 {
		lookbackWindowSeconds = 300
	}

	// Build query dynamically based on available columns
	query, err := ipt.buildStatementsQuery(ctx, lookbackWindowSeconds, ipt.Dbm.Metric.DmExecQueryStatsRowLimit)
	if err != nil {
		return nil, fmt.Errorf("failed to build metric query: %w", err)
	}

	// Measure database query time
	queryStart := time.Now()
	rows, err := ipt.db.QueryContext(ctx, query)
	dbmSQLQueryDuration.WithLabelValues("statements", "query").Observe(time.Since(queryStart).Seconds())
	if err != nil {
		return nil, fmt.Errorf("failed to query statement metrics: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			l.Errorf("failed to close rows: %v", err)
		}
	}()

	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("failed to get columns: %w", err)
	}

	var statementRows []*dbmStatementRow
	for rows.Next() {
		columnMap, err := GetColumnMap(rows, columns)
		if err != nil {
			l.Errorf("failed to get column map: %v", err)
			continue
		}

		row, err := ipt.buildDbmStatementRow(columnMap)
		if err != nil {
			l.Errorf("build dbm statement row: %v", err)
			continue
		}
		if row == nil {
			continue
		}

		statementRows = append(statementRows, row)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	if len(statementRows) == 0 {
		return nil, nil
	}

	// Process statements: merge by query_hash, compute derivatives, and update cache
	resultRows := ipt.processStatements(statementRows)

	if len(resultRows) == 0 {
		return nil, nil
	}

	// Sort by delta elapsed time and take top N
	sort.Slice(resultRows, func(i, j int) bool {
		return resultRows[i].deltaElapsedTime > resultRows[j].deltaElapsedTime
	})

	maxQueries := ipt.Dbm.Metric.MaxQueries
	if len(resultRows) > maxQueries {
		resultRows = resultRows[:maxQueries]
	}

	// Build points from processed rows (report derivative values, not cumulative)
	pts := ipt.buildStatementPoints(resultRows, ptsTime)
	if len(pts) > 0 {
		if err := ipt.feeder.Feed(point.Metric, pts,
			dkio.WithCollectCost(time.Since(start)),
			dkio.WithElection(ipt.Election),
			dkio.WithSource(dbmFeedName),
			dkio.WithMeasurement(inputs.GetOverrideMeasurement(ipt.MeasurementVersion, measurementSQLServer)),
		); err != nil {
			ipt.feeder.FeedLastError(err.Error(),
				metrics.WithLastErrorInput(inputName),
				metrics.WithLastErrorCategory(point.Metric),
			)
			l.Errorf("feed dbm metric failed: %s", err.Error())
		}
	}

	// Build set of top querySignature + queryPlanHash combinations
	topKeys := make(map[string]bool)
	for _, row := range resultRows {
		key := row.querySignature + ":" + row.queryPlanHash
		topKeys[key] = true
	}

	// Filter original rows to return only those matching top query+plan combinations
	var filteredRows []*dbmStatementRow
	for _, row := range statementRows {
		key := row.querySignature + ":" + row.queryPlanHash
		if topKeys[key] {
			filteredRows = append(filteredRows, row)
		}
	}

	return filteredRows, nil
}

// generateQuerySignature generates a unique signature for a SQL statement.
func generateQuerySignature(databaseName, procedureName, queryHash string) string {
	h := xxhash.New()
	_, _ = h.WriteString(databaseName)
	_, _ = h.WriteString(":")
	_, _ = h.WriteString(procedureName)
	_, _ = h.WriteString(":")
	_, _ = h.WriteString(queryHash)

	return fmt.Sprintf("%016x", h.Sum64())
}

// dbmMetricCache stores cumulative values needed to compute derivatives.
type dbmMetricCache struct {
	totalElapsedTime             int64
	executionCount               int64
	totalWorkerTime              int64
	totalPhysicalReads           int64
	totalLogicalWrites           int64
	totalLogicalReads            int64
	totalClrTime                 int64
	totalRows                    int64
	totalDop                     int64
	totalGrantKb                 int64
	totalUsedGrantKb             int64
	totalIdealGrantKb            int64
	totalReservedThreads         int64
	totalUsedThreads             int64
	totalColumnstoreSegmentReads int64
	totalColumnstoreSegmentSkips int64
	totalSpills                  int64
}

func (ipt *Input) processStatements(currentRows []*dbmStatementRow) []*dbmStatementRow {
	// SQL query already groups by query_hash, query_plan_hash, so each row is already unique
	// No need to aggregate again, just compute derivatives and update cache
	newCache := make(map[string]*dbmMetricCache)
	var resultRows []*dbmStatementRow

	for _, row := range currentRows {
		// Use querySignature + queryPlanHash as cache key to distinguish different execution plans
		key := row.querySignature + ":" + row.queryPlanHash

		// Store all cumulative values in cache
		newCache[key] = &dbmMetricCache{
			totalElapsedTime:             row.totalElapsedTime,
			executionCount:               row.executionCount,
			totalWorkerTime:              row.totalWorkerTime,
			totalPhysicalReads:           row.totalPhysicalReads,
			totalLogicalWrites:           row.totalLogicalWrites,
			totalLogicalReads:            row.totalLogicalReads,
			totalClrTime:                 row.totalClrTime,
			totalRows:                    row.totalRows,
			totalDop:                     row.totalDop,
			totalGrantKb:                 row.totalGrantKb,
			totalUsedGrantKb:             row.totalUsedGrantKb,
			totalIdealGrantKb:            row.totalIdealGrantKb,
			totalReservedThreads:         row.totalReservedThreads,
			totalUsedThreads:             row.totalUsedThreads,
			totalColumnstoreSegmentReads: row.totalColumnstoreSegmentReads,
			totalColumnstoreSegmentSkips: row.totalColumnstoreSegmentSkips,
			totalSpills:                  row.totalSpills,
		}

		old, ok := ipt.dbmMetricCache[key]
		if !ok {
			// first time to see this sql, only set baseline, not report metric
			continue
		}

		// Calculate derivatives for all fields
		diffExecCount := row.executionCount - old.executionCount
		if diffExecCount <= 0 {
			continue
		}

		// Calculate and store delta values directly in the row、
		if row.deltaElapsedTime = row.totalElapsedTime - old.totalElapsedTime; row.deltaElapsedTime < 0 {
			continue
		}
		if row.deltaExecutionCount = diffExecCount; row.deltaExecutionCount < 0 {
			continue
		}
		if row.deltaWorkerTime = row.totalWorkerTime - old.totalWorkerTime; row.deltaWorkerTime < 0 {
			continue
		}
		if row.deltaPhysicalReads = row.totalPhysicalReads - old.totalPhysicalReads; row.deltaPhysicalReads < 0 {
			continue
		}
		if row.deltaLogicalWrites = row.totalLogicalWrites - old.totalLogicalWrites; row.deltaLogicalWrites < 0 {
			continue
		}
		if row.deltaLogicalReads = row.totalLogicalReads - old.totalLogicalReads; row.deltaLogicalReads < 0 {
			continue
		}
		if row.deltaClrTime = row.totalClrTime - old.totalClrTime; row.deltaClrTime < 0 {
			continue
		}
		if row.deltaRows = row.totalRows - old.totalRows; row.deltaRows < 0 {
			continue
		}
		if row.deltaDop = row.totalDop - old.totalDop; row.deltaDop < 0 {
			continue
		}
		if row.deltaGrantKb = row.totalGrantKb - old.totalGrantKb; row.deltaGrantKb < 0 {
			continue
		}
		if row.deltaUsedGrantKb = row.totalUsedGrantKb - old.totalUsedGrantKb; row.deltaUsedGrantKb < 0 {
			continue
		}
		if row.deltaIdealGrantKb = row.totalIdealGrantKb - old.totalIdealGrantKb; row.deltaIdealGrantKb < 0 {
			continue
		}
		if row.deltaReservedThreads = row.totalReservedThreads - old.totalReservedThreads; row.deltaReservedThreads < 0 {
			continue
		}
		if row.deltaUsedThreads = row.totalUsedThreads - old.totalUsedThreads; row.deltaUsedThreads < 0 {
			continue
		}
		if row.deltaColumnstoreSegmentReads = row.totalColumnstoreSegmentReads - old.totalColumnstoreSegmentReads; row.deltaColumnstoreSegmentReads < 0 {
			continue
		}
		if row.deltaColumnstoreSegmentSkips = row.totalColumnstoreSegmentSkips - old.totalColumnstoreSegmentSkips; row.deltaColumnstoreSegmentSkips < 0 {
			continue
		}
		if row.deltaSpills = row.totalSpills - old.totalSpills; row.deltaSpills < 0 {
			continue
		}

		resultRows = append(resultRows, row)
	}

	// cache result rows
	ipt.dbmMetricCache = newCache

	return resultRows
}

func (ipt *Input) buildDbmStatementRow(columnMap map[string]*interface{}) (*dbmStatementRow, error) {
	row := &dbmStatementRow{}

	// Extract basic fields
	row.databaseName = getStringField(columnMap, "database_name")
	row.sqlserverHost = getStringField(columnMap, "sqlserver_host")

	// Extract hash values (convert binary to hex string with 0x prefix)
	planHandle := getBytesField(columnMap, "plan_handle")
	queryHash := getBytesField(columnMap, "query_hash")
	queryPlanHash := getBytesField(columnMap, "query_plan_hash")

	if len(planHandle) > 0 {
		row.planHandle = "0x" + hex.EncodeToString(planHandle)
	}
	if len(queryHash) > 0 {
		row.queryHash = "0x" + hex.EncodeToString(queryHash)
	}
	if len(queryPlanHash) > 0 {
		row.queryPlanHash = "0x" + hex.EncodeToString(queryPlanHash)
	}

	row.statementText = getStringField(columnMap, "statement_text")
	row.isEncrypted = getBoolField(columnMap, "is_encrypted")

	// Extract schema name and procedure name first (needed for query signature calculation)
	procedureName := getStringField(columnMap, "procedure_name")
	schemaName := getStringField(columnMap, "schema_name")
	row.schemaName = schemaName
	if procedureName != "" {
		if schemaName != "" {
			row.procedureName = fmt.Sprintf("%s.%s", schemaName, procedureName)
		} else {
			row.procedureName = procedureName
		}
	}

	// Extract statement offsets
	row.statementStartOffset = getInt64Field(columnMap, "statement_start_offset")
	row.statementEndOffset = getInt64Field(columnMap, "statement_end_offset")

	// Generate querySignature
	row.querySignature = generateQuerySignature(row.databaseName, row.procedureName, row.queryHash)

	// Extract cumulative metric values from SQL Server (counters)
	row.executionCount = getInt64Field(columnMap, "execution_count")
	row.totalWorkerTime = getInt64Field(columnMap, "total_worker_time")
	row.totalPhysicalReads = getInt64Field(columnMap, "total_physical_reads")
	row.totalLogicalWrites = getInt64Field(columnMap, "total_logical_writes")
	row.totalLogicalReads = getInt64Field(columnMap, "total_logical_reads")
	row.totalClrTime = getInt64Field(columnMap, "total_clr_time")
	row.totalElapsedTime = getInt64Field(columnMap, "total_elapsed_time")
	row.totalRows = getInt64Field(columnMap, "total_rows")
	row.totalDop = getInt64Field(columnMap, "total_dop")
	row.totalGrantKb = getInt64Field(columnMap, "total_grant_kb")
	row.totalUsedGrantKb = getInt64Field(columnMap, "total_used_grant_kb")
	row.totalIdealGrantKb = getInt64Field(columnMap, "total_ideal_grant_kb")
	row.totalReservedThreads = getInt64Field(columnMap, "total_reserved_threads")
	row.totalUsedThreads = getInt64Field(columnMap, "total_used_threads")
	row.totalColumnstoreSegmentReads = getInt64Field(columnMap, "total_columnstore_segment_reads")
	row.totalColumnstoreSegmentSkips = getInt64Field(columnMap, "total_columnstore_segment_skips")
	row.totalSpills = getInt64Field(columnMap, "total_spills")

	return row, nil
}

// buildStatementPoints builds point.Point from statement rows.
func (ipt *Input) buildStatementPoints(rows []*dbmStatementRow, ptsTime time.Time) []*point.Point {
	var pts []*point.Point
	opts := append(point.DefaultMetricOptions(), point.WithTime(ptsTime))

	for _, row := range rows {
		kvs := ipt.getKVs()

		// Tags
		kvs = kvs.AddTag("database_name", row.databaseName)
		kvs = kvs.AddTag("query_hash", row.queryHash)
		if row.queryPlanHash != "" {
			kvs = kvs.AddTag("query_plan_hash", row.queryPlanHash)
		}
		kvs = kvs.AddTag("sqlserver_host", row.sqlserverHost)
		if row.schemaName != "" {
			kvs = kvs.AddTag("schema_name", row.schemaName)
		}
		if row.procedureName != "" {
			kvs = kvs.AddTag("procedure_name", row.procedureName)
		}
		if row.querySignature != "" {
			kvs = kvs.AddTag("query_signature", row.querySignature)
		}

		// Fields - report both total and delta values
		// Total values (cumulative values from SQL Server)
		kvs = kvs.Set("total_execution_count", row.executionCount)
		kvs = kvs.Set("total_worker_time", row.totalWorkerTime)
		kvs = kvs.Set("total_physical_reads", row.totalPhysicalReads)
		kvs = kvs.Set("total_logical_writes", row.totalLogicalWrites)
		kvs = kvs.Set("total_logical_reads", row.totalLogicalReads)
		kvs = kvs.Set("total_clr_time", row.totalClrTime)
		kvs = kvs.Set("total_elapsed_time", row.totalElapsedTime)
		kvs = kvs.Set("total_rows", row.totalRows)
		kvs = kvs.Set("total_dop", row.totalDop)
		kvs = kvs.Set("total_grant_kb", row.totalGrantKb)
		kvs = kvs.Set("total_used_grant_kb", row.totalUsedGrantKb)
		kvs = kvs.Set("total_ideal_grant_kb", row.totalIdealGrantKb)
		kvs = kvs.Set("total_reserved_threads", row.totalReservedThreads)
		kvs = kvs.Set("total_used_threads", row.totalUsedThreads)
		kvs = kvs.Set("total_columnstore_segment_reads", row.totalColumnstoreSegmentReads)
		kvs = kvs.Set("total_columnstore_segment_skips", row.totalColumnstoreSegmentSkips)
		kvs = kvs.Set("total_spills", row.totalSpills)

		// Delta values (change between collection intervals)
		kvs = kvs.Set("delta_execution_count", row.deltaExecutionCount)
		kvs = kvs.Set("delta_worker_time", row.deltaWorkerTime)
		kvs = kvs.Set("delta_physical_reads", row.deltaPhysicalReads)
		kvs = kvs.Set("delta_logical_writes", row.deltaLogicalWrites)
		kvs = kvs.Set("delta_query_logical_reads", row.deltaLogicalReads)
		kvs = kvs.Set("delta_clr_time", row.deltaClrTime)
		kvs = kvs.Set("delta_elapsed_time", row.deltaElapsedTime)
		// Calculate average elapsed time per execution
		if row.deltaExecutionCount > 0 {
			kvs = kvs.Set("avg_elapsed_time", row.deltaElapsedTime/row.deltaExecutionCount)
		}
		kvs = kvs.Set("delta_rows", row.deltaRows)
		kvs = kvs.Set("delta_dop", row.deltaDop)
		kvs = kvs.Set("delta_grant_kb", row.deltaGrantKb)
		kvs = kvs.Set("delta_used_grant_kb", row.deltaUsedGrantKb)
		kvs = kvs.Set("delta_ideal_grant_kb", row.deltaIdealGrantKb)
		kvs = kvs.Set("delta_reserved_threads", row.deltaReservedThreads)
		kvs = kvs.Set("delta_used_threads", row.deltaUsedThreads)
		kvs = kvs.Set("delta_columnstore_segment_reads", row.deltaColumnstoreSegmentReads)
		kvs = kvs.Set("delta_columnstore_segment_skips", row.deltaColumnstoreSegmentSkips)
		kvs = kvs.Set("delta_spills", row.deltaSpills)

		pts = append(pts, point.NewPoint(metricNameSQLServerDbmMetric, kvs, opts...))
	}

	return pts
}
