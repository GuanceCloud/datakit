package smart

import (
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

type smartMeasurement struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
	ts     time.Time
}

func (s *smartMeasurement) LineProto() (*io.Point, error) {
	return io.MakePoint(s.name, s.tags, s.fields, s.ts)
}

func (s *smartMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   inputName,
		Tags:   map[string]interface{}{},
		Fields: map[string]interface{}{},
	}
}
