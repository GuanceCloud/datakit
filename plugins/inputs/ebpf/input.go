// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

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

var _ inputs.ReadEnv = (*Input)(nil)

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
	external.ExternalInput
	K8sConf
	EnabledPlugins []string      `toml:"enabled_plugins"`
	L7NetDisabled  []string      `toml:"l7net_disabled"`
	L7NetEnabled   []string      `toml:"l7net_enabled"`
	IPv6Disabled   bool          `toml:"ipv6_disabled"`
	semStop        *cliutils.Sem // start stop signal
}

func (ipt *Input) Run() {
	l = logger.SLogger(inputName)
	tick := time.NewTicker(time.Second * 60)
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

		cmd := strings.Split(ipt.ExternalInput.Cmd, " ")
		var execFile string
		if len(cmd) > 0 {
			execFile = cmd[0]
		} else {
			execFile = filepath.Join(datakit.InstallDir, "externals", "datakit-ebpf")
			ipt.ExternalInput.Cmd = execFile
		}
		if _, err := os.Stat(execFile); err == nil && ok {
			break loop
		} else {
			l.Errorf("please run `datakit install --ebpf`")
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
	if ipt.ExternalInput.Args == nil {
		ipt.ExternalInput.Args = []string{}
	}
	if ipt.ExternalInput.Envs == nil {
		ipt.ExternalInput.Envs = []string{}
	}
	for _, arg := range ipt.ExternalInput.Args {
		haveHostNameArg = matchHost.MatchString(arg)
		if haveHostNameArg {
			break
		}
	}
	if !haveHostNameArg {
		ipt.ExternalInput.Args = append(ipt.ExternalInput.Args, "--hostname", config.Cfg.Hostname)
	}

	if ipt.K8sURL != "" {
		ipt.ExternalInput.Envs = append(ipt.ExternalInput.Envs,
			fmt.Sprintf("K8S_URL=%s", ipt.K8sConf.K8sURL))
	}
	if ipt.K8sBearerToken != "" {
		ipt.ExternalInput.Envs = append(ipt.ExternalInput.Envs,
			fmt.Sprintf("K8S_BEARER_TOKEN_PATH=%s", ipt.K8sConf.K8sBearerToken))
	}
	if ipt.K8sBearerTokenStr != "" {
		ipt.ExternalInput.Envs = append(ipt.ExternalInput.Envs,
			fmt.Sprintf("K8S_BEARER_TOKEN_STRING=%s", ipt.K8sConf.K8sBearerTokenStr))
	}

	if ipt.L7NetDisabled == nil && ipt.L7NetEnabled == nil {
		ipt.L7NetEnabled = []string{"httpflow"}
	}

	if len(ipt.L7NetEnabled) > 0 {
		ipt.ExternalInput.Args = append(ipt.ExternalInput.Args,
			"--l7net-enabled", strings.Join(ipt.L7NetEnabled, ","))
	} else if len(ipt.L7NetDisabled) > 0 {
		ipt.ExternalInput.Args = append(ipt.ExternalInput.Args,
			"--l7net-disabled", strings.Join(ipt.L7NetDisabled, ","))
	}

	if ipt.IPv6Disabled {
		ipt.ExternalInput.Args = append(ipt.ExternalInput.Args,
			"--ipv6-disabled", "true")
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
		ipt.ExternalInput.Args = append(ipt.ExternalInput.Args,
			"--enabled", strings.Join(enablePlugins, ","))
		l.Infof("ebpf input started")
		ipt.ExternalInput.Run()
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
	ipt.ExternalInput.Terminate()
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
	return []string{datakit.OSLabelLinux}
}

// ReadEnv support envs：
//   ENV_INPUT_EBPF_ENABLED_PLUGINS : []string
//   ENV_INPUT_EBPF_L7NET_ENABLED  : []string
//   ENV_INPUT_EBPF_IPV6_DISABLED   : bool
func (ipt *Input) ReadEnv(envs map[string]string) {
	if pluginList, ok := envs["ENV_INPUT_EBPF_ENABLED_PLUGINS"]; ok {
		l.Debugf("add enabled_plugins from ENV: %v", pluginList)
		ipt.EnabledPlugins = strings.Split(pluginList, ",")
	}

	if l7netEnabledList, ok := envs["ENV_INPUT_EBPF_L7NET_ENABLED"]; ok {
		l.Debugf("add l7net_enabled from ENV: %v", l7netEnabledList)
		ipt.L7NetEnabled = strings.Split(l7netEnabledList, ",")
	}

	if v, ok := envs["ENV_INPUT_EBPF_IPV6_DISABLED"]; ok {
		switch v {
		case "", "f", "false", "FALSE", "False", "0":
			ipt.IPv6Disabled = false
		default:
			ipt.IPv6Disabled = true
		}
	}
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return &Input{
			semStop:        cliutils.NewSem(),
			EnabledPlugins: []string{},
			ExternalInput:  *external.NewExternalInput(),
		}
	})
}
