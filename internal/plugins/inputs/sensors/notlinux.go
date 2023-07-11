// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build !linux
// +build !linux

package sensors

import (
	"runtime"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

// Input redefine here for sample checking on non-linux platform.
type Input struct{}

var _ inputs.InputV2 = (*Input)(nil)

func (*Input) Catalog() string {
	return inputName
}

func (*Input) Terminate() {}

func (*Input) SampleConfig() string {
	return sampleConfig
}

func (*Input) AvailableArchs() []string {
	return []string{datakit.OSLabelLinux}
}

func (*Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{&sensorsMeasurement{}}
}

func (*Input) Run() {
	l.Errorf("Can not run input %q on %s-%s.", inputName, runtime.GOOS, runtime.GOARCH)
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input { return &Input{} })
}
