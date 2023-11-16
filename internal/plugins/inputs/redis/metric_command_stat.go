// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package redis

import (
	"bufio"
	"strconv"
	"strings"
	"time"

	"github.com/GuanceCloud/cliutils/point"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

type commandMeasurement struct{}

// see also: https://redis.io/commands/info/
//
//nolint:lll
func (m *commandMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "redis_command_stat",
		Type: "metric",
		Fields: map[string]interface{}{
			"calls":          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The number of calls that reached command execution."},
			"usec":           &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationUS, Desc: "The total CPU time consumed by these commands."},
			"usec_per_call":  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationUS, Desc: "The average CPU consumed per command execution."},
			"rejected_calls": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The number of rejected calls (errors prior command execution)."},
			"failed_calls":   &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The number of failed calls (errors within the command execution)."},
		},
		Tags: map[string]interface{}{
			"host":         &inputs.TagInfo{Desc: "Hostname"},
			"method":       &inputs.TagInfo{Desc: "Command type"},
			"server":       &inputs.TagInfo{Desc: "Server addr"},
			"service_name": &inputs.TagInfo{Desc: "Service name"},
		},
	}
}

func (ipt *Input) parseCommandData(list string) ([]*point.Point, error) {
	collectCache := []*point.Point{}
	opts := point.DefaultMetricOptions()
	opts = append(opts, point.WithTime(time.Now()))

	rdr := strings.NewReader(list)
	scanner := bufio.NewScanner(rdr)
	for scanner.Scan() {
		var kvs point.KVs

		line := scanner.Text()
		if len(line) == 0 || line[0] == '#' {
			continue
		}

		parts := strings.Split(line, ":")
		if len(parts) != 2 {
			continue
		}

		// example data:
		// cmdstat_client|list:calls=1,usec=25,usec_per_call=25.00,rejected_calls=0,failed_calls=0
		kvs = kvs.AddTag("method", parts[0])

		itemStrs := strings.Split(parts[1], ",")
		for _, itemStr := range itemStrs {
			item := strings.Split(itemStr, "=")

			f, err := strconv.ParseFloat(item[1], 64)
			if err != nil {
				continue
			}

			kvs = kvs.Add(item[0], f, false, false)
		}

		if kvs.FieldCount() > 0 {
			for k, v := range ipt.mergedTags {
				kvs = kvs.AddTag(k, v)
			}
			collectCache = append(collectCache, point.NewPointV2(redisCommandStat, kvs, opts...))
		}
	}

	return collectCache, nil
}
