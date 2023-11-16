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

type clientMeasurement struct{}

// see also: https://redis.io/commands/client-list/
//
//nolint:lll
func (m *clientMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: redisClient,
		Type: "metric",
		Fields: map[string]interface{}{
			"id": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Unique 64-bit client ID."},
			// Not number, discard. "addr":      &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "address/port of the client."},
			// Not number, discard. "laddr":     &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "address/port of local address client connected to (bind address)."},
			"fd": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "File descriptor corresponding to the socket."},
			// Not number, discard. "name":      &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "the name set by the client with CLIENT SETNAME."},
			"age":  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationSecond, Desc: "Total duration of the connection in seconds"},
			"idle": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationSecond, Desc: "Idle time of the connection in seconds"},
			// Not number, discard. "flags":     &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "client flags (see below)."},
			"db":        &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Current database ID."},
			"sub":       &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of channel subscriptions"},
			"psub":      &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of pattern matching subscriptions"},
			"ssub":      &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of shard channel subscriptions. Added in Redis 7.0.3."},
			"multi":     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of commands in a MULTI/EXEC context."},
			"qbuf":      &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Query buffer length (0 means no query pending)."},
			"qbuf_free": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Free space of the query buffer (0 means the buffer is full)."},
			"argv_mem":  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Incomplete arguments for the next command (already extracted from query buffer)."},
			"multi_mem": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Memory is used up by buffered multi commands. Added in Redis 7.0."},
			"obl":       &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Output buffer length."},
			"oll":       &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Output list length (replies are queued in this list when the buffer is full)."},
			"omem":      &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Output buffer memory usage."},
			"tot_mem":   &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Total memory consumed by this client in its various buffers."},
			// Not number, discard. "events":    &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "File descriptor events (see below)."},
			// Not number, discard. "cmd":       &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Last command played."},
			// Not number, discard. "user":      &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The authenticated username of the client."},
			"redir": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Client id of current client tracking redirection."},
			"resp":  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Client RESP protocol version. Added in Redis 7.0."},
		},
		Tags: map[string]interface{}{
			"addr":         &inputs.TagInfo{Desc: "Address without port of the client"},
			"host":         &inputs.TagInfo{Desc: "Hostname"},
			"name":         &inputs.TagInfo{Desc: "The name set by the client with `CLIENT SETNAME`, default unknown"},
			"server":       &inputs.TagInfo{Desc: "Server addr"},
			"service_name": &inputs.TagInfo{Desc: "Service name"},
		},
	}
}

func (ipt *Input) parseClientData(list string) ([]*point.Point, error) {
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

		parts := strings.Split(line, " ")

		for _, part := range parts {
			item := strings.Split(part, "=")

			key := item[0]
			key = strings.ReplaceAll(key, "-", "_")
			val := strings.TrimSpace(item[1])

			// https://gitlab.jiagouyun.com/cloudcare-tools/datakit/-/issues/1743
			switch key {
			case "addr", "name":
				if val == "" {
					val = "unknown"
				}

				if key == "addr" {
					// exclude port.
					arr := strings.Split(val, ":")
					kvs = kvs.AddTag(key, arr[0])
				} else {
					// "name"
					kvs = kvs.AddTag(key, val)
				}

			case "id": // drop it.
			default:
				f, err := strconv.ParseFloat(val, 64)
				if err != nil {
					continue
				}
				if _, has := clientFieldMap[key]; has {
					// key in the MeasurementInfo
					kvs = kvs.Add(key, f, false, false)
				}
			}
		}

		if kvs.FieldCount() > 0 {
			for k, v := range ipt.mergedTags {
				kvs = kvs.AddTag(k, v)
			}
			collectCache = append(collectCache, point.NewPointV2(redisClient, kvs, opts...))
		}
	}

	return collectCache, nil
}

var clientFieldMap = map[string]struct{}{}

func getClientFieldMap() {
	m := clientMeasurement{}
	for k := range m.Info().Fields {
		clientFieldMap[k] = struct{}{}
	}
}
