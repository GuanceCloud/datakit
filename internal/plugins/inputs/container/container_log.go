// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package container

import (
	"context"
	"fmt"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/container/runtime"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/goroutine"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/tailer"
)

const defaultActiveDuration = time.Hour * 1

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

	for _, cfg := range ins.configs {
		if cfg.Disable {
			continue
		}
		if c.logTable.inTable(ins.id, cfg.Path) {
			continue
		}

		path := cfg.Path
		if cfg.hostFilePath != "" {
			path = cfg.hostFilePath
			l.Infof("container log %s redirect to host path %s", cfg.Path, cfg.hostFilePath)
		}

		mergedTags := inputs.MergeTags(c.extraTags, cfg.Tags, "")
		pathAtRootfs := joinLogsAtRootfs(path)

		hostDir := cfg.hostDir
		insideDir := cfg.insideDir

		insideFilepathFunc := func(filepath string) string {
			pathAtInside := trimLogsFromRootfs(filepath)
			if insidePath := joinInsideFilepath(hostDir, insideDir, pathAtInside); insidePath != pathAtInside {
				return insidePath
			}
			return ""
		}

		opts := []tailer.Option{
			tailer.WithSource(cfg.Source),
			tailer.WithService(cfg.Service),
			tailer.WithPipeline(cfg.Pipeline),
			tailer.EnableDebugFields(config.Cfg.EnableDebugFields),
			tailer.WithCharacterEncoding(cfg.CharacterEncoding),
			tailer.EnableMultiline(c.ipt.LoggingEnableMultline),
			tailer.WithMultilinePatterns(cfg.MultilinePatterns),
			tailer.WithGlobalTags(mergedTags),
			tailer.WithMaxMultilineLength(int64(float64(config.Cfg.Dataway.MaxRawBodySize) * 0.8)),
			tailer.WithMaxMultilineLifeDuration(c.ipt.LoggingMaxMultilineLifeDuration),
			tailer.WithRemoveAnsiEscapeCodes(cfg.RemoveAnsiEscapeCodes || c.ipt.LoggingRemoveAnsiEscapeCodes),
			tailer.WithMaxForceFlushLimit(c.ipt.LoggingForceFlushLimit),
			tailer.WithFromBeginning(cfg.FromBeginning || c.ipt.LoggingFileFromBeginning),
			tailer.WithFileFromBeginningThresholdSize(int64(c.ipt.LoggingFileFromBeginningThresholdSize)),
			tailer.WithIgnoreDeadLog(defaultActiveDuration),
			tailer.WithInsideFilepathFunc(insideFilepathFunc),
		}

		switch cfg.Type {
		case "file":
			opts = append(opts, tailer.WithTextParserMode(tailer.FileMode))
		case runtime.DockerRuntime:
			opts = append(opts, tailer.WithTextParserMode(tailer.DockerJSONLogMode))
		default:
			opts = append(opts, tailer.WithTextParserMode(tailer.CriLogdMode))
		}

		tail, err := tailer.NewTailer([]string{pathAtRootfs}, opts...)
		if err != nil {
			l.Warnf("failed to create container-tailer %s for %s, err: %s", pathAtRootfs, ins.containerName, err)
			continue
		}

		l.Infof("add container log collection with path %s from source %s", pathAtRootfs, cfg.Source)

		c.logTable.addToTable(ins.id, cfg.Path, tail.Close)
		g.Go(func(ctx context.Context) error {
			defer func() {
				c.logTable.removePathFromTable(ins.id, cfg.Path)
				l.Infof("remove container log collection from source %s", cfg.Source)
			}()
			tail.Start()
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

			ins.podIP = podInfo.pod.Status.PodIP
			ins.podLabels = podInfo.pod.Labels
			ins.ownerKind, ins.ownerName = podInfo.owner()

			// use Image from Pod Container
			if img := podInfo.containerImage(ins.containerName); img != "" {
				ins.image = img
			}
		}
	}

	// ex: DATAKIT_LOGS_CONFIG
	if info.Envs != nil {
		if str, ok := info.Envs["DATAKIT_LOGS_CONFIG"]; ok {
			ins.configStr = str
		}
	}

	ins.imageName, ins.imageShortName, ins.imageTag = runtime.ParseImage(ins.image)

	l.Debugf("container %s use config: %v", ins.containerName, ins)
	return ins
}
