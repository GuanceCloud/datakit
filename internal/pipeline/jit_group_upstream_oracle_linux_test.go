// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build pipeline_jit && linux && (amd64 || arm64)

// Adapted from pipeline-go v1.4.3 ptinput/funcs/fn_group_in_test.go.
// Unless explicitly stated otherwise all files in that repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package pipeline

import "testing"

func TestJITGroupInUpstreamOracle(t *testing.T) {
	// Keep both duplicated upstream entries so indices map exactly to TestGroupIn.
	runJSONOracle(t, []jsonOracleCase{
		{name: "0", source: `json(_,status); group_in(status,[true],"ok","newkey")`, input: `{"status":true,"age":"47"}`, key: "newkey", want: "ok"},
		{name: "1", source: `json(_,status); group_in(status,[true],"ok","newkey")`, input: `{"status":true,"age":"47"}`, key: "newkey", want: "ok"},
		{name: "2", source: `json(_,status); group_in(status,[true],"ok","newkey")`, input: `{"status":"aa","age":"47"}`, key: "status", want: "aa"},
		{name: "3", source: `json(_,status); group_in(status,["aa"],"ok","newkey")`, input: `{"status":"aa","age":"47"}`, key: "newkey", want: "ok"},
		{name: "4", source: `json(_,log_level); group_in(log_level,[200,"test"],119)`, input: `{"log_level":"test","age":"47"}`, key: "log_level", want: int64(119)},
		{name: "5", source: `json(_,log_level); group_in(log_level,[200,"test1"],119)`, input: `{"log_level":"test","age":"47"}`, key: "log_level", want: "test"},
		{name: "6", source: `json(_,log_level); group_in(log_level,[200,"test"],119,"hh")`, input: `{"log_level":"test","age":"47"}`, key: "hh", want: int64(119)},
	})
}

// Additional compositions, not counted as upstream original test cases.
func TestJITGroupInLoopOracle(t *testing.T) {
	runJSONOracle(t, []jsonOracleCase{
		{name: "changing_candidates", source: `choices=["miss"]; total=0; for i=0; i<4; i=i+1 { choices[0]="miss"; if i==2 { choices[0]="hit" }; drop_key(result); group_in(message,choices,true,result); if result==true { total=total+1 } }; add_key(total,total)`, input: "hit", key: "total", want: int64(1)},
		{name: "nested_changing_candidates", source: `choices=["miss"]; total=0; for i=0; i<4; i=i+1 { choices[0]="miss"; if i==2 { choices[0]="hit" }; drop_key(result); returned=group_in(message,choices,true,result)==nil; if result==true { total=total+1 } }; add_key(total,total)`, input: "hit", key: "total", want: int64(1)},
	})
}
