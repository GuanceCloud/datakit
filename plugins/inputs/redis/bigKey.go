package redis

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
	"time"
)

type bigKeyMeasurement struct {
	client *redis.Client
	name   string
	tags   map[string]string
	fields map[string]interface{}
	ts     time.Time
}

func (m *bigKeyMeasurement) LineProto() (*io.Point, error) {
	return io.MakePoint(m.name, m.tags, m.fields, m.ts)
}

func (m *bigKeyMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "redis_bigkey",
		Fields: map[string]interface{}{
			"value_length": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Desc:     "Key length",
			},
		},
		Tags: map[string]interface{}{
			"server": &inputs.TagInfo{
				Desc: "Server addr",
			},
			"db_name": &inputs.TagInfo{
				Desc: "db",
			},
			"key": &inputs.TagInfo{
				Desc: "monitor key",
			},
		},
	}
}

func (i *Input) getKeys() ([]string, error) {
	var res []string
	for _, pattern := range i.Keys {
		var cursor uint64
		for {
			var keys []string
			var err error
			ctx := context.Background()

			keys, cursor, err = i.client.Scan(ctx, cursor, pattern, 10).Result()
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

// 数据源获取数据
func (i *Input) getData(resKeys []string) ([]inputs.Measurement, error) {
	var collectCache []inputs.Measurement

	for _, key := range resKeys {
		found := false

		m := &commandMeasurement{
			name:   "redis_bigkey",
			tags:   make(map[string]string),
			fields: make(map[string]interface{}),
		}

		for key, value := range i.Tags {
			m.tags[key] = value
		}

		m.tags["db_name"] = fmt.Sprintf("%d", i.DB)
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
			if val, err := i.client.Do(ctx, op, key).Result(); err == nil && val != nil {
				found = true
				m.fields["value_length"] = val
				break
			}
		}

		if !found {
			if i.WarnOnMissingKeys {
				l.Warnf("%s key not found in redis", key)
			}

			m.fields["value_length"] = 0
		}

		if len(m.fields) > 0 {
			collectCache = append(collectCache, m)
		}
	}

	return collectCache, nil
}
