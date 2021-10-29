// Package netebpf wrap ebpf external input to collect eBPF-network metrics
package netebpf

import (
	"fmt"
	"regexp"
	"runtime"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
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

loop:
	for {
		select {
		case <-tick.C:
			// not linux/amd64
			if !(runtime.GOOS == "linux" && runtime.GOARCH == "amd64") {
				l.Error("unsupport OS/Arch")

				io.FeedLastError(inputName,
					fmt.Sprintf("net_ebpf not support %s/%s ",
						runtime.GOOS, runtime.GOARCH))
			}

			if ok, err := checkLinuxKernelVesion(""); err != nil || !ok {
				if err != nil {
					l.Errorf("checkLinuxKernelVesion: %s", err)
				}

				io.FeedLastError(inputName, err.Error())
			} else {
				break loop
			}
		case <-datakit.Exit.Wait():
			l.Infof("net_ebpf input exit")
			return
		}
	}

	l.Infof("net_ebpf input started")
	matchHost := regexp.MustCompile("--hostname")
	haveHostNameArg := false
	for _, arg := range i.ExernalInput.Args {
		haveHostNameArg = matchHost.MatchString(arg)
		if haveHostNameArg {
			break
		}
	}
	if !haveHostNameArg {
		if envHostname, ok := config.Cfg.Environments["ENV_HOSTNAME"]; ok && envHostname != "" {
			i.ExernalInput.Args = append(i.ExernalInput.Args, "--hostname", envHostname)
		}
	}

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

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return &Input{}
	})
}
