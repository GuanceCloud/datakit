// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package dameng

import (
	"database/sql"
	"fmt"
	"strconv"

	"github.com/GuanceCloud/cliutils/point"
)

var queries = map[string]string{
	metricNameMemory: `
        SELECT
            (SELECT SUM(n_pages) * page()/1024/1024 FROM v$bufferpool) AS "buffer_size_mb",
            (SELECT SUM(total_size)/1024/1024 FROM v$mem_pool) AS "mem_pool_size_mb",
            (SELECT SUM(n_pages) * page()/1024/1024 FROM v$bufferpool) + (SELECT SUM(total_size)/1024/1024 FROM v$mem_pool) AS "total_size_mb"
        FROM dual`,
	metricNameMemPool: `
        SELECT
            name AS "pool_name",
            is_shared AS "is_shared",
            is_overflow AS "is_overflow",
            org_size / 1024.0 / 1024.0 AS "org_size_mb",
            total_size / 1024.0 / 1024.0 AS "total_size_mb",
            reserved_size / 1024.0 / 1024.0 AS "reserved_size_mb",
            data_size / 1024.0 / 1024.0 AS "data_size_mb",
            extend_size AS "extend_size_mb",
            target_size AS "target_size_mb",
            n_extend_normal AS "n_extend_normal",
            n_extend_exclusive AS "n_extend_exclusive"
        FROM v$mem_pool
        ORDER BY total_size DESC`,
	metricNameTablespace: `
        SELECT
            UPPER(F.TABLESPACE_NAME) AS "tablespace_name",
            D.TOT_GROOTTE_MB AS "total_size_mb",
            D.TOT_GROOTTE_MB - F.TOTAL_BYTES AS "used_size_mb",
            ROUND((D.TOT_GROOTTE_MB - F.TOTAL_BYTES) / D.TOT_GROOTTE_MB * 100, 2) AS "usage_ratio",
            F.TOTAL_BYTES AS "free_size_mb",
            F.MAX_BYTES AS "max_block_mb"
        FROM
            (SELECT TABLESPACE_NAME,
                    ROUND(SUM(BYTES) / (1024 * 1024), 2) TOTAL_BYTES,
                    ROUND(MAX(BYTES) / (1024 * 1024), 2) MAX_BYTES
             FROM SYS.DBA_FREE_SPACE
             GROUP BY TABLESPACE_NAME) F,
            (SELECT DD.TABLESPACE_NAME,
                    ROUND(SUM(DD.BYTES) / (1024 * 1024), 2) TOT_GROOTTE_MB
             FROM SYS.DBA_DATA_FILES DD
             GROUP BY DD.TABLESPACE_NAME) D
        WHERE D.TABLESPACE_NAME = F.TABLESPACE_NAME
        ORDER BY 2 DESC`,
	metricNameConnection: `
		SELECT 
			SUM(CASE WHEN state = 'ACTIVE' THEN 1 ELSE 0 END) AS "active_connections",
			SUM(CASE WHEN state = 'IDLE' THEN 1 ELSE 0 END) AS "idle_connections",
			(SELECT CAST(VALUE AS BIGINT) FROM V$PARAMETER WHERE NAME = 'MAX_SESSIONS') AS "max_connections"
		FROM V$SESSIONS`,
	metricNameRates: `
		SELECT
   			name AS "name",
    		stat_val AS "stat_val"
			FROM sys.v$sysstat
		WHERE name IN ('sql executed count', 'transaction commit count', 'transaction rollback count');`,
	metricNameSlowQueries: `
		SELECT
            SESS_ID AS "sess_id",
            SQL_ID AS "sql_id",
            SQL_TEXT AS "sql_text",
            EXEC_TIME AS "exec_time",
            N_RUNS AS "n_runs"
        FROM V$LONG_EXEC_SQLS
        WHERE EXEC_TIME > ?
        ORDER BY EXEC_TIME DESC
        LIMIT 100`,

	metricNameLocks: `
		SELECT 
        	COUNT(*) AS "waiting_locks"
    	FROM V$LOCK
    	WHERE BLOCKED = 1`,
	metricNameDeadlock: `
        SELECT
            dh.trx_id AS "deadlock_trx_id",
            dh.sess_id AS "deadlock_sess_id",
            COUNT(*) AS "deadlock_count"
        FROM
            V$DEADLOCK_HISTORY dh,
            V$SQL_HISTORY sh
        WHERE
            dh.trx_id = sh.trx_id
            AND dh.sess_id = sh.sess_id
			AND dh.happen_time < SYSDATE - 1/24
        GROUP BY
            dh.trx_id, dh.sess_id`,
	metricNameBufferCache: `
		SELECT
            name AS "pool_name",
            SUM(page_size) * SF_GET_PAGE_SIZE() AS "total_size_bytes",
            ROUND(SUM(page_size) * SF_GET_PAGE_SIZE() / 1073741824.0, 3) AS "total_size_gb",
            SUM(rat_hit) / COUNT(*) AS "buffer_hit_ratio"
        FROM V$BUFFERPOOL
        GROUP BY name`,
	metricNameBlockSessions: `
		SELECT
            DS.SESS_ID AS "blocked_sess_id",
            DS.TRX_ID AS "blocked_trx_id",
            (CASE L.LTYPE 
                WHEN 'OBJECT' THEN 'object_lock' 
                WHEN 'TID' THEN 'transaction_lock' 
                ELSE 'other' 
            END) AS "blocked_lock_type",
            DS.CREATE_TIME AS "blocked_start_time",
            ROUND((SYSDATE - DS.CREATE_TIME) * 1440, 2) AS "block_duration_min",
            SS.SESS_ID AS "blocking_sess_id",
            SS.CLNT_IP AS "blocking_ip",
            L.TID AS "blocking_trx_id"
        FROM
            V$LOCK L
        LEFT JOIN V$SESSIONS DS ON DS.TRX_ID = L.TRX_ID
        LEFT JOIN V$SESSIONS SS ON SS.TRX_ID = L.TID
        WHERE
            L.BLOCKED = 1
            AND DS.CREATE_TIME < SYSDATE - 1/24 `,
}

func (ipt *Input) addCommonTags(kvs point.KVs) point.KVs {
	kvs = kvs.AddTag("database", ipt.Database)
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
	err := ipt.db.Get(&result, queries[metricNameConnection])
	if err != nil {
		return fmt.Errorf("failed to collect connections: %w", err)
	}

	var kvs point.KVs
	kvs = ipt.addCommonTags(kvs)
	kvs = kvs.AddTag("database", ipt.Database)
	kvs = kvs.AddTag("host", ipt.Host)
	kvs = kvs.Set("active_connections", result.ActiveConnections)
	kvs = kvs.Set("idle_connections", result.IdleConnections)
	kvs = kvs.Set("max_connections", result.MaxConnections)

	opts := point.DefaultMetricOptions()
	opts = append(opts, point.WithTime(ipt.ptsTime))
	ipt.collectCache = append(ipt.collectCache, point.NewPoint(metricNameConnection, kvs, opts...))
	return nil
}

func (ipt *Input) collectBufferCache() error {
	var results []struct {
		PoolName       string  `db:"pool_name"`
		TotalSizeBytes int64   `db:"total_size_bytes"`
		TotalSizeGb    float64 `db:"total_size_gb"`
		BufferHitRatio float64 `db:"buffer_hit_ratio"`
	}

	err := ipt.db.Select(&results, queries[metricNameBufferCache])
	if err != nil {
		return fmt.Errorf("failed to collect buffer cache: %w", err)
	}
	if len(results) == 0 {
		return fmt.Errorf("no buffer cache metrics found")
	}

	for _, result := range results {
		kvs := point.KVs{}
		kvs = ipt.addCommonTags(kvs)
		kvs = kvs.AddTag("pool_name", result.PoolName)
		kvs = kvs.Set("buffer_hit_ratio", result.BufferHitRatio)
		kvs = kvs.Set("total_size_bytes", result.TotalSizeBytes)
		kvs = kvs.Set("total_size_gb", result.TotalSizeGb)

		opts := point.DefaultMetricOptions()
		opts = append(opts, point.WithTime(ipt.ptsTime))
		ipt.collectCache = append(ipt.collectCache, point.NewPoint(metricNameBufferCache, kvs, opts...))
	}
	return nil
}

func (ipt *Input) collectBlockSessions() error {
	var results []struct {
		BlockedSessID    int64   `db:"blocked_sess_id"`
		BlockedTrxID     int64   `db:"blocked_trx_id"`
		BlockedLockType  string  `db:"blocked_lock_type"`
		BlockedStartTime string  `db:"blocked_start_time"`
		BlockDurationMin float64 `db:"block_duration_min"`
		BlockingSessID   int64   `db:"blocking_sess_id"`
		BlockingIP       string  `db:"blocking_ip"`
		BlockingTrxID    int64   `db:"blocking_trx_id"`
	}

	err := ipt.db.Select(&results, queries[metricNameBlockSessions])
	if err != nil {
		return fmt.Errorf("failed to collect blocked sessions: %w", err)
	}
	if len(results) == 0 {
		return nil
	}

	var kvs point.KVs
	for _, result := range results {
		kvs = point.KVs{}
		kvs = ipt.addCommonTags(kvs)
		kvs = kvs.AddTag("blocked_sess_id", strconv.FormatInt(result.BlockedSessID, 10))
		kvs = kvs.AddTag("blocked_trx_id", strconv.FormatInt(result.BlockedTrxID, 10))
		kvs = kvs.AddTag("blocked_lock_type", result.BlockedLockType)
		kvs = kvs.AddTag("blocked_start_time", result.BlockedStartTime)
		kvs = kvs.AddTag("blocking_sess_id", strconv.FormatInt(result.BlockingSessID, 10))
		kvs = kvs.AddTag("blocking_ip", result.BlockingIP)
		kvs = kvs.AddTag("blocking_trx_id", strconv.FormatInt(result.BlockingTrxID, 10))
		kvs = kvs.Set("block_duration_min", result.BlockDurationMin)

		opts := point.DefaultMetricOptions()
		opts = append(opts, point.WithTime(ipt.ptsTime))
		ipt.collectCache = append(ipt.collectCache, point.NewPoint(metricNameBlockSessions, kvs, opts...))
	}
	return nil
}

func (ipt *Input) collectDeadlock() error {
	var results []struct {
		DeadlockTrxID  string `db:"deadlock_trx_id"`
		DeadlockSessID string `db:"deadlock_sess_id"`
		DeadlockCount  int64  `db:"deadlock_count"`
	}

	err := ipt.db.Select(&results, queries[metricNameDeadlock])
	if err != nil {
		return fmt.Errorf("failed to collect deadlocks: %w", err)
	}

	for _, result := range results {
		var kvs point.KVs
		kvs = ipt.addCommonTags(kvs)
		kvs = kvs.AddTag("trx_id", result.DeadlockTrxID)
		kvs = kvs.AddTag("sess_id", result.DeadlockSessID)
		kvs = kvs.Set("deadlock_count", result.DeadlockCount)

		opts := point.DefaultMetricOptions()
		opts = append(opts, point.WithTime(ipt.ptsTime))
		ipt.collectCache = append(ipt.collectCache, point.NewPoint(metricNameDeadlock, kvs, opts...))
	}

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
	kvs = kvs.Set("waiting_locks", result.WaitingLocks)

	opts := point.DefaultMetricOptions()
	opts = append(opts, point.WithTime(ipt.ptsTime))
	ipt.collectCache = append(ipt.collectCache, point.NewPoint(metricNameLocks, kvs, opts...))
	return nil
}

func (ipt *Input) collectMemory() error {
	var result struct {
		BufferSizeMB  float64 `db:"buffer_size_mb"`
		MemPoolSizeMB float64 `db:"mem_pool_size_mb"`
		TotalSizeMB   float64 `db:"total_size_mb"`
	}

	err := ipt.db.Get(&result, queries[metricNameMemory])
	if err != nil {
		return fmt.Errorf("failed to collect memory metrics: %w", err)
	}

	var kvs point.KVs
	kvs = ipt.addCommonTags(kvs)
	kvs = kvs.AddTag("database", ipt.Database)
	kvs = kvs.AddTag("host", ipt.Host)
	kvs = kvs.Set("buffer_size_mb", result.BufferSizeMB)
	kvs = kvs.Set("mem_pool_size_mb", result.MemPoolSizeMB)
	kvs = kvs.Set("total_size_mb", result.TotalSizeMB)

	opts := point.DefaultMetricOptions()
	opts = append(opts, point.WithTime(ipt.ptsTime))
	ipt.collectCache = append(ipt.collectCache, point.NewPoint(metricNameMemory, kvs, opts...))
	return nil
}

func (ipt *Input) collectMemPool() error {
	var results []struct {
		PoolName         string  `db:"pool_name"`
		IsShared         string  `db:"is_shared"`
		IsOverflow       string  `db:"is_overflow"`
		OrgSizeMB        float64 `db:"org_size_mb"`
		TotalSizeMB      float64 `db:"total_size_mb"`
		ReservedSizeMB   float64 `db:"reserved_size_mb"`
		DataSizeMB       float64 `db:"data_size_mb"`
		ExtendSize       float64 `db:"extend_size_mb"`
		TargetSize       float64 `db:"target_size_mb"`
		NExtendNormal    int64   `db:"n_extend_normal"`
		NExtendExclusive int64   `db:"n_extend_exclusive"`
	}

	err := ipt.db.Select(&results, queries[metricNameMemPool])
	if err != nil {
		return fmt.Errorf("failed to collect mem_pool metrics: %w", err)
	}

	if len(results) == 0 {
		return fmt.Errorf("no mem_pool metrics found")
	}

	for _, result := range results {
		kvs := point.KVs{}
		kvs = ipt.addCommonTags(kvs)
		kvs = kvs.AddTag("pool_name", result.PoolName)
		kvs = kvs.AddTag("is_shared", result.IsShared)
		kvs = kvs.AddTag("is_overflow", result.IsOverflow)

		kvs = kvs.Set("org_size_mb", result.OrgSizeMB)
		kvs = kvs.Set("total_size_mb", result.TotalSizeMB)
		kvs = kvs.Set("reserved_size_mb", result.ReservedSizeMB)
		kvs = kvs.Set("data_size_mb", result.DataSizeMB)
		kvs = kvs.Set("extend_size_mb", result.ExtendSize)
		kvs = kvs.Set("target_size_mb", result.TargetSize)
		kvs = kvs.Set("n_extend_normal", result.NExtendNormal)
		kvs = kvs.Set("n_extend_exclusive", result.NExtendExclusive)

		opts := point.DefaultMetricOptions()
		opts = append(opts, point.WithTime(ipt.ptsTime))
		ipt.collectCache = append(ipt.collectCache, point.NewPoint(metricNameMemPool, kvs, opts...))
	}

	return nil
}

func (ipt *Input) collectRates() error {
	var results []struct {
		Name    string `db:"name"`
		StatVal int64  `db:"stat_val"`
	}
	err := ipt.db.Select(&results, queries[metricNameRates])
	if err != nil {
		return fmt.Errorf("failed to collect QPS and TPS: %w", err)
	}

	if len(results) == 0 {
		return fmt.Errorf("no QPS/TPS metrics found")
	}

	currentStats := make(map[string]int64)
	for _, result := range results {
		currentStats[result.Name] = result.StatVal
	}

	requiredKeys := []string{"sql executed count", "transaction commit count", "transaction rollback count"}
	for _, key := range requiredKeys {
		if _, ok := currentStats[key]; !ok {
			return fmt.Errorf("missing stat key: %s, skipping QPS/TPS calculation", key)
		}
		if _, ok := ipt.LastStatValues[key]; !ok && !ipt.LastStatTime.IsZero() {
			return fmt.Errorf("missing last stat key: %s, skipping QPS/TPS calculation", key)
		}
	}

	if ipt.LastStatValues == nil || ipt.LastStatTime.IsZero() {
		ipt.LastStatValues = currentStats
		ipt.LastStatTime = ipt.ptsTime
		return nil
	}

	timeDelta := ipt.ptsTime.Sub(ipt.LastStatTime).Seconds()
	if timeDelta <= 0 {
		return fmt.Errorf("invalid time delta for QPS/TPS calculation")
	}

	qps := float64(currentStats["sql executed count"]-ipt.LastStatValues["sql executed count"]) / timeDelta
	tps := float64((currentStats["transaction commit count"]-ipt.LastStatValues["transaction commit count"])+
		(currentStats["transaction rollback count"]-ipt.LastStatValues["transaction rollback count"])) / timeDelta

	kvs := point.KVs{}
	kvs = ipt.addCommonTags(kvs)
	kvs = kvs.Set("qps", qps)
	kvs = kvs.Set("tps", tps)

	opts := point.DefaultMetricOptions()
	opts = append(opts, point.WithTime(ipt.ptsTime))
	ipt.collectCache = append(ipt.collectCache, point.NewPoint(metricNameRates, kvs, opts...))

	ipt.LastStatValues = currentStats
	ipt.LastStatTime = ipt.ptsTime

	return nil
}

func (ipt *Input) collectSlowQueries() error {
	threshold := ipt.SlowQueryThreshold
	if threshold <= 0 {
		threshold = 1000 // Default to 1 second if not set
	}
	var results []struct {
		SessID   int64          `db:"sess_id"`
		SQLID    sql.NullString `db:"sql_id"`
		SQLText  sql.NullString `db:"sql_text"`
		ExecTime int64          `db:"exec_time"`
		NRuns    int64          `db:"n_runs"`
	}
	err := ipt.db.Select(&results, queries[metricNameSlowQueries], float64(threshold))
	if err != nil {
		return fmt.Errorf("failed to collect slow queries: %w", err)
	}

	if len(results) == 0 {
		return nil
	}

	for _, result := range results {
		kvs := point.KVs{}
		kvs = ipt.addCommonTags(kvs)
		kvs = kvs.AddTag("sess_id", strconv.FormatInt(result.SessID, 10))
		kvs = kvs.AddTag("sql_id", result.SQLID.String)
		kvs = kvs.AddTag("sql_text", result.SQLText.String)
		kvs = kvs.Set("exec_time", result.ExecTime)
		kvs = kvs.Set("n_runs", result.NRuns)

		opts := point.DefaultMetricOptions()
		opts = append(opts, point.WithTime(ipt.ptsTime))
		ipt.collectCache = append(ipt.collectCache, point.NewPoint(metricNameSlowQueries, kvs, opts...))
	}
	return nil
}

func (ipt *Input) collectTablespace() error {
	var results []struct {
		TablespaceName string  `db:"tablespace_name"`
		TotalSizeMB    float64 `db:"total_size_mb"`
		UsedSizeMB     float64 `db:"used_size_mb"`
		UsageRatio     float64 `db:"usage_ratio"`
		FreeSizeMB     float64 `db:"free_size_mb"`
		MaxBlockMB     float64 `db:"max_block_mb"`
	}
	err := ipt.db.Select(&results, queries[metricNameTablespace])
	if err != nil {
		return fmt.Errorf("failed to collect tablespace metrics: %w", err)
	}

	if len(results) == 0 {
		return fmt.Errorf("no tablespace metrics found")
	}

	for _, result := range results {
		kvs := point.KVs{}
		kvs = ipt.addCommonTags(kvs)
		kvs = kvs.AddTag("tablespace_name", result.TablespaceName)
		kvs = kvs.Set("usage_ratio", result.UsageRatio)
		kvs = kvs.Set("total_size_mb", result.TotalSizeMB)
		kvs = kvs.Set("used_size_mb", result.UsedSizeMB)
		kvs = kvs.Set("free_size_mb", result.FreeSizeMB)
		kvs = kvs.Set("max_block_mb", result.MaxBlockMB)

		opts := point.DefaultMetricOptions()
		opts = append(opts, point.WithTime(ipt.ptsTime))
		ipt.collectCache = append(ipt.collectCache, point.NewPoint(metricNameTablespace, kvs, opts...))
	}
	return nil
}
