// +build !linux

package sensors

import (
	"runtime"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

// redefine here for sample checking on non-linux platform
type Input struct {
	Path     string            `toml:"path"`
	Interval datakit.Duration  `toml:"interval"`
	Timeout  datakit.Duration  `toml:"timeout"`
	Tags     map[string]string `toml:"tags"`
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
	l.Errorf("Can not run input %q on %s-%s.", inputName, runtime.GOOS, runtime.GOARCH)
}

func init() {
	inputs.Add(inputName, func() inputs.Input { return &Input{} })
}
