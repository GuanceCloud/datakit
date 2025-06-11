// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package container

import (
	"context"
	"fmt"
	"regexp"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/container/filter"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/container/runtime"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/goroutine"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/kubernetes/podutil"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/logtail/multiline"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/tailer"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const defaultActiveDuration = time.Hour * 1

func (c *containerCollector) gatherLogging() {
	list, err := c.runtime.ListContainers()
	if len(list) == 0 && err != nil {
		l.Warn("not found containers, err: %s", err)
		return
	}

	var activeIDs []string
	for _, item := range list {
		activeIDs = append(activeIDs, item.ID)
	}
	c.cleanMissingContainerLog(activeIDs)

	for _, item := range list {
		if isPauseContainer(item) {
			continue
		}

		l.Debugf("find container %s info: %#v", item.Name, item)

		instance := c.queryContainerLogInfo(item)
		if instance == nil {
			continue
		}

		if err := instance.parseLogConfigs(); err != nil {
			l.Warn(err)
			continue
		}

		if !c.shouldPullContainerLog(instance) {
			continue
		}

		instance.addStdout()
		instance.fillLogType(item.RuntimeName)
		instance.fillSource()
		instance.checkTagsKey()

		instance.setTagsToLogConfigs(instance.tags())
		instance.setTagsToLogConfigs(c.extraTags)
		if c.enableExtractK8sLabelAsTagsV1 {
			instance.setLabelAsTags(instance.podLabels, true /*all labels*/, nil)
		} else {
			instance.setLabelAsTags(instance.podLabels, c.podLabelAsTagsForNonMetric.all, c.podLabelAsTagsForNonMetric.keys)
		}

		setLoggingExtraSourceMapToLogConfigs(c.ipt, instance.configs)
		setLoggingSourceMultilineMapToLogConfigs(c.ipt, instance.configs)
		setLoggingAutoMultilineToLogConfigs(c.ipt, instance.configs)

		c.tailingLogs(instance)
	}

	l.Debugf("current container logtable: %s", c.logTable.String())
}

func (c *containerCollector) shouldPullContainerLog(ins *logInstance) bool {
	if len(ins.configs) != 0 {
		disable := true
		for _, cfg := range ins.configs {
			disable = disable && cfg.Disable
		}
		return !disable
	}

	if ins.ownerKind == "job" || ins.ownerKind == "cronjob" {
		return false
	}

	pass := c.logFilter.Match(filter.FilterImage, ins.image) &&
		c.logFilter.Match(filter.FilterImageName, ins.imageName) &&
		c.logFilter.Match(filter.FilterImageShortName, ins.imageShortName) &&
		c.logFilter.Match(filter.FilterNamespace, ins.podNamespace)

	return pass
}

func (c *containerCollector) cleanMissingContainerLog(activeIDs []string) {
	missingIDs := c.logTable.findDifferences(activeIDs)
	for _, id := range missingIDs {
		l.Infof("clean log collection for container id %s", id)
		c.logTable.closeFromTable(id)
		c.logTable.removeFromTable(id)
	}
}

func (c *containerCollector) tailingLogs(ins *logInstance) {
	if config.IsKVTemplate(ins.configTemplate) {
		if err := config.GetKV().Register("container-logs", ins.configTemplate, c.ReloadConfigKV,
			&config.KVOpt{
				IsMultiConf:              true,
				IsUnRegisterBeforeReload: true,
				ConfName:                 ins.containerName + "-" + ins.id,
			},
		); err != nil {
			l.Warnf("register KV err: %s", err)
		}
	}

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
			tailer.WithRemoveAnsiEscapeCodes(cfg.RemoveAnsiEscapeCodes || c.ipt.LoggingRemoveAnsiEscapeCodes),
			tailer.WithMaxOpenFiles(c.ipt.LoggingMaxOpenFiles),
			tailer.WithFromBeginning(cfg.FromBeginning || c.ipt.LoggingFileFromBeginning),
			tailer.WithFileFromBeginningThresholdSize(int64(c.ipt.LoggingFileFromBeginningThresholdSize)),
			tailer.WithIgnoreDeadLog(defaultActiveDuration),
			tailer.WithFieldWhiteList(c.ipt.LoggingFieldWhiteList),
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

		cfgPath := cfg.Path
		cfgSource := cfg.Source

		g.Go(func(ctx context.Context) error {
			defer func() {
				c.logTable.removePathFromTable(ins.id, cfgPath)
				l.Infof("remove container log collection from source %s", cfgSource)
			}()
			tail.Start()
			return nil
		})
	}
}

func (c *containerCollector) queryContainerLogInfo(item *runtime.Container) *logInstance {
	podName := getPodNameForLabels(item.Labels)
	namespace := getPodNamespaceForLabels(item.Labels)

	ins := &logInstance{
		id:            item.ID,
		containerName: item.Name,
		image:         item.Image,
		logPath:       item.LogPath,
		podName:       podName,
		podNamespace:  namespace,
		volMounts:     item.Mounts,
	}

	if name := getContainerNameForLabels(item.Labels); name != "" {
		ins.containerName = name
	}

	if c.k8sClient != nil && podName != "" {
		pod, err := c.k8sClient.GetPods(namespace).Get(context.Background(), podName, metav1.GetOptions{ResourceVersion: "0"})
		if err != nil {
			l.Warnf("query pod failed, err: %s", err)
		} else {
			// ex: datakit/logs
			if v := pod.Annotations[fmt.Sprintf(logConfigAnnotationKeyFormat, "")]; v != "" {
				ins.configTemplate = v
			}
			// ex: datakit/nginx.logs
			if v := pod.Annotations[fmt.Sprintf(logConfigAnnotationKeyFormat, ins.containerName+".")]; v != "" {
				ins.configTemplate = v
			}

			ins.podIP = pod.Status.PodIP
			ins.podLabels = pod.Labels
			ins.ownerKind, ins.ownerName = podutil.PodOwner(pod)

			if img := podutil.ContainerImageFromPod(ins.containerName, pod); img != "" {
				ins.image = img
			}
		}
	}

	// ex: DATAKIT_LOGS_CONFIG
	if item.Envs != nil {
		if str, ok := item.Envs["DATAKIT_LOGS_CONFIG"]; ok {
			ins.configTemplate = str
		}
	}

	var err error
	ins.configStr, err = config.GetKV().ReplaceKV(ins.configTemplate)
	if err != nil {
		ins.configStr = ins.configTemplate
		l.Warnf("failed of replace kv for container %s, err: %s", ins.containerName, ins.configTemplate)
	} else {
		l.Debugf("replace kv for container %s, old: %s, new:%s", ins.containerName, ins.configTemplate, ins.configStr)
	}

	ins.imageName, ins.imageShortName, ins.imageTag = runtime.ParseImage(ins.image)
	l.Debugf("container %s use config: %v", ins.containerName, ins)

	return ins
}

func setLoggingAutoMultilineToLogConfigs(ipt *Input, configs logConfigs) {
	if !ipt.LoggingAutoMultilineDetection {
		return
	}
	for _, cfg := range configs {
		if len(cfg.MultilinePatterns) != 0 {
			continue
		}
		cfg.MultilinePatterns = ipt.LoggingAutoMultilineExtraPatterns
		cfg.MultilinePatterns = append(cfg.MultilinePatterns, multiline.GlobalPatterns...)
	}
}

func setLoggingExtraSourceMapToLogConfigs(ipt *Input, configs logConfigs) {
	for re, newSource := range ipt.LoggingExtraSourceMap {
		for _, cfg := range configs {
			match, err := regexp.MatchString(re, cfg.Source)
			if err != nil {
				l.Warnf("invalid global_extra_source_map '%s', err %s, skip", re, err)
			}
			if match {
				l.Infof("replaced source '%s' with '%s'", cfg.Source, newSource)
				cfg.Source = newSource
				break
			}
		}
	}
}

func setLoggingSourceMultilineMapToLogConfigs(ipt *Input, configs logConfigs) {
	if len(ipt.LoggingSourceMultilineMap) == 0 {
		return
	}
	for _, cfg := range configs {
		if cfg.Multiline != "" {
			continue
		}

		source := cfg.Source
		mult := ipt.LoggingSourceMultilineMap[source]
		if mult != "" {
			l.Infof("replaced multiline '%s' with '%s' to source %s", cfg.Multiline, mult, source)
			cfg.MultilinePatterns = []string{mult}
		}
	}
}
