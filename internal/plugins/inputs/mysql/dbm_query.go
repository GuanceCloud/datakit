// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package mysql

import (
	"fmt"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"github.com/hashicorp/golang-lru/v2/expirable"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

const (
	dbmQueryObjectName = "db_query"
)

type dbmQueryObjectMeasurement struct{}

func (*dbmQueryObjectMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: dbmQueryObjectName,
		Cat:  point.Object,
		//nolint:lll
		Desc:   "MySQL DBM query objects. Each object represents a unique SQL query identified by query_signature, containing the obfuscated SQL text.",
		DescZh: "MySQL DBM SQL 文本对象。每个对象代表一个由 query_signature 唯一标识的 SQL 查询，包含脱敏后的 SQL 文本。",
		Tags: map[string]interface{}{
			"name":              inputs.NewTagInfo("Object identity built from server, database_instance, and query_signature"),
			"digest":            inputs.NewTagInfo("The digest hash value computed from the original normalized statement"),
			"server":            inputs.NewTagInfo("The server address (host:port)"),
			"database_instance": inputs.NewTagInfo("MySQL instance identifier from configured tag or @@server_uuid."),
			"database_type":     inputs.NewTagInfo("The type of the database. The value is `MySQL`"),
			"schema_name":       inputs.NewTagInfo("The schema name"),
			"query_signature":   inputs.NewTagInfo("The hash signature computed from schema and digest"),
		},
		Fields: map[string]interface{}{
			"message": &inputs.FieldInfo{
				DataType: inputs.String,
				Unit:     inputs.UnknownUnit,
				Desc:     "The obfuscated/normalized SQL text (digest_text)",
			},
		},
	}
}

func (ipt *Input) collectDbmQueries(dbmRows []dbmRow, ptsTime time.Time) {
	if len(dbmRows) == 0 {
		return
	}

	// Initialize query object cache if not already initialized
	if ipt.dbmQueryObjectCache == nil {
		ipt.dbmQueryObjectCache = expirable.NewLRU[string, struct{}](
			100000,
			nil,
			24*time.Hour,
		)
	}

	start := time.Now()

	opts := append(point.DefaultObjectOptions(), point.WithTime(ptsTime))
	var pts []*point.Point

	for _, row := range dbmRows {
		if row.querySignature == "" {
			continue
		}
		// Check cache to avoid duplicate reporting
		if ipt.dbmQueryObjectCache != nil {
			if _, ok := ipt.dbmQueryObjectCache.Get(row.querySignature); ok {
				l.Debugf("skip duplicate dbm query object: %s", row.querySignature)
				continue
			}
		}

		kvs := ipt.getKVs()

		// Tags
		objectName := fmt.Sprintf("%s-%s", ipt.Object.name, row.querySignature)
		if ipt.InstanceName != "" {
			objectName = fmt.Sprintf("%s-%s-%s", ipt.Object.name, ipt.InstanceName, row.querySignature)
		}
		kvs = kvs.AddTag("name", objectName)
		kvs = kvs.AddTag("digest", row.digest)
		kvs = kvs.AddTag("server", ipt.Object.name)
		kvs = kvs.AddTag("database_type", "MySQL")
		kvs = kvs.AddTag("schema_name", row.schemaName)
		kvs = kvs.AddTag("query_signature", row.querySignature)

		// Fields - digest_text is already obfuscated in getCleanSummaryRows
		kvs = kvs.Set("message", row.digestText)

		pt := point.NewPoint(dbmQueryObjectName, kvs, opts...)
		pts = append(pts, pt)

		// Add to reported cache after successfully building the point
		if ipt.dbmQueryObjectCache != nil {
			ipt.dbmQueryObjectCache.Add(row.querySignature, struct{}{})
		}
	}

	if len(pts) > 0 {
		if err := ipt.feeder.Feed(point.Object, pts,
			dkio.WithCollectCost(time.Since(start)),
			dkio.WithElection(ipt.Election),
			dkio.WithSource(dbmFeedName),
		); err != nil {
			l.Errorf("feed dbm query DBO failed: %s", err.Error())
		}
	}
}
