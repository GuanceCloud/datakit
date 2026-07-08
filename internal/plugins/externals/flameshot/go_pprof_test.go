// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package flameshot

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	pprofile "github.com/google/pprof/profile"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildGoPProfURL(t *testing.T) {
	base, err := url.Parse("http://127.0.0.1:6060/base/?token=abc")
	require.NoError(t, err)

	got := buildGoPProfURL(base, goPProfItems[goPProfTypeCPU], 17)
	u, err := url.Parse(got)
	require.NoError(t, err)

	assert.Equal(t, "http", u.Scheme)
	assert.Equal(t, "127.0.0.1:6060", u.Host)
	assert.Equal(t, "/base/debug/pprof/profile", u.Path)
	assert.Equal(t, "17", u.Query().Get("seconds"))
	assert.Equal(t, "abc", u.Query().Get("token"))
}

func TestResolveGoPProfTypes(t *testing.T) {
	assert.Equal(t, []string{goPProfTypeCPU}, resolveGoPProfTypes(&triggerStats{}))
	assert.Equal(t, allGoPProfTypes, resolveGoPProfTypes(&triggerStats{PProfTypes: []string{"all"}}))
	assert.Equal(t,
		[]string{goPProfTypeCPU, goPProfTypeGoroutine, goPProfTypeHeap, goPProfTypeMutex, goPProfTypeBlock},
		resolveGoPProfTypes(&triggerStats{PProfTypes: []string{"cpu", "goroutines", "heap", "mutex", "block", "cpu"}}),
	)
}

func TestApplyProcessConfigDefaultsGoPProfToCPU(t *testing.T) {
	m := NewMonitor(&Config{})
	stats := newTriggerStats("", "10s", nil)

	m.applyProcessConfigToStats(stats, &Process{Language: "go"})

	assert.Equal(t, []string{goPProfTypeCPU}, stats.PProfTypes)
}

func TestRunGoPProfCollectsAllTypesWithDelta(t *testing.T) {
	var (
		heapSeq  int64
		mutexSeq int64
		blockSeq int64
	)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/debug/pprof/profile":
			assert.Equal(t, "3", r.URL.Query().Get("seconds"))
			_, _ = w.Write([]byte("cpu-profile"))
		case "/debug/pprof/goroutine":
			assert.Equal(t, "0", r.URL.Query().Get("debug"))
			_, _ = w.Write([]byte("goroutine-profile"))
		case "/debug/pprof/heap":
			heapSeq++
			_, _ = w.Write(testPProfBytes(t, []goPProfValueType{
				{Type: "alloc_objects", Unit: "count"},
				{Type: "alloc_space", Unit: "bytes"},
			}, []int64{heapSeq * 10, heapSeq * 100}, heapSeq))
		case "/debug/pprof/mutex":
			mutexSeq++
			_, _ = w.Write(testPProfBytes(t, []goPProfValueType{
				{Type: "contentions", Unit: "count"},
				{Type: "delay", Unit: "nanoseconds"},
			}, []int64{mutexSeq * 2, mutexSeq * 20}, mutexSeq))
		case "/debug/pprof/block":
			blockSeq++
			_, _ = w.Write(testPProfBytes(t, []goPProfValueType{
				{Type: "contentions", Unit: "count"},
				{Type: "delay", Unit: "nanoseconds"},
			}, []int64{blockSeq * 3, blockSeq * 30}, blockSeq))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	stats := &triggerStats{
		Language:      "go",
		PProfURL:      srv.URL,
		PProfTypes:    allGoPProfTypes,
		Duration:      3,
		Service:       "go-svc",
		PID:           123,
		CommandName:   "go-app",
		goPProfDeltas: newGoPProfDeltaStore(),
	}

	require.NoError(t, runGoPProf(t.Context(), stats))
	assert.ElementsMatch(t, []string{"cpu.pprof", "goroutines.pprof"}, stats.attachmentFileNames())

	require.NoError(t, runGoPProf(t.Context(), stats))
	assert.ElementsMatch(t,
		[]string{"cpu.pprof", "goroutines.pprof", "delta-heap.pprof", "delta-mutex.pprof", "delta-block.pprof"},
		stats.attachmentFileNames(),
	)
}

func TestUploadGoPProfToDataKitMultipart(t *testing.T) {
	var gotEvent Event
	var gotFiles []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.NoError(t, r.ParseMultipartForm(10<<20))
		for _, files := range r.MultipartForm.File {
			for _, file := range files {
				gotFiles = append(gotFiles, file.Filename)
			}
		}

		eventFiles := r.MultipartForm.File["event"]
		require.Len(t, eventFiles, 1)
		eventReader, err := eventFiles[0].Open()
		require.NoError(t, err)
		defer eventReader.Close() //nolint:errcheck
		body, err := io.ReadAll(eventReader)
		require.NoError(t, err)
		require.NoError(t, json.Unmarshal(body, &gotEvent))
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	stats := &triggerStats{
		Language:    "golang",
		PID:         99,
		CommandName: "go-app",
		Service:     "go-svc",
		Reason:      []string{"service:go-svc", "host:test-host", "env:test", "version:v1"},
		startTime:   time.Now().Add(-time.Second).Format(time.RFC3339Nano),
		endTime:     time.Now().Format(time.RFC3339Nano),
		Attachments: []*profileAttachment{
			{FieldName: "cpu.pprof", FileName: "cpu.pprof", Data: []byte("cpu")},
			{FieldName: "goroutines.pprof", FileName: "goroutines.pprof", Data: []byte("goroutine")},
		},
	}

	require.NoError(t, uploadFileToDataKit(stats, srv.URL))
	assert.Contains(t, gotFiles, "cpu.pprof")
	assert.Contains(t, gotFiles, "goroutines.pprof")
	assert.Contains(t, gotFiles, "event.json")
	assert.Equal(t, "go", gotEvent.Family)
	assert.Equal(t, "pprof", gotEvent.Format)
	assert.Equal(t, "pprof", gotEvent.Profiler)
	assert.ElementsMatch(t, []string{"cpu.pprof", "goroutines.pprof"}, gotEvent.Attachments)
	assert.Contains(t, gotEvent.TagProfiler, "language:golang")
	assert.Contains(t, gotEvent.TagProfiler, "service:go-svc")
}

func TestRunProfilingDispatchesGoPProf(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/debug/pprof/profile", r.URL.Path)
		_, _ = w.Write([]byte("cpu-profile"))
	}))
	defer srv.Close()

	stats := &triggerStats{
		Language:    "go",
		PProfURL:    srv.URL,
		PProfTypes:  []string{goPProfTypeCPU},
		Duration:    1,
		Service:     "go-svc",
		PID:         123,
		CommandName: "go-app",
	}

	require.NoError(t, runProfiling(t.Context(), stats))
	assert.Equal(t, []string{"cpu.pprof"}, stats.attachmentFileNames())
}

func testPProfBytes(t *testing.T, sampleTypes []goPProfValueType, values []int64, seq int64) []byte {
	t.Helper()

	prof := &pprofile.Profile{
		TimeNanos:     seq * int64(time.Second),
		DurationNanos: int64(time.Second),
		Sample: []*pprofile.Sample{
			{Value: values},
		},
	}
	for _, typ := range sampleTypes {
		prof.SampleType = append(prof.SampleType, &pprofile.ValueType{
			Type: typ.Type,
			Unit: typ.Unit,
		})
	}

	buf := bytes.NewBuffer(nil)
	require.NoError(t, prof.Write(buf))
	return buf.Bytes()
}
