// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build pipeline_jit && linux && (amd64 || arm64)

// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// Copyright 2026-present Guance, Inc.

package pipeline

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"github.com/GuanceCloud/pipeline-go/constants"
	pljit "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline/jit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline/plval"
)

func TestPipelineRecordTimeoutGoAndNative(t *testing.T) {
	for _, native := range []bool{false, true} {
		t.Run(map[bool]string{false: "go", true: "jit"}[native], func(t *testing.T) {
			original, _ := plval.GetManager()
			m := plval.NewScriptManager(nil, nil)
			plval.SetManager(m)
			t.Cleanup(func() {
				if err := closeJITGeneration(jitRunners.replace(nil)); err != nil {
					t.Error(err)
				}
				plval.SetManager(original)
			})
			if err := m.LoadScriptWithCatChecked(point.Logging, constants.NSDefault, map[string]string{"main.p": "add_key(before,1);for ; repeat; {};add_key(after,2)"}, nil); err != nil {
				t.Fatal(err)
			}
			m.UpdateDefaultScript(map[point.Category]string{point.Logging: "main.p"})
			if err := initJIT(&plval.PipelineCfg{JIT: &plval.JITCfg{Enabled: native, RuntimePath: os.Getenv("PLATYPUS_JIT_RUNTIME")}}, ""); err != nil {
				t.Fatal(err)
			}
			for _, budget := range []time.Duration{0, time.Millisecond} {
				p := point.NewPoint("clock", point.NewKVs(map[string]any{"repeat": true}))
				counter := jitRouteRecordsVec.WithLabelValues("logging", "attempted", "submitted")
				before := readPrometheusCounter(t, counter)
				ctx := pljit.WithRecordTimeout(context.Background(), budget)
				start := time.Now()
				result, err := RunPlContext(ctx, point.Logging, []*point.Point{p}, nil)
				if err != nil {
					t.Fatal(err)
				}
				result.Release()
				if time.Since(start) > time.Second {
					t.Fatal("clock timeout failed to stop execution")
				}
				if native && readPrometheusCounter(t, counter) != before+1 {
					t.Fatal("native timeout silently routed to Go")
				}
				if (plval.EnableAppendRunInfo() && p.Get(plStatus) != sFailed) || p.Get("after") != nil {
					t.Fatal("timeout published completion", p.KVMap())
				}
				if budget == 0 && p.Get("before") != nil {
					t.Fatal("expired budget ran a statement")
				}
				if budget > 0 && p.Get("before") != int64(1) {
					t.Fatal("lost committed prefix", p.KVMap())
				}
				if ctx.Err() != nil || ctx.Done() != nil {
					t.Fatal("clock-only budget mutated context cancellation")
				}
			}
		})
	}
}
