package ddtrace

import (
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/trace"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

type DDTraceMeasurement struct {
	trace.TraceMeasurement
}

func (d *DDTraceMeasurement) LineProto() (*io.Point, error) {
	return d.TraceMeasurement.LineProto()
}

func (d *DDTraceMeasurement) Info() *inputs.MeasurementInfo {
	dm := d.TraceMeasurement.Info()
	dm.Name = inputName

	return dm
}
