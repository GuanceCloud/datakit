// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package jaeger

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/uber/jaeger-client-go/thrift-gen/jaeger"
	itrace "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/trace"
)

func TestCollectorSourceIP_JaegerBatchToDkTrace(t *testing.T) {
	// 创建测试 batch 数据
	serviceName := "test-jaeger-service"
	batch := &jaeger.Batch{
		Process: &jaeger.Process{
			ServiceName: serviceName,
		},
		Spans: []*jaeger.Span{
			{
				TraceIdLow:    12345,
				TraceIdHigh:   0,
				SpanId:        67890,
				ParentSpanId:  0,
				OperationName: "test-operation",
				StartTime:     1609459200000000, // 2021-01-01 00:00:00 UTC in microseconds
				Duration:      1000,
				Tags:          []*jaeger.Tag{},
			},
		},
	}

	testRemoteIP := "192.168.100.50"

	// 调用 batchToDkTrace 并传入 remoteIP
	dktrace := batchToDkTrace(batch, testRemoteIP)

	require.NotEmpty(t, dktrace, "dktrace should not be empty")

	// 验证 collector_source_ip tag
	pt := dktrace[0].Point
	collectorIP := pt.GetTag(itrace.TagCollectorSourceIP)
	assert.Equal(t, testRemoteIP, collectorIP, "collector_source_ip should be set correctly")
}

func TestCollectorSourceIP_JaegerEmptyRemoteIP(t *testing.T) {
	serviceName := "test-jaeger-service"
	batch := &jaeger.Batch{
		Process: &jaeger.Process{
			ServiceName: serviceName,
		},
		Spans: []*jaeger.Span{
			{
				TraceIdLow:    12345,
				TraceIdHigh:   0,
				SpanId:        67890,
				ParentSpanId:  0,
				OperationName: "test-operation",
				StartTime:     1609459200000000,
				Duration:      1000,
				Tags:          []*jaeger.Tag{},
			},
		},
	}

	// 测试空的 remoteIP
	dktrace := batchToDkTrace(batch, "")

	require.NotEmpty(t, dktrace, "dktrace should not be empty")

	// 验证 collector_source_ip tag 为空
	pt := dktrace[0].Point
	collectorIP := pt.GetTag(itrace.TagCollectorSourceIP)
	assert.Equal(t, "", collectorIP, "collector_source_ip should be empty when remoteIP is empty")
}

func TestCollectorSourceIP_JaegerTraceParameters(t *testing.T) {
	// 创建 TraceParameters 并设置 RemoteIP
	param := &itrace.TraceParameters{
		Body:     bytes.NewBuffer([]byte{}),
		RemoteIP: "10.20.30.40",
	}

	// 验证 TraceParameters 正确存储了 RemoteIP
	assert.Equal(t, "10.20.30.40", param.RemoteIP, "TraceParameters should store RemoteIP correctly")
}
