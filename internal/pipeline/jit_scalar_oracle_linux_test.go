// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build pipeline_jit && linux && (amd64 || arm64)

// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2026-present Guance, Inc.

package pipeline

import (
	"fmt"
	"math"
	"os"
	"strings"
	"testing"

	"github.com/GuanceCloud/cliutils/point"
	pljit "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline/jit"
)

// Keep these scripts free of added sentinel instructions: those would change
// compiler eligibility and accidentally test only the general execution path.
func TestJITScalarExecutionOracle(t *testing.T) {
	runtimePath := os.Getenv("PLATYPUS_JIT_RUNTIME")
	if runtimePath == "" {
		t.Fatal("requires real PLATYPUS_JIT_RUNTIME")
	}
	runner, err := pljit.NewRunnerWithHost(runtimePath, "pipeline-go-1.4.3-datakit", 16, pljit.NewPipelineGoHost(nil))
	if err != nil {
		t.Fatal(err)
	}
	defer runner.Close()
	values := []any{
		"42", "-0", "9223372036854775807", "-9223372036854775808", "9007199254740993", "1e300", "0x1.8p2", "1_234.5", "1__2",
		"NaN", "+Inf", "-Inf", "true", "TRUE", "True", "TrUe", "t", "", "中文", "\xff",
		true, false, int64(42), int64(math.MaxInt64), int64(math.MinInt64), float64(-1.25),
	}
	for _, target := range []string{"int", "float", "bool", "string"} {
		t.Run(target, func(t *testing.T) {
			source := fmt.Sprintf("drop_key(remove_me)\ncast(n, %q)\ncast(absent, %q)", target, target)
			script, err := NewPlScriptSimple(point.Logging, "scalar.p", source)
			if err != nil {
				t.Fatal(err)
			}
			if check := runner.Check(source); check.Route != pljit.RouteJITNative {
				t.Fatalf("not native: %+v", check)
			}
			for _, batch := range []int{1, 10, 32} {
				for offset := range values {
					indexes := make([]int, batch)
					runs := make([]pointRun, batch)
					expected := make([]*point.Point, batch)
					for i := range runs {
						fields := map[string]any{"message": "unchanged", "n": values[(offset+i)%len(values)], "remove_me": "gone", "sentinel": "preserved"}
						actual := newRealScriptPoint("scalar", fields)
						expected[i] = newRealScriptPoint("scalar", fields)
						if i%4 == 1 {
							// A tag forces pre-execution fallback for this record.
							actual.AddTag("n", "42")
							expected[i].AddTag("n", "42")
						}
						if i%4 == 2 {
							actual.AddKVs(point.NewKV("status", "WARNING"))
							expected[i].AddKVs(point.NewKV("status", "WARNING"))
						}
						runs[i] = pointRun{point: actual, script: script}
						indexes[i] = i
						reference := pointRun{point: expected[i], script: script}
						runPipelineGo(point.Logging, &reference, nil)
						if reference.output == nil || reference.dropped {
							t.Fatal("Go unexpectedly dropped scalar record")
						}
						expected[i] = reference.output
					}
					runJITGroup(runner, point.Logging, script, indexes, runs, nil)
					for i, run := range runs {
						if run.output == nil || run.dropped || len(run.created) != 0 {
							t.Fatal("unexpected JIT terminal")
						}
						if equal, reason := equalJSONOraclePoints(run.output, expected[i]); !equal {
							t.Fatalf("batch=%d offset=%d record=%d: %s", batch, offset, i, reason)
						}
					}
				}
			}
		})
	}
}

func TestJITJSONSlotExecutionOracle(t *testing.T) {
	for _, mixed := range []bool{false, true} {
		t.Run(fmt.Sprintf("mixed=%t", mixed), func(t *testing.T) {
			previous := point.EnableMixedArrayField
			point.EnableMixedArrayField = mixed
			defer func() { point.EnableMixedArrayField = previous }()
			runJSONSlotExecutionOracle(t, mixed)
		})
	}
}

func runJSONSlotExecutionOracle(t *testing.T, mixed bool) {
	runtimePath := os.Getenv("PLATYPUS_JIT_RUNTIME")
	if runtimePath == "" {
		t.Fatal("requires real PLATYPUS_JIT_RUNTIME")
	}
	runner, err := pljit.NewRunnerWithServices(runtimePath, "pipeline-go-1.4.3-datakit", 32, pljit.NewPipelineGoHost(nil), pljit.ServiceConfig{Version: 1, EnableMixedArrayField: mixed})
	if err != nil {
		t.Fatal(err)
	}
	defer runner.Close()
	values := []any{
		`{"a":"42","b":"  value  ","c":true}`,
		`{"a":42,"b":null,"c":false}`,
		`{"a":{"key":1},"b":[1,2],"c":[1,"mixed"]}`,
		`{"a":{"nested":{"x":1}},"b":[{},null],"c":[]}`,
		`{"a":1,"a":2,"b":"\ud800","unused":1e400}`,
		`{"a":42,"unused":}`, "{}", "[]", "null", "not JSON", "\xff",
	}
	for index, source := range []string{
		"json(message, a, a)",
		"drop_key(remove_me)\njson(message, a, a)\njson(message, b, b)\njson(message, c, c)",
		"json(message, a, a)\njson(message, b, b)\ncast(a, \"int\")\ncast(b, \"bool\")",
		"json(message, a, a)\njson(other, b, b)\ncast(a, \"float\")\ndrop_key(b)",
		"json(_, b, b, trim_space=true)",
	} {
		t.Run(fmt.Sprint(index), func(t *testing.T) {
			script, err := NewPlScriptSimple(point.Logging, "scalar.p", source)
			if err != nil {
				t.Fatal(err)
			}
			if check := runner.Check(source); check.Route != pljit.RouteJITNative {
				t.Fatalf("not native: %+v", check)
			}
			for _, batch := range []int{1, 10, 32} {
				for offset := range values {
					indexes := make([]int, batch)
					runs := make([]pointRun, batch)
					expected := make([]*point.Point, batch)
					for i := range runs {
						fields := map[string]any{"message": values[(offset+i)%len(values)], "other": `{"b":"independent"}`, "a": int64(42), "b": "keep", "remove_me": "gone", "sentinel": "preserved"}
						actual := newRealScriptPoint("scalar", fields)
						expected[i] = newRealScriptPoint("scalar", fields)
						if i%4 == 1 {
							// A tag forces pre-execution fallback for this record.
							actual.AddTag("a", "42")
							expected[i].AddTag("a", "42")
						}
						if i%4 == 2 {
							actual.AddKVs(point.NewKV("status", "WARNING"))
							expected[i].AddKVs(point.NewKV("status", "WARNING"))
						}
						runs[i] = pointRun{point: actual, script: script}
						indexes[i] = i
						reference := pointRun{point: expected[i], script: script}
						runPipelineGo(point.Logging, &reference, nil)
						if reference.output == nil || reference.dropped {
							t.Fatal("Go unexpectedly dropped JSON record")
						}
						expected[i] = reference.output
					}
					runJITGroup(runner, point.Logging, script, indexes, runs, nil)
					for i, run := range runs {
						if run.output == nil || run.dropped || len(run.created) != 0 {
							t.Fatal("unexpected JIT terminal")
						}
						if equal, reason := equalJSONOraclePoints(run.output, expected[i]); !equal {
							t.Fatalf("batch=%d offset=%d record=%d: %s", batch, offset, i, reason)
						}
					}
				}
			}
		})
	}
}

func TestJITGrokSlotExecutionOracle(t *testing.T) {
	runtimePath := os.Getenv("PLATYPUS_JIT_RUNTIME")
	if runtimePath == "" {
		t.Fatal("requires real PLATYPUS_JIT_RUNTIME")
	}
	runner, err := pljit.NewRunnerWithHost(runtimePath, "pipeline-go-1.4.3-datakit", 16, pljit.NewPipelineGoHost(nil))
	if err != nil {
		t.Fatal(err)
	}
	defer runner.Close()
	values := []any{"one two", "42 1.25 True", "", " \tvalue\n ", "中文 🦀", "a\xffb", int64(42), true, float64(1.5), strings.Repeat("x", 5000), "１２", "a\u2003b", "\v", `\w`}
	for number, source := range []string{
		`grok(message, "%{WORD:word}")`,
		`drop_key(remove_me)
grok(message, "%{WORD:dup} %{WORD:dup}")`,
		`grok(message, "^%{INT:n:int} %{NUMBER:f:float} %{WORD:b:bool}$")`,
		`grok(message, "(?s)^%{GREEDYDATA:all}$", false)`,
		`grok(message, "^%{DATA:empty}$")`,
		`grok(message, "^absent %{WORD:word}")
grok(message, "%{WORD:word}")
grok(other, "^%{INT:n:int}$")`,
		`add_pattern("INNER", "%{WORD:dup}")
grok(message, "%{INNER:outer}")`,
		`grok(message, "^(?P<ascii_word>\\w+)$")`,
		`grok(message, "^(?P<notspace>\\S+)$")`,
		`grok(message, "^(?P<digits>[\\d]+)$")`,
		`grok(message, "(?P<unicode>\\p{Han}+)")`,
		`grok(message, "(?P<notword>[\\W]+)")`,
		`grok(message, "(?P<literal>\\\\w)")`,
	} {
		t.Run(fmt.Sprint(number), func(t *testing.T) {
			script, err := NewPlScriptSimple(point.Logging, "grok-slots.p", source)
			if err != nil {
				t.Fatal(err)
			}
			if check := runner.Check(source); check.Route != pljit.RouteJITNative {
				t.Fatalf("not native: %+v", check)
			}
			for _, batch := range []int{1, 10, 32} {
				for offset := range values {
					indexes := make([]int, batch)
					runs := make([]pointRun, batch)
					expected := make([]*point.Point, batch)
					for i := range runs {
						fields := map[string]any{"message": values[(offset+i)%len(values)], "other": "42", "dup": "keep", "remove_me": "gone", "sentinel": "preserved"}
						if i%7 == 6 {
							delete(fields, "message")
						}
						actual := newRealScriptPoint("grok-slots", fields)
						expected[i] = newRealScriptPoint("grok-slots", fields)
						if i%4 == 1 {
							actual.AddTag("message", "tag value")
							expected[i].AddTag("message", "tag value")
						}
						runs[i] = pointRun{point: actual, script: script}
						indexes[i] = i
						reference := pointRun{point: expected[i], script: script}
						runPipelineGo(point.Logging, &reference, nil)
						if reference.output == nil || reference.dropped {
							t.Fatal("Go unexpectedly dropped Grok record")
						}
						expected[i] = reference.output
					}
					runJITGroup(runner, point.Logging, script, indexes, runs, nil)
					for i, run := range runs {
						if run.output == nil || run.dropped || len(run.created) != 0 {
							t.Fatal("unexpected JIT terminal")
						}
						if equal, reason := equalJSONOraclePoints(run.output, expected[i]); !equal {
							t.Fatalf("batch=%d offset=%d record=%d: %s", batch, offset, i, reason)
						}
					}
				}
			}
		})
	}
}
