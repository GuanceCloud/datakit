// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package postgresql

import (
	"fmt"
	"strings"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"github.com/spf13/cast"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/metrics"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

const (
	metricNamePostgreSQLDbmSession = "postgresql_dbm_session"
	waitGroupLock                  = "Lock"
	waitGroupIO                    = "I/O"
	waitGroupConcurrency           = "Concurrency"
	waitGroupNetwork               = "Network"
	waitGroupCPU                   = "CPU"
	waitGroupCommitLog             = "Commit/Log"
	waitGroupOther                 = "Other"
)

type dbmSessionMeasurement struct{}

func (*dbmSessionMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: metricNamePostgreSQLDbmSession,
		Desc: "PostgreSQL DBM session metrics aggregated from pg_stat_activity by db/usename/session_status/wait_event dimensions.",
		Cat:  point.Metric,
		Fields: map[string]interface{}{
			"session_group_count": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of sessions in this dimension group.",
			},
			"session_blocked_count": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of blocked sessions in this dimension group.",
			},
		},
		Tags: map[string]interface{}{
			"server":           inputs.NewTagInfo("The server address"),
			"db":               inputs.NewTagInfo("The database name."),
			"usename":          inputs.NewTagInfo("Name of the user logged into this backend."),
			"application_name": inputs.NewTagInfo("Name of the application connected to this backend."),
			"client_addr":      inputs.NewTagInfo("IP address of the client connected to this backend."),
			"session_status":   inputs.NewTagInfo("Derived session status: active / idle / blocked."),
			"wait_event_type":  inputs.NewTagInfo("Type of event for which the backend is waiting."),
			"wait_event":       inputs.NewTagInfo("Wait event name if backend is currently waiting."),
			"wait_group":       inputs.NewTagInfo("Datakit unified wait group: Lock, I/O, Concurrency, Network, CPU, Commit/Log, Other."),
		},
	}
}

type aggregatedPgSession struct {
	datname         string
	usename         string
	applicationName string
	clientAddr      string
	sessionStatus   string
	waitEventType   string
	waitEvent       string
	waitGroup       string

	sessionCount        int64
	sessionBlockedCount int64
}

type pgSessionSnapshot struct {
	datname         string
	usename         string
	applicationName string
	clientAddr      string
	state           string
	waitEventType   string
	waitEvent       string
	waitGroup       string
	isBlocked       bool
}

func (ipt *Input) collectDbmSessionMetrics(activityRows []map[string]any, ptsTime time.Time) {
	if !ipt.Dbm || !ipt.DbmActivity.Enabled || len(activityRows) == 0 {
		return
	}

	start := time.Now()
	aggregated := aggregatePgSessions(activityRows)
	if len(aggregated) == 0 {
		return
	}

	pts := ipt.buildDbmSessionPoints(aggregated, ptsTime)
	if len(pts) == 0 {
		return
	}

	if err := ipt.feeder.Feed(point.Metric, pts,
		dkio.WithCollectCost(time.Since(start)),
		dkio.WithElection(ipt.Election),
		dkio.WithSource(dbmFeedName),
		dkio.WithMeasurement(inputs.GetOverrideMeasurement(ipt.MeasurementVersion, measurementPostgreSQL)),
	); err != nil {
		ipt.feeder.FeedLastError(err.Error(),
			metrics.WithLastErrorInput(inputName),
			metrics.WithLastErrorCategory(point.Metric),
		)
		l.Errorf("feed dbm session metrics failed: %s", err.Error())
	}
}

func aggregatePgSessions(rows []map[string]any) []*aggregatedPgSession {
	if len(rows) == 0 {
		return nil
	}

	snapshots := make([]pgSessionSnapshot, 0, len(rows))
	for _, row := range rows {
		snapshots = append(snapshots, buildPgSessionSnapshot(row))
	}

	merged := make(map[string]*aggregatedPgSession)
	for _, snapshot := range snapshots {
		sessionStatus := getPGSessionStatus(snapshot.state, snapshot.isBlocked)
		waitGroup := snapshot.waitGroup
		key := fmt.Sprintf("%s|%s|%s|%s|%s|%s|%s|%s", snapshot.datname, snapshot.usename, snapshot.applicationName,
			snapshot.clientAddr, sessionStatus, snapshot.waitEventType, snapshot.waitEvent, waitGroup)
		if _, ok := merged[key]; !ok {
			merged[key] = &aggregatedPgSession{
				datname:         snapshot.datname,
				usename:         snapshot.usename,
				applicationName: snapshot.applicationName,
				clientAddr:      snapshot.clientAddr,
				sessionStatus:   sessionStatus,
				waitEventType:   snapshot.waitEventType,
				waitEvent:       snapshot.waitEvent,
				waitGroup:       waitGroup,
			}
		}
		agg := merged[key]

		agg.sessionCount++
		if sessionStatus == "blocked" {
			agg.sessionBlockedCount++
		}
	}

	out := make([]*aggregatedPgSession, 0, len(merged))
	for _, row := range merged {
		out = append(out, row)
	}

	return out
}

func buildPgSessionSnapshot(row map[string]any) pgSessionSnapshot {
	state := strings.ToLower(strings.TrimSpace(cast.ToString(row["state"])))
	waitEventType := cast.ToString(row["wait_event_type"])
	waitEvent := cast.ToString(row["wait_event"])
	waitGroup := cast.ToString(row["wait_group"])
	isBlocked := strings.TrimSpace(cast.ToString(row["blocking_pids"])) != ""

	return pgSessionSnapshot{
		datname:         cast.ToString(row["datname"]),
		usename:         cast.ToString(row["usename"]),
		applicationName: cast.ToString(row["application_name"]),
		clientAddr:      cast.ToString(row["client_addr"]),
		state:           state,
		waitEventType:   waitEventType,
		waitEvent:       waitEvent,
		waitGroup:       waitGroup,
		isBlocked:       isBlocked,
	}
}

func (ipt *Input) buildDbmSessionPoints(rows []*aggregatedPgSession, ptsTime time.Time) []*point.Point {
	if len(rows) == 0 {
		return nil
	}

	opts := append(point.DefaultMetricOptions(), point.WithTime(ptsTime))
	pts := make([]*point.Point, 0, len(rows))
	for _, row := range rows {
		kvs := ipt.getKVs()
		if row.datname != "" {
			kvs = kvs.AddTag("db", row.datname)
		}
		if row.usename != "" {
			kvs = kvs.AddTag("usename", row.usename)
		}
		if row.applicationName != "" {
			kvs = kvs.AddTag("application_name", row.applicationName)
		}
		if row.clientAddr != "" {
			kvs = kvs.AddTag("client_addr", row.clientAddr)
		}
		kvs = kvs.AddTag("session_status", row.sessionStatus)
		if row.waitEventType != "" {
			kvs = kvs.AddTag("wait_event_type", row.waitEventType)
		}
		if row.waitEvent != "" {
			kvs = kvs.AddTag("wait_event", row.waitEvent)
		}
		if row.waitGroup != "" {
			kvs = kvs.AddTag("wait_group", row.waitGroup)
		}

		kvs = kvs.Set("session_group_count", row.sessionCount)
		kvs = kvs.Set("session_blocked_count", row.sessionBlockedCount)

		pts = append(pts, point.NewPoint(metricNamePostgreSQLDbmSession, kvs, opts...))
	}

	return pts
}
