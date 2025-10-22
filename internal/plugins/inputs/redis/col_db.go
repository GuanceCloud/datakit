// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package redis

import (
	"bufio"
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

func (i *instance) collectDB(ctx context.Context) {
	collectStart := time.Now()
	list, err := i.curCli.info(ctx, "keyspace")
	if err != nil {
		l.Warnf("Info.keyspace: %s, ignored", err.Error())
		return
	}

	pts, err := i.parseDBData(list)
	if err != nil {
		l.Warnf("parseDBData: %s, ignored", err.Error())
		return
	}

	if err := i.ipt.feeder.Feed(point.Metric, pts,
		dkio.WithCollectCost(time.Since(collectStart)),
		dkio.WithElection(i.ipt.Election),
		dkio.WithSource(dkio.FeedSource(inputName, "db")),
		dkio.WithMeasurement(inputs.GetOverrideMeasurement(i.ipt.MeasurementVersion, measureuemtRedis)),
	); err != nil {
		l.Warnf("feed measurement: %s, ignored", err)
	}
}

func (i *instance) parseDBData(list string) ([]*point.Point, error) {
	collectCache := []*point.Point{}
	opts := point.DefaultMetricOptions()
	opts = append(opts, point.WithTime(i.ipt.ptsTime))

	rdr := strings.NewReader(list)
	scanner := bufio.NewScanner(rdr)
	dbIndexSlice := i.ipt.DBs

	// example data
	// db0:keys=43706,expires=117,avg_ttl=30904274304765
	for scanner.Scan() {
		var kvs point.KVs

		line := scanner.Text()

		if len(line) == 0 || line[0] == '#' {
			continue
		}

		parts := strings.Split(line, ":")
		if len(parts) != 2 {
			continue
		}
		db := parts[0]

		// Only process lines starting with "db" (e.g., db0, db1, db15)
		if !strings.HasPrefix(db, "db") || len(db) < 3 {
			continue
		}

		dbIndex, err := strconv.Atoi(db[2:])
		if err != nil {
			// Skip lines that don't match db{number} format
			continue
		}

		kvs = kvs.AddTag("db_name", parts[0])

		itemStrs := strings.Split(parts[1], ",")
		for _, itemStr := range itemStrs {
			item := strings.Split(itemStr, "=")

			f, err := strconv.ParseFloat(item[1], 64)
			if err != nil {
				continue
			}

			kvs = kvs.Add(item[0], f)
		}

		if len(i.ipt.DBs) == 0 {
			// if !IsSlicesHave(ipt.keyDBS, dbIndex) {
			//	ipt.keyDBS = append(ipt.keyDBS, dbIndex)
			//}
			if kvs.FieldCount() > 0 {
				for k, v := range i.mergedTags {
					kvs = kvs.AddTag(k, v)
				}
				collectCache = append(collectCache, point.NewPoint(measureuemtRedisDB, kvs, opts...))
			}
		} else if isSlicesHave(dbIndexSlice, dbIndex) && kvs.FieldCount() > 0 {
			for k, v := range i.mergedTags {
				kvs = kvs.AddTag(k, v)
			}
			collectCache = append(collectCache, point.NewPoint(measureuemtRedisDB, kvs, opts...))
		}
	}

	return collectCache, nil
}

type dbMeasurement struct{}

func (m *dbMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: measureuemtRedisDB,
		Cat:  point.Metric,
		Fields: map[string]interface{}{
			"keys":    &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Desc: "Key."},
			"expires": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Desc: "expires time."},
			"avg_ttl": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Desc: "Average ttl."},
		},
		Tags: map[string]interface{}{
			"db_name":      &inputs.TagInfo{Desc: "DB name."},
			"host":         &inputs.TagInfo{Desc: "Hostname."},
			"server":       &inputs.TagInfo{Desc: "Server addr."},
			"service_name": &inputs.TagInfo{Desc: "Service name."},
		},
	}
}
