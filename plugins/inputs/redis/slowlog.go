package redis

import (
	"context"
	"errors"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

type slowlogMeasurement struct {
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
func (i *Input) getSlowData() ([]inputs.Measurement, error) {
	var collectCache []inputs.Measurement
	var maxSlowEntries int
	maxSlowEntries = i.SlowlogMaxLen

	ctx := context.Background()
	slowlogs, err := i.client.Do(ctx, "SLOWLOG", "GET", maxSlowEntries).Result()
	if err != nil {
		l.Error("redis exec SLOWLOG `get`, happen error,", err)
		return nil, err
	}

	var maxTs int64
	for _, slowlog := range slowlogs.([]interface{}) {
		if entry, ok := slowlog.([]interface{}); ok {
			if entry == nil || len(entry) != 6 {
				return nil, errors.New("slowlog get protocol error")
			}

			m := &slowlogMeasurement{
				tags:              make(map[string]string),
				fields:            make(map[string]interface{}),
				lastTimestampSeen: make(map[string]int64),
				slowlogMaxLen:     i.SlowlogMaxLen,
			}

			m.name = "redis_slowlog"

			for k, v := range i.Tags {
				m.tags[k] = v
			}

			startTime := entry[1].(int64)
			if startTime <= m.lastTimestampSeen["server"] {
				continue
			}

			if startTime > maxTs {
				maxTs = startTime
			}
			duration := entry[2].(int64)

			var command string
			if obj, ok := entry[3].([]interface{}); ok {
				for _, arg := range obj {
					command += arg.(string) + " "
				}
			}

			m.ts = time.Unix(startTime, 0)

			m.fields["command"] = command
			m.fields["slowlog_micros"] = duration

			collectCache = append(collectCache, m)

			addr := m.tags["server"]
			m.lastTimestampSeen[addr] = maxTs
		}
	}
	return collectCache, nil
}
