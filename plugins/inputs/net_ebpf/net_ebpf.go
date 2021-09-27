package net_ebpf

import (
	"fmt"
	"runtime"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
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
	tick := time.NewTicker(time.Second * 60)
	defer tick.Stop()
	var err error

loop:
	for {
		select {
		case <-tick.C:
			// not linux/amd64
			if !(runtime.GOOS == "linux" && runtime.GOARCH == "amd64") {
				err = fmt.Errorf("net_ebpf not support %s/%s ", runtime.GOOS, runtime.GOARCH)
				io.FeedLastError(inputName, err.Error())
				l.Error(err)
			}
			if ok, err := checkLinuxKernelVesion(""); err != nil || !ok {
				io.FeedLastError(inputName, err.Error())
				l.Error(err)
			} else {
				break loop
			}
		case <-datakit.Exit.Wait():
			l.Infof("net_ebpf input exit")
			return
		}
	}

	l.Infof("net_ebpf input started")
	i.ExernalInput.Run()
	l.Infof("net_ebpf input exit")
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
