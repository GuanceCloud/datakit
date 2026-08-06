// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package flameshot

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAddJSONConfigDeduplicatesProfilerTags(t *testing.T) {
	stats := &triggerStats{
		PID:         14,
		CommandName: "java",
		Reason: []string{
			"host:test-host",
			"env:uat",
			"version:1.0.0",
			"service:flameshot-test-app",
			"env:process-env",
			"version:process-version",
			"service:duplicated-service",
			"trigger:cpu",
			"cpu_avg:70.23",
			"pid:14",
		},
		startTime: "2026-05-07T02:36:59Z",
		endTime:   "2026-05-07T02:37:14Z",
	}

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	require.NoError(t, addJSONConfig(writer, stats))
	require.NoError(t, writer.Close())

	reader := multipart.NewReader(&buf, writer.Boundary())
	part, err := reader.NextPart()
	require.NoError(t, err)
	defer part.Close() //nolint:errcheck

	body, err := io.ReadAll(part)
	require.NoError(t, err)

	var event Event
	require.NoError(t, json.Unmarshal(body, &event))

	tags := strings.Split(event.TagProfiler, ",")
	assert.Equal(t, 1, countProfilerTagKey(tags, "host"))
	assert.Equal(t, 1, countProfilerTagKey(tags, "env"))
	assert.Equal(t, 1, countProfilerTagKey(tags, "version"))
	assert.Equal(t, 1, countProfilerTagKey(tags, "service"))
	assert.Contains(t, tags, "trigger:cpu")
	assert.Contains(t, tags, "cpu_avg:70.23")
	assert.Contains(t, tags, "pid:14")
}

func TestUploadPythonPySpyToDataKitMultipart(t *testing.T) {
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
		Language:    "python",
		PID:         123,
		CommandName: "python",
		Service:     "py-svc",
		Reason:      []string{"service:py-svc", "host:test-host", "env:test", "version:v1"},
		startTime:   "2026-07-27T06:00:00Z",
		endTime:     "2026-07-27T06:00:10Z",
		Attachments: []*profileAttachment{
			{FieldName: "prof", FileName: "prof", Data: []byte("process 123:\"python app.py\";thread (1);main (/app/app.py:1) 1\n")},
		},
	}

	require.NoError(t, uploadFileToDataKit(stats, srv.URL))
	assert.Contains(t, gotFiles, "prof")
	assert.Contains(t, gotFiles, "event.json")
	assert.Equal(t, "python", gotEvent.Family)
	assert.Equal(t, "collapse", gotEvent.Format)
	assert.Equal(t, "pyspy", gotEvent.Profiler)
	assert.ElementsMatch(t, []string{"prof"}, gotEvent.Attachments)
	assert.Contains(t, gotEvent.TagProfiler, "language:python")
	assert.Contains(t, gotEvent.TagProfiler, "profiler:pyspy")
	assert.Contains(t, gotEvent.TagProfiler, "service:py-svc")
}

func countProfilerTagKey(tags []string, key string) int {
	count := 0
	for _, tag := range tags {
		tagKey, _, ok := strings.Cut(tag, ":")
		if ok && tagKey == key {
			count++
		}
	}
	return count
}
