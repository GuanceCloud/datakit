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
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"sync/atomic"
	"syscall"
	"testing"
	"time"

	"github.com/GuanceCloud/cliutils/logger"
	"github.com/GuanceCloud/cliutils/point"
	"github.com/GuanceCloud/pipeline-go/lang"
	"github.com/GuanceCloud/pipeline-go/ptinput"
	"github.com/GuanceCloud/pipeline-go/ptinput/funcs"
	plruntime "github.com/GuanceCloud/platypus/pkg/engine/runtime"
	pljit "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline/jit"
)

var (
	realScriptBenchmarkBytesSink   []byte
	realScriptBenchmarkBatchSink   pljit.Batch
	realScriptBenchmarkNativeSink  pljit.BenchmarkEncodedBatch
	realScriptBenchmarkBooleanSink bool
)

type realScriptBenchmarkWorkload struct {
	name           string
	script         string
	source         string
	message        string
	invalidTime    bool
	requireFast    bool
	extraFields    map[string]any
	expectedFields map[string]any
}

func BenchmarkPipelineJITRealScriptsCrossover(b *testing.B) {
	if err := logger.InitRoot(&logger.Option{Level: logger.WARN, Flags: logger.OPT_DEFAULT}); err != nil {
		b.Fatalf("reduce benchmark log noise: %v", err)
	}
	// Function loggers are cached at package initialization. Rebind them after
	// setting the root level so parse_duration does not log every timed record.
	funcs.InitLog()
	runtimePath := os.Getenv("PLATYPUS_JIT_RUNTIME")
	if runtimePath == "" {
		b.Skip("set PLATYPUS_JIT_RUNTIME")
	}
	scriptDirectory := os.Getenv("PIPELINE_SCRIPT_DIRECTORY")
	if scriptDirectory == "" {
		scriptDirectory = filepath.Join("testdata", "jit-real-scripts")
	}

	var callbacks atomic.Uint64
	var timestampCallbacks atomic.Uint64
	host := pljit.NewPipelineGoHost(nil)
	pljit.BenchmarkSetHostObserver(host, func(operation string, _ uintptr) {
		callbacks.Add(1)
		if operation == "default_time" {
			timestampCallbacks.Add(1)
		}
	})
	runner, err := pljit.NewRunnerWithHost(
		runtimePath,
		"pipeline-go-1.4.3-datakit",
		32,
		host,
	)
	if err != nil {
		b.Fatalf("open JIT runner: %v", err)
	}
	defer runner.Close()
	var groupProcessor jitBatchProcessor = runner
	if os.Getenv("JIT_BENCH_CANCELLABLE") == "1" {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		groupProcessor = contextJITProcessor{ctx: ctx, processor: runner}
	}

	workloads := []realScriptBenchmarkWorkload{
		{
			name:        "light-drop-cast",
			script:      "light-drop-cast.p",
			source:      "drop_key(remove_me)\ncast(n, \"int\")\n",
			message:     "light static ABI control",
			requireFast: true,
			extraFields: map[string]any{"n": "123", "remove_me": "gone"},
		},
		{
			name:   "json-3-cast",
			script: "json-3-cast.p",
			source: "json(message, payload.user.id, user_id)\n" +
				"json(message, payload.request.path, request_path)\n" +
				"json(message, payload.ok, ok)\n" +
				"cast(user_id, \"int\")\n",
			message:     `{"payload":{"user":{"id":"42"},"request":{"path":"/health"},"ok":true}}`,
			requireFast: true,
		},
		{
			name:        "single-grok",
			script:      "single-grok.p",
			source:      "grok(_, \"%{IPORHOST:client_ip} %{WORD:http_method} %{URIPATHPARAM:http_url} %{INT:status_code:int} %{INT:bytes:int}\")\n",
			message:     "127.0.0.1 GET /health?q=1 200 1234",
			requireFast: true,
		},
		{
			name:        "regex-chain-static",
			script:      "regex-chain-static.p",
			source:      strings.Repeat("replace(n, \"[a-z]+\", \"x\")\n", 16),
			message:     "static replace chain",
			requireFast: true,
			extraFields: map[string]any{"n": "abc123def456"},
		},
		{
			name:    "nginx-native",
			script:  "nginx.p",
			message: `127.0.0.1 - - [24/Mar/2021:13:54:19 +0800] "GET /basic_status HTTP/1.1" 200 97 "-" "Mozilla/5.0 (Macintosh; Intel Mac OS X 11_1_0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/89.0.4389.72 Safari/537.36"`,
		},
		{
			name:    "nginx-stage-access-base",
			script:  "nginx-stage-access-base.p",
			source:  `grok(_, "%{NOTSPACE:client_ip} %{NOTSPACE:http_ident} %{NOTSPACE:http_auth} \\[%{HTTPDATE:time}\\] \"%{DATA:http_method} %{GREEDYDATA:http_url} HTTP/%{NUMBER:http_version}\" %{INT:status_code} %{INT:bytes}")`,
			message: `127.0.0.1 - - [24/Mar/2021:13:54:19 +0800] "GET /basic_status HTTP/1.1" 200 97 "-" "Mozilla/5.0"`,
		},
		{
			name:   "nginx-stage-access-agent-grok",
			script: "nginx-stage-access-agent-grok.p",
			source: "add_pattern(\"access_common\", \"%{NOTSPACE:client_ip} %{NOTSPACE:http_ident} %{NOTSPACE:http_auth} \\\\[%{HTTPDATE:time}\\\\] \\\"%{DATA:http_method} %{GREEDYDATA:http_url} HTTP/%{NUMBER:http_version}\\\" %{INT:status_code} %{INT:bytes}\")\n" +
				`grok(_, '%{access_common} "%{NOTSPACE:referrer}" "%{GREEDYDATA:agent}"')` + "\n",
			message: `127.0.0.1 - - [24/Mar/2021:13:54:19 +0800] "GET /basic_status HTTP/1.1" 200 97 "-" "Mozilla/5.0 (Macintosh; Intel Mac OS X 11_1_0) AppleWebKit/537.36 Chrome/89.0.4389.72 Safari/537.36"`,
		},
		{
			name:   "nginx-stage-access-agent",
			script: "nginx-stage-access-agent.p",
			source: "add_pattern(\"access_common\", \"%{NOTSPACE:client_ip} %{NOTSPACE:http_ident} %{NOTSPACE:http_auth} \\\\[%{HTTPDATE:time}\\\\] \\\"%{DATA:http_method} %{GREEDYDATA:http_url} HTTP/%{NUMBER:http_version}\\\" %{INT:status_code} %{INT:bytes}\")\n" +
				`grok(_, '%{access_common} "%{NOTSPACE:referrer}" "%{GREEDYDATA:agent}"')` + "\nuser_agent(agent)\n",
			message: `127.0.0.1 - - [24/Mar/2021:13:54:19 +0800] "GET /basic_status HTTP/1.1" 200 97 "-" "Mozilla/5.0 (Macintosh; Intel Mac OS X 11_1_0) AppleWebKit/537.36 Chrome/89.0.4389.72 Safari/537.36"`,
		},
		{
			name:        "nginx-stage-user-agent",
			script:      "nginx-stage-user-agent.p",
			source:      "user_agent(agent)\n",
			message:     "user-agent stage",
			extraFields: map[string]any{"agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 11_1_0) AppleWebKit/537.36 Chrome/89.0.4389.72 Safari/537.36"},
		},
		{
			name:   "nginx-stage-error-misses",
			script: "nginx-stage-error-misses.p",
			source: "add_pattern(\"date2\", \"%{YEAR}[./]%{MONTHNUM}[./]%{MONTHDAY} %{TIME}\")\n" +
				`grok(_, "%{date2:time} \\[%{LOGLEVEL:status}\\] %{GREEDYDATA:msg}, client: %{NOTSPACE:client_ip}, server: %{NOTSPACE:server}, request: \"%{DATA:http_method} %{GREEDYDATA:http_url} HTTP/%{NUMBER:http_version}\", (upstream: \"%{GREEDYDATA:upstream}\", )?host: \"%{NOTSPACE:ip_or_host}\"")` + "\n" +
				`grok(_, "%{date2:time} \\[%{LOGLEVEL:status}\\] %{GREEDYDATA:msg}, client: %{NOTSPACE:client_ip}, server: %{NOTSPACE:server}, request: \"%{GREEDYDATA:http_method} %{GREEDYDATA:http_url} HTTP/%{NUMBER:http_version}\", host: \"%{NOTSPACE:ip_or_host}\"")` + "\n" +
				`grok(_,"%{date2:time} \\[%{LOGLEVEL:status}\\] %{GREEDYDATA:msg}")` + "\n",
			message: `127.0.0.1 - - [24/Mar/2021:13:54:19 +0800] "GET /basic_status HTTP/1.1" 200 97 "-" "Mozilla/5.0"`,
		},
		{
			name:        "default-time-native",
			script:      "default-time-native.p",
			source:      "default_time(time, \"UTC\")\n",
			message:     "native timestamp control",
			requireFast: true,
			extraFields: map[string]any{"time": "2026-05-19T13:47:01.004Z"},
		},
		{
			name:        "default-time-host",
			script:      "default-time-host.p",
			source:      "default_time(time)\n",
			message:     "host callback control",
			invalidTime: true,
		},
		{
			name:        "elasticsearch-native",
			script:      "elasticsearch.p",
			message:     `[2021-06-01T11:56:06,712][WARN ][i.s.s.query              ] [master] [shopping][0] took[36.3ms], took_millis[36], total_hits[5 hits], types[], stats[], search_type[QUERY_THEN_FETCH], total_shards[1]`,
			extraFields: map[string]any{"shard": "0"},
		},
		{
			name:    "tdengine-duration",
			script:  "tdengine.p",
			message: `08/22 13:44:34.290731 01081508 TAOS_ADAPTER info "| 204 |    1.641678ms |     172.16.5.29 | POST | /influxdb/v1/write?db=biz " model=web sessionID=48847`,
		},
		{
			name:    "tdengine-error",
			script:  "tdengine.p",
			message: `08/22 13:44:34.290731 01081508 TAOS_ADAPTER error "" error_code=512 error_msg=db is not specified`,
		},
		{
			name:    "consul-native",
			script:  "consul.p",
			message: `Sep 18 19:30:23 derrick-ThinkPad-X230 consul[11803]: 2021-09-18T19:30:23.522+0800 [INFO]  agent.server.connect: initialized primary datacenter CA with provider: provider=consul`,
		},
	}

	if os.Getenv("JIT_BENCH_LOG_MATRIX") == "1" {
		workloads = loadLogMatrixWorkloads(b)
	}

	for _, workload := range workloads {
		workload := workload
		b.Run(workload.name, func(b *testing.B) {
			source := workload.source
			if source == "" {
				sourceBytes, err := os.ReadFile(filepath.Join(scriptDirectory, workload.script))
				if err != nil {
					b.Fatal(err)
				}
				source = string(sourceBytes)
			}
			script, err := NewPlScriptSimple(point.Logging, workload.script, source)
			if err != nil {
				b.Fatalf("create pipeline-go script: %v", err)
			}
			projection, err := runner.Projection(source)
			if err != nil {
				b.Fatalf("query JIT projection: %v", err)
			}
			check := runner.Check(source)
			if workload.requireFast &&
				(check.Capabilities.InputProjection.Mode != "keys" ||
					check.Capabilities.StaticOutput.Mode != "slots") {
				b.Fatalf(
					"benchmark requires datakit fast path, got projection=%s static=%s route=%s",
					check.Capabilities.InputProjection.Mode,
					check.Capabilities.StaticOutput.Mode,
					check.Route,
				)
			}
			b.Logf(
				"route=%s backend=%s tier=%s host=%v projection=%s static=%s keys=%d",
				check.Route,
				check.Capabilities.Backend,
				check.Capabilities.ExecutionTier,
				check.Capabilities.HostCalls,
				check.Capabilities.InputProjection.Mode,
				check.Capabilities.StaticOutput.Mode,
				len(projection.Keys()),
			)

			newPoints := func(batchSize int) []*point.Point {
				points := make([]*point.Point, batchSize)
				for index := range points {
					fields := map[string]any{
						"message":  workload.message,
						"sentinel": "keep",
						"status":   "unknown",
					}
					if workload.invalidTime {
						fields["time"] = "11-Oct-2012::12:53:54"
					}
					for key, value := range workload.extraFields {
						fields[key] = value
					}
					points[index] = newRealScriptPoint(
						strings.TrimSuffix(workload.script, ".p"),
						fields,
					)
				}
				return points
			}

			for _, batchSize := range []int{1, 4, 8, 10, 16, 32, 64, 128, 256, 1024} {
				batchSize := batchSize
				b.Run(fmt.Sprintf("batch-%03d", batchSize), func(b *testing.B) {
					phasePoints := newPoints(batchSize)
					input, err := encodeProjectedJITPoints(point.Logging, phasePoints, projection)
					if err != nil {
						b.Fatalf("encode warmup: %v", err)
					}
					raw, err := runner.BenchmarkNative(source, input)
					if err != nil {
						b.Fatalf("native warmup: %v", err)
					}
					decoded, err := raw.Decode()
					if err != nil {
						b.Fatalf("decode warmup: %v", err)
					}
					if err := applyRealScriptBenchmarkBatch(point.Logging, phasePoints, decoded); err != nil {
						b.Fatalf("apply warmup: %v", err)
					}

					b.Run("pipeline-go-e2e", func(b *testing.B) {
						b.ReportAllocs()
						b.ResetTimer()
						cpuStarted := benchmarkProcessCPU(b)
						for range b.N {
							for _, pt := range newPoints(batchSize) {
								run := pointRun{point: pt, script: script, started: time.Now()}
								runPipelineGo(point.Logging, &run, nil)
								if run.output == nil || run.dropped {
									b.Fatal("pipeline-go unexpectedly dropped the benchmark point")
								}
							}
						}
						b.ReportMetric(float64(benchmarkProcessCPU(b)-cpuStarted)/float64(b.N), "cpu-ns/op")
						b.ReportMetric(recordsPerSecond(b, batchSize), "records/s")
					})

					b.Run("pipeline-go-runtime", func(b *testing.B) {
						// Only the checked local log fixtures use this narrower entry;
						// host services and wrapper status processing are not installed.
						if os.Getenv("JIT_BENCH_LOG_MATRIX") != "1" {
							b.Skip("runtime comparison requires checked log fixtures")
						}
						engine := script.Engine()
						signal := pipelineContextSignal{context.Background()}
						actual, expected := newPoints(batchSize), newPoints(batchSize)
						for i, pt := range actual {
							wrapped := ptinput.PtWrap(point.Logging, pt)
							ctx := plruntime.InitCtx(plruntime.GetContext(), wrapped, engine, signal)
							err := plruntime.RunStmts(ctx, engine.Ast)
							// Match Script.Run's status and timestamp finalization for
							// complete output parity, outside the RunStmts timing.
							if wrapped.GetStatusMapping() {
								lang.ProcLoggingStatus(wrapped, false, nil)
							}
							wrapped.KeyTime2Time()
							pt = wrapped.Point()
							plruntime.PutContext(ctx)
							if err != nil {
								b.Fatal(err)
							}
							ref := pointRun{point: expected[i], script: script}
							runPipelineGo(point.Logging, &ref, nil)
							if ref.output == nil || ref.dropped || len(ref.created) != 0 {
								b.Fatal("unexpected reference output")
							}
							if equal, reason := pt.EqualWithReason(ref.output, point.EqualWithoutKeys(plFieldCost)); !equal {
								b.Fatal(reason)
							}
							for key, want := range workload.expectedFields {
								if got := pt.Get(key); !reflect.DeepEqual(got, want) {
									b.Fatalf("runtime field %s = %#v, want %#v", key, got, want)
								}
							}
						}
						// Prepare fresh inputs and contexts outside timing in bounded
						// chunks. Never repeatedly execute a previously mutated Point.
						const chunkBatches = 32
						b.ResetTimer()
						b.StopTimer()
						for done := 0; done < b.N; {
							n := min(chunkBatches, b.N-done)
							contexts := make([]*plruntime.Task, 0, n*batchSize)
							for range n {
								for _, pt := range newPoints(batchSize) {
									contexts = append(contexts, plruntime.InitCtx(plruntime.GetContext(), ptinput.PtWrap(point.Logging, pt), engine, signal))
								}
							}
							b.StartTimer()
							for _, ctx := range contexts {
								if err := plruntime.RunStmts(ctx, engine.Ast); err != nil {
									b.Fatal(err)
								}
							}
							b.StopTimer()
							for _, ctx := range contexts {
								plruntime.PutContext(ctx)
							}
							done += n
						}
						b.ReportMetric(recordsPerSecond(b, batchSize), "records/s")
					})

					b.Run("jit-group-e2e", func(b *testing.B) {
						if os.Getenv("JIT_BENCH_PHASE_PROFILE") == "1" {
							// Rust diagnostic counters are thread-local. Keep a complete
							// batch window on one native thread; ordinary timing is unchanged.
							runtime.LockOSThread()
							defer runtime.UnlockOSThread()
						}

						// Verify the production entry before timing. A prefix-error
						// result can still contain a Point, so checking non-nil alone
						// would accidentally reward incomplete execution.
						inputs, expected := newPoints(batchSize), newPoints(batchSize)
						indexes := make([]int, batchSize)
						runs := make([]pointRun, batchSize)
						for index, pt := range inputs {
							indexes[index] = index
							runs[index] = pointRun{point: pt, script: script}
							reference := pointRun{point: expected[index], script: script}
							runPipelineGo(point.Logging, &reference, nil)
							if reference.output == nil || reference.dropped {
								b.Fatal("reference unexpectedly dropped the benchmark point")
							}
							for key, want := range workload.expectedFields {
								if got := reference.output.Get(key); !reflect.DeepEqual(got, want) {
									b.Fatalf("Go reference field %s = %#v (%T), want %#v (%T)", key, got, got, want, want)
								}
							}
							expected[index] = reference.output
						}
						runJITGroup(groupProcessor, point.Logging, script, indexes, runs, nil)
						for index, run := range runs {
							if run.output == nil || run.dropped || len(run.created) != 0 {
								b.Fatal("unexpected JIT group output")
							}
							if equal, reason := run.output.EqualWithReason(expected[index], point.EqualWithoutKeys(plFieldCost)); !equal {
								b.Fatalf("group differs from pipeline-go: %s", reason)
							}
						}
						callbacks.Store(0)
						timestampCallbacks.Store(0)
						b.ReportAllocs()
						b.ResetTimer()
						cpuStarted := benchmarkProcessCPU(b)
						for range b.N {
							points := newPoints(batchSize)
							indexes := make([]int, batchSize)
							runs := make([]pointRun, batchSize)
							for index, pt := range points {
								indexes[index] = index
								runs[index] = pointRun{point: pt, script: script}
							}
							runJITGroup(groupProcessor, point.Logging, script, indexes, runs, nil)
							for _, run := range runs {
								if run.output == nil || run.dropped {
									b.Fatal("JIT group unexpectedly failed or dropped a point")
								}
							}
						}
						b.ReportMetric(float64(benchmarkProcessCPU(b)-cpuStarted)/float64(b.N), "cpu-ns/op")
						b.ReportMetric(recordsPerSecond(b, batchSize), "records/s")
						reportCallbackMetrics(b, batchSize, callbacks.Load(), timestampCallbacks.Load())
					})

					b.Run("jit-e2e", func(b *testing.B) {
						callbacks.Store(0)
						timestampCallbacks.Store(0)
						var encoder projectedJITEncoder
						b.ReportAllocs()
						b.ResetTimer()
						for range b.N {
							points := newPoints(batchSize)
							input, err := encoder.encode(point.Logging, points, projection)
							if err != nil {
								b.Fatal(err)
							}
							batch, err := runner.Process(source, input)
							if err != nil {
								b.Fatal(err)
							}
							if err := applyRealScriptBenchmarkBatch(point.Logging, points, batch); err != nil {
								b.Fatal(err)
							}
						}
						b.ReportMetric(recordsPerSecond(b, batchSize), "records/s")
						reportCallbackMetrics(b, batchSize, callbacks.Load(), timestampCallbacks.Load())
					})

					b.Run("encode", func(b *testing.B) {
						points := newPoints(batchSize)
						b.ReportAllocs()
						b.SetBytes(int64(len(input)))
						b.ResetTimer()
						for range b.N {
							realScriptBenchmarkBytesSink, err = encodeProjectedJITPoints(point.Logging, points, projection)
							if err != nil {
								b.Fatal(err)
							}
						}
						b.ReportMetric(recordsPerSecond(b, batchSize), "records/s")
						b.ReportMetric(float64(len(input))/float64(batchSize), "input-bytes/record")
					})

					b.Run("encode-prepared", func(b *testing.B) {
						if projection.IsAll() && !projection.AllowsRawStringValues() && !projection.AllowsRawTagValues() {
							b.Skip("production uses the general encoder for this projection")
						}
						points := newPoints(batchSize)
						indexes, runs := make([]int, batchSize), make([]pointRun, batchSize)
						for i, pt := range points {
							indexes[i], runs[i].point = i, pt
						}
						var encoder projectedJITEncoder
						encode := func() ([]byte, error) {
							end := encoder.prepareChunk(point.Logging, indexes, runs, projection, 0)
							if end != batchSize {
								b.Skip("conversion fixture requires multiple production chunks")
							}
							return encoder.encodePrepared(end, projection)
						}
						for range 2 {
							got, err := encode()
							if err != nil || !bytes.Equal(got, input) {
								b.Fatalf("prepared encoding differs from reference: %v", err)
							}
						}
						b.ReportAllocs()
						b.SetBytes(int64(len(input)))
						b.ResetTimer()
						for range b.N {
							realScriptBenchmarkBytesSink, err = encode()
							if err != nil {
								b.Fatal(err)
							}
						}
						b.ReportMetric(float64(len(input))/float64(batchSize), "input-bytes/record")
					})

					b.Run("native", func(b *testing.B) {
						callbacks.Store(0)
						timestampCallbacks.Store(0)
						b.ReportAllocs()
						b.SetBytes(int64(len(input)))
						b.ResetTimer()
						for range b.N {
							realScriptBenchmarkNativeSink, err = runner.BenchmarkNative(source, input)
							if err != nil {
								b.Fatal(err)
							}
						}
						b.ReportMetric(recordsPerSecond(b, batchSize), "records/s")
						b.ReportMetric(float64(raw.Len())/float64(batchSize), "output-bytes/record")
						reportCallbackMetrics(b, batchSize, callbacks.Load(), timestampCallbacks.Load())
					})

					b.Run("process", func(b *testing.B) {
						b.ReportAllocs()
						b.SetBytes(int64(len(input)))
						b.ResetTimer()
						for range b.N {
							realScriptBenchmarkBatchSink, err = runner.Process(source, input)
							if err != nil {
								b.Fatal(err)
							}
						}
						b.ReportMetric(recordsPerSecond(b, batchSize), "records/s")
					})

					b.Run("decode", func(b *testing.B) {
						b.ReportAllocs()
						b.ResetTimer()
						for range b.N {
							realScriptBenchmarkBatchSink, err = raw.Decode()
							if err != nil {
								b.Fatal(err)
							}
						}
						b.ReportMetric(recordsPerSecond(b, batchSize), "records/s")
					})

					b.Run("decode-reused", func(b *testing.B) {
						var scratch pljit.Batch
						for range 2 {
							got, err := raw.DecodeInto(&scratch)
							if err != nil || !reflect.DeepEqual(got, decoded) {
								b.Fatalf("reused decoding differs from reference: %v", err)
							}
						}
						b.ReportAllocs()
						b.ResetTimer()
						for range b.N {
							realScriptBenchmarkBatchSink, err = raw.DecodeInto(&scratch)
							if err != nil {
								b.Fatal(err)
							}
						}
					})

					b.Run("apply", func(b *testing.B) {
						b.ReportAllocs()
						b.ResetTimer()
						for range b.N {
							b.StopTimer()
							points := newPoints(batchSize)
							b.StartTimer()
							if err := applyRealScriptBenchmarkBatch(point.Logging, points, decoded); err != nil {
								b.Fatal(err)
							}
						}
						realScriptBenchmarkBooleanSink = len(decoded.Records) == batchSize
						b.ReportMetric(recordsPerSecond(b, batchSize), "records/s")
					})
				})
			}
		})
	}
}

func applyRealScriptBenchmarkBatch(
	category point.Category,
	points []*point.Point,
	batch pljit.Batch,
) error {
	if len(batch.Records) != len(points) {
		return fmt.Errorf("JIT returned %d records for %d points", len(batch.Records), len(points))
	}
	for index := range points {
		if err := applyBenchmarkJITRecord(category, points[index], batch, index); err != nil {
			return err
		}
		ptinput.PtWrap(category, points[index]).KeyTime2Time()
	}
	return nil
}

func recordsPerSecond(b *testing.B, batchSize int) float64 {
	return float64(b.N*batchSize) / b.Elapsed().Seconds()
}

func reportCallbackMetrics(b *testing.B, batchSize int, total, timestamp uint64) {
	records := float64(b.N * batchSize)
	b.ReportMetric(float64(total)/records, "host-callbacks/record")
	b.ReportMetric(float64(timestamp)/records, "timestamp-callbacks/record")
}

// CPU time is diagnostic alongside wall time on shared development machines.
// It includes Go GC work and native execution, excluding time descheduled by
// Linux. Wall-clock end-to-end throughput remains the acceptance criterion.
func benchmarkProcessCPU(b *testing.B) int64 {
	b.Helper()
	var usage syscall.Rusage
	if err := syscall.Getrusage(syscall.RUSAGE_SELF, &usage); err != nil {
		b.Fatal(err)
	}
	return usage.Utime.Nano() + usage.Stime.Nano()
}
