// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build pipeline_jit && linux && jitbench && (amd64 || arm64)

// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2026-present Guance, Inc.

package pipeline

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	pljit "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline/jit"

	"github.com/GuanceCloud/cliutils/logger"
	"github.com/GuanceCloud/cliutils/point"
	"github.com/GuanceCloud/pipeline-go/constants"
	"github.com/GuanceCloud/pipeline-go/ptinput/funcs"
	dto "github.com/prometheus/client_model/go"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline/plval"
)

// Frozen baseline binaries retain the earlier benchmark source. The clock
// mode is an opt-in addition and never changes ordinary/cancellable fixtures.
// E2E includes Point construction, RunPlContext routing, native and writeback.
func BenchmarkPipelineJITRolloutE2E(b *testing.B) {
	path := os.Getenv("PLATYPUS_JIT_RUNTIME")
	if path == "" {
		b.Skip("set PLATYPUS_JIT_RUNTIME")
	}
	if err := logger.InitRoot(&logger.Option{Level: logger.ERROR, Flags: logger.OPT_DEFAULT}); err != nil {
		b.Fatal(err)
	}
	funcs.InitLog()
	targets := map[string]bool{"log-json-1-small": true, "log-json-10-small": true, "log-gjson-1-small": true, "log-load-json-once-10-small": true,
		"log-grok-3-line": true, "log-exec-loop-100": true, "log-exec-assign-20": true, "log-collection-map-write-20": true, "log-collection-list-write-20": true, "log-string-plus-20": true}
	for _, workload := range loadLogMatrixWorkloads(b) {
		if !targets[workload.name] && os.Getenv("JIT_BENCH_ALL") != "1" && !(os.Getenv("JIT_BENCH_JSON_FOCUS") == "1" && (strings.HasPrefix(workload.name, "log-load-json-") || strings.HasPrefix(workload.name, "log-json-"))) {
			continue
		}
		b.Run(workload.name, func(b *testing.B) {
			for _, size := range []int{1, 10} {
				b.Run(fmt.Sprintf("batch-%03d", size), func(b *testing.B) {
					for _, engine := range []string{"go", "jit"} {
						b.Run(engine, func(b *testing.B) {
							original, _ := plval.GetManager()
							m := plval.NewScriptManager(nil, nil)
							plval.SetManager(m)
							defer plval.SetManager(original)
							if err := m.LoadScriptWithCatChecked(point.Logging, constants.NSDefault, map[string]string{workload.script: workload.source}, nil); err != nil {
								b.Fatal(err)
							}
							m.UpdateDefaultScript(map[point.Category]string{point.Logging: workload.script})
							if err := initJIT(&plval.PipelineCfg{JIT: &plval.JITCfg{Enabled: false}}, ""); err != nil {
								b.Fatal(err)
							}
							ctx := context.Background()
							if os.Getenv("PIPELINE_ROLLOUT_CANCELLABLE") == "1" {
								var cancel context.CancelFunc
								ctx, cancel = context.WithCancel(ctx)
								defer cancel()
							}
							if os.Getenv("PIPELINE_ROLLOUT_CLOCK") == "1" {
								ctx = pljit.WithRecordTimeout(ctx, 30*time.Second)
							}
							reference := newCancellableContainerPoints(workload, size)
							if _, err := RunPlContext(ctx, point.Logging, reference, nil); err != nil {
								b.Fatal(err)
							}
							if engine == "jit" {
								if err := initJIT(&plval.PipelineCfg{JIT: &plval.JITCfg{Enabled: true, RuntimePath: path, MaxCachedPrograms: 32}}, ""); err != nil {
									b.Fatal(err)
								}
								defer func() {
									if err := closeJITGeneration(jitRunners.replace(nil)); err != nil {
										b.Error(err)
									}
								}()
							}
							counter := jitRouteRecordsVec.WithLabelValues("logging", "attempted", "submitted")
							var before, after dto.Metric
							if err := counter.Write(&before); err != nil {
								b.Fatal(err)
							}
							points := newCancellableContainerPoints(workload, size)
							if _, err := RunPlContext(ctx, point.Logging, points, nil); err != nil {
								b.Fatal(err)
							}
							if err := counter.Write(&after); err != nil {
								b.Fatal(err)
							}
							if engine == "jit" && after.GetCounter().GetValue()-before.GetCounter().GetValue() != float64(size) {
								b.Fatal("benchmark did not submit native records")
							}
							for i := range points {
								if os.Getenv("JIT_BENCH_JSON_FOCUS") == "1" {
									for key, want := range workload.expectedFields {
										if got := points[i].Get(key); !reflect.DeepEqual(got, want) {
											b.Fatalf("field %s = %#v, want %#v", key, got, want)
										}
									}
								}
								if equal, why := preflightEqualPoints(points[i], reference[i]); !equal {
									b.Fatal("full Point mismatch:", why)
								}
							}
							b.ReportAllocs()
							b.ResetTimer()
							cpuStarted := benchmarkProcessCPU(b)
							for range b.N {
								points := newCancellableContainerPoints(workload, size)
								result, err := RunPlContext(ctx, point.Logging, points, nil)
								if err != nil || len(result.Pts()) != size {
									b.Fatal("pipeline execution failed", err)
								}
								result.Release()
							}
							b.StopTimer()
							b.ReportMetric(float64(benchmarkProcessCPU(b)-cpuStarted)/float64(b.N*size), "cpu-ns/record")
							b.ReportMetric(float64(b.Elapsed().Nanoseconds())/float64(b.N*size), "ns/record")
						})
					}
				})
			}
		})
	}
}
