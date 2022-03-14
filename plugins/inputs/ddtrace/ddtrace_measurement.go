// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

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
