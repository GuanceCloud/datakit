// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build pipeline_jit && linux && (amd64 || arm64)

package pipeline

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"
	"unicode"

	"github.com/GuanceCloud/cliutils/point"
	"github.com/GuanceCloud/pipeline-go/constants"
	"github.com/GuanceCloud/pipeline-go/ptinput"
	pljit "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline/jit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline/plval"
)

func TestJITNativeManagerReloadWaitsForOldBatch(t *testing.T) {
	path := os.Getenv("PLATYPUS_JIT_RUNTIME")
	if path == "" {
		t.Fatal("requires real PLATYPUS_JIT_RUNTIME")
	}
	runner, err := pljit.NewRunnerWithHost(path, "pipeline-go-1.4.3-datakit", 2, pljit.NewPipelineGoHost(nil))
	if err != nil {
		t.Fatal(err)
	}
	defer runner.Close()
	const oldSource = "add_key(version, \"old\")\n"
	const newSource = "add_key(version, \"new\")\n"
	manager := plval.NewScriptManager(nil, nil)
	checked := make(chan struct{}, 1)
	invalidated := make(chan string, 2)
	manager.SetJITCheck(func(source string) bool {
		ok := runner.Check(source).Route == pljit.RouteJITNative
		if source == newSource {
			checked <- struct{}{}
		}
		return ok
	})
	manager.SetJITInvalidate(func(source string) { runner.Invalidate(source); invalidated <- source })
	if err := manager.LoadScriptWithCatChecked(point.Logging, constants.NSDefault, map[string]string{"reload.p": oldSource}, nil); err != nil {
		t.Fatal(err)
	}
	lease, ok := manager.Acquire()
	if !ok {
		t.Fatal("missing old lease")
	}
	defer lease.Release()
	updated := make(chan error, 1)
	go func() {
		updated <- manager.LoadScriptWithCatChecked(point.Logging, constants.NSDefault, map[string]string{"reload.p": newSource}, nil)
	}()
	select {
	case <-checked:
	case <-time.After(5 * time.Second):
		t.Fatal("new source check did not finish")
	}
	select {
	case err := <-updated:
		t.Fatalf("reload crossed old batch: %v", err)
	default:
	}
	select {
	case source := <-invalidated:
		t.Fatalf("source retired early: %s", source)
	default:
	}
	old, ok := lease.Manager().QueryScript(point.Logging, "reload.p")
	if !ok || old.Content() != oldSource || !lease.JITEligible(oldSource) {
		t.Fatal("old batch lost its generation")
	}
	run := func(source string) string {
		t.Helper()
		pt := newRealScriptPoint("reload", map[string]any{"message": "test"})
		projection, err := runner.Projection(source)
		if err != nil {
			t.Fatal(err)
		}
		input, err := encodeProjectedJITPoints(point.Logging, []*point.Point{pt}, projection)
		if err != nil {
			t.Fatal(err)
		}
		batch, err := runner.Process(source, input)
		if err != nil {
			t.Fatal(err)
		}
		if len(batch.Records) != 1 || batch.Records[0].Status != pljit.TerminalOK {
			t.Fatalf("bad native result: %#v", batch)
		}
		if batch.Static != nil {
			_, _, err = applyJITStatic(point.Logging, pt, batch.Static, 0, nil)
		} else {
			_, _, err = applyJITRecord(point.Logging, pt, batch.Records[0], 0, nil)
		}
		if err != nil {
			t.Fatal(err)
		}
		value, _ := pt.Get("version").(string)
		return value
	}
	if got := run(old.Content()); got != "old" {
		t.Fatalf("old batch result: %q", got)
	}
	lease.Release()
	select {
	case err := <-updated:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("reload did not drain")
	}
	select {
	case source := <-invalidated:
		if source != oldSource {
			t.Fatalf("retired wrong source: %s", source)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("old native instance not invalidated")
	}
	current, ok := manager.QueryScript(point.Logging, "reload.p")
	if !ok || current.Content() != newSource {
		t.Fatal("new script not published")
	}
	if got := run(current.Content()); got != "new" {
		t.Fatalf("new batch result: %q", got)
	}
}

func TestJITNativeBadEncodingDoesNotPoisonBatch(t *testing.T) {
	path := os.Getenv("PLATYPUS_JIT_RUNTIME")
	if path == "" {
		t.Fatal("requires real PLATYPUS_JIT_RUNTIME")
	}
	runner, err := pljit.NewRunnerWithHost(path, "pipeline-go-1.4.3-datakit", 2, pljit.NewPipelineGoHost(nil))
	if err != nil {
		t.Fatal(err)
	}
	defer runner.Close()
	const source = "value = len(message)\nadd_key(length, value)\n"
	script, err := NewPlScriptSimple(point.Logging, "isolated-native.p", source)
	if err != nil {
		t.Fatal(err)
	}
	if check := runner.Check(source); check.Route != pljit.RouteJITNative {
		t.Fatalf("route: %#v", check)
	}
	runs := make([]pointRun, 3)
	for i, message := range []string{"first", string([]byte{0xff}), "last"} {
		runs[i] = pointRun{point: newRealScriptPoint("isolated", map[string]any{"message": message}), script: script}
	}
	runJITGroup(runner, point.Logging, script, []int{0, 1, 2}, runs, nil)
	for i, want := range map[int]int64{0: 5, 2: 4} {
		if runs[i].output == nil || runs[i].output.Get("length") != want || runs[i].output.Get(plStatus) == sFailed {
			t.Fatalf("valid record %d: %#v", i, runs[i].output)
		}
	}
	if runs[1].output == nil || runs[1].output.Get(plStatus) == sFailed || runs[1].output.Get("length") != int64(1) || runs[1].output.Get("message") != string([]byte{0xff}) {
		t.Fatalf("raw string record was not preserved: %#v", runs[1].output)
	}
	// Measurement is a Go string too: raw bytes must not reject or poison the
	// record, and neighboring records must retain their own names.
	for i := range runs {
		runs[i] = pointRun{point: newRealScriptPoint("isolated", map[string]any{"message": "ok"}), script: script}
	}
	runs[1].point.AddKVs(point.NewKV("message", "ok"))
	runs[1].point.PBPoint().Name = string([]byte{0xff})
	runJITGroup(runner, point.Logging, script, []int{0, 1, 2}, runs, nil)
	for i := range runs {
		if runs[i].output == nil || runs[i].output.Get(plStatus) == sFailed || runs[i].output.Get("length") != int64(2) {
			t.Fatalf("measurement record %d failed: %#v", i, runs[i].output)
		}
	}
	if runs[1].output.Name() != string([]byte{0xff}) || runs[0].output.Name() != "isolated" || runs[2].output.Name() != "isolated" {
		t.Fatalf("raw measurement or neighbor isolation lost: %q %q %q", runs[0].output.Name(), runs[1].output.Name(), runs[2].output.Name())
	}
}

func TestJITNativeProductionSubpointTermination(t *testing.T) {
	path := os.Getenv("PLATYPUS_JIT_RUNTIME")
	if path == "" {
		t.Fatal("requires real PLATYPUS_JIT_RUNTIME")
	}
	const source = `add_key(before, true)
create_point("before", {}, {"value": sequence}, ts=1)
if sequence % 2 == 0 { drop() }
value = 1 / sequence
create_point("after", {}, {"value": sequence}, ts=2)
add_key(after, true)
`
	runner, err := pljit.NewRunnerWithHost(path, "pipeline-go-1.4.3-datakit", 2, pljit.NewPipelineGoHost(nil))
	if err != nil {
		t.Fatal(err)
	}
	defer runner.Close()
	if check := runner.Check(source); check.Route != pljit.RouteJITNative {
		t.Fatalf("not native: %#v", check)
	}
	script, err := NewPlScriptSimple(point.Logging, "subpoint-termination.p", source)
	if err != nil {
		t.Fatal(err)
	}
	for _, batchSize := range []int{1, 8, 128} {
		t.Run(fmt.Sprintf("batch-%d", batchSize), func(t *testing.T) {
			const count = 129
			actual, expected := make([]pointRun, count), make([]pointRun, count)
			for i := range actual {
				fields := map[string]any{"sequence": int64(i % 5), "message": "unchanged"}
				actual[i] = pointRun{point: newRealScriptPoint("parent", fields), script: script}
				expected[i] = pointRun{point: newRealScriptPoint("parent", fields), script: script}
				runPipelineGo(point.Logging, &expected[i], nil)
			}
			for start := 0; start < count; start += batchSize {
				indexes := make([]int, min(batchSize, count-start))
				for i := range indexes {
					indexes[i] = start + i
				}
				runJITGroup(runner, point.Logging, script, indexes, actual, nil)
			}
			for i := range actual {
				got, want := &actual[i], &expected[i]
				if got.dropped != want.dropped || got.offload != want.offload {
					t.Fatalf("record %d disposition mismatch", i)
				}
				if !got.dropped {
					if got.output == nil || want.output == nil {
						t.Fatalf("record %d missing output", i)
					}
					if equal, reason := got.output.EqualWithReason(want.output); !equal {
						t.Fatalf("record %d parent: %s", i, reason)
					}
				}
				if len(got.created) != len(want.created) {
					t.Fatalf("record %d category count differs", i)
				}
				for category, children := range want.created {
					if len(got.created[category]) != len(children) {
						t.Fatalf("record %d child count differs", i)
					}
					for j, child := range children {
						if equal, reason := got.created[category][j].EqualWithReason(child); !equal {
							t.Fatalf("record %d child %d: %s", i, j, reason)
						}
					}
				}
				// Independent assertions keep an accidental empty oracle from passing.
				if i%5 == 0 {
					if got.dropped || len(got.created) != 0 || got.output.Get("before") != true || got.output.Get("after") != nil {
						t.Fatalf("record %d error publication wrong", i)
					}
				} else if len(got.created[point.Metric]) != 2 {
					t.Fatalf("record %d lost children", i)
				}
			}
		})
	}
}

func TestJITNativeAbortedUpdateReleasesCandidateCapacity(t *testing.T) {
	path := os.Getenv("PLATYPUS_JIT_RUNTIME")
	if path == "" {
		t.Fatal("requires real PLATYPUS_JIT_RUNTIME")
	}
	runner, err := pljit.NewRunnerWithHost(path, "pipeline-go-1.4.3-datakit", 2, pljit.NewPipelineGoHost(nil))
	if err != nil {
		t.Fatal(err)
	}
	defer runner.Close()
	manager := plval.NewScriptManager(nil, nil)
	manager.SetJITCheck(func(source string) bool { return runner.Check(source).Route == pljit.RouteJITNative })
	manager.SetJITInvalidate(runner.Invalidate)
	const current = "add_key(version, \"current\")\n"
	if err := manager.LoadScriptWithCatChecked(point.Logging, constants.NSDefault, map[string]string{"current.p": current}, nil); err != nil {
		t.Fatal(err)
	}
	candidate, err := manager.PrepareRemoteUpdate(plval.RemoteManagerUpdate{
		ReplaceScripts: true,
		Scripts:        map[point.Category]map[string]string{point.Logging: {"candidate.p": "add_key(version, \"candidate\")\n"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer candidate.Abort()
	const third = "add_key(version, \"third\")\n"
	if err := runner.Prepare(third); err == nil {
		t.Fatal("candidate did not reserve capacity")
	}
	candidate.Abort()
	if check := runner.Check(third); check.Route != pljit.RouteJITNative {
		t.Fatalf("aborted candidate leaked capacity: %#v", check)
	}
	lease, ok := manager.Acquire()
	if !ok {
		t.Fatal("missing published manager")
	}
	defer lease.Release()
	if !lease.JITEligible(current) {
		t.Fatal("abort changed active route")
	}
	if _, ok := lease.Manager().QueryScript(point.Logging, "candidate.p"); ok {
		t.Fatal("aborted script was published")
	}
}

// Repeated aborted candidates must not exhaust capacity or replace active code.
func TestJITNativeRepeatedAbortedUpdatesRetainCurrent(t *testing.T) {
	path := os.Getenv("PLATYPUS_JIT_RUNTIME")
	if path == "" {
		t.Fatal("requires real PLATYPUS_JIT_RUNTIME")
	}
	runner, err := pljit.NewRunner(path, "pipeline-go-1.4.3-datakit", 2)
	if err != nil {
		t.Fatal(err)
	}
	defer runner.Close()
	manager := plval.NewScriptManager(nil, nil)
	manager.SetJITCheck(func(source string) bool { return runner.Check(source).Route == pljit.RouteJITNative })
	manager.SetJITInvalidate(runner.Invalidate)
	const current = "add_key(version,\"current\")\n"
	if err := manager.LoadScriptWithCatChecked(point.Logging, constants.NSDefault, map[string]string{"current.p": current}, nil); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 32; i++ {
		source := fmt.Sprintf("add_key(version,\"candidate%d\")\n", i)
		candidate, err := manager.PrepareRemoteUpdate(plval.RemoteManagerUpdate{ReplaceScripts: true, Scripts: map[point.Category]map[string]string{point.Logging: {"candidate.p": source}}})
		if err != nil {
			t.Fatalf("round%d prepare: %v", i, err)
		}
		// Abort must release the candidate's reservation on every generation.
		candidate.Abort()
		candidate.Abort()
		probe := fmt.Sprintf("add_key(probe,%d)\n", i)
		if err := runner.Prepare(probe); err != nil {
			t.Fatalf("round%d leaked capacity: %v", i, err)
		}
		runner.Invalidate(probe)
		lease, ok := manager.Acquire()
		if !ok {
			t.Fatalf("round%d missing active generation", i)
		}
		active, ok := lease.Manager().QueryScript(point.Logging, "current.p")
		eligible := lease.JITEligible(current)
		_, publishedCandidate := lease.Manager().QueryScript(point.Logging, "candidate.p")
		lease.Release()
		if !ok || active.Content() != current || !eligible || publishedCandidate {
			t.Fatalf("round%d aborted candidate changed publication", i)
		}
		pt := newRealScriptPoint("abort", map[string]any{"message": "test"})
		projection, err := runner.Projection(current)
		if err != nil {
			t.Fatal(err)
		}
		input, err := encodeProjectedJITPoints(point.Logging, []*point.Point{pt}, projection)
		if err != nil {
			t.Fatal(err)
		}
		batch, err := runner.Process(current, input)
		if err != nil {
			t.Fatal(err)
		}
		if len(batch.Records) != 1 || batch.Records[0].Status != pljit.TerminalOK {
			t.Fatalf("round%d active native execution failed", i)
		}
		if batch.Static != nil {
			_, _, err = applyJITStatic(point.Logging, pt, batch.Static, 0, nil)
		} else {
			_, _, err = applyJITRecord(point.Logging, pt, batch.Records[0], 0, nil)
		}
		if err != nil || pt.Get("version") != "current" {
			t.Fatalf("round%d current output=%v err=%v", i, pt.Get("version"), err)
		}
	}
}

// Named crash reproducer, not a passing native differential.
func TestJITUpstreamListSlicePanicReproducer(t *testing.T) {
	script, err := NewPlScriptSimple(point.Logging, "slice-panic.p", "items = [1, 2, 3, 4, 5]\nadd_key(value, items[99:])\n")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if recovered := recover(); recovered == nil {
			t.Fatal("upstream no longer panics; reevaluate compatibility boundary")
		}
	}()
	_ = script.Run(ptinput.PtWrap(point.Logging, newRealScriptPoint("slice", nil)), nil, nil)
}

func TestJITNativeSliceProjection(t *testing.T) {
	path := os.Getenv("PLATYPUS_JIT_RUNTIME")
	if path == "" {
		t.Fatal("requires real PLATYPUS_JIT_RUNTIME")
	}
	runner, err := pljit.NewRunnerWithHost(path, "pipeline-go-1.4.3-datakit", 2, pljit.NewPipelineGoHost(nil))
	if err != nil {
		t.Fatal(err)
	}
	defer runner.Close()
	const source = "add_key(result, message[start:end:step])\n"
	projection, err := runner.Projection(source)
	if err != nil {
		t.Fatal(err)
	}
	fields := map[string]any{"message": "abcdef", "start": int64(1), "end": int64(5), "step": int64(2), "result": "old", "unrelated": "keep"}
	pt := newRealScriptPoint("slice", fields)
	input, err := encodeProjectedJITPoints(point.Logging, []*point.Point{pt}, projection)
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := pljit.DecodeFlatPoints(input)
	if err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"message", "start", "end", "step", "result"} {
		if decoded[0].Fields[key] != fields[key] {
			t.Fatalf("projection lost %s: %#v", key, decoded[0])
		}
	}
	if _, exists := decoded[0].Fields["unrelated"]; exists {
		t.Fatal("pure slice did not use projection")
	}
	batch, err := runner.Process(source, input)
	if err != nil {
		t.Fatal(err)
	}
	if len(batch.Records) != 1 || batch.Records[0].Status != pljit.TerminalOK {
		t.Fatalf("native: %#v", batch)
	}
	if batch.Static == nil {
		t.Fatal("missing static output for projected slice")
	}
	if _, _, err := applyJITStatic(point.Logging, pt, batch.Static, 0, nil); err != nil {
		t.Fatal(err)
	}
	if pt.Get("result") != "bd" || pt.Get("unrelated") != "keep" {
		t.Fatalf("wrong slice output: %#v", pt.KVMap())
	}
}

func TestJITNativeRawStringFunctions(t *testing.T) {
	path := os.Getenv("PLATYPUS_JIT_RUNTIME")
	if path == "" {
		t.Fatal("requires real PLATYPUS_JIT_RUNTIME")
	}
	runner, err := pljit.NewRunnerWithHost(path, "pipeline-go-1.4.3-datakit", 4, pljit.NewPipelineGoHost(nil))
	if err != nil {
		t.Fatal(err)
	}
	defer runner.Close()
	// Hash methods mirror pipeline-go v1.4.3 fn_hash_test.go::TestHash;
	// extend its text vectors with binary and malformed UTF-8 inputs.
	for _, target := range []string{"int", "float", "bool", "str", "md5", "sha1", "sha256", "sha512", "xx", "b64enc", "b64dec", "len", "strlen", "lowercase", "uppercase", "url_decode", "strfmt_q", "strfmt_s", "strfmt_x", "strfmt_T", "concat", "concat_error", "compare_error", "[:]", "[1:3]", "[::-1]", "[-3:-1]", "[99:]", "[::0]", "[true:]", "[0]"} {
		source := fmt.Sprintf("cast(message, %q)\nadd_key(done, true)\n", target)
		if target == "concat" {
			source = "original = message\ncopy = message + \"\"\nif copy == original { add_key(before, true) }\ncopy += message\nadd_key(message, \"[\" + copy + \"]\")\nadd_key(done, true)\n"
		}
		if target == "concat_error" {
			source = "add_key(before, true)\nadd_key(message, message + 1)\nadd_key(done, true)\n"
		}
		if target == "compare_error" {
			source = "add_key(before, true)\nif message < \"z\" { add_key(done, true) }\n"
		}
		if target[0] == '[' {
			source = fmt.Sprintf("add_key(before, true)\nadd_key(message, message%s)\nadd_key(done, true)\n", target)
		}
		if target == "md5" || target == "sha1" || target == "sha256" || target == "sha512" || target == "xx" {
			source = fmt.Sprintf("add_key(message, hash(message, %q))\nadd_key(done, true)\n", target)
		}
		if target == "b64enc" || target == "b64dec" || target == "lowercase" || target == "uppercase" || target == "url_decode" {
			source = fmt.Sprintf("add_key(before, true)\n%s(message)\nadd_key(done, true)\n", target)
		}
		if target == "len" || target == "strlen" {
			source = fmt.Sprintf("add_key(message, %s(message))\nadd_key(done, true)\n", target)
		}
		if strings.HasPrefix(target, "strfmt_") {
			source = fmt.Sprintf("strfmt(message, %q, message)\nadd_key(done, true)\n", "%"+strings.TrimPrefix(target, "strfmt_"))
		}
		script, err := NewPlScriptSimple(point.Logging, "raw-cast.p", source)
		if err != nil {
			t.Fatal(err)
		}
		inputs := []string{"", "42", "-12.5", "1e3", "true", "T", "FALSE", "\xff", "42\x00", " 42", "测"}
		if target == "url_decode" {
			inputs = append(inputs, "%FF", "%ff%00", "%E2%82", "a+b%2Bc", "%25FF", "%", "%0", "%GG", "ok%20bad%", "\xff+%41", "%E6%B5%8B")
		}
		if target == "lowercase" || target == "uppercase" {
			inputs = append(inputs, "a\xe2\x82Z", "a\xf0\x9f\x92Z", "测\x00试🦀�", "\xed\xa0\x80", "\xc0\xaf")
			if unicode.Version != "15.0.0" {
				t.Fatalf("Unicode baseline changed: %s; refresh Rust case table before accepting results", unicode.Version)
			}
			// Every valid Unicode scalar, in bounded chunks, through the actual
			// Go script and native helper. Includes unmapped and supplementary runes.
			chunk := make([]rune, 0, 4096)
			for r := rune(0); r <= unicode.MaxRune; r++ {
				if r >= 0xd800 && r <= 0xdfff {
					continue
				}
				chunk = append(chunk, r)
				if len(chunk) == cap(chunk) {
					inputs = append(inputs, string(chunk))
					chunk = chunk[:0]
				}
			}
			if len(chunk) != 0 {
				inputs = append(inputs, string(chunk))
			}
		}
		if target == "len" || target == "strlen" {
			inputs = append(inputs, "\xe2\x82", "\xf0\x9f\x92", "测\x00试🦀�", "\xed\xa0\x80", "\xc0\xaf")
		}
		if target == "b64dec" {
			inputs = append(inputs, "/w==", "4oI=", "YQ==", "YR==", "YWJ=", "Y\r\nQ==\n", "YQ", "YQ===", "Y Q==", "YQ==x", "YWJj!", "\r\n")
		}
		for inputIndex, raw := range inputs {
			name := fmt.Sprintf("%s/%x", target, raw)
			if len(raw) > 100 {
				name = fmt.Sprintf("%s/unicode-chunk-%d", target, inputIndex)
			}
			t.Run(name, func(t *testing.T) {
				expected := newRealScriptPoint("cast", map[string]any{"message": raw})
				wrapped := ptinput.PtWrap(point.Logging, expected)
				goErr := script.Run(wrapped, nil, nil)
				input, err := pljit.EncodeFlatPoints([]pljit.Point{{Version: 1, Category: "logging", Measurement: "cast", Fields: map[string]any{"message": pljit.RawString(raw)}}})
				if err != nil {
					t.Fatal(err)
				}
				batch, err := runner.Process(source, input)
				if err != nil {
					t.Fatal(err)
				}
				wantStatus := pljit.TerminalOK
				if goErr != nil {
					wantStatus = pljit.TerminalError
				}
				if len(batch.Records) != 1 || batch.Records[0].Status != wantStatus {
					t.Fatalf("native: %#v", batch)
				}
				actual := newRealScriptPoint("cast", map[string]any{"message": raw})
				if batch.Static != nil {
					_, _, err = applyJITStatic(point.Logging, actual, batch.Static, 0, nil)
				} else {
					_, _, err = applyJITRecord(point.Logging, actual, batch.Records[0], 0, nil)
				}
				if err != nil {
					t.Fatal(err)
				}
				for _, key := range []string{"message", "before", "done"} {
					if actual.Get(key) != wrapped.Point().Get(key) {
						t.Fatalf("%s: native %#v, Go %#v", key, actual.Get(key), wrapped.Point().Get(key))
					}
				}
			})
		}
	}
}

func TestJITNativeRawStringOutput(t *testing.T) {
	path := os.Getenv("PLATYPUS_JIT_RUNTIME")
	if path == "" {
		t.Fatal("requires real PLATYPUS_JIT_RUNTIME")
	}
	runner, err := pljit.NewRunnerWithHost(path, "pipeline-go-1.4.3-datakit", 2, pljit.NewPipelineGoHost(nil))
	if err != nil {
		t.Fatal(err)
	}
	defer runner.Close()
	for _, raw := range []string{"", "测\x00试", "\xff\x00", "\xe2\x82", "\xb2\xe2\xca\xd4"} {
		t.Run(fmt.Sprintf("%x", raw), func(t *testing.T) {
			input, err := pljit.EncodeFlatPoints([]pljit.Point{{Version: 1, Category: "logging", Measurement: "raw", Fields: map[string]any{"message": pljit.RawString(raw)}}})
			if err != nil {
				t.Fatal(err)
			}
			batch, err := runner.Process("add_key(copied, message)\nvalue = [message]\ncast(value, \"str\")\nadd_key(encoded, value)\n", input)
			if err != nil {
				t.Fatal(err)
			}
			if len(batch.Records) != 1 || batch.Records[0].Status != pljit.TerminalOK {
				t.Fatalf("raw copy failed: %#v", batch)
			}
			pt := newRealScriptPoint("raw", map[string]any{"message": raw})
			if batch.Static != nil {
				_, _, err = applyJITStatic(point.Logging, pt, batch.Static, 0, nil)
			} else {
				_, _, err = applyJITRecord(point.Logging, pt, batch.Records[0], 0, nil)
			}
			if err != nil {
				t.Fatal(err)
			}
			for _, key := range []string{"message", "copied"} {
				if got, ok := pt.Get(key).(string); !ok || got != raw {
					t.Fatalf("%s: got %#v, want bytes %x", key, pt.Get(key), raw)
				}
			}
			wantJSON, err := json.Marshal([]string{raw})
			if err != nil {
				t.Fatal(err)
			}
			if got := pt.Get("encoded"); got != string(wantJSON) {
				t.Fatalf("JSON: got %#v, want %s", got, wantJSON)
			}
		})
	}
}

func TestJITNativeRawGBKDecode(t *testing.T) {
	path := os.Getenv("PLATYPUS_JIT_RUNTIME")
	if path == "" {
		t.Fatal("requires real PLATYPUS_JIT_RUNTIME")
	}
	runner, err := pljit.NewRunnerWithHost(path, "pipeline-go-1.4.3-datakit", 2, pljit.NewPipelineGoHost(nil))
	if err != nil {
		t.Fatal(err)
	}
	defer runner.Close()
	const source = "decode(message, \"gbk\")\nadd_key(done, true)\n"
	// Explicit experimental value type: normal production encoding is still
	// disabled until raw-value capability negotiation and all consumers agree.
	input, err := pljit.EncodeFlatPoints([]pljit.Point{{Version: 1, Category: "logging", Measurement: "gbk", Fields: map[string]any{
		"message": pljit.RawString(string([]byte{0xb2, 0xe2, 0xca, 0xd4, 0xd2, 0xbb, 0xcf, 0xc2})),
	}}})
	if err != nil {
		t.Fatal(err)
	}
	batch, err := runner.Process(source, input)
	if err != nil {
		t.Fatal(err)
	}
	if len(batch.Records) != 1 || batch.Records[0].Status != pljit.TerminalOK {
		t.Fatalf("raw decode failed: %#v", batch)
	}
	pt := newRealScriptPoint("gbk", map[string]any{})
	if batch.Static != nil {
		_, _, err = applyJITStatic(point.Logging, pt, batch.Static, 0, nil)
	} else {
		_, _, err = applyJITRecord(point.Logging, pt, batch.Records[0], 0, nil)
	}
	if err != nil {
		t.Fatal(err)
	}
	if pt.Get("message") != "测试一下" || pt.Get("done") != true {
		t.Fatalf("raw bytes decoded incorrectly: %#v", pt.KVMap())
	}
}

// Call the actual native runner directly: no routing decision may turn this
// test into a Go-against-Go comparison.
func TestJITNativeStateSurvivesCapacityPressure(t *testing.T) {
	path := os.Getenv("PLATYPUS_JIT_RUNTIME")
	if path == "" {
		t.Fatal("requires real PLATYPUS_JIT_RUNTIME")
	}
	runner, err := pljit.NewRunnerWithHost(path, "pipeline-go-1.4.3-datakit", 1, pljit.NewPipelineGoHost(nil))
	if err != nil {
		t.Fatal(err)
	}
	defer runner.Close()
	const source = `if write == true { cache_set("capacity-state", message) }
previous = cache_get("capacity-state")
if previous != nil { add_key(retained, previous) }
`
	if check := runner.Check(source); check.Route != pljit.RouteJITNative {
		t.Fatalf("route: %#v", check)
	}
	process := func(write bool, message string) *point.Point {
		t.Helper()
		pt := newRealScriptPoint("capacity", map[string]any{"write": write, "message": message})
		projection, err := runner.Projection(source)
		if err != nil {
			t.Fatal(err)
		}
		input, err := encodeProjectedJITPoints(point.Logging, []*point.Point{pt}, projection)
		if err != nil {
			t.Fatal(err)
		}
		batch, err := runner.Process(source, input)
		if err != nil {
			t.Fatal(err)
		}
		if len(batch.Records) != 1 || batch.Records[0].Status != pljit.TerminalOK {
			t.Fatalf("native failure: %#v", batch)
		}
		if batch.Static != nil {
			_, _, err = applyJITStatic(point.Logging, pt, batch.Static, 0, nil)
		} else {
			_, _, err = applyJITRecord(point.Logging, pt, batch.Records[0], 0, nil)
		}
		if err != nil {
			t.Fatal(err)
		}
		return pt
	}
	if got := process(true, "seed").Get("retained"); got != "seed" {
		t.Fatalf("initial state: %#v", got)
	}
	for i := 0; i < 8; i++ {
		if err := runner.Prepare(fmt.Sprintf("add_key(pressure, %d)\n", i)); err == nil {
			t.Fatal("capacity limit evicted pinned native state")
		}
		if got := process(false, "ignored").Get("retained"); got != "seed" {
			t.Fatalf("state lost under pressure %d: %#v", i, got)
		}
	}
	runner.Invalidate(source)
	if check := runner.Check(source); check.Route != pljit.RouteJITNative {
		t.Fatalf("new route: %#v", check)
	}
	if got := process(false, "ignored").Get("retained"); got != nil {
		t.Fatalf("retired state leaked to new instance: %#v", got)
	}
}

// Native compile rejection, failure and mismatched output are never replayed.
// These oracle fixtures pin upstream loop behavior before machine-code lowering
// is enabled. Passing this test alone is not evidence of JIT loop support.
//
//go:noinline
func appendCapacityProbe(values []any, item any) []any { return append(values, item) }

func TestGoListCapacityOracle(t *testing.T) {
	var values []any
	previous := 0
	expected := []int{1, 2, 4, 8, 16, 32, 71, 143, 303, 591, 1023, 1535, 2560, 3584, 5120}
	change := 0
	for i := 0; i < 4096; i++ {
		values = appendCapacityProbe(values, i)
		if cap(values) != previous {
			if change >= len(expected) || cap(values) != expected[change] {
				t.Fatalf("Go capacity baseline changed at len=%d: cap=%d", len(values), cap(values))
			}
			change++
			t.Logf("len=%d cap=%d", len(values), cap(values))
			previous = cap(values)
		}
	}
	if change != len(expected) {
		t.Fatalf("missing capacity transitions: %d", change)
	}
}

func TestPipelineGoLoopOracle(t *testing.T) {
	cases := []struct {
		name, source string
		want         int64
		fails        bool
	}{
		{"continue_step_break", `total = 0
for i = 0; i < 6; i += 1 {
 if i == 1 { continue }
 if i == 4 { break }
 total += i
}
add_key(total, total)`, 5, false},
		{"nested_break", `total = 0
for i = 0; i < 3; i += 1 {
 for j = 0; j < 4; j += 1 {
  if j == 2 { break }
  total += 1
 }
}
add_key(total, total)`, 6, false},
		{"list_values", `total = 0
for item in [2, 4, 6] { total += item }
add_key(total, total)`, 12, false},
		{"list_live_slots", `total = 0
items = [1, 2]
for item in items {
 total += item
 items[1] = 9
}
add_key(total, total)`, 10, false},
		{"unicode_runes", `total = 0
for item in "a中😀" { total += 1 }
add_key(total, total)`, 3, false},
		{"map_keys", `total = 0
values = {"a": 2, "b": 4}
for key in values { total += values[key] }
add_key(total, total)`, 6, false},
		{"error_prefix", `add_key(total, 0)
for item in [2, 1, 0, 4] {
 value = 4 / item
 add_key(total, total + 1)
}
add_key(after, true)`, 2, true},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			script, err := NewPlScriptSimple(point.Logging, test.name+".p", test.source)
			if err != nil {
				t.Fatal(err)
			}
			input := ptinput.PtWrap(point.Logging, newRealScriptPoint("loop", map[string]any{"message": "test"}))
			err = script.Run(input, nil, nil)
			if (err != nil) != test.fails {
				t.Fatalf("error outcome: %v", err)
			}
			if got := input.Point().Get("total"); got != test.want {
				t.Fatalf("total: got %v want %d", got, test.want)
			}
			if test.fails && input.Point().Get("after") != nil {
				t.Fatal("continued after loop error")
			}
		})
	}
}

func TestJITComplexStrictDifferential(t *testing.T) {
	path := os.Getenv("PLATYPUS_JIT_RUNTIME")
	if path == "" {
		t.Fatal("strict differential requires PLATYPUS_JIT_RUNTIME; skipping is not a release pass")
	}
	cases := []struct{ name, source string }{
		{"decode_utf16_and_invalid_sequence", `add_key(text, "A\u0000B\u0000")
decode(text, "utf-16le")
add_key(decoded, text)
add_key(odd, "A")
decode(odd, "utf-16le")
add_key(unknown, "unchanged")
decode(unknown, "UTF-16LE")
add_key(done, true)
`},
		{"literal_container_isolation", `base = []
items = append(base, [7])
add_key(before, items[0][0])
items[0][0] = sequence
add_key(after, items[0][0])
`},
		{"parallel_assignment", `left, right = sequence, sequence + 1
left, right = right, left
items = [1, 2]
items[0], items[1] = items[1], items[0]
add_key(left_result, left)
add_key(right_result, right)
add_key(first, items[0])
add_key(second, items[1])
`},
		{"parallel_assignment_rhs_error", `add_key(before, "retained")
sequence, temporary = 999, 128 / sequence
add_key(after, sequence)
`},
		{"parallel_assignment_target_error", `items = [1]
add_key(before, "retained")
sequence, items[2] = 999, 3
add_key(after, sequence)
`},
		{"index_assignment_alias", `values = [1, 2, [3]]
alias = values
values[-1][0] = sequence
values[0] += 2
object = {"items": values}
object["new"] = message
add_key(first, alias[0])
add_key(nested, object["items"][-1][0])
add_key(inserted, object["new"])
`},
		{"index_assignment_range_error", `values = [7, 8]
add_key(before, values[0])
values[sequence] = 99
add_key(after, values[0])
`},
		{"index_assignment_type_error", `object = {"value": "text"}
add_key(before, object["value"])
object["value"] += sequence
add_key(after, object["value"])
`},
		{"dynamic_containers", `values = [sequence, message, [sequence + 1]]
object = {"values": values, "length": len(message)}
add_key(first, object["values"][0])
add_key(nested, object["values"][-1][0])
add_key(size, object["length"])
add_key(dynamic_argument, len([message, sequence]))
add_key(parenthesized, (sequence + 2))
`},
		{"set_map_nil_members_nonraw", `values = {"nil_value": nil, "nested_list": [nil, 1], "nested_map": {"value": nil}, "text": "ok"}
count = pt_kvs_set_map(values, key_patterns=["*"])
add_key(set_count, count)
add_key(after, true)
`},
		{"set_map_nil_members_raw", `values = {"nil_value": nil, "nested_list": [nil, 1], "nested_map": {"value": nil}, "text": "ok"}
count = pt_kvs_set_map(values, key_patterns=["*"], raw=true)
add_key(set_count, count)
add_key(after, true)
`},
		{"set_map_nil_members_tags", `values = {"nil_value": nil, "number": 7, "flag": true, "text": "ok"}
count = pt_kvs_set_map(values, key_patterns=["*"], as_tag=true)
add_key(set_count, count)
add_key(after, true)
`},
		{"set_map_container_tags", `values = {"list": [1, nil, "x\"y"], "map": {"number": 7, "nil_value": nil}, "nested": [{"flag": true}, nil], "nil_value": nil}
count = pt_kvs_set_map(values, key_patterns=["*"], as_tag=true, raw=true)
add_key(set_count, count)
add_key(after, true)
`},
		{"set_map_deep_alias_nonraw", `leaf = {"nil_value": nil, "text": "ok"}
nested = {"level1": [{"level2": [leaf]}]}
alias = nested["level1"][0]["level2"][0]
alias["sequence"] = sequence
count = pt_kvs_set_map({"deep": nested}, key_patterns=["*"])
add_key(set_count, count)
add_key(after, true)
`},
		{"set_map_deep_alias_raw", `leaf = {"nil_value": nil, "text": "ok"}
nested = {"level1": [{"level2": [leaf]}]}
alias = nested["level1"][0]["level2"][0]
alias["sequence"] = sequence
count = pt_kvs_set_map({"deep": nested}, key_patterns=["*"], raw=true)
add_key(set_count, count)
add_key(after, true)
`},
		{"set_map_deep_alias_tag", `leaf = {"nil_value": nil, "text": "ok"}
nested = {"level1": [{"level2": [leaf]}]}
alias = nested["level1"][0]["level2"][0]
alias["sequence"] = sequence
count = pt_kvs_set_map({"deep": nested}, key_patterns=["*"], as_tag=true)
add_key(set_count, count)
add_key(after, true)
`},
		{"attribute_flat_key_lifecycle", `add_key("obj.name", " Value ")
trim(obj.name)
rename(obj.name, renamed)
set_tag(renamed)
add_key(after, true)
`},
		{"origin_alias_lifecycle", `add_key(_, " Value ")
trim(_)
rename(_, renamed_message)
add_key(observed, renamed_message)
add_key(after, true)
`},
		{"container_element_error", `add_key(before, "retained")
values = [sequence, 128 / sequence]
add_key(after, values[-1])
`},
		{"compound_mixed_types", `value = true
value += 2.5
add_key(mixed, value)
value = "prefix"
add_key(before, value)
value += sequence
add_key(after, value)
`},
		{"compound_assignments", `value = sequence
value += 7
value *= 3
value -= 2
value /= 2
value %= 5
add_key(result, value)
text = "prefix"
text += message
add_key(text_result, text)
value += len(message)
add_key(call_result, value)
`},
		{"compound_assignment_error", `add_key(before, "retained")
sequence /= sequence
add_key(after, sequence)
`},
		{"value_assignment_scope", `value = 3
value = value + sequence
if sequence > 0 { value = value * 2
  branch_local = "private"
  add_key(in_branch, value)
}
add_key(result, value)
add_key(local_visible, branch_local)
copy = message
add_key(copied, copy)
`},
		{"value_assignment_error", `value = 9
add_key(before, value)
value = 128 / sequence
add_key(after, value)
`},
		// Adapted from GuanceCloud/platypus (MIT), pkg/engine/runtime/
		// runtime_test.go::TestCondOp, via pipeline-go's pinned dependency.
		// Copyright 2021-present Guance, Inc.
		{"upstream_mixed_equality", `add_key(nil_equal, nil == nil)
add_key(int_nil, 1 == nil)
add_key(int_float, 1 == 1.0)
add_key(int_bool, 1 == true)
add_key(bool_float, true == 1.0)
add_key(false_float, false == 1.0)
add_key(int_string, 1 == "")
add_key(empty_string, "" == "")
add_key(string_bool, "" == true)
add_key(nil_bool, nil == true)
`},
		{"arithmetic_error_prefix", `add_key(before, "retained")
add_key(quotient, 128 / sequence)
add_key(after, "success")
`},
		{"logical_type_error_prefix", `add_key(before, "retained")
if sequence && true { add_key(inside, true) }
add_key(after, "success")
`},
		{"condition_truthiness", `if sequence { add_key(nonzero, true) } else { add_key(nonzero, false) }
if message { add_key(nonempty, true) }
if load_json("[]") { add_key(wrong_empty, true) }
if load_json("[0]") { add_key(list_present, true) }
if load_json("{}") { add_key(wrong_map, true) }
if missing_field { add_key(wrong_missing, true) }
add_key(done, true)
`},
		{"arithmetic_range", `add_key(calculated, (sequence + 3) * 2 - sequence % 3)
if sequence >= 1 && sequence < 64 { add_key(range, "middle") } else { add_key(range, "outside") }
if sequence == 0 || sequence > 127 { add_key(edge, true) }
`},
		{"short_circuit", `if false && (1 / sequence > 0) { add_key(unreachable, true) }
if true || (1 / sequence > 0) { add_key(always, true) }
add_key(done, true)
`},
		// Adapted from GuanceCloud/pipeline-go v1.4.3 (MIT),
		// ptinput/funcs/fn_load_json_test.go, TestLoadJson.
		// Copyright 2021-present Guance, Inc.
		{"upstream_negative_index", `abc = load_json("[2.2, 1.1]")
add_key(abc, abc[-1])
`},
		{"drop_branch_state", `if sequence % 2 == 0 {
  drop()
  cache_set("drop-state", "after-drop")
} else {
  seen = cache_get("drop-state")
  add_key(seen, seen)
}
`},
		{"error_after_drop", `add_key(before, true)
drop()
value = 1 / sequence
add_key(after, true)
`},
		{"error_before_drop", `add_key(before, true)
value = 1 / sequence
drop()
`},
		{"compound_rhs_exit_missing", `add_key(before, true)
missing += exit()
add_key(after, true)
`},
		{"compound_rhs_exit_present", `value = 1
add_key(before, true)
value += exit()
add_key(after, true)
`},
		{"compound_rhs_creates_target", `add_key(before, true)
created += add_key(created, 7)
add_key(after, true)
`},
		{"compound_rhs_deletes_target", `add_key(target, 7)
target += drop_key(target)
add_key(after, true)
`},
		{"compound_rhs_error", `add_key(before, true)
missing += b64dec(message)
add_key(after, true)
`},
		{"missing_compound_rhs_effect", `missing += add_key(rhs, "executed")
add_key(done, true)
`},
		{"present_compound_rhs_effect", `value = 1
add_key(before, true)
value += add_key(rhs, "executed")
add_key(after, true)
`},
		{"compound_rhs_call", `value = 1
value += len(message)
value *= len("ab")
add_key(result, value)
`},
		{"missing_compound", `missing += 1
missing -= 1
missing *= 2
missing /= 2
missing %= 2
add_key(done, true)
`},
		{"missing_compound_rhs_error", `add_key(before, true)
missing += 1 / sequence
add_key(after, true)
`},
		{"present_nil_compound", `value = nil
add_key(before, true)
value += 1
add_key(after, true)
`},
		{"indexed_string_compound", `items = [message]
alias = items
items[0] += "-suffix"
add_key(list_result, alias[0])
object = {"items": items, "text": "prefix-"}
object["items"][0] += "-nested"
object["text"] += message
add_key(nested_result, alias[0])
add_key(map_result, object["text"])
add_key(before, true)
object["items"][0] += 1
add_key(after, true)
`},
		{"slice_alias_lifecycle", `items = [[1], [2], [3]]
add_key(initial, items[0][0])
part = items[:2]
part[0][0] = 7
add_key(shared_nested, items[0][0])
part[1] = [9]
add_key(original_second, items[1][0])
add_key(replaced_second, part[1][0])
reversed = items[::-1]
reversed[0][0] = 8
add_key(reverse_nested, items[2][0])
part[0] = [11]
add_key(original_first, items[0][0])
add_key(part_first, part[0][0])
add_key(before, true)
part[99] = 1
add_key(after, true)
`},
		{"slice_composition", `items = [1, 2, 3, 4, 5]
add_key(reverse, items[::-1])
add_key(stride, items[1:5:2])
add_key(negative, items[-3:-1])
add_key(clipped, items[-99:99])
add_key(empty, items[5:])
add_key(nested, len(items[1:3]))
part = message[1:3]
add_key(part, part)
add_key(before, true)
add_key(invalid, items[::0])
add_key(after, true)
`},
		{"upstream_nested_calls", `abc = load_json('{"a":{"first":[2.2,1.1],"ff":"[2.2,1.1]"}}')
add_key(abc, abc["a"]["first"][-1])
add_key(len_abc, len(load_json(abc["a"]["ff"])))
`},
		{"nested_json", `data = load_json(message)
if data["enabled"] != nil {
  if data["enabled"] == true {
    add_key(result, data["value"])
    lowercase(result)
  } else { add_key(result, "disabled") }
} else { add_key(result, "missing") }
add_key(done, true)
`},
		{"grok_branch", `add_pattern("CUSTOM_WORD", "[A-Za-z]+")
grok(message, "%{CUSTOM_WORD:word}:%{INT:count}")
if word != nil { lowercase(word)
  if word == "error" { add_key(level, "high") } else { add_key(level, "low") }
}
drop_key(temporary)
`},
		{"error_prefix", `add_key(before, "retained")
duration_precision(duration, "ns", "ms")
add_key(after, "success")
`},
		{"for_continue_break", `total = 0
for i = 0; i < 6; i += 1 {
 if i == 1 { continue }
 if i == 4 { break }
 total += i
}
add_key(total, total)
add_key(scope_i, i)
`},
		{"for_nested_control", `total = 0
for i = 0; i < 4; i += 1 {
 for j = 0; j < 5; j += 1 {
  if j == 1 { continue }
  if j == 3 { break }
  total += 1
 }
 if i == 2 { continue }
 total += 10
}
add_key(total, total)
`},
		{"for_control_scope", `total = 0
for i = 0; i < 5; i += 1 {
 if i % 2 == 0 { local = 99
  if i == 4 { break }
  continue
 }
 total += i
}
add_key(total, total)
add_key(leaked, local)
`},
		{"slice_reverse_append_capacity", `items = [1, 2, 3, 4]
copied = items[3:0:-1]
extended = append(copied, 9)
copied[1] = 7
add_key(copied, copied)
add_key(extended, extended)
`},
		{"slice_append_capacity", `items = [1, 2, 3, 4]
copied = items[0:3]
extended = append(copied, 9)
copied[1] = 7
add_key(copied, copied)
add_key(extended, extended)
`},
		{"append_self_mutate", `items = [1, 2, 3]
extended = append(items, items)
items[1] = 7
add_key(extended, extended)
add_key(items, items)
`},
		{"append_self_extract", `items = [1, 2, 3]
extended = append(items, items)
nested = extended[3]
items = nil
extended = nil
nested[1] = 7
add_key(nested, nested)
`},
		{"append_self_view", `items = [1, 2, 3]
extended = append(items, items)
add_key(extended, extended)
add_key(done, true)
`},
		{"append_shared_backing", `items = [1, 2, 3]
extended = append(items, 9)
items[1] = 7
add_key(items, items)
add_key(extended, extended)
`},
		{"forin_append", `items = [1, 2, 3]
total = 0
for item in items { total += item
 items = append(items, 9)
}
add_key(total, total)
add_key(items, items)
`},
		{"forin_append_alias", `items = [1, 2]
alias = items
total = 0
for item in items { total += item
 items = append(items, 9)
 alias[1] = 7
}
add_key(total, total)
add_key(items, items)
add_key(alias, alias)
`},
		{"forin_append_spare_capacity", `items = [1, 2, 3]
alias = items
total = 0
for item in items { total += item
 items = append(items, 9)
 alias[1] = 7
}
add_key(total, total)
add_key(items, items)
add_key(alias, alias)
`},
		{"forin_append_ignored", `items = [1, 2]
total = 0
for item in items { total += item
 append(items, 9)
}
add_key(total, total)
add_key(items, items)
`},
		{"forin_dynamic", `add_key(before, true)
total = 0
for item in duration { total += 1 }
add_key(total, total)
add_key(after, true)
`},
		{"forin_existing_local", `item = 99
for item in [1, 2, 3] { add_key(last, item) }
add_key(final_item, item)
`},
		{"forin_existing_field", `for sequence in [7, 8] { add_key(last, sequence) }
add_key(final_sequence, sequence)
`},
		{"forin_empty", `item = 99
for item in [] { add_key(unexpected, true) }
add_key(final_item, item)
`},
		{"forin_rebind", `items = [1, 2, 3]
total = 0
for item in items { total += item
 items = [99]
}
add_key(total, total)
`},
		{"forin_nested_alias", `items = [[1], [2]]
total = 0
for item in items { total += item[0]
 items[1][0] = 9
}
add_key(total, total)
`},
		{"forin_live", `total = 0
items = [1, 2]
for item in items { total += item
 items[1] = 9
}
add_key(total, total)
`},
		{"forin_unicode", `total = 0
for ch in "a中😀" { total += 1 }
add_key(total, total)
`},
		{"forin_nested", `total = 0
for item in [1, 2, 3] {
 for n in [1, 2, 3, 4] {
  if n == 2 { continue }
  if n == 4 { break }
  total += item
 }
}
add_key(total, total)
`},
		{"forin_map", `total = 0
values = {"a": 2, "b": 4}
for key in values { total += values[key] }
add_key(total, total)
`},
		{"forin_error", `add_key(before, true)
for item in [1, 2] { value = item / sequence
 add_key(progress, item)
}
add_key(after, true)
`},
		{"for_counted", `total = 0
for i = 0; i < sequence % 7; i += 1 { total += i }
add_key(total, total)
add_key(scope_i, i)
`},
		{"for_nested", `total = 0
for i = 0; i < 3; i += 1 {
 for j = 0; j < 2; j += 1 { total += 1 }
}
add_key(total, total)
`},
		{"for_error_prefix", `add_key(before, true)
for i = 0; i < 3; i += 1 {
 value = i / sequence
 add_key(progress, i)
}
add_key(after, true)
`},
		{"for_exit", `for i = 0; i < 3; i += 1 {
 add_key(progress, i)
 if i == 1 { exit() }
}
add_key(after, true)
`},
		{"created_points", `create_point("first", {"source": "diff"}, {"value": sequence}, ts=1)
create_point("second", {}, {"message": message}, category="L", ts=2)
create_point("third", {}, {"value": sequence}, ts=3)
`},
		{"created_then_drop", `create_point("child", {}, {"value": sequence}, ts=1)
drop()
create_point("after_drop", {}, {"value": sequence}, ts=2)
`},
		{"created_then_error", `create_point("child", {}, {"value": sequence}, ts=1)
value = 1 / sequence
create_point("after_divide", {}, {"value": value}, ts=2)
`},
		{"cache_across_batches", `if sequence == 0 { cache_set("strict-diff-cache", "seed") }
previous = cache_get("strict-diff-cache")
add_key(previous, previous)
cache_set("strict-diff-cache", message)
add_key(done, true)
`},
	}
	for _, test := range cases {
		for _, batchSize := range []int{1, 8, 128} {
			t.Run(fmt.Sprintf("%s/batch-%d", test.name, batchSize), func(t *testing.T) {
				runner, err := pljit.NewRunnerWithHost(path, "pipeline-go-1.4.3-datakit", 8, pljit.NewPipelineGoHost(nil))
				if err != nil {
					t.Fatal(err)
				}
				defer runner.Close()
				script, err := NewPlScriptSimple(point.Logging, test.name+".p", test.source)
				if err != nil {
					t.Fatalf("Go compile: %v", err)
				}
				if err := runner.Prepare(test.source); err != nil {
					t.Fatalf("native compile rejected: %v", err)
				}
				projection, err := runner.Projection(test.source)
				if err != nil {
					t.Fatal(err)
				}
				const count = 129
				actual := make([]*point.Point, count)
				expected := make([]pointRun, count)
				oracleErrors := make([]error, count)
				messages := []string{`{"enabled":true,"value":"HELLO"}`, `{"enabled":false}`, `{}`, `not-json`, `ERROR:42`, `ok:7`, ``}
				for i := 0; i < count; i++ {
					var duration any = int64(123456789)
					if i%3 == 1 {
						duration = "invalid"
					}
					fields := map[string]any{"message": messages[i%len(messages)], "sequence": int64(i), "temporary": "remove", "sentinel": "keep", "duration": duration}
					actual[i] = newRealScriptPoint(test.name, fields)
					expected[i] = pointRun{point: newRealScriptPoint(test.name, fields), script: script}
					// Call Go once and retain its error. runPipelineGo logs and
					// discards that error, which can hide native termination drift.
					wrapped := ptinput.PtWrap(point.Logging, expected[i].point)
					oracleErrors[i] = script.Run(wrapped, nil, nil)
					// DataKit publishes subpoints only on successful script completion.
					if oracleErrors[i] == nil {
						for _, child := range wrapped.GetSubPoint() {
							if !child.Dropped() {
								if expected[i].created == nil {
									expected[i].created = make(map[point.Category][]*point.Point)
								}
								expected[i].created[child.Category()] = append(expected[i].created[child.Category()], child.Point())
							}
						}
					}
					expected[i].dropped = wrapped.Dropped()
					expected[i].output = wrapped.Point()
				}
				for start := 0; start < count; start += batchSize {
					end := min(start+batchSize, count)
					input, err := encodeProjectedJITPoints(point.Logging, actual[start:end], projection)
					if err != nil {
						t.Fatal(err)
					}
					batch, err := runner.Process(test.source, input)
					if err != nil {
						t.Fatalf("native process failed (no replay): %v", err)
					}
					if len(batch.Records) != end-start {
						t.Fatal("native record count mismatch")
					}
					for offset, record := range batch.Records {
						i := start + offset
						emitted, err := decodeJITEmitted(record.Emitted)
						if err != nil {
							t.Fatalf("record %d emitted decode: %v", i, err)
						}
						compareCreated := func(created map[point.Category][]*point.Point) {
							t.Helper()
							if oracleErrors[i] != nil {
								return // Production suppresses side outputs on terminal error.
							}
							for category, children := range emitted {
								if created == nil {
									created = make(map[point.Category][]*point.Point)
								}
								created[category] = append(created[category], children...)
							}
							if len(created) != len(expected[i].created) {
								t.Fatalf("record %d child categories differ: native=%v Go=%v", i, created, expected[i].created)
							}
							for category, want := range expected[i].created {
								got := created[category]
								if len(got) != len(want) {
									t.Fatalf("record %d category %v child count: native=%d Go=%d", i, category, len(got), len(want))
								}
								for j := range want {
									if equal, reason := got[j].EqualWithReason(want[j]); !equal {
										t.Fatalf("record %d category %v child %d differs: %s", i, category, j, reason)
									}
								}
							}
						}
						if (record.Status == pljit.TerminalError) != (oracleErrors[i] != nil) {
							t.Fatalf("record %d error outcome differs: native status=%v error=%s; Go=%v", i, record.Status, record.Error, oracleErrors[i])
						}
						if oracleErrors[i] == nil {
							if (record.Status == pljit.TerminalDropped) != expected[i].dropped {
								t.Fatalf("record %d drop differs: native=%v Go=%v", i, record.Status, expected[i].dropped)
							}
							if expected[i].dropped {
								compareCreated(nil)
								if record.HasMutations() {
									t.Fatalf("dropped record %d unexpectedly has mutations", i)
								}
								continue // No surviving point; subsequent records still verify shared state.
							}
						}
						if record.Status != pljit.TerminalOK && !(record.Status == pljit.TerminalError && record.CommitPrefixError) {
							t.Fatalf("native terminal at %d (no replay): status=%v error=%s", i, record.Status, record.Error)
						}
						var dropped bool
						var created map[point.Category][]*point.Point
						if batch.Static != nil {
							created, dropped, err = applyJITStatic(point.Logging, actual[i], batch.Static, offset, nil)
						} else {
							created, dropped, err = applyJITRecord(point.Logging, actual[i], record, uint64(offset), nil)
						}
						if err != nil {
							t.Fatal(err)
						}
						if dropped != expected[i].dropped {
							t.Fatalf("record %d applied drop differs: native=%v Go=%v", i, dropped, expected[i].dropped)
						}
						compareCreated(created)
						if expected[i].output == nil {
							t.Fatalf("Go unexpectedly dropped record %d", i)
						}
						if equal, reason := actual[i].EqualWithReason(expected[i].output); !equal {
							t.Fatalf("record %d differs: %s\nJIT: %#v\nGo: %#v", i, reason, actual[i].KVMap(), expected[i].output.KVMap())
						}
					}
				}
			})
		}
	}
}
