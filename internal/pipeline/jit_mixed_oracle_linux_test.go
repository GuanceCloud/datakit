// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build pipeline_jit && linux && (amd64 || arm64)

package pipeline

import (
	"fmt"
	"os"
	"reflect"
	"testing"

	"github.com/GuanceCloud/cliutils/point"
	"github.com/GuanceCloud/pipeline-go/ptinput"
	pljit "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline/jit"
)

func TestJITMixedRecordErrorsAndReuse(t *testing.T) {
	path := os.Getenv("PLATYPUS_JIT_RUNTIME")
	if path == "" {
		t.Fatal("requires real PLATYPUS_JIT_RUNTIME")
	}
	runner, err := pljit.NewRunner(path, "pipeline-go-1.4.3-datakit", 16)
	if err != nil {
		t.Fatal(err)
	}
	defer runner.Close()
	for _, tc := range []struct{ name, body, good, bad string }{
		{"base64", `b64dec(message)`, "aGVsbG8=", "aGVsbG8=invalid"},
		{"url", `url_decode(message)`, "hello%20world", "hello%20world%GG"},
		{"json", `obj=load_json(message); add_key(parsed,obj)`, `{"x":1}`, `{"x":`},
		{"grok-then-error", `grok(message,"^%{WORD:word}(?: %{WORD:optional})?$"); result=1/divisor`, "hello world", "hello"},
		{"cache-then-error", `previous=cache_get("mixed-error-state"); if previous != nil { add_key(previous,previous) }; cache_set("mixed-error-state",message); result=1/divisor`, "success-state", "failed-state"},
		{"nested-cache-then-error", `previous=cache_get("mixed-nested-state"); if previous != nil { add_key(previous,previous) }; result=(cache_set("mixed-nested-state",message) == nil) && 1/divisor == 1`, "success-state", "failed-state"},
		{"nested-cache-invalid-expiry", `previous=cache_get("mixed-nested-expiry"); if previous != nil { add_key(previous,previous) }; expiry=100; if divisor == 0 { expiry="100" }; add_key(result,cache_set("mixed-nested-expiry",message,expiry) == nil)`, "success-state", "must-not-be-stored"},
		{"cache-invalid-expiry", `previous=cache_get("mixed-expiry-state"); if previous != nil { add_key(previous,previous) }; expiry=100; if divisor == 0 { expiry="100" }; cache_set("mixed-expiry-state",message,expiry)`, "success-state", "must-not-be-stored"},
		{"cache-list-then-error", `previous=cache_get("mixed-list"); if previous != nil { add_key(previous,previous) }; a=[identity]; cache_set("mixed-list",a); a[0]=message; result=1/divisor`, "success-list", "failed-list"},
		{"cache-map-then-error", `previous=cache_get("mixed-map"); if previous != nil { pt_kvs_set("previous",previous) }; a={"value":identity}; cache_set("mixed-map",a); a["value"]=message; result=1/divisor`, "success-map", "failed-map"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			source := "add_key(before,true); " + tc.body + "; add_key(after,true)"
			goScript, err := NewPlScriptSimple(point.Logging, "mixed.p", source)
			if err != nil {
				t.Fatal(err)
			}
			if check := runner.Check(source); check.Route != pljit.RouteJITNative {
				t.Fatalf("not native: %+v", check)
			}
			projection, err := runner.Projection(source)
			if err != nil {
				t.Fatal(err)
			}
			var previousCacheValue any
			for _, size := range []int{1, 2, 4, 8, 10, 128} {
				for round := 0; round < 4; round++ {
					actual, expected := make([]*point.Point, size), make([]*point.Point, size)
					failed := make([]bool, size)
					for i := range actual {
						bad := (i+round)%2 == 0
						message, divisor := tc.good, int64(1)
						if bad {
							message, divisor = tc.bad, 0
						}
						fields := map[string]any{"message": message, "divisor": divisor, "identity": fmt.Sprintf("%d/%d/%d", size, round, i), "optional": "old"}
						actual[i], expected[i] = newRealScriptPoint("mixed", fields), newRealScriptPoint("mixed", fields)
						failed[i] = goScript.Run(ptinput.PtWrap(point.Logging, expected[i]), nil, nil) != nil
						if tc.name == "cache-then-error" || tc.name == "cache-invalid-expiry" || tc.name == "nested-cache-then-error" || tc.name == "nested-cache-invalid-expiry" {
							if got := expected[i].Get("previous"); got != previousCacheValue {
								t.Fatalf("Go cache state got=%v want=%v", got, previousCacheValue)
							}
							if tc.name == "cache-then-error" || tc.name == "nested-cache-then-error" || !bad {
								previousCacheValue = message
							}
						}
						if tc.name == "cache-list-then-error" || tc.name == "cache-map-then-error" {
							if got := expected[i].Get("previous"); !reflect.DeepEqual(got, previousCacheValue) {
								t.Fatalf("Go container cache got=%#v want=%#v", got, previousCacheValue)
							}
							if tc.name == "cache-list-then-error" {
								previousCacheValue = []string{message}
							} else {
								previousCacheValue = fmt.Sprintf(`{"value":%q}`, message)
							}
						}
						// load_json returns nil for malformed JSON rather than failing.
						expectFailure := bad && tc.name != "json"
						if failed[i] != expectFailure {
							t.Fatalf("fixture error: record=%d bad=%v failed=%v", i, bad, failed[i])
						}
					}
					input, err := encodeProjectedJITPoints(point.Logging, actual, projection)
					if err != nil {
						t.Fatal(err)
					}
					batch, err := runner.Process(source, input)
					if err != nil {
						t.Fatal(err)
					}
					if len(batch.Records) != size {
						t.Fatal("record count mismatch")
					}
					for i, record := range batch.Records {
						wantStatus := pljit.TerminalOK
						if failed[i] {
							wantStatus = pljit.TerminalError
						}
						if record.Status != wantStatus || record.CommitPrefixError != failed[i] {
							t.Fatalf("size=%d round=%d record=%d terminal=%+v", size, round, i, record)
						}
						if batch.Static != nil {
							_, _, err = applyJITStatic(point.Logging, actual[i], batch.Static, i, nil)
						} else {
							_, _, err = applyJITRecord(point.Logging, actual[i], record, uint64(i), nil)
						}
						if err != nil {
							t.Fatal(err)
						}
						if equal, why := actual[i].EqualWithReason(expected[i]); !equal {
							t.Fatalf("size=%d round=%d record=%d: %s", size, round, i, why)
						}
					}
				}
			}
		})
	}
}
