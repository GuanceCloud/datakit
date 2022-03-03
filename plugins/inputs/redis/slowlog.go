package redis

import (
	"context"
	"fmt"
	"reflect"
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

//nolint:lll
func (m *slowlogMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "redis_slowlog",
		Tags: map[string]interface{}{
			"server": &inputs.TagInfo{
				Desc: "server",
			},
		},
		Fields: map[string]interface{}{
			"command": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.Gauge,
				Unit:     inputs.UnknownUnit,
				Desc:     "slow command",
			},
			"slowlog_micros": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.DurationUS,
				Desc:     "cost time",
			},
		},
	}
}

// 数据源获取数据.
func (i *Input) getSlowData() ([]inputs.Measurement, error) {
	var collectCache []inputs.Measurement
	maxSlowEntries := i.SlowlogMaxLen

	ctx := context.Background()
	slowlogs, err := i.client.Do(ctx, "SLOWLOG", "GET", maxSlowEntries).Result()
	if err != nil {
		l.Error("redis exec SLOWLOG `get`, happen error,", err)
		return nil, err
	}

	var maxTS int64
	for _, slowlog := range slowlogs.([]interface{}) {
		if entry, ok := slowlog.([]interface{}); ok {
			if entry == nil || len(entry) != 6 {
				return nil, fmt.Errorf("protocol error: slowlog expect 6 fields, got %+#v", entry)
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

			var startTime int64
			if x, isi64 := entry[1].(int64); isi64 {
				startTime = x
			} else {
				return nil, fmt.Errorf("startTime expect int64, got %s", reflect.TypeOf(entry[1]).String())
			}

			if !ok {
				return nil, fmt.Errorf("%v expect to be int64", entry[1])
			}

			if startTime <= m.lastTimestampSeen["server"] {
				continue
			}

			if startTime > maxTS {
				maxTS = startTime
			}

			var duration int64
			if x, isi64 := entry[2].(int64); isi64 {
				duration = x
			} else {
				return nil, fmt.Errorf("duration expect int64, got %s", reflect.TypeOf(entry[2]).String())
			}

			var command string
			if obj, isok := entry[3].([]interface{}); isok {
				for _, arg := range obj {
					command += arg.(string) + " "
				}
			}

			m.ts = time.Unix(startTime, 0)

			m.fields["command"] = command
			m.fields["slowlog_micros"] = duration

			collectCache = append(collectCache, m)

			addr := m.tags["server"]
			m.lastTimestampSeen[addr] = maxTS
		}
	}
	return collectCache, nil
}
