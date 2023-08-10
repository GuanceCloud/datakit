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

type logInstance struct {
	id            string
	containerName string
	logPath       string
	configStr     string
	configs       logConfigs

	podName      string
	podNamespace string
	ownerKind    ownerKind
	ownerName    string
}

func (lc *logInstance) enabled() bool {
	b := false
	for _, cfg := range lc.configs {
		b = b || !cfg.Disable
	}
	return b
}

func (lc *logInstance) parseLogConfigs() error {
	if lc.configStr != "" {
		var configs logConfigs
		if err := json.Unmarshal([]byte(lc.configStr), &configs); err != nil {
			return fmt.Errorf("failed to parse configs from container %s, err: %w, data: %s",
				lc.containerName, err, lc.configStr)
		}
		lc.configs = configs
	}
	return nil
}

func (lc *logInstance) addStdout() {
	if len(lc.configs) == 0 {
		lc.configs = append(lc.configs, &logConfig{
			Path:   lc.logPath,
			Source: lc.containerName,
		})
		return
	}

	for _, cfg := range lc.configs {
		if cfg.Type == "" && cfg.Path == "" {
			cfg.Path = lc.logPath
		}
	}
}

func (lc *logInstance) fillLogType(runtimeName string) {
	for _, cfg := range lc.configs {
		if cfg.Type != "" {
			continue
		}
		cfg.Type = runtimeName
	}
}

func (lc *logInstance) setTagsToLogConfigs(m map[string]string) {
	for _, cfg := range lc.configs {
		if cfg.Tags == nil {
			cfg.Tags = make(map[string]string)
		}
		for k, v := range m {
			if _, ok := cfg.Tags[k]; !ok {
				cfg.Tags[k] = v
			}
		}
	}
}

func (lc *logInstance) tags() map[string]string {
	m := map[string]string{
		"container_id":   lc.id,
		"container_name": lc.containerName,
	}

	if lc.podName != "" {
		m["pod_name"] = lc.podName
		m["namespace"] = lc.podNamespace
	}

	switch lc.ownerKind {
	case deploymentKind:
		m["deployment"] = lc.ownerName
	case daemonsetKind:
		m["daemonset"] = lc.ownerName
	case statefulsetKind:
		m["statefulset"] = lc.ownerName
	default:
		// skip
	}

	return m
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
