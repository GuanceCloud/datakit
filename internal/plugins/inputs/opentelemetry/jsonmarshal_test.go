// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package opentelemetry

import (
	T "testing"
	"time"

	trace "github.com/GuanceCloud/tracing-protos/opentelemetry-gen-go/collector/trace/v1"
	cv1 "github.com/GuanceCloud/tracing-protos/opentelemetry-gen-go/common/v1"
	rv1 "github.com/GuanceCloud/tracing-protos/opentelemetry-gen-go/resource/v1"
	tv1 "github.com/GuanceCloud/tracing-protos/opentelemetry-gen-go/trace/v1"

	"github.com/stretchr/testify/assert"
)

// 创建模拟的 Trace 数据 (包含 4-10 个 spans)
func createTestTraceData(nspan int) *trace.ExportTraceServiceRequest {
	traces := &trace.ExportTraceServiceRequest{
		ResourceSpans: []*tv1.ResourceSpans{
			{
				Resource: &rv1.Resource{
					Attributes: []*cv1.KeyValue{
						{Key: "service.name", Value: &cv1.AnyValue{Value: &cv1.AnyValue_StringValue{StringValue: "test-service"}}},
						{Key: "environment", Value: &cv1.AnyValue{Value: &cv1.AnyValue_StringValue{StringValue: "production"}}},
					},
				},
				ScopeSpans: []*tv1.ScopeSpans{{
					Scope: &cv1.InstrumentationScope{Name: "test-scope", Version: "1.0"},
					Spans: generateSpans(nspan),
				}},
			},
		},
	}

	return traces
}

// 生成指定数量的 Span
func generateSpans(count int) []*tv1.Span {
	spans := make([]*tv1.Span, 0, count)
	startTime := time.Now()

	for i := 0; i < count; i++ {
		endTime := startTime.Add(time.Duration(i) * time.Millisecond)

		span := &tv1.Span{
			TraceId:           []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15},
			SpanId:            []byte{0, 1, 2, 3, 4, 5, 6, byte(i)},
			ParentSpanId:      []byte{0, 1, 2, 3, 4, 5, 6, 7},
			Name:              "span-" + string(rune('A'+i)),
			Kind:              tv1.Span_SPAN_KIND_SERVER,
			StartTimeUnixNano: uint64(startTime.UnixNano()),
			EndTimeUnixNano:   uint64(endTime.UnixNano()),
			Attributes: []*cv1.KeyValue{
				{Key: "http.method", Value: &cv1.AnyValue{Value: &cv1.AnyValue_StringValue{StringValue: "GET"}}},
				{Key: "http.status_code", Value: &cv1.AnyValue{Value: &cv1.AnyValue_IntValue{IntValue: 200}}},
			},
			Events: []*tv1.Span_Event{
				{TimeUnixNano: uint64(startTime.UnixNano() + 1), Name: "event1"},
				{TimeUnixNano: uint64(endTime.UnixNano() - 1), Name: "event2"},
			},
			Status: &tv1.Status{Code: tv1.Status_STATUS_CODE_OK},
		}

		spans = append(spans, span)
	}
	return spans
}

func Benchmark_protojson(b *T.B) {
	traces := createTestTraceData(10)

	b.Run(`protojson`, func(b *T.B) {
		ipt := defaultInput()
		ipt.jmarshaler = &protojsonMarshaler{}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			ipt.parseResourceSpans(traces.ResourceSpans)
		}
	})
}

func Benchmark_golangjson(b *T.B) {
	traces := createTestTraceData(10)
	b.Run(`golang-json`, func(b *T.B) {
		ipt := defaultInput()
		ipt.jmarshaler = &gojsonMarshaler{}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			ipt.parseResourceSpans(traces.ResourceSpans)
		}
	})
}

func Benchmark_jsoniter(b *T.B) {
	traces := createTestTraceData(10)
	b.Run(`jsoniter`, func(b *T.B) {
		ipt := defaultInput()
		ipt.jmarshaler = &jsoniterMarshaler{}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			ipt.parseResourceSpans(traces.ResourceSpans)
		}
	})
}

func Benchmark_dropmsg(b *T.B) {
	traces := createTestTraceData(10)
	b.Run(`del-message`, func(b *T.B) {
		ipt := defaultInput()
		ipt.jmarshaler = &jsoniterMarshaler{}
		ipt.DelMessage = true
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			ipt.parseResourceSpans(traces.ResourceSpans)
		}
	})
}

func Test_parseResourceSpans(t *T.T) {
	traces := createTestTraceData(4)

	t.Run(`jsoniter`, func(t *T.T) {
		ipt := defaultInput()

		ipt.jmarshaler = &jsoniterMarshaler{}
		traces := ipt.parseResourceSpans(traces.ResourceSpans)
		assert.Len(t, traces, 1)
	})

	t.Run(`protojson`, func(t *T.T) {
		ipt := defaultInput()

		ipt.jmarshaler = &protojsonMarshaler{}
		traces := ipt.parseResourceSpans(traces.ResourceSpans)
		assert.Len(t, traces, 1)
	})

	t.Run(`golang-json`, func(t *T.T) {
		ipt := defaultInput()

		ipt.jmarshaler = &gojsonMarshaler{}
		traces := ipt.parseResourceSpans(traces.ResourceSpans)
		assert.Len(t, traces, 1)
	})
}

func TestJSONMarshal(t *T.T) {
	traces := createTestTraceData(4)

	t.Run(`jsoniter`, func(t *T.T) {
		jmarshaler := jsoniterMarshaler{}
		j, err := jmarshaler.Marshal(traces)
		assert.NoError(t, err)

		t.Logf("json:  %s", string(j))
	})

	t.Run(`protojson`, func(t *T.T) {
		jmarshaler := protojsonMarshaler{}
		j, err := jmarshaler.Marshal(traces)
		assert.NoError(t, err)
		t.Logf("json:  %s", string(j))
	})

	t.Run(`golang-json`, func(t *T.T) {
		jmarshaler := gojsonMarshaler{}
		j, err := jmarshaler.Marshal(traces)
		assert.NoError(t, err)

		t.Logf("json:  %s", string(j))
	})
}
