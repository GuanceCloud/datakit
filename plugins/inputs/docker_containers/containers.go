package docker_containers

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"strconv"
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

    # valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h"
    # require, cannot be less than zero
    interval = "5s"

    # Is all containers
    all = false

    # Timeout for Docker API calls.
    timeout = "5s"

    ## Optional TLS Config
    # tls_ca = "/tmp/ca.pem"
    # tls_cert = "/tmp/cert.pem"
    # tls_key = "/tmp/key.pem"
    ## Use TLS but skip chain & host verification
    # insecure_skip_verify = false
`
	defaultEndpoint = "unix:///var/run/docker.sock"
)

var l = logger.DefaultSLogger(inputName)

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return &DockerContainers{
			Interval:     datakit.Cfg.MainCfg.Interval,
			newEnvClient: NewEnvClient,
			newClient:    NewClient,
		}
	})
}

type DockerContainers struct {
	Endpoint string `toml:"endpoint"`
	Interval string `toml:"interval"`
	Timeout  string `toml:"timeout"`
	All      bool   `toml:"all"`

	timeoutDuration  time.Duration
	intervalDuration time.Duration
	host             string
	ClientConfig

	newEnvClient func() (Client, error)
	newClient    func(string, *tls.Config) (Client, error)

	client Client
	opts   types.ContainerListOptions

	objects []*ContainerObject
}

func (*DockerContainers) SampleConfig() string {
	return sampleCfg
}

func (*DockerContainers) Catalog() string {
	return "docker"
}

func (d *DockerContainers) Test() (result *inputs.TestResult, err error) {
	l = logger.SLogger(inputName)
	// default
	result.Desc = "数据指标获取失败，详情见错误信息"

	if err = d.loadCfg(); err != nil {
		return
	}

	var data []byte
	data, err = d.gather()
	if err != nil {
		return
	}

	result.Result = data
	result.Desc = "数据指标获取成功"
	return
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
		err = fmt.Errorf("invalid interval, %s", err.Error())
		return
	} else if d.intervalDuration <= 0 {
		err = fmt.Errorf("invalid interval, cannot be less than zero")
		return
	}

	d.timeoutDuration, err = time.ParseDuration(d.Timeout)
	if err != nil {
		err = fmt.Errorf("invalid timeout, %s", err.Error())
		return
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

	if strings.HasPrefix(d.Endpoint, "tcp") {
		d.host = d.Endpoint
	} else {
		d.host = datakit.Cfg.MainCfg.Hostname
	}
	d.opts.All = d.All

	return
}

func (d *DockerContainers) gather() ([]byte, error) {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, d.timeoutDuration)
	defer cancel()
	containers, err := d.client.ContainerList(ctx, d.opts)
	if err != nil {
		l.Error(err)
		return nil, err
	}

	for _, container := range containers {
		if err = d.gatherContainer(container); err != nil {
			l.Error(err)
		}
	}

	data, err := json.Marshal(d.objects)
	if err != nil {
		l.Error(err)
		return nil, err
	}

	d.objects = d.objects[:0]
	return data, nil
}

func (d *DockerContainers) gatherContainer(container types.Container) error {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, d.timeoutDuration)
	defer cancel()
	containerJSON, err := d.client.ContainerInspect(ctx, container.ID)
	if err != nil {
		return err
	}

	var obj = ContainerObject{
		Name:  containerName(container.Names),
		Class: "docker_containers",
	}
	obj.Name = containerName(container.Names)

	content := ObjectContent{
		ContainerName:   obj.Name,
		ContainerID:     container.ID,
		ContainerImage:  container.Image,
		ContainerStatue: container.State,
		Host:            d.host,
		PID:             strconv.Itoa(containerJSON.State.Pid),
	}
	content.Carated = containerTime(containerJSON.Created)
	content.Started = containerTime(containerJSON.State.StartedAt)
	content.Finished = containerTime(containerJSON.State.FinishedAt)
	content.Path = containerJSON.Path
	inspect, _ := json.Marshal(containerJSON)
	content.Inspect = string(inspect)

	jd, err := json.Marshal(content)
	if err != nil {
		return err

	}
	obj.Content = string(jd)

	d.objects = append(d.objects, &obj)
	return nil
}

func containerName(names []string) string {
	if len(names) > 0 {
		return strings.TrimPrefix(names[0], "/")
	}
	return ""
}

func containerTime(tim string) int64 {
	t, err := time.Parse(time.RFC3339Nano, tim)
	if err != nil {
		return 0
	}
	if t.UnixNano() < 0 {
		return -1
	}
	return t.UnixNano()
}
