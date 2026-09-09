// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build pipeline_jit && linux && (amd64 || arm64)

// TestExit adapted from pipeline-go v1.4.3 ptinput/funcs/fn_exit_test.go.
// Unless explicitly stated otherwise all files in that repository are licensed
// under the MIT License.

package pipeline

import "testing"

func TestJITExitUpstreamOracle(t *testing.T) {
	const input = `162.62.81.1 - - [29/Nov/2021:07:30:50 +0000] "123 /?signature=b8d8ea&timestamp=1638171049 HTTP/1.1" 200 413 "-" "Mozilla/4.0"`
	const source = `grok(_, "%{IPORHOST:client_ip} %{NOTSPACE} %{NOTSPACE} \\[%{HTTPDATE:time}\\] \"%{DATA:data} %{GREEDYDATA} HTTP/%{NUMBER}\" %{INT:status_code} %{INT:bytes}"); exit(); cast(data,"int")`
	runJSONOracle(t, []jsonOracleCase{{
		name: "cast-int-stopped", source: source, input: input, key: "data", want: "123",
	}})
}
