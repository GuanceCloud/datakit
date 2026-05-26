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
)

type containerLogInfo struct {
	containerID   string
	containerName string
	runtime       string
	image         string
	logPath       string

	podUID                string
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
	Disable                    bool              `json:"disable"`
	Type                       string            `json:"type"`
	Path                       string            `json:"path"`
	Source                     string            `json:"source"`
	StorageIndex               string            `json:"storage_index"`
	Service                    string            `json:"service"`
	CharacterEncoding          string            `json:"character_encoding"`
	Pipeline                   string            `json:"pipeline"`
	Multiline                  string            `json:"multiline_match"`
	RemoveAnsiEscapeCodes      bool              `json:"remove_ansi_escape_codes"`
	FromBeginning              bool              `json:"from_beginning"`
	FromBeginningThresholdSize int64             `json:"from_beginning_threshold_size"`
	Tags                       map[string]string `json:"tags"`

	multilinePattern string   `json:"-"`
	autoMultiline    bool     `json:"-"`
	extraPatterns    []string `json:"-"`
	hostDir          string   `json:"-"`
	insideDir        string   `json:"-"`
	hostFilePath     string   `json:"-"`
}

func newLogConfigs(defaults *loggingDefaults, info *containerLogInfo, str string) ([]*logConfig, bool, error) {
	var configs []*logConfig
	var useDefaultStdoutConfigs bool

	// add default stdout
	if str == "" {
		configs = append(configs, &logConfig{
			Type:   info.runtime,
			Path:   info.logPath,
			Source: info.containerName,
		})
		useDefaultStdoutConfigs = true
	} else {
		if err := json.Unmarshal([]byte(str), &configs); err != nil {
			return nil, false, fmt.Errorf("faild to parse log configs, container %s, err %w", info.containerName, err)
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
		return nil, false, fmt.Errorf("configs(len=%d) has duplicate path", len(configs))
	}

	return configs, useDefaultStdoutConfigs, nil
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

func fillLogConfigsWithCRDLogging(
	defaults *loggingDefaults,
	info *containerLogInfo,
	crdConfigs *crdLoggingConfig,
) ([]*logConfig, error) {
	b, err := json.Marshal(crdConfigs.configs)
	if err != nil {
		return nil, err
	}

	configs, _, err := newLogConfigs(defaults, info, string(b))
	if err != nil {
		return nil, err
	}

	for _, cfg := range configs {
		if cfg.Disable {
			continue
		}
		cfg.addTags(setLabelAsTags(false, labelsOption{keys: crdConfigs.podTargetLabels}, info.podLabels))
	}

	return configs, nil
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

	baseDir, err := filepath.Abs(mergedDir)
	if err != nil {
		return "", fmt.Errorf("abs rootfs base directory failed: %w", err)
	}

	cleanInside := filepath.Clean(insidePath)
	if filepath.IsAbs(cleanInside) {
		cleanInside = strings.TrimPrefix(cleanInside, string(filepath.Separator))
	}

	full := filepath.Join(baseDir, cleanInside)
	rel, err := filepath.Rel(baseDir, full)
	if err != nil {
		return "", fmt.Errorf("calc relative path failed: %w", err)
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("resolved path escaped rootfs base directory: %s", insidePath)
	}

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
		cfg.multilinePattern = cfg.Multiline
		cfg.autoMultiline = false
		cfg.extraPatterns = nil
		return
	}

	if !defaults.enableMultiline {
		cfg.multilinePattern = ""
		cfg.autoMultiline = false
		cfg.extraPatterns = nil
		return
	}

	cfg.multilinePattern = ""
	cfg.autoMultiline = true
	cfg.extraPatterns = append([]string{}, defaults.autoMultilineExtraPatterns...)
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
		cfg.multilinePattern = mult
		cfg.autoMultiline = false
		cfg.extraPatterns = nil
	}
}

func isAllDisable(configs []*logConfig) bool {
	disable := true
	for _, config := range configs {
		disable = disable && config.Disable
	}
	return disable
}
