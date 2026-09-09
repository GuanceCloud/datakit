// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build pipeline_jit && linux && (amd64 || arm64)

// Adapted from pipeline-go v1.4.3 ptinput/funcs/fn_group_test.go, TestGroup.
// Unless explicitly stated otherwise all files in that repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package pipeline

import "testing"

func TestJITGroupBetweenUpstreamOracle(t *testing.T) {
	const input = `{"status":200,"age":47}`
	runJSONOracle(t, []jsonOracleCase{
		{name: "0", source: `json(_,status); group_between(status,[200,400],false,newkey)`, input: input, key: "newkey", want: false},
		{name: "1", source: `json(_,status); group_between(status,[200,400],10,newkey)`, input: input, key: "newkey", want: int64(10)},
		{name: "2", source: `json(_,status); group_between(status,[200,400],"ok",newkey)`, input: input, key: "newkey", want: "ok"},
		// The upstream PlPoint harness does not apply DataKit status mapping.
		// Disable it explicitly to retain the original assertion; test the
		// actual DataKit default separately below.
		{name: "3", source: `setopt(status_mapping=false); json(_,status); group_between(status,[200,299],"ok")`, input: input, key: "status", want: "ok"},
		{name: "4", source: `json(_,status); group_between(status,[200,299],"ok",newkey)`, input: input, key: "newkey", want: "ok"},
		{name: "5", source: `json(_,status); group_between(status,[300,400],"ok",newkey)`, input: input, key: "newkey", want: nil},
	})
}

func TestJITGroupBetweenStatusMappingOracle(t *testing.T) {
	runJSONOracle(t, []jsonOracleCase{
		{name: "datakit_default", source: `json(_,status); group_between(status,[200,299],"ok")`, input: `{"status":200,"age":47}`, key: "status", want: "OK"},
	})
}
