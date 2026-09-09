// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build pipeline_jit && linux && (amd64 || arm64)

package pipeline

import (
	"crypto/sha256"
	"fmt"
	"os"
	"testing"

	"github.com/GuanceCloud/cliutils/point"
	"github.com/GuanceCloud/pipeline-go/ptinput"
	pljit "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline/jit"
)

const jsonStringChainSource = `
doc=load_json(message)
items=doc["items"]
method=doc["method"]
results=[]
for i=0; i<len(items); i=i+1 {
  text=items[i]
  pt_kvs_set("started",i)
  if text != nil {
    piece=slice_string(end=strlen(text),name=text,start=0)
    digest=hash(method=method,text=piece)
    results=append(results,digest)
    pt_kvs_set("results",results)
  }
  pt_kvs_set("completed",i)
}
add_key(count,len(results))
`

func TestJITJSONStringChainOracle(t *testing.T) {
	const source = jsonStringChainSource
	digest := func(s string) string { return fmt.Sprintf("%x", sha256.Sum256([]byte(s))) }
	runJSONOracle(t, []jsonOracleCase{
		{name: "unicode_and_empty", source: source, input: `{"items":["你好","","abc"],"method":"sha256"}`, key: "results", want: fmt.Sprintf(`["%s","%s","%s"]`, digest("你好"), digest(""), digest("abc"))},
		{name: "empty_list", source: source, input: `{"items":[],"method":"sha256"}`, key: "count", want: int64(0)},
		{name: "skip_nil", source: source, input: `{"items":[null,"abc"],"method":"sha256"}`, key: "count", want: int64(1)},
		{name: "unsupported_algorithm", source: source, input: `{"items":["abc","def"],"method":"SHA256"}`, key: "results", want: `["",""]`},
		{name: "middle_type_error", source: source, input: `{"items":["abc",42,"def"],"method":"sha256"}`, key: "results", want: fmt.Sprintf(`["%s"]`, digest("abc")), fails: true},
		{name: "method_type_error", source: source, input: `{"items":["abc"],"method":42}`, key: "started", want: int64(0), fails: true},
	})
}

func TestJITJSONStringChainMixedBatchOracle(t *testing.T) {
	runJSONStringChainMixedBatchOracle(t, false)
}

func TestJITJSONStringChainConcurrentBatchOracle(t *testing.T) {
	runJSONStringChainMixedBatchOracle(t, true)
}

func runJSONStringChainMixedBatchOracle(t *testing.T, concurrent bool) {
	t.Helper()
	path := os.Getenv("PLATYPUS_JIT_RUNTIME")
	if path == "" {
		t.Fatal("requires real PLATYPUS_JIT_RUNTIME")
	}
	runner, err := pljit.NewRunner(path, "pipeline-go-1.4.3-datakit", 16)
	if err != nil {
		t.Fatal(err)
	}
	// Parallel subtests finish after the parent function returns.
	t.Cleanup(func() { runner.Close() })
	const source = "add_key(before,true)\n" + jsonStringChainSource + "\nadd_key(after,true)"
	if check := runner.Check(source); check.Route != pljit.RouteJITNative {
		t.Fatalf("not native: %+v", check)
	}
	projection, err := runner.Projection(source)
	if err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		input string
		fails bool
	}{
		{`{"items":["abc",42,"def"],"method":"sha256"}`, true},
		{`{"items":[],"method":"sha256"}`, false},
		{`{"items":["你好",""],"method":"sha256"}`, false},
		{`{"items":["abc"],"method":42}`, true},
		{`{"items":[null,"abc"],"method":"sha256"}`, false},
	}
	for _, count := range []int{1, 2, 4, 8, 10, 128} {
		for offset := range cases {
			count, offset := count, offset
			t.Run(fmt.Sprintf("batch%d/offset%d", count, offset), func(t *testing.T) {
				// Each oracle has independent interpreter state; only JIT runner
				// and its immutable projection are shared by concurrent calls.
				script, err := NewPlScriptSimple(point.Logging, "mixed-chain.p", source)
				if err != nil {
					t.Fatal(err)
				}
				if concurrent {
					t.Parallel()
				}
				actual, expected := make([]*point.Point, count), make([]*point.Point, count)
				for i := range actual {
					tc := cases[(i+offset)%len(cases)]
					fields := map[string]any{"message": tc.input, "sentinel": fmt.Sprintf("%d/%d/%d", count, offset, i)}
					actual[i], expected[i] = newRealScriptPoint("mixed", fields), newRealScriptPoint("mixed", fields)
					goErr := script.Run(ptinput.PtWrap(point.Logging, expected[i]), nil, nil)
					if (goErr != nil) != tc.fails {
						t.Fatalf("record%d Go error=%v", i, goErr)
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
				if len(batch.Records) != count {
					t.Fatalf("records=%d want%d", len(batch.Records), count)
				}
				for i, record := range batch.Records {
					fails := cases[(i+offset)%len(cases)].fails
					wantStatus := pljit.TerminalOK
					if fails {
						wantStatus = pljit.TerminalError
					}
					if record.Status != wantStatus || (fails && !record.CommitPrefixError) {
						t.Fatalf("record%d status=%v prefix=%v", i, record.Status, record.CommitPrefixError)
					}
					if batch.Static != nil {
						_, _, err = applyJITStatic(point.Logging, actual[i], batch.Static, i, nil)
					} else {
						_, _, err = applyJITRecord(point.Logging, actual[i], record, uint64(i), nil)
					}
					if err != nil {
						t.Fatal(err)
					}
					if !fails {
						ptinput.PtWrap(point.Logging, actual[i]).KeyTime2Time()
					}
					if equal, reason := equalJSONOraclePoints(actual[i], expected[i]); !equal {
						t.Fatalf("record%d: %s\nactual=%s\nexpected=%s", i, reason, actual[i].Pretty(), expected[i].Pretty())
					}
				}
			})
		}
	}
}
