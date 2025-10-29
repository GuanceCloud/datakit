// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package container

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/container/runtime"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/logtail/multiline"
)

type containerLogInfo struct {
	containerID   string
	containerName string
	runtime       string
	image         string
	logPath       string

	podNamespace, podName string
	podIP                 string
	podLabels             map[string]string
	ownerKind, ownerName  string

	mergedDir string
	mounts    runtime.Mounts
}

func (info *containerLogInfo) buildTags() map[string]string {
	tags := map[string]string{
		"container_id":   info.containerID,
		"container_name": info.containerName,
		"image":          info.image,
	}
	if info.podName != "" {
		tags["pod_name"] = info.podName
		tags["pod_ip"] = info.podIP
		tags["namespace"] = info.podNamespace
	}
	if info.ownerKind != "" && info.ownerName != "" {
		tags[info.ownerKind] = info.ownerName
	}
	return tags
}

type logConfig struct {
	Disable               bool              `json:"disable"`
	Type                  string            `json:"type"`
	Path                  string            `json:"path"`
	Source                string            `json:"source"`
	StorageIndex          string            `json:"storage_index"`
	Service               string            `json:"service"`
	CharacterEncoding     string            `json:"character_encoding"`
	Pipeline              string            `json:"pipeline"`
	Multiline             string            `json:"multiline_match"`
	RemoveAnsiEscapeCodes bool              `json:"remove_ansi_escape_codes"`
	FromBeginning         bool              `json:"from_beginning"`
	Tags                  map[string]string `json:"tags"`

	multilinePatterns []string `json:"-"`
	hostDir           string   `json:"-"`
	insideDir         string   `json:"-"`
	hostFilePath      string   `json:"-"`
}

func newLogConfigs(defaults *loggingDefaults, info *containerLogInfo, str string) ([]*logConfig, error) {
	var configs []*logConfig

	// add default stdout
	if str == "" {
		configs = append(configs, &logConfig{
			Type:   info.runtime,
			Path:   info.logPath,
			Source: info.containerName,
		})
	} else {
		if err := json.Unmarshal([]byte(str), &configs); err != nil {
			return nil, fmt.Errorf("faild to parse log configs, container %s, err %w", info.containerName, err)
		}
	}

	for _, cfg := range configs {
		if cfg.Disable {
			continue
		}

		cfg.fillDefaultStdout(info)

		if cfg.Path == "" {
			continue
		}

		if err := cfg.setVolumePath(info); err != nil {
			l.Warnf("resolve host path failed: container=%s path=%s mergedDir=%s err=%v", info.containerName, cfg.Path, info.mergedDir, err)
			continue
		}

		cfg.fillDefaultSource(info)
		cfg.addTags(info.buildTags())
		cfg.addTags(defaults.extraTags)
		cfg.addTags(defaults.setLabelAsTags(info.podLabels))
		cfg.replacedTagsKey()

		cfg.setAutoMultiline(defaults)
		cfg.setExtraSourceMap(defaults)
		cfg.setSourceMultilineMap(defaults)
	}

	if hasDuplicatePath(configs) {
		return nil, fmt.Errorf("configs(len=%d) has duplicate path", len(configs))
	}

	return configs, nil
}

func hasDuplicatePath(configs []*logConfig) bool {
	paths := make(map[string]interface{})
	for _, cfg := range configs {
		if _, exists := paths[cfg.Path]; exists {
			return true
		}
		paths[cfg.Path] = nil
	}
	return false
}

func fillLogConfigs(defaults *loggingDefaults, info *containerLogInfo, configs []*logConfig) ([]*logConfig, error) {
	b, err := json.Marshal(configs)
	if err != nil {
		return nil, err
	}
	return newLogConfigs(defaults, info, string(b))
}

func (cfg *logConfig) getStructHash() string {
	data, err := json.Marshal(cfg)
	if err != nil {
		return ""
	}
	hash := sha256.Sum256(data)
	return fmt.Sprintf("%x", hash)
}

func (cfg *logConfig) fillDefaultSource(info *containerLogInfo) {
	if cfg.Source == "" {
		cfg.Source = info.containerName
	}
}

func (cfg *logConfig) fillDefaultStdout(info *containerLogInfo) {
	if cfg.Type == "" || cfg.Type == "stdout" {
		cfg.Type = info.runtime
		cfg.Path = info.logPath
	}
}

func (cfg *logConfig) setVolumePath(info *containerLogInfo) error {
	// skip stdout
	if cfg.Type == info.runtime {
		return nil
	}

	destinationPath, sourcePath, ok := info.mounts.FindBestMount(cfg.Path)
	if !ok {
		// 当无法通过挂载映射时，尝试通过 mergedDir + 相对路径 寻找宿主机文件：
		hostPath, err := resolveHostPathFromMergedDir(info.mergedDir, cfg.Path)
		if err != nil {
			if info.runtime == "crio" || info.runtime == "cri-o" {
				l.Warnf("runtime=%s does not support resolving path via mergedDir; please mount an extra emptyDir for logs", info.runtime)
			}
			return err
		}
		cfg.hostFilePath = hostPath
		cfg.hostDir = info.mergedDir
		l.Infof("use fallback rootfs mapping runtime=%s base=%s for mapped path=%s", info.runtime, cfg.hostDir, cfg.hostFilePath)
		return nil
	}

	cfg.insideDir = destinationPath
	cfg.hostDir = sourcePath
	cfg.hostFilePath, _ = runtime.ResolveToSourcePath(destinationPath, sourcePath, cfg.Path)

	l.Infof("use volMount destination=%s, source=%s for mapped path=%s", cfg.insideDir, cfg.hostDir, cfg.hostFilePath)
	return nil
}

func resolveHostPathFromMergedDir(mergedDir, insidePath string) (string, error) {
	stat, err := os.Stat(mergedDir)
	if err != nil || !stat.IsDir() {
		return "", fmt.Errorf("rootfs base directory not found: %s", mergedDir)
	}

	relInside := strings.TrimPrefix(insidePath, "/")
	full := filepath.Join(mergedDir, relInside)

	return full, nil
}

func (cfg *logConfig) addTags(m map[string]string) {
	if cfg.Tags == nil {
		cfg.Tags = make(map[string]string)
	}
	for k, v := range m {
		cfg.Tags[k] = v
	}
}

func (cfg *logConfig) replacedTagsKey() {
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

func (cfg *logConfig) setAutoMultiline(defaults *loggingDefaults) {
	if cfg.Multiline != "" {
		cfg.multilinePatterns = []string{cfg.Multiline}
		return
	}

	if !defaults.autoMultilineDetection {
		return
	}
	cfg.multilinePatterns = defaults.autoMultilineExtraPatterns
	cfg.multilinePatterns = append(cfg.multilinePatterns, multiline.GlobalPatterns...)
}

func (cfg *logConfig) setExtraSourceMap(defaults *loggingDefaults) {
	for re, newSource := range defaults.extraSourceMap {
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

func (cfg *logConfig) setSourceMultilineMap(defaults *loggingDefaults) {
	if defaults.sourceMultilineMap == nil {
		return
	}
	mult := defaults.sourceMultilineMap[cfg.Source]
	if mult != "" {
		l.Infof("replaced multiline '%s' with '%s' to source %s", cfg.Multiline, mult, cfg.Source)
		cfg.multilinePatterns = []string{mult}
	}
}

func isAllDisable(configs []*logConfig) bool {
	disable := true
	for _, config := range configs {
		disable = disable && config.Disable
	}
	return disable
}
