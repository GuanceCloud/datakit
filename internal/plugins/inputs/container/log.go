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

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
)

const logConfigAnnotationKeyFormat = "datakit/%slogs"

type logConfig struct {
	Disable               bool              `json:"disable"`
	Type                  string            `json:"type"`
	Path                  string            `json:"path"`
	Source                string            `json:"source"`
	Service               string            `json:"service"`
	CharacterEncoding     string            `json:"character_encoding"`
	Pipeline              string            `json:"pipeline"`
	Multiline             string            `json:"multiline_match"`
	RemoveAnsiEscapeCodes bool              `json:"remove_ansi_escape_codes"`
	Tags                  map[string]string `json:"tags"`

	MultilinePatterns []string `json:"-"`

	hostDir      string `json:"-"`
	insideDir    string `json:"-"`
	hostFilePath string `json:"-"`
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

	podName, podIP, podNamespace string
	podLabels                    map[string]string
	ownerKind                    string
	ownerName                    string

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
					cfg.hostDir = hostdir
					cfg.insideDir = vol
					cfg.hostFilePath = joinHostFilepath(hostdir, vol, path)

					foundHostPath = true
				}
			}

			if !foundHostPath {
				return fmt.Errorf("unexpected log path %s, no matched mounts(%d) found", cfg.Path, len(lc.volMounts))
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
		if (cfg.Type == "" || cfg.Type == "stdout") && cfg.Path == "" {
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

func (lc *logInstance) checkTagsKey() {
	for _, cfg := range lc.configs {
		for k, v := range cfg.Tags {
			if idx := strings.Index(k, "."); idx == -1 {
				continue
			}
			newkey := replaceLabelKey(k)
			if _, ok := cfg.Tags[newkey]; !ok {
				cfg.Tags[newkey] = v
				delete(cfg.Tags, k)
			}
		}
	}
}

func (lc *logInstance) setTagsToLogConfigs(m map[string]string) {
	if len(m) == 0 {
		return
	}
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

func (lc *logInstance) setLabelAsTags(m map[string]string, all bool, keys []string) {
	if len(m) == 0 {
		return
	}

	if all {
		for _, cfg := range lc.configs {
			if cfg.Tags == nil {
				cfg.Tags = make(map[string]string)
			}
			for k, v := range m {
				if _, ok := cfg.Tags[k]; !ok {
					newkey := replaceLabelKey(k)
					cfg.Tags[newkey] = v
				}
			}
		}
		return
	}

	for _, cfg := range lc.configs {
		if cfg.Tags == nil {
			cfg.Tags = make(map[string]string)
		}
		for _, key := range keys {
			v, ok := m[key]
			if !ok {
				continue
			}
			if _, ok := cfg.Tags[key]; !ok {
				newkey := replaceLabelKey(key)
				cfg.Tags[newkey] = v
			}
		}
	}
}

func replaceLabelKey(s string) string {
	return strings.ReplaceAll(s, ".", "_")
}

func (lc *logInstance) tags() map[string]string {
	m := map[string]string{
		"container_id":   lc.id,
		"container_name": lc.containerName,
		"image":          lc.image,
	}

	if lc.podName != "" {
		m["pod_name"] = lc.podName
		m["pod_ip"] = lc.podIP
		m["namespace"] = lc.podNamespace
	}

	if lc.ownerKind != "" && lc.ownerName != "" {
		m[lc.ownerKind] = lc.ownerName
	}

	return m
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
