// +build !linux

package sensors

import (
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

type Input struct {
}

func (*Input) Catalog() string {
	return inputName
}

func (*Input) SampleConfig() string {
	return sampleConfig
}

func (*Input) AvailableArchs() []string {
	return []string{datakit.OSLinux}
}

func (*Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{&sensorsMeasurement{}}
}

func (s *Input) Run() {
	l.Errorf("Can not run %s on this system.", inputName)
}

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return &Input{}
	})
}
