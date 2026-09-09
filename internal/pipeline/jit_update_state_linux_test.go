// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build pipeline_jit && linux && (amd64 || arm64)

package pipeline

import (
	"fmt"
	"os"
	"testing"

	"github.com/GuanceCloud/cliutils/point"
	"github.com/GuanceCloud/pipeline-go/constants"
	pljit "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline/jit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline/plval"
)

func TestLiveRemoteUpdatesRetainGoAndNativeCache(t *testing.T) {
	for _, native := range []bool{false, true} {
		t.Run(fmt.Sprintf("native=%v", native), func(t *testing.T) {
			isolateProductionOrchestrationMetrics(t)
			runner, err := pljit.NewRunner(os.Getenv("PLATYPUS_JIT_RUNTIME"), "pipeline-go-1.4.3-datakit", 64)
			if err != nil {
				t.Fatal(err)
			}
			installProductionOrchestrationRunner(t, runner)
			previous, _ := plval.GetManager()
			m := plval.NewScriptManager(nil, nil)
			plval.SetManager(m)
			t.Cleanup(func() { plval.SetManager(previous) })
			if native {
				m.SetJITCheck(func(source string) bool { return runner.Check(source).Route == pljit.RouteJITNative })
				m.SetJITInvalidate(runner.Invalidate)
			}
			if err := m.LoadScriptWithCatChecked(point.Logging, constants.NSDefault, map[string]string{
				"counter.p": `n=cache_get("count"); cache_set("count","retained"); add_key(count,n)`,
				"json.p":    `json(_,value); add_key(result,value+1)`,
				"grok.p":    `grok(_,"%{INT:value}"); cast(value,"int"); add_key(result,value+1)`,
			}, nil); err != nil {
				t.Fatal(err)
			}
			attempt := jitRouteRecordsVec.WithLabelValues("logging", "attempted", "submitted")
			before := readPrometheusCounter(t, attempt)
			run := func(name, message string, count int) ([]*point.Point, error) {
				pts := make([]*point.Point, count)
				for i := range pts {
					pts[i] = newRealScriptPoint(name, map[string]any{"message": message})
				}
				result, err := RunPl(point.Logging, pts, nil)
				if err != nil {
					return nil, err
				}
				return result.Pts(), nil
			}
			if _, err := run("counter", "seed", 1); err != nil {
				t.Fatal(err)
			}
			updates := make(chan error, 1)
			go func() {
				for round := 0; round < 100; round++ {
					candidate, err := m.PrepareRemoteUpdate(plval.RemoteManagerUpdate{ReplaceScripts: true, Scripts: map[point.Category]map[string]string{point.Logging: {"other.p": fmt.Sprintf("add_key(round,%d)", round)}}})
					if err != nil {
						updates <- err
						return
					}
					if round%3 == 0 {
						candidate.Abort()
					} else if err = candidate.Commit(); err != nil {
						updates <- err
						return
					}
				}
				updates <- nil
			}()
			defer func() {
				if err := <-updates; err != nil {
					t.Error(err)
				}
			}()
			for round := 0; round < 100; round++ {
				for _, name := range []string{"counter", "json", "grok"} {
					message := "41"
					if name == "json" {
						message = `{"value":41}`
					}
					pts, err := run(name, message, 32)
					if err != nil {
						t.Fatal(err)
					}
					if len(pts) != 32 {
						t.Fatal("lost points")
					}
					for _, pt := range pts {
						if name == "counter" {
							if pt.Get("count") != "retained" {
								t.Fatalf("round %d cache lost: %v", round, pt.Get("count"))
							}
						} else if pt.Get("result") != int64(42) && pt.Get("result") != float64(42) {
							t.Fatalf("incorrect parse result: %v", pt.Get("result"))
						}
						if pt.Get("_pl_status") == "failed" {
							t.Fatal("script execution failed")
						}
					}
				}
			}
			submitted := readPrometheusCounter(t, attempt) - before
			if native && submitted != 9601 {
				t.Fatalf("native route records=%v", submitted)
			}
			if !native && submitted != 0 {
				t.Fatalf("Go baseline used native: %v", submitted)
			}
		})
	}
}

func TestJITRunPlCastBoundaryContinues(t *testing.T) {
	isolateProductionOrchestrationMetrics(t)
	runner, err := pljit.NewRunner(os.Getenv("PLATYPUS_JIT_RUNTIME"), "pipeline-go-1.4.3-datakit", 16)
	if err != nil {
		t.Fatal(err)
	}
	installProductionOrchestrationRunner(t, runner)
	previous, _ := plval.GetManager()
	m := plval.NewScriptManager(nil, nil)
	plval.SetManager(m)
	t.Cleanup(func() { plval.SetManager(previous) })
	source := `add_key(before_cast,true); cast(n,"int"); add_key(after_cast,true)`
	if err := m.LoadScriptWithCatChecked(point.Logging, constants.NSDefault, map[string]string{"boundary.p": source}, nil); err != nil {
		t.Fatal(err)
	}
	if check := runner.Check(source); check.Route != pljit.RouteJITNative {
		t.Fatalf("expected native cast: %+v", check)
	}
	for _, value := range []string{"42", "9223372036854775807", "-9223372036854775808", "9223372036854775808", "-9223372036854775809", "1e300", "-1e300", "NaN", "+Inf", "-Inf"} {
		t.Run(value, func(t *testing.T) {
			m.SetJITCheck(nil)
			reference, err := RunPl(point.Logging, []*point.Point{newRealScriptPoint("boundary", map[string]any{"message": "keep", "n": value})}, nil)
			if err != nil {
				t.Fatal(err)
			}
			m.SetJITCheck(func(s string) bool { return runner.Check(s).Route == pljit.RouteJITNative })
			m.SetJITInvalidate(runner.Invalidate)
			attempt := jitRouteRecordsVec.WithLabelValues("logging", "attempted", "submitted")
			before := readPrometheusCounter(t, attempt)
			actual, err := RunPl(point.Logging, []*point.Point{newRealScriptPoint("boundary", map[string]any{"message": "keep", "n": value})}, nil)
			if err != nil {
				t.Fatal(err)
			}
			if readPrometheusCounter(t, attempt)-before != 1 {
				t.Fatal("native record was not submitted")
			}
			if len(actual.Pts()) != 1 || len(reference.Pts()) != 1 {
				t.Fatal("missing output")
			}
			got, want := actual.Pts()[0], reference.Pts()[0]
			if got.Get("before_cast") != true || got.Get("after_cast") != true {
				t.Fatalf("cast interrupted script: n=%v after=%v", got.Get("n"), got.Get("after_cast"))
			}
			if equal, reason := equalJSONOraclePoints(got, want); !equal {
				t.Fatal(reason)
			}
		})
	}
}
