// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package postgresql

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMapPostgreSQLWaitGroup(t *testing.T) {
	tests := []struct {
		name          string
		sessionStatus string
		waitEventType string
		waitEvent     string
		want          string
	}{
		{
			name:          "active session without wait event is cpu",
			sessionStatus: "active",
			waitEventType: "",
			waitEvent:     "",
			want:          waitGroupCPU,
		},
		{
			name:          "lock wait",
			sessionStatus: "active",
			waitEventType: "Lock",
			waitEvent:     "transactionid",
			want:          waitGroupLock,
		},
		{
			name:          "lwlock wal maps to commit log",
			sessionStatus: "active",
			waitEventType: "LWLock",
			waitEvent:     "WALWrite",
			want:          waitGroupCommitLog,
		},
		{
			name:          "bufferpin maps to concurrency",
			sessionStatus: "active",
			waitEventType: "BufferPin",
			waitEvent:     "BufferPin",
			want:          waitGroupConcurrency,
		},
		{
			name:          "io wal maps to commit log",
			sessionStatus: "active",
			waitEventType: "IO",
			waitEvent:     "WALSync",
			want:          waitGroupCommitLog,
		},
		{
			name:          "client wait maps to network",
			sessionStatus: "active",
			waitEventType: "Client",
			waitEvent:     "ClientRead",
			want:          waitGroupNetwork,
		},
		{
			name:          "ipc syncrep maps to commit log",
			sessionStatus: "active",
			waitEventType: "IPC",
			waitEvent:     "SyncRep",
			want:          waitGroupCommitLog,
		},
		{
			name:          "timeout wait maps to other",
			sessionStatus: "idle",
			waitEventType: "Timeout",
			waitEvent:     "PgSleep",
			want:          waitGroupOther,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mapPostgreSQLWaitGroup(tt.sessionStatus, tt.waitEventType, tt.waitEvent)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestAggregatePgSessions(t *testing.T) {
	rows := []map[string]any{
		{
			"datname":         "app",
			"usename":         "testuser",
			"backend_type":    "client backend",
			"pid":             int64(10),
			"state":           "active",
			"wait_event_type": "",
			"wait_event":      "",
			"wait_group":      waitGroupCPU,
			"blocking_pids":   "",
			"now":             int64(200),
			"query_start":     int64(100),
		},
		{
			"datname":         "app",
			"usename":         "testuser",
			"backend_type":    "client backend",
			"pid":             int64(11),
			"state":           "active",
			"wait_event_type": "Lock",
			"wait_event":      "transactionid",
			"wait_group":      waitGroupLock,
			"blocking_pids":   "10",
			"now":             int64(260),
			"query_start":     int64(160),
		},
		{
			"datname":         "app",
			"usename":         "testuser",
			"backend_type":    "client backend",
			"pid":             int64(12),
			"state":           "active",
			"wait_event_type": "Lock",
			"wait_event":      "transactionid",
			"wait_group":      waitGroupLock,
			"blocking_pids":   "10",
			"now":             int64(300),
			"query_start":     int64(260),
		},
	}

	aggregated := aggregatePgSessions(rows)
	if assert.Len(t, aggregated, 2) {
		var cpuGroup *aggregatedPgSession
		var blockedLockGroup *aggregatedPgSession
		for _, row := range aggregated {
			switch {
			case row.waitGroup == waitGroupCPU:
				cpuGroup = row
			case row.sessionStatus == "blocked" && row.waitGroup == waitGroupLock:
				blockedLockGroup = row
			}
		}

		if assert.NotNil(t, cpuGroup) {
			assert.Equal(t, "active", cpuGroup.sessionStatus)
			assert.Equal(t, int64(1), cpuGroup.sessionCount)
			assert.Equal(t, int64(0), cpuGroup.sessionBlockedCount)
		}

		if assert.NotNil(t, blockedLockGroup) {
			assert.Equal(t, "blocked", blockedLockGroup.sessionStatus)
			assert.Equal(t, "Lock", blockedLockGroup.waitGroup)
			assert.Equal(t, "Lock", blockedLockGroup.waitEventType)
			assert.Equal(t, "transactionid", blockedLockGroup.waitEvent)
			assert.Equal(t, int64(2), blockedLockGroup.sessionCount)
			assert.Equal(t, int64(2), blockedLockGroup.sessionBlockedCount)
		}
	}
}

func TestAggregatePgSessionsSkipsIdleClientBackends(t *testing.T) {
	rows := []map[string]any{
		{
			"datname":         "app",
			"usename":         "alice",
			"backend_type":    "client backend",
			"pid":             int64(10),
			"state":           "idle",
			"wait_event_type": "Client",
			"wait_event":      "ClientRead",
			"wait_group":      waitGroupNetwork,
			"blocking_pids":   "",
			"now":             int64(300),
			"query_start":     int64(100),
		},
		{
			"datname":         "app",
			"usename":         "alice",
			"backend_type":    "client backend",
			"pid":             int64(11),
			"state":           "idle in transaction",
			"wait_event_type": "Client",
			"wait_event":      "ClientRead",
			"wait_group":      waitGroupNetwork,
			"blocking_pids":   "",
			"now":             int64(320),
			"query_start":     int64(120),
		},
	}

	filtered := make([]map[string]any, 0, len(rows))
	for _, row := range rows {
		if shouldCollectPGActivity(row) {
			filtered = append(filtered, row)
		}
	}

	aggregated := aggregatePgSessions(filtered)
	if assert.Len(t, aggregated, 1) {
		row := aggregated[0]
		assert.Equal(t, "idle", row.sessionStatus)
		assert.Equal(t, waitGroupNetwork, row.waitGroup)
		assert.Equal(t, int64(1), row.sessionCount)
	}
}

func TestBuildDbmSessionPoints(t *testing.T) {
	ipt := &Input{
		mergedTags: map[string]string{
			"server":            "127.0.0.1:5432",
			"database_instance": "pg-test",
		},
	}

	rows := []*aggregatedPgSession{
		{
			datname:             "app",
			usename:             "testuser",
			applicationName:     "psql",
			clientAddr:          "10.0.0.8",
			sessionStatus:       "blocked",
			waitEventType:       "Lock",
			waitEvent:           "transactionid",
			waitGroup:           waitGroupLock,
			sessionCount:        2,
			sessionBlockedCount: 2,
		},
	}

	pts := ipt.buildDbmSessionPoints(rows, time.Unix(1700000000, 0))
	if assert.Len(t, pts, 1) {
		pt := pts[0]
		tags := pt.Tags()
		fields := pt.Fields()

		assert.Equal(t, "127.0.0.1:5432", tags.GetTag("server"))
		assert.Equal(t, "pg-test", tags.GetTag("database_instance"))
		assert.Equal(t, "app", tags.GetTag("db"))
		assert.Equal(t, "testuser", tags.GetTag("usename"))
		assert.Equal(t, "psql", tags.GetTag("application_name"))
		assert.Equal(t, "10.0.0.8", tags.GetTag("client_addr"))
		assert.Equal(t, "blocked", tags.GetTag("session_status"))
		assert.Equal(t, "Lock", tags.GetTag("wait_event_type"))
		assert.Equal(t, "transactionid", tags.GetTag("wait_event"))
		assert.Equal(t, "Lock", tags.GetTag("wait_group"))

		if sessionCount := fields.Get("session_group_count"); sessionCount != nil {
			assert.Equal(t, int64(2), sessionCount.Raw())
		}
		if blockedCount := fields.Get("session_blocked_count"); blockedCount != nil {
			assert.Equal(t, int64(2), blockedCount.Raw())
		}
		assert.Nil(t, fields.Get("session_blocking_count"))
		assert.Nil(t, fields.Get("session_total_query_duration"))
		assert.Nil(t, fields.Get("session_max_query_duration"))
	}
}

func TestBuildDbmSessionPointsSetsZeroBlockedCountForActiveGroup(t *testing.T) {
	ipt := &Input{
		mergedTags: map[string]string{
			"server":            "127.0.0.1:5432",
			"database_instance": "pg-test",
		},
	}

	rows := []*aggregatedPgSession{
		{
			datname:             "app",
			usename:             "testuser",
			sessionStatus:       "active",
			waitEventType:       "",
			waitEvent:           "",
			waitGroup:           waitGroupCPU,
			sessionCount:        3,
			sessionBlockedCount: 0,
		},
	}

	pts := ipt.buildDbmSessionPoints(rows, time.Unix(1700000000, 0))
	if assert.Len(t, pts, 1) {
		fields := pts[0].Fields()
		if sessionCount := fields.Get("session_group_count"); sessionCount != nil {
			assert.Equal(t, int64(3), sessionCount.Raw())
		}
		if blockedCount := fields.Get("session_blocked_count"); blockedCount != nil {
			assert.Equal(t, int64(0), blockedCount.Raw())
		}
	}
}
