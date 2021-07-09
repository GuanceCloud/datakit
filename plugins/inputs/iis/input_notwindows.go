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
)

type Input struct {
}

func (i *Input) SampleConfig() string {
	return sampleConfig

}

func (i *Input) Catalog() string {
	return "iis"
}

// TODO
func (*Input) RunPipeline() {
}

func (i *Input) AvailableArchs() []string {
	return []string{
		// datakit.OSArchWin386,
		datakit.OSArchWinAmd64,
	}
}

func (i *Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{
		&IISAppPoolWas{},
		&IISWebService{},
	}
}

func (i *Input) Run() {
}

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return &Input{}
	})
}
