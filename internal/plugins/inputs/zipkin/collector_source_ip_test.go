// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package zipkin

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	itrace "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/trace"
)

func TestCollectorSourceIP_ZipkinV1JSON(t *testing.T) {
	var capturedTraces itrace.DatakitTraces
	afterGatherRun = itrace.AfterGatherFunc(func(inputName string, dktraces itrace.DatakitTraces) {
		capturedTraces = dktraces
	})

	// 创建测试 span 数据
	spans := []*ZipkinSpanV1{
		{
			TraceID:  "abc123",
			ID:       "def456",
			ParentID: "",
			Name:     "test-span",
			Annotations: []*Annotation{
				{
					Timestamp: 1609459200000000, // 2021-01-01 00:00:00 UTC in microseconds
					Value:     "cs",
					Host: &Endpoint{
						ServiceName: "test-service",
						Ipv4:        "127.0.0.1",
						Port:        8080,
					},
				},
			},
			Duration: 1000,
		},
	}

	body, err := json.Marshal(spans)
	require.NoError(t, err)

	// 创建带有 X-Forwarded-For 头的请求（需要包含端口）
	req := httptest.NewRequest("POST", "/api/v1/spans", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Forwarded-For", "192.168.1.100:8080")

	resp := httptest.NewRecorder()
	handleZipkinTraceV1(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)
	require.NotEmpty(t, capturedTraces)
	require.NotEmpty(t, capturedTraces[0])

	// 验证 collector_source_ip tag
	pt := capturedTraces[0][0].Point
	collectorIP := pt.GetTag(itrace.TagCollectorSourceIP)
	assert.Equal(t, "192.168.1.100", collectorIP, "collector_source_ip should be set from X-Forwarded-For header")
}

func TestCollectorSourceIP_ZipkinV2JSON(t *testing.T) {
	var capturedTraces itrace.DatakitTraces
	afterGatherRun = itrace.AfterGatherFunc(func(inputName string, dktraces itrace.DatakitTraces) {
		capturedTraces = dktraces
	})

	// 创建测试 span 数据 (Zipkin V2 JSON 格式)
	spansJSON := `[{
		"traceId": "abc123def456789",
		"id": "def456789abc123",
		"name": "test-span-v2",
		"timestamp": 1609459200000000,
		"duration": 1000,
		"localEndpoint": {
			"serviceName": "test-service-v2"
		}
	}]`

	req := httptest.NewRequest("POST", "/api/v2/spans", bytes.NewReader([]byte(spansJSON)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Real-Ip", "10.0.0.50:9090")

	resp := httptest.NewRecorder()
	handleZipkinTraceV2(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)
	require.NotEmpty(t, capturedTraces)
	require.NotEmpty(t, capturedTraces[0])

	// 验证 collector_source_ip tag
	pt := capturedTraces[0][0].Point
	collectorIP := pt.GetTag(itrace.TagCollectorSourceIP)
	assert.Equal(t, "10.0.0.50", collectorIP, "collector_source_ip should be set from X-Real-Ip header (port stripped)")
}

func TestCollectorSourceIP_ZipkinRemoteAddr(t *testing.T) {
	var capturedTraces itrace.DatakitTraces
	afterGatherRun = itrace.AfterGatherFunc(func(inputName string, dktraces itrace.DatakitTraces) {
		capturedTraces = dktraces
	})

	spans := []*ZipkinSpanV1{
		{
			TraceID:  "abc123",
			ID:       "def456",
			ParentID: "",
			Name:     "test-span",
			Annotations: []*Annotation{
				{
					Timestamp: 1609459200000000,
					Value:     "cs",
					Host: &Endpoint{
						ServiceName: "test-service",
						Ipv4:        "127.0.0.1",
						Port:        8080,
					},
				},
			},
			Duration: 1000,
		},
	}

	body, err := json.Marshal(spans)
	require.NoError(t, err)

	// 创建没有代理头的请求，使用 RemoteAddr
	req := httptest.NewRequest("POST", "/api/v1/spans", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.RemoteAddr = "172.16.0.1:12345"

	resp := httptest.NewRecorder()
	handleZipkinTraceV1(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)
	require.NotEmpty(t, capturedTraces)
	require.NotEmpty(t, capturedTraces[0])

	// 验证 collector_source_ip tag 来自 RemoteAddr
	pt := capturedTraces[0][0].Point
	collectorIP := pt.GetTag(itrace.TagCollectorSourceIP)
	assert.Equal(t, "172.16.0.1", collectorIP, "collector_source_ip should be set from RemoteAddr")
}
