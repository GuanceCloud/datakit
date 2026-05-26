// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package mysql

import (
	"strings"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

// dbmSessionMeasurement describes mysql_dbm_session metrics for documentation/measurement info.
type dbmSessionMeasurement struct{}

func (m *dbmSessionMeasurement) Point() *point.Point {
	// This measurement is only used for metadata (Info), points are built in buildMysqlSessionPoints.
	return nil
}

func (m *dbmSessionMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: metricNameMySQLDbmSession,
		Desc: "MySQL DBM session metrics aggregated by user/host/db/session_status/wait_event (with wait_group for unified taxonomy).",
		Cat:  point.Metric,
		Fields: map[string]interface{}{
			"session_group_count": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of sessions in this dimension group",
			},
			"session_blocked_count": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of sessions that are being blocked",
			},
		},
		Tags: map[string]interface{}{
			"host":              &inputs.TagInfo{Desc: "The server host address"},
			"server":            &inputs.TagInfo{Desc: "The address of the server. The value is `host:port`"},
			"database_instance": &inputs.TagInfo{Desc: "MySQL instance identifier from configured tag or @@server_uuid."},
			"processlist_user":  &inputs.TagInfo{Desc: "The MySQL user name"},
			"processlist_host":  &inputs.TagInfo{Desc: "The host name of the client"},
			"processlist_db":    &inputs.TagInfo{Desc: "The default database from processlist."},
			"session_status":    &inputs.TagInfo{Desc: "Derived session status: active / idle / blocked."},
			"wait_event":        &inputs.TagInfo{Desc: "The MySQL wait event name from performance_schema (or CPU/other sentinel)."},
			//nolint:lll
			"wait_group": &inputs.TagInfo{Desc: "Datakit unified wait group: Lock, I/O, Concurrency, Memory, Network, CPU, Commit/Log, Other (derived from wait_event)."},
		},
	}
}

// mysqlAggregatedSession stores aggregated values for sessions grouped by dimensions.
type mysqlAggregatedSession struct {
	// Tags
	processlistUser string
	processlistHost string
	processlistDB   string
	sessionStatus   string
	waitEvent       string
	waitGroup       string

	// Aggregated fields
	sessionCount int64
	blockedCount int64
}

// aggregateMysqlSessions aggregates activity rows into session metrics.
// Aggregates by: db + user + host + session_status + wait_event + wait_group.
// wait_event/wait_group are normalized defensively here so callers do not rely on a specific upstream step.
func (ipt *Input) aggregateMysqlSessions(rows activityRowSlice) []*mysqlAggregatedSession {
	if len(rows) == 0 {
		return nil
	}

	merged := make(map[string]*mysqlAggregatedSession)

	for _, row := range rows {
		user := row.ProcesslistUser.String
		host := row.ProcesslistHost.String
		db := row.ProcesslistDB.String
		sessionStatus := getMySQLSessionStatus(row)
		waitEvent := row.WaitEvent.String
		waitGrp := row.WaitGroup

		key := strings.Join([]string{db, user, host, sessionStatus, waitEvent, waitGrp}, "|")

		if _, ok := merged[key]; !ok {
			merged[key] = &mysqlAggregatedSession{
				processlistUser: user,
				processlistHost: host,
				processlistDB:   db,
				sessionStatus:   sessionStatus,
				waitEvent:       waitEvent,
				waitGroup:       waitGrp,
			}
		}
		agg := merged[key]

		agg.sessionCount++
		if sessionStatus == "blocked" {
			agg.blockedCount++
		}
	}

	out := make([]*mysqlAggregatedSession, 0, len(merged))
	for _, v := range merged {
		out = append(out, v)
	}
	return out
}

// buildMysqlSessionPoints builds metric points from aggregated session rows.
func (ipt *Input) buildMysqlSessionPoints(rows []*mysqlAggregatedSession, ptsTime time.Time) []*point.Point {
	if len(rows) == 0 {
		return nil
	}

	var pts []*point.Point
	opts := append(point.DefaultMetricOptions(), point.WithTime(ptsTime))

	for _, row := range rows {
		kvs := ipt.getKVs()

		// Tags
		kvs = kvs.AddTag("processlist_user", row.processlistUser)
		kvs = kvs.AddTag("processlist_host", row.processlistHost)
		kvs = kvs.AddTag("processlist_db", row.processlistDB)
		kvs = kvs.AddTag("session_status", row.sessionStatus)
		kvs = kvs.AddTag("wait_event", row.waitEvent)
		kvs = kvs.AddTag("wait_group", row.waitGroup)

		// Fields
		kvs = kvs.Set("session_group_count", row.sessionCount)
		kvs = kvs.Set("session_blocked_count", row.blockedCount)

		pts = append(pts, point.NewPoint(metricNameMySQLDbmSession, kvs, opts...))
	}

	return pts
}

func getMySQLSessionStatus(row activityRow) string {
	if row.BlockingThreadID.Valid && strings.TrimSpace(row.BlockingThreadID.String) != "" {
		return "blocked"
	}

	command := strings.ToLower(strings.TrimSpace(row.ProcesslistCommand.String))
	state := strings.ToLower(strings.TrimSpace(row.ProcesslistState.String))

	if command == "sleep" || strings.HasPrefix(state, "idle") || strings.Contains(state, "sleep") {
		return "idle"
	}

	return "active"
}

// metricCollectMysqlDbmSession aggregates activity rows into session metrics and returns metric points.
func (ipt *Input) metricCollectMysqlDbmSession(rows activityRowSlice, ptsTime time.Time) []*point.Point {
	aggregated := ipt.aggregateMysqlSessions(rows)
	if len(aggregated) == 0 {
		return nil
	}
	return ipt.buildMysqlSessionPoints(aggregated, ptsTime)
}
