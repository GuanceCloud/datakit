// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build pipeline_jit && linux && (amd64 || arm64)

package pipeline

import "testing"

func TestJITAddKeyAssignmentOracle(t *testing.T) {
	runJSONOracle(t, []jsonOracleCase{
		{name: "attribute_assignment", source: `chosen="old"; obj={"name":"new"}; add_key(result,chosen=obj.name); add_key(binding,chosen)`, key: "binding", want: nil},
		{name: "parenthesized_attribute", source: `obj={"name":"new"}; add_key(result,(obj.name))`, key: "result", want: nil},
		{name: "parenthesized_assignment_attribute", source: `chosen="old"; obj={"name":"new"}; add_key(result,chosen=(obj.name)); add_key(binding,chosen)`, key: "binding", want: nil},
		{name: "named_value", source: `add_key(result,value=7); add_key(binding,value)`, key: "binding", want: int64(7)},
		{name: "arbitrary_name", source: `add_key(result,chosen=7); add_key(binding,chosen)`, key: "binding", want: int64(7)},
		{name: "rhs_error", source: `chosen="old"; add_key(result,chosen=1/divisor)`, key: "result", want: nil, fails: true},
		{name: "nested_write", source: `add_key(result,chosen=pt_kvs_set("side",true)); add_key(binding,chosen)`, key: "binding", want: true},
	})
}

// The pinned Go RunStmt returns nil/Void for AttrExpr without evaluating Obj.
func TestJITAddKeyAttributeValueOracle(t *testing.T) {
	runJSONOracle(t, []jsonOracleCase{
		{name: "map", source: `obj={"name":"value"}; add_key(result,obj.name)`, key: "result", want: nil},
		{name: "existing", source: `add_key(result,"old"); obj={"name":"value"}; add_key(result,obj.name)`, key: "result", want: nil},
		{name: "scalar", source: `obj=7; add_key(result,obj.name)`, key: "result", want: nil},
		{name: "missing", source: `add_key(result,missing.name)`, key: "result", want: nil},
		{name: "indexed_object", source: `obj=[{"name":"value"}]; add_key(result,obj[0].name)`, key: "result", want: nil},
	})
}
