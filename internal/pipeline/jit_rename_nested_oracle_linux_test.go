// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build pipeline_jit && linux && (amd64 || arm64)

package pipeline

import "testing"

func TestJITNestedRenameOracle(t *testing.T) {
	runJSONOracle(t, []jsonOracleCase{
		{name: "move", source: `add_key(old,7); add_key(result,rename(new,old) == nil)`, key: "new", want: int64(7)},
		{name: "overwrite", source: `add_key(old,7); add_key(new,8); add_key(result,rename(new,old) == nil)`, key: "new", want: int64(7)},
		{name: "tag", source: `set_tag(old,"value"); add_key(result,rename(new,old) == nil)`, key: "new", want: "value"},
		{name: "missing", source: `add_key(result,rename(new,absent) == nil)`, key: "new", want: nil},
		{name: "short_circuit", source: `add_key(old,7); result=false && rename(new,old) == nil`, key: "old", want: int64(7)},
		{name: "error_prefix", source: `add_key(old,7); result=(rename(new,old) == nil) && 1/divisor == 0`, key: "new", want: int64(7), fails: true},
		{name: "attribute", source: `add_key("obj.old",7); add_key(result,rename(obj.new,obj.old) == nil)`, key: "obj.new", want: int64(7)},
		{name: "alias", source: `add_key(result,rename(new,_) == nil)`, input: "original", key: "new", want: "original"},
	})
}
