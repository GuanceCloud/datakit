// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package redis

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/GuanceCloud/cliutils/point"

	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

func (i *instance) collectLatency(ctx context.Context) {
	collectStart := time.Now()
	list := i.curCli.do(ctx, "latency", "latest").String()
	pts := i.parseLatencyData(list)

	if len(pts) == 0 {
		return
	}

	if err := i.ipt.feeder.Feed(point.Logging, pts,
		dkio.WithElection(i.ipt.Election),
		dkio.WithCollectCost(time.Since(collectStart)),
		dkio.WithSource(dkio.FeedSource(inputName, "latency"))); err != nil {
		l.Warnf("feed: %s, ignored", err.Error())
	}
}

// example data:
// "latency latest: [[command 1699346177 250 1000] [xxxxx 1699346178 251 1001]...]".
func (i *instance) parseLatencyData(list string) []*point.Point {
	pts := []*point.Point{}
	opts := point.DefaultLoggingOptions()

	part := strings.Split(list, "[[")

	// redis have not new latency event
	if len(part) != 2 {
		return nil
	}

	line := strings.Split(part[1], "] [")
	for _, finalParts := range line {
		var kvs point.KVs

		finalParts = strings.ReplaceAll(finalParts, "[", "")
		finalParts = strings.ReplaceAll(finalParts, "]", "")

		// example data:
		// command 1699346177 250 1000
		finalPart := strings.Split(finalParts, " ")

		if len(finalPart) != 4 {
			continue
		}

		typeName := finalPart[0]
		startTime, err := strconv.ParseInt(finalPart[1], 10, 64)
		if err != nil {
			continue
		}

		ts := time.Unix(startTime, 0)
		if ts == i.latencyLastTime[typeName] {
			continue
		}

		// kvs = kvs.AddTag("server_addr", i.server)

		fieldName := []string{"event_name", "occur_time", "cost_time", "max_cost_time"}
		for index, info := range fieldName {
			value, err := strconv.Atoi(finalPart[index])
			if err != nil {
				kvs = kvs.Add(info, finalPart[index])
			}
			kvs = kvs.Add(info, value)
		}

		kvs = kvs.Add("message",
			finalPart[0]+" cost time "+finalPart[2]+"ms"+",max_cost_time "+finalPart[3]+"ms",
		)

		for k, v := range i.mergedTags {
			kvs = kvs.AddTag(k, v)
		}

		opts = append(opts, point.WithTime(ts))
		pt := point.NewPoint(measureuemtRedisLatency, kvs, opts...)
		pts = append(pts, pt)

		i.latencyLastTime[typeName] = ts
	}

	return pts
}

type latencyMeasurement struct{}

//nolint:lll,funlen
func (m *latencyMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: measureuemtRedisLatency,
		Cat:  point.Logging,
		Fields: map[string]interface{}{
			"occur_time": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.TimestampSec,
				Desc:     "Unix timestamp of the latest latency spike for the event.",
			},
			"cost_time": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.DurationMS,
				Desc:     "Latest event latency in millisecond.",
			},
			"max_cost_time": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.DurationMS,
				Desc:     "All-time maximum latency for this event.",
			},
			"event_name": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.Gauge,
				Unit:     inputs.NoUnit,
				Desc:     "Event name.",
			},
		},
		Tags: map[string]interface{}{
			"server": &inputs.TagInfo{Desc: "Server addr"},
		},
	}
}
