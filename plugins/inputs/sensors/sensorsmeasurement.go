package sensors

import (
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

type sensorsMeasurement struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
	ts     time.Time
}

func (m *sensorsMeasurement) LineProto() (*io.Point, error) {
	return io.MakePoint(m.name, m.tags, m.fields, m.ts)
}

func (m *sensorsMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "sensors",
		Tags: map[string]interface{}{
			"hostname": &inputs.TagInfo{Desc: "host name"},
			"adapter":  &inputs.TagInfo{Desc: "device adapter"},
			"chip":     &inputs.TagInfo{Desc: "chip id"},
			"feature":  &inputs.TagInfo{Desc: "gathering target"},
		},
		Fields: map[string]interface{}{
			"tmep*_crit":       &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.Celsius, Desc: `Critical temperature of this chip, '*' is the order number in the chip list.`},
			"temp*_crit_alarm": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.Celsius, Desc: `Alarm count, '*' is the order number in the chip list.`},
			"temp*_input":      &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.Celsius, Desc: `Current input temperature of this chip, '*' is the order number in the chip list.`},
			"tmep*_max":        &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.Celsius, Desc: `Max temperature of this chip, '*' is the order number in the chip list.`},
		},
	}
}
