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
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

type latencyMeasurement struct {
	name     string
	tags     map[string]string
	fields   map[string]interface{}
	resData  map[string]interface{}
	ts       time.Time
	election bool
}

func (m *latencyMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: redisLatency,
		Type: "logging",
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
				Unit:     inputs.UnknownUnit,
				Desc:     "Event name.",
			},
		},
		Tags: map[string]interface{}{
			"server": &inputs.TagInfo{
				Desc: "Server addr",
			},
			"service_name": &inputs.TagInfo{Desc: "Service name"},
		},
	}
}

// GetLatencyData 解析数据并返回指定的数据.
func (ipt *Input) GetLatencyData() error {
	ctx := context.Background()
	list := ipt.client.Do(ctx, "latency", "latest").String()

	// [latency latest:  command 1640151523 324 1000] ]]
	part := strings.Split(list, "[[")

	// redis没有最新延迟事件
	if len(part) != 2 {
		return nil
	}
	// "command 1640151523 324 1000"
	part1 := strings.Split(part[1], "]]")

	// command 1640151523 324 1000
	finalPart := strings.Split(part1[0], " ")

	// 长度不足4
	if len(finalPart) != 4 {
		return nil
	}

	fieldName := []string{"event_name", "occur_time", "cost_time", "max_cost_time"}

	m := &latencyMeasurement{
		name:     redisLatency,
		tags:     make(map[string]string),
		fields:   make(map[string]interface{}),
		resData:  make(map[string]interface{}),
		election: ipt.Election,
	}
	m.tags["server_addr"] = ipt.Addr
	setHostTagIfNotLoopback(m.tags, ipt.Host)
	var pts []*point.Point
	for index, info := range fieldName {
		m.fields[info] = finalPart[index]
	}
	startTime, err := strconv.ParseInt(finalPart[1], 10, 64)
	m.fields["message"] = finalPart[0] + " cost time " + finalPart[2] + "ms" + ",max_cost_time " + finalPart[3] + "ms"
	if err != nil {
		l.Warnf("input redis latency unexpected data %s, ignored", finalPart[1])
	}
	m.ts = time.Unix(startTime, 0)

	opts := point.DefaultLoggingOptions()
	opts = append(opts, point.WithTime(m.ts))
	pt := point.NewPointV2(redisLatency,
		append(point.NewTags(m.tags), point.NewKVs(m.fields)...),
		opts...)

	if m.ts == ipt.latencyLastTime {
		return nil
	}

	pts = append(pts, pt)
	err = ipt.feeder.Feed(m.name, point.Logging, pts, &io.Option{})
	if err != nil {
		return err
	}
	ipt.latencyLastTime = m.ts
	return nil
}
