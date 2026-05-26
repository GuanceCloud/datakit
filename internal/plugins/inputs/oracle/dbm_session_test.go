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

func TestAggregateSessions(t *testing.T) {
	ipt := &Input{
		Object: oracleObject{
			name: "testhost:1521",
		},
	}

	tests := []struct {
		name     string
		rows     []*OracleActivityRow
		wantLen  int
		validate func(t *testing.T, agg []*aggregatedSession)
	}{
		{
			name:    "empty rows",
			rows:    []*OracleActivityRow{},
			wantLen: 0,
		},
		{
			name: "single active session",
			rows: []*OracleActivityRow{
				{
					SessionID:      123,
					User:           "testuser",
					Status:         "ACTIVE",
					PdbName:        "TESTPDB",
					WaitEventClass: "CPU",
					WaitTimeMicro:  1000,
				},
			},
			wantLen: 1,
			validate: func(t *testing.T, agg []*aggregatedSession) {
				t.Helper()
				assert.Len(t, agg, 1)
				session := agg[0]
				assert.Equal(t, "TESTPDB", session.pdbName)
				assert.Equal(t, "testuser", session.username)
				assert.Equal(t, "active", session.status)
				assert.Equal(t, "CPU", session.waitClass)
				assert.Equal(t, int64(1), session.sessionCount)
				assert.Equal(t, int64(1), session.totalWaitTime) // 1000 microseconds / 1000 = 1 millisecond
			},
		},
		{
			name: "single idle session",
			rows: []*OracleActivityRow{
				{
					SessionID:      456,
					User:           "testuser",
					Status:         "INACTIVE",
					PdbName:        "TESTPDB",
					WaitEventClass: "Idle",
					WaitTimeMicro:  0,
				},
			},
			wantLen: 1,
			validate: func(t *testing.T, agg []*aggregatedSession) {
				t.Helper()
				assert.Len(t, agg, 1)
				session := agg[0]
				assert.Equal(t, "idle", session.status)
				assert.Equal(t, "Idle", session.waitClass)
			},
		},
		{
			name: "blocked session",
			rows: []*OracleActivityRow{
				{
					SessionID:            789,
					User:                 "testuser",
					Status:               "ACTIVE",
					PdbName:              "TESTPDB",
					WaitEventClass:       "Application",
					BlockingSession:      100,
					FinalBlockingSession: 100,
					WaitTimeMicro:        5000,
				},
			},
			wantLen: 1,
			validate: func(t *testing.T, agg []*aggregatedSession) {
				t.Helper()
				assert.Len(t, agg, 1)
				session := agg[0]
				assert.Equal(t, "blocked", session.status)
				assert.Equal(t, int64(1), session.blockedCount)
			},
		},
		{
			name: "blocking session (is_blocker)",
			rows: []*OracleActivityRow{
				{
					SessionID:            100,
					User:                 "blocker",
					Status:               "ACTIVE",
					PdbName:              "TESTPDB",
					WaitEventClass:       "CPU",
					BlockingSession:      0,
					FinalBlockingSession: 0,
				},
				{
					SessionID:            200,
					User:                 "blocked",
					Status:               "ACTIVE",
					PdbName:              "TESTPDB",
					WaitEventClass:       "Application",
					BlockingSession:      100,
					FinalBlockingSession: 100,
				},
			},
			wantLen: 2,
			validate: func(t *testing.T, agg []*aggregatedSession) {
				t.Helper()
				assert.Len(t, agg, 2)
				// Find the blocking session
				var blocker *aggregatedSession
				for _, s := range agg {
					if s.username == "blocker" {
						blocker = s
						break
					}
				}
				assert.NotNil(t, blocker)
				assert.Equal(t, int64(1), blocker.blockingCount)
			},
		},
		{
			name: "aggregate by pdb_name + username + status + wait_class",
			rows: []*OracleActivityRow{
				{
					SessionID:      1,
					User:           "user1",
					Status:         "ACTIVE",
					PdbName:        "PDB1",
					WaitEventClass: "CPU",
					WaitTimeMicro:  1000,
				},
				{
					SessionID:      2,
					User:           "user1",
					Status:         "ACTIVE",
					PdbName:        "PDB1",
					WaitEventClass: "CPU",
					WaitTimeMicro:  2000,
				},
				{
					SessionID:      3,
					User:           "user1",
					Status:         "ACTIVE",
					PdbName:        "PDB1",
					WaitEventClass: "I/O",
					WaitTimeMicro:  3000,
				},
			},
			wantLen: 2,
			validate: func(t *testing.T, agg []*aggregatedSession) {
				t.Helper()
				assert.Len(t, agg, 2)
				// Find CPU group
				var cpuGroup *aggregatedSession
				for _, s := range agg {
					if s.waitClass == "CPU" {
						cpuGroup = s
						break
					}
				}
				assert.NotNil(t, cpuGroup)
				assert.Equal(t, int64(2), cpuGroup.sessionCount)
				assert.Equal(t, int64(3), cpuGroup.totalWaitTime) // (1000 + 2000) / 1000
			},
		},
		{
			name: "empty wait_class defaults to Other",
			rows: []*OracleActivityRow{
				{
					SessionID:      123,
					User:           "testuser",
					Status:         "ACTIVE",
					PdbName:        "TESTPDB",
					WaitEventClass: "",
					WaitTimeMicro:  1000,
				},
			},
			wantLen: 1,
			validate: func(t *testing.T, agg []*aggregatedSession) {
				t.Helper()
				assert.Len(t, agg, 1)
				session := agg[0]
				assert.Equal(t, "Other", session.waitGroup)
			},
		},
		{
			name: "final_blocking_session takes precedence",
			rows: []*OracleActivityRow{
				{
					SessionID:            789,
					User:                 "testuser",
					Status:               "ACTIVE",
					PdbName:              "TESTPDB",
					WaitEventClass:       "Application",
					BlockingSession:      100,
					FinalBlockingSession: 200,
				},
			},
			wantLen: 1,
			validate: func(t *testing.T, agg []*aggregatedSession) {
				t.Helper()
				assert.Len(t, agg, 1)
				session := agg[0]
				assert.Equal(t, "blocked", session.status)
				assert.Equal(t, int64(1), session.blockedCount)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agg := ipt.aggregateSessions(tt.rows)
			assert.Len(t, agg, tt.wantLen)
			if tt.validate != nil {
				tt.validate(t, agg)
			}
		})
	}
}

func TestBuildSessionPoints(t *testing.T) {
	ipt := &Input{
		Object: oracleObject{
			name: "testhost:1521",
		},
		cdbName: "TESTCDB",
	}

	tests := []struct {
		name     string
		sessions []*aggregatedSession
		wantLen  int
		validate func(t *testing.T, pts []*point.Point)
	}{
		{
			name:     "empty sessions",
			sessions: []*aggregatedSession{},
			wantLen:  0,
		},
		{
			name: "single session",
			sessions: []*aggregatedSession{
				{
					server:        "testhost:1521",
					pdbName:       "TESTPDB",
					username:      "testuser",
					program:       "sqlplus",
					client:        "client-host",
					status:        "active",
					waitClass:     "CPU",
					waitGroup:     "CPU",
					sessionCount:  5,
					totalWaitTime: 1000,
					blockedCount:  0,
					blockingCount: 0,
				},
			},
			wantLen: 1,
			validate: func(t *testing.T, pts []*point.Point) {
				t.Helper()
				assert.Len(t, pts, 1)
				pt := pts[0]
				tags := pt.Tags()
				fields := pt.Fields()

				assert.Equal(t, "testhost:1521", tags.GetTag("server"))
				assert.Equal(t, "TESTCDB", tags.GetTag("cdb_name"))
				assert.Equal(t, "TESTPDB", tags.GetTag("pdb_name"))
				assert.Equal(t, "testuser", tags.GetTag("username"))
				assert.Equal(t, "sqlplus", tags.GetTag("program"))
				assert.Equal(t, "client-host", tags.GetTag("client"))
				assert.Equal(t, "active", tags.GetTag("session_status"))
				assert.Equal(t, "CPU", tags.GetTag("wait_group"))

				if sessionCount := fields.Get("session_group_count"); sessionCount != nil {
					assert.Equal(t, int64(5), sessionCount.Raw())
				}
				if totalWaitTime := fields.Get("session_total_wait_time"); totalWaitTime != nil {
					assert.Equal(t, int64(1000), totalWaitTime.Raw())
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ptsTime := time.Now()
			pts := ipt.buildSessionPoints(tt.sessions, ptsTime)
			assert.Len(t, pts, tt.wantLen)
			if tt.validate != nil {
				tt.validate(t, pts)
			}
		})
	}
}

func TestAggregateSessions_StatusDetermination(t *testing.T) {
	ipt := &Input{
		Object: oracleObject{
			name: "testhost:1521",
		},
	}

	tests := []struct {
		name           string
		row            *OracleActivityRow
		expectedStatus string
	}{
		{
			name: "ACTIVE status",
			row: &OracleActivityRow{
				SessionID:            123,
				User:                 "testuser",
				Status:               "ACTIVE",
				PdbName:              "TESTPDB",
				WaitEventClass:       "CPU",
				BlockingSession:      0,
				FinalBlockingSession: 0,
			},
			expectedStatus: "active",
		},
		{
			name: "INACTIVE status",
			row: &OracleActivityRow{
				SessionID:            123,
				User:                 "testuser",
				Status:               "INACTIVE",
				PdbName:              "TESTPDB",
				WaitEventClass:       "Idle",
				BlockingSession:      0,
				FinalBlockingSession: 0,
			},
			expectedStatus: "idle",
		},
		{
			name: "blocked session (blocking_session > 0)",
			row: &OracleActivityRow{
				SessionID:            123,
				User:                 "testuser",
				Status:               "ACTIVE",
				PdbName:              "TESTPDB",
				WaitEventClass:       "Application",
				BlockingSession:      100,
				FinalBlockingSession: 0,
			},
			expectedStatus: "blocked",
		},
		{
			name: "blocked session (final_blocking_session > 0)",
			row: &OracleActivityRow{
				SessionID:            123,
				User:                 "testuser",
				Status:               "INACTIVE",
				PdbName:              "TESTPDB",
				WaitEventClass:       "Application",
				BlockingSession:      0,
				FinalBlockingSession: 100,
			},
			expectedStatus: "blocked",
		},
		{
			name: "lowercase status",
			row: &OracleActivityRow{
				SessionID:            123,
				User:                 "testuser",
				Status:               "active",
				PdbName:              "TESTPDB",
				WaitEventClass:       "CPU",
				BlockingSession:      0,
				FinalBlockingSession: 0,
			},
			expectedStatus: "active", // lowercase "active" is converted to "ACTIVE" and matches
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agg := ipt.aggregateSessions([]*OracleActivityRow{tt.row})
			assert.Len(t, agg, 1)
			assert.Equal(t, tt.expectedStatus, agg[0].status)
		})
	}
}
