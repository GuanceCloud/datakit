// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package oracle

import (
	"fmt"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	dkpt "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

const (
	oracleProcess    = "oracle_process"
	oracleTablespace = "oracle_tablespace"
	oracleSystem     = "oracle_system"
)

type processMeasurement struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
	ts     time.Time
	opt    point.Option
}

// Point implement MeasurementV2.
func (m *processMeasurement) Point() *point.Point {
	opts := point.DefaultMetricOptions()
	opts = append(opts, point.WithTime(m.ts), m.opt)

	return point.NewPointV2([]byte(m.name),
		append(point.NewTags(m.tags), point.NewKVs(m.fields)...),
		opts...)
}

// 生成行协议.
func (m *processMeasurement) LineProto() (*dkpt.Point, error) {
	// return point.NewPoint(m.name, m.tags, m.fields, point.MOptElection())
	return nil, fmt.Errorf("not implement")
}

// 指定指标.
func (m *processMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: oracleProcess,
		Fields: map[string]interface{}{
			// status
			"pga_used_mem": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.SizeByte,
				Desc:     "PGA memory used by process",
			},
			"pga_alloc_mem": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.SizeByte,
				Desc:     "PGA memory allocated by process",
			},
			"pga_freeable_mem": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.SizeByte,
				Desc:     "PGA memory freeable by process",
			},
			"pga_max_mem": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.SizeByte,
				Desc:     "PGA maximum memory ever allocated by process",
			},
		},
		Tags: map[string]interface{}{
			"oracle_server": &inputs.TagInfo{
				Desc: "Server addr",
			},
			"oracle_service": &inputs.TagInfo{
				Desc: "Server service",
			},
			"program": &inputs.TagInfo{
				Desc: "Program",
			},
			"host": &inputs.TagInfo{
				Desc: "host",
			},
		},
	}
}

type tablespaceMeasurement struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
	ts     time.Time
	opt    point.Option
}

// Point implement MeasurementV2.
func (m *tablespaceMeasurement) Point() *point.Point {
	opts := point.DefaultMetricOptions()
	opts = append(opts, point.WithTime(m.ts), m.opt)

	return point.NewPointV2([]byte(m.name),
		append(point.NewTags(m.tags), point.NewKVs(m.fields)...),
		opts...)
}

// 生成行协议.
func (m *tablespaceMeasurement) LineProto() (*dkpt.Point, error) {
	// return point.NewPoint(m.name, m.tags, m.fields, dkpt.MOptElection())
	return nil, fmt.Errorf("not implement")
}

// 指定指标.
func (m *tablespaceMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: oracleTablespace,
		Fields: map[string]interface{}{
			// status
			"used_space": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Used space",
			},
			"ts_size": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.SizeByte,
				Desc:     "Table space size",
			},
			"in_use": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Table space in-use",
			},
			"off_use": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Table space offline",
			},
		},
		Tags: map[string]interface{}{
			"oracle_server": &inputs.TagInfo{
				Desc: "Server addr",
			},
			"oracle_service": &inputs.TagInfo{
				Desc: "Server service",
			},
			"tablespace_name": &inputs.TagInfo{
				Desc: "Table space",
			},
			"host": &inputs.TagInfo{
				Desc: "host",
			},
		},
	}
}

type systemMeasurement struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
	ts     time.Time
	opt    point.Option
}

// Point implement MeasurementV2.
func (m *systemMeasurement) Point() *point.Point {
	opts := point.DefaultMetricOptions()
	opts = append(opts, point.WithTime(m.ts), m.opt)

	return point.NewPointV2([]byte(m.name),
		append(point.NewTags(m.tags), point.NewKVs(m.fields)...),
		opts...)
}

// 生成行协议.
func (m *systemMeasurement) LineProto() (*dkpt.Point, error) {
	// return point.NewPoint(m.name, m.tags, m.fields, dkpt.MOpt())
	return nil, fmt.Errorf("not implement")
}

// 指定指标.
func (m *systemMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: oracleSystem,
		Desc: "You have to wait for a few minutes to see these metrics when your running Oracle database's version is earlier than 12c.",
		Fields: map[string]interface{}{
			// status
			"buffer_cachehit_ratio": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Ratio of buffer cache hits",
			},
			"cursor_cachehit_ratio": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Ratio of cursor cache hits",
			},
			"library_cachehit_ratio": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Ratio of library cache hits",
			},
			"shared_pool_free": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.SizeByte,
				Desc:     "Shared pool free memory %",
			},
			"physical_reads": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Physical reads per sec",
			},
			"physical_writes": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Physical writes per sec",
			},
			"enqueue_timeouts": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Enqueue timeouts per sec",
			},
			"gc_cr_block_received": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "GC CR block received",
			},
			"cache_blocks_corrupt": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Corrupt cache blocks",
			},
			"cache_blocks_lost": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Lost cache blocks",
			},
			"active_sessions": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of active sessions",
			},
			"service_response_time": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Service response time",
			},
			"user_rollbacks": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of user rollbacks",
			},
			"sorts_per_user_call": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Sorts per user call",
			},
			"rows_per_sort": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Rows per sort",
			},
			"disk_sorts": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Disk sorts per second",
			},
			"memory_sorts_ratio": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Memory sorts ratio",
			},
			"database_wait_time_ratio": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Memory sorts per second",
			},
			"session_limit_usage": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Session limit usage",
			},
			"session_count": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Session count",
			},
			"temp_space_used": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Temp space used",
			},
		},
		Tags: map[string]interface{}{
			"oracle_server": &inputs.TagInfo{
				Desc: "Server addr",
			},
			"oracle_service": &inputs.TagInfo{
				Desc: "Server service",
			},
			"host": &inputs.TagInfo{
				Desc: "host",
			},
		},
	}
}
