// Package ebpf wrap ebpf external input to collect eBPF metrics
package ebpf

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/shirou/gopsutil/host"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/external"
)

var (
	inputName           = "ebpf"
	catalogName         = "host"
	l                   = logger.DefaultSLogger("ebpf")
	AllSupportedPlugins = map[string]bool{
		"ebpf-bash": true,
		"ebpf-net":  true,
	}
)

type K8sConf struct {
	K8sURL            string `toml:"kubernetes_url"`
	K8sBearerToken    string `toml:"bearer_token"`
	K8sBearerTokenStr string `toml:"bearer_token_string"`
}

type Input struct {
	external.ExernalInput
	K8sConf
	EnabledPlugins []string      `toml:"enabled_plugins"`
	semStop        *cliutils.Sem // start stop signal
}

func (ipt *Input) Run() {
	l = logger.SLogger(inputName)
	tick := time.NewTicker(time.Second * 60)
	io.FeedEventLog(&io.Reporter{Message: "ebpf start ok, ready for collecting metrics.", Logtype: "event"})
	defer tick.Stop()

loop:
	for {
		// not linux/amd64 or linux/arm64
		if !(runtime.GOOS == "linux" && (runtime.GOARCH == "amd64" || runtime.GOARCH == "arm64")) {
			l.Error("unsupport OS/Arch")

			io.FeedLastError(inputName,
				fmt.Sprintf("ebpf not support %s/%s ",
					runtime.GOOS, runtime.GOARCH))
		}

		ok, err := checkLinuxKernelVesion("")
		if err != nil || !ok {
			if err != nil {
				if p, _, v, err := host.PlatformInformation(); err == nil {
					if checkIsCentos76Ubuntu1604(p, v) {
						break loop
					}
				}
				l.Errorf("checkLinuxKernelVesion: %s", err)
			}
			io.FeedLastError(inputName, err.Error())
		}

		cmd := strings.Split(ipt.ExernalInput.Cmd, " ")
		var execFile string
		if len(cmd) > 0 {
			execFile = cmd[0]
		} else {
			execFile = filepath.Join(datakit.InstallDir, "externals", "datakit-ebpf")
			ipt.ExernalInput.Cmd = execFile
		}
		if _, err := os.Stat(execFile); err == nil && ok {
			break loop
		} else {
			l.Errorf("please run `datakit install --datakit-ebpf`")
		}

		select {
		case <-tick.C:
		case <-datakit.Exit.Wait():
			l.Info("ebpf input exit")
			return

		case <-ipt.semStop.Wait():
			l.Info("ebpf input return")
			return
		}
	}

	matchHost := regexp.MustCompile("--hostname")
	haveHostNameArg := false
	if ipt.ExernalInput.Args == nil {
		ipt.ExernalInput.Args = []string{}
	}
	if ipt.ExernalInput.Envs == nil {
		ipt.ExernalInput.Envs = []string{}
	}
	for _, arg := range ipt.ExernalInput.Args {
		haveHostNameArg = matchHost.MatchString(arg)
		if haveHostNameArg {
			break
		}
	}
	if !haveHostNameArg {
		if envHostname, ok := config.Cfg.Environments["ENV_HOSTNAME"]; ok && envHostname != "" {
			ipt.ExernalInput.Args = append(ipt.ExernalInput.Args, "--hostname", envHostname)
		}
	}

	if ipt.K8sURL != "" {
		ipt.Envs = append(ipt.Envs,
			fmt.Sprintf("K8S_URL=%s", ipt.K8sURL))
	}
	if ipt.K8sBearerToken != "" {
		ipt.Envs = append(ipt.Envs,
			fmt.Sprintf("K8S_BEARER_TOKEN_PATH=%s", ipt.K8sBearerToken))
	}
	if ipt.K8sBearerTokenStr != "" {
		ipt.Envs = append(ipt.Envs,
			fmt.Sprintf("K8S_BEARER_TOKEN_STRING=%s", ipt.K8sBearerTokenStr))
	}

	if len(ipt.EnabledPlugins) == 0 {
		ipt.EnabledPlugins = []string{"ebpf-net"}
	}

	enablePlugins := []string{}
	for _, nameP := range ipt.EnabledPlugins {
		if v, ok := AllSupportedPlugins[nameP]; ok && v {
			enablePlugins = append(enablePlugins, nameP)
		}
	}
	if len(enablePlugins) > 0 {
		ipt.ExernalInput.Args = append(ipt.ExernalInput.Args,
			"--enabled", strings.Join(enablePlugins, ","))
		l.Infof("ebpf input started")
		ipt.ExernalInput.Run()
	} else {
		l.Warn("no ebpf plugins enabled")
		io.FeedLastError(inputName, "no ebpf plugins enabled")
	}
	l.Infof("ebpf input exit")
}

func (ipt *Input) Terminate() {
	if ipt.semStop != nil {
		ipt.semStop.Close()
	}
	ipt.ExernalInput.Terminate()
}

func (*Input) Catalog() string { return catalogName }

func (*Input) SampleConfig() string { return configSample }

func (*Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{
		&ConnStatsM{},
		&DNSStatsM{},
		&BashM{},
		&HTTPFlowM{},
	}
}

func (*Input) AvailableArchs() []string {
	return []string{datakit.OSArchLinuxAmd64}
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return &Input{
			semStop:        cliutils.NewSem(),
			EnabledPlugins: []string{},
			ExernalInput:   *external.NewExternalInput(),
		}
	})
}
