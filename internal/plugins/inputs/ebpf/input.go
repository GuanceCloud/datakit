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
	"strconv"
	"strings"
	"time"

	"github.com/GuanceCloud/cliutils"
	"github.com/GuanceCloud/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/external"
)

var (
	_ inputs.ReadEnv   = (*Input)(nil)
	_ inputs.Singleton = (*Input)(nil)
)

var (
	inputName           = "ebpf"
	catalogName         = "host"
	l                   = logger.DefaultSLogger("ebpf")
	AllSupportedPlugins = map[string]bool{
		"ebpf-bash":      true,
		"ebpf-net":       true,
		"ebpf-conntrack": true,
		"ebpf-trace":     true,
	}
)

type K8sConf struct {
	K8sURL            string `toml:"kubernetes_url"`
	K8sBearerToken    string `toml:"bearer_token"`
	K8sBearerTokenStr string `toml:"bearer_token_string"`
}

type Input struct {
	external.Input
	K8sConf
	EnabledPlugins []string `toml:"enabled_plugins"`
	L7NetDisabled  []string `toml:"l7net_disabled"`
	L7NetEnabled   []string `toml:"l7net_enabled"`

	TraceServer     string `toml:"trace_server"`
	Conv2DD         bool   `toml:"conv_to_ddtrace"`
	TraceAllProcess bool   `toml:"trace_all_process"`

	TraceENVList       []string `toml:"trace_env_list"`
	TraceENVBlacklist  []string `toml:"trace_env_blacklist"`
	TraceNameList      []string `toml:"trace_name_list"`
	TraceNameBlacklist []string `toml:"trace_name_blacklist"`

	IPv6Disabled  bool          `toml:"ipv6_disabled"`
	EphemeralPort int32         `toml:"ephemeral_port"`
	Interval      string        `toml:"interval"`
	semStop       *cliutils.Sem // start stop signal
}

func (ipt *Input) Singleton() {
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

		cmd := strings.Split(ipt.Input.Cmd, " ")
		var execFile string
		if len(cmd) > 0 {
			execFile = cmd[0]
		} else {
			execFile = filepath.Join(datakit.InstallDir, "externals", "datakit-ebpf")
			ipt.Input.Cmd = execFile
		}
		if _, err := os.Stat(execFile); err == nil {
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
	if ipt.Input.Args == nil {
		ipt.Input.Args = []string{}
	}
	if ipt.Input.Envs == nil {
		ipt.Input.Envs = []string{}
	}
	for _, arg := range ipt.Input.Args {
		haveHostNameArg = matchHost.MatchString(arg)
		if haveHostNameArg {
			break
		}
	}
	if !haveHostNameArg {
		ipt.Input.Args = append(ipt.Input.Args, "--hostname", config.Cfg.Hostname)
	}

	if ipt.K8sURL != "" {
		ipt.Input.Envs = append(ipt.Input.Envs,
			fmt.Sprintf("K8S_URL=%s", ipt.K8sConf.K8sURL))
	}
	if ipt.K8sBearerToken != "" {
		ipt.Input.Envs = append(ipt.Input.Envs,
			fmt.Sprintf("K8S_BEARER_TOKEN_PATH=%s", ipt.K8sConf.K8sBearerToken))
	}
	if ipt.K8sBearerTokenStr != "" {
		ipt.Input.Envs = append(ipt.Input.Envs,
			fmt.Sprintf("K8S_BEARER_TOKEN_STRING=%s", ipt.K8sConf.K8sBearerTokenStr))
	}

	if ipt.L7NetDisabled == nil && ipt.L7NetEnabled == nil {
		ipt.L7NetEnabled = []string{"httpflow"}
	}

	if len(ipt.L7NetEnabled) > 0 {
		ipt.Input.Args = append(ipt.Input.Args,
			"--l7net-enabled", strings.Join(ipt.L7NetEnabled, ","))
	} else if len(ipt.L7NetDisabled) > 0 {
		ipt.Input.Args = append(ipt.Input.Args,
			"--l7net-disabled", strings.Join(ipt.L7NetDisabled, ","))
	}

	if ipt.IPv6Disabled {
		ipt.Input.Args = append(ipt.Input.Args,
			"--ipv6-disabled", "true")
	}

	if ipt.EphemeralPort >= 0 {
		ipt.Input.Args = append(ipt.Input.Args,
			"--ephemeral_port", strconv.FormatInt(int64(ipt.EphemeralPort), 10))
	}

	if ipt.Interval != "" {
		ipt.Input.Args = append(ipt.Input.Args,
			"--interval", ipt.Interval)
	}

	if ipt.TraceServer != "" {
		ipt.Input.Args = append(ipt.Input.Args,
			"--trace-server", ipt.TraceServer)
	}

	if ipt.TraceAllProcess {
		ipt.Input.Args = append(ipt.Input.Args,
			"--trace-allprocess", "true")
	}
	if len(ipt.TraceENVList) > 0 {
		ipt.Input.Args = append(ipt.Input.Args,
			"--trace-env-list", strings.Join(ipt.TraceENVList, ","))
	}
	if len(ipt.TraceENVBlacklist) > 0 {
		ipt.Input.Args = append(ipt.Input.Args,
			"--trace-env-blacklist", strings.Join(ipt.TraceENVBlacklist, ","))
	}

	if len(ipt.TraceNameList) > 0 {
		ipt.Input.Args = append(ipt.Input.Args,
			"--trace-name-list", strings.Join(ipt.TraceNameList, ","))
	}
	if len(ipt.TraceNameBlacklist) > 0 {
		ipt.Input.Args = append(ipt.Input.Args,
			"--trace-name-blacklist", strings.Join(ipt.TraceNameBlacklist, ","))
	}

	if ipt.Conv2DD {
		ipt.Input.Args = append(ipt.Input.Args,
			"--conv-to-ddtrace", "true")
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
		ipt.Input.Args = append(ipt.Input.Args,
			"--enabled", strings.Join(enablePlugins, ","))
		l.Infof("ebpf input started")
		ipt.Input.Run()
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
	ipt.Input.Terminate()
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
	return []string{datakit.OSLabelLinux, datakit.LabelK8s}
}

// ReadEnv support envs：
//
// ENV_INPUT_EBPF_ENABLED_PLUGINS : []string
// ENV_INPUT_EBPF_L7NET_ENABLED   : []string
// ENV_INPUT_EBPF_IPV6_DISABLED   : bool
// ENV_INPUT_EBPF_EPHEMERAL_PORT  : int32
// ENV_INPUT_EBPF_INTERVAL        : string
//
// ENV_INPUT_EBPF_TRACE_ALL_PROCESS    : bool
// ENV_INPUT_EBPF_CONV_TO_DDTRACE      : bool
// ENV_INPUT_EBPF_TRACE_SERVER         : string
// ENV_INPUT_EBPF_TRACE_ENV_LIST       : string
// ENV_INPUT_EBPF_TRACE_ENV_BLACKLIST  : string
// ENV_INPUT_EBPF_TRACE_NAME_LIST      : string
// ENV_INPUT_EBPF_TRACE_NAME_BLACKLIST : string.
func (ipt *Input) ReadEnv(envs map[string]string) {
	if pluginList, ok := envs["ENV_INPUT_EBPF_ENABLED_PLUGINS"]; ok {
		l.Debugf("add enabled_plugins from ENV: %v", pluginList)
		ipt.EnabledPlugins = strings.Split(pluginList, ",")
	}

	if l7netEnabledList, ok := envs["ENV_INPUT_EBPF_L7NET_ENABLED"]; ok {
		l.Debugf("add l7net_enabled from ENV: %v", l7netEnabledList)
		ipt.L7NetEnabled = strings.Split(l7netEnabledList, ",")
	}

	if envset, ok := envs["ENV_INPUT_EBPF_TRACE_ENV_LIST"]; ok {
		l.Debug("add env list from ENV: %v", envset)
		ipt.TraceENVList = strings.Split(envset, ",")
	}

	if envBlack, ok := envs["ENV_INPUT_EBPF_TRACE_ENV_BLACKLIST"]; ok {
		l.Debug("add env blacklist from ENV: %v", envBlack)
		ipt.TraceENVBlacklist = strings.Split(envBlack, ",")
	}

	if nameset, ok := envs["ENV_INPUT_EBPF_TRACE_NAME_LIST"]; ok {
		l.Debug("add process name list from ENV: %v", nameset)
		ipt.TraceNameList = strings.Split(nameset, ",")
	}

	if nameBlack, ok := envs["ENV_INPUT_EBPF_TRACE_NAME_BLACKLIST"]; ok {
		l.Debug("add process name blacklist from ENV: %v", nameBlack)
		ipt.TraceNameBlacklist = strings.Split(nameBlack, ",")
	}

	if v, ok := envs["ENV_INPUT_EBPF_IPV6_DISABLED"]; ok {
		switch v {
		case "", "f", "false", "FALSE", "False", "0":
			ipt.IPv6Disabled = false
		default:
			ipt.IPv6Disabled = true
		}
	}

	if v, ok := envs["ENV_INPUT_EBPF_TRACE_ALL_PROCESS"]; ok {
		switch v {
		case "", "f", "false", "FALSE", "False", "0":
			ipt.TraceAllProcess = false
		default:
			ipt.TraceAllProcess = true
		}
	}

	if v, ok := envs["ENV_INPUT_EBPF_EPHEMERAL_PORT"]; ok {
		if p, err := strconv.ParseInt(v, 10, 32); err != nil {
			l.Warn("parse ENV_INPUT_EBPF_EPHEMERAL_PORT: %w", err)
		} else {
			ipt.EphemeralPort = int32(p)
		}
	}

	if v, ok := envs["ENV_INPUT_EBPF_INTERVAL"]; ok {
		ipt.Interval = v
	}

	if v, ok := envs["ENV_INPUT_EBPF_TRACE_SERVER"]; ok {
		ipt.TraceServer = v
	}

	if v, ok := envs["ENV_INPUT_EBPF_CONV_TO_DDTRACE"]; ok {
		switch v {
		case "", "f", "false", "FALSE", "False", "0":
			ipt.Conv2DD = false
		default:
			ipt.Conv2DD = true
		}
	}
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		ret := &Input{
			semStop:        cliutils.NewSem(),
			EnabledPlugins: []string{},
			Input:          *external.NewInput(),
			EphemeralPort:  -1,
		}
		ret.Input.Election = false
		return ret
	})
}
