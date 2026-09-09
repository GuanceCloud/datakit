// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build pipeline_jit && linux && (amd64 || arm64)

package pipeline

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/GuanceCloud/cliutils/point"
	"github.com/GuanceCloud/pipeline-go/lang"
	"github.com/GuanceCloud/pipeline-go/lang/platypus"
	"github.com/GuanceCloud/pipeline-go/ptinput"
	"github.com/GuanceCloud/pipeline-go/ptinput/funcs"
	"github.com/GuanceCloud/platypus/pkg/engine"
	pljit "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline/jit"
)

// Strict differential over the real module ABI and DataKit Point application.
// No fallback is permitted, including compile failures and terminal errors.
func TestJITModuleStrictDifferential(t *testing.T) {
	path := os.Getenv("PLATYPUS_JIT_RUNTIME")
	if path == "" {
		t.Fatal("requires real PLATYPUS_JIT_RUNTIME")
	}
	for _, child := range []string{
		`local = 9; add_key(child, local)`,
		`add_key(child, 1); exit(); add_key(unreachable, true)`,
		`add_key(child, 2); value = 1 / divisor; add_key(result, value)`,
		`use("leaf.p"); add_key(child, 3)`,
		`use("dir/nested.p"); add_key(child, 4)`,
		`create_point("generated", {}, {"divisor": divisor}, after_use="parse.p"); add_key(child, 5)`,
	} {
		for _, size := range []int{1, 8, 128} {
			t.Run(fmt.Sprintf("%s/batch%d", child, size), func(t *testing.T) {
				sources := map[string]string{
					"main.p":  `local = 7; add_key(before, true); use("child.p"); add_key(parent_local, local); add_key(after, true)`,
					"child.p": child, "leaf.p": `add_key(leaf, true); exit(); add_key(unreachable, true)`,
					"dir/nested.p": `use("leaf.p")`,
					"dir/leaf.p":   `add_key(wrong_relative_resolution, true)`,
					"parse.p":      `add_key(child_before, true); quotient = 1 / divisor; drop()`,
				}
				scripts, failures := platypus.NewScripts(sources, lang.WithCat(point.Logging))
				defer func() {
					for _, script := range scripts {
						script.Cleanup()
					}
				}()
				if len(failures) != 0 {
					t.Fatal(failures)
				}
				runner, err := pljit.NewRunnerWithHost(path, "pipeline-go-1.4.3-datakit", 4, pljit.NewPipelineGoHost(nil))
				if err != nil {
					t.Fatal(err)
				}
				defer runner.Close()
				snapshot, err := pljit.NewModuleSnapshot("main.p", sources)
				if err != nil {
					t.Fatal(err)
				}
				if check := runner.CheckModules(snapshot); check.Route != pljit.RouteJITNative {
					t.Fatalf("not native: %+v", check)
				}
				for start := 0; start < 129; start += size {
					count := min(size, 129-start)
					actual := make([]*point.Point, count)
					expected := make([]*point.Point, count)
					errors := make([]error, count)
					for i := range actual {
						fields := map[string]any{"message": "keep", "divisor": int64((start + i) % 3), "sequence": int64(start + i)}
						actual[i] = newRealScriptPoint("module", fields)
						wrapped := ptinput.PtWrap(point.Logging, newRealScriptPoint("module", fields))
						if runErr := scripts["main.p"].Run(wrapped, nil, nil); runErr != nil {
							errors[i] = runErr
						}
						expected[i] = wrapped.Point()
					}
					input, err := encodeProjectedJITPoints(point.Logging, actual, pljit.InputProjection{})
					if err != nil {
						t.Fatal(err)
					}
					batch, err := runner.ProcessModules(snapshot, input)
					if err != nil {
						t.Fatal(err)
					}
					if len(batch.Records) != count {
						t.Fatal("record count mismatch")
					}
					for i, record := range batch.Records {
						if (record.Status == pljit.TerminalError) != (errors[i] != nil) {
							t.Fatalf("terminal mismatch: %s / %v", record.Error, errors[i])
						}
						if errors[i] != nil && !record.CommitPrefixError {
							t.Fatal("missing error prefix")
						}
						if errors[i] != nil && strings.Contains(child, "after_use") {
							var diagnostic struct {
								SourceName string   `json:"source_name"`
								CallChain  []string `json:"call_chain"`
							}
							if err := json.Unmarshal(record.Error, &diagnostic); err != nil {
								t.Fatal(err)
							}
							if diagnostic.SourceName != "parse.p" || !strings.Contains(strings.Join(diagnostic.CallChain, "/"), "create_point(after_use=parse.p)") || !strings.Contains(strings.Join(diagnostic.CallChain, "/"), "use(child.p)") {
								t.Fatalf("missing nested error chain: %s", record.Error)
							}
						}
						created, dropped, err := applyJITRecord(point.Logging, actual[i], record, uint64(i), nil)
						if err != nil {
							t.Fatal(err)
						}
						if dropped || len(created) != 0 || len(record.Emitted) != 0 {
							t.Fatal("unexpected drop or children")
						}
						if equal, reason := actual[i].EqualWithReason(expected[i]); !equal {
							t.Fatalf("point mismatch: %s native=%v Go=%v", reason, actual[i].KVMap(), expected[i].KVMap())
						}
					}
				}
			})
		}
	}
}

// A module call is also a value expression in pipeline-go: use returns nil after
// the child frame completes.  Keep this separate from the statement-form matrix
// above so expression placement, child exit isolation, and error boundaries are
// all checked through the real shared-library ABI.
func TestJITNestedUseStrictDifferential(t *testing.T) {
	path := os.Getenv("PLATYPUS_JIT_RUNTIME")
	if path == "" {
		t.Fatal("requires real PLATYPUS_JIT_RUNTIME")
	}
	for _, tc := range []struct {
		name  string
		child string
	}{
		{name: "normal", child: `add_key(child, true)`},
		{name: "exit", child: `add_key(child, true); exit(); add_key(unreachable, true)`},
		{name: "error", child: `add_key(child, true); value = 1 / divisor; add_key(unreachable, true)`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			sources := map[string]string{
				"main.p":  `add_key(before, true); add_key(returned_nil, use("child.p") == nil); add_key(after, true)`,
				"child.p": tc.child,
			}
			scripts, failures := platypus.NewScripts(sources, lang.WithCat(point.Logging))
			defer func() {
				for _, script := range scripts {
					script.Cleanup()
				}
			}()
			if len(failures) != 0 {
				t.Fatal(failures)
			}
			runner, err := pljit.NewRunner(path, "pipeline-go-1.4.3-datakit", 1)
			if err != nil {
				t.Fatal(err)
			}
			defer runner.Close()
			snapshot, err := pljit.NewModuleSnapshot("main.p", sources)
			if err != nil {
				t.Fatal(err)
			}
			if check := runner.CheckModules(snapshot); check.Route != pljit.RouteJITNative {
				t.Fatalf("not native: %+v", check)
			}

			actual := newRealScriptPoint("module", map[string]any{"divisor": int64(0)})
			wrapped := ptinput.PtWrap(point.Logging, newRealScriptPoint("module", map[string]any{"divisor": int64(0)}))
			wantErr := scripts["main.p"].Run(wrapped, nil, nil)
			expected := wrapped.Point()
			input, err := encodeProjectedJITPoints(point.Logging, []*point.Point{actual}, pljit.InputProjection{})
			if err != nil {
				t.Fatal(err)
			}
			batch, err := runner.ProcessModules(snapshot, input)
			if err != nil {
				t.Fatal(err)
			}
			if len(batch.Records) != 1 {
				t.Fatalf("record count = %d", len(batch.Records))
			}
			record := batch.Records[0]
			if (record.Status == pljit.TerminalError) != (wantErr != nil) {
				t.Fatalf("terminal mismatch: native=%s Go=%v", record.Error, wantErr)
			}
			if wantErr != nil && !record.CommitPrefixError {
				t.Fatal("missing committed child/error prefix")
			}
			created, dropped, err := applyJITRecord(point.Logging, actual, record, 0, nil)
			if err != nil {
				t.Fatal(err)
			}
			if dropped || len(created) != 0 || len(record.Emitted) != 0 {
				t.Fatal("unexpected drop or children")
			}
			if equal, reason := actual.EqualWithReason(expected); !equal {
				t.Fatalf("point mismatch: %s native=%v Go=%v", reason, actual.KVMap(), expected.KVMap())
			}
		})
	}
}

func TestJITModuleProductionBatchApplication(t *testing.T) {
	path := os.Getenv("PLATYPUS_JIT_RUNTIME")
	if path == "" {
		t.Fatal("requires real PLATYPUS_JIT_RUNTIME")
	}
	sources := map[string]string{
		"main.p":  `local = 7; use("child.p"); add_key(parent_local, local)`,
		"child.p": `local = 9; add_key(child, local); if sequence == 0 { exit() }; add_key(after, true)`,
	}
	scripts, failures := platypus.NewScripts(sources, lang.WithCat(point.Logging))
	defer func() {
		for _, script := range scripts {
			script.Cleanup()
		}
	}()
	if len(failures) != 0 {
		t.Fatal(failures)
	}
	runner, err := pljit.NewRunner(path, "pipeline-go-1.4.3-datakit", 4)
	if err != nil {
		t.Fatal(err)
	}
	defer runner.Close()
	snapshot, err := pljit.NewScopedModuleSnapshot("test", "logging", "main.p", sources)
	if err != nil {
		t.Fatal(err)
	}
	bound, err := runner.BindModules(snapshot)
	if err != nil {
		t.Fatal(err)
	}
	defer bound.Close()
	// Invalidation cannot replace the code between projection and processing.
	runner.InvalidateModules(snapshot)
	for _, size := range []int{1, 8, 128} {
		actual, expected := make([]pointRun, 129), make([]pointRun, 129)
		for i := range actual {
			fields := map[string]any{"sequence": int64(i % 3), "message": "keep"}
			actual[i] = pointRun{point: newRealScriptPoint("module", fields), script: scripts["main.p"]}
			expected[i] = pointRun{point: newRealScriptPoint("module", fields), script: scripts["main.p"]}
			runPipelineGo(point.Logging, &expected[i], nil)
		}
		for start := 0; start < len(actual); start += size {
			indexes := make([]int, min(size, len(actual)-start))
			for i := range indexes {
				indexes[i] = start + i
			}
			runJITGroup(bound, point.Logging, scripts["main.p"], indexes, actual, nil)
		}
		for i := range actual {
			if actual[i].output == nil || expected[i].output == nil {
				t.Fatal("missing output")
			}
			if equal, reason := actual[i].output.EqualWithReason(expected[i].output); !equal {
				t.Fatalf("record %d: %s", i, reason)
			}
		}
	}
	if err := bound.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := bound.Process(sources["main.p"], nil); err == nil {
		t.Fatal("closed binding still executable")
	}
}

func TestJITAfterUseProductionDifferential(t *testing.T) {
	path := os.Getenv("PLATYPUS_JIT_RUNTIME")
	if path == "" {
		t.Fatal("requires real PLATYPUS_JIT_RUNTIME")
	}
	for _, child := range []string{`add_key(parsed, value + 1)`, `add_key(parsed, true); exit(); add_key(unreachable, true)`, `drop()`,
		`create_point("grandchild", {}, {"value": value}); add_key(parsed, true)`,
		`create_point("grandchild", {}, {"value": value}, after_use="leaf.p"); add_key(parsed, true)`,
		`create_point("grandchild", {}, {"value": value}, after_use="leaf.p"); drop()`,
	} {
		t.Run(child, func(t *testing.T) {
			sources := map[string]string{"main.p": `create_point("child", {}, {"value": sequence}, after_use="child.p"); add_key(parent, true)`, "child.p": child,
				"leaf.p": `add_key(parsed_leaf, true)`,
			}
			scripts, failures := platypus.NewScripts(sources, lang.WithCat(point.Logging))
			defer func() {
				for _, script := range scripts {
					script.Cleanup()
				}
			}()
			if len(failures) != 0 {
				t.Fatal(failures)
			}
			runner, err := pljit.NewRunner(path, "pipeline-go-1.4.3-datakit", 4)
			if err != nil {
				t.Fatal(err)
			}
			defer runner.Close()
			snapshot, err := pljit.NewModuleSnapshot("main.p", sources)
			if err != nil {
				t.Fatal(err)
			}
			bound, err := runner.BindModules(snapshot)
			if err != nil {
				t.Fatal(err)
			}
			defer bound.Close()
			for _, size := range []int{1, 8, 128} {
				actual, expected := make([]pointRun, 129), make([]pointRun, 129)
				for i := range actual {
					fields := map[string]any{"sequence": int64(i), "message": "keep"}
					actual[i] = pointRun{point: newRealScriptPoint("parent", fields), script: scripts["main.p"]}
					expected[i] = pointRun{point: newRealScriptPoint("parent", fields), script: scripts["main.p"]}
					runPipelineGo(point.Logging, &expected[i], nil)
				}
				for start := 0; start < 129; start += size {
					indexes := make([]int, min(size, 129-start))
					for i := range indexes {
						indexes[i] = start + i
					}
					runJITGroup(bound, point.Logging, scripts["main.p"], indexes, actual, nil)
				}
				for i := range actual {
					if actual[i].output == nil || expected[i].output == nil {
						t.Fatal("missing parent")
					}
					if equal, reason := actual[i].output.EqualWithReason(expected[i].output); !equal {
						t.Fatalf("parent: %s", reason)
					}
					if len(actual[i].created) != len(expected[i].created) {
						t.Fatalf("child categories: %v vs %v", actual[i].created, expected[i].created)
					}
					for category, want := range expected[i].created {
						got := actual[i].created[category]
						if len(got) != len(want) {
							t.Fatal("child count differs")
						}
						for j := range want {
							if equal, reason := got[j].EqualWithReason(want[j]); !equal {
								t.Fatalf("child: %s", reason)
							}
						}
					}
				}
			}
		})
	}
}

// Oracle for native module call frames, not evidence of JIT support.
// Based on pipeline-go v1.4.3 ptinput/funcs/fn_use_test.go (TestUse),
// MIT License, Copyright 2021-present Guance, Inc.; extended with call isolation.
func TestPipelineGoUseCallFrameOracle(t *testing.T) {
	cases := []struct {
		name, child string
		fails       bool
		wantChild   int64
	}{
		{"normal", `local = 9; add_key(child, local)`, false, 9},
		{"child_exit", `add_key(child, 1); exit(); add_key(unreachable, true)`, false, 1},
		{"child_error", `add_key(child, 2); value = 1 / divisor; add_key(unreachable, true)`, true, 2},
		{"nested_exit", `use("leaf.p"); add_key(child, 3)`, false, 3},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			scripts, failures := engine.ParseScript(map[string]string{
				"parent.p": `local = 7; add_key(before, true); use("child.p"); add_key(parent_local, local); add_key(after, true)`,
				"child.p":  tc.child,
				"leaf.p":   `exit(); add_key(unreachable, true)`,
			}, funcs.FuncsMap, funcs.FuncsCheckMap)
			if len(failures) != 0 || scripts["parent.p"] == nil {
				t.Fatalf("parse: %v", failures)
			}
			for i := 0; i < 3; i++ {
				input := ptinput.PtWrap(point.Logging, newRealScriptPoint("use", map[string]any{"message": "test", "divisor": int64(0)}))
				err := scripts["parent.p"].Run(input, nil)
				if (err != nil) != tc.fails {
					t.Fatalf("error: %v", err)
				}
				pt := input.Point()
				if pt.Get("before") != true || pt.Get("child") != tc.wantChild || pt.Get("unreachable") != nil {
					t.Fatalf("incorrect shared point/prefix: %v", pt)
				}
				if tc.fails {
					if pt.Get("after") != nil || pt.Get("parent_local") != nil {
						t.Fatal("parent continued after child error")
					}
				} else if pt.Get("after") != true || pt.Get("parent_local") != int64(7) {
					t.Fatalf("child changed parent frame or exit state: %v", pt)
				}
			}
		})
	}
}

// Retains the valid and invalid dependency sets from upstream TestUse.
func TestPipelineGoUseDependencyOracle(t *testing.T) {
	scripts, failures := engine.ParseScript(map[string]string{
		"a.p": `if true { use("b.p") }`,
		"b.p": `add_key(b, 1)`,
		"d.p": `use("c.p")`,
		"c.p": `use("a.p"); use("d.p"); use("fcName.p")`,
	}, funcs.FuncsMap, funcs.FuncsCheckMap)
	if len(scripts) != 2 || len(failures) != 2 {
		t.Fatalf("valid=%v invalid=%v", scripts, failures)
	}
	for _, name := range []string{"a.p", "b.p"} {
		if scripts[name] == nil {
			t.Fatalf("missing valid script %s", name)
		}
		input := ptinput.PtWrap(point.Logging, newRealScriptPoint("use", map[string]any{"message": "unchanged"}))
		if err := scripts[name].Run(input, nil); err != nil {
			t.Fatal(err)
		}
		if input.Point().Get("b") != int64(1) || input.Point().Get("message") != "unchanged" || input.Dropped() {
			t.Fatalf("unexpected output: %v", input.Point())
		}
	}
	for _, name := range []string{"c.p", "d.p"} {
		if failures[name] == nil {
			t.Fatalf("missing dependency failure for %s", name)
		}
	}
}
