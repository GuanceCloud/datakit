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
	"fmt"
	"os"
	"testing"
)

func TestGetKeyPipelineGoUpstreamAndBoundaryCorpus(t *testing.T) {
	path := os.Getenv("PLATYPUS_JIT_RUNTIME")
	if path == "" {
		t.Fatal("requires real PLATYPUS_JIT_RUNTIME")
	}
	runner, err := NewRunner(path, "pipeline-go-1.4.3-datakit", 8)
	if err != nil {
		t.Fatal(err)
	}
	defer runner.Close()

	upstream := []struct {
		source string
		want   any
	}{
		{`key="shanghai"; add_key(key); key="tokyo"; add_key(add_new_key,key); add_key(after,true)`, "tokyo"},
		{`key="shanghai"; add_key(key); key="tokyo"; add_key(add_new_key,get_key(key)); add_key(after,true)`, "shanghai"},
	}
	for index, tc := range upstream {
		check := runner.Check(tc.source)
		if check.Route != RouteJITNative || check.Capabilities.ExecutionMode != "native" {
			t.Fatalf("upstream %d route=%+v", index, check)
		}
		points := make([]Point, 10)
		for i := range points {
			points[i] = Point{Version: 1, Category: "logging", Measurement: "upstream-get-key", Fields: map[string]any{"message": "original", "sequence": int64(i)}}
		}
		assertGetKeyBatch(t, runner, tc.source, points, "add_new_key", func(int) any { return tc.want })
	}

	values := []struct{ input, want any }{
		{"text", "text"}, {int64(-7), int64(-7)}, {float64(1.25), float64(1.25)}, {true, true}, {nil, nil},
		{[]any{"x", int64(2)}, `["x",2]`}, {map[string]any{"nested": int64(3)}, `{"nested":3}`}, {RawString("A\xffB"), RawString("A\xffB")},
	}
	const source = `add_key(result,get_key(target)); add_key(after,true)`
	check := runner.Check(source)
	if check.Route != RouteJITNative || check.Capabilities.ExecutionMode != "native" {
		t.Fatalf("boundary route=%+v", check)
	}
	for _, size := range []int{1, 2, 4, 8, 10, 128} {
		t.Run("field-batch-"+itoa(size), func(t *testing.T) {
			points := make([]Point, size)
			for i := range points {
				fields := map[string]any{"sequence": int64(i)}
				value := values[i%len(values)]
				if value.input != nil {
					fields["target"] = value.input
				}
				points[i] = Point{Version: 1, Category: "logging", Measurement: "field-get-key", Fields: fields}
			}
			assertGetKeyBatch(t, runner, source, points, "result", func(i int) any { return values[i%len(values)].want })
		})
	}

	const tagSource = `add_key(result,get_key(env)); add_key(after,true)`
	points := make([]Point, 10)
	for i := range points {
		points[i] = Point{Version: 1, Category: "logging", Measurement: "tag-get-key", Tags: map[string]string{"env": "prod"}, Fields: map[string]any{"sequence": int64(i)}}
	}
	assertGetKeyBatch(t, runner, tagSource, points, "result", func(int) any { return "prod" })
}

func assertGetKeyBatch(t *testing.T, runner *Runner, source string, points []Point, resultKey string, want func(int) any) {
	t.Helper()
	input, err := EncodeFlatPoints(points)
	if err != nil {
		t.Fatal(err)
	}
	batch, err := runner.Process(source, input)
	if err != nil || len(batch.Records) != len(points) {
		t.Fatalf("records=%d error=%v", len(batch.Records), err)
	}
	for i, record := range batch.Records {
		if record.Status != TerminalOK {
			t.Fatalf("record %d terminal=%d error=%s", i, record.Status, record.Error)
		}
		applyHostCompatRecord(t, batch, i, &points[i])
		if got := points[i].Fields[resultKey]; !equalGetKeyValue(got, want(i)) {
			t.Fatalf("record %d %s=%#v want=%#v", i, resultKey, got, want(i))
		}
		if points[i].Fields["after"] != true || points[i].Fields["sequence"] != int64(i) {
			t.Fatalf("record %d lost input or continuation: %#v", i, points[i])
		}
	}
}

func equalGetKeyValue(got, want any) bool {
	return valueCanonical(got) == valueCanonical(want)
}

func valueCanonical(value any) string {
	return fmt.Sprintf("%#v", value)
}
