// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package flameshot

import (
	"container/list"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMonitorBuildProcessExitEvent(t *testing.T) {
	profilingPath := t.TempDir()
	hprofDir := filepath.Join(profilingPath, "dumps")
	assert.NoError(t, os.MkdirAll(hprofDir, 0o755))

	detectedAt := time.Now()
	jfrPath := filepath.Join(profilingPath, "profiler_alloc_1234.jfr")
	hprofPath := filepath.Join(hprofDir, "app.hprof")

	assert.NoError(t, os.WriteFile(jfrPath, []byte("jfr"), 0o644))
	assert.NoError(t, os.WriteFile(hprofPath, []byte("hprof"), 0o644))
	assert.NoError(t, os.Chtimes(jfrPath, detectedAt.Add(-30*time.Second), detectedAt.Add(-30*time.Second)))
	assert.NoError(t, os.Chtimes(hprofPath, detectedAt.Add(-20*time.Second), detectedAt.Add(-20*time.Second)))

	pm := &processM{
		Name:          "java",
		Cmdline:       "java -XX:+HeapDumpOnOutOfMemoryError -XX:HeapDumpPath=" + hprofDir + " -jar app.jar",
		Pid:           1234,
		SampleMaxSize: 10,
		SampleHistory: list.New(),
		configProcess: &Process{
			Service: "svc-exit",
			Tags:    []string{"env:test"},
		},
	}

	pm.addResourceSample(resourceSample{
		Timestamp:  detectedAt.Add(-40 * time.Second),
		CPUPercent: 10,
		MemUsageMB: 256,
		MemPercent: 50,
	})
	pm.addResourceSample(resourceSample{
		Timestamp:  detectedAt.Add(-20 * time.Second),
		CPUPercent: 15,
		MemUsageMB: 512,
		MemPercent: 80,
	})
	pm.markProfileTriggered(detectedAt.Add(-15 * time.Second))
	pm.markMemoryPressure(detectedAt.Add(-10 * time.Second))
	pm.markProfileArtifact(&profileArtifactSummary{
		OutputPath:  jfrPath,
		UploadedAt:  detectedAt.Add(-5 * time.Second),
		StartTime:   detectedAt.Add(-15 * time.Second).Format(time.RFC3339),
		EndTime:     detectedAt.Add(-5 * time.Second).Format(time.RFC3339),
		Event:       "alloc",
		DurationSec: 10,
	})

	m := &monitor{
		config: &Config{
			Tags:          []string{"cluster:test"},
			ProfilingPath: profilingPath,
		},
	}

	event := m.buildProcessExitEvent(pm, detectedAt, errors.New("process exited"))
	if assert.NotNil(t, event) {
		assert.Equal(t, "svc-exit", event.Service)
		assert.Equal(t, int32(1234), event.PID)
		assert.Equal(t, "process_exit_with_heap_dump", event.SuspectedReason)
		assert.Len(t, event.RecentSamples, 2)
		assert.InDelta(t, 80.0, event.RecentMemPeakPercent, 0.0001)
		assert.InDelta(t, 512.0, event.RecentMemPeakMB, 0.0001)
		if assert.NotNil(t, event.RecentJFRArtifact) {
			assert.Equal(t, jfrPath, event.RecentJFRArtifact.Path)
		}
		if assert.NotNil(t, event.RecentHProfArtifact) {
			assert.Equal(t, hprofPath, event.RecentHProfArtifact.Path)
		}
		if assert.NotNil(t, event.LastProfileArtifact) {
			assert.Equal(t, jfrPath, event.LastProfileArtifact.OutputPath)
		}
	}
}

func TestMonitorHandleProcessGone(t *testing.T) {
	canceled := false
	pm := &processM{
		Name:          "java",
		Pid:           4321,
		SampleMaxSize: 10,
		SampleHistory: list.New(),
		configProcess: &Process{
			Service: "svc-gone",
		},
	}

	watcher := newCgroupWatcher("cg-test", cgroupVersionV2, "/tmp/mock", func() {
		canceled = true
	})
	watcher.addMember(pm)

	m := NewMonitor(&Config{ProfilingPath: t.TempDir()})
	m.exitChan = make(chan *processExitEvent, 1)
	m.cs = []*processM{pm}
	m.watchers[watcher.key] = watcher
	m.watcherKeyByPID[pm.Pid] = watcher.key

	m.handleProcessGone(pm, errors.New("process disappeared"))

	assert.Empty(t, m.cs)
	assert.Empty(t, m.watchers)
	assert.Empty(t, m.watcherKeyByPID)
	assert.True(t, canceled)

	select {
	case event := <-m.exitChan:
		assert.Equal(t, pm.Pid, event.PID)
		assert.Equal(t, "svc-gone", event.Service)
		assert.Equal(t, "process_exit", event.SuspectedReason)
	default:
		t.Fatal("expected process exit event to be emitted")
	}
}

func TestMonitorHandleProcessGoneDoesNotBlockWhenExitQueueFull(t *testing.T) {
	pm := &processM{
		Name:          "java",
		Pid:           4322,
		SampleMaxSize: 10,
		SampleHistory: list.New(),
		configProcess: &Process{
			Service: "svc-gone",
		},
	}

	m := NewMonitor(&Config{ProfilingPath: t.TempDir()})
	m.exitChan = make(chan *processExitEvent, 1)
	m.exitChan <- &processExitEvent{PID: 1}
	m.cs = []*processM{pm}

	done := make(chan struct{})
	go func() {
		m.handleProcessGone(pm, errors.New("process disappeared"))
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("handleProcessGone blocked on a full exit event queue")
	}

	assert.Empty(t, m.cs)
}

func TestMonitorHandleProcessExitUploadsRecentHProf(t *testing.T) {
	hprofPath := filepath.Join(t.TempDir(), "app.hprof")
	assert.NoError(t, os.WriteFile(hprofPath, []byte("dump"), 0o644))
	modTime := time.Now().Add(-10 * time.Second)
	assert.NoError(t, os.Chtimes(hprofPath, modTime, modTime))

	var putPath string
	var putBody string
	var logBody string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		switch r.Method {
		case http.MethodPut:
			putPath = r.URL.Path
			putBody = string(body)
		case http.MethodPost:
			logBody = string(body)
		default:
			t.Fatalf("unexpected request method %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	m := NewMonitor(&Config{
		DataKitAddr:                server.URL + "/profiling/v1/input",
		HProfUploadEnabled:         true,
		HProfUploadProvider:        hprofUploadProviderS3,
		HProfUploadEndpoint:        server.URL,
		HProfUploadBucket:          "dump-bucket",
		HProfUploadAccessKeyID:     "ak",
		HProfUploadAccessKeySecret: "sk",
		HProfUploadPathTemplate:    "{service}/{filename}",
		HProfUploadS3PathStyle:     boolPtr(true),
		HProfUploadTimeout:         "5s",
	})
	event := &processExitEvent{
		Service:         "svc-exit",
		PID:             1234,
		ProcessName:     "java",
		DetectedAt:      time.Now(),
		SuspectedReason: "process_exit_with_heap_dump",
		Tags:            []string{"pod_name:pod-a"},
		RecentHProfArtifact: &localArtifact{
			Path:      hprofPath,
			SizeBytes: 4,
			ModTime:   modTime,
		},
	}

	m.handleProcessExit(event)

	assert.Equal(t, "/dump-bucket/svc-exit/app.hprof", putPath)
	assert.Equal(t, "dump", putBody)
	assert.Equal(t, "ok", event.HProfUploadStatus)
	assert.Equal(t, hprofUploadProviderS3, event.HProfUploadProvider)
	assert.Equal(t, "svc-exit/app.hprof", event.HProfObjectKey)
	assert.Contains(t, event.HProfDownloadURL, "/dump-bucket/svc-exit/app.hprof")
	assert.Contains(t, logBody, "flameshot_process_exit")
	assert.Contains(t, logBody, "hprof_upload_status")
	assert.Contains(t, logBody, "hprof_download_url")
	assert.True(t, m.processedHProf.seen(processExitHProfKey(event.RecentHProfArtifact)))
}
