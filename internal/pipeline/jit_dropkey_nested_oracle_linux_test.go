// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build pipeline_jit && linux && (amd64 || arm64)

package pipeline

import "testing"

func TestJITNestedDropKeyOracle(t *testing.T) {
	runJSONOracle(t, []jsonOracleCase{
		{name: "delete", source: `add_key(victim,7); add_key(result,drop_key(victim) == nil)`, key: "victim", want: nil},
		{name: "missing", source: `add_key(result,drop_key(absent) == nil)`, key: "result", want: true},
		{name: "short_circuit", source: `add_key(victim,7); result=false && drop_key(victim) == nil`, key: "victim", want: int64(7)},
		{name: "prefix", source: `add_key(victim,7); result=(drop_key(victim) == nil) && 1/divisor == 0`, key: "victim", want: nil, fails: true},
		{name: "attribute", source: `add_key("obj.name",7); add_key(result,drop_key(obj.name) == nil)`, key: "obj.name", want: nil},
		{name: "alias", source: `add_key(result,drop_key(_) == nil)`, input: "original", key: "message", want: nil},
	})
}
