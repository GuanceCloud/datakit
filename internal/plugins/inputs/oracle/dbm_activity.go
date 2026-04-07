// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package oracle

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"

	"github.com/DataDog/datadog-agent/pkg/obfuscate"
)

const (
	metricNameOracleDbmActivity = "oracle_dbm_activity"
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
		Desc:   "Collect the currently active queries, session information, wait events, and blocking information for Oracle.",
		DescZh: "收集 Oracle 当前活动的查询、会话信息、等待事件和阻塞信息",
		Name:   metricNameOracleDbmActivity,
		Cat:    point.Logging,
		Fields: map[string]interface{}{
			"message": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.String,
				Unit:     inputs.NoUnit,
				Desc:     "The text of the normalized/obfuscated SQL text",
			},
			"session_id": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The session ID (SID)",
			},
			"serial_number": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The session serial number (SERIAL#)",
			},
			"wait_time": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.DurationUS,
				Desc:     "The wait time in microseconds",
			},
			"blocking_session_id": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The blocking session ID. 0 means not blocked.",
			},
			"final_blocking_session_id": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The final blocking session ID in the blocking chain. 0 means not blocked.",
			},
			"event": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.String,
				Unit:     inputs.NoUnit,
				Desc:     "The wait event name",
			},
			"client_port": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.String,
				Unit:     inputs.NoUnit,
				Desc:     "The client TCP port",
			},
			"process": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.String,
				Unit:     inputs.NoUnit,
				Desc:     "The client process ID",
			},
			"client_info": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.String,
				Unit:     inputs.NoUnit,
				Desc:     "The client information set by DBMS_APPLICATION_INFO",
			},
			"client_identifier": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.String,
				Unit:     inputs.NoUnit,
				Desc:     "The client identifier",
			},
			"logon_time": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.String,
				Unit:     inputs.NoUnit,
				Desc:     "The time when the session logged on",
			},
			"sql_exec_start": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.String,
				Unit:     inputs.NoUnit,
				Desc:     "The time when the SQL statement started executing",
			},
			"command_name": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.String,
				Unit:     inputs.NoUnit,
				Desc:     "The command name (e.g., SELECT, INSERT, UPDATE)",
			},
			"blocking_instance": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The blocking instance ID (for RAC). 0 means not blocked.",
			},
			"final_blocking_instance": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The final blocking instance ID in the blocking chain (for RAC). 0 means not blocked.",
			},
		},
		Tags: map[string]interface{}{
			"server":                   inputs.NewTagInfo("The server address (host:port)"),
			"database_instance":        inputs.NewTagInfo("Oracle instance identifier from configured tag or v$instance.host_name."),
			"cdb_name":                 inputs.NewTagInfo("The name of the CDB (Container Database)"),
			"con_id":                   inputs.NewTagInfo("The container ID (con_id) in Oracle multi tenant architecture"),
			"pdb_name":                 inputs.NewTagInfo("The name of the PDB (Pluggable Database)"),
			"username":                 inputs.NewTagInfo("The Oracle username"),
			"program":                  inputs.NewTagInfo("The program name"),
			"machine":                  inputs.NewTagInfo("The machine name"),
			"terminal":                 inputs.NewTagInfo("The terminal name"),
			"module":                   inputs.NewTagInfo("The module name"),
			"action":                   inputs.NewTagInfo("The action name"),
			"service_name":             inputs.NewTagInfo("The service name"),
			"sql_id":                   inputs.NewTagInfo("The SQL ID"),
			"force_matching_signature": inputs.NewTagInfo("The force matching signature"),
			"plan_hash_value":          inputs.NewTagInfo("The plan hash value"),
			"query_signature":          inputs.NewTagInfo("Hash signature generated from normalized SQL text to link metrics and objects"),
			"session_status":           inputs.NewTagInfo("The session status (ACTIVE, INACTIVE, KILLED, etc.)"),
			"session_type":             inputs.NewTagInfo("The session type (USER, BACKGROUND)"),
			"wait_class":               inputs.NewTagInfo("The wait event class."),
			"wait_group":               inputs.NewTagInfo("Datakit unified wait group: Lock, I/O, Concurrency, Memory, Network, CPU, Commit/Log, Other."),
		},
	}
}

type OracleSQLRow struct {
	SQLID                  string `json:"sql_id,omitempty"`
	ForceMatchingSignature uint64 `json:"force_matching_signature,omitempty"`
	SQLPlanHashValue       uint64 `json:"sql_plan_hash_value,omitempty"`
	SQLExecStart           string `json:"sql_exec_start,omitempty"`
}

//nolint:revive // TODO(DBM) Fix revive linter
type OracleActivityRow struct {
	Now           string `json:"now"`
	UtcMs         float64
	SessionID     uint64 `json:"sid,omitempty"`
	SessionSerial uint64 `json:"serial,omitempty"`
	User          string `json:"user,omitempty"`
	Status        string `json:"status"`
	OsUser        string `json:"os_user,omitempty"`
	Process       string `json:"process,omitempty"`
	Client        string `json:"client,omitempty"`
	Port          string `json:"port,omitempty"`
	Program       string `json:"program,omitempty"`
	Type          string `json:"type,omitempty"`
	OracleSQLRow
	Module                string `json:"module,omitempty"`
	Action                string `json:"action,omitempty"`
	ClientInfo            string `json:"client_info,omitempty"`
	LogonTime             string `json:"logon_time,omitempty"`
	ClientIdentifier      string `json:"client_identifier,omitempty"`
	BlockingInstance      uint64 `json:"blocking_instance,omitempty"`
	BlockingSession       uint64 `json:"blocking_session,omitempty"`
	FinalBlockingInstance uint64 `json:"final_blocking_instance,omitempty"`
	FinalBlockingSession  uint64 `json:"final_blocking_session,omitempty"`
	WaitEvent             string `json:"wait_event,omitempty"`
	WaitEventClass        string `json:"wait_event_class,omitempty"`
	WaitTimeMicro         uint64 `json:"wait_time_micro,omitempty"`
	Statement             string `json:"statement,omitempty"`
	PdbName               string `json:"pdb_name,omitempty"`
	CdbName               string `json:"cdb_name,omitempty"`
	QuerySignature        string `json:"query_signature,omitempty"`
	CommandName           string `json:"command_name,omitempty"`
	PreviousSQL           bool   `json:"previous_sql,omitempty"`
	OpFlags               uint64 `json:"op_flags,omitempty"`
}

type OracleActivityRowDB struct {
	SampleID                   uint64         `db:"SAMPLE_ID"`
	Now                        string         `db:"NOW"`
	UtcMs                      float64        `db:"UTC_MS"`
	SessionID                  uint64         `db:"SID"`
	SessionSerial              uint64         `db:"SERIAL#"`
	User                       sql.NullString `db:"USERNAME"`
	Status                     string         `db:"STATUS"`
	OsUser                     sql.NullString `db:"OSUSER"`
	Process                    sql.NullString `db:"PROCESS"`
	Client                     sql.NullString `db:"MACHINE"`
	Port                       sql.NullInt64  `db:"PORT"`
	Program                    sql.NullString `db:"PROGRAM"`
	Type                       sql.NullString `db:"TYPE"`
	SQLID                      sql.NullString `db:"SQL_ID"`
	ForceMatchingSignature     *string        `db:"FORCE_MATCHING_SIGNATURE"`
	SQLPlanHashValue           *uint64        `db:"SQL_PLAN_HASH_VALUE"`
	SQLExecStart               sql.NullString `db:"SQL_EXEC_START"`
	SQLAddress                 sql.NullString `db:"SQL_ADDRESS"`
	PrevSQLID                  sql.NullString `db:"PREV_SQL_ID"`
	PrevForceMatchingSignature *string        `db:"PREV_FORCE_MATCHING_SIGNATURE"`
	PrevSQLPlanHashValue       *uint64        `db:"PREV_SQL_PLAN_HASH_VALUE"`
	PrevSQLExecStart           sql.NullString `db:"PREV_SQL_EXEC_START"`
	PrevSQLAddress             sql.NullString `db:"PREV_SQL_ADDRESS"`
	Module                     sql.NullString `db:"MODULE"`
	Action                     sql.NullString `db:"ACTION"`
	ClientInfo                 sql.NullString `db:"CLIENT_INFO"`
	LogonTime                  sql.NullString `db:"LOGON_TIME"`
	ClientIdentifier           sql.NullString `db:"CLIENT_IDENTIFIER"`
	OpFlags                    uint64         `db:"OP_FLAGS"`
	BlockingInstance           *uint64        `db:"BLOCKING_INSTANCE"`
	BlockingSession            *uint64        `db:"BLOCKING_SESSION"`
	FinalBlockingInstance      *uint64        `db:"FINAL_BLOCKING_INSTANCE"`
	FinalBlockingSession       *uint64        `db:"FINAL_BLOCKING_SESSION"`
	WaitEvent                  sql.NullString `db:"EVENT"`
	WaitEventClass             sql.NullString `db:"WAIT_CLASS"`
	WaitTimeMicro              *uint64        `db:"WAIT_TIME_MICRO"`
	Statement                  sql.NullString `db:"SQL_FULLTEXT"`
	PrevSQLFullText            sql.NullString `db:"PREV_SQL_FULLTEXT"`
	PdbName                    sql.NullString `db:"PDB_NAME"`
	CommandName                sql.NullString `db:"COMMAND_NAME"`
}

// collectDbmActivity collects DBM activity (currently active queries) from Oracle.
func (ipt *Input) collectDbmActivity(duration time.Duration, ptsTime time.Time) ([]*point.Point, error) {
	ctx, cancel := context.WithTimeout(context.Background(), duration)
	defer cancel()

	if ipt.Dbm == nil || ipt.Dbm.Activity == nil || !ipt.Dbm.Activity.Enabled {
		return nil, nil
	}

	start := time.Now()

	// Sample sessions based on configuration
	activityRows, err := ipt.sampleSession(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to sample sessions: %w", err)
	}
	l.Debugf("activityRows: %d", len(activityRows))

	if len(activityRows) == 0 {
		return nil, nil
	}

	// Build points from activity rows
	pts := ipt.buildDbmActivityPoints(activityRows, ptsTime)
	if len(pts) == 0 {
		return nil, nil
	}

	// Collect session metrics (aggregated statistics)
	ipt.collectDbmSessionMetrics(activityRows, ptsTime)

	// Collect connection metrics
	if err := ipt.collectDbmConnections(ctx, ptsTime); err != nil {
		l.Errorf("collectDbmConnections failed: %s", err.Error())
	}

	l.Debugf("collectDbmActivity completed, collected %d rows, time taken: %s", len(activityRows), time.Since(start))

	return pts, nil
}

// sampleSession samples active sessions from Oracle based on configuration.
//
//nolint:funlen
func (ipt *Input) sampleSession(ctx context.Context) ([]*OracleActivityRow, error) {
	var sessionRows []*OracleActivityRow
	sessionSamples := []OracleActivityRowDB{}
	var activityQuery string
	maxSQLTextLength := ipt.sqlSubstringLength

	isDBVersionGreaterOrEqualThan := isDBVersionGreaterOrEqualThan(ipt.dbVersion, "12")
	if isDBVersionGreaterOrEqualThan {
		activityQuery = activityQueryDirect12
	} else {
		activityQuery = activityQueryDirect11
	}

	if !ipt.Dbm.Activity.IncludeAllSessions {
		activityQuery = fmt.Sprintf("%s%s", activityQuery, ` AND (
    -- Condition A: Active and non-idle sessions (Waiters or busy sessions)
    (
        s.status = 'ACTIVE' 
        AND (
            NOT (s.state = 'WAITING' AND s.wait_class = 'Idle') 
            OR (s.state = 'WAITING' AND s.event = 'fbar timer' AND s.type = 'USER')
        )
    )
    OR 
    -- Condition B: Blocking sources (Root cause), including INACTIVE sessions
    s.sid IN (SELECT b.blocking_session FROM v$session b WHERE b.blocking_session IS NOT NULL)
)`)
	}

	limitCondition := ""
	if isDBVersionGreaterOrEqualThan {
		limitCondition = `
			FETCH FIRST :limit ROWS ONLY`
	} else {
		limitCondition = `
			AND ROWNUM <= :limit`
	}
	activityQuery = fmt.Sprintf("%s %s", activityQuery, limitCondition)

	// Get DBRowsLimit, default to 1000 if not configured
	dbRowsLimit := ipt.Dbm.Activity.DBRowsLimit
	if dbRowsLimit <= 0 {
		dbRowsLimit = 1000
	}

	queryStart := time.Now()
	err := selectWrapperWithBinds(ipt, ctx, &sessionSamples, activityQuery, maxSQLTextLength, maxSQLTextLength, dbRowsLimit)
	dbmSQLQueryDuration.WithLabelValues("activity", "direct_query").Observe(time.Since(queryStart).Seconds())
	if err != nil {
		if strings.Contains(err.Error(), "ORA-06502") {
			if maxSQLTextLength > 1000 {
				ipt.sqlSubstringLength = maxInt(maxSQLTextLength-500, 1000)
				l.Infof("decrease sql substring length to %d", ipt.sqlSubstringLength)
				return nil, nil
			}
		}
		return nil, fmt.Errorf("failed to collect session sampling activity: %w \n%s", err, activityQuery)
	}

	o := obfuscate.NewObfuscator(obfuscate.Config{})
	for _, sample := range sessionSamples {
		sessionRow := &OracleActivityRow{}

		sessionRow.Now = sample.Now
		sessionRow.UtcMs = sample.UtcMs

		sessionRow.SessionID = sample.SessionID
		sessionRow.SessionSerial = sample.SessionSerial
		if sample.User.Valid {
			sessionRow.User = sample.User.String
		}
		sessionRow.Status = sample.Status
		if sample.OsUser.Valid {
			sessionRow.OsUser = sample.OsUser.String
		}
		if sample.Process.Valid {
			sessionRow.Process = sample.Process.String
		}
		if sample.Client.Valid {
			sessionRow.Client = sample.Client.String
		}
		if sample.Port.Valid {
			sessionRow.Port = strconv.FormatInt(sample.Port.Int64, 10)
		}

		program := ""
		if sample.Program.Valid {
			sessionRow.Program = sample.Program.String
			program = sample.Program.String
		}

		sessionType := ""
		if sample.Type.Valid {
			if sample.Type.String == "FOREGROUND" {
				sample.Type.String = "USER"
			}
			sessionRow.Type = sample.Type.String
			sessionType = sample.Type.String
		}

		commandName := ""
		if sample.CommandName.Valid {
			commandName = sample.CommandName.String
		}
		sessionRow.CommandName = commandName
		previousSQL := false
		sqlCurrentSQL, err := ipt.getSQLRow(sample.SQLID, sample.ForceMatchingSignature, sample.SQLPlanHashValue, sample.SQLExecStart)
		if err != nil {
			l.Errorf("error getting SQL row %s", err)
		}

		var sqlPrevSQL OracleSQLRow
		if sqlCurrentSQL.SQLID != "" {
			sessionRow.OracleSQLRow = sqlCurrentSQL
		} else {
			sqlPrevSQL, err = ipt.getSQLRow(sample.PrevSQLID, sample.PrevForceMatchingSignature, sample.PrevSQLPlanHashValue, sample.PrevSQLExecStart)
			if err != nil {
				l.Errorf("error getting SQL row %s", err)
			}
			if sqlPrevSQL.SQLID != "" {
				sessionRow.OracleSQLRow = sqlPrevSQL
				previousSQL = true
			}
		}
		sessionRow.PreviousSQL = previousSQL

		if sample.Module.Valid {
			sessionRow.Module = sample.Module.String
		}
		if sample.Action.Valid {
			sessionRow.Action = sample.Action.String
		}
		if sample.ClientInfo.Valid {
			sessionRow.ClientInfo = sample.ClientInfo.String
		}
		if sample.LogonTime.Valid {
			sessionRow.LogonTime = sample.LogonTime.String
		}
		if sample.ClientIdentifier.Valid {
			sessionRow.ClientIdentifier = sample.ClientIdentifier.String
		}
		if sample.BlockingInstance != nil {
			sessionRow.BlockingInstance = *sample.BlockingInstance
		}
		if sample.BlockingSession != nil {
			sessionRow.BlockingSession = *sample.BlockingSession
		}
		if sample.FinalBlockingInstance != nil {
			sessionRow.FinalBlockingInstance = *sample.FinalBlockingInstance
		}
		if sample.FinalBlockingSession != nil {
			sessionRow.FinalBlockingSession = *sample.FinalBlockingSession
		}
		if sample.WaitEvent.Valid {
			sessionRow.WaitEvent = sample.WaitEvent.String
		}
		if sample.WaitEventClass.Valid {
			sessionRow.WaitEventClass = sample.WaitEventClass.String
		}
		if sample.WaitTimeMicro != nil {
			sessionRow.WaitTimeMicro = *sample.WaitTimeMicro
		}
		sessionRow.OpFlags = sample.OpFlags

		statement := ""
		obfuscate := true
		var hasRealSQLText bool
		switch {
		case sample.Statement.Valid && sample.Statement.String != "" && !previousSQL:
			// If we captured the statement, we are assigning the value
			statement = sample.Statement.String
			hasRealSQLText = true
		case previousSQL && sample.PrevSQLFullText.Valid && sample.PrevSQLFullText.String != "":
			statement = sample.PrevSQLFullText.String
			hasRealSQLText = true
		case (sample.OpFlags & 8) == 8:
			statement = "LOG ON/LOG OFF"
			obfuscate = false
		case commandName != "":
			statement = commandName
		case sessionType == "BACKGROUND":
			statement = program
			obfuscate = false
		case sample.Module.Valid && sample.Module.String == "DBMS_SCHEDULER":
			statement = sample.Module.String
			obfuscate = false
		default:
			l.Debugf("activity sql text empty for %#v \n", sample)
		}

		if hasRealSQLText {
			/*
			 * If the statement length is maxSQLTextLength characters, we are assuming that the statement was truncated,
			 * so we are trying to fetch it complete. The full statement is stored in a LOB, so we are calling
			 * getFullSQLText which doesn't leak PGA memory
			 */
			if len(statement) == ipt.sqlSubstringLength && sessionRow.SQLID != "" {
				var fetchedStatement string
				err = ipt.getFullSQLText(&fetchedStatement, "sql_id", sessionRow.SQLID)
				if err != nil {
					l.Errorf("failed to fetch full sql text for the current sql_id: %s", err)
				}
				if fetchedStatement != "" {
					statement = fetchedStatement
				}
			}
		} else {
			if (sample.OpFlags & 128) == 128 {
				statement += " IN HARD PARSE"
			} else if (sample.OpFlags & 16) == 16 {
				statement += " IN PARSE"
			}
			if (sample.OpFlags & 65536) == 65536 {
				statement += " IN CURSOR CLOSING"
			}
		}

		sessionRow.Statement = statement
		if statement != "" && obfuscate {
			obfuscatedStatement, err := o.ObfuscateSQLString(statement)
			if err != nil {
				l.Warnf("failed to obfuscate statement: %s", err)
			} else {
				sessionRow.Statement = obfuscatedStatement.Query
			}
		}

		if sample.PdbName.Valid {
			sessionRow.PdbName = sample.PdbName.String
		}

		sessionRow.PdbName = ipt.getPdbName(sessionRow.PdbName)
		queryHash := sessionRow.OracleSQLRow.SQLID
		if sessionRow.OracleSQLRow.ForceMatchingSignature != 0 {
			queryHash = strconv.FormatUint(sessionRow.OracleSQLRow.ForceMatchingSignature, 10)
		}
		sessionRow.QuerySignature = generateQuerySignature(sessionRow.PdbName, queryHash)

		sessionRow.CdbName = ipt.cdbName
		sessionRows = append(sessionRows, sessionRow)
	}

	return sessionRows, nil
}

// maxInt returns the maximum of two int values.
// This is a compatibility function for Go versions < 1.21.
func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// mapOracleWaitClassToGroup maps Oracle wait_class to Datakit unified wait_group.
// Unified groups: Lock, I/O, Concurrency, Memory, Network, CPU, Commit/Log, Other.
func mapOracleWaitClassToGroup(waitClass string) string {
	switch strings.TrimSpace(strings.ToLower(waitClass)) {
	case "application":
		return "Lock"
	case "user i/o", "system i/o":
		return "I/O"
	case "concurrency":
		return "Concurrency"
	case "network":
		return "Network"
	case "cpu":
		return "CPU"
	case "commit":
		return "Commit/Log"
	case "memory":
		return "Memory"
	default:
		return "Other"
	}
}

// buildDbmActivityPoints builds point.Point from activity rows.
func (ipt *Input) buildDbmActivityPoints(rows []*OracleActivityRow, ptsTime time.Time) []*point.Point {
	var pts []*point.Point
	opts := append(point.DefaultLoggingOptions(), point.WithTime(ptsTime))

	for _, row := range rows {
		if row.SQLID == "" || row.Statement == "" {
			continue
		}

		kvs := ipt.getKVs()

		// Tags
		if ipt.cdbName != "" {
			kvs = kvs.AddTag("cdb_name", ipt.cdbName)
		}
		if row.PdbName != "" {
			kvs = kvs.AddTag("pdb_name", row.PdbName)
		}
		if row.User != "" {
			kvs = kvs.AddTag("username", row.User)
		}
		if row.Program != "" {
			kvs = kvs.AddTag("program", row.Program)
		}
		if row.Client != "" {
			kvs = kvs.AddTag("machine", row.Client)
		}
		if row.OsUser != "" {
			kvs = kvs.AddTag("terminal", row.OsUser)
		}
		if row.Module != "" {
			kvs = kvs.AddTag("module", row.Module)
		}
		if row.Action != "" {
			kvs = kvs.AddTag("action", row.Action)
		}
		if row.SQLID != "" {
			kvs = kvs.AddTag("sql_id", row.SQLID)
		}
		kvs = kvs.AddTag("force_matching_signature", fmt.Sprintf("%d", row.ForceMatchingSignature))
		if row.SQLPlanHashValue > 0 {
			kvs = kvs.AddTag("plan_hash_value", fmt.Sprintf("%d", row.SQLPlanHashValue))
		}
		if row.QuerySignature != "" {
			kvs = kvs.AddTag("query_signature", row.QuerySignature)
		}
		if row.Status != "" {
			kvs = kvs.AddTag("session_status", row.Status)
		}
		if row.Type != "" {
			kvs = kvs.AddTag("session_type", row.Type)
		}
		if row.WaitEventClass != "" {
			kvs = kvs.AddTag("wait_class", row.WaitEventClass)
			kvs = kvs.AddTag("wait_group", mapOracleWaitClassToGroup(row.WaitEventClass))
		}

		// Fields
		// message field contains the obfuscated/normalized SQL statement
		kvs = kvs.Set("message", row.Statement)
		kvs = kvs.Set("session_id", int64(row.SessionID))
		kvs = kvs.Set("serial_number", int64(row.SessionSerial))
		kvs = kvs.Set("wait_time", int64(row.WaitTimeMicro))

		// Blocking session
		kvs = kvs.Set("blocking_session_id", int64(row.BlockingSession))
		kvs = kvs.Set("final_blocking_session_id", int64(row.FinalBlockingSession))
		kvs = kvs.Set("blocking_instance", int64(row.BlockingInstance))
		kvs = kvs.Set("final_blocking_instance", int64(row.FinalBlockingInstance))
		kvs = kvs.Set("event", row.WaitEvent)

		// Client and connection information
		kvs = kvs.Set("client_port", row.Port)
		kvs = kvs.Set("process", row.Process)
		kvs = kvs.Set("client_info", row.ClientInfo)
		kvs = kvs.Set("client_identifier", row.ClientIdentifier)
		kvs = kvs.Set("logon_time", row.LogonTime)
		kvs = kvs.Set("sql_exec_start", row.SQLExecStart)
		kvs = kvs.Set("command_name", row.CommandName)

		pt := point.NewPoint(metricNameOracleDbmActivity, kvs, opts...)
		pts = append(pts, pt)
	}

	return pts
}

//nolint:lll
func (ipt *Input) getSQLRow(sqlID sql.NullString, forceMatchingSignature *string, sqlPlanHashValue *uint64, sqlExecStart sql.NullString) (OracleSQLRow, error) {
	SQLRow := OracleSQLRow{}
	if sqlID.Valid {
		SQLRow.SQLID = sqlID.String
	} else {
		SQLRow.SQLID = ""
		return SQLRow, nil
	}
	if forceMatchingSignature != nil {
		forceMatchingSignatureUint64, err := strconv.ParseUint(*forceMatchingSignature, 10, 64)
		if err != nil {
			return SQLRow, fmt.Errorf("failed converting force_matching_signature to uint64 %w", err)
		}
		SQLRow.ForceMatchingSignature = forceMatchingSignatureUint64
	} else {
		SQLRow.ForceMatchingSignature = 0
	}
	if sqlPlanHashValue != nil {
		SQLRow.SQLPlanHashValue = *sqlPlanHashValue
	}
	if sqlExecStart.Valid {
		SQLRow.SQLExecStart = sqlExecStart.String
	}
	return SQLRow, nil
}
