// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

const (
	redisBigkey      = "redis_bigkey"
	redisClient      = "redis_client"
	redisCluster     = "redis_cluster"
	redisCommandStat = "redis_command_stat"
	redisDB          = "redis_db"
	redisLatency     = "redis_latency"
	redisInfoM       = "redis_info"
	redisReplica     = "redis_replica"
	redisSlowlog     = "redis_slowlog"
)

type bigKeyMeasurement struct{}

//nolint:lll
func (m *bigKeyMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: redisBigkey,
		Type: "metric",
		Fields: map[string]interface{}{
			"value_length": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.UnknownUnit,
				Desc:     "Key length",
			},
			/*"key": &inputs.FieldInfo{
				DataType: inputs.String,
				Type: inputs.String,
				Desc: "monitor key",
			},*/
		},
		Tags: map[string]interface{}{
			"host": &inputs.TagInfo{
				Desc: "Hostname",
			},
			"server": &inputs.TagInfo{
				Desc: "Server addr",
			},
			"service_name": &inputs.TagInfo{Desc: "Service name"},
			"db_name": &inputs.TagInfo{
				Desc: "DB name.",
			},
			"key": &inputs.TagInfo{
				Desc: "Monitor key",
			},
		},
	}
}

func (ipt *Input) getKeys() ([]string, error) {
	var res []string
	for _, pattern := range ipt.Keys {
		var cursor uint64
		for {
			var keys []string
			var err error
			ctx := context.Background()

			keys, cursor, err = ipt.client.Scan(ctx, cursor, pattern, 10).Result()
			if err != nil {
				l.Errorf("redis pattern key %s scan fail error %v", pattern, err)
				return nil, err
			}
			// keys: []string{"key1","key2"...}

			res = append(res, keys...)
			if cursor == 0 {
				break
			}
		}
	}
	return res, nil
}

func (ipt *Input) getData(resKeys []string) ([]*point.Point, error) {
	collectCache := []*point.Point{}
	opts := point.DefaultMetricOptions()
	opts = append(opts, point.WithTime(time.Now()))

	for _, key := range resKeys {
		var kvs point.KVs
		kvs = kvs.AddTag("db_name", fmt.Sprintf("%d", ipt.DB))
		kvs = kvs.AddTag("key", key)

		found := false
		ctx := context.Background()
		for _, op := range []string{
			"HLEN",
			"LLEN",
			"SCARD",
			"ZCARD",
			"PFCOUNT",
			"STRLEN",
		} {
			if val, err := ipt.client.Do(ctx, op, key).Result(); err == nil && val != nil {
				// op:"STRLEN", key:"key1", val=interface{}(int64)5
				found = true
				kvs = kvs.Add("value_length", val, false, true)
				break
			}
		}

		if !found {
			if ipt.WarnOnMissingKeys {
				l.Warnf("%s key not found in redis", key)
			}

			kvs = kvs.Add("value_length", 0, false, true)
		}

		if kvs.FieldCount() > 0 {
			for k, v := range ipt.mergedTags {
				kvs = kvs.AddTag(k, v)
			}
			collectCache = append(collectCache, point.NewPointV2(redisBigkey, kvs, opts...))
		}
	}

	return collectCache, nil
}
