// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package container

import (
	"context"
	"fmt"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/container/runtime"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/goroutine"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/tailer"
)

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

		opt := &tailer.Option{
			Source:                   cfg.Source,
			Service:                  cfg.Service,
			Pipeline:                 cfg.Pipeline,
			CharacterEncoding:        cfg.CharacterEncoding,
			MultilinePatterns:        cfg.MultilinePatterns,
			GlobalTags:               cfg.Tags,
			BlockingMode:             c.ipt.LoggingBlockingMode,
			MinFlushInterval:         c.ipt.LoggingMinFlushInterval,
			MaxMultilineLifeDuration: c.ipt.LoggingMaxMultilineLifeDuration,
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

		path := logsJoinRootfs(cfg.Path)

		l.Infof("add container log collection with path: %s from source %s", path, opt.Source)

		tail, err := tailer.NewTailerSingle(path, opt)
		if err != nil {
			l.Errorf("failed to create container-log collection %s for %s, err: %s", path, ins.containerName, err)
			continue
		}

		c.logTable.addToTable(ins.id, cfg.Path, done)

		g.Go(func(ctx context.Context) error {
			defer func() {
				c.logTable.removePathFromTable(ins.id, cfg.Path)
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

			ins.ownerKind, ins.ownerName = podInfo.owner()

			// use Image from Pod Container
			ins.image = podInfo.containerImage(ins.containerName)
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
