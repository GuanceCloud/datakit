// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2024-present Guance, Inc.

package awslambda

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/GuanceCloud/cliutils/point"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/goroutine"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	lambdaextsrv "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/awslambda/extension"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/awslambda/lambdaapi/telemetry"
	lambdatrace "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/awslambda/trace"
	itrace "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/trace"
)

var traceSnapshotMu sync.Mutex

func TestMain(m *testing.M) {
	_ = os.Remove(filepath.Join(".", "test", "test.output", "input-tracing-points.ndjson"))
	os.Exit(m.Run())
}

func TestInputTracingOnDemand(t *testing.T) {
	ipt, feeder := newTraceTestInput(false)
	server := startLifecycleServerForTest(t, ipt)
	defer server.Close()

	fixture := mustReadTestFile(t, filepath.Join("test", "fixtures", "api_gateway_rest.json"))

	runInvocationFlow(t, ipt, server.Addr, "req-on-demand-1", fixture, false, 20*time.Millisecond, 120*time.Millisecond)
	runInvocationFlow(t, ipt, server.Addr, "req-on-demand-2", fixture, false, 0, 110*time.Millisecond)

	pts, err := feeder.NPoints(5, time.Second)
	if err != nil {
		t.Fatalf("wait tracing points failed: %v", err)
	}

	assertTraceContains(t, pts, "aws.lambda")
	assertTraceContains(t, pts, "aws.lambda.cold_start")
	assertTraceContains(t, pts, "aws.apigateway")

	coldStartCount := countByOperation(pts, "aws.lambda.cold_start")
	if coldStartCount != 1 {
		t.Fatalf("expected exactly one cold start point, got %d", coldStartCount)
	}

	lambdaCount := countByOperation(pts, "aws.lambda")
	if lambdaCount < 2 {
		t.Fatalf("expected at least two lambda points, got %d", lambdaCount)
	}

	writeTraceSnapshot(t, t.Name(), pts)
}

func TestInputTracingManagedInstance(t *testing.T) {
	ipt, feeder := newTraceTestInput(true)
	server := startLifecycleServerForTest(t, ipt)
	defer server.Close()

	fixture := mustReadTestFile(t, filepath.Join("test", "fixtures", "api_gateway_rest.json"))

	var wg sync.WaitGroup
	requestIDs := []string{"req-managed-1", "req-managed-2"}
	for _, requestID := range requestIDs {
		wg.Add(1)
		go func(id string) {
			defer wg.Done()
			runInvocationFlow(t, ipt, server.Addr, id, fixture, true, 0, 130*time.Millisecond)
		}(requestID)
	}
	wg.Wait()

	pts, err := feeder.NPoints(4, time.Second)
	if err != nil {
		t.Fatalf("wait tracing points failed: %v", err)
	}

	assertTraceContains(t, pts, "aws.lambda")
	assertTraceContains(t, pts, "aws.apigateway")
	if countByOperation(pts, "aws.lambda.cold_start") != 0 {
		t.Fatalf("managed instance should not emit cold start span")
	}
	if countManagedLambdaPoints(pts) < 2 {
		t.Fatalf("expected at least two managed lambda points")
	}

	writeTraceSnapshot(t, t.Name(), pts)
}

func TestInputTracingUsesDatadogLambdaNaming(t *testing.T) {
	t.Setenv("AWS_LAMBDA_FUNCTION_NAME", "lambda-go-datadog-zhengb")
	t.Setenv("AWS_LAMBDA_FUNCTION_VERSION", "$LATEST")
	t.Setenv("DD_TRACE_AWS_SERVICE_REPRESENTATION_ENABLED", "true")

	ipt, feeder := newTraceTestInput(false)
	server := startLifecycleServerForTest(t, ipt)
	defer server.Close()

	fixture := mustReadTestFile(t, filepath.Join("test", "fixtures", "api_gateway_rest.json"))
	runInvocationFlow(t, ipt, server.Addr, "req-naming-1", fixture, false, 0, 100*time.Millisecond)

	pts, err := feeder.NPoints(2, time.Second)
	if err != nil {
		t.Fatalf("wait tracing points failed: %v", err)
	}

	service, resource := findOperationServiceAndResource(t, pts, "aws.lambda")
	if service != "lambda-go-datadog-zhengb" {
		t.Fatalf("expected lambda service to match function name, got %q", service)
	}
	if resource != "lambda-go-datadog-zhengb" {
		t.Fatalf("expected lambda resource to match function name, got %q", resource)
	}
}

func newTraceTestInput(managed bool) (*Input, *dkio.MockedFeeder) {
	feeder := dkio.NewMockedFeeder()
	ipt := &Input{
		EnableMetricCollection: true,
		EnableLogCollection:    true,
		feeder:                 feeder,
		tags:                   map[string]string{},
		g:                      goroutine.G("awslambda-test"),
		runtimeDoneChan:        make(chan string, 32),
	}
	if managed {
		ipt.tags[LambdaInitializationType] = "lambda-managed-instances"
	}
	ipt.traceProcessor = lambdatrace.NewProcessor(lambdatrace.NewPointSink(inputName, ipt.feeder, ipt.tags), managed)
	return ipt, feeder
}

func TestRuntimeDoneRequestIDs(t *testing.T) {
	events := []*telemetry.Event{
		{Record: &telemetry.PlatformStart{RequestID: "start-1"}},
		{Record: &telemetry.PlatformRuntimeDone{RequestID: "runtime-1"}},
		{Record: &telemetry.PlatformRuntimeDone{RequestID: ""}},
		{Record: &telemetry.PlatformRuntimeDone{RequestID: "runtime-2"}},
	}

	requestIDs := runtimeDoneRequestIDs(events)
	if len(requestIDs) != 2 {
		t.Fatalf("expected 2 runtime done request IDs, got %d", len(requestIDs))
	}
	if requestIDs[0] != "runtime-1" || requestIDs[1] != "runtime-2" {
		t.Fatalf("unexpected runtime done request IDs: %#v", requestIDs)
	}
}

func startLifecycleServerForTest(t *testing.T, ipt *Input) *http.Server {
	t.Helper()
	if !canBind("127.0.0.1:0") {
		t.Skip("sandbox does not allow binding local TCP sockets")
	}
	server, err := lambdaextsrv.StartLifecycleServer("127.0.0.1:0", ipt.traceProcessor)
	if err != nil {
		t.Fatalf("start lifecycle server failed: %v", err)
	}
	return server
}

func runInvocationFlow(t *testing.T, ipt *Input, addr, requestID string, payload []byte, managed bool, initDuration, reportDuration time.Duration) {
	t.Helper()
	if initDuration > 0 {
		ipt.traceProcessor.OnPlatformInitReport(float64(initDuration) / float64(time.Millisecond))
	}
	start := time.Now()

	postLifecycle(t, "http://"+addr+"/lambda/start-invocation", requestID, payload, nil)
	ipt.traceProcessor.OnPlatformStart(requestID, start)
	postLifecycle(t, "http://"+addr+"/lambda/end-invocation", requestID, []byte(`"ok"`), nil)
	if !managed {
		ipt.traceProcessor.OnPlatformRuntimeDone(requestID, float64(reportDuration)/float64(time.Millisecond), "success")
	}
	if err := ipt.traceProcessor.OnPlatformReport(requestID, float64(reportDuration)/float64(time.Millisecond), "success"); err != nil {
		t.Fatalf("platform report failed: %v", err)
	}
}

func postLifecycle(t *testing.T, url, requestID string, payload []byte, headers map[string]string) {
	t.Helper()
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		t.Fatalf("new request failed: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Lambda-Runtime-Aws-Request-Id", requestID)
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("do lifecycle request failed: %v", err)
	}
	defer resp.Body.Close() //nolint:errcheck
}

func mustReadTestFile(t *testing.T, rel string) []byte {
	t.Helper()
	body, err := os.ReadFile(filepath.Join(".", rel))
	if err != nil {
		t.Fatalf("read fixture failed: %v", err)
	}
	return body
}

func assertTraceContains(t *testing.T, pts []*point.Point, operation string) {
	t.Helper()
	for _, pt := range pts {
		if pt.Get(itrace.TagOperation) == operation {
			return
		}
	}
	t.Fatalf("expected operation %q in tracing points", operation)
}

func countByOperation(pts []*point.Point, operation string) int {
	count := 0
	for _, pt := range pts {
		if pt.Get(itrace.TagOperation) == operation {
			count++
		}
	}
	return count
}

func countManagedLambdaPoints(pts []*point.Point) int {
	count := 0
	for _, pt := range pts {
		if pt.Get(itrace.TagOperation) == "aws.lambda" && strings.Contains(pt.Pretty(), "lambda-managed-instances") {
			count++
		}
	}
	return count
}

func findOperationServiceAndResource(t *testing.T, pts []*point.Point, operation string) (string, string) {
	t.Helper()
	for _, pt := range pts {
		if pt.Get(itrace.TagOperation) == operation {
			service, _ := pt.Get(itrace.TagService).(string)
			resource, _ := pt.Get(itrace.FieldResource).(string)
			return service, resource
		}
	}
	t.Fatalf("expected operation %q in tracing points", operation)
	return "", ""
}

func canBind(addr string) bool {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return false
	}
	_ = ln.Close()
	return true
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

func writeTraceSnapshot(t *testing.T, testName string, pts []*point.Point) {
	t.Helper()

	traceSnapshotMu.Lock()
	defer traceSnapshotMu.Unlock()

	tmpDir := filepath.Join(".", "test", "test.output")
	if err := os.MkdirAll(tmpDir, 0o755); err != nil {
		t.Fatalf("create trace snapshot dir failed: %v", err)
	}

	filePath := filepath.Join(tmpDir, "input-tracing-points.ndjson")

	line := pointSnapshotLine{
		Timestamp: time.Now().Format(time.RFC3339Nano),
		TestName:  testName,
		Points:    make([]pointSnapshotRecord, 0, len(pts)),
	}
	for _, pt := range pts {
		record := pointSnapshotRecord{
			TraceID:    stringifyPointValue(pt.Get(itrace.FieldTraceID)),
			SpanID:     stringifyPointValue(pt.Get(itrace.FieldSpanid)),
			ParentID:   stringifyPointValue(pt.Get(itrace.FieldParentID)),
			Operation:  pt.GetTag(itrace.TagOperation),
			Service:    pt.GetTag(itrace.TagService),
			Resource:   stringifyPointValue(pt.Get(itrace.FieldResource)),
			Start:      int64Value(pt.Get(itrace.FieldStart)),
			Duration:   int64Value(pt.Get(itrace.FieldDuration)),
			Status:     pt.GetTag(itrace.TagSpanStatus),
			SpanType:   pt.GetTag(itrace.TagSpanType),
			Source:     pt.GetTag(itrace.TagSource),
			SourceType: pt.GetTag(itrace.TagSourceType),
			RequestID:  pt.GetTag("request_id"),
			Trigger:    pt.GetTag("trigger"),
			ColdStart:  pt.GetTag("cold_start") == "true",
			InitType:   pt.GetTag(LambdaInitializationType),
		}
		line.Points = append(line.Points, record)
	}

	fd, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		t.Fatalf("open trace snapshot file failed: %v", err)
	}
	defer fd.Close() //nolint:errcheck

	encoded, err := json.Marshal(line)
	if err != nil {
		t.Fatalf("marshal trace snapshot failed: %v", err)
	}
	if _, err := fd.Write(append(encoded, '\n')); err != nil {
		t.Fatalf("write trace snapshot failed: %v", err)
	}
}

func stringifyPointValue(v any) string {
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

func int64Value(v any) int64 {
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
