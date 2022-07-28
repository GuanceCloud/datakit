// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build !windows || !amd64
// +build !windows !amd64

package iis

import (
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

var (
	inputName            = "iis"
	metricNameWebService = "iis_web_service"
	metricNameAppPoolWas = "iis_app_pool_was"

	_ inputs.InputV2 = (*Input)(nil)
)

// Input redefine them here for conf-sample checking.
type Input struct {
	Interval datakit.Duration

	Tags map[string]string

	Log *iisLog `toml:"log"`
}

type iisLog struct {
	Files    []string `toml:"files"`
	Pipeline string   `toml:"pipeline"`
}

func (i *Input) SampleConfig() string {
	return sampleConfig
}

func (*Input) Terminate() {
	// Do nothing
}

func (i *Input) Catalog() string {
	return "iis"
}

func (*Input) RunPipeline() {
	// TODO
}

func (i *Input) AvailableArchs() []string {
	return []string{
		datakit.OSArchWinAmd64,
	}
}

func (i *Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{
		&IISAppPoolWas{},
		&IISWebService{},
	}
}

func (*Input) PipelineConfig() map[string]string {
	pipelineConfig := map[string]string{
		inputName: pipelineCfg,
	}
	return pipelineConfig
}

func (i *Input) Run() {}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return &Input{}
	})
}
