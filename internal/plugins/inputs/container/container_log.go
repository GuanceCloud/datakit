// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package container

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/container/filter"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/container/runtime"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/kubernetes/podutil"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (c *containerCollector) gatherLogging() {
	list, err := c.runtime.ListContainers()
	if err != nil {
		l.Warn("not found containers, err: %s", err)
		return
	}

	var activeContainers []string

	for _, item := range list {
		if isPauseContainer(item) {
			continue
		}

		info, configStr := c.queryContainerLogInfoAndConfig(item)
		if info.ownerKind == "job" || info.ownerKind == "cronjob" {
			continue
		}

		// l.Debugf("find container %s info: %#v", item.Name, info)

		imageMatch := c.logFilter.Match(filter.FilterImage, info.image)
		nsMatch := c.logFilter.Match(filter.FilterNamespace, info.podNamespace)
		if !(imageMatch && nsMatch) {
			l.Debugf("log filter matched: containerID=%s, namespace=%s, image=%s, skip", item.ID, info.podNamespace, info.image)
			continue
		}

		c.logCoordinator.addTask(item.ID, info, configStr)
		activeContainers = append(activeContainers, item.ID)
	}

	c.logCoordinator.cleanMissingContainerLog(activeContainers)
}

const logConfigAnnotationKeyFormat = "datakit/%slogs"

func (c *containerCollector) queryContainerLogInfoAndConfig(item *runtime.Container) (info *containerLogInfo, configStr string) {
	podName := getPodNameForLabels(item.Labels)
	namespace := getPodNamespaceForLabels(item.Labels)

	info = &containerLogInfo{
		containerID:   item.ID,
		containerName: item.Name,
		runtime:       item.RuntimeName,
		image:         item.Image,
		logPath:       item.LogPath,
		podName:       podName,
		podNamespace:  namespace,
		mergedDir:     item.MergedDir,
		mounts:        item.Mounts,
	}
	if name := getContainerNameForLabels(item.Labels); name != "" {
		info.containerName = name
	}

	if c.k8sClient != nil && podName != "" {
		pod, err := c.k8sClient.GetPods(namespace).Get(context.Background(), podName, metav1.GetOptions{ResourceVersion: "0"})
		if err != nil {
			l.Warnf("query pod fail, err: %s", err)
		} else {
			// ex: datakit/logs
			if v := pod.Annotations[fmt.Sprintf(logConfigAnnotationKeyFormat, "")]; v != "" {
				configStr = v
			}
			// ex: datakit/nginx.logs
			if v := pod.Annotations[fmt.Sprintf(logConfigAnnotationKeyFormat, info.containerName+".")]; v != "" {
				configStr = v
			}

			info.podIP = pod.Status.PodIP
			info.podLabels = pod.Labels
			info.ownerKind, info.ownerName = podutil.PodOwner(pod)
			if img := podutil.ContainerImageFromPod(info.containerName, pod); img != "" {
				info.image = img
			}
		}
	}
	// ex: DATAKIT_LOGS_CONFIG
	if item.Envs != nil {
		if str, ok := item.Envs["DATAKIT_LOGS_CONFIG"]; ok && str != "" {
			configStr = str
		}
	}
	return info, configStr
}

type loggingDefaults struct {
	enableDebugFields bool
	enableMultiline   bool
	extraSourceMap    map[string]string

	sourceMultilineMap         map[string]string
	autoMultilineDetection     bool
	autoMultilineExtraPatterns []string
	maxMultilineLength         int64

	fileFromBeginning     bool
	fileSizeThreshold     int64
	removeAnsiEscapeCodes bool
	fieldWhitelist        []string
	maxOpenFiles          int
	ignoreDeadLog         time.Duration

	insideFilepathFunc func(hostDir, insideDir, filepath string) string
	extraTags          map[string]string
	setLabelAsTags     func(map[string]string) map[string]string
}

func newLoggingDefaults(ipt *Input) *loggingDefaults {
	optForNonMetric := buildLabelsOption(ipt.ExtractK8sLabelAsTagsV2, config.Cfg.Dataway.GlobalCustomerKeys)
	l.Infof("use labels %s for logging", optForNonMetric.keys)

	return &loggingDefaults{
		enableDebugFields: config.Cfg.EnableDebugFields,
		enableMultiline:   ipt.LoggingEnableMultline,
		extraSourceMap:    ipt.LoggingExtraSourceMap,

		sourceMultilineMap:         ipt.LoggingSourceMultilineMap,
		autoMultilineDetection:     ipt.LoggingAutoMultilineDetection,
		autoMultilineExtraPatterns: ipt.LoggingAutoMultilineExtraPatterns,
		maxMultilineLength:         int64(float64(config.Cfg.Dataway.MaxRawBodySize) * 0.8),

		fileFromBeginning:     ipt.LoggingFileFromBeginning,
		fileSizeThreshold:     int64(ipt.LoggingFileFromBeginningThresholdSize),
		removeAnsiEscapeCodes: ipt.LoggingRemoveAnsiEscapeCodes,
		fieldWhitelist:        ipt.LoggingFieldWhiteList,
		maxOpenFiles:          ipt.LoggingMaxOpenFiles,
		ignoreDeadLog:         time.Minute * 10,

		insideFilepathFunc: func(hostDir, insideDir, filepath string) string {
			pathAtInside := trimLogsFromRootfs(filepath)
			insidePath, ok := runtime.ReverseResolveFromSourcePath(pathAtInside, insideDir, hostDir)
			if ok && insidePath != pathAtInside {
				return insidePath
			}
			return ""
		},
		extraTags: inputs.MergeTags(ipt.Tagger.HostTags(), ipt.Tags, ""),
		setLabelAsTags: func(labels map[string]string) map[string]string {
			return setLabelAsTags(ipt.EnableExtractK8sLabelAsTags, optForNonMetric, labels)
		},
	}
}

func setLabelAsTags(enableExtractK8sLabelAsTagsV1 bool, podLabelAsTagsForNonMetric labelsOption, labels map[string]string) map[string]string {
	if len(labels) == 0 {
		return nil
	}

	tags := make(map[string]string)
	if enableExtractK8sLabelAsTagsV1 || podLabelAsTagsForNonMetric.all {
		for k, v := range labels {
			tags[k] = v
		}
	} else {
		for _, key := range podLabelAsTagsForNonMetric.keys {
			v, ok := labels[key]
			if !ok {
				continue
			}
			tags[key] = v
		}
	}
	return tags
}

const defaultContainerLogMountPoint = "/rootfs"

func joinLogsAtRootfs(logs string) string {
	if !datakit.Docker && !config.IsKubernetes() {
		return logs
	}
	if v := os.Getenv("HOST_ROOT"); v != "" {
		return filepath.Join(v, logs)
	}
	return filepath.Join(defaultContainerLogMountPoint, logs)
}

func trimLogsFromRootfs(logs string) string {
	if v := os.Getenv("HOST_ROOT"); v != "" {
		return strings.TrimPrefix(logs, v)
	}
	return strings.TrimPrefix(logs, defaultContainerLogMountPoint)
}

func joinHostFilepath(hostDir, insideDir, insidePath string) string {
	if hostDir == "" || insideDir == "" {
		return insidePath
	}
	partialPath := strings.TrimPrefix(insidePath, insideDir)
	return filepath.Join(hostDir, filepath.Clean(partialPath))
}

func joinInsideFilepath(hostDir, insideDir, hostPath string) string {
	if hostDir == "" || insideDir == "" {
		return hostPath
	}
	partialPath := strings.TrimPrefix(hostPath, hostDir)
	return filepath.Join(insideDir, filepath.Clean(partialPath))
}

func replaceLabelKey(s string) string {
	return strings.ReplaceAll(s, ".", "_")
}
