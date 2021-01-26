package dockercontainers

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"strings"
	"time"

	"github.com/docker/docker/api/types"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

const (
	inputName = "docker_containers"

	sampleCfg = `
[[inputs.docker_containers]]
    # Docker Endpoint
    # To use TCP, set endpoint = "tcp://[ip]:[port]"
    # To use environment variables (ie, docker-machine), set endpoint = "ENV"
    endpoint = "unix:///var/run/docker.sock"

    # Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h"
    # Require. Cannot be less than zero, minimum 5m and maximum 1h.
    interval = "5m"

    # Is all containers, Return all containers. By default, only running containers are shown.
    all = false

    ## Optional TLS Config
    # tls_ca = "/tmp/ca.pem"
    # tls_cert = "/tmp/cert.pem"
    # tls_key = "/tmp/key.pem"
    ## Use TLS but skip chain & host verification
    # insecure_skip_verify = false
    
    ## Use containerID link kubernetes pods
    # [inputs.docker_containers.kubernetes]
    #   ## URL for the kubelet
    #   url = "http://127.0.0.1:10255"
    #
    #   ## Use bearer token for authorization. ('bearer_token' takes priority)
    #   ## If both of these are empty, we'll use the default serviceaccount:
    #   ## at: /run/secrets/kubernetes.io/serviceaccount/token
    #   # bearer_token = "/path/to/bearer/token"
    #   ## OR
    #   # bearer_token_string = "abc_123"
    #
    #   ## Optional TLS Config
    #   # tls_ca = /path/to/cafile
    #   # tls_cert = /path/to/certfile
    #   # tls_key = /path/to/keyfile
    #   ## Use TLS but skip chain & host verification
    #   # insecure_skip_verify = false

`
	defaultEndpoint       = "unix:///var/run/docker.sock"
	defaultGetherInterval = time.Minute * 5
	maxGetherInterval     = time.Hour
	defaultAPITimeout     = time.Second * 10
)

var l = logger.DefaultSLogger(inputName)

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return &DockerContainers{
			Endpoint:     defaultEndpoint,
			All:          false,
			newEnvClient: NewEnvClient,
			newClient:    NewClient,
		}
	})
}

type DockerContainers struct {
	Endpoint string `toml:"endpoint"`
	Interval string `toml:"interval"`
	All      bool   `toml:"all"`

	intervalDuration time.Duration
	ClientConfig

	newEnvClient func() (Client, error)
	newClient    func(string, *tls.Config) (Client, error)

	client Client
	opts   types.ContainerListOptions

	Kubernetes *Kubernetes `toml:"kubernetes"`
}

func (*DockerContainers) SampleConfig() string {
	return sampleCfg
}

func (*DockerContainers) Catalog() string {
	return "docker"
}

func (*DockerContainers) PipelineConfig() map[string]string {
	return nil
}

func (d *DockerContainers) Test() (*inputs.TestResult, error) {
	l = logger.SLogger(inputName)

	var result = inputs.TestResult{Desc: "数据指标获取失败，详情见错误信息"}
	var err error

	if err = d.loadCfg(); err != nil {
		return &result, err
	}

	var data []byte
	data, err = d.gather()
	if err != nil {
		return &result, err
	}

	result.Result = data
	result.Desc = "数据指标获取成功"

	return &result, err
}

func (d *DockerContainers) Run() {
	l = logger.SLogger(inputName)

	if d.initCfg() {
		return
	}

	ticker := time.NewTicker(d.intervalDuration)
	defer ticker.Stop()

	l.Info("docker_containers input start")
	for {
		select {
		case <-datakit.Exit.Wait():
			l.Info("exit")
			return

		case <-ticker.C:
			data, err := d.gather()
			if err != nil {
				continue
			}
			if err := io.NamedFeed(data, io.Object, inputName); err != nil {
				l.Error(err)
			}
		}
	}
}

func (d *DockerContainers) initCfg() bool {
	for {
		select {
		case <-datakit.Exit.Wait():
			l.Info("exit")
			return true
		default:
			// nil
		}

		if err := d.loadCfg(); err != nil {
			l.Error(err)
			time.Sleep(time.Second)
		} else {
			break
		}
	}
	return false
}

func (d *DockerContainers) loadCfg() (err error) {
	d.intervalDuration, err = time.ParseDuration(d.Interval)
	if err != nil {
		l.Warnf("invalid interval, %s", err)
		l.Warn("Use default interval 5m")

		d.intervalDuration = defaultGetherInterval
	} else {
		if d.intervalDuration <= 0 ||
			d.intervalDuration < defaultGetherInterval ||
			maxGetherInterval < d.intervalDuration {

			l.Warn("invalid interval, cannot be less than zero, between 5m and 1h.")
			l.Warn("Use default interval 5m")

			d.intervalDuration = defaultGetherInterval
		}
	}

	if d.Endpoint == "ENV" {
		d.client, err = d.newEnvClient()
		if err != nil {
			return
		}
	} else {
		tlsConfig, _err := d.ClientConfig.TLSConfig()
		if _err != nil {
			return _err
		}
		d.client, err = d.newClient(d.Endpoint, tlsConfig)
		if err != nil {
			return
		}
	}
	d.opts.All = d.All

	return
}

func (d *DockerContainers) gather() ([]byte, error) {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, defaultAPITimeout)
	defer cancel()

	containers, err := d.client.ContainerList(ctx, d.opts)
	if err != nil {
		l.Error(err)
		return nil, err
	}

	var buffer bytes.Buffer

	for _, container := range containers {
		data, err := d.gatherContainer(container)
		if err != nil {
			l.Error(err)
		} else {
			buffer.Write(data)
			buffer.WriteString("\n")
		}
	}

	return buffer.Bytes(), nil
}

func (d *DockerContainers) gatherContainer(container types.Container) ([]byte, error) {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, defaultAPITimeout)
	defer cancel()

	containerJSON, err := d.client.ContainerInspect(ctx, container.ID)
	if err != nil {
		return nil, err
	}

	containerProcessList, err := d.client.ContainerTop(ctx, container.ID, nil)
	if err != nil {
		return nil, err
	}

	msg, err := json.Marshal(struct {
		types.ContainerJSON
		Process containerTop `json:"Process"`
	}{
		containerJSON,
		containerProcessList,
	})

	if err != nil {
		return nil, err
	}

	fields := map[string]interface{}{"message": string(msg)}
	fields["container_id"] = container.ID
	fields["images_name"] = container.Image
	fields["created_time"] = container.Created
	fields["container_name"] = getContainerName(container.Names)
	fields["restart_count"] = containerJSON.RestartCount
	fields["status"] = containerJSON.State.Status
	fields["start_time"] = containerJSON.State.StartedAt

	if d.Kubernetes != nil {
		m, err := d.Kubernetes.GatherPodInfo(container.ID)
		if err != nil {
			l.Warn(err)
		}
		for k, v := range m {
			fields[k] = v
		}
	}

	tags := map[string]string{"name": container.ID}

	return io.MakeMetric(inputName, tags, fields, time.Now())
}

func getContainerName(names []string) string {
	if len(names) > 0 {
		return strings.TrimPrefix(names[0], "/")
	}
	return ""
}
