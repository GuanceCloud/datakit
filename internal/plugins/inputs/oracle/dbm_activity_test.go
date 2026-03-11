// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package oracle

import (
	"testing"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"github.com/stretchr/testify/assert"
)

func TestBuildDbmActivityPoints(t *testing.T) {
	ipt := &Input{
		Object: oracleObject{
			name: "testhost:1521",
		},
		cdbName: "TESTCDB",
	}

	tests := []struct {
		name     string
		rows     []*OracleActivityRow
		wantLen  int
		validate func(t *testing.T, pts []*point.Point)
	}{
		{
			name:    "empty rows",
			rows:    []*OracleActivityRow{},
			wantLen: 0,
		},
		{
			name: "single row with all fields",
			rows: []*OracleActivityRow{
				{
					SessionID:             123,
					SessionSerial:         456,
					User:                  "testuser",
					PdbName:               "TESTPDB",
					Status:                "ACTIVE",
					Type:                  "USER",
					WaitEventClass:        "CPU",
					WaitEvent:             "CPU",
					WaitTimeMicro:         1000,
					BlockingSession:       100,
					FinalBlockingSession:  100,
					BlockingInstance:      1,
					FinalBlockingInstance: 1,
					Statement:             "SELECT * FROM users",
					QuerySignature:        "abc123def456",
					OracleSQLRow: OracleSQLRow{
						SQLID:                  "abc123",
						SQLPlanHashValue:       987654321,
						ForceMatchingSignature: 1234567890,
						SQLExecStart:           "2024-01-01T00:01:00Z",
					},
					Port:             "1521",
					Process:          "12345",
					ClientInfo:       "test_client",
					ClientIdentifier: "test_identifier",
					LogonTime:        "2024-01-01T00:00:00Z",
					CommandName:      "SELECT",
				},
			},
			wantLen: 1,
			validate: func(t *testing.T, pts []*point.Point) {
				t.Helper()
				assert.Len(t, pts, 1)
				pt := pts[0]
				tags := pt.Tags()
				fields := pt.Fields()

				assert.Equal(t, "TESTCDB", tags.GetTag("cdb_name"))
				assert.Equal(t, "TESTPDB", tags.GetTag("pdb_name"))
				assert.Equal(t, "abc123def456", tags.GetTag("query_signature"))
				assert.Equal(t, "abc123", tags.GetTag("sql_id"))
				assert.Equal(t, "987654321", tags.GetTag("plan_hash_value"))
				assert.Equal(t, "1234567890", tags.GetTag("force_matching_signature"))
				assert.Equal(t, "ACTIVE", tags.GetTag("session_status"))
				assert.Equal(t, "USER", tags.GetTag("session_type"))
				assert.Equal(t, "CPU", tags.GetTag("wait_class"))

				if message := fields.Get("message"); message != nil {
					assert.Equal(t, "SELECT * FROM users", message.GetS())
				}
				if sessionID := fields.Get("session_id"); sessionID != nil {
					assert.Equal(t, int64(123), sessionID.Raw())
				}
				if waitTime := fields.Get("wait_time"); waitTime != nil {
					assert.Equal(t, int64(1000), waitTime.Raw())
				}
				if blockingSession := fields.Get("blocking_session_id"); blockingSession != nil {
					assert.Equal(t, int64(100), blockingSession.Raw())
				}
			},
		},
		{
			name: "row with zero blocking_session_id",
			rows: []*OracleActivityRow{
				{
					SessionID:     123,
					SessionSerial: 456,
					User:          "testuser",
					PdbName:       "TESTPDB",
					Status:        "ACTIVE",
					Statement:     "SELECT 1",
					OracleSQLRow: OracleSQLRow{
						SQLID: "test-sql-id",
					},
					BlockingSession:       0,
					FinalBlockingSession:  0,
					BlockingInstance:      0,
					FinalBlockingInstance: 0,
				},
			},
			wantLen: 1,
			validate: func(t *testing.T, pts []*point.Point) {
				t.Helper()
				assert.Len(t, pts, 1)
				pt := pts[0]
				fields := pt.Fields()
				if blockingSession := fields.Get("blocking_session_id"); blockingSession != nil {
					assert.Equal(t, int64(0), blockingSession.Raw())
				}
			},
		},
		{
			name: "row with empty statement",
			rows: []*OracleActivityRow{
				{
					SessionID:     123,
					SessionSerial: 456,
					User:          "testuser",
					PdbName:       "TESTPDB",
					Status:        "INACTIVE",
					Statement:     "",
				},
			},
			// Rows without SQLID or statement should be filtered out and not produce points.
			wantLen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ptsTime := time.Now()
			pts := ipt.buildDbmActivityPoints(tt.rows, ptsTime)
			assert.Len(t, pts, tt.wantLen)
			if tt.validate != nil {
				tt.validate(t, pts)
			}
		})
	}
}

func TestBuildDbmActivityPoints_OptionalFields(t *testing.T) {
	ipt := &Input{
		Object: oracleObject{
			name: "testhost:1521",
		},
		cdbName: "TESTCDB",
	}

	rows := []*OracleActivityRow{
		{
			SessionID:     123,
			SessionSerial: 456,
			User:          "testuser",
			PdbName:       "TESTPDB",
			Status:        "ACTIVE",
			Statement:     "SELECT 1",
			// Optional fields not set in OracleSQLRow
			OracleSQLRow: OracleSQLRow{
				SQLID:                  "abc123",
				SQLPlanHashValue:       0,
				ForceMatchingSignature: 0,
			},
		},
	}

	pts := ipt.buildDbmActivityPoints(rows, time.Now())
	assert.Len(t, pts, 1)
	pt := pts[0]
	tags := pt.Tags()

	// Optional tags should not be set when empty (plan_hash_value),
	// but sql_id should be set when present.
	assert.Equal(t, "abc123", tags.GetTag("sql_id"))
	assert.Empty(t, tags.GetTag("plan_hash_value"))
	// force_matching_signature is always set, even when 0 (it will be "0")
	assert.Equal(t, "0", tags.GetTag("force_matching_signature"))
}
