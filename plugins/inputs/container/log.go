// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package container

import (
	"encoding/json"
	"regexp"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/tailer"
)

type containerLogBasisInfo struct {
	name, id       string
	logPath        string
	image          string
	labels         map[string]string
	tags           map[string]string
	created        string
	extraSourceMap map[string]string
}

func composeTailerOption(k8sClient k8sClientX, info *containerLogBasisInfo) *tailer.Option {
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
	info.tags["pod_name"] = getPodNameForLabels(info.labels)
	info.tags["namespace"] = getPodNamespaceForLabels(info.labels)

	if info.image != "" {
		imageName, imageShortName, imageTag := ParseImage(info.image)
		info.tags["image"] = info.image
		info.tags["image_name"] = imageName
		info.tags["image_short_name"] = imageShortName
		info.tags["image_tag"] = imageTag
	}

	opt := &tailer.Option{GlobalTags: info.tags}

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

	if !checkContainerIsOlder(info.created, time.Minute) {
		opt.FromBeginning = true
	}

	var logconf *containerLogConfig

	func() {
		if !datakit.Docker || info.tags["pod_name"] == "" {
			return
		}
		meta, err := queryPodMetaData(k8sClient, info.tags["pod_name"], info.tags["namespace"])
		if err != nil {
			l.Warnf("failed of get pod data, err: %s", err)
			return
		}

		info.tags["pod_ip"] = meta.Status.PodIP
		// if replicaSet := meta.replicaSet(); replicaSet != "" {
		// 	info.tags["replicaSet"] = replicaSet
		// }
		if deployment := getDeployment(meta.labels()["app"], info.tags["namespace"]); deployment != "" {
			info.tags["deployment"] = deployment
		}

		c, err := getContainerLogConfig(meta.annotations())
		if err != nil {
			l.Warnf("failed of parse logconfig from annotations, err: %s", err)
			return
		}

		logconf = c
	}()

	if logconf == nil {
		c, err := getContainerLogConfig(info.labels)
		if err != nil {
			l.Warnf("failed of parse logconfig from labels, err: %s", err)
		} else {
			logconf = c
		}
	}

	if logconf != nil {
		if !useExtraSource && logconf.Source != "" {
			opt.Source = logconf.Source
		}
		if logconf.Service != "" {
			opt.Service = logconf.Service
		}
		opt.Pipeline = logconf.Pipeline
		opt.MultilineMatch = logconf.Multiline

		for k, v := range logconf.Tags {
			opt.GlobalTags[k] = v
		}

		l.Debugf("use container logconfig:%#v, containerId: %s, source: %s, logpath: %s", logconf, info.id, opt.Source, info.logPath)
	}

	_ = opt.Init()

	l.Debugf("use container-log opt:%#v, containerId: %s", logconf, opt)

	return opt
}

type containerLogConfig struct {
	Disable    bool              `json:"disable"`
	Source     string            `json:"source"`
	Pipeline   string            `json:"pipeline"`
	Service    string            `json:"service"`
	Multiline  string            `json:"multiline_match"`
	OnlyImages []string          `json:"only_images"`
	Tags       map[string]string `json:"tags"`
}

const containerLogConfigKey = "datakit/logs"

func getContainerLogConfig(m map[string]string) (*containerLogConfig, error) {
	if m == nil || m[containerLogConfigKey] == "" {
		return nil, nil
	}
	return parseContainerLogConfig(m[containerLogConfigKey])
}

func parseContainerLogConfig(cfg string) (*containerLogConfig, error) {
	var configs []containerLogConfig
	if err := json.Unmarshal([]byte(cfg), &configs); err != nil {
		return nil, err
	}

	if len(configs) < 1 {
		return nil, nil
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
