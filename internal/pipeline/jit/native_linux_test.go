// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build pipeline_jit && linux && (amd64 || arm64)

package jit

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"
	"testing"
	"unsafe"

	"github.com/GuanceCloud/pipeline-go/ptinput/funcs"
)

var hostCompatTestOutputLength uintptr

type hostCompatCallRecorder struct {
	mu    sync.Mutex
	calls map[uint32][]uintptr
}

func newObservedPipelineGoHost() (*HostCompat, *hostCompatCallRecorder) {
	recorder := &hostCompatCallRecorder{calls: make(map[uint32][]uintptr)}
	host := NewPipelineGoHost(nil)
	host.observeInvoke = func(operation uint32, recordIndex uintptr) {
		recorder.mu.Lock()
		defer recorder.mu.Unlock()
		recorder.calls[operation] = append(recorder.calls[operation], recordIndex)
	}
	return host, recorder
}

func (recorder *hostCompatCallRecorder) reset() {
	recorder.mu.Lock()
	defer recorder.mu.Unlock()
	recorder.calls = make(map[uint32][]uintptr)
}

func (recorder *hostCompatCallRecorder) operationCalls(operation uint32) []uintptr {
	recorder.mu.Lock()
	defer recorder.mu.Unlock()
	return append([]uintptr(nil), recorder.calls[operation]...)
}

func requireHostCompatCalls(t *testing.T, recorder *hostCompatCallRecorder, operation uint32, want ...uintptr) {
	t.Helper()
	got := recorder.operationCalls(operation)
	sort.Slice(got, func(i, j int) bool { return got[i] < got[j] })
	sort.Slice(want, func(i, j int) bool { return want[i] < want[j] })
	if fmt.Sprint(got) != fmt.Sprint(want) {
		t.Fatalf("host operation %d record indexes = %v, want %v", operation, got, want)
	}
}

func requireDefaultTimeNativeCapabilities(t *testing.T, runner *Runner, source string) CheckResult {
	t.Helper()
	check := runner.Check(source)
	if check.Route != RouteJITNative || check.Capabilities.ExecutionMode != "native" ||
		check.Capabilities.RequiredHostFlags != 0 || len(check.Capabilities.HostCalls) != 0 ||
		check.Capabilities.InputProjection.Mode != "keys" || check.Capabilities.StaticOutput.Mode != "slots" {
		t.Fatalf("default_time native capabilities = %#v", check)
	}
	projection, err := runner.Projection(source)
	if err != nil {
		t.Fatalf("default_time projection: %v", err)
	}
	keys := make(map[string]bool, len(projection.Keys()))
	for _, key := range projection.Keys() {
		keys[key] = true
	}
	for _, key := range []string{"status", "time", "pl_msg"} {
		if !keys[key] {
			t.Fatalf("default_time projection keys = %v, missing %q", projection.Keys(), key)
		}
	}
	return check
}

func requireDefaultTimeStaticBatch(t *testing.T, batch Batch) {
	t.Helper()
	if batch.Static == nil {
		t.Fatalf("default_time returned %q instead of static output: %#v", batch.Protocol, batch)
	}
	keys := make(map[string]bool, len(batch.Static.Schema.Keys))
	for _, key := range batch.Static.Schema.Keys {
		keys[key] = true
	}
	for _, key := range []string{"time", "pl_msg"} {
		if !keys[key] {
			t.Fatalf("default_time static schema keys = %v, missing %q", batch.Static.Schema.Keys, key)
		}
	}
}

func applyHostCompatRecord(t *testing.T, batch Batch, recordIndex int, point *Point) {
	t.Helper()
	if batch.Static == nil {
		deltas, err := batch.Records[recordIndex].MutationDeltas()
		if err != nil || len(deltas) != 1 {
			t.Fatalf("decode host compatibility delta: count=%d err=%v", len(deltas), err)
		}
		if err := deltas[0].Apply(point); err != nil {
			t.Fatalf("apply host compatibility delta: %v", err)
		}
		return
	}
	static := batch.Static
	if recordIndex < 0 || recordIndex+1 >= len(static.StateOffsets) || recordIndex+1 >= len(static.ValueOffsets) {
		t.Fatalf("static record index %d is invalid", recordIndex)
	}
	stateStart, stateEnd := int(static.StateOffsets[recordIndex]), int(static.StateOffsets[recordIndex+1])
	valueIndex, valueEnd := int(static.ValueOffsets[recordIndex]), int(static.ValueOffsets[recordIndex+1])
	if stateEnd-stateStart != len(static.Schema.Keys) || stateStart < 0 || stateEnd > len(static.States) ||
		valueIndex < 0 || valueEnd < valueIndex || valueEnd > len(static.Values) {
		t.Fatalf("invalid static record %d ranges", recordIndex)
	}
	for slotIndex, key := range static.Schema.Keys {
		switch static.States[stateStart+slotIndex] {
		case StaticNoop:
		case StaticSetField:
			if valueIndex >= valueEnd {
				t.Fatalf("static record %d field %q has no value", recordIndex, key)
			}
			if point.Fields == nil {
				point.Fields = make(map[string]any)
			}
			point.Fields[key] = static.Values[valueIndex]
			valueIndex++
		case StaticDeleteField:
			delete(point.Fields, key)
		case StaticSetTag:
			if valueIndex >= valueEnd {
				t.Fatalf("static record %d tag %q has no value", recordIndex, key)
			}
			value, ok := static.Values[valueIndex].(string)
			if !ok {
				t.Fatalf("static record %d tag %q value is %T", recordIndex, key, static.Values[valueIndex])
			}
			if point.Tags == nil {
				point.Tags = make(map[string]string)
			}
			point.Tags[key] = value
			valueIndex++
		case StaticDeleteTag:
			delete(point.Tags, key)
		default:
			t.Fatalf("static record %d slot %q has state %d", recordIndex, key, static.States[stateStart+slotIndex])
		}
	}
	if valueIndex != valueEnd {
		t.Fatalf("static record %d contains %d unused values", recordIndex, valueEnd-valueIndex)
	}
}

func TestNativeRawMeasurementReadAndMutation(t *testing.T) {
	path := os.Getenv("PLATYPUS_JIT_RUNTIME")
	if path == "" {
		t.Fatal("requires real PLATYPUS_JIT_RUNTIME")
	}
	runner, err := NewRunner(path, "pipeline-go-1.4.3-datakit", 8)
	if err != nil {
		t.Fatal(err)
	}
	defer runner.Close()

	const source = `add_key(before,pt_name()); set_measurement(message); add_key(after,pt_name())`
	point := Point{
		Version: 1, Category: "logging", Measurement: "metric\xff\x00",
		Fields: map[string]any{"message": RawString("renamed\xfe\x00")},
	}
	input, err := EncodeFlatPoints([]Point{point})
	if err != nil {
		t.Fatal(err)
	}
	batch, err := runner.Process(source, input)
	if err != nil || len(batch.Records) != 1 || batch.Records[0].Status != TerminalOK {
		t.Fatalf("records=%d error=%v batch=%+v", len(batch.Records), err, batch)
	}
	applyHostCompatRecord(t, batch, 0, &point)
	if point.Measurement != "renamed\xfe\x00" || point.Fields["before"] != RawString("metric\xff\x00") ||
		point.Fields["after"] != RawString("renamed\xfe\x00") {
		t.Fatalf("raw measurement was not preserved: %#v", point)
	}
}

func TestNativeDynamicBuiltinArguments(t *testing.T) {
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
cast_type="int"; cast(number,cast_type)
precision="s"; fmt="RFC3339"; timezone="UTC"; datetime(epoch,precision,fmt,timezone)
layout="20060102"; default_time_with_fmt(day,layout,timezone)
old_precision="ms"; new_precision="ns"; duration_precision(duration,old_precision,new_precision)
candidates=["warn","error"]; group_in(level,candidates,"matched",grouped)
group_between(score,[1,10],"inside",bucket)
cover_range=[2,4]; cover(secret,cover_range)
data=load_json("{\"a\":1,\"b\":2}"); delete_key="a"; delete(data,delete_key); add_key(deleted_map,data)
`
	check := runner.Check(source)
	if check.Route != RouteJITNative || check.Capabilities.Backend != ExecutionBackendMachineCode ||
		check.Capabilities.RequiredHostFlags != 0 || len(check.Capabilities.HostCalls) != 0 {
		t.Fatalf("route=%+v", check)
	}
	point := Point{Version: 1, Category: "logging", Measurement: "test", Fields: map[string]any{
		"message": "hello", "number": "42", "epoch": int64(0), "day": "20240102",
		"duration": int64(2), "level": "warn", "score": int64(5),
		"secret": "abcdef",
	}}
	input, err := EncodeFlatPoints([]Point{point})
	if err != nil {
		t.Fatal(err)
	}
	batch, err := runner.Process(source, input)
	if err != nil || len(batch.Records) != 1 || batch.Records[0].Status != TerminalOK {
		t.Fatalf("batch=%#v error=%v", batch, err)
	}
	applyHostCompatRecord(t, batch, 0, &point)
	for key, want := range map[string]any{
		"number": int64(42), "epoch": "1970-01-01T00:00:00+00:00",
		"day": int64(1704153600000000000), "duration": int64(2000000),
		"grouped": "matched", "bucket": "inside",
		"secret":      "a***ef",
		"deleted_map": `{"b":2}`,
	} {
		if got := point.Fields[key]; got != want {
			t.Fatalf("%s=%#v want %#v; point=%#v", key, got, want, point)
		}
	}
}

func TestNativeJSONResourceLimitCommitsPrefixWithoutReplay(t *testing.T) {
	path := os.Getenv("PLATYPUS_JIT_RUNTIME")
	if path == "" {
		t.Fatal("requires real PLATYPUS_JIT_RUNTIME")
	}
	runner, err := NewRunner(path, "pipeline-go-1.4.3-datakit", 1)
	if err != nil {
		t.Fatal(err)
	}
	defer runner.Close()

	const source = `add_key(before,"kept"); path="a"; json(payload,path,extracted); add_key(after,"bad")`
	if check := runner.Check(source); check.Route != RouteJITNative {
		t.Fatalf("route=%+v", check)
	}
	// The bounded value tree counts the root object, its member and nested array
	// as nodes, leaving 65,533 scalar slots at the 65,536-node boundary.
	atLimit := `{"a":[` + strings.Repeat("0,", 65532) + `0]}`
	point := Point{Version: 1, Category: "logging", Measurement: "json-limit", Fields: map[string]any{
		"message": "input", "payload": atLimit,
	}}
	input, err := EncodeFlatPoints([]Point{point})
	if err != nil {
		t.Fatal(err)
	}
	batch, err := runner.Process(source, input)
	if err != nil || len(batch.Records) != 1 || batch.Records[0].Status != TerminalOK {
		t.Fatalf("at-limit batch=%#v error=%v", batch, err)
	}
	applyHostCompatRecord(t, batch, 0, &point)
	if point.Fields["before"] != "kept" || point.Fields["after"] != "bad" {
		t.Fatalf("at-limit mutations missing: before=%#v after=%#v", point.Fields["before"], point.Fields["after"])
	}

	payload := `{"a":[` + strings.Repeat("0,", 65536) + `0]}`
	point = Point{Version: 1, Category: "logging", Measurement: "json-limit", Fields: map[string]any{
		"message": "input", "payload": payload,
	}}
	input, err = EncodeFlatPoints([]Point{point})
	if err != nil {
		t.Fatal(err)
	}
	batch, err = runner.Process(source, input)
	if err != nil || len(batch.Records) != 1 {
		t.Fatalf("records=%d error=%v", len(batch.Records), err)
	}
	record := batch.Records[0]
	if record.Status != TerminalError || !record.CommitPrefixError {
		t.Fatalf("record=%+v", record)
	}
	var diagnostic struct {
		Code string `json:"code"`
	}
	if err := json.Unmarshal(record.Error, &diagnostic); err != nil || diagnostic.Code != "E_OUTPUT_LIMIT" {
		t.Fatalf("diagnostic=%s error=%v", record.Error, err)
	}
	applyHostCompatRecord(t, batch, 0, &point)
	if point.Fields["before"] != "kept" || point.Fields["after"] != nil || point.Fields["extracted"] != nil {
		t.Fatalf("prefix=%#v", point.Fields)
	}

	// The production runner is persistent. A record-local resource failure must
	// not poison its compiled program or retained execution context.
	point = Point{Version: 1, Category: "logging", Measurement: "json-limit", Fields: map[string]any{
		"message": "input", "payload": `{"a":[1]}`,
	}}
	input, err = EncodeFlatPoints([]Point{point})
	if err != nil {
		t.Fatal(err)
	}
	batch, err = runner.Process(source, input)
	if err != nil || len(batch.Records) != 1 || batch.Records[0].Status != TerminalOK {
		t.Fatalf("recovery batch=%#v error=%v", batch, err)
	}
	applyHostCompatRecord(t, batch, 0, &point)
	if point.Fields["before"] != "kept" || point.Fields["after"] != "bad" || point.Fields["extracted"] != nil {
		t.Fatalf("recovery output=%#v", point.Fields)
	}
}

func TestHostCompatNativeCallback(t *testing.T) {
	if got, want := unsafe.Sizeof(hostCompatFlatV1{}), uintptr(24); got != want {
		t.Fatalf("host ABI struct size = %d, want %d", got, want)
	}
	if got, want := unsafe.Offsetof(hostCompatFlatV1{}.Invoke), uintptr(16); got != want {
		t.Fatalf("host ABI invoke offset = %d, want %d", got, want)
	}
	var observed []string
	registration := registerHostCompat(NewPipelineGoHostObserved(nil, func(operation string) {
		observed = append(observed, operation)
	}))
	request := encodeHostCompatStrings("curl/8.0")
	hostCompatTestOutputLength = 0
	status := invokeHostCompatCallback(registration.token, hostCompatUserAgentOp, 11, 22,
		uintptr(unsafe.Pointer(unsafe.SliceData(request))), uintptr(len(request)), 0, 0,
		uintptr(unsafe.Pointer(&hostCompatTestOutputLength)))
	if status != hostCompatCallbackBufferTooSmall || hostCompatTestOutputLength == 0 {
		t.Fatalf("sizing callback status = %d, required = %d", status, hostCompatTestOutputLength)
	}
	output := make([]byte, hostCompatTestOutputLength)
	status = invokeHostCompatCallback(registration.token, hostCompatUserAgentOp, 11, 22,
		uintptr(unsafe.Pointer(unsafe.SliceData(request))), uintptr(len(request)),
		uintptr(unsafe.Pointer(unsafe.SliceData(output))), uintptr(len(output)),
		uintptr(unsafe.Pointer(&hostCompatTestOutputLength)))
	if status != hostCompatCallbackOK || len(output) == 0 || output[0] != 1 {
		t.Fatalf("callback status = %d, output = %v", status, output)
	}
	status = invokeHostCompatCallback(registration.token, 999, 11, 22, 0, 0, 0, 0,
		uintptr(unsafe.Pointer(&hostCompatTestOutputLength)))
	if status != hostCompatCallbackInvalidArgument {
		t.Fatalf("unknown callback status = %d", status)
	}
	if got, want := fmt.Sprint(observed), "[user_agent user_agent unknown]"; got != want {
		t.Fatalf("observed callback invocations = %s, want %s", got, want)
	}
	registration.close()
	status = invokeHostCompatCallback(registration.token, hostCompatUserAgentOp, 0, 0, 0, 0, 0, 0,
		uintptr(unsafe.Pointer(&hostCompatTestOutputLength)))
	if status != hostCompatCallbackHostError {
		t.Fatalf("closed callback status = %d", status)
	}
	if len(observed) != 3 {
		t.Fatalf("closed host callback was counted: %v", observed)
	}
}

func TestHostCompatRuntimeEndToEnd(t *testing.T) {
	runtimePath := os.Getenv("PLATYPUS_JIT_RUNTIME")
	if runtimePath == "" {
		t.Skip("set PLATYPUS_JIT_RUNTIME to a trusted platypus_jit cdylib")
	}
	host, recorder := newObservedPipelineGoHost()
	runner, err := NewRunnerWithHost(runtimePath, "pipeline-go-1.4.3-datakit", 8, host)
	if err != nil {
		t.Fatalf("open runner with compatibility host: %v", err)
	}
	defer runner.Close()
	nativeCheck := runner.Check(strings.Repeat("cast(n, \"int\")\n", 16))
	if nativeCheck.Route != RouteJITNative {
		t.Fatalf("native capability route = %#v", nativeCheck)
	}
	requireDefaultTimeNativeCapabilities(t, runner, "default_time(time)\n")

	const pointTime = int64(1_700_000_000_000_000_123)
	t.Run("default-time-success", func(t *testing.T) {
		recorder.reset()
		got := runHostCompatPoint(t, runner, "default_time(time)\n", Point{
			Version: 1, Category: "logging", Measurement: "time-success", TimeUnixNano: pointTime,
			Fields: map[string]any{"time": "2026-05-19T13:47:01.004+0800"},
		})
		if value, ok := got.Fields["time"].(int64); !ok || value != 1_779_169_621_004_000_000 {
			t.Fatalf("parsed time = %#v", got.Fields["time"])
		}
		if _, ok := got.Fields["pl_msg"]; ok {
			t.Fatalf("successful default_time set pl_msg: %#v", got.Fields)
		}
		requireHostCompatCalls(t, recorder, hostCompatTimestampOp)
	})

	t.Run("default-time-at-any-position", func(t *testing.T) {
		cases := []struct {
			name      string
			source    string
			hasBefore bool
			hasAfter  bool
		}{
			{name: "first", source: "default_time(time)\nadd_key(after, \"ok\")\n", hasAfter: true},
			{name: "middle", source: "add_key(before, \"ok\")\ndefault_time(time)\nadd_key(after, \"ok\")\n", hasBefore: true, hasAfter: true},
			{name: "last", source: "add_key(before, \"ok\")\ndefault_time(time)\n", hasBefore: true},
		}
		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				recorder.reset()
				got := runHostCompatPoint(t, runner, tc.source, Point{
					Version: 1, Category: "logging", Measurement: "default-time-position", TimeUnixNano: pointTime,
					Fields: map[string]any{"time": "2026-05-19T13:47:01.004+0800"},
				})
				if value, ok := got.Fields["time"].(int64); !ok || value != 1_779_169_621_004_000_000 {
					t.Fatalf("default_time position result = %#v", got.Fields["time"])
				}
				if _, ok := got.Fields["before"]; ok != tc.hasBefore {
					t.Fatalf("before field presence = %v, want %v: %#v", ok, tc.hasBefore, got.Fields)
				}
				if _, ok := got.Fields["after"]; ok != tc.hasAfter {
					t.Fatalf("after field presence = %v, want %v: %#v", ok, tc.hasAfter, got.Fields)
				}
				requireHostCompatCalls(t, recorder, hostCompatTimestampOp)
			})
		}
	})

	t.Run("default-time-failure-stays-native-and-continues", func(t *testing.T) {
		recorder.reset()
		point := Point{
			Version: 1, Category: "logging", Measurement: "default-time-failure", TimeUnixNano: pointTime,
			Fields: map[string]any{"time": "not-a-time"},
		}
		input, err := EncodeFlatPoints([]Point{point})
		if err != nil {
			t.Fatal(err)
		}
		batch, err := runner.Process("add_key(before, \"ok\")\ndefault_time(time)\nadd_key(after, \"ok\")\n", input)
		if err != nil {
			t.Fatal(err)
		}
		if len(batch.Records) != 1 || batch.Records[0].Status != TerminalOK {
			t.Fatalf("default_time native failure did not complete: %#v", batch)
		}
		requireDefaultTimeStaticBatch(t, batch)
		applyHostCompatRecord(t, batch, 0, &point)
		_, timestampErr := funcs.TimestampHandle("not-a-time", "")
		wantMessage := fmt.Sprintf("time convert failed: %v", timestampErr)
		if point.Fields["time"] != pointTime || point.Fields["pl_msg"] != wantMessage ||
			point.Fields["before"] != "ok" || point.Fields["after"] != "ok" {
			t.Fatalf("default_time native failure semantics = %#v, want time=%d pl_msg=%q and continued script", point.Fields, pointTime, wantMessage)
		}
		requireHostCompatCalls(t, recorder, hostCompatTimestampOp)
	})

	t.Run("default-time-prefix-not-replayed-after-native-error", func(t *testing.T) {
		recorder.reset()
		point := Point{
			Version: 1, Category: "logging", Measurement: "default-time-native-error", TimeUnixNano: pointTime,
			Fields: map[string]any{"time": "2021-01-02 03:04:05", "divisor": int64(0)},
		}
		input, err := EncodeFlatPoints([]Point{point})
		if err != nil {
			t.Fatalf("encode native error point: %v", err)
		}
		batch, err := runner.Process("default_time(time)\nvalue = 1 / divisor\nadd_key(after, \"must-not-run\")\n", input)
		if err != nil {
			t.Fatalf("process native error point: %v", err)
		}
		if len(batch.Records) != 1 || batch.Records[0].Status != TerminalError || !batch.Records[0].CommitPrefixError {
			t.Fatalf("unexpected native error record: %#v", batch.Records)
		}
		applyHostCompatRecord(t, batch, 0, &point)
		if _, ok := point.Fields["time"].(int64); !ok {
			t.Fatalf("native error lost committed time prefix: %#v", point.Fields)
		}
		if _, ok := point.Fields["after"]; ok {
			t.Fatalf("post-error instruction mutated fallback point: %#v", point.Fields)
		}
		requireHostCompatCalls(t, recorder, hostCompatTimestampOp)
	})

	t.Run("default-time-native-error-is-record-local", func(t *testing.T) {
		recorder.reset()
		points := []Point{
			{Version: 1, Category: "logging", Measurement: "time-ok-0", TimeUnixNano: pointTime, Fields: map[string]any{"time": "2026-05-19T13:47:01.004+0800"}},
			{Version: 1, Category: "logging", Measurement: "time-error-1", TimeUnixNano: pointTime + 1, Fields: map[string]any{"time": "not-a-time"}},
			{Version: 1, Category: "logging", Measurement: "time-ok-2", TimeUnixNano: pointTime + 2, Fields: map[string]any{"time": "2026-05-19T13:47:01.004+0800"}},
			{Version: 1, Category: "logging", Measurement: "time-error-3", TimeUnixNano: pointTime + 3, Fields: map[string]any{"time": "still-not-a-time"}},
		}
		input, err := EncodeFlatPoints(points)
		if err != nil {
			t.Fatal(err)
		}
		batch, err := runner.Process("default_time(time)\nadd_key(after, \"ok\")\n", input)
		if err != nil {
			t.Fatal(err)
		}
		if len(batch.Records) != len(points) {
			t.Fatalf("record-local timestamp batch size = %d, want %d", len(batch.Records), len(points))
		}
		for index := range points {
			if batch.Records[index].Status != TerminalOK {
				t.Fatalf("record %d unexpectedly requested replay: %#v", index, batch.Records[index])
			}
			requireDefaultTimeStaticBatch(t, batch)
			applyHostCompatRecord(t, batch, index, &points[index])
			if points[index].Fields["after"] != "ok" {
				t.Fatalf("record %d did not continue after default_time: %#v", index, points[index].Fields)
			}
		}
		if points[1].Fields["time"] != pointTime+1 || points[3].Fields["time"] != pointTime+3 {
			t.Fatalf("failed records did not use their own point time: %#v %#v", points[1], points[3])
		}
		requireHostCompatCalls(t, recorder, hostCompatTimestampOp)
	})

	t.Run("user-agent", func(t *testing.T) {
		got := runNativeUserAgentPoint(t, runner, "user_agent(agent)\n", Point{
			Version: 1, Category: "logging", Measurement: "ua", TimeUnixNano: pointTime,
			Fields: map[string]any{"agent": "Mozilla/5.0 (iPhone; CPU iPhone OS 17_0 like Mac OS X) AppleWebKit/605.1.15 Version/17.0 Mobile/15E148 Safari/604.1"},
		})
		if mobile, ok := got.Fields["isMobile"].(bool); !ok || !mobile {
			t.Fatalf("isMobile = %#v; fields = %#v", got.Fields["isMobile"], got.Fields)
		}
		if browser, ok := got.Fields["browser"].(string); !ok || browser == "" {
			t.Fatalf("browser = %#v; fields = %#v", got.Fields["browser"], got.Fields)
		}
	})

	t.Run("geoip-without-database", func(t *testing.T) {
		got := runHostCompatPoint(t, runner, "geoip(ip)\n", Point{
			Version: 1, Category: "logging", Measurement: "geo", TimeUnixNano: pointTime,
			Fields: map[string]any{"ip": "203.0.113.1"},
		})
		if got.Fields["ip"] != "203.0.113.1" {
			t.Fatalf("geoip without a database mutated point: %#v", got.Fields)
		}
		for _, key := range []string{"city", "province", "country", "isp"} {
			if _, ok := got.Fields[key]; ok {
				t.Fatalf("geoip without a database set %q: %#v", key, got.Fields)
			}
		}
	})
}

func TestHostCompatRealScriptCompileBoundary(t *testing.T) {
	runtimePath := os.Getenv("PLATYPUS_JIT_RUNTIME")
	if runtimePath == "" {
		t.Skip("set PLATYPUS_JIT_RUNTIME to a trusted platypus_jit cdylib")
	}
	runner, err := NewRunnerWithHost(runtimePath, "pipeline-go-1.4.3-datakit", 8, NewPipelineGoHost(nil))
	if err != nil {
		t.Fatalf("open runner with compatibility host: %v", err)
	}
	defer runner.Close()

	const arbiter = `grok(_, "%{TIMESTAMP_ISO8601:time}%{SPACE}%{LOGLEVEL:status}%{SPACE}%{NOTSPACE:service}%{SPACE}%{DATA:source_file}:%{INT:source_line}%{SPACE}\\[workspace_uuid=%{NOTSPACE:workspace_uuid} monitor_checker_id=%{NOTSPACE:monitor_checker_id} task_id=%{NOTSPACE:task_id}\\] query time range=%{DATA:query_start} ~ %{GREEDYDATA:query_end}")
cast(source_line, "int")
default_time(time)
`
	if err := runner.Prepare(arbiter); err != nil {
		t.Fatalf("real default_time-last script should compile: %v", err)
	}
	got := runHostCompatPoint(t, runner, arbiter, Point{
		Version: 1, Category: "logging", Measurement: "arbiter", TimeUnixNano: 1_700_000_000_000_000_123,
		Fields: map[string]any{"message": "2026-05-19T13:47:01.004+0800 INFO svc file.go:42 [workspace_uuid=wk monitor_checker_id=mc task_id=t] query time range=a ~ b"},
	})
	if value, ok := got.Fields["time"].(int64); !ok || value != 1_779_169_621_004_000_000 {
		t.Fatalf("real default_time-last script did not parse time: %#v", got.Fields)
	}

	const coredns = `grok(_,"%{GREEDYDATA:ip_or_host} - - \\[%{HTTPDATE:time}\\] \"%{DATA:http_method} %{GREEDYDATA:http_url} HTTP/%{NUMBER:http_version}\" %{NUMBER:http_code} ")
grok(_,"%{GREEDYDATA:ip_or_host} - - \\[%{HTTPDATE:time}\\] \"-\" %{NUMBER:http_code} ")
default_time(time)
cast(http_code,"int")
grok(_,"\\[%{HTTPDERROR_DATE:time}\\] \\[%{GREEDYDATA:type}:%{GREEDYDATA:status}\\] \\[pid %{GREEDYDATA:pid}:tid %{GREEDYDATA:tid}\\] ")
grok(_,"\\[%{HTTPDERROR_DATE:time}\\] \\[%{GREEDYDATA:type}:%{GREEDYDATA:status}\\] \\[pid %{INT:pid}\\] ")
default_time(time)
`
	if err := runner.Prepare(coredns); err != nil {
		t.Fatalf("coredns with non-terminal default_time should compile: %v", err)
	}
}

func runHostCompatPoint(t *testing.T, runner *Runner, source string, point Point) Point {
	t.Helper()
	input, err := EncodeFlatPoints([]Point{point})
	if err != nil {
		t.Fatalf("encode host compatibility point: %v", err)
	}
	batch, err := runner.Process(source, input)
	if err != nil {
		t.Fatalf("process host compatibility point: %v", err)
	}
	if len(batch.Records) != 1 || batch.Records[0].Status != TerminalOK {
		t.Fatalf("unexpected host compatibility batch: %#v", batch)
	}
	if strings.Contains(source, "default_time(") {
		requireDefaultTimeStaticBatch(t, batch)
	}
	applyHostCompatRecord(t, batch, 0, &point)
	return point
}

func TestFlatRuntimeProcessIndexed(t *testing.T) {
	runtimePath := os.Getenv("PLATYPUS_JIT_RUNTIME")
	if runtimePath == "" {
		t.Skip("set PLATYPUS_JIT_RUNTIME to a trusted platypus_jit cdylib")
	}
	runner, err := NewRunner(runtimePath, "pipeline-go-1.4.3", 8)
	if err != nil {
		t.Fatalf("open runner: %v", err)
	}
	defer runner.Close()

	input, err := EncodeFlatPoints([]Point{
		{Version: 1, Category: "logging", Measurement: "ok", Fields: map[string]any{"foo": "remove"}},
		{Version: 1, Category: "logging", Measurement: "bad", Fields: map[string]any{"foo": []any{func() {}}}},
	})
	if err == nil {
		t.Fatal("unsupported Go value should not be encoded")
	}
	input, err = EncodeFlatPoints([]Point{
		{Version: 1, Category: "logging", Measurement: "ok", Fields: map[string]any{"foo": "remove"}},
		{Version: 1, Category: "logging", Measurement: "also-ok", Fields: map[string]any{"foo": "remove"}},
	})
	if err != nil {
		t.Fatalf("encode flat points: %v", err)
	}
	batch, err := runner.Process("drop_key(foo)\n", input)
	if err != nil {
		t.Fatalf("process indexed batch: %v", err)
	}
	if len(batch.Records) != 2 || batch.Records[0].Status != TerminalOK || batch.Records[1].Status != TerminalOK || !batch.Records[0].HasMutations() {
		t.Fatalf("unexpected records: %#v", batch.Records)
	}
	if _, err := runner.Process("geoip(ip)\n", input); err == nil {
		t.Fatal("pipeline-go-1.4.3 changed builtin should request fallback")
	}
}

func TestDatakitProfileRunsEmptyAndCommentOnlyScriptsNatively(t *testing.T) {
	runtimePath := os.Getenv("PLATYPUS_JIT_RUNTIME")
	if runtimePath == "" {
		t.Skip("set PLATYPUS_JIT_RUNTIME to a trusted platypus_jit cdylib")
	}
	runner, err := NewRunner(runtimePath, "pipeline-go-1.4.3-datakit", 8)
	if err != nil {
		t.Fatalf("open runner: %v", err)
	}
	defer runner.Close()

	for _, source := range []string{"", "  \n\t", "# comment only\n"} {
		check := runner.Check(source)
		if check.Route != RouteJITNative || check.Capabilities.ExecutionMode != "native" {
			t.Fatalf("no-op source %q route = %#v", source, check)
		}
		got := runHostCompatPoint(t, runner, source, Point{
			Version: 1, Category: "logging", Measurement: "noop",
			Tags: map[string]string{"env": "prod"}, Fields: map[string]any{"message": "unchanged", "count": int64(42)},
			TimeUnixNano: 1_725_000_000_123_456_789,
		})
		if got.Measurement != "noop" || got.TimeUnixNano != 1_725_000_000_123_456_789 ||
			got.Tags["env"] != "prod" || got.Fields["message"] != "unchanged" ||
			got.Fields["count"] != int64(42) || got.Fields["status"] != "info" {
			t.Fatalf("no-op source %q changed Point semantics: %#v", source, got)
		}
	}
}

func TestDatakitProfileUsesStaticAndGeneralPrograms(t *testing.T) {
	runtimePath := os.Getenv("PLATYPUS_JIT_RUNTIME")
	if runtimePath == "" {
		t.Skip("set PLATYPUS_JIT_RUNTIME to a trusted platypus_jit cdylib")
	}
	runner, err := NewRunner(runtimePath, "pipeline-go-1.4.3-datakit", 8)
	if err != nil {
		t.Fatalf("open runner: %v", err)
	}
	defer runner.Close()

	short := "cast(n, \"int\")\n"
	if err := runner.Prepare(short); err != nil {
		t.Fatalf("short compatible program should compile: %v", err)
	}
	projection, err := runner.Projection(short)
	if err != nil || strings.Join(projection.Keys(), ",") != "n,status" {
		t.Fatalf("short cast projection = %v, %v", projection.Keys(), err)
	}
	if err := runner.Prepare(strings.Repeat("cast(n, \"string\")\n", 16)); err != nil {
		t.Fatalf("pipeline-go string target (nil conversion) should compile: %v", err)
	}
	profitable := strings.Repeat("cast(n, \"int\")\n", 16)
	if err := runner.Prepare(profitable); err != nil {
		t.Fatalf("prepare profitable program: %v", err)
	}
	projection, err = runner.Projection(profitable)
	if err != nil || strings.Join(projection.Keys(), ",") != "n,status" {
		t.Fatalf("cast projection = %v, %v", projection.Keys(), err)
	}
	jsonSource := strings.Repeat("json(message, payload.value, value)\n", 16)
	if err := runner.Prepare(jsonSource); err != nil {
		t.Fatalf("prepare fused JSON projection program: %v", err)
	}
	projection, err = runner.Projection(jsonSource)
	if err != nil || strings.Join(projection.Keys(), ",") != "message,status,value" {
		t.Fatalf("JSON projection = %v, %v", projection.Keys(), err)
	}
	regexSource := strings.Repeat("replace(n, \"[a-z]+\", \"x\")\n", 16)
	if err := runner.Prepare(regexSource); err != nil {
		t.Fatalf("prepare statically compiled regex program: %v", err)
	}
	projection, err = runner.Projection(regexSource)
	if err != nil || strings.Join(projection.Keys(), ",") != "n,status" {
		t.Fatalf("regex projection = %v, %v", projection.Keys(), err)
	}
	input, err := EncodeFlatPoints([]Point{{
		Version: 1, Category: "logging", Measurement: "cast", Fields: map[string]any{"n": "42"},
	}})
	if err != nil {
		t.Fatalf("encode flat points: %v", err)
	}
	batch, err := runner.Process(profitable, input)
	if err != nil {
		t.Fatalf("process profitable program: %v", err)
	}
	if len(batch.Records) != 1 || batch.Records[0].Status != TerminalOK {
		t.Fatalf("unexpected records: %#v", batch.Records)
	}
	if batch.Static == nil || strings.Join(batch.Static.Schema.Keys, ",") != "n,status" {
		t.Fatalf("profitable program did not use its static output schema: %#v", batch.Static)
	}

	hybrid := "set_measurement(\"renamed\")\ncreate_point(\"child\", {\"env\": \"prod\"}, {\"count\": 1}, 123, \"M\")\n"
	if err := runner.Prepare(hybrid); err != nil {
		t.Fatalf("prepare general mutation program: %v", err)
	}
	projection, err = runner.Projection(hybrid)
	if err != nil || !projection.IsAll() {
		t.Fatalf("general mutation projection = %#v, %v", projection, err)
	}
	hybridBatch, err := runner.Process(hybrid, input)
	if err != nil {
		t.Fatalf("process general mutation program: %v", err)
	}
	if hybridBatch.Static != nil || len(hybridBatch.Records) != 1 || !hybridBatch.Records[0].HasMutations() {
		t.Fatalf("general mutation program did not use MutationDelta: %#v", hybridBatch)
	}
	droppedBatch, err := runner.Process("drop()\n", input)
	if err != nil {
		t.Fatalf("process drop program: %v", err)
	}
	if droppedBatch.Static != nil || len(droppedBatch.Records) != 1 || droppedBatch.Records[0].Status != TerminalDropped {
		t.Fatalf("drop program did not return a dropped terminal: %#v", droppedBatch)
	}

	for _, test := range []struct {
		name       string
		source     string
		fields     map[string]any
		projection string
		wantValue  any
	}{
		{
			name:       "origin-key",
			source:     strings.Repeat("replace(_, \"old\", \"new\")\n", 16),
			fields:     map[string]any{"message": "old value"},
			projection: "message,status",
			wantValue:  "new value",
		},
		{
			name:       "replace-non-string",
			source:     strings.Repeat("replace(n, \"2\", \"x\")\n", 16),
			fields:     map[string]any{"n": int64(123)},
			projection: "n,status",
			wantValue:  "1x3",
		},
		{
			name:       "cast-large-int",
			source:     strings.Repeat("cast(n, \"int\")\n", 16),
			fields:     map[string]any{"n": int64(9_007_199_254_740_993)},
			projection: "n,status",
			wantValue:  int64(9_007_199_254_740_992),
		},
		{
			name:       "cast-explicit-nil",
			source:     strings.Repeat("cast(n, \"str\")\n", 16),
			fields:     map[string]any{"n": nil},
			projection: "n,status",
			wantValue:  "",
		},
		{
			name:       "cast-composite",
			source:     strings.Repeat("cast(n, \"str\")\n", 16),
			fields:     map[string]any{"n": map[string]any{"b": true, "a": int64(1)}},
			projection: "n,status",
			wantValue:  `{"a":1,"b":true}`,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			projection, err := runner.Projection(test.source)
			if err != nil || strings.Join(projection.Keys(), ",") != test.projection {
				t.Fatalf("projection = %v, %v", projection.Keys(), err)
			}
			input, err := EncodeFlatPoints([]Point{{
				Version: 1, Category: "logging", Measurement: test.name, Fields: test.fields,
			}})
			if err != nil {
				t.Fatalf("encode flat points: %v", err)
			}
			batch, err := runner.Process(test.source, input)
			if err != nil {
				t.Fatalf("process: %v", err)
			}
			if batch.Static == nil || len(batch.Static.Values) < 1 || batch.Static.Values[0] != test.wantValue {
				t.Fatalf("static values = %#v, want first value %#v", batch.Static, test.wantValue)
			}
		})
	}

	strictJSONSource := strings.Repeat("json(message, object[0], value)\n", 16)
	strictInput, err := EncodeFlatPoints([]Point{{
		Version: 1, Category: "logging", Measurement: "strict-json",
		Fields: map[string]any{"message": `{"object":{"0":"must-not-match"}}`},
	}})
	if err != nil {
		t.Fatalf("encode strict JSON point: %v", err)
	}
	strictBatch, err := runner.Process(strictJSONSource, strictInput)
	if err != nil {
		t.Fatalf("process strict JSON path: %v", err)
	}
	if strictBatch.Static == nil || len(strictBatch.Static.States) != 2 ||
		strictBatch.Static.States[0] != StaticSetField || strictBatch.Static.States[1] != StaticNoop {
		t.Fatalf("numeric object key matched an array index: %#v", strictBatch.Static)
	}
}
