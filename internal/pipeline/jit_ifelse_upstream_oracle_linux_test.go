// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build pipeline_jit && linux && (amd64 || arm64)

// TestIfelse adapted from pipeline-go v1.4.3 ptinput/funcs/ifelse_test.go.
// Unless explicitly stated otherwise all files in that repository are licensed
// under the MIT License.

package pipeline

import "testing"

func TestJITIfElseUpstreamOracle(t *testing.T) {
	const logInput = `1.2.3.4 "POST /?signature=b8d8ea&timestamp=1638171049 HTTP/1.1" 200 413`
	const grok = `grok(_, "%{IPORHOST:client_ip} \"%{DATA} %{GREEDYDATA} HTTP/%{NUMBER}\" %{INT:status_code} %{INT:bytes}")`
	runJSONOracle(t, []jsonOracleCase{
		{name: "0-missing-equals-nil", source: grok + `; if invalid_status_code==nil { add_key(add_new_key,"OK") }`, input: logInput, key: "add_new_key", want: "OK"},
		{name: "1-int-one", source: `if 1 { add_key(add_new_key,"OK") }`, key: "add_new_key", want: "OK"},
		{name: "2-float-one", source: `if 1.1 { add_key(add_new_key,"OK") }`, key: "add_new_key", want: "OK"},
		{name: "3-int-zero", source: `if 0 { add_key(add_new_key,"OK") }`, key: "add_new_key", want: nil},
		{name: "4-true", source: `if true { add_key(add_new_key,"OK") }`, key: "add_new_key", want: "OK"},
		{name: "5-string", source: `if "str" { add_key(add_new_key,"OK") }`, key: "add_new_key", want: "OK"},
		{name: "6-nil", source: `if nil { add_key(add_new_key,"OK") }`, key: "add_new_key", want: nil},
		{name: "7-list", source: `if [123] { add_key(add_new_key,"OK") }`, key: "add_new_key", want: "OK"},
		{name: "8-list-len", source: `if len([123]) { add_key(add_new_key,"OK") }`, key: "add_new_key", want: "OK"},
		{name: "9-missing-not-equals-nil", source: grok + `; if invalid_status_code!=nil { add_key(add_new_key,"OK") }`, input: logInput, key: "add_new_key", want: nil},
	})
}
