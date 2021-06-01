package docker

import (
	"fmt"
	"net/url"
	"path/filepath"
	"regexp"
	"sync"
	"time"

	"github.com/docker/docker/api/types"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline"
)

// ## To use environment variables (ie, docker-machine), set endpoint = "ENV"

const (
	inputName = "docker"

	sampleCfg = `
[inputs.docker]
    # Docker server host
    # To use TCP, set endpoint = "tcp://[ip]:[port]"
    endpoint = "unix:///var/run/docker.sock"

    collect_metric = false  # enable metric collect
    collect_object = true   # enable object collect
    collect_logging = true  # enable logging collect

    collect_metric_interval = "10s"

    # If enabled, collect exited container info
    include_exited = false

    ## param type: string - optional: TLS Config
    # tls_ca = "/path/to/ca.pem"
    # tls_cert = "/path/to/cert.pem"
    # tls_key = "/path/to/key.pem"
    ## param type: boolean - optional: Use TLS but skip chain & host verification
    # insecure_skip_verify = false

    ## Logging filter(if collect_logging enabled)
    #[[inputs.docker.logfilter]]
        # filter_message = [
        #   '''<this-is-message-regexp''',
        #   '''<this-is-another-message-regexp''',
        # ]

        # source = "<your-source-name>"
        # service = "<your-service-name>"
        # pipeline = "<pipeline.p>"

    [inputs.docker.tags]
    # some_tag = "some_value"
    # more_tag = "some_other_value"
`
	defaultEndpoint     = "unix:///var/run/docker.sock"
	defaultEndpointPath = "/var/run/docker.sock"
	// Docker API 超时时间
	defaultAPITimeout = time.Second * 5
	// 最小指标采集间隔
	minCollectMetricDuration = time.Second * 5
	// 最大指标采集间隔
	maxCollectMetricDuration = time.Second * 60
	// 对象采集间隔
	collectObjectDuration = time.Minute * 5
	// 定时发现新日志源
	loggingHitDuration = time.Second * 5
)

var l = logger.DefaultSLogger(inputName)

type DeprecatedLogOption struct {
	NameMatch string `toml:"container_name_match"`
	Source    string `toml:"source"`
	Service   string `toml:"service"`
	Pipeline  string `toml:"pipeline"`
}

type LogFilters []*LogFilter

type LogFilter struct {
	FilterMessage []string `toml:"filter_message"`
	// # filter_mutltiline = '''^\s+'''
	FilterMultiline string `toml:"-"`
	Source          string `toml:"source"`
	Service         string `toml:"service"`
	Pipeline        string `toml:"pipeline"`

	pipelinePool sync.Pool

	multilinePattern *regexp.Regexp
	messagePattern   []*regexp.Regexp
}

func (lf *LogFilter) Init() error {
	if lf.Service == "" {
		lf.Service = lf.Source
	}

	if lf.FilterMultiline != "" {
		pattern, err := regexp.Compile(lf.FilterMultiline)
		if err != nil {
			return fmt.Errorf("config FilterMultiline, error: %s", err)
		}
		lf.multilinePattern = pattern
	}

	// regexp
	for idx, m := range lf.FilterMessage {
		pattern, err := regexp.Compile(m)
		if err != nil {
			return fmt.Errorf("config FilterMessage index[%d], error: %s", idx, err)
		}
		lf.messagePattern = append(lf.messagePattern, pattern)
	}

	// pipeline 不是并发安全，无法支持多个 goroutine 使用同一个 pipeline 对象
	// 所以在此处使用 pool
	// 另，regexp 是并发安全的
	lf.pipelinePool = sync.Pool{
		New: func() interface{} {
			if lf.Pipeline == "" {
				return nil
			}

			// 即使 pipeline 配置错误，也不会影响全局
			p, err := pipeline.NewPipelineFromFile(filepath.Join(datakit.PipelineDir, lf.Pipeline))
			if err != nil {
				l.Debugf("new pipeline error: %s", err)
				return nil
			}
			return p
		},
	}

	return nil
}

func (lf *LogFilter) RunPipeline(message string) (map[string]interface{}, error) {
	pipe := lf.pipelinePool.Get()
	// pipe 为空指针（即没有配置 pipeline），将返回默认值
	if pipe == nil {
		return map[string]interface{}{"message": message}, nil
	}

	return pipe.(*pipeline.Pipeline).Run(message).Result()
}

func (lf *LogFilter) MatchMessage(message string) bool {
	for _, pattern := range lf.messagePattern {
		if pattern.MatchString(message) {
			return true
		}
	}
	return false
}

func (lf *LogFilter) MatchMultiline(message string) bool {
	if lf.multilinePattern == nil {
		return false
	}
	return lf.multilinePattern.MatchString(message)
}

func (this *Input) loadCfg() (err error) {
	// new docker client
	if this.Endpoint == "ENV" {
		this.client, err = this.newEnvClient()
		return
	}
	// tlsConfig 可以为空指针，即没有配置tls
	tlsConfig, _err := this.ClientConfig.TLSConfig()
	if _err != nil {
		return _err
	}

	this.client, err = this.newClient(this.Endpoint, tlsConfig)
	if err != nil {
		return
	}

	// 始终认为，docker和k8s在同一台主机上
	// 避免进行冗杂的k8s连接配置
	var k8sURL = fmt.Sprintf(defaultKubernetesURL, "127.0.0.1")
	if this.Endpoint != defaultEndpoint {
		if u, err := url.Parse(this.Endpoint); err == nil {
			k8sURL = fmt.Sprintf(defaultKubernetesURL, u.Hostname())
		}
	}

	l.Debugf("use k8sURL %s", k8sURL)

	this.kubernetes = func() *Kubernetes {
		k := Kubernetes{URL: k8sURL}
		if err := k.Init(); err != nil {
			l.Debugf("read k8s token error (use empty tokne): %s", err)
			// use empty token
			k.BearerTokenString = ""
		}
		return &k
	}()

	if this.CollectMetric {
		var dur time.Duration
		dur, err = time.ParseDuration(this.CollectMetricInterval)
		if err != nil {
			return
		}
		this.metricDuration = config.ProtectedInterval(minCollectMetricDuration, maxCollectMetricDuration, dur)
		l.Debugf("collect metrics interval %s", this.metricDuration)
	}

	if err = this.initLoggingConf(); err != nil {
		return
	}

	return
}

func (this *Input) initLoggingConf() error {
	this.opts = types.ContainerListOptions{All: this.IncludeExited}

	if len(this.DeprecatedLogOption) != 0 {
		l.Warn("log_option is deprecated")
	}

	this.containerLogsOptions = types.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Timestamps: true,
		Details:    false,
		Follow:     true,
		Tail:       "0", // 默认关闭FromBeginning，避免数据量巨大。开启为 'all'
	}

	for _, lf := range this.LogFilters {
		if err := lf.Init(); err != nil {
			return err
		}
	}

	return nil
}
