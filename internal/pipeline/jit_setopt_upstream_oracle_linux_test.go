// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build pipeline_jit && linux && (amd64 || arm64)

// TestSetopt cases adapted from pipeline-go v1.4.3
// ptinput/funcs/fn_setopt_test.go.
// Unless explicitly stated otherwise all files in that repository are licensed
// under the MIT License.

package pipeline

import (
	"os"
	"testing"

	"github.com/GuanceCloud/cliutils/point"
	pljit "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline/jit"
)

const observeStatusMapping = `; add_key(status,200); group_between(status,[200,299],"ok")`

func TestJITSetoptUpstreamOracle(t *testing.T) {
	runJSONOracle(t, []jsonOracleCase{
		{name: "0-disabled", source: `1+1; setopt(status_mapping=false)` + observeStatusMapping, key: "status", want: "ok"},
		{name: "1-enabled", source: `setopt(status_mapping=true)` + observeStatusMapping, key: "status", want: "OK"},
		{name: "3-dead-branch", source: `if false { setopt(status_mapping=false) }` + observeStatusMapping, key: "status", want: "OK"},
		{name: "4-live-branch", source: `if true { setopt(status_mapping=false) }` + observeStatusMapping, key: "status", want: "ok"},
		{name: "6-empty", source: `setopt()` + observeStatusMapping, key: "status", want: "OK"},
	})
}

func TestJITSetoptUpstreamInvalidCalls(t *testing.T) {
	path := os.Getenv("PLATYPUS_JIT_RUNTIME")
	if path == "" {
		t.Fatal("requires real PLATYPUS_JIT_RUNTIME")
	}
	runner, err := pljit.NewRunner(path, "pipeline-go-1.4.3-datakit", 2)
	if err != nil {
		t.Fatal(err)
	}
	defer runner.Close()
	for _, source := range []string{
		`setopt(true)`,
		`if false { setopt(status_mapping=false,x=1) }`,
	} {
		t.Run(source, func(t *testing.T) {
			if _, err := NewPlScriptSimple(point.Logging, "setopt-upstream.p", source); err == nil {
				t.Fatal("pipeline-go accepted invalid setopt call")
			}
			check := runner.Check(source)
			if check.Route != pljit.RoutePipelineGo || check.Reason != pljit.CheckReasonCompileError || check.Detail == "" {
				t.Fatalf("Rust accepted invalid setopt call: %+v", check)
			}
		})
	}
}

func TestJITDiscardedExpressionStatementOracle(t *testing.T) {
	runJSONOracle(t, []jsonOracleCase{
		{name: "pure", source: `1+1; add_key(after,true)`, key: "after", want: true},
		{name: "nil-add-errors", source: `add_key(before,true); nil+1; add_key(after,true)`, key: "before", want: true, fails: true},
		{name: "error-stops-continuation", source: `add_key(before,true); 1/divisor; add_key(after,true)`, key: "before", want: true, fails: true},
		{name: "side-effect-and-bool-add", source: `pt_kvs_set("side",7)+1; add_key(after,true)`, key: "side", want: int64(7)},
	})
}
