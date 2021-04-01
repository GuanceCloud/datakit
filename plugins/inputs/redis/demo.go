package redis

import (
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)


type demoMeasurement struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
	ts     time.Time
}



func (m *demoMeasurement) LineProto() (io.Point, error) {
	return io.MakeMetric(m.name, m.tags, m.fields, m.ts)
}

func (m *demoMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "demo",
		Fields: map[string]*inputs.FieldInfo{
			"usage": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type: inputs.Gauge,
				Unit: inputs.Percent,
				Desc: "this is CPU usage",
			},
			"disk_size": &inputs.FieldInfo{
				DataType: inputs.Int, Type:
				inputs.Gauge,
				Unit: inputs.SizeIByte,
				Desc: "this is disk size",
			},
			"some_string": &inputs.FieldInfo{
				DataType: inputs.String,
				Type: inputs.Gauge,
				Unit: inputs.UnknownUnit,
				Desc: "some string field",
			},
			"ok": &inputs.FieldInfo{
				DataType: inputs.Bool,
				Type: inputs.Gauge,
				Unit: inputs.UnknownUnit,
				Desc: "some boolean field",
			},
		},
	}
}

func CollectDemoMeasurement() *demoMeasurement {
	m := &demoMeasurement{}
	m.collect()
}

func (m *demoMeasurement) collect() *demoMeasurement {
	demo := &demoMeasurement{
			name: "demo",
			tags: map[string]string{"tag_a": "a", "tag_b": "b"},
			fields: map[string]interface{}{
				"usage":       "12.3",
				"disk_size":   5e9,
				"some_string": "hello world",
				"ok":          true,
			},
			ts: time.Now(),
		},
	}

	return demo
}