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
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"github.com/GuanceCloud/pipeline-go/constants"
	plstats "github.com/GuanceCloud/pipeline-go/stats"
	"github.com/prometheus/client_golang/prometheus"
	pljit "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline/jit"
	plval "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline/plval"
)

const productionOrchestrationBatchSize = 128

func TestRunPlContextCancelsBothEngines(t *testing.T) {
	path := os.Getenv("PLATYPUS_JIT_RUNTIME")
	if path == "" {
		t.Fatal("requires real PLATYPUS_JIT_RUNTIME")
	}
	for _, native := range []bool{false, true} {
		t.Run(map[bool]string{false: "go", true: "rust"}[native], func(t *testing.T) {
			isolateProductionOrchestrationMetrics(t)
			runner, err := pljit.NewRunner(path, "pipeline-go-1.4.3-datakit", 4)
			if err != nil {
				t.Fatal(err)
			}
			installProductionOrchestrationRunner(t, runner)
			previous, _ := plval.GetManager()
			manager := plval.NewScriptManager(nil, nil)
			plval.SetManager(manager)
			t.Cleanup(func() { plval.SetManager(previous) })
			if native {
				manager.SetJITModuleHooks(func(s *pljit.ModuleSnapshot) bool {
					return s.HasDependencies() && runner.CheckModules(s).Route == pljit.RouteJITNative
				}, runner.InvalidateModules)
			}
			if err := manager.LoadScriptWithCatChecked(point.Logging, constants.NSDefault, map[string]string{
				"cancel.p": `use("child.p"); add_key(after, true)`,
				"child.p":  "for ; repeat; {}\nadd_key(done, true)",
			}, nil); err != nil {
				t.Fatal(err)
			}
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Millisecond)
			defer cancel()
			result, err := RunPlContext(ctx, point.Logging, []*point.Point{newRealScriptPoint("cancel", map[string]any{"repeat": true})}, nil)
			if !errors.Is(err, context.DeadlineExceeded) {
				t.Fatalf("expected request deadline, got %v", err)
			}
			if result == nil || len(result.Pts()) != 1 {
				t.Fatal("missing partial result")
			}
			if result.Pts()[0].Get("after") != nil {
				t.Fatal("continued after cancellation")
			}
			result, err = RunPlContext(context.Background(), point.Logging, []*point.Point{newRealScriptPoint("cancel", map[string]any{"repeat": false})}, nil)
			if err != nil || len(result.Pts()) != 1 || result.Pts()[0].Get("after") != true {
				t.Fatalf("next request failed: %v", err)
			}
		})
	}
}

func TestJITModuleAutomaticProductionRoute(t *testing.T) {
	path := os.Getenv("PLATYPUS_JIT_RUNTIME")
	if path == "" {
		t.Fatal("requires real PLATYPUS_JIT_RUNTIME")
	}
	isolateProductionOrchestrationMetrics(t)
	runner, err := pljit.NewRunner(path, "pipeline-go-1.4.3-datakit", 8)
	if err != nil {
		t.Fatal(err)
	}
	installProductionOrchestrationRunner(t, runner)
	previous, _ := plval.GetManager()
	manager := plval.NewScriptManager(nil, nil)
	plval.SetManager(manager)
	t.Cleanup(func() { plval.SetManager(previous) })
	manager.SetJITModuleHooks(func(snapshot *pljit.ModuleSnapshot) bool {
		return snapshot.HasDependencies() && runner.CheckModules(snapshot).Route == pljit.RouteJITNative
	}, runner.InvalidateModules)
	for _, version := range []int{1, 2} {
		child := "add_key(version, 1)"
		if version == 2 {
			child = "add_key(version, 2)"
		}
		if err := manager.LoadScriptWithCatChecked(point.Logging, constants.NSDefault,
			map[string]string{"modules.p": `use("child.p"); add_key(parent, true)`, "child.p": child,
				"after.p": `create_point("generated", {}, {}, category="logging", after_use="child.p"); add_key(parent, true)`,
			}, nil); err != nil {
			t.Fatal(err)
		}
		attempted := jitRouteRecordsVec.WithLabelValues(point.Logging.String(), "attempted", "submitted")
		before := readPrometheusCounter(t, attempted)
		points := []*point.Point{newRealScriptPoint("modules", map[string]any{"message": "keep"})}
		result, err := RunPl(point.Logging, points, nil)
		if err != nil {
			t.Fatal(err)
		}
		if len(result.Pts()) != 1 || result.Pts()[0].Get("version") != int64(version) || result.Pts()[0].Get("parent") != true {
			t.Fatalf("bad output: %+v", result.Pts())
		}
		if readPrometheusCounter(t, attempted)-before != 1 {
			t.Fatal("module script was not submitted to native engine")
		}
		for _, size := range []int{1, 8, 128} {
			points := make([]*point.Point, size)
			for i := range points {
				points[i] = newRealScriptPoint("after", map[string]any{"message": "keep"})
			}
			before := readPrometheusCounter(t, attempted)
			result, err := RunPl(point.Logging, points, nil)
			if err != nil {
				t.Fatal(err)
			}
			if readPrometheusCounter(t, attempted)-before != float64(size) {
				t.Fatal("after_use did not reach native engine")
			}
			children := result.PtsCreated()[point.Logging]
			if len(result.Pts()) != size || len(children) != size {
				t.Fatalf("parent/child count differs: %d/%d", len(result.Pts()), len(children))
			}
			for i, parent := range result.Pts() {
				if parent.Get("parent") != true || parent.Get("version") != nil {
					t.Fatal("child mutated parent")
				}
				if children[i].Get("version") != int64(version) {
					t.Fatalf("stale child version: %v", children[i].Get("version"))
				}
			}
		}
	}
}

type productionOrchestrationStats struct {
	points float64
	drops  float64
	errors float64
}

func (*productionOrchestrationStats) Metrics() []prometheus.Collector { return nil }
func (*productionOrchestrationStats) WriteEvent(*plstats.ChangeEvent, map[string]string) {
}
func (*productionOrchestrationStats) ReadEvents(events []*plstats.ChangeEvent) []*plstats.ChangeEvent {
	return events
}
func (*productionOrchestrationStats) WriteUpdateTime(map[string]string) {}
func (stats *productionOrchestrationStats) WriteMetric(
	_ map[string]string,
	points, drops, errors float64,
	_ time.Duration,
) {
	stats.points += points
	stats.drops += drops
	stats.errors += errors
}

// TestJITProductionOrchestrationMatchesPipelineGo crosses the same global
// manager and runner registries used by RunPl. It is deliberately opt-in: the
// two environment variables must name a real Linux runtime and the production
// integration-script directory.
func TestJITProductionOrchestrationMatchesPipelineGo(t *testing.T) {
	runtimePath := os.Getenv("PLATYPUS_JIT_RUNTIME")
	scriptDirectory := os.Getenv("PIPELINE_SCRIPT_DIRECTORY")
	if runtimePath == "" || scriptDirectory == "" {
		t.Skip("set PLATYPUS_JIT_RUNTIME and PIPELINE_SCRIPT_DIRECTORY")
	}
	if _, err := os.Stat(runtimePath); err != nil {
		t.Fatalf("stat PLATYPUS_JIT_RUNTIME: %v", err)
	}
	if info, err := os.Stat(scriptDirectory); err != nil {
		t.Fatalf("stat PIPELINE_SCRIPT_DIRECTORY: %v", err)
	} else if !info.IsDir() {
		t.Fatalf("PIPELINE_SCRIPT_DIRECTORY is not a directory: %s", scriptDirectory)
	}

	isolateProductionOrchestrationMetrics(t)
	plstats.SetStats(nil)
	t.Cleanup(func() { plstats.SetStats(nil) })

	host := pljit.NewPipelineGoHostObserved(nil, observeJITHostCallback)
	runner, err := pljit.NewRunnerWithHost(
		runtimePath,
		"pipeline-go-1.4.3-datakit",
		16,
		host,
	)
	if err != nil {
		t.Fatalf("open real JIT runtime: %v", err)
	}
	installProductionOrchestrationRunner(t, runner)

	previousManager, hadPreviousManager := plval.GetManager()
	manager := plval.NewScriptManager(nil, nil)
	manager.SetJITInvalidate(runner.Invalidate)
	manager.SetJITCheck(func(source string) bool {
		check := runner.Check(source)
		return check.Route == pljit.RouteJITNative || check.Route == pljit.RouteJITWithHost
	})
	plval.SetManager(manager)
	t.Cleanup(func() {
		if hadPreviousManager {
			plval.SetManager(previousManager)
		} else {
			plval.SetManager(nil)
		}
	})
	t.Cleanup(func() {
		if err := manager.LoadScriptWithCatChecked(point.Logging, constants.NSDefault, nil, nil); err != nil {
			t.Errorf("clean production orchestration scripts: %v", err)
		}
	})

	tests := []struct {
		name                    string
		script                  string
		measurement             string
		primaryFields           map[string]any
		normalFields            map[string]any
		wantNativeOK            float64
		wantCommittedErrors     float64
		wantErrors              float64
		minimumDefaultTimeCalls float64
		wantDefaultTimeError    bool
	}{
		{
			name:        "nginx-native-success",
			script:      "nginx.p",
			measurement: "nginx",
			primaryFields: map[string]any{
				"message":  `127.0.0.1 - - [24/Mar/2021:13:54:19 +0800] "GET /basic_status HTTP/1.1" 200 97 "-" "Mozilla/5.0 (Macintosh; Intel Mac OS X 11_1_0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/89.0.4389.72 Safari/537.36"`,
				"sentinel": "keep",
				"status":   "unknown",
			},
			wantNativeOK: productionOrchestrationBatchSize,
		},
		{
			name:        "elasticsearch-committed-prefix-error",
			script:      "elasticsearch.p",
			measurement: "elasticsearch",
			primaryFields: map[string]any{
				"message":  `[2021-06-01T11:45:15,927][WARN ][o.e.c.r.a.DiskThresholdMonitor] [master] high disk watermark [90%] exceeded on [A2kEFgMLQ1-vhMdZMJV3Iw][master][/tmp/elasticsearch-cluster/nodes/0] free: 17.1gb[7.3%], shards will be relocated away from this node`,
				"sentinel": "keep",
				"status":   "unknown",
			},
			normalFields: map[string]any{
				"message":  `[2021-06-01T11:56:06,712][WARN ][i.s.s.query              ] [master] [shopping][0] took[36.3ms], took_millis[36], total_hits[5 hits], types[], stats[], search_type[QUERY_THEN_FETCH], total_shards[1]`,
				"sentinel": "keep",
				"status":   "unknown",
			},
			wantNativeOK:        productionOrchestrationBatchSize - 1,
			wantCommittedErrors: 1,
			wantErrors:          1,
		},
		{
			name:        "nginx-default-time-host-error",
			script:      "nginx.p",
			measurement: "nginx",
			primaryFields: map[string]any{
				"message":  "not a recognized log format",
				"time":     "not-a-time",
				"sentinel": "keep",
				"status":   "unknown",
			},
			normalFields: map[string]any{
				"message":  `127.0.0.1 - - [24/Mar/2021:13:54:19 +0800] "GET /basic_status HTTP/1.1" 200 97 "-" "Mozilla/5.0 (Macintosh; Intel Mac OS X 11_1_0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/89.0.4389.72 Safari/537.36"`,
				"sentinel": "keep",
				"status":   "unknown",
			},
			wantNativeOK:            productionOrchestrationBatchSize,
			minimumDefaultTimeCalls: 0,
			wantDefaultTimeError:    true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			sourceBytes, err := os.ReadFile(filepath.Join(scriptDirectory, test.script))
			if err != nil {
				t.Fatalf("read production script %s: %v", test.script, err)
			}
			source := string(sourceBytes)
			check := runner.Check(source)
			if check.Route != pljit.RouteJITNative {
				t.Fatalf("production route = %s (%s), want %s: %s",
					check.Route, check.Reason, pljit.RouteJITNative, check.Detail)
			}
			if containsString(check.Capabilities.HostCalls, "default_time") || check.Capabilities.RequiredHostFlags != 0 {
				t.Fatalf("production capabilities retain default_time host dependency: %#v", check.Capabilities)
			}
			if err := manager.LoadScriptWithCatChecked(point.Logging, constants.NSDefault,
				map[string]string{test.script: source}, nil); err != nil {
				t.Fatalf("publish %s through ScriptManager: %v", test.script, err)
			}
			manager.UpdateDefaultScript(map[point.Category]string{point.Logging: test.script})

			wantPrimary := runProductionOrchestrationBaseline(
				t, test.measurement, test.script, source, test.primaryFields)
			if test.wantDefaultTimeError {
				message, ok := wantPrimary.Get("pl_msg").(string)
				if !ok || !strings.HasPrefix(message, "time convert failed:") {
					t.Fatalf("pipeline-go baseline did not expose default_time error: %#v", wantPrimary.Get("pl_msg"))
				}
				if !wantPrimary.Time().Equal(time.Unix(1_700_000_000, 123)) {
					t.Fatalf("pipeline-go default_time fallback point time = %s", wantPrimary.Time())
				}
			}
			wantNormal := wantPrimary
			if test.normalFields != nil {
				wantNormal = runProductionOrchestrationBaseline(
					t, test.measurement, test.script, source, test.normalFields)
			}

			points := make([]*point.Point, productionOrchestrationBatchSize)
			for index := range points {
				fields := test.normalFields
				if index == 0 || fields == nil {
					fields = test.primaryFields
				}
				points[index] = newRealScriptPoint(test.measurement, fields)
			}
			attemptedCounter := jitRouteRecordsVec.WithLabelValues(
				point.Logging.String(), "attempted", "submitted")
			nativeOKCounter := jitRouteRecordsVec.WithLabelValues(
				point.Logging.String(), "native", "ok")
			committedCounter := jitRouteRecordsVec.WithLabelValues(
				point.Logging.String(), "native", "committed_error")
			beforeAttempted := readPrometheusCounter(t, attemptedCounter)
			beforeNativeOK := readPrometheusCounter(t, nativeOKCounter)
			beforeCommitted := readPrometheusCounter(t, committedCounter)
			beforeFallback := sumProductionOrchestrationCounterVec(t, jitFallbackVec)
			beforeDefaultTimeCalls := readPrometheusCounter(t, jitHostDefaultTimeCounter)

			stats := &productionOrchestrationStats{}
			plstats.SetStats(stats)
			result, err := RunPl(point.Logging, points, nil)
			plstats.SetStats(nil)
			if err != nil {
				t.Fatalf("RunPl production orchestration: %v", err)
			}
			if result == nil || len(result.Pts()) != productionOrchestrationBatchSize ||
				len(result.PtsOffload()) != 0 || len(result.PtsCreated()) != 0 {
				t.Fatalf("RunPl result = points:%d offload:%d created:%d",
					productionResultLength(result, "points"),
					productionResultLength(result, "offload"),
					productionResultLength(result, "created"))
			}
			for index, got := range result.Pts() {
				want := wantNormal
				if index == 0 {
					want = wantPrimary
				}
				if equal, reason := got.EqualWithReason(want); !equal {
					t.Fatalf("record %d differs from pipeline-go: %s\nJIT: %#v\npipeline-go: %#v",
						index, reason, got.KVMap(), want.KVMap())
				}
			}

			if delta := readPrometheusCounter(t, attemptedCounter) - beforeAttempted; delta != productionOrchestrationBatchSize {
				t.Fatalf("attempted records = %v, want %d", delta, productionOrchestrationBatchSize)
			}
			if delta := readPrometheusCounter(t, nativeOKCounter) - beforeNativeOK; delta != test.wantNativeOK {
				t.Fatalf("native ok records = %v, want %v", delta, test.wantNativeOK)
			}
			if delta := readPrometheusCounter(t, committedCounter) - beforeCommitted; delta != test.wantCommittedErrors {
				t.Fatalf("native committed-error records = %v, want %v",
					delta, test.wantCommittedErrors)
			}
			if delta := sumProductionOrchestrationCounterVec(t, jitFallbackVec) - beforeFallback; delta != 0 {
				t.Fatalf("runtime fallback records = %v, want 0", delta)
			}
			if delta := readPrometheusCounter(t, jitHostDefaultTimeCounter) - beforeDefaultTimeCalls; delta != test.minimumDefaultTimeCalls {
				t.Fatalf("default_time host calls = %v, want %v",
					delta, test.minimumDefaultTimeCalls)
			}
			if stats.points != productionOrchestrationBatchSize || stats.drops != 0 ||
				stats.errors != test.wantErrors {
				t.Fatalf("pipeline stats = points:%v drops:%v errors:%v, want %d/0/%v",
					stats.points, stats.drops, stats.errors,
					productionOrchestrationBatchSize, test.wantErrors)
			}
		})
	}
}

func runProductionOrchestrationBaseline(
	t *testing.T,
	measurement, scriptName, source string,
	fields map[string]any,
) *point.Point {
	t.Helper()
	script, err := NewPlScriptSimple(point.Logging, scriptName, source)
	if err != nil {
		t.Fatalf("compile pipeline-go baseline: %v", err)
	}
	defer script.Cleanup()
	run := pointRun{
		point:   newRealScriptPoint(measurement, fields),
		script:  script,
		started: time.Now(),
	}
	runPipelineGo(point.Logging, &run, nil)
	if run.dropped || run.output == nil {
		t.Fatalf("pipeline-go baseline unexpectedly dropped output: dropped=%v output=%#v",
			run.dropped, run.output)
	}
	return run.output
}

func installProductionOrchestrationRunner(t *testing.T, runner jitProcessor) {
	t.Helper()
	installed := &jitRunnerGeneration{runner: runner, drained: make(chan struct{})}
	jitRunners.mu.Lock()
	previous := jitRunners.current
	jitRunners.current = installed
	jitRunners.mu.Unlock()

	t.Cleanup(func() {
		jitRunners.mu.Lock()
		if jitRunners.current != installed {
			t.Errorf("production orchestration JIT generation changed during test")
		} else {
			jitRunners.current = previous
		}
		installed.retired = true
		if installed.refs == 0 {
			close(installed.drained)
		}
		jitRunners.mu.Unlock()
		if err := closeJITGeneration(installed); err != nil {
			t.Errorf("close production orchestration JIT runner: %v", err)
		}
	})
}

func isolateProductionOrchestrationMetrics(t *testing.T) {
	t.Helper()
	previousGroupMetrics := jitGroupMetricsByCategory
	jitGroupMetricsByCategory = new(sync.Map)
	previousPhase := jitPhaseCostVec
	previousBatch := jitBatchRecordsVec
	previousBytes := jitBytesVec
	previousFallback := jitFallbackVec
	previousRoute := jitRouteRecordsVec
	previousHost := jitHostCallbacksVec
	previousHostGeoIP := jitHostGeoIPCounter
	previousHostUserAgent := jitHostUserAgentCounter
	previousHostDefaultTime := jitHostDefaultTimeCounter
	previousHostUnknown := jitHostUnknownCounter

	jitPhaseCostVec = prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Name: "pipeline_jit_production_orchestration_phase_seconds",
		Help: "Isolated production orchestration test metric.",
	}, []string{"category", "phase"})
	jitBatchRecordsVec = prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Name: "pipeline_jit_production_orchestration_batch_records",
		Help: "Isolated production orchestration test metric.",
	}, []string{"category"})
	jitBytesVec = prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Name: "pipeline_jit_production_orchestration_bytes",
		Help: "Isolated production orchestration test metric.",
	}, []string{"category", "direction", "protocol"})
	jitFallbackVec = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "pipeline_jit_production_orchestration_fallback_records_total",
		Help: "Isolated production orchestration test metric.",
	}, []string{"category", "reason"})
	jitRouteRecordsVec = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "pipeline_jit_production_orchestration_route_records_total",
		Help: "Isolated production orchestration test metric.",
	}, []string{"category", "outcome", "reason"})
	jitHostCallbacksVec = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "pipeline_jit_production_orchestration_host_callbacks_total",
		Help: "Isolated production orchestration test metric.",
	}, []string{"operation"})
	jitHostGeoIPCounter = jitHostCallbacksVec.WithLabelValues("geoip")
	jitHostUserAgentCounter = jitHostCallbacksVec.WithLabelValues("user_agent")
	jitHostDefaultTimeCounter = jitHostCallbacksVec.WithLabelValues("default_time")
	jitHostUnknownCounter = jitHostCallbacksVec.WithLabelValues("unknown")

	t.Cleanup(func() {
		jitGroupMetricsByCategory = previousGroupMetrics
		jitPhaseCostVec = previousPhase
		jitBatchRecordsVec = previousBatch
		jitBytesVec = previousBytes
		jitFallbackVec = previousFallback
		jitRouteRecordsVec = previousRoute
		jitHostCallbacksVec = previousHost
		jitHostGeoIPCounter = previousHostGeoIP
		jitHostUserAgentCounter = previousHostUserAgent
		jitHostDefaultTimeCounter = previousHostDefaultTime
		jitHostUnknownCounter = previousHostUnknown
	})
}

func sumProductionOrchestrationCounterVec(t *testing.T, vector *prometheus.CounterVec) float64 {
	t.Helper()
	registry := prometheus.NewRegistry()
	if err := registry.Register(vector); err != nil {
		t.Fatalf("register isolated counter: %v", err)
	}
	families, err := registry.Gather()
	if err != nil {
		t.Fatalf("gather isolated counter: %v", err)
	}
	var sum float64
	for _, family := range families {
		for _, metric := range family.GetMetric() {
			sum += metric.GetCounter().GetValue()
		}
	}
	return sum
}

func containsString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func productionResultLength(result *ScriptResult, kind string) int {
	if result == nil {
		return 0
	}
	switch kind {
	case "points":
		return len(result.Pts())
	case "offload":
		return len(result.PtsOffload())
	case "created":
		return len(result.PtsCreated())
	default:
		return 0
	}
}
