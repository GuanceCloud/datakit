// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package sqlserver

import (
	"context"
	"fmt"
	"strconv"
	"strings"
)

// getAvailableColumns discovers available columns from a SQL Server DMV (Dynamic Management View).
// It executes "SELECT TOP 0 * FROM <table>" to get column metadata without fetching data.
func (ipt *Input) getAvailableColumns(ctx context.Context, tableName string, expectedColumns []string) ([]string, error) {
	// Check cache first
	cacheKey := tableName
	ipt.dbmColumnsCacheMu.Lock()
	cached, ok := ipt.dbmColumnsCache[cacheKey]
	ipt.dbmColumnsCacheMu.Unlock()
	if ok {
		return cached, nil
	}

	// Execute query to get column metadata
	query := "SELECT TOP 0 * FROM " + tableName
	rows, err := ipt.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query %s for column discovery: %w", tableName, err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			l.Errorf("failed to close rows: %v", err)
		}
	}()

	// Get column names from result set
	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("failed to get columns from %s: %w", tableName, err)
	}

	// Create a set of available columns for fast lookup
	availableSet := make(map[string]bool)
	for _, col := range columns {
		availableSet[strings.ToLower(col)] = true
	}

	// Filter expected columns to only include those that exist
	var availableColumns []string
	var missingColumns []string
	for _, col := range expectedColumns {
		if availableSet[strings.ToLower(col)] {
			availableColumns = append(availableColumns, col)
		} else {
			missingColumns = append(missingColumns, col)
		}
	}

	if len(missingColumns) > 0 {
		l.Debugf("missing the following expected columns from %s: %v", tableName, missingColumns)
	}
	l.Debugf("found available %s columns: %v", tableName, availableColumns)

	// Cache the result
	ipt.dbmColumnsCacheMu.Lock()
	ipt.dbmColumnsCache[cacheKey] = availableColumns
	ipt.dbmColumnsCacheMu.Unlock()

	return availableColumns, nil
}

// These columns are used to dynamically build the SELECT statement.
var expectedDbmStatementsColumns = []string{
	"execution_count",
	"total_worker_time",
	"total_physical_reads",
	"total_logical_writes",
	"total_logical_reads",
	"total_clr_time",
	"total_elapsed_time",
	"total_rows",
	"total_dop",
	"total_grant_kb",
	"total_used_grant_kb",
	"total_ideal_grant_kb",
	"total_reserved_threads",
	"total_used_threads",
	"total_columnstore_segment_reads",
	"total_columnstore_segment_skips",
	"total_spills",
}

// buildStatementsQuery dynamically builds the statements query based on available columns.
func (ipt *Input) buildStatementsQuery(ctx context.Context, lookbackWindowSeconds, rowLimit int) (string, error) {
	// Get available columns from sys.dm_exec_query_stats
	availableColumns, err := ipt.getAvailableColumns(ctx, "sys.dm_exec_query_stats", expectedDbmStatementsColumns)
	if err != nil {
		return "", fmt.Errorf("failed to discover columns: %w", err)
	}

	if len(availableColumns) == 0 {
		return "", fmt.Errorf("no available columns found in sys.dm_exec_query_stats")
	}

	// Build query_metrics_columns (for SELECT in qstats CTE)
	queryMetricsColumns := make([]string, len(availableColumns))
	for i, col := range availableColumns {
		queryMetricsColumns[i] = fmt.Sprintf("qs.%s as %s", col, col)
	}

	// Build query_metrics_column_sums (for SUM in qstats_aggr CTE)
	queryMetricsColumnSums := make([]string, len(availableColumns))
	for i, col := range availableColumns {
		queryMetricsColumnSums[i] = fmt.Sprintf("SUM(qs.%s) AS %s", col, col)
	}

	// Build the query
	//nolint:lll
	query := fmt.Sprintf(`SET DEADLOCK_PRIORITY -10;
IF SERVERPROPERTY('EngineEdition') NOT IN (2,3,4) BEGIN /*NOT IN Standard,Enterprise,Express*/
	DECLARE @ErrorMessage AS nvarchar(500) = 'DataKit - Connection string Server:'+ @@ServerName + ',Database:' + DB_NAME() +' is not a SQL Server Standard,Enterprise or Express. Check the database_type parameter in the datakit configuration.';
	RAISERROR (@ErrorMessage,11,1)
	RETURN
END;

WITH qstats AS (
    SELECT
        qs.query_hash,
        qs.query_plan_hash,
        qs.last_execution_time,
        qs.last_elapsed_time,
            CONVERT(VARCHAR(64), CONVERT(binary(64), qs.plan_handle), 1) +
            CONVERT(VARCHAR(10), CONVERT(varbinary(4), qs.statement_start_offset), 1) +
            CONVERT(VARCHAR(10), CONVERT(varbinary(4), qs.statement_end_offset), 1) AS plan_handle_and_offsets,
        (SELECT value FROM sys.dm_exec_plan_attributes(qs.plan_handle) WHERE attribute = 'dbid') AS dbid,
        eps.object_id AS sproc_object_id,
        %s
    FROM sys.dm_exec_query_stats qs
    LEFT JOIN sys.dm_exec_procedure_stats eps ON eps.plan_handle = qs.plan_handle
),
qstats_aggr AS (
    SELECT
        query_hash,
        query_plan_hash,
        CAST(qs.dbid AS int) AS dbid,
        dbs.name AS database_name,
        MAX(plan_handle_and_offsets) AS plan_handle_and_offsets,
        MAX(last_execution_time) AS last_execution_time,
        MAX(last_elapsed_time) AS last_elapsed_time,
        sproc_object_id,
        %s
    FROM qstats qs
    LEFT JOIN sys.databases dbs ON qs.dbid = dbs.database_id
    GROUP BY query_hash, query_plan_hash, qs.dbid, dbs.name, sproc_object_id
),
qstats_aggr_split AS (
    SELECT TOP %d
        CONVERT(varbinary(64), CONVERT(binary(64), SUBSTRING(plan_handle_and_offsets, 1, 64), 1)) AS plan_handle,
        CONVERT(int, CONVERT(varbinary(10), SUBSTRING(plan_handle_and_offsets, 64+1, 10), 1)) AS statement_start_offset,
        CONVERT(int, CONVERT(varbinary(10), SUBSTRING(plan_handle_and_offsets, 64+11, 10), 1)) AS statement_end_offset,
        *
    FROM qstats_aggr
	WHERE DATEADD(ms, last_elapsed_time / 1000, last_execution_time) > DATEADD(second, -%d, GETDATE())
)
SELECT
    SUBSTRING(qt.text, (statement_start_offset / 2) + 1,
        ((CASE statement_end_offset
            WHEN -1 THEN DATALENGTH(qt.text)
            ELSE statement_end_offset
        END - statement_start_offset) / 2) + 1) AS statement_text,
    SUBSTRING(qt.text, 1, %d) AS text,
    encrypted AS is_encrypted,
    OBJECT_SCHEMA_NAME(s.sproc_object_id, s.dbid) AS schema_name,
    OBJECT_NAME(s.sproc_object_id, s.dbid) AS procedure_name,
    REPLACE(@@SERVERNAME,'\',':') AS [sqlserver_host],
    s.*
FROM qstats_aggr_split s
CROSS APPLY sys.dm_exec_sql_text(s.plan_handle) qt`,
		strings.Join(queryMetricsColumns, ", "),
		strings.Join(queryMetricsColumnSums, ", "),
		rowLimit,
		lookbackWindowSeconds,
		ipt.Dbm.StoredProcedureCharactersLimit)

	return query, nil
}

func (ipt *Input) buildIdleBlockingSessionsQuery(blockingSessionIDs []int64) string {
	if len(blockingSessionIDs) == 0 {
		return ""
	}
	var validIDs []string
	for _, id := range blockingSessionIDs {
		if id > 0 {
			validIDs = append(validIDs, strconv.FormatInt(id, 10))
		}
	}
	if len(validIDs) == 0 {
		return ""
	}
	idsStr := strings.Join(validIDs, ",")
	var query string
	if ipt.MajorVersion <= 2008 {
		query = strings.ReplaceAll(dbmIdleBlockingSessionsQuery08, "__BLOCKING_SESSION_IDS__", idsStr)
	} else {
		query = strings.ReplaceAll(dbmIdleBlockingSessionsQuery, "__BLOCKING_SESSION_IDS__", idsStr)
	}
	return query
}

// expectedDbmActivityColumns defines the list of columns expected from sys.dm_exec_requests.
var expectedDbmActivityColumns = []string{
	"command",
	"blocking_session_id",
	"wait_type",
	"wait_time",
	"last_wait_type",
	"wait_resource",
	"open_transaction_count",
	"transaction_id",
	"percent_complete",
	"estimated_completion_time",
	"cpu_time",
	"total_elapsed_time",
	"reads",
	"writes",
	"logical_reads",
	"transaction_isolation_level",
	"lock_timeout",
	"deadlock_priority",
	"row_count",
	"query_hash",
	"query_plan_hash",
	"context_info",
}

// buildActivityQueryWithColumns dynamically builds the activity query based on available columns.
func (ipt *Input) buildActivityQueryWithColumns(ctx context.Context, rowLimit int) (string, error) {
	// Get available columns from sys.dm_exec_requests (cached)
	availableColumns, err := ipt.getAvailableColumns(ctx, "sys.dm_exec_requests", expectedDbmActivityColumns)
	if err != nil {
		return "", fmt.Errorf("failed to discover columns: %w", err)
	}

	if len(availableColumns) == 0 {
		return "", fmt.Errorf("no available columns found in sys.dm_exec_requests")
	}

	// Build exec_request_columns (for SELECT from req)
	execRequestColumns := make([]string, len(availableColumns))
	for i, col := range availableColumns {
		execRequestColumns[i] = fmt.Sprintf("req.%s", col)
	}

	//nolint:lll
	query := fmt.Sprintf(`SET DEADLOCK_PRIORITY -10;
IF SERVERPROPERTY('EngineEdition') NOT IN (2,3,4) BEGIN /*NOT IN Standard,Enterprise,Express*/
	DECLARE @ErrorMessage AS nvarchar(500) = 'DataKit - Connection string Server:'+ @@ServerName + ',Database:' + DB_NAME() +' is not a SQL Server Standard,Enterprise or Express. Check the database_type parameter in the datakit configuration.';
	RAISERROR (@ErrorMessage,11,1)
	RETURN
END

SELECT TOP %d
    CONVERT(NVARCHAR, TODATETIMEOFFSET(CURRENT_TIMESTAMP, DATEPART(TZOFFSET, SYSDATETIMEOFFSET())), 126) AS now,
    CONVERT(NVARCHAR, TODATETIMEOFFSET(req.start_time, DATEPART(TZOFFSET, SYSDATETIMEOFFSET())), 126) AS query_start,
    sess.login_name AS user_name,
    sess.last_request_start_time AS last_request_start_time,
    sess.session_id AS id,
    DB_NAME(req.database_id) AS database_name,
    sess.status AS session_status,
    req.status AS request_status,
    SUBSTRING(qt.text, (req.statement_start_offset / 2) + 1,
        ((CASE req.statement_end_offset
            WHEN -1 THEN DATALENGTH(qt.text)
            ELSE req.statement_end_offset
        END - req.statement_start_offset) / 2) + 1) AS statement_text,
    SUBSTRING(qt.text, 1, %d) AS text,
    CASE
        WHEN LEN(qt.text) > %d THEN RIGHT(qt.text, 200)
        ELSE ''
    END AS tail_text,
    OBJECT_SCHEMA_NAME(qt.objectid, req.database_id) AS schema_name,
    OBJECT_NAME(qt.objectid, req.database_id) AS procedure_name,
    REPLACE(@@SERVERNAME,'\',':') AS [sqlserver_host],
    c.client_tcp_port AS client_port,
    c.client_net_address AS client_address,
    sess.host_name AS host_name,
    sess.program_name AS program_name,
    sess.is_user_process AS is_user_process,
    sess.client_interface_name AS client_interface_name,
    %s
FROM sys.dm_exec_sessions sess
INNER JOIN sys.dm_exec_connections c ON sess.session_id = c.session_id
INNER JOIN sys.dm_exec_requests req ON c.connection_id = req.connection_id
CROSS APPLY sys.dm_exec_sql_text(req.sql_handle) qt
WHERE
    sess.session_id != @@spid AND
    sess.status != 'sleeping'`,
		rowLimit,
		ipt.Dbm.StoredProcedureCharactersLimit,
		ipt.Dbm.StoredProcedureCharactersLimit,
		strings.Join(execRequestColumns, ", "))

	return query, nil
}
