// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package dk

import (
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

func (*Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{meas{}}
}

type meas struct{}

//nolint:lll
func (meas) Info() *inputs.MeasurementInfo {
	return nil
}
