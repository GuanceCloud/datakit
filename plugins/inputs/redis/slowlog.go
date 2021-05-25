package redis

import (
	"errors"
	"time"

	"github.com/go-redis/redis"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

type slowlogMeasurement struct {
	client            *redis.Client
	name              string
	tags              map[string]string
	fields            map[string]interface{}
	ts                time.Time
	slowlogMaxLen     int
	lastTimestampSeen map[string]int64
}

func (m *slowlogMeasurement) LineProto() (*io.Point, error) {
	return io.MakePoint(m.name, m.tags, m.fields, m.ts)
}

func (m *slowlogMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "redis_slowlog",
		Fields: map[string]interface{}{
			"command": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.Gauge,
				Desc:     "slow command",
			},
			"slowlog_micros": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Desc:     "cost time",
			},
		},
	}
}

// 数据源获取数据
func (m *slowlogMeasurement) getData() ([]inputs.Measurement, error) {
	var maxSlowEntries int
	maxSlowEntries = m.slowlogMaxLen

	slowlogs, err := m.client.Do("SLOWLOG", "GET", maxSlowEntries).Result()
	if err != nil {
		return nil, err
	}

	var maxTs int64
	for _, slowlog := range slowlogs.([]interface{}) {
		if entry, ok := slowlog.([]interface{}); ok {
			if entry == nil || len(entry) != 4 {
				return nil, errors.New("slowlog get protocol error")
			}

			startTime := entry[1].(int64)
			if startTime <= m.lastTimestampSeen["server"] {
				continue
			}
			if startTime > maxTs {
				maxTs = startTime
			}
			duration := entry[2].(int64)

			var command []string
			if obj, ok := entry[3].([]interface{}); ok {
				for _, arg := range obj {
					command = append(command, string(arg.([]uint8)))
				}
			}

			m.fields["command"] = command[0]
			m.fields["slowlog_micros"] = duration
		}
	}

	addr := m.tags["server"]
	m.lastTimestampSeen[addr] = maxTs

	return []inputs.Measurement{m}, nil
}
