// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package sqlserver

import (
	"context"
	"database/sql"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"

	"github.com/DataDog/datadog-agent/pkg/obfuscate"
)

const (
	metricNameSQLServerDbmActivity = "sqlserver_dbm_activity"
)

type dbmActivityMeasurement struct {
	name     string
	tags     map[string]string
	fields   map[string]interface{}
	election bool
}

// Point implement MeasurementV2.
func (m *dbmActivityMeasurement) Point() *point.Point {
	opts := point.DefaultLoggingOptions()

	if m.election {
		opts = append(opts, point.WithExtraTags(datakit.GlobalElectionTags()))
	}

	return point.NewPoint(m.name,
		append(point.NewTags(m.tags), point.NewKVs(m.fields)...),
		opts...)
}

func (m *dbmActivityMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Desc:   "Collect the currently active queries, session information, wait events, and blocking information.",
		DescZh: "收集当前活动的查询、会话信息、等待事件和阻塞信息",
		Name:   metricNameSQLServerDbmActivity,
		Cat:    point.Logging,
		Fields: map[string]interface{}{
			"message": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.String,
				Unit:     inputs.NoUnit,
				Desc:     "The text of the normalized SQL text",
			},
			"session_id": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The session ID",
			},
			"percent_complete": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The percentage of work completed for the current operation.",
			},
			"estimated_completion_time": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.DurationMS,
				Desc:     "The estimated completion time (milliseconds) for the current operation.",
			},
			"wait_time": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.DurationMS,
				Desc:     "The wait time in milliseconds",
			},
			"wait_resource": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.String,
				Unit:     inputs.NoUnit,
				Desc:     "The resource being waited on",
			},
			"blocking_session_id": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The ID of the blocking session. 0 means not blocked.",
			},
			"last_wait_type": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.String,
				Unit:     inputs.NoUnit,
				Desc:     "The last wait type encountered by this request.",
			},
			"open_transaction_count": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The number of open transactions for the session.",
			},
			"transaction_id": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The ID of the transaction that the request is part of.",
			},
			"transaction_isolation_level": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NoUnit,
				Desc:     "The transaction isolation level of the request.",
			},
			"lock_timeout": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.DurationMS,
				Desc:     "The lock timeout period in milliseconds.",
			},
			"deadlock_priority": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NoUnit,
				Desc:     "The deadlock priority for the session.",
			},
			"context_info": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.String,
				Unit:     inputs.UnknownUnit,
				Desc:     "The CONTEXT_INFO value for the session.",
			},
			"cpu_time": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.DurationMS,
				Desc:     "The CPU time in milliseconds",
			},
			"total_elapsed_time": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.DurationMS,
				Desc:     "The total elapsed time in milliseconds",
			},
			"reads": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The number of reads",
			},
			"writes": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The number of writes",
			},
			"logical_reads": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The number of logical reads",
			},
			"row_count": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The number of rows",
			},
			"client_address": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.String,
				Unit:     inputs.NoUnit,
				Desc:     "The client network address",
			},
			"client_port": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.String,
				Unit:     inputs.NoUnit,
				Desc:     "The client TCP port",
			},
			"query_start": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.String,
				Unit:     inputs.NoUnit,
				Desc:     "The time when the request started executing.",
			},
			"last_request_start_time": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.DurationMS,
				Desc:     "The last time a request started in the session.",
			},
			"is_user_process": &inputs.FieldInfo{
				DataType: inputs.Bool,
				Type:     inputs.Bool,
				Unit:     inputs.UnknownUnit,
				Desc:     "Indicates whether the session is a user process (1) or a system process (0).",
			},
			"client_interface_name": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.String,
				Unit:     inputs.NoUnit,
				Desc:     "The name of the client interface used by the application (e.g., 'Microsoft JDBC Driver for SQL Server', 'ODBC').",
			},
			"procedure_text": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.String,
				Unit:     inputs.NoUnit,
				Desc:     "The obfuscated/normalized stored procedure text (if applicable).",
			},
		},
		Tags: map[string]interface{}{
			"server":          inputs.NewTagInfo("The server address (host:port)"),
			"sqlserver_host":  inputs.NewTagInfo("Host name which installed SQLServer"),
			"database_name":   inputs.NewTagInfo("The name of the database"),
			"user_name":       inputs.NewTagInfo("The login name of the user"),
			"host_name":       inputs.NewTagInfo("The host name of the client"),
			"procedure_name":  inputs.NewTagInfo("The name of the stored procedure in the format 'schema_name.procedure_name' (if applicable)"),
			"schema_name":     inputs.NewTagInfo("The schema name of the stored procedure (if applicable)"),
			"program_name":    inputs.NewTagInfo("The name of the client program"),
			"query_hash":      inputs.NewTagInfo("The hash value computed from the query."),
			"query_plan_hash": inputs.NewTagInfo("The hash value computed from the query plan."),
			"session_status":  inputs.NewTagInfo("The status of the session."),
			"request_status":  inputs.NewTagInfo("The status of the request."),
			"command":         inputs.NewTagInfo("The type of command being executed."),
			"wait_type":       inputs.NewTagInfo("The type of wait."),
			"wait_group":      inputs.NewTagInfo("Datakit unified wait group: Lock, I/O, Concurrency, Memory, Network, CPU, Commit/Log, Other."),
			"query_signature": inputs.NewTagInfo("Hash signature generated from database_name:procedure_name:query_hash to link metrics and objects"),
		},
	}
}

type dbmActivityRow struct {
	querySignature string

	// Identifiers
	sessionID     int64
	userName      string
	hostName      string
	databaseName  string
	sqlserverHost string

	// SQL information
	obfuscatedText string
	schemaName     string
	procedureName  string
	procedureText  string // obfuscated stored procedure text
	queryHash      string
	queryPlanHash  string

	// Status
	sessionStatus string
	requestStatus string
	command       string

	// Wait information
	waitType          string
	waitCategory      string // Categorized wait type (Lock, I/O, Concurrency, Memory, Network, CPU, Other)
	waitTime          int64
	waitResource      string
	blockingSessionID int64
	lastWaitType      string

	// Transaction information
	openTransactionCount      int64
	transactionID             int64
	transactionIsolationLevel int64
	lockTimeout               int64
	deadlockPriority          int64

	// Progress information
	percentComplete         int64
	estimatedCompletionTime int64

	// Resource usage
	cpuTime          int64
	totalElapsedTime int64
	reads            int64
	writes           int64
	logicalReads     int64
	rowCount         int64

	// Connection information
	clientAddress       string
	clientPort          string
	programName         string
	isUserProcess       bool
	clientInterfaceName string

	// Timestamps
	queryStart           string
	lastRequestStartTime time.Time

	// Additional metadata
	contextInfo string
}

func (ipt *Input) collectDbmActivity(duration time.Duration, ptsTime time.Time) ([]*point.Point, error) {
	if ipt.Dbm == nil || ipt.Dbm.Activity == nil || !ipt.Dbm.Activity.Enabled {
		return nil, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), duration)
	defer cancel()

	// Build query dynamically based on available columns
	query, err := ipt.buildActivityQueryWithColumns(ctx, ipt.Dbm.Activity.DmExecSessionsRowLimit)
	if err != nil {
		return nil, fmt.Errorf("failed to build activity query: %w", err)
	}

	// Measure database query time for main query
	queryStart := time.Now()
	rows, err := ipt.db.QueryContext(ctx, query)
	dbmSQLQueryDuration.WithLabelValues("activity", "query").Observe(time.Since(queryStart).Seconds())
	if err != nil {
		return nil, fmt.Errorf("failed to query activity: %w", err)
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

	var activeRows []*dbmActivityRow
	obfuscator := obfuscate.NewObfuscator(obfuscate.Config{
		SQL: obfuscate.SQLConfig{
			DBMS: obfuscate.DBMSSQLServer,
		},
	})
	for rows.Next() {
		columnMap, err := GetColumnMap(rows, columns)
		if err != nil {
			l.Errorf("failed to get column map: %v", err)
			continue
		}

		row, err := buildDbmActivityRow(columnMap, obfuscator)
		if err != nil {
			l.Errorf("%v", err)
			continue
		}
		if row == nil {
			continue // Skip empty statement text
		}

		activeRows = append(activeRows, row)
	}
	if err := rows.Err(); err != nil {
		l.Errorf("rows error: %v", err)
		return nil, err
	}
	if len(activeRows) == 0 {
		return nil, nil
	}

	// Find idle blocking sessions
	// Note: blocking_session_id = 0 means "not blocked" in SQL Server, so we exclude 0
	sessionIDs := make(map[int64]struct{})
	blockingSessionIDs := make(map[int64]struct{})
	for _, row := range activeRows {
		if row.sessionID > 0 {
			sessionIDs[row.sessionID] = struct{}{}
		}
		if row.blockingSessionID > 0 {
			blockingSessionIDs[row.blockingSessionID] = struct{}{}
		}
	}

	idleBlockingSessionIDs := make([]int64, 0, len(blockingSessionIDs))
	for id := range blockingSessionIDs {
		if _, ok := sessionIDs[id]; !ok {
			idleBlockingSessionIDs = append(idleBlockingSessionIDs, id)
		}
	}

	// Build points for active sessions
	pts := ipt.buildActivityPoints(activeRows, ptsTime)

	// Handle idle blocking sessions
	var idleRows []*dbmActivityRow
	if len(idleBlockingSessionIDs) > 0 {
		l.Debugf("idle blocking session ids: %v", idleBlockingSessionIDs)
		idleQueryStart := time.Now()
		idleRows, err = ipt.getIdleBlockingSessions(ctx, idleBlockingSessionIDs)
		dbmSQLQueryDuration.WithLabelValues("activity", "idle_blocking").Observe(time.Since(idleQueryStart).Seconds())
		if err != nil {
			l.Warnf("failed to get idle blocking sessions: %v", err)
		} else {
			idlePts := ipt.buildActivityPoints(idleRows, ptsTime)
			if len(idlePts) > 0 {
				pts = append(pts, idlePts...)
			}
		}
	}

	// Collect session metrics from all activity rows (active + idle blocking)
	activeRows = append(activeRows, idleRows...)
	ipt.collectDbmSessionMetrics(activeRows, ptsTime)

	// Collect connection metrics
	if err := ipt.collectDbmConnections(ctx, ptsTime); err != nil {
		l.Errorf("collectDbmConnections failed: %s", err.Error())
	}

	return pts, nil
}

func buildDbmActivityRow(columnMap map[string]*interface{}, obfuscator *obfuscate.Obfuscator) (*dbmActivityRow, error) {
	row := &dbmActivityRow{}

	// Extract basic fields
	row.sessionID = getInt64Field(columnMap, "id")
	row.userName = getStringField(columnMap, "user_name")
	row.hostName = getStringField(columnMap, "host_name")
	row.databaseName = getStringField(columnMap, "database_name")
	row.sqlserverHost = getStringField(columnMap, "sqlserver_host")

	// Extract statement text
	statementText := getStringField(columnMap, "statement_text")
	if statementText == "" {
		l.Debugf("skip activity row due to empty statement text")
		return nil, nil
	}

	// Obfuscate SQL and compute signature
	obfStart := time.Now()
	obfuscated, err := obfuscator.ObfuscateSQLString(statementText)
	obfuscateTime := time.Since(obfStart)
	dbmObfuscateDuration.WithLabelValues("activity", "statement").Observe(obfuscateTime.Seconds())
	if err != nil {
		l.Warnf("failed to obfuscate SQL statement: %v", err)
		row.obfuscatedText = statementText
	} else {
		row.obfuscatedText = obfuscated.Query
	}
	// Extract other fields (status, command, wait info, resource usage, connection info)
	row.sessionStatus = getStringField(columnMap, "session_status")
	row.requestStatus = getStringField(columnMap, "request_status")
	row.command = getStringField(columnMap, "command")

	// NULL values default to 0
	row.percentComplete = getInt64Field(columnMap, "percent_complete")
	row.estimatedCompletionTime = getInt64Field(columnMap, "estimated_completion_time")

	// Wait information
	row.waitType = getStringField(columnMap, "wait_type")
	row.waitTime = getInt64Field(columnMap, "wait_time")
	row.waitResource = getStringField(columnMap, "wait_resource")
	row.blockingSessionID = getInt64Field(columnMap, "blocking_session_id")

	// Calculate wait category
	row.waitCategory = categorizeWaitType(row.sessionStatus, row.waitType)

	// Additional fields from sys.dm_exec_requests
	row.lastWaitType = getStringField(columnMap, "last_wait_type")
	row.openTransactionCount = getInt64Field(columnMap, "open_transaction_count")
	row.transactionID = getInt64Field(columnMap, "transaction_id")
	row.transactionIsolationLevel = getInt64Field(columnMap, "transaction_isolation_level")
	row.lockTimeout = getInt64Field(columnMap, "lock_timeout")
	row.deadlockPriority = getInt64Field(columnMap, "deadlock_priority")
	queryHash := getBytesField(columnMap, "query_hash")
	queryPlanHash := getBytesField(columnMap, "query_plan_hash")
	if len(queryHash) > 0 {
		row.queryHash = "0x" + hex.EncodeToString(queryHash)
	}
	if len(queryPlanHash) > 0 {
		row.queryPlanHash = "0x" + hex.EncodeToString(queryPlanHash)
	}
	row.contextInfo = getStringField(columnMap, "context_info")

	// Resource usage
	row.cpuTime = getInt64Field(columnMap, "cpu_time")
	row.totalElapsedTime = getInt64Field(columnMap, "total_elapsed_time")
	row.reads = getInt64Field(columnMap, "reads")
	row.writes = getInt64Field(columnMap, "writes")
	row.logicalReads = getInt64Field(columnMap, "logical_reads")
	row.rowCount = getInt64Field(columnMap, "row_count")

	// Connection information
	row.clientAddress = getStringField(columnMap, "client_address")
	row.clientPort = getStringFromInt64(columnMap, "client_port")
	row.programName = getStringField(columnMap, "program_name")
	row.isUserProcess = getBoolField(columnMap, "is_user_process")
	row.clientInterfaceName = getStringField(columnMap, "client_interface_name")

	// Timestamps
	row.queryStart = getStringField(columnMap, "query_start")
	row.lastRequestStartTime = getTimeField(columnMap, "last_request_start_time")

	// Procedure name (merge schema_name and procedure_name)
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
	row.querySignature = generateQuerySignature(row.databaseName, row.procedureName, row.queryHash)

	// Extract and obfuscate stored procedure text if available
	// The 'text' field contains the full stored procedure text (limited by stored_procedure_characters_limit)
	procedureText := getStringField(columnMap, "text")
	if procedureText != "" && row.procedureName != "" {
		// Obfuscate stored procedure text
		procObfStart := time.Now()
		obfResult, err := obfuscator.ObfuscateSQLString(procedureText)
		procObfuscateTime := time.Since(procObfStart)
		dbmObfuscateDuration.WithLabelValues("activity", "procedure").Observe(procObfuscateTime.Seconds())
		if err != nil {
			l.Warnf("failed to obfuscate stored procedure text: %v", err)
			// Continue even if obfuscation fails
		} else {
			row.procedureText = obfResult.Query
		}
	}

	return row, nil
}

type rawIdleBlockingSessionRow struct {
	Now                  sql.NullString
	UserName             sql.NullString
	LastRequestStartTime sql.NullTime
	ID                   sql.NullInt64
	DatabaseName         sql.NullString
	SessionStatus        sql.NullString
	StatementText        sql.NullString
	SchemaName           sql.NullString
	ProcedureName        sql.NullString
	SqlserverHost        sql.NullString
	ClientPort           sql.NullInt64
	ClientAddress        sql.NullString
	HostName             sql.NullString
	ProgramName          sql.NullString
	IsUserProcess        sql.NullBool
	ClientInterfaceName  sql.NullString
}

func (ipt *Input) getIdleBlockingSessions(ctx context.Context, blockingSessionIDs []int64) ([]*dbmActivityRow, error) {
	query := ipt.buildIdleBlockingSessionsQuery(blockingSessionIDs)
	if query == "" {
		return nil, nil
	}
	rows, err := ipt.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query idle blocking sessions: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			l.Errorf("failed to close rows: %v", err)
		}
	}()

	var idleRows []*dbmActivityRow
	obfuscator := obfuscate.NewObfuscator(obfuscate.Config{
		SQL: obfuscate.SQLConfig{
			DBMS: obfuscate.DBMSSQLServer,
		},
	})

	for rows.Next() {
		var rawRow rawIdleBlockingSessionRow
		if err := rows.Scan(
			&rawRow.Now,
			&rawRow.UserName,
			&rawRow.LastRequestStartTime,
			&rawRow.ID,
			&rawRow.DatabaseName,
			&rawRow.SessionStatus,
			&rawRow.StatementText,
			&rawRow.SchemaName,
			&rawRow.ProcedureName,
			&rawRow.SqlserverHost,
			&rawRow.ClientPort,
			&rawRow.ClientAddress,
			&rawRow.HostName,
			&rawRow.ProgramName,
			&rawRow.IsUserProcess,
			&rawRow.ClientInterfaceName,
		); err != nil {
			l.Errorf("failed to scan idle blocking session row: %v", err)
			continue
		}

		row, err := buildIdleBlockingActivityRow(&rawRow, obfuscator)
		if err != nil {
			l.Errorf("%v", err)
			continue
		}
		if row == nil {
			continue
		}

		idleRows = append(idleRows, row)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return idleRows, nil
}

func buildIdleBlockingActivityRow(rawRow *rawIdleBlockingSessionRow, obfuscator *obfuscate.Obfuscator) (*dbmActivityRow, error) {
	row := &dbmActivityRow{}

	// Extract basic fields
	if !rawRow.ID.Valid {
		return nil, fmt.Errorf("idle blocking session row has no session ID")
	}
	row.sessionID = rawRow.ID.Int64
	row.userName = rawRow.UserName.String
	row.hostName = rawRow.HostName.String
	row.databaseName = rawRow.DatabaseName.String
	row.sqlserverHost = rawRow.SqlserverHost.String

	// Extract statement text
	statementText := rawRow.StatementText.String
	if statementText == "" {
		l.Warnf("skip idle blocking session row due to empty statement text")
		return nil, nil
	}

	// Obfuscate SQL and compute signature
	obfStart := time.Now()
	obfuscated, err := obfuscator.ObfuscateSQLString(statementText)
	obfuscateTime := time.Since(obfStart)
	dbmObfuscateDuration.WithLabelValues("activity", "statement").Observe(obfuscateTime.Seconds())
	if err != nil {
		l.Warnf("failed to obfuscate SQL statement: %v", err)
		row.obfuscatedText = statementText
	} else {
		row.obfuscatedText = obfuscated.Query
	}

	// Extract other fields (idle blocking sessions don't have request_status, query_start, etc.)
	row.sessionStatus = rawRow.SessionStatus.String

	// Idle blocking sessions don't have wait_type (they are sleeping)
	row.waitType = ""

	// Calculate wait category
	row.waitCategory = categorizeWaitType(row.sessionStatus, row.waitType)

	// Procedure name (merge schema_name and procedure_name)
	procedureName := rawRow.ProcedureName.String
	schemaName := rawRow.SchemaName.String
	row.schemaName = schemaName
	if procedureName != "" {
		if schemaName != "" {
			row.procedureName = fmt.Sprintf("%s.%s", schemaName, procedureName)
		} else {
			row.procedureName = procedureName
		}
	}

	// Connection information
	row.clientAddress = rawRow.ClientAddress.String
	if rawRow.ClientPort.Int64 != 0 {
		row.clientPort = fmt.Sprintf("%d", rawRow.ClientPort.Int64)
	}
	row.programName = rawRow.ProgramName.String
	row.isUserProcess = rawRow.IsUserProcess.Bool
	row.clientInterfaceName = rawRow.ClientInterfaceName.String

	// Timestamps
	row.lastRequestStartTime = rawRow.LastRequestStartTime.Time

	return row, nil
}

// buildActivityPoints builds point.Point from activity rows.
func (ipt *Input) buildActivityPoints(rows []*dbmActivityRow, ptsTime time.Time) []*point.Point {
	var pts []*point.Point
	opts := append(point.DefaultLoggingOptions(), point.WithTime(ptsTime))

	for _, row := range rows {
		kvs := ipt.getKVs()

		// Set Tags
		kvs = kvs.AddTag("database_name", row.databaseName)
		kvs = kvs.AddTag("sqlserver_host", row.sqlserverHost)

		kvs = kvs.AddTag("query_signature", row.querySignature)
		if row.queryHash != "" {
			kvs = kvs.AddTag("query_hash", row.queryHash)
		}
		if row.queryPlanHash != "" {
			kvs = kvs.AddTag("query_plan_hash", row.queryPlanHash)
		}
		if row.userName != "" {
			kvs = kvs.AddTag("user_name", row.userName)
		}
		if row.hostName != "" {
			kvs = kvs.AddTag("host_name", row.hostName)
		}
		if row.procedureName != "" {
			kvs = kvs.AddTag("procedure_name", row.procedureName)
		}
		if row.schemaName != "" {
			kvs = kvs.AddTag("schema_name", row.schemaName)
		}
		if row.programName != "" {
			kvs = kvs.AddTag("program_name", row.programName)
		}
		if row.sessionStatus != "" {
			kvs = kvs.AddTag("session_status", row.sessionStatus)
		}
		if row.requestStatus != "" {
			kvs = kvs.AddTag("request_status", row.requestStatus)
		}
		if row.command != "" {
			kvs = kvs.AddTag("command", row.command)
		}
		if row.waitType != "" {
			kvs = kvs.AddTag("wait_type", row.waitType)
		}
		if row.waitCategory != "" {
			kvs = kvs.AddTag("wait_group", row.waitCategory)
		}

		// ========== Common fields (Session level - both active and idle blocking sessions) ==========
		kvs = kvs.Set("message", row.obfuscatedText)
		kvs = kvs.Set("procedure_text", row.procedureText)
		kvs = kvs.Set("session_id", row.sessionID)
		kvs = kvs.Set("client_address", row.clientAddress)
		kvs = kvs.Set("client_port", row.clientPort)
		kvs = kvs.Set("is_user_process", row.isUserProcess)
		kvs = kvs.Set("client_interface_name", row.clientInterfaceName)
		var lastRequestStartTime int64
		if !row.lastRequestStartTime.IsZero() {
			lastRequestStartTime = row.lastRequestStartTime.UnixNano() / int64(time.Millisecond)
		}
		kvs = kvs.Set("last_request_start_time", lastRequestStartTime)

		// ========== Active session only fields (Request level - not available for idle blocking sessions) ==========
		// Note: These fields will be zero/empty for idle blocking sessions
		kvs = kvs.Set("query_start", row.queryStart)
		kvs = kvs.Set("wait_time", row.waitTime)
		kvs = kvs.Set("wait_resource", row.waitResource)
		kvs = kvs.Set("blocking_session_id", row.blockingSessionID)
		kvs = kvs.Set("last_wait_type", row.lastWaitType)
		kvs = kvs.Set("open_transaction_count", row.openTransactionCount)
		kvs = kvs.Set("transaction_id", row.transactionID)
		kvs = kvs.Set("transaction_isolation_level", row.transactionIsolationLevel)
		kvs = kvs.Set("lock_timeout", row.lockTimeout)
		kvs = kvs.Set("deadlock_priority", row.deadlockPriority)
		kvs = kvs.Set("context_info", row.contextInfo)
		kvs = kvs.Set("cpu_time", row.cpuTime)
		kvs = kvs.Set("total_elapsed_time", row.totalElapsedTime)
		kvs = kvs.Set("reads", row.reads)
		kvs = kvs.Set("writes", row.writes)
		kvs = kvs.Set("logical_reads", row.logicalReads)
		kvs = kvs.Set("row_count", row.rowCount)
		kvs = kvs.Set("percent_complete", row.percentComplete)
		kvs = kvs.Set("estimated_completion_time", row.estimatedCompletionTime)

		pts = append(pts, point.NewPoint(metricNameSQLServerDbmActivity, kvs, opts...))
	}

	return pts
}
