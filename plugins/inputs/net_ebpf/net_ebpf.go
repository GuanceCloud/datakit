package net_ebpf

import (
	"runtime"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/external"
)

var (
	inputName   = "net_ebpf"
	catalogName = "host"
	l           = logger.DefaultSLogger("net_ebpf")
)

type Input struct {
	external.ExernalInput
}

func (i *Input) Run() {
	l = logger.SLogger(inputName)
	if runtime.GOOS == "linux" && runtime.GOARCH == "amd64" {
		i.ExernalInput.Run()
	} else {
		l.Error("net_ebpf not support ", runtime.GOOS, "/", runtime.GOARCH)
	}
}

func (i *Input) Catalog() string { return catalogName }

func (i *Input) SampleConfig() string { return configSample }

func (i *Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{
		&ConnStatsM{},
	}
}

func (i *Input) AvailableArchs() []string {
	return []string{datakit.OSArchLinuxAmd64}
}

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return &Input{}
	})
}
