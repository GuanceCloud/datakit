// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package tailer wraps logging file collection
package tailer

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/GuanceCloud/cliutils/logger"
	"github.com/fsnotify/fsnotify"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/logtail/fileprovider"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/logtail/openfile"
)

const (
	// 定期寻找符合条件的新文件.
	scanNewFileInterval = time.Second * 10
)

var g = datakit.G("tailer")

type Tailer struct {
	options []Option

	source        string
	fromBeginning bool
	ignoreDeadLog time.Duration

	fileList map[string]context.CancelFunc

	fileScanner *fileprovider.Scanner
	fileFilter  *fileprovider.GlobFilter
	fileInotify fileprovider.InotifyInterface

	done chan interface{}
	mu   sync.Mutex

	log *logger.Logger
}

func NewTailer(patterns []string, opts ...Option) (*Tailer, error) {
	c := defaultOption()
	for _, opt := range opts {
		opt(c)
	}

	tailer := &Tailer{
		options:       opts,
		source:        c.source,
		fromBeginning: c.fromBeginning,
		ignoreDeadLog: c.ignoreDeadLog,
		fileList:      make(map[string]context.CancelFunc),

		done: make(chan interface{}),
		log:  logger.SLogger("tailer/" + c.source),
	}

	var err error

	tailer.fileScanner, err = fileprovider.NewScanner(patterns)
	if err != nil {
		return nil, fmt.Errorf("failed to new scanner, err: %w", err)
	}

	tailer.fileFilter, err = fileprovider.NewGlobFilter(patterns, c.ignorePatterns)
	if err != nil {
		return nil, fmt.Errorf("failed to new filter, err: %w", err)
	}

	if runtime.GOOS == datakit.OSLinux {
		tailer.fileInotify, err = fileprovider.NewInotify(patterns)
	} else {
		tailer.fileInotify = fileprovider.NewNopInotify()
	}

	return tailer, err
}

func (t *Tailer) Start() {
	defer func() {
		t.closeAllFiles()
		_ = g.Wait()
		t.log.Info("all exit")
	}()

	ticker := time.NewTicker(scanNewFileInterval)
	defer ticker.Stop()

	ctx := context.Background()

	for {
		files, err := t.fileScanner.ScanFiles()
		if err != nil {
			t.log.Warn(err)
		} else {
			t.tryCreateWorkFromFiles(ctx, files)
		}

		t.log.Debugf("list of recivering: %v", t.getFileList())

		select {
		case <-datakit.Exit.Wait():
			return
		case <-t.done:
			return

		case event, ok := <-t.fileInotify.Events():
			if !(ok || event.Has(fsnotify.Create)) {
				continue
			}
			if stat, err := os.Stat(event.Name); err != nil || stat.IsDir() {
				continue
			}
			t.tryCreateWorkFromFiles(ctx, []string{event.Name})

		case <-ticker.C:
			// next
		}
	}
}

func (t *Tailer) tryCreateWorkFromFiles(ctx context.Context, files []string) {
	files = t.fileFilter.IncludeFilterFiles(files)
	files = t.fileFilter.ExcludeFilterFiles(files)

	for _, file := range files {
		if t.ignoreDeadLog > 0 && !openfile.FileIsActive(file, t.ignoreDeadLog) {
			continue
		}

		if t.inFileList(file) {
			continue
		}

		t.log.Infof("new logging file %s with source %s", file, t.source)

		single, err := NewTailerSingle(file, t.options...)
		if err != nil {
			t.log.Warnf("new tailer file %s error: %s", file, err)
			continue
		}

		ctx, cancel := context.WithCancel(ctx)
		t.addToFileList(file, cancel)

		func(file string) {
			g.Go(func(_ context.Context) error {
				single.Run(ctx)
				t.removeFromFileList(file)
				t.log.Infof("file %s exit", file)
				return nil
			})
		}(file)
	}
}

func (t *Tailer) Close() {
	select {
	case <-t.done:
		return
	default:
		close(t.done)
	}
}

func (t *Tailer) addToFileList(file string, cancel context.CancelFunc) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.fileList[file] = cancel
}

func (t *Tailer) removeFromFileList(file string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.fileList, file)
}

func (t *Tailer) inFileList(file string) bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	_, ok := t.fileList[file]
	return ok
}

func (t *Tailer) closeAllFiles() {
	t.mu.Lock()
	defer t.mu.Unlock()

	for _, cancel := range t.fileList {
		if cancel != nil {
			cancel()
		}
	}
}

func (t *Tailer) getFileList() []string {
	t.mu.Lock()
	defer t.mu.Unlock()
	var list []string
	for file := range t.fileList {
		list = append(list, file)
	}
	return list
}
