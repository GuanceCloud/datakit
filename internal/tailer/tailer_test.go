// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package tailer

import (
	"fmt"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/GuanceCloud/cliutils/logger"
	"github.com/stretchr/testify/require"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/goroutine"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/logtail/fileprovider"
)

func newLoopTestTailer(shortScanEnabled bool) *Tailer {
	return &Tailer{
		shutdownChan:      make(chan struct{}),
		updateChan:        make(chan []Option, 1),
		fileWatcher:       fileprovider.NewNopInotify(),
		shortScanEnabled:  shortScanEnabled,
		log:               logger.SLogger("tailer/test"),
		g:                 goroutine.NewGroup(goroutine.Option{Name: "tailer-test"}),
		monitoredFiles:    map[string]*Single{},
		additionalOptions: nil,
	}
}

func runLoopAndCapturePanic(t *testing.T, tailer *Tailer, shortTick, longTick, runFor time.Duration) (panicked bool) {
	t.Helper()

	panicCh := make(chan interface{}, 1)
	done := make(chan struct{})

	go func() {
		defer close(done)
		defer func() {
			if r := recover(); r != nil {
				panicCh <- r
			}
		}()

		shortTicker := time.NewTicker(shortTick)
		defer shortTicker.Stop()
		longTicker := time.NewTicker(longTick)
		defer longTicker.Stop()

		tailer.runEventLoop(shortTicker, longTicker)
	}()

	time.Sleep(runFor)
	select {
	case <-tailer.shutdownChan:
	default:
		close(tailer.shutdownChan)
	}

	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("event loop did not exit in time")
	}

	select {
	case <-panicCh:
		return true
	default:
		return false
	}
}

func TestRunEventLoopShortScanDisabledSkipsShortTicker(t *testing.T) {
	tailer := newLoopTestTailer(false)
	// fileScanner is intentionally nil: if short scan executes, it panics.
	panicked := runLoopAndCapturePanic(t, tailer, 5*time.Millisecond, 1*time.Second, 80*time.Millisecond)
	require.False(t, panicked)
}

func TestRunEventLoopShortScanEnabledTriggersShortTickerScan(t *testing.T) {
	tailer := newLoopTestTailer(true)
	// fileScanner is intentionally nil: short tick should trigger scan and panic.
	panicked := runLoopAndCapturePanic(t, tailer, 5*time.Millisecond, 1*time.Second, 80*time.Millisecond)
	require.True(t, panicked)
}

func TestInitializeFileProvidersInotifySuccessDisablesShortScan(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("inotify success path is Linux-only")
	}

	tmpDir := t.TempDir()
	pattern := filepath.Join(tmpDir, "app.log")
	tailer := &Tailer{
		log: logger.SLogger("tailer/test"),
	}

	err := tailer.initializeFileProviders([]string{pattern}, defaultConfig())
	require.NoError(t, err)
	require.False(t, tailer.shortScanEnabled)
	require.NotNil(t, tailer.fileWatcher)
	require.NoError(t, tailer.fileWatcher.Close())
}

func TestInitializeFileProvidersInotifyFallbackEnablesShortScan(t *testing.T) {
	tailer := &Tailer{
		log: logger.SLogger("tailer/test"),
	}

	invalidPattern := filepath.Join("/path/does/not/exist", fmt.Sprintf("tailer-%d.log", time.Now().UnixNano()))
	err := tailer.initializeFileProviders([]string{invalidPattern}, defaultConfig())
	require.NoError(t, err)
	require.True(t, tailer.shortScanEnabled)
	require.NotNil(t, tailer.fileWatcher)
	require.NoError(t, tailer.fileWatcher.Close())
}
