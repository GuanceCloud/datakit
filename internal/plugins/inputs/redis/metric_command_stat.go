// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package redis

import (
	"bufio"
	"strconv"
	"strings"

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
		Cat:  point.Metric,
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
	var (
		collectCache = []*point.Point{}
		rdr          = strings.NewReader(list)
		scanner      = bufio.NewScanner(rdr)
		opts         = append(point.DefaultMetricOptions(), point.WithTimestamp(ipt.alignTS))
	)

	for scanner.Scan() {
		line := scanner.Text()
		if len(line) == 0 || line[0] == '#' {
			l.Warnf("ignore comment or empty line %q", line)
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
		kvs = kvs.AddTag("method", parts[0])

		itemStrs := strings.Split(parts[1], ",")
		for _, itemStr := range itemStrs {
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

			kvs = kvs.Add(arr[0], f, false, false)
		}

		if kvs.FieldCount() > 0 {
			for k, v := range ipt.mergedTags {
				kvs = kvs.AddTag(k, v)
			}
			collectCache = append(collectCache, point.NewPointV2(redisCommandStat, kvs, opts...))
		} else {
			l.Warnf("no field, ignored line %q", line)
		}
	}

	return collectCache, nil
}
