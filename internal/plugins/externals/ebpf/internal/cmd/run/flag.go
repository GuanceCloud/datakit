// Package run implements datakit-ebpf run command
package run

import (
	"os"
	"strings"
)

var (
	EnvPrefix = "DKE_"

	EnvK8sURL = "K8S_URL"

	//nolint:gosec
	EnvBearerToken = "K8S_BEARER_TOKEN"
	//nolint:gosec
	EnvBearerTokenPath = "K8S_BEARER_TOKEN_PATH"

	EnvNetlogNetFilter = "NETLOG_NET_FILTER"
)

type Flag struct {
	DataKitAPIServer string `toml:"datakit_api"`
	PprofHost        string `toml:"pprof_host"`
	PprofPort        string `toml:"pprof_port"`

	HostName string   `toml:"hostname"`
	Service  string   `toml:"service"`
	Tags     []string `toml:"tags"`

	Interval string `toml:"interval"`

	Log      string `toml:"log"`
	LogLevel string `toml:"log_level"`

	PIDFile string `toml:"pidfile"`

	Enabled []string `toml:"enabled"`

	K8sInfo       FlagK8s       `toml:"k8s_info"`
	EBPFNet       FlagNet       `toml:"ebpf_net"`
	EBPFTrace     FlagTrace     `toml:"ebpf_trace"`
	BPFNetLog     FlagBPFNetLog `toml:"bpf_netlog"`
	ResourceLimit FlagResLimit  `toml:"resource_limit"`
}

type FlagNet struct {
	L7NetEnabled  []string `toml:"l7net_enabled"`
	L7NetDisabled []string `toml:"l7net_disabled"`

	EphemeralPort int32 `toml:"ephemeral_port"`
	IPv6Disabled  bool  `toml:"ipv6_diabled"`
}

type FlagBPFNetLog struct {
	EnableLog    bool   `toml:"enable_log"`
	EnableMetric bool   `toml:"enable_metric"`
	NetFilter    string `toml:"net_filter"`
}

type FlagTrace struct {
	TraceServer        string   `toml:"trace_server"`
	TraceAllProc       bool     `toml:"trace_all_proc"`
	TraceEnvList       []string `toml:"trace_env_list"`
	TraceNameList      []string `toml:"trace_name_list"`
	TraceEnvBlacklist  []string `toml:"trace_env_blacklist"`
	TraceNameBlacklist []string `toml:"trace_name_blacklist"`
	ConvTraceToDD      bool     `toml:"conv_trace_to_dd"`
}

type FlagK8s struct {
	URL             string `toml:"url"`
	BearerToken     string `toml:"bearer_token"`
	BearerTokenPath string `toml:"bearer_token_path"`
}

type FlagResLimit struct {
	LimitCPU       float64 `toml:"limit_cpu"`
	LimitMem       string  `toml:"limit_mem"`
	LimitBandwidth string  `toml:"limit_bandwidth"`
}

func readEnv(flag *Flag) {
	for _, env := range os.Environ() {
		i := strings.Index(env, "=")
		if i < 0 {
			continue
		}
		key := env[:i]
		key = strings.TrimSpace(key)
		if strings.HasPrefix(key, EnvPrefix) {
			key = strings.TrimPrefix(key, EnvPrefix)
		} else {
			continue
		}

		var v string
		if i+1 < len(env) {
			v = env[i+1:]
			v = strings.TrimSpace(v)
		}

		switch key {
		case EnvK8sURL:
			flag.K8sInfo.URL = v
		case EnvBearerToken:
			flag.K8sInfo.BearerToken = v
		case EnvBearerTokenPath:
			flag.K8sInfo.BearerTokenPath = v
		case EnvNetlogNetFilter:
			flag.BPFNetLog.NetFilter = v
		}
	}
}
