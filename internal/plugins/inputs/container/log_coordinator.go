// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package container

import (
	"context"
	"regexp"
	"sync"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/container/runtime"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	loggingv1alpha1 "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/kubernetes/pkg/apis/datakits/v1alpha1"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/kubernetes/pkg/labels"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/tailer"
)

type containerLogCoordinator struct {
	containerTasks map[string]*containerLogTask
	taskMutex      sync.RWMutex

	crdConfigs  []*crdLoggingConfig
	configMutex sync.RWMutex

	defaults *loggingDefaults
}

type containerLogTask struct {
	containerID                  string
	info                         *containerLogInfo
	useAnnotationOrEnvLogConfigs bool
	configStr                    string

	configs []*logConfig
	tailers []TailerItem
}

type TailerItem struct {
	path       string
	configHash string
	tailer     *tailer.Tailer
}

func newContainerLogCoordinator(defaults *loggingDefaults) *containerLogCoordinator {
	return &containerLogCoordinator{
		containerTasks: make(map[string]*containerLogTask),
		crdConfigs:     make([]*crdLoggingConfig, 0),
		defaults:       defaults,
	}
}

func (c *containerLogCoordinator) addTask(containerID string, info *containerLogInfo, configStr string) {
	c.taskMutex.Lock()
	defer c.taskMutex.Unlock()

	l.Debugf("addTask start - containerID=%s, existingTasks=%d", containerID, len(c.containerTasks))

	task, exists := c.containerTasks[containerID]
	if !exists {
		task = &containerLogTask{
			containerID: containerID,
			info:        info,
		}
	}
	task.useAnnotationOrEnvLogConfigs = configStr != ""

	if exists {
		if task.useAnnotationOrEnvLogConfigs && task.configStr != configStr {
			return
		}
	} else {
		l.Infof("creating new log task for container %s", containerID)
	}

	task.configStr = configStr

	var configs []*logConfig
	var err error

	if !task.useAnnotationOrEnvLogConfigs {
		if crd := c.matchCRDConfigs(task); crd != nil {
			configs = crd
		} else {
			configs, err = newLogConfigs(c.defaults, info, configStr)
			if err != nil {
				l.Errorf("failed to parse log configs for container %s: %v", containerID, err)
				return
			}
			// no-op
		}
	} else {
		configs, err = newLogConfigs(c.defaults, info, configStr)
		if err != nil {
			l.Errorf("failed to parse log configs for container %s: %v", containerID, err)
			return
		}
	}

	l.Debugf("task exists for container %s but config changed, updating task", containerID)

	// 当路径集合发生变化（例如新增或缺少某个 path）时，关闭现有 tailer，后续按正常流程重新创建
	c.closeTailersIfPathChanged(task, configs)
	// 关闭被显式 Disable 的 path 对应的 tailer（仅针对被禁用的路径，不影响其他路径）
	c.closeTailersForDisabledPaths(task, configs)

	if isAllDisable(configs) {
		return
	}

	for _, cfg := range configs {
		if cfg == nil || cfg.Disable {
			continue
		}

		if updated, idx := c.updateTailerForTask(task, cfg); idx >= 0 {
			if updated {
				task.configs[idx] = cfg
			}
			continue
		}
		c.createTailerForTask(task, cfg)
		task.configs = append(task.configs, cfg)
	}

	c.containerTasks[containerID] = task
	l.Infof("added task for container %s, tailers=%d", containerID, len(task.tailers))
}

func (c *containerLogCoordinator) cleanMissingContainerLog(activeContainers []string) {
	activeSet := make(map[string]bool)
	for _, id := range activeContainers {
		activeSet[id] = true
	}

	// 在读锁下收集需要移除的容器 ID，避免在持有写锁时调用 removeTask 导致死锁
	c.taskMutex.RLock()
	toRemove := make([]string, 0)

	for containerID := range c.containerTasks {
		if !activeSet[containerID] {
			toRemove = append(toRemove, containerID)
		}
	}
	// 预估移除后的剩余任务数量：len(c.containerTasks) - len(toRemove)
	l.Debugf("cleanMissingContainerLog: will remove=%d, remaining(after)=%d", len(toRemove), len(c.containerTasks)-len(toRemove))
	c.taskMutex.RUnlock()

	// 在不持锁的情况下逐个删除（removeTask 内部自行加锁）
	for _, containerID := range toRemove {
		c.removeTask(containerID)
		l.Infof("found inactive container %s, cleaning up log task", containerID)
	}
}

func (c *containerLogCoordinator) removeTask(containerID string) {
	c.taskMutex.Lock()
	defer c.taskMutex.Unlock()

	if task, exists := c.containerTasks[containerID]; exists {
		for _, it := range task.tailers {
			if it.tailer != nil {
				it.tailer.Close()
			}
		}
		delete(c.containerTasks, containerID)
		l.Infof("removed task for container %s", containerID)
	}
}

func (c *containerLogCoordinator) updateCRDLoggingConfig(key string, config *crdLoggingConfig) {
	c.configMutex.Lock()
	defer c.configMutex.Unlock()

	// 根据 key 替换已存在的配置；若不存在则追加到末尾，保持匹配顺序稳定
	replaced := false
	for i := range c.crdConfigs {
		if c.crdConfigs[i] != nil && c.crdConfigs[i].key == key {
			c.crdConfigs[i] = config
			replaced = true
			break
		}
	}
	if !replaced {
		c.crdConfigs = append(c.crdConfigs, config)
	}

	l.Infof("updated CRD config for key %s, current total=%d", key, len(c.crdConfigs))
}

func (c *containerLogCoordinator) deleteCRDLoggingConfig(key string) {
	c.configMutex.Lock()
	defer c.configMutex.Unlock()

	// 按 key 删除，同时保持切片顺序不变
	for i := range c.crdConfigs {
		if c.crdConfigs[i] != nil && c.crdConfigs[i].key == key {
			c.crdConfigs = append(c.crdConfigs[:i], c.crdConfigs[i+1:]...)
			break
		}
	}
	l.Infof("deleted CRD config for key %s, current total=%d", key, len(c.crdConfigs))
}

func (c *containerLogCoordinator) matchCRDConfigs(task *containerLogTask) []*logConfig {
	c.configMutex.RLock()
	defer c.configMutex.RUnlock()

	// 遍历 CRD 配置，返回第一个匹配项，以确保匹配顺序稳定
	for _, crdConfig := range c.crdConfigs {
		if c.matchesCRDConfig(task, crdConfig) {
			newConfigs, err := fillLogConfigs(c.defaults, task.info, crdConfig.configs)
			if err != nil {
				l.Warnf("apply CRD config failed: key=%s container=%s err=%v", crdConfig.key, task.containerID, err)
				break
			}

			l.Infof("applied CRD config %s to container %s", crdConfig.key, task.containerID)
			return newConfigs
		}
	}
	return nil
}

func (c *containerLogCoordinator) matchesCRDConfig(task *containerLogTask, crdConfig *crdLoggingConfig) bool {
	info := task.info

	if crdConfig.namespaceMatch != nil && !crdConfig.namespaceMatch.MatchString(info.podNamespace) {
		l.Debugf("matchesCRDConfig: container=%s namespace not match, ns=%s regex=%s", task.containerID, info.podNamespace, crdConfig.namespaceRegex)
		return false
	}
	if crdConfig.podMatch != nil && !crdConfig.podMatch.MatchString(info.podName) {
		l.Debugf("matchesCRDConfig: container=%s pod not match, pod=%s regex=%s", task.containerID, info.podName, crdConfig.podRegex)
		return false
	}
	if crdConfig.containerMatch != nil && !crdConfig.containerMatch.MatchString(info.containerName) {
		l.Debugf("matchesCRDConfig: container=%s container not match, name=%s regex=%s", task.containerID, info.containerName, crdConfig.containerRegex)
		return false
	}
	if crdConfig.podLabelSelectorMatch != nil && info.podLabels != nil {
		if !crdConfig.podLabelSelectorMatch.Matches(labels.Set(info.podLabels)) {
			l.Debugf("matchesCRDConfig: container=%s pod labels not match selector=%s", task.containerID, crdConfig.podLabelSelector)
			return false
		}
	}
	l.Debugf("matchesCRDConfig: container=%s matched", task.containerID)
	return true
}

func (c *containerLogCoordinator) updateTailerForTask(task *containerLogTask, newCfg *logConfig) (bool, int) {
	var existing *tailer.Tailer
	targetPath := newCfg.Path
	tailerIndex := -1

	for idx, item := range task.tailers {
		if item.path == targetPath {
			existing = item.tailer
			tailerIndex = idx
			break
		}
	}
	if existing == nil {
		return false, -1
	}

	newHash := newCfg.getStructHash()
	l.Debugf("updateTailerForTask: currentHash=%s newHash=%s", task.tailers[tailerIndex].configHash, newHash)
	if task.tailers[tailerIndex].configHash == newHash {
		return false, tailerIndex
	}

	// 配置存在差异，执行更新
	l.Infof("config changed for container %s, path %s, will update", task.containerID, targetPath)

	opts := c.buildTailerOptions(task.info, newCfg)
	if err := existing.UpdateOptions(opts); err != nil {
		l.Errorf("failed to update tailer options for container %s: %v", task.containerID, err)
		return false, tailerIndex
	}

	task.tailers[tailerIndex].configHash = newHash
	l.Infof("successfully updated existing tailer %s config for container %s", targetPath, task.containerID)
	return true, tailerIndex
}

func (c *containerLogCoordinator) createTailerForTask(task *containerLogTask, cfg *logConfig) {
	path := cfg.Path
	for _, item := range task.tailers {
		if item.path == path {
			return
		}
	}

	tailer, err := c.createTailer(task.info, cfg)
	if err != nil {
		l.Errorf("failed to create tailer for container %s: %v", task.containerID, err)
		return
	}

	task.tailers = append(task.tailers, TailerItem{path: path, configHash: cfg.getStructHash(), tailer: tailer})

	datakit.G("container-log-tailer").Go(func(ctx context.Context) error {
		tailer.Start()
		return nil
	})

	l.Infof("started tailer for container %s, path %s", task.containerID, cfg.Path)
}

func (c *containerLogCoordinator) createTailer(info *containerLogInfo, cfg *logConfig) (*tailer.Tailer, error) {
	opts := c.buildTailerOptions(info, cfg)
	opts = append(opts, tailer.WithInsideFilepathFunc(func(filepath string) string {
		return c.defaults.insideFilepathFunc(cfg.hostDir, cfg.insideDir, filepath)
	}))

	path := cfg.Path
	if cfg.hostFilePath != "" {
		path = cfg.hostFilePath
	}

	pathAtRootfs := joinLogsAtRootfs(path)

	return tailer.NewTailer([]string{pathAtRootfs}, opts...)
}

func (c *containerLogCoordinator) closeTailersIfPathChanged(task *containerLogTask, configs []*logConfig) bool {
	if len(task.tailers) == 0 {
		return false
	}

	newPathSet := make(map[string]struct{})
	for _, cfg := range configs {
		if cfg == nil || cfg.Disable {
			continue
		}
		newPathSet[cfg.Path] = struct{}{}
	}

	runningPathSet := make(map[string]struct{})
	for _, it := range task.tailers {
		runningPathSet[it.path] = struct{}{}
	}

	pathsEqual := len(newPathSet) == len(runningPathSet)
	if pathsEqual {
		for p := range newPathSet {
			if _, ok := runningPathSet[p]; !ok {
				pathsEqual = false
				break
			}
		}
	}

	if pathsEqual {
		return false
	}

	l.Infof("log paths changed for container %s, will close all tailers", task.containerID)
	for _, it := range task.tailers {
		if it.tailer != nil {
			it.tailer.Close()
		}
	}
	task.tailers = nil
	task.configs = nil

	return true
}

func (c *containerLogCoordinator) closeTailersForDisabledPaths(task *containerLogTask, configs []*logConfig) bool {
	if len(task.tailers) == 0 || len(configs) == 0 {
		return false
	}

	disabled := make(map[string]struct{})
	for _, cfg := range configs {
		if cfg != nil && cfg.Disable {
			disabled[cfg.Path] = struct{}{}
		}
	}
	if len(disabled) == 0 {
		return false
	}

	// 关闭并过滤 tailers
	updatedTailers := make([]TailerItem, 0, len(task.tailers))
	closedAny := false
	for _, it := range task.tailers {
		if _, ok := disabled[it.path]; ok {
			if it.tailer != nil {
				it.tailer.Close()
			}
			closedAny = true
			continue
		}
		updatedTailers = append(updatedTailers, it)
	}
	if !closedAny {
		return false
	}
	task.tailers = updatedTailers

	// 同步过滤 configs（按 path 对齐）
	if len(task.configs) > 0 {
		kept := make([]*logConfig, 0, len(task.configs))
		for _, c := range task.configs {
			if c == nil {
				continue
			}
			if _, ok := disabled[c.Path]; ok {
				continue
			}
			kept = append(kept, c)
		}
		task.configs = kept
	}

	for p := range disabled {
		l.Infof("disabled log path for container %s, closed tailer: %s", task.containerID, p)
	}
	return true
}

func (c *containerLogCoordinator) buildTailerOptions(info *containerLogInfo, cfg *logConfig) []tailer.Option {
	opts := []tailer.Option{
		tailer.WithStorageIndex(cfg.StorageIndex),
		tailer.WithSource(cfg.Source),
		tailer.WithService(cfg.Service),
		tailer.WithPipeline(cfg.Pipeline),
		tailer.EnableDebugFields(c.defaults.enableDebugFields),

		tailer.WithCharacterEncoding(cfg.CharacterEncoding),
		tailer.WithExtraTags(cfg.Tags),
		tailer.WithFromBeginning(cfg.FromBeginning || c.defaults.fileFromBeginning),
		tailer.WithRemoveAnsiEscapeCodes(cfg.RemoveAnsiEscapeCodes || c.defaults.removeAnsiEscapeCodes),

		tailer.EnableMultiline(c.defaults.enableMultiline),
		tailer.WithMultilinePatterns(cfg.multilinePatterns),
		tailer.WithMaxMultilineLength(c.defaults.maxMultilineLength),

		tailer.WithMaxOpenFiles(c.defaults.maxOpenFiles),
		tailer.WithFileSizeThreshold(c.defaults.fileSizeThreshold),
		tailer.WithIgnoreDeadLog(c.defaults.ignoreDeadLog),
		tailer.WithFieldWhitelist(c.defaults.fieldWhitelist),
	}

	switch cfg.Type {
	case "file":
		opts = append(opts, tailer.WithTextParserMode(tailer.FileMode))
	case runtime.DockerRuntime:
		opts = append(opts, tailer.WithTextParserMode(tailer.DockerJSONLogMode))
	default:
		opts = append(opts, tailer.WithTextParserMode(tailer.CriLogdMode))
	}

	return opts
}

type crdLoggingConfig struct {
	key        string
	lastUpdate time.Time

	namespaceRegex   string
	podRegex         string
	podLabelSelector string
	containerRegex   string

	namespaceMatch        *regexp.Regexp
	podMatch              *regexp.Regexp
	containerMatch        *regexp.Regexp
	podLabelSelectorMatch labels.Selector

	configs []*logConfig
}

func newCRDLoggingConfig(key string, item *loggingv1alpha1.ClusterLoggingConfig) (*crdLoggingConfig, error) {
	cfg := &crdLoggingConfig{
		key:              key,
		lastUpdate:       time.Now(),
		namespaceRegex:   item.Spec.Selector.NamespaceRegex,
		podRegex:         item.Spec.Selector.PodRegex,
		podLabelSelector: item.Spec.Selector.PodLabelSelector,
		containerRegex:   item.Spec.Selector.ContainerRegex,
	}

	var err error

	if cfg.namespaceRegex != "" {
		cfg.namespaceMatch, err = regexp.Compile(cfg.namespaceRegex)
		if err != nil {
			return nil, err
		}
	}
	if cfg.podRegex != "" {
		cfg.podMatch, err = regexp.Compile(cfg.podRegex)
		if err != nil {
			return nil, err
		}
	}
	if cfg.containerRegex != "" {
		cfg.containerMatch, err = regexp.Compile(cfg.containerRegex)
		if err != nil {
			return nil, err
		}
	}
	if cfg.podLabelSelector != "" {
		cfg.podLabelSelectorMatch, err = labels.Parse(cfg.podLabelSelector)
		if err != nil {
			return nil, err
		}
	}

	for _, config := range item.Spec.Configs {
		cfg.configs = append(cfg.configs, &logConfig{
			Type:                  config.Type,
			Source:                config.Source,
			Disable:               config.Disable,
			Path:                  config.Path,
			StorageIndex:          config.StorageIndex,
			Service:               config.Service,
			CharacterEncoding:     config.CharacterEncoding,
			Pipeline:              config.Pipeline,
			Multiline:             config.Multiline,
			RemoveAnsiEscapeCodes: config.RemoveAnsiEscapeCodes,
			FromBeginning:         config.FromBeginning,
			Tags:                  config.Tags,
		})
	}

	return cfg, nil
}
