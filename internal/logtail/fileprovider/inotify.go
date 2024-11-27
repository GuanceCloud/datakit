// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package fileprovider

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/fsnotify/fsnotify"
)

type InotifyInterface interface {
	Events() chan fsnotify.Event
	Close() error
}

type Inotify struct {
	watcher *fsnotify.Watcher
}

func NewInotify(patterns []string) (*Inotify, error) {
	var dirs []string

	for _, pattern := range patterns {
		var dir string
		if starIdx := strings.Index(pattern, "*"); starIdx != -1 {
			dir = filepath.Join(filepath.Dir(pattern[:starIdx]), "...")
		} else {
			dir = filepath.Dir(pattern)
		}
		dirs = append(dirs, dir)
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("unable create inotfiy, err: %w", err)
	}

	for _, dir := range dirs {
		err := watcher.Add(dir)
		if err != nil {
			_ = watcher.Close()
			return nil, fmt.Errorf("failed to add watcher %s, err: %w", dir, err)
		}
	}

	return &Inotify{watcher}, nil
}

func (in *Inotify) Events() chan fsnotify.Event {
	return in.watcher.Events
}

func (in *Inotify) Close() error {
	return in.watcher.Close()
}

type NopInotify struct{}

func NewNopInotify() *NopInotify { return &NopInotify{} }

func (*NopInotify) Events() chan fsnotify.Event { return nil }

func (*NopInotify) Close() error { return nil }
