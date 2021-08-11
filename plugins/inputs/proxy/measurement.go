package proxy

// type measurement struct {
// 	name   string
// 	tags   map[string]string
// 	fields map[string]interface{}
// 	ts     time.Time
// }

// func (m *measurement) LineProto() (*io.Point, error) {
// 	return io.MakePoint(m.name, m.tags, m.fields, m.ts)
// }

// func (m *measurement) Info() *inputs.MeasurementInfo {
// 	return &inputs.MeasurementInfo{
// 		Name: "proxy",
// 		Tags: map[string]interface{}{
// 			"request_host": &inputs.TagInfo{Desc: "host name which send the request to the proxy"},
// 		},
// 		Fields: map[string]interface{}{
// 			"delivered":     &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "the total bytes delivered to destination url."},
// 			"increment":     &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "the total increment bytes delivered to destination url."},
// 			"request_count": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "totoal request count."},
// 		},
// 	}
// }
