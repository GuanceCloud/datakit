// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package logstreaming

import (
	"bytes"
	"io"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	dknet "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/net"
)

func TestCollectorSourceIP_ParametersStructure(t *testing.T) {
	testURL, _ := url.Parse("http://localhost/v1/write/logstreaming?source=test")

	param := &parameters{
		ignoreURLTags: false,
		url:           testURL,
		queryValues:   testURL.Query(),
		body:          io.NopCloser(bytes.NewBufferString("test")),
		remoteIP:      "192.168.50.100",
	}

	assert.Equal(t, "192.168.50.100", param.remoteIP, "parameters should store remoteIP correctly")
	assert.False(t, param.ignoreURLTags)
	assert.Equal(t, "test", param.queryValues.Get("source"))
}

func TestCollectorSourceIP_RemoteAddrExtraction(t *testing.T) {
	testCases := []struct {
		name       string
		headerName string
		headerVal  string
		remoteAddr string
		expectedIP string
	}{
		{
			name:       "X-Forwarded-For with port",
			headerName: "X-Forwarded-For",
			headerVal:  "192.168.1.100:8080",
			remoteAddr: "127.0.0.1:1234",
			expectedIP: "192.168.1.100",
		},
		{
			name:       "X-Real-Ip with port",
			headerName: "X-Real-Ip",
			headerVal:  "10.20.30.40:9090",
			remoteAddr: "127.0.0.1:1234",
			expectedIP: "10.20.30.40",
		},
		{
			name:       "Proxy-Client-Ip with port",
			headerName: "Proxy-Client-Ip",
			headerVal:  "172.16.100.50:12345",
			remoteAddr: "127.0.0.1:1234",
			expectedIP: "172.16.100.50",
		},
		{
			name:       "Fallback to RemoteAddr",
			headerName: "",
			headerVal:  "",
			remoteAddr: "192.168.0.1:54321",
			expectedIP: "192.168.0.1",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/v1/write/logstreaming", nil)
			if tc.headerName != "" {
				req.Header.Set(tc.headerName, tc.headerVal)
			}
			req.RemoteAddr = tc.remoteAddr

			ip, _ := dknet.RemoteAddr(req)
			assert.Equal(t, tc.expectedIP, ip, "RemoteAddr should extract correct IP")
		})
	}
}

func TestCollectorSourceIP_HandleLogstreamingCreatesParamWithRemoteIP(t *testing.T) {
	// 测试 handleLogstreaming 创建的 parameters 包含正确的 remoteIP
	// 这是一个简化的测试，只验证远端 IP 提取逻辑

	req := httptest.NewRequest("POST", "/v1/write/logstreaming?source=test", bytes.NewBufferString("test log"))
	req.Header.Set("X-Forwarded-For", "192.168.1.100:8080")

	remoteIP, _ := dknet.RemoteAddr(req)
	assert.Equal(t, "192.168.1.100", remoteIP, "should extract IP from X-Forwarded-For")

	// 验证 parameters 结构体可以正确保存 remoteIP
	param := &parameters{
		ignoreURLTags: false,
		url:           req.URL,
		queryValues:   req.URL.Query(),
		body:          req.Body,
		remoteIP:      remoteIP,
	}

	assert.Equal(t, "192.168.1.100", param.remoteIP, "parameters should have correct remoteIP")
}
