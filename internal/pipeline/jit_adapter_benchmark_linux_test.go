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
	"testing"

	"github.com/GuanceCloud/cliutils/logger"
	"github.com/GuanceCloud/cliutils/point"
	pljit "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline/jit"
)

var adapterBatchSink pljit.Batch
var adapterBytesSink []byte

func BenchmarkPipelineJITAdapter(b *testing.B) {
	if err := logger.InitRoot(&logger.Option{Level: logger.WARN, Flags: logger.OPT_DEFAULT}); err != nil {
		b.Fatal(err)
	}
	r, err := pljit.NewRunnerWithHost(os.Getenv("PLATYPUS_JIT_RUNTIME"), "pipeline-go-1.4.3-datakit", 32, pljit.NewPipelineGoHost(nil))
	if err != nil {
		b.Fatal(err)
	}
	b.Cleanup(func() {
		if err := r.Close(); err != nil {
			b.Error(err)
		}
	})
	source := "json(message, payload.user.id, user_id)\njson(message, payload.request.path, request_path)\njson(message, payload.ok, ok)\ncast(user_id, \"int\")\n"
	projection, err := r.Projection(source)
	if err != nil {
		b.Fatal(err)
	}
	script, err := NewPlScriptSimple(point.Logging, "json-3-cast.p", source)
	if err != nil {
		b.Fatal(err)
	}
	for _, n := range []int{1, 10} {
		b.Run(fmt.Sprintf("batch-%d", n), func(b *testing.B) {
			points := make([]*point.Point, n)
			for i := range points {
				points[i] = newRealScriptPoint("json-3-cast", map[string]any{"message": `{"payload":{"user":{"id":"42"},"request":{"path":"/health"},"ok":true}}`, "sentinel": "keep", "status": "unknown"})
			}
			indexes, runs := make([]int, n), make([]pointRun, n)
			for i, pt := range points {
				indexes[i], runs[i].point = i, pt
			}
			var warm projectedJITEncoder
			input, err := warm.encode(point.Logging, points, projection)
			if err != nil {
				b.Fatal(err)
			}
			bound, err := r.BindSource(source)
			if err != nil {
				b.Fatal(err)
			}
			b.Cleanup(func() {
				if err := bound.Close(); err != nil {
					b.Error(err)
				}
			})
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			for _, mode := range []string{"encode-reused", "chunk-original", "chunk-prepared", "encode-group-lifetime", "process-static", "process-cancellable", "e2e-static", "e2e-cancellable", "e2e-group", "projection-query"} {
				b.Run(mode, func(b *testing.B) {
					b.ReportAllocs()
					b.ResetTimer()
					for i := 0; i < b.N; i++ {
						switch mode {
						case "projection-query":
							projection, err = bound.Projection(source)
						case "chunk-original":
							end := nextJITChunkEnd(point.Logging, indexes, runs, projection, 0)
							adapterBytesSink, err = warm.encode(point.Logging, points[:end], projection)
						case "chunk-prepared":
							end := warm.prepareChunk(point.Logging, indexes, runs, projection, 0)
							adapterBytesSink, err = warm.encodePrepared(end, projection)
						case "encode-reused":
							adapterBytesSink, err = warm.encode(point.Logging, points, projection)
						case "encode-group-lifetime":
							var encoder projectedJITEncoder
							adapterBytesSink, err = encoder.encode(point.Logging, points, projection)
						case "process-static":
							adapterBatchSink, err = bound.Process(source, input)
						case "process-cancellable":
							adapterBatchSink, err = bound.ProcessContext(ctx, source, input)
						case "e2e-static", "e2e-cancellable", "e2e-group":
							pts := make([]*point.Point, n)
							for j := range pts {
								pts[j] = newRealScriptPoint("json-3-cast", map[string]any{"message": `{"payload":{"user":{"id":"42"},"request":{"path":"/health"},"ok":true}}`, "sentinel": "keep", "status": "unknown"})
							}
							if mode == "e2e-group" {
								indexes := make([]int, n)
								runs := make([]pointRun, n)
								for j := range pts {
									indexes[j] = j
									runs[j] = pointRun{point: pts[j], script: script}
								}
								runJITGroup(r, point.Logging, script, indexes, runs, nil)
								for j := range runs {
									if runs[j].output == nil {
										b.Fatal("group failed")
									}
								}
							} else {
								var encoded []byte
								encoded, err = warm.encode(point.Logging, pts, projection)
								if err != nil {
									b.Fatal(err)
								}
								if mode == "e2e-static" {
									adapterBatchSink, err = bound.Process(source, encoded)
								} else {
									adapterBatchSink, err = bound.ProcessContext(ctx, source, encoded)
								}
								if err != nil {
									b.Fatal(err)
								}
								err = applyRealScriptBenchmarkBatch(point.Logging, pts, adapterBatchSink)
							}
							for _, pt := range pts {
								if pt.Get("user_id") != int64(42) {
									b.Fatal("wrong result")
								}
							}
						}
						if err != nil {
							b.Fatal(err)
						}
					}
					if mode == "process-static" && adapterBatchSink.Static == nil {
						b.Fatal("static fast path missing")
					}
					if mode == "process-cancellable" && adapterBatchSink.Static != nil {
						b.Fatal("unexpected static cancellable path")
					}
				})
			}
		})
	}
}
