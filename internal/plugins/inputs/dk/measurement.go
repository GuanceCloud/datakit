// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package dk

import (
	"fmt"

	dkpt "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

func (*Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{meas{}}
}

type meas struct{}

// LineProto not implemented.
func (meas) LineProto() (*dkpt.Point, error) {
	return nil, fmt.Errorf("dk not implement interface LineProto()")
}

//nolint:lll
func (meas) Info() *inputs.MeasurementInfo {
	return nil
}
