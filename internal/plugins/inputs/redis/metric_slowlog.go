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
	"strings"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

type slowlogMeasurement struct {
	name          string
	tags          map[string]string
	fields        map[string]interface{}
	ts            time.Time
	slowlogMaxLen int
	election      bool
}

func (m *slowlogMeasurement) LineProto() (*point.Point, error) {
	return point.NewPoint(m.name, m.tags, m.fields, point.LOptElectionV2(m.election))
}

//nolint:lll
func (m *slowlogMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Desc: "Redis 慢查询命令历史，这里我们将其以日志的形式采集",
		Name: redisSlowlog,
		Type: "logging",
		Tags: map[string]interface{}{
			"server": &inputs.TagInfo{
				Desc: "server",
			},
			"service_name": &inputs.TagInfo{Desc: "Service name"},
			"message": &inputs.TagInfo{
				Desc: "log message",
			},
			"host": &inputs.TagInfo{
				Desc: "host",
			},
		},
		Fields: map[string]interface{}{
			"slowlog_id": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.UnknownUnit,
				Desc:     "Slow log unique id",
			},
			"slowlog_micros": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.DurationUS,
				Desc:     "Cost time",
			},
			"command": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.DurationUS,
				Desc:     "Slow command",
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

	if _, ok := slowlogs.([]interface{}); !ok {
		return fmt.Errorf("unexpected slowlogs, expect []inerface{}, got type %s", reflect.TypeOf(slowlogs).String())
	}

	m := &slowlogMeasurement{
		name: redisSlowlog,
		tags: func() map[string]string {
			x := map[string]string{
				"service": "redis",
				"host":    i.Host,
			}

			for k, v := range i.Tags {
				x[k] = v
			}
			return x
		}(),

		fields: make(map[string]interface{}),

		slowlogMaxLen: i.SlowlogMaxLen,

		election: i.Election,
	}

	var pts []*point.Point

	for _, slowlog := range slowlogs.([]interface{}) {
		entry, ok := slowlog.([]interface{})
		if !ok {
			l.Warnf("unexpected slowlog, expect []inerface{}, got type %s, ignored", reflect.TypeOf(slowlog).String())
			continue
		}

		if entry == nil || len(entry) != 6 {
			return fmt.Errorf("protocol error: slowlog expect 6 fields, got %+#v", entry)
		}

		var id int64
		if x, ok := entry[0].(int64); ok {
			id = x
		} else {
			return fmt.Errorf("id expect int64, got %s", reflect.TypeOf(entry[0]).String())
		}

		var startTime int64
		if x, ok := entry[1].(int64); ok {
			startTime = x
		} else {
			return fmt.Errorf("startTime expect int64, got %s", reflect.TypeOf(entry[1]).String())
		}

		m.ts = time.Now()

		var duration int64
		if x, ok := entry[2].(int64); ok {
			duration = x
		} else {
			return fmt.Errorf("duration expect int64, got %s", reflect.TypeOf(entry[2]).String())
		}

		// Skip collected slowlog.
		if !i.AllSlowLog {
			if startTime < i.startUpUnix {
				continue
			}
			hashRes := md5.Sum([]byte(strconv.FormatInt(startTime, 10) + string(rune(id)))) //nolint:gosec
			if i.hashMap[int32(id%int64(i.SlowlogMaxLen))] == hashRes {
				continue // ignore old slow-logs
			}
			i.hashMap[int32(id%int64(i.SlowlogMaxLen))] = hashRes
		}

		var command string
		if obj, ok := entry[3].([]interface{}); ok {
			arr := []string{}
			for _, arg := range obj {
				if x, ok := arg.(string); ok {
					arr = append(arr, x)
				}
			}

			command = strings.Join(arr, " ")
		}

		m.fields = map[string]interface{}{
			"slowlog_micros": duration,
			"slowlog_id":     id,
			"command":        command,
			"message":        command + " cost time " + strconv.FormatInt(duration, 10) + "us",
			"status":         "WARNING",
		}

		pt, err := point.NewPoint(redisSlowlog, m.tags, m.fields,
			&point.PointOption{Time: m.ts, Category: datakit.Logging, Strict: true})
		if err != nil {
			l.Warnf("make metric failed: %s", err.Error)
			return err
		}

		pts = append(pts, pt)
	}

	err = io.Feed(m.name, datakit.Logging, pts, &io.Option{})
	if err != nil {
		return err
	}
	return nil
}
