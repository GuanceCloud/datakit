// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package kingbase

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
