// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package container

import (
	"context"
	"fmt"
	"strings"

	"github.com/docker/docker/api/types"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/goroutine"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/tailer"
)

func (d *dockerInput) tailingLogs(info *containerLogInfo) {
	g := goroutine.NewGroup(goroutine.Option{Name: "docker-logs/" + info.containerName})
	done := make(chan interface{})

	for _, cfg := range info.logConfigs {
		if cfg.Disable {
			continue
		}

		if d.logTable.inTable(info.id, cfg.Path) {
			continue
		}

		opt := &tailer.Option{
			Source:                   cfg.Source,
			Service:                  cfg.Service,
			Pipeline:                 cfg.Pipeline,
			CharacterEncoding:        cfg.CharacterEncoding,
			MultilinePatterns:        cfg.MultilinePatterns,
			GlobalTags:               cfg.Tags,
			BlockingMode:             d.ipt.LoggingBlockingMode,
			MinFlushInterval:         d.ipt.LoggingMinFlushInterval,
			MaxMultilineLifeDuration: d.ipt.LoggingMaxMultilineLifeDuration,
			Done:                     done,
		}

		if cfg.Type == "file" {
			opt.Mode = tailer.FileMode
		} else {
			opt.Mode = tailer.DockerMode
		}
		_ = opt.Init()

		path := logsJoinRootfs(cfg.Path)
		l.Debugf("tailer option: %v, path: %s", opt, path)

		tail, err := tailer.NewTailerSingle(path, opt)
		if err != nil {
			l.Errorf("failed to create docker-log collection %s for %s, err: %s", path, info.containerName, err)
			continue
		}

		d.logTable.addToTable(info.id, cfg.Path, done)

		g.Go(func(ctx context.Context) error {
			defer d.logTable.removePathFromTable(info.id, cfg.Path)
			tail.Run()
			return nil
		})
	}
}

func (d *dockerInput) queryContainerLogInfo(ctx context.Context, container *types.Container) *containerLogInfo {
	inspect, err := d.client.ContainerInspect(ctx, container.ID)
	if err != nil {
		l.Warnf("failed to query docker %s inspect, err: %s, skip", container.Names, err)
		return nil
	}

	originalName := getContainerName(container.Names)

	labels := container.Labels
	info := &containerLogInfo{
		runtimeType:   "docker",
		id:            container.ID,
		originalName:  originalName,
		containerName: getContainerNameForLabels(labels),
		podName:       getPodNameForLabels(labels),
		podNamespace:  getPodNamespaceForLabels(labels),
		image:         container.Image,
		logPath:       inspect.LogPath,
		createdAt:     container.Created,
	}

	if info.containerName == "" {
		info.containerName = originalName
	}

	if d.k8sClient != nil && info.podName != "" {
		meta, err := queryPodMetaData(d.k8sClient, info.podName, info.podNamespace)
		if err != nil {
			l.Warnf("failed to query docker %s info from k8s, err: %s, skip", info.containerName, err)
		} else {
			img := meta.containerImage(info.containerName)
			if img != "" {
				info.image = img
			}

			info.podLabels = meta.labels()
			annotations := meta.annotations()

			// example: datakit/logs
			if v := annotations[fmt.Sprintf(logConfigAnnotationKeyFormat, "")]; v != "" {
				info.logConfigStr = v
			}

			// example: datakit/nginx.logs
			if v := annotations[fmt.Sprintf(logConfigAnnotationKeyFormat, info.containerName+".")]; v != "" {
				info.logConfigStr = v
			}
		}
	}

	if inspect.Config != nil {
		// example: DATAKIT_LOGS_CONFIG
		if v := findDockerEnv(inspect.Config.Env, "DATAKIT_LOGS_CONFIG"); v != "" {
			info.logConfigStr = v
		}
	}

	l.Debugf("docker container %s use logConfig: '%v'", info.containerName, info.logConfigStr)
	return info
}

// findDockerEnv, return the value corresponding to this key in 'envs'
//    example: "PATH=/usr/local/sbin:/usr/local/bin", return "/usr/local/sbin:/usr/local/bin"
func findDockerEnv(envs []string, key string) string {
	for _, envStr := range envs {
		array := strings.Split(envStr, "=")
		if len(array) > 1 && array[0] == key {
			return array[1]
		}
	}
	return ""
}
