// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package flameshot

import (
	"context"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type jcmdRoundTripFunc func(*http.Request) (*http.Response, error)

func (f jcmdRoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestBuildJcmdOutputPath(t *testing.T) {
	path := buildJcmdOutputPath("/tmp/flameshot", 1234, "gc_class_histogram", time.Date(2026, 4, 15, 10, 0, 0, 0, time.UTC))
	assert.Equal(t, "/tmp/flameshot/jcmd_gc_class_histogram_1234_20260415_100000.txt", path)
}

func TestHandleJcmdSnapshot(t *testing.T) {
	oldTransport := http.DefaultClient.Transport
	oldFindBinary := findJcmdBinary
	oldRun := runJcmdCommand
	t.Cleanup(func() {
		http.DefaultClient.Transport = oldTransport
		findJcmdBinary = oldFindBinary
		runJcmdCommand = oldRun
	})

	findJcmdBinary = func() string { return "/usr/bin/jcmd" }
	runJcmdCommand = func(ctx context.Context, binary string, pid int32, command string) ([]byte, error) {
		return []byte("header line\nbody line\n"), nil
	}

	var posts int
	http.DefaultClient.Transport = jcmdRoundTripFunc(func(req *http.Request) (*http.Response, error) {
		posts++
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader("OK")),
			Header:     make(http.Header),
		}, nil
	})

	dir := t.TempDir()
	m := NewMonitor(&Config{
		DataKitAddr:         "http://datakit.test:9529/profiling/v1/input",
		ProfilingPath:       dir,
		JCmdSnapshotEnabled: true,
	})

	req := &jcmdSnapshotRequest{
		Service:     "svc-a",
		PID:         1234,
		ProcessName: "java",
		DetectedAt:  time.Date(2026, 4, 15, 10, 0, 0, 0, time.UTC),
		MemPercent:  96.5,
		Tags:        []string{"env:test"},
	}

	m.handleJcmdSnapshot(req)

	assert.Equal(t, 2, posts)
	files, err := filepath.Glob(filepath.Join(dir, "jcmd_*.txt"))
	assert.NoError(t, err)
	assert.Len(t, files, 2)
	content, err := os.ReadFile(files[0])
	assert.NoError(t, err)
	assert.Contains(t, string(content), "header line")
}

func TestSummarizeClassHistogram(t *testing.T) {
	output := []byte(`
 num     #instances         #bytes  class name
----------------------------------------------
   1:          1000        3200000  [B
   2:           200         800000  java.lang.String
   3:            50         120000  java.util.HashMap
   4:            10          40000  java.lang.Object
`)
	summary := summarizeJcmdOutput("gc_class_histogram", output)
	assert.Contains(t, summary, "1:          1000        3200000  [B")
	assert.Contains(t, summary, "2:           200         800000  java.lang.String")
}

func TestSummarizeThreadPrint(t *testing.T) {
	output := []byte(`
"main" #1 prio=5 os_prio=0 cpu=10.00ms elapsed=12.00s tid=0x1 nid=0x2 runnable
   java.lang.Thread.State: RUNNABLE

"Reference Handler" #2 daemon prio=10 os_prio=0 cpu=0.20ms elapsed=12.00s tid=0x3 nid=0x4 waiting on condition
   java.lang.Thread.State: WAITING
`)
	summary := summarizeJcmdOutput("thread_print", output)
	assert.Contains(t, summary, "\"main\" #1")
	assert.Contains(t, summary, "java.lang.Thread.State: RUNNABLE")
	assert.Contains(t, summary, "\"Reference Handler\" #2")
}

func TestMonitorGetJcmdTimeout(t *testing.T) {
	assert.Equal(t, defaultJcmdTimeout, (&monitor{}).getJcmdTimeout())
	assert.Equal(t, defaultJcmdTimeout, (&monitor{config: &Config{}}).getJcmdTimeout())
	assert.Equal(t, 20*time.Second, (&monitor{config: &Config{JCmdTimeout: "20s"}}).getJcmdTimeout())
	assert.Equal(t, defaultJcmdTimeout, (&monitor{config: &Config{JCmdTimeout: "bad"}}).getJcmdTimeout())
}
