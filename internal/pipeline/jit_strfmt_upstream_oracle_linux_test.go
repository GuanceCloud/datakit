// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build pipeline_jit && linux && (amd64 || arm64)

// Adapted from pipeline-go v1.4.3 ptinput/funcs/fn_strfmt_test.go.
// Copyright 2021-present Guance, Inc. MIT License; see testdata/PIPELINE_GO_TEST_LICENSE.txt.
package pipeline

import (
	"github.com/GuanceCloud/cliutils/point"
	pljit "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline/jit"
	"os"
	"testing"
)

func TestJITUpstreamStrfmtOracle(t *testing.T) {
	const input = `{"a":{"first":2.3,"second":2,"third":"abc","forth":true},"age":47}`
	const floatPrefix = `json(_, a.first, a_first); cast(a_first, "float"); `
	runJSONOracle(t, []jsonOracleCase{
		{name: "0-mixed", source: `json(_, a.second, a_second)
json(_, a.third, a_third)
cast(a_second, "int")
json(_, a.forth, a_forth)
strfmt(bb, "%d %s %v", a_second, a_third, a_forth)`, input: input, key: "bb", want: "2 abc true"},
		{name: "1-precision", source: floatPrefix + `strfmt(bb, "%.4f", a_first)`, input: input, key: "bb", want: "2.3000"},
		{name: "2-float-int", source: floatPrefix + `strfmt(bb, "%.4f%d", a_first, 3)`, input: input, key: "bb", want: "2.30003"},
		{name: "3-two-floats", source: floatPrefix + `strfmt(bb, "%.4f%.1f", a_first, 3.5)`, input: input, key: "bb", want: "2.30003.5"},
		{name: "4-bool-string", source: `json(_, a.forth, a_forth); strfmt(bb, "%v%s", a_forth, "tone")`, input: `{"a":{"first":2.3,"second":2,"third":"abcd","forth":true},"age":47}`, key: "bb", want: "truetone"},
	})
}

func TestJITUpstreamStrfmtInvalidArguments(t *testing.T) {
	path := os.Getenv("PLATYPUS_JIT_RUNTIME")
	if path == "" {
		t.Fatal("requires real runtime")
	}
	runner, err := pljit.NewRunner(path, "pipeline-go-1.4.3-datakit", 2)
	if err != nil {
		t.Fatal(err)
	}
	defer runner.Close()
	for _, call := range []string{`strfmt(bb)`, `strfmt(bb, 1)`} {
		source := `json(_, a.first, a_first); cast(a_first, "float"); ` + call
		t.Run(call, func(t *testing.T) {
			if _, err := NewPlScriptSimple(point.Logging, "strfmt.p", source); err == nil {
				t.Fatal("expected Go checking rejection")
			}
			check := runner.Check(source)
			if check.Route != pljit.RoutePipelineGo || check.Reason != pljit.CheckReasonCompileError {
				t.Fatalf("expected JIT checking rejection: %+v", check)
			}
		})
	}
}
