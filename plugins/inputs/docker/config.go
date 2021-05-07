package docker

import (
	"fmt"
	"net/url"
	"regexp"
	"sync"
	"time"

	"github.com/docker/docker/api/types"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline"
)

const (
	inputName = "docker"

	sampleCfg = `
[inputs.docker]
  # Docker Endpoint
  # To use TCP, set endpoint = "tcp://[ip]:[port]"
  # To use environment variables (ie, docker-machine), set endpoint = "ENV"
  endpoint = "unix:///var/run/docker.sock"
  
  collect_metric = true
  collect_object = true
  collect_logging = true
  
  # Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h"
  collect_metric_interval = "10s"
  
  # Is all containers, Return all containers. By default, only running containers are shown.
  include_exited = false
  
  ## Optional TLS Config
  # tls_ca = "/path/to/ca.pem"
  # tls_cert = "/path/to/cert.pem"
  # tls_key = "/path/to/key.pem"
  ## Use TLS but skip chain & host verification
  # insecure_skip_verify = false
  
  #[[inputs.docker.log_option]]
    # container_name_match = '''<regexp-container-name>'''
    # source = "<your-source>"
    # service = "<your-service>"
    # pipeline = "<this-is-pipeline>"
  
  [inputs.docker.tags]
    # tags1 = "value1"
`
	defaultEndpoint          = "unix:///var/run/docker.sock"
	defaultAPITimeout        = time.Second * 5
	minCollectMetricDuration = time.Second * 5
	maxCollectMetricDuration = time.Second * 60
	collectObjectDuration    = time.Minute * 5
)

var l = logger.DefaultSLogger(inputName)

type LogOption struct {
	NameMatch string `toml:"container_name_match"`
	Source    string `toml:"source"`
	Service   string `toml:"service"`
	Pipeline  string `toml:"pipeline"`

	pipelinePool sync.Pool
	nameCompile  *regexp.Regexp
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
			l.Debugf("init kubernetes connect error: %s", err)
			return nil
		}
		return &k
	}()

	if this.CollectMetric {
		var dur time.Duration
		dur, err = time.ParseDuration(this.CollectMetricInterval)
		if err != nil {
			return
		}
		this.collectMetricDuration = datakit.ProtectedInterval(minCollectMetricDuration, maxCollectMetricDuration, dur)
		l.Debugf("collect metrics interval %s", this.collectMetricDuration)
	}

	if err = this.initLogOption(); err != nil {
		return
	}

	return
}

func (this *Input) initLogOption() (err error) {
	this.opts = types.ContainerListOptions{All: this.IncludeExited}

	this.containerLogsOptions = types.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Timestamps: true,
		Details:    false,
		Follow:     true,
		Tail:       "0", // 默认关闭FromBeginning，避免数据量巨大。开启为 'all'
	}

	for _, opt := range this.LogOption {
		// 此为基本配置，为空值时直接continue
		if opt.NameMatch == "" {
			continue
		}
		// opt.Source为空时，会默认使用 container_name
		// opt.Service为空时，会默认使用 container_name
		if opt.Service == "" {
			opt.Service = opt.Source
		}

		opt.nameCompile, err = regexp.Compile(opt.NameMatch)
		if err != nil {
			return
		}

		func(pipelinePath string) {
			opt.pipelinePool = sync.Pool{
				New: func() interface{} {
					p, err := pipeline.NewPipelineFromFile(pipelinePath)
					if err != nil {
						l.Debug(err)
						return nil
					}
					return p
				},
			}
		}(opt.Pipeline)
	}

	return
}
