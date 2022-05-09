// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

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
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/multiline"
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

func (d *dockerInput) watchingContainerLog(ctx context.Context, container *types.Container) error {
	tags := getContainerInfo(container, d.k8sClient)

	source := getContainerLogSource(tags["image_short_name"])
	if n := container.Labels[containerLableForPodContainerName]; n != "" {
		source = n
	}

	// add extra tags
	for k, v := range d.cfg.extraTags {
		if _, ok := tags[k]; !ok {
			tags[k] = v
		}
	}

	logconf := func() *containerLogConfig {
		if datakit.Docker && tags["pod_name"] != "" {
			return getContainerLogConfigForK8s(d.k8sClient, tags["pod_name"], tags["namespace"])
		}
		return getContainerLogConfigForDocker(container.Labels)
	}()

	if logconf != nil {
		if logconf.Source == "" {
			logconf.Source = source
		}
		if logconf.Service == "" {
			logconf.Service = logconf.Source
		}
		logconf.tags = tags
		logconf.containerID = container.ID
		l.Debugf("use container logconfig:%#v, containerName:%s", logconf, tags["container_name"])
	} else {
		logconf = &containerLogConfig{
			Source:      source,
			Service:     source,
			tags:        tags,
			containerID: container.ID,
		}
	}

	if err := logconf.checking(); err != nil {
		return err
	}

	return d.tailContainerLog(ctx, logconf)
}

func (d *dockerInput) tailContainerLog(ctx context.Context, logconf *containerLogConfig) error {
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
	ctx, cancel := context.WithTimeout(ctx, timeoutDuration)
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
	Disable    bool     `json:"disable"`
	Source     string   `json:"source"`
	Pipeline   string   `json:"pipeline"`
	Service    string   `json:"service"`
	Multiline  string   `json:"multiline_match"`
	OnlyImages []string `json:"only_images"`

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
		Multiline:   c.Multiline,
		containerID: c.containerID,
		tags:        t,
	}
}

// multiline maxLines.
const maxLines = 1000

func (c *containerLogConfig) checking() error {
	if c.Multiline == "" {
		return nil
	}

	_, err := multiline.New(c.Multiline, maxLines)
	return err
}

const (
	containerLableForPodName          = "io.kubernetes.pod.name"
	containerLableForPodNamespace     = "io.kubernetes.pod.namespace"
	containerLableForPodContainerName = "io.kubernetes.container.name"
	containerLogConfigKey             = "datakit/logs"
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

	logconf.tags["service"] = logconf.Service
	logconf.tags["stream"] = stream

	containerName := logconf.tags["container_name"]

	mult, err := multiline.New(logconf.Multiline, maxLines)
	if err != nil {
		// unreachable
		return err
	}

	newTask := func() *worker.TaskTemplate {
		return &worker.TaskTemplate{
			TaskName:        "containerlog/" + logconf.Source,
			Source:          logconf.Source,
			ScriptName:      logconf.Pipeline,
			MaxMessageLen:   d.cfg.maxLoggingLength,
			Tags:            logconf.tags,
			ContentDataType: worker.ContentString,
			TS:              time.Now(),
		}
	}

	r := readbuf.NewReadBuffer(reader, readBuffSize)

	timeout := time.NewTicker(timeoutDuration)
	defer timeout.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-timeout.C:
			if text := mult.Flush(); len(text) != 0 {
				task := newTask()
				task.Content = []string{string(removeAnsiEscapeCodes(text, d.cfg.removeLoggingAnsiCodes))}
				if err := worker.FeedPipelineTaskBlock(task); err != nil {
					l.Errorf("failed to feed log, containerName:%s, err:%w", containerName, err)
				}
			}
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

		// 接收到数据，重置 ticker
		timeout.Reset(timeoutDuration)
		content := []string{}

		for _, line := range lines {
			if len(line) == 0 {
				continue
			}

			text := mult.ProcessLine(line)
			if len(text) == 0 {
				continue
			}

			content = append(content,
				string(removeAnsiEscapeCodes(text, d.cfg.removeLoggingAnsiCodes)),
			)
		}

		if len(content) == 0 {
			continue
		}

		task := newTask()
		task.Content = content

		if err := worker.FeedPipelineTaskBlock(task); err != nil {
			l.Errorf("failed to feed log, containerName:%s, err:%w", containerName, err)
		}
	}
}

type containerLog struct{}

func (c *containerLog) LineProto() (*iod.Point, error) { return nil, nil }

//nolint:lll
func (c *containerLog) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "容器日志",
		Type: "logging",
		Desc: "日志来源设置，参见[这里](container#6de978c3)",
		Tags: map[string]interface{}{
			"container_name": inputs.NewTagInfo(`容器名称`),
			"container_id":   inputs.NewTagInfo(`容器ID`),
			"container_type": inputs.NewTagInfo(`容器类型，表明该容器由谁创建，kubernetes/docker`),
			"stream":         inputs.NewTagInfo(`数据流方式，stdout/stderr/tty`),
			"pod_name":       inputs.NewTagInfo(`pod 名称（容器由 k8s 创建时存在）`),
			"namespace":      inputs.NewTagInfo(`pod 的 k8s 命名空间（k8s 创建容器时，会打上一个形如 'io.kubernetes.pod.namespace' 的 label，DataKit 将其命名为 'namespace'）`),
			"deployment":     inputs.NewTagInfo(`deployment 名称（容器由 k8s 创建时存在）`),
			"service":        inputs.NewTagInfo(`服务名称`),
		},
		Fields: map[string]interface{}{
			"status":  &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "日志状态，info/emerg/alert/critical/error/warning/debug/OK/unknown"},
			"message": &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "日志源数据"},
		},
	}
}

func getContainerLogSource(image string) string {
	if image != "" {
		return image
	}
	return "unknown"
}

//nolint:gochecknoinits
func init() {
	registerMeasurement(&containerLog{})
}
