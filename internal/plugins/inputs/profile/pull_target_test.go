// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package profile

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGoProfilerTargetPathAndCPUDuration(t *testing.T) {
	var requestPath, seconds string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		requestPath = req.URL.Path
		seconds = req.URL.Query().Get("seconds")
		_, err := w.Write([]byte("profile"))
		assert.NoError(t, err)
	}))
	defer server.Close()

	profiler := &GoProfiler{
		URL:             server.URL + "/custom/pprof",
		ProfileDuration: "23s",
		HTTPTimeout:     "30s",
		input:           DefaultInput(),
	}
	require.NoError(t, profiler.init())
	data, err := profiler.pullProfileItem(context.Background(), "cpu", profileConfigMap["cpu"], profiler.duration)
	require.NoError(t, err)
	require.NotNil(t, data)
	assert.Equal(t, "/custom/pprof/profile", requestPath)
	assert.Equal(t, "23", seconds)
	assert.Equal(t, "profile", data.buf.String())
}

func TestGoProfilerDefaultsPProfPath(t *testing.T) {
	var requestPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		requestPath = req.URL.Path
		_, err := w.Write([]byte("goroutine"))
		assert.NoError(t, err)
	}))
	defer server.Close()

	profiler := &GoProfiler{URL: server.URL, input: DefaultInput()}
	require.NoError(t, profiler.init())
	data, err := profiler.pullProfileItem(context.Background(), "goroutine", profileConfigMap["goroutine"], time.Second)
	require.NoError(t, err)
	require.NotNil(t, data)
	assert.Equal(t, "/debug/pprof/goroutine", requestPath)
	assert.Equal(t, 10*time.Second, profiler.duration)
	assert.Equal(t, 15*time.Second, profiler.timeout)
}

func TestGoProfilerRejectsShortHTTPTimeout(t *testing.T) {
	profiler := &GoProfiler{
		URL:             "http://127.0.0.1:6060",
		ProfileDuration: "30s",
		HTTPTimeout:     "30s",
	}
	err := profiler.init()
	require.EqualError(t, err, "http_timeout must be greater than profile_duration")
}
