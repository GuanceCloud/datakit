// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package oracle

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

const (
	oracleType                  = "Oracle"
	oracleObjectMeasurementName = "database"
)

type objectMertric struct {
	Queries     int64
	SlowQueries int64
	QPS         float64
	TPS         float64
	TranCommits float64
	TranRolls   float64
	Time        time.Time
}

type oracleObjectMeasurement struct{}

//nolint:lll
func (*oracleObjectMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: oracleObjectMeasurementName,
		Cat:  point.Object,
		Desc: "Oracle object metrics([:octicons-tag-24: Version-1.77.0](../datakit/changelog-2025.md#cl-1.77.0))",
		Tags: map[string]interface{}{
			"host":          &inputs.TagInfo{Desc: "The hostname of the Oracle server"},
			"server":        &inputs.TagInfo{Desc: "The address of the server. The value is `host:port`"},
			"version":       &inputs.TagInfo{Desc: "The version of the Oracle server"},
			"name":          &inputs.TagInfo{Desc: "The name of the database. The value is `host:port` in default"},
			"database_type": &inputs.TagInfo{Desc: "The type of the database. The value is `Oracle`"},
			"port":          &inputs.TagInfo{Desc: "The port of the Oracle server"},
		},
		Fields: map[string]interface{}{
			"uptime":         &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.DurationSecond, Desc: "The number of seconds that the server has been up"},
			"slow_query_log": &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "Whether the slow query log is enabled according to whether `slow_query_time` is greater than 0 . The value can be OFF to disable the log or ON to enable the log."},
			"slow_queries":   &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "The number of queries that have taken more than `slow_query_time`."},
			"avg_query_time": &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.TimestampUS, Desc: "The average time taken by a query to execute"},
			"qps":            &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.Gauge, Desc: "The number of queries executed by the database per second"},
			"tps":            &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.Gauge, Desc: "The number of transactions executed by the database per second"},
		},
	}
}

func (ipt *Input) collectDatabaseObject() {
	if !ipt.Object.lastCollectionTime.IsZero() &&
		ipt.Object.lastCollectionTime.Add(ipt.Object.Interval.Duration).After(time.Now()) {
		l.Debugf("skip oracle_object collection, time interval not reached")
	}
	start := time.Now()

	ipt.Object.lastCollectionTime = time.Now()

	kvs := ipt.getKVs()
	opts := ipt.getKVsOpts(point.Object)

	opts = append(opts, point.WithTimestamp(ipt.ptsTime.UnixNano()))

	kvs = kvs.AddTag("version", ipt.mainVersion).
		AddTag("database_type", oracleType).
		AddTag("name", ipt.Object.name).
		AddTag("host", ipt.Host).
		AddTag("port", fmt.Sprintf("%d", ipt.Port)).
		AddV2("uptime", ipt.Uptime, false)

	if ipt.objectMetric != nil {
		ipt.objectMetric.TPS = ipt.objectMetric.TranCommits + ipt.objectMetric.TranRolls
		kvs = kvs.AddV2("qps", ipt.objectMetric.QPS, false).
			AddV2("tps", ipt.objectMetric.TPS, false)
	}

	slowQueryLog := "OFF"
	if ipt.slowQueryTime > 0 {
		slowQueryLog = "ON"
		kvs = kvs.AddV2("slow_queries", ipt.objectMetric.SlowQueries, false)
	}
	kvs = kvs.AddV2("slow_query_log", slowQueryLog, false)

	if avgQueryTime, err := ipt.getAvgQueryTime(); err != nil {
		l.Warnf("failed to get avg query time: %s", err)
	} else {
		kvs = kvs.AddV2("avg_query_time", avgQueryTime, false)
	}

	pts := []*point.Point{point.NewPointV2("database", kvs, opts...)}

	if err := ipt.feeder.Feed(point.Object,
		pts,
		dkio.WithCollectCost(time.Since(start)),
		dkio.WithElection(ipt.Election),
		dkio.WithSource(objectFeedName)); err != nil {
		l.Warnf("feeder.Feed: %s, ignored", err)
	}
}

const sqlGetQueryTime = `
select 
	sum(elapsed_time)/sum(executions) as AVG_QUERY_TIME
from v$sqlarea
where executions > 0
`

type avgQueryTimeRow struct {
	AvgQueryTime sql.NullFloat64 `db:"AVG_QUERY_TIME"`
}

func (ipt *Input) getAvgQueryTime() (float64, error) {
	rows := []avgQueryTimeRow{}
	if err := selectWrapper(ipt, &rows, sqlGetQueryTime, getMetricName(oracleObjectMeasurementName, "avg_query_time")); err != nil {
		l.Warnf("failed to get avg query time: %s, oracle version %s", err, ipt.mainVersion)
		return 0, fmt.Errorf("failed to query: %w", err)
	}

	if len(rows) == 0 {
		return 0, fmt.Errorf("query average query execution time failed: empty rows")
	}

	return rows[0].AvgQueryTime.Float64, nil
}
