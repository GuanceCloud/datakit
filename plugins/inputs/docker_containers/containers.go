package docker_containers

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
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

func (d *DockerContainers) Run() {
	l = logger.SLogger(inputName)

	if d.initCfg() {
		return
	}

	ticker := time.NewTicker(d.intervalDuration)
	defer ticker.Stop()

	l.Info("docker_containers input start")

	// 采集器开启时，先运行一次，然后等待ticker
	d.do()

	for {
		select {
		case <-datakit.Exit.Wait():
			l.Info("exit")
			return

		case <-ticker.C:
			d.do()
		}
	}
}

func (d *DockerContainers) do() {
	data, err := d.gather()
	if err != nil {
		return
	}
	if err := io.NamedFeed(data, datakit.Object, inputName); err != nil {
		l.Error(err)
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
		l.Warnf("invalid interval: %s, use default interval 5m", err)
		d.intervalDuration = defaultGetherInterval
	} else {
		if d.intervalDuration <= 0 ||
			d.intervalDuration < defaultGetherInterval ||
			maxGetherInterval < d.intervalDuration {

			l.Warn("invalid interval, cannot be less than zero, between 5m and 1h. Use default interval 5m")
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
			// 忽略某一个container的错误
			// 继续gather下一个
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

	msg, err := d.composeMessage(ctx, container.ID, &containerJSON)
	if err != nil {
		return nil, err
	}

	stats, err := d.gatherStats(ctx, container.ID)
	if err != nil {
		l.Warnf("gather stats error, %s", err)
	}

	podInfo, err := d.gatherK8sPodInfo(container.ID)
	if err != nil {
		l.Warnf("gather k8s pod error, %s", err)
	}

	t, err := time.Parse(time.RFC3339Nano, containerJSON.State.StartedAt)
	if err != nil {
		return nil, fmt.Errorf("parse start_time error, %s", err)
	}

	fields := map[string]interface{}{
		"container_id":   container.ID,
		"images_name":    container.Image,
		"created_time":   container.Created,
		"container_name": getContainerName(container.Names),
		"restart_count":  containerJSON.RestartCount,
		"status":         containerJSON.State.Status,
		"start_time":     t.UnixNano() / int64(time.Millisecond),
		"message":        string(msg),
	}

	for k, v := range stats {
		fields[k] = v
	}

	for k, v := range podInfo {
		fields[k] = v
	}

	tags := map[string]string{"name": container.ID}

	return io.MakeMetric(inputName, tags, fields, time.Now())
}

const streamStats = false

func (d *DockerContainers) gatherStats(ctx context.Context, id string) (map[string]interface{}, error) {
	resp, err := d.client.ContainerStats(ctx, id, streamStats)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var v *types.StatsJSON

	if err := json.NewDecoder(resp.Body).Decode(&v); err != nil {
		return nil, err
	}

	if resp.OSType == "windows" {
		return nil, nil
	}

	cpuPercent := calculateCPUPercentUnix(v.PreCPUStats.CPUUsage.TotalUsage, v.PreCPUStats.SystemUsage, v)
	mem := calculateMemUsageUnixNoCache(v.MemoryStats)
	memLimit := float64(v.MemoryStats.Limit)
	memPercent := calculateMemPercentUnixNoCache(memLimit, mem)
	netRx, netTx := calculateNetwork(v.Networks)
	blkRead, blkWrite := calculateBlockIO(v.BlkioStats)

	return map[string]interface{}{
		"cpu_usage":        cpuPercent,
		"mem_usage":        mem,
		"mem_limit":        memLimit,
		"mem_used_percent": memPercent,
		"network_in":       netRx,
		"network_out":      netTx,
		"block_in":         float64(blkRead),
		"block_out":        float64(blkWrite),
	}, nil
}

func (d *DockerContainers) gatherK8sPodInfo(id string) (map[string]string, error) {
	if d.Kubernetes == nil {
		return nil, nil
	}
	return d.Kubernetes.GatherPodInfo(id)
}

func (d *DockerContainers) composeMessage(ctx context.Context, id string, v *types.ContainerJSON) ([]byte, error) {
	// 容器未启动时，无法进行containerTop，此处会得到error
	// 与 opt.All 冲突，忽略此error即可
	t, _ := d.client.ContainerTop(ctx, id, nil)

	return json.Marshal(struct {
		types.ContainerJSON
		Process containerTop `json:"Process"`
	}{
		*v,
		t,
	})
}

func getContainerName(names []string) string {
	if len(names) > 0 {
		return strings.TrimPrefix(names[0], "/")
	}
	return ""
}
