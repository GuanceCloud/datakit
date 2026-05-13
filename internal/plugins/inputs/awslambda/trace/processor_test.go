// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2024-present Guance, Inc.

package lambdatrace

import (
	"context"
	"fmt"
	"testing"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/awslambda/lambdaapi/model"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/awslambda/lambdaapi/telemetry"
)

type testSink struct{}

func (testSink) Consume(context.Context, []Span) error {
	return nil
}

type recordingSink struct {
	spans []Span
}

func (s *recordingSink) Consume(_ context.Context, spans []Span) error {
	s.spans = append(s.spans, spans...)
	return nil
}

func TestProcessorFlushesOnRuntimeDoneLikeDatadog(t *testing.T) {
	sink := &recordingSink{}
	p := NewProcessor(sink, false)

	p.OnUniversalStart("request-1", map[string]string{"x-datadog-trace-id": "123"}, nil)
	if err := p.OnUniversalEnd("request-1", nil, nil); err != nil {
		t.Fatalf("universal end failed: %v", err)
	}
	if len(sink.spans) != 0 {
		t.Fatalf("expected no spans before runtime done, got %d", len(sink.spans))
	}

	if err := p.OnPlatformRuntimeDone("request-1", 123, "success"); err != nil {
		t.Fatalf("runtime done failed: %v", err)
	}
	if len(sink.spans) == 0 {
		t.Fatal("expected runtime done to flush context spans")
	}
	if _, ok := p.contexts["request-1"]; ok {
		t.Fatal("expected flushed context to be removed")
	}

	if err := p.OnPlatformReport("request-1", 150, "success"); err != nil {
		t.Fatalf("late platform report failed: %v", err)
	}
	if _, ok := p.contexts["request-1"]; ok {
		t.Fatal("late platform report should not recreate flushed context")
	}
}

func TestProcessorDoesNotFlushOnOnDemandReportBeforeRuntimeDone(t *testing.T) {
	sink := &recordingSink{}
	p := NewProcessor(sink, false)

	p.OnUniversalStart("request-1", map[string]string{"x-datadog-trace-id": "123"}, nil)
	if err := p.OnUniversalEnd("request-1", nil, nil); err != nil {
		t.Fatalf("universal end failed: %v", err)
	}
	if err := p.OnPlatformReport("request-1", 150, "success"); err != nil {
		t.Fatalf("platform report failed: %v", err)
	}
	if len(sink.spans) != 0 {
		t.Fatalf("expected on-demand report not to flush before runtime done, got %d spans", len(sink.spans))
	}
}

func TestProcessorLateReportInitDurationEmitsDeferredColdStart(t *testing.T) {
	sink := &recordingSink{}
	p := NewProcessor(sink, false)

	p.OnUniversalStart("request-1", map[string]string{"x-datadog-trace-id": "123"}, nil)
	if err := p.OnUniversalEnd("request-1", nil, nil); err != nil {
		t.Fatalf("universal end failed: %v", err)
	}
	if err := p.OnPlatformRuntimeDone("request-1", 123, "success"); err != nil {
		t.Fatalf("runtime done failed: %v", err)
	}

	if got := countSpanByName(sink.spans, "aws.lambda.cold_start"); got != 0 {
		t.Fatalf("expected no cold start span before init duration arrives, got %d", got)
	}

	initDuration := 549.04
	event := &telemetry.Event{
		Record: &telemetry.PlatformReport{
			RequestID: "request-1",
			Status:    model.StatusSuccess,
			Metrics: model.ReportMetrics{
				BaseMetrics:    model.BaseMetrics{DurationMs: 549.04},
				InitDurationMs: &initDuration,
			},
		},
	}
	if err := p.OnTelemetryEvent(event); err != nil {
		t.Fatalf("late platform report failed: %v", err)
	}

	if got := countSpanByName(sink.spans, "aws.lambda.cold_start"); got != 1 {
		t.Fatalf("expected deferred cold start span after late report, got %d", got)
	}
}

func TestProcessorUsesUpstreamDatadogContextWithInferredSpan(t *testing.T) {
	sink := &recordingSink{}
	p := NewProcessor(sink, false)

	payload := []byte(`{
		"httpMethod":"GET",
		"path":"/orders",
		"requestContext":{"requestId":"request-1"},
		"headers":{
			"x-datadog-trace-id":"12345",
			"x-datadog-parent-id":"67890",
			"x-datadog-sampling-priority":"1"
		}
	}`)

	p.OnUniversalStart("request-1", nil, payload)
	if err := p.OnUniversalEnd("request-1", nil, nil); err != nil {
		t.Fatalf("universal end failed: %v", err)
	}
	if err := p.OnPlatformRuntimeDone("request-1", 123, "success"); err != nil {
		t.Fatalf("runtime done failed: %v", err)
	}

	inferredSpan := findSpanByName(t, sink.spans, "aws.apigateway")
	lambdaSpan := findSpanByName(t, sink.spans, "aws.lambda")
	if inferredSpan.TraceID != 12345 || lambdaSpan.TraceID != 12345 {
		t.Fatalf("expected upstream trace id 12345, got inferred=%d lambda=%d", inferredSpan.TraceID, lambdaSpan.TraceID)
	}
	if inferredSpan.ParentID != 67890 {
		t.Fatalf("expected inferred span parent to be upstream parent 67890, got %d", inferredSpan.ParentID)
	}
	if lambdaSpan.ParentID != inferredSpan.SpanID {
		t.Fatalf("expected lambda parent to be inferred span %d, got %d", inferredSpan.SpanID, lambdaSpan.ParentID)
	}
	if inferredSpan.Meta["_sampling_priority_v1"] != "1" {
		t.Fatalf("expected sampling priority on inferred span, got %#v", inferredSpan.Meta)
	}
}

func TestProcessorUsesUpstreamDatadogContextWithoutInferredSpan(t *testing.T) {
	sink := &recordingSink{}
	p := NewProcessor(sink, false)

	payload := []byte(`{
		"headers":{
			"x-datadog-trace-id":"223344",
			"x-datadog-parent-id":"998877"
		}
	}`)

	p.OnUniversalStart("request-1", nil, payload)
	if err := p.OnUniversalEnd("request-1", nil, nil); err != nil {
		t.Fatalf("universal end failed: %v", err)
	}
	if err := p.OnPlatformRuntimeDone("request-1", 123, "success"); err != nil {
		t.Fatalf("runtime done failed: %v", err)
	}

	lambdaSpan := findSpanByName(t, sink.spans, "aws.lambda")
	if lambdaSpan.TraceID != 223344 {
		t.Fatalf("expected upstream trace id 223344, got %d", lambdaSpan.TraceID)
	}
	if lambdaSpan.ParentID != 998877 {
		t.Fatalf("expected lambda parent to be upstream parent 998877, got %d", lambdaSpan.ParentID)
	}
}

func TestProcessorUsesUpstreamW3CContext(t *testing.T) {
	sink := &recordingSink{}
	p := NewProcessor(sink, false)

	payload := []byte(`{
		"httpMethod":"GET",
		"path":"/w3c",
		"requestContext":{"requestId":"request-1"},
		"headers":{
			"traceparent":"00-0000000000000001123456789abcdef0-00000000000000ff-01",
			"tracestate":"dd=s:1;o:synthetics;t.dm:-4"
		}
	}`)

	p.OnUniversalStart("request-1", nil, payload)
	if err := p.OnUniversalEnd("request-1", nil, nil); err != nil {
		t.Fatalf("universal end failed: %v", err)
	}
	if err := p.OnPlatformRuntimeDone("request-1", 123, "success"); err != nil {
		t.Fatalf("runtime done failed: %v", err)
	}

	inferredSpan := findSpanByName(t, sink.spans, "aws.apigateway")
	if inferredSpan.TraceID != 1311768467463790320 {
		t.Fatalf("expected W3C low 64-bit trace id, got %d", inferredSpan.TraceID)
	}
	if inferredSpan.ParentID != 255 {
		t.Fatalf("expected W3C parent id 255, got %d", inferredSpan.ParentID)
	}
	if inferredSpan.Meta["_dd.p.tid"] != "0000000000000001" {
		t.Fatalf("expected high 64-bit trace id tag, got %#v", inferredSpan.Meta)
	}
	if inferredSpan.Meta["_sampling_priority_v1"] != "1" || inferredSpan.Meta["_dd.origin"] != "synthetics" {
		t.Fatalf("expected tracestate Datadog meta, got %#v", inferredSpan.Meta)
	}
}

func TestProcessorUsesUpstreamB3Context(t *testing.T) {
	sink := &recordingSink{}
	p := NewProcessor(sink, false)

	payload := []byte(`{
		"httpMethod":"GET",
		"path":"/b3",
		"requestContext":{"requestId":"request-1"},
		"headers":{
			"x-b3-traceid":"0000000000000001123456789abcdef0",
			"x-b3-spanid":"0000000000000abc",
			"x-b3-sampled":"1"
		}
	}`)

	p.OnUniversalStart("request-1", nil, payload)
	if err := p.OnUniversalEnd("request-1", nil, nil); err != nil {
		t.Fatalf("universal end failed: %v", err)
	}
	if err := p.OnPlatformRuntimeDone("request-1", 123, "success"); err != nil {
		t.Fatalf("runtime done failed: %v", err)
	}

	inferredSpan := findSpanByName(t, sink.spans, "aws.apigateway")
	if inferredSpan.TraceID != 1311768467463790320 || inferredSpan.ParentID != 2748 {
		t.Fatalf("expected B3 upstream context, got trace=%d parent=%d", inferredSpan.TraceID, inferredSpan.ParentID)
	}
	if inferredSpan.Meta["_dd.p.tid"] != "0000000000000001" || inferredSpan.Meta["_sampling_priority_v1"] != "1" {
		t.Fatalf("expected B3 meta, got %#v", inferredSpan.Meta)
	}
}

func TestProcessorUsesUpstreamXRayContext(t *testing.T) {
	sink := &recordingSink{}
	p := NewProcessor(sink, false)

	payload := []byte(`{
		"httpMethod":"GET",
		"path":"/xray",
		"requestContext":{"requestId":"request-1"},
		"headers":{
			"x-amzn-trace-id":"Root=1-00000001-123456789abcdef012345678;Parent=0000000000000abc;Sampled=1"
		}
	}`)

	p.OnUniversalStart("request-1", nil, payload)
	if err := p.OnUniversalEnd("request-1", nil, nil); err != nil {
		t.Fatalf("universal end failed: %v", err)
	}
	if err := p.OnPlatformRuntimeDone("request-1", 123, "success"); err != nil {
		t.Fatalf("runtime done failed: %v", err)
	}

	inferredSpan := findSpanByName(t, sink.spans, "aws.apigateway")
	if inferredSpan.TraceID != 11150031900141442680 || inferredSpan.ParentID != 2748 {
		t.Fatalf("expected X-Ray upstream context, got trace=%d parent=%d", inferredSpan.TraceID, inferredSpan.ParentID)
	}
	if inferredSpan.Meta["_dd.p.tid"] != "0000000112345678" || inferredSpan.Meta["_sampling_priority_v1"] != "1" {
		t.Fatalf("expected X-Ray meta, got %#v", inferredSpan.Meta)
	}
}

func TestProcessorUsesUpstreamDatadogContextFromSQSMessageAttributes(t *testing.T) {
	sink := &recordingSink{}
	p := NewProcessor(sink, false)

	payload := []byte(`{
		"Records":[{
			"eventSource":"aws:sqs",
			"eventSourceARN":"arn:aws:sqs:queue",
			"messageAttributes":{
				"x-datadog-trace-id":{"stringValue":"321","dataType":"String"},
				"x-datadog-parent-id":{"stringValue":"654","dataType":"String"}
			}
		}]
	}`)

	p.OnUniversalStart("request-1", nil, payload)
	if err := p.OnUniversalEnd("request-1", nil, nil); err != nil {
		t.Fatalf("universal end failed: %v", err)
	}
	if err := p.OnPlatformRuntimeDone("request-1", 123, "success"); err != nil {
		t.Fatalf("runtime done failed: %v", err)
	}

	inferredSpan := findSpanByName(t, sink.spans, "aws.sqs")
	if inferredSpan.TraceID != 321 || inferredSpan.ParentID != 654 {
		t.Fatalf("expected SQS upstream context, got trace=%d parent=%d", inferredSpan.TraceID, inferredSpan.ParentID)
	}
}

func TestProcessorUsesUpstreamDatadogContextFromSNSMessageAttributes(t *testing.T) {
	sink := &recordingSink{}
	p := NewProcessor(sink, false)

	payload := []byte(`{
		"Records":[{
			"EventSource":"aws:sns",
			"Sns":{
				"TopicArn":"arn:aws:sns:topic",
				"MessageAttributes":{
					"x-datadog-trace-id":{"Type":"String","Value":"4321"},
					"x-datadog-parent-id":{"Type":"String","Value":"8765"}
				}
			}
		}]
	}`)

	p.OnUniversalStart("request-1", nil, payload)
	if err := p.OnUniversalEnd("request-1", nil, nil); err != nil {
		t.Fatalf("universal end failed: %v", err)
	}
	if err := p.OnPlatformRuntimeDone("request-1", 123, "success"); err != nil {
		t.Fatalf("runtime done failed: %v", err)
	}

	inferredSpan := findSpanByName(t, sink.spans, "aws.sns")
	if inferredSpan.TraceID != 4321 || inferredSpan.ParentID != 8765 {
		t.Fatalf("expected SNS upstream context, got trace=%d parent=%d", inferredSpan.TraceID, inferredSpan.ParentID)
	}
}

func TestProcessorUsesUpstreamDatadogContextFromEventBridgeDetailHeaders(t *testing.T) {
	sink := &recordingSink{}
	p := NewProcessor(sink, false)

	payload := []byte(`{
		"source":"custom.app",
		"detail-type":"app.event",
		"detail":{
			"headers":{
				"x-datadog-trace-id":"9876",
				"x-datadog-parent-id":"5432"
			}
		}
	}`)

	p.OnUniversalStart("request-1", nil, payload)
	if err := p.OnUniversalEnd("request-1", nil, nil); err != nil {
		t.Fatalf("universal end failed: %v", err)
	}
	if err := p.OnPlatformRuntimeDone("request-1", 123, "success"); err != nil {
		t.Fatalf("runtime done failed: %v", err)
	}

	inferredSpan := findSpanByName(t, sink.spans, "aws.eventbridge")
	if inferredSpan.TraceID != 9876 || inferredSpan.ParentID != 5432 {
		t.Fatalf("expected EventBridge upstream context, got trace=%d parent=%d", inferredSpan.TraceID, inferredSpan.ParentID)
	}
}

func TestProcessorUsesUpstreamDatadogContextFromKinesisBase64JSON(t *testing.T) {
	sink := &recordingSink{}
	p := NewProcessor(sink, false)

	payload := []byte(`{
		"Records":[{
			"eventSource":"aws:kinesis",
			"eventSourceARN":"arn:aws:kinesis:stream",
			"kinesis":{
				"data":"eyJoZWFkZXJzIjp7IngtZGF0YWRvZy10cmFjZS1pZCI6IjI0NjgiLCJ4LWRhdGFkb2ctcGFyZW50LWlkIjoiMTM1NyJ9fQ=="
			}
		}]
	}`)

	p.OnUniversalStart("request-1", nil, payload)
	if err := p.OnUniversalEnd("request-1", nil, nil); err != nil {
		t.Fatalf("universal end failed: %v", err)
	}
	if err := p.OnPlatformRuntimeDone("request-1", 123, "success"); err != nil {
		t.Fatalf("runtime done failed: %v", err)
	}

	inferredSpan := findSpanByName(t, sink.spans, "aws.kinesis")
	if inferredSpan.TraceID != 2468 || inferredSpan.ParentID != 1357 {
		t.Fatalf("expected Kinesis upstream context, got trace=%d parent=%d", inferredSpan.TraceID, inferredSpan.ParentID)
	}
}

func TestProcessorUsesUpstreamDatadogContextFromSQSBodyJSON(t *testing.T) {
	sink := &recordingSink{}
	p := NewProcessor(sink, false)

	payload := []byte(`{
		"Records":[{
			"eventSource":"aws:sqs",
			"eventSourceARN":"arn:aws:sqs:queue",
			"body":"{\"headers\":{\"x-datadog-trace-id\":\"1122\",\"x-datadog-parent-id\":\"3344\"}}"
		}]
	}`)

	p.OnUniversalStart("request-1", nil, payload)
	if err := p.OnUniversalEnd("request-1", nil, nil); err != nil {
		t.Fatalf("universal end failed: %v", err)
	}
	if err := p.OnPlatformRuntimeDone("request-1", 123, "success"); err != nil {
		t.Fatalf("runtime done failed: %v", err)
	}

	inferredSpan := findSpanByName(t, sink.spans, "aws.sqs")
	if inferredSpan.TraceID != 1122 || inferredSpan.ParentID != 3344 {
		t.Fatalf("expected SQS body upstream context, got trace=%d parent=%d", inferredSpan.TraceID, inferredSpan.ParentID)
	}
}

func TestProcessorUsesUpstreamDatadogContextFromSNSMessageJSON(t *testing.T) {
	sink := &recordingSink{}
	p := NewProcessor(sink, false)

	payload := []byte(`{
		"Records":[{
			"EventSource":"aws:sns",
			"Sns":{
				"TopicArn":"arn:aws:sns:topic",
				"Message":"{\"headers\":{\"x-datadog-trace-id\":\"5566\",\"x-datadog-parent-id\":\"7788\"}}"
			}
		}]
	}`)

	p.OnUniversalStart("request-1", nil, payload)
	if err := p.OnUniversalEnd("request-1", nil, nil); err != nil {
		t.Fatalf("universal end failed: %v", err)
	}
	if err := p.OnPlatformRuntimeDone("request-1", 123, "success"); err != nil {
		t.Fatalf("runtime done failed: %v", err)
	}

	inferredSpan := findSpanByName(t, sink.spans, "aws.sns")
	if inferredSpan.TraceID != 5566 || inferredSpan.ParentID != 7788 {
		t.Fatalf("expected SNS message upstream context, got trace=%d parent=%d", inferredSpan.TraceID, inferredSpan.ParentID)
	}
}

func TestProcessorUsesUpstreamDatadogContextFromDynamoDBAttributeMap(t *testing.T) {
	sink := &recordingSink{}
	p := NewProcessor(sink, false)

	payload := []byte(`{
		"Records":[{
			"eventSource":"aws:dynamodb",
			"eventSourceARN":"arn:aws:dynamodb:stream",
			"dynamodb":{
				"NewImage":{
					"headers":{
						"M":{
							"x-datadog-trace-id":{"S":"9911"},
							"x-datadog-parent-id":{"S":"1199"}
						}
					}
				}
			}
		}]
	}`)

	p.OnUniversalStart("request-1", nil, payload)
	if err := p.OnUniversalEnd("request-1", nil, nil); err != nil {
		t.Fatalf("universal end failed: %v", err)
	}
	if err := p.OnPlatformRuntimeDone("request-1", 123, "success"); err != nil {
		t.Fatalf("runtime done failed: %v", err)
	}

	inferredSpan := findSpanByName(t, sink.spans, "aws.dynamodb")
	if inferredSpan.TraceID != 9911 || inferredSpan.ParentID != 1199 {
		t.Fatalf("expected DynamoDB upstream context, got trace=%d parent=%d", inferredSpan.TraceID, inferredSpan.ParentID)
	}
}

func TestProcessorUsesUpstreamDatadogContextFromKafkaHeaders(t *testing.T) {
	sink := &recordingSink{}
	p := NewProcessor(sink, false)

	payload := []byte(`{
		"eventSource":"aws:kafka",
		"records":{
			"topic-0":[
				{
					"topic":"topic",
					"partition":0,
					"offset":1,
					"headers":[
						{"x-datadog-trace-id":[49,50,49,50]},
						{"x-datadog-parent-id":[51,52,51,52]}
					],
					"value":"e30="
				}
			]
		}
	}`)

	p.OnUniversalStart("request-1", nil, payload)
	if err := p.OnUniversalEnd("request-1", nil, nil); err != nil {
		t.Fatalf("universal end failed: %v", err)
	}
	if err := p.OnPlatformRuntimeDone("request-1", 123, "success"); err != nil {
		t.Fatalf("runtime done failed: %v", err)
	}

	inferredSpan := findSpanByName(t, sink.spans, "aws.msk")
	lambdaSpan := findSpanByName(t, sink.spans, "aws.lambda")
	if inferredSpan.TraceID != 1212 || inferredSpan.ParentID != 3434 {
		t.Fatalf("expected Kafka upstream context, got trace=%d parent=%d", inferredSpan.TraceID, inferredSpan.ParentID)
	}
	if lambdaSpan.ParentID != inferredSpan.SpanID {
		t.Fatalf("expected lambda parent to be Kafka inferred span %d, got %d", inferredSpan.SpanID, lambdaSpan.ParentID)
	}
}

func TestProcessorUsesUpstreamDatadogContextFromStepFunctionsInput(t *testing.T) {
	sink := &recordingSink{}
	p := NewProcessor(sink, false)

	payload := []byte(`{
		"Execution":{"Id":"execution-1"},
		"StateMachine":{"Name":"state-machine"},
		"Input":{
			"headers":{
				"x-datadog-trace-id":"5656",
				"x-datadog-parent-id":"7878"
			}
		}
	}`)

	p.OnUniversalStart("request-1", nil, payload)
	if err := p.OnUniversalEnd("request-1", nil, nil); err != nil {
		t.Fatalf("universal end failed: %v", err)
	}
	if err := p.OnPlatformRuntimeDone("request-1", 123, "success"); err != nil {
		t.Fatalf("runtime done failed: %v", err)
	}

	inferredSpan := findSpanByName(t, sink.spans, "aws.step_functions")
	if inferredSpan.TraceID != 5656 || inferredSpan.ParentID != 7878 {
		t.Fatalf("expected Step Functions upstream context, got trace=%d parent=%d", inferredSpan.TraceID, inferredSpan.ParentID)
	}
}

func TestProcessorCleanupKeepsOldPendingStateWithinCapacity(t *testing.T) {
	p := NewProcessor(testSink{}, false)
	now := time.Now()
	old := now.Add(-24 * time.Hour)

	p.contexts["old"] = &InvocationContext{RequestID: "old", CreatedAt: old, UpdatedAt: old}
	p.contexts["new"] = &InvocationContext{RequestID: "new", CreatedAt: now, UpdatedAt: now}
	p.contextOrder = []string{"old", "new"}
	p.pendingInvokeIDs = []pendingRequestID{
		{id: "old", createdAt: old},
		{id: "new", createdAt: now},
	}
	p.pendingRuntimeIDs = []pendingRequestID{
		{id: "old-runtime", createdAt: old},
		{id: "new-runtime", createdAt: now},
	}
	p.pendingStarts = []*InvocationContext{
		{RequestID: "old-start", CreatedAt: old, UpdatedAt: old},
		{RequestID: "new-start", CreatedAt: now, UpdatedAt: now},
	}
	p.pendingEnds = []pendingUniversalEnd{
		{createdAt: old},
		{createdAt: now},
	}
	p.pendingTracer[tracerPlaceholderKey{traceID: 1, invocationSpanID: 1}] = &tracerPlaceholder{createdAt: old}
	p.pendingTracer[tracerPlaceholderKey{traceID: 2, invocationSpanID: 2}] = &tracerPlaceholder{createdAt: now}

	p.cleanupLocked()

	if _, ok := p.contexts["old"]; !ok {
		t.Fatal("old context should be retained while capacity is available")
	}
	if _, ok := p.contexts["new"]; !ok {
		t.Fatal("new context was removed")
	}
	if len(p.pendingInvokeIDs) != 2 || p.pendingInvokeIDs[0].id != "old" {
		t.Fatalf("unexpected pending invoke IDs: %#v", p.pendingInvokeIDs)
	}
	if len(p.pendingRuntimeIDs) != 2 || p.pendingRuntimeIDs[0].id != "old-runtime" {
		t.Fatalf("unexpected pending runtime IDs: %#v", p.pendingRuntimeIDs)
	}
	if len(p.pendingStarts) != 2 || p.pendingStarts[0].RequestID != "old-start" {
		t.Fatalf("unexpected pending starts: %#v", p.pendingStarts)
	}
	if len(p.pendingEnds) != 2 {
		t.Fatalf("unexpected pending ends: %#v", p.pendingEnds)
	}
	if _, ok := p.pendingTracer[tracerPlaceholderKey{traceID: 1, invocationSpanID: 1}]; !ok {
		t.Fatal("old pending tracer should be retained while capacity is available")
	}
	if _, ok := p.pendingTracer[tracerPlaceholderKey{traceID: 2, invocationSpanID: 2}]; !ok {
		t.Fatal("new pending tracer was removed")
	}
}

func TestProcessorCleanupCapsPendingState(t *testing.T) {
	p := NewProcessor(testSink{}, false)
	now := time.Now()

	for i := 0; i < processorPendingItemCapacity+2; i++ {
		p.pendingInvokeIDs = append(p.pendingInvokeIDs, pendingRequestID{
			id:        fmt.Sprintf("invoke-%d", i),
			createdAt: now.Add(time.Duration(i) * time.Millisecond),
		})
		p.pendingRuntimeIDs = append(p.pendingRuntimeIDs, pendingRequestID{
			id:        fmt.Sprintf("runtime-%d", i),
			createdAt: now.Add(time.Duration(i) * time.Millisecond),
		})
		p.pendingStarts = append(p.pendingStarts, &InvocationContext{
			RequestID: fmt.Sprintf("start-%d", i),
			CreatedAt: now.Add(time.Duration(i) * time.Millisecond),
			UpdatedAt: now.Add(time.Duration(i) * time.Millisecond),
		})
		p.pendingEnds = append(p.pendingEnds, pendingUniversalEnd{
			createdAt: now.Add(time.Duration(i) * time.Millisecond),
		})
	}

	p.cleanupLocked()

	if len(p.pendingInvokeIDs) != processorPendingItemCapacity {
		t.Fatalf("expected capped pending invoke IDs, got %d", len(p.pendingInvokeIDs))
	}
	if p.pendingInvokeIDs[0].id != "invoke-2" {
		t.Fatalf("expected oldest pending invoke IDs to be dropped, got %s", p.pendingInvokeIDs[0].id)
	}
	if len(p.pendingRuntimeIDs) != processorPendingItemCapacity {
		t.Fatalf("expected capped pending runtime IDs, got %d", len(p.pendingRuntimeIDs))
	}
	if p.pendingRuntimeIDs[0].id != "runtime-2" {
		t.Fatalf("expected oldest pending runtime IDs to be dropped, got %s", p.pendingRuntimeIDs[0].id)
	}
	if len(p.pendingStarts) != processorPendingItemCapacity {
		t.Fatalf("expected capped pending starts, got %d", len(p.pendingStarts))
	}
	if p.pendingStarts[0].RequestID != "start-2" {
		t.Fatalf("expected oldest pending starts to be dropped, got %s", p.pendingStarts[0].RequestID)
	}
	if len(p.pendingEnds) != processorPendingItemCapacity {
		t.Fatalf("expected capped pending ends, got %d", len(p.pendingEnds))
	}
}

func TestProcessorCleanupCapsContextsByFIFOOrder(t *testing.T) {
	p := NewProcessor(testSink{}, false)

	for i := 0; i < processorContextCapacity+2; i++ {
		requestID := fmt.Sprintf("request-%d", i)
		p.contexts[requestID] = &InvocationContext{RequestID: requestID}
		p.contextOrder = append(p.contextOrder, requestID)
	}

	p.cleanupLocked()

	if len(p.contexts) != processorContextCapacity {
		t.Fatalf("expected capped contexts, got %d", len(p.contexts))
	}
	if _, ok := p.contexts["request-0"]; ok {
		t.Fatal("expected oldest context request-0 to be dropped")
	}
	if _, ok := p.contexts["request-1"]; ok {
		t.Fatal("expected oldest context request-1 to be dropped")
	}
	if _, ok := p.contexts["request-2"]; !ok {
		t.Fatal("expected request-2 to remain")
	}
}

func countSpanByName(spans []Span, name string) int {
	count := 0
	for _, span := range spans {
		if span.Name == name {
			count++
		}
	}
	return count
}

func findSpanByName(t *testing.T, spans []Span, name string) Span {
	t.Helper()
	for _, span := range spans {
		if span.Name == name {
			return span
		}
	}
	t.Fatalf("span %q not found in %#v", name, spans)
	return Span{}
}
