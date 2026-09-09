// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build pipeline_jit && linux && (amd64 || arm64)

package pipeline

import (
	"github.com/GuanceCloud/cliutils/point"
	pljit "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline/jit"
	"os"
	"testing"
)

func TestJITPatternConditionalCheckingOracle(t *testing.T) {
	runJSONOracle(t, []jsonOracleCase{
		{name: "nested_expression_declares", source: `add_key(declared, add_pattern("CUSTOM","[a-z]+") == nil); grok(_,"%{CUSTOM:capture}")`, input: "abc123", key: "capture", want: "abc"},
		{name: "loop_body_declares", source: `for i=0; i<2; i=i+1 { add_pattern("CUSTOM","[a-z]+"); grok(_,"%{CUSTOM:capture}") }`, input: "abc123", key: "capture", want: "abc"},
		{name: "zero_iteration_no_writes", source: `add_key(capture,"keep"); for i=0; i<0; i=i+1 { add_pattern("CUSTOM","[a-z]+"); grok(_,"%{CUSTOM:capture}") }`, input: "abc123", key: "capture", want: "keep"},
		{name: "false_branch_no_writes", source: `add_key(capture,"keep"); if false { add_pattern("CUSTOM","[a-z]+"); grok(_,"%{CUSTOM:capture}") }`, input: "abc123", key: "capture", want: "keep"},
		{name: "true_branch_declares", source: `if true { add_pattern("CUSTOM","[a-z]+"); grok(_,"%{CUSTOM:capture}") }`, input: "abc123", key: "capture", want: "abc"},
		{name: "sibling_isolation", source: `if false { add_pattern("CUSTOM","[a-z]+"); grok(_,"%{CUSTOM:capture}") } else { add_pattern("CUSTOM","[0-9]+"); grok(_,"%{CUSTOM:capture}") }`, input: "abc123", key: "capture", want: "123"},
		{name: "nested_inherits", source: `if true { add_pattern("CUSTOM","[a-z]+"); if true { add_key(result,grok(_,"%{CUSTOM:capture}")) } }`, input: "abc123", key: "capture", want: "abc"},
		{name: "earlier_call_keeps_pattern", source: `add_pattern("CUSTOM","[a-z]+")
grok(_,"%{CUSTOM:first}")
if false { add_pattern("CUSTOM","[0-9]+") }
grok(_,"%{CUSTOM:capture}")`, input: "abc123", key: "capture", want: "abc"},
	})
}

func TestJITPatternAliasSnapshotOracle(t *testing.T) {
	runJSONOracle(t, []jsonOracleCase{
		{name: "builtin_snapshot", source: `add_pattern("ALIAS","%{NUMBER}")
add_pattern("NUMBER","[a-z]+")
grok(_,"%{ALIAS:capture}")`, input: "abc123", key: "capture", want: "123"},
		{name: "typed_multilevel", source: `add_pattern("BASE","[0-9]+")
add_pattern("MID","%{BASE:number:int}")
add_pattern("ALIAS","%{MID}")
add_pattern("BASE","[a-z]+")
grok(_,"%{ALIAS:capture}")`, input: "abc123", key: "number", want: int64(123)},
		{name: "private_name_collision", source: `add_pattern("PPFROZEN0","[a-z]+")
add_pattern("ALIAS","%{PPFROZEN0}")
add_pattern("PPFROZEN0","[0-9]+")
grok(_,"%{ALIAS:capture}")`, input: "abc123", key: "capture", want: "abc"},
		{name: "alias_before_redefinition", source: `add_pattern("BASE","[a-z]+")
add_pattern("ALIAS","%{BASE}")
add_pattern("BASE","[0-9]+")
grok(_,"%{ALIAS:capture}")`, input: "abc123", key: "capture", want: "abc"},
		{name: "alias_in_child", source: `add_pattern("BASE","[a-z]+")
add_pattern("ALIAS","%{BASE}")
if true { add_pattern("BASE","[0-9]+"); grok(_,"%{ALIAS:capture}") }`, input: "abc123", key: "capture", want: "abc"},
	})
}

func TestJITPatternScopeRejectsEscapingDefinitions(t *testing.T) {
	path := os.Getenv("PLATYPUS_JIT_RUNTIME")
	if path == "" {
		t.Fatal("requires real runtime")
	}
	runner, err := pljit.NewRunner(path, "pipeline-go-1.4.3-datakit", 8)
	if err != nil {
		t.Fatal(err)
	}
	defer runner.Close()
	for _, source := range []string{
		`if true { add_pattern("CUSTOM","x") }; grok(_,"%{CUSTOM:capture}")`,
		`if false { add_pattern("CUSTOM","x") }; grok(_,"%{CUSTOM:capture}")`,
		`if true { add_pattern("CUSTOM","x") } else { grok(_,"%{CUSTOM:capture}") }`,
	} {
		t.Run(source, func(t *testing.T) {
			if _, err := NewPlScriptSimple(point.Logging, "scope.p", source); err == nil {
				t.Fatal("Go unexpectedly accepted escaped pattern")
			}
			check := runner.Check(source)
			if check.Route != pljit.RoutePipelineGo || check.Reason != pljit.CheckReasonCompileError {
				t.Fatalf("JIT accepted escaped pattern: %+v", check)
			}
		})
	}
}
