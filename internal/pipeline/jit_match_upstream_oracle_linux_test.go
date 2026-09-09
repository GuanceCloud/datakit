// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build pipeline_jit && linux && (amd64 || arm64)

// Adapted from pipeline-go v1.4.3 ptinput/funcs/fn_match_test.go.
// Copyright 2021-present Guance, Inc. MIT License; see testdata/PIPELINE_GO_TEST_LICENSE.txt.
package pipeline

import (
	"github.com/GuanceCloud/cliutils/point"
	pljit "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline/jit"
	"os"
	"testing"
)

func TestJITUpstreamMatchOracle(t *testing.T) {
	runJSONOracle(t, []jsonOracleCase{
		{name: "0-word", source: `abc = "sss"; add_key(abc, match("\\w+", abc))`, input: "test", key: "abc", want: true},
		{name: "1-literal", source: `abc = "sss"; add_key(abc, match("sss", abc))`, input: "test", key: "abc", want: true},
	})
}

func TestJITMatchDynamicPatternExtension(t *testing.T) {
	const source = `abc = "sss"; add_key(abc, match(abc, abc))`
	if _, err := NewPlScriptSimple(point.Logging, "match.p", source); err == nil {
		t.Fatal("expected Go checking rejection")
	}
	path := os.Getenv("PLATYPUS_JIT_RUNTIME")
	if path == "" {
		t.Fatal("requires real runtime")
	}
	runner, err := pljit.NewRunner(path, "pipeline-go-1.4.3-datakit", 1)
	if err != nil {
		t.Fatal(err)
	}
	defer runner.Close()
	check := runner.Check(source)
	if check.Route != pljit.RouteJITNative || check.Capabilities.Backend != pljit.ExecutionBackendMachineCode {
		t.Fatalf("dynamic match extension rejected: %+v", check)
	}
}
