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
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

type dbMeasurement struct{}

func (m *dbMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: redisDB,
		Type: "metric",
		Fields: map[string]interface{}{
			"keys":    &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Desc: "Key."},
			"expires": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Desc: "expires time."},
			"avg_ttl": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Desc: "Average ttl."},
		},
		Tags: map[string]interface{}{
			"db":           &inputs.TagInfo{Desc: "DB name."},
			"host":         &inputs.TagInfo{Desc: "Hostname."},
			"server":       &inputs.TagInfo{Desc: "Server addr."},
			"service_name": &inputs.TagInfo{Desc: "Service name."},
		},
	}
}

func (ipt *Input) collectDBMeasurement() ([]*point.Point, error) {
	ctx := context.Background()
	list, err := ipt.client.Info(ctx, "keyspace").Result()
	if err != nil {
		return nil, err
	}

	return ipt.parseDBData(list)
}

func (ipt *Input) parseDBData(list string) ([]*point.Point, error) {
	collectCache := []*point.Point{}
	opts := point.DefaultMetricOptions()
	opts = append(opts, point.WithTime(time.Now()))

	rdr := strings.NewReader(list)
	scanner := bufio.NewScanner(rdr)
	dbIndexSlice := ipt.DBS

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
		dbIndex, err := strconv.Atoi(db[2:])
		if err != nil {
			return collectCache, err
		}

		kvs = kvs.AddTag("db_name", parts[0])

		itemStrs := strings.Split(parts[1], ",")
		for _, itemStr := range itemStrs {
			item := strings.Split(itemStr, "=")

			f, err := strconv.ParseFloat(item[1], 64)
			if err != nil {
				continue
			}

			kvs = kvs.Add(item[0], f, false, false)
		}

		if len(ipt.DBS) == 0 {
			if !IsSlicesHave(ipt.keyDBS, dbIndex) {
				ipt.keyDBS = append(ipt.keyDBS, dbIndex)
			}
			if kvs.FieldCount() > 0 {
				for k, v := range ipt.mergedTags {
					kvs = kvs.AddTag(k, v)
				}
				collectCache = append(collectCache, point.NewPointV2(redisDB, kvs, opts...))
			}
		} else if IsSlicesHave(dbIndexSlice, dbIndex) && kvs.FieldCount() > 0 {
			for k, v := range ipt.mergedTags {
				kvs = kvs.AddTag(k, v)
			}
			collectCache = append(collectCache, point.NewPointV2(redisDB, kvs, opts...))
		}
	}

	return collectCache, nil
}
