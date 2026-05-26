// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package httpapi

import (
	"bytes"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCollectorSourceIP_APIWriteLogging(t *testing.T) {
	// 创建测试请求 - 发送行协议格式的日志数据
	body := []byte("test_log,host=myhost message=\"test log message\" 1609459200000000000")

	req := httptest.NewRequest("POST", "/v1/write/logging", bytes.NewReader(body))
	req.Header.Set("Content-Type", "text/plain")
	req.Header.Set("X-Forwarded-For", "192.168.1.100:8080")

	wr := GetAPIWriteResult()
	defer PutAPIWriteResult(wr)

	err := wr.APIV1Write(req)
	require.NoError(t, err, "APIV1Write should not return error")
	require.NotEmpty(t, wr.Points, "should have points")

	// 验证 collector_source_ip tag
	pt := wr.Points[0]
	collectorIP := pt.GetTag("collector_source_ip")
	assert.Equal(t, "192.168.1.100", collectorIP, "collector_source_ip should be set from X-Forwarded-For")
}

func TestCollectorSourceIP_APIWriteLoggingWithXRealIP(t *testing.T) {
	body := []byte("test_log,host=myhost message=\"test log message\" 1609459200000000000")

	req := httptest.NewRequest("POST", "/v1/write/logging", bytes.NewReader(body))
	req.Header.Set("Content-Type", "text/plain")
	req.Header.Set("X-Real-Ip", "10.20.30.40:9090")

	wr := GetAPIWriteResult()
	defer PutAPIWriteResult(wr)

	err := wr.APIV1Write(req)
	require.NoError(t, err, "APIV1Write should not return error")
	require.NotEmpty(t, wr.Points, "should have points")

	// 验证 collector_source_ip tag
	pt := wr.Points[0]
	collectorIP := pt.GetTag("collector_source_ip")
	assert.Equal(t, "10.20.30.40", collectorIP, "collector_source_ip should be set from X-Real-Ip")
}

func TestCollectorSourceIP_APIWriteLoggingRemoteAddr(t *testing.T) {
	body := []byte("test_log,host=myhost message=\"test log message\" 1609459200000000000")

	req := httptest.NewRequest("POST", "/v1/write/logging", bytes.NewReader(body))
	req.Header.Set("Content-Type", "text/plain")
	req.RemoteAddr = "172.16.0.50:54321"

	wr := GetAPIWriteResult()
	defer PutAPIWriteResult(wr)

	err := wr.APIV1Write(req)
	require.NoError(t, err, "APIV1Write should not return error")
	require.NotEmpty(t, wr.Points, "should have points")

	// 验证 collector_source_ip tag 来自 RemoteAddr
	pt := wr.Points[0]
	collectorIP := pt.GetTag("collector_source_ip")
	assert.Equal(t, "172.16.0.50", collectorIP, "collector_source_ip should be set from RemoteAddr")
}

func TestCollectorSourceIP_APIWriteMetricNoTag(t *testing.T) {
	// Metric 类型不应该添加 collector_source_ip
	body := []byte("test_metric,host=myhost value=100 1609459200000000000")

	req := httptest.NewRequest("POST", "/v1/write/metric", bytes.NewReader(body))
	req.Header.Set("Content-Type", "text/plain")
	req.Header.Set("X-Forwarded-For", "192.168.1.100:8080")

	wr := GetAPIWriteResult()
	defer PutAPIWriteResult(wr)

	err := wr.APIV1Write(req)
	require.NoError(t, err, "APIV1Write should not return error")
	require.NotEmpty(t, wr.Points, "should have points")

	// 验证 Metric 类型不包含 collector_source_ip tag
	pt := wr.Points[0]
	collectorIP := pt.GetTag("collector_source_ip")
	assert.Equal(t, "", collectorIP, "Metric should not have collector_source_ip tag")
}

func TestCollectorSourceIP_ProxyHeaders(t *testing.T) {
	testCases := []struct {
		name       string
		headerName string
		headerVal  string
		expectedIP string
	}{
		{
			name:       "X-Forwarded-For single IP with port",
			headerName: "X-Forwarded-For",
			headerVal:  "192.168.1.100:8080",
			expectedIP: "192.168.1.100",
		},
		{
			name:       "X-Forwarded-For multiple IPs with port",
			headerName: "X-Forwarded-For",
			headerVal:  "192.168.1.100:8080, 10.0.0.1:9090, 172.16.0.1:7070",
			expectedIP: "192.168.1.100",
		},
		{
			name:       "X-Real-Ip with port",
			headerName: "X-Real-Ip",
			headerVal:  "10.20.30.40:9090",
			expectedIP: "10.20.30.40",
		},
		{
			name:       "Proxy-Client-Ip with port",
			headerName: "Proxy-Client-Ip",
			headerVal:  "172.16.100.50:12345",
			expectedIP: "172.16.100.50",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			body := []byte("test_log,host=myhost message=\"test\" 1609459200000000000")

			req := httptest.NewRequest("POST", "/v1/write/logging", bytes.NewReader(body))
			req.Header.Set("Content-Type", "text/plain")
			req.Header.Set(tc.headerName, tc.headerVal)

			wr := GetAPIWriteResult()
			defer PutAPIWriteResult(wr)

			err := wr.APIV1Write(req)
			require.NoError(t, err)
			require.NotEmpty(t, wr.Points)

			pt := wr.Points[0]
			collectorIP := pt.GetTag("collector_source_ip")
			assert.Equal(t, tc.expectedIP, collectorIP, "collector_source_ip should be extracted from %s", tc.headerName)
		})
	}
}
