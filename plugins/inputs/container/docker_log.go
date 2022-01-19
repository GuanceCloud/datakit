package container

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"sync"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/pkg/stdcopy"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/readbuf"
	iod "gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/worker"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

const (
	// Maximum bytes of a log line before it will be split, size is mirroring
	// docker code:
	// https://github.com/moby/moby/blob/master/daemon/logger/copier.go#L21
	// maxLineBytes = 16 * 1024.

	readBuffSize = 2 * 1024
)

var containerLogsOptions = types.ContainerLogsOptions{
	ShowStdout: true,
	ShowStderr: true,
	Follow:     true,
	Tail:       "0", // 默认关闭FromBeginning，避免数据量巨大。开启为 'all'
}

func (d *dockerInput) addToContainerList(containerID string, cancel context.CancelFunc) {
	d.containerLogList[containerID] = cancel
}

func (d *dockerInput) removeFromContainerList(containerID string) {
	delete(d.containerLogList, containerID)
}

func (d *dockerInput) containerInContainerList(containerID string) bool {
	_, ok := d.containerLogList[containerID]
	return ok
}

func (d *dockerInput) cancelTails() {
	for _, cancel := range d.containerLogList {
		cancel()
	}
}

func (d *dockerInput) watchingContainerLogs(ctx context.Context, container *types.Container) error {
	tags := getContainerInfo(container, d.k8sClient)
	// add extra tags
	for k, v := range d.cfg.extraTags {
		if _, ok := tags[k]; !ok {
			tags[k] = v
		}
	}

	logconf := func() *containerLogConfig {
		if datakit.Docker && tags["pod_name"] != "" {
			return getContainerLogConfigForK8s(d.k8sClient, tags["pod_name"], tags["pod_namesapce"])
		}
		return getContainerLogConfigForDocker(container.Labels)
	}()

	if logconf != nil {
		l.Debugf("use contaier logconfig %#v, container_name:%s", logconf, tags["container_name"])
		if logconf.Disable {
			l.Debug("disable contaier log, container_name:%s pod_name:%s", tags["container_name"], tags["pod_name"])
			return nil
		}

		logconf.tags = tags
		logconf.containerID = container.ID
	} else {
		logconf = &containerLogConfig{
			Source:      getContainerLogSource(tags["image_short_name"]),
			Service:     tags["image_short_name"],
			tags:        tags,
			containerID: container.ID,
		}
	}

	return d.tailContainerLogs(ctx, logconf)
}

func (d *dockerInput) tailContainerLogs(ctx context.Context, logconf *containerLogConfig) error {
	hasTTY, err := d.hasTTY(ctx, logconf.containerID)
	if err != nil {
		return err
	}

	logReader, err := d.client.ContainerLogs(ctx, logconf.containerID, containerLogsOptions)
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
		return d.tailStream(ctx, logReader, "tty", logconf)
	} else {
		return d.tailMultiplexed(ctx, logReader, logconf)
	}
}

func (d *dockerInput) hasTTY(ctx context.Context, containerID string) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, apiTimeoutDuration)
	defer cancel()
	c, err := d.client.ContainerInspect(ctx, containerID)
	if err != nil {
		return false, err
	}
	return c.Config.Tty, nil
}

func (d *dockerInput) tailMultiplexed(ctx context.Context, src io.ReadCloser, logconf *containerLogConfig) error {
	outReader, outWriter := io.Pipe()
	errReader, errWriter := io.Pipe()

	// 避免goroutine共享数据，clone一份
	logconfStderr := logconf.clone()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		err := d.tailStream(ctx, outReader, "stdout", logconf)
		if err != nil {
			l.Error(err)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		err := d.tailStream(ctx, errReader, "stderr", logconfStderr)
		if err != nil {
			l.Error(err)
		}
	}()

	defer func() {
		wg.Wait()

		if err := outWriter.Close(); err != nil {
			l.Warnf("Close: %s", err)
		}

		if err := errWriter.Close(); err != nil {
			l.Warnf("Close: %s", err)
		}

		if err := src.Close(); err != nil {
			l.Warnf("Close: %s", err)
		}
	}()

	_, err := stdcopy.StdCopy(outWriter, errWriter, src)
	if err != nil {
		l.Warnf("StdCopy: %s", err)
		return err
	}

	return nil
}

type containerLogConfig struct {
	Disable  bool   `json:"disable"`
	Source   string `json:"source"`
	Pipeline string `json:"pipeline"`
	Service  string `json:"service"`

	containerID string
	tags        map[string]string
}

func (c *containerLogConfig) clone() *containerLogConfig {
	t := make(map[string]string)
	for k, v := range c.tags {
		t[k] = v
	}
	return &containerLogConfig{
		Disable:     c.Disable,
		Source:      c.Source,
		Pipeline:    c.Pipeline,
		Service:     c.Service,
		containerID: c.containerID,
		tags:        t,
	}
}

const (
	containerLableForPodName      = "io.kubernetes.pod.name"
	containerLableForPodNamespace = "io.kubernetes.pod.namespace"
	containerLogConfigKey         = "datakit/logs"
)

func getContainerLogConfig(m map[string]string) (*containerLogConfig, error) {
	configStr := m[containerLogConfigKey]
	if configStr == "" {
		return nil, nil
	}

	var configs []containerLogConfig
	if err := json.Unmarshal([]byte(configStr), &configs); err != nil {
		return nil, err
	}

	if len(configs) < 1 {
		return nil, nil
	}

	temp := configs[0]
	return &temp, nil
}

func getContainerLogConfigForK8s(k8sClient k8sClientX, podname, podnamespace string) *containerLogConfig {
	annotations, err := getPodAnnotations(k8sClient, podname, podnamespace)
	if err != nil {
		l.Errorf("failed to get pod annotations, %w", err)
		return nil
	}

	c, err := getContainerLogConfig(annotations)
	if err != nil {
		l.Errorf("failed to get container logConfig: %w", err)
		return nil
	}
	return c
}

func getContainerLogConfigForDocker(labels map[string]string) *containerLogConfig {
	c, err := getContainerLogConfig(labels)
	if err != nil {
		l.Errorf("failed to get container logConfig: %w", err)
		return nil
	}
	return c
}

func (d *dockerInput) tailStream(ctx context.Context, reader io.ReadCloser, stream string, logconf *containerLogConfig) error {
	defer reader.Close() //nolint:errcheck

	logconf.tags["stream"] = stream
	shortImageName := logconf.tags["image_short_name"]

	r := readbuf.NewReadBuffer(reader, readBuffSize)

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			// nil
		}

		lines, err := r.ReadLines()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return err
		}

		if len(lines) == 0 {
			// 如果没有数据，那就等待1秒，避免频繁read
			time.Sleep(time.Second)
			continue
		}

		workerData := []worker.TaskData{}

		for _, line := range lines {
			if len(line) == 0 {
				continue
			}
			workerData = append(workerData,
				&taskData{
					tags: logconf.tags,
					log:  string(removeAnsiEscapeCodes(line, d.cfg.removeLoggingAnsiCodes)),
				},
			)
		}

		task := &worker.Task{
			TaskName:   "containerlogging/" + shortImageName,
			Source:     logconf.Source,
			ScriptName: logconf.Pipeline,
			Data:       workerData,
			TS:         time.Now(),
		}

		if err := worker.FeedPipelineTaskBlock(task); err != nil {
			l.Errorf("failed to fedd log, containerName: %s, err: %w", logconf.tags["container_name"], err)
		}
	}
}

type containerLog struct{}

func (c *containerLog) LineProto() (*iod.Point, error) { return nil, nil }

//nolint:lll
func (c *containerLog) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "日志数据",
		Type: "logging",
		Desc: "source 默认使用容器 image_short_name",
		Tags: map[string]interface{}{
			"container_name": inputs.NewTagInfo(`容器名称`),
			"container_id":   inputs.NewTagInfo(`容器ID`),
			"container_type": inputs.NewTagInfo(`容器类型，表明该容器由谁创建，kubernetes/docker`),
			"stream":         inputs.NewTagInfo(`数据流方式，stdout/stderr/tty`),
			"pod_name":       inputs.NewTagInfo(`pod 名称（容器由 k8s 创建时存在）`),
			"pod_namesapce":  inputs.NewTagInfo(`pod 命名空间（容器由 k8s 创建时存在）`),
			"deployment":     inputs.NewTagInfo(`deployment 名称（容器由 k8s 创建时存在）`),
			"service":        inputs.NewTagInfo(`服务名称`),
		},
		Fields: map[string]interface{}{
			"status":  &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "日志状态，info/emerg/alert/critical/error/warning/debug/OK"},
			"message": &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "日志源数据"},
		},
	}
}

type taskData struct {
	tags map[string]string
	log  string
}

func (t *taskData) GetContent() string {
	return t.log
}

func (t *taskData) Handler(r *worker.Result) error {
	for k, v := range t.tags {
		if _, err := r.GetTag(k); err != nil {
			r.SetTag(k, v)
		}
	}
	return nil
}

func getContainerLogSource(image string) string {
	if image != "" {
		return image
	}
	return "default"
}

//nolint:gochecknoinits
func init() {
	registerMeasurement(&containerLog{})
}
