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
	"github.com/montanaflynn/stats"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
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
				Type:     inputs.UnknownType,
				Unit:     inputs.UnknownUnit,
				Desc:     "Slow log unique ID",
			},
			"slowlog_micros": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.DurationUS,
				Desc:     "Cost time",
			},
			"command": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.UnknownType,
				Unit:     inputs.UnknownUnit,
				Desc:     "Slow command",
			},
			"slowlog_avg": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.DurationUS,
				Desc:     "Slow average duration",
			},
			"slowlog_max": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.DurationUS,
				Desc:     "Slow maximum duration",
			},
			"slowlog_median": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.DurationUS,
				Desc:     "Slow median duration",
			},
			"slowlog_95percentile": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.DurationUS,
				Desc:     "Slow 95th percentile duration",
			},

			"client_addr": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.UnknownType,
				Unit:     inputs.UnknownUnit,
				Desc:     "The client ip:port that run the slow query",
			},

			"client_name": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.UnknownType,
				Unit:     inputs.UnknownUnit,
				Desc:     "The client name that run the slow query(if `client setname` executed on client-side)",
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
		if err := ipt.feeder.FeedV2(point.Logging, pts,
			dkio.WithElection(ipt.Election),
			dkio.WithInputName(redisSlowlog)); err != nil {
			return err
		}
	}

	return nil
}

func (ipt *Input) parseSlowData(slowlogs any) ([]*point.Point, error) {
	collectCache := []*point.Point{}
	opts := point.DefaultLoggingOptions()
	opts = append(opts, point.WithTime(time.Now()))

	var durationList []float64
	for _, slowlog := range slowlogs.([]interface{}) {
		var kvs point.KVs

		kvs = kvs.AddTag("service", "redis")
		kvs = kvs.AddTag("host", ipt.Host)

		entry, ok := slowlog.([]interface{})
		if !ok {
			l.Warnf("unexpected slowlog, expect []inerface{}, got type %s, ignored", reflect.TypeOf(slowlog).String())
			continue
		}

		if entry == nil || len(entry) < 4 {
			return nil, fmt.Errorf("protocol error: slowlog expect at least 4 fields, got %+#v", entry)
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
			durationList = append(durationList, float64(x))
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

		// calculate avg, max, median, P95, count
		maxDur, _ := stats.Max(durationList)
		avgDur, _ := stats.Mean(durationList)
		midDur, _ := stats.Median(durationList)
		p95Dur, err := stats.Percentile(durationList, 0.95)
		if err != nil {
			p95Dur = maxDur
		}

		kvs = kvs.Add("slowlog_avg", avgDur, false, false)
		kvs = kvs.Add("slowlog_max", int64(maxDur), false, false)
		kvs = kvs.Add("slowlog_median", int64(midDur), false, false)
		kvs = kvs.Add("slowlog_95percentile", p95Dur, false, false)

		if len(entry) >= 6 { // redis 4.0+
			if x, ok := entry[4].(string); ok {
				kvs = kvs.Add("client_addr", x, false, false)
			}

			if x, ok := entry[5].(string); ok {
				kvs = kvs.Add("client_name", x, false, false)
			}
		}

		for k, v := range ipt.mergedTags {
			kvs = kvs.AddTag(k, v)
		}
		collectCache = append(collectCache, point.NewPointV2(redisSlowlog, kvs, opts...))
	}
	return collectCache, nil
}
