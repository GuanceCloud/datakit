// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package redis

import (
	"context"
	"fmt"

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
				Desc: "db",
			},
			"key": &inputs.TagInfo{
				Desc: "monitor key",
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

			res = append(res, keys...)
			if cursor == 0 {
				break
			}
		}
	}
	return res, nil
}

// 数据源获取数据.
func (ipt *Input) getData(resKeys []string) ([]*point.Point, error) {
	var collectCache []*point.Point

	for _, key := range resKeys {
		found := false

		m := &commandMeasurement{
			name:     redisBigkey,
			tags:     make(map[string]string),
			fields:   make(map[string]interface{}),
			election: ipt.Election,
		}

		for key, value := range ipt.Tags {
			m.tags[key] = value
		}

		m.tags["db_name"] = fmt.Sprintf("%d", ipt.DB)
		m.tags["key"] = key
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
				found = true
				m.fields["value_length"] = val
				break
			}
		}

		if !found {
			if ipt.WarnOnMissingKeys {
				l.Warnf("%s key not found in redis", key)
			}

			m.fields["value_length"] = 0
		}

		if len(m.fields) > 0 {
			var opts []point.Option

			if m.election {
				m.tags = inputs.MergeTags(ipt.Tagger.ElectionTags(), m.tags, ipt.Host)
			} else {
				m.tags = inputs.MergeTags(ipt.Tagger.HostTags(), m.tags, ipt.Host)
			}

			pt := point.NewPointV2(m.name,
				append(point.NewTags(m.tags), point.NewKVs(m.fields)...),
				opts...)
			collectCache = append(collectCache, pt)
		}
	}

	return collectCache, nil
}
