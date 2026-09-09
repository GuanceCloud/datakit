// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build pipeline_jit && linux && (amd64 || arm64)

// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2026-present Guance, Inc.

package jit

import (
	"fmt"
	"os"
	"strconv"
	"testing"

	"github.com/GuanceCloud/pipeline-go/ptinput/funcs"
)

var timestampBenchmarkBatch Batch

func TestNativeTimestampCorpusEndToEnd(t *testing.T) {
	runtimePath := os.Getenv("PLATYPUS_JIT_RUNTIME")
	if runtimePath == "" {
		t.Skip("set PLATYPUS_JIT_RUNTIME to a trusted production platypus_jit cdylib")
	}
	host, recorder := newObservedPipelineGoHost()
	runner, err := NewRunnerWithHost(runtimePath, "pipeline-go-1.4.3-datakit", 16, host)
	if err != nil {
		t.Fatalf("open runner with compatibility host: %v", err)
	}
	defer runner.Close()

	for index, rawVector := range dateparseTimestampTestVectors {
		vector := normalizedDateparseTimestampVector(rawVector)
		t.Run(fmt.Sprintf("dateparse-%03d", index), func(t *testing.T) {
			source := "default_time(time, " + strconv.Quote(vector.timezone) + ")\n"
			compareTimestampBatchWithPipelineGo(t, runner, recorder, source, []timestampTestVector{vector})
		})
	}
	for index, vector := range pipelineGoTimezoneTestVectors {
		t.Run(fmt.Sprintf("timezone-%02d", index), func(t *testing.T) {
			source := "default_time(time, " + strconv.Quote(vector.timezone) + ")\n"
			compareTimestampBatchWithPipelineGo(t, runner, recorder, source, []timestampTestVector{vector})
		})
	}

	// Generate malformed grammar families instead of pinning only the two
	// production samples that originally exposed dateparse diagnostics.
	var malformed []timestampTestVector
	for _, month := range []string{"January", "march", "September"} {
		for _, suffix := range []string{"RR", "XYZ", "Bad"} {
			malformed = append(malformed, timestampTestVector{
				value:    month + suffix + " 7th, 1970",
				timezone: "Etc/UTC",
			})
		}
	}
	for _, month := range []string{"Feb", "Oct", "DEC"} {
		malformed = append(malformed, timestampTestVector{
			value:    "11-" + month + "-2012::12:53:54",
			timezone: "Etc/UTC",
		})
	}
	t.Run("generated-malformed-grammar", func(t *testing.T) {
		compareTimestampBatchWithPipelineGo(
			t, runner, recorder, `default_time(time, "Etc/UTC")`+"\n", malformed,
		)
	})
}

// TestTimestampHandle, pipeline-go v1.4.3 ptinput/funcs/handle_test.go.
func TestNativeTimestampHandleUpstreamOracle(t *testing.T) {
	runtimePath := os.Getenv("PLATYPUS_JIT_RUNTIME")
	if runtimePath == "" {
		t.Skip("set PLATYPUS_JIT_RUNTIME to a trusted production platypus_jit cdylib")
	}
	host, recorder := newObservedPipelineGoHost()
	runner, err := NewRunnerWithHost(runtimePath, "pipeline-go-1.4.3-datakit", 4, host)
	if err != nil {
		t.Fatalf("open runner with compatibility host: %v", err)
	}
	defer runner.Close()

	vectors := make([]timestampTestVector, 0, 10)
	for year := 1970; year < 1980; year++ {
		vectors = append(vectors, timestampTestVector{
			value:    fmt.Sprintf("Thu Jan 16 10:05:19 %d", year),
			timezone: "+0",
		})
	}
	compareTimestampBatchWithPipelineGo(
		t,
		runner,
		recorder,
		`default_time(time, "+0")`+"\n",
		vectors,
	)
}

func compareTimestampBatchWithPipelineGo(t *testing.T, runner *Runner, recorder *hostCompatCallRecorder, source string, vectors []timestampTestVector) {
	t.Helper()
	recorder.reset()
	requireDefaultTimeNativeCapabilities(t, runner, source)
	const basePointTime = int64(1_700_000_000_000_000_000)
	points := make([]Point, len(vectors))
	expected := make([]int64, len(vectors))
	expectedErrors := make([]error, len(vectors))
	for index, vector := range vectors {
		pointTime := basePointTime + int64(index)
		points[index] = Point{
			Version:      1,
			Category:     "logging",
			Measurement:  "timestamp-differential",
			TimeUnixNano: pointTime,
			Fields:       map[string]any{"time": vector.value},
		}
		expected[index], expectedErrors[index] = funcs.TimestampHandle(vector.value, vector.timezone)
	}
	input, err := EncodeFlatPoints(points)
	if err != nil {
		t.Fatalf("encode timestamp corpus: %v", err)
	}
	batch, err := runner.Process(source, input)
	if err != nil {
		t.Fatalf("process timestamp corpus: %v", err)
	}
	if len(batch.Records) != len(points) {
		t.Fatalf("unexpected timestamp corpus batch: records=%d want=%d", len(batch.Records), len(points))
	}
	requireDefaultTimeStaticBatch(t, batch)
	for index := range points {
		record := batch.Records[index]
		if record.Status != TerminalOK {
			t.Fatalf("record %d value=%q returned terminal %d: %s", index, vectors[index].value, record.Status, record.Error)
		}
		applyHostCompatRecord(t, batch, index, &points[index])
		actual := points[index].Fields["time"]
		if expectedErrors[index] != nil {
			wantMessage := fmt.Sprintf("time convert failed: %v", expectedErrors[index])
			if actual != points[index].TimeUnixNano || points[index].Fields["pl_msg"] != wantMessage {
				t.Errorf("record %d value=%q failure fields=%#v, want point time %d and pl_msg %q",
					index, vectors[index].value, points[index].Fields, points[index].TimeUnixNano, wantMessage)
			}
			continue
		}
		if actual != expected[index] {
			t.Errorf("record %d value=%q timezone=%q time=%#v, want pipeline-go %d",
				index, vectors[index].value, vectors[index].timezone, actual, expected[index])
		}
		if _, ok := points[index].Fields["pl_msg"]; ok {
			t.Errorf("record %d successful parse set pl_msg: %#v", index, points[index].Fields)
		}
	}
	requireHostCompatCalls(t, recorder, hostCompatTimestampOp)
}

func BenchmarkNativeDefaultTimeEndToEnd(b *testing.B) {
	runtimePath := os.Getenv("PLATYPUS_JIT_RUNTIME")
	if runtimePath == "" {
		b.Skip("set PLATYPUS_JIT_RUNTIME to a trusted production platypus_jit cdylib")
	}
	runner, err := NewRunnerWithHost(runtimePath, "pipeline-go-1.4.3-datakit", 8, NewPipelineGoHost(nil))
	if err != nil {
		b.Fatalf("open runner with compatibility host: %v", err)
	}
	defer runner.Close()

	const source = "default_time(time)\n"
	if err := runner.Prepare(source); err != nil {
		b.Fatalf("prepare timestamp benchmark: %v", err)
	}
	for _, benchmark := range []struct {
		name  string
		value string
	}{
		{name: "native-explicit-offset", value: "2026-05-19T13:47:01.004+0800"},
		{name: "native-dateparse", value: "oct 7, 1970"},
		{name: "native-offset-with-zone-label", value: "2012-08-03 18:31:59.000+00:00 PST"},
	} {
		b.Run(benchmark.name, func(b *testing.B) {
			const batchSize = 128
			points := make([]Point, batchSize)
			for index := range points {
				points[index] = Point{
					Version:      1,
					Category:     "logging",
					Measurement:  "timestamp-benchmark",
					TimeUnixNano: 1_700_000_000_000_000_000 + int64(index),
					Fields:       map[string]any{"time": benchmark.value},
				}
			}
			input, err := EncodeFlatPoints(points)
			if err != nil {
				b.Fatalf("encode timestamp benchmark batch: %v", err)
			}
			if _, err := runner.Process(source, input); err != nil {
				b.Fatalf("warm timestamp benchmark: %v", err)
			}
			b.ReportAllocs()
			b.SetBytes(int64(len(input)))
			b.ResetTimer()
			for range b.N {
				timestampBenchmarkBatch, timestampBenchmarkError = runner.Process(source, input)
				if timestampBenchmarkError != nil {
					b.Fatal(timestampBenchmarkError)
				}
			}
			b.ReportMetric(float64(b.N*batchSize)/b.Elapsed().Seconds(), "records/s")
		})
	}
}
