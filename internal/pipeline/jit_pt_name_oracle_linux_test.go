// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build pipeline_jit && linux && (amd64 || arm64)

package pipeline

import (
	"os"
	"testing"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"github.com/GuanceCloud/pipeline-go/ptinput"
	"github.com/GuanceCloud/pipeline-go/ptinput/funcs"
	"github.com/GuanceCloud/platypus/pkg/engine"
	pljit "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline/jit"
)

func TestJITSetMeasurementCheckingOracle(t *testing.T) {
	path := os.Getenv("PLATYPUS_JIT_RUNTIME")
	if path == "" {
		t.Fatal("requires real runtime")
	}
	runner, err := pljit.NewRunner(path, "pipeline-go-1.4.3-datakit", 32)
	if err != nil {
		t.Fatal(err)
	}
	defer runner.Close()
	for _, source := range []string{
		`set_measurement(obj["name"])`, `set_measurement(obj[0],true)`,
		`set_measurement(message,delete_key=true)`, `set_measurement(name="new")`,
		`set_measurement(1)`, `set_measurement(pt_name())`,
	} {
		t.Run(source, func(t *testing.T) {
			_, errs := engine.ParseScript(map[string]string{"check.p": source}, funcs.FuncsMap, funcs.FuncsCheckMap)
			if len(errs) == 0 {
				t.Fatal("upstream accepted source expected to fail")
			}
			if check := runner.Check(source); check.Route == pljit.RouteJITNative {
				t.Fatalf("JIT accepted invalid Go source: %+v", check)
			}
		})
	}
	if check := runner.Check(`flag=true; set_measurement(message,flag)`); check.Route != pljit.RouteJITNative {
		t.Fatalf("dynamic delete_key extension rejected: %+v", check)
	}
}

// MIT, Guance Inc.; both TestSetMeasurement cases from the upstream file
// fn_set_mesaurement_test.go (the filename spelling is intentional).
func TestJITSetMeasurementUpstreamOracle(t *testing.T) {
	const input = `162.62.81.1 - - [29/Nov/2021:07:30:50 +0000] "123 /?signature=b8d8ea&timestamp=1638171049 HTTP/1.1" 200 413 "-" "Mozilla/4.0"`
	const grok = `grok(_, "%{IPORHOST:client_ip} %{NOTSPACE} %{NOTSPACE} \\[%{HTTPDATE:time}\\] \"%{DATA:data} %{GREEDYDATA} HTTP/%{NUMBER}\" %{INT:status_code} %{INT:bytes}")`
	var cases []jsonOracleCase
	for _, tc := range []struct {
		name, source string
		deleted      bool
	}{
		{"set_measurement 0", "set_tag(client_ip)\n" + grok + "\ncast(data,\"int\")\nset_measurement(client_ip)", false},
		{"set_measurement 1", grok + "\nset_tag(client_ip)\ncast(data,\"int\")\nset_measurement(client_ip,true)", true},
	} {
		t.Run(tc.name+"/original_assertions", func(t *testing.T) {
			scripts, errs := engine.ParseScript(map[string]string{"upstream.p": tc.source}, funcs.FuncsMap, funcs.FuncsCheckMap)
			if len(errs) != 0 {
				t.Fatal(errs)
			}
			pt := ptinput.NewPlPt(point.Logging, "test", nil, map[string]any{"message": input}, time.Unix(1700000000, 123))
			if err := scripts["upstream.p"].Run(pt, nil); err != nil {
				t.Fatal(err)
			}
			_, _, err := pt.Get("client_ip")
			if (err != nil) != tc.deleted {
				t.Fatalf("deleted=%v error=%v", tc.deleted, err)
			}
			if pt.GetPtName() != "162.62.81.1" {
				t.Fatalf("measurement=%q", pt.GetPtName())
			}
		})
		cases = append(cases, jsonOracleCase{name: tc.name, source: tc.source, input: input})
	}
	runJSONOracle(t, cases, func(pt *point.Point) { ptinput.PtWrap(point.Logging, pt).SetPtName("test") })
}

// Adapted from pipeline-go v1.4.3 ptinput/funcs/fn_pt_name_test.go.
// MIT License; Copyright 2021-present Guance, Inc.
// Preserve all five scripts, including a=1 followed by the distinct aa name.
func TestJITPointNameUpstreamOracle(t *testing.T) {
	const input = `162.62.81.1 - - [29/Nov/2021:07:30:50 +0000] "123 /?signature=b8d8ea&timestamp=1638171049 HTTP/1.1" 200 413 "-" "Mozilla/4.0"`
	const grok = `grok(_, "%{IPORHOST:client_ip} %{NOTSPACE} %{NOTSPACE} \\[%{HTTPDATE:time}\\] \"%{DATA:data} %{GREEDYDATA} HTTP/%{NUMBER}\" %{INT:status_code} %{INT:bytes}")`
	table := []struct{ name, prefix, tail, want string }{
		{"pt_name_set_and_get", "set_tag(client_ip)\n" + grok, `add_key("pt_name",pt_name(client_ip))`, "162.62.81.1"},
		{"pt_name_set_str_and_get", "set_tag(client_ip)\n" + grok, `add_key("pt_name",pt_name("aa"))`, "aa"},
		{"pt_name_set_int_1_and_get", "set_tag(client_ip)\n" + grok, `add_key("pt_name",pt_name(1))`, "default"},
		{"pt_name_set_int_var_and_get", "set_tag(client_ip)\n" + grok, "a=1\n" + `add_key("pt_name",pt_name(aa))`, "default"},
		{"pt_name_get", grok + "\nset_tag(client_ip)", `add_key("pt_name",pt_name())`, "default"},
	}
	var cases []jsonOracleCase
	for _, tc := range table {
		source := tc.prefix + "\ncast(data,\"int\")\n" + tc.tail
		t.Run(tc.name+"/original_assertions", func(t *testing.T) {
			scripts, errs := engine.ParseScript(map[string]string{"upstream.p": source}, funcs.FuncsMap, funcs.FuncsCheckMap)
			if len(errs) != 0 {
				t.Fatal(errs)
			}
			pt := ptinput.NewPlPt(point.Logging, "default", nil, map[string]any{"message": input}, time.Unix(1700000000, 123))
			if err := scripts["upstream.p"].Run(pt, nil); err != nil {
				t.Fatal(err)
			}
			if _, _, err := pt.Get("pt_name"); err != nil {
				t.Fatal(err)
			}
			if pt.GetPtName() != tc.want {
				t.Fatalf("name=%q want=%q", pt.GetPtName(), tc.want)
			}
		})
		cases = append(cases, jsonOracleCase{name: tc.name, source: source, input: input, key: "pt_name", want: tc.want})
	}
	runJSONOracle(t, cases, func(pt *point.Point) { ptinput.PtWrap(point.Logging, pt).SetPtName("default") })
}

// Supplementary cases derived from pipeline-go v1.4.3 fn_pt_name.go.
func TestJITPointNameOracle(t *testing.T) {
	runJSONOracle(t, []jsonOracleCase{
		{name: "named_assignment", source: `add_key(result,pt_name(chosen="renamed")); add_key(chosen,chosen)`, key: "chosen", want: "renamed"},
		{name: "named_argument_error", source: `chosen="old"; add_key(result,pt_name(chosen=1/divisor)); add_key(chosen,chosen)`, key: "chosen", want: "old"},
		{name: "read", source: `add_key(result,pt_name())`, key: "result", want: "valid"},
		{name: "set", source: `add_key(result,pt_name("renamed"))`, key: "result", want: "renamed"},
		{name: "integer_ignored", source: `pt_name(7); add_key(result,pt_name())`, key: "result", want: "valid"},
		{name: "nil_ignored", source: `pt_name(nil); add_key(result,pt_name())`, key: "result", want: "valid"},
		{name: "list_ignored", source: `add_key(result,pt_name([1,2]))`, key: "result", want: "valid"},
		{name: "nested_updates", source: `add_key(result,pt_name(pt_name("inner")))`, key: "result", want: "inner"},
		{name: "short_circuit", source: `result=false && pt_name("unused") == "unused"; add_key(result,pt_name())`, key: "result", want: "valid"},
		{name: "argument_error", source: `add_key(result,pt_name(1/divisor))`, key: "result", want: "valid"},
		{name: "error_after_effect", source: `add_key(result,pt_name(pt_kvs_set("side",true) && 1/divisor == 0))`, key: "side", want: true},
	})
}

func TestJITRawMeasurementOracle(t *testing.T) {
	rawBefore := "metric\xff\x00"
	rawAfter := "renamed\xfe\x00"
	runJSONOracle(t, []jsonOracleCase{{
		name:   "raw_measurement_read_and_field_mutation",
		source: `add_key(before,pt_name()); set_measurement(message); add_key(after,pt_name())`,
		input:  rawAfter,
	}}, func(pt *point.Point) {
		ptinput.PtWrap(point.Logging, pt).SetPtName(rawBefore)
	})
}

func TestJITSetMeasurementOracle(t *testing.T) {
	runJSONOracle(t, []jsonOracleCase{
		{name: "attribute_bracket_name", source: "add_key(\"obj.[name\",\"wrong\"); set_measurement(obj.`[name`,true); add_key(result,pt_name())", key: "result", want: "valid"},
		{name: "attribute_dollar_name", source: "add_key(\"obj.$name\",\"wrong\"); set_measurement(obj.`$name`,true); add_key(result,pt_name())", key: "result", want: "valid"},
		{name: "local_shadow", source: `message="local"; set_measurement(message,true); add_key(result,pt_name()); add_key(local,message)`, input: "field", key: "result", want: "local"},
		{name: "attribute", source: `obj={"name":"nested"}; set_measurement(obj.name,true); add_key(result,pt_name())`, key: "result", want: "valid"},
		{name: "attribute_error", source: `obj=7; set_measurement(obj.name,true); add_key(result,pt_name())`, key: "result", want: "valid"},
		{name: "dotted_field", source: `add_key("obj.name","dotted"); set_measurement(obj.name,true); add_key(result,pt_name())`, key: "result", want: "valid"},
		{name: "field", source: `set_measurement(message); add_key(result,pt_name())`, input: "renamed", key: "result", want: "renamed"},
		{name: "delete", source: `set_measurement(message,true); add_key(result,pt_name())`, input: "renamed", key: "message", want: nil},
		{name: "origin_alias", source: `set_measurement(_,true); add_key(result,pt_name())`, input: "renamed", key: "message", want: nil},
		{name: "literal_not_deleted", source: `set_measurement("message",true); add_key(result,pt_name())`, input: "original", key: "message", want: "original"},
		{name: "numeric_deleted", source: `add_key(source,7); set_measurement(source,true); add_key(result,pt_name())`, key: "result", want: "valid"},
		{name: "missing", source: `set_measurement(absent,true); add_key(result,pt_name())`, key: "result", want: "valid"},
		{name: "nested", source: `add_key(result,set_measurement(message,true) == nil)`, input: "renamed", key: "result", want: true},
		{name: "short_circuit", source: `result=false && set_measurement(message,true) == nil; add_key(result,pt_name())`, input: "renamed", key: "result", want: "valid"},
	})
}
