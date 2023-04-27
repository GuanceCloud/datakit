// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package container

import (
	"context"
	"encoding/json"
	"path/filepath"
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

	logpath := logsJoinRootfs(inspect.LogPath)

	if !tailer.FileIsActive(logpath, ignoreDeadLogDuration) {
		l.Debugf("container %s file %s is not active, larger than %s, ignored", getContainerName(container.Names), logpath, ignoreDeadLogDuration)
		return nil
	}

	inspectJSON, _ := json.Marshal(inspect)
	l.Debugf("containerId: %s inspect %v", container.ID, string(inspectJSON))

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
		logPath:               logpath,
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

	{
		info.configKey = containerInsiderLogConfigKey
		opt, paths := composeTailerOption(d.k8sClient, info)
		opt.Mode = tailer.FileMode
		opt.BlockingMode = d.ipt.LoggingBlockingMode
		opt.MinFlushInterval = d.ipt.LoggingMinFlushInterval
		opt.MaxMultilineLifeDuration = d.ipt.LoggingMaxMultilineLifeDuration
		opt.Done = d.ipt.semStop.Wait()
		_ = opt.Init()

		l.Debugf("use container-log opt:%#v, paths:%#v, containerId: %s", opt, paths, container.ID)

		if len(paths) != 0 && info.tags["container_type"] == "docker" {
			if inspect.GraphDriver.Name == "overlay2" && inspect.GraphDriver.Data["MergedDir"] != "" {
				opt2 := deepCopyTailerOption(opt)
				opt2.Mode = tailer.FileMode
				_ = opt2.Init()
				tailMerged, err := tailer.NewTailer(completePaths(inspect.GraphDriver.Data["MergedDir"], paths), opt2)
				if err != nil {
					l.Warnf("failed to new paths, containerID: %s, source: %s, paths: %s, err: %s", container.ID, opt2.Source, paths, err)
					return err
				}
				go tailMerged.Start()
			}
		}
	}

	{
		info.configKey = containerLogConfigKey
		opt, _ := composeTailerOption(d.k8sClient, info)
		opt.Mode = tailer.DockerMode
		opt.BlockingMode = d.ipt.LoggingBlockingMode
		opt.MinFlushInterval = d.ipt.LoggingMinFlushInterval
		opt.MaxMultilineLifeDuration = d.ipt.LoggingMaxMultilineLifeDuration
		opt.Done = d.ipt.semStop.Wait()
		_ = opt.Init()
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
	}
	return nil
}

type containerLog struct{}

func (c *containerLog) LineProto() (*point.Point, error) { return nil, nil }

//nolint:lll
func (c *containerLog) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "Use Logging Source",
		Desc: "The logging of the container.",
		Type: "logging",
		Tags: map[string]interface{}{
			"container_id":           inputs.NewTagInfo(`Container ID`),
			"container_name":         inputs.NewTagInfo(`Container name from k8s (label 'io.kubernetes.container.name'). If empty then use $container_runtime_name.`),
			"container_runtime_name": inputs.NewTagInfo(`Container name from runtime (like 'docker ps'). If empty then use 'unknown' ([:octicons-tag-24: Version-1.4.6](changelog.md#cl-1.4.6)).`),
			"container_type":         inputs.NewTagInfo(`The type of the container (this container is created by kubernetes/docker/containerd).`),
			// "stream":                 inputs.NewTagInfo(`数据流方式，stdout/stderr/tty（containerd 日志缺少此字段）`),
			"pod_name":    inputs.NewTagInfo(`The pod name of the container (label 'io.kubernetes.pod.name').`),
			"namespace":   inputs.NewTagInfo(`The pod namespace of the container (label 'io.kubernetes.pod.namespace').`),
			"deployment":  inputs.NewTagInfo(`The deployment name of the container's pod (unsupported containerd).`),
			"service":     inputs.NewTagInfo("The name of the service, if `service` is empty then use `source`."),
			"[POD_LABEL]": inputs.NewTagInfo("The pod labels will be extracted as tags if `extract_k8s_label_as_tags` is enabled."),
		},
		Fields: map[string]interface{}{
			"status":          &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "The status of the logging, only supported info/emerg/alert/critical/error/warning/debug/OK/unknown."},
			"log_read_lines":  &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "The lines of the read file ([:octicons-tag-24: Version-1.4.6](changelog.md#cl-1.4.6))."},
			"log_read_offset": &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.UnknownUnit, Desc: "The offset of the read file ([:octicons-tag-24: Version-1.4.8](changelog.md#cl-1.4.8) · [:octicons-beaker-24: Experimental](index.md#experimental))."},
			"log_read_time":   &inputs.FieldInfo{DataType: inputs.DurationSecond, Unit: inputs.UnknownUnit, Desc: "The timestamp of the read file."},
			"message_length":  &inputs.FieldInfo{DataType: inputs.SizeByte, Unit: inputs.NCount, Desc: "The length of the message content."},
			"message":         &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "The text of the logging."},
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

func completePaths(target string, paths []string) []string {
	var res []string
	for _, path := range paths {
		res = append(res, filepath.Join(target, path))
	}
	return res
}

func deepCopyTailerOption(in *tailer.Option) *tailer.Option {
	out := &tailer.Option{}
	out.IgnoreStatus = in.IgnoreStatus
	out.Sockets = in.Sockets
	out.InputName = in.InputName
	out.Source = in.Source
	out.Service = in.Service
	out.Pipeline = in.Pipeline
	out.CharacterEncoding = in.CharacterEncoding
	out.MultilinePatterns = in.MultilinePatterns
	out.FromBeginning = in.FromBeginning
	out.RemoveAnsiEscapeCodes = in.RemoveAnsiEscapeCodes
	out.DisableAddStatusField = in.DisableAddStatusField
	out.DisableHighFreqIODdata = in.DisableHighFreqIODdata
	out.ForwardFunc = in.ForwardFunc
	out.IgnoreDeadLog = in.IgnoreDeadLog
	out.BlockingMode = in.BlockingMode
	out.MinFlushInterval = in.MinFlushInterval
	out.MaxMultilineLifeDuration = in.MaxMultilineLifeDuration

	out.GlobalTags = make(map[string]string)
	for k, v := range in.GlobalTags {
		out.GlobalTags[k] = v
	}

	return out
}

//nolint:gochecknoinits
func init() {
	registerMeasurement(&containerLog{})
}
