// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package mysql

import (
	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

type dbmConnectionMeasurement struct {
	name string
	tags map[string]string
}

func (m *dbmConnectionMeasurement) Point() *point.Point {
	opts := point.DefaultMetricOptions()
	return point.NewPoint(m.name,
		point.NewTags(m.tags),
		opts...)
}

// Info describes the mysql_dbm_connection metrics: connection_count grouped by user/host/db/state.
func (m *dbmConnectionMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: metricNameMySQLDbmConnection,
		Desc: "MySQL DBM connection metrics, aggregated by user/host/db/state",
		Cat:  point.Metric,
		Fields: map[string]interface{}{
			"connection_count": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of connections for this (user, host, db, state) group",
			},
		},
		Tags: map[string]interface{}{
			"host":              &inputs.TagInfo{Desc: "The server host address"},
			"server":            &inputs.TagInfo{Desc: "The address of the server. The value is `host:port`"},
			"database_instance": &inputs.TagInfo{Desc: "MySQL instance identifier from configured tag or @@server_uuid."},
			"processlist_user":  &inputs.TagInfo{Desc: "The MySQL user name"},
			"processlist_host":  &inputs.TagInfo{Desc: "The host name of the client"},
			"processlist_db":    &inputs.TagInfo{Desc: "The default database from processlist."},
			"processlist_state": &inputs.TagInfo{Desc: "The MySQL processlist state."},
		},
	}
}

// metricCollectMysqlDbmConnections builds connection metric points from already-fetched connection rows,
// so the caller can query once and pass the same data to both activity and connection collection.
func (ipt *Input) metricCollectMysqlDbmConnections(connections []connectionRow) []*point.Point {
	var pts []*point.Point
	opts := ipt.getKVsOpts(point.Metric)

	for _, connection := range connections {
		kvs := ipt.getKVs()

		kvs = kvs.AddTag("processlist_user", connection.processlistUser.String)
		kvs = kvs.AddTag("processlist_host", connection.processlistHost.String)
		kvs = kvs.AddTag("processlist_db", connection.processlistDB.String)
		kvs = kvs.AddTag("processlist_state", connection.processlistState.String)
		kvs = kvs.Set("connection_count", connection.connections.Int64)

		pts = append(pts, point.NewPoint(metricNameMySQLDbmConnection, kvs, opts...))
	}

	return pts
}
