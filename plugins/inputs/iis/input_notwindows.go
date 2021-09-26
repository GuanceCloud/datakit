// +build !windows !amd64

package iis

import (
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/tailer"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

var (
	inputName            = "iis"
	metricNameWebService = "iis_web_service"
	metricNameAppPoolWas = "iis_app_pool_was"
)

// redefine them here for conf-sample checking.
type Input struct {
	Interval datakit.Duration

	Tags map[string]string

	Log  *iisLog `toml:"log"`
	tail *tailer.Tailer

	collectCache []inputs.Measurement
}

type iisLog struct {
	Files    []string `toml:"files"`
	Pipeline string   `toml:"pipeline"`
}

func (i *Input) SampleConfig() string {
	return sampleConfig
}

func (i *Input) Catalog() string {
	return "iis"
}

// TODO.
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
