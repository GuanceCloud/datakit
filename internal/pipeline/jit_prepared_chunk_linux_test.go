// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build pipeline_jit && linux && (amd64 || arm64)

// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// Copyright 2026-present Guance, Inc.

package pipeline

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/GuanceCloud/cliutils/point"
	pljit "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline/jit"
)

func preparedChunkProjection(t *testing.T) pljit.InputProjection {
	t.Helper()
	runner, err := pljit.NewRunner(os.Getenv("PLATYPUS_JIT_RUNTIME"), "pipeline-go-1.4.3-datakit", 4)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := runner.Close(); err != nil {
			t.Error(err)
		}
	})
	projection, err := runner.Projection("cast(message, \"str\")\ncast(sequence, \"int\")\ncast(values, \"str\")")
	if err != nil {
		t.Fatal(err)
	}
	if projection.IsAll() {
		t.Fatal("expected a proven projection")
	}
	return projection
}

func TestPreparedJITChunkMatchesExistingBoundariesAndWire(t *testing.T) {
	projection := preparedChunkProjection(t)
	makePoint := func(message any) *point.Point {
		return point.NewPoint("prepared", point.NewKVs(map[string]any{
			"message": message, "sequence": int64(7), "values": []int64{1, 2}, "unread": "keep",
		}))
	}
	tagged := makePoint("tagged")
	tagged.SetTag("sequence", "17")
	raw := makePoint("raw\xfftext")
	large := makePoint(strings.Repeat("x", 3<<20))
	cases := [][]*point.Point{
		{makePoint("normal"), tagged, raw},
		{makePoint("first"), nil, makePoint("last")},
		{large, large, large, makePoint("tail")},
		{makePoint([]byte(strings.Repeat("x", 5<<20))), makePoint("tail")},
	}
	var prepared, ordinary projectedJITEncoder
	for _, points := range cases {
		runs, indexes := make([]pointRun, len(points)), make([]int, len(points))
		for i, pt := range points {
			runs[i].point, indexes[i] = pt, i
		}
		for start := 0; start < len(points); {
			end := prepared.prepareChunk(point.Logging, indexes, runs, projection, start)
			if want := nextJITChunkEnd(point.Logging, indexes, runs, projection, start); end != want {
				t.Fatalf("boundary=%d, want %d", end, want)
			}
			// The production retry loop repeatedly encodes shrinking prefixes.
			for count := end - start; count >= 1; count /= 2 {
				got, gotErr := prepared.encodePrepared(count, projection)
				want, wantErr := ordinary.encode(point.Logging, points[start:start+count], projection)
				if (gotErr == nil) != (wantErr == nil) {
					t.Fatalf("error mismatch: %v / %v", gotErr, wantErr)
				}
				if gotErr != nil {
					if gotErr.Error() != wantErr.Error() {
						t.Fatalf("errors: %v / %v", gotErr, wantErr)
					}
				} else if !bytes.Equal(got, want) {
					t.Fatal("prepared wire differs")
				}
			}
			start = end
		}
	}
}

type projectedChunkRecorder struct {
	chunkRecordingJITProcessor
	projection pljit.InputProjection
}

func (p *projectedChunkRecorder) Projection(string) (pljit.InputProjection, error) {
	return p.projection, nil
}

func TestPreparedJITGroupSplitsLargeBytesAndIsolatesBadField(t *testing.T) {
	projection := preparedChunkProjection(t)
	script, err := NewPlScriptSimple(point.Logging, "prepared.p", "drop()")
	if err != nil {
		t.Fatal(err)
	}
	for _, bad := range []int{-1, 6} {
		for _, large := range []bool{false, true} {
			runs, indexes := make([]pointRun, 13), make([]int, 13)
			for i := range runs {
				indexes[i] = i
				var message any = "valid"
				if large {
					message = []byte(strings.Repeat("x", 1<<20))
				}
				runs[i] = pointRun{point: point.NewPoint("prepared", point.NewKVs(map[string]any{
					"message": message, "sequence": int64(i),
				})), script: script}
				if i == bad {
					for _, kv := range runs[i].point.KVs() {
						if kv.Key == "message" {
							kv.Val = nil
						}
					}
				}
			}
			processor := &projectedChunkRecorder{projection: projection}
			runJITGroup(processor, point.Logging, script, indexes, runs, nil)
			expected := int64(0)
			for _, sequence := range processor.sequences {
				if expected == int64(bad) {
					expected++
				}
				if sequence != expected {
					t.Fatalf("large=%v: replay/order mismatch: %v", large, processor.sequences)
				}
				expected++
			}
			wanted := 13
			if bad >= 0 {
				wanted--
			}
			if len(processor.sequences) != wanted {
				t.Fatalf("lost neighbors: %v", processor.sequences)
			}
			for _, size := range processor.bytes {
				if size > jitChunkTargetBytes {
					t.Fatalf("submitted oversized chunk: %d", size)
				}
			}
			for i := range runs {
				if i != bad && !runs[i].dropped {
					t.Fatalf("valid record %d was not applied", i)
				}
			}
		}
	}
}

func TestPreparedJITChunkDefersCompositesAndClearsDiscardedViews(t *testing.T) {
	projection := preparedChunkProjection(t)
	points := make([]*point.Point, 3)
	runs, indexes := make([]pointRun, 3), []int{0, 1, 2}
	for i := range points {
		kvs := make(point.KVs, 24)
		for j := range kvs {
			kvs[j] = point.NewKV(string(rune('a'+j)), []int64{int64(i), int64(j)})
		}
		points[i] = point.NewPoint("growth", kvs)
		// Repeated selected keys exceed the projection's capacity hint and force
		// entries to grow while earlier records retain slices of older arrays.
		for _, kv := range points[i].KVs() {
			kv.Key = "values"
		}
		runs[i].point = points[i]
	}
	var encoder projectedJITEncoder
	if end := encoder.prepareChunk(point.Logging, indexes, runs, projection, 0); end != 3 {
		t.Fatal(end)
	}
	for _, record := range encoder.records {
		for _, entry := range record.Entries {
			if _, ok := entry.Value.(*point.Field); !ok {
				t.Fatal("composite decoded during estimation")
			}
		}
	}
	if _, err := encoder.encodePrepared(3, projection); err != nil {
		t.Fatal(err)
	}
	discarded := append([]pljit.FlatPointRecord(nil), encoder.records[1:]...)
	got, err := encoder.encodePrepared(1, projection)
	if err != nil {
		t.Fatal(err)
	}
	var ordinary projectedJITEncoder
	want, err := ordinary.encode(point.Logging, points[:1], projection)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want) {
		t.Fatal("growth/retry changed prefix")
	}
	for _, record := range discarded {
		for _, entry := range record.Entries {
			if entry.Key != "" || entry.Value != nil {
				t.Fatal("discarded view retains a payload")
			}
		}
	}
}
