package oracle

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
		Fields: map[string]*inputs.FieldInfo{
			// status
			"pga_used_memory": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.Gauge,
				Unit:     inputs.DurationMS,
				Desc:     "The latency of the redis INFO command.",
			},
			"pga_allocated_memory": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.Gauge,
				Desc:     "",
			},
			"pga_freeable_memory": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.Gauge,
				Desc:     "",
			},
			"pga_maximum_memory": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.Gauge,
				Desc:     "",
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
		Fields: map[string]*inputs.FieldInfo{
			// status
			"used": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.Gauge,
				Unit:     inputs.DurationMS,
				Desc:     "The latency of the redis INFO command.",
			},
			"size": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.Gauge,
				Desc:     "",
			},
			"in_use": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.Gauge,
				Desc:     "",
			},
			"offline": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.Gauge,
				Desc:     "",
			},
		},
	}
}
