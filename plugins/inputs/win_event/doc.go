//+build !windows

package win_event

import (
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

type EvtHandle uintptr

type Input struct {
	Query string            `toml:"xpath_query"`
	Tags  map[string]string `toml:"tags,omitempty"`

	subscription EvtHandle
	buf          []byte
	collectCache []inputs.Measurement
}

func (_ *Input) SampleConfig() string {
	return sample
}

func (_ *Input) Catalog() string {
	return "windows"
}

// TODO
func (*Input) RunPipeline() {
}

func (_ *Input) AvailableArchs() []string {
	return []string{datakit.OSWindows}
}

func (w *Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{
		&Measurement{},
	}
}

func (w *Input) Run() {
}

func init() {
	inputs.Add(inputName, func() inputs.Input {
		s := &Input{}
		return s
	})
}
