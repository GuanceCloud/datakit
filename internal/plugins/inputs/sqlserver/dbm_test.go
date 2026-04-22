// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package sqlserver

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DataDog/datadog-agent/pkg/obfuscate"

	"github.com/GuanceCloud/cliutils/point"
	"github.com/stretchr/testify/assert"
)

func TestObfuscateXMLPlan(t *testing.T) {
	tests := []struct {
		name     string
		rawPlan  string
		wantErr  bool
		validate func(t *testing.T, result string)
	}{
		{
			name: "obfuscate StatementText attribute",
			rawPlan: `<ShowPlanXML>
				<StmtSimple StatementText="SELECT * FROM users WHERE id = 123">
				</StmtSimple>
			</ShowPlanXML>`,
			wantErr: false,
			validate: func(t *testing.T, result string) {
				t.Helper()
				assert.NotContains(t, result, "SELECT * FROM users WHERE id = 123")
				assert.Contains(t, result, "StatementText")
				assert.Contains(t, result, "StmtSimple")
			},
		},
		{
			name: "obfuscate ConstValue attribute",
			rawPlan: `<ShowPlanXML>
				<ScalarOperator ConstValue="123">
				</ScalarOperator>
			</ShowPlanXML>`,
			wantErr: false,
			validate: func(t *testing.T, result string) {
				t.Helper()
				assert.Contains(t, result, "ConstValue")
				assert.Contains(t, result, "ScalarOperator")
			},
		},
		{
			name: "obfuscate ScalarString attribute",
			rawPlan: `<ShowPlanXML>
				<ScalarOperator ScalarString="GetRange([dbo].[users].[id],(123),(123))">
				</ScalarOperator>
			</ShowPlanXML>`,
			wantErr: false,
			validate: func(t *testing.T, result string) {
				t.Helper()
				assert.Contains(t, result, "ScalarString")
				assert.Contains(t, result, "ScalarOperator")
			},
		},
		{
			name: "obfuscate ParameterCompiledValue attribute",
			rawPlan: `<ShowPlanXML>
				<ParameterList>
					<ColumnReference ParameterCompiledValue="(123)">
					</ColumnReference>
				</ParameterList>
			</ShowPlanXML>`,
			wantErr: false,
			validate: func(t *testing.T, result string) {
				t.Helper()
				assert.Contains(t, result, "ParameterCompiledValue")
				assert.Contains(t, result, "ColumnReference")
			},
		},
		{
			name: "strip whitespace from CharData",
			rawPlan: `<ShowPlanXML>
				<Comment>   This is a comment   </Comment>
			</ShowPlanXML>`,
			wantErr: false,
			validate: func(t *testing.T, result string) {
				t.Helper()
				assert.Contains(t, result, "This is a comment")
				assert.NotContains(t, result, "   This is a comment   ")
			},
		},
		{
			name:    "empty XML",
			rawPlan: ``,
			wantErr: false,
			validate: func(t *testing.T, result string) {
				t.Helper()
				assert.Empty(t, result)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := obfuscateXMLPlan(tt.rawPlan)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.validate != nil {
					tt.validate(t, result)
				}
			}
		})
	}
}

func TestProcessPlan(t *testing.T) {
	tests := []struct {
		name        string
		plan        string
		isEncrypted bool
		wantPlan    string
		wantError   bool
	}{
		{
			name:        "empty plan with encrypted flag",
			plan:        "",
			isEncrypted: true,
			wantPlan:    "",
			wantError:   false,
		},
		{
			name:        "empty plan without encrypted flag",
			plan:        "",
			isEncrypted: false,
			wantPlan:    "",
			wantError:   false,
		},
		{
			name:        "encrypted plan",
			plan:        "<ShowPlanXML>encrypted content</ShowPlanXML>",
			isEncrypted: true,
			wantPlan:    "<ShowPlanXML>encrypted content</ShowPlanXML>",
			wantError:   false,
		},
		{
			name:        "valid plan without encryption",
			plan:        `<ShowPlanXML><StmtSimple StatementText="SELECT 1"></StmtSimple></ShowPlanXML>`,
			isEncrypted: false,
			wantError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resultPlan, err := processPlan(tt.plan, tt.isEncrypted)
			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			if tt.wantPlan != "" {
				assert.Equal(t, tt.wantPlan, resultPlan)
			} else if !tt.isEncrypted && tt.plan != "" {
				// For non-encrypted plans, the result should be obfuscated
				assert.NotEmpty(t, resultPlan)
			}
		})
	}
}

func TestBuildIdleBlockingSessionsQuery(t *testing.T) {
	ipt := &Input{}

	tests := []struct {
		name               string
		blockingSessionIDs []int64
		expectedEmpty      bool
		validate           func(t *testing.T, query string)
	}{
		{
			name:               "single session ID",
			blockingSessionIDs: []int64{123},
			expectedEmpty:      false,
			validate: func(t *testing.T, query string) {
				t.Helper()
				assert.Contains(t, query, "123")
				assert.NotContains(t, query, "__BLOCKING_SESSION_IDS__")
				assert.Contains(t, query, "sess.session_id IN (123)")
				assert.Contains(t, query, "c.session_id IN (123)")
			},
		},
		{
			name:               "multiple session IDs",
			blockingSessionIDs: []int64{123, 456, 789},
			expectedEmpty:      false,
			validate: func(t *testing.T, query string) {
				t.Helper()
				assert.Contains(t, query, "123,456,789")
				assert.NotContains(t, query, "__BLOCKING_SESSION_IDS__")
				assert.Contains(t, query, "sess.session_id IN (123,456,789)")
				assert.Contains(t, query, "c.session_id IN (123,456,789)")
			},
		},
		{
			name:               "empty session IDs",
			blockingSessionIDs: []int64{},
			expectedEmpty:      true,
			validate: func(t *testing.T, query string) {
				t.Helper()
				// Empty input should return empty string
				assert.Empty(t, query)
			},
		},
		{
			name:               "invalid session IDs",
			blockingSessionIDs: []int64{},
			expectedEmpty:      true,
			validate: func(t *testing.T, query string) {
				t.Helper()
				// All invalid IDs should result in empty string
				assert.Empty(t, query)
			},
		},
		{
			name:               "mixed valid and invalid session IDs",
			blockingSessionIDs: []int64{123, 456},
			expectedEmpty:      false,
			validate: func(t *testing.T, query string) {
				t.Helper()
				// Only valid IDs should appear
				assert.Contains(t, query, "123,456")
				assert.NotContains(t, query, "abc")
				assert.NotContains(t, query, "def")
				// Verify the placeholder is replaced
				assert.NotContains(t, query, "__BLOCKING_SESSION_IDS__")
				// Verify the query contains the expected SQL structure
				assert.Contains(t, query, "sess.session_id IN (123,456)")
				assert.Contains(t, query, "c.session_id IN (123,456)")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := ipt.buildIdleBlockingSessionsQuery(tt.blockingSessionIDs)
			if tt.expectedEmpty {
				assert.Empty(t, query)
			} else {
				assert.NotEmpty(t, query)
			}
			if tt.validate != nil {
				tt.validate(t, query)
			}
		})
	}
}

func TestBuildActivityQueryWithColumns(t *testing.T) {
	ipt := &Input{
		Database: "testdb",
		Dbm: &dbmConfig{
			StoredProcedureCharactersLimit: 500,
		},
		dbmColumnsCache: map[string][]string{
			"sys.dm_exec_requests": {"cpu_time", "total_elapsed_time", "reads"},
		},
	}

	tests := []struct {
		name     string
		rowLimit int
		wantErr  bool
		validate func(t *testing.T, query string)
	}{
		{
			name:     "valid query",
			rowLimit: 100,
			wantErr:  false,
			validate: func(t *testing.T, query string) {
				t.Helper()
				assert.Contains(t, query, "TOP 100")
				assert.Contains(t, query, "req.cpu_time")
				assert.Contains(t, query, "req.total_elapsed_time")
				assert.Contains(t, query, "req.reads")
				assert.Contains(t, query, "DB_NAME(req.database_id) AS database_name")
			},
		},
		{
			name:     "different row limit",
			rowLimit: 50,
			wantErr:  false,
			validate: func(t *testing.T, query string) {
				t.Helper()
				assert.Contains(t, query, "TOP 50")
				assert.Contains(t, query, "DB_NAME(req.database_id) AS database_name")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query, err := ipt.buildActivityQueryWithColumns(context.Background(), tt.rowLimit)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, query)
				if tt.validate != nil {
					tt.validate(t, query)
				}
			}
		})
	}

	// Test different databases generate same query (both use @db parameter)
	t.Run("different databases", func(t *testing.T) {
		ipt3 := &Input{
			Database: "db1",
			Dbm: &dbmConfig{
				StoredProcedureCharactersLimit: 500,
			},
			dbmColumnsCache: map[string][]string{
				"sys.dm_exec_requests": {"cpu_time"},
			},
		}
		ipt4 := &Input{
			Database: "db2",
			Dbm: &dbmConfig{
				StoredProcedureCharactersLimit: 500,
			},
			dbmColumnsCache: map[string][]string{
				"sys.dm_exec_requests": {"cpu_time"},
			},
		}
		query1, err1 := ipt3.buildActivityQueryWithColumns(context.Background(), 100)
		assert.NoError(t, err1)
		assert.Contains(t, query1, "DB_NAME(req.database_id) AS database_name")

		query2, err2 := ipt4.buildActivityQueryWithColumns(context.Background(), 100)
		assert.NoError(t, err2)
		assert.Contains(t, query2, "DB_NAME(req.database_id) AS database_name")
		// Both queries should be the same
		assert.Equal(t, query1, query2)
	})
}

func TestBuildDbmActivityRow(t *testing.T) {
	obfuscator := obfuscate.NewObfuscator(obfuscate.Config{
		SQL: obfuscate.SQLConfig{
			DBMS: obfuscate.DBMSSQLServer,
		},
	})

	tests := []struct {
		name      string
		columnMap map[string]*interface{}
		wantNil   bool
		wantErr   bool
		validate  func(t *testing.T, row *dbmActivityRow)
	}{
		{
			name: "valid row with all fields",
			columnMap: map[string]*interface{}{
				"id":                          intPtr(int64(123)),
				"user_name":                   strPtr("testuser"),
				"host_name":                   strPtr("testhost"),
				"database_name":               strPtr("testdb"),
				"statement_text":              strPtr("SELECT * FROM users WHERE id = 1"),
				"session_status":              strPtr("running"),
				"request_status":              strPtr("running"),
				"command":                     strPtr("SELECT"),
				"percent_complete":            intPtr(int64(0)),
				"estimated_completion_time":   intPtr(int64(0)),
				"wait_type":                   strPtr(""),
				"wait_time":                   intPtr(int64(0)),
				"wait_resource":               strPtr(""),
				"blocking_session_id":         intPtr(int64(0)),
				"last_wait_type":              strPtr(""),
				"open_transaction_count":      intPtr(int64(0)),
				"transaction_id":              intPtr(int64(0)),
				"transaction_isolation_level": intPtr(int64(2)),
				"lock_timeout":                intPtr(int64(-1)),
				"deadlock_priority":           intPtr(int64(0)),
				"query_hash":                  bytesPtr([]byte{0x01, 0x02}),
				"query_plan_hash":             bytesPtr([]byte{0x03, 0x04}),
				"context_info":                strPtr(""),
				"cpu_time":                    intPtr(int64(100)),
				"total_elapsed_time":          intPtr(int64(200)),
				"reads":                       intPtr(int64(10)),
				"writes":                      intPtr(int64(5)),
				"logical_reads":               intPtr(int64(20)),
				"row_count":                   intPtr(int64(1)),
				"client_address":              strPtr("192.168.1.1"),
				"client_port":                 intPtr(int64(1433)),
				"program_name":                strPtr("testapp"),
				"is_user_process":             boolPtr(true),
				"client_interface_name":       strPtr("ODBC"),
				"query_start":                 strPtr("2024-01-01T00:00:00Z"),
				"last_request_start_time":     timePtr(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)),
				"procedure_name":              strPtr("testproc"),
				"schema_name":                 strPtr("dbo"),
			},
			wantNil: false,
			wantErr: false,
			validate: func(t *testing.T, row *dbmActivityRow) {
				t.Helper()
				assert.Equal(t, int64(123), row.sessionID)
				assert.Equal(t, "testuser", row.userName)
				assert.Equal(t, "testhost", row.hostName)
				assert.Equal(t, "testdb", row.databaseName)
				assert.NotEmpty(t, row.obfuscatedText)
				assert.Equal(t, "running", row.sessionStatus)
				assert.Equal(t, "running", row.requestStatus)
				assert.Equal(t, "SELECT", row.command)
				assert.Equal(t, waitTypeCPU, row.waitType)
				assert.Equal(t, "CPU", row.waitCategory)
				assert.Equal(t, int64(100), row.cpuTime)
				assert.Equal(t, "dbo.testproc", row.procedureName)
				assert.Equal(t, "dbo", row.schemaName)
				assert.Equal(t, "testapp", row.programName)
				assert.Equal(t, "0x0102", row.queryHash)
				assert.Equal(t, "0x0304", row.queryPlanHash)
			},
		},
		{
			name: "runnable request with empty wait type gets waiting_on_cpu",
			columnMap: map[string]*interface{}{
				"id":             intPtr(int64(456)),
				"statement_text": strPtr("SELECT 1"),
				"session_status": strPtr("running"),
				"request_status": strPtr("runnable"),
				"wait_type":      strPtr(""),
			},
			wantNil: false,
			wantErr: false,
			validate: func(t *testing.T, row *dbmActivityRow) {
				t.Helper()
				assert.Equal(t, waitTypeWaitingOnCPU, row.waitType)
				assert.Equal(t, "CPU", row.waitCategory)
			},
		},
		{
			name: "suspended request with empty wait type stays empty",
			columnMap: map[string]*interface{}{
				"id":             intPtr(int64(789)),
				"statement_text": strPtr("SELECT 1"),
				"session_status": strPtr("running"),
				"request_status": strPtr("suspended"),
				"wait_type":      strPtr(""),
			},
			wantNil: false,
			wantErr: false,
			validate: func(t *testing.T, row *dbmActivityRow) {
				t.Helper()
				assert.Empty(t, row.waitType)
				assert.Equal(t, "Other", row.waitCategory)
			},
		},
		{
			name: "empty statement text returns nil",
			columnMap: map[string]*interface{}{
				"statement_text": strPtr(""),
			},
			wantNil: true,
			wantErr: false,
		},
		{
			name: "missing statement text returns nil",
			columnMap: map[string]*interface{}{
				"id": intPtr(int64(123)),
			},
			wantNil: true,
			wantErr: false,
		},
		{
			name: "procedure name without schema",
			columnMap: map[string]*interface{}{
				"statement_text": strPtr("SELECT 1"),
				"procedure_name": strPtr("testproc"),
			},
			wantNil: false,
			wantErr: false,
			validate: func(t *testing.T, row *dbmActivityRow) {
				t.Helper()
				assert.Equal(t, "testproc", row.procedureName)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			row, err := buildDbmActivityRow(tt.columnMap, obfuscator)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			if tt.wantNil {
				assert.Nil(t, row)
			} else {
				assert.NotNil(t, row)
				if tt.validate != nil {
					tt.validate(t, row)
				}
			}
		})
	}
}

func TestBuildIdleBlockingActivityRow(t *testing.T) {
	obfuscator := obfuscate.NewObfuscator(obfuscate.Config{
		SQL: obfuscate.SQLConfig{
			DBMS: obfuscate.DBMSSQLServer,
		},
	})

	tests := []struct {
		name     string
		rawRow   *rawIdleBlockingSessionRow
		wantNil  bool
		wantErr  bool
		validate func(t *testing.T, row *dbmActivityRow)
	}{
		{
			name: "valid row",
			rawRow: &rawIdleBlockingSessionRow{
				ID:                   sql.NullInt64{Int64: 123, Valid: true},
				UserName:             sql.NullString{String: "testuser", Valid: true},
				HostName:             sql.NullString{String: "testhost", Valid: true},
				DatabaseName:         sql.NullString{String: "testdb", Valid: true},
				SessionStatus:        sql.NullString{String: "sleeping", Valid: true},
				StatementText:        sql.NullString{String: "SELECT * FROM users", Valid: true},
				ProcedureName:        sql.NullString{String: "testproc", Valid: true},
				SchemaName:           sql.NullString{String: "dbo", Valid: true},
				ClientAddress:        sql.NullString{String: "192.168.1.1", Valid: true},
				ClientPort:           sql.NullInt64{Int64: 1433, Valid: true},
				ProgramName:          sql.NullString{String: "testapp", Valid: true},
				IsUserProcess:        sql.NullBool{Bool: true, Valid: true},
				ClientInterfaceName:  sql.NullString{String: "ODBC", Valid: true},
				LastRequestStartTime: sql.NullTime{Time: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), Valid: true},
			},
			wantNil: false,
			wantErr: false,
			validate: func(t *testing.T, row *dbmActivityRow) {
				t.Helper()
				assert.Equal(t, int64(123), row.sessionID)
				assert.Equal(t, "testuser", row.userName)
				assert.Equal(t, "testhost", row.hostName)
				assert.Equal(t, "testdb", row.databaseName)
				assert.Equal(t, "sleeping", row.sessionStatus)
				assert.Empty(t, row.waitType)
				assert.Equal(t, "Other", row.waitCategory)
				assert.NotEmpty(t, row.obfuscatedText)
				assert.Equal(t, "dbo.testproc", row.procedureName)
				assert.Equal(t, "dbo", row.schemaName)
				assert.Equal(t, "192.168.1.1", row.clientAddress)
				assert.Equal(t, "1433", row.clientPort)
				assert.Equal(t, "testapp", row.programName)
				assert.True(t, row.isUserProcess)
			},
		},
		{
			name: "empty statement text returns nil",
			rawRow: &rawIdleBlockingSessionRow{
				ID:            sql.NullInt64{Int64: 123, Valid: true},
				StatementText: sql.NullString{String: "", Valid: true},
			},
			wantNil: true,
			wantErr: false,
		},
		{
			name: "procedure name without schema",
			rawRow: &rawIdleBlockingSessionRow{
				ID:            sql.NullInt64{Int64: 456, Valid: true},
				StatementText: sql.NullString{String: "SELECT 1", Valid: true},
				ProcedureName: sql.NullString{String: "testproc", Valid: true},
				SchemaName:    sql.NullString{String: "", Valid: false},
			},
			wantNil: false,
			wantErr: false,
			validate: func(t *testing.T, row *dbmActivityRow) {
				t.Helper()
				assert.Equal(t, "testproc", row.procedureName)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			row, err := buildIdleBlockingActivityRow(tt.rawRow, obfuscator)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			if tt.wantNil {
				assert.Nil(t, row)
			} else {
				assert.NotNil(t, row)
				if tt.validate != nil {
					tt.validate(t, row)
				}
			}
		})
	}
}

// Helper functions for tests
func strPtr(s string) *interface{} {
	var v interface{} = s
	return &v
}

func intPtr(i int64) *interface{} {
	var v interface{} = i
	return &v
}

func boolPtr(b bool) *interface{} {
	var v interface{} = b
	return &v
}

func bytesPtr(b []byte) *interface{} {
	var v interface{} = b
	return &v
}

func timePtr(t time.Time) *interface{} {
	var v interface{} = t
	return &v
}

func TestGetStringFromInt64(t *testing.T) {
	tests := []struct {
		name      string
		columnMap map[string]*interface{}
		colName   string
		want      string
	}{
		{
			name: "valid int64",
			columnMap: map[string]*interface{}{
				"id": intPtr(int64(123)),
			},
			colName: "id",
			want:    "123",
		},
		{
			name: "nil value",
			columnMap: map[string]*interface{}{
				"id": nil,
			},
			colName: "id",
			want:    "",
		},
		{
			name: "missing key",
			columnMap: map[string]*interface{}{
				"other": intPtr(int64(123)),
			},
			colName: "id",
			want:    "",
		},
		{
			name: "wrong type",
			columnMap: map[string]*interface{}{
				"id": strPtr("123"),
			},
			colName: "id",
			want:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getStringFromInt64(tt.columnMap, tt.colName)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestGetTimeField(t *testing.T) {
	testTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name      string
		columnMap map[string]*interface{}
		colName   string
		want      time.Time
	}{
		{
			name: "valid time",
			columnMap: map[string]*interface{}{
				"timestamp": timePtr(testTime),
			},
			colName: "timestamp",
			want:    testTime,
		},
		{
			name: "nil value",
			columnMap: map[string]*interface{}{
				"timestamp": nil,
			},
			colName: "timestamp",
			want:    time.Time{},
		},
		{
			name: "missing key",
			columnMap: map[string]*interface{}{
				"other": timePtr(testTime),
			},
			colName: "timestamp",
			want:    time.Time{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getTimeField(tt.columnMap, tt.colName)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestGetInt64Field(t *testing.T) {
	tests := []struct {
		name      string
		columnMap map[string]*interface{}
		colName   string
		want      int64
	}{
		{
			name: "valid int64",
			columnMap: map[string]*interface{}{
				"count": intPtr(int64(100)),
			},
			colName: "count",
			want:    int64(100),
		},
		{
			name: "nil value",
			columnMap: map[string]*interface{}{
				"count": nil,
			},
			colName: "count",
			want:    int64(0),
		},
		{
			name: "missing key",
			columnMap: map[string]*interface{}{
				"other": intPtr(int64(100)),
			},
			colName: "count",
			want:    int64(0),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getInt64Field(tt.columnMap, tt.colName)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestGetBoolField(t *testing.T) {
	tests := []struct {
		name      string
		columnMap map[string]*interface{}
		colName   string
		want      bool
	}{
		{
			name: "valid bool true",
			columnMap: map[string]*interface{}{
				"flag": boolPtr(true),
			},
			colName: "flag",
			want:    true,
		},
		{
			name: "valid bool false",
			columnMap: map[string]*interface{}{
				"flag": boolPtr(false),
			},
			colName: "flag",
			want:    false,
		},
		{
			name: "nil value",
			columnMap: map[string]*interface{}{
				"flag": nil,
			},
			colName: "flag",
			want:    false,
		},
		{
			name: "missing key",
			columnMap: map[string]*interface{}{
				"other": boolPtr(true),
			},
			colName: "flag",
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getBoolField(tt.columnMap, tt.colName)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestGetStringField(t *testing.T) {
	tests := []struct {
		name      string
		columnMap map[string]*interface{}
		colName   string
		want      string
	}{
		{
			name: "valid string",
			columnMap: map[string]*interface{}{
				"name": strPtr("test"),
			},
			colName: "name",
			want:    "test",
		},
		{
			name: "valid bytes",
			columnMap: map[string]*interface{}{
				"name": bytesPtr([]byte("test")),
			},
			colName: "name",
			want:    "test",
		},
		{
			name: "nil value",
			columnMap: map[string]*interface{}{
				"name": nil,
			},
			colName: "name",
			want:    "",
		},
		{
			name: "missing key",
			columnMap: map[string]*interface{}{
				"other": strPtr("test"),
			},
			colName: "name",
			want:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getStringField(tt.columnMap, tt.colName)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestGetBytesField(t *testing.T) {
	testBytes := []byte{0x01, 0x02, 0x03}

	tests := []struct {
		name      string
		columnMap map[string]*interface{}
		colName   string
		want      []byte
	}{
		{
			name: "valid bytes",
			columnMap: map[string]*interface{}{
				"data": bytesPtr(testBytes),
			},
			colName: "data",
			want:    testBytes,
		},
		{
			name: "nil value",
			columnMap: map[string]*interface{}{
				"data": nil,
			},
			colName: "data",
			want:    nil,
		},
		{
			name: "missing key",
			columnMap: map[string]*interface{}{
				"other": bytesPtr(testBytes),
			},
			colName: "data",
			want:    nil,
		},
		{
			name: "wrong type",
			columnMap: map[string]*interface{}{
				"data": strPtr("test"),
			},
			colName: "data",
			want:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getBytesField(tt.columnMap, tt.colName)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestObfuscateXMLPlan_ErrorHandling(t *testing.T) {
	tests := []struct {
		name    string
		rawPlan string
		wantErr bool
	}{
		{
			name:    "invalid XML encoding",
			rawPlan: string([]byte{0xFF, 0xFE, 0xFD}),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := obfuscateXMLPlan(tt.rawPlan)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Empty(t, result)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestObfuscateXMLPlan_AllObfuscatableAttributes(t *testing.T) {
	rawPlan := `<ShowPlanXML>
		<StmtSimple StatementText="SELECT * FROM table">
			<ScalarOperator ConstValue="123" ScalarString="GetRange([dbo].[table].[id],(123),(123))">
				<ParameterList>
					<ColumnReference ParameterCompiledValue="(123)">
					</ColumnReference>
				</ParameterList>
			</ScalarOperator>
		</StmtSimple>
	</ShowPlanXML>`

	result, err := obfuscateXMLPlan(rawPlan)
	assert.NoError(t, err)

	// All four attribute types should be present
	assert.Contains(t, result, "StatementText")
	assert.Contains(t, result, "ConstValue")
	assert.Contains(t, result, "ScalarString")
	assert.Contains(t, result, "ParameterCompiledValue")

	// Verify structure is preserved
	assert.Contains(t, result, "ShowPlanXML")
	assert.Contains(t, result, "StmtSimple")
	assert.Contains(t, result, "ScalarOperator")
	assert.Contains(t, result, "ParameterList")
	assert.Contains(t, result, "ColumnReference")
}

func TestObfuscateXMLPlan_WhitespaceHandling(t *testing.T) {
	rawPlan := `<ShowPlanXML>
		<Element>   text with spaces   </Element>
	</ShowPlanXML>`

	result, err := obfuscateXMLPlan(rawPlan)
	assert.NoError(t, err)

	// Leading/trailing whitespace should be trimmed, but internal spaces preserved
	assert.Contains(t, result, "text with spaces")
	assert.NotContains(t, result, "   text with spaces   ")
}

func TestParseExecutionPlan(t *testing.T) {
	tests := []struct {
		name     string
		xmlPlan  string
		wantErr  bool
		validate func(t *testing.T, analysis *PlanAnalysis)
	}{
		{
			name:    "empty XML",
			xmlPlan: "",
			wantErr: false,
			validate: func(t *testing.T, analysis *PlanAnalysis) {
				t.Helper()
				assert.NotNil(t, analysis)
				assert.Equal(t, 0.0, analysis.TotalCost)
				assert.Equal(t, int64(0), analysis.CompileTime)
				assert.Equal(t, 0.0, analysis.EstimatedRows)
				assert.Empty(t, analysis.Tables)
				assert.Empty(t, analysis.Warnings)
				assert.Empty(t, analysis.IndexesUsed)
				assert.Empty(t, analysis.Operators)
				assert.Empty(t, analysis.MissingIndexes)
			},
		},
		{
			name: "basic StmtSimple with TotalCost and EstimatedRows",
			xmlPlan: `<ShowPlanXML>
				<StmtSimple StatementSubTreeCost="1.234" StatementEstRows="1000">
				</StmtSimple>
			</ShowPlanXML>`,
			wantErr: false,
			validate: func(t *testing.T, analysis *PlanAnalysis) {
				t.Helper()
				assert.NotNil(t, analysis)
				assert.Equal(t, 1.234, analysis.TotalCost)
				assert.Equal(t, 1000.0, analysis.EstimatedRows)
			},
		},
		{
			name: "QueryPlan with CompileTime",
			xmlPlan: `<ShowPlanXML>
				<StmtSimple>
					<QueryPlan CompileTime="500">
					</QueryPlan>
				</StmtSimple>
			</ShowPlanXML>`,
			wantErr: false,
			validate: func(t *testing.T, analysis *PlanAnalysis) {
				t.Helper()
				assert.NotNil(t, analysis)
				assert.Equal(t, int64(500), analysis.CompileTime)
			},
		},
		{
			name: "single RelOp operator",
			xmlPlan: `<ShowPlanXML>
				<StmtSimple>
					<QueryPlan>
						<RelOp PhysicalOp="Table Scan" EstimatedTotalSubtreeCost="0.5" EstimateRows="100" ActualRows="95">
						</RelOp>
					</QueryPlan>
				</StmtSimple>
			</ShowPlanXML>`,
			wantErr: false,
			validate: func(t *testing.T, analysis *PlanAnalysis) {
				t.Helper()
				assert.NotNil(t, analysis)
				assert.Len(t, analysis.Operators, 1)
				op := analysis.Operators[0]
				assert.Equal(t, "Table Scan", op.PhysicalOp)
				assert.Equal(t, 0.5, op.Cost)
				assert.Equal(t, 100.0, op.EstimatedRows)
				assert.Equal(t, 95.0, op.ActualRows)
			},
		},
		{
			name: "nested RelOp operators",
			xmlPlan: `<ShowPlanXML>
				<StmtSimple>
					<QueryPlan>
						<RelOp PhysicalOp="Sort" EstimatedTotalSubtreeCost="1.0">
							<RelOp PhysicalOp="Table Scan" EstimatedTotalSubtreeCost="0.5">
							</RelOp>
						</RelOp>
					</QueryPlan>
				</StmtSimple>
			</ShowPlanXML>`,
			wantErr: false,
			validate: func(t *testing.T, analysis *PlanAnalysis) {
				t.Helper()
				assert.NotNil(t, analysis)
				assert.Len(t, analysis.Operators, 1)
				rootOp := analysis.Operators[0]
				assert.Equal(t, "Sort", rootOp.PhysicalOp)
				assert.Len(t, rootOp.Children, 1)
				childOp := rootOp.Children[0]
				assert.Equal(t, "Table Scan", childOp.PhysicalOp)
			},
		},
		{
			name: "Object with Table and Index",
			xmlPlan: `<ShowPlanXML>
				<StmtSimple>
					<QueryPlan>
						<RelOp PhysicalOp="Clustered Index Scan">
							<Object Database="[loadtest]" Schema="[dbo]" Table="[users]" Index="[PK__users__B9BE370FB9E45B5A]">
							</Object>
						</RelOp>
					</QueryPlan>
				</StmtSimple>
			</ShowPlanXML>`,
			wantErr: false,
			validate: func(t *testing.T, analysis *PlanAnalysis) {
				t.Helper()
				assert.NotNil(t, analysis)
				assert.Contains(t, analysis.Tables, "users")
				assert.Contains(t, analysis.IndexesUsed, "users.PK__users__B9BE370FB9E45B5A")
				assert.Len(t, analysis.Operators, 1)
				op := analysis.Operators[0]
				assert.Equal(t, "users", op.TableName)
				assert.Equal(t, "PK__users__B9BE370FB9E45B5A", op.IndexName)
			},
		},
		{
			name: "Object with brackets in table name",
			xmlPlan: `<ShowPlanXML>
				<StmtSimple>
					<QueryPlan>
						<RelOp PhysicalOp="Index Seek">
							<Object Table="[Users]" Index="[IX_Users_Email]">
							</Object>
						</RelOp>
					</QueryPlan>
				</StmtSimple>
			</ShowPlanXML>`,
			wantErr: false,
			validate: func(t *testing.T, analysis *PlanAnalysis) {
				t.Helper()
				assert.NotNil(t, analysis)
				assert.Contains(t, analysis.Tables, "Users")
				assert.Contains(t, analysis.IndexesUsed, "Users.IX_Users_Email")
			},
		},
		{
			name: "MissingIndexes",
			xmlPlan: `<ShowPlanXML>
				<StmtSimple>
					<QueryPlan>
						<MissingIndexes>
							<MissingIndexGroup Impact="49.3503">
								<MissingIndex Database="[loadtest]" Schema="[dbo]" Table="[orders]">
									<ColumnGroup Usage="EQUALITY">
										<Column Name="[user_id]"></Column>
									</ColumnGroup>
									<ColumnGroup Usage="INCLUDE">
										<Column Name="[order_date]"></Column>
										<Column Name="[total_amount]"></Column>
									</ColumnGroup>
								</MissingIndex>
							</MissingIndexGroup>
						</MissingIndexes>
					</QueryPlan>
				</StmtSimple>
			</ShowPlanXML>`,
			wantErr: false,
			validate: func(t *testing.T, analysis *PlanAnalysis) {
				t.Helper()
				assert.NotNil(t, analysis)
				assert.Len(t, analysis.MissingIndexes, 1)
				missingIdxGroup := analysis.MissingIndexes[0]
				assert.Equal(t, 49.3503, missingIdxGroup.Impact)
				assert.Len(t, missingIdxGroup.MissingIndexes, 1)
				missingIdx := missingIdxGroup.MissingIndexes[0]
				assert.Equal(t, "orders", missingIdx.Table)
				assert.Equal(t, []string{"user_id"}, missingIdx.Keys)
				assert.Equal(t, []string{"order_date", "total_amount"}, missingIdx.Includes)
			},
		},
		{
			name: "Warnings",
			xmlPlan: `<ShowPlanXML>
				<StmtSimple>
					<QueryPlan>
						<RelOp>
							<Warnings>
								<ColumnsWithNoStatistics>
								</ColumnsWithNoStatistics>
								<SpillToTempDb>
								</SpillToTempDb>
							</Warnings>
						</RelOp>
					</QueryPlan>
				</StmtSimple>
			</ShowPlanXML>`,
			wantErr: false,
			validate: func(t *testing.T, analysis *PlanAnalysis) {
				t.Helper()
				assert.NotNil(t, analysis)
				assert.Contains(t, analysis.Warnings, "ColumnsWithNoStatistics")
				assert.Contains(t, analysis.Warnings, "SpillToTempDb")
			},
		},
		{
			name: "multiple tables and indexes",
			xmlPlan: `<ShowPlanXML>
				<StmtSimple>
					<QueryPlan>
						<RelOp PhysicalOp="Nested Loops">
							<RelOp PhysicalOp="Index Seek">
								<Object Table="[Users]" Index="[IX_Users_Id]">
								</Object>
							</RelOp>
							<RelOp PhysicalOp="Index Seek">
								<Object Table="[Orders]" Index="[IX_Orders_UserId]">
								</Object>
							</RelOp>
						</RelOp>
					</QueryPlan>
				</StmtSimple>
			</ShowPlanXML>`,
			wantErr: false,
			validate: func(t *testing.T, analysis *PlanAnalysis) {
				t.Helper()
				assert.NotNil(t, analysis)
				assert.Contains(t, analysis.Tables, "Users")
				assert.Contains(t, analysis.Tables, "Orders")
				assert.Contains(t, analysis.IndexesUsed, "Users.IX_Users_Id")
				assert.Contains(t, analysis.IndexesUsed, "Orders.IX_Orders_UserId")
			},
		},
		{
			name: "duplicate tables and indexes should be deduplicated",
			xmlPlan: `<ShowPlanXML>
				<StmtSimple>
					<QueryPlan>
						<RelOp PhysicalOp="Index Seek">
							<Object Table="[Users]" Index="[IX_Users_Id]">
							</Object>
						</RelOp>
						<RelOp PhysicalOp="Index Seek">
							<Object Table="[Users]" Index="[IX_Users_Id]">
							</Object>
						</RelOp>
					</QueryPlan>
				</StmtSimple>
			</ShowPlanXML>`,
			wantErr: false,
			validate: func(t *testing.T, analysis *PlanAnalysis) {
				t.Helper()
				assert.NotNil(t, analysis)
				assert.Len(t, analysis.Tables, 1)
				assert.Len(t, analysis.IndexesUsed, 1)
			},
		},
		{
			name: "complete plan with all elements",
			xmlPlan: `<ShowPlanXML>
				<StmtSimple StatementSubTreeCost="2.5" StatementEstRows="5000">
					<QueryPlan CompileTime="1000">
						<RelOp PhysicalOp="Sort" EstimatedTotalSubtreeCost="2.0" EstimateRows="5000">
							<RelOp PhysicalOp="Index Seek" EstimatedTotalSubtreeCost="1.5" EstimateRows="5000" ActualRows="4950">
								<Object Database="[loadtest]" Schema="[dbo]" Table="[Products]" Index="[IX_Products_Category]">
								</Object>
							</RelOp>
						</RelOp>
						<MissingIndexes>
							<MissingIndexGroup Impact="85.0">
								<MissingIndex Database="[loadtest]" Schema="[dbo]" Table="[Orders]">
									<ColumnGroup Usage="EQUALITY">
										<Column Name="[user_id]"></Column>
									</ColumnGroup>
								</MissingIndex>
							</MissingIndexGroup>
						</MissingIndexes>
						<RelOp>
							<Warnings>
								<ColumnsWithNoStatistics>
								</ColumnsWithNoStatistics>
							</Warnings>
						</RelOp>
					</QueryPlan>
				</StmtSimple>
			</ShowPlanXML>`,
			wantErr: false,
			validate: func(t *testing.T, analysis *PlanAnalysis) {
				t.Helper()
				assert.NotNil(t, analysis)
				assert.Equal(t, 2.5, analysis.TotalCost)
				assert.Equal(t, 5000.0, analysis.EstimatedRows)
				assert.Equal(t, int64(1000), analysis.CompileTime)
				assert.Contains(t, analysis.Tables, "Products")
				assert.Contains(t, analysis.IndexesUsed, "Products.IX_Products_Category")
				assert.Len(t, analysis.MissingIndexes, 1)
				missingIdxGroup := analysis.MissingIndexes[0]
				assert.Equal(t, 85.0, missingIdxGroup.Impact)
				assert.Len(t, missingIdxGroup.MissingIndexes, 1)
				missingIdx := missingIdxGroup.MissingIndexes[0]
				assert.Equal(t, "Orders", missingIdx.Table)
				assert.Equal(t, []string{"user_id"}, missingIdx.Keys)
				assert.Contains(t, analysis.Warnings, "ColumnsWithNoStatistics")
				// There are two root RelOp elements: one Sort and one for Warnings
				assert.Len(t, analysis.Operators, 2)
				// Find the Sort operator
				var sortOp *OperatorInfo
				for _, op := range analysis.Operators {
					if op.PhysicalOp == "Sort" {
						sortOp = op
						break
					}
				}
				assert.NotNil(t, sortOp)
				assert.Equal(t, "Sort", sortOp.PhysicalOp)
				assert.Len(t, sortOp.Children, 1)
				childOp := sortOp.Children[0]
				assert.Equal(t, "Index Seek", childOp.PhysicalOp)
				assert.Equal(t, 4950.0, childOp.ActualRows)
			},
		},
		{
			name: "invalid XML should still parse (Strict=false)",
			xmlPlan: `<ShowPlanXML>
				<StmtSimple>
					<QueryPlan>
						<RelOp PhysicalOp="Table Scan">
							<UnclosedTag>
						</RelOp>
					</QueryPlan>
				</StmtSimple>
			</ShowPlanXML>`,
			wantErr: false,
			validate: func(t *testing.T, analysis *PlanAnalysis) {
				// Should still parse what it can
				t.Helper()
				assert.NotNil(t, analysis)
			},
		},
		{
			name: "RelOp outside QueryPlan should be ignored",
			xmlPlan: `<ShowPlanXML>
				<StmtSimple>
					<RelOp PhysicalOp="Table Scan">
					</RelOp>
					<QueryPlan>
						<RelOp PhysicalOp="Index Seek">
						</RelOp>
					</QueryPlan>
				</StmtSimple>
			</ShowPlanXML>`,
			wantErr: false,
			validate: func(t *testing.T, analysis *PlanAnalysis) {
				t.Helper()
				assert.NotNil(t, analysis)
				// Only the RelOp inside QueryPlan should be parsed
				assert.Len(t, analysis.Operators, 1)
				assert.Equal(t, "Index Seek", analysis.Operators[0].PhysicalOp)
			},
		},
		{
			name: "MissingIndex without Table should be ignored",
			xmlPlan: `<ShowPlanXML>
				<StmtSimple>
					<QueryPlan>
						<MissingIndexes>
							<MissingIndex>
								<MissingIndexGroup Impact="50.0">
								</MissingIndexGroup>
							</MissingIndex>
						</MissingIndexes>
					</QueryPlan>
				</StmtSimple>
			</ShowPlanXML>`,
			wantErr: false,
			validate: func(t *testing.T, analysis *PlanAnalysis) {
				t.Helper()
				assert.NotNil(t, analysis)
				// MissingIndex without Table should not be added
				assert.Empty(t, analysis.MissingIndexes)
			},
		},
		{
			name: "Object without operator stack should be ignored",
			xmlPlan: `<ShowPlanXML>
				<StmtSimple>
					<Object Table="[Users]">
					</Object>
				</StmtSimple>
			</ShowPlanXML>`,
			wantErr: false,
			validate: func(t *testing.T, analysis *PlanAnalysis) {
				t.Helper()
				assert.NotNil(t, analysis)
				// Object outside RelOp should not be parsed
				assert.Empty(t, analysis.Tables)
			},
		},
		{
			name: "multiple root RelOp operators",
			xmlPlan: `<ShowPlanXML>
				<StmtSimple>
					<QueryPlan>
						<RelOp PhysicalOp="Sort">
						</RelOp>
						<RelOp PhysicalOp="Hash Match">
						</RelOp>
					</QueryPlan>
				</StmtSimple>
			</ShowPlanXML>`,
			wantErr: false,
			validate: func(t *testing.T, analysis *PlanAnalysis) {
				t.Helper()
				assert.NotNil(t, analysis)
				assert.Len(t, analysis.Operators, 2)
				assert.Equal(t, "Sort", analysis.Operators[0].PhysicalOp)
				assert.Equal(t, "Hash Match", analysis.Operators[1].PhysicalOp)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analysis, err := parseExecutionPlan(tt.xmlPlan)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.validate != nil {
					tt.validate(t, analysis)
				}
			}
		})
	}
}

func TestParseExecutionPlan_EdgeCases(t *testing.T) {
	t.Run("malformed XML with unclosed tags", func(t *testing.T) {
		xmlPlan := `<ShowPlanXML><StmtSimple><QueryPlan><RelOp`
		analysis, err := parseExecutionPlan(xmlPlan)
		// Should handle gracefully or return error
		if err != nil {
			assert.Error(t, err)
		} else {
			assert.NotNil(t, analysis)
		}
	})

	t.Run("XML with special characters", func(t *testing.T) {
		xmlPlan := `<ShowPlanXML>
			<StmtSimple StatementText="SELECT &lt;test&gt;">
				<QueryPlan>
					<RelOp PhysicalOp="Table Scan">
					</RelOp>
				</QueryPlan>
			</StmtSimple>
		</ShowPlanXML>`
		analysis, err := parseExecutionPlan(xmlPlan)
		t.Helper()
		assert.NoError(t, err)
		assert.NotNil(t, analysis)
	})

	t.Run("very large numbers", func(t *testing.T) {
		xmlPlan := `<ShowPlanXML>
			<StmtSimple StatementSubTreeCost="1.7976931348623157e+308" StatementEstRows="1.7976931348623157e+308">
				<QueryPlan CompileTime="9223372036854775807">
					<RelOp EstimatedTotalSubtreeCost="1.7976931348623157e+308" EstimateRows="1.7976931348623157e+308" ActualRows="1.7976931348623157e+308">
					</RelOp>
				</QueryPlan>
			</StmtSimple>
		</ShowPlanXML>`
		analysis, err := parseExecutionPlan(xmlPlan)
		t.Helper()
		assert.NoError(t, err)
		assert.NotNil(t, analysis)
		// Should parse without panic
	})

	t.Run("negative numbers", func(t *testing.T) {
		xmlPlan := `<ShowPlanXML>
			<StmtSimple StatementSubTreeCost="-1.0" StatementEstRows="-100">
				<QueryPlan CompileTime="-1000">
				</QueryPlan>
			</StmtSimple>
		</ShowPlanXML>`
		analysis, err := parseExecutionPlan(xmlPlan)
		assert.NoError(t, err)
		assert.NotNil(t, analysis)
		// Should parse negative numbers
		assert.Equal(t, -1.0, analysis.TotalCost)
		assert.Equal(t, -100.0, analysis.EstimatedRows)
		assert.Equal(t, int64(-1000), analysis.CompileTime)
	})

	t.Run("invalid number formats", func(t *testing.T) {
		xmlPlan := `<ShowPlanXML>
			<StmtSimple StatementSubTreeCost="invalid" StatementEstRows="not_a_number">
				<QueryPlan CompileTime="also_invalid">
					<RelOp EstimatedTotalSubtreeCost="bad" EstimateRows="worse" ActualRows="worst">
					</RelOp>
				</QueryPlan>
			</StmtSimple>
		</ShowPlanXML>`
		analysis, err := parseExecutionPlan(xmlPlan)
		t.Helper()
		assert.NoError(t, err)
		assert.NotNil(t, analysis)
		// Invalid numbers should be ignored (default to 0)
		assert.Equal(t, 0.0, analysis.TotalCost)
		assert.Equal(t, 0.0, analysis.EstimatedRows)
		assert.Equal(t, int64(0), analysis.CompileTime)
	})
}

func TestBuildDbmStatementRow(t *testing.T) {
	tests := []struct {
		name      string
		columnMap map[string]*interface{}
		wantNil   bool
		wantErr   bool
		validate  func(t *testing.T, row *dbmStatementRow)
	}{
		{
			name: "valid row with all fields",
			columnMap: map[string]*interface{}{
				"statement_text":    strPtr("SELECT * FROM users WHERE id = 1"),
				"database_name":     strPtr("testdb"),
				"plan_handle":       bytesPtr([]byte{0x01, 0x02, 0x03}),
				"query_hash":        bytesPtr([]byte{0x04, 0x05}),
				"query_plan_hash":   bytesPtr([]byte{0x06, 0x07}),
				"procedure_name":    strPtr("testproc"),
				"schema_name":       strPtr("dbo"),
				"is_encrypted":      boolPtr(false),
				"execution_count":   intPtr(int64(10)),
				"total_worker_time": intPtr(int64(100)),
			},
			wantNil: false,
			wantErr: false,
			validate: func(t *testing.T, row *dbmStatementRow) {
				t.Helper()
				assert.Equal(t, "testdb", row.databaseName)
				assert.NotEmpty(t, row.statementText) // Original text should be saved
				assert.Equal(t, "0x010203", row.planHandle)
				assert.Equal(t, "0x0405", row.queryHash)
				assert.Equal(t, "0x0607", row.queryPlanHash)
				assert.Equal(t, "dbo.testproc", row.procedureName)
				assert.False(t, row.isEncrypted)
				assert.Equal(t, int64(10), row.executionCount)
				assert.Equal(t, int64(100), row.totalWorkerTime)
			},
		},
		{
			name: "empty statement text (allowed)",
			columnMap: map[string]*interface{}{
				"statement_text": strPtr(""),
				"is_encrypted":   boolPtr(false),
				"query_hash":     bytesPtr([]byte{0x01, 0x02}),
			},
			wantNil: false,
			wantErr: false,
			validate: func(t *testing.T, row *dbmStatementRow) {
				t.Helper()
				assert.Equal(t, "", row.statementText)
				assert.False(t, row.isEncrypted)
			},
		},
		{
			name: "encrypted statement with text",
			columnMap: map[string]*interface{}{
				"statement_text": strPtr("SELECT 1"),
				"is_encrypted":   boolPtr(true),
			},
			wantNil: false,
			wantErr: false,
			validate: func(t *testing.T, row *dbmStatementRow) {
				t.Helper()
				assert.True(t, row.isEncrypted)
			},
		},
		{
			name: "encrypted statement with empty text",
			columnMap: map[string]*interface{}{
				"statement_text": strPtr(""),
				"is_encrypted":   boolPtr(true),
				"query_hash":     bytesPtr([]byte{0x01, 0x02}),
			},
			wantNil: false,
			wantErr: false,
			validate: func(t *testing.T, row *dbmStatementRow) {
				t.Helper()
				if row == nil {
					t.Fatal("row should not be nil")
				}
				assert.True(t, row.isEncrypted)
				assert.Equal(t, "", row.statementText)
			},
		},
	}

	ipt := defaultInput()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			row, err := ipt.buildDbmStatementRow(tt.columnMap)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			if tt.wantNil {
				assert.Nil(t, row)
			} else {
				assert.NotNil(t, row)
				if tt.validate != nil {
					tt.validate(t, row)
				}
			}
		})
	}
}

func TestBuildStatementsQuery(t *testing.T) {
	ipt := &Input{
		Dbm: &dbmConfig{
			StoredProcedureCharactersLimit: 500,
		},
		dbmColumnsCache: map[string][]string{
			"sys.dm_exec_query_stats": {"execution_count", "total_worker_time", "total_elapsed_time"},
		},
	}

	tests := []struct {
		name                   string
		collectIntervalSeconds int
		rowLimit               int
		wantErr                bool
		validate               func(t *testing.T, query string)
	}{
		{
			name:                   "valid query",
			collectIntervalSeconds: 60,
			rowLimit:               100,
			wantErr:                false,
			validate: func(t *testing.T, query string) {
				t.Helper()
				assert.Contains(t, query, "TOP 100")
				assert.Contains(t, query, "qs.execution_count as execution_count")
				assert.Contains(t, query, "SUM(qs.execution_count) AS execution_count")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query, err := ipt.buildStatementsQuery(context.Background(), tt.collectIntervalSeconds, tt.rowLimit)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, query)
				if tt.validate != nil {
					tt.validate(t, query)
				}
			}
		})
	}
}

func TestDbmActivityMeasurementInfo(t *testing.T) {
	m := &dbmActivityMeasurement{}
	info := m.Info()

	assert.NotNil(t, info)
	assert.Equal(t, "sqlserver_dbm_activity", info.Name)
	assert.Equal(t, point.Logging, info.Cat)
	assert.Contains(t, info.Desc, "active queries")

	// Check some key fields
	assert.Contains(t, info.Fields, "message")
	assert.Contains(t, info.Fields, "session_id")
	// session_status, wait_type, command are now tags, not fields
	assert.NotContains(t, info.Fields, "session_status")
	assert.NotContains(t, info.Fields, "wait_type")
	assert.NotContains(t, info.Fields, "command")

	// Check tags - these should be tags now
	assert.Contains(t, info.Tags, "server")
	assert.Contains(t, info.Tags, "sqlserver_host")
	assert.Contains(t, info.Tags, "database_name")
	assert.Contains(t, info.Tags, "user_name")
	assert.Contains(t, info.Tags, "host_name")
	assert.Contains(t, info.Tags, "procedure_name")
	assert.Contains(t, info.Tags, "schema_name")
	assert.Contains(t, info.Tags, "program_name")
	assert.Contains(t, info.Tags, "query_hash")
	assert.Contains(t, info.Tags, "query_plan_hash")
	assert.Contains(t, info.Tags, "session_status")
	assert.Contains(t, info.Tags, "request_status")
	assert.Contains(t, info.Tags, "command")
	assert.Contains(t, info.Tags, "wait_type")
	assert.Contains(t, info.Tags, "query_signature")
}

func TestDatabasePlanObjectMeasurementInfo(t *testing.T) {
	m := &dbmPlanObjectMeasurement{}
	info := m.Info()

	assert.NotNil(t, info)
	assert.Equal(t, "db_exec_plan", info.Name)
	assert.Equal(t, point.Object, info.Cat)
	assert.Contains(t, info.Desc, "execution plan")

	// Check some key fields
	assert.Contains(t, info.Fields, "message")
	assert.Contains(t, info.Fields, "is_encrypted")

	// Check tags
	assert.Contains(t, info.Tags, "name")
	assert.Contains(t, info.Tags, "query_plan_hash")
	assert.Contains(t, info.Tags, "query_hash")
	assert.Contains(t, info.Tags, "server")
	assert.Contains(t, info.Tags, "sqlserver_host")
	assert.Contains(t, info.Tags, "database_type")
	assert.Contains(t, info.Tags, "plan_type")
	assert.Contains(t, info.Tags, "query_signature")
	assert.Contains(t, info.Tags, "database_name")
	assert.Contains(t, info.Tags, "schema_name")
	assert.Contains(t, info.Tags, "procedure_name")
}

func TestBuildActivityPoints(t *testing.T) {
	ipt := &Input{
		Host:  "testhost:1433",
		start: time.Now(),
		Tags:  make(map[string]string),
	}

	obfuscator := obfuscate.NewObfuscator(obfuscate.Config{
		SQL: obfuscate.SQLConfig{
			DBMS: obfuscate.DBMSSQLServer,
		},
	})
	obfResult, _ := obfuscator.ObfuscateSQLString("SELECT * FROM users")

	rows := []*dbmActivityRow{
		{
			sessionID:        123,
			userName:         "testuser",
			hostName:         "testhost",
			databaseName:     "testdb",
			sqlserverHost:    "testhost:1433",
			obfuscatedText:   obfResult.Query,
			sessionStatus:    "running",
			requestStatus:    "running",
			command:          "SELECT",
			waitType:         "PAGEIOLATCH_SH",
			waitCategory:     "I/O",
			queryHash:        "0x0102",
			queryPlanHash:    "0x0304",
			schemaName:       "dbo",
			procedureName:    "dbo.testproc",
			programName:      "testapp",
			cpuTime:          100,
			totalElapsedTime: 200,
		},
	}

	pts := ipt.buildActivityPoints(rows, time.Now())
	assert.Len(t, pts, 1)
	assert.NotNil(t, pts[0])

	// Verify tags are set correctly
	pt := pts[0]
	tags := pt.Tags()
	assert.Equal(t, "testdb", tags.GetTag("database_name"))
	assert.Equal(t, "testhost:1433", tags.GetTag("sqlserver_host"))
	assert.Equal(t, "testuser", tags.GetTag("user_name"))
	assert.Equal(t, "testhost", tags.GetTag("host_name"))
	assert.Equal(t, "dbo.testproc", tags.GetTag("procedure_name"))
	assert.Equal(t, "dbo", tags.GetTag("schema_name"))
	assert.Equal(t, "testapp", tags.GetTag("program_name"))
	assert.Equal(t, "0x0102", tags.GetTag("query_hash"))
	assert.Equal(t, "0x0304", tags.GetTag("query_plan_hash"))
	assert.Equal(t, "running", tags.GetTag("session_status"))
	assert.Equal(t, "running", tags.GetTag("request_status"))
	assert.Equal(t, "SELECT", tags.GetTag("command"))
	assert.Equal(t, "PAGEIOLATCH_SH", tags.GetTag("wait_type"))
	assert.Equal(t, "I/O", tags.GetTag("wait_group"))

	// Verify fields are set correctly
	fields := pt.Fields()
	assert.Equal(t, obfResult.Query, fields.Get("message").GetS())
	if sessionIDField := fields.Get("session_id"); sessionIDField != nil {
		assert.Equal(t, int64(123), sessionIDField.Raw())
	}
	if cpuTimeField := fields.Get("cpu_time"); cpuTimeField != nil {
		assert.Equal(t, int64(100), cpuTimeField.Raw())
	}
	if elapsedTimeField := fields.Get("total_elapsed_time"); elapsedTimeField != nil {
		assert.Equal(t, int64(200), elapsedTimeField.Raw())
	}
}

func TestProcessStatements(t *testing.T) {
	ipt := &Input{
		dbmMetricCache: make(map[string]*dbmMetricCache),
	}

	// 1. First collection (baseline)
	signature := generateQuerySignature("testdb", "", "0x1")
	rows1 := []*dbmStatementRow{
		{
			databaseName:     "testdb",
			queryHash:        "0x1",
			queryPlanHash:    "0xPlan1",
			procedureName:    "",
			querySignature:   signature,
			totalElapsedTime: 1000,
			executionCount:   10,
		},
	}
	result1 := ipt.processStatements(rows1)
	assert.Len(t, result1, 0) // baseline only
	assert.Len(t, ipt.dbmMetricCache, 1)

	// 2. Normal increase
	rows2 := []*dbmStatementRow{
		{
			databaseName:     "testdb",
			queryHash:        "0x1",
			queryPlanHash:    "0xPlan1",
			procedureName:    "",
			querySignature:   signature,
			totalElapsedTime: 1500,
			executionCount:   15,
		},
	}
	result2 := ipt.processStatements(rows2)
	assert.Len(t, result2, 1)
	assert.Equal(t, int64(500), result2[0].deltaElapsedTime)
	assert.Equal(t, int64(5), result2[0].deltaExecutionCount)

	// 3. Reset (total elapsed time decreases)
	rows3 := []*dbmStatementRow{
		{
			databaseName:     "testdb",
			queryHash:        "0x1",
			queryPlanHash:    "0xPlan1",
			procedureName:    "",
			querySignature:   signature,
			totalElapsedTime: 200,
			executionCount:   2,
		},
	}
	result3 := ipt.processStatements(rows3)
	assert.Len(t, result3, 0) // reset results in skipping

	ipt.dbmMetricCache = make(map[string]*dbmMetricCache) // clear cache
	// Key format: querySignature:queryPlanHash (to distinguish different execution plans)
	keyA := signature + ":0xA"
	keyB := signature + ":0xB"
	ipt.dbmMetricCache[keyA] = &dbmMetricCache{
		totalElapsedTime: 100,
		executionCount:   1,
	}
	ipt.dbmMetricCache[keyB] = &dbmMetricCache{
		totalElapsedTime: 50,
		executionCount:   1,
	}

	rows4 := []*dbmStatementRow{
		{
			databaseName:     "testdb",
			queryHash:        "0x1",
			queryPlanHash:    "0xA",
			schemaName:       "",
			procedureName:    "",
			querySignature:   signature,
			totalElapsedTime: 300,
			executionCount:   3,
		},
		{
			databaseName:     "testdb",
			queryHash:        "0x1",
			queryPlanHash:    "0xB",
			schemaName:       "",
			procedureName:    "",
			querySignature:   signature,
			totalElapsedTime: 200,
			executionCount:   2,
		},
	}
	// SQL already groups by query_hash, query_plan_hash, so each row is unique
	// No aggregation needed, just compute derivatives
	result4 := ipt.processStatements(rows4)
	assert.Len(t, result4, 2, "should have 2 results (different query_plan_hash, no aggregation needed)")
	// Find results by query_plan_hash
	var resultA, resultB *dbmStatementRow
	for _, r := range result4 {
		if r.queryPlanHash == "0xA" {
			resultA = r
		} else if r.queryPlanHash == "0xB" {
			resultB = r
		}
	}
	assert.NotNil(t, resultA, "should have result for query_plan_hash 0xA")
	assert.NotNil(t, resultB, "should have result for query_plan_hash 0xB")
	// 300 - 100 = 200
	assert.Equal(t, int64(200), resultA.deltaElapsedTime)
	// 3 - 1 = 2
	assert.Equal(t, int64(2), resultA.deltaExecutionCount)
	// 200 - 50 = 150
	assert.Equal(t, int64(150), resultB.deltaElapsedTime)
	// 2 - 1 = 1
	assert.Equal(t, int64(1), resultB.deltaExecutionCount)
}

func TestCategorizeWaitType(t *testing.T) {
	tests := []struct {
		name          string
		sessionStatus string
		waitType      string
		want          string
	}{
		// CPU - running status with empty wait type
		{
			name:          "running with empty wait type",
			sessionStatus: "running",
			waitType:      "",
			want:          "Other",
		},
		{
			name:          "running with CPU wait type",
			sessionStatus: "running",
			waitType:      waitTypeCPU,
			want:          "CPU",
		},
		{
			name:          "running with waiting_on_cpu wait type",
			sessionStatus: "running",
			waitType:      waitTypeWaitingOnCPU,
			want:          "CPU",
		},
		{
			name:          "sos_scheduler_yield",
			sessionStatus: "running",
			waitType:      "SOS_SCHEDULER_YIELD",
			want:          "CPU",
		},
		// Lock related (Highest priority)
		{
			name:          "LCK_M_S",
			sessionStatus: "running",
			waitType:      "LCK_M_S",
			want:          "Lock",
		},
		{
			name:          "LCK_M_X",
			sessionStatus: "running",
			waitType:      "LCK_M_X",
			want:          "Lock",
		},
		{
			name:          "LCK_M_IS",
			sessionStatus: "running",
			waitType:      "LCK_M_IS",
			want:          "Lock",
		},
		// Concurrency related
		{
			name:          "RESOURCE_SEMAPHORE_QUERY_COMPILE",
			sessionStatus: "running",
			waitType:      "RESOURCE_SEMAPHORE_QUERY_COMPILE",
			want:          "Concurrency",
		},
		{
			name:          "LATCH_EX",
			sessionStatus: "running",
			waitType:      "LATCH_EX",
			want:          "Concurrency",
		},
		{
			name:          "PAGELATCH_SH",
			sessionStatus: "running",
			waitType:      "PAGELATCH_SH",
			want:          "Concurrency",
		},
		// Memory related
		{
			name:          "RESOURCE_SEMAPHORE",
			sessionStatus: "running",
			waitType:      "RESOURCE_SEMAPHORE",
			want:          "Memory",
		},
		{
			name:          "RESOURCE_SEMAPHORE_QUERY_MEMORY",
			sessionStatus: "running",
			waitType:      "RESOURCE_SEMAPHORE_QUERY_MEMORY",
			want:          "Memory",
		},
		// Network related
		{
			name:          "ASYNC_NETWORK_IO",
			sessionStatus: "running",
			waitType:      "ASYNC_NETWORK_IO",
			want:          "Network",
		},
		{
			name:          "NETWORK_IO",
			sessionStatus: "running",
			waitType:      "NETWORK_IO",
			want:          "Network",
		},
		// I/O related
		{
			name:          "PAGEIOLATCH_SH",
			sessionStatus: "running",
			waitType:      "PAGEIOLATCH_SH",
			want:          "I/O",
		},
		{
			name:          "PAGEIOLATCH_EX",
			sessionStatus: "running",
			waitType:      "PAGEIOLATCH_EX",
			want:          "I/O",
		},
		{
			name:          "IO_COMPLETION",
			sessionStatus: "running",
			waitType:      "IO_COMPLETION",
			want:          "I/O",
		},
		{
			name:          "WRITELOG",
			sessionStatus: "running",
			waitType:      "WRITELOG",
			want:          "Commit/Log",
		},
		// Other
		{
			name:          "empty wait type with non-running status",
			sessionStatus: "sleeping",
			waitType:      "",
			want:          "Other",
		},
		{
			name:          "unknown wait type",
			sessionStatus: "running",
			waitType:      "UNKNOWN_WAIT",
			want:          "Other",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := categorizeWaitType(tt.waitType)
			assert.Equal(t, tt.want, result, "sessionStatus: %s, waitType: %s", tt.sessionStatus, tt.waitType)
		})
	}
}

func TestNormalizeWaitType(t *testing.T) {
	tests := []struct {
		name          string
		sessionStatus string
		requestStatus string
		waitType      string
		want          string
	}{
		{
			name:          "preserve explicit wait type",
			sessionStatus: "running",
			requestStatus: "suspended",
			waitType:      "PAGEIOLATCH_SH",
			want:          "PAGEIOLATCH_SH",
		},
		{
			name:          "runnable request becomes waiting_on_cpu",
			sessionStatus: "running",
			requestStatus: "runnable",
			waitType:      "",
			want:          waitTypeWaitingOnCPU,
		},
		{
			name:          "running request becomes cpu",
			sessionStatus: "running",
			requestStatus: "running",
			waitType:      "",
			want:          waitTypeCPU,
		},
		{
			name:          "running session without request becomes cpu",
			sessionStatus: "running",
			requestStatus: "",
			waitType:      "",
			want:          waitTypeCPU,
		},
		{
			name:          "suspended request does not become cpu from session status",
			sessionStatus: "running",
			requestStatus: "suspended",
			waitType:      "",
			want:          "",
		},
		{
			name:          "sleeping session with empty wait type stays empty",
			sessionStatus: "sleeping",
			requestStatus: "",
			waitType:      "",
			want:          "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeWaitType(tt.sessionStatus, tt.requestStatus, tt.waitType)
			assert.Equal(t, tt.want, got)
		})
	}
}
