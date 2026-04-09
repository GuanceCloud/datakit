// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build !windows

package ddtrace

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDDTracesResetKeepInPool(t *testing.T) {
	span := &DDSpan{
		Service: "svc",
		Name:    "op",
		Meta: map[string]string{
			"k": "v",
		},
		Metrics: map[string]float64{
			"m": 1,
		},
	}
	traces := DDTraces{DDTrace{span}}

	require.True(t, traces.shouldKeepInPool())

	traceBacking := traces[0]

	traces.reset(true)

	assert.Len(t, traces, 0)
	require.NotNil(t, traceBacking[0])
	assert.Empty(t, traceBacking[0].Service)
	assert.Empty(t, traceBacking[0].Name)
	assert.Empty(t, traceBacking[0].Meta)
	assert.Empty(t, traceBacking[0].Metrics)
}

func TestDDTracesResetDropOversized(t *testing.T) {
	traces := make(DDTraces, maxPooledTraceCount+1)
	for i := range traces {
		traces[i] = DDTrace{
			&DDSpan{
				Service: "svc",
				Meta: map[string]string{
					"k": "v",
				},
			},
		}
	}

	require.False(t, traces.shouldKeepInPool())

	traceBacking := traces[0]
	span := traceBacking[0]

	traces.reset(false)

	assert.Nil(t, traces)
	assert.Nil(t, traceBacking[0])
	assert.Equal(t, DDSpan{}, *span)
}

func TestHandleDDTracesRejectOversizedBodyOnRead(t *testing.T) {
	ipt := defaultInput()
	ipt.maxTraceBody = 8

	req := httptest.NewRequest(http.MethodPost, v4, strings.NewReader("0123456789"))
	req.Header.Set("Content-Type", "application/msgpack")
	req.Header.Set("X-Datadog-Trace-Count", "1")

	resp := httptest.NewRecorder()
	ipt.handleDDTraces(resp, req)

	assert.Equal(t, http.StatusRequestEntityTooLarge, resp.Code)
}

func TestHandleDDTracesRejectOversizedBodyByHeader(t *testing.T) {
	ipt := defaultInput()
	ipt.maxTraceBody = 8

	req := httptest.NewRequest(http.MethodPost, v4, strings.NewReader("0123456789"))
	req.Header.Set("Content-Type", "application/msgpack")
	req.Header.Set("X-Datadog-Trace-Count", "1")
	req.Header.Set("Content-Length", "10")

	resp := httptest.NewRecorder()
	ipt.handleDDTraces(resp, req)

	assert.Equal(t, http.StatusRequestEntityTooLarge, resp.Code)
}
