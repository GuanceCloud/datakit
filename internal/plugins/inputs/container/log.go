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

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
)

const logConfigAnnotationKeyFormat = "datakit/%slogs"

type logConfig struct {
	Disable           bool              `json:"disable"`
	Type              string            `json:"type"`
	Path              string            `json:"path"`
	Source            string            `json:"source"`
	Service           string            `json:"service"`
	CharacterEncoding string            `json:"character_encoding"`
	Pipeline          string            `json:"pipeline"`
	Multiline         string            `json:"multiline_match"`
	MultilinePatterns []string          `json:"-"`
	Tags              map[string]string `json:"tags"`
}

type logConfigs []*logConfig

func (lc logConfigs) enabled() bool {
	b := false
	for _, c := range lc {
		b = b || !c.Disable
	}
	return b
}

func parseLogConfig(cfg string) (logConfigs, error) {
	if cfg == "" {
		return nil, fmt.Errorf("logsconf is empty")
	}

	var configs logConfigs
	if err := json.Unmarshal([]byte(cfg), &configs); err != nil {
		return nil, err
	}

	return configs, nil
}

type containerLogInfo struct {
	runtimeType   string
	id            string
	originalName  string
	containerName string
	image         string
	podName       string
	podNamespace  string
	logPath       string

	tags      map[string]string
	podLabels map[string]string

	logConfigStr string
	logConfigs   logConfigs
}

func (info *containerLogInfo) enabled() bool {
	return info.logConfigs.enabled()
}

func (info *containerLogInfo) fillTags() {
	info.tags = map[string]string{
		"container_id":           info.id,
		"container_runtime_name": info.originalName,
		"container_name":         info.containerName,
	}

	if info.podName != "" {
		info.tags["pod_name"] = info.podName
	}
	if info.podNamespace != "" {
		info.tags["namespace"] = info.podNamespace
	}

	if info.image != "" {
		imageName, imageShortName, imageTag := ParseImage(info.image)
		info.tags["image"] = info.image
		info.tags["image_name"] = imageName
		info.tags["image_short_name"] = imageShortName
		info.tags["image_tag"] = imageTag
	}

	if containerIsFromKubernetes(info.containerName) {
		info.tags["container_type"] = "kubernetes"
	} else {
		info.tags["container_type"] = info.runtimeType
	}
}

func (info *containerLogInfo) parseLogConfigs() error {
	if info.logConfigStr != "" {
		configs, err := parseLogConfig(info.logConfigStr)
		if err != nil {
			return fmt.Errorf("failed to parse configs from container %s, err: %w", info.containerName, err)
		}
		info.logConfigs = configs
	}
	return nil
}

func (info *containerLogInfo) addStdout() {
	if len(info.logConfigs) == 0 {
		info.logConfigs = append(info.logConfigs, &logConfig{
			Type:   "stdout/stderr",
			Path:   info.logPath,
			Source: info.containerName,
		})
		return
	}
	for _, cfg := range info.logConfigs {
		if (cfg.Type == "" || cfg.Type == "stdout") && cfg.Path == "" {
			cfg.Path = info.logPath
		}
	}
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
