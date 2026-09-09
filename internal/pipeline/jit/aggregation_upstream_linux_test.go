// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build pipeline_jit && linux && (amd64 || arm64)

// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package jit

import (
	"errors"
	"os"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestAggregationPipelineGoCountFlushCorpus(t *testing.T) {
	path := os.Getenv("PLATYPUS_JIT_RUNTIME")
	if path == "" {
		t.Fatal("requires real PLATYPUS_JIT_RUNTIME")
	}
	runner, err := NewRunner(path, "pipeline-go-1.4.3-datakit", 4)
	if err != nil {
		t.Fatal(err)
	}
	defer runner.Close()

	// This is the upstream TestAgg program with a deterministic count trigger.
	// The exact 5s idle interval is covered separately by the event-drain gate.
	const source = `
set_tag("t0","t1111")
set_tag("t1","t1111")
set_tag("t2","t2__")
f0=_
cast(f0,"int")
agg_create("abc",on_interval="5s",on_count=2,const_tags={"a":"b"})
agg_create("def",on_count=2)
agg_metric("abc","f1","avg",["t2","t0"],"f0")
agg_metric("def","f1","avg",["t1","t2"],"f0")
agg_create("abc",on_interval="5s",on_count=2,const_tags={"a":"b"},category="logging")
agg_create("def",on_count=2,category="logging")
agg_metric("abc","f1","avg",["t2","t0"],"f0",category="logging")
agg_metric("def","f1","avg",["t1","t2"],"f0","logging")
`
	check := runner.Check(source)
	if check.Route != RouteJITNative || check.Capabilities.Backend != ExecutionBackendMachineCode ||
		check.Capabilities.RequiredHostFlags != 0 || len(check.Capabilities.HostCalls) != 0 {
		t.Fatalf("route=%+v", check)
	}
	points := []Point{
		{Version: 1, Category: "logging", Measurement: "test", Fields: map[string]any{"message": "1"}},
		{Version: 1, Category: "logging", Measurement: "test", Fields: map[string]any{"message": "2"}},
	}
	input, err := EncodeFlatPoints(points)
	if err != nil {
		t.Fatal(err)
	}
	batch, err := runner.Process(source, input)
	if err != nil || len(batch.Records) != 2 {
		t.Fatalf("records=%d error=%v", len(batch.Records), err)
	}
	if len(batch.Records[0].Emitted) != 0 || len(batch.Records[1].Emitted) != 4 {
		t.Fatalf("aggregate emission counts=%d/%d", len(batch.Records[0].Emitted), len(batch.Records[1].Emitted))
	}
	type aggregatePoint struct {
		Category    string         `json:"category"`
		Measurement string         `json:"measurement"`
		Tags        map[string]any `json:"tags"`
		Fields      map[string]any `json:"fields"`
	}
	seen := map[string]aggregatePoint{}
	for index, payload := range batch.Records[1].Emitted {
		var point aggregatePoint
		if err := decodeEmittedTestPayload(payload, &point); err != nil {
			t.Fatalf("aggregate %d: %v", index, err)
		}
		seen[point.Category+"/"+point.Measurement] = point
	}
	for _, category := range []string{"metric", "logging"} {
		for _, bucket := range []string{"abc", "def"} {
			point, ok := seen[category+"/"+bucket]
			if !ok || point.Fields["f1"] != float64(1.5) {
				t.Fatalf("%s/%s=%#v", category, bucket, point)
			}
			if bucket == "abc" {
				if point.Tags["t0"] != "t1111" || point.Tags["a"] != "b" {
					t.Fatalf("%s/%s tags=%#v", category, bucket, point.Tags)
				}
			} else if point.Tags["t1"] != "t1111" {
				t.Fatalf("%s/%s tags=%#v", category, bucket, point.Tags)
			}
		}
	}
}

func TestAggregationPreservesRawGroupingTagBytes(t *testing.T) {
	path := os.Getenv("PLATYPUS_JIT_RUNTIME")
	if path == "" {
		t.Fatal("requires real PLATYPUS_JIT_RUNTIME")
	}
	runner, err := NewRunner(path, "pipeline-go-1.4.3-datakit", 4)
	if err != nil {
		t.Fatal(err)
	}
	defer runner.Close()
	const source = `f0=_; cast(f0,"int"); agg_create("raw",on_count=2); agg_metric("raw","value","avg",["host"],"f0")`
	points := []Point{
		{Version: 1, Category: "logging", Measurement: "test", Tags: map[string]string{"host": "A\xff\x00B"}, Fields: map[string]any{"message": "1"}},
		{Version: 1, Category: "logging", Measurement: "test", Tags: map[string]string{"host": "A\xff\x00B"}, Fields: map[string]any{"message": "3"}},
	}
	projection, err := runner.Projection(source)
	if err != nil {
		t.Fatal(err)
	}
	records := make([]FlatPointRecord, len(points))
	for index := range points {
		records[index] = FlatPointRecord{
			Version: 1, Category: "logging", Measurement: "test",
			Entries: []FlatPointEntry{
				{Key: "host", Value: "A\xff\x00B", IsTag: true},
				{Key: "message", Value: points[index].Fields["message"]},
			},
		}
	}
	input, err := EncodeProjectedPointRecords(records, projection)
	if err != nil {
		t.Fatal(err)
	}
	batch, err := runner.Process(source, input)
	if err != nil || len(batch.Records) != 2 {
		t.Fatalf("records=%d error=%v", len(batch.Records), err)
	}
	if len(batch.Records[1].Emitted) != 1 {
		t.Fatalf("emitted=%d batch=%+v", len(batch.Records[1].Emitted), batch)
	}
	emittedPoints, err := DecodeFlatPoints(batch.Records[1].Emitted[0])
	if err != nil || len(emittedPoints) != 1 {
		t.Fatal(err)
	}
	emitted := emittedPoints[0]
	if emitted.Tags["host"] != "A\xff\x00B" || emitted.Fields["value"] != float64(2) {
		t.Fatalf("raw aggregate tag was not preserved: %#v payload=%q", emitted, batch.Records[1].Emitted[0])
	}
}

func TestAggregationNestedValueCallsUseOneRuntimeState(t *testing.T) {
	path := os.Getenv("PLATYPUS_JIT_RUNTIME")
	if path == "" {
		t.Fatal("requires real PLATYPUS_JIT_RUNTIME")
	}
	runner, err := NewRunner(path, "pipeline-go-1.4.3-datakit", 4)
	if err != nil {
		t.Fatal(err)
	}
	defer runner.Close()

	const source = `
set_tag("host","a")
f0=_
cast(f0,"int")
bucket="abc"; interval="1h"; count=2
output="f1"; aggregate="avg"; grouping=["host"]; input="f0"
add_key(created,agg_create(bucket,on_interval=interval,on_count=count)==nil)
add_key(metric,agg_metric(bucket,output,aggregate,grouping,input)==nil)
`
	check := runner.Check(source)
	if check.Route != RouteJITNative || check.Capabilities.Backend != ExecutionBackendMachineCode ||
		check.Capabilities.RequiredHostFlags != 0 {
		t.Fatalf("route=%+v", check)
	}
	input, err := EncodeFlatPoints([]Point{
		{Version: 1, Category: "logging", Measurement: "test", Fields: map[string]any{"message": "1"}},
		{Version: 1, Category: "logging", Measurement: "test", Fields: map[string]any{"message": "2"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	batch, err := runner.Process(source, input)
	if err != nil || len(batch.Records) != 2 {
		t.Fatalf("records=%d error=%v", len(batch.Records), err)
	}
	if len(batch.Records[0].Emitted) != 0 || len(batch.Records[1].Emitted) != 1 {
		t.Fatalf("aggregate emission counts=%d/%d", len(batch.Records[0].Emitted), len(batch.Records[1].Emitted))
	}
	var point struct {
		Fields map[string]any `json:"fields"`
	}
	if err := decodeEmittedTestPayload(batch.Records[1].Emitted[0], &point); err != nil {
		t.Fatal(err)
	}
	if point.Fields["f1"] != float64(1.5) {
		t.Fatalf("aggregate=%#v", point.Fields)
	}
}

func TestAggregationPipelineGoExactUpstreamRetirementFlush(t *testing.T) {
	path := os.Getenv("PLATYPUS_JIT_RUNTIME")
	if path == "" {
		t.Fatal("requires real PLATYPUS_JIT_RUNTIME")
	}
	runner, err := NewRunner(path, "pipeline-go-1.4.3-datakit", 4)
	if err != nil {
		t.Fatal(err)
	}
	events := make(chan []Point, 2)
	if err := runner.SetAggregateEventHandler(func(points []Point) error {
		events <- append([]Point(nil), points...)
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	const source = `
set_tag("t0","t1111")
set_tag("t1","t1111")
set_tag("t2","t2__")
f0=_
cast(f0,"int")
agg_create("abc",on_interval="5s",on_count=0,const_tags={"a":"b"})
agg_create("def")
agg_metric("abc","f1","avg",["t2","t0"],"f0")
agg_metric("def","f1","avg",["t1","t2"],"f0")
agg_create("abc",on_interval="5s",on_count=0,const_tags={"a":"b"},category="logging")
agg_create("def",category="logging")
agg_metric("abc","f1","avg",["t2","t0"],"f0",category="logging")
agg_metric("def","f1","avg",["t1","t2"],"f0","logging")
`
	if check := runner.Check(source); check.Route != RouteJITNative || check.Capabilities.AggregateEvents == nil || !*check.Capabilities.AggregateEvents {
		t.Fatalf("route=%+v", check)
	}
	input, err := EncodeFlatPoints([]Point{
		{Version: 1, Category: "logging", Measurement: "test", Fields: map[string]any{"message": "1"}},
		{Version: 1, Category: "logging", Measurement: "test", Fields: map[string]any{"message": "2"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	batch, err := runner.Process(source, input)
	if err != nil || len(batch.Records) != 2 {
		t.Fatalf("records=%d error=%v", len(batch.Records), err)
	}
	if len(batch.Records[0].Emitted) != 0 || len(batch.Records[1].Emitted) != 0 {
		t.Fatalf("5s upstream interval flushed before retirement: %+v", batch.Records)
	}
	if err := runner.Close(); err != nil {
		t.Fatal(err)
	}
	select {
	case points := <-events:
		if len(points) != 4 {
			t.Fatalf("retirement points=%#v", points)
		}
		seen := map[string]Point{}
		for _, point := range points {
			seen[point.Category+"/"+point.Measurement] = point
		}
		for _, category := range []string{"metric", "logging"} {
			for _, bucket := range []string{"abc", "def"} {
				point, ok := seen[category+"/"+bucket]
				if !ok || point.Fields["f1"] != float64(1.5) {
					t.Fatalf("%s/%s=%#v", category, bucket, point)
				}
			}
		}
	case <-time.After(2 * time.Second):
		t.Fatal("retirement aggregate flush timed out")
	}
}

func TestAggregationIntervalAndRetirementDrainABI(t *testing.T) {
	path := os.Getenv("PLATYPUS_JIT_RUNTIME")
	if path == "" {
		t.Fatal("requires real PLATYPUS_JIT_RUNTIME")
	}
	runner, err := NewRunner(path, "pipeline-go-1.4.3-datakit", 4)
	if err != nil {
		t.Fatal(err)
	}
	defer runner.Close()
	const source = `
set_tag("host","a")
f0=_
cast(f0,"int")
agg_create("abc",on_interval="5s",on_count=0,const_tags={"env":"test"})
agg_metric("abc","f1","avg",["host"],"f0")
`
	lease, err := runner.acquireProgram(source)
	if err != nil {
		t.Fatal(err)
	}
	defer lease.Release()
	program, ok := lease.program.(AggregateEventProgram)
	if !ok {
		t.Fatal("native program lacks aggregate event drain")
	}
	process := func(values ...string) {
		t.Helper()
		points := make([]Point, len(values))
		for i, value := range values {
			points[i] = Point{Version: 1, Category: "logging", Measurement: "source", Fields: map[string]any{"message": value}}
		}
		input, err := EncodeFlatPoints(points)
		if err != nil {
			t.Fatal(err)
		}
		batch, err := program.ProcessIndexed(input)
		if err != nil || len(batch.Records) != len(values) {
			t.Fatalf("records=%d error=%v", len(batch.Records), err)
		}
	}
	assertAverage := func(points []Point, want float64) {
		t.Helper()
		if len(points) != 1 || points[0].Measurement != "abc" || points[0].Category != "metric" ||
			points[0].Fields["f1"] != want || points[0].Tags["host"] != "a" || points[0].Tags["env"] != "test" {
			t.Fatalf("aggregate points=%#v want average=%v", points, want)
		}
	}

	process("1", "2")
	due, err := program.DrainAggregateEvents(time.Now().Add(6*time.Second), false)
	if err != nil {
		t.Fatal(err)
	}
	assertAverage(due, 1.5)
	empty, err := program.DrainAggregateEvents(time.Now().Add(6*time.Second), false)
	if err != nil || len(empty) != 0 {
		t.Fatalf("repeat due drain=%#v error=%v", empty, err)
	}

	process("3", "5")
	forced, err := program.DrainAggregateEvents(time.Now(), true)
	if err != nil {
		t.Fatal(err)
	}
	assertAverage(forced, 4)
	empty, err = program.DrainAggregateEvents(time.Now(), true)
	if err != nil || len(empty) != 0 {
		t.Fatalf("repeat force drain=%#v error=%v", empty, err)
	}
}

func TestAggregationAutomaticIdleAndRunnerCloseFlush(t *testing.T) {
	path := os.Getenv("PLATYPUS_JIT_RUNTIME")
	if path == "" {
		t.Fatal("requires real PLATYPUS_JIT_RUNTIME")
	}
	newRunner := func(interval string) (*Runner, chan []Point, string) {
		t.Helper()
		runner, err := NewRunner(path, "pipeline-go-1.4.3-datakit", 2)
		if err != nil {
			t.Fatal(err)
		}
		events := make(chan []Point, 4)
		if err := runner.SetAggregateEventHandler(func(points []Point) error {
			events <- append([]Point(nil), points...)
			return nil
		}); err != nil {
			t.Fatal(err)
		}
		source := `set_tag("host","a"); f0=_; cast(f0,"int"); ` +
			`agg_create("abc",on_interval="` + interval + `",on_count=0); ` +
			`agg_metric("abc","f1","avg",["host"],"f0")`
		return runner, events, source
	}
	process := func(runner *Runner, source string) {
		t.Helper()
		input, err := EncodeFlatPoints([]Point{
			{Version: 1, Category: "logging", Fields: map[string]any{"message": "1"}},
			{Version: 1, Category: "logging", Fields: map[string]any{"message": "2"}},
		})
		if err != nil {
			t.Fatal(err)
		}
		batch, err := runner.Process(source, input)
		if err != nil || len(batch.Records) != 2 {
			t.Fatalf("records=%d error=%v", len(batch.Records), err)
		}
	}
	assertEvent := func(events <-chan []Point, reason string) {
		t.Helper()
		select {
		case points := <-events:
			if len(points) != 1 || points[0].Fields["f1"] != float64(1.5) {
				t.Fatalf("%s points=%#v", reason, points)
			}
		case <-time.After(4 * time.Second):
			t.Fatalf("%s aggregate event timed out", reason)
		}
	}

	idleRunner, idleEvents, idleSource := newRunner("1s")
	process(idleRunner, idleSource)
	assertEvent(idleEvents, "idle interval")
	if err := idleRunner.Close(); err != nil {
		t.Fatal(err)
	}
	select {
	case duplicate := <-idleEvents:
		t.Fatalf("idle close duplicated flushed aggregate: %#v", duplicate)
	default:
	}

	closeRunner, closeEvents, closeSource := newRunner("1h")
	process(closeRunner, closeSource)
	if err := closeRunner.Close(); err != nil {
		t.Fatal(err)
	}
	assertEvent(closeEvents, "runner close")
}

func TestAggregationUploadFailureRetriesWithoutDuplicate(t *testing.T) {
	path := os.Getenv("PLATYPUS_JIT_RUNTIME")
	if path == "" {
		t.Fatal("requires real PLATYPUS_JIT_RUNTIME")
	}
	runner, err := NewRunner(path, "pipeline-go-1.4.3-datakit", 2)
	if err != nil {
		t.Fatal(err)
	}
	defer runner.Close()
	var attempts atomic.Int32
	delivered := make(chan []Point, 2)
	if err := runner.SetAggregateEventHandler(func(points []Point) error {
		if attempts.Add(1) == 1 {
			return errors.New("temporary upload failure")
		}
		delivered <- append([]Point(nil), points...)
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	const source = `set_tag("host","a"); f0=_; cast(f0,"int"); agg_create("abc",on_interval="1s",on_count=0); agg_metric("abc","f1","avg",["host"],"f0")`
	input, err := EncodeFlatPoints([]Point{
		{Version: 1, Category: "logging", Fields: map[string]any{"message": "1"}},
		{Version: 1, Category: "logging", Fields: map[string]any{"message": "2"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if batch, err := runner.Process(source, input); err != nil || len(batch.Records) != 2 {
		t.Fatalf("batch=%+v error=%v", batch, err)
	}
	select {
	case points := <-delivered:
		if attempts.Load() < 2 || len(points) != 1 || points[0].Fields["f1"] != float64(1.5) {
			t.Fatalf("attempts=%d points=%#v", attempts.Load(), points)
		}
	case <-time.After(5 * time.Second):
		t.Fatalf("pending aggregate was not retried; attempts=%d", attempts.Load())
	}
	select {
	case duplicate := <-delivered:
		t.Fatalf("aggregate delivered twice: %#v", duplicate)
	default:
	}
}

func TestAggregationPersistentUploadFailureKeepsSingleBoundedBatch(t *testing.T) {
	path := os.Getenv("PLATYPUS_JIT_RUNTIME")
	if path == "" {
		t.Fatal("requires real PLATYPUS_JIT_RUNTIME")
	}
	runner, err := NewRunner(path, "pipeline-go-1.4.3-datakit", 2)
	if err != nil {
		t.Fatal(err)
	}
	var attempts atomic.Int32
	lengths := make(chan int, 8)
	if err := runner.SetAggregateEventHandler(func(points []Point) error {
		attempts.Add(1)
		lengths <- len(points)
		return errors.New("persistent upload failure")
	}); err != nil {
		t.Fatal(err)
	}
	const source = `set_tag("host","a"); f0=_; cast(f0,"int"); agg_create("abc",on_interval="1s",on_count=0); agg_metric("abc","f1","avg",["host"],"f0")`
	input, err := EncodeFlatPoints([]Point{
		{Version: 1, Category: "logging", Fields: map[string]any{"message": "1"}},
		{Version: 1, Category: "logging", Fields: map[string]any{"message": "2"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if batch, err := runner.Process(source, input); err != nil || len(batch.Records) != 2 {
		t.Fatalf("batch=%+v error=%v", batch, err)
	}
	// The runtime deliberately uses one low-overhead maintenance worker. Under
	// CPU-starved CI/WSL runs, unrelated native tests can delay that worker by
	// more than one nominal one-second tick; verify eventual bounded retry
	// without turning scheduler latency into a semantic failure.
	deadline := time.After(8 * time.Second)
	for attempts.Load() < 3 {
		select {
		case length := <-lengths:
			if length != 1 {
				t.Fatalf("persistent retry batch length=%d, want 1", length)
			}
		case <-deadline:
			t.Fatalf("persistent aggregate was not retried; attempts=%d", attempts.Load())
		}
	}
	if err := runner.Close(); err == nil || !strings.Contains(err.Error(), "persistent upload failure") {
		t.Fatalf("close error=%v, want surfaced persistent upload failure", err)
	}
}

func TestAggregationGenerationRetirementFlushesOldAndIsolatesNew(t *testing.T) {
	path := os.Getenv("PLATYPUS_JIT_RUNTIME")
	if path == "" {
		t.Fatal("requires real PLATYPUS_JIT_RUNTIME")
	}
	runner, err := NewRunner(path, "pipeline-go-1.4.3-datakit", 4)
	if err != nil {
		t.Fatal(err)
	}
	events := make(chan []Point, 4)
	if err := runner.SetAggregateEventHandler(func(points []Point) error {
		events <- append([]Point(nil), points...)
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	source := func(bucket string) string {
		return `set_tag("host","a"); f0=_; cast(f0,"int"); ` +
			`agg_create("` + bucket + `",on_interval="1h",on_count=0); ` +
			`agg_metric("` + bucket + `","f1","avg",["host"],"f0")`
	}
	encode := func(values ...string) []byte {
		t.Helper()
		points := make([]Point, len(values))
		for i, value := range values {
			points[i] = Point{Version: 1, Category: "logging", Fields: map[string]any{"message": value}}
		}
		input, err := EncodeFlatPoints(points)
		if err != nil {
			t.Fatal(err)
		}
		return input
	}

	oldSource, newSource := source("old"), source("new")
	oldLease, err := runner.acquireProgram(oldSource)
	if err != nil {
		t.Fatal(err)
	}
	oldInput := encode("1", "3")
	oldDone := make(chan error, 1)
	go func() {
		batch, err := oldLease.program.ProcessIndexed(oldInput)
		if err == nil && len(batch.Records) != 2 {
			err = errors.New("old generation returned the wrong record count")
		}
		oldDone <- err
	}()
	runner.Invalidate(oldSource)
	if err := <-oldDone; err != nil {
		t.Fatal(err)
	}
	select {
	case premature := <-events:
		t.Fatalf("old generation flushed while its lease was active: %#v", premature)
	default:
	}
	if batch, err := runner.Process(newSource, encode("10", "20")); err != nil || len(batch.Records) != 2 {
		t.Fatalf("new generation batch=%+v error=%v", batch, err)
	}
	oldLease.Release()
	select {
	case points := <-events:
		if len(points) != 1 || points[0].Measurement != "old" || points[0].Fields["f1"] != float64(2) {
			t.Fatalf("old retirement points=%#v", points)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("old generation did not flush on final lease release")
	}
	select {
	case duplicate := <-events:
		t.Fatalf("old generation flushed more than once: %#v", duplicate)
	default:
	}
	if err := runner.Close(); err != nil {
		t.Fatal(err)
	}
	select {
	case points := <-events:
		if len(points) != 1 || points[0].Measurement != "new" || points[0].Fields["f1"] != float64(15) {
			t.Fatalf("new retirement points=%#v", points)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("new generation did not flush on runner close")
	}
	select {
	case duplicate := <-events:
		t.Fatalf("retirement produced duplicate events: %#v", duplicate)
	default:
	}
}
