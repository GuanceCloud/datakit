// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2026-present Guance, Inc.

package pipeline

import (
	"strings"
	"testing"

	"github.com/GuanceCloud/cliutils/point"
	pljit "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline/jit"
)

type chunkRecordingJITProcessor struct {
	fakeJITProcessor
	counts    []int
	bytes     []int
	sequences []int64
}

func (processor *chunkRecordingJITProcessor) Process(_ string, input []byte) (pljit.Batch, error) {
	points, err := pljit.DecodeFlatPoints(input)
	if err != nil {
		return pljit.Batch{}, err
	}
	processor.counts = append(processor.counts, len(points))
	processor.bytes = append(processor.bytes, len(input))
	for _, pt := range points {
		if sequence, ok := pt.Fields["sequence"].(int64); ok {
			processor.sequences = append(processor.sequences, sequence)
		}
	}
	records := make([]pljit.Record, len(points))
	for index := range records {
		records[index].Status = pljit.TerminalDropped
	}
	return pljit.Batch{Records: records}, nil
}

func TestRunJITGroupChunksLargeRecordCount(t *testing.T) {
	script, err := NewPlScriptSimple(point.Logging, "chunk-count.p", "drop()\n")
	if err != nil {
		t.Fatal(err)
	}
	runs := make([]pointRun, jitChunkMaxRecords+17)
	indexes := make([]int, len(runs))
	for index := range runs {
		indexes[index] = index
		runs[index] = pointRun{
			point:  point.NewPoint("chunk-count", point.NewKVs(map[string]any{"message": "ok"})),
			script: script,
		}
	}
	processor := &chunkRecordingJITProcessor{}
	submitted := jitRouteRecordsVec.WithLabelValues(point.Logging.String(), "attempted", "submitted")
	dropped := jitRouteRecordsVec.WithLabelValues(point.Logging.String(), "native", "dropped")
	beforeSubmitted, beforeDropped := readPrometheusCounter(t, submitted), readPrometheusCounter(t, dropped)
	runJITGroup(processor, point.Logging, script, indexes, runs, nil)
	if got := readPrometheusCounter(t, submitted) - beforeSubmitted; got != float64(len(runs)) {
		t.Fatalf("submitted count across chunks = %v, want %d", got, len(runs))
	}
	if got := readPrometheusCounter(t, dropped) - beforeDropped; got != float64(len(runs)) {
		t.Fatalf("dropped count across chunks = %v, want %d", got, len(runs))
	}
	if len(processor.counts) != 2 || processor.counts[0] != jitChunkMaxRecords || processor.counts[1] != 17 {
		t.Fatalf("JIT chunk record counts = %v", processor.counts)
	}
	for index := range runs {
		if !runs[index].dropped {
			t.Fatalf("record %d was not applied", index)
		}
	}
}

func TestRunJITGroupIsolatesBadEncodingWithoutLosingValidNeighbors(t *testing.T) {
	script, err := NewPlScriptSimple(point.Logging, "isolated.p", "drop()\n")
	if err != nil {
		t.Fatal(err)
	}
	for _, bad := range []int{0, 3, 8} {
		runs := make([]pointRun, 9)
		indexes := make([]int, len(runs))
		for i := range runs {
			message := "valid"
			if i == bad {
				message = string([]byte{0xff})
			}
			runs[i] = pointRun{point: point.NewPoint("isolated", point.NewKVs(map[string]any{"message": message, "sequence": int64(i)})), script: script}
			indexes[i] = i
		}
		processor := &chunkRecordingJITProcessor{}
		runJITGroup(processor, point.Logging, script, indexes, runs, nil)
		expectedSequence := 0
		for _, sequence := range processor.sequences {
			if expectedSequence == bad {
				expectedSequence++
			}
			if sequence != int64(expectedSequence) {
				t.Fatalf("bad=%d submission order/duplication: %v", bad, processor.sequences)
			}
			expectedSequence++
		}
		if len(processor.sequences) != 8 {
			t.Fatalf("bad=%d submitted %v", bad, processor.sequences)
		}
		for i := range runs {
			if i == bad {
				if runs[i].dropped || runs[i].output == nil || runs[i].output.Get(plStatus) != sFailed {
					t.Fatalf("bad record %d replayed or unreported", i)
				}
			} else if !runs[i].dropped {
				t.Fatalf("valid neighbor %d lost for bad=%d", i, bad)
			}
		}
	}
}

func TestRunJITGroupChunksLargeEncodedInput(t *testing.T) {
	script, err := NewPlScriptSimple(point.Logging, "chunk-bytes.p", "drop()\n")
	if err != nil {
		t.Fatal(err)
	}
	const records = 12
	value := strings.Repeat("x", 1<<20)
	runs := make([]pointRun, records)
	indexes := make([]int, records)
	for index := range runs {
		indexes[index] = index
		runs[index] = pointRun{
			point:  point.NewPoint("chunk-bytes", point.NewKVs(map[string]any{"message": value})),
			script: script,
		}
	}
	processor := &chunkRecordingJITProcessor{}
	runJITGroup(processor, point.Logging, script, indexes, runs, nil)
	if len(processor.bytes) < 2 {
		t.Fatalf("large encoded input was not chunked: %v", processor.bytes)
	}
	for _, inputBytes := range processor.bytes {
		if inputBytes > jitChunkTargetBytes {
			t.Fatalf("JIT input chunk = %d bytes, target <= %d", inputBytes, jitChunkTargetBytes)
		}
	}
}
