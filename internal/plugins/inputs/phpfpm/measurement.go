// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package phpfpm

import (
	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

type baseMeasurement struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
	ts     int64
}

//nolint:lll
func (*baseMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: metricName,
		Cat:  point.Metric,
		Fields: map[string]interface{}{
			"accepted_connections": &inputs.FieldInfo{
				Type: inputs.Count, DataType: inputs.Int, Unit: inputs.NCount,
				Desc: "The number of requests accepted by the pool.",
			},
			"listen_queue": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.NCount,
				Desc: "The number of requests in the queue of pending connections.",
			},
			"max_listen_queue": &inputs.FieldInfo{
				Type: inputs.Count, DataType: inputs.Int, Unit: inputs.NCount,
				Desc: "The maximum number of requests in the queue of pending connections since FPM has started.",
			},
			"listen_queue_length": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.NCount,
				Desc: "The size of the socket queue of pending connections.",
			},
			"idle_processes": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.NCount,
				Desc: "The number of idle processes.",
			},
			"active_processes": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.NCount,
				Desc: "The number of active processes.",
			},
			"total_processes": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.NCount,
				Desc: "The number of idle + active processes.",
			},
			"max_active_processes": &inputs.FieldInfo{
				Type: inputs.Count, DataType: inputs.Int, Unit: inputs.NCount,
				Desc: "The maximum number of active processes since FPM has started.",
			},
			"max_children_reached": &inputs.FieldInfo{
				Type: inputs.Count, DataType: inputs.Int, Unit: inputs.NCount,
				Desc: "The number of times, the process limit has been reached, when pm tries to start more children (works only for pm 'dynamic' and 'ondemand').",
			},
			"slow_requests": &inputs.FieldInfo{
				Type: inputs.Count, DataType: inputs.Int, Unit: inputs.NCount,
				Desc: "The number of requests that exceeded your 'request_slowlog_timeout' value.",
			},
			"process_requests": &inputs.FieldInfo{
				Type: inputs.Count, DataType: inputs.Int, Unit: inputs.NCount,
				Desc: "The number of requests the process has served.",
			},
			"process_last_request_memory": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.NCount,
				Desc: "The max amount of memory the last request consumed.",
			},
			"process_last_request_cpu": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.Percent,
				Desc: "The %cpu the last request consumed.",
			},
			"process_request_duration": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.DurationUS,
				Desc: "The duration in microseconds of the requests.",
			},
		},
		Tags: map[string]interface{}{
			"address":         &inputs.TagInfo{Desc: "Pool Address."},
			"pool":            &inputs.TagInfo{Desc: "The Pools Name of the FPM."},
			"process_manager": &inputs.TagInfo{Desc: "The type of the Process Manager (static, dynamic, ondemand)."},
			"process_state":   &inputs.TagInfo{Desc: "The state of the process (Idle, Running, ...)."},
			"pid":             &inputs.TagInfo{Desc: "The pid of the process."},
		},
	}
}

// Point implement MeasurementV2.
func (m *baseMeasurement) Point() *point.Point {
	opts := point.DefaultMetricOptions()
	opts = append(opts, point.WithTimestamp(m.ts))

	return point.NewPointV2(m.name,
		append(point.NewTags(m.tags), point.NewKVs(m.fields)...),
		opts...)
}
