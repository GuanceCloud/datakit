// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package pipeline

import (
	"testing"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"github.com/GuanceCloud/pipeline-go/ptinput"
	pljit "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline/jit"
)

func TestPointFromJITMatchesCategoryConstruction(t *testing.T) {
	for _, category := range []point.Category{point.Logging, point.Metric, point.Tracing, point.RUM, point.Network, point.Object, point.CustomObject, point.Security, point.DialTesting} {
		t.Run(category.String(), func(t *testing.T) {
			fields := map[string]any{"message": "hello", "text": "value", "count": int64(3), "dotted.key": true}
			tags := map[string]string{"source": "diff"}
			timestamp := time.Unix(0, 123)
			want := ptinput.NewPlPt(category, "child", tags, fields, timestamp).Point()
			got, err := pointFromJIT(&pljit.Point{Category: category.String(), Measurement: "child", Tags: tags, Fields: fields, TimeUnixNano: 123})
			if err != nil {
				t.Fatal(err)
			}
			if equal, reason := got.EqualWithReason(want); !equal {
				t.Fatalf("category construction differs: %s", reason)
			}
		})
	}
}

func TestRunJITGroupDroppedParentPreservesEmittedPoints(t *testing.T) {
	script, err := NewPlScriptSimple(point.Logging, "window.p", "drop()")
	if err != nil {
		t.Fatal(err)
	}
	pt := point.NewPoint("parent", point.NewKVs(map[string]any{"message": "parent"}))
	runs := []pointRun{{point: pt, script: script, started: time.Now()}}
	runner := &fakeJITProcessor{batch: pljit.Batch{Records: []pljit.Record{{
		Status:  pljit.TerminalDropped,
		Emitted: [][]byte{[]byte(`{"version":1,"category":"logging","measurement":"restored","fields":{"sequence":9007199254740993},"time":"2026-08-31T00:00:00Z"}`)},
	}}}}
	runJITGroup(runner, point.Logging, script, []int{0}, runs, nil)
	if !runs[0].dropped || len(runs[0].created[point.Logging]) != 1 {
		t.Fatalf("dropped parent lost recovered points: %#v", runs[0])
	}
	restored := runs[0].created[point.Logging][0]
	if got := restored.Get("sequence"); got != int64(9007199254740993) {
		t.Fatalf("emitted integer lost precision or type: %T(%v)", got, got)
	}
}

func TestRunJITGroupMalformedEmittedPointDoesNotMutateOrReplay(t *testing.T) {
	script, err := NewPlScriptSimple(point.Logging, "no-replay.p", "rename(bar, foo)")
	if err != nil {
		t.Fatal(err)
	}
	delta, err := pljit.EncodeMutationDeltas([]pljit.MutationDelta{{RecordIndex: 0, Operations: []pljit.MutationOp{
		{Kind: pljit.MutationDeleteField, Key: "foo"},
	}}})
	if err != nil {
		t.Fatal(err)
	}
	pt := point.NewPoint("original", point.NewKVs(map[string]any{"foo": "original"}))
	runs := []pointRun{{point: pt, script: script, started: time.Now()}}
	runner := &fakeJITProcessor{batch: pljit.Batch{Records: []pljit.Record{{
		Status: pljit.TerminalOK, Delta: delta, Emitted: [][]byte{[]byte(`{"version":1`)},
	}}}}
	runJITGroup(runner, point.Logging, script, []int{0}, runs, nil)
	if pt.Get("foo") != "original" || pt.Get("bar") != nil || runs[0].output != pt || len(runs[0].created) != 0 {
		t.Fatalf("invalid side output applied mutations or replayed Go: %#v", pt.KVMap())
	}
	if pt.Get(plStatus) != sFailed || pt.Get("pl_msg") == nil {
		t.Fatalf("invalid side output was not reported as failure: %#v", pt.KVMap())
	}
}

func TestDecodeJITEmittedRejectsInvalidEnvelope(t *testing.T) {
	for _, payload := range []string{
		`{"version":2,"category":"logging"}`,
		`{"version":1,"category":"logging"} {}`,
		`{"version":1,"category":"invalid-category"}`,
	} {
		if _, err := decodeJITEmitted([][]byte{[]byte(payload)}); err == nil {
			t.Fatalf("accepted invalid emitted point: %s", payload)
		}
	}
}

func TestDecodeJITEmittedBinaryEnvelope(t *testing.T) {
	payload, err := pljit.EncodeFlatPoints([]pljit.Point{{
		Version: 1, Category: "metric", Measurement: "binary-child",
		Tags: map[string]string{"source": "binary"}, Fields: map[string]any{"count": int64(3)},
	}})
	if err != nil {
		t.Fatal(err)
	}
	created, err := decodeJITEmitted([][]byte{payload})
	if err != nil {
		t.Fatal(err)
	}
	points := created[point.Metric]
	if len(points) != 1 || points[0].GetTag("source") != "binary" || points[0].Get("count") != int64(3) {
		if len(points) == 1 {
			t.Fatalf("binary emitted point changed: kv=%#v source=%q count=%T(%v)", points[0].KVs(), points[0].GetTag("source"), points[0].Get("count"), points[0].Get("count"))
		}
		t.Fatalf("binary emitted point count=%d", len(points))
	}
}
