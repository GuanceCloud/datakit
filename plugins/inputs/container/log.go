// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package container

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/tailer"
)

const ignoreDeadLogDuration = time.Hour * 12

type containerLogBasisInfo struct {
	name, id              string
	logPath               string
	image                 string
	labels                map[string]string
	tags                  map[string]string
	created               string
	extraSourceMap        map[string]string
	sourceMultilineMap    map[string]string
	autoMultilinePatterns []string
	extractK8sLabelAsTags bool

	configKey string
}

func composeTailerOption(k8sClient k8sClientX, info *containerLogBasisInfo) (*tailer.Option, []string) {
	if info.tags == nil {
		info.tags = make(map[string]string)
	}

	info.tags["container_runtime_name"] = info.name
	if getContainerNameForLabels(info.labels) != "" {
		info.tags["container_name"] = getContainerNameForLabels(info.labels)
	} else {
		info.tags["container_name"] = info.tags["container_runtime_name"]
	}
	info.tags["container_id"] = info.id

	if n := getPodNameForLabels(info.labels); n != "" {
		info.tags["pod_name"] = n
	}
	if ns := getPodNamespaceForLabels(info.labels); ns != "" {
		info.tags["namespace"] = ns
	}

	if info.image != "" {
		imageName, imageShortName, imageTag := ParseImage(info.image)
		info.tags["image"] = info.image
		info.tags["image_name"] = imageName
		info.tags["image_short_name"] = imageShortName
		info.tags["image_tag"] = imageTag
	}

	opt := &tailer.Option{
		GlobalTags:    info.tags,
		IgnoreDeadLog: ignoreDeadLogDuration,
	}

	switch {
	case getContainerNameForLabels(info.labels) != "":
		opt.Source = getContainerNameForLabels(info.labels)
	case info.tags["image_short_name"] != "":
		opt.Source = info.tags["image_short_name"]
	default:
		opt.Source = info.name
	}

	// 如果 opt.Source 能够匹配到 extra source，就不再使用 logconf.Source 的值 (#903)
	useExtraSource := false
	for re, newSource := range info.extraSourceMap {
		match, err := regexp.MatchString(re, opt.Source)
		if err != nil {
			l.Warnf("invalid global_extra_source_map '%s', err %s, ignored", re, err)
		}
		if match {
			opt.Source = newSource
			useExtraSource = true
			break
		}
	}

	// 创建时间小于等于 5 分钟的新容器，日志从首部开始采集
	if !checkContainerIsOlder(info.created, time.Minute*10) {
		opt.FromBeginning = true
	}

	var logconf *containerLogConfig

	func() {
		if !datakit.Docker || info.tags["pod_name"] == "" || k8sClient == nil {
			return
		}
		meta, err := queryPodMetaData(k8sClient, info.tags["pod_name"], info.tags["namespace"])
		if err != nil {
			l.Warnf("failed of get pod data, err: %s", err)
			return
		}

		info.tags["pod_ip"] = meta.Status.PodIP
		if deployment := getDeployment(meta.labels()["app"], info.tags["namespace"]); deployment != "" {
			info.tags["deployment"] = deployment
		}

		// extract pod labels to tags
		if info.extractK8sLabelAsTags {
			for k, v := range meta.Labels {
				if _, ok := opt.GlobalTags[k]; !ok {
					opt.GlobalTags[k] = v
				}
			}
		}

		var conf string
		if meta.annotations() != nil && meta.annotations()[info.configKey] != "" {
			conf = meta.annotations()[info.configKey]
			l.Infof("use annotation datakit/logs, conf: %s, pod_name %s", conf, info.tags["pod_name"])
		} else {
			globalCRDLogsConfList.mu.Lock()
			crdLogsConf := globalCRDLogsConfList.list[string(meta.UID)]
			globalCRDLogsConfList.mu.Unlock()

			if crdLogsConf != "" {
				conf = crdLogsConf
				l.Infof("use crd datakit/logs, conf: %s, pod_name %s", conf, info.tags["pod_name"])
			}
		}

		if conf == "" {
			l.Debugf("not found datakit/logs conf")
			return
		}

		c, err := parseContainerLogConfig(conf)
		if err != nil {
			l.Warnf("failed of parse logconfig from annotations, err: %s", err)
			return
		}

		logconf = c
	}()

	if logconf == nil {
		c, err := getContainerLogConfig(info.configKey, info.labels)
		if err != nil {
			l.Warnf("failed of parse logconfig from labels, err: %s", err)
		} else {
			logconf = c
		}
	}

	multilineMatch := ""

	if logconf != nil {
		if !useExtraSource {
			if logconf.Source != "" {
				opt.Source = logconf.Source
			}
			opt.Pipeline = logconf.Pipeline
			multilineMatch = logconf.Multiline
		}
		if logconf.Service != "" {
			opt.Service = logconf.Service
		}

		for k, v := range logconf.Tags {
			opt.GlobalTags[k] = v
		}

		l.Debugf("use container logconfig:%#v, containerId: %s, source: %s, logpath: %s", logconf, info.id, opt.Source, info.logPath)
	}

	if multilineMatch == "" && info.sourceMultilineMap != nil {
		mult := info.sourceMultilineMap[opt.Source]
		if mult != "" {
			multilineMatch = mult
			l.Debugf("use multiline_match '%s' to source %s", multilineMatch, opt.Source)
		}
	}

	if multilineMatch != "" {
		opt.MultilinePatterns = []string{multilineMatch}
	} else if len(info.autoMultilinePatterns) != 0 {
		opt.MultilinePatterns = info.autoMultilinePatterns
		l.Infof("source %s, filename %s, automatic-multiline on, patterns %v", opt.Source, info.logPath, info.autoMultilinePatterns)
	}

	if logconf != nil && len(logconf.Paths) != 0 {
		return opt, logconf.Paths
	}

	return opt, nil
}

type containerLogConfig struct {
	Disable    bool              `json:"disable"`
	Source     string            `json:"source"`
	Paths      []string          `json:"paths"`
	Pipeline   string            `json:"pipeline"`
	Service    string            `json:"service"`
	Multiline  string            `json:"multiline_match"`
	OnlyImages []string          `json:"only_images"`
	Tags       map[string]string `json:"tags"`
}

const (
	containerLogConfigKey        = "datakit/logs"
	containerInsiderLogConfigKey = "datakit/logs/inside"
)

func getContainerLogConfig(key string, m map[string]string) (*containerLogConfig, error) {
	if m == nil || m[key] == "" {
		return nil, nil
	}
	return parseContainerLogConfig(m[key])
}

func parseContainerLogConfig(cfg string) (*containerLogConfig, error) {
	if cfg == "" {
		return nil, fmt.Errorf("logsconf is empty")
	}

	var configs []containerLogConfig
	if err := json.Unmarshal([]byte(cfg), &configs); err != nil {
		return nil, err
	}

	if len(configs) < 1 {
		return nil, fmt.Errorf("invalid logsconf")
	}

	temp := configs[0]
	return &temp, nil
}

func getPodNameForLabels(labels map[string]string) string {
	return labels["io.kubernetes.pod.name"]
}

func getPodNamespaceForLabels(labels map[string]string) string {
	return labels["io.kubernetes.pod.namespace"]
}

func getContainerNameForLabels(labels map[string]string) string {
	return labels["io.kubernetes.container.name"]
}

func logsJoinRootfs(logs string) string {
	if !datakit.Docker {
		return logs
	}
	if v := os.Getenv("HOST_ROOT"); v != "" {
		return filepath.Join(v, logs)
	}
	return filepath.Join("/rootfs", logs)
}
