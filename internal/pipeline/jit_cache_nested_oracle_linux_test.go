// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build pipeline_jit && linux && (amd64 || arm64)

package pipeline

import "testing"

// Supplemental combinations of fn_cache.go semantics, not upstream test counts.
func TestJITNestedCacheSetOracle(t *testing.T) {
	runJSONOracle(t, []jsonOracleCase{
		{name: "write", source: `add_key(result,cache_set("nested","stored") == nil); add_key(value,cache_get("nested"))`, key: "value", want: "stored"},
		{name: "short_circuit", source: `cache_set("nested","old"); result=false && cache_set("nested","new") == nil; add_key(value,cache_get("nested"))`, key: "value", want: "old"},
		{name: "zero_ttl", source: `cache_set("nested","old"); add_key(result,cache_set("nested","new",0) == nil); add_key(value,cache_get("nested"))`, key: "value", want: "old"},
		{name: "named_order", source: `add_key(result,cache_set("nested",exp=pt_kvs_set("order",2),value=pt_kvs_set("order",1)) == nil)`, key: "order", want: int64(2), fails: true},
		{name: "invalid_key_stops_value", source: `add_key(result,cache_set(7,pt_kvs_set("unexpected",1)) == nil)`, key: "unexpected", want: nil, fails: true},
		{name: "value_error_stops_expiration", source: `add_key(result,cache_set("nested",1/divisor,pt_kvs_set("unexpected",1)) == nil)`, key: "unexpected", want: nil, fails: true},
	})
}
