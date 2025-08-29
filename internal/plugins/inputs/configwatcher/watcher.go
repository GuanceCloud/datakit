// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package configwatcher monitors files for changes.
package configwatcher

import (
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/cespare/xxhash/v2"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/diff"
)

const defaultMaxOpenFiles = 1000

type fileState struct {
	path    string
	exists  bool
	size    int64
	modTime time.Time
	hash    string
	content []byte
}

type changeType int

const (
	noChange changeType = iota
	created
	modified
	deleted
)

func (t changeType) String() string {
	switch t {
	case noChange:
		return "NoChange"
	case created:
		return "Created"
	case modified:
		return "Modified"
	case deleted:
		return "Deleted"
	default:
		return ""
	}
}

type changeEvent struct {
	typ      changeType
	path     string
	oldState *fileState
	newState *fileState
	diff     string
}

type fileWatcher struct {
	path       string
	lastStates map[string]*fileState // 存储每个文件的上次状态

	maxOpenFiles int
	maxDiffSize  int64
	recursive    bool
}

type option func(w *fileWatcher)

func withMaxDiffSize(n int64) option { return func(w *fileWatcher) { w.maxDiffSize = n } }
func withRecursive(b bool) option    { return func(w *fileWatcher) { w.recursive = b } }

func newFileWatcher(path string, opts ...option) (*fileWatcher, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("invalid path: %w", err)
	}

	watcher := &fileWatcher{
		path:         absPath,
		lastStates:   make(map[string]*fileState),
		maxOpenFiles: defaultMaxOpenFiles,
	}
	for _, opt := range opts {
		opt(watcher)
	}

	// 初始扫描
	states := watcher.scanPath()
	for path, state := range states {
		watcher.lastStates[path] = state
	}

	return watcher, nil
}

func (w *fileWatcher) scanPath() map[string]*fileState {
	states := make(map[string]*fileState)

	// 扫描函数
	scanFunc := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // 忽略无法访问的文件或目录
		}

		if isHiddenFile(path) {
			return nil
		}

		if len(states) > w.maxOpenFiles {
			return nil
		}

		// 如果是目录，根据递归设置决定是否继续
		if info.IsDir() {
			if path == w.path || w.recursive {
				return nil // 继续遍历
			}
			return filepath.SkipDir // 跳过子目录
		}

		// 处理文件
		state := &fileState{
			path:    path,
			exists:  true,
			size:    info.Size(),
			modTime: info.ModTime(),
		}

		if state.size <= w.maxDiffSize {
			content, hash, err := w.readFileContent(path)
			if err == nil {
				state.hash = hash
				state.content = content
			}
		} else {
			hash, err := w.calculateFileHash(path)
			if err == nil {
				state.hash = hash
			}
		}

		states[path] = state
		return nil
	}

	filepath.Walk(w.path, scanFunc) //nolint:errcheck,gosec
	return states
}

func (w *fileWatcher) readFileContent(path string) ([]byte, string, error) {
	content, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return nil, "", err
	}

	hasher := xxhash.New()
	if _, err := hasher.Write(content); err != nil {
		return nil, "", err
	}
	return content, hex.EncodeToString(hasher.Sum(nil)), nil
}

func (w *fileWatcher) calculateFileHash(path string) (string, error) {
	file, err := os.Open(filepath.Clean(path))
	if err != nil {
		return "", err
	}
	defer file.Close() //nolint:errcheck,gosec

	hasher := xxhash.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hasher.Sum(nil)), nil
}

func (w *fileWatcher) checkChanges() ([]changeEvent, error) {
	currentStates := w.scanPath()
	var events []changeEvent

	// 检查已删除的文件
	for path, lastState := range w.lastStates {
		if _, exists := currentStates[path]; !exists && lastState.exists {
			events = append(events, changeEvent{
				typ:      deleted,
				path:     path,
				oldState: lastState,
				newState: &fileState{path: path, exists: false},
			})
		}
	}

	// 检查新增或修改的文件
	for path, currentState := range currentStates {
		lastState, exists := w.lastStates[path]

		if !exists {
			// 新增文件
			event := changeEvent{
				typ:      created,
				path:     path,
				oldState: nil,
				newState: currentState,
			}
			if currentState.size <= w.maxDiffSize {
				event.diff = w.generateDiff(nil, currentState.content)
			}
			events = append(events, event)
		} else if lastState.exists && currentState.exists {
			// 检查文件是否修改
			if lastState.modTime != currentState.modTime || lastState.size != currentState.size {
				// 如果文件大小合适，检查哈希是否变化
				if lastState.size <= w.maxDiffSize && currentState.size <= w.maxDiffSize {
					if lastState.hash != currentState.hash {
						event := changeEvent{
							typ:      modified,
							path:     path,
							oldState: lastState,
							newState: currentState,
						}
						event.diff = w.generateDiff(lastState.content, currentState.content)
						events = append(events, event)
					}
				} else {
					// 对于大文件，只要 ModTime 或 size 变化就认为是修改
					events = append(events, changeEvent{
						typ:      modified,
						path:     path,
						oldState: lastState,
						newState: currentState,
					})
				}
			}
		}
	}

	w.lastStates = currentStates
	return events, nil
}

func (w *fileWatcher) generateDiff(oldContent, newContent []byte) string {
	return diff.LineDiffWithContextLines(string(oldContent), string(newContent), 4)
}

func (w *fileWatcher) getCurrentStates() map[string]*fileState {
	return w.lastStates
}

func (w *fileWatcher) Reset() {
	w.lastStates = w.scanPath()
}
