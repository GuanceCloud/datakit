// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package container

import (
	"context"
	"fmt"
	"path/filepath"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/container/runtime"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/goroutine"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/tailer"
	apicorev1 "k8s.io/api/core/v1"
)

const emptyDirMountToHost = "/var/lib/kubelet/pods/%s/volumes/kubernetes.io~empty-dir/%s/"

func (c *container) cleanMissingContainerLog(newIDs []string) {
	missingIDs := c.logTable.findDifferences(newIDs)
	for _, id := range missingIDs {
		l.Infof("clean log collection for container id %s", id)
		c.logTable.closeFromTable(id)
		c.logTable.removeFromTable(id)
	}
}

func (c *container) tailingLogs(ins *logInstance) {
	g := goroutine.NewGroup(goroutine.Option{Name: "container-logs/" + ins.containerName})
	done := make(chan interface{})

	for _, cfg := range ins.configs {
		if cfg.Disable {
			continue
		}

		if c.logTable.inTable(ins.id, cfg.Path) {
			continue
		}

		path := cfg.Path
		if cfg.TargetPath != "" {
			path = cfg.TargetPath
			l.Infof("container log %s redirect to host path %s", cfg.Path, cfg.TargetPath)
		}

		mergedTags := inputs.MergeTags(c.extraTags, cfg.Tags, "")

		opt := &tailer.Option{
			Source:                   cfg.Source,
			Service:                  cfg.Service,
			Pipeline:                 cfg.Pipeline,
			CharacterEncoding:        cfg.CharacterEncoding,
			MultilinePatterns:        cfg.MultilinePatterns,
			GlobalTags:               mergedTags,
			BlockingMode:             c.ipt.LoggingBlockingMode,
			MinFlushInterval:         c.ipt.LoggingMinFlushInterval,
			MaxMultilineLifeDuration: c.ipt.LoggingMaxMultilineLifeDuration,
			RemoveAnsiEscapeCodes:    c.ipt.LoggingRemoveAnsiEscapeCodes,
			Done:                     done,
		}

		switch cfg.Type {
		case "file":
			opt.Mode = tailer.FileMode
		case runtime.DockerRuntime:
			opt.Mode = tailer.DockerMode
		default:
			opt.Mode = tailer.ContainerdMode
		}

		_ = opt.Init()

		path = logsJoinRootfs(path)

		l.Infof("add container log collection with path %s(%s) from source %s", cfg.Path, path, opt.Source)

		tail, err := tailer.NewTailerSingle(path, opt)
		if err != nil {
			l.Errorf("failed to create container-log collection %s(%s) for %s, err: %s", cfg.Path, path, ins.containerName, err)
			continue
		}

		c.logTable.addToTable(ins.id, cfg.Path, done)

		g.Go(func(ctx context.Context) error {
			defer func() {
				c.logTable.removePathFromTable(ins.id, path)
				l.Infof("remove container log collection from source %s", opt.Source)
			}()
			tail.Run()
			return nil
		})
	}
}

func (c *container) queryContainerLogInfo(info *runtime.Container) *logInstance {
	podName := getPodNameForLabels(info.Labels)
	podNamespace := getPodNamespaceForLabels(info.Labels)

	ins := &logInstance{
		id:            info.ID,
		containerName: info.Name,
		image:         info.Image,
		logPath:       info.LogPath,
		podName:       podName,
		podNamespace:  podNamespace,
		volMounts:     info.Mounts,
	}

	if name := getContainerNameForLabels(info.Labels); name != "" {
		ins.containerName = name
	}

	if c.k8sClient != nil && podName != "" {
		podInfo, err := c.queryPodInfo(context.Background(), podName, podNamespace)
		if err != nil {
			l.Warn(err)
		} else {
			// ex: datakit/logs
			if v := podInfo.pod.Annotations[fmt.Sprintf(logConfigAnnotationKeyFormat, "")]; v != "" {
				ins.configStr = v
			}

			// ex: datakit/nginx.logs
			if v := podInfo.pod.Annotations[fmt.Sprintf(logConfigAnnotationKeyFormat, ins.containerName+".")]; v != "" {
				ins.configStr = v
			}

			ins.podLabels = podInfo.pod.Labels
			ins.ownerKind, ins.ownerName = podInfo.owner()

			// use Image from Pod Container
			ins.image = podInfo.containerImage(ins.containerName)

			for _, volume := range podInfo.pod.Spec.Volumes {
				if volume.EmptyDir == nil {
					continue
				}

				mountPath := findContainerVolumeMount(podInfo.pod.Spec.Containers, volume.Name)
				if mountPath != "" {
					ins.volMounts[filepath.Clean(mountPath)] = fmt.Sprintf(emptyDirMountToHost, string(podInfo.pod.UID), volume.Name)
				}
			}
		}
	}

	// ex: DATAKIT_LOGS_CONFIG
	if info.Envs != nil {
		if str, ok := info.Envs["DATAKIT_LOGS_CONFIG"]; ok {
			ins.configStr = str
		}
	}

	l.Debugf("container %s use config: %v", ins.containerName, ins)
	return ins
}

func findContainerVolumeMount(containers []apicorev1.Container, mountName string) string {
	for _, container := range containers {
		for _, mount := range container.VolumeMounts {
			if mount.Name == mountName {
				return mount.MountPath
			}
		}
	}
	return ""
}
