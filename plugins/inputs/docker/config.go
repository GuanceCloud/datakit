package docker

import (
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
)

const (
	inputName = "docker"

	sampleCfg = `
[inputs.docker]
    # Docker Endpoint
    # To use TCP, set endpoint = "tcp://[ip]:[port]"
    # To use environment variables (ie, docker-machine), set endpoint = "ENV"
    endpoint = "unix:///var/run/docker.sock"

    # Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h"
    collect_metric_interval = "10s"
    collect_object_interval = "5m"

    collect_logging = true

    # Is all containers, Return all containers. By default, only running containers are shown.
    include_exited = false

    ## Optional TLS Config
    # tls_ca = "/path/to/ca.pem"
    # tls_cert = "/path/to/cert.pem"
    # tls_key = "/path/to/key.pem"
    ## Use TLS but skip chain & host verification
    # insecure_skip_verify = false

    [[inputs.docker.log_pipeline]]
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

type DockerLogPipeline struct {
	ContainerNameMatch string `toml:"container_name_match"`
	Source             string `toml:"source"`
	Service            string `toml:"service"`
	Pipeline           string `toml:"pipeline"`
}

func (d *DockerUtil) loadCfg() (err error) {
	// new docker client
	if d.Endpoint == "ENV" {
		d.client, err = d.newEnvClient()
		return
	}
	// tlsConfig 可以为空指针，即没有配置tls
	tlsConfig, _err := d.ClientConfig.TLSConfig()
	if _err != nil {
		return _err
	}

	d.client, err = d.newClient(d.Endpoint, tlsConfig)
	if err != nil {
		return
	}

	d.collectMetricDuration, err = time.ParseDuration(d.CollectMetricInterval)
	if err != nil {
		return
	}

	d.collectObjectDuration, err = time.ParseDuration(d.CollectObjectInterval)
	if err != nil {
		return
	}

	// 限制最小采集间隔
	if 0 < d.collectMetricDuration &&
		d.collectMetricDuration < minimumCollectMetricDuration {
		l.Warn("invalid collect_metric_interval, cannot be less than 5s. Use default interval 5s")
		d.collectMetricDuration = minimumCollectMetricDuration
	}

	if 0 < d.collectObjectDuration &&
		d.collectObjectDuration < minimumCollectObjectDuration {
		l.Warn("invalid collect_object_interval, cannot be less than 5m. Use default interval 5m")
		d.collectObjectDuration = minimumCollectObjectDuration
	}

	if d.Tags == nil {
		d.Tags = make(map[string]string)
	}

	return
}
