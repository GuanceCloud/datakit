package container

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/pkg/stdcopy"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	iod "gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/worker"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

const (
	// Maximum bytes of a log line before it will be split, size is mirroring
	// docker code:
	// https://github.com/moby/moby/blob/master/daemon/logger/copier.go#L21
	maxLineBytes = 16 * 1024
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

func (d *dockerInput) tailContainerLogs(ctx context.Context, container *types.Container) error {
	hasTTY, err := d.hasTTY(ctx, container.ID)
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

func (d *dockerInput) hasTTY(ctx context.Context, containerID string) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, apiTimeoutDuration)
	defer cancel()
	c, err := d.client.ContainerInspect(ctx, containerID)
	if err != nil {
		return false, err
	}
	return c.Config.Tty, nil
}

func (d *dockerInput) tailMultiplexed(ctx context.Context, src io.ReadCloser, container *types.Container) error {
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

func (d *dockerInput) tailStream(ctx context.Context, reader io.ReadCloser, stream string, container *types.Container) error {
	defer reader.Close() //nolint:errcheck

	tags := getContainerInfo(container, d.k8sClient)
	tags["stream"] = stream

	task := &worker.Task{
		TaskName: "log-" + tags["image_short_name"],
		Source:   getContainerLogSource(container.Image),
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
		task.Source = logconf.Source
		task.ScriptName = logconf.Pipeline
		tags["service"] = logconf.Service
	}

	// add extra tags
	for k, v := range d.cfg.extraTags {
		tags[k] = v
	}

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

		lineStr := Bytes2String(removeAnsiEscapeCodes(line, d.cfg.removeLoggingAnsiCodes))

		task.Data = []worker.TaskData{
			&taskData{tags: tags, log: lineStr},
		}
		task.TS = time.Now()

		if err := worker.FeedPipelineTask(task); err != nil {
			l.Errorf("feed log error: %w", err)
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
	_, imageShortName, imageVersion := ParseImage(image)

	if strings.HasPrefix(imageShortName, "sha256") {
		if len(imageVersion) > 12 {
			return "image_id_" + imageVersion[:12]
		}
		return "image_id_" + imageVersion
	}

	if imageShortName != "" {
		return imageShortName
	}

	return "default"
}

//nolint:gochecknoinits
func init() {
	registerMeasurement(&containerLog{})
}
