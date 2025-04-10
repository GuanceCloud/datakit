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
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/metrics"
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
		"bpf-netlog":     true,
	}
)

type K8sConf struct {
	URL            string `toml:"kubernetes_url"`
	BearerToken    string `toml:"bearer_token"`
	BearerTokenStr string `toml:"bearer_token_string"`

	WorkloadLabels      []string `toml:"workload_labels"`
	WorkloadLabelPrefix string   `toml:"workload_label_prefix"`
}

type Input struct {
	external.Input
	K8sConf

	NetlogBlacklist  string `toml:"netlog_blacklist"`
	NetlogMetricOnly bool   `toml:"netlog_metric_only"`
	NetlogMetric     bool   `toml:"netlog_metric"`
	NetlogLog        bool   `toml:"netlog_log"`

	EnabledPlugins []string `toml:"enabled_plugins"`
	L7NetDisabled  []string `toml:"l7net_disabled"`
	L7NetEnabled   []string `toml:"l7net_enabled"`

	TraceServer     string `toml:"trace_server"`
	Conv2DD         bool   `toml:"conv_to_ddtrace"`
	TraceAllProcess bool   `toml:"trace_all_process"`

	PprofHost string `toml:"pprof_host"`
	PprofPort string `toml:"pprof_port"`

	CPULimit string `toml:"cpu_limit"`
	MemLimit string `toml:"mem_limit"`
	NetLimit string `toml:"net_limit"`

	TraceENVList       []string `toml:"trace_env_list"`
	TraceENVBlacklist  []string `toml:"trace_env_blacklist"`
	TraceNameList      []string `toml:"trace_name_list"`
	TraceNameBlacklist []string `toml:"trace_name_blacklist"`

	IPv6Disabled          bool   `toml:"ipv6_disabled"`
	EphemeralPort         int32  `toml:"ephemeral_port"`
	Interval              string `toml:"interval"`
	SamplingRate          string `toml:"sampling_rate"`
	SamplingRatePtsPerMin string `toml:"sampling_rate_pts_per_min"`

	semStop *cliutils.Sem // start stop signal
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

			metrics.FeedLastError(inputName,
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

	if ipt.URL != "" {
		ipt.Input.Envs = append(ipt.Input.Envs,
			fmt.Sprintf("DKE_K8S_URL=%s", ipt.K8sConf.URL))
	}
	if ipt.BearerToken != "" {
		ipt.Input.Envs = append(ipt.Input.Envs,
			fmt.Sprintf("DKE_K8S_BEARER_TOKEN_PATH=%s", ipt.K8sConf.BearerToken))
	}
	if ipt.BearerTokenStr != "" {
		ipt.Input.Envs = append(ipt.Input.Envs,
			fmt.Sprintf("DKE_K8S_BEARER_TOKEN=%s", ipt.K8sConf.BearerTokenStr))
	}
	if len(ipt.WorkloadLabels) > 0 {
		ipt.Input.Envs = append(
			ipt.Input.Envs,
			fmt.Sprintf("DKE_K8S_WORKLOAD_LABELS=%s",
				strings.Join(ipt.WorkloadLabels, ","),
			),
		)
		if ipt.WorkloadLabelPrefix != "" {
			ipt.Input.Envs = append(
				ipt.Input.Envs,
				fmt.Sprintf("K8S_WORKLOAD_LABEL_PREFIX=%s",
					ipt.WorkloadLabelPrefix),
			)
		}
	}
	if ipt.NetlogBlacklist != "" {
		ipt.Input.Envs = append(ipt.Input.Envs,
			fmt.Sprintf("DKE_NETLOG_NET_FILTER=%s", ipt.NetlogBlacklist))
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

	if ipt.PprofHost != "" {
		ipt.Input.Args = append(ipt.Input.Args,
			"--pprof-host", ipt.PprofHost)
	}

	if ipt.PprofPort != "" {
		ipt.Input.Args = append(ipt.Input.Args,
			"--pprof-port", ipt.PprofPort)
	}

	if ipt.IPv6Disabled {
		ipt.Input.Args = append(ipt.Input.Args,
			"--ipv6-disabled")
	}

	if ipt.EphemeralPort >= 0 {
		ipt.Input.Args = append(ipt.Input.Args,
			"--ephemeral_port", strconv.FormatInt(int64(ipt.EphemeralPort), 10))
	}

	if ipt.Interval != "" {
		ipt.Input.Args = append(ipt.Input.Args,
			"--interval", ipt.Interval)
	}

	{
		var netlogArgs []string
		if ipt.NetlogMetric {
			netlogArgs = append(netlogArgs, "--netlog-metric")
		}
		if !ipt.NetlogMetricOnly || ipt.NetlogLog {
			netlogArgs = append(netlogArgs, "--netlog-log")
		}

		ipt.Input.Args = append(ipt.Input.Args, netlogArgs...)
	}

	if ipt.TraceServer != "" {
		ipt.Input.Args = append(ipt.Input.Args,
			"--trace-server", ipt.TraceServer)
	}

	if ipt.TraceAllProcess {
		ipt.Input.Args = append(ipt.Input.Args,
			"--trace-allprocess")
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
			"--conv-to-ddtrace")
	}

	if ipt.CPULimit != "" {
		ipt.Input.Args = append(ipt.Input.Args,
			"--res-cpu", ipt.CPULimit)
	}

	if ipt.MemLimit != "" {
		ipt.Input.Args = append(ipt.Input.Args,
			"--res-mem", ipt.MemLimit)
	}

	if ipt.NetLimit != "" {
		ipt.Input.Args = append(ipt.Input.Args,
			"--res-net", ipt.NetLimit)
	}

	if ipt.SamplingRate != "" {
		ipt.Input.Args = append(ipt.Input.Args,
			"--sampling-rate", ipt.SamplingRate)
	} else if ipt.SamplingRatePtsPerMin != "" {
		ipt.Input.Args = append(ipt.Input.Args,
			"--sampling-rate-ptsperminute", ipt.SamplingRatePtsPerMin)
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
		ipt.Input.Args = append([]string{"run"}, ipt.Input.Args...)
		l.Infof("ebpf input started")
		ipt.Input.Run()
	} else {
		l.Warn("no ebpf plugins enabled")
		metrics.FeedLastError(inputName, "no ebpf plugins enabled")
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
		&BPFL4Log{},
		&BPFL7Log{},
		&EBPFTrace{},
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
// ENV_INPUT_EBPF_PPROF_HOST      : string
// ENV_INPUT_EBPF_PPROF_PORT      : string
//
// ENV_INPUT_EBPF_NETLOG_BLACKLIST   : string
// ENV_INPUT_EBPF_NETLOG_METRIC_ONLY : bool
// ENV_INPUT_EBPF_NETLOG_METRIC      : bool
// ENV_INPUT_EBPF_NETLOG_LOG         : bool
//
// ENV_INPUT_EBPF_CPU_LIMIT : string
// ENV_INPUT_EBPF_MEM_LIMIT : string
// ENV_INPUT_EBPF_NET_LIMIT : string
//
// ENV_INPUT_EBPF_TRACE_ALL_PROCESS    : bool
// ENV_INPUT_EBPF_CONV_TO_DDTRACE      : bool
// ENV_INPUT_EBPF_TRACE_SERVER         : string
// ENV_INPUT_EBPF_TRACE_ENV_LIST       : string
// ENV_INPUT_EBPF_TRACE_ENV_BLACKLIST  : string
// ENV_INPUT_EBPF_TRACE_NAME_LIST      : string
// ENV_INPUT_EBPF_TRACE_NAME_BLACKLIST : string
//
// ENV_INPUT_EBPF_SAMPLING_RATE           : string
// ENV_INPUT_EBPF_SAMPLING_RATE_PTSPERMIN : string.
//
// ENV_INPUT_EBPF_WORKLOAD_LABELS      : []string
// ENV_INPUT_EBPF_WORKLOAD_LABEL_PREFIX: string.
func (ipt *Input) ReadEnv(envs map[string]string) {
	if v, ok := envs["ENV_INPUT_EBPF_PPROF_HOST"]; ok {
		ipt.PprofHost = v
	}
	if v, ok := envs["ENV_INPUT_EBPF_PPROF_PORT"]; ok {
		ipt.PprofPort = v
	}

	if pluginList, ok := envs["ENV_INPUT_EBPF_ENABLED_PLUGINS"]; ok {
		l.Debugf("add enabled_plugins from ENV: %v", pluginList)
		ipt.EnabledPlugins = strings.Split(pluginList, ",")
	}

	if v, ok := envs["ENV_INPUT_EBPF_WORKLOAD_LABELS"]; ok {
		ipt.WorkloadLabels = strings.Split(v, ",")
	}

	if v, ok := envs["ENV_INPUT_EBPF_WORKLOAD_LABEL_PREFIX"]; ok {
		ipt.WorkloadLabelPrefix = v
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
		case "", "f", "false", "FALSE", "False", "0": //nolint:goconst
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

	if v, ok := envs["ENV_NETLOG_BLACKLIST"]; ok {
		ipt.NetlogBlacklist = v
	}

	if v, ok := envs["ENV_INPUT_EBPF_NETLOG_BLACKLIST"]; ok {
		ipt.NetlogBlacklist = v
	}

	if v, ok := envs["ENV_NETLOG_METRIC_ONLY"]; ok {
		switch v {
		case "", "f", "false", "FALSE", "False", "0":
			ipt.NetlogMetricOnly = false
		default:
			ipt.NetlogMetricOnly = true
		}
	}

	if v, ok := envs["ENV_INPUT_EBPF_NETLOG_METRIC_ONLY"]; ok {
		switch v {
		case "", "f", "false", "FALSE", "False", "0":
			ipt.NetlogMetricOnly = false
		default:
			ipt.NetlogMetricOnly = true
		}
	}

	if v, ok := envs["ENV_NETLOG_METRIC"]; ok {
		switch v {
		case "", "f", "false", "FALSE", "False", "0":
			ipt.NetlogMetric = false
		default:
			ipt.NetlogMetric = true
		}
	}

	if v, ok := envs["ENV_INPUT_EBPF_NETLOG_METRIC"]; ok {
		switch v {
		case "", "f", "false", "FALSE", "False", "0":
			ipt.NetlogMetric = false
		default:
			ipt.NetlogMetric = true
		}
	}

	if v, ok := envs["ENV_NETLOG_LOG"]; ok {
		switch v {
		case "", "f", "false", "FALSE", "False", "0":
			ipt.NetlogLog = false
		default:
			ipt.NetlogLog = true
		}
	}

	if v, ok := envs["ENV_INPUT_EBPF_NETLOG_LOG"]; ok {
		switch v {
		case "", "f", "false", "FALSE", "False", "0":
			ipt.NetlogLog = false
		default:
			ipt.NetlogLog = true
		}
	}

	if v, ok := envs["ENV_INPUT_EBPF_CPU_LIMIT"]; ok {
		ipt.CPULimit = v
	}
	if v, ok := envs["ENV_INPUT_EBPF_MEM_LIMIT"]; ok {
		ipt.MemLimit = v
	}
	if v, ok := envs["ENV_INPUT_EBPF_NET_LIMIT"]; ok {
		ipt.NetLimit = v
	}

	// ENV_INPUT_EBPF_SAMPLING_RATE           : string
	// ENV_INPUT_EBPF_SAMPLING_RATE_PTSPERMIN : string
	if v, ok := envs["ENV_INPUT_EBPF_SAMPLING_RATE"]; ok {
		ipt.SamplingRate = v
	}

	if v, ok := envs["ENV_INPUT_EBPF_SAMPLING_RATE_PTSPERMIN"]; ok {
		ipt.SamplingRatePtsPerMin = v
	}
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		ret := &Input{
			semStop:          cliutils.NewSem(),
			EnabledPlugins:   []string{},
			Input:            *external.NewInput(),
			EphemeralPort:    -1,
			NetlogMetric:     true,
			NetlogLog:        false,
			NetlogMetricOnly: true,
		}
		ret.Input.Election = false
		return ret
	})
}
