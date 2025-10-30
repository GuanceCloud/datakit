// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

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
	"github.com/fsnotify/fsnotify"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/logtail"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/logtail/fileprovider"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/logtail/openfile"
)

const (
	defaultScanIntervalShort = time.Second * 10
	defaultScanIntervalLong  = time.Minute * 1
	defaultMaxOpenFiles      = 500
	defaultUpdateChannelSize = 10
)

var globalGoroutineGroup = datakit.G("tailer")

type Tailer struct {
	initialOptions    []Option
	additionalOptions []Option

	source        string
	fromBeginning bool
	ignoreDeadLog time.Duration
	maxOpenFiles  int

	monitoredFiles map[string]*Single
	openFileCount  atomic.Int64
	fileMutex      sync.RWMutex

	fileScanner *fileprovider.Scanner
	fileFilter  *fileprovider.GlobFilter
	fileWatcher fileprovider.InotifyInterface

	shutdownChan chan struct{}
	updateChan   chan []Option

	log *logger.Logger

	// 状态管理
	isRunning atomic.Bool
	startTime time.Time
}

func NewTailer(patterns []string, opts ...Option) (*Tailer, error) {
	_ = logtail.InitDefault()

	cfg := buildConfig(opts)
	if err := validateConfig(cfg); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	patterns = cleanPatterns(patterns)
	if len(patterns) == 0 {
		return nil, fmt.Errorf("no patterns provided")
	}

	tailer := &Tailer{
		initialOptions:    opts,
		additionalOptions: nil,
		source:            cfg.source,
		maxOpenFiles:      cfg.maxOpenFiles,
		fromBeginning:     cfg.fromBeginning,
		ignoreDeadLog:     cfg.ignoreDeadLog,
		monitoredFiles:    make(map[string]*Single),

		shutdownChan: make(chan struct{}),
		log:          logger.SLogger("tailer/" + cfg.source),
		updateChan:   make(chan []Option, defaultUpdateChannelSize),
	}

	if err := tailer.initializeFileProviders(patterns, cfg); err != nil {
		return nil, fmt.Errorf("failed to initialize file providers: %w", err)
	}

	return tailer, nil
}

func (t *Tailer) initializeFileProviders(patterns []string, cfg *config) error {
	var err error

	t.fileScanner, err = fileprovider.NewScanner(patterns)
	if err != nil {
		return fmt.Errorf("failed to create scanner: %w", err)
	}

	t.fileFilter, err = fileprovider.NewGlobFilter(patterns, cfg.ignorePatterns)
	if err != nil {
		return fmt.Errorf("failed to create filter: %w", err)
	}

	if runtime.GOOS == datakit.OSLinux {
		t.fileWatcher, err = fileprovider.NewInotify(patterns)
		if err != nil {
			t.log.Warnf("failed to create inotify: %s, using fallback", err)
			t.fileWatcher = fileprovider.NewNopInotify()
		}
	} else {
		t.fileWatcher = fileprovider.NewNopInotify()
	}

	return nil
}

func (t *Tailer) Start() {
	if !t.isRunning.CompareAndSwap(false, true) {
		t.log.Warn("tailer is already running")
		return
	}

	t.startTime = time.Now()
	t.log.Infof("starting tailer with source: %s, maxOpenFiles: %d", t.source, t.maxOpenFiles)

	defer func() {
		t.isRunning.Store(false)
		t.cleanup()
	}()

	shortTicker := time.NewTicker(defaultScanIntervalShort)
	defer shortTicker.Stop()
	longTicker := time.NewTicker(defaultScanIntervalLong)
	defer longTicker.Stop()

	if err := t.performInitialScan(); err != nil {
		t.log.Errorf("initial scan failed: %v", err)
		return
	}

	t.runEventLoop(shortTicker, longTicker)
}

func (t *Tailer) cleanup() {
	t.closeAllFiles()
	_ = globalGoroutineGroup.Wait()
	t.log.Infof("all tailers exited, source: %s", t.source)
}

func (t *Tailer) performInitialScan() error {
	ctx := context.Background()
	files, err := t.fileScanner.ScanFiles()
	if err != nil {
		return fmt.Errorf("failed to scan files: %w", err)
	}

	t.log.Debugf("initial scan found %d files", len(files))
	t.processFiles(ctx, files)
	return nil
}

func (t *Tailer) runEventLoop(shortTicker, longTicker *time.Ticker) {
	ctx := context.Background()

	for {
		select {
		case <-datakit.Exit.Wait():
			t.log.Info("received exit signal, stopping tailer")
			return
		case <-t.shutdownChan:
			t.log.Info("received shutdown signal, stopping tailer")
			return
		case newOpts := <-t.updateChan:
			t.handleConfigUpdate(newOpts)
		case event, ok := <-t.fileWatcher.Events():
			t.handleFileEvent(ctx, event, ok, shortTicker)
		case <-shortTicker.C:
			t.handleShortIntervalScan(ctx, longTicker)
		case <-longTicker.C:
			t.handleLongIntervalScan(ctx, shortTicker)
		}
	}
}

func (t *Tailer) handleConfigUpdate(newOpts []Option) {
	t.log.Info("received options update")
	t.updateAllSingles(newOpts)
}

func (t *Tailer) handleFileEvent(ctx context.Context, event fsnotify.Event, ok bool, shortTicker *time.Ticker) {
	if !ok {
		t.log.Warn("file watcher events channel closed")
		return
	}

	stat, err := os.Stat(event.Name)
	if err != nil {
		t.log.Warnf("invalid event name: %v", err)
		return
	}

	if stat.IsDir() {
		t.log.Debugf("ignoring directory event: %s", event.Name)
		createEventCounter.WithLabelValues(t.source, "directory").Inc()
		return
	}

	createEventCounter.WithLabelValues(t.source, "file").Inc()
	file := filepath.Clean(event.Name)
	t.log.Debugf("processing file event: %s", file)

	t.processFiles(ctx, []string{file})
	shortTicker.Reset(defaultScanIntervalShort)
}

func (t *Tailer) handleShortIntervalScan(ctx context.Context, longTicker *time.Ticker) {
	t.scanFiles(ctx)
	longTicker.Reset(defaultScanIntervalLong)
}

func (t *Tailer) handleLongIntervalScan(ctx context.Context, shortTicker *time.Ticker) {
	t.scanFiles(ctx)
	shortTicker.Reset(defaultScanIntervalShort)
}

func (t *Tailer) UpdateOptions(newOpts []Option) error {
	cfg := buildConfig(newOpts)
	if err := checkConfig(cfg); err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}

	select {
	case t.updateChan <- newOpts:
		t.log.Debug("options update sent to channel")
	default:
		t.log.Warn("update channel full, dropping options update")
	}

	return nil
}

func (t *Tailer) updateAllSingles(newOpts []Option) {
	t.log.Infof("updating options for %d single tailers", len(t.monitoredFiles))
	for _, single := range t.monitoredFiles {
		single.UpdateOptions(newOpts)
	}
	t.additionalOptions = newOpts // 保存补充的配置选项
	t.log.Debug("all single tailers options updated")
}

func (t *Tailer) scanFiles(ctx context.Context) {
	files, err := t.fileScanner.ScanFiles()
	if err != nil {
		t.log.Warnf("scan files failed: %v", err)
		return
	}
	t.log.Debugf("scan found %d files", len(files))
	t.processFiles(ctx, files)
}

func (t *Tailer) processFiles(ctx context.Context, files []string) {
	files = t.fileFilter.IncludeFilterFiles(files)
	files = t.fileFilter.ExcludeFilterFiles(files)

	t.log.Debugf("processing %d filtered files", len(files))

	for _, file := range files {
		if !t.shouldOpenFile(file) {
			continue
		}

		t.log.Infof("creating new tailer for file %s with source %s", file, t.source)
		t.createFileTailer(ctx, file)
	}
}

func (t *Tailer) createFileTailer(ctx context.Context, file string) {
	// 合并初始配置和补充配置
	allOptions := make([]Option, 0, len(t.initialOptions)+len(t.additionalOptions))
	allOptions = append(allOptions, t.initialOptions...)
	allOptions = append(allOptions, t.additionalOptions...)
	single, err := NewTailerSingle(file, allOptions...)
	if err != nil {
		t.log.Warnf("failed to create tailer for file %s: %v", file, err)
		return
	}

	t.addToMonitoredFiles(file, single)
	openFilesGauge.WithLabelValues(t.source, strconv.Itoa(t.maxOpenFiles)).Inc()

	// 启动文件采集器协程
	globalGoroutineGroup.Go(func(_ context.Context) error {
		single.Run(ctx)
		t.removeFromMonitoredFiles(file)
		t.log.Infof("file %s tailer exited", file)
		openFilesGauge.WithLabelValues(t.source, strconv.Itoa(t.maxOpenFiles)).Dec()
		return nil
	})
}

func (t *Tailer) Close() {
	t.log.Info("closing tailer, source: %s", t.source)

	if t.fileWatcher != nil {
		if err := t.fileWatcher.Close(); err != nil {
			t.log.Warnf("close file watcher error: %v", err)
		}
	}

	select {
	case <-t.shutdownChan:
		t.log.Debug("tailer already closed")
		return
	default:
		close(t.shutdownChan)
		t.log.Debug("tailer close signal sent")
	}
}

func (t *Tailer) addToMonitoredFiles(file string, single *Single) {
	t.fileMutex.Lock()
	defer t.fileMutex.Unlock()

	if _, exists := t.monitoredFiles[file]; exists {
		t.log.Warnf("file %s already in monitored files, skipping", file)
		return
	}

	t.monitoredFiles[file] = single
	t.openFileCount.Add(1)
	t.log.Debugf("added file %s to monitored files, current open files: %d", file, t.openFileCount.Load())
}

func (t *Tailer) removeFromMonitoredFiles(file string) {
	t.fileMutex.Lock()
	defer t.fileMutex.Unlock()

	if _, exists := t.monitoredFiles[file]; !exists {
		t.log.Warnf("file %s not in monitored files, skipping removal", file)
		return
	}

	delete(t.monitoredFiles, file)
	t.openFileCount.Add(-1)
	t.log.Debugf("removed file %s from monitored files, current open files: %d", file, t.openFileCount.Load())
}

func (t *Tailer) isFileMonitored(file string) bool {
	t.fileMutex.RLock()
	defer t.fileMutex.RUnlock()
	_, ok := t.monitoredFiles[file]
	return ok
}

func (t *Tailer) closeAllFiles() {
	t.fileMutex.Lock()
	defer t.fileMutex.Unlock()

	t.log.Infof("closing %d single tailers", len(t.monitoredFiles))
	for file, single := range t.monitoredFiles {
		if single != nil {
			single.Close()
			t.log.Debugf("closed tailer for file %s", file)
		}
	}
	t.monitoredFiles = make(map[string]*Single)
}

func (t *Tailer) shouldOpenFile(file string) bool {
	if t.isFileMonitored(file) {
		return false
	}
	if t.maxOpenFiles != -1 && t.openFileCount.Load() >= int64(t.maxOpenFiles) {
		t.log.Warnf("too many open files, limit %d, current %d", t.maxOpenFiles, t.openFileCount.Load())
		return false
	}
	if t.ignoreDeadLog > 0 && !openfile.FileIsActive(file, t.ignoreDeadLog) {
		t.log.Debugf("file %s is not active, skipping", file)
		return false
	}
	return true
}

func validateConfig(cfg *config) error {
	if cfg.source == "" {
		return fmt.Errorf("source cannot be empty")
	}
	if cfg.maxOpenFiles < -1 {
		return fmt.Errorf("maxOpenFiles must be >= -1, got %d", cfg.maxOpenFiles)
	}
	if cfg.ignoreDeadLog < 0 {
		return fmt.Errorf("ignoreDeadLog must be >= 0, got %v", cfg.ignoreDeadLog)
	}
	return nil
}

func cleanPatterns(patterns []string) []string {
	newPatterns := make([]string, 0, len(patterns))
	for _, pattern := range patterns {
		if pattern == "" {
			continue
		}
		cleaned := filepath.Clean(pattern)
		cleaned = filepath.ToSlash(cleaned)
		newPatterns = append(newPatterns, cleaned)
	}
	return newPatterns
}
