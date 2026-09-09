// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2026-present Guance, Inc.

package pipeline

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"github.com/GuanceCloud/pipeline-go/lang"
	pljit "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline/jit"
)

func TestJITDeltaDeleteRespectsFieldKind(t *testing.T) {
	for _, tag := range []bool{false, true} {
		pt := point.NewPoint("test", point.NewKVs(map[string]any{"key": "old"}))
		set := pljit.MutationOp{Kind: pljit.MutationSetField, Key: "key", Value: "new"}
		del := pljit.MutationOp{Kind: pljit.MutationDeleteTag, Key: "key"}
		if tag {
			set = pljit.MutationOp{Kind: pljit.MutationSetTag, Key: "key", String: "new"}
			del.Kind = pljit.MutationDeleteField
		} else {
			pt.SetTag("key", "old")
		}
		data, err := pljit.EncodeMutationDeltas([]pljit.MutationDelta{{RecordIndex: 0, Operations: []pljit.MutationOp{set, del}}})
		if err != nil {
			t.Fatal(err)
		}
		if _, _, err = applyJITDelta(point.Logging, pt, data, 0, nil); err != nil {
			t.Fatal(err)
		}
		if pt.Get("key") != "new" {
			t.Fatalf("tag=%v: replacement lost", tag)
		}
	}
}

func applyBenchmarkJITRecord(
	category point.Category,
	target *point.Point,
	batch pljit.Batch,
	recordIndex int,
) error {
	if recordIndex < 0 || recordIndex >= len(batch.Records) {
		return fmt.Errorf("JIT benchmark record index %d is out of range", recordIndex)
	}
	if batch.Records[recordIndex].Status != pljit.TerminalOK {
		return fmt.Errorf("JIT benchmark record %d status is %d", recordIndex, batch.Records[recordIndex].Status)
	}
	if batch.Static != nil {
		_, _, err := applyJITStatic(category, target, batch.Static, recordIndex, nil)
		return err
	}
	_, _, err := applyJITRecord(category, target, batch.Records[recordIndex], uint64(recordIndex), nil)
	return err
}

func TestJITAppliesRawNilMapAtomically(t *testing.T) {
	target := point.NewPoint("safe", point.KVs{point.NewKV("keep", int64(1))}, point.DefaultLoggingOptions()...)
	encoded, err := pljit.EncodeMutationDeltas([]pljit.MutationDelta{{RecordIndex: 0, Operations: []pljit.MutationOp{
		{Kind: pljit.MutationSetField, Key: "keep", Value: int64(2)},
		{Kind: pljit.MutationSetField, Key: "unsafe", Value: map[string]any{"x": nil}},
	}}})
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := applyJITDelta(point.Logging, target, encoded, 0, nil); err != nil {
		t.Fatal(err)
	}
	if target.Get("keep") != int64(2) {
		t.Fatalf("raw nil map was not committed exactly: %#v", target.KVMap())
	}
	assertPhysicalPointMap(t, target, "unsafe", map[string]*point.BasicTypes{"x": nil})
}

func assertPhysicalPointMap(t *testing.T, target *point.Point, key string, values map[string]*point.BasicTypes) {
	t.Helper()
	field := point.KVs(target.PBPoint().Fields).Get(key)
	if field == nil || field.GetA() == nil {
		t.Fatalf("%s is not a physical Point map: %#v", key, target.PBPoint().Fields)
	}
	want, err := point.NewAny(&point.Map{Map: values})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(field.GetA(), want) {
		t.Fatalf("%s physical map=%#v want=%#v", key, field.GetA(), want)
	}
}

func BenchmarkPipelineJITBatch(b *testing.B) {
	runtimePath := os.Getenv("PLATYPUS_JIT_RUNTIME")
	if runtimePath == "" {
		b.Skip("set PLATYPUS_JIT_RUNTIME to benchmark the JIT")
	}
	const (
		batchSize = 1024
		source    = "json(message, payload.user.id, user_id)\njson(message, payload.request.path, request_path)\njson(message, payload.ok, ok)\ncast(user_id, \"int\")\n"
	)
	script, err := NewPlScriptSimple(point.Logging, "benchmark.p", source)
	if err != nil {
		b.Fatalf("create benchmark script: %v", err)
	}
	runner, err := pljit.NewRunner(runtimePath, "pipeline-go-1.4.3", 8)
	if err != nil {
		b.Fatalf("open JIT runner: %v", err)
	}
	defer runner.Close()

	newPoints := func() []*point.Point {
		points := make([]*point.Point, batchSize)
		for index := range points {
			message := fmt.Sprintf(`{"payload":{"user":{"id":"%d","name":"bench"},"request":{"path":"/api/v1/items/%d"},"ok":true}}`, 1000+index, index)
			points[index] = point.NewPoint(
				"benchmark",
				point.NewKVs(map[string]any{"message": message}),
				point.WithTime(time.Unix(1_700_000_000, int64(index))),
			)
		}
		return points
	}

	warmup, err := encodeJITPoints(point.Logging, newPoints())
	if err != nil {
		b.Fatalf("encode warmup: %v", err)
	}
	warmupBatch, err := runner.Process(source, warmup)
	if err != nil {
		b.Fatalf("warm up JIT: %v", err)
	}

	b.Run("encode-flat", func(b *testing.B) {
		points := newPoints()
		b.ReportAllocs()
		b.ResetTimer()
		for range b.N {
			if _, err := encodeJITPoints(point.Logging, points); err != nil {
				b.Fatal(err)
			}
		}
		b.ReportMetric(float64(b.N*batchSize)/b.Elapsed().Seconds(), "records/s")
	})

	b.Run("jit-native", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for range b.N {
			if _, err := runner.Process(source, warmup); err != nil {
				b.Fatal(err)
			}
		}
		b.ReportMetric(float64(b.N*batchSize)/b.Elapsed().Seconds(), "records/s")
	})

	b.Run("apply-delta", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for range b.N {
			points := newPoints()
			for index := range warmupBatch.Records {
				if err := applyBenchmarkJITRecord(point.Logging, points[index], warmupBatch, index); err != nil {
					b.Fatal(err)
				}
			}
		}
		b.ReportMetric(float64(b.N*batchSize)/b.Elapsed().Seconds(), "records/s")
	})

	b.Run("pipeline-go", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for range b.N {
			for _, pt := range newPoints() {
				input := pointRun{point: pt, script: script, started: time.Now()}
				runPipelineGo(point.Logging, &input, nil)
				if input.output == nil {
					b.Fatal("pipeline-go returned no point")
				}
			}
		}
		b.ReportMetric(float64(b.N*batchSize)/b.Elapsed().Seconds(), "records/s")
	})

	b.Run("jit", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for range b.N {
			points := newPoints()
			input, err := encodeJITPoints(point.Logging, points)
			if err != nil {
				b.Fatal(err)
			}
			batch, err := runner.Process(source, input)
			if err != nil {
				b.Fatal(err)
			}
			for index := range batch.Records {
				if err := applyBenchmarkJITRecord(point.Logging, points[index], batch, index); err != nil {
					b.Fatal(err)
				}
			}
		}
		b.ReportMetric(float64(b.N*batchSize)/b.Elapsed().Seconds(), "records/s")
	})
}

func BenchmarkPipelineJITInputProjection(b *testing.B) {
	runtimePath := os.Getenv("PLATYPUS_JIT_RUNTIME")
	if runtimePath == "" {
		b.Skip("set PLATYPUS_JIT_RUNTIME to benchmark the JIT")
	}
	const batchSize = 1024
	var sourceBuilder strings.Builder
	var messageBuilder strings.Builder
	messageBuilder.WriteString(`{"payload":{`)
	for index := range 16 {
		if index != 0 {
			messageBuilder.WriteByte(',')
		}
		fmt.Fprintf(&sourceBuilder, "json(message, payload.value%02d, value%02d)\n", index, index)
		if index == 15 {
			fmt.Fprint(&messageBuilder, `"value15":{"b":2,"a":[1,true,null]}`)
		} else {
			fmt.Fprintf(&messageBuilder, `"value%02d":"result-%02d"`, index, index)
		}
	}
	messageBuilder.WriteString("}}")
	source := sourceBuilder.String()
	script, err := NewPlScriptSimple(point.Logging, "input-projection.p", source)
	if err != nil {
		b.Fatalf("create benchmark script: %v", err)
	}
	runner, err := pljit.NewRunner(runtimePath, "pipeline-go-1.4.3-datakit", 8)
	if err != nil {
		b.Fatalf("open JIT runner: %v", err)
	}
	defer runner.Close()
	projection, err := runner.Projection(source)
	if err != nil {
		b.Fatalf("query JIT projection: %v", err)
	}
	newPoints := func() []*point.Point {
		points := make([]*point.Point, batchSize)
		for recordIndex := range points {
			fields := make(map[string]any, 34)
			fields["message"] = messageBuilder.String()
			fields["value00"] = "old-field"
			for fieldIndex := range 32 {
				fields[fmt.Sprintf("irrelevant_%02d", fieldIndex)] = strings.Repeat("not-sent", 16)
			}
			kvs := point.NewKVs(fields).SetTag("value01", "old-tag")
			points[recordIndex] = point.NewPoint("projection", kvs)
		}
		return points
	}
	points := newPoints()
	fullInput, err := encodeJITPoints(point.Logging, points)
	if err != nil {
		b.Fatalf("encode full input: %v", err)
	}
	projectedInput, err := encodeProjectedJITPoints(point.Logging, points, projection)
	if err != nil {
		b.Fatalf("encode projected input: %v", err)
	}

	for _, benchmark := range []struct {
		name       string
		projection pljit.InputProjection
		input      []byte
	}{
		{name: "full", input: fullInput},
		{name: "projected", projection: projection, input: projectedInput},
	} {
		benchmark := benchmark
		b.Run("encode-"+benchmark.name, func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			for range b.N {
				if _, err := encodeProjectedJITPoints(point.Logging, points, benchmark.projection); err != nil {
					b.Fatal(err)
				}
			}
			b.ReportMetric(float64(b.N*batchSize)/b.Elapsed().Seconds(), "records/s")
			b.ReportMetric(float64(len(benchmark.input)), "input-bytes")
		})
		b.Run("native-"+benchmark.name, func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			for range b.N {
				if _, err := runner.Process(source, benchmark.input); err != nil {
					b.Fatal(err)
				}
			}
			b.ReportMetric(float64(b.N*batchSize)/b.Elapsed().Seconds(), "records/s")
			b.ReportMetric(float64(len(benchmark.input)), "input-bytes")
		})
		b.Run("end-to-end-"+benchmark.name, func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			for range b.N {
				points := newPoints()
				input, err := encodeProjectedJITPoints(point.Logging, points, benchmark.projection)
				if err != nil {
					b.Fatal(err)
				}
				batch, err := runner.Process(source, input)
				if err != nil {
					b.Fatal(err)
				}
				for index := range batch.Records {
					if err := applyBenchmarkJITRecord(point.Logging, points[index], batch, index); err != nil {
						b.Fatal(err)
					}
				}
			}
			b.ReportMetric(float64(b.N*batchSize)/b.Elapsed().Seconds(), "records/s")
		})
	}

	b.Run("pipeline-go", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for range b.N {
			for _, pt := range newPoints() {
				input := pointRun{point: pt, script: script, started: time.Now()}
				runPipelineGo(point.Logging, &input, nil)
			}
		}
		b.ReportMetric(float64(b.N*batchSize)/b.Elapsed().Seconds(), "records/s")
	})
}

func BenchmarkPipelineJITSelectiveJSON(b *testing.B) {
	runtimePath := os.Getenv("PLATYPUS_JIT_RUNTIME")
	if runtimePath == "" {
		b.Skip("set PLATYPUS_JIT_RUNTIME to benchmark the JIT")
	}
	const batchSize = 1024
	var sourceBuilder strings.Builder
	var messageBuilder strings.Builder
	messageBuilder.WriteString(`{"payload":{`)
	for index := range 16 {
		if index != 0 {
			messageBuilder.WriteByte(',')
		}
		fmt.Fprintf(&sourceBuilder, "json(message, payload.value%02d, value%02d)\n", index, index)
		fmt.Fprintf(&messageBuilder, `"value%02d":"result-%02d"`, index, index)
	}
	messageBuilder.WriteString(`},"noise":{`)
	for index := range 128 {
		if index != 0 {
			messageBuilder.WriteByte(',')
		}
		fmt.Fprintf(&messageBuilder, `"branch%03d":{"text":"ignored-%03d","values":[1,2,3,4]}`, index, index)
	}
	messageBuilder.WriteString("}}")
	source := sourceBuilder.String()
	script, err := NewPlScriptSimple(point.Logging, "selective-json.p", source)
	if err != nil {
		b.Fatalf("create benchmark script: %v", err)
	}
	runner, err := pljit.NewRunner(runtimePath, "pipeline-go-1.4.3-datakit", 8)
	if err != nil {
		b.Fatalf("open JIT runner: %v", err)
	}
	defer runner.Close()
	projection, err := runner.Projection(source)
	if err != nil {
		b.Fatalf("query JIT projection: %v", err)
	}
	newPoints := func() []*point.Point {
		points := make([]*point.Point, batchSize)
		for index := range points {
			points[index] = point.NewPoint("selective-json", point.NewKVs(map[string]any{
				"message": messageBuilder.String(),
			}))
		}
		return points
	}

	b.Run("pipeline-go", func(b *testing.B) {
		b.ReportAllocs()
		for range b.N {
			for _, pt := range newPoints() {
				input := pointRun{point: pt, script: script, started: time.Now()}
				runPipelineGo(point.Logging, &input, nil)
			}
		}
		b.ReportMetric(float64(b.N*batchSize)/b.Elapsed().Seconds(), "records/s")
	})
	b.Run("jit", func(b *testing.B) {
		b.ReportAllocs()
		for range b.N {
			points := newPoints()
			input, err := encodeProjectedJITPoints(point.Logging, points, projection)
			if err != nil {
				b.Fatal(err)
			}
			batch, err := runner.Process(source, input)
			if err != nil {
				b.Fatal(err)
			}
			for index := range batch.Records {
				if err := applyBenchmarkJITRecord(point.Logging, points[index], batch, index); err != nil {
					b.Fatal(err)
				}
			}
		}
		b.ReportMetric(float64(b.N*batchSize)/b.Elapsed().Seconds(), "records/s")
	})
}

func BenchmarkPipelineJITComputeHeavy(b *testing.B) {
	runtimePath := os.Getenv("PLATYPUS_JIT_RUNTIME")
	if runtimePath == "" {
		b.Skip("set PLATYPUS_JIT_RUNTIME to benchmark the JIT")
	}
	const batchSize = 1024
	runner, err := pljit.NewRunner(runtimePath, "pipeline-go-1.4.3", 8)
	if err != nil {
		b.Fatalf("open JIT runner: %v", err)
	}
	defer runner.Close()
	newPoints := func() []*point.Point {
		points := make([]*point.Point, batchSize)
		for index := range points {
			points[index] = point.NewPoint("benchmark", point.NewKVs(map[string]any{"n": int64(index)}))
		}
		return points
	}

	for _, callCount := range []int{1, 2, 4, 8, 16, 32, 64} {
		callCount := callCount
		b.Run(fmt.Sprintf("calls-%02d", callCount), func(b *testing.B) {
			source := strings.Repeat("cast(n, \"int\")\n", callCount)
			script, err := NewPlScriptSimple(point.Logging, "compute-heavy.p", source)
			if err != nil {
				b.Fatalf("create benchmark script: %v", err)
			}
			b.Run("pipeline-go", func(b *testing.B) {
				b.ReportAllocs()
				for range b.N {
					for _, pt := range newPoints() {
						input := pointRun{point: pt, script: script, started: time.Now()}
						runPipelineGo(point.Logging, &input, nil)
					}
				}
				b.ReportMetric(float64(b.N*batchSize)/b.Elapsed().Seconds(), "records/s")
			})
			b.Run("jit", func(b *testing.B) {
				b.ReportAllocs()
				for range b.N {
					points := newPoints()
					input, err := encodeJITPoints(point.Logging, points)
					if err != nil {
						b.Fatal(err)
					}
					batch, err := runner.Process(source, input)
					if err != nil {
						b.Fatal(err)
					}
					for index := range batch.Records {
						if err := applyBenchmarkJITRecord(point.Logging, points[index], batch, index); err != nil {
							b.Fatal(err)
						}
					}
				}
				b.ReportMetric(float64(b.N*batchSize)/b.Elapsed().Seconds(), "records/s")
			})
		})
	}
}

func BenchmarkPipelineJITRegex(b *testing.B) {
	runtimePath := os.Getenv("PLATYPUS_JIT_RUNTIME")
	if runtimePath == "" {
		b.Skip("set PLATYPUS_JIT_RUNTIME to benchmark the JIT")
	}
	const batchSize = 1024
	runner, err := pljit.NewRunner(runtimePath, "pipeline-go-1.4.3", 8)
	if err != nil {
		b.Fatalf("open JIT runner: %v", err)
	}
	defer runner.Close()
	newPoints := func() []*point.Point {
		points := make([]*point.Point, batchSize)
		for index := range points {
			points[index] = point.NewPoint("benchmark", point.NewKVs(map[string]any{"n": "abc123def456"}))
		}
		return points
	}

	for _, callCount := range []int{1, 2, 4, 8, 16, 32} {
		callCount := callCount
		b.Run(fmt.Sprintf("calls-%02d", callCount), func(b *testing.B) {
			source := strings.Repeat("replace(n, \"[a-z]+\", \"x\")\n", callCount)
			script, err := NewPlScriptSimple(point.Logging, "regex.p", source)
			if err != nil {
				b.Fatalf("create benchmark script: %v", err)
			}
			b.Run("pipeline-go", func(b *testing.B) {
				b.ReportAllocs()
				for range b.N {
					for _, pt := range newPoints() {
						input := pointRun{point: pt, script: script, started: time.Now()}
						runPipelineGo(point.Logging, &input, nil)
					}
				}
				b.ReportMetric(float64(b.N*batchSize)/b.Elapsed().Seconds(), "records/s")
			})
			b.Run("jit", func(b *testing.B) {
				b.ReportAllocs()
				for range b.N {
					points := newPoints()
					input, err := encodeJITPoints(point.Logging, points)
					if err != nil {
						b.Fatal(err)
					}
					batch, err := runner.Process(source, input)
					if err != nil {
						b.Fatal(err)
					}
					for index := range batch.Records {
						if err := applyBenchmarkJITRecord(point.Logging, points[index], batch, index); err != nil {
							b.Fatal(err)
						}
					}
				}
				b.ReportMetric(float64(b.N*batchSize)/b.Elapsed().Seconds(), "records/s")
			})
		})
	}
}

func BenchmarkPipelineJITJSONProjection(b *testing.B) {
	runtimePath := os.Getenv("PLATYPUS_JIT_RUNTIME")
	if runtimePath == "" {
		b.Skip("set PLATYPUS_JIT_RUNTIME to benchmark the JIT")
	}
	const batchSize = 1024
	runner, err := pljit.NewRunner(runtimePath, "pipeline-go-1.4.3", 8)
	if err != nil {
		b.Fatalf("open JIT runner: %v", err)
	}
	defer runner.Close()
	newPoints := func() []*point.Point {
		points := make([]*point.Point, batchSize)
		for index := range points {
			var message strings.Builder
			message.WriteString(`{"payload":{`)
			for valueIndex := range 32 {
				if valueIndex != 0 {
					message.WriteByte(',')
				}
				fmt.Fprintf(&message, `"value%02d":"%d"`, valueIndex, index+valueIndex)
			}
			message.WriteString("}}")
			points[index] = point.NewPoint("benchmark", point.NewKVs(map[string]any{"message": message.String()}))
		}
		return points
	}

	for _, callCount := range []int{1, 2, 4, 8, 16, 32} {
		callCount := callCount
		b.Run(fmt.Sprintf("calls-%02d", callCount), func(b *testing.B) {
			var sourceBuilder strings.Builder
			for index := range callCount {
				fmt.Fprintf(&sourceBuilder, "json(message, payload.value%02d, value%02d)\n", index, index)
			}
			source := sourceBuilder.String()
			script, err := NewPlScriptSimple(point.Logging, "json-projection.p", source)
			if err != nil {
				b.Fatalf("create benchmark script: %v", err)
			}
			b.Run("pipeline-go", func(b *testing.B) {
				b.ReportAllocs()
				for range b.N {
					for _, pt := range newPoints() {
						input := pointRun{point: pt, script: script, started: time.Now()}
						runPipelineGo(point.Logging, &input, nil)
					}
				}
				b.ReportMetric(float64(b.N*batchSize)/b.Elapsed().Seconds(), "records/s")
			})
			b.Run("jit", func(b *testing.B) {
				b.ReportAllocs()
				for range b.N {
					points := newPoints()
					input, err := encodeJITPoints(point.Logging, points)
					if err != nil {
						b.Fatal(err)
					}
					batch, err := runner.Process(source, input)
					if err != nil {
						b.Fatal(err)
					}
					for index := range batch.Records {
						if err := applyBenchmarkJITRecord(point.Logging, points[index], batch, index); err != nil {
							b.Fatal(err)
						}
					}
				}
				b.ReportMetric(float64(b.N*batchSize)/b.Elapsed().Seconds(), "records/s")
			})
		})
	}
}

func BenchmarkPipelineJITProductionCrossover(b *testing.B) {
	runtimePath := os.Getenv("PLATYPUS_JIT_RUNTIME")
	if runtimePath == "" {
		b.Skip("set PLATYPUS_JIT_RUNTIME to benchmark the JIT")
	}
	runner, err := pljit.NewRunner(runtimePath, "pipeline-go-1.4.3", 8)
	if err != nil {
		b.Fatalf("open JIT runner: %v", err)
	}
	defer runner.Close()
	var jsonSourceBuilder strings.Builder
	for index := range 8 {
		fmt.Fprintf(&jsonSourceBuilder, "json(message, payload.value%02d, value%02d)\n", index, index)
	}
	workloads := []struct {
		name   string
		source string
		fields func(int) map[string]any
	}{
		{
			name:   "json-08",
			source: jsonSourceBuilder.String(),
			fields: func(record int) map[string]any {
				return map[string]any{"message": fmt.Sprintf(`{"payload":{"value00":"%d","value01":"1","value02":"2","value03":"3","value04":"4","value05":"5","value06":"6","value07":"7"}}`, record)}
			},
		},
		{
			name:   "regex-16",
			source: strings.Repeat("replace(n, \"[a-z]+\", \"x\")\n", 16),
			fields: func(int) map[string]any { return map[string]any{"n": "abc123def456"} },
		},
	}
	for _, workload := range workloads {
		workload := workload
		b.Run(workload.name, func(b *testing.B) {
			script, err := NewPlScriptSimple(point.Logging, "crossover.p", workload.source)
			if err != nil {
				b.Fatalf("create benchmark script: %v", err)
			}
			for _, batchSize := range []int{128, 256, 512, 1024} {
				batchSize := batchSize
				newPoints := func() []*point.Point {
					points := make([]*point.Point, batchSize)
					for index := range points {
						points[index] = point.NewPoint("benchmark", point.NewKVs(workload.fields(index)))
					}
					return points
				}
				b.Run(fmt.Sprintf("batch-%04d", batchSize), func(b *testing.B) {
					b.Run("pipeline-go", func(b *testing.B) {
						for range b.N {
							for _, pt := range newPoints() {
								input := pointRun{point: pt, script: script, started: time.Now()}
								runPipelineGo(point.Logging, &input, nil)
							}
						}
						b.ReportMetric(float64(b.N*batchSize)/b.Elapsed().Seconds(), "records/s")
					})
					b.Run("jit", func(b *testing.B) {
						for range b.N {
							points := newPoints()
							input, err := encodeJITPoints(point.Logging, points)
							if err != nil {
								b.Fatal(err)
							}
							batch, err := runner.Process(workload.source, input)
							if err != nil {
								b.Fatal(err)
							}
							for index := range batch.Records {
								if err := applyBenchmarkJITRecord(point.Logging, points[index], batch, index); err != nil {
									b.Fatal(err)
								}
							}
						}
						b.ReportMetric(float64(b.N*batchSize)/b.Elapsed().Seconds(), "records/s")
					})
				})
			}
		})
	}
}

func BenchmarkPipelineJITNativeParallel(b *testing.B) {
	runtimePath := os.Getenv("PLATYPUS_JIT_RUNTIME")
	if runtimePath == "" {
		b.Skip("set PLATYPUS_JIT_RUNTIME to benchmark the JIT")
	}
	const batchSize = 128
	source := strings.Repeat("cast(n, \"int\")\n", 16)
	runner, err := pljit.NewRunner(runtimePath, "pipeline-go-1.4.3", 8)
	if err != nil {
		b.Fatalf("open JIT runner: %v", err)
	}
	defer runner.Close()
	points := make([]*point.Point, batchSize)
	for index := range points {
		points[index] = point.NewPoint("benchmark", point.NewKVs(map[string]any{"n": int64(index)}))
	}
	input, err := encodeJITPoints(point.Logging, points)
	if err != nil {
		b.Fatalf("encode JIT points: %v", err)
	}
	if _, err := runner.Process(source, input); err != nil {
		b.Fatalf("warm up JIT: %v", err)
	}

	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			if _, err := runner.Process(source, input); err != nil {
				b.Error(err)
				return
			}
		}
	})
	b.ReportMetric(float64(b.N*batchSize)/b.Elapsed().Seconds(), "records/s")
}

func BenchmarkPipelineJITStaticGrok(b *testing.B) {
	runtimePath := os.Getenv("PLATYPUS_JIT_RUNTIME")
	if runtimePath == "" {
		b.Skip("set PLATYPUS_JIT_RUNTIME to benchmark the JIT")
	}
	const (
		batchSize = 1024
		source    = "grok(_, \"%{IPORHOST:client_ip} %{WORD:http_method} %{URIPATHPARAM:http_url} %{INT:status_code:int} %{INT:bytes:int}\")\n"
		message   = "10.20.30.40 GET /api/v1/pipeline?q=json 200 512"
	)
	script, err := NewPlScriptSimple(point.Logging, "grok-benchmark.p", source)
	if err != nil {
		b.Fatalf("create Grok benchmark script: %v", err)
	}
	runner, err := pljit.NewRunner(runtimePath, "pipeline-go-1.4.3-datakit", 8)
	if err != nil {
		b.Fatalf("open JIT runner: %v", err)
	}
	defer runner.Close()
	projection, err := runner.Projection(source)
	if err != nil {
		b.Fatalf("query Grok projection: %v", err)
	}

	newPoints := func() []*point.Point {
		points := make([]*point.Point, batchSize)
		for index := range points {
			points[index] = point.NewPoint(
				"grok-benchmark",
				point.NewKVs(map[string]any{"message": message}),
				point.WithTime(time.Unix(1_700_000_000, int64(index))),
			)
		}
		return points
	}

	b.Run("pipeline-go", func(b *testing.B) {
		b.ReportAllocs()
		for range b.N {
			for _, pt := range newPoints() {
				run := pointRun{point: pt, script: script, started: time.Now()}
				runPipelineGo(point.Logging, &run, nil)
				if run.output == nil {
					b.Fatal("pipeline-go returned no point")
				}
			}
		}
		b.ReportMetric(float64(b.N*batchSize)/b.Elapsed().Seconds(), "records/s")
	})

	b.Run("jit", func(b *testing.B) {
		b.ReportAllocs()
		for range b.N {
			points := newPoints()
			input, err := encodeProjectedJITPoints(point.Logging, points, projection)
			if err != nil {
				b.Fatal(err)
			}
			batch, err := runner.Process(source, input)
			if err != nil {
				b.Fatal(err)
			}
			if batch.Static == nil {
				b.Fatal("Grok JIT did not use static output")
			}
			for index := range points {
				if _, dropped, err := applyJITStatic(point.Logging, points[index], batch.Static, index, nil); err != nil || dropped {
					b.Fatalf("apply Grok JIT output: dropped=%v err=%v", dropped, err)
				}
			}
		}
		b.ReportMetric(float64(b.N*batchSize)/b.Elapsed().Seconds(), "records/s")
	})
}

func TestJITMatchesPipelineGo(t *testing.T) {
	runtimePath := os.Getenv("PLATYPUS_JIT_RUNTIME")
	if runtimePath == "" {
		t.Skip("set PLATYPUS_JIT_RUNTIME to test the JIT runtime")
	}
	const source = "json(message, payload.user.id, user_id)\njson(message, payload.request.path, request_path)\njson(message, payload.ok, ok)\ncast(user_id, \"int\")\ndrop_key(unused)\n"
	script, err := NewPlScriptSimple(point.Logging, "equivalence.p", source)
	if err != nil {
		t.Fatalf("create script: %v", err)
	}
	runner, err := pljit.NewRunner(runtimePath, "pipeline-go-1.4.3", 8)
	if err != nil {
		t.Fatalf("open JIT runner: %v", err)
	}
	defer runner.Close()

	newPoint := func() *point.Point {
		return point.NewPoint(
			"equivalence",
			point.NewKVs(map[string]any{
				"message": `{"payload":{"user":{"id":"42"},"request":{"path":"/api/v1/items/42"},"ok":true}}`,
				"unused":  "remove",
			}),
			point.WithTime(time.Unix(1_700_000_000, 123)),
		)
	}
	want := newPoint()
	run := pointRun{point: want, script: script, started: time.Now()}
	runPipelineGo(point.Logging, &run, nil)
	if run.output == nil {
		t.Fatal("pipeline-go returned no point")
	}

	got := newPoint()
	input, err := encodeJITPoints(point.Logging, []*point.Point{got})
	if err != nil {
		t.Fatalf("encode JIT point: %v", err)
	}
	batch, err := runner.Process(source, input)
	if err != nil {
		t.Fatalf("run JIT: %v", err)
	}
	if len(batch.Records) != 1 || batch.Records[0].Status != pljit.TerminalOK {
		t.Fatalf("unexpected JIT batch: %#v", batch)
	}
	if _, dropped, err := applyJITRecord(point.Logging, got, batch.Records[0], 0, nil); err != nil || dropped {
		t.Fatalf("apply JIT delta: dropped=%v err=%v", dropped, err)
	}
	if equal, reason := got.EqualWithReason(want); !equal {
		t.Fatalf("JIT differs from pipeline-go: %s\nJIT: %#v\npipeline-go: %#v", reason, got.KVMap(), want.KVMap())
	}
}

func TestJITHybridMutatorsMatchPipelineGo(t *testing.T) {
	runtimePath := os.Getenv("PLATYPUS_JIT_RUNTIME")
	if runtimePath == "" {
		t.Skip("set PLATYPUS_JIT_RUNTIME to test the JIT runtime")
	}
	const source = `set_measurement("renamed")
pt_kvs_set(target_name, copied)
json_all(json_blob, [], ["*"])
drop_key(remove_me)
exit()
add_key(after_exit, "bad")
`
	script, err := NewPlScriptSimple(point.Logging, "hybrid-equivalence.p", source)
	if err != nil {
		t.Fatalf("create script: %v", err)
	}
	runner, err := pljit.NewRunner(runtimePath, "pipeline-go-1.4.3-datakit", 8)
	if err != nil {
		t.Fatalf("open JIT runner: %v", err)
	}
	defer runner.Close()
	projection, err := runner.Projection(source)
	if err != nil {
		t.Fatalf("query hybrid projection: %v", err)
	}
	if !projection.IsAll() {
		t.Fatalf("dynamic mutators must receive the complete Point: %v", projection.Keys())
	}

	newPoint := func() *point.Point {
		return point.NewPoint(
			"original",
			point.NewKVs(map[string]any{
				"target_name": "dynamic_field",
				"copied":      "copied-value",
				"json_blob":   `{"left":"one","right":"two"}`,
				"remove_me":   "remove",
			}),
			point.WithTime(time.Unix(1_700_000_000, 123)),
		)
	}
	want := newPoint()
	run := pointRun{point: want, script: script, started: time.Now()}
	runPipelineGo(point.Logging, &run, nil)
	if run.output == nil {
		t.Fatal("pipeline-go returned no point")
	}

	got := newPoint()
	input, err := encodeJITPoints(point.Logging, []*point.Point{got})
	if err != nil {
		t.Fatalf("encode JIT point: %v", err)
	}
	batch, err := runner.Process(source, input)
	if err != nil {
		t.Fatalf("run hybrid JIT: %v", err)
	}
	if batch.Static != nil || len(batch.Records) != 1 || batch.Records[0].Status != pljit.TerminalOK {
		t.Fatalf("unexpected hybrid JIT batch: %#v", batch)
	}
	if _, dropped, err := applyJITRecord(point.Logging, got, batch.Records[0], 0, nil); err != nil || dropped {
		t.Fatalf("apply hybrid JIT delta: dropped=%v err=%v", dropped, err)
	}
	if equal, reason := got.EqualWithReason(run.output); !equal {
		t.Fatalf("hybrid JIT differs from pipeline-go: %s\nJIT: %#v\npipeline-go: %#v", reason, got.KVMap(), run.output.KVMap())
	}
}

func TestJITControlFlowMatchesPipelineGo(t *testing.T) {
	runtimePath := os.Getenv("PLATYPUS_JIT_RUNTIME")
	if runtimePath == "" {
		t.Skip("set PLATYPUS_JIT_RUNTIME to test the JIT runtime")
	}
	const source = `data = load_json(_)
json(_, recorder)
if recorder == "gunicorn" {
  add_key(db_session.init_count, data["db_session.init_count"])
  cast(db_session.init_count, "int")
  add_key(db_session.instance_count, data["db_session.instance_count"])
  cast(db_session.instance_count, "int")
  add_key(db_session.max_nest_layer_num, data["db_session.max_nest_layer_num"])
  cast(db_session.max_nest_layer_num, "int")
  add_key(db_session.unclosed_count, data["db_session.unclosed_count"])
  cast(db_session.unclosed_count, "int")
}
`
	script, err := NewPlScriptSimple(point.Logging, "control-flow-equivalence.p", source)
	if err != nil {
		t.Fatalf("create script: %v", err)
	}
	runner, err := pljit.NewRunnerWithHost(
		runtimePath,
		"pipeline-go-1.4.3-datakit",
		8,
		pljit.NewPipelineGoHost(nil),
	)
	if err != nil {
		t.Fatalf("open JIT runner: %v", err)
	}
	defer runner.Close()

	for _, test := range []struct {
		name    string
		message string
	}{
		{
			name: "matching branch",
			message: `{"recorder":"gunicorn","db_session.init_count":"11",` +
				`"db_session.instance_count":"12","db_session.max_nest_layer_num":"13",` +
				`"db_session.unclosed_count":"14"}`,
		},
		{name: "skipped branch", message: `{"recorder":"other","db_session.init_count":"11"}`},
	} {
		t.Run(test.name, func(t *testing.T) {
			newPoint := func() *point.Point {
				return point.NewPoint(
					"control-flow",
					point.NewKVs(map[string]any{"message": test.message}),
					point.WithTime(time.Unix(1_700_000_000, 123)),
				)
			}
			want := newPoint()
			run := pointRun{point: want, script: script, started: time.Now()}
			runPipelineGo(point.Logging, &run, nil)
			if run.output == nil {
				t.Fatal("pipeline-go returned no point")
			}

			got := newPoint()
			input, err := encodeJITPoints(point.Logging, []*point.Point{got})
			if err != nil {
				t.Fatalf("encode JIT point: %v", err)
			}
			batch, err := runner.Process(source, input)
			if err != nil {
				t.Fatalf("run control-flow JIT: %v", err)
			}
			if batch.Static != nil || len(batch.Records) != 1 || batch.Records[0].Status != pljit.TerminalOK {
				t.Fatalf("unexpected control-flow JIT batch: %#v", batch)
			}
			if _, dropped, err := applyJITRecord(point.Logging, got, batch.Records[0], 0, nil); err != nil || dropped {
				t.Fatalf("apply control-flow JIT delta: dropped=%v err=%v", dropped, err)
			}
			if equal, reason := got.EqualWithReason(run.output); !equal {
				t.Fatalf("control-flow JIT differs from pipeline-go: %s\nJIT: %#v\npipeline-go: %#v", reason, got.KVMap(), run.output.KVMap())
			}
		})
	}
}

func TestJITProjectedInputMatchesPipelineGo(t *testing.T) {
	runtimePath := os.Getenv("PLATYPUS_JIT_RUNTIME")
	if runtimePath == "" {
		t.Skip("set PLATYPUS_JIT_RUNTIME to test the JIT runtime")
	}
	var sourceBuilder strings.Builder
	var messageBuilder strings.Builder
	messageBuilder.WriteString(`{"payload":{`)
	for index := range 16 {
		if index != 0 {
			messageBuilder.WriteByte(',')
		}
		fmt.Fprintf(&sourceBuilder, "json(message, payload.value%02d, value%02d)\n", index, index)
		fmt.Fprintf(&messageBuilder, `"value%02d":"result-%02d"`, index, index)
	}
	messageBuilder.WriteString("}}")
	source := sourceBuilder.String()
	script, err := NewPlScriptSimple(point.Logging, "projected-equivalence.p", source)
	if err != nil {
		t.Fatalf("create script: %v", err)
	}
	runner, err := pljit.NewRunner(runtimePath, "pipeline-go-1.4.3-datakit", 8)
	if err != nil {
		t.Fatalf("open JIT runner: %v", err)
	}
	defer runner.Close()
	projection, err := runner.Projection(source)
	if err != nil {
		t.Fatalf("query JIT projection: %v", err)
	}
	if !projection.Includes("message") || !projection.Includes("value01") || projection.Includes("irrelevant") {
		t.Fatalf("unexpected JIT projection keys: %v", projection.Keys())
	}

	newPoint := func() *point.Point {
		kvs := point.NewKVs(map[string]any{
			"message":    messageBuilder.String(),
			"irrelevant": strings.Repeat("not-sent", 512),
			"status":     "warning",
			"value00":    "old-field",
		})
		kvs = kvs.SetTag("value01", "old-tag")
		kvs = kvs.SetTag("irrelevant_tag", "not-sent")
		return point.NewPoint(
			"projected",
			kvs,
			point.WithTime(time.Unix(1_700_000_000, 123)),
		)
	}
	want := newPoint()
	run := pointRun{point: want, script: script, started: time.Now()}
	runPipelineGo(point.Logging, &run, nil)
	if run.output == nil {
		t.Fatal("pipeline-go returned no point")
	}

	got := newPoint()
	input, err := encodeProjectedJITPoints(point.Logging, []*point.Point{got}, projection)
	if err != nil {
		t.Fatalf("encode projected JIT point: %v", err)
	}
	decoded, err := pljit.DecodeFlatPoints(input)
	if err != nil {
		t.Fatalf("decode projected JIT point: %v", err)
	}
	if len(decoded) != 1 || decoded[0].Fields["irrelevant"] != nil || decoded[0].Tags["irrelevant_tag"] != "" {
		t.Fatalf("projected input retained unrelated values: %#v", decoded)
	}
	if decoded[0].Fields["value00"] != "old-field" || decoded[0].Tags["value01"] != "old-tag" {
		t.Fatalf("projected input lost output target state: %#v", decoded[0])
	}
	if decoded[0].Fields["status"] != "warning" {
		t.Fatalf("projected input lost implicit profile field state: %#v", decoded[0])
	}
	batch, err := runner.Process(source, input)
	if err != nil {
		t.Fatalf("run projected JIT: %v", err)
	}
	if len(batch.Records) != 1 || batch.Records[0].Status != pljit.TerminalOK {
		t.Fatalf("unexpected JIT batch: %#v", batch)
	}
	if batch.Static == nil {
		t.Fatal("DataKit profile did not return a static output batch")
	}
	if _, dropped, err := applyJITStatic(point.Logging, got, batch.Static, 0, nil); err != nil || dropped {
		t.Fatalf("apply projected JIT static output: dropped=%v err=%v", dropped, err)
	}
	if equal, reason := got.EqualWithReason(want); !equal {
		t.Fatalf("projected JIT differs from pipeline-go: %s\nJIT: %#v\npipeline-go: %#v", reason, got.KVMap(), want.KVMap())
	}
}

func TestJITStaticGrokMatchesPipelineGo(t *testing.T) {
	runtimePath := os.Getenv("PLATYPUS_JIT_RUNTIME")
	if runtimePath == "" {
		t.Skip("set PLATYPUS_JIT_RUNTIME to test the JIT runtime")
	}
	runner, err := pljit.NewRunner(runtimePath, "pipeline-go-1.4.3-datakit", 8)
	if err != nil {
		t.Fatalf("open JIT runner: %v", err)
	}
	defer runner.Close()

	tests := []struct {
		name   string
		source string
		fields map[string]any
		tags   map[string]string
	}{
		{
			name: "builtin-patterns",
			source: `grok(_, "%{IPORHOST:client_ip} %{WORD:method} %{URIPATHPARAM:path} %{INT:status_code:int}")
`,
			fields: map[string]any{"message": "10.20.30.40 GET /api/items?q=1 200"},
		},
		{
			name: "scoped-pattern-and-trim",
			source: `add_pattern("APP", "%{WORD:service}:%{INT:code:int}")
grok(_, "%{APP} %{GREEDYDATA:msg}", true)
`,
			fields: map[string]any{"message": "api:42   hello world   "},
		},
		{
			name: "no-match-preserves-existing-targets",
			source: `grok(_, "%{WORD:service}:%{INT:code:int}")
`,
			fields: map[string]any{"message": "not a match", "code": int64(7)},
			tags:   map[string]string{"service": "existing"},
		},
		{
			name: "non-string-source",
			source: `grok(value, "%{INT:number:int}")
`,
			fields: map[string]any{"value": int64(123)},
		},
		{
			name: "optional-capture",
			source: `grok(_, "(?:%{WORD:prefix}:)?%{WORD:value}")
`,
			fields: map[string]any{"message": "payload", "prefix": "old"},
		},
		{
			name: "duplicate-output-name",
			source: `grok(_, "%{WORD:value}-%{WORD:value}")
`,
			fields: map[string]any{"message": "first-second"},
		},
		{
			name: "custom-overrides-default",
			source: `add_pattern("WORD", "[0-9]+")
grok(_, "%{WORD:value}")
`,
			fields: map[string]any{"message": "123"},
		},
		{
			name: "typed-capture-conversions",
			source: `grok(_, "%{NOTSPACE:octal:int} %{NOTSPACE:hex:int} %{NOTSPACE:float:float} %{NOTSPACE:flag:bool}")
`,
			fields: map[string]any{"message": "010 0x10 0x1p2 T"},
		},
		{
			name: "mixed-static-mutators",
			source: `grok(_, "%{WORD:service}")
uppercase(service)
rename(app, service)
add_key(env, "prod")
set_tag(env)
drop_key(remove_me)
gjson(payload, "value", "json_value")
parse_duration(duration)
datetime(epoch, "s", "%Y-%m-%d", "UTC")
default_time_with_fmt(parsed_at, "2006-01-02 15:04:05", "UTC")
pt_kvs_del("pt_remove")
pt_kvs_set("copied", app)
pt_kvs_set("literal_tag", "yes", true)
cover(secret, [2, 4])
group_between(score, [10, 20], "mid", score_group)
group_in(level, ["warn", "error"], "problem", level_group)
json_all(json_blob, ["a", "b"])
kv_split(kv_blob, " ", "=", "", "", ["x", "y"], "kv_")
drop_origin_data()
`,
			fields: map[string]any{
				"message": "api", "payload": `{"value":42}`, "remove_me": "gone",
				"duration": "1.5s", "epoch": int64(0),
				"parsed_at": "1970-01-01 00:00:01", "pt_remove": "gone",
				"json_blob": `{"a":1,"b":"two","c":3}`,
				"kv_blob":   "x=one y=two z=three", "level": "warn",
				"score": int64(15), "secret": "abcdef",
			},
		},
		{
			name:   "datetime-invalid-input",
			source: "grok(_, \"%{GREEDYDATA:noop}\")\ndatetime(epoch, \"s\", \"%Y-%m-%d\", \"UTC\")\n",
			fields: map[string]any{"message": "x", "epoch": "invalid"},
		},
		{
			name:   "datetime-rfc3339",
			source: "grok(_, \"%{GREEDYDATA:noop}\")\ndatetime(epoch, \"s\", \"RFC3339\", \"UTC\")\n",
			fields: map[string]any{"message": "x", "epoch": int64(0)},
		},
		{
			name:   "default-time-date-only",
			source: "grok(_, \"%{GREEDYDATA:noop}\")\ndefault_time_with_fmt(parsed_at, \"2006-01-02\", \"UTC\")\n",
			fields: map[string]any{"message": "x", "parsed_at": "1970-01-02"},
		},
		{
			name:   "cover-unicode-han-radical",
			source: "grok(_, \"%{GREEDYDATA:noop}\")\ncover(secret, [1, 1])\n",
			fields: map[string]any{"message": "x", "secret": "⺀A"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			script, err := NewPlScriptSimple(point.Logging, test.name+".p", test.source)
			if err != nil {
				t.Fatalf("create pipeline-go script: %v", err)
			}
			newPoint := func() *point.Point {
				kvs := point.NewKVs(test.fields)
				for key, value := range test.tags {
					kvs = kvs.SetTag(key, value)
				}
				return point.NewPoint(
					"grok",
					kvs,
					point.WithTime(time.Unix(1_700_000_000, 123)),
				)
			}

			want := newPoint()
			run := pointRun{point: want, script: script, started: time.Now()}
			runPipelineGo(point.Logging, &run, nil)
			if run.output == nil {
				t.Fatal("pipeline-go returned no point")
			}

			projection, err := runner.Projection(test.source)
			if err != nil {
				t.Fatalf("query static Grok projection: %v", err)
			}
			got := newPoint()
			input, err := encodeProjectedJITPoints(point.Logging, []*point.Point{got}, projection)
			if err != nil {
				t.Fatalf("encode projected JIT point: %v", err)
			}
			batch, err := runner.Process(test.source, input)
			if err != nil {
				t.Fatalf("run JIT: %v", err)
			}
			if len(batch.Records) != 1 || batch.Records[0].Status != pljit.TerminalOK || batch.Static == nil {
				t.Fatalf("unexpected JIT batch: %#v", batch)
			}
			if _, dropped, err := applyJITStatic(point.Logging, got, batch.Static, 0, nil); err != nil || dropped {
				t.Fatalf("apply static Grok output: dropped=%v err=%v", dropped, err)
			}
			if equal, reason := got.EqualWithReason(run.output); !equal {
				t.Fatalf("JIT Grok differs from pipeline-go: %s\nJIT: %#v\npipeline-go: %#v", reason, got.KVMap(), run.output.KVMap())
			}
		})
	}
}

func TestJITFastCastMatchesPipelineGo(t *testing.T) {
	runtimePath := os.Getenv("PLATYPUS_JIT_RUNTIME")
	if runtimePath == "" {
		t.Skip("set PLATYPUS_JIT_RUNTIME to test the JIT runtime")
	}
	runner, err := pljit.NewRunner(runtimePath, "pipeline-go-1.4.3", 8)
	if err != nil {
		t.Fatalf("open JIT runner: %v", err)
	}
	defer runner.Close()

	tests := []struct {
		name   string
		value  any
		target string
		tag    bool
	}{
		{name: "integer-no-op", value: int64(42), target: "int"},
		{name: "string-to-integer", value: "42", target: "int"},
		{name: "invalid-number-to-zero", value: "invalid", target: "int"},
		{name: "integer-to-float", value: int64(42), target: "float"},
		{name: "boolean-to-string", value: true, target: "str"},
		{name: "string-alias-to-nil", value: true, target: "string"},
		{name: "string-to-boolean", value: "true", target: "bool"},
		{name: "tag-to-boolean", value: "true", target: "bool", tag: true},
		{name: "tag-string-alias-to-empty", value: "value", target: "string", tag: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			source := fmt.Sprintf("cast(value, %q)\n", test.target)
			script, err := NewPlScriptSimple(point.Logging, test.name+".p", source)
			if err != nil {
				t.Fatalf("create script: %v", err)
			}
			newPoint := func() *point.Point {
				kvs := point.NewKVs(map[string]any{"value": test.value})
				if test.tag {
					kvs = point.NewKVs(nil).SetTag("value", fmt.Sprint(test.value))
				}
				return point.NewPoint(
					"cast",
					kvs,
					point.WithTime(time.Unix(1_700_000_000, 123)),
				)
			}

			want := newPoint()
			run := pointRun{point: want, script: script, started: time.Now()}
			runPipelineGo(point.Logging, &run, nil)
			if run.output == nil {
				t.Fatal("pipeline-go returned no point")
			}

			got := newPoint()
			input, err := encodeJITPoints(point.Logging, []*point.Point{got})
			if err != nil {
				t.Fatalf("encode JIT point: %v", err)
			}
			batch, err := runner.Process(source, input)
			if err != nil {
				t.Fatalf("run JIT: %v", err)
			}
			if len(batch.Records) != 1 || batch.Records[0].Status != pljit.TerminalOK {
				t.Fatalf("unexpected JIT batch: %#v", batch)
			}
			if _, dropped, err := applyJITRecord(point.Logging, got, batch.Records[0], 0, nil); err != nil || dropped {
				t.Fatalf("apply JIT delta: dropped=%v err=%v", dropped, err)
			}
			if equal, reason := got.EqualWithReason(run.output); !equal {
				t.Fatalf("JIT differs from pipeline-go: %s\nJIT: %#v\npipeline-go: %#v", reason, got.KVMap(), run.output.KVMap())
			}
		})
	}
}

type fakeJITProcessor struct {
	batch      pljit.Batch
	prepareErr error
	err        error
}

func (f *fakeJITProcessor) Prepare(string) error { return f.prepareErr }

func (f *fakeJITProcessor) Check(string) pljit.CheckResult {
	if f.prepareErr != nil {
		return pljit.CheckResult{Route: pljit.RoutePipelineGo, Reason: pljit.CheckReasonCompileError}
	}
	return pljit.CheckResult{Route: pljit.RouteJITNative, Reason: pljit.CheckReasonJITReady}
}

func (*fakeJITProcessor) Invalidate(string) {}

func (f *fakeJITProcessor) Projection(string) (pljit.InputProjection, error) {
	return pljit.InputProjection{}, f.prepareErr
}

func (f *fakeJITProcessor) MinBatchSize() int { return 0 }

func (f *fakeJITProcessor) Process(string, []byte) (pljit.Batch, error) {
	return f.batch, f.err
}

func (f *fakeJITProcessor) Close() error { return nil }

func TestRunJITGroupUsesIndexedResultsAndRecordFallback(t *testing.T) {
	script, err := NewPlScriptSimple(point.Logging, "test.p", "drop_key(foo)\n")
	if err != nil {
		t.Fatalf("create script: %v", err)
	}
	timestamp := time.Unix(1_700_000_000, 123).UTC()
	jitTimestamp := timestamp.Add(time.Minute)
	first := point.NewPoint("first", point.NewKVs(map[string]any{
		"foo":  "jit",
		"keep": int64(1),
		"time": jitTimestamp.UnixNano(),
	}), point.WithTime(timestamp))
	second := point.NewPoint("second", point.NewKVs(map[string]any{"foo": "fallback", "keep": int64(2)}), point.WithTime(timestamp))
	runs := []pointRun{
		{point: first, script: script, started: time.Now()},
		{point: second, script: script, started: time.Now()},
	}
	delta, err := pljit.EncodeMutationDeltas([]pljit.MutationDelta{{
		RecordIndex: 0,
		Operations: []pljit.MutationOp{
			{Kind: pljit.MutationSetMeasurement, String: "jit-first"},
			{Kind: pljit.MutationDeleteField, Key: "foo"},
			{Kind: pljit.MutationSetTag, Key: "source", String: "jit"},
		},
	}})
	if err != nil {
		t.Fatalf("encode mutation delta: %v", err)
	}

	runner := &fakeJITProcessor{batch: pljit.Batch{Records: []pljit.Record{
		{
			Status: pljit.TerminalOK,
			Delta:  delta,
		},
		{Status: pljit.TerminalError, Error: []byte(`{"code":"E_TEST"}`)},
	}}}
	runJITGroup(runner, point.Logging, script, []int{0, 1}, runs, &lang.LogOption{})
	if runs[0].output == nil || runs[0].output.Name() != "jit-first" || runs[0].output.GetTag("source") != "jit" {
		t.Fatalf("first JIT output was not applied: %#v", runs[0].output)
	}
	if runs[0].output.Get("foo") != nil || runs[0].output.Get("keep") != int64(1) {
		t.Fatalf("unexpected first fields: %#v", runs[0].output.KVMap())
	}
	if runs[0].output.Get("time") != nil || !runs[0].output.Time().Equal(jitTimestamp) {
		t.Fatalf("JIT output did not finalize the time field: %#v", runs[0].output)
	}
	if runs[1].output == nil || runs[1].output.Get("foo") == nil || runs[1].output.Get("keep") != int64(2) || runs[1].output.Get(plStatus) != sFailed {
		t.Fatalf("second record was replayed or unreported: %#v", runs[1].output)
	}
}

func TestRunJITGroupReportsFailedWholeBatchWithoutReplay(t *testing.T) {
	script, err := NewPlScriptSimple(point.Logging, "test.p", "drop_key(foo)\n")
	if err != nil {
		t.Fatalf("create script: %v", err)
	}
	pt := point.NewPoint("test", point.NewKVs(map[string]any{"foo": "remove", "keep": int64(1)}))
	runs := []pointRun{{point: pt, script: script, started: time.Now()}}
	runner := &fakeJITProcessor{err: errors.New("native failure")}
	runJITGroup(runner, point.Logging, script, []int{0}, runs, nil)
	if runs[0].output == nil || runs[0].output.Get("foo") != "remove" || runs[0].output.Get("keep") != int64(1) || runs[0].output.Get(plStatus) != sFailed {
		t.Fatalf("whole-batch failure was replayed or not reported: %#v", runs[0].output)
	}
}

func TestRunJITGroupProjectionFailureKeepsEngineAndSkipsSubmission(t *testing.T) {
	script, err := NewPlScriptSimple(point.Logging, "test.p", "drop_key(foo)\n")
	if err != nil {
		t.Fatalf("create script: %v", err)
	}
	pt := point.NewPoint("test", point.NewKVs(map[string]any{"foo": "remove", "keep": int64(1)}))
	runs := []pointRun{{point: pt, script: script, started: time.Now()}}
	runner := &countingJITProcessor{fakeJITProcessor: fakeJITProcessor{prepareErr: errors.New("projection unavailable")}}
	runJITGroup(runner, point.Logging, script, []int{0}, runs, nil)
	if runner.processes != 0 || runs[0].output == nil || runs[0].output.Get("foo") != "remove" || runs[0].output.Get("keep") != int64(1) || runs[0].output.Get(plStatus) != sFailed {
		t.Fatalf("projection failure changed engine or submitted input: %#v", runs[0].output)
	}
}

func TestApplyJITDeltaPreservesRawString(t *testing.T) {
	for _, raw := range []string{"", "测\x00试", "\xff\x00", "\xe2\x82", "\xb2\xe2\xca\xd4"} {
		t.Run(fmt.Sprintf("%x", raw), func(t *testing.T) {
			pt := point.NewPoint("raw", point.NewKVs(map[string]any{"copied": "old"}))
			encoded, err := pljit.EncodeMutationDeltas([]pljit.MutationDelta{{RecordIndex: 0, Operations: []pljit.MutationOp{
				{Kind: pljit.MutationSetField, Key: "copied", Value: pljit.RawString(raw)},
			}}})
			if err != nil {
				t.Fatal(err)
			}
			if _, _, err := applyJITDelta(point.Logging, pt, encoded, 0, nil); err != nil {
				t.Fatal(err)
			}
			if got, ok := pt.Get("copied").(string); !ok || got != raw {
				t.Fatalf("got %#v, want bytes %x", pt.Get("copied"), raw)
			}
		})
	}
}

func TestApplyJITDeltaIsAtomicOnValidationError(t *testing.T) {
	pt := point.NewPoint("test", point.NewKVs(map[string]any{"foo": "keep"}))
	encoded, err := pljit.EncodeMutationDeltas([]pljit.MutationDelta{{
		RecordIndex: 0,
		Operations: []pljit.MutationOp{
			{Kind: pljit.MutationDeleteField, Key: "foo"},
			{Kind: pljit.MutationSetCategory, String: "metric"},
		},
	}})
	if err != nil {
		t.Fatalf("encode mutation delta: %v", err)
	}
	if _, _, err := applyJITDelta(point.Logging, pt, encoded, 0, nil); err == nil {
		t.Fatal("category change should be rejected")
	}
	if got := pt.Get("foo"); got != "keep" {
		t.Fatalf("failed delta partially mutated point: %v", got)
	}
}

func TestApplyJITStaticIsAtomicOnValidationError(t *testing.T) {
	pt := point.NewPoint("test", point.NewKVs(map[string]any{"first": "keep", "second": "tag"}))
	batch := &pljit.StaticBatch{
		Schema:       pljit.StaticOutputSchema{Keys: []string{"first", "second"}},
		States:       []pljit.StaticMutationState{pljit.StaticSetField, pljit.StaticSetTag},
		StateOffsets: []uint32{0, 2},
		ValueOffsets: []uint32{0, 2},
		Values:       []any{"changed", int64(42)},
	}
	if _, _, err := applyJITStatic(point.Logging, pt, batch, 0, nil); err == nil {
		t.Fatal("non-string static tag should be rejected")
	}
	if got := pt.Get("first"); got != "keep" {
		t.Fatalf("failed static result partially mutated point: %v", got)
	}
}
