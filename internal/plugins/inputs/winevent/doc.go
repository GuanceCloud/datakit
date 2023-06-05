// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build !windows
// +build !windows

package winevent

// this file used to build documents.

import (
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

var _ inputs.InputV2 = (*Input)(nil)

type EvtHandle uintptr

type Input struct {
	Query string            `toml:"xpath_query"`
	Tags  map[string]string `toml:"tags,omitempty"`
}

func (*Input) SampleConfig() string {
	return sample
}

func (*Input) Catalog() string {
	return "windows"
}

func (*Input) RunPipeline() {
	// TODO.
}

func (*Input) Terminate() {
	// Do nothing
}

func (*Input) AvailableArchs() []string {
	return []string{datakit.OSLabelWindows}
}

func (*Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{
		&Measurement{},
	}
}

func (*Input) Run() {}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		s := &Input{}
		return s
	})
}
