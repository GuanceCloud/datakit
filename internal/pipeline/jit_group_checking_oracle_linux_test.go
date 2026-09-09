// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build pipeline_jit && linux && (amd64 || arm64)

package pipeline

import (
	"os"
	"testing"

	"github.com/GuanceCloud/cliutils/point"
	pljit "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline/jit"
)

func TestJITGroupInInvalidKeysOracle(t *testing.T) {
	path := os.Getenv("PLATYPUS_JIT_RUNTIME")
	if path == "" {
		t.Fatal("requires real PLATYPUS_JIT_RUNTIME")
	}
	runner, err := pljit.NewRunner(path, "pipeline-go-1.4.3-datakit", 16)
	if err != nil {
		t.Fatal(err)
	}
	defer runner.Close()
	for name, source := range map[string]string{
		"assigned_source":                  `group_in(key=message,["hit"],true,result)`,
		"numeric_source":                   `group_in(42,["hit"],true,result)`,
		"numeric_target":                   `group_in(message,["hit"],true,42)`,
		"assignment_middle_invalid_target": `group_in(message,range=["hit"],true,42)`,
		"dynamic_list_invalid_target":      `item="hit"; group_in(message,[item],true,42)`,
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := NewPlScriptSimple(point.Logging, "group-check.p", source); err == nil {
				t.Fatal("Go unexpectedly accepts fixture")
			}
			check := runner.Check(source)
			if check.Route != pljit.RoutePipelineGo || check.Reason != pljit.CheckReasonCompileError || check.Detail == "" {
				t.Fatalf("expected compile rejection, got %+v", check)
			}
		})
	}
}

func TestJITGroupBetweenInvalidRangeOracle(t *testing.T) {
	path := os.Getenv("PLATYPUS_JIT_RUNTIME")
	if path == "" {
		t.Fatal("requires real PLATYPUS_JIT_RUNTIME")
	}
	runner, err := pljit.NewRunner(path, "pipeline-go-1.4.3-datakit", 16)
	if err != nil {
		t.Fatal(err)
	}
	defer runner.Close()
	for name, source := range map[string]string{
		"reversed":                        `group_between(message,[3,1],true,result)`,
		"short":                           `group_between(message,[1],true,result)`,
		"long":                            `group_between(message,[1,2,3],true,result)`,
		"string_bound":                    `group_between(message,["1",3],true,result)`,
		"bool_bound":                      `group_between(message,[false,3],true,result)`,
		"variable_bound":                  `limit=1; group_between(message,[limit,3],true,result)`,
		"effect_bound":                    `group_between(message,[pt_kvs_set("side",1),3],true,result)`,
		"assigned_source":                 `group_between(key=message,[1,3],true,result)`,
		"assigned_range":                  `group_between(message,between=[1,3],true,result)`,
		"assigned_replacement_bad_target": `group_between(message,[1,3],new_value=true,42)`,
		"nested_bad_target":               `add_key(returned,group_between(message,[1,3],new_value=true,42)==nil)`,
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := NewPlScriptSimple(point.Logging, "between-check.p", source); err == nil {
				t.Fatal("Go unexpectedly accepts fixture")
			}
			check := runner.Check(source)
			if check.Route != pljit.RoutePipelineGo || check.Reason != pljit.CheckReasonCompileError || check.Detail == "" {
				t.Fatalf("expected compile rejection, got %+v", check)
			}
		})
	}
	if check := runner.Check(`limits=[1,3]; group_between(message,limits,true,result)`); check.Route != pljit.RouteJITNative {
		t.Fatalf("dynamic range extension rejected: %+v", check)
	}
}

// pipeline-go rejects a trailing new_key binding because its legacy checker
// treats it as an assignment expression. Rust intentionally provides the
// regular named-argument form as a backwards-compatible language extension.
func TestJITGroupTrailingNamedTargetExtension(t *testing.T) {
	path := os.Getenv("PLATYPUS_JIT_RUNTIME")
	if path == "" {
		t.Fatal("requires real PLATYPUS_JIT_RUNTIME")
	}
	runner, err := pljit.NewRunner(path, "pipeline-go-1.4.3-datakit", 16)
	if err != nil {
		t.Fatal(err)
	}
	defer runner.Close()
	for _, source := range []string{
		`group_in(message,["hit"],true,new_key=result)`,
		`group_between(number,[1,3],true,new_key=result)`,
	} {
		if check := runner.Check(source); check.Route != pljit.RouteJITNative {
			t.Fatalf("named target extension rejected: source=%s check=%+v", source, check)
		}
	}
}
