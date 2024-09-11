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
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/logtail/fileprovider"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
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

		path := cfg.Path
		if cfg.hostFilePath != "" {
			path = cfg.hostFilePath
			l.Infof("container log %s redirect to host path %s", cfg.Path, cfg.hostFilePath)
		}

		mergedTags := inputs.MergeTags(c.extraTags, cfg.Tags, "")

		opts := []tailer.Option{
			tailer.WithSource(cfg.Source),
			tailer.WithService(cfg.Service),
			tailer.WithPipeline(cfg.Pipeline),
			tailer.WithCharacterEncoding(cfg.CharacterEncoding),
			tailer.WithMultilinePatterns(cfg.MultilinePatterns),
			tailer.WithGlobalTags(mergedTags),
			tailer.WithMaxMultilineLifeDuration(c.ipt.LoggingMaxMultilineLifeDuration),
			tailer.WithRemoveAnsiEscapeCodes(cfg.RemoveAnsiEscapeCodes || c.ipt.LoggingRemoveAnsiEscapeCodes),
			tailer.WithMaxForceFlushLimit(c.ipt.LoggingForceFlushLimit),
			tailer.WithFileFromBeginningThresholdSize(int64(c.ipt.LoggingFileFromBeginningThresholdSize)),
			tailer.WithDone(done),
		}

		switch cfg.Type {
		case "file":
			opts = append(opts, tailer.WithTextParserMode(tailer.FileMode))
		case runtime.DockerRuntime:
			opts = append(opts, tailer.WithTextParserMode(tailer.DockerJSONLogMode))
		default:
			opts = append(opts, tailer.WithTextParserMode(tailer.CriLogdMode))
		}

		pathAtRootfs := joinLogsAtRootfs(path)

		filelist, err := fileprovider.NewProvider().SearchFiles([]string{pathAtRootfs}).Result()
		if err != nil {
			l.Warnf("failed to scan container-log collection %s(%s) for %s, err: %s", cfg.Path, pathAtRootfs, ins.containerName, err)
			continue
		}

		if len(filelist) == 0 {
			l.Infof("container %s not found any log file for path %s, skip", ins.containerName, pathAtRootfs)
			continue
		}

		for _, file := range filelist {
			if c.logTable.inTable(ins.id, file) {
				continue
			}

			l.Infof("add container log collection with path %s from source %s", file, cfg.Source)

			newOpts := opts
			pathAtInside := trimLogsFromRootfs(file)
			if insidePath := joinInsideFilepath(cfg.hostDir, cfg.insideDir, pathAtInside); insidePath != pathAtInside {
				newOpts = append(newOpts, tailer.WithTag("inside_filepath", insidePath))
				newOpts = append(newOpts, tailer.WithTag("host_filepath", file))
			}

			tail, err := tailer.NewTailerSingle(file, newOpts...)
			if err != nil {
				l.Errorf("failed to create container-log collection %s for %s, err: %s", file, ins.containerName, err)
				continue
			}

			c.logTable.addToTable(ins.id, file, done)

			func(file string) {
				g.Go(func(ctx context.Context) error {
					defer func() {
						c.logTable.removePathFromTable(ins.id, file)
						l.Infof("remove container log collection from source %s", cfg.Source)
					}()
					tail.Run()
					return nil
				})
			}(file)
		}
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
