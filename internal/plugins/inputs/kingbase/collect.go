// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package kingbase

import (
	"fmt"

	"github.com/GuanceCloud/cliutils/point"
	"github.com/jmoiron/sqlx"
)

func q[T any](db *sqlx.DB, query string, args ...any) ([]T, error) {
	rows, err := db.Queryx(query, args...)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close() //nolint:errcheck

	var results []T
	for rows.Next() {
		rowMap := map[string]any{}
		if err := rows.MapScan(rowMap); err != nil {
			return nil, fmt.Errorf("rows.MapScan failed: %w", err)
		}

		var result T
		if err := mapToStruct(rowMap, &result); err != nil {
			return nil, fmt.Errorf("decode failed: %w", err)
		}
		results = append(results, result)
	}
	return results, nil
}

func (ipt *Input) addCommonTags(kvs point.KVs) point.KVs {
	kvs = kvs.AddTag("database", ipt.Database)
	if ipt.Version != "" {
		kvs = kvs.AddTag("db_version", ipt.Version)
	}
	// kvs = kvs.AddTag("host", ipt.Host) // 放在了 ipt.mergedTags 里
	for k, v := range ipt.mergedTags {
		kvs = kvs.AddTag(k, v)
	}
	return kvs
}

func (ipt *Input) collectConnections() error {
	results, err := q[connectionResult](
		ipt.db,
		queries[metricNameConnection],
		ipt.Database,
	)
	if err != nil {
		return fmt.Errorf("failed to collect connections: %w", err)
	}
	if len(results) == 0 {
		return nil
	}
	r := results[0]

	var kvs point.KVs
	kvs = ipt.addCommonTags(kvs)
	kvs = kvs.Set("active_connections", r.ActiveConnections)
	kvs = kvs.Set("idle_connections", r.IdleConnections)
	kvs = kvs.Set("max_connections", r.MaxConnections)

	opts := point.DefaultMetricOptions()
	opts = append(opts, point.WithTime(ipt.ptsTime))
	ipt.collectCache = append(ipt.collectCache, point.NewPoint(metricNameConnection, kvs, opts...))
	return nil
}

func (ipt *Input) collectQueryStats() error {
	results, err := q[queryStat](
		ipt.db,
		fmt.Sprintf(queries[metricNameQueryStats], "total_exec_time"),
		ipt.Database,
	)
	if err != nil {
		return fmt.Errorf("failed to collect query stats: %w", err)
	}
	for _, r := range results {
		var kvs point.KVs
		kvs = ipt.addCommonTags(kvs)
		kvs = kvs.AddTag("queryid", r.QueryID)
		kvs = kvs.Set("total_time", r.TotalTime)
		kvs = kvs.Set("calls", r.Calls)
		kvs = kvs.Set("rows", r.Rows)
		kvs = kvs.Set("shared_blks_hit", r.SharedBlksHit)
		kvs = kvs.Set("shared_blks_read", r.SharedBlksRead)

		opts := point.DefaultMetricOptions()
		opts = append(opts, point.WithTime(ipt.ptsTime))
		ipt.collectCache = append(ipt.collectCache, point.NewPoint(metricNameQueryStats, kvs, opts...))
	}
	return nil
}

func (ipt *Input) collectBufferCache() error {
	results, err := q[bufferCache](
		ipt.db,
		queries[metricNameBufferCache],
		ipt.Database,
	)
	if err != nil {
		return fmt.Errorf("failed to collect buffer cache: %w", err)
	}
	if len(results) == 0 {
		return nil
	}
	r := results[0]
	var kvs point.KVs
	kvs = ipt.addCommonTags(kvs)
	kvs = kvs.Set("buffer_hit_ratio", r.BufferHitRatio)
	kvs = kvs.Set("shared_blks_hit", r.SharedBlksHit)
	kvs = kvs.Set("shared_blks_read", r.SharedBlksRead)

	opts := point.DefaultMetricOptions()
	opts = append(opts, point.WithTime(ipt.ptsTime))
	ipt.collectCache = append(ipt.collectCache, point.NewPoint(metricNameBufferCache, kvs, opts...))
	return nil
}

func (ipt *Input) collectIndexUsage() error {
	results, err := q[indexUsage](
		ipt.db,
		queries[metricNameIndexUsage],
	)
	if err != nil {
		return fmt.Errorf("failed to collect index usage: %w", err)
	}
	if len(results) == 0 {
		return nil
	}
	r := results[0]
	var kvs point.KVs
	kvs = ipt.addCommonTags(kvs)
	kvs = kvs.Set("idx_scan", r.IdxScan)
	kvs = kvs.Set("seq_scan", r.SeqScan)
	kvs = kvs.Set("index_hit_ratio", r.IndexHitRatio)

	opts := point.DefaultMetricOptions()
	opts = append(opts, point.WithTime(ipt.ptsTime))
	ipt.collectCache = append(ipt.collectCache, point.NewPoint(metricNameIndexUsage, kvs, opts...))
	return nil
}

func (ipt *Input) collectBackgroundWriter() error {
	results, err := q[backgroundWriter](
		ipt.db,
		queries[metricNameBackgroundWriter],
	)
	if err != nil {
		return fmt.Errorf("failed to collect background writer: %w", err)
	}
	if len(results) == 0 {
		return nil
	}
	r := results[0]
	var kvs point.KVs
	kvs = ipt.addCommonTags(kvs)
	kvs = kvs.Set("buffers_clean", r.BuffersClean)
	kvs = kvs.Set("buffers_backend", r.BuffersBackend)
	kvs = kvs.Set("checkpoints_timed", r.CheckpointsTimed)
	kvs = kvs.Set("checkpoints_req", r.CheckpointsReq)

	opts := point.DefaultMetricOptions()
	opts = append(opts, point.WithTime(ipt.ptsTime))
	ipt.collectCache = append(ipt.collectCache, point.NewPoint(metricNameBackgroundWriter, kvs, opts...))
	return nil
}

func (ipt *Input) collectTransactions() error {
	results, err := q[transaction](
		ipt.db,
		queries[metricNameTransactions],
		ipt.Database,
	)
	if err != nil {
		return fmt.Errorf("failed to collect transactions: %w", err)
	}
	if len(results) == 0 {
		return nil
	}
	r := results[0]
	var kvs point.KVs
	kvs = ipt.addCommonTags(kvs)
	kvs = kvs.Set("commits", r.Commits)
	kvs = kvs.Set("rollbacks", r.Rollbacks)

	opts := point.DefaultMetricOptions()
	opts = append(opts, point.WithTime(ipt.ptsTime))
	ipt.collectCache = append(ipt.collectCache, point.NewPoint(metricNameTransactions, kvs, opts...))
	return nil
}

func (ipt *Input) collectLocks() error {
	results, err := q[lock](
		ipt.db,
		queries[metricNameLocks],
	)
	if err != nil {
		return fmt.Errorf("failed to collect locks: %w", err)
	}
	if len(results) == 0 {
		return nil
	}
	r := results[0]
	var kvs point.KVs
	kvs = ipt.addCommonTags(kvs)
	kvs = kvs.Set("waiting_locks", r.WaitingLocks)

	opts := point.DefaultMetricOptions()
	opts = append(opts, point.WithTime(ipt.ptsTime))
	ipt.collectCache = append(ipt.collectCache, point.NewPoint(metricNameLocks, kvs, opts...))
	return nil
}

func (ipt *Input) collectTablespace() error {
	results, err := q[tablespace](ipt.db, queries[metricNameTablespace])
	if err != nil {
		return fmt.Errorf("failed to collect tablespace: %w", err)
	}

	for _, r := range results {
		var kvs point.KVs
		kvs = ipt.addCommonTags(kvs)
		kvs = kvs.AddTag("spcname", r.SpcName)
		kvs = kvs.Set("size_bytes", r.SizeBytes)

		opts := point.DefaultMetricOptions()
		opts = append(opts, point.WithTime(ipt.ptsTime))
		ipt.collectCache = append(ipt.collectCache, point.NewPoint(metricNameTablespace, kvs, opts...))
	}
	return nil
}

func (ipt *Input) collectQueryPerformance() error {
	results, err := q[queryPerformance](
		ipt.db,
		queries[metricNameQueryPerformance],
		ipt.Database,
	)
	if err != nil {
		return fmt.Errorf("failed to collect query performance: %w", err)
	}

	for _, r := range results {
		var kvs point.KVs
		kvs = ipt.addCommonTags(kvs)
		kvs = kvs.AddTag("queryid", r.QueryID)
		kvs = kvs.AddTag("query", r.Query)
		kvs = kvs.Set("mean_exec_time", r.MeanExecTime)

		opts := point.DefaultMetricOptions()
		opts = append(opts, point.WithTime(ipt.ptsTime))
		ipt.collectCache = append(ipt.collectCache, point.NewPoint(metricNameQueryPerformance, kvs, opts...))
	}
	return nil
}

func (ipt *Input) collectDatabaseStatus() error {
	results, err := q[databaseStatus](ipt.db, queries[metricNameDatabaseStatus], ipt.Database)
	if err != nil {
		return fmt.Errorf("failed to collect database status: %w", err)
	}
	if len(results) == 0 {
		return nil
	}
	r := results[0]
	var kvs point.KVs
	kvs = ipt.addCommonTags(kvs)
	kvs = kvs.Set("numbackends", r.Numbackends)
	kvs = kvs.Set("blks_hit", r.BlksHit)
	kvs = kvs.Set("blks_read", r.BlksRead)
	kvs = kvs.Set("tup_inserted", r.TupInserted)
	kvs = kvs.Set("tup_updated", r.TupUpdated)
	kvs = kvs.Set("tup_deleted", r.TupDeleted)
	kvs = kvs.Set("conflicts", r.Conflicts)

	opts := point.DefaultMetricOptions()
	opts = append(opts, point.WithTime(ipt.ptsTime))
	ipt.collectCache = append(ipt.collectCache, point.NewPoint(metricNameDatabaseStatus, kvs, opts...))
	return nil
}

func (ipt *Input) collectLockDetails() error {
	results, err := q[lockDetail](
		ipt.db,
		queries[metricNameLockDetails],
		ipt.Database,
	)
	if err != nil {
		return fmt.Errorf("failed to collect lock details: %w", err)
	}

	for _, r := range results {
		var kvs point.KVs
		kvs = ipt.addCommonTags(kvs)
		kvs = kvs.AddTag("lock_type", r.LockType)
		kvs = kvs.Set("lock_count", r.LockCount)

		opts := point.DefaultMetricOptions()
		opts = append(opts, point.WithTime(ipt.ptsTime))
		ipt.collectCache = append(ipt.collectCache, point.NewPoint(metricNameLockDetails, kvs, opts...))
	}
	return nil
}

func (ipt *Input) collectSessionActivity() error {
	results, err := q[sessionActivity](ipt.db, queries[metricNameSessionActivity], ipt.Database)
	if err != nil {
		return fmt.Errorf("failed to collect session activity: %w", err)
	}
	for _, r := range results {
		var kvs point.KVs
		kvs = ipt.addCommonTags(kvs)
		kvs = kvs.AddTag("state", r.State)
		kvs = kvs.AddTag("wait_event", r.WaitEvent)
		kvs = kvs.Set("session_count", r.SessionCount)

		opts := point.DefaultMetricOptions()
		opts = append(opts, point.WithTime(ipt.ptsTime))
		ipt.collectCache = append(ipt.collectCache, point.NewPoint(metricNameSessionActivity, kvs, opts...))
	}
	return nil
}

func (ipt *Input) collectQueryCancellation() error {
	results, err := q[queryCancellation](ipt.db, queries[metricNameQueryCancellation], ipt.Database)
	if err != nil {
		return fmt.Errorf("failed to collect query cancellation: %w", err)
	}
	if len(results) == 0 {
		return nil
	}
	r := results[0]
	var kvs point.KVs
	kvs = ipt.addCommonTags(kvs)
	kvs = kvs.Set("temp_files", r.TempFiles)
	kvs = kvs.Set("deadlocks", r.Deadlocks)

	opts := point.DefaultMetricOptions()
	opts = append(opts, point.WithTime(ipt.ptsTime))
	ipt.collectCache = append(ipt.collectCache, point.NewPoint(metricNameQueryCancellation, kvs, opts...))
	return nil
}

func (ipt *Input) collectFunctionStats() error {
	results, err := q[FunctionStat](ipt.db, queries[metricNameFunctionStats])
	if err != nil {
		return fmt.Errorf("failed to collect function stats: %w", err)
	}

	for _, r := range results {
		var kvs point.KVs
		kvs = ipt.addCommonTags(kvs)
		kvs = kvs.AddTag("schemaname", r.Schemaname)
		kvs = kvs.AddTag("funcname", r.Funcname)
		kvs = kvs.Set("calls", r.Calls)
		kvs = kvs.Set("total_time", r.TotalTime)
		kvs = kvs.Set("self_time", r.SelfTime)

		opts := point.DefaultMetricOptions()
		opts = append(opts, point.WithTime(ipt.ptsTime))
		ipt.collectCache = append(ipt.collectCache, point.NewPoint(metricNameFunctionStats, kvs, opts...))
	}
	return nil
}

func (ipt *Input) collectSlowQueries() error {
	threshold := ipt.SlowQueryThreshold
	if threshold <= 0 {
		threshold = 1000 // Default to 1 second if not set
	}

	results, err := q[slowQuery](ipt.db, queries[metricNameSlowQueries], ipt.Database, float64(threshold))
	if err != nil {
		return fmt.Errorf("failed to collect slow queries: %w", err)
	}

	for _, r := range results {
		var kvs point.KVs
		kvs = ipt.addCommonTags(kvs)
		kvs = kvs.AddTag("queryid", r.QueryID)
		kvs = kvs.AddTag("query", r.Query)
		kvs = kvs.Set("total_exec_time", r.TotalExecTime)
		kvs = kvs.Set("calls", r.Calls)
		kvs = kvs.Set("mean_exec_time", r.MeanExecTime)

		opts := point.DefaultMetricOptions()
		opts = append(opts, point.WithTime(ipt.ptsTime))
		ipt.collectCache = append(ipt.collectCache, point.NewPoint(metricNameSlowQueries, kvs, opts...))
	}
	return nil
}
