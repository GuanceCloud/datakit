package mysql

import (
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

type schemaMeasurement struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
	ts     time.Time
}

// 生成行协议.
func (m *schemaMeasurement) LineProto() (*io.Point, error) {
	return io.MakePoint(m.name, m.tags, m.fields, m.ts)
}

// 指定指标.
func (m *schemaMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "mysql_schema",
		Desc: "具体字段，以实际采集上来的数据为准，部分字段，会因 MySQL 配置、已有数据等原因，采集不到",
		Fields: map[string]interface{}{
			"schema_size": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.SizeMiB,
				Desc:     "Size of schemas(MiB)",
			},
			"query_run_time_avg": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.DurationNS,
				Desc:     "Avg query response time per schema.",
			},
		},
		Tags: map[string]interface{}{
			"server": &inputs.TagInfo{
				Desc: "Server addr",
			},
			"schema_name": &inputs.TagInfo{
				Desc: "Schema name",
			},
		},
	}
}
