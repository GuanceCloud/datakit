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
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

func (d *dockerInput) addToContainerList(containerID string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.containerLogList[containerID] = nil
}

func (d *dockerInput) removeFromContainerList(containerID string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	delete(d.containerLogList, containerID)
}

func (d *dockerInput) containerInContainerList(containerID string) bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	_, ok := d.containerLogList[containerID]
	return ok
}

func (d *dockerInput) tailingLog(ctx context.Context, container *types.Container) error {
	inspect, err := d.client.ContainerInspect(ctx, container.ID)
	if err != nil {
		return err
	}

	if !tailer.FileIsActive(inspect.LogPath, ignoreDeadLogDuration) {
		l.Infof("container %s file %s is not active, larger than %s, ignored", getContainerName(container.Names), inspect.LogPath, ignoreDeadLogDuration)
		return nil
	}

	image := container.Image

	if d.k8sClient != nil {
		podname := getPodNameForLabels(container.Labels)
		podnamespace := getPodNamespaceForLabels(container.Labels)
		podContainerName := getContainerNameForLabels(container.Labels)

		meta, err := queryPodMetaData(d.k8sClient, podname, podnamespace)
		if err == nil {
			image = meta.containerImage(podContainerName)
		}
	}

	info := &containerLogBasisInfo{
		name:                  getContainerName(container.Names),
		id:                    container.ID,
		logPath:               inspect.LogPath,
		labels:                container.Labels,
		image:                 image,
		tags:                  make(map[string]string),
		created:               inspect.Created,
		extraSourceMap:        d.ipt.LoggingExtraSourceMap,
		sourceMultilineMap:    d.ipt.LoggingSourceMultilineMap,
		autoMultilinePatterns: d.ipt.getAutoMultilinePatterns(),
		extractK8sLabelAsTags: d.ipt.ExtractK8sLabelAsTags,
	}

	if containerIsFromKubernetes(getContainerName(container.Names)) {
		info.tags["container_type"] = "kubernetes"
	} else {
		info.tags["container_type"] = "docker"
	}

	// add extra tags
	for k, v := range d.ipt.Tags {
		if _, ok := info.tags[k]; !ok {
			info.tags[k] = v
		}
	}

	opt := composeTailerOption(d.k8sClient, info)
	opt.Mode = tailer.DockerMode
	opt.BlockingMode = d.ipt.LoggingBlockingMode
	opt.MinFlushInterval = d.ipt.LoggingMinFlushInterval
	opt.MaxMultilineLifeDuration = d.ipt.LoggingMaxMultilineLifeDuration
	opt.Done = d.ipt.semStop.Wait()
	_ = opt.Init()

	l.Debugf("use container-log opt:%#v, containerId: %s", opt, container.ID)

	t, err := tailer.NewTailerSingle(info.logPath, opt)
	if err != nil {
		l.Warnf("failed to new docker log, containerID: %s, source: %s, logpath: %s, err: %s", container.ID, opt.Source, info.logPath, err)
		return err
	}

	d.addToContainerList(container.ID)
	l.Infof("add docker log, containerId: %s, source: %s, logpath: %s", container.ID, opt.Source, info.logPath)
	defer func() {
		d.removeFromContainerList(container.ID)
		l.Debugf("remove docker log, containerName: %s image: %s", getContainerName(container.Names), container.Image)
	}()

	t.Run()
	return nil
}

type containerLog struct{}

func (c *containerLog) LineProto() (*point.Point, error) { return nil, nil }

//nolint:lll
func (c *containerLog) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "容器日志",
		Type: "logging",
		Tags: map[string]interface{}{
			"container_name":         inputs.NewTagInfo(`k8s 命名的容器名（在 labels 中取 'io.kubernetes.container.name'），如果值为空则跟 container_runtime_name 相同`),
			"container_runtime_name": inputs.NewTagInfo(`由 runtime 命名的容器名（例如 docker ps 查看），如果值为空则默认是 unknown（[:octicons-tag-24: Version-1.4.6](changelog.md#cl-1.4.6)）`),
			"container_id":           inputs.NewTagInfo(`容器ID`),
			"container_type":         inputs.NewTagInfo(`容器类型，表明该容器由谁创建，kubernetes/docker`),
			// "stream":                 inputs.NewTagInfo(`数据流方式，stdout/stderr/tty（containerd 日志缺少此字段）`),
			"pod_name":    inputs.NewTagInfo(`pod 名称（容器由 k8s 创建时存在）`),
			"namespace":   inputs.NewTagInfo(`pod 的 k8s 命名空间（k8s 创建容器时，会打上一个形如 'io.kubernetes.pod.namespace' 的 label，DataKit 将其命名为 'namespace'）`),
			"deployment":  inputs.NewTagInfo(`deployment 名称（容器由 k8s 创建时存在，containerd 日志缺少此字段）`),
			"service":     inputs.NewTagInfo(`服务名称`),
			"[POD_LABEL]": inputs.NewTagInfo("如果该容器是由 k8s 创建，且配置参数 `extract_k8s_label_as_tags` 开启，则会将 pod 的 label 添加至标签中"),
		},
		Fields: map[string]interface{}{
			"status":          &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "日志状态，info/emerg/alert/critical/error/warning/debug/OK/unknown"},
			"log_read_lines":  &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "采集到的行数计数，多行数据算成一行（[:octicons-tag-24: Version-1.4.6](changelog.md#cl-1.4.6)）"},
			"log_read_offset": &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.UnknownUnit, Desc: "当前数据在文件中的偏移位置（[:octicons-tag-24: Version-1.4.8](changelog.md#cl-1.4.8) · [:octicons-beaker-24: Experimental](index.md#experimental)）"},
			"log_read_time":   &inputs.FieldInfo{DataType: inputs.DurationSecond, Unit: inputs.UnknownUnit, Desc: "数据从文件中读取到的这一刻的时间戳，单位是秒"},
			"message_length":  &inputs.FieldInfo{DataType: inputs.SizeByte, Unit: inputs.NCount, Desc: "message 字段的长度，单位字节"},
			"message":         &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "日志源数据"},
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
