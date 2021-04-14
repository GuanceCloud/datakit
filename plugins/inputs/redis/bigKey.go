package redis

import (
	"time"
	// "fmt"

	"github.com/go-redis/redis"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

type bigKeyMeasurement struct {
	client            *redis.Client
	name              string
	tags              map[string]string
	fields            map[string]interface{}
	ts                time.Time
	lastTimestampSeen map[string]int64
	keys              []string
	WarnOnMissingKeys bool
}

func (m *bigKeyMeasurement) LineProto() (*io.Point, error) {
	return io.MakePoint(m.name, m.tags, m.fields, m.ts)
}

func (m *bigKeyMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "redis_bigkey_scan",
		Fields: map[string]interface{}{
			"key_length": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Desc:     "Key length",
			},
		},
		Tags: map[string]interface{}{
			"server": &inputs.TagInfo{
				Desc: "Server addr",
			},
		},
	}
}

func CollectBigKeyMeasurement(input *Input) *bigKeyMeasurement {
	m := &bigKeyMeasurement{
		client:            input.client,
		tags:              make(map[string]string),
		fields:            make(map[string]interface{}),
		lastTimestampSeen: make(map[string]int64),
		keys:              input.Keys,
		WarnOnMissingKeys: input.WarnOnMissingKeys,
	}

	m.name = "redis_bigkey"
	m.tags = input.Tags
	// m.tags["db_name"] = fmt.Sprintf("%d", input.DB)
	m.getData()

	return m
}

// 数据源获取数据
func (m *bigKeyMeasurement) getData() error {
	for _, key := range m.keys {
		found := false
		m.tags["key"] = key

		for _, op := range []string{
			"HLEN",
			"LLEN",
			"SCARD",
			"ZCARD",
			"PFCOUNT",
			"STRLEN",
		} {
			if val, err := m.client.Do(op, key).Result(); err == nil && val != nil {
				found = true
				m.fields["key_length"] = val
				break
			}
		}

		if !found {
			if m.WarnOnMissingKeys {
				l.Warnf("%s key not found in redis", key)
			}

			m.fields["key_length"] = 0
		}
	}

	return nil
}
