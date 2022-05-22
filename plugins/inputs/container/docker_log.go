// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package container

import (
	"context"
	"encoding/json"
	"time"

	"github.com/docker/docker/api/types"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/tailer"
	iod "gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

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
	if n := getContainerNameForLabels(container.Labels); n != "" {
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

	if logconf == nil {
		logconf = &containerLogConfig{}
	}
	if logconf.Source == "" {
		logconf.Source = source
	}
	if logconf.Service == "" {
		logconf.Service = logconf.Source
	}

	logconf.tags = tags
	logconf.containerID = container.ID

	l.Debugf("use container logconfig:%#v, containerName:%s", logconf, tags["container_name"])

	inspect, err := d.client.ContainerInspect(ctx, container.ID)
	if err != nil {
		return err
	}
	logconf.logpath = inspect.LogPath
	logconf.created = inspect.Created

	return d.tailContainerLog(logconf)
}

func (d *dockerInput) tailContainerLog(logconf *containerLogConfig) error {
	opt := &tailer.Option{
		Source:         logconf.Source,
		Service:        logconf.Service,
		Pipeline:       logconf.Pipeline,
		MultilineMatch: logconf.Multiline,
		GlobalTags:     logconf.tags,
		DockerMode:     true,
	}
	if !checkContainerIsOlder(logconf.created, time.Minute) {
		opt.FromBeginning = true
	}

	l.Debugf("use container logconfig:%#v, containerID: %s, source: %s, logpath: %s", logconf, logconf.containerID, opt.Source, logconf.logpath)

	_ = opt.Init()

	t, err := tailer.NewTailerSingle(logconf.logpath, opt)
	if err != nil {
		l.Errorf("failed to new containerd log, containerID: %s, source: %s, logpath: %s", logconf.containerID, opt.Source, logconf.logpath)
		return err
	}

	d.addToContainerList(logconf.containerID, t.Close)

	l.Infof("add containerd log, containerID: %s, source: %s, logpath: %s", logconf.containerID, opt.Source, logconf.logpath)

	t.Run()
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
	created     string
	logpath     string
	tags        map[string]string
}

const containerLogConfigKey = "datakit/logs"

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
			"stream":         inputs.NewTagInfo(`数据流方式，stdout/stderr/tty（containerd 日志缺少此字段）`),
			"pod_name":       inputs.NewTagInfo(`pod 名称（容器由 k8s 创建时存在）`),
			"namespace":      inputs.NewTagInfo(`pod 的 k8s 命名空间（k8s 创建容器时，会打上一个形如 'io.kubernetes.pod.namespace' 的 label，DataKit 将其命名为 'namespace'）`),
			"deployment":     inputs.NewTagInfo(`deployment 名称（容器由 k8s 创建时存在，containerd 日志缺少此字段）`),
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

func checkContainerIsOlder(createdTime string, limit time.Duration) bool {
	t, err := time.Parse(time.RFC3339, createdTime)
	if err != nil {
		return false
	}
	return time.Since(t) > limit
}

//nolint:gochecknoinits
func init() {
	registerMeasurement(&containerLog{})
}
