// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build pipeline_jit && linux && (amd64 || arm64)

// Adapted from pipeline-go v1.4.3 ptinput/funcs/fn_hash_test.go.
// Copyright 2021-present Guance, Inc. MIT License; see testdata/PIPELINE_GO_TEST_LICENSE.txt.
package pipeline

import (
	"crypto/sha256"
	"fmt"
	"os"
	"testing"

	"github.com/GuanceCloud/cliutils/point"
	pljit "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline/jit"
)

func TestJITUpstreamHashOracle(t *testing.T) {
	var cases []jsonOracleCase
	for i, tc := range []struct{ method, want string }{
		{"md5", "900150983cd24fb0d6963f7d28e17f72"},
		{"xx", ""},
		{"sha1", "a9993e364706816aba3e25717850c26c9cd0d89d"},
		{"sha256", "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad"},
		{"sha512", "ddaf35a193617abacc417349ae20413112e6fa4e89a97ea20a9eeee64b55d39a2192992a274fc1a836ba3c23a3feebbd454d4423643ce80e2a9ac94fa54ca49f"},
	} {
		cases = append(cases, jsonOracleCase{name: fmt.Sprintf("%d-%s", i, tc.method),
			source: fmt.Sprintf("sum = hash(\"abc\", %q)\npt_kvs_set(\"result\", sum)", tc.method),
			key:    "result", want: tc.want})
	}
	runJSONOracle(t, cases)
}

func TestJITHashArgumentOrderOracle(t *testing.T) {
	runJSONOracle(t, []jsonOracleCase{
		{name: "positional_text_error", source: `src=divisor; hash(src,pt_kvs_set("side",1))`, key: "side", want: nil, fails: true},
		{name: "named_text_error", source: `src=divisor; hash(method=pt_kvs_set("side",1),text=src)`, key: "side", want: nil, fails: true},
		{name: "nested_text_error", source: `src=divisor; add_key(result,hash(method=pt_kvs_set("side",1),text=src))`, key: "side", want: nil, fails: true},
		{name: "text_effect_before_error", source: `hash(method=pt_kvs_set("later",2),text=pt_kvs_set("side",1))`, key: "side", want: int64(1), fails: true},
		{name: "method_effect_before_error", source: `hash(method=pt_kvs_set("side",1),text="abc")`, key: "side", want: int64(1), fails: true},
	})
}

func TestJITUpstreamHashInvalidArguments(t *testing.T) {
	path := os.Getenv("PLATYPUS_JIT_RUNTIME")
	if path == "" {
		t.Fatal("requires real runtime")
	}
	runner, err := pljit.NewRunner(path, "pipeline-go-1.4.3-datakit", 2)
	if err != nil {
		t.Fatal(err)
	}
	defer runner.Close()
	for i, call := range []string{`hash("abc", )`, `hash(method= "abc", )`} {
		t.Run(fmt.Sprintf("%d", i+5), func(t *testing.T) {
			source := "sum = " + call + "\npt_kvs_set(\"result\", sum)"
			if _, err := NewPlScriptSimple(point.Logging, "hash.p", source); err == nil {
				t.Fatal("expected Go checking rejection")
			}
			check := runner.Check(source)
			if check.Route != pljit.RoutePipelineGo || check.Reason != pljit.CheckReasonCompileError {
				t.Fatalf("expected JIT checking rejection: %+v", check)
			}
		})
	}
}

func TestJITHashCompositionOracle(t *testing.T) {
	const digest = "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad"
	input := "中文\x00abc"
	runJSONOracle(t, []jsonOracleCase{
		{name: "dynamic-method", source: `method = "sha256"; text = "abc"; sum = hash(text, method); pt_kvs_set("result", sum)`, key: "result", want: digest},
		{name: "named-reordered", source: `sum = hash(method="sha256", text="abc"); pt_kvs_set("result", sum)`, key: "result", want: digest},
		{name: "nested", source: `pt_kvs_set("result", hash("abc", "sha256"))`, key: "result", want: digest},
		{name: "branch-loop", source: `method = "md5"; for i=0; i<2; i=i+1 { if i == 1 { method = "sha256" }; pt_kvs_set("result", hash("abc", method)) }`, key: "result", want: digest},
		{name: "case-sensitive-method", source: `pt_kvs_set("result", hash("abc", "SHA256"))`, key: "result", want: ""},
		{name: "utf8-nul", source: `pt_kvs_set("result", hash(message, "sha256"))`, input: input, key: "result", want: fmt.Sprintf("%x", sha256.Sum256([]byte(input)))},
	})
}
