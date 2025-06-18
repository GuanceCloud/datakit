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
	"path/filepath"
	"runtime"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/GuanceCloud/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/logtail"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/logtail/fileprovider"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/logtail/openfile"
)

const (
	// 定期寻找符合条件的新文件.
	scanIntervalShort = time.Second * 10
	scanIntervalLong  = time.Minute * 1
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

	maxOpenFiles     int
	currentOpenFiles atomic.Int64

	done chan interface{}
	mu   sync.Mutex

	log *logger.Logger
}

func NewTailer(patterns []string, opts ...Option) (*Tailer, error) {
	_ = logtail.InitDefault()

	c := getOption(opts...)
	patterns = cleanPatterns(patterns)

	tailer := &Tailer{
		options:       opts,
		source:        c.source,
		maxOpenFiles:  c.maxOpenFiles,
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
		if err != nil {
			tailer.log.Warnf("failed to new inotify, err: %s, ingored", err)
			tailer.fileInotify = fileprovider.NewNopInotify()
			return tailer, nil
		}
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

	shortTicker := time.NewTicker(scanIntervalShort)
	defer shortTicker.Stop()
	longTicker := time.NewTicker(scanIntervalLong)
	defer longTicker.Stop()

	ctx := context.Background()

	// first scan
	files, err := t.fileScanner.ScanFiles()
	if err != nil {
		t.log.Warn(err)
	} else {
		t.tryCreateWorkFromFiles(ctx, files)
	}

	for {
		select {
		case <-datakit.Exit.Wait():
			return
		case <-t.done:
			return

		case event, ok := <-t.fileInotify.Events():
			if !ok {
				continue
			}
			stat, err := os.Stat(event.Name)
			if err != nil {
				t.log.Warnf("invalid event name: %s", err)
				continue
			}
			if stat.IsDir() {
				receiveCreateEventVec.WithLabelValues(t.source, "directory").Inc()
				continue
			}
			receiveCreateEventVec.WithLabelValues(t.source, "file").Inc()

			file := filepath.Clean(event.Name)
			t.tryCreateWorkFromFiles(ctx, []string{file})
			shortTicker.Reset(scanIntervalShort)

		case <-shortTicker.C:
			t.scanFiles(ctx)
			longTicker.Reset(scanIntervalLong)

		case <-longTicker.C:
			t.scanFiles(ctx)
			shortTicker.Reset(scanIntervalShort)
		}
	}
}

func (t *Tailer) scanFiles(ctx context.Context) {
	files, err := t.fileScanner.ScanFiles()
	if err != nil {
		t.log.Warn(err)
		return
	}
	t.tryCreateWorkFromFiles(ctx, files)
}

func (t *Tailer) tryCreateWorkFromFiles(ctx context.Context, files []string) {
	files = t.fileFilter.IncludeFilterFiles(files)
	files = t.fileFilter.ExcludeFilterFiles(files)

	for _, file := range files {
		if !t.shouldOpenFile(file) {
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

		openfilesVec.WithLabelValues(t.source, strconv.Itoa(t.maxOpenFiles)).Inc()

		func(file string) {
			g.Go(func(_ context.Context) error {
				single.Run(ctx)
				t.removeFromFileList(file)
				t.log.Infof("file %s exit", file)

				openfilesVec.WithLabelValues(t.source, strconv.Itoa(t.maxOpenFiles)).Dec()
				return nil
			})
		}(file)
	}
}

func (t *Tailer) Close() {
	if t.fileInotify != nil {
		if err := t.fileInotify.Close(); err != nil {
			t.log.Warnf("close inotify error: %s", err)
		}
	}
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
	t.currentOpenFiles.Add(1)
}

func (t *Tailer) removeFromFileList(file string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	delete(t.fileList, file)
	t.currentOpenFiles.Add(-1)
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

func (t *Tailer) shouldOpenFile(file string) bool {
	if t.inFileList(file) {
		return false
	}
	if t.maxOpenFiles != -1 && t.currentOpenFiles.Load() >= int64(t.maxOpenFiles) {
		t.log.Warnf("too many open files, limit %d", t.maxOpenFiles)
		return false
	}
	if t.ignoreDeadLog > 0 && !openfile.FileIsActive(file, t.ignoreDeadLog) {
		return false
	}
	return true
}

func cleanPatterns(patterns []string) []string {
	newPatterns := make([]string, len(patterns))
	copy(newPatterns, patterns)
	for i := range newPatterns {
		newPatterns[i] = filepath.Clean(newPatterns[i])
		newPatterns[i] = filepath.ToSlash(newPatterns[i])
	}
	return newPatterns
}
