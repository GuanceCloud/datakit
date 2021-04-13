package docker

import (
	"context"
	"regexp"
	"sync"
	"time"

	"github.com/docker/docker/api/types"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
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
  collect_object_interval = "5m"
  
  # Is all containers, Return all containers. By default, only running containers are shown.
  include_exited = false
  
  ## Optional TLS Config
  # tls_ca = "/path/to/ca.pem"
  # tls_cert = "/path/to/cert.pem"
  # tls_key = "/path/to/key.pem"
  ## Use TLS but skip chain & host verification
  # insecure_skip_verify = false
  
  [[inputs.docker.log_option]]
    # container_name_match = "<regexp-container-name>"
    # source = "<your-source>"
    # service = "<your-service>"
    # pipeline = "<this-is-pipeline>"
  
  [inputs.docker.tags]
    # tags1 = "value1"
`
	defaultEndpoint              = "unix:///var/run/docker.sock"
	defaultAPITimeout            = time.Second * 5
	minimumCollectMetricDuration = time.Second * 5
	minimumCollectObjectDuration = time.Minute * 5
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

	this.kubernetes = func() *Kubernetes {
		k := Kubernetes{URL: defaultKubernetesURL}
		if err := k.Init(); err != nil {
			l.Warn(err)
			return nil
		}
		return &k
	}()

	if this.CollectMetric {
		this.collectMetricDuration, err = time.ParseDuration(this.CollectMetricInterval)
		if err != nil {
			return
		}
	}

	if this.CollectObject {
		this.collectObjectDuration, err = time.ParseDuration(this.CollectObjectInterval)
		if err != nil {
			return
		}
	}

	// 限制最小采集间隔
	if 0 < this.collectMetricDuration &&
		this.collectMetricDuration < minimumCollectMetricDuration {
		l.Warn("invalid collect_metric_interval, cannot be less than 5s. Use default interval 5s")
		this.collectMetricDuration = minimumCollectMetricDuration
	}

	// if 0 < this.collectObjectDuration &&
	// 	this.collectObjectDuration < minimumCollectObjectDuration {
	// 	l.Warn("invalid collect_object_interval, cannot be less than 5m. Use default interval 5m")
	// 	this.collectObjectDuration = minimumCollectObjectDuration
	// }

	if this.Tags == nil {
		this.Tags = make(map[string]string)
	}

	if err = this.initLogOption(); err != nil {
		return
	}

	this.timeoutDuration = defaultAPITimeout

	return
}

func (this *Input) initLogOption() (err error) {
	this.opts = types.ContainerListOptions{All: this.IncludeExited}
	this.containerLogList = make(map[string]context.CancelFunc)

	this.containerLogsOptions = types.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Timestamps: true,
		Details:    false,
		Follow:     true,
		Tail:       "0", // 默认关闭FromBeginning，避免数据量巨大。开启为 'all'
	}

	for _, opt := range this.LogOption {
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
						l.Error(err)
						return nil
					}
					return p
				},
			}
		}(opt.Pipeline)
	}

	return
}
