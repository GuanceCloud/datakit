// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package flameshot

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRenderHeapDumpPath(t *testing.T) {
	cfg := &Config{
		ProfilingPath:        "/flameshot-data",
		HeapDumpPathTemplate: "{profiling_path}/dumps/{service}_{pod_name}_{pid}_{timestamp}.hprof",
	}
	pm := &processM{
		Pid: 1234,
		configProcess: &Process{
			Service: "svc-a",
		},
	}

	got := renderHeapDumpPath(cfg, pm, []string{"pod_name:pod-a"}, time.Date(2026, 6, 25, 7, 11, 22, 0, time.UTC))
	assert.Equal(t, "/flameshot-data/dumps/svc-a_pod-a_1234_20260625T071122Z.hprof", got)
}

func TestHandleHeapDumpTaskRunsCommandAndReportsSummary(t *testing.T) {
	oldRun := runHeapDumpCommand
	t.Cleanup(func() {
		runHeapDumpCommand = oldRun
	})

	runHeapDumpCommand = func(ctx context.Context, jmapPath string, outputPath string, pid int32) (string, string, error) {
		require.Equal(t, "jmap", jmapPath)
		require.Equal(t, int32(1234), pid)
		require.NoError(t, os.WriteFile(outputPath, []byte("dump"), 0o644))
		return "ok", "", nil
	}

	var gotBody string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		gotBody = string(body)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	dir := t.TempDir()
	cfg := &Config{
		DataKitAddr:          server.URL + "/profiling/v1/input",
		ProfilingPath:        dir,
		HeapDumpEnabled:      true,
		HeapDumpJMapPath:     "jmap",
		HeapDumpTimeout:      "5s",
		HeapDumpPathTemplate: "{profiling_path}/dumps/{service}_{pid}.hprof",
	}
	m := NewMonitor(cfg)
	pm := &processM{
		Name: "java",
		Pid:  1234,
		configProcess: &Process{
			Service:  "svc-a",
			Language: "java",
		},
	}

	m.handleHeapDumpTask(&heapDumpTask{
		pm:         pm,
		trigger:    "process_memory_emergency",
		tags:       []string{"env:test", "service:svc-a"},
		detectedAt: time.Now(),
	})

	assert.FileExists(t, filepath.Join(dir, "dumps", "svc-a_1234.hprof"))
	assert.Contains(t, gotBody, "flameshot_oom_hprof")
	assert.Contains(t, gotBody, "heap dump generated")
}

func TestCgroupMemoryPressureProfilingDisabledStillEnqueuesHeapDump(t *testing.T) {
	profilingEnabled := false
	m := NewMonitor(&Config{
		ProfilingEnabled: &profilingEnabled,
		HeapDumpEnabled:  true,
		Tags:             []string{"env:test"},
	})
	m.statsChan = make(chan *triggerStats, 1)
	m.heapDumpChan = make(chan *heapDumpTask, 1)

	pm := &processM{
		Name: "java",
		Pid:  1234,
		configProcess: &Process{
			Service:                  "svc-a",
			Language:                 "java",
			Events:                   "cpu",
			EmergencyDuration:        "10s",
			MEMUsagePercentEmergency: 90,
		},
	}
	watcher := newCgroupWatcher("cg-heap-dump", cgroupVersionV2, "/sys/fs/cgroup/mock", func() {})
	watcher.addMember(pm)

	m.handleCgroupMemoryStats(watcher, &cgroupMemoryStats{
		Current: 95,
		Max:     100,
		OOMKill: 0,
	}, 0, time.Now())

	select {
	case task := <-m.heapDumpChan:
		assert.Equal(t, int32(1234), task.pm.Pid)
		assert.True(t, strings.Contains(strings.Join(task.tags, ","), "trigger:cgroup_memory_pressure"))
	default:
		t.Fatal("expected cgroup memory pressure to enqueue heap dump")
	}
	select {
	case <-m.statsChan:
		t.Fatal("expected profiling to be disabled")
	default:
	}
}
