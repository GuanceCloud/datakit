// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package sqlserver

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
	metricNameSQLServerDbmSession = "sqlserver_dbm_session"
	waitTypeCPU                   = "CPU"
	waitTypeWaitingOnCPU          = "WAITING_ON_CPU"
)

// aggregatedSession stores aggregated values for sessions grouped by dimensions.
type aggregatedSession struct {
	// Aggregation key (for grouping)
	aggKey string

	// Tags
	server        string
	sqlserverHost string
	databaseName  string
	userName      string
	programName   string
	clientAddress string
	status        string // active/idle/blocked
	waitCategory  string // CPU/IO/Lock/Network/Other
	waitType      string

	// Aggregated fields
	sessionCount              int64
	totalCPUTime              int64
	totalWaitTime             int64
	totalElapsedTime          int64
	blockedCount              int64
	blockingCount             int64
	totalReads                int64
	totalWrites               int64
	totalLogicalReads         int64
	totalOpenTransactionCount int64
	maxElapsedTime            int64
}

// aggregateSessions aggregates activity rows into session metrics.
// Aggregates by: database_name + user_name + program_name + client_address + session_status + wait_group + wait_type
// Center platform can sum these metrics by any subset of tags as needed.
func (ipt *Input) aggregateSessions(activityRows []*dbmActivityRow) []*aggregatedSession {
	if len(activityRows) == 0 {
		return nil
	}

	// First pass: collect all blocking_session_id to determine blocking_count
	// Note: blocking_session_id = 0 means "not blocked" in SQL Server, so we exclude 0
	blockingSessionIDs := make(map[int64]struct{})
	for _, row := range activityRows {
		if row.blockingSessionID > 0 {
			blockingSessionIDs[row.blockingSessionID] = struct{}{}
		}
	}

	// Second pass: aggregate by database + status + wait_class
	mergedMap := make(map[string]*aggregatedSession)

	for _, row := range activityRows {
		isActive := row.requestStatus != ""
		// Note: blocking_session_id = 0 means "not blocked" in SQL Server, so we exclude 0
		isBlocked := row.blockingSessionID > 0

		status := "idle"
		if isBlocked {
			status = "blocked" // Idle session that is being blocked
		} else if isActive {
			status = "active" // Active session (has request), even if blocked
		}

		// Use wait_class from activity row (already calculated)
		waitCategory := row.waitCategory

		// Calculate is_blocker
		isBlocker := false
		if _, ok := blockingSessionIDs[row.sessionID]; ok {
			isBlocker = true
		}

		// Generate aggregation key
		aggKey := fmt.Sprintf("%s|%s|%s|%s|%s|%s|%s", row.databaseName, row.userName, row.programName,
			row.clientAddress, status, waitCategory, row.waitType)

		if _, ok := mergedMap[aggKey]; !ok {
			mergedMap[aggKey] = &aggregatedSession{
				aggKey:        aggKey,
				server:        ipt.Object.name,
				sqlserverHost: row.sqlserverHost,
				databaseName:  row.databaseName,
				userName:      row.userName,
				programName:   row.programName,
				clientAddress: row.clientAddress,
				status:        status,
				waitCategory:  waitCategory,
				waitType:      row.waitType,
			}
		}
		agg := mergedMap[aggKey]

		// Aggregate values
		agg.sessionCount++

		// Only aggregate request-level fields for active sessions (have requestStatus)
		// Idle blocking sessions don't have request-level data, so these fields remain 0
		if row.requestStatus != "" {
			// Active session: aggregate all request-level fields
			agg.totalCPUTime += row.cpuTime
			agg.totalWaitTime += row.waitTime
			agg.totalElapsedTime += row.totalElapsedTime
			agg.totalReads += row.reads
			agg.totalWrites += row.writes
			agg.totalLogicalReads += row.logicalReads
			agg.totalOpenTransactionCount += row.openTransactionCount
			if row.totalElapsedTime > agg.maxElapsedTime {
				agg.maxElapsedTime = row.totalElapsedTime
			}
		}
		// Session-level fields (available for both active and idle sessions)
		// Note: blocking_session_id = 0 means "not blocked" in SQL Server, so we exclude 0
		agg.blockedCount += boolToInt64(row.blockingSessionID > 0)
		agg.blockingCount += boolToInt64(isBlocker)
	}

	// Convert map to slice
	result := make([]*aggregatedSession, 0, len(mergedMap))
	for _, agg := range mergedMap {
		result = append(result, agg)
	}

	return result
}

// normalizeWaitType synthesizes CPU-related wait labels so wait-event
// dimensions remain visible even when SQL Server reports an empty wait_type.
func normalizeWaitType(sessionStatus, requestStatus, waitType string) string {
	wt := strings.TrimSpace(waitType)
	if wt != "" {
		return wt
	}

	rs := strings.ToLower(strings.TrimSpace(requestStatus))
	switch rs {
	case "runnable":
		return waitTypeWaitingOnCPU
	case "running":
		return waitTypeCPU
	}

	if rs == "" && strings.ToLower(strings.TrimSpace(sessionStatus)) == "running" {
		return waitTypeCPU
	}

	return ""
}

// categorizeWaitType maps a normalized SQL Server wait type to a unified category.
func categorizeWaitType(waitType string) string {
	wt := strings.ToUpper(strings.TrimSpace(waitType))

	if wt == waitTypeCPU || wt == waitTypeWaitingOnCPU {
		return "CPU"
	}

	switch {
	case wt == "SOS_SCHEDULER_YIELD":
		return "CPU"
	case strings.HasPrefix(wt, "LCK_"):
		return "Lock"
	case wt == "RESOURCE_SEMAPHORE_QUERY_COMPILE" || strings.HasPrefix(wt, "LATCH_") || strings.HasPrefix(wt, "PAGELATCH_"):
		return "Concurrency"
	case strings.Contains(wt, "RESOURCE_SEMAPHORE"):
		return "Memory"
	case strings.Contains(wt, "NETWORK"):
		return "Network"
	case wt == "WRITELOG" || wt == "LOGBUFFER":
		return "Commit/Log"
	case strings.HasPrefix(wt, "PAGEIOLATCH_") || wt == "IO_COMPLETION":
		return "I/O"
	default:
		return "Other"
	}
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
		kvs = kvs.AddTag("sqlserver_host", row.sqlserverHost)
		kvs = kvs.AddTag("database_name", row.databaseName)
		kvs = kvs.AddTag("user_name", row.userName)
		if row.programName != "" {
			kvs = kvs.AddTag("program_name", row.programName)
		}
		if row.clientAddress != "" {
			kvs = kvs.AddTag("client_address", row.clientAddress)
		}
		kvs = kvs.AddTag("session_status", row.status)
		if row.waitType != "" {
			kvs = kvs.AddTag("wait_type", row.waitType)
		}
		kvs = kvs.AddTag("wait_group", row.waitCategory)

		// Fields
		kvs = kvs.Set("session_group_count", row.sessionCount)
		kvs = kvs.Set("session_total_cpu_time", row.totalCPUTime)
		kvs = kvs.Set("session_total_wait_time", row.totalWaitTime)
		kvs = kvs.Set("session_total_elapsed_time", row.totalElapsedTime)
		kvs = kvs.Set("session_blocked_count", row.blockedCount)
		kvs = kvs.Set("session_blocking_count", row.blockingCount)
		kvs = kvs.Set("session_total_reads", row.totalReads)
		kvs = kvs.Set("session_total_writes", row.totalWrites)
		kvs = kvs.Set("session_total_logical_reads", row.totalLogicalReads)
		kvs = kvs.Set("session_total_open_transaction_count", row.totalOpenTransactionCount)
		kvs = kvs.Set("session_max_elapsed_time", row.maxElapsedTime)

		pts = append(pts, point.NewPoint(metricNameSQLServerDbmSession, kvs, opts...))
	}

	return pts
}

// collectDbmSessionMetrics collects and feeds session metrics from activity data.
func (ipt *Input) collectDbmSessionMetrics(activityRows []*dbmActivityRow, ptsTime time.Time) {
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
		dkio.WithMeasurement(inputs.GetOverrideMeasurement(ipt.MeasurementVersion, measurementSQLServer)),
	); err != nil {
		ipt.feeder.FeedLastError(err.Error(),
			metrics.WithLastErrorInput(inputName),
			metrics.WithLastErrorCategory(point.Metric),
		)
		l.Errorf("feed session metrics failed: %s", err.Error())
	}
}
