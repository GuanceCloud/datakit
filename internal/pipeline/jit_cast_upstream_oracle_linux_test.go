// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build pipeline_jit && linux && (amd64 || arm64)

// Cases adapted from pipeline-go v1.4.3 ptinput/funcs/fn_cast_test.go, TestCast.
// Unless explicitly stated otherwise all files in that repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package pipeline

import (
	"fmt"
	"os"
	"testing"

	"github.com/GuanceCloud/cliutils/point"
	pljit "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline/jit"
)

// Keep the upstream grok expression and input template: testing cast alone
// would miss the original grok -> captured field -> cast integration.
const upstreamCastGrok = `grok(_, "%{IPORHOST:client_ip} %{NOTSPACE} %{NOTSPACE} \\[%{HTTPDATE:time}\\] \"%{DATA:data} %{GREEDYDATA} HTTP/%{NUMBER}\" %{INT:status_code} %{INT:bytes}")`

func TestJITUpstreamCastOracle(t *testing.T) {
	var cases []jsonOracleCase
	for index, tc := range []struct {
		name, captured, target string
		want                   any
	}{
		{"cast int", "123", "int", int64(123)},
		{"cast string", "123", "str", "123"},
		{"cast bool", "true", "bool", true},
		{"cast float", "123", "float", float64(123)},
		{"cast float", "12.3", "float", float64(12.3)},
		{"cast float ", "-123.", "float", float64(-123)},
		{"cast float ", "-.12", "float", float64(-.12)},
		{"cast int ", "+12.6", "int", int64(12)},
	} {
		cases = append(cases, jsonOracleCase{
			name:   fmt.Sprintf("%02d-%s", index, tc.name),
			source: upstreamCastGrok + fmt.Sprintf("\ncast(data, %q)", tc.target),
			input:  fmt.Sprintf(`162.62.81.1 - - [29/Nov/2021:07:30:50 +0000] "%s /?signature=b8d8ea&timestamp=1638171049 HTTP/1.1" 200 413 "-" "Mozilla/4.0"`, tc.captured),
			key:    "data", want: tc.want,
		})
	}
	runJSONOracle(t, cases)
}

func TestJITUpstreamCastInvalidTarget(t *testing.T) {
	// Upstream case 8 has fail=true: the checker rejects intx, so the
	// upstream output expectation "+12.6" is not reached on this path.
	source := upstreamCastGrok + "\ncast(data, \"intx\")"
	if _, err := NewPlScriptSimple(point.Logging, "cast-upstream.p", source); err == nil {
		t.Fatal("upstream invalid target unexpectedly accepted by Go")
	}
	path := os.Getenv("PLATYPUS_JIT_RUNTIME")
	if path == "" {
		t.Fatal("requires real PLATYPUS_JIT_RUNTIME")
	}
	runner, err := pljit.NewRunner(path, "pipeline-go-1.4.3-datakit", 1)
	if err != nil {
		t.Fatal(err)
	}
	defer runner.Close()
	check := runner.Check(source)
	if check.Route != pljit.RoutePipelineGo || check.Reason != pljit.CheckReasonCompileError || check.Detail == "" {
		t.Fatalf("expected rejected invalid target with diagnostic, got %+v", check)
	}
}
