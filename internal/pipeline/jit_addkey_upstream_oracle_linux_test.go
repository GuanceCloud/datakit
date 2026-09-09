// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build pipeline_jit && linux && (amd64 || arm64)

package pipeline

// Adapted from pipeline-go v1.4.3 ptinput/funcs/fn_addkey_test.go.
// MIT License; Copyright 2021-present Guance, Inc.
import (
	"github.com/GuanceCloud/cliutils/point"
	"github.com/GuanceCloud/pipeline-go/ptinput"
	"github.com/GuanceCloud/pipeline-go/ptinput/funcs"
	"github.com/GuanceCloud/platypus/pkg/engine"
	pljit "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline/jit"
	"os"
	"reflect"
	"testing"
	"time"
)

func TestJITAddKeyUpstreamOracle(t *testing.T) {
	const input = `1.2.3.4 - - [29/Nov/2021:07:30:50 +0000] "POST /?signature=b8d8ea&timestamp=1638171049 HTTP/1.1" 200 413 "-" "Mozilla/4.0"`
	const grok = `grok(_, "%{IPORHOST:client_ip} %{NOTSPACE} %{NOTSPACE} \\[%{HTTPDATE:time}\\] \"%{DATA} %{GREEDYDATA} HTTP/%{NUMBER}\" %{INT:status_code} %{INT:bytes}")` + "\n"
	table := []struct {
		name, source, input string
		want                any
		export, invalid     bool
	}{
		{"string", `add_key(add_new_key,"shanghai")`, input, "shanghai", false, false},
		{"int", grok + `add_key(add_new_key,-1)`, input, int64(-1), false, false},
		{"float", grok + `add_key(add_new_key,1.)`, input, float64(1), false, false},
		{"invalid_float", grok + `add_key(add_new_key,.1)`, input, float64(.1), false, true},
		{"true", grok + `add_key(add_new_key,true)`, input, true, false, false},
		{"mixed_case_true", grok + `add_key(add_new_key,tRue)`, input, true, false, false},
		{"false", grok + `add_key(add_new_key,false)`, input, false, false, false},
		{"nil", grok + `add_key(add_new_key,nil)`, input, nil, false, false},
		{"list", `add_key(add_new_key,[1,2])`, "test", "[1,2]", true, false},
		{"map", `add_key(add_new_key,{"a":1,"b":"x"})`, "test", `{"a":1,"b":"x"}`, true, false},
	}
	path := os.Getenv("PLATYPUS_JIT_RUNTIME")
	if path == "" {
		t.Fatal("requires real runtime")
	}
	runner, err := pljit.NewRunner(path, "pipeline-go-1.4.3-datakit", 32)
	if err != nil {
		t.Fatal(err)
	}
	defer runner.Close()
	var cases []jsonOracleCase
	for _, tc := range table {
		t.Run(tc.name+"/original_assertions", func(t *testing.T) {
			scripts, errs := engine.ParseScript(map[string]string{"upstream.p": tc.source}, funcs.FuncsMap, funcs.FuncsCheckMap)
			if tc.invalid {
				if len(errs) == 0 {
					t.Fatal("Go accepted original invalid float")
				}
				if runner.Check(tc.source).Route == pljit.RouteJITNative {
					t.Fatal("JIT accepted original invalid float")
				}
				return
			}
			if len(errs) != 0 {
				t.Fatal(errs)
			}
			pt := ptinput.NewPlPt(point.Logging, "test", nil, map[string]any{"message": tc.input}, time.Unix(1700000000, 123))
			if err := scripts["upstream.p"].Run(pt, nil); err != nil {
				t.Fatal(err)
			}
			value, _, err := pt.Get("add_new_key")
			if err != nil || !reflect.DeepEqual(value, tc.want) {
				t.Fatalf("value=%#v want=%#v err=%v", value, tc.want, err)
			}
			if tc.export && (!reflect.DeepEqual(pt.Fields()["add_new_key"], tc.want) || !reflect.DeepEqual(pt.Point().KVs().InfluxFields()["add_new_key"], tc.want)) {
				t.Fatal("export mismatch")
			}
		})
		if !tc.invalid {
			cases = append(cases, jsonOracleCase{name: tc.name, source: tc.source, input: tc.input, key: "add_new_key", want: tc.want})
		}
	}
	runJSONOracle(t, cases)
}

func TestJITNumericSpellingCheckingOracle(t *testing.T) {
	path := os.Getenv("PLATYPUS_JIT_RUNTIME")
	if path == "" {
		t.Fatal("requires real runtime")
	}
	runner, err := pljit.NewRunner(path, "pipeline-go-1.4.3-datakit", 32)
	if err != nil {
		t.Fatal(err)
	}
	defer runner.Close()
	for _, tc := range []struct {
		source string
		valid  bool
	}{
		{`add_key(n,.1)`, false}, {`add_key(n,-.1)`, false}, {`add_key(n,+.1)`, false},
		{`add_key(n,(.1))`, false}, {`add_key(n,-(.1))`, false}, {`n=[.1]; add_key(n,n)`, false},
		{`add_key(n,0.1)`, true}, {`add_key(n,-0.1)`, true}, {`add_key(n,1.)`, true},
		{`add_key(n,1.e2)`, true}, {`add_key(n,".1")`, true}, {"# .1\nadd_key(n,1)", true},
	} {
		t.Run(tc.source, func(t *testing.T) {
			_, errs := engine.ParseScript(map[string]string{"numeric.p": tc.source}, funcs.FuncsMap, funcs.FuncsCheckMap)
			if (len(errs) == 0) != tc.valid {
				t.Fatalf("Go valid=%v want=%v errors=%v", len(errs) == 0, tc.valid, errs)
			}
			check := runner.Check(tc.source)
			if (check.Route == pljit.RouteJITNative) != tc.valid {
				t.Fatalf("JIT check=%+v want valid=%v", check, tc.valid)
			}
		})
	}
}

func TestJITAddKeyMessageMapUpstreamOracle(t *testing.T) {
	const source = `result_msg = {}
all_keys = pt_kvs_keys()
for k in all_keys {
 if k != "message" {
  result_msg[k] = pt_kvs_get(k)
 }
}
add_key("message", result_msg)`
	const want = `{"a":"x","b":1,"status":"info"}`
	scripts, errs := engine.ParseScript(map[string]string{"upstream.p": source}, funcs.FuncsMap, funcs.FuncsCheckMap)
	if len(errs) != 0 {
		t.Fatal(errs)
	}
	pt := ptinput.NewPlPt(point.Logging, "test", nil, map[string]any{"message": "test", "a": "x", "b": int64(1)}, time.Unix(1700000000, 123))
	if err := scripts["upstream.p"].Run(pt, nil); err != nil {
		t.Fatal(err)
	}
	value, _, err := pt.Get("message")
	if err != nil || value != want || pt.Fields()["message"] != want || pt.Point().KVs().InfluxFields()["message"] != want {
		t.Fatalf("original exports: value=%#v fields=%#v influx=%#v err=%v", value, pt.Fields(), pt.Point().KVs().InfluxFields(), err)
	}
	// Keep the exact script. The shared differential harness adds sentinel and
	// before/after fields, so compare the whole Go result rather than reusing
	// the narrower original JSON expectation for this enriched input.
	runJSONOracle(t, []jsonOracleCase{{name: "original_script", source: source, input: "test"}}, func(pt *point.Point) {
		pt.AddKVs(point.NewKV("a", "x"), point.NewKV("b", int64(1)))
	})
}
