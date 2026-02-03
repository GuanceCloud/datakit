// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build !windows
// +build !windows

package hostchange

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/changes"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/diff"
)

const (
	ChangeIDFileChange = "host_change_03_01"
)

type FileChange struct {
	FilePath   string
	OldContent string
	NewContent string
	Diff       string
}

type FileChangesByPath map[string]*FileChange

type FileChecker struct {
	Enabled     bool     `toml:"enabled"`
	Files       []string `toml:"files"`
	IgnorePaths []string `toml:"ignore_paths"`
	MaxFileSize int64    `toml:"max_file_size"` // Maximum file size to store full content (in bytes, 0 means no limit, larger files only store hash)
	input       *Input

	fileChange *Change[string, *FileChange]
}

func (fc *FileChecker) Init(ipt *Input) error {
	files := []string{}

	for _, file := range fc.Files {
		if !filepath.IsAbs(file) {
			l.Warnf("file path must be absolute: %s", file)
			continue
		}

		info, err := os.Stat(file)
		if err == nil {
			if info.IsDir() {
				l.Warnf("file path must be a regular file, not a directory: %s", file)
				continue
			}
		} else {
			l.Warnf("file not found: %s", file)
		}

		files = append(files, file)
	}
	fc.Files = files

	fc.input = ipt
	return nil
}

func readFileAndCalculateHash(file string, useHashOnly bool) (content string, hash uint64, err error) {
	if useHashOnly {
		fileHash, err := GetFileHash(file)
		if err != nil {
			return "", 0, fmt.Errorf("failed to get file hash: %w", err)
		}
		return "", fileHash, nil
	}

	contentBytes, err := ReadFile(file)
	if err != nil {
		return "", 0, fmt.Errorf("failed to read file: %w", err)
	}
	contentStr := string(contentBytes)

	return contentStr, GetHashCode(contentBytes), nil
}

func (fc *FileChecker) add(changesByPath FileChangesByPath) func(key string, newValues, oldValues []*FileChange,
	parentChanges ...*Change[string, *FileChange]) *ChangeItem {
	return func(key string, newValues, oldValues []*FileChange, parentChanges ...*Change[string, *FileChange]) *ChangeItem {
		if len(newValues) == 0 {
			return nil
		}

		fileChange := newValues[0]
		changesByPath[key] = fileChange
		return nil
	}
}

func (fc *FileChecker) delete(changesByPath FileChangesByPath) func(key string, newValues,
	oldValues []*FileChange, parentChanges ...*Change[string, *FileChange]) *ChangeItem {
	return func(key string, newValues, oldValues []*FileChange, parentChanges ...*Change[string, *FileChange]) *ChangeItem {
		if len(oldValues) == 0 {
			return nil
		}

		fileChange := oldValues[0]
		fileChange.OldContent = fileChange.NewContent
		fileChange.NewContent = ""
		fileChange.Diff = fmt.Sprintf("File deleted: %s", fileChange.FilePath)
		changesByPath[key] = fileChange
		return nil
	}
}

func (fc *FileChecker) modify(changesByPath FileChangesByPath) func(key string,
	newValues, oldValues []*FileChange,
	parentChanges ...*Change[string, *FileChange]) *ChangeItem {
	return func(key string, newValues, oldValues []*FileChange, parentChanges ...*Change[string, *FileChange]) *ChangeItem {
		if len(newValues) == 0 || len(oldValues) == 0 {
			return nil
		}

		newFileChange := newValues[0]
		oldFileChange := oldValues[0]

		mergedFileChange := &FileChange{
			FilePath:   newFileChange.FilePath,
			OldContent: oldFileChange.NewContent,
			NewContent: newFileChange.NewContent,
		}

		if fc.MaxFileSize > 0 && len(mergedFileChange.NewContent) > int(fc.MaxFileSize) {
			mergedFileChange.Diff = fmt.Sprintf("Large file changed (size: %d bytes, max allowed: %d bytes)",
				len(mergedFileChange.NewContent), fc.MaxFileSize)
		} else {
			diffResult := diff.LineDiffWithContextLines(mergedFileChange.OldContent, mergedFileChange.NewContent, 4)
			mergedFileChange.Diff = diffResult
		}

		changesByPath[key] = mergedFileChange
		return nil
	}
}

func (fc *FileChecker) getChangeEvent(file string, fileChange *FileChange) *ChangeItem {
	changeData := map[string]interface{}{
		"FilePath": fileChange.FilePath,
		"Diff":     fileChange.Diff,
	}

	title, message, err := changes.RenderHostTemplate(defaultChangeLanguage, ChangeIDFileChange, changeData)
	if err != nil {
		l.Warnf("RenderHostTemplate failed for %s: %v, using default message", ChangeIDFileChange, err.Error())
		title = fmt.Sprintf("File Changed: %s", file)
		message = fmt.Sprintf("File %s has been modified", file)
	}

	return &ChangeItem{
		ChangeID:             ChangeIDFileChange,
		ChangeTimestampMicro: time.Now().UnixMicro(),
		Title:                title,
		Message:              message,
	}
}

func (fc *FileChecker) Collect() ([]*ChangeItem, error) {
	l.Debugf("collecting file changes")

	changesByPath := make(FileChangesByPath)

	newChange := NewChange[string, *FileChange]()

	for _, file := range fc.Files {
		if fc.shouldIgnore(file) {
			continue
		}

		fileInfo, err := os.Stat(file)
		if err != nil {
			if os.IsNotExist(err) {
				l.Infof("file %s not exist", file)
				if fc.fileChange != nil {
					oldChild := fc.fileChange.GetChild(file)
					if oldChild != nil {
						for _, oldFileChanges := range oldChild.data {
							for _, oldFileChange := range oldFileChanges {
								oldFileChange.OldContent = oldFileChange.NewContent
								oldFileChange.NewContent = ""
								oldFileChange.Diff = fmt.Sprintf("File deleted: %s", file)
								changesByPath[file] = oldFileChange
							}
						}
					}
				}
			} else {
				l.Warnf("failed to stat file %s: %s", file, err.Error())
			}
			continue
		}

		if fileInfo.IsDir() {
			l.Debugf("skip dir %s", file)
			continue
		}

		useHashOnly := fc.MaxFileSize > 0 && fileInfo.Size() > fc.MaxFileSize

		currentContent, contentHash, err := readFileAndCalculateHash(file, useHashOnly)
		if err != nil {
			l.Warnf("failed to read file %s: %s", file, err.Error())
			continue
		}

		childChange := NewChange[string, *FileChange]()
		childChange.LastUpdateTime = fileInfo.ModTime()
		childChange.ContentHash = contentHash
		childChange.RawKey = file

		if fc.fileChange != nil {
			oldChild := fc.fileChange.GetChild(file)
			if oldChild != nil {
				if oldChild.LastUpdateTime.Equal(fileInfo.ModTime()) {
					childChange = oldChild
					newChange.SetChild(file, childChange)
					continue
				}

				if oldChild.ContentHash == contentHash {
					childChange.LastUpdateTime = fileInfo.ModTime()
					newChange.SetChild(file, childChange)
					continue
				}
			}
		}

		fileChange := &FileChange{
			FilePath:   file,
			NewContent: currentContent,
		}

		if useHashOnly {
			fileChange.Diff = fmt.Sprintf("Large file changed (size: %d bytes, max allowed: %d bytes)", fileInfo.Size(), fc.MaxFileSize)
		} else {
			fileChange.Diff = fmt.Sprintf("New file created: %s", file)
		}

		childChange.Add(file, fileChange)
		newChange.SetChild(file, childChange)
	}

	if fc.fileChange != nil {
		_, err := newChange.GetChangeEvent(fc.fileChange, &ChangeItemConfig[string, *FileChange]{
			Add:    fc.add(changesByPath),
			Delete: fc.delete(changesByPath),
			Modify: fc.modify(changesByPath),
		})
		if err != nil {
			l.Warnf("failed to get file change event: %s", err.Error())
		}
	}

	fc.fileChange = newChange

	var changeItems []*ChangeItem
	for file, fileChange := range changesByPath {
		changeItem := fc.getChangeEvent(file, fileChange)
		changeItems = append(changeItems, changeItem)
	}

	return changeItems, nil
}

func (fc *FileChecker) shouldIgnore(file string) bool {
	for _, ignorePath := range fc.IgnorePaths {
		if match, _ := filepath.Match(ignorePath, file); match {
			return true
		}
		if strings.HasPrefix(file, ignorePath) {
			return true
		}
	}
	return false
}
