package self

import (
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

type datakitMeasurement struct {
	inputs.CommonMeasurement
	ts time.Time
}

func (m *datakitMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   measurementName,
		Fields: map[string]interface{}{},
		Tags:   map[string]interface{}{},
	}
}
