// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package opentelemetry

import (
	"fmt"
	"os"
	"strconv"
	T "testing"
	"time"

	trace "github.com/GuanceCloud/tracing-protos/opentelemetry-gen-go/collector/trace/v1"
	cv1 "github.com/GuanceCloud/tracing-protos/opentelemetry-gen-go/common/v1"
	rv1 "github.com/GuanceCloud/tracing-protos/opentelemetry-gen-go/resource/v1"
	tv1 "github.com/GuanceCloud/tracing-protos/opentelemetry-gen-go/trace/v1"
	itrace "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/trace"
	"google.golang.org/protobuf/proto"

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
						{Key: itrace.FieldRuntimeID, Value: &cv1.AnyValue{Value: &cv1.AnyValue_StringValue{StringValue: "mocked-runtime-id"}}},
					},
				},
				ScopeSpans: []*tv1.ScopeSpans{{
					Scope: &cv1.InstrumentationScope{
						Name:    "test-scope",
						Version: "1.0",
						Attributes: []*cv1.KeyValue{
							{Key: "scope-str-attr", Value: &cv1.AnyValue{Value: &cv1.AnyValue_StringValue{StringValue: "mocked-str-attr"}}},
							{Key: "scope-bool-attr", Value: &cv1.AnyValue{Value: &cv1.AnyValue_BoolValue{BoolValue: true}}},
							{Key: "scope-double-attr", Value: &cv1.AnyValue{Value: &cv1.AnyValue_DoubleValue{DoubleValue: 3.14}}},
							{Key: "scope-byte-attr", Value: &cv1.AnyValue{Value: &cv1.AnyValue_BytesValue{BytesValue: []byte("byte-string")}}},
							{Key: "scope-int-attr", Value: &cv1.AnyValue{Value: &cv1.AnyValue_IntValue{IntValue: int64(42)}}},
							{
								Key: "scope-list-attr", Value: &cv1.AnyValue{
									Value: &cv1.AnyValue_ArrayValue{
										ArrayValue: &cv1.ArrayValue{
											Values: []*cv1.AnyValue{
												{Value: &cv1.AnyValue_IntValue{IntValue: int64(42)}},
												{Value: &cv1.AnyValue_DoubleValue{DoubleValue: 3.14}},
												{Value: &cv1.AnyValue_BytesValue{BytesValue: []byte("byte-string")}},
												{Value: &cv1.AnyValue_BoolValue{BoolValue: true}},
											},
										},
									},
								},
							},

							{
								Key: "scope-kv-attr", Value: &cv1.AnyValue{
									Value: &cv1.AnyValue_KvlistValue{
										KvlistValue: &cv1.KeyValueList{
											Values: []*cv1.KeyValue{
												{Key: "int", Value: &cv1.AnyValue{Value: &cv1.AnyValue_IntValue{IntValue: int64(42)}}},
												{Key: "double", Value: &cv1.AnyValue{Value: &cv1.AnyValue_DoubleValue{DoubleValue: 3.14}}},
												{Key: "byte-string", Value: &cv1.AnyValue{Value: &cv1.AnyValue_BytesValue{BytesValue: []byte("byte-string")}}},
												{Key: "boolean", Value: &cv1.AnyValue{Value: &cv1.AnyValue_BoolValue{BoolValue: true}}},
											},
										},
									},
								},
							},
						},
					},
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
				{Key: "project.id", Value: &cv1.AnyValue{Value: &cv1.AnyValue_StringValue{StringValue: "project-001"}}},
				{Key: "app_id", Value: &cv1.AnyValue{Value: &cv1.AnyValue_StringValue{StringValue: "app-001"}}},
				{Key: "not.used", Value: &cv1.AnyValue{Value: &cv1.AnyValue_StringValue{StringValue: "not-used-value"}}},
				{Key: "db.system", Value: &cv1.AnyValue{Value: &cv1.AnyValue_StringValue{StringValue: "postgresql"}}},
			},
			Events: []*tv1.Span_Event{
				{TimeUnixNano: uint64(startTime.UnixNano() + 1), Name: "event1"},
				{TimeUnixNano: uint64(endTime.UnixNano() - 1), Name: "event2"},
				{
					TimeUnixNano: uint64(endTime.UnixNano() - 1),
					Name:         ExceptionEventName,
					Attributes: []*cv1.KeyValue{
						{
							Key:   ExceptionTypeKey,
							Value: &cv1.AnyValue{Value: &cv1.AnyValue_StringValue{StringValue: "mocked-exception-type"}},
						},
						{
							Key:   ExceptionMessageKey,
							Value: &cv1.AnyValue{Value: &cv1.AnyValue_StringValue{StringValue: "mocked-exception-message"}},
						},
						{
							Key:   ExceptionStacktraceKey,
							Value: &cv1.AnyValue{Value: &cv1.AnyValue_StringValue{StringValue: "mocked-exception-stack"}},
						},
					},
				},
			},
			Status: &tv1.Status{Code: tv1.Status_STATUS_CODE_OK},
		}

		spans = append(spans, span)
	}
	return spans
}

func Benchmark_protojson(b *T.B) {
	traces := createTestTraceData(10)
	b.ResetTimer()

	b.Run(`protojson`, func(b *T.B) {
		ipt := defaultInput()
		ipt.CleanMessage = false
		ipt.jmarshaler = &protojsonMarshaler{}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			ipt.parseResourceSpans(traces.ResourceSpans)
		}
	})
}

func Benchmark_protojsonCleanMessage(b *T.B) {
	traces := createTestTraceData(10)
	b.ResetTimer()

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
	nspan := 10
	traces := createTestTraceData(nspan)
	pb, err := proto.Marshal(traces)

	assert.NoError(b, err)

	ipt := defaultInput()
	if v := os.Getenv("CLEAN_MESSAGE"); v != "" {
		if x, err := strconv.ParseBool(v); err == nil {
			ipt.CleanMessage = x
		}
	}

	ipt.jmarshaler = &gojsonMarshaler{}
	b.ResetTimer()

	b.Run(fmt.Sprintf(`golang-json-pb(#%d)-%d-span`, len(pb), nspan), func(b *T.B) {
		for i := 0; i < b.N; i++ {
			tsreq := &trace.ExportTraceServiceRequest{}
			proto.Unmarshal(pb, tsreq)
			ipt.parseResourceSpans(tsreq.ResourceSpans)
		}
	})
}

func Benchmark_jsonmarshal(b *T.B) {
	traces := createTestTraceData(10)

	b.Run(`jsoniter`, func(b *T.B) {
		m := &jsoniterMarshaler{}
		for i := 0; i < b.N; i++ {
			m.Marshal(traces)
		}
	})

	b.Run(`golang-json`, func(b *T.B) {
		m := &gojsonMarshaler{}
		for i := 0; i < b.N; i++ {
			m.Marshal(traces)
		}
	})

	b.Run(`pb-json`, func(b *T.B) {
		m := &protojsonMarshaler{}
		for i := 0; i < b.N; i++ {
			m.Marshal(traces)
		}
	})
}

func Benchmark_jsoniter(b *T.B) {
	traces := createTestTraceData(10)
	b.Run(`jsoniter`, func(b *T.B) {
		ipt := defaultInput()
		ipt.CleanMessage = false
		ipt.jmarshaler = &jsoniterMarshaler{}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			ipt.parseResourceSpans(traces.ResourceSpans)
		}
	})
}

func Benchmark_jsoniterCleanMessage(b *T.B) {
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
