// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2024-present Guance, Inc.

package awslambda_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/GuanceCloud/cliutils/point"

	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	awslambda "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/awslambda"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/ddtrace"
	itrace "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/trace"
)

var handlerSnapshotMu sync.Mutex

func TestMain(m *testing.M) {
	_ = os.Remove(filepath.Join(".", "test.output", "input-ddtrace-points.ndjson"))
	os.Exit(m.Run())
}

func TestInputTracingWithDDTraceHandlerSpanOnDemand(t *testing.T) {
	if !awslambda.CanBindForTest("127.0.0.1:0") {
		t.Skip("sandbox does not allow binding local TCP sockets")
	}

	feeder := dkio.NewMockedFeeder()
	awsInput := awslambda.NewTracingTestInput(feeder, false)
	lifecycleServer, err := awslambda.StartLifecycleServerForTest(awsInput)
	if err != nil {
		t.Fatalf("start lifecycle server failed: %v", err)
	}
	defer lifecycleServer.Close()

	ddInput := ddtrace.NewTestInput(feeder)
	restoreAfterGather := ddtrace.SetAfterGatherForTest(feeder)
	defer restoreAfterGather()

	ddServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ddInput.HandleTracesForTest(w, r)
	}))
	defer ddServer.Close()

	fixture := mustReadFixture(t, "api_gateway_rest.json")

	awsInput.TraceProcessorForTest().OnPlatformInitReport(25)
	traceID, parentID := postStartInvocation(t, "http://"+lifecycleServer.Addr+"/lambda/start-invocation", "req-handler-1", fixture)
	start := time.Now()
	awsInput.TraceProcessorForTest().OnPlatformStart("req-handler-1", start)

	sendDDTracePayload(t, ddServer.URL+"/v0.4/traces", buildHandlerTrace(traceID, parentID, start))

	postEndInvocation(t, "http://"+lifecycleServer.Addr+"/lambda/end-invocation", "req-handler-1", []byte(`"ok"`))
	awsInput.TraceProcessorForTest().OnPlatformRuntimeDone("req-handler-1", 120, "success")
	if err := awsInput.TraceProcessorForTest().OnPlatformReport("req-handler-1", 120, "success"); err != nil {
		t.Fatalf("platform report failed: %v", err)
	}

	pts, err := feeder.NPoints(5, time.Second)
	if err != nil {
		t.Fatalf("wait tracing points failed: %v", err)
	}

	assertOperation(t, pts, "aws.apigateway")
	assertOperation(t, pts, "aws.lambda")
	assertOperation(t, pts, "aws.lambda.cold_start")
	assertOperation(t, pts, "child.span")
	assertOperation(t, pts, "http.request")
	assertSingleTraceID(t, pts)
	assertParent(t, pts, "child.span", strconv.FormatUint(parentID, 10))
	assertParent(t, pts, "http.request", strconv.FormatUint(parentID, 10))

	writePointSnapshot(t, "TestInputTracingWithDDTraceHandlerSpanOnDemand", pts, "test.output/input-ddtrace-points.ndjson")
}

func TestInputTracingWithDDTraceHandlerSpanManagedInstance(t *testing.T) {
	if !awslambda.CanBindForTest("127.0.0.1:0") {
		t.Skip("sandbox does not allow binding local TCP sockets")
	}

	feeder := dkio.NewMockedFeeder()
	awsInput := awslambda.NewTracingTestInput(feeder, true)
	lifecycleServer, err := awslambda.StartLifecycleServerForTest(awsInput)
	if err != nil {
		t.Fatalf("start lifecycle server failed: %v", err)
	}
	defer lifecycleServer.Close()

	ddInput := ddtrace.NewTestInput(feeder)
	restoreAfterGather := ddtrace.SetAfterGatherForTest(feeder)
	defer restoreAfterGather()

	ddServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ddInput.HandleTracesForTest(w, r)
	}))
	defer ddServer.Close()

	fixture := mustReadFixture(t, "api_gateway_rest.json")

	traceID, parentID := postStartInvocation(t, "http://"+lifecycleServer.Addr+"/lambda/start-invocation", "req-handler-managed-1", fixture)
	start := time.Now()
	awsInput.TraceProcessorForTest().OnPlatformStart("req-handler-managed-1", start)

	sendDDTracePayload(t, ddServer.URL+"/v0.4/traces", buildHandlerTrace(traceID, parentID, start))

	postEndInvocation(t, "http://"+lifecycleServer.Addr+"/lambda/end-invocation", "req-handler-managed-1", []byte(`"ok"`))
	if err := awsInput.TraceProcessorForTest().OnPlatformReport("req-handler-managed-1", 130, "success"); err != nil {
		t.Fatalf("platform report failed: %v", err)
	}

	pts, err := feeder.NPoints(4, time.Second)
	if err != nil {
		t.Fatalf("wait tracing points failed: %v", err)
	}

	assertOperation(t, pts, "aws.apigateway")
	assertOperation(t, pts, "aws.lambda")
	assertOperation(t, pts, "child.span")
	assertOperation(t, pts, "http.request")
	if countOperation(pts, "aws.lambda.cold_start") != 0 {
		t.Fatalf("managed instance should not emit cold start span")
	}
	assertSingleTraceID(t, pts)
	assertParent(t, pts, "child.span", strconv.FormatUint(parentID, 10))
	assertParent(t, pts, "http.request", strconv.FormatUint(parentID, 10))

	writePointSnapshot(t, "TestInputTracingWithDDTraceHandlerSpanManagedInstance", pts, "test.output/input-ddtrace-points.ndjson")
}

func TestInputTracingWithDDTraceServerlessPlaceholderSpan(t *testing.T) {
	if !awslambda.CanBindForTest("127.0.0.1:0") {
		t.Skip("sandbox does not allow binding local TCP sockets")
	}
	t.Setenv(awslambda.EnvLambdaFunctionName, "test-lambda")

	feeder := dkio.NewMockedFeeder()
	awsInput := awslambda.NewTracingTestInput(feeder, false)
	lifecycleServer, err := awslambda.StartLifecycleServerForTest(awsInput)
	if err != nil {
		t.Fatalf("start lifecycle server failed: %v", err)
	}
	defer lifecycleServer.Close()

	ddInput := ddtrace.NewTestInput(feeder)
	restoreAfterGather := ddtrace.SetAfterGatherForTest(feeder)
	defer restoreAfterGather()

	ddServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ddInput.HandleTracesForTest(w, r)
	}))
	defer ddServer.Close()

	fixture := mustReadFixture(t, "api_gateway_rest.json")

	awsInput.TraceProcessorForTest().OnPlatformInitReport(25)
	requestID := "req-handler-placeholder-1"
	traceID, invocationSpanID := postStartInvocation(t, "http://"+lifecycleServer.Addr+"/lambda/start-invocation", requestID, fixture)
	start := time.Now()
	awsInput.TraceProcessorForTest().OnPlatformStart(requestID, start)

	sendDDTracePayload(t, ddServer.URL+"/v0.4/traces", buildHandlerTraceWithPlaceholder(traceID, invocationSpanID, requestID, start))

	postEndInvocation(t, "http://"+lifecycleServer.Addr+"/lambda/end-invocation", requestID, []byte(`"ok"`))
	awsInput.TraceProcessorForTest().OnPlatformRuntimeDone(requestID, 120, "success")
	if err := awsInput.TraceProcessorForTest().OnPlatformReport(requestID, 120, "success"); err != nil {
		t.Fatalf("platform report failed: %v", err)
	}

	pts, err := feeder.NPoints(5, time.Second)
	if err != nil {
		t.Fatalf("wait tracing points failed: %v", err)
	}

	assertOperation(t, pts, "aws.apigateway")
	assertOperation(t, pts, "aws.lambda")
	assertOperation(t, pts, "aws.lambda.cold_start")
	assertOperation(t, pts, "http.request")
	assertOperation(t, pts, "child.span")
	assertSingleTraceID(t, pts)
	assertOperationCount(t, pts, "aws.lambda", 1)
	assertResourceAbsent(t, pts, "dd-tracer-serverless-span")
	assertParent(t, pts, "http.request", strconv.FormatUint(invocationSpanID, 10))
	assertParent(t, pts, "child.span", strconv.FormatUint(invocationSpanID, 10))
	assertInvocationMessageContains(t, pts, `"datadog_lambda":"v2.6.0-dev.1"`)
	assertInvocationMessageContains(t, pts, `"language":"go"`)
	assertInvocationMessageContains(t, pts, `"_sampling_priority_v1":1`)

	writePointSnapshot(t, "TestInputTracingWithDDTraceServerlessPlaceholderSpan", pts, "test.output/input-ddtrace-points.ndjson")
}

func TestInputTracingWithDDTraceDuplicatePayloadDedupedInLambda(t *testing.T) {
	if !awslambda.CanBindForTest("127.0.0.1:0") {
		t.Skip("sandbox does not allow binding local TCP sockets")
	}
	t.Setenv(awslambda.EnvLambdaFunctionName, "test-lambda")

	feeder := dkio.NewMockedFeeder()
	awsInput := awslambda.NewTracingTestInput(feeder, false)
	lifecycleServer, err := awslambda.StartLifecycleServerForTest(awsInput)
	if err != nil {
		t.Fatalf("start lifecycle server failed: %v", err)
	}
	defer lifecycleServer.Close()

	ddInput := ddtrace.NewTestInput(feeder)
	restoreAfterGather := ddtrace.SetAfterGatherForTest(feeder)
	defer restoreAfterGather()

	ddServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ddInput.HandleTracesForTest(w, r)
	}))
	defer ddServer.Close()

	fixture := mustReadFixture(t, "api_gateway_rest.json")

	awsInput.TraceProcessorForTest().OnPlatformInitReport(25)
	requestID := "req-handler-dedup-1"
	traceID, invocationSpanID := postStartInvocation(t, "http://"+lifecycleServer.Addr+"/lambda/start-invocation", requestID, fixture)
	start := time.Now()
	awsInput.TraceProcessorForTest().OnPlatformStart(requestID, start)

	tracePayload := buildHandlerTraceWithPlaceholder(traceID, invocationSpanID, requestID, start)
	sendDDTracePayload(t, ddServer.URL+"/v0.4/traces", tracePayload)
	sendDDTracePayload(t, ddServer.URL+"/v0.4/traces", tracePayload)

	postEndInvocation(t, "http://"+lifecycleServer.Addr+"/lambda/end-invocation", requestID, []byte(`"ok"`))
	awsInput.TraceProcessorForTest().OnPlatformRuntimeDone(requestID, 120, "success")
	if err := awsInput.TraceProcessorForTest().OnPlatformReport(requestID, 120, "success"); err != nil {
		t.Fatalf("platform report failed: %v", err)
	}

	pts, err := feeder.NPoints(5, time.Second)
	if err != nil {
		t.Fatalf("wait tracing points failed: %v", err)
	}

	assertOperationCount(t, pts, "http.request", 1)
	assertOperationCount(t, pts, "child.span", 1)
	assertResourceAbsent(t, pts, "dd-tracer-serverless-span")
	assertParent(t, pts, "http.request", strconv.FormatUint(invocationSpanID, 10))
	assertParent(t, pts, "child.span", strconv.FormatUint(invocationSpanID, 10))
}

func buildHandlerTrace(traceID, parentID uint64, start time.Time) ddtrace.DDTraces {
	httpSpan := &ddtrace.DDSpan{
		Service:  "http-client",
		Name:     "http.request",
		Resource: "GET https://www.datadoghq.com",
		TraceID:  traceID,
		SpanID:   parentID + 1,
		ParentID: parentID,
		Start:    start.Add(5 * time.Millisecond).UnixNano(),
		Duration: int64(300 * time.Microsecond),
		Meta: map[string]string{
			"http.method":      "GET",
			"http.url":         "https://www.datadoghq.com",
			"http.status_code": "200",
		},
		Metrics: map[string]float64{},
		Type:    "http",
	}
	childSpan := &ddtrace.DDSpan{
		Service:  "go-lambda-demo",
		Name:     "child.span",
		Resource: "child.span",
		TraceID:  traceID,
		SpanID:   parentID + 2,
		ParentID: parentID,
		Start:    start.Add(10 * time.Millisecond).UnixNano(),
		Duration: int64(100 * time.Millisecond),
		Meta:     map[string]string{},
		Metrics:  map[string]float64{},
		Type:     "custom",
	}
	return ddtrace.DDTraces{ddtrace.DDTrace{httpSpan, childSpan}}
}

func buildHandlerTraceWithPlaceholder(traceID, invocationSpanID uint64, requestID string, start time.Time) ddtrace.DDTraces {
	placeholderSpanID := invocationSpanID + 100
	placeholderSpan := &ddtrace.DDSpan{
		Service:  "aws.lambda",
		Name:     "aws.lambda",
		Resource: "dd-tracer-serverless-span",
		TraceID:  traceID,
		SpanID:   placeholderSpanID,
		ParentID: invocationSpanID,
		Start:    start.UnixNano(),
		Duration: int64(120 * time.Millisecond),
		Meta: map[string]string{
			"datadog_lambda": "v2.6.0-dev.1",
			"language":       "go",
			"request_id":     requestID,
			"resource_names": "test-lambda",
		},
		Metrics: map[string]float64{
			"_sampling_priority_v1": 1,
			"_dd.top_level":         1,
		},
		Type: "serverless",
	}
	httpSpan := &ddtrace.DDSpan{
		Service:  "http-client",
		Name:     "http.request",
		Resource: "GET https://www.datadoghq.com",
		TraceID:  traceID,
		SpanID:   placeholderSpanID + 1,
		ParentID: placeholderSpanID,
		Start:    start.Add(5 * time.Millisecond).UnixNano(),
		Duration: int64(300 * time.Microsecond),
		Meta: map[string]string{
			"http.method":      "GET",
			"http.url":         "https://www.datadoghq.com",
			"http.status_code": "200",
		},
		Metrics: map[string]float64{},
		Type:    "http",
	}
	childSpan := &ddtrace.DDSpan{
		Service:  "go-lambda-demo",
		Name:     "child.span",
		Resource: "child.span",
		TraceID:  traceID,
		SpanID:   placeholderSpanID + 2,
		ParentID: placeholderSpanID,
		Start:    start.Add(10 * time.Millisecond).UnixNano(),
		Duration: int64(100 * time.Millisecond),
		Meta:     map[string]string{},
		Metrics:  map[string]float64{},
		Type:     "custom",
	}
	return ddtrace.DDTraces{ddtrace.DDTrace{placeholderSpan, httpSpan, childSpan}}
}

func sendDDTracePayload(t *testing.T, url string, traces ddtrace.DDTraces) {
	t.Helper()
	body, err := json.Marshal(traces)
	if err != nil {
		t.Fatalf("marshal ddtrace payload failed: %v", err)
	}
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		t.Fatalf("new ddtrace request failed: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Datadog-Trace-Count", "1")
	req.Header.Set("Content-Length", strconv.Itoa(len(body)))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("send ddtrace request failed: %v", err)
	}
	defer resp.Body.Close() //nolint:errcheck
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected ddtrace response status: %d", resp.StatusCode)
	}
}

func postStartInvocation(t *testing.T, url, requestID string, payload []byte) (uint64, uint64) {
	t.Helper()
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		t.Fatalf("new start request failed: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Lambda-Runtime-Aws-Request-Id", requestID)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("do start request failed: %v", err)
	}
	defer resp.Body.Close() //nolint:errcheck

	traceID, err := strconv.ParseUint(resp.Header.Get("x-datadog-trace-id"), 10, 64)
	if err != nil {
		t.Fatalf("parse trace id failed: %v", err)
	}
	parentID, err := strconv.ParseUint(resp.Header.Get("x-datadog-parent-id"), 10, 64)
	if err != nil {
		t.Fatalf("parse parent id failed: %v", err)
	}
	return traceID, parentID
}

func postEndInvocation(t *testing.T, url, requestID string, payload []byte) {
	t.Helper()
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		t.Fatalf("new end request failed: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Lambda-Runtime-Aws-Request-Id", requestID)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("do end request failed: %v", err)
	}
	defer resp.Body.Close() //nolint:errcheck
}

func mustReadFixture(t *testing.T, name string) []byte {
	t.Helper()
	body, err := os.ReadFile(filepath.Join(".", "fixtures", name))
	if err != nil {
		t.Fatalf("read fixture failed: %v", err)
	}
	return body
}

func assertOperation(t *testing.T, pts []*point.Point, operation string) {
	t.Helper()
	for _, pt := range pts {
		if pt.Get(itrace.TagOperation) == operation {
			return
		}
	}
	t.Fatalf("expected operation %q in tracing points", operation)
}

func countOperation(pts []*point.Point, operation string) int {
	count := 0
	for _, pt := range pts {
		if pt.Get(itrace.TagOperation) == operation {
			count++
		}
	}
	return count
}

func assertOperationCount(t *testing.T, pts []*point.Point, operation string, expected int) {
	t.Helper()
	if got := countOperation(pts, operation); got != expected {
		t.Fatalf("expected operation %q count %d, got %d", operation, expected, got)
	}
}

func assertResourceAbsent(t *testing.T, pts []*point.Point, resource string) {
	t.Helper()
	for _, pt := range pts {
		if stringify(pt.Get(itrace.FieldResource)) == resource {
			t.Fatalf("expected resource %q to be absent", resource)
		}
	}
}

func assertSingleTraceID(t *testing.T, pts []*point.Point) {
	t.Helper()
	seen := map[string]struct{}{}
	for _, pt := range pts {
		if traceID, ok := pt.Get(itrace.FieldTraceID).(string); ok && traceID != "" {
			seen[traceID] = struct{}{}
		}
	}
	if len(seen) != 1 {
		t.Fatalf("expected exactly one trace id, got %d", len(seen))
	}
}

func assertParent(t *testing.T, pts []*point.Point, operation, expectedParent string) {
	t.Helper()
	for _, pt := range pts {
		if pt.Get(itrace.TagOperation) == operation {
			if got, _ := pt.Get(itrace.FieldParentID).(string); got != expectedParent {
				t.Fatalf("operation %s expected parent %s, got %s", operation, expectedParent, got)
			}
			return
		}
	}
	t.Fatalf("operation %s not found", operation)
}

func assertInvocationMessageContains(t *testing.T, pts []*point.Point, expected string) {
	t.Helper()
	for _, pt := range pts {
		if pt.Get(itrace.TagOperation) != "aws.lambda" {
			continue
		}
		message := stringify(pt.Get(itrace.FieldMessage))
		if !strings.Contains(message, expected) {
			t.Fatalf("expected invocation message to contain %s, got %s", expected, message)
		}
		return
	}
	t.Fatalf("operation aws.lambda not found")
}

type pointSnapshotLine struct {
	Timestamp string                `json:"timestamp"`
	TestName  string                `json:"test"`
	Points    []pointSnapshotRecord `json:"points"`
}

type pointSnapshotRecord struct {
	TraceID    string `json:"trace_id"`
	SpanID     string `json:"span_id"`
	ParentID   string `json:"parent_id"`
	Operation  string `json:"operation"`
	Service    string `json:"service"`
	Resource   string `json:"resource"`
	Start      int64  `json:"start"`
	Duration   int64  `json:"duration"`
	Status     string `json:"status"`
	SpanType   string `json:"span_type"`
	Source     string `json:"source"`
	SourceType string `json:"source_type"`
	RequestID  string `json:"request_id"`
	Trigger    string `json:"trigger"`
	ColdStart  bool   `json:"cold_start"`
	InitType   string `json:"init_type"`
}

func writePointSnapshot(t *testing.T, testName string, pts []*point.Point, relFile string) {
	t.Helper()
	handlerSnapshotMu.Lock()
	defer handlerSnapshotMu.Unlock()

	if err := os.MkdirAll(filepath.Dir(relFile), 0o755); err != nil {
		t.Fatalf("create snapshot dir failed: %v", err)
	}

	line := pointSnapshotLine{
		Timestamp: time.Now().Format(time.RFC3339Nano),
		TestName:  testName,
		Points:    make([]pointSnapshotRecord, 0, len(pts)),
	}
	for _, pt := range pts {
		line.Points = append(line.Points, pointSnapshotRecord{
			TraceID:    stringify(pt.Get(itrace.FieldTraceID)),
			SpanID:     stringify(pt.Get(itrace.FieldSpanid)),
			ParentID:   stringify(pt.Get(itrace.FieldParentID)),
			Operation:  pt.GetTag(itrace.TagOperation),
			Service:    pt.GetTag(itrace.TagService),
			Resource:   stringify(pt.Get(itrace.FieldResource)),
			Start:      asInt64(pt.Get(itrace.FieldStart)),
			Duration:   asInt64(pt.Get(itrace.FieldDuration)),
			Status:     pt.GetTag(itrace.TagSpanStatus),
			SpanType:   pt.GetTag(itrace.TagSpanType),
			Source:     pt.GetTag(itrace.TagSource),
			SourceType: pt.GetTag(itrace.TagSourceType),
			RequestID:  pt.GetTag("request_id"),
			Trigger:    pt.GetTag("trigger"),
			ColdStart:  pt.GetTag("cold_start") == "true",
			InitType:   pt.GetTag(awslambda.LambdaInitializationType),
		})
	}

	fd, err := os.OpenFile(relFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		t.Fatalf("open snapshot failed: %v", err)
	}
	defer fd.Close() //nolint:errcheck

	data, err := json.Marshal(line)
	if err != nil {
		t.Fatalf("marshal snapshot failed: %v", err)
	}
	if _, err := fd.Write(append(data, '\n')); err != nil {
		t.Fatalf("write snapshot failed: %v", err)
	}
}

func stringify(v any) string {
	switch x := v.(type) {
	case nil:
		return ""
	case string:
		return x
	case []byte:
		return string(x)
	default:
		return fmt.Sprint(x)
	}
}

func asInt64(v any) int64 {
	switch x := v.(type) {
	case nil:
		return 0
	case int64:
		return x
	case int:
		return int64(x)
	case uint64:
		return int64(x)
	case float64:
		return int64(x)
	case string:
		n, _ := strconv.ParseInt(x, 10, 64)
		return n
	default:
		return 0
	}
}
