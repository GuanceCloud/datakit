// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package postgresql

import (
	"context"
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
	metricNamePostgreSQLDbmConnection = "postgresql_dbm_connection"
)

const sqlGetDbmActiveConnections = `
	SELECT
		application_name,
		state,
		usename,
		datname,
		count(*) AS connections
	FROM pg_stat_activity
	WHERE pid != pg_backend_pid()
		AND client_port IS NOT NULL %s
	GROUP BY application_name, state, usename, datname
`

type dbmConnectionRow struct {
	applicationName string
	state           string
	usename         string
	datname         string
	connectionCount int64
}

type dbmConnectionMeasurement struct{}

func (*dbmConnectionMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "postgresql_dbm_connection",
		Desc: "PostgreSQL DBM active connection metrics grouped by application_name, state, usename and db.",
		Cat:  point.Metric,
		Fields: map[string]interface{}{
			"connection_count": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of active user connections grouped by application, state, user and database.",
			},
		},
		Tags: map[string]interface{}{
			"server":           inputs.NewTagInfo("The server address"),
			"application_name": inputs.NewTagInfo("Name of the application connected to this backend"),
			"state":            inputs.NewTagInfo("Current overall state of this backend"),
			"usename":          inputs.NewTagInfo("Name of the user logged into this backend"),
			"db":               inputs.NewTagInfo("The database name"),
		},
	}
}

func (ipt *Input) collectDbmConnections(ptsTime time.Time) error {
	if !ipt.Dbm || !ipt.DbmActivity.Enabled {
		return nil
	}

	start := time.Now()
	filters := ""
	if len(ipt.IgnoredDatabases) > 0 {
		filters += fmt.Sprintf(" AND datname NOT IN ('%s')", strings.Join(ipt.IgnoredDatabases, "','"))
	} else if len(ipt.Databases) > 0 {
		filters += fmt.Sprintf(" AND datname IN ('%s')", strings.Join(ipt.Databases, "','"))
	}

	query := fmt.Sprintf(sqlGetDbmActiveConnections, filters)
	ctx, cancel := context.WithTimeout(context.Background(), ipt.Timeout.Duration)
	defer cancel()

	rows, err := ipt.service.Query(ctx, query)
	if err != nil {
		return fmt.Errorf("query dbm active connections failed: %w", err)
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return fmt.Errorf("get dbm active connection columns failed: %w", err)
	}

	connRows := make([]dbmConnectionRow, 0, 64)
	for rows.Next() {
		columnMap, err := ipt.service.GetColumnMap(rows, columns)
		if err != nil {
			l.Errorf("scan dbm active connection row failed: %s", err.Error())
			continue
		}

		row := dbmConnectionRow{}
		for col, v := range columnMap {
			if v == nil || *v == nil {
				continue
			}

			switch col {
			case "application_name":
				row.applicationName = interfaceToString(*v)
			case "state":
				row.state = interfaceToString(*v)
			case "usename":
				row.usename = interfaceToString(*v)
			case "datname":
				row.datname = interfaceToString(*v)
			case "connections":
				row.connectionCount = cast.ToInt64(*v)
			}
		}

		if row.connectionCount > 0 {
			connRows = append(connRows, row)
		}
	}

	if len(connRows) == 0 {
		return nil
	}

	pts := ipt.buildDbmConnectionPoints(connRows, ptsTime)
	if len(pts) == 0 {
		return nil
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
		l.Errorf("feed dbm connection metrics failed: %s", err.Error())
		return err
	}

	return nil
}

func (ipt *Input) buildDbmConnectionPoints(rows []dbmConnectionRow, ptsTime time.Time) []*point.Point {
	if len(rows) == 0 {
		return nil
	}

	opts := append(point.DefaultMetricOptions(), point.WithTime(ptsTime))
	pts := make([]*point.Point, 0, len(rows))
	for _, row := range rows {
		kvs := ipt.getKVs()
		kvs = kvs.AddTag("application_name", row.applicationName)
		kvs = kvs.AddTag("state", row.state)
		kvs = kvs.AddTag("usename", row.usename)
		kvs = kvs.AddTag("db", row.datname)

		kvs = kvs.Set("connection_count", row.connectionCount)

		pts = append(pts, point.NewPoint(metricNamePostgreSQLDbmConnection, kvs, opts...))
	}

	return pts
}
