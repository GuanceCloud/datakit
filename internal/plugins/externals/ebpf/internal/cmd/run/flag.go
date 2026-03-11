//go:build linux
// +build linux

// Package run implements datakit-ebpf run command
package run

import (
	"crypto/tls"
	"net/http"
	"os"
	"strings"
	"time"

	k8sclient "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/kubernetes/client"
	k8scli "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/pkg/cli"
)

var (
	EnvPrefix = "DKE_"

	EnvK8sURL = "K8S_URL"

	//nolint:gosec
	EnvBearerToken = "K8S_BEARER_TOKEN"
	//nolint:gosec
	EnvBearerTokenPath = "K8S_BEARER_TOKEN_PATH"
	EnvKubeConfig      = "K8S_KUBECONFIG"
	EnvOperatorURL     = "K8S_OPERATOR_URL"
	EnvK8sOperatorOnly = "K8S_OPERATOR_ONLY"

	EnvWorkloadLabels      = "K8S_WORKLOAD_LABELS"
	EnvWorkloadLabelPrefix = "K8S_WORKLOAD_LABEL_PREFIX"

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

	K8sInfo       k8scli.K8sConfig `toml:"k8s_info"`
	ContainerInfo FlagContainer    `toml:"container_info"`
	EBPFNet       FlagNet          `toml:"ebpf_net"`
	EBPFTrace     FlagTrace        `toml:"ebpf_trace"`
	BPFNetLog     FlagBPFNetLog    `toml:"bpf_netlog"`
	ResourceLimit FlagResLimit     `toml:"resource_limit"`

	Sampling FlagSampling `toml:"sampling"`
}

type FlagSampling struct {
	Rate             string `toml:"rate"`
	RatePtsPerMinute string `toml:"rate_pts_per_min"`
}

type FlagNet struct {
	L7NetEnabled  []string `toml:"l7net_enabled"`
	L7NetDisabled []string `toml:"l7net_disabled"`

	EphemeralPort int32 `toml:"ephemeral_port"`
	IPv6Disabled  bool  `toml:"ipv6_diabled"`
}

type FlagBPFNetLog struct {
	EnableLog      bool     `toml:"enable_log"`
	EnableMetric   bool     `toml:"enable_metric"`
	L7LogProtocols []string `toml:"l7log_protocols"`
	NetFilter      string   `toml:"net_filter"`
}

type FlagTrace struct {
	TraceServer         string   `toml:"trace_server"`
	TraceAllProc        bool     `toml:"trace_all_proc"`
	TraceEnvList        []string `toml:"trace_env_list"`
	TraceNameList       []string `toml:"trace_name_list"`
	TraceProtoList      []string `toml:"trace_proto_list"`
	TraceEnvBlacklist   []string `toml:"trace_env_blacklist"`
	TraceNameBlacklist  []string `toml:"trace_name_blacklist"`
	TraceProtoBlacklist []string `toml:"trace_proto_blacklist"`
	ConvTraceToDD       bool     `toml:"conv_trace_to_dd"`
}

type FlagContainer struct {
	Endpoints []string `toml:"endpoints"`
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
		case EnvKubeConfig:
			flag.K8sInfo.KubeConfig = v
		case EnvOperatorURL:
			flag.K8sInfo.OperatorURL = strings.TrimSpace(v)
		case EnvWorkloadLabels:
			s := strings.Split(v, ",")
			for i := range s {
				s[i] = strings.TrimSpace(s[i])
			}
			flag.K8sInfo.WorkloadLabels = s
		case EnvWorkloadLabelPrefix:
			flag.K8sInfo.WorkloadLabelPrefix = v
		case EnvNetlogNetFilter:
			flag.BPFNetLog.NetFilter = v
		}
	}
}

const operatorProbeTimeout = 2 * time.Second

// probeOperatorURL 检测 operator 地址是否可达（短超时），用于未配置 operator_url 时尝试默认地址。
func probeOperatorURL(baseURL string) bool {
	if baseURL == "" {
		return false
	}
	url := strings.TrimSuffix(baseURL, "/") + "/v1/cluster/api/v1/pods"
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return false
	}
	client := &http.Client{
		Timeout: operatorProbeTimeout,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // nolint:gosec
		},
	}
	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	_ = resp.Body.Close()

	// 2xx 表示服务存在
	return resp.StatusCode >= 200 && resp.StatusCode < 300
}

// ApplyDefaultOperatorURLIfReachable 若未配置 operator_url，则探测 DefaultOperatorBaseURL，可达则设为默认。
func ApplyDefaultOperatorURLIfReachable(flag *Flag) {
	if flag.K8sInfo.OperatorURL != "" {
		return
	}

	if probeOperatorURL(k8sclient.DefaultOperatorBaseURL) {
		flag.K8sInfo.OperatorURL = k8sclient.DefaultOperatorBaseURL
	} else {
		log.Warn("Default operator address is not reachable, please set operator_url manually")
	}
}
