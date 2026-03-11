// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package sqlserver

import (
	"context"
	"fmt"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/metrics"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

const (
	metricNameSQLServerDbmConnection = "sqlserver_dbm_connection"
)

type dbmConnectionRow struct {
	userName        string
	databaseName    string
	status          string
	connectionCount int64
}

// collectDbmConnections collects and feeds connection metrics.
func (ipt *Input) collectDbmConnections(ctx context.Context, ptsTime time.Time) error {
	if ipt.Dbm == nil || ipt.Dbm.Activity == nil || !ipt.Dbm.Activity.Enabled {
		return nil
	}

	start := time.Now()

	var query string
	if ipt.MajorVersion <= 2008 {
		query = dbmConnectionQuery08
	} else {
		query = dbmConnectionQuery
	}

	// Query connection counts
	queryStart := time.Now()
	rows, err := ipt.db.QueryContext(ctx, query)
	dbmSQLQueryDuration.WithLabelValues("connection", "query").Observe(time.Since(queryStart).Seconds())
	if err != nil {
		return fmt.Errorf("failed to query connection metrics: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			l.Errorf("failed to close rows: %v", err)
		}
	}()

	columns, err := rows.Columns()
	if err != nil {
		return fmt.Errorf("failed to get columns: %w", err)
	}

	var connectionRows []*dbmConnectionRow
	for rows.Next() {
		columnMap, err := GetColumnMap(rows, columns)
		if err != nil {
			l.Errorf("failed to get column map: %v", err)
			continue
		}

		row := &dbmConnectionRow{}
		row.userName = getStringField(columnMap, "user_name")
		row.databaseName = getStringField(columnMap, "database_name")
		row.status = getStringField(columnMap, "status")
		row.connectionCount = getInt64Field(columnMap, "connections")

		if row.connectionCount > 0 {
			connectionRows = append(connectionRows, row)
		}
	}
	if err := rows.Err(); err != nil {
		return err
	}

	if len(connectionRows) == 0 {
		return nil
	}

	// Build and feed points
	pts := ipt.buildConnectionPoints(connectionRows, ptsTime)
	if len(pts) > 0 {
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
			l.Errorf("feed dbm connection metrics failed: %s", err.Error())
			return err
		}
	}

	return nil
}

// buildConnectionPoints builds point.Point from connection rows.
func (ipt *Input) buildConnectionPoints(rows []*dbmConnectionRow, ptsTime time.Time) []*point.Point {
	if len(rows) == 0 {
		return nil
	}

	var pts []*point.Point
	opts := append(point.DefaultMetricOptions(), point.WithTime(ptsTime))

	for _, row := range rows {
		kvs := ipt.getKVs()

		// Tags
		kvs = kvs.AddTag("server", ipt.Object.name)
		kvs = kvs.AddTag("database_name", row.databaseName)
		kvs = kvs.AddTag("user_name", row.userName)
		kvs = kvs.AddTag("connection_status", row.status)

		// Fields
		kvs = kvs.Set("connection_count", row.connectionCount)

		pts = append(pts, point.NewPoint(metricNameSQLServerDbmConnection, kvs, opts...))
	}

	return pts
}
