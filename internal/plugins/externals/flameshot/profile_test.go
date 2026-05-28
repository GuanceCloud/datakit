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
