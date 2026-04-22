// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package mysql

import (
	"database/sql"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGetMetricRowsBaselineAndDelta(t *testing.T) {
	var snapshot map[string]dbmMetricCache

	// First collection: baseline only, no metric output.
	first := []dbmRow{
		{
			schemaName:         "db1",
			querySignature:     "q1",
			countStar:          10,
			sumTimerWait:       10000,
			sumLockTime:        2000,
			sumErrors:          2,
			sumRowsAffected:    20,
			sumRowsSent:        30,
			sumRowsExamined:    40,
			sumSelectScan:      5,
			sumSelectFullJoin:  6,
			sumNoIndexUsed:     7,
			sumNoGoodIndexUsed: 8,
		},
	}

	metricRows, snapshot := getMetricRows(first, snapshot)
	assert.Len(t, metricRows, 0)

	// Second collection: should generate delta values.
	second := []dbmRow{
		{
			schemaName:         "db1",
			querySignature:     "q1",
			countStar:          15,    // +5
			sumTimerWait:       25000, // +15000 => 15(ns after /1000)
			sumLockTime:        5000,  // +3000 => 3(ns after /1000)
			sumErrors:          5,     // +3
			sumRowsAffected:    33,    // +13
			sumRowsSent:        50,    // +20
			sumRowsExamined:    70,    // +30
			sumSelectScan:      7,     // +2
			sumSelectFullJoin:  10,    // +4
			sumNoIndexUsed:     9,     // +2
			sumNoGoodIndexUsed: 11,    // +3
		},
	}

	metricRows, snapshot = getMetricRows(second, snapshot)
	if assert.Len(t, metricRows, 1) {
		row := metricRows[0]
		assert.Equal(t, uint64(5), row.deltaCountStar)
		assert.Equal(t, uint64(15), row.deltaTimerWait)
		assert.Equal(t, uint64(3), row.deltaLockTime)
		assert.Equal(t, uint64(3), row.deltaErrors)
		assert.Equal(t, uint64(13), row.deltaRowsAffected)
		assert.Equal(t, uint64(20), row.deltaRowsSent)
		assert.Equal(t, uint64(30), row.deltaRowsExamined)
		assert.Equal(t, uint64(2), row.deltaSelectScan)
		assert.Equal(t, uint64(4), row.deltaSelectFullJoin)
		assert.Equal(t, uint64(2), row.deltaNoIndexUsed)
		assert.Equal(t, uint64(3), row.deltaNoGoodIndexUsed)
	}
	assert.Contains(t, snapshot, getRowKey("db1", "q1"))
}

func TestAggregateMysqlSessions(t *testing.T) {
	ipt := &Input{}
	rows := activityRowSlice{
		{
			ThreadID:           sql.NullString{String: "1", Valid: true},
			ProcesslistUser:    sql.NullString{String: "u1", Valid: true},
			ProcesslistHost:    sql.NullString{String: "h1", Valid: true},
			ProcesslistDB:      sql.NullString{String: "d1", Valid: true},
			ProcesslistCommand: sql.NullString{String: "Query", Valid: true},
			ProcesslistState:   sql.NullString{String: "running", Valid: true},
			WaitEvent:          sql.NullString{String: "wait/lock/table/sql/handler", Valid: true},
			WaitGroup:          "Lock",
		},
		{
			ThreadID:           sql.NullString{String: "2", Valid: true},
			ProcesslistUser:    sql.NullString{String: "u1", Valid: true},
			ProcesslistHost:    sql.NullString{String: "h1", Valid: true},
			ProcesslistDB:      sql.NullString{String: "d1", Valid: true},
			ProcesslistCommand: sql.NullString{String: "Query", Valid: true},
			ProcesslistState:   sql.NullString{String: "running", Valid: true},
			WaitEvent:          sql.NullString{String: "wait/lock/table/sql/handler", Valid: true},
			WaitGroup:          "Lock",
			BlockingThreadID:   sql.NullString{String: "1", Valid: true}, // thread 2 blocked by thread 1
		},
		{
			ThreadID:           sql.NullString{String: "3", Valid: true},
			ProcesslistUser:    sql.NullString{String: "u1", Valid: true},
			ProcesslistHost:    sql.NullString{String: "h1", Valid: true},
			ProcesslistDB:      sql.NullString{String: "d1", Valid: true},
			ProcesslistCommand: sql.NullString{String: "Query", Valid: true},
			ProcesslistState:   sql.NullString{String: "running", Valid: true},
			WaitEvent:          sql.NullString{String: "CPU", Valid: true},
			WaitGroup:          "CPU",
		},
	}

	aggregated := ipt.aggregateMysqlSessions(rows)
	assert.Len(t, aggregated, 3)

	var blockedLockGroup *mysqlAggregatedSession
	for _, g := range aggregated {
		if g.waitEvent == "wait/lock/table/sql/handler" && g.sessionStatus == "blocked" {
			blockedLockGroup = g
			break
		}
	}
	if assert.NotNil(t, blockedLockGroup) {
		assert.Equal(t, int64(1), blockedLockGroup.sessionCount)
		assert.Equal(t, int64(1), blockedLockGroup.blockedCount)
	}
}

func TestAggregateMysqlSessionsMergeSameSessionStatus(t *testing.T) {
	ipt := &Input{}
	rows := activityRowSlice{
		{
			ThreadID:           sql.NullString{String: "1", Valid: true},
			ProcesslistUser:    sql.NullString{String: "u1", Valid: true},
			ProcesslistHost:    sql.NullString{String: "h1", Valid: true},
			ProcesslistDB:      sql.NullString{String: "d1", Valid: true},
			ProcesslistCommand: sql.NullString{String: "Query", Valid: true},
			ProcesslistState:   sql.NullString{String: "running", Valid: true},
			WaitEvent:          sql.NullString{String: "CPU", Valid: true},
			WaitGroup:          "CPU",
		},
		{
			ThreadID:           sql.NullString{String: "2", Valid: true},
			ProcesslistUser:    sql.NullString{String: "u1", Valid: true},
			ProcesslistHost:    sql.NullString{String: "h1", Valid: true},
			ProcesslistDB:      sql.NullString{String: "d1", Valid: true},
			ProcesslistCommand: sql.NullString{String: "Query", Valid: true},
			ProcesslistState:   sql.NullString{String: "waiting for handler commit", Valid: true},
			WaitEvent:          sql.NullString{String: "CPU", Valid: true},
			WaitGroup:          "CPU",
		},
	}

	aggregated := ipt.aggregateMysqlSessions(rows)
	if assert.Len(t, aggregated, 1) {
		assert.Equal(t, "active", aggregated[0].sessionStatus)
		assert.Equal(t, int64(2), aggregated[0].sessionCount)
	}
}

func TestGetMySQLSessionStatus(t *testing.T) {
	t.Run("blocked takes precedence", func(t *testing.T) {
		status := getMySQLSessionStatus(activityRow{
			ProcesslistCommand: sql.NullString{String: "Sleep", Valid: true},
			ProcesslistState:   sql.NullString{String: "User sleep", Valid: true},
			BlockingThreadID:   sql.NullString{String: "10", Valid: true},
		})
		assert.Equal(t, "blocked", status)
	})

	t.Run("sleep command becomes idle", func(t *testing.T) {
		status := getMySQLSessionStatus(activityRow{
			ProcesslistCommand: sql.NullString{String: "Sleep", Valid: true},
			ProcesslistState:   sql.NullString{String: "", Valid: true},
		})
		assert.Equal(t, "idle", status)
	})

	t.Run("running query becomes active", func(t *testing.T) {
		status := getMySQLSessionStatus(activityRow{
			ProcesslistCommand: sql.NullString{String: "Query", Valid: true},
			ProcesslistState:   sql.NullString{String: "running", Valid: true},
		})
		assert.Equal(t, "active", status)
	})
}

func TestAggregateMysqlSessionsSplitBySessionStatus(t *testing.T) {
	ipt := &Input{}
	rows := activityRowSlice{
		{
			ThreadID:           sql.NullString{String: "1", Valid: true},
			ProcesslistUser:    sql.NullString{String: "u1", Valid: true},
			ProcesslistHost:    sql.NullString{String: "h1", Valid: true},
			ProcesslistDB:      sql.NullString{String: "d1", Valid: true},
			ProcesslistCommand: sql.NullString{String: "Query", Valid: true},
			ProcesslistState:   sql.NullString{String: "running", Valid: true},
			WaitEvent:          sql.NullString{String: "CPU", Valid: true},
			WaitGroup:          "CPU",
		},
		{
			ThreadID:           sql.NullString{String: "2", Valid: true},
			ProcesslistUser:    sql.NullString{String: "u1", Valid: true},
			ProcesslistHost:    sql.NullString{String: "h1", Valid: true},
			ProcesslistDB:      sql.NullString{String: "d1", Valid: true},
			ProcesslistCommand: sql.NullString{String: "Sleep", Valid: true},
			ProcesslistState:   sql.NullString{String: "User sleep", Valid: true},
			WaitEvent:          sql.NullString{String: "CPU", Valid: true},
			WaitGroup:          "CPU",
		},
	}

	aggregated := ipt.aggregateMysqlSessions(rows)
	if assert.Len(t, aggregated, 2) {
		var statuses []string
		for _, g := range aggregated {
			statuses = append(statuses, g.sessionStatus)
			assert.Equal(t, int64(1), g.sessionCount)
		}
		assert.ElementsMatch(t, []string{"active", "idle"}, statuses)
	}
}

func TestMetricCollectMysqlDbmSessionPoints(t *testing.T) {
	ipt := &Input{
		Host:       "127.0.0.1",
		mergedTags: map[string]string{"server": "127.0.0.1:3306", "database_instance": "mysql-test"},
	}
	rows := activityRowSlice{
		{
			ThreadID:           sql.NullString{String: "1", Valid: true},
			ProcesslistUser:    sql.NullString{String: "u1", Valid: true},
			ProcesslistHost:    sql.NullString{String: "h1", Valid: true},
			ProcesslistDB:      sql.NullString{String: "d1", Valid: true},
			ProcesslistCommand: sql.NullString{String: "Query", Valid: true},
			ProcesslistState:   sql.NullString{String: "running", Valid: true},
			WaitEvent:          sql.NullString{String: "CPU", Valid: true},
			WaitGroup:          "CPU",
		},
	}

	pts := ipt.metricCollectMysqlDbmSession(rows, time.Now())
	if assert.Len(t, pts, 1) {
		pt := pts[0]
		assert.Equal(t, metricNameMySQLDbmSession, pt.Name())
		s := pt.LineProto()
		assert.True(t, strings.Contains(s, "session_group_count=1i") || strings.Contains(s, "session_group_count=1"))
		assert.True(t, strings.Contains(s, "session_blocked_count=0i") || strings.Contains(s, "session_blocked_count=0"))

		tags := pt.Tags()
		fields := pt.Fields()
		assert.Equal(t, "127.0.0.1:3306", tags.GetTag("server"))
		assert.Equal(t, "mysql-test", tags.GetTag("database_instance"))
		assert.Equal(t, "u1", tags.GetTag("processlist_user"))
		assert.Equal(t, "h1", tags.GetTag("processlist_host"))
		assert.Equal(t, "d1", tags.GetTag("processlist_db"))
		assert.Equal(t, "active", tags.GetTag("session_status"))
		assert.Equal(t, "", tags.GetTag("processlist_state"))
		assert.Equal(t, "CPU", tags.GetTag("wait_event"))
		assert.Equal(t, "CPU", tags.GetTag("wait_group"))
		if sessionCount := fields.Get("session_group_count"); sessionCount != nil {
			assert.Equal(t, int64(1), sessionCount.Raw())
		}
		if blockedCount := fields.Get("session_blocked_count"); blockedCount != nil {
			assert.Equal(t, int64(0), blockedCount.Raw())
		}
		assert.Nil(t, fields.Get("session_blocking_count"))
	}
}

func TestMetricCollectMysqlDbmConnectionsPoints(t *testing.T) {
	ipt := &Input{
		Host:       "127.0.0.1",
		mergedTags: map[string]string{"server": "127.0.0.1:3306"},
	}
	connections := []connectionRow{
		{
			processlistUser:  sql.NullString{String: "u1", Valid: true},
			processlistHost:  sql.NullString{String: "h1", Valid: true},
			processlistDB:    sql.NullString{String: "d1", Valid: true},
			processlistState: sql.NullString{String: "running", Valid: true},
			connections:      sql.NullInt64{Int64: 3, Valid: true},
		},
	}

	pts := ipt.metricCollectMysqlDbmConnections(connections)
	if assert.Len(t, pts, 1) {
		assert.Equal(t, metricNameMySQLDbmConnection, pts[0].Name())
		s := pts[0].LineProto()
		assert.True(t, strings.Contains(s, "connection_count=3i") || strings.Contains(s, "connection_count=3"))
		assert.Contains(t, s, "processlist_user=u1")
		assert.Contains(t, s, "processlist_db=d1")
	}
}
