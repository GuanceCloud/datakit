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

func (i *instance) collectSlowLog(ctx context.Context) {
	if !i.ipt.EnableSlowLog {
		return
	}

	collectStart := time.Now()

	slowlogs, err := i.curCli.do(ctx, "SLOWLOG", "GET", i.ipt.SlowlogMaxLen).Result()
	if err != nil {
		l.Warnf("SLOWLOG GET: %s", err)
		return
	}

	if _, ok := slowlogs.([]interface{}); !ok {
		l.Warnf("unexpected slowlogs, expect []inerface{}, got type %s", reflect.TypeOf(slowlogs).String())
		return
	}

	pts, err := i.parseSlowData(slowlogs)
	if err != nil {
		l.Warnf("parseSlowData: %s, ignored", err.Error())
		return
	}

	if len(pts) > 0 {
		if err := i.ipt.feeder.Feed(point.Logging, pts,
			dkio.WithElection(i.ipt.Election),
			dkio.WithCollectCost(time.Since(collectStart)),
			dkio.WithSource(dkio.FeedSource(inputName, "slowlog"))); err != nil {
			l.Warnf("feed: %s", err.Error())
		}
	}
}

func (i *instance) parseSlowData(slowlogs any) ([]*point.Point, error) {
	var (
		pts  = []*point.Point{}
		opts = append(point.DefaultLoggingOptions(), point.WithTime(time.Now()))
	)

	var costArr []float64
	for _, slowlog := range slowlogs.([]interface{}) {
		var (
			kvs                 point.KVs
			id, startTime, cost int64
			command             string
		)

		entry, ok := slowlog.([]interface{})
		if !ok {
			l.Warnf("unexpected slowlog, expect []inerface{}, got type %s, ignored", reflect.TypeOf(slowlog).String())
			continue
		}

		if entry == nil || len(entry) < 4 {
			return nil, fmt.Errorf("protocol error: slowlog expect at least 4 fields, got %+#v", entry)
		}

		if x, ok := entry[0].(int64); ok {
			id = x
			kvs = kvs.Add("slowlog_id", id)
		} else {
			return nil, fmt.Errorf("id expect int64, got %s", reflect.TypeOf(entry[0]).String())
		}

		if x, ok := entry[1].(int64); ok {
			startTime = x
		} else {
			return nil, fmt.Errorf("startTime expect int64, got %s", reflect.TypeOf(entry[1]).String())
		}

		if x, ok := entry[2].(int64); ok {
			cost = x
			kvs = kvs.Add("slowlog_micros", cost)
			costArr = append(costArr, float64(x))
		} else {
			return nil, fmt.Errorf("command cost expect int64, got %s", reflect.TypeOf(entry[2]).String())
		}

		// Skip collected slowlog.
		if !i.ipt.CollectAllSlowLog {
			if startTime < i.ipt.startUpUnix {
				continue
			}

			slogHash := md5.Sum([]byte(strconv.FormatInt(startTime, 10) + string(rune(id)))) //nolint:gosec
			if i.slowlogHash[int32(id%int64(i.ipt.SlowlogMaxLen))] == slogHash {
				continue // ignore old slow-logs
			}

			i.slowlogHash[int32(id%int64(i.ipt.SlowlogMaxLen))] = slogHash
		}

		// parse slow command details
		if obj, ok := entry[3].([]interface{}); ok {
			arr := []string{}
			for _, arg := range obj {
				if x, ok := arg.(string); ok {
					arr = append(arr, x)
				}
			}

			command = strings.Join(arr, " ")
			kvs = kvs.Set("command", command)
		}

		kvs = kvs.Add("slowlog_micros", cost)
		kvs = kvs.Add("slowlog_id", id)
		kvs = kvs.Add("command", command)
		kvs = kvs.Add("message", command+" cost time "+strconv.FormatInt(cost, 10)+"us")
		kvs = kvs.Add("status", "WARNING")

		// calculate avg, max, median, P95, count
		maxDur, _ := stats.Max(costArr)
		avgDur, _ := stats.Mean(costArr)
		midDur, _ := stats.Median(costArr)
		p95Dur, err := stats.Percentile(costArr, 0.95)
		if err != nil {
			p95Dur = maxDur
		}

		kvs = kvs.Add("slowlog_avg", avgDur)
		kvs = kvs.Add("slowlog_max", int64(maxDur))
		kvs = kvs.Add("slowlog_median", int64(midDur))
		kvs = kvs.Add("slowlog_95percentile", p95Dur)

		if len(entry) >= 6 { // redis 4.0+
			if x, ok := entry[4].(string); ok {
				kvs = kvs.Add("client_addr", x)
			}

			if x, ok := entry[5].(string); ok {
				kvs = kvs.Add("client_name", x)
			}
		}

		for k, v := range i.mergedTags {
			kvs = kvs.AddTag(k, v)
		}
		pts = append(pts, point.NewPoint(measureuemtRedisSlowLog, kvs, opts...))
	}

	return pts, nil
}

type slowlogMeasurement struct{}

//nolint:lll
func (m *slowlogMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Desc:   "Redis slow query history logging",
		DescZh: "采集 Redis 慢查询历史",
		Name:   measureuemtRedisSlowLog,
		Cat:    point.Logging,
		Tags: map[string]interface{}{
			"server": &inputs.TagInfo{
				Desc: "server",
			},
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
				Unit:     inputs.NoUnit,
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
				Unit:     inputs.NoUnit,
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
				Unit:     inputs.NoUnit,
				Desc:     "The client ip:port that run the slow query",
			},

			"client_name": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.UnknownType,
				Unit:     inputs.NoUnit,
				Desc:     "The client name that run the slow query(if `client setname` executed on client-side)",
			},
		},
	}
}
