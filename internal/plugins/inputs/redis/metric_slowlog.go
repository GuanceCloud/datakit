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

	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

type slowlogMeasurement struct{}

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

func (ipt *Input) getSlowData() error {
	ctx := context.Background()
	slowlogs, err := ipt.client.Do(ctx, "SLOWLOG", "GET", ipt.SlowlogMaxLen).Result()
	if err != nil {
		l.Error("redis exec SLOWLOG `get`, happen error,", err)
		return err
	}

	if _, ok := slowlogs.([]interface{}); !ok {
		return fmt.Errorf("unexpected slowlogs, expect []inerface{}, got type %s", reflect.TypeOf(slowlogs).String())
	}

	pts, err := ipt.parseSlowData(slowlogs)
	if err != nil {
		return err
	}

	if len(pts) > 0 {
		err = ipt.feeder.Feed(redisSlowlog, point.Logging, pts, &io.Option{})
		if err != nil {
			return err
		}
	}

	return nil
}

func (ipt *Input) parseSlowData(slowlogs any) ([]*point.Point, error) {
	collectCache := []*point.Point{}
	opts := point.DefaultLoggingOptions()
	opts = append(opts, point.WithTime(time.Now()))

	for _, slowlog := range slowlogs.([]interface{}) {
		var kvs point.KVs

		kvs = kvs.AddTag("service", "redis")
		kvs = kvs.AddTag("host", ipt.Host)

		entry, ok := slowlog.([]interface{})
		if !ok {
			l.Warnf("unexpected slowlog, expect []inerface{}, got type %s, ignored", reflect.TypeOf(slowlog).String())
			continue
		}

		if entry == nil || len(entry) != 6 {
			return nil, fmt.Errorf("protocol error: slowlog expect 6 fields, got %+#v", entry)
		}

		var id int64
		if x, ok := entry[0].(int64); ok {
			id = x
		} else {
			return nil, fmt.Errorf("id expect int64, got %s", reflect.TypeOf(entry[0]).String())
		}

		var startTime int64
		if x, ok := entry[1].(int64); ok {
			startTime = x
		} else {
			return nil, fmt.Errorf("startTime expect int64, got %s", reflect.TypeOf(entry[1]).String())
		}

		var duration int64
		if x, ok := entry[2].(int64); ok {
			duration = x
		} else {
			return nil, fmt.Errorf("duration expect int64, got %s", reflect.TypeOf(entry[2]).String())
		}

		// Skip collected slowlog.
		if !ipt.AllSlowLog {
			if startTime < ipt.startUpUnix {
				continue
			}
			hashRes := md5.Sum([]byte(strconv.FormatInt(startTime, 10) + string(rune(id)))) //nolint:gosec
			if ipt.hashMap[int32(id%int64(ipt.SlowlogMaxLen))] == hashRes {
				continue // ignore old slow-logs
			}
			ipt.hashMap[int32(id%int64(ipt.SlowlogMaxLen))] = hashRes
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

		kvs = kvs.Add("slowlog_micros", duration, false, false)
		kvs = kvs.Add("slowlog_id", id, false, false)
		kvs = kvs.Add("command", command, false, false)
		kvs = kvs.Add("message", command+" cost time "+strconv.FormatInt(duration, 10)+"us", false, false)
		kvs = kvs.Add("status", "WARNING", false, false)

		for k, v := range ipt.mergedTags {
			kvs = kvs.AddTag(k, v)
		}
		collectCache = append(collectCache, point.NewPointV2(redisSlowlog, kvs, opts...))
	}

	return collectCache, nil
}
