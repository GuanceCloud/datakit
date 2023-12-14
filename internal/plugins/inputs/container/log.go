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
	"strings"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
)

const logConfigAnnotationKeyFormat = "datakit/%slogs"

type logConfig struct {
	Disable           bool              `json:"disable"`
	Type              string            `json:"type"`
	Path              string            `json:"path"`
	HostFilePath      string            `json:"-"`
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
	id                                  string
	containerName                       string
	image                               string
	imageName, imageShortName, imageTag string
	logPath                             string
	configStr                           string
	configs                             logConfigs

	podName      string
	podNamespace string
	podLabels    map[string]string
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
			if cfg.Multiline != "" {
				cfg.MultilinePatterns = []string{cfg.Multiline}
			}

			if cfg.Path == "" {
				continue
			}

			path := filepath.Clean(cfg.Path)
			foundHostPath := false

			for vol, hostdir := range lc.volMounts {
				if strings.HasPrefix(path, vol) {
					file := strings.TrimPrefix(path, vol)
					cfg.HostFilePath = filepath.Join(hostdir, filepath.Clean(file))

					// add target fileapath
					if cfg.Tags == nil {
						cfg.Tags = make(map[string]string)
					}
					cfg.Tags["inside_filepath"] = path
					cfg.Tags["host_filepath"] = cfg.HostFilePath

					foundHostPath = true
				}
			}

			if !foundHostPath {
				return fmt.Errorf("unexpected log path %s, no matched mounts found", cfg.Path)
			}
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

func (lc *logInstance) fillSource() {
	for _, cfg := range lc.configs {
		if cfg.Source != "" {
			continue
		}
		cfg.Source = lc.containerName
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

func (lc *logInstance) setLabelsToLogConfigs(m map[string]string) {
	for _, cfg := range lc.configs {
		if cfg.Tags == nil {
			cfg.Tags = make(map[string]string)
		}
		for k, v := range m {
			key := replaceLabelKey(k)
			if _, ok := cfg.Tags[key]; !ok {
				cfg.Tags[key] = v
			}
		}
	}
}

func replaceLabelKey(s string) string {
	return strings.ReplaceAll(s, ".", "_")
}

func (lc *logInstance) setCustomerTags(m map[string]string, keys []string) {
	if len(keys) == 0 || len(m) == 0 {
		return
	}

	for _, cfg := range lc.configs {
		if cfg.Tags == nil {
			cfg.Tags = make(map[string]string)
		}
		for _, key := range keys {
			if v, ok := m[key]; ok {
				cfg.Tags[key] = v
			}
		}
	}
}

func (lc *logInstance) tags() map[string]string {
	m := map[string]string{
		"container_id":     lc.id,
		"container_name":   lc.containerName,
		"image":            lc.image,
		"image_name":       lc.imageName,
		"image_short_name": lc.imageShortName,
		"image_tag":        lc.imageTag,
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
