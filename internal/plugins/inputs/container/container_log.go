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

func (c *container) tailingLogs(instance *logInstance) {
	g := goroutine.NewGroup(goroutine.Option{Name: "container-logs/" + instance.containerName})
	done := make(chan interface{})

	for _, cfg := range instance.configs {
		if cfg.Disable {
			continue
		}

		if c.logTable.inTable(instance.id, cfg.Path) {
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
			l.Errorf("failed to create container-log collection %s for %s, err: %s", path, instance.containerName, err)
			continue
		}

		c.logTable.addToTable(instance.id, cfg.Path, done)

		g.Go(func(ctx context.Context) error {
			defer func() {
				c.logTable.removePathFromTable(instance.id, cfg.Path)
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

	instance := &logInstance{
		id:           info.ID,
		logPath:      info.LogPath,
		podName:      podName,
		podNamespace: podNamespace,
	}

	containerName := getContainerNameForLabels(info.Labels)
	if containerName != "" {
		instance.containerName = info.Name
	} else {
		instance.containerName = containerName
	}

	if c.k8sClient != nil && podName != "" {
		owner, err := c.queryOwnerFromK8s(context.Background(), podName, podNamespace)
		if err != nil {
			l.Warn(err)
		} else {
			// ex: datakit/logs
			if v := owner.podAnnotations[fmt.Sprintf(logConfigAnnotationKeyFormat, "")]; v != "" {
				instance.configStr = v
			}

			// ex: datakit/nginx.logs
			if v := owner.podAnnotations[fmt.Sprintf(logConfigAnnotationKeyFormat, instance.containerName+".")]; v != "" {
				instance.configStr = v
			}

			instance.ownerKind = owner.ownerKind
			instance.ownerName = owner.ownerName
		}
	}

	// ex: DATAKIT_LOGS_CONFIG
	if info.Envs != nil {
		if str, ok := info.Envs["DATAKIT_LOGS_CONFIG"]; ok {
			instance.configStr = str
		}
	}

	l.Debugf("container %s use config: %v", instance.containerName, instance)
	return instance
}
