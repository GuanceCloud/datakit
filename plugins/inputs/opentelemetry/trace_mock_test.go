package opentelemetry

import (
	"context"
	"testing"
	"time"

	DKtrace "gitlab.jiagouyun.com/cloudcare-tools/datakit/io/trace"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/instrumentation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
	tracev1 "go.opentelemetry.io/otel/trace"
)

/*
got :
&{
	TraceID:00000000000000000000000000000001
	ParentID:0
	SpanID:0000000000000002
	Service:global.ServerName
	Resource:test-server
	Operation:span_name
	Source:opentelemetry
	SpanType:SPAN_KIND_UNSPECIFIED
	SourceType: Env: Project: Version:
	Tags:map[a:b int:123]
	EndPoint: HTTPMethod: HTTPStatusCode: ContainerHost: PID:
	Start:1645413248574237500
	Duration:1000000000
	Status:info
	Content:{"TraceID":"00000000000000000000000000000001","ParentID":"0","SpanID":"0000000000000002","Service":"global.ServerName","Resource":"test-server","Operation":"span_name","Source":"opentelemetry","SpanType":"SPAN_KIND_UNSPECIFIED","SourceType":"","Env":"","Project":"","Version":"","Tags":{"a":"b","int":"123"},"EndPoint":"","HTTPMethod":"","HTTPStatusCode":"","ContainerHost":"","PID":"","Start":1645413248574237500,"Duration":1000000000,"Status":"info","Content":"","Priority":0,"SamplingRateGlobal":0}
	Priority:0
	SamplingRateGlobal:0
},

{
"trace_id":"AAAAAAAAAAAAAAAAAAAAAQ==",
"span_id":"AAAAAAAAAAI=",
"name":"span_name",
"start_time_unix_nano":1645423252614239800,
"end_time_unix_nano":1645423253614239800,
"attributes":[
{"key":"a","value":{"Value":{"StringValue":"b"}}},
{"key":"int","value":{"Value":{"IntValue":123}}}
]

want:
&{
	TraceID:1
	ParentID:0
	SpanID:2
	Service:global.ServerName
	Resource:test-server
	Operation:span_name
	Source:
	SpanType: SourceType: Env: Project: Version:
	Tags:map[a:b int:123]
	EndPoint: HTTPMethod: HTTPStatusCode: ContainerHost: PID:
	Start:1645413248574237500
	Duration:1645413249574237500
	Status:
	Content:
	Priority:0
	SamplingRateGlobal:0
}

*/

func mockRoSpans(t *testing.T) (roSpans []sdktrace.ReadOnlySpan, want []DKtrace.DatakitTrace) {
	startTime := time.Now()
	endTime := startTime.Add(time.Second)
	traceID, err := tracev1.TraceIDFromHex("00000000000000000000000000000001")
	if err != nil {
		t.Errorf("err=%v", err)
		return
	}
	spanID, err := tracev1.SpanIDFromHex("0000000000000002")
	if err != nil {
		t.Errorf("err=%v", err)
		return
	}
	ctx := context.Background()
	res, err := resource.New(ctx,
		resource.WithAttributes(
			// the service name used to display traces in backends
			semconv.ServiceNameKey.String("global.ServerName"),
			// semconv.FaaSIDKey.String(""),
		),
	)

	spanCxt := tracev1.NewSpanContext(tracev1.SpanContextConfig{
		TraceID:    traceID,
		SpanID:     spanID,
		TraceFlags: 0,
		TraceState: tracev1.TraceState{},
		Remote:     false,
	})

	//  创建mockSpan数组 和dktrace数组  进行mock测试
	roSpans = tracetest.SpanStubs{tracetest.SpanStub{
		Name:                   "span_name",
		SpanContext:            spanCxt,
		Parent:                 tracev1.SpanContext{},
		SpanKind:               0,
		StartTime:              startTime,
		EndTime:                endTime,
		Attributes:             []attribute.KeyValue{attribute.String("a", "b"), attribute.Int("int", 123)},
		Events:                 nil,
		Links:                  nil,
		Status:                 sdktrace.Status{},
		DroppedAttributes:      0,
		DroppedEvents:          0,
		DroppedLinks:           0,
		ChildSpanCount:         0,
		Resource:               res,
		InstrumentationLibrary: instrumentation.Library{Name: "test-server"},
	}}.Snapshots()

	wantContent := `{"trace_id":"AAAAAAAAAAAAAAAAAAAAAQ==","span_id":"AAAAAAAAAAI=","name":"span_name","start_time_unix_nano":1645423573257862600,"end_time_unix_nano":1645423574257862600,"attributes":[{"key":"a","value":{"Value":{"StringValue":"b"}}},{"key":"int","value":{"Value":{"IntValue":123}}}],"status":{}}`
	want = []DKtrace.DatakitTrace{[]*DKtrace.DatakitSpan{&DKtrace.DatakitSpan{
		TraceID:            "00000000000000000000000000000001",
		ParentID:           "0",
		SpanID:             "0000000000000002",
		Service:            "global.ServerName",
		Resource:           "test-server",
		Operation:          "span_name",
		Source:             inputName,
		SpanType:           "SPAN_KIND_UNSPECIFIED",
		SourceType:         "",
		Env:                "",
		Project:            "",
		Version:            "",
		Tags:               map[string]string{"a": "b", "int": "123"},
		EndPoint:           "",
		HTTPMethod:         "",
		HTTPStatusCode:     "",
		ContainerHost:      "",
		PID:                "",
		Start:              startTime.UnixNano(),
		Duration:           endTime.UnixNano() - startTime.UnixNano(),
		Status:             "info",
		Content:            wantContent,
		Priority:           0,
		SamplingRateGlobal: 0,
	}}}

	return roSpans, want
}
