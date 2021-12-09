//go:build !windows
// +build !windows

package winevent

import (
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

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

func (*Input) AvailableArchs() []string {
	return []string{datakit.OSWindows}
}

func (w *Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{
		&Measurement{},
	}
}

func (w *Input) Run() {}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		s := &Input{}
		return s
	})
}
