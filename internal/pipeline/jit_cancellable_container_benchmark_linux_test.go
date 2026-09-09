// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build pipeline_jit && linux && jitbench && (amd64 || arm64)

// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// Copyright 2026-present Guance, Inc.

package pipeline

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/GuanceCloud/cliutils/point"
	"github.com/GuanceCloud/pipeline-go/constants"
	"github.com/GuanceCloud/pipeline-go/lang/platypus"
	pljit "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline/jit"
	plval "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline/plval"
)

// BenchmarkPipelineJITCancellableContainers exercises the context-aware runner
// entry for the same real list/map fixtures used by the crossover matrix.
func BenchmarkPipelineJITCancellableContainers(b *testing.B) {
	workloads := loadLogMatrixWorkloads(b)
	targets := map[string]bool{
		"log-collection-list-write-20": true,
		"log-collection-map-write-20":  true,
	}
	for _, workload := range workloads {
		if !targets[workload.name] {
			continue
		}
		workload := workload
		b.Run(workload.name, func(b *testing.B) {
			runner, err := pljit.NewRunnerWithHost(os.Getenv("PLATYPUS_JIT_RUNTIME"), "pipeline-go-1.4.3-datakit", 32, pljit.NewPipelineGoHost(nil))
			if err != nil {
				b.Fatal(err)
			}
			b.Cleanup(func() {
				if err := runner.Close(); err != nil {
					b.Error(err)
				}
			})
			bound, err := runner.BindSource(workload.source)
			if err != nil {
				b.Fatal(err)
			}
			b.Cleanup(func() {
				if err := bound.Close(); err != nil {
					b.Error(err)
				}
			})
			projection, err := bound.Projection(workload.source)
			if err != nil {
				b.Fatal(err)
			}
			script, err := NewPlScriptSimple(point.Logging, workload.name+".p", workload.source)
			if err != nil {
				b.Fatal(err)
			}
			for _, batchSize := range []int{1, 10} {
				batchSize := batchSize
				b.Run(fmt.Sprintf("batch-%03d", batchSize), func(b *testing.B) {
					b.Run("pipeline-go-e2e", func(b *testing.B) {
						b.ReportAllocs()
						cpuStarted := benchmarkProcessCPU(b)
						b.ResetTimer()
						for range b.N {
							for _, pt := range newCancellableContainerPoints(workload, batchSize) {
								run := pointRun{point: pt, script: script}
								runPipelineGo(point.Logging, &run, nil)
								if run.output == nil || run.dropped {
									b.Fatal("pipeline-go unexpectedly dropped the benchmark point")
								}
							}
						}
						b.ReportMetric(float64(benchmarkProcessCPU(b)-cpuStarted)/float64(b.N), "cpu-ns/op")
						b.ReportMetric(recordsPerSecond(b, batchSize), "records/s")
					})
					b.Run("jit-group-cancellable", func(b *testing.B) {
						ctx, cancel := context.WithCancel(context.Background())
						defer cancel()
						if err := checkCancellableContainerAgreement(bound, workload, projection, script, ctx, batchSize); err != nil {
							b.Fatal(err)
						}
						processor := contextJITProcessor{ctx: ctx, processor: runner}
						runGroup := func() {
							points := newCancellableContainerPoints(workload, batchSize)
							indexes, runs := make([]int, batchSize), make([]pointRun, batchSize)
							for i, pt := range points {
								indexes[i], runs[i] = i, pointRun{point: pt, script: script}
							}
							runJITGroup(processor, point.Logging, script, indexes, runs, nil)
							for _, run := range runs {
								if run.output == nil || run.dropped || run.output.Get("ok") != true {
									b.Fatal("cancellable group failed")
								}
							}
						}
						runGroup()
						b.ReportAllocs()
						cpuStarted := benchmarkProcessCPU(b)
						b.ResetTimer()
						for range b.N {
							runGroup()
						}
						b.ReportMetric(float64(benchmarkProcessCPU(b)-cpuStarted)/float64(b.N), "cpu-ns/op")
						b.ReportMetric(recordsPerSecond(b, batchSize), "records/s")
					})
					b.Run("jit-e2e-cancellable", func(b *testing.B) {
						ctx, cancel := context.WithCancel(context.Background())
						defer cancel()
						if err := checkCancellableContainerAgreement(bound, workload, projection, script, ctx, batchSize); err != nil {
							b.Fatal(err)
						}
						b.ReportAllocs()
						cpuStarted := benchmarkProcessCPU(b)
						b.ResetTimer()
						for range b.N {
							points := newCancellableContainerPoints(workload, batchSize)
							input, err := encodeProjectedJITPoints(point.Logging, points, projection)
							if err != nil {
								b.Fatal(err)
							}
							batch, err := bound.ProcessContext(ctx, workload.source, input)
							if err != nil {
								b.Fatal(err)
							}
							if err := applyCancellableContainerBatch(points, batch); err != nil {
								b.Fatal(err)
							}
						}
						b.ReportMetric(float64(benchmarkProcessCPU(b)-cpuStarted)/float64(b.N), "cpu-ns/op")
						b.ReportMetric(recordsPerSecond(b, batchSize), "records/s")
					})
				})
			}
		})
	}
}

func TestRunPlContextContainerCorrectness(t *testing.T) {
	path := os.Getenv("PLATYPUS_JIT_RUNTIME")
	if path == "" {
		t.Skip("requires real PLATYPUS_JIT_RUNTIME")
	}
	isolateProductionOrchestrationMetrics(t)
	runner, err := pljit.NewRunner(path, "pipeline-go-1.4.3-datakit", 4)
	if err != nil {
		t.Fatal(err)
	}
	installProductionOrchestrationRunner(t, runner)
	previous, _ := plval.GetManager()
	manager := plval.NewScriptManager(nil, nil)
	plval.SetManager(manager)
	t.Cleanup(func() {
		plval.SetManager(previous)
		if err := runner.Close(); err != nil {
			t.Error(err)
		}
	})
	manager.SetJITCheck(func(source string) bool {
		return runner.Check(source).Route == pljit.RouteJITNative
	})
	scripts := map[string]string{
		"list-write.p": "x=[seed,2]; x[0]+=seed; add_key(ok,x[0]==seed*2)",
		"map-write.p":  "x={\"state\":seed}; x[\"state\"]+=seed; add_key(ok,x[\"state\"]==seed*2)",
	}
	if err := manager.LoadScriptWithCatChecked(point.Logging, constants.NSDefault, scripts, nil); err != nil {
		t.Fatal(err)
	}
	for scriptName := range scripts {
		scriptName := scriptName
		name := strings.TrimSuffix(scriptName, ".p")
		t.Run(name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			attempted := jitRouteRecordsVec.WithLabelValues(point.Logging.String(), "attempted", "submitted")
			before := readPrometheusCounter(t, attempted)
			result, err := RunPlContext(ctx, point.Logging, []*point.Point{newRealScriptPoint(name, map[string]any{"message": "keep", "seed": int64(12)})}, nil)
			if err != nil || result == nil || len(result.Pts()) != 1 {
				t.Fatalf("RunPlContext failed: %v", err)
			}
			if got := readPrometheusCounter(t, attempted) - before; got != 1 {
				t.Fatalf("expected one actual native record, got %v", got)
			}
			if result.Pts()[0].Get("ok") != true {
				t.Fatalf("unexpected container result: %#v", result.Pts()[0].Get("ok"))
			}
		})
	}
}

func applyCancellableContainerBatch(points []*point.Point, batch pljit.Batch) error {
	if batch.Static != nil {
		return fmt.Errorf("cancellable runner unexpectedly returned static output")
	}
	if len(batch.Records) != len(points) {
		return fmt.Errorf("cancellable runner returned %d records for %d points", len(batch.Records), len(points))
	}
	for i, record := range batch.Records {
		if record.Status != pljit.TerminalOK {
			return fmt.Errorf("cancellable record %d terminal=%d error=%s", i, record.Status, record.Error)
		}
		if len(record.Error) != 0 {
			return fmt.Errorf("cancellable record %d has error=%s", i, record.Error)
		}
		if err := applyBenchmarkJITRecord(point.Logging, points[i], batch, i); err != nil {
			return err
		}
	}
	return nil
}

func checkCancellableContainerAgreement(bound *pljit.BoundModules, workload realScriptBenchmarkWorkload, projection pljit.InputProjection, script *platypus.PlScript, ctx context.Context, batchSize int) error {
	points := newCancellableContainerPoints(workload, batchSize)
	references := newCancellableContainerPoints(workload, batchSize)
	for _, pt := range references {
		run := pointRun{point: pt, script: script}
		runPipelineGo(point.Logging, &run, nil)
		if run.output == nil || run.dropped {
			return fmt.Errorf("pipeline-go unexpectedly dropped the benchmark point")
		}
	}
	input, err := encodeProjectedJITPoints(point.Logging, points, projection)
	if err != nil {
		return err
	}
	batch, err := bound.ProcessContext(ctx, workload.source, input)
	if err != nil {
		return err
	}
	if err := applyCancellableContainerBatch(points, batch); err != nil {
		return err
	}
	for i := range points {
		if equal, reason := points[i].EqualWithReason(references[i], point.EqualWithoutKeys(plFieldCost)); !equal {
			return fmt.Errorf("cancellable/Go point %d mismatch: %s", i, reason)
		}
	}
	return nil
}

func newCancellableContainerPoints(workload realScriptBenchmarkWorkload, batchSize int) []*point.Point {
	points := make([]*point.Point, batchSize)
	for i := range points {
		fields := map[string]any{"message": workload.message, "sentinel": "keep", "status": "unknown"}
		for key, value := range workload.extraFields {
			fields[key] = value
		}
		points[i] = newRealScriptPoint(workload.name, fields)
	}
	return points
}
