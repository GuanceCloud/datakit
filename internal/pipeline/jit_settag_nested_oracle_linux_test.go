// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build pipeline_jit && linux && (amd64 || arm64)

package pipeline

import "testing"

// TestSetTag, pipeline-go v1.4.3 ptinput/funcs/fn_set_tag_test.go.
// Keep all seven upstream scripts, including the ordering-sensitive pre-grok
// promotion and list/map Go JSON stringification cases.
func TestJITSetTagUpstreamOracle(t *testing.T) {
	const input = `162.62.81.1 - - [29/Nov/2021:07:30:50 +0000] "123 /?signature=b8d8ea&timestamp=1638171049 HTTP/1.1" 200 413 "-" "Mozilla/4.0"`
	const grok = `grok(_, "%{IPORHOST:client_ip} %{NOTSPACE} %{NOTSPACE} \\[%{HTTPDATE:time}\\] \"%{DATA:data} %{GREEDYDATA} HTTP/%{NUMBER}\" %{INT:status_code} %{INT:bytes}")`
	runJSONOracle(t, []jsonOracleCase{
		{name: "0-promote-before-grok", source: `set_tag(client_ip); ` + grok + `; cast(data,"int")`, input: input, key: "client_ip", want: "162.62.81.1"},
		{name: "1-promote-after-grok", source: grok + `; set_tag(client_ip); cast(data,"int")`, input: input, key: "client_ip", want: "162.62.81.1"},
		{name: "2-explicit-literal", source: grok + `; set_tag(client_ip,"1"); cast(data,"int")`, input: input, key: "client_ip", want: "1"},
		{name: "3-string-source", source: `json(_,str_a); json(_,str_b); set_tag(str_a,str_b)`, input: `{"str_a":"2","str_b":"3"}`, key: "str_a", want: "3"},
		{name: "4-integer-source", source: `json(_,str_a); json(_,str_b); set_tag(str_a,str_b)`, input: `{"str_a":"2","str_b":3}`, key: "str_a", want: "3"},
		{name: "5-list-stringify", source: `a=[1,2]; set_tag("arr_tag",a)`, input: "test", key: "arr_tag", want: "[1,2]"},
		{name: "6-map-stringify", source: `a={"a":1,"b":"x"}; set_tag("obj_tag",a)`, input: "test", key: "obj_tag", want: `{"a":1,"b":"x"}`},
	})
}

func TestJITNestedSetTagOracle(t *testing.T) {
	runJSONOracle(t, []jsonOracleCase{
		{name: "literal", source: `add_key(result,set_tag(destination,"value") == nil)`, key: "destination", want: "value"},
		{name: "promote", source: `add_key(destination,7); add_key(result,set_tag(destination) == nil)`, key: "destination", want: "7"},
		{name: "source", source: `value="source"; add_key(result,set_tag(destination,value) == nil)`, key: "destination", want: "source"},
		{name: "missing", source: `add_key(result,set_tag(destination) == nil)`, key: "destination", want: ""},
		{name: "short_circuit", source: `add_key(destination,7); result=false && set_tag(destination) == nil`, key: "destination", want: int64(7)},
		{name: "prefix", source: `result=(set_tag(destination,"kept") == nil) && 1/divisor == 0`, key: "destination", want: "kept", fails: true},
		{name: "attribute_key", source: `add_key(result,set_tag(obj.name,"value") == nil)`, key: "obj.name", want: "value"},
		{name: "attribute_value", source: `obj={"name":"not-read"}; add_key(result,set_tag(destination,obj.name) == nil)`, key: "destination", want: ""},
		{name: "attribute_scalar", source: `obj=7; add_key(result,set_tag(destination,obj.name) == nil)`, key: "destination", want: ""},
		{name: "alias_promote", source: `add_key(result,set_tag(_) == nil)`, input: "original", key: "message", want: "original"},
		{name: "alias_explicit", source: `add_key(result,set_tag(_,"new") == nil)`, input: "original", key: "message", want: "new"},
	})
}
