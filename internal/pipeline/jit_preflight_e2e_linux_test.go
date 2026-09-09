// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build pipeline_jit && jitbench && linux && (amd64 || arm64)

package pipeline

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/GuanceCloud/cliutils/logger"
	"github.com/GuanceCloud/cliutils/point"
	"github.com/GuanceCloud/pipeline-go/constants"
	"github.com/GuanceCloud/pipeline-go/ptinput/funcs"
	pljit "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline/jit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline/plval"
)

type preflightCase struct {
	Name, Source string
	Inputs       []map[string]any
}
type preflightRow struct {
	Case        string   `json:"case"`
	Context     string   `json:"context"`
	Batch       int      `json:"batch"`
	Route       string   `json:"route"`
	Native      int      `json:"native_submitted"`
	Output      int      `json:"output"`
	Differences []string `json:"differences,omitempty"`
}
type preflightSnapshot struct {
	result *ScriptResult
	err    string
}

// This is the production Pipeline entry, not a direct native/function oracle.
// Global manager and JIT generations are deliberately changed serially.
func TestJITPreflightE2E(t *testing.T) {
	path := os.Getenv("PLATYPUS_JIT_RUNTIME")
	if path == "" {
		t.Fatal("requires a frozen ordinary JIT library")
	}
	if err := logger.InitRoot(&logger.Option{Level: logger.ERROR, Flags: logger.OPT_DEFAULT}); err != nil {
		t.Fatal(err)
	}
	funcs.InitLog()
	var rows []preflightRow
	t.Cleanup(func() {
		if path := os.Getenv("JIT_PREFLIGHT_REPORT"); path != "" {
			data, err := json.MarshalIndent(rows, "", "  ")
			if err != nil {
				t.Error(err)
				return
			}
			if err := os.WriteFile(path, data, 0644); err != nil {
				t.Error(err)
			}
		}
	})
	var cases []preflightCase
	for _, c := range realIntegrationCorpora() {
		source, err := os.ReadFile(filepath.Join("testdata", "jit-real-scripts", c.script))
		if err != nil {
			t.Fatal(err)
		}
		samples := append(append([]string{}, c.samples...), extraRealScriptSamples(c.script)...)
		samples = append(samples, "not a recognized log format", "", "broken\nmultiline\nlog")
		cse := preflightCase{Name: "builtin/" + c.script, Source: string(source)}
		for _, message := range samples {
			cse.Inputs = append(cse.Inputs, map[string]any{"message": message, "sentinel": "keep"})
		}
		cse.Inputs = append(cse.Inputs, map[string]any{"sentinel": "missing-message"}, map[string]any{"message": "unmatched", "time": "invalid-time", "status": "existing"})
		cases = append(cases, cse)
	}
	for _, w := range loadLogMatrixWorkloads(t) {
		fields := map[string]any{"message": w.message, "sentinel": "keep", "status": "unknown"}
		for key, value := range w.extraFields {
			fields[key] = value
		}
		cases = append(cases, preflightCase{Name: w.name, Source: w.source, Inputs: []map[string]any{fields}})
	}
	cases = append(cases,
		preflightCase{Name: "dynamic/type-change", Source: `key=pt_kvs_get("selector"); pt_kvs_set(key,42); pt_kvs_set(key,"changed"); pt_kvs_set(key,{"x":[1,2]},false,true); pt_kvs_set("copy",pt_kvs_get(key,true),false,true)`, Inputs: []map[string]any{{"selector": "dynamic", "message": "keep"}, {"selector": "existing_tag", "message": "keep"}}},
		preflightCase{Name: "lifecycle/created", Source: `create_point("child", {"source":"test"}, {"value":1}, category="L", ts=1); add_key(after,true)`, Inputs: []map[string]any{{"message": "keep"}}},
		preflightCase{Name: "lifecycle/drop", Source: `if message == "drop" { drop() }; add_key(after,true)`, Inputs: []map[string]any{{"message": "drop"}, {"message": "keep"}}},
		preflightCase{Name: "errors/committed-prefix", Source: `add_key(before,true); x=1/divisor; add_key(after,true)`, Inputs: []map[string]any{{"message": "error", "divisor": int64(0)}}},
		preflightCase{Name: "json/malformed", Source: `doc=load_json(message); pt_kvs_set("parsed",doc)`, Inputs: []map[string]any{{"message": "{"}, {"message": ""}, {"message": "null"}, {"message": `{"x":1}`}}},
		preflightCase{Name: "dynamic/raw-bytes", Source: `pt_kvs_set(pt_kvs_get("selector"),message)`, Inputs: []map[string]any{{"selector": "key\xff", "message": "value\xff\x00"}}},
	)
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			previous, _ := plval.GetManager()
			manager := plval.NewScriptManager(nil, nil)
			plval.SetManager(manager)
			t.Cleanup(func() {
				if err := closeJITGeneration(jitRunners.replace(nil)); err != nil {
					t.Error(err)
				}
				plval.SetManager(previous)
			})
			if err := initJIT(&plval.PipelineCfg{JIT: &plval.JITCfg{Enabled: false}}, ""); err != nil {
				t.Fatal(err)
			}
			if err := manager.LoadScriptWithCatChecked(point.Logging, constants.NSDefault, map[string]string{"main.p": c.Source}, nil); err != nil {
				t.Fatal(err)
			}
			manager.UpdateDefaultScript(map[point.Category]string{point.Logging: "main.p"})
			sizes := []int{1, 10, 128}
			if c.Name == "lifecycle/drop" || c.Name == "log-load-json-dynamic-pt-set-large" {
				sizes = append(sizes, 4097)
			}
			modes := []string{"background", "cancellable", "record-timeout"}
			expected := map[string]preflightSnapshot{}
			run := func(size int, mode string) preflightSnapshot {
				pts := make([]*point.Point, size)
				for i := range pts {
					pts[i] = newRealScriptPoint(c.Name, c.Inputs[i%len(c.Inputs)])
				}
				ctx := context.Background()
				cancel := func() {}
				if mode == "cancellable" {
					ctx, cancel = context.WithCancel(ctx)
				}
				if mode == "record-timeout" {
					ctx = pljit.WithRecordTimeout(ctx, 30*time.Second)
				}
				defer cancel()
				result, err := RunPlContext(ctx, point.Logging, pts, nil)
				snapshot := preflightSnapshot{result: result}
				if err != nil {
					snapshot.err = err.Error()
				}
				return snapshot
			}
			counter := jitRouteRecordsVec.WithLabelValues("logging", "attempted", "submitted")
			beforeGo := readPrometheusCounter(t, counter)
			for _, mode := range modes {
				for _, size := range sizes {
					key := fmt.Sprintf("%s/%d", mode, size)
					expected[key] = run(size, mode)
				}
			}
			if readPrometheusCounter(t, counter) != beforeGo {
				t.Fatal("disabled JIT submitted native records")
			}
			if err := initJIT(&plval.PipelineCfg{JIT: &plval.JITCfg{Enabled: true, RuntimePath: path, MaxCachedPrograms: 64}}, ""); err != nil {
				t.Fatal(err)
			}
			lease := jitRunners.acquire()
			check := lease.runner().Check(c.Source)
			lease.release()
			for _, mode := range modes {
				for _, size := range sizes {
					key := fmt.Sprintf("%s/%d", mode, size)
					before := readPrometheusCounter(t, counter)
					actual := run(size, mode)
					row := preflightRow{Case: c.Name, Context: mode, Batch: size, Route: string(check.Route), Native: int(readPrometheusCounter(t, counter) - before)}
					want := expected[key]
					if actual.err != want.err {
						row.Differences = append(row.Differences, fmt.Sprintf("error: Go=%q JIT=%q", want.err, actual.err))
					}
					compare := func(label string, got, want []*point.Point) {
						if len(got) != len(want) {
							row.Differences = append(row.Differences, fmt.Sprintf("%s count Go=%d JIT=%d", label, len(want), len(got)))
							return
						}
						for i := range got {
							if equal, why := preflightEqualPoints(got[i], want[i]); !equal {
								row.Differences = append(row.Differences, fmt.Sprintf("%s[%d]: %s; Go=%s JIT=%s", label, i, why, want[i].Pretty(), got[i].Pretty()))
								if len(row.Differences) >= 3 {
									return
								}
							}
						}
					}
					if actual.result == nil || want.result == nil {
						if (actual.result == nil) != (want.result == nil) {
							row.Differences = append(row.Differences, "nil result mismatch")
						}
					} else {
						row.Output = len(actual.result.Pts())
						compare("points", actual.result.Pts(), want.result.Pts())
						compare("offload", actual.result.PtsOffload(), want.result.PtsOffload())
						cats := map[point.Category]bool{}
						for cat := range actual.result.PtsCreated() {
							cats[cat] = true
						}
						for cat := range want.result.PtsCreated() {
							cats[cat] = true
						}
						for cat := range cats {
							compare("created/"+cat.String(), actual.result.PtsCreated()[cat], want.result.PtsCreated()[cat])
						}
					}
					if check.Route == pljit.RouteJITNative && row.Native != size {
						row.Differences = append(row.Differences, fmt.Sprintf("native admission: submitted %d of %d", row.Native, size))
					}
					rows = append(rows, row)
					if len(row.Differences) > 0 {
						t.Errorf("%s: %v", key, row.Differences)
					}
					if actual.result != nil {
						actual.result.Release()
					}
					if want.result != nil {
						want.result.Release()
					}
				}
			}
		})
	}
}

// Concurrent calls use the real generation leases and per-call native scratch.
func TestJITPreflightConcurrentE2E(t *testing.T) {
	path := os.Getenv("PLATYPUS_JIT_RUNTIME")
	if path == "" {
		t.Fatal("requires native runtime")
	}
	previous, _ := plval.GetManager()
	manager := plval.NewScriptManager(nil, nil)
	plval.SetManager(manager)
	t.Cleanup(func() {
		if err := closeJITGeneration(jitRunners.replace(nil)); err != nil {
			t.Error(err)
		}
		plval.SetManager(previous)
	})
	source := `doc=load_json(message); pt_kvs_set(doc["key"],doc["value"]); pt_kvs_set("copy",pt_kvs_get(doc["key"]))`
	if err := initJIT(&plval.PipelineCfg{JIT: &plval.JITCfg{Enabled: false}}, ""); err != nil {
		t.Fatal(err)
	}
	if err := manager.LoadScriptWithCatChecked(point.Logging, constants.NSDefault, map[string]string{"main.p": source}, nil); err != nil {
		t.Fatal(err)
	}
	manager.UpdateDefaultScript(map[point.Category]string{point.Logging: "main.p"})
	makePoints := func() []*point.Point {
		pts := make([]*point.Point, 16)
		for i := range pts {
			pts[i] = newRealScriptPoint("concurrent", map[string]any{"message": `{"key":"dynamic","value":42}`})
		}
		return pts
	}
	reference, err := RunPlContext(context.Background(), point.Logging, makePoints(), nil)
	if err != nil {
		t.Fatal(err)
	}
	defer reference.Release()
	for _, enabled := range []bool{false, true} {
		if err := initJIT(&plval.PipelineCfg{JIT: &plval.JITCfg{Enabled: enabled, RuntimePath: path, MaxCachedPrograms: 16}}, ""); err != nil {
			t.Fatal(err)
		}
		counter := jitRouteRecordsVec.WithLabelValues("logging", "attempted", "submitted")
		before := readPrometheusCounter(t, counter)
		var wg sync.WaitGroup
		for worker := 0; worker < 8; worker++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for iteration := 0; iteration < 50; iteration++ {
					result, err := RunPlContext(context.Background(), point.Logging, makePoints(), nil)
					if err != nil || result == nil {
						t.Errorf("enabled=%v: result=%v error=%v", enabled, result, err)
						return
					}
					if len(result.Pts()) != len(reference.Pts()) {
						t.Errorf("output count mismatch")
						result.Release()
						return
					}
					for i, pt := range result.Pts() {
						if equal, why := preflightEqualPoints(pt, reference.Pts()[i]); !equal {
							t.Errorf("enabled=%v: %s", enabled, why)
						}
					}
					result.Release()
				}
			}()
		}
		wg.Wait()
		delta := readPrometheusCounter(t, counter) - before
		want := float64(0)
		if enabled {
			want = 8 * 50 * 16
		}
		if delta != want {
			t.Errorf("enabled=%v native=%v want=%v", enabled, delta, want)
		}
	}
}

// EqualWithReason with excluded keys does not check key counts and walks only
// its receiver. Both directions are required to catch missing OR extra fields.
func preflightEqualPoints(got, want *point.Point) (bool, string) {
	if got == nil || want == nil {
		return got == want, "nil Point mismatch"
	}
	if ok, why := got.EqualWithReason(want, point.EqualWithoutKeys(plFieldCost)); !ok {
		return false, why
	}
	return want.EqualWithReason(got, point.EqualWithoutKeys(plFieldCost))
}

func TestJITPreflightComparator(t *testing.T) {
	expected := newRealScriptPoint("test", map[string]any{"message": "x", "count": int64(1)})
	for _, fields := range []map[string]any{
		{"message": "x"},
		{"message": "x", "count": int64(1), "extra": true},
		{"message": "x", "count": float64(1)},
	} {
		if ok, _ := preflightEqualPoints(newRealScriptPoint("test", fields), expected); ok {
			t.Fatalf("accepted field mismatch: %#v", fields)
		}
	}
	actual := newRealScriptPoint("test", map[string]any{"message": "x", "count": int64(1), plFieldCost: float64(0.1)})
	if ok, why := preflightEqualPoints(actual, expected); !ok {
		t.Fatal("execution cost should be excluded:", why)
	}
	changedTag := newRealScriptPoint("test", map[string]any{"message": "x", "count": int64(1)})
	changedTag.AddTag("other_tag", "value")
	if ok, _ := preflightEqualPoints(changedTag, expected); ok {
		t.Fatal("accepted tag mismatch")
	}
}
