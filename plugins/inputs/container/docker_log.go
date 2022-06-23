// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package container

import (
	"context"
	"time"

	"github.com/docker/docker/api/types"
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

func (d *dockerInput) tailingLog(ctx context.Context, container *types.Container) error {
	inspect, err := d.client.ContainerInspect(ctx, container.ID)
	if err != nil {
		return err
	}

	info := &containerLogBasisInfo{
		name:    getContainerName(container.Names),
		id:      container.ID,
		logPath: inspect.LogPath,
		labels:  container.Labels,
		image:   container.Image,
		tags:    make(map[string]string),
		created: inspect.Created,
	}

	if containerIsFromKubernetes(getContainerName(container.Names)) {
		info.tags["container_type"] = "kubernetes"
	} else {
		info.tags["container_type"] = "docker"
	}

	// add extra tags
	for k, v := range d.cfg.extraTags {
		if _, ok := info.tags[k]; !ok {
			info.tags[k] = v
		}
	}

	opt := composeTailerOption(d.k8sClient, info)
	opt.Mode = tailer.DockerMode

	t, err := tailer.NewTailerSingle(info.logPath, opt)
	if err != nil {
		l.Warnf("failed to new docker log, containerID: %s, source: %s, logpath: %s, err: %s", container.ID, opt.Source, info.logPath, err)
		return err
	}

	d.addToContainerList(container.ID, t.Close)
	l.Infof("add docker log, containerId: %s, source: %s, logpath: %s", container.ID, opt.Source, info.logPath)
	defer func() {
		d.removeFromContainerList(container.ID)
		l.Debugf("remove docker log, containerName: %s image: %s", getContainerName(container.Names), container.Image)
	}()

	t.Run()
	return nil
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
			"container_name":         inputs.NewTagInfo(`k8s 命名的容器名（在 labels 中取 'io.kubernetes.container.name'），如果值为空则跟 container_runtime_name 相同`),
			"container_runtime_name": inputs.NewTagInfo(`由 runtime 命名的容器名（例如 docker ps 查看），如果值为空则默认是 unknown`),
			"container_id":           inputs.NewTagInfo(`容器ID`),
			"container_type":         inputs.NewTagInfo(`容器类型，表明该容器由谁创建，kubernetes/docker`),
			// "stream":                 inputs.NewTagInfo(`数据流方式，stdout/stderr/tty（containerd 日志缺少此字段）`),
			"pod_name":   inputs.NewTagInfo(`pod 名称（容器由 k8s 创建时存在）`),
			"namespace":  inputs.NewTagInfo(`pod 的 k8s 命名空间（k8s 创建容器时，会打上一个形如 'io.kubernetes.pod.namespace' 的 label，DataKit 将其命名为 'namespace'）`),
			"deployment": inputs.NewTagInfo(`deployment 名称（容器由 k8s 创建时存在，containerd 日志缺少此字段）`),
			"service":    inputs.NewTagInfo(`服务名称`),
		},
		Fields: map[string]interface{}{
			"status":  &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "日志状态，info/emerg/alert/critical/error/warning/debug/OK/unknown"},
			"message": &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "日志源数据"},
		},
	}
}

func checkContainerIsOlder(createdTime string, limit time.Duration) bool {
	// default older
	if createdTime == "" {
		return true
	}
	t, err := time.Parse(time.RFC3339, createdTime)
	if err != nil {
		return true
	}
	return time.Since(t) > limit
}

//nolint:gochecknoinits
func init() {
	registerMeasurement(&containerLog{})
}
