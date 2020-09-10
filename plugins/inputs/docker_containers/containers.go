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
    interval = "5s"

    # Is all containers
    all = false

    # Timeout for Docker API calls.
    timeout = "5s"

    ## Optional TLS Config
    # tls_ca = "/etc/telegraf/ca.pem"
    # tls_cert = "/etc/telegraf/cert.pem"
    # tls_key = "/etc/telegraf/key.pem"
    ## Use TLS but skip chain & host verification
    # insecure_skip_verify = false
`
	defaultEndpoint = "unix:///var/run/docker.sock"
)

var l = logger.DefaultSLogger(inputName)

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return &DockerContainers{
			Endpoint:     defaultEndpoint,
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
	intervalDutation time.Duration
	host             string
	ClientConfig

	newEnvClient func() (Client, error)
	newClient    func(string, *tls.Config) (Client, error)

	client Client
	opts   types.ContainerListOptions

	objects []*DockerObject
}

func (*DockerContainers) SampleConfig() string {
	return sampleCfg
}

func (*DockerContainers) Catalog() string {
	return "docker"
}

func (d *DockerContainers) Run() {
	l = logger.SLogger(inputName)

	if d.loadCfg() {
		return
	}

	ticker := time.NewTicker(d.intervalDutation)
	defer ticker.Stop()

	l.Info("docker_containers input start")
	for {
		select {
		case <-datakit.Exit.Wait():
			l.Info("exit")
			return

		case <-ticker.C:
			d.gather()
		}
	}
}

func (d *DockerContainers) loadCfg() bool {
	var err error
	for {
		select {
		case <-datakit.Exit.Wait():
			l.Info("exit")
			return true
		default:
			// nil
		}

		d.intervalDutation, err = time.ParseDuration(d.Interval)
		if err != nil || d.intervalDutation <= 0 {
			err = fmt.Errorf("invalid interval, %s", err.Error())
			goto label
		}

		d.timeoutDuration, err = time.ParseDuration(d.Timeout)
		if err != nil || d.timeoutDuration <= 0 {
			err = fmt.Errorf("invalid timeout, %s", err.Error())
			goto label
		}

		if d.Endpoint == "ENV" {
			d.client, err = d.newEnvClient()
			if err != nil {
				goto label
			}
		} else {
			tlsConfig, err := d.ClientConfig.TLSConfig()
			if err != nil {
				goto label
			}
			d.client, err = d.newClient(d.Endpoint, tlsConfig)
			if err != nil {
				goto label
			}
		}
		break
	label:
		l.Error(err)
		time.Sleep(time.Second)
	}

	if strings.HasPrefix(d.Endpoint, "tcp") {
		d.host = d.Endpoint
	} else {
		d.host = datakit.Cfg.MainCfg.Hostname
	}
	d.opts.All = d.All
	return false
}

func (d *DockerContainers) gather() {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, d.timeoutDuration)
	defer cancel()
	containers, err := d.client.ContainerList(ctx, d.opts)
	if err != nil {
		l.Error(err)
		return
	}

	for _, container := range containers {
		if err = d.gatherContainer(container); err != nil {
			l.Error(err)
		}
	}

	data, err := json.Marshal(d.objects)
	if err != nil {
		l.Error(err)
	} else {
		if err := io.NamedFeed(data, io.Object, inputName); err != nil {
			l.Error(err)
		}
	}

	d.objects = d.objects[:0]
}

func (d *DockerContainers) gatherContainer(container types.Container) error {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, d.timeoutDuration)
	defer cancel()
	containerJSON, err := d.client.ContainerInspect(ctx, container.ID)
	if err != nil {
		return err
	}

	var obj = DockerObject{}
	obj.Name = containerName(container.Names)
	obj.Tags = Tags{
		Class:           "docker_containers",
		ContainerName:   obj.Name,
		ContainerID:     container.ID,
		ContainerImage:  container.Image,
		ContainerStatue: container.State,
		Host:            d.host,
		PID:             strconv.Itoa(containerJSON.State.Pid),
	}
	obj.Carated = containerTime(containerJSON.Created)
	obj.Started = containerTime(containerJSON.State.StartedAt)
	obj.Finished = containerTime(containerJSON.State.FinishedAt)
	obj.Path = containerJSON.Path
	// obj.Inspect = containerJSON
	description, _ := json.Marshal(containerJSON)
	obj.Description = string(description)

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
