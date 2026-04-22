// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package oracle

import (
	"fmt"
	"strings"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/metrics"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

const (
	metricNameOracleDbmSession = "oracle_dbm_session"
)

// aggregatedSession stores aggregated values for sessions grouped by dimensions.
type aggregatedSession struct {
	// Aggregation key (for grouping)
	aggKey string

	// Tags
	server    string
	pdbName   string
	username  string
	program   string
	client    string
	status    string // active/idle/blocked
	waitClass string // Wait event class from Oracle
	waitEvent string
	waitGroup string

	// Aggregated fields
	sessionCount  int64
	totalWaitTime int64
	blockedCount  int64
	blockingCount int64
}

// aggregateSessions aggregates activity rows into session metrics.
func (ipt *Input) aggregateSessions(activityRows []*OracleActivityRow) []*aggregatedSession {
	if len(activityRows) == 0 {
		return nil
	}

	// First pass: collect all blocking session IDs to determine blocking_count
	// Note: blocking_session = 0 means "not blocked" in Oracle, so we exclude 0
	blockingSessionIDs := make(map[uint64]struct{})
	for _, row := range activityRows {
		if row.FinalBlockingSession > 0 {
			blockingSessionIDs[row.FinalBlockingSession] = struct{}{}
		} else if row.BlockingSession > 0 {
			blockingSessionIDs[row.BlockingSession] = struct{}{}
		}
	}

	// Second pass: aggregate by pdb_name + username + program + client + status + wait_group + wait_class + wait_event
	mergedMap := make(map[string]*aggregatedSession)

	for _, row := range activityRows {
		// Determine status: active/idle/blocked
		status := "idle"
		isBlocked := row.FinalBlockingSession > 0 || row.BlockingSession > 0
		if isBlocked {
			status = "blocked" // Session that is being blocked
		} else if strings.ToUpper(row.Status) == "ACTIVE" {
			status = "active" // Active session
		}

		// Use wait_class from activity row
		waitClass := row.WaitEventClass
		waitEvent := row.WaitEvent
		waitGroup := mapOracleWaitClassToGroup(waitClass)

		// Calculate is_blocker
		isBlocker := false
		if _, ok := blockingSessionIDs[row.SessionID]; ok {
			isBlocker = true
		}

		// Generate aggregation key
		aggKey := fmt.Sprintf("%s|%s|%s|%s|%s|%s|%s|%s", row.PdbName, row.User, row.Program, row.Client, status, waitGroup, waitClass, waitEvent)

		if _, ok := mergedMap[aggKey]; !ok {
			mergedMap[aggKey] = &aggregatedSession{
				aggKey:    aggKey,
				server:    ipt.Object.name,
				pdbName:   row.PdbName,
				username:  row.User,
				program:   row.Program,
				client:    row.Client,
				status:    status,
				waitClass: waitClass,
				waitEvent: waitEvent,
				waitGroup: waitGroup,
			}
		}
		agg := mergedMap[aggKey]

		// Aggregate values
		agg.sessionCount++

		// Aggregate wait time (convert from microseconds to milliseconds)
		if row.WaitTimeMicro > 0 {
			agg.totalWaitTime += int64(row.WaitTimeMicro / 1000)
		}

		// Session-level fields (available for both active and idle sessions)
		// Note: blocking_session = 0 means "not blocked" in Oracle, so we exclude 0
		agg.blockedCount += boolToInt64(isBlocked)
		agg.blockingCount += boolToInt64(isBlocker)
	}

	// Convert map to slice
	result := make([]*aggregatedSession, 0, len(mergedMap))
	for _, agg := range mergedMap {
		result = append(result, agg)
	}

	return result
}

// boolToInt64 converts boolean to int64 (1 for true, 0 for false).
func boolToInt64(b bool) int64 {
	if b {
		return 1
	}
	return 0
}

// buildSessionPoints builds point.Point from aggregated session rows.
func (ipt *Input) buildSessionPoints(rows []*aggregatedSession, ptsTime time.Time) []*point.Point {
	if len(rows) == 0 {
		return nil
	}

	var pts []*point.Point
	opts := append(point.DefaultMetricOptions(), point.WithTime(ptsTime))

	for _, row := range rows {
		kvs := ipt.getKVs()

		// Tags
		kvs = kvs.AddTag("server", row.server)
		if ipt.cdbName != "" {
			kvs = kvs.AddTag("cdb_name", ipt.cdbName)
		}
		if row.pdbName != "" {
			kvs = kvs.AddTag("pdb_name", row.pdbName)
		}
		if row.username != "" {
			kvs = kvs.AddTag("username", row.username)
		}
		if row.program != "" {
			kvs = kvs.AddTag("program", row.program)
		}
		if row.client != "" {
			kvs = kvs.AddTag("client", row.client)
		}
		kvs = kvs.AddTag("session_status", row.status)
		if row.waitClass != "" {
			kvs = kvs.AddTag("wait_class", row.waitClass)
		}
		if row.waitGroup != "" {
			kvs = kvs.AddTag("wait_group", row.waitGroup)
		}
		if row.waitEvent != "" {
			kvs = kvs.AddTag("event", row.waitEvent)
		}

		// Fields
		kvs = kvs.Set("session_group_count", row.sessionCount)
		kvs = kvs.Set("session_total_wait_time", row.totalWaitTime)
		kvs = kvs.Set("session_blocked_count", row.blockedCount)
		kvs = kvs.Set("session_blocking_count", row.blockingCount)

		pts = append(pts, point.NewPoint(metricNameOracleDbmSession, kvs, opts...))
	}

	return pts
}

// collectDbmSessionMetrics collects and feeds session metrics from activity data.
func (ipt *Input) collectDbmSessionMetrics(activityRows []*OracleActivityRow, ptsTime time.Time) {
	if len(activityRows) == 0 {
		return
	}

	start := time.Now()

	// Aggregate sessions at multiple granularities
	aggregatedSessions := ipt.aggregateSessions(activityRows)
	if len(aggregatedSessions) == 0 {
		return
	}

	// Build points
	pts := ipt.buildSessionPoints(aggregatedSessions, ptsTime)
	if len(pts) == 0 {
		return
	}

	// Feed points
	if err := ipt.feeder.Feed(point.Metric, pts,
		dkio.WithCollectCost(time.Since(start)),
		dkio.WithElection(ipt.Election),
		dkio.WithSource(dbmFeedName),
		dkio.WithMeasurement(inputs.GetOverrideMeasurement(ipt.MeasurementVersion, measurementOracle)),
	); err != nil {
		ipt.feeder.FeedLastError(err.Error(),
			metrics.WithLastErrorInput(inputName),
			metrics.WithLastErrorCategory(point.Metric),
		)
		l.Errorf("feed session metrics failed: %s", err.Error())
	}
}
