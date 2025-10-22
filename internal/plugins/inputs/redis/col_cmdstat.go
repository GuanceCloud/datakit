// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package redis

import (
	"bufio"
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/GuanceCloud/cliutils/point"

	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

func (i *instance) collectCommandStats(ctx context.Context) {
	collectStart := time.Now()

	list, err := i.curCli.info(ctx, "commandstats")
	if err != nil {
		l.Error("command stats error,", err)
		return
	}

	pts, err := i.parseCommandData(list)
	if err != nil {
		l.Warnf("parseCommandData: %s, ignored", err)
		return
	}

	if err := i.ipt.feeder.Feed(point.Metric, pts,
		dkio.WithCollectCost(time.Since(collectStart)),
		dkio.WithElection(i.ipt.Election),
		dkio.WithSource(dkio.FeedSource(inputName, "cmdstat")),
		dkio.WithMeasurement(inputs.GetOverrideMeasurement(i.ipt.MeasurementVersion, measureuemtRedis)),
	); err != nil {
		l.Warnf("feed measurement: %s, ignored", err)
	}
}

func (i *instance) parseCommandData(list string) ([]*point.Point, error) {
	var (
		pts     = []*point.Point{}
		scanner = bufio.NewScanner(strings.NewReader(list))
		opts    = append(point.DefaultMetricOptions(), point.WithTime(i.ipt.ptsTime))
	)

	for scanner.Scan() {
		line := scanner.Text()
		if len(line) == 0 || line[0] == '#' {
			l.Debugf("ignore comment or empty line %q", line)
			continue
		}

		parts := strings.Split(line, ":")
		if len(parts) != 2 {
			l.Warnf("ignore line %q", line)
			continue
		}

		// example data:
		// cmdstat_client|list:calls=1,usec=25,usec_per_call=25.00,rejected_calls=0,failed_calls=0
		var kvs point.KVs
		if len(parts[0]) > 8 { // remove `cmdstat_' prefix
			kvs = kvs.AddTag("method", parts[0][8:])
		} else {
			kvs = kvs.AddTag("method", parts[0]) // NOTE: should not been here
		}

		// following `,' splited fields.
		arr := strings.Split(parts[1], ",")
		for _, itemStr := range arr {
			arr := strings.Split(itemStr, "=")
			if len(arr) != 2 {
				l.Warnf("ignore itemStr %q within %q", itemStr, parts[1])
				continue
			}

			f, err := strconv.ParseFloat(arr[1], 64)
			if err != nil {
				l.Warnf("ignore value %q on key %q within %q", arr[1], arr[0], parts[1])
				continue
			}

			kvs = kvs.Add(arr[0], f)
		}

		if kvs.FieldCount() == 0 {
			l.Warnf("ignore no-field point on line %q", line)
			continue
		}

		for k, v := range i.mergedTags {
			kvs = kvs.AddTag(k, v)
		}
		pts = append(pts, point.NewPoint(measureuemtRedisCommandStat, kvs, opts...))
	}

	return pts, nil
}

type commandMeasurement struct{}

// see also: https://redis.io/commands/info/
//
//nolint:lll
func (m *commandMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "redis_command_stat",
		Cat:  point.Metric,
		Fields: map[string]interface{}{
			"calls": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The number of calls that reached command execution.",
			},
			"usec": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.DurationUS,
				Desc:     "The total CPU time consumed by these commands.",
			},
			"usec_per_call": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.DurationUS,
				Desc:     "The average CPU consumed per command execution.",
			},
			"rejected_calls": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The number of rejected calls (errors prior command execution).",
			},
			"failed_calls": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The number of failed calls (errors within the command execution).",
			},
		},
		Tags: map[string]interface{}{
			"host":         &inputs.TagInfo{Desc: "Hostname"},
			"method":       &inputs.TagInfo{Desc: "Command type"},
			"server":       &inputs.TagInfo{Desc: "Server addr"},
			"service_name": &inputs.TagInfo{Desc: "Service name"},
		},
	}
}
