// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package flameshot

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type oomRoundTripFunc func(*http.Request) (*http.Response, error)

func (f oomRoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestDeriveLoggingURL(t *testing.T) {
	u, err := deriveLoggingURL("http://127.0.0.1:9529/profiling/v1/input")
	assert.NoError(t, err)
	assert.Equal(t, "http://127.0.0.1:9529/v1/write/logging?input=flameshot", u)
}

func TestFindMatchingHProf(t *testing.T) {
	dir := t.TempDir()
	oldFile := filepath.Join(dir, "old.hprof")
	newFile := filepath.Join(dir, "new.hprof")

	assert.NoError(t, os.WriteFile(oldFile, []byte("old"), 0o644))
	assert.NoError(t, os.WriteFile(newFile, []byte("new"), 0o644))

	now := time.Now()
	assert.NoError(t, os.Chtimes(oldFile, now.Add(-10*time.Minute), now.Add(-10*time.Minute)))
	assert.NoError(t, os.Chtimes(newFile, now, now))

	file, err := findMatchingHProf(dir, now, 2*time.Minute)
	assert.NoError(t, err)
	assert.NotNil(t, file)
	assert.Equal(t, newFile, file.Path)
}

func TestFindMatchingHProfFilePath(t *testing.T) {
	filePath := filepath.Join(t.TempDir(), "app.hprof")
	now := time.Now()
	assert.NoError(t, os.WriteFile(filePath, []byte("dump"), 0o644))
	assert.NoError(t, os.Chtimes(filePath, now, now))

	file, err := findMatchingHProf(filePath, now, 2*time.Minute)
	assert.NoError(t, err)
	assert.NotNil(t, file)
	assert.Equal(t, filePath, file.Path)
}

func TestDetectOOMHProfPath(t *testing.T) {
	pm := &processM{
		Cmdline: `java -XX:+HeapDumpOnOutOfMemoryError -XX:HeapDumpPath=/data/dumps/app.hprof -jar app.jar`,
	}
	assert.Equal(t, "/data/dumps/app.hprof", detectOOMHProfPath(pm))
}

func TestDetectOOMHProfPathRequiresHeapDumpOnOOM(t *testing.T) {
	pm := &processM{
		Cmdline: `java -XX:HeapDumpPath=/data/dumps/app.hprof -jar app.jar`,
	}
	assert.Equal(t, "", detectOOMHProfPath(pm))
}

func TestUploadOOMSummaryLog(t *testing.T) {
	oldTransport := http.DefaultClient.Transport
	t.Cleanup(func() {
		http.DefaultClient.Transport = oldTransport
	})

	var gotPath string
	var gotBody string
	http.DefaultClient.Transport = oomRoundTripFunc(func(req *http.Request) (*http.Response, error) {
		gotPath = req.URL.String()
		body, _ := io.ReadAll(req.Body)
		gotBody = string(body)
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader("OK")),
			Header:     make(http.Header),
		}, nil
	})

	err := uploadOOMSummaryLog(&hprofSummary{
		Service:        "svc-a",
		PID:            1234,
		ProcessName:    "java",
		DetectedAt:     time.Now(),
		HProfPath:      "/data/dumps/app.hprof",
		HProfSizeBytes: 1024,
		MatchedBy:      "path+time_window",
		SummaryText:    "OOM detected",
		OOMKillDelta:   1,
	}, "http://datakit.test:9529/profiling/v1/input", []string{"env:test"})
	assert.NoError(t, err)
	assert.Equal(t, "http://datakit.test:9529/v1/write/logging?input=flameshot", gotPath)
	assert.Contains(t, gotBody, "flameshot_oom_hprof")
	assert.Contains(t, gotBody, "OOM detected")
}

func TestHandleOOMEventMarksProcessed(t *testing.T) {
	dir := t.TempDir()
	hprof := filepath.Join(dir, "app.hprof")
	assert.NoError(t, os.WriteFile(hprof, []byte("dump"), 0o644))
	now := time.Now()
	assert.NoError(t, os.Chtimes(hprof, now, now))

	oldTransport := http.DefaultClient.Transport
	t.Cleanup(func() {
		http.DefaultClient.Transport = oldTransport
	})

	calls := 0
	http.DefaultClient.Transport = oomRoundTripFunc(func(req *http.Request) (*http.Response, error) {
		calls++
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader("OK")),
			Header:     make(http.Header),
		}, nil
	})

	m := NewMonitor(&Config{
		DataKitAddr:         "http://datakit.test:9529/profiling/v1/input",
		OOMHProfEnabled:     true,
		OOMHProfMatchWindow: "2m",
	})

	evt := &OOMEvent{
		DetectedAt:   now,
		OOMKillDelta: 1,
		Candidates: []*oomProcessCandidate{
			{
				Service:     "svc-a",
				PID:         1234,
				ProcessName: "java",
				HProfPath:   dir,
				Tags:        []string{"env:test"},
			},
		},
	}

	m.handleOOMEvent(evt)
	m.handleOOMEvent(evt)

	assert.Equal(t, 1, calls)
}

func TestFindOOMHProfMatchChoosesMatchingCandidate(t *testing.T) {
	dirA := t.TempDir()
	dirB := t.TempDir()
	now := time.Now()

	fileA := filepath.Join(dirA, "a.hprof")
	fileB := filepath.Join(dirB, "b.hprof")
	assert.NoError(t, os.WriteFile(fileA, []byte("a"), 0o644))
	assert.NoError(t, os.WriteFile(fileB, []byte("b"), 0o644))
	assert.NoError(t, os.Chtimes(fileA, now.Add(-time.Minute), now.Add(-time.Minute)))
	assert.NoError(t, os.Chtimes(fileB, now, now))

	m := NewMonitor(&Config{OOMHProfMatchWindow: "2m"})
	match, candidate, err := m.findOOMHProfMatch(&OOMEvent{
		DetectedAt:   now,
		OOMKillDelta: 1,
		Candidates: []*oomProcessCandidate{
			{Service: "svc-a", PID: 1, ProcessName: "java-a", HProfPath: dirA},
			{Service: "svc-b", PID: 2, ProcessName: "java-b", HProfPath: dirB},
		},
	})
	assert.NoError(t, err)
	assert.NotNil(t, match)
	assert.NotNil(t, candidate)
	assert.Equal(t, fileB, match.Path)
	assert.Equal(t, "svc-b", candidate.Service)
	assert.Equal(t, int32(2), candidate.PID)
}
