// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package oracle

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/metrics"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

const (
	metricNameOracleDbmConnection = "oracle_dbm_connection"
)

type dbmConnectionRowDB struct {
	UserName        sql.NullString `db:"USER_NAME"`
	Status          string         `db:"STATUS"`
	PdbName         sql.NullString `db:"PDB_NAME"`
	ConnectionCount int64          `db:"CONNECTION_COUNT"`
}

// collectDbmConnections collects and feeds connection metrics.
func (ipt *Input) collectDbmConnections(ctx context.Context, ptsTime time.Time) error {
	if ipt.Dbm == nil || ipt.Dbm.Activity == nil || !ipt.Dbm.Activity.Enabled {
		return nil
	}

	start := time.Now()

	// Build connection query based on Oracle version
	var query string
	if isDBVersionGreaterOrEqualThan(ipt.dbVersion, "12") {
		query = connectionQuery12
	} else {
		query = connectionQuery11
	}

	// Query connection counts
	queryStart := time.Now()
	var connectionRowsDB []dbmConnectionRowDB
	err := selectWrapperWithBinds(ipt, ctx, &connectionRowsDB, query)
	dbmSQLQueryDuration.WithLabelValues("connection", "query").Observe(time.Since(queryStart).Seconds())
	if err != nil {
		return fmt.Errorf("failed to query connection metrics: %w", err)
	}

	// Build and feed points
	pts := ipt.buildConnectionPoints(connectionRowsDB, ptsTime)
	if len(pts) > 0 {
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
			l.Errorf("feed dbm connection metrics failed: %s", err.Error())
			return err
		}
	}

	return nil
}

// buildConnectionPoints builds point.Point from connection rows.
func (ipt *Input) buildConnectionPoints(rows []dbmConnectionRowDB, ptsTime time.Time) []*point.Point {
	if len(rows) == 0 {
		return nil
	}

	var pts []*point.Point
	opts := append(point.DefaultMetricOptions(), point.WithTime(ptsTime))

	for _, row := range rows {
		if row.ConnectionCount <= 0 {
			continue
		}

		kvs := ipt.getKVs()

		// Tags
		kvs = kvs.AddTag("server", ipt.Object.name)
		kvs = kvs.AddTag("cdb_name", ipt.cdbName)
		kvs = kvs.AddTag("pdb_name", ipt.getPdbName(row.PdbName.String))
		if row.UserName.Valid {
			kvs = kvs.AddTag("username", row.UserName.String)
		}
		kvs = kvs.AddTag("connection_status", row.Status)

		// Fields
		kvs = kvs.Set("connection_count", row.ConnectionCount)

		pts = append(pts, point.NewPoint(metricNameOracleDbmConnection, kvs, opts...))
	}

	return pts
}
