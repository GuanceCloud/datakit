// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build pipeline_jit && linux && (amd64 || arm64)

package pipeline

import (
	"reflect"
	"strings"
	"testing"

	"github.com/GuanceCloud/cliutils/point"
	"github.com/GuanceCloud/pipeline-go/ptinput"
)

// Go-only characterization; passing this test is not Rust compatibility evidence.
func TestGoCacheContainerCharacterization(t *testing.T) {
	for _, tc := range []struct{ name, source string }{
		{"list-type", `cache_set("c",[1,2]); x=cache_get("c"); add_key(kind,value_type(x)); add_key(result,x)`},
		{"map-type", `cache_set("c",{"a":1}); x=cache_get("c"); add_key(kind,value_type(x)); add_key(result,x)`},
		{"list-len", `cache_set("c",[1,2]); x=cache_get("c"); add_key(size,len(x))`},
		{"list-index", `cache_set("c",[1,2]); x=cache_get("c"); add_key(element,x[0])`},
		{"map-source-alias", `a={"x":1}; cache_set("c",a); a["x"]=2; x=cache_get("c"); add_key(result,x)`},
		{"list-source-alias", `a=[1,2]; cache_set("c",a); a[0]=3; x=cache_get("c"); add_key(result,x)`},
		{"list-output-snapshot", `a=[1,2]; cache_set("c",a); x=cache_get("c"); add_key(result,x); a[0]=9`},
		{"map-output-snapshot", `a={"x":1}; cache_set("c",a); x=cache_get("c"); add_key(result,x); a["x"]=9`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				p := recover()
				if tc.name == "list-len" {
					if p == nil {
						t.Error("expected known upstream type-assertion panic")
					}
				} else if p != nil {
					t.Errorf("unexpected Go panic: %v", p)
				}
			}()
			script, err := NewPlScriptSimple(point.Logging, "cache-characterize.p", tc.source)
			if err != nil {
				t.Fatal(err)
			}
			pt := newRealScriptPoint("cache", map[string]any{"message": "input"})
			errRun := script.Run(ptinput.PtWrap(point.Logging, pt), nil, nil)
			if tc.name == "list-index" {
				if errRun == nil || !strings.Contains(errRun.Error(), "unindexable type: str") {
					t.Fatalf("unexpected index result: %v", errRun)
				}
				return
			}
			if errRun != nil {
				t.Fatal(errRun)
			}
			switch tc.name {
			case "list-type":
				if pt.Get("kind") != "str" || !reflect.DeepEqual(pt.Get("result"), []int64{1, 2}) {
					t.Fatalf("unexpected list type/value: %#v", pt.KVMap())
				}
			case "map-type":
				if pt.Get("kind") != "str" {
					t.Fatalf("unexpected map type: %#v", pt.KVMap())
				}
			case "list-source-alias":
				if !reflect.DeepEqual(pt.Get("result"), []int64{3, 2}) {
					t.Fatalf("source alias lost: %#v", pt.KVMap())
				}
			case "list-output-snapshot":
				if !reflect.DeepEqual(pt.Get("result"), []int64{1, 2}) {
					t.Fatalf("Point write did not snapshot array: %#v", pt.KVMap())
				}
			}
			if strings.HasPrefix(tc.name, "map-") {
				payload, ok := pt.GetA("result")
				if !ok || payload == nil {
					t.Fatal("map result lost protobuf payload")
				}
				if payload.TypeUrl != "type.googleapis.com/point.Map" {
					t.Fatalf("unexpected payload type: %s", payload.TypeUrl)
				}
				var decoded point.Map
				if err := decoded.Unmarshal(payload.Value); err != nil {
					t.Fatal(err)
				}
				key, want := "x", int64(1)
				if tc.name == "map-type" {
					key = "a"
				} else if tc.name == "map-source-alias" {
					want = 2
				}
				if len(decoded.Map) != 1 || decoded.Map[key] == nil {
					t.Fatalf("unexpected map entries: %v", decoded.Map)
				}
				integer, ok := decoded.Map[key].X.(*point.BasicTypes_I)
				if !ok || integer.I != want {
					t.Fatalf("map payload must be int %d, got %v", want, decoded.Map[key])
				}
			}
			t.Logf("error=%v fields=%#v", errRun, pt.KVMap())
		})
	}
}
