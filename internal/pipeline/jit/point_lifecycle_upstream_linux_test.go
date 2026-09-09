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
	"os"
	"testing"
)

func TestDropAndDropOriginDataPipelineGoUpstreamCorpus(t *testing.T) {
	runner := newLifecycleRunner(t)
	defer runner.Close()

	for _, size := range []int{1, 2, 4, 8, 10, 128} {
		t.Run("batch-"+itoa(size), func(t *testing.T) {
			points := lifecyclePoints(size)
			input, err := EncodeFlatPoints(points)
			if err != nil {
				t.Fatal(err)
			}

			dropSource := `grok(_, "%{GREEDYDATA:parsed}"); drop(); add_key(after, true)`
			assertLifecycleNative(t, runner, dropSource)
			dropped, err := runner.Process(dropSource, input)
			if err != nil || len(dropped.Records) != size {
				t.Fatalf("drop records=%d error=%v", len(dropped.Records), err)
			}
			for i, record := range dropped.Records {
				if record.Status != TerminalDropped || record.Point != nil || record.HasMutations() {
					t.Fatalf("drop record %d=%+v", i, record)
				}
			}

			originSource := `grok(_, "%{GREEDYDATA:parsed}"); drop_origin_data(); add_key(after, true)`
			assertLifecycleNative(t, runner, originSource)
			origin, err := runner.Process(originSource, input)
			if err != nil || len(origin.Records) != size {
				t.Fatalf("drop_origin_data records=%d error=%v", len(origin.Records), err)
			}
			for i := range origin.Records {
				if origin.Records[i].Status != TerminalOK {
					t.Fatalf("drop_origin_data record %d=%+v", i, origin.Records[i])
				}
				applyHostCompatRecord(t, origin, i, &points[i])
				if _, ok := points[i].Fields["message"]; ok {
					t.Fatalf("record %d retained origin data: %#v", i, points[i])
				}
				if points[i].Fields["parsed"] == nil || points[i].Fields["after"] != true {
					t.Fatalf("record %d lost parsed/continuation fields: %#v", i, points[i])
				}
			}
		})
	}
}

func TestCreatePointPipelineGoUpstreamCategoryAndTimestampCorpus(t *testing.T) {
	runner := newLifecycleRunner(t)
	defer runner.Close()

	const source = `
create_point("n1", {"a": "1", "ignored": 2}, {"d": "x1", "b": 2, "ignored": [1]}, 0)
create_point("n1", {"a": "1"}, {"b": 2}, 1, category="L")
create_point("n1", {"a": "1"}, {"b": 2}, 2, category="M")
create_point("n1", {"a": "1"}, {"b": 2}, 3, category="T")
create_point("n1", {"a": "1"}, {"b": 2}, 4, category="R")
create_point("n1", {"a": "1"}, {"b": 2}, 5, category="N")
create_point("n1", {"a": "1"}, {"b": 2}, 6, category="O")
create_point("n1", {"a": "1"}, {"b": 2}, 7, category="CO")
create_point("n1", {"a": "1"}, {"b": 2}, 8, category="S")
create_point("n1", {"a": "1"}, {"b": 2}, 10, category="logging")
create_point("n1", {"a": "1"}, {"b": 2}, 11, category="metric")
create_point("n1", {"a": "1"}, {"b": 2}, 12, category="tracing")
create_point("n1", {"a": "1"}, {"b": 2}, 13, category="rum")
create_point("n1", {"a": "1"}, {"b": 2}, 14, category="network")
create_point("n1", {"a": "1"}, {"b": 2}, 15, category="object")
create_point("n1", {"a": "1"}, {"b": 2}, 16, category="custom_object")
create_point("n1", {"a": "1"}, {"b": 2}, 17, category="security")
create_point("ignored", {}, {}, 9, category="unknown")
add_key(after, true)
`
	assertLifecycleNative(t, runner, source)
	for _, size := range []int{1, 2, 4, 8, 10, 128} {
		points := lifecyclePoints(size)
		input, err := EncodeFlatPoints(points)
		if err != nil {
			t.Fatal(err)
		}
		batch, err := runner.Process(source, input)
		if err != nil || len(batch.Records) != size {
			t.Fatalf("batch %d records=%d error=%v", size, len(batch.Records), err)
		}
		wantCategories := []string{
			"metric", "logging", "metric", "tracing", "rum", "network", "object", "custom_object", "security",
			"logging", "metric", "tracing", "rum", "network", "object", "custom_object", "security",
		}
		for i := range batch.Records {
			if batch.Records[i].Status != TerminalOK {
				t.Fatalf("batch %d record %d=%+v", size, i, batch.Records[i])
			}
			applyHostCompatRecord(t, batch, i, &points[i])
			if points[i].Fields["after"] != true || len(points[i].Subpoints) != len(wantCategories) {
				t.Fatalf("batch %d record %d output=%#v", size, i, points[i])
			}
			for childIndex, category := range wantCategories {
				child := points[i].Subpoints[childIndex]
				if child.Measurement != "n1" || child.Category != category || child.Tags["a"] != "1" {
					t.Fatalf("batch %d record %d child %d=%#v", size, i, childIndex, child)
				}
				wantTime := int64(childIndex)
				if childIndex == 0 {
					wantTime = points[i].TimeUnixNano
				} else if childIndex >= 9 {
					wantTime = int64(childIndex + 1)
				}
				if child.TimeUnixNano != wantTime {
					t.Fatalf("batch %d record %d child %d time=%d want=%d", size, i, childIndex, child.TimeUnixNano, wantTime)
				}
			}
			if _, ok := points[i].Subpoints[0].Tags["ignored"]; ok {
				t.Fatalf("non-string tag was retained: %#v", points[i].Subpoints[0])
			}
			if _, ok := points[i].Subpoints[0].Fields["ignored"]; ok {
				t.Fatalf("unsupported field was retained: %#v", points[i].Subpoints[0])
			}
		}
	}
}

func TestCreatePointPreservesRawTagBytes(t *testing.T) {
	runner := newLifecycleRunner(t)
	defer runner.Close()
	const source = `create_point("child", {"copied":raw}, {}, 1, "M")`
	assertLifecycleNative(t, runner, source)
	point := Point{Version: 1, Category: "logging", Measurement: "parent", Tags: map[string]string{"raw": "A\xff\x00B"}}
	projection, err := runner.Projection(source)
	if err != nil {
		t.Fatal(err)
	}
	input, err := EncodeProjectedPointRecords([]FlatPointRecord{{
		Version: 1, Category: "logging", Measurement: "parent",
		Entries: []FlatPointEntry{{Key: "raw", Value: "A\xff\x00B", IsTag: true}},
	}}, projection)
	if err != nil {
		t.Fatal(err)
	}
	batch, err := runner.Process(source, input)
	if err != nil || len(batch.Records) != 1 || batch.Records[0].Status != TerminalOK {
		t.Fatalf("records=%d error=%v batch=%+v", len(batch.Records), err, batch)
	}
	applyHostCompatRecord(t, batch, 0, &point)
	if len(point.Subpoints) != 1 || point.Subpoints[0].Tags["copied"] != "A\xff\x00B" {
		t.Fatalf("raw child tag was not preserved: %#v", point.Subpoints)
	}
}

func TestCreatePointDropOrderingAndTerminalErrors(t *testing.T) {
	runner := newLifecycleRunner(t)
	defer runner.Close()

	dropSource := `create_point("before", {}, {"count": 1}, 123); drop(); create_point("after", {}, {"count": 2}, 124)`
	assertLifecycleNative(t, runner, dropSource)
	input, err := EncodeFlatPoints(lifecyclePoints(2))
	if err != nil {
		t.Fatal(err)
	}
	batch, err := runner.Process(dropSource, input)
	if err != nil || len(batch.Records) != 2 {
		t.Fatalf("drop/create records=%d error=%v", len(batch.Records), err)
	}
	for i, record := range batch.Records {
		if record.Status != TerminalDropped || len(record.Emitted) != 2 || record.Point != nil || record.HasMutations() {
			t.Fatalf("drop/create record %d=%+v", i, record)
		}
		for childIndex, payload := range record.Emitted {
			var child struct {
				Measurement string         `json:"measurement"`
				Category    string         `json:"category"`
				Fields      map[string]any `json:"fields"`
			}
			if err := decodeEmittedTestPayload(payload, &child); err != nil {
				t.Fatalf("record %d child %d: %v", i, childIndex, err)
			}
			wantName := []string{"before", "after"}[childIndex]
			if child.Measurement != wantName || child.Category != "metric" || child.Fields["count"] != float64(childIndex+1) {
				t.Fatalf("record %d child %d=%#v", i, childIndex, child)
			}
		}
	}

	errorCases := []struct {
		name   string
		source string
		field  string
		value  any
	}{
		{"name", `add_key(before,true); create_point(bad, {}, {}, 1, "M"); add_key(after,true)`, "bad", int64(1)},
		{"tags", `add_key(before,true); create_point("child", bad, {}, 1, "M"); add_key(after,true)`, "bad", "not-a-map"},
		{"fields", `add_key(before,true); create_point("child", {}, bad, 1, "M"); add_key(after,true)`, "bad", "not-a-map"},
		{"timestamp", `add_key(before,true); create_point("child", {}, {}, bad, "M"); add_key(after,true)`, "bad", "1"},
		{"category", `add_key(before,true); create_point("child", {}, {}, 1, bad); add_key(after,true)`, "bad", int64(1)},
	}
	for _, tc := range errorCases {
		t.Run(tc.name, func(t *testing.T) {
			assertLifecycleNative(t, runner, tc.source)
			point := lifecyclePoints(1)[0]
			point.Fields[tc.field] = tc.value
			input, err := EncodeFlatPoints([]Point{point})
			if err != nil {
				t.Fatal(err)
			}
			result, err := runner.Process(tc.source, input)
			if err != nil || len(result.Records) != 1 {
				t.Fatalf("records=%d error=%v", len(result.Records), err)
			}
			record := result.Records[0]
			if record.Status != TerminalError || !record.CommitPrefixError || len(record.Error) == 0 || !record.HasMutations() {
				t.Fatalf("terminal=%+v", record)
			}
			applyHostCompatRecord(t, result, 0, &point)
			if point.Fields["before"] != true {
				t.Fatalf("committed prefix missing: %#v", point)
			}
			if _, ok := point.Fields["after"]; ok || len(point.Subpoints) != 0 {
				t.Fatalf("execution continued after error: %#v", point)
			}
		})
	}
}

func TestCreatePointAfterUsePipelineGoUpstreamCorpus(t *testing.T) {
	runner := newLifecycleRunner(t)
	defer runner.Close()

	for _, tc := range []struct {
		name       string
		main       string
		wantTime   int64
		wantMetric bool
	}{
		{"default-time", `create_point("n1", {"a":"1"}, {"b":2}, after_use="abc.p")`, 4242, true},
		{"explicit-time", `create_point("n1", {"a":"1"}, {"b":2}, ts=1, after_use="abc.p")`, 1, true},
		{"logging-category", `create_point("n1", {"a":"1"}, {"b":2}, category="L", after_use="abc.p")`, 4242, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			snapshot, err := NewModuleSnapshot("main.p", map[string]string{
				"main.p": tc.main + `; add_key(parent, true)`,
				"abc.p":  `add_key("aa", 1); add_key(child_only, true)`,
			})
			if err != nil {
				t.Fatal(err)
			}
			check := runner.CheckModules(snapshot)
			if check.Route != RouteJITNative || check.Capabilities.Backend != ExecutionBackendMachineCode ||
				check.Capabilities.RequiredHostFlags != 0 || len(check.Capabilities.HostCalls) != 0 {
				t.Fatalf("route=%+v", check)
			}
			for _, size := range []int{1, 2, 4, 8, 10, 128} {
				points := lifecyclePoints(size)
				for i := range points {
					points[i].TimeUnixNano = 4242
				}
				input, err := EncodeFlatPoints(points)
				if err != nil {
					t.Fatal(err)
				}
				batch, err := runner.ProcessModules(snapshot, input)
				if err != nil || len(batch.Records) != size {
					t.Fatalf("batch %d records=%d error=%v", size, len(batch.Records), err)
				}
				for i := range batch.Records {
					if batch.Records[i].Status != TerminalOK {
						t.Fatalf("batch %d record %d=%+v", size, i, batch.Records[i])
					}
					applyHostCompatRecord(t, batch, i, &points[i])
					if points[i].Fields["parent"] != true || points[i].Fields["child_only"] != nil || len(points[i].Subpoints) != 1 {
						t.Fatalf("parent/child isolation failed: %#v", points[i])
					}
					child := points[i].Subpoints[0]
					wantCategory := "logging"
					if tc.wantMetric {
						wantCategory = "metric"
					}
					if child.Measurement != "n1" || child.Category != wantCategory || child.TimeUnixNano != tc.wantTime ||
						child.Tags["a"] != "1" || child.Fields["b"] != int64(2) || child.Fields["aa"] != int64(1) || child.Fields["child_only"] != true {
						t.Fatalf("batch %d record %d child=%#v", size, i, child)
					}
				}
			}
		})
	}

	missing, err := NewModuleSnapshot("main.p", map[string]string{
		"main.p": `create_point("n1", {}, {}, after_use="missing.p")`,
	})
	if err != nil {
		t.Fatal(err)
	}
	if check := runner.CheckModules(missing); check.Route != RoutePipelineGo || check.Reason != CheckReasonCompileError {
		t.Fatalf("missing after_use dependency route=%+v", check)
	}
}

func TestCreatePointPipelineGoExactDynamicMapCorpus(t *testing.T) {
	runner := newLifecycleRunner(t)
	defer runner.Close()

	const prelude = `
d = load_json(_)
r = {}
for x in d["c"] {
  r[x] = d["c"][x]
}
r["b"] = d["b"]
`
	for _, tc := range []struct {
		name, call, child string
		wantTime          int64
	}{
		{"default", `create_point("n1", {"a": d["a"]}, r)`, "", 777},
		{"after-use", `create_point("n1", {"a": d["a"]}, r, category="M", after_use="abc.p")`, `add_key("aa", 1)`, 777},
		{"after-use-ts", `create_point("n1", {"a": d["a"]}, r, ts=1, after_use="abc.p")`, `add_key("aa", 1)`, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			sources := map[string]string{"main.p": prelude + tc.call}
			if tc.child != "" {
				sources["abc.p"] = tc.child
			}
			snapshot, err := NewModuleSnapshot("main.p", sources)
			if err != nil {
				t.Fatal(err)
			}
			check := runner.CheckModules(snapshot)
			if check.Route != RouteJITNative || check.Capabilities.Backend != ExecutionBackendMachineCode ||
				check.Capabilities.RequiredHostFlags != 0 || len(check.Capabilities.HostCalls) != 0 {
				t.Fatalf("route=%+v", check)
			}
			for _, size := range []int{1, 8, 128} {
				points := lifecyclePoints(size)
				for i := range points {
					points[i].TimeUnixNano = 777
					points[i].Fields["message"] = `{"a":"1","b":2,"c":{"d":"x1"}}`
				}
				input, err := EncodeFlatPoints(points)
				if err != nil {
					t.Fatal(err)
				}
				batch, err := runner.ProcessModules(snapshot, input)
				if err != nil || len(batch.Records) != size {
					t.Fatalf("batch %d records=%d error=%v", size, len(batch.Records), err)
				}
				for i := range batch.Records {
					if batch.Records[i].Status != TerminalOK {
						t.Fatalf("batch %d record %d=%+v", size, i, batch.Records[i])
					}
					applyHostCompatRecord(t, batch, i, &points[i])
					if len(points[i].Subpoints) != 1 {
						t.Fatalf("batch %d record %d subpoints=%#v", size, i, points[i].Subpoints)
					}
					child := points[i].Subpoints[0]
					if child.Measurement != "n1" || child.Category != "metric" || child.TimeUnixNano != tc.wantTime ||
						child.Tags["a"] != "1" || child.Fields["d"] != "x1" || child.Fields["b"] != float64(2) {
						t.Fatalf("batch %d record %d child=%#v", size, i, child)
					}
					if tc.child != "" && child.Fields["aa"] != int64(1) {
						t.Fatalf("batch %d record %d after_use child=%#v", size, i, child)
					}
				}
			}
		})
	}
}

func newLifecycleRunner(t *testing.T) *Runner {
	t.Helper()
	path := os.Getenv("PLATYPUS_JIT_RUNTIME")
	if path == "" {
		t.Fatal("requires real PLATYPUS_JIT_RUNTIME")
	}
	runner, err := NewRunner(path, "pipeline-go-1.4.3-datakit", 16)
	if err != nil {
		t.Fatal(err)
	}
	return runner
}

func lifecyclePoints(size int) []Point {
	points := make([]Point, size)
	for i := range points {
		points[i] = Point{
			Version: 1, Category: "logging", Measurement: "source", TimeUnixNano: int64(1000 + i),
			Fields: map[string]any{"message": "hello world", "sequence": int64(i)},
		}
	}
	return points
}

func assertLifecycleNative(t *testing.T, runner *Runner, source string) {
	t.Helper()
	check := runner.Check(source)
	if check.Route != RouteJITNative || check.Capabilities.Backend != ExecutionBackendMachineCode ||
		check.Capabilities.RequiredHostFlags != 0 || len(check.Capabilities.HostCalls) != 0 {
		t.Fatalf("route=%+v", check)
	}
}
