// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package postgresql

import (
	"fmt"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"github.com/hashicorp/golang-lru/v2/expirable"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/metrics"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

const (
	dbmQueryObjectCacheMaxSize  = 100000
	dbmQueryObjectCacheTTL      = 24 * time.Hour
	dbmQueryObjectMeasurementID = "db_query"
)

type dbmQueryObjectMeasurement struct{}

func (*dbmQueryObjectMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: dbmQueryObjectMeasurementID,
		Cat:  point.Object,
		//nolint:lll
		Desc: "PostgreSQL DBM query object. Each object represents a unique normalized SQL statement identified by query_signature, which is derived from db, rolname, and query text.",
		Tags: map[string]interface{}{
			"name":              inputs.NewTagInfo("Object identity built from server, optional database_instance, and query_signature."),
			"server":            inputs.NewTagInfo("The PostgreSQL server address"),
			"database_instance": inputs.NewTagInfo("PostgreSQL instance identifier from configured tag `database_instance` or system_identifier."),
			"database_type":     inputs.NewTagInfo("The type of database. The value is `PostgreSQL`"),
			"db":                inputs.NewTagInfo("The database name"),
			"rolname":           inputs.NewTagInfo("The role name"),
			"queryid":           inputs.NewTagInfo("The query ID reported by pg_stat_statements, if available."),
			"query_signature":   inputs.NewTagInfo("The hash signature computed from db, rolname, and normalized query text."),
		},
		Fields: map[string]interface{}{
			"message": &inputs.FieldInfo{
				DataType: inputs.String,
				Unit:     inputs.UnknownUnit,
				Desc:     "The normalized/obfuscated SQL text.",
			},
		},
	}
}

func (ipt *Input) collectDbmQueries(rows []dbmMetricRow, ptsTime time.Time) {
	if len(rows) == 0 {
		return
	}
	if ipt.dbmQueryObjectCache == nil {
		ipt.dbmQueryObjectCache = expirable.NewLRU[string, struct{}](
			dbmQueryObjectCacheMaxSize,
			nil,
			dbmQueryObjectCacheTTL,
		)
	}

	start := time.Now()
	opts := append(point.DefaultObjectOptions(), point.WithTime(ptsTime))
	pts := make([]*point.Point, 0, len(rows))
	for _, row := range rows {
		if row.querySignature == "" || row.message == "" {
			continue
		}

		cacheKey := row.querySignature
		if _, ok := ipt.dbmQueryObjectCache.Get(cacheKey); ok {
			continue
		}

		objectName := fmt.Sprintf("%s-%s", ipt.Object.name, row.querySignature)
		if ipt.databaseInstance != "" {
			objectName = fmt.Sprintf("%s-%s-%s", ipt.Object.name, ipt.databaseInstance, row.querySignature)
		}

		kvs := ipt.getKVs()
		kvs = kvs.AddTag("name", objectName)
		kvs = kvs.AddTag("server", ipt.Object.name)
		kvs = kvs.AddTag("database_type", "PostgreSQL")
		if row.db != "" {
			kvs = kvs.AddTag("db", row.db)
		}
		if row.rolname != "" {
			kvs = kvs.AddTag("rolname", row.rolname)
		}
		if row.queryID != "" {
			kvs = kvs.AddTag("queryid", row.queryID)
		}
		kvs = kvs.AddTag("query_signature", row.querySignature)
		kvs = kvs.Set("message", row.message)

		pts = append(pts, point.NewPoint(dbmQueryObjectMeasurementID, kvs, opts...))
		ipt.dbmQueryObjectCache.Add(cacheKey, struct{}{})
	}

	if len(pts) == 0 {
		return
	}

	if err := ipt.feeder.Feed(point.Object, pts,
		dkio.WithCollectCost(time.Since(start)),
		dkio.WithElection(ipt.Election),
		dkio.WithSource(dbmFeedName),
	); err != nil {
		ipt.feeder.FeedLastError(err.Error(),
			metrics.WithLastErrorInput(inputName),
			metrics.WithLastErrorCategory(point.Object),
		)
		l.Errorf("feed dbm query object failed: %s", err.Error())
	}
}
