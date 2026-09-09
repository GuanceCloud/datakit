// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package pipeline

import (
	"errors"
	"testing"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"github.com/GuanceCloud/pipeline-go/lang"
	plstats "github.com/GuanceCloud/pipeline-go/stats"
	"github.com/prometheus/client_golang/prometheus"
	pljit "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline/jit"
)

type committedErrorStats struct {
	points float64
	drops  float64
	errors float64
}

type countingJITProcessor struct {
	fakeJITProcessor
	processes int
}

func (processor *countingJITProcessor) Process(source string, input []byte) (pljit.Batch, error) {
	processor.processes++
	return processor.fakeJITProcessor.Process(source, input)
}

func TestRunJITGroupEncodeFailureDoesNotSwitchEngine(t *testing.T) {
	script, err := NewPlScriptSimple(point.Logging, "bad-utf8.p", "rename(bar, foo)\n")
	if err != nil {
		t.Fatal(err)
	}
	pt := point.NewPoint("bad-utf8", point.NewKVs(map[string]any{"foo": "original", "message": string([]byte{0xff})}))
	runs := []pointRun{{point: pt, script: script}}
	runner := &countingJITProcessor{}
	runJITGroup(runner, point.Logging, script, []int{0}, runs, nil)
	if runner.processes != 0 || pt.Get("foo") != "original" || pt.Get("bar") != nil || pt.Get(plStatus) != sFailed {
		t.Fatalf("encoding failure submitted or changed engine: calls=%d point=%#v", runner.processes, pt.KVMap())
	}
}

type highThresholdProcessor struct{ countingJITProcessor }

func (*highThresholdProcessor) MinBatchSize() int { return 128 }

func TestRunJITGroupSmallBatchKeepsSelectedNativeRoute(t *testing.T) {
	script, err := NewPlScriptSimple(point.Logging, "small-native.p", "add_key(engine, \"go\")\n")
	if err != nil {
		t.Fatal(err)
	}
	delta, err := pljit.EncodeMutationDeltas([]pljit.MutationDelta{{RecordIndex: 0, Operations: []pljit.MutationOp{
		{Kind: pljit.MutationSetField, Key: "engine", Value: "native"},
	}}})
	if err != nil {
		t.Fatal(err)
	}
	runner := &highThresholdProcessor{countingJITProcessor{fakeJITProcessor: fakeJITProcessor{batch: pljit.Batch{
		Records: []pljit.Record{{Status: pljit.TerminalOK, Delta: delta}},
	}}}}
	runs := []pointRun{{point: point.NewPoint("small", point.NewKVs(map[string]any{"message": "x"})), script: script}}
	runJITGroup(runner, point.Logging, script, []int{0}, runs, nil)
	if runner.processes != 1 || runs[0].output == nil || runs[0].output.Get("engine") != "native" {
		t.Fatalf("small batch changed engine: calls=%d output=%#v", runner.processes, runs[0].output)
	}
}

func TestRunJITGroupCancelledAndInvalidTerminalsDoNotReplay(t *testing.T) {
	script, err := NewPlScriptSimple(point.Logging, "terminal-no-replay.p", "rename(bar, foo)\n")
	if err != nil {
		t.Fatal(err)
	}
	statuses := []pljit.TerminalStatus{pljit.TerminalCancelled, pljit.TerminalStatus(255), pljit.TerminalOK}
	runs := make([]pointRun, len(statuses))
	records := make([]pljit.Record, len(statuses))
	for i, status := range statuses {
		pt := point.NewPoint("original", point.NewKVs(map[string]any{"foo": "retained"}))
		runs[i] = pointRun{point: pt, script: script}
		delta, err := pljit.EncodeMutationDeltas([]pljit.MutationDelta{{RecordIndex: uint64(i), Operations: []pljit.MutationOp{
			{Kind: pljit.MutationSetField, Key: "native", Value: "applied"},
		}}})
		if err != nil {
			t.Fatal(err)
		}
		records[i] = pljit.Record{Status: status, Delta: delta, Error: []byte("interrupted")}
	}
	runner := &countingJITProcessor{fakeJITProcessor: fakeJITProcessor{batch: pljit.Batch{Records: records}}}
	runJITGroup(runner, point.Logging, script, []int{0, 1, 2}, runs, nil)
	if runner.processes != 1 {
		t.Fatalf("process count=%d", runner.processes)
	}
	for i := 0; i < 2; i++ {
		pt := runs[i].output
		if pt == nil || pt.Get("foo") != "retained" || pt.Get("bar") != nil || pt.Get("native") != nil || pt.Get(plStatus) != sFailed {
			t.Fatalf("terminal %v applied or replayed output: %#v", statuses[i], pt)
		}
	}
	if runs[2].output == nil || runs[2].output.Get("native") != "applied" || runs[2].output.Get(plStatus) == sFailed {
		t.Fatalf("valid neighbor failed: %#v", runs[2].output)
	}
}

func TestRunJITGroupStaticApplyFailureDoesNotMutateOrReplay(t *testing.T) {
	for _, status := range []pljit.TerminalStatus{pljit.TerminalOK, pljit.TerminalError} {
		script, err := NewPlScriptSimple(point.Logging, "static-no-replay.p", "rename(bar, foo)\n")
		if err != nil {
			t.Fatal(err)
		}
		pt := point.NewPoint("original", point.NewKVs(map[string]any{"foo": "value"}))
		runs := []pointRun{{point: pt, script: script}}
		runner := &fakeJITProcessor{batch: pljit.Batch{
			Records: []pljit.Record{{Status: status, CommitPrefixError: status == pljit.TerminalError}},
			Static: &pljit.StaticBatch{
				Schema:       pljit.StaticOutputSchema{Keys: []string{"foo", "bad_tag"}},
				States:       []pljit.StaticMutationState{pljit.StaticDeleteField, pljit.StaticSetTag},
				StateOffsets: []uint32{0, 2}, ValueOffsets: []uint32{0, 1}, Values: []any{int64(9)},
			},
		}}
		runJITGroup(runner, point.Logging, script, []int{0}, runs, nil)
		if pt.Get("foo") != "value" || pt.Get("bar") != nil || pt.Get(plStatus) != sFailed {
			t.Fatalf("status %v mutated or replayed: %#v", status, pt.KVMap())
		}
	}
}

func (*committedErrorStats) Metrics() []prometheus.Collector                    { return nil }
func (*committedErrorStats) WriteEvent(*plstats.ChangeEvent, map[string]string) {}
func (*committedErrorStats) ReadEvents(events []*plstats.ChangeEvent) []*plstats.ChangeEvent {
	return events
}
func (*committedErrorStats) WriteUpdateTime(map[string]string) {}
func (stats *committedErrorStats) WriteMetric(_ map[string]string, points, drops, errors float64, _ time.Duration) {
	stats.points += points
	stats.drops += drops
	stats.errors += errors
}

// Malformed native output must neither mutate the input nor replay the script.
func TestRunJITGroupApplyFailurePreservesOriginalWithoutReplay(t *testing.T) {
	script, err := NewPlScriptSimple(point.Logging, "apply-fallback.p", "rename(bar, foo)\n")
	if err != nil {
		t.Fatal(err)
	}
	pt := point.NewPoint("apply-fallback", point.NewKVs(map[string]any{"foo": "value"}), point.WithTime(time.Unix(1_700_000_000, 0)))
	runs := []pointRun{{point: pt, script: script, started: time.Now()}}
	delta, err := pljit.EncodeMutationDeltas([]pljit.MutationDelta{{
		RecordIndex: 0,
		Operations: []pljit.MutationOp{
			{Kind: pljit.MutationDeleteField, Key: "foo"},
			{Kind: pljit.MutationSetCategory, String: "metric"},
		},
	}})
	if err != nil {
		t.Fatal(err)
	}
	runner := &fakeJITProcessor{batch: pljit.Batch{Records: []pljit.Record{{Status: pljit.TerminalOK, Delta: delta}}}}
	runJITGroup(runner, point.Logging, script, []int{0}, runs, &lang.LogOption{})
	if runs[0].output == nil {
		t.Fatal("fallback did not produce output")
	}
	if got := runs[0].output.Get("bar"); got != nil {
		t.Fatalf("Go rename was replayed: %#v", runs[0].output.KVMap())
	}
	if got := runs[0].output.Get("foo"); got != "value" || runs[0].output.Get(plStatus) != sFailed {
		t.Fatalf("original point not retained with failure status: %#v", runs[0].output.KVMap())
	}
}

// A failed native call may already have performed side effects. Keep the
// original point explicitly failed, without replaying the Go drop_key script.
func TestRunJITGroupProcessFailureDoesNotReplay(t *testing.T) {
	script, err := NewPlScriptSimple(point.Logging, "process-fallback.p", "drop_key(foo)\n")
	if err != nil {
		t.Fatal(err)
	}
	pt := point.NewPoint("process-fallback", point.NewKVs(map[string]any{"foo": "value", "keep": int64(1)}))
	runs := []pointRun{{point: pt, script: script, started: time.Now()}}
	runner := &fakeJITProcessor{err: errors.New("native process failure")}
	inputBytes := jitBytesVec.WithLabelValues(point.Logging.String(), "input", "flat-v1")
	beforeCount, beforeBytes := readPrometheusSummary(t, inputBytes)
	runJITGroup(runner, point.Logging, script, []int{0}, runs, nil)
	afterCount, afterBytes := readPrometheusSummary(t, inputBytes)
	if afterCount-beforeCount != 1 || afterBytes <= beforeBytes {
		t.Fatalf("failed native submission input bytes: count delta=%d bytes delta=%v, want one positive sample",
			afterCount-beforeCount, afterBytes-beforeBytes)
	}
	if runs[0].output == nil || runs[0].created != nil {
		t.Fatalf("unexpected failed-native state: output=%#v created=%#v", runs[0].output, runs[0].created)
	}
	if runs[0].output.Get("foo") != "value" || runs[0].output.Get("keep") != int64(1) {
		t.Fatalf("failed native submission was replayed: %#v", runs[0].output.KVMap())
	}
	if runs[0].output.Get(plStatus) != sFailed || runs[0].output.Get("pl_msg") == nil {
		t.Fatalf("native failure was not explicitly reported: %#v", runs[0].output.KVMap())
	}
}

// A hosted builtin can handle a record-local native miss and still return a
// successful static result. DataKit must apply that result directly: replaying
// the script in pipeline-go would duplicate the work and overwrite mutations
// from instructions that ran after the host callback.
func TestRunJITGroupHostedDefaultTimeSuccessDoesNotReplay(t *testing.T) {
	script, err := NewPlScriptSimple(point.Logging, "hosted-default-time.p",
		"default_time(time)\nadd_key(after, \"pipeline-go\")\n")
	if err != nil {
		t.Fatal(err)
	}
	pointTime := time.Unix(1_700_000_000, 123)
	pt := point.NewPoint("hosted-default-time", point.NewKVs(map[string]any{"time": "not-a-time"}), point.WithTime(pointTime))
	runs := []pointRun{{point: pt, script: script, started: time.Now()}}
	runner := &fakeJITProcessor{batch: pljit.Batch{
		Records: []pljit.Record{{Status: pljit.TerminalOK}},
		Static: &pljit.StaticBatch{
			Schema:       pljit.StaticOutputSchema{Keys: []string{"time", "pl_msg", "after"}},
			States:       []pljit.StaticMutationState{pljit.StaticSetField, pljit.StaticSetField, pljit.StaticSetField},
			StateOffsets: []uint32{0, 3},
			ValueOffsets: []uint32{0, 3},
			Values:       []any{pointTime.UnixNano(), "native host parse error", "native continued"},
		},
	}}

	runJITGroup(runner, point.Logging, script, []int{0}, runs, nil)
	if runs[0].output == nil {
		t.Fatal("hosted default_time result produced no output")
	}
	if got := runs[0].output.Time(); !got.Equal(pointTime) {
		t.Fatalf("hosted default_time point time = %s, want %s", got, pointTime)
	}
	if got := runs[0].output.Get("pl_msg"); got != "native host parse error" {
		t.Fatalf("hosted default_time was replayed in pipeline-go: pl_msg=%#v", got)
	}
	if got := runs[0].output.Get("after"); got != "native continued" {
		t.Fatalf("instruction after hosted default_time was overwritten: after=%#v", got)
	}
}

func TestRunJITGroupCommittedErrorAppliesPrefixWithoutReplay(t *testing.T) {
	script, err := NewPlScriptSimple(point.Logging, "committed-error.p", "add_key(after, \"pipeline-go\")\n")
	if err != nil {
		t.Fatal(err)
	}
	originalTime := time.Unix(1_700_000_000, 123)
	pt := point.NewPoint("committed-error", point.NewKVs(map[string]any{"keep": "original"}), point.WithTime(originalTime))
	runs := []pointRun{{point: pt, script: script, started: time.Now()}}
	delta, err := pljit.EncodeMutationDeltas([]pljit.MutationDelta{{
		RecordIndex: 0,
		Operations: []pljit.MutationOp{
			{Kind: pljit.MutationSetField, Key: "before", Value: "native"},
			{Kind: pljit.MutationSetField, Key: "time", Value: int64(42)},
			{Kind: pljit.MutationSetDropped, Dropped: true},
			{Kind: pljit.MutationAddSubpoint, Point: &pljit.Point{Version: 1, Category: "logging", Measurement: "must-not-publish"}},
		},
	}})
	if err != nil {
		t.Fatal(err)
	}
	runner := &fakeJITProcessor{batch: pljit.Batch{Records: []pljit.Record{{
		Status:            pljit.TerminalError,
		Delta:             delta,
		Error:             []byte(`{"code":"E_PIPELINE_RUNTIME","message":"key not found","span":{"line":1,"column":1}}`),
		CommitPrefixError: true,
	}}}}
	stats := &committedErrorStats{}
	plstats.SetStats(stats)
	t.Cleanup(func() { plstats.SetStats(nil) })
	committedCounter := jitRouteRecordsVec.WithLabelValues(point.Logging.String(), "native", "committed_error")
	replayedCounter := jitRouteRecordsVec.WithLabelValues(point.Logging.String(), "replayed", "record_error")
	fallbackCounter := jitFallbackVec.WithLabelValues(point.Logging.String(), "record_error")
	beforeCommitted := readPrometheusCounter(t, committedCounter)
	beforeReplayed := readPrometheusCounter(t, replayedCounter)
	beforeFallback := readPrometheusCounter(t, fallbackCounter)

	runJITGroup(runner, point.Logging, script, []int{0}, runs, nil)
	if runs[0].output == nil || runs[0].output.Get("before") != "native" || runs[0].output.Get("after") != nil {
		t.Fatalf("committed error output = %#v", runs[0].output)
	}
	if runs[0].output.Get("time") != int64(42) || !runs[0].output.Time().Equal(originalTime) {
		t.Fatalf("committed error incorrectly finalized time: field=%#v point=%s", runs[0].output.Get("time"), runs[0].output.Time())
	}
	if runs[0].created != nil || runs[0].dropped {
		t.Fatalf("committed error published side effects: created=%#v dropped=%v", runs[0].created, runs[0].dropped)
	}
	if stats.points != 1 || stats.drops != 0 || stats.errors != 1 {
		t.Fatalf("committed error pipeline stats = points=%v drops=%v errors=%v", stats.points, stats.drops, stats.errors)
	}
	if got := readPrometheusCounter(t, committedCounter) - beforeCommitted; got != 1 {
		t.Fatalf("committed route metric delta = %v, want 1", got)
	}
	if got := readPrometheusCounter(t, replayedCounter) - beforeReplayed; got != 0 {
		t.Fatalf("committed error replay metric delta = %v, want 0", got)
	}
	if got := readPrometheusCounter(t, fallbackCounter) - beforeFallback; got != 0 {
		t.Fatalf("committed error fallback metric delta = %v, want 0", got)
	}
}

func TestRunJITGroupOrdinaryErrorDoesNotReplay(t *testing.T) {
	script, err := NewPlScriptSimple(point.Logging, "ordinary-error.p", "add_key(after, \"pipeline-go\")\n")
	if err != nil {
		t.Fatal(err)
	}
	runs := []pointRun{{
		point:   point.NewPoint("ordinary-error", point.NewKVs(map[string]any{"keep": "original"})),
		script:  script,
		started: time.Now(),
	}}
	runner := &fakeJITProcessor{batch: pljit.Batch{Records: []pljit.Record{{
		Status: pljit.TerminalError,
		Error:  []byte(`{"code":"E_PIPELINE_COMPAT"}`),
	}}}}
	runJITGroup(runner, point.Logging, script, []int{0}, runs, nil)
	if runs[0].output == nil || runs[0].output.Get("after") != nil || runs[0].output.Get(plStatus) != sFailed {
		t.Fatalf("ordinary terminal error was replayed or unreported: %#v", runs[0].output)
	}
}

func TestRunJITGroupCommittedErrorApplyFailureDoesNotReplay(t *testing.T) {
	script, err := NewPlScriptSimple(point.Logging, "committed-apply-failure.p", "duration_precision(foo, \"ms\", \"ns\")\n")
	if err != nil {
		t.Fatal(err)
	}
	pt := point.NewPoint("committed-apply-failure", point.NewKVs(map[string]any{"foo": int64(1)}))
	runs := []pointRun{{point: pt, script: script, started: time.Now()}}
	delta, err := pljit.EncodeMutationDeltas([]pljit.MutationDelta{{
		RecordIndex: 0,
		Operations: []pljit.MutationOp{
			{Kind: pljit.MutationDeleteField, Key: "foo"},
			{Kind: pljit.MutationSetCategory, String: "metric"},
		},
	}})
	if err != nil {
		t.Fatal(err)
	}
	runner := &countingJITProcessor{fakeJITProcessor: fakeJITProcessor{batch: pljit.Batch{Records: []pljit.Record{{
		Status:            pljit.TerminalError,
		Delta:             delta,
		Error:             []byte(`{"code":"E_PIPELINE_RUNTIME","message":"key not found"}`),
		CommitPrefixError: true,
	}}}}}
	runJITGroup(runner, point.Logging, script, []int{0}, runs, nil)
	if runner.processes != 1 || runs[0].output == nil || runs[0].output.Get("foo") != int64(1) || runs[0].output.Get(plStatus) != sFailed {
		t.Fatalf("committed apply failure replayed or mutated original: processes=%d output=%#v", runner.processes, runs[0].output)
	}
}

func TestRunJITGroupCommittedErrorAcceptsEmptyPrefixDelta(t *testing.T) {
	script, err := NewPlScriptSimple(point.Logging, "committed-empty-prefix.p", "add_key(after, \"pipeline-go\")\n")
	if err != nil {
		t.Fatal(err)
	}
	pt := point.NewPoint("committed-empty-prefix", point.NewKVs(map[string]any{"keep": "original"}))
	runs := []pointRun{{point: pt, script: script, started: time.Now()}}
	delta, err := pljit.EncodeMutationDeltas([]pljit.MutationDelta{{RecordIndex: 0}})
	if err != nil {
		t.Fatal(err)
	}
	runner := &fakeJITProcessor{batch: pljit.Batch{Records: []pljit.Record{{
		Status:            pljit.TerminalError,
		Delta:             delta,
		Error:             []byte(`{"code":"E_PIPELINE_RUNTIME","message":"key not found"}`),
		CommitPrefixError: true,
	}}}}
	runJITGroup(runner, point.Logging, script, []int{0}, runs, nil)
	if runs[0].output == nil || runs[0].output.Get("keep") != "original" || runs[0].output.Get("after") != nil {
		t.Fatalf("empty committed prefix was replayed or lost: %#v", runs[0].output)
	}
}

func TestAppendJITFailedRunInfo(t *testing.T) {
	run := &pointRun{point: point.NewPoint("run-info", nil), started: time.Now().Add(-time.Millisecond)}
	appendJITFailedRunInfoEnabled(run, true)
	if run.point.Get(plStatus) != sFailed {
		t.Fatalf("failed run status = %#v", run.point.Get(plStatus))
	}
	if cost, ok := run.point.Get(plFieldCost).(float64); !ok || cost <= 0 {
		t.Fatalf("failed run cost = %#v", run.point.Get(plFieldCost))
	}
	if got := committedJITError("duration.p", []byte(`{"message":"param value type expect int","span":{"line":2,"column":3}}`)).Error(); got != "duration.p:2:3: param value type expect int" {
		t.Fatalf("committed error log = %q", got)
	}
}

func TestExecutionCostStartsAtActualEngineEntry(t *testing.T) {
	script, err := NewPlScriptSimple(point.Logging, "cost-start.p", "drop_key(foo)\n")
	if err != nil {
		t.Fatal(err)
	}

	goRun := pointRun{
		point:   point.NewPoint("go", point.NewKVs(map[string]any{"foo": "remove"})),
		script:  script,
		started: time.Now().Add(-time.Hour),
	}
	goThreshold := time.Now()
	runPipelineGo(point.Logging, &goRun, nil)
	if goRun.started.Before(goThreshold) {
		t.Fatalf("pipeline-go cost retained routing start: %s < %s", goRun.started, goThreshold)
	}

	delta, err := pljit.EncodeMutationDeltas([]pljit.MutationDelta{{RecordIndex: 0}})
	if err != nil {
		t.Fatal(err)
	}
	jitRun := pointRun{
		point:   point.NewPoint("jit", point.NewKVs(map[string]any{"foo": "keep"})),
		script:  script,
		started: time.Now().Add(-time.Hour),
	}
	runner := &fakeJITProcessor{batch: pljit.Batch{Records: []pljit.Record{{
		Status: pljit.TerminalOK,
		Delta:  delta,
	}}}}
	runs := []pointRun{jitRun}
	jitThreshold := time.Now()
	runJITGroup(runner, point.Logging, script, []int{0}, runs, nil)
	if runs[0].started.Before(jitThreshold) {
		t.Fatalf("JIT cost retained routing start: %s < %s", runs[0].started, jitThreshold)
	}
}
