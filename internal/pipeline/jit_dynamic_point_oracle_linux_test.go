// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build pipeline_jit && linux && (amd64 || arm64)

package pipeline

import (
	"os"
	"testing"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"github.com/GuanceCloud/pipeline-go/lang"
	"github.com/GuanceCloud/pipeline-go/lang/platypus"
	pljit "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline/jit"
)

func TestJITDynamicPointUpstreamOracle(t *testing.T) {
	path := os.Getenv("PLATYPUS_JIT_RUNTIME")
	if path == "" {
		t.Fatal("requires real PLATYPUS_JIT_RUNTIME")
	}
	runner, err := pljit.NewRunnerWithHost(path, "pipeline-go-1.4.3-datakit", 8, pljit.NewPipelineGoHost(nil))
	if err != nil {
		t.Fatal(err)
	}
	defer runner.Close()
	cases := []struct {
		name, source string
		dropped      bool
	}{
		{name: "dynamic_read_write", source: `key = pt_kvs_get("selector"); pt_kvs_set(key, pt_kvs_get(key) + 1); pt_kvs_set(key + "_copy", pt_kvs_get(key))`},
		{name: "dynamic_json", source: `doc = load_json(message); pt_kvs_set(doc["key"], doc["value"]); pt_kvs_set("readback", pt_kvs_get(doc["key"]))`},
		{name: "dynamic_grok", source: `grok(message, "%{GREEDYDATA:captured}"); key = pt_kvs_get("selector"); pt_kvs_set(key, pt_kvs_get("captured"))`},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			scripts, failures := platypus.NewScripts(map[string]string{"main.p": test.source}, lang.WithCat(point.Logging))
			if len(failures) != 0 || len(scripts) != 1 {
				t.Fatalf("pipeline-go compile failures=%v scripts=%d", failures, len(scripts))
			}
			script := scripts["main.p"]
			if script == nil {
				t.Fatal("pipeline-go did not return main.p")
			}
			defer script.Cleanup()
			expected := newRealScriptPoint("nested-control", map[string]any{"message": `{"key":"customer","value":42}`, "selector": "counter", "counter": int64(7), "untouched": "keep", "pattern": `%{GREEDYDATA:captured}`})
			run := pointRun{point: expected, script: script, started: time.Now()}
			runPipelineGo(point.Logging, &run, nil)

			if check := runner.Check(test.source); check.Route != pljit.RouteJITNative {
				t.Fatalf("nested control did not route native: %#v", check)
			}
			actual := newRealScriptPoint("nested-control", map[string]any{"message": `{"key":"customer","value":42}`, "selector": "counter", "counter": int64(7), "untouched": "keep", "pattern": `%{GREEDYDATA:captured}`})
			projection, err := runner.Projection(test.source)
			if err != nil {
				t.Fatal(err)
			}
			if !projection.IsAll() {
				t.Fatal("dynamic read must preserve all input keys")
			}
			input, err := encodeProjectedJITPoints(point.Logging, []*point.Point{actual}, projection)
			if err != nil {
				t.Fatal(err)
			}
			batch, err := runner.Process(test.source, input)
			if err != nil || len(batch.Records) != 1 {
				t.Fatalf("native process records=%d err=%v", len(batch.Records), err)
			}
			if batch.Protocol != "dynamic-v3" {
				t.Fatalf("dynamic protocol not selected: %s", batch.Protocol)
			}
			record := batch.Records[0]
			if (record.Status == pljit.TerminalDropped) != test.dropped {
				t.Fatalf("native terminal=%v want dropped=%v", record.Status, test.dropped)
			}
			if test.dropped {
				if record.Point != nil || record.HasMutations() {
					t.Fatalf("dropped record must not publish a Point or delta: %#v", record)
				}
				return
			}
			var nativeDropped bool
			if batch.Static != nil {
				_, nativeDropped, err = applyJITStatic(point.Logging, actual, batch.Static, 0, nil)
			} else {
				_, nativeDropped, err = applyJITRecord(point.Logging, actual, record, 0, nil)
			}
			if err != nil || nativeDropped != test.dropped {
				t.Fatalf("apply dropped=%v err=%v", nativeDropped, err)
			}
			if equal, reason := equalJSONOraclePoints(actual, expected); !equal {
				t.Fatalf("%s\nactual=%s\nexpected=%s", reason, actual.Pretty(), expected.Pretty())
			}
		})
	}
}
