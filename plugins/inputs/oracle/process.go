package oracle

import (
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

type processMeasurement struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
	ts     time.Time
}

// 生成行协议
func (m *processMeasurement) LineProto() (*io.Point, error) {
	return io.MakePoint(m.name, m.tags, m.fields, m.ts)
}

// 指定指标
func (m *processMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "oracle_process",
		Fields: map[string]interface{}{
			// status
			"pga_used_memory": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.SizeByte,
				Desc:     "PGA memory used by process",
			},
			"pga_allocated_memory": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.SizeByte,
				Desc:     "PGA memory allocated by process",
			},
			"pga_freeable_memory": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.SizeByte,
				Desc:     "PGA memory freeable by process",
			},
			"pga_maximum_memory": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.SizeByte,
				Desc:     "PGA maximum memory ever allocated by process",
			},
		},
		Tags: map[string]interface{}{
			"server": &inputs.TagInfo{
				Desc: "server addr",
			},
			"program": &inputs.TagInfo{
				Desc: "program",
			},
		},
	}
}

type tablespaceMeasurement struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
	ts     time.Time
}

// 生成行协议
func (m *tablespaceMeasurement) LineProto() (*io.Point, error) {
	return io.MakePoint(m.name, m.tags, m.fields, m.ts)
}

// 指定指标
func (m *tablespaceMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "oracle_tablespace",
		Fields: map[string]interface{}{
			// status
			"used_space": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Desc:     "The latency of the redis INFO command.",
			},
			"ts_size": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Desc:     "tablespace size",
			},
			"in_use": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Desc:     "tablespace in-use",
			},
			"off_use": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Desc:     "tablespace offline",
			},
		},
		Tags: map[string]interface{}{
			"server": &inputs.TagInfo{
				Desc: "server addr",
			},
			"tablespace": &inputs.TagInfo{
				Desc: "table space",
			},
		},
	}
}

type systemMeasurement struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
	ts     time.Time
}

// 生成行协议
func (m *systemMeasurement) LineProto() (*io.Point, error) {
	return io.MakePoint(m.name, m.tags, m.fields, m.ts)
}

// 指定指标
func (m *systemMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "oracle_system",
		Fields: map[string]interface{}{
			// status
			"buffer_cachehit_ratio": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Desc:     "Ratio of buffer cache hits",
			},
			"cursor_cachehit_ratio": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Desc:     "Ratio of cursor cache hits",
			},
			"library_cachehit_ratio": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Desc:     "Ratio of library cache hits",
			},
			"shared_pool_free": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Desc:     "shared pool free memory %",
			},
			"physical_reads": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Desc:     "physical reads per sec",
			},
			"physical_writes": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Desc:     "physical writes per sec",
			},
			"enqueue_timeouts": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Desc:     "enqueue timeouts per sec",
			},
			"gc_cr_block_received": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Desc:     "GC CR block received",
			},
			"cache_blocks_corrupt": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Desc:     "corrupt cache blocks",
			},
			"cache_blocks_lost": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Desc:     "lost cache blocks",
			},
			"active_sessions": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Desc:     "number of active sessions",
			},
			"service_response_time": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Desc:     "service response time",
			},
			"user_rollbacks": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Desc:     "number of user rollbacks",
			},
			"sorts_per_user_call": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Desc:     "sorts per user call",
			},
			"rows_per_sort": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Desc:     "rows per sort",
			},
			"disk_sorts": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Desc:     "disk sorts per second",
			},
			"memory_sorts_ratio": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Desc:     "memory sorts ratio",
			},
			"database_wait_time_ratio": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Desc:     "memory sorts per second",
			},
			"session_limit_usage": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Desc:     "session limit usage",
			},
			"session_count": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Desc:     "session count",
			},
			"temp_space_used": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Desc:     "temp space used",
			},
		},
		Tags: map[string]interface{}{
			"server": &inputs.TagInfo{
				Desc: "server addr",
			},
		},
	}
}
