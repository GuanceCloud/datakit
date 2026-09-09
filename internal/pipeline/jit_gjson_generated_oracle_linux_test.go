// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build pipeline_jit && linux && (amd64 || arm64)

package pipeline

import (
	"fmt"
	"strings"
	"testing"

	"github.com/tidwall/gjson"
)

// TestJITGJSONGeneratedGrammarOracle exercises grammar families rather than
// adding one regression path at a time. Expectations come directly from the
// locked tidwall/gjson v1.17.0 implementation used by pipeline-go.
func TestJITGJSONGeneratedGrammarOracle(t *testing.T) {
	const input = `{
		"name":{"first":"Tom","last":"Anderson"},
		"a.b":{"literal":"dot"},
		"children":["Sara","Alex","Jack"],
		"friends":[
			{"first":"Dale","age":44,"active":true,"nets":["fb","tw"]},
			{"first":"Roger","age":68,"active":false,"nets":["tw"]},
			{"first":"Jane","age":47,"active":true,"nets":[]}
		],
		"groups":[
			{"name":"ops","members":[{"role":"admin"},{"role":"reader"}]},
			{"name":"dev","members":[{"role":"writer"}]}
		],
		"objects":[{"a":1,"b":2},{"b":3,"c":4}],
		"encoded":"{\"inner\":{\"value\":9}}"
	}`

	paths := []string{
		"name.first", `a\.b.literal`, "children.0", "children.#",
		"friends.#.first", "friends.#.nets.#.0", "friends.#",
		"groups.#.members.#.role", "groups.#(members.#(role==\"admin\")).name",
		"{first:name.first,last:name.last}", "[name.first,children.1,friends.2.age]",
		"friends|@reverse|0|first", "objects|@join|b", "encoded|@fromstr|inner.value",
		"@this", "@valid", "@ugly", "@keys", "@values", "@group", "@dig:role",
	}

	for _, all := range []bool{false, true} {
		allSuffix := ""
		if all {
			allSuffix = "#"
		}
		for _, operator := range []string{"==", "!=", ">", ">=", "<", "<=", "%", "!%"} {
			operand := "47"
			field := "age"
			if strings.Contains(operator, "%") {
				field, operand = "first", `"*a*"`
			}
			paths = append(paths, fmt.Sprintf("friends.#(%s%s%s)%s.first", field, operator, operand, allSuffix))
		}
	}

	for _, argument := range []string{
		"", `:{"deep":true}`, `:{"deep":false}`, `:{"deep":true,"ignored":1}`,
	} {
		paths = append(paths, "@flatten"+argument)
	}
	for _, argument := range []string{
		"", `:{"preserve":true}`, `:{"preserve":false}`, `:{"ignored":1,"preserve":true}`,
	} {
		paths = append(paths, "@join"+argument, "objects|@join"+argument)
	}
	for _, argument := range []string{
		"", `:{"sortKeys":true}`, `:{"sortKeys":false,"indent":"  "}`,
		`:{"prefix":"> ","indent":"\t","width":40}`,
	} {
		paths = append(paths, "@pretty"+argument)
	}

	cases := make([]jsonOracleCase, 0, len(paths))
	for index, path := range paths {
		result := gjson.Get(input, path)
		var want any
		switch result.Type {
		case gjson.Number:
			want = result.Float()
		case gjson.True, gjson.False:
			want = result.Bool()
		case gjson.String, gjson.JSON:
			want = result.String()
		case gjson.Null:
			want = nil
		default:
			t.Fatalf("generated path %q returned unexpected type %s", path, result.Type)
		}
		cases = append(cases, jsonOracleCase{
			name:   fmt.Sprintf("generated_%03d", index),
			source: fmt.Sprintf("gjson(message,%q,\"result\")", path),
			input:  input,
			key:    "result",
			want:   want,
		})
	}
	runJSONOracle(t, cases)
}
