// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build pipeline_jit && linux && (amd64 || arm64)

package jit

import (
	"context"
	"os"
	"reflect"
	"testing"
	"time"
)

func TestDynamicNativeMatchesIndexedLifecycle(t *testing.T) {
	rt, err := openRuntime(os.Getenv("PLATYPUS_JIT_RUNTIME"), nil)
	if err != nil {
		t.Fatal(err)
	}
	defer rt.Close()
	for _, source := range []string{
		`key=pt_kvs_get("key"); pt_kvs_set(key, pt_kvs_get("value")); pt_kvs_set(key+"_tag","tag",true); pt_kvs_set("copy",pt_kvs_get(key))`,
		`doc=load_json(message); pt_kvs_set(doc["key"],doc["value"],false,true)`,
		`key=pt_kvs_get("key"); pt_kvs_set(key,1); for ; repeat; {}`,
		`key=pt_kvs_get("key"); pt_kvs_set(key,pt_kvs_get(["invalid"]))`,
	} {
		t.Run(source, func(t *testing.T) {
			compiled, err := rt.Compile(source, "pipeline-go-1.4.3-datakit")
			if err != nil {
				t.Fatal(err)
			}
			defer compiled.Close()
			program := compiled.(*nativeProgram)
			schema := program.static
			if schema == nil || !schema.Dynamic {
				t.Fatal("missing dynamic descriptor")
			}
			points := []Point{
				{Version: 1, Category: "logging", Measurement: "test", Fields: map[string]any{"key": "a", "value": int64(42), "repeat": true, "message": `{"key":"object","value":{"items":[1,"two"]}}`}},
				{Version: 1, Category: "logging", Measurement: "test", Fields: map[string]any{"key": "a", "value": "replacement", "repeat": false, "message": "invalid JSON"}},
			}
			input, err := EncodeFlatPoints(points)
			if err != nil {
				t.Fatal(err)
			}
			ctx := WithRecordTimeout(context.Background(), time.Millisecond)
			actual, err := program.ProcessIndexedContext(ctx, input)
			if err != nil {
				t.Fatal(err)
			}
			program.static = nil // independent, retained indexed encoder; only a test override
			expected, err := program.ProcessIndexedContext(ctx, input)
			if err != nil {
				t.Fatal(err)
			}
			program.static = schema
			for i, record := range actual.Records {
				want := expected.Records[i]
				if record.Status != want.Status || record.CommitPrefixError != want.CommitPrefixError || !reflect.DeepEqual(record.Error, want.Error) || !reflect.DeepEqual(record.Emitted, want.Emitted) {
					t.Fatalf("record %d lifecycle differs: %+v / %+v", i, record, want)
				}
				if record.HasMutations() {
					got, err := record.MutationDeltas()
					if err != nil {
						t.Fatal(err)
					}
					delta, err := want.MutationDeltas()
					if err != nil {
						t.Fatal(err)
					}
					if !reflect.DeepEqual(got, delta) {
						t.Fatalf("record %d mutations differ: %#v / %#v", i, got, delta)
					}
				}
			}
		})
	}
}
