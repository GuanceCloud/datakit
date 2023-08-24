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

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/container/runtime"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
)

const logConfigAnnotationKeyFormat = "datakit/%slogs"

type logConfig struct {
	Disable           bool              `json:"disable"`
	Type              string            `json:"type"`
	Path              string            `json:"path"`
	TargetPath        string            `json:"-"`
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
	image         string
	logPath       string
	configStr     string
	configs       logConfigs

	podName      string
	podNamespace string
	ownerKind    string
	ownerName    string

	// volMounts Source to HostTarget
	// example: map["/tmp/opt"] = "/var/lib/docker/volumes/<id>/_data"
	//          map["/tmp/opt"] = "/var/lib/kubelet/pods/<pod-id>/volumes/kubernetes.io~empty-dir/<volume-name>/"
	volMounts map[string]string
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

		for _, cfg := range lc.configs {
			base := filepath.Base(cfg.Path)
			dir := filepath.Dir(cfg.Path)

			target, ok := lc.volMounts[filepath.Clean(dir)]
			if !ok {
				continue
			}
			cfg.TargetPath = filepath.Join(filepath.Clean(target), base)

			// add target fileapath
			if cfg.Tags == nil {
				cfg.Tags = make(map[string]string)
			}
			cfg.Tags["filepath"] = cfg.Path
			cfg.Tags["target_filepath"] = cfg.TargetPath
		}
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
	imageName, shortName, imageTag := runtime.ParseImage(lc.image)
	m := map[string]string{
		"container_id":     lc.id,
		"container_name":   lc.containerName,
		"image":            lc.image,
		"image_name":       imageName,
		"image_short_name": shortName,
		"image_tag":        imageTag,
	}

	if lc.podName != "" {
		m["pod_name"] = lc.podName
		m["namespace"] = lc.podNamespace
	}

	if lc.ownerKind != "" && lc.ownerName != "" {
		m[lc.ownerKind] = lc.ownerName
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
