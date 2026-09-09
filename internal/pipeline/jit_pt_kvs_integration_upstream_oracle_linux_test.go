// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build pipeline_jit && linux && (amd64 || arm64)

// Cases adapted from pipeline-go v1.4.3
// ptinput/funcs/fn_pt_kvs_integration_test.go.
// Unless explicitly stated otherwise all files in that repository are licensed
// under the MIT License.

package pipeline

import (
	"testing"

	"github.com/GuanceCloud/cliutils/point"
)

func TestJITPtKvsIntegrationUpstreamOracle(t *testing.T) {
	runJSONOracle(t, []jsonOracleCase{
		{
			name:   "get-in-operator",
			source: `arr=pt_kvs_get("nums",true); if 2 in arr { pt_kvs_set("hit",true) }`,
			key:    "hit", want: true,
		},
		{
			name:   "get-append-len-chain",
			source: `arr=pt_kvs_get("nums",true); arr=append(arr,4); pt_kvs_set("size",len(arr)); pt_kvs_set("arr",arr,false,true)`,
			key:    "size", want: int64(4),
		},
		{
			name:   "set-map-string",
			source: `obj={"a":1,"b":"x"}; pt_kvs_set("obj",obj)`,
			key:    "obj", want: `{"a":1,"b":"x"}`,
		},
		{
			name:   "set-message-map-compatibility",
			source: `result_msg={}; all_keys=pt_kvs_keys(); for k in all_keys { if k!="message" { result_msg[k]=pt_kvs_get(k) } }; pt_kvs_set("message",result_msg)`,
			key:    "message", want: `{"before":true,"divisor":0,"existing_tag":"keep","nums":"[1,2,3]","sentinel":"preserved"}`,
		},
		{
			name:   "get-map-compatibility",
			source: `obj={"a":1,"b":"x"}; pt_kvs_set("obj",obj); add_key("obj_type",value_type(pt_kvs_get("obj")))`,
			key:    "obj_type", want: "str",
		},
		{
			name:   "get-raw-option",
			source: `pt_kvs_set("obj",{"a":1},false,true); pt_kvs_set("raw_type",value_type(pt_kvs_get("obj",true))); pt_kvs_set("plain_type",value_type(pt_kvs_get("obj",false))); pt_kvs_set("plain_value",pt_kvs_get("obj",false))`,
			key:    "raw_type", want: "map",
		},
		{
			name:   "set-raw-option",
			source: `obj={"a":1}; pt_kvs_set("obj_raw",obj,false,true); pt_kvs_set("obj_str",obj,false,false)`,
			key:    "obj_str", want: `{"a":1}`,
		},
	}, func(pt *point.Point) { pt.AddKVs(point.NewKV("nums", []int{1, 2, 3})) })
}
