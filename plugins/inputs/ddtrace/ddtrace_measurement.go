package ddtrace

import (
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/trace"
)

type DdtraceMeasurement struct {
	trace.TraceMeasurement
}

func (d *DdtraceMeasurement) LineProto() (*io.Point, error) {
	return d.TraceMeasurement.LineProto()
}

func (d *DdtraceMeasurement) Info() *inputs.MeasurementInfo {
	dm := d.TraceMeasurement.Info()
	dm.Name = inputName
	return dm
}
