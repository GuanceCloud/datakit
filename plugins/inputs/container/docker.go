package container

import (
	"bufio"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/docker/docker/pkg/stdcopy"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/tailer"
	iod "gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline"

	"github.com/docker/docker/api/types"
	docker "github.com/docker/docker/client"
)

// 日志 source 选择
// 0. 默认是 container_name
// 1. 当容器由 k8s 创建，且 deployment 不为空时，使用 deployment 值
// 2. 当 cotainer_name 或 deployment 匹配 log filter 时，使用该 filter 的 source，即用户配置为最高优先
// service 同上，service 是一个 tag 而非 field
// commit: 71bfb0731dedcbb624cecb23a1ede9846311767b

const (
	// Maximum bytes of a log line before it will be split, size is mirroring
	// docker code:
	// https://github.com/moby/moby/blob/master/daemon/logger/copier.go#L21
	maxLineBytes = 16 * 1024

	// ES value can be at most 32766 bytes long
	maxFieldsLength = 32766

	pipelineTimeField = "time"

	useIOHighFreq = true

	containerIDPrefix = "docker://"
)

type dockerClient struct {
	client *docker.Client
	K8s    *Kubernetes

	IgnoreImageName     []string
	IgnoreContainerName []string

	ProcessTags func(tags map[string]string)
	Logs        Logs

	containerLogsOptions types.ContainerLogsOptions
	containerLogList     map[string]context.CancelFunc

	mu sync.Mutex
	wg sync.WaitGroup
}

/*This file is inherited from telegraf docker input plugin*/
var (
	version        = "1.24"
	defaultHeaders = map[string]string{"User-Agent": "engine-api-cli-1.0"}

	// 容器日志的连接参数
	containerLogsOptions = types.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     true,
		Tail:       "0", // 默认关闭FromBeginning，避免数据量巨大。开启为 'all'
	}
)

func newDockerClient(host string, tlsConfig *tls.Config) (*dockerClient, error) {
	transport := &http.Transport{
		TLSClientConfig: tlsConfig,
	}

	httpClient := &http.Client{Transport: transport}

	client, err := docker.NewClientWithOpts(
		docker.WithHTTPHeaders(defaultHeaders),
		docker.WithHTTPClient(httpClient),
		docker.WithVersion(version),
		docker.WithHost(host))
	if err != nil {
		return nil, err
	}

	return &dockerClient{
		client:               client,
		containerLogList:     make(map[string]context.CancelFunc),
		containerLogsOptions: containerLogsOptions,
	}, nil
}

func newDockerClientFromEnv() (*dockerClient, error) {
	client, err := docker.NewClientWithOpts(docker.FromEnv)
	if err != nil {
		return nil, err
	}

	return &dockerClient{client: client}, nil
}

func (d *dockerClient) Stop() {
	d.cancelTails()
	d.wg.Wait()
	return
}

func (d *dockerClient) Metric(ctx context.Context, in chan<- []*job) {
	var jobs []*job
	fn := func(c types.Container) {
		if ignoreCommand(c.Command) {
			return
		}

		if d.ignoreImageName(c.Image) || d.ignoreContainerName(c.ID) {
			return
		}

		result, err := d.gather(c)
		if err != nil {
			l.Error(err)
			return
		}

		result.setMetric()
		jobs = append(jobs, result)
	}

	if err := d.do(ctx, fn, types.ContainerListOptions{All: containerAllForMetric}); err != nil {
		l.Error(err)
		return
	}
	l.Debugf("get len(%d) container metric", len(jobs))
	in <- jobs
}

func (d *dockerClient) Object(ctx context.Context, in chan<- []*job) {
	var jobs []*job
	fn := func(c types.Container) {
		if ignoreCommand(c.Command) {
			return
		}

		if d.ignoreImageName(c.Image) || d.ignoreContainerName(c.ID) {
			return
		}

		result, err := d.gather(c)
		if err != nil {
			l.Error(err)
			return
		}

		result.addTag("name", c.ID)
		if hostname, err := d.getContainerHostname(c.ID); err != nil {
			result.addTag("container_host", hostname)
		}
		result.addTag("status", c.Status)
		result.addField("age", time.Since(time.Unix(c.Created, 0)).Milliseconds()/1e3) // 毫秒除以1000得秒数，不使用Second()因为它返回浮点
		result.addField("from_kubernetes", contianerIsFromKubernetes(getContainerName(c.Names)))

		if message, err := result.marshal(); err != nil {
			l.Warnf("failed of marshal json, %s", err)
		} else {
			result.addField("message", string(message))
		}

		if process, err := d.gatherSingleContainerProcessToJSON(c); err != nil {
			l.Debug(err)
		} else {
			result.addField("process", process)
		}

		result.setObject()
		jobs = append(jobs, result)
	}

	if err := d.do(ctx, fn, types.ContainerListOptions{All: containerAllForObject}); err != nil {
		l.Error(err)
		return
	}
	l.Debugf("get len(%d) container object", len(jobs))
	in <- jobs
}

func (d *dockerClient) do(ctx context.Context, processFunc func(types.Container), opt types.ContainerListOptions) error {
	cList, err := d.client.ContainerList(ctx, opt)
	if err != nil {
		l.Error(err)
		return err
	}

	var wg sync.WaitGroup
	for _, container := range cList {
		wg.Add(1)
		go func(c types.Container) {
			defer wg.Done()
			processFunc(c)
		}(container)
	}

	wg.Wait()
	return nil
}

func (d *dockerClient) gather(container types.Container) (*job, error) {
	startTime := time.Now()
	tags := d.gatherContainerInfo(container)

	var fields = make(map[string]interface{})
	var err error

	// 注意，此处如果没有 fields，构建 point 会失败
	// 需要在上层手动 addFiedls
	if container.State == "running" {
		fields, err = d.gatherSingleContainerStats(container)
		if err != nil {
			l.Error(err)
			return nil, err
		}
	}
	cost := time.Since(startTime)

	return &job{measurement: containerName, tags: tags, fields: fields, ts: time.Now(), cost: cost}, nil
}

func (d *dockerClient) ignoreImageName(name string) bool {
	return regexpMatchString(d.IgnoreImageName, name)
}

func (d *dockerClient) ignoreContainerName(name string) bool {
	return regexpMatchString(d.IgnoreContainerName, name)
}

func (d *dockerClient) gatherContainerInfo(container types.Container) map[string]string {
	imageName, imageShortName, imageTag := ParseImage(container.Image)
	var tags = map[string]string{
		"state":            container.State,
		"docker_image":     container.Image,
		"image_name":       imageName,
		"image_short_name": imageShortName,
		"image_tag":        imageTag,
	}

	for k, v := range d.getContainerTags(container) {
		tags[k] = v
	}

	if d.ProcessTags != nil {
		d.ProcessTags(tags)
	}

	return tags
}

func (d *dockerClient) gatherSingleContainerProcess(container types.Container) ([]map[string]string, error) {
	// query parameters: top
	// default "-ef"
	// The arguments to pass to ps. For example, aux
	top, err := d.client.ContainerTop(context.TODO(), container.ID, nil)
	if err != nil {
		return nil, err
	}

	var res []map[string]string

	for _, proc := range top.Processes {
		if len(proc) != len(top.Titles) {
			continue
		}

		var p = make(map[string]string)

		for idx, title := range top.Titles {
			p[title] = proc[idx]
		}

		res = append(res, p)
	}

	return res, nil
}

func (d *dockerClient) gatherSingleContainerProcessToJSON(container types.Container) (string, error) {
	process, err := d.gatherSingleContainerProcess(container)
	if err != nil {
		return "", err
	}

	j, err := json.Marshal(process)
	if err != nil {
		return "", err
	}

	return string(j), nil
}

const streamStats = false

func (d *dockerClient) gatherSingleContainerStats(container types.Container) (map[string]interface{}, error) {
	resp, err := d.client.ContainerStats(context.TODO(), container.ID, streamStats)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.OSType == "windows" {
		return nil, nil
	}

	var v *types.StatsJSON
	if err := json.NewDecoder(resp.Body).Decode(&v); err != nil {
		return nil, err
	}

	return d.calculateContainerStats(v), nil
}

func (d *dockerClient) calculateContainerStats(v *types.StatsJSON) map[string]interface{} {
	mem := calculateMemUsageUnixNoCache(v.MemoryStats)
	memPercent := calculateMemPercentUnixNoCache(float64(v.MemoryStats.Limit), float64(mem))
	netRx, netTx := calculateNetwork(v.Networks)
	blkRead, blkWrite := calculateBlockIO(v.BlkioStats)

	return map[string]interface{}{
		"cpu_usage":          calculateCPUPercentUnix(v.PreCPUStats.CPUUsage.TotalUsage, v.PreCPUStats.SystemUsage, v), /*float64*/
		"cpu_delta":          calculateCPUDelta(v),
		"cpu_system_delta":   calculateCPUSystemDelta(v),
		"cpu_numbers":        calculateCPUNumbers(v),
		"mem_limit":          int64(v.MemoryStats.Limit),
		"mem_usage":          mem,
		"mem_used_percent":   memPercent, /*float64*/
		"mem_failed_count":   int64(v.MemoryStats.Failcnt),
		"network_bytes_rcvd": netRx,
		"network_bytes_sent": netTx,
		"block_read_byte":    blkRead,
		"block_write_byte":   blkWrite,
	}
}

func (d *dockerClient) getContainerHostname(id string) (string, error) {
	containerJson, err := d.client.ContainerInspect(context.TODO(), id)
	if err != nil {
		return "", err
	}
	return containerJson.Config.Hostname, nil
}

func getContainerName(names []string) string {
	if len(names) > 0 {
		return strings.TrimPrefix(names[0], "/")
	}
	return "invalidContainerName"
}

// contianerIsFromKubernetes 判断该容器是否由kubernetes创建
// 所有kubernetes启动的容器的containerNamePrefix都是k8s，依据链接如下
// https://github.com/rootsongjc/kubernetes-handbook/blob/master/practice/monitor.md#%E5%AE%B9%E5%99%A8%E7%9A%84%E5%91%BD%E5%90%8D%E8%A7%84%E5%88%99
func contianerIsFromKubernetes(containerName string) bool {
	const kubernetesContainerNamePrefix = "k8s"
	return strings.HasPrefix(containerName, kubernetesContainerNamePrefix)
}

//
////////////////////////////////////// LOGGING ////////////////////////////////////////////////
//

func (d *dockerClient) addToContainerList(containerID string, cancel context.CancelFunc) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.containerLogList[containerID] = cancel
	return nil
}

func (d *dockerClient) removeFromContainerList(containerID string) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	delete(d.containerLogList, containerID)
	return nil
}

func (d *dockerClient) containerInContainerList(containerID string) bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	_, ok := d.containerLogList[containerID]
	return ok
}

func (d *dockerClient) cancelTails() error {
	d.mu.Lock()
	defer d.mu.Unlock()
	for _, cancel := range d.containerLogList {
		cancel()
	}
	return nil
}

func (d *dockerClient) hasTTY(ctx context.Context, container types.Container) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, apiTimeoutDuration)
	defer cancel()
	c, err := d.client.ContainerInspect(ctx, container.ID)
	if err != nil {
		return false, err
	}
	return c.Config.Tty, nil
}

func (d *dockerClient) Logging(ctx context.Context) {
	ctx, cancel := context.WithTimeout(ctx, apiTimeoutDuration)
	defer cancel()

	cList, err := d.client.ContainerList(ctx, types.ContainerListOptions{All: containerAllForLogging})
	if err != nil {
		return
	}

	for _, container := range cList {
		if ignoreCommand(container.Command) {
			continue
		}

		// ParseImage() return imageName imageShortName and imageVersion, discard imageShortName and imageVersion
		imageName, _, _ := ParseImage(container.Image)
		if d.ignoreImageName(imageName) ||
			d.ignoreContainerName(container.ID) {
			continue
		}

		if d.containerInContainerList(container.ID) {
			continue
		}

		ctx, cancel := context.WithCancel(context.Background())
		d.addToContainerList(container.ID, cancel)

		// Start a new goroutine for every new container that has logs to collect
		d.wg.Add(1)
		go func(container types.Container) {
			defer d.wg.Done()
			defer d.removeFromContainerList(container.ID)

			err = d.tailContainerLogs(ctx, container)
			if err != nil && err != context.Canceled {
				l.Error(err)
				iod.FeedLastError(inputName, fmt.Sprintf("failed of gather logging, restart this container logging, name:%s ID:%s error: %s",
					getContainerName(container.Names), container.ID, err))
			}
		}(container)
	}
}

func (d *dockerClient) getContainerTags(container types.Container) map[string]string {
	name := getContainerName(container.Names)
	tags := map[string]string{
		"container_name": name,
		"container_id":   container.ID,
	}

	if !contianerIsFromKubernetes(name) {
		tags["container_type"] = "docker"
	} else {
		tags["container_type"] = "kubernetes"
	}

	if d.K8s != nil {
		func() {
			pods, err := d.K8s.getPods()
			if err != nil {
				l.Warn(err)
				return
			}
			id := "docker://" + container.ID
			if name := pods.GetContainerPodName(id); name != "" {
				tags["pod_name"] = name
			}
			if namespace := pods.GetContainerPodNamespace(id); namespace != "" {
				tags["pod_namespace"] = namespace
			}
			if deploymentName := pods.GetContainerDeploymentName(id); deploymentName != "" {
				tags["deployment"] = deploymentName
			}
		}()
	}

	return tags
}

func (d *dockerClient) tailContainerLogs(ctx context.Context, container types.Container) error {
	hasTTY, err := d.hasTTY(ctx, container)
	if err != nil {
		return err
	}

	logReader, err := d.client.ContainerLogs(ctx, container.ID, containerLogsOptions)
	if err != nil {
		return err
	}

	// If the container is using a TTY, there is only a single stream
	// (stdout), and data is copied directly from the container output stream,
	// no extra multiplexing or headers.
	//
	// If the container is *not* using a TTY, streams for stdout and stderr are
	// multiplexed.
	if hasTTY {
		return d.tailStream(ctx, logReader, "tty", container)
	} else {
		return d.tailMultiplexed(ctx, logReader, container)
	}

}

func (d *dockerClient) tailStream(ctx context.Context, reader io.ReadCloser, stream string, container types.Container) error {
	defer reader.Close()

	var (
		tags       = d.getContainerTags(container)
		name       = tags["container_name"]
		deployment = tags["deployment"]

		pipe *pipeline.Pipeline
		err  error

		source = func() string {
			// measurement 默认使用容器名，如果该容器是 k8s 创建，则尝试使用它的 deployment name
			if tags["container_type"] == "kubernetes" && deployment != "" {
				return deployment
			}
			return name
		}()
		service = name
	)

	l.Debugf("matched name:%s deployment:%s", name, deployment)

	if n := d.Logs.MatchName(deployment, name); n != -1 {
		pipe, err = pipeline.NewPipelineFromFile(filepath.Join(datakit.PipelineDir, d.Logs[n].Pipeline))
		if err != nil {
			l.Debugf("new pipeline error: %s", err)
		}
		source = d.Logs[n].Source
		service = d.Logs[n].Service
	}

	tags["stream"] = stream
	tags["service"] = service

	r := bufio.NewReaderSize(reader, maxLineBytes)

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			// nil
		}

		line, err := r.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}

		if len(line) == 0 {
			continue
		}

		text := strings.TrimSpace(string(line))

		if err := tailer.NewLogs(text).
			Pipeline(pipe).
			CheckFieldsLength().
			AddStatus(false).
			TakeTime().
			Point(source, tags).
			Feed(inputName).
			Error(); err != nil {
			l.Error("logging gather failed, container_id: %s, container_name:%s, err: %s", err.Error())
		}
	}
}

func (d *dockerClient) tailMultiplexed(ctx context.Context, src io.ReadCloser, container types.Container) error {
	outReader, outWriter := io.Pipe()
	errReader, errWriter := io.Pipe()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		err := d.tailStream(ctx, outReader, "stdout", container)
		if err != nil {
			l.Error(err)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		err := d.tailStream(ctx, errReader, "stderr", container)
		if err != nil {
			l.Error(err)
		}
	}()

	_, err := stdcopy.StdCopy(outWriter, errWriter, src)
	outWriter.Close()
	errWriter.Close()
	src.Close()
	wg.Wait()
	return err
}

// ignoreCommand 忽略 k8s pod 的 init container
func ignoreCommand(command string) bool {
	return command == "/pause"
}
