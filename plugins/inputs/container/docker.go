package container

import (
	"bufio"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/docker/docker/pkg/stdcopy"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	iod "gitlab.jiagouyun.com/cloudcare-tools/datakit/io"

	"github.com/docker/docker/api/types"
	docker "github.com/docker/docker/client"
)

const (
	// Maximum bytes of a log line before it will be split, size is mirroring
	// docker code:
	// https://github.com/moby/moby/blob/master/daemon/logger/copier.go#L21
	maxLineBytes = 16 * 1024

	// ES value can be at most 32766 bytes long
	maxFieldsLength = 32766

	pipelineTimeField = "time"

	useIOHighFreq = true
)

type dockerClient struct {
	client *docker.Client
	K8s    *Kubernetes

	IgnoreImageName     []string
	IgnoreContainerName []string

	ProcessTags func(tags map[string]string)
	LogFilters  LogFilters

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

func (d *dockerClient) Metric(ctx context.Context, in chan<- *job) {
	fn := func(c types.Container) {
		if d.ignoreImageNameFromContainer(c.Image) || d.ignoreContainerNameFromContainer(c.ID) {
			return
		}

		result, err := d.gather(c)
		if err != nil {
			l.Error(err)
			return
		}

		result.setMetric()
		in <- result
	}

	if err := d.do(ctx, fn, types.ContainerListOptions{All: containerAllForMetric}); err != nil {
		l.Error(err)
	}
}

func (d *dockerClient) Object(ctx context.Context, in chan<- *job) {
	fn := func(c types.Container) {
		if d.ignoreImageNameFromContainer(c.Image) || d.ignoreContainerNameFromContainer(c.ID) {
			return
		}

		result, err := d.gather(c)
		if err != nil {
			l.Error(err)
			return
		}

		result.addTag("name", c.ID)
		if hostname, err := d.getContainerHostname(ctx, c.ID); err != nil {
			result.addTag("container_host", hostname)
		}
		result.addTag("status", c.Status)
		result.addField("age", time.Since(time.Unix(c.Created, 0)).Milliseconds()/1e3) // 毫秒除以1000得秒数，不使用Second()因为它返回浮点

		result.setObject()
		in <- result
	}

	if err := d.do(ctx, fn, types.ContainerListOptions{All: containerAllForObject}); err != nil {
		l.Error(err)
	}
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
	tags := d.gatherSingleContainerInfo(container)

	var fields = make(map[string]interface{})
	var err error

	// 注意，此处如果没有 fields，构建 point 会失败
	// 需要在上层手动 addFiedls
	if container.State == "running" {
		fields, err = d.gatherSingleContainerStats(context.Background(), container)
		if err != nil {
			l.Error(err)
			return nil, err
		}
	}
	cost := time.Since(startTime)

	return &job{measurement: containerName, tags: tags, fields: fields, ts: time.Now(), cost: cost}, nil
}

func (d *dockerClient) ignoreImageNameFromContainer(name string) bool {
	return regexpMatchString(d.IgnoreImageName, name)
}

func (d *dockerClient) ignoreContainerNameFromContainer(name string) bool {
	return regexpMatchString(d.IgnoreContainerName, name)
}

func (d *dockerClient) gatherSingleContainerInfo(container types.Container) map[string]string {
	var tags = map[string]string{
		"container_id":   container.ID,
		"container_name": getContainerName(container.Names),
		"state":          container.State,
	}

	if !contianerIsFromKubernetes(getContainerName(container.Names)) {
		imageName, imageShortName, imageTag := ParseImage(container.Image)
		tags["docker_image"] = container.Image
		tags["image_name"] = imageName
		tags["image_short_name"] = imageShortName
		tags["image_tag"] = imageTag
		tags["container_type"] = "docker"
	} else {
		tags["container_type"] = "kubernetes"

	}

	if d.K8s != nil {
		name, _ := d.K8s.GetContainerPodName(container.ID)
		if name != "" {
			tags["pod_name"] = name
		}

		namespace, _ := d.K8s.GetContainerPodNamespace(container.ID)
		if namespace != "" {
			tags["pod_namespace"] = namespace
		}
	}

	return tags
}

const streamStats = false

func (d *dockerClient) gatherSingleContainerStats(ctx context.Context, container types.Container) (map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(ctx, apiTimeoutDuration)
	defer cancel()

	resp, err := d.client.ContainerStats(ctx, container.ID, streamStats)
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

func (d *dockerClient) getContainerHostname(ctx context.Context, id string) (string, error) {
	containerJson, err := d.client.ContainerInspect(context.Background(), id)
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
		// ParseImage() return imageName and imageVersion, discard imageVersion
		imageName, _, _ := ParseImage(container.Image)
		if d.ignoreImageNameFromContainer(imageName) ||
			d.ignoreContainerNameFromContainer(container.ID) {
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
				iod.FeedLastError(inputName, fmt.Sprintf("gather logging: %s", err.Error()))
			}
		}(container)
	}
}

func (d *dockerClient) getTags(container types.Container) map[string]string {
	imageName, _, _ := ParseImage(container.Image)
	tags := map[string]string{
		"container_name": containerName,
		"container_id":   container.ID,
		"image_name":     imageName,
	}

	if !contianerIsFromKubernetes(getContainerName(container.Names)) {
		imageName, _, _ := ParseImage(container.Image)
		tags["image_name"] = imageName
		tags["container_type"] = "docker"
	} else {
		tags["container_type"] = "kubernetes"
	}

	if d.K8s != nil {
		namespace, err := d.K8s.GetContainerPodNamespace(container.ID)
		if err != nil {
			l.Debugf("gather k8s pod error, %s", err)
		}
		if namespace != "" {
			tags["namespace"] = namespace
		}
	}
	if d.ProcessTags != nil {
		d.ProcessTags(tags)
	}
	return tags
}

func (d *dockerClient) getSource(container types.Container) string {
	// measurement 默认使用容器名，如果该容器是 k8s 创建，则尝试获取它的 work name（work-load）
	// 如果该字段值（即 source 参数）不为空，则使用
	var source = getContainerName(container.Names)

	if contianerIsFromKubernetes(getContainerName(container.Names)) && d.K8s != nil {
		name, err := d.K8s.GetContainerWorkname(container.ID)
		if err != nil {
		} else {
			source = name
		}
	}

	return source

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
		return d.tailStream(logReader, "tty", container)
	} else {
		return d.tailMultiplexed(logReader, container)
	}

}

func (d *dockerClient) tailStream(reader io.ReadCloser, stream string, container types.Container) error {
	defer reader.Close()

	var tags = d.getTags(container)
	tags["stream"] = stream

	var containerName = getContainerName(container.Names)
	var source = d.getSource(container)

	r := bufio.NewReaderSize(reader, maxLineBytes)

	for {
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

		measurement := source
		message := strings.TrimSpace(string(line))

		var fields = make(map[string]interface{})

		for _, lf := range d.LogFilters {
			if lf.MatchMessage(message) {
				if lf.Source != "" {
					measurement = lf.Source
				}

				var err error
				fields, err = lf.RunPipeline(message)
				if err != nil {
					l.Debug(err)
				}

				if lf.Service != "" {
					fields["service"] = lf.Service
				}
				break
			}
		}

		// 没有对应的 logFilters
		if len(fields) == 0 {
			fields["service"] = containerName
			fields["message"] = message
		}

		// l.Debugf("get %d bytes from source: %s", len(message), measurement)

		if err := checkFieldsLength(fields, maxFieldsLength); err != nil {
			// 只有在碰到非 message 字段，且长度超过最大限制时才会返回 error
			// 防止通过 pipeline 添加巨长字段的恶意行为
			l.Error(err)
			continue
		}

		addStatus(fields)

		// pipeline切割的日志时间
		ts, err := takeTime(fields)
		if err != nil {
			ts = time.Now()
			l.Error(err)
		}

		pt, err := iod.MakePoint(measurement, tags, fields, ts)
		if err != nil {
			l.Error(err)
		} else {
			if err := iod.Feed(inputName, datakit.Logging, []*iod.Point{pt}, &iod.Option{HighFreq: useIOHighFreq}); err != nil {
				l.Error("logging gather failed, container_id: %s, container_name:%s, err: %s", err.Error())
			}
		}
	}
}

func (d *dockerClient) tailMultiplexed(src io.ReadCloser, container types.Container) error {
	outReader, outWriter := io.Pipe()
	errReader, errWriter := io.Pipe()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		err := d.tailStream(outReader, "stdout", container)
		if err != nil {
			l.Error(err)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		err := d.tailStream(errReader, "stderr", container)
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
