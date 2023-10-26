// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package coredns collect coreDNS metrics by using input prom
package coredns

import (
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

const (
	inputName = "coredns"
)

type Input struct{}

var _ inputs.InputV2 = (*Input)(nil)

func (i *Input) Terminate() {
	// do nothing
}

func (i *Input) Catalog() string {
	return inputName
}

func (i *Input) SampleConfig() string {
	return configSample
}

func (i *Input) Run() {
}

func (i *Input) AvailableArchs() []string {
	return datakit.AllOSWithElection
}

func (i *Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{
		&aCLMeasurement{},
		&cacheMeasurement{},
		&dnsSecMeasurement{},
		&forwardMeasurement{},
		&grpcMeasurement{},
		&hostsMeasurement{},
		&templateMeasurement{},
		&promMeasurement{},
	}
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return &Input{}
	})
}
