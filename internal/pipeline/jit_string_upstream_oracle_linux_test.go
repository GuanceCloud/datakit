// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build pipeline_jit && linux && (amd64 || arm64)

// Adapted from pipeline-go v1.4.3 TestSliceString and TestStrlen.
// Unless explicitly stated otherwise all files in that repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package pipeline

import (
	"os"
	"testing"

	"github.com/GuanceCloud/cliutils/point"
	pljit "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline/jit"
)

func TestJITStringUpstreamOracle(t *testing.T) {
	runJSONOracle(t, []jsonOracleCase{
		{name: "slice_normal1", source: `substring=slice_string("█汉字15384073392",0,5); pt_kvs_set("result",substring)`, key: "result", want: "█汉字15"},
		{name: "slice_normal2", source: `substring=slice_string("15384073392",5,10); pt_kvs_set("result",substring)`, key: "result", want: "07339"},
		{name: "slice_normal3", source: `substring=slice_string("abcdefghijklmnop",0,10); pt_kvs_set("result",substring)`, key: "result", want: "abcdefghij"},
		{name: "slice_negative", source: `substring=slice_string("abcdefghijklmnop",-1,10); pt_kvs_set("result",substring)`, key: "result", want: ""},
		{name: "slice_clamped", source: `substring=slice_string("abcdefghijklmnop",0,100); pt_kvs_set("result",substring)`, key: "result", want: "abcdefghijklmnop"},
		{name: "slice_panic_regression", source: `val="123你好123123123123123123123123123"; add_key("result",slice_string(val,0,len(val)))`, key: "result", want: "123你好123123123123123123123123123"},
		{name: "strlen_t1", source: `add_key("k1",strlen("你好"))`, key: "k1", want: int64(2)},
		{name: "strlen_t2", source: `add_key("k1",strlen("hello"))`, key: "k1", want: int64(5)},
		{name: "strlen_t3", source: `add_key("k1",strlen("你好hello"))`, key: "k1", want: int64(7)},
		{name: "strlen_t4", source: `v=[]; v=append(v,strlen("hello你好")); v=append(v,len("hello你好")); add_key("v",v)`, key: "v", want: "[7,11]"},
	})
}

// Supplementary cases: the upstream tables do not assert argument side effects.
func TestJITSliceStringNamedOrderOracle(t *testing.T) {
	runJSONOracle(t, []jsonOracleCase{
		{name: "source_failure", source: `src=divisor; slice_string(end=len(pt_kvs_set("side",1)),start=0,name=src)`, key: "side", want: nil, fails: true},
		{name: "start_failure", source: `idx=message; slice_string(end=len(pt_kvs_set("side",1)),name="abc",start=idx)`, input: "bad", key: "side", want: nil, fails: true},
		{name: "nested_start_failure", source: `idx=message; add_key(result,slice_string(end=len(pt_kvs_set("side",1)),name="abc",start=idx))`, input: "bad", key: "side", want: nil, fails: true},
		{name: "start_effect", source: `slice_string(end=len(pt_kvs_set("later",2)),name="abc",start=pt_kvs_set("side",1))`, key: "side", want: int64(1), fails: true},
		{name: "mixed_failure", source: `idx=message; slice_string("abc",end=len(pt_kvs_set("side",1)),start=idx)`, input: "bad", key: "side", want: nil, fails: true},
		{name: "reordered_success", source: `src="你好abc"; i=1; j=4; add_key(result,slice_string(end=j,start=i,name=src))`, key: "result", want: "好ab"},
	})
}

// Supplementary positional evaluation cases.
func TestJITSliceStringEvaluationOrderOracle(t *testing.T) {
	runJSONOracle(t, []jsonOracleCase{
		{name: "source_error_skips_start", source: `src=divisor; slice_string(src,len(pt_kvs_set("side",1)),2)`, key: "side", want: nil, fails: true},
		{name: "start_error_skips_end", source: `idx=message; slice_string("abc",idx,len(pt_kvs_set("side",1)))`, input: "bad", key: "side", want: nil, fails: true},
		{name: "start_effect_before_type_error", source: `slice_string("abc",pt_kvs_set("side",1),2)`, key: "side", want: int64(1), fails: true},
		{name: "nested_start_effect_before_type_error", source: `add_key(result,slice_string("abc",pt_kvs_set("side",1),2))`, key: "side", want: int64(1), fails: true},
		{name: "end_effect_before_type_error", source: `slice_string("abc",0,pt_kvs_set("side",2))`, key: "side", want: int64(2), fails: true},
		{name: "short_circuit", source: `if false { add_key(result,slice_string("abc",pt_kvs_set("side",1),2)) }; add_key(done,true)`, key: "side", want: nil},
	})
}

func TestJITSliceStringUpstreamCheckingOracle(t *testing.T) {
	path := os.Getenv("PLATYPUS_JIT_RUNTIME")
	if path == "" {
		t.Fatal("requires real PLATYPUS_JIT_RUNTIME")
	}
	runner, err := pljit.NewRunner(path, "pipeline-go-1.4.3-datakit", 16)
	if err != nil {
		t.Fatal(err)
	}
	defer runner.Close()
	for name, source := range map[string]string{
		"not_integer1": `substring=slice_string("abcdefghijklmnop","a","b"); pt_kvs_set("result",substring)`,
		"not_integer2": `substring=slice_string("abcdefghijklmnop","abc","def"); pt_kvs_set("result",substring)`,
		"not_string":   `substring=slice_string(12345,0,3); pt_kvs_set("result",substring)`,
		"too_few":      `substring=slice_string("abcdefghijklmnop",0); pt_kvs_set("result",substring)`,
		"too_many":     `substring=slice_string("abcdefghijklmnop",0,1,2); pt_kvs_set("result",substring)`,
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := NewPlScriptSimple(point.Logging, "slice-check.p", source); err == nil {
				t.Fatal("Go unexpectedly accepts fixture")
			}
			check := runner.Check(source)
			if check.Route != pljit.RoutePipelineGo || check.Reason != pljit.CheckReasonCompileError || check.Detail == "" {
				t.Fatalf("expected compile rejection: %+v", check)
			}
		})
	}
}
