// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package redis

// nolint
import (
	"context"
	"crypto/md5"
	"fmt"
	"reflect"
	"strconv"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
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
			"slowlog_id": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.Gauge,
				Unit:     inputs.UnknownUnit,
				Desc:     "slowlog unique id",
			},
			"slowlog_micros": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.DurationUS,
				Desc:     "cost time",
			},
			"command": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.DurationUS,
				Desc:     "slow command",
			},
		},
	}
}

// 数据源获取数据.
func (i *Input) getSlowData() error {
	maxSlowEntries := i.SlowlogMaxLen
	ctx := context.Background()
	slowlogs, err := i.client.Do(ctx, "SLOWLOG", "GET", maxSlowEntries).Result()
	if err != nil {
		l.Error("redis exec SLOWLOG `get`, happen error,", err)
		return err
	}

	m := &slowlogMeasurement{
		tags:              make(map[string]string),
		fields:            make(map[string]interface{}),
		lastTimestampSeen: make(map[string]int64),
		slowlogMaxLen:     i.SlowlogMaxLen,
	}

	var maxTS int64
	for _, slowlog := range slowlogs.([]interface{}) {
		if entry, ok := slowlog.([]interface{}); ok {
			if entry == nil || len(entry) != 6 {
				return fmt.Errorf("protocol error: slowlog expect 6 fields, got %+#v", entry)
			}

			m.name = "redis_slowlog"

			for k, v := range i.Tags {
				m.tags[k] = v
			}
			var id int64
			if x, isint := entry[0].(int64); isint {
				id = x
			} else {
				return fmt.Errorf("id expect int64, got %s", reflect.TypeOf(entry[1]).String())
			}

			var startTime int64
			if x, isi64 := entry[1].(int64); isi64 {
				startTime = x
			} else {
				return fmt.Errorf("startTime expect int64, got %s", reflect.TypeOf(entry[1]).String())
			}

			if !ok {
				return fmt.Errorf("%v expect to be int64", entry[1])
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
				return fmt.Errorf("duration expect int64, got %s", reflect.TypeOf(entry[2]).String())
			}
			hashStr := strconv.FormatInt(startTime, 10) + string(rune(id))
			// nolint
			hashRes := md5.Sum([]byte(hashStr))

			var command string
			if obj, isok := entry[3].([]interface{}); isok {
				for _, arg := range obj {
					command += arg.(string) + " "
				}
			}

			m.ts = time.Unix(startTime, 0)

			m.fields["slowlog_micros"] = duration
			m.fields["slowlog_id"] = id
			m.fields["command"] = command
			addr := m.tags["server"]
			m.lastTimestampSeen[addr] = maxTS
			if i.hashMap[int32(id%int64(i.SlowlogMaxLen))] != hashRes {
				m.tags["message"] = command + " cost time " + strconv.FormatInt(duration, 10) + "us"
				data, err := io.NewPoint("redis_slowlog", m.tags, m.fields, &io.PointOption{Time: m.ts, Category: datakit.Logging, Strict: true})
				if err != nil {
					l.Warnf("make metric failed: %s", err.Error)
					return err
				}

				var pts []*io.Point
				pts = append(pts, data)
				err = io.Feed(m.name, datakit.Logging, pts, &io.Option{})
				if err != nil {
					return err
				}
				i.hashMap[int32(id%int64(i.SlowlogMaxLen))] = hashRes
			}
		}
	}
	return nil
}
