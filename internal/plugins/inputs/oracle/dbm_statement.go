// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package oracle

import (
	"context"
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
	metricNameOracleDbmMetric = "oracle_dbm_metric"
)

type StatementMetricsDB struct {
	StatementMetricsKeyDB
	SQLText       string `db:"SQL_TEXT"`
	SQLTextLength int16  `db:"SQL_TEXT_LENGTH"`
	StatementMetricsMonotonicCountDB
	StatementMetricsGaugeDB
}

type StatementMetricsKeyDB struct {
	ConID                  int    `db:"CON_ID"`
	PDBName                string `db:"PDB_NAME"`
	SQLID                  string `db:"SQL_ID"`
	ForceMatchingSignature string `db:"FORCE_MATCHING_SIGNATURE"`
	PlanHashValue          uint64 `db:"PLAN_HASH_VALUE"`
}

type StatementMetricsMonotonicCountDB struct {
	ParseCalls                 float64 `db:"PARSE_CALLS"`
	DiskReads                  float64 `db:"DISK_READS"`
	DirectWrites               float64 `db:"DIRECT_WRITES"`
	DirectReads                float64 `db:"DIRECT_READS"`
	BufferGets                 float64 `db:"BUFFER_GETS"`
	RowsProcessed              float64 `db:"ROWS_PROCESSED"`
	SerializableAborts         float64 `db:"SERIALIZABLE_ABORTS"`
	Fetches                    float64 `db:"FETCHES"`
	Executions                 float64 `db:"EXECUTIONS"`
	EndOfFetchCount            float64 `db:"END_OF_FETCH_COUNT"`
	Loads                      float64 `db:"LOADS"`
	Invalidations              float64 `db:"INVALIDATIONS"`
	PxServersExecutions        float64 `db:"PX_SERVERS_EXECUTIONS"`
	CPUTime                    float64 `db:"CPU_TIME"`
	ElapsedTime                float64 `db:"ELAPSED_TIME"`
	ApplicationWaitTime        float64 `db:"APPLICATION_WAIT_TIME"`
	ConcurrencyWaitTime        float64 `db:"CONCURRENCY_WAIT_TIME"`
	ClusterWaitTime            float64 `db:"CLUSTER_WAIT_TIME"`
	UserIOWaitTime             float64 `db:"USER_IO_WAIT_TIME"`
	PLSQLExecTime              float64 `db:"PLSQL_EXEC_TIME"`
	JavaExecTime               float64 `db:"JAVA_EXEC_TIME"`
	Sorts                      float64 `db:"SORTS"`
	IOCellOffloadEligibleBytes float64 `db:"IO_CELL_OFFLOAD_ELIGIBLE_BYTES"`
	IOCellUncompressedBytes    float64 `db:"IO_CELL_UNCOMPRESSED_BYTES"`
	IOCellOffloadReturnedBytes float64 `db:"IO_CELL_OFFLOAD_RETURNED_BYTES"`
	IOInterconnectBytes        float64 `db:"IO_INTERCONNECT_BYTES"`
	PhysicalReadRequests       float64 `db:"PHYSICAL_READ_REQUESTS"`
	PhysicalReadBytes          float64 `db:"PHYSICAL_READ_BYTES"`
	PhysicalWriteRequests      float64 `db:"PHYSICAL_WRITE_REQUESTS"`
	PhysicalWriteBytes         float64 `db:"PHYSICAL_WRITE_BYTES"`
	ObsoleteCount              float64 `db:"OBSOLETE_COUNT"`
	AvoidedExecutions          float64 `db:"AVOIDED_EXECUTIONS"`
}

type StatementMetricsGaugeDB struct {
	VersionCount float64 `db:"VERSION_COUNT"`
	SharableMem  float64 `db:"SHARABLE_MEM"`
	TypecheckMem float64 `db:"TYPECHECK_MEM"`
}

type OracleRowMonotonicCount struct {
	ParseCalls                 float64 `json:"parse_calls,omitempty"`
	DiskReads                  float64 `json:"disk_reads,omitempty"`
	DirectWrites               float64 `json:"direct_writes,omitempty"`
	DirectReads                float64 `json:"direct_reads,omitempty"`
	BufferGets                 float64 `json:"buffer_gets,omitempty"`
	RowsProcessed              float64 `json:"rows_processed,omitempty"`
	SerializableAborts         float64 `json:"serializable_aborts,omitempty"`
	Fetches                    float64 `json:"fetches,omitempty"`
	Executions                 float64 `json:"executions,omitempty"`
	EndOfFetchCount            float64 `json:"end_of_fetch_count,omitempty"`
	Loads                      float64 `json:"loads,omitempty"`
	Invalidations              float64 `json:"invalidations,omitempty"`
	PxServersExecutions        float64 `json:"px_servers_executions,omitempty"`
	CPUTime                    float64 `json:"cpu_time,omitempty"`
	ElapsedTime                float64 `json:"elapsed_time,omitempty"`
	ApplicationWaitTime        float64 `json:"application_wait_time,omitempty"`
	ConcurrencyWaitTime        float64 `json:"concurrency_wait_time,omitempty"`
	ClusterWaitTime            float64 `json:"cluster_wait_time,omitempty"`
	UserIOWaitTime             float64 `json:"user_io_wait_time,omitempty"`
	PLSQLExecTime              float64 `json:"plsql_exec_time,omitempty"`
	JavaExecTime               float64 `json:"java_exec_time,omitempty"`
	Sorts                      float64 `json:"sorts,omitempty"`
	IOCellOffloadEligibleBytes float64 `json:"io_cell_offload_eligible_bytes,omitempty"`
	IOCellUncompressedBytes    float64 `json:"io_cell_uncompressed_bytes,omitempty"`
	IOCellOffloadReturnedBytes float64 `json:"io_cell_offload_returned_bytes,omitempty"`
	IOInterconnectBytes        float64 `json:"io_interconnect_bytes,omitempty"`
	PhysicalReadRequests       float64 `json:"physical_read_requests,omitempty"`
	PhysicalReadBytes          float64 `json:"physical_read_bytes,omitempty"`
	PhysicalWriteRequests      float64 `json:"physical_write_requests,omitempty"`
	PhysicalWriteBytes         float64 `json:"physical_write_bytes,omitempty"`
	ObsoleteCount              float64 `json:"obsolete_count,omitempty"`
	AvoidedExecutions          float64 `json:"avoided_executions,omitempty"`
}

type OracleRow struct {
	querySignature string
	RawData        StatementMetricsDB
	DeltaData      OracleRowMonotonicCount
}

// isAllDeltasZero checks if all delta values in OracleRowMonotonicCount are zero.
func isAllDeltasZero(diff OracleRowMonotonicCount) bool {
	return diff.Executions == 0 && diff.ParseCalls == 0 && diff.DiskReads == 0 && diff.DirectWrites == 0 &&
		diff.DirectReads == 0 && diff.BufferGets == 0 && diff.RowsProcessed == 0 && diff.SerializableAborts == 0 &&
		diff.Fetches == 0 && diff.EndOfFetchCount == 0 && diff.Loads == 0 && diff.Invalidations == 0 &&
		diff.PxServersExecutions == 0 && diff.CPUTime == 0 && diff.ElapsedTime == 0 &&
		diff.ApplicationWaitTime == 0 && diff.ConcurrencyWaitTime == 0 && diff.ClusterWaitTime == 0 &&
		diff.UserIOWaitTime == 0 && diff.PLSQLExecTime == 0 && diff.JavaExecTime == 0 &&
		diff.Sorts == 0 && diff.IOCellOffloadEligibleBytes == 0 && diff.IOCellUncompressedBytes == 0 &&
		diff.IOCellOffloadReturnedBytes == 0 && diff.IOInterconnectBytes == 0 &&
		diff.PhysicalReadRequests == 0 && diff.PhysicalReadBytes == 0 &&
		diff.PhysicalWriteRequests == 0 && diff.PhysicalWriteBytes == 0 &&
		diff.ObsoleteCount == 0 && diff.AvoidedExecutions == 0
}

// collectDbmMetric collects DBM statement metrics from Oracle.
func (ipt *Input) collectDbmMetric(ctx context.Context, ptsTime time.Time) ([]*OracleRow, error) {
	if ipt.Dbm == nil || ipt.Dbm.Metric == nil || !ipt.Dbm.Metric.Enabled {
		return nil, nil
	}

	start := time.Now()

	oracleRows, err := queryOracleRows(ipt, ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to query Oracle rows: %w", err)
	}
	l.Debugf("queryOracleRows completed, collected %d rows, time taken: %s", len(oracleRows), time.Since(start))

	if len(oracleRows) == 0 {
		return nil, nil
	}

	// Sort by multiple dimensions (descending) to get top queries
	sort.Slice(oracleRows, func(i, j int) bool {
		return oracleRows[i].DeltaData.ElapsedTime > oracleRows[j].DeltaData.ElapsedTime
	})

	// Take top N queries based on MaxQueries configuration
	maxQueries := ipt.Dbm.Metric.MaxQueries
	if maxQueries <= 0 {
		maxQueries = 500 // Default to 500 if not configured
	}
	if len(oracleRows) > maxQueries {
		oracleRows = oracleRows[:maxQueries]
	}

	l.Debugf("collectDbmMetric completed, sorted and filtered to top %d rows by elapsed time/cpu time/executions/buffer gets, total time taken: %s",
		len(oracleRows), time.Since(start))

	pts := ipt.buildStatementPoints(oracleRows, ptsTime)
	if len(pts) > 0 {
		if err := ipt.feeder.Feed(point.Metric, pts,
			dkio.WithCollectCost(time.Since(start)),
			dkio.WithElection(ipt.Election),
			dkio.WithSource(dbmFeedName),
			dkio.WithMeasurement(inputs.GetOverrideMeasurement(ipt.MeasurementVersion, measurementOracle)),
		); err != nil {
			ipt.feeder.FeedLastError(err.Error(),
				metrics.WithLastErrorInput(inputName),
				metrics.WithLastErrorCategory(point.Metric),
			)
			l.Errorf("feed dbm metric failed: %s", err.Error())
		}
	}
	return oracleRows, nil
}

func queryOracleRows(ipt *Input, ctx context.Context) ([]*OracleRow, error) {
	// Calculate lookback window: use configured value if set, otherwise use 300 seconds
	lookbackWindowSeconds := ipt.Dbm.Metric.LookbackWindow
	if ipt.Dbm.Metric.LookbackWindow <= 0 {
		lookbackWindowSeconds = 300
	}

	// Get queries based on Oracle version
	queries := getStatementMetricsQueries(ipt.dbVersion)
	if len(queries) == 0 {
		return nil, fmt.Errorf("no queries available for Oracle version: %s", ipt.dbVersion)
	}

	// Select query based on configuration
	var sqlQuery string
	if ipt.Dbm.Metric.DisableLastActive {
		sqlQuery = queries[fmsRandomQuery]
	} else {
		sqlQuery = queries[fmsLastActiveQuery]
	}

	// Execute first query: force_matching_signature
	var statementMetrics []StatementMetricsDB
	queryStart := time.Now()
	err := selectWrapperWithBinds(ipt, ctx, &statementMetrics, sqlQuery, lookbackWindowSeconds, ipt.Dbm.Metric.DBRowsLimit)
	dbmSQLQueryDuration.WithLabelValues("statements", "fms_query").Observe(time.Since(queryStart).Seconds())
	if err != nil {
		return nil, fmt.Errorf("failed to query statement metrics: %w", err)
	}

	// Execute second query: SQL_ID (force_matching_signature = 0)
	sqlIDQuery := queries[sqlIDQuery]
	var statementMetricsSQLID []StatementMetricsDB
	queryStart = time.Now()
	err = selectWrapperWithBinds(ipt, ctx, &statementMetricsSQLID, sqlIDQuery, lookbackWindowSeconds, ipt.Dbm.Metric.DBRowsLimit)
	dbmSQLQueryDuration.WithLabelValues("statements", "sqlid_query").Observe(time.Since(queryStart).Seconds())
	if err != nil {
		return nil, fmt.Errorf("failed to query statement metrics for SQL_ID: %w", err)
	}
	// Merge both query results
	statementMetricsAll := make([]StatementMetricsDB, 0, len(statementMetrics)+len(statementMetricsSQLID))
	statementMetricsAll = append(statementMetricsAll, statementMetrics...)
	statementMetricsAll = append(statementMetricsAll, statementMetricsSQLID...)

	newCache := make(map[StatementMetricsKeyDB]StatementMetricsMonotonicCountDB)
	if ipt.statementMetricsMonotonicCountsPrevious == nil {
		ipt.statementMetricsMonotonicCountsPrevious = make(map[StatementMetricsKeyDB]StatementMetricsMonotonicCountDB)
	}
	var diff OracleRowMonotonicCount
	l.Debugf("queryOracleRows statementMetricsAll: %d", len(statementMetricsAll))
	oracleRows := make([]*OracleRow, 0, len(statementMetricsAll))
	for _, statementMetricRow := range statementMetricsAll {
		cacheKey := statementMetricRow.StatementMetricsKeyDB
		if cacheKey.ForceMatchingSignature != "" {
			cacheKey.SQLID = ""
		}

		newCache[cacheKey] = statementMetricRow.StatementMetricsMonotonicCountDB
		previousMonotonic, exists := ipt.statementMetricsMonotonicCountsPrevious[cacheKey]
		if !exists {
			continue
		}
		diff = OracleRowMonotonicCount{}
		if diff.Executions = statementMetricRow.Executions - previousMonotonic.Executions; diff.Executions < 0 {
			continue
		}
		if diff.ParseCalls = statementMetricRow.ParseCalls - previousMonotonic.ParseCalls; diff.ParseCalls < 0 {
			continue
		}
		if diff.DiskReads = statementMetricRow.DiskReads - previousMonotonic.DiskReads; diff.DiskReads < 0 {
			continue
		}
		if diff.DirectWrites = statementMetricRow.DirectWrites - previousMonotonic.DirectWrites; diff.DirectWrites < 0 {
			continue
		}
		if diff.DirectReads = statementMetricRow.DirectReads - previousMonotonic.DirectReads; diff.DirectReads < 0 {
			continue
		}
		if diff.BufferGets = statementMetricRow.BufferGets - previousMonotonic.BufferGets; diff.BufferGets < 0 {
			continue
		}
		if diff.RowsProcessed = statementMetricRow.RowsProcessed - previousMonotonic.RowsProcessed; diff.RowsProcessed < 0 {
			continue
		}
		if diff.SerializableAborts = statementMetricRow.SerializableAborts - previousMonotonic.SerializableAborts; diff.SerializableAborts < 0 {
			continue
		}
		if diff.Fetches = statementMetricRow.Fetches - previousMonotonic.Fetches; diff.Fetches < 0 {
			continue
		}
		if diff.EndOfFetchCount = statementMetricRow.EndOfFetchCount - previousMonotonic.EndOfFetchCount; diff.EndOfFetchCount < 0 {
			continue
		}
		if diff.Loads = statementMetricRow.Loads - previousMonotonic.Loads; diff.Loads < 0 {
			continue
		}
		if diff.Invalidations = statementMetricRow.Invalidations - previousMonotonic.Invalidations; diff.Invalidations < 0 {
			continue
		}
		if diff.PxServersExecutions = statementMetricRow.PxServersExecutions - previousMonotonic.PxServersExecutions; diff.PxServersExecutions < 0 {
			continue
		}
		if diff.CPUTime = statementMetricRow.CPUTime - previousMonotonic.CPUTime; diff.CPUTime < 0 {
			continue
		}
		if diff.ElapsedTime = statementMetricRow.ElapsedTime - previousMonotonic.ElapsedTime; diff.ElapsedTime < 0 {
			continue
		}
		if diff.ApplicationWaitTime = statementMetricRow.ApplicationWaitTime - previousMonotonic.ApplicationWaitTime; diff.ApplicationWaitTime < 0 {
			continue
		}
		if diff.ConcurrencyWaitTime = statementMetricRow.ConcurrencyWaitTime - previousMonotonic.ConcurrencyWaitTime; diff.ConcurrencyWaitTime < 0 {
			continue
		}
		if diff.ClusterWaitTime = statementMetricRow.ClusterWaitTime - previousMonotonic.ClusterWaitTime; diff.ClusterWaitTime < 0 {
			continue
		}
		if diff.UserIOWaitTime = statementMetricRow.UserIOWaitTime - previousMonotonic.UserIOWaitTime; diff.UserIOWaitTime < 0 {
			continue
		}
		if diff.PLSQLExecTime = statementMetricRow.PLSQLExecTime - previousMonotonic.PLSQLExecTime; diff.PLSQLExecTime < 0 {
			continue
		}
		if diff.JavaExecTime = statementMetricRow.JavaExecTime - previousMonotonic.JavaExecTime; diff.JavaExecTime < 0 {
			continue
		}
		if diff.Sorts = statementMetricRow.Sorts - previousMonotonic.Sorts; diff.Sorts < 0 {
			continue
		}
		//nolint:lll
		if diff.IOCellOffloadEligibleBytes = statementMetricRow.IOCellOffloadEligibleBytes - previousMonotonic.IOCellOffloadEligibleBytes; diff.IOCellOffloadEligibleBytes < 0 {
			continue
		}
		//nolint:lll
		if diff.IOCellUncompressedBytes = statementMetricRow.IOCellUncompressedBytes - previousMonotonic.IOCellUncompressedBytes; diff.IOCellUncompressedBytes < 0 {
			continue
		}
		//nolint:lll
		if diff.IOCellOffloadReturnedBytes = statementMetricRow.IOCellOffloadReturnedBytes - previousMonotonic.IOCellOffloadReturnedBytes; diff.IOCellOffloadReturnedBytes < 0 {
			continue
		}
		if diff.IOInterconnectBytes = statementMetricRow.IOInterconnectBytes - previousMonotonic.IOInterconnectBytes; diff.IOInterconnectBytes < 0 {
			continue
		}
		if diff.PhysicalReadRequests = statementMetricRow.PhysicalReadRequests - previousMonotonic.PhysicalReadRequests; diff.PhysicalReadRequests < 0 {
			continue
		}
		if diff.PhysicalReadBytes = statementMetricRow.PhysicalReadBytes - previousMonotonic.PhysicalReadBytes; diff.PhysicalReadBytes < 0 {
			continue
		}
		//nolint:lll
		if diff.PhysicalWriteRequests = statementMetricRow.PhysicalWriteRequests - previousMonotonic.PhysicalWriteRequests; diff.PhysicalWriteRequests < 0 {
			continue
		}
		if diff.PhysicalWriteBytes = statementMetricRow.PhysicalWriteBytes - previousMonotonic.PhysicalWriteBytes; diff.PhysicalWriteBytes < 0 {
			continue
		}
		if diff.ObsoleteCount = statementMetricRow.ObsoleteCount - previousMonotonic.ObsoleteCount; diff.ObsoleteCount < 0 {
			continue
		}
		if diff.AvoidedExecutions = statementMetricRow.AvoidedExecutions - previousMonotonic.AvoidedExecutions; diff.AvoidedExecutions < 0 {
			continue
		}

		// Skip if all delta values are zero
		if isAllDeltasZero(diff) {
			continue
		}

		// Generate query signature
		statementMetricRow.PDBName = ipt.getPdbName(statementMetricRow.PDBName)
		var queryHash string
		if statementMetricRow.ForceMatchingSignature != "" && statementMetricRow.ForceMatchingSignature != "0" {
			queryHash = statementMetricRow.ForceMatchingSignature
		} else {
			queryHash = statementMetricRow.SQLID
		}
		querySignature := generateQuerySignature(statementMetricRow.PDBName, queryHash)

		oracleRows = append(oracleRows, &OracleRow{
			querySignature: querySignature,
			RawData:        statementMetricRow,
			DeltaData:      diff,
		})
	}
	// Update cache for next collection
	ipt.statementMetricsMonotonicCountsPrevious = newCache

	return oracleRows, nil
}

// buildStatementPoints builds point.Point from statement rows.
func (ipt *Input) buildStatementPoints(rows []*OracleRow, ptsTime time.Time) []*point.Point {
	var pts []*point.Point
	opts := append(point.DefaultMetricOptions(), point.WithTime(ptsTime))

	for _, row := range rows {
		kvs := ipt.getKVs()

		// Tags
		kvs = kvs.AddTag("query_signature", row.querySignature)
		if row.RawData.ConID > 0 {
			kvs = kvs.AddTag("con_id", fmt.Sprintf("%d", row.RawData.ConID))
		}
		if ipt.cdbName != "" {
			kvs = kvs.AddTag("cdb_name", ipt.cdbName)
		}
		if row.RawData.PDBName != "" {
			kvs = kvs.AddTag("pdb_name", row.RawData.PDBName)
		}
		if row.RawData.ForceMatchingSignature != "" {
			kvs = kvs.AddTag("force_matching_signature", row.RawData.ForceMatchingSignature)
		} else {
			kvs = kvs.AddTag("force_matching_signature", "0")
		}

		kvs = kvs.AddTag("plan_hash_value", fmt.Sprintf("%d", row.RawData.PlanHashValue))

		// Total fields (cumulative values from Oracle) - only keep core metrics
		kvs = kvs.Set("total_executions", int64(row.RawData.Executions))
		kvs = kvs.Set("total_elapsed_time", int64(row.RawData.ElapsedTime))
		kvs = kvs.Set("total_cpu_time", int64(row.RawData.CPUTime))
		kvs = kvs.Set("total_buffer_gets", int64(row.RawData.BufferGets))
		kvs = kvs.Set("total_rows_processed", int64(row.RawData.RowsProcessed))

		// Gauge fields (current values from Oracle)
		kvs = kvs.Set("version_count", int64(row.RawData.VersionCount))
		kvs = kvs.Set("sharable_mem", int64(row.RawData.SharableMem))
		kvs = kvs.Set("typecheck_mem", int64(row.RawData.TypecheckMem))

		delta := row.DeltaData
		kvs = kvs.Set("delta_parse_calls", int64(delta.ParseCalls))
		kvs = kvs.Set("delta_disk_reads", int64(delta.DiskReads))
		kvs = kvs.Set("delta_direct_writes", int64(delta.DirectWrites))
		kvs = kvs.Set("delta_direct_reads", int64(delta.DirectReads))
		kvs = kvs.Set("delta_buffer_gets", int64(delta.BufferGets))
		kvs = kvs.Set("delta_rows_processed", int64(delta.RowsProcessed))
		kvs = kvs.Set("delta_serializable_aborts", int64(delta.SerializableAborts))
		kvs = kvs.Set("delta_fetches", int64(delta.Fetches))
		kvs = kvs.Set("delta_executions", int64(delta.Executions))
		kvs = kvs.Set("delta_end_of_fetch_count", int64(delta.EndOfFetchCount))
		kvs = kvs.Set("delta_loads", int64(delta.Loads))
		kvs = kvs.Set("delta_invalidations", int64(delta.Invalidations))
		kvs = kvs.Set("delta_px_servers_executions", int64(delta.PxServersExecutions))
		kvs = kvs.Set("delta_cpu_time", int64(delta.CPUTime))
		kvs = kvs.Set("delta_elapsed_time", int64(delta.ElapsedTime))
		// Calculate average elapsed time per execution
		if delta.Executions > 0 {
			avgElapsedTime := int64(delta.ElapsedTime / delta.Executions)
			kvs = kvs.Set("avg_elapsed_time", avgElapsedTime)
		}
		kvs = kvs.Set("delta_application_wait_time", int64(delta.ApplicationWaitTime))
		kvs = kvs.Set("delta_concurrency_wait_time", int64(delta.ConcurrencyWaitTime))
		kvs = kvs.Set("delta_cluster_wait_time", int64(delta.ClusterWaitTime))
		kvs = kvs.Set("delta_user_io_wait_time", int64(delta.UserIOWaitTime))
		kvs = kvs.Set("delta_plsql_exec_time", int64(delta.PLSQLExecTime))
		kvs = kvs.Set("delta_java_exec_time", int64(delta.JavaExecTime))
		kvs = kvs.Set("delta_sorts", int64(delta.Sorts))
		kvs = kvs.Set("delta_io_cell_offload_eligible_bytes", int64(delta.IOCellOffloadEligibleBytes))
		kvs = kvs.Set("delta_io_cell_uncompressed_bytes", int64(delta.IOCellUncompressedBytes))
		kvs = kvs.Set("delta_io_cell_offload_returned_bytes", int64(delta.IOCellOffloadReturnedBytes))
		kvs = kvs.Set("delta_io_interconnect_bytes", int64(delta.IOInterconnectBytes))
		kvs = kvs.Set("delta_physical_read_requests", int64(delta.PhysicalReadRequests))
		kvs = kvs.Set("delta_physical_read_bytes", int64(delta.PhysicalReadBytes))
		kvs = kvs.Set("delta_physical_write_requests", int64(delta.PhysicalWriteRequests))
		kvs = kvs.Set("delta_physical_write_bytes", int64(delta.PhysicalWriteBytes))
		kvs = kvs.Set("delta_obsolete_count", int64(delta.ObsoleteCount))
		kvs = kvs.Set("delta_avoided_executions", int64(delta.AvoidedExecutions))

		pt := point.NewPoint(metricNameOracleDbmMetric, kvs, opts...)
		pts = append(pts, pt)
	}

	return pts
}

// generateQuerySignature generates a unique signature for a SQL statement using xxhash.
func generateQuerySignature(pdbName, queryHash string) string {
	h := xxhash.New()
	_, _ = h.WriteString(pdbName)
	_, _ = h.WriteString(":")
	_, _ = h.WriteString(queryHash)

	return fmt.Sprintf("%016x", h.Sum64())
}

func (ipt *Input) getPdbName(pdbName string) string {
	if pdbName != "" {
		return pdbName
	}
	if ipt.isMultitenant {
		return "CDB$ROOT"
	}
	return ipt.cdbName
}
