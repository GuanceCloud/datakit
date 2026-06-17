// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package ibm_i exposes documentation for the IBM i external collector.
package ibm_i

import (
	"github.com/GuanceCloud/cliutils/logger"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/external"
)

const (
	inputName   = "ibm_i"
	catalogName = "host"

	measurementName = "ibm_i"
)

var l = logger.DefaultSLogger(inputName)

type Input struct {
	external.Input
}

func (ipt *Input) Run() {
	l.Info("Only for measurement documentation information, should not be here.")
}

func (*Input) Catalog() string { return catalogName }

func (*Input) SampleConfig() string { return sampleConfig }

func (*Input) AvailableArchs() []string {
	return []string{datakit.OSLabelLinux, datakit.LabelElection}
}

func (*Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{
		&ibmiMeasurement{},
		&inputs.UpMeasurement{},
	}
}

func defaultInput() *Input {
	return &Input{Input: *external.NewInput()}
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return defaultInput()
	})
}
