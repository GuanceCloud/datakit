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
	"sort"
	"strings"
	"sync"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
)

const logConfigAnnotationKeyFormat = "datakit/%slogs"

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
	configTemplate                      string
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

func (ins *logInstance) parseLogConfigs() error {
	if ins.configStr != "" {
		var configs logConfigs
		if err := json.Unmarshal([]byte(ins.configStr), &configs); err != nil {
			return fmt.Errorf("failed to parse configs from container %s, err: %w, data: %s",
				ins.containerName, err, ins.configStr)
		}
		ins.configs = configs

		for _, cfg := range ins.configs {
			if cfg.Disable {
				continue
			}
			if cfg.Multiline != "" {
				cfg.MultilinePatterns = []string{cfg.Multiline}
			}

			if cfg.Path == "" {
				continue
			}

			path := filepath.Clean(cfg.Path)
			foundHostPath := false

			for vol, hostdir := range ins.volMounts {
				if strings.HasPrefix(path, vol) {
					cfg.hostDir = hostdir
					cfg.insideDir = vol
					cfg.hostFilePath = joinHostFilepath(hostdir, vol, path)

					foundHostPath = true
				}
			}

			if !foundHostPath {
				return fmt.Errorf("unexpected log path %s, no matched mounts(%d) found", cfg.Path, len(ins.volMounts))
			}
		}
	}
	return nil
}

func (ins *logInstance) addStdout() {
	if len(ins.configs) == 0 {
		ins.configs = append(ins.configs, &logConfig{
			Path:   ins.logPath,
			Source: ins.containerName,
		})
		return
	}

	for _, cfg := range ins.configs {
		if cfg.Disable {
			continue
		}
		if (cfg.Type == "" || cfg.Type == "stdout") && cfg.Path == "" {
			cfg.Path = ins.logPath
		}
	}
}

func (ins *logInstance) fillLogType(runtimeName string) {
	for _, cfg := range ins.configs {
		if cfg.Type != "" {
			continue
		}
		cfg.Type = runtimeName
	}
}

func (ins *logInstance) fillSource() {
	for _, cfg := range ins.configs {
		if cfg.Source != "" {
			continue
		}
		cfg.Source = ins.containerName
	}
}

func (ins *logInstance) checkTagsKey() {
	for _, cfg := range ins.configs {
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

func (ins *logInstance) setTagsToLogConfigs(m map[string]string) {
	if len(m) == 0 {
		return
	}
	for _, cfg := range ins.configs {
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

func (ins *logInstance) setLabelAsTags(m map[string]string, all bool, keys []string) {
	if len(m) == 0 {
		return
	}

	if all {
		for _, cfg := range ins.configs {
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

	for _, cfg := range ins.configs {
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

func (ins *logInstance) tags() map[string]string {
	m := map[string]string{
		"container_id":   ins.id,
		"container_name": ins.containerName,
		"image":          ins.image,
	}

	if ins.podName != "" {
		m["pod_name"] = ins.podName
		m["pod_ip"] = ins.podIP
		m["namespace"] = ins.podNamespace
	}

	if ins.ownerKind != "" && ins.ownerName != "" {
		m[ins.ownerKind] = ins.ownerName
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

type logTable struct {
	table map[string]map[string]func()
	mu    sync.Mutex
}

func newLogTable() *logTable {
	return &logTable{
		table: make(map[string]map[string]func()),
	}
}

func (t *logTable) addToTable(id, path string, cancel func()) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.table[id] == nil {
		t.table[id] = make(map[string]func())
	}
	t.table[id][path] = cancel
}

func (t *logTable) closeFromTable(id string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	for _, cancel := range t.table[id] {
		if cancel != nil {
			cancel()
		}
	}
}

func (t *logTable) closeAll() {
	t.mu.Lock()
	defer t.mu.Unlock()

	for _, cancels := range t.table {
		for _, cancel := range cancels {
			if cancel != nil {
				cancel()
			}
		}
	}
}

func (t *logTable) removeFromTable(id string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.table, id)
}

func (t *logTable) removePathFromTable(id, path string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	paths, ok := t.table[id]
	if !ok {
		return
	}
	if paths != nil {
		delete(paths, path)
	}
}

func (t *logTable) inTable(id, path string) bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.table[id] == nil {
		return false
	}
	_, ok := t.table[id][path]
	return ok
}

func (t *logTable) findDifferences(ids []string) []string {
	t.mu.Lock()
	defer t.mu.Unlock()

	var differences []string
	for id := range t.table {
		found := false
		for _, k := range ids {
			if k == id {
				found = true
				break
			}
		}
		if !found {
			differences = append(differences, id)
		}
	}
	return differences
}

func (t *logTable) String() string {
	t.mu.Lock()
	defer t.mu.Unlock()

	var str []string
	var ids []string

	for id := range t.table {
		ids = append(ids, id)
	}

	sort.Strings(ids)

	for _, id := range ids {
		paths, ok := t.table[id]
		if !ok {
			continue
		}

		var p []string

		for path := range paths {
			p = append(p, path)
		}
		sort.Strings(p)

		shortID := id
		if len(id) > 12 {
			shortID = id[:12]
		}

		str = append(str, fmt.Sprintf("{id:%s,paths:[%s]}", shortID, strings.Join(p, ",")))
	}

	return strings.Join(str, ", ")
}

func IsClosed(ch <-chan interface{}) bool {
	select {
	case <-ch:
		return true
	default:
	}
	return false
}
