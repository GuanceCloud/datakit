// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package kingbase

import (
	"database/sql"
	"fmt"

	"github.com/GuanceCloud/cliutils/point"
)

var queries = map[string]string{
	metricNameConnection: `
		SELECT 
			count(*) FILTER (WHERE state = 'active') AS active_connections,
			count(*) FILTER (WHERE state = 'idle') AS idle_connections,
			(SELECT setting::int FROM sys_catalog.sys_settings WHERE name = 'max_connections') AS max_connections
		FROM sys_catalog.sys_stat_activity
		WHERE datname = ?`,
	metricNameQueryStats: `
		SELECT 
			queryid,
			%s AS total_time,
			calls,
			rows,
			shared_blks_hit,
			shared_blks_read
		FROM public.sys_stat_statements
		WHERE dbid IN (SELECT oid FROM sys_catalog.sys_database WHERE datname = ?)`,
	metricNameBufferCache: `
		SELECT 
			sum(blks_hit) AS shared_blks_hit,
			sum(blks_read) AS shared_blks_read,
			CASE 
				WHEN sum(blks_hit + blks_read) > 0 
				THEN (sum(blks_hit) * 100.0) / sum(blks_hit + blks_read)
				ELSE 0
			END AS buffer_hit_ratio
		FROM sys_catalog.sys_stat_database
		WHERE datname = ?`,
	metricNameIndexUsage: `
		SELECT 
			sum(idx_scan) AS idx_scan,
			sum(seq_scan) AS seq_scan,
			CASE 
				WHEN sum(idx_scan + seq_scan) > 0 
				THEN (sum(idx_scan) * 100.0) / sum(idx_scan + seq_scan)
				ELSE 0
			END AS index_hit_ratio
		FROM sys_catalog.sys_stat_user_tables
		WHERE schemaname != 'sys_catalog'`,
	metricNameBackgroundWriter: `
		SELECT 
			buffers_clean,
			buffers_backend,
			checkpoints_timed,
			checkpoints_req
		FROM sys_catalog.sys_stat_bgwriter`,
	metricNameTransactions: `
		SELECT 
			sum(xact_commit) AS commits,
			sum(xact_rollback) AS rollbacks
		FROM sys_catalog.sys_stat_database
		WHERE datname = ?`,
	metricNameLocks: `
		SELECT 
			count(*) AS waiting_locks
		FROM sys_catalog.sys_locks
		WHERE granted = false`,
	metricNameTablespace: `
		SELECT 
			spcname,
			sys_tablespace_size(oid) AS size_bytes
		FROM sys_catalog.sys_tablespace`,
	metricNameQueryPerformance: `
		SELECT 
			queryid,
			query,
			mean_exec_time
		FROM public.sys_stat_statements
		WHERE dbid IN (SELECT oid FROM sys_catalog.sys_database WHERE datname = ?)`,
	metricNameDatabaseStatus: `
		SELECT 
			datname,
			numbackends,
			blks_hit,
			blks_read,
			tup_inserted,
			tup_updated,
			tup_deleted,
			conflicts
		FROM sys_catalog.sys_stat_database
		WHERE datname = ?`,
	metricNameLockDetails: `
		SELECT 
			mode AS lock_type,
			count(*) AS lock_count
		FROM sys_catalog.sys_locks
		JOIN sys_catalog.sys_stat_activity ON sys_catalog.sys_locks.pid = sys_catalog.sys_stat_activity.pid
		WHERE sys_catalog.sys_stat_activity.datname = ?
		GROUP BY mode`,
	metricNameSessionActivity: `
		SELECT 
			state,
			wait_event,
			count(*) AS session_count
		FROM sys_catalog.sys_stat_activity
		WHERE datname = ?
		GROUP BY state, wait_event`,
	metricNameQueryCancellation: `
		SELECT 
			sum(temp_files) AS temp_files,
			sum(deadlocks) AS deadlocks
		FROM sys_catalog.sys_stat_database
		WHERE datname = ?`,
	metricNameFunctionStats: `
		SELECT 
			schemaname,
			funcname,
			calls,
			total_time,
			self_time
		FROM sys_catalog.sys_stat_user_functions`,
	metricNameSlowQueries: `
		SELECT 
			queryid,
			query,
			total_exec_time,
			calls,
			mean_exec_time
		FROM public.sys_stat_statements
		WHERE dbid IN (SELECT oid FROM sys_catalog.sys_database WHERE datname = ?)
			AND mean_exec_time > ?
		ORDER BY mean_exec_time DESC
		LIMIT 100`,
}

func (ipt *Input) addCommonTags(kvs point.KVs) point.KVs {
	kvs = kvs.AddTag("database", ipt.Database)
	// kvs = kvs.AddTag("host", ipt.Host) // 放在了 ipt.mergedTags 里
	for k, v := range ipt.mergedTags {
		kvs = kvs.AddTag(k, v)
	}
	return kvs
}

func (ipt *Input) collectConnections() error {
	var result struct {
		ActiveConnections int64 `db:"active_connections"`
		IdleConnections   int64 `db:"idle_connections"`
		MaxConnections    int64 `db:"max_connections"`
	}
	err := ipt.db.Get(&result, queries[metricNameConnection], ipt.Database)
	if err != nil {
		return fmt.Errorf("failed to collect connections: %w", err)
	}

	var kvs point.KVs
	kvs = kvs.AddTag("database", ipt.Database)
	kvs = kvs.AddTag("host", ipt.Host)
	for k, v := range ipt.mergedTags {
		kvs = kvs.AddTag(k, v)
	}
	kvs = kvs.Add("active_connections", result.ActiveConnections, false, true)
	kvs = kvs.Add("idle_connections", result.IdleConnections, false, true)
	kvs = kvs.Add("max_connections", result.MaxConnections, false, true)

	opts := point.DefaultMetricOptions()
	opts = append(opts, point.WithTime(ipt.ptsTime))
	ipt.collectCache = append(ipt.collectCache, point.NewPointV2(metricNameConnection, kvs, opts...))
	return nil
}

func (ipt *Input) collectQueryStats() error {
	query := fmt.Sprintf(queries[metricNameQueryStats], "total_exec_time")
	rows, err := ipt.db.Queryx(query, ipt.Database)
	if err != nil {
		return fmt.Errorf("failed to collect query stats: %w", err)
	}
	defer rows.Close() //nolint:errcheck

	for rows.Next() {
		var result struct {
			QueryID        sql.NullString `db:"queryid"`
			TotalTime      float64        `db:"total_time"`
			Calls          int64          `db:"calls"`
			Rows           int64          `db:"rows"`
			SharedBlksHit  int64          `db:"shared_blks_hit"`
			SharedBlksRead int64          `db:"shared_blks_read"`
		}
		if err := rows.StructScan(&result); err != nil {
			l.Errorf("Scan query stats result failed: %v", err)
			continue
		}

		var kvs point.KVs
		kvs = ipt.addCommonTags(kvs)
		queryID := ""
		if result.QueryID.Valid {
			queryID = result.QueryID.String
		}
		kvs = kvs.AddTag("queryid", queryID)
		kvs = kvs.Add("total_time", result.TotalTime, false, true)
		kvs = kvs.Add("calls", result.Calls, false, true)
		kvs = kvs.Add("rows", result.Rows, false, true)
		kvs = kvs.Add("shared_blks_hit", result.SharedBlksHit, false, true)
		kvs = kvs.Add("shared_blks_read", result.SharedBlksRead, false, true)

		opts := point.DefaultMetricOptions()
		opts = append(opts, point.WithTime(ipt.ptsTime))
		ipt.collectCache = append(ipt.collectCache, point.NewPointV2(metricNameQueryStats, kvs, opts...))
	}
	return nil
}

func (ipt *Input) collectBufferCache() error {
	var result struct {
		SharedBlksHit  int64   `db:"shared_blks_hit"`
		SharedBlksRead int64   `db:"shared_blks_read"`
		BufferHitRatio float64 `db:"buffer_hit_ratio"`
	}
	err := ipt.db.Get(&result, queries[metricNameBufferCache], ipt.Database)
	if err != nil {
		return fmt.Errorf("failed to collect buffer cache: %w", err)
	}

	var kvs point.KVs
	kvs = ipt.addCommonTags(kvs)
	kvs = kvs.Add("buffer_hit_ratio", result.BufferHitRatio, false, true)
	kvs = kvs.Add("shared_blks_hit", result.SharedBlksHit, false, true)
	kvs = kvs.Add("shared_blks_read", result.SharedBlksRead, false, true)

	opts := point.DefaultMetricOptions()
	opts = append(opts, point.WithTime(ipt.ptsTime))
	ipt.collectCache = append(ipt.collectCache, point.NewPointV2(metricNameBufferCache, kvs, opts...))
	return nil
}

func (ipt *Input) collectIndexUsage() error {
	var result struct {
		IdxScan       int64   `db:"idx_scan"`
		SeqScan       int64   `db:"seq_scan"`
		IndexHitRatio float64 `db:"index_hit_ratio"`
	}
	err := ipt.db.Get(&result, queries[metricNameIndexUsage])
	if err != nil {
		return fmt.Errorf("failed to collect index usage: %w", err)
	}

	var kvs point.KVs
	kvs = ipt.addCommonTags(kvs)
	kvs = kvs.Add("idx_scan", result.IdxScan, false, true)
	kvs = kvs.Add("seq_scan", result.SeqScan, false, true)
	kvs = kvs.Add("index_hit_ratio", result.IndexHitRatio, false, true)

	opts := point.DefaultMetricOptions()
	opts = append(opts, point.WithTime(ipt.ptsTime))
	ipt.collectCache = append(ipt.collectCache, point.NewPointV2(metricNameIndexUsage, kvs, opts...))
	return nil
}

func (ipt *Input) collectBackgroundWriter() error {
	var result struct {
		BuffersClean     int64 `db:"buffers_clean"`
		BuffersBackend   int64 `db:"buffers_backend"`
		CheckpointsTimed int64 `db:"checkpoints_timed"`
		CheckpointsReq   int64 `db:"checkpoints_req"`
	}
	err := ipt.db.Get(&result, queries[metricNameBackgroundWriter])
	if err != nil {
		return fmt.Errorf("failed to collect background writer: %w", err)
	}

	var kvs point.KVs
	kvs = ipt.addCommonTags(kvs)
	kvs = kvs.Add("buffers_clean", result.BuffersClean, false, true)
	kvs = kvs.Add("buffers_backend", result.BuffersBackend, false, true)
	kvs = kvs.Add("checkpoints_timed", result.CheckpointsTimed, false, true)
	kvs = kvs.Add("checkpoints_req", result.CheckpointsReq, false, true)

	opts := point.DefaultMetricOptions()
	opts = append(opts, point.WithTime(ipt.ptsTime))
	ipt.collectCache = append(ipt.collectCache, point.NewPointV2(metricNameBackgroundWriter, kvs, opts...))
	return nil
}

func (ipt *Input) collectTransactions() error {
	var result struct {
		Commits   int64 `db:"commits"`
		Rollbacks int64 `db:"rollbacks"`
	}
	err := ipt.db.Get(&result, queries[metricNameTransactions], ipt.Database)
	if err != nil {
		return fmt.Errorf("failed to collect transactions: %w", err)
	}

	var kvs point.KVs
	kvs = ipt.addCommonTags(kvs)
	kvs = kvs.Add("commits", result.Commits, false, true)
	kvs = kvs.Add("rollbacks", result.Rollbacks, false, true)

	opts := point.DefaultMetricOptions()
	opts = append(opts, point.WithTime(ipt.ptsTime))
	ipt.collectCache = append(ipt.collectCache, point.NewPointV2(metricNameTransactions, kvs, opts...))
	return nil
}

func (ipt *Input) collectLocks() error {
	var result struct {
		WaitingLocks int64 `db:"waiting_locks"`
	}
	err := ipt.db.Get(&result, queries[metricNameLocks])
	if err != nil {
		return fmt.Errorf("failed to collect locks: %w", err)
	}

	var kvs point.KVs
	kvs = ipt.addCommonTags(kvs)
	kvs = kvs.Add("waiting_locks", result.WaitingLocks, false, true)

	opts := point.DefaultMetricOptions()
	opts = append(opts, point.WithTime(ipt.ptsTime))
	ipt.collectCache = append(ipt.collectCache, point.NewPointV2(metricNameLocks, kvs, opts...))
	return nil
}

func (ipt *Input) collectTablespace() error {
	rows, err := ipt.db.Queryx(queries[metricNameTablespace])
	if err != nil {
		return fmt.Errorf("failed to collect tablespace: %w", err)
	}
	defer rows.Close() //nolint:errcheck

	for rows.Next() {
		var result struct {
			SpcName   sql.NullString `db:"spcname"`
			SizeBytes int64          `db:"size_bytes"`
		}
		if err := rows.StructScan(&result); err != nil {
			l.Errorf("Scan tablespace result failed: %v", err)
			continue
		}

		var kvs point.KVs
		kvs = ipt.addCommonTags(kvs)
		spcName := ""
		if result.SpcName.Valid {
			spcName = result.SpcName.String
		}
		kvs = kvs.AddTag("spcname", spcName)
		kvs = kvs.Add("size_bytes", result.SizeBytes, false, true)

		opts := point.DefaultMetricOptions()
		opts = append(opts, point.WithTime(ipt.ptsTime))
		ipt.collectCache = append(ipt.collectCache, point.NewPointV2(metricNameTablespace, kvs, opts...))
	}
	return nil
}

func (ipt *Input) collectQueryPerformance() error {
	rows, err := ipt.db.Queryx(queries[metricNameQueryPerformance], ipt.Database)
	if err != nil {
		return fmt.Errorf("failed to collect query performance: %w", err)
	}
	defer rows.Close() //nolint:errcheck

	for rows.Next() {
		var result struct {
			QueryID      sql.NullString `db:"queryid"`
			Query        sql.NullString `db:"query"`
			MeanExecTime float64        `db:"mean_exec_time"`
		}
		if err := rows.StructScan(&result); err != nil {
			l.Errorf("Scan query performance result failed: %v", err)
			continue
		}

		var kvs point.KVs
		kvs = ipt.addCommonTags(kvs)
		queryID := ""
		if result.QueryID.Valid {
			queryID = result.QueryID.String
		}
		query := ""
		if result.Query.Valid {
			query = truncateQuery(result.Query.String, 64)
		}
		kvs = kvs.AddTag("queryid", queryID)
		kvs = kvs.AddTag("query", query)
		kvs = kvs.Add("mean_exec_time", result.MeanExecTime, false, true)

		opts := point.DefaultMetricOptions()
		opts = append(opts, point.WithTime(ipt.ptsTime))
		ipt.collectCache = append(ipt.collectCache, point.NewPointV2(metricNameQueryPerformance, kvs, opts...))
	}
	return nil
}

func (ipt *Input) collectDatabaseStatus() error {
	var result struct {
		Datname     sql.NullString `db:"datname"`
		Numbackends int64          `db:"numbackends"`
		BlksHit     int64          `db:"blks_hit"`
		BlksRead    int64          `db:"blks_read"`
		TupInserted int64          `db:"tup_inserted"`
		TupUpdated  int64          `db:"tup_updated"`
		TupDeleted  int64          `db:"tup_deleted"`
		Conflicts   int64          `db:"conflicts"`
	}
	err := ipt.db.Get(&result, queries[metricNameDatabaseStatus], ipt.Database)
	if err != nil {
		return fmt.Errorf("failed to collect database status: %w", err)
	}

	var kvs point.KVs
	kvs = ipt.addCommonTags(kvs)
	kvs = kvs.Add("numbackends", result.Numbackends, false, true)
	kvs = kvs.Add("blks_hit", result.BlksHit, false, true)
	kvs = kvs.Add("blks_read", result.BlksRead, false, true)
	kvs = kvs.Add("tup_inserted", result.TupInserted, false, true)
	kvs = kvs.Add("tup_updated", result.TupUpdated, false, true)
	kvs = kvs.Add("tup_deleted", result.TupDeleted, false, true)
	kvs = kvs.Add("conflicts", result.Conflicts, false, true)

	opts := point.DefaultMetricOptions()
	opts = append(opts, point.WithTime(ipt.ptsTime))
	ipt.collectCache = append(ipt.collectCache, point.NewPointV2(metricNameDatabaseStatus, kvs, opts...))
	return nil
}

func (ipt *Input) collectLockDetails() error {
	rows, err := ipt.db.Queryx(queries[metricNameLockDetails], ipt.Database)
	if err != nil {
		return fmt.Errorf("failed to collect lock details: %w", err)
	}
	defer rows.Close() //nolint:errcheck

	for rows.Next() {
		var result struct {
			LockType  sql.NullString `db:"lock_type"`
			LockCount int64          `db:"lock_count"`
		}
		if err := rows.StructScan(&result); err != nil {
			l.Errorf("Scan lock details result failed: %v", err)
			continue
		}

		var kvs point.KVs
		kvs = ipt.addCommonTags(kvs)
		lockType := ""
		if result.LockType.Valid {
			lockType = result.LockType.String
		}
		kvs = kvs.AddTag("lock_type", lockType)
		kvs = kvs.Add("lock_count", result.LockCount, false, true)

		opts := point.DefaultMetricOptions()
		opts = append(opts, point.WithTime(ipt.ptsTime))
		ipt.collectCache = append(ipt.collectCache, point.NewPointV2(metricNameLockDetails, kvs, opts...))
	}
	return nil
}

func (ipt *Input) collectSessionActivity() error {
	rows, err := ipt.db.Queryx(queries[metricNameSessionActivity], ipt.Database)
	if err != nil {
		return fmt.Errorf("failed to collect session activity: %w", err)
	}
	defer rows.Close() //nolint:errcheck

	for rows.Next() {
		var result struct {
			State        sql.NullString `db:"state"`
			WaitEvent    sql.NullString `db:"wait_event"`
			SessionCount int64          `db:"session_count"`
		}
		if err := rows.StructScan(&result); err != nil {
			l.Errorf("Scan session activity result failed: %v", err)
			continue
		}

		var kvs point.KVs
		kvs = ipt.addCommonTags(kvs)
		state := ""
		if result.State.Valid {
			state = result.State.String
		}
		waitEvent := ""
		if result.WaitEvent.Valid {
			waitEvent = result.WaitEvent.String
		}
		kvs = kvs.AddTag("state", state)
		kvs = kvs.AddTag("wait_event", waitEvent)
		kvs = kvs.Add("session_count", result.SessionCount, false, true)

		opts := point.DefaultMetricOptions()
		opts = append(opts, point.WithTime(ipt.ptsTime))
		ipt.collectCache = append(ipt.collectCache, point.NewPointV2(metricNameSessionActivity, kvs, opts...))
	}
	return nil
}

// Truncate query to avoid oversized tags.
func truncateQuery(query string, maxLen int) string {
	if len(query) > maxLen {
		return query[:maxLen-3] + "..."
	}
	return query
}

func (ipt *Input) collectQueryCancellation() error {
	var result struct {
		TempFiles int64 `db:"temp_files"`
		Deadlocks int64 `db:"deadlocks"`
	}
	err := ipt.db.Get(&result, queries[metricNameQueryCancellation], ipt.Database)
	if err != nil {
		return fmt.Errorf("failed to collect query cancellation: %w", err)
	}

	var kvs point.KVs
	kvs = ipt.addCommonTags(kvs)
	kvs = kvs.Add("temp_files", result.TempFiles, false, true)
	kvs = kvs.Add("deadlocks", result.Deadlocks, false, true)

	opts := point.DefaultMetricOptions()
	opts = append(opts, point.WithTime(ipt.ptsTime))
	ipt.collectCache = append(ipt.collectCache, point.NewPointV2(metricNameQueryCancellation, kvs, opts...))
	return nil
}

func (ipt *Input) collectFunctionStats() error {
	rows, err := ipt.db.Queryx(queries[metricNameFunctionStats])
	if err != nil {
		return fmt.Errorf("failed to collect function stats: %w", err)
	}
	defer rows.Close() //nolint:errcheck

	for rows.Next() {
		var result struct {
			Schemaname sql.NullString `db:"schemaname"`
			Funcname   sql.NullString `db:"funcname"`
			Calls      int64          `db:"calls"`
			TotalTime  float64        `db:"total_time"`
			SelfTime   float64        `db:"self_time"`
		}
		if err := rows.StructScan(&result); err != nil {
			l.Errorf("Scan function stats result failed: %v", err)
			continue
		}

		var kvs point.KVs
		kvs = ipt.addCommonTags(kvs)
		schemaname := ""
		if result.Schemaname.Valid {
			schemaname = result.Schemaname.String
		}
		funcname := ""
		if result.Funcname.Valid {
			funcname = result.Funcname.String
		}
		kvs = kvs.AddTag("schemaname", schemaname)
		kvs = kvs.AddTag("funcname", funcname)
		kvs = kvs.Add("calls", result.Calls, false, true)
		kvs = kvs.Add("total_time", result.TotalTime, false, true)
		kvs = kvs.Add("self_time", result.SelfTime, false, true)

		opts := point.DefaultMetricOptions()
		opts = append(opts, point.WithTime(ipt.ptsTime))
		ipt.collectCache = append(ipt.collectCache, point.NewPointV2(metricNameFunctionStats, kvs, opts...))
	}
	return nil
}

func (ipt *Input) collectSlowQueries() error {
	threshold := ipt.SlowQueryThreshold
	if threshold <= 0 {
		threshold = 1000 // Default to 1 second if not set
	}
	rows, err := ipt.db.Queryx(queries[metricNameSlowQueries], ipt.Database, float64(threshold))
	if err != nil {
		return fmt.Errorf("failed to collect slow queries: %w", err)
	}
	defer rows.Close() //nolint:errcheck

	for rows.Next() {
		var result struct {
			QueryID       sql.NullString `db:"queryid"`
			Query         sql.NullString `db:"query"`
			TotalExecTime float64        `db:"total_exec_time"`
			Calls         int64          `db:"calls"`
			MeanExecTime  float64        `db:"mean_exec_time"`
		}
		if err := rows.StructScan(&result); err != nil {
			l.Errorf("Scan slow query result failed: %v", err)
			continue
		}

		var kvs point.KVs
		kvs = ipt.addCommonTags(kvs)
		queryID := ""
		if result.QueryID.Valid {
			queryID = result.QueryID.String
		}
		query := ""
		if result.Query.Valid {
			query = truncateQuery(result.Query.String, 30)
		}
		kvs = kvs.AddTag("queryid", queryID)
		kvs = kvs.AddTag("query", query)
		kvs = kvs.Add("total_exec_time", result.TotalExecTime, false, true)
		kvs = kvs.Add("calls", result.Calls, false, true)
		kvs = kvs.Add("mean_exec_time", result.MeanExecTime, false, true)

		opts := point.DefaultMetricOptions()
		opts = append(opts, point.WithTime(ipt.ptsTime))
		ipt.collectCache = append(ipt.collectCache, point.NewPointV2(metricNameSlowQueries, kvs, opts...))
	}
	return nil
}
