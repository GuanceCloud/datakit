// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build pipeline_jit && linux && (amd64 || arm64)

// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2026-present Guance, Inc.

package pipeline

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/GuanceCloud/cliutils/point"
	"github.com/GuanceCloud/pipeline-go/constants"
	pljit "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline/jit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline/plval"
)

func TestRolloutNativeAdmissionAndPolicyReuse(t *testing.T) {
	input := os.Getenv("PLATYPUS_JIT_RUNTIME")
	if input == "" {
		t.Skip("set PLATYPUS_JIT_RUNTIME to the current scalar-slots library")
	}
	data, err := os.ReadFile(input)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "libplatypus_jit.so")
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
	digest := fmt.Sprintf("%x", sha256.Sum256(data))
	metadata, err := json.Marshal(map[string]any{
		"schema_version": 1, "source_revision": "0123456789012345678901234567890123456789", "sha256": digest,
		"target":   map[string]string{"amd64": "x86_64-unknown-linux-gnu", "arm64": "aarch64-unknown-linux-gnu"}[runtime.GOARCH],
		"artifact": "libplatypus_jit.so", "profile": "pipeline-go-1.4.3-datakit", "features": []string{"scalar-slots"}, "rust_toolchain": "1.97.1",
		"runtime_abi": "pp-jit-flat-v1+ppr3+pps2+modules-v2+services-v1+cancel-v1+cache-v1+aggregate-ppf1-v2",
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(filepath.Dir(path), "manifest.json"), metadata, 0o644); err != nil {
		t.Fatal(err)
	}
	original, _ := plval.GetManager()
	m := plval.NewScriptManager(nil, nil)
	plval.SetManager(m)
	t.Cleanup(func() {
		if err := closeJITGeneration(jitRunners.replace(nil)); err != nil {
			t.Error(err)
		}
		plval.SetManager(original)
	})
	sources := map[string]string{"main.p": "add_key(engine, \"executed\")", "cached.p": "cache_set(\"key\", 1)"}
	if err := m.LoadScriptWithCatChecked(point.Logging, constants.NSDefault, sources, nil); err != nil {
		t.Fatal(err)
	}
	m.UpdateDefaultScript(map[point.Category]string{point.Logging: "main.p"})
	lease, _ := m.Acquire()
	script, _ := lease.Manager().QueryScript(point.Logging, "main.p", struct{}{})
	snapshot, err := lease.JITModuleSnapshot(point.Logging, script)
	lease.Release()
	if err != nil {
		t.Fatal(err)
	}
	cfg := &plval.PipelineCfg{JIT: &plval.JITCfg{Enabled: true, Mode: "allowlist", RuntimePath: path, RuntimeSHA256: digest, MaxCachedPrograms: 16,
		Allow: []pljit.AllowRule{{Category: "logging", Namespace: "default", Script: "main.p", BundleSHA256: fmt.Sprintf("%x", snapshot.BundleIdentity())}}}}
	if err := initJIT(cfg, ""); err != nil {
		t.Fatal(err)
	}
	runnerLease := jitRunners.acquire()
	firstRunner := runnerLease.runner()
	runnerLease.release()
	lease, _ = m.Acquire()
	a := lease.JITAdmission(script)
	lease.Release()
	if a == nil || !a.Stateless {
		t.Fatal("native effect descriptor missing or incorrectly stateful")
	}
	run := func(wantNative bool) {
		t.Helper()
		counter := jitRouteRecordsVec.WithLabelValues("logging", "attempted", "submitted")
		before := readPrometheusCounter(t, counter)
		pt := point.NewPoint("test", point.NewKVs(map[string]any{"message": "x"}))
		if _, err := RunPlContext(context.Background(), point.Logging, []*point.Point{pt}, nil); err != nil {
			t.Fatal(err)
		}
		if pt.Get("engine") != "executed" {
			t.Fatal(pt.KVMap())
		}
		delta := readPrometheusCounter(t, counter) - before
		if (delta == 1) != wantNative {
			t.Fatalf("native submissions %v, wantNative %v", delta, wantNative)
		}
	}
	run(true)
	cfg.JIT.ForceGo = []pljit.ScriptIdentity{{Category: "logging", Namespace: "default", Script: "main.p"}}
	if err := initJIT(cfg, ""); err != nil {
		t.Fatal(err)
	}
	run(false)
	runnerLease = jitRunners.acquire()
	if runnerLease.runner() != firstRunner {
		t.Error("policy update replaced runtime")
	}
	runnerLease.release()
	current, _ := plval.GetManager()
	if current != m {
		t.Fatal("policy update replaced manager")
	}
	cfg.JIT.ForceGo = nil
	if err := initJIT(cfg, ""); err != nil {
		t.Fatal(err)
	}
	run(true)
	sources["main.p"] = "add_key(engine, \"executed\"); add_key(changed, true)"
	if err := m.LoadScriptWithCatChecked(point.Logging, constants.NSDefault, sources, nil); err != nil {
		t.Fatal(err)
	}
	run(false) // Changed bundle needs a new allow entry.
	lease, _ = m.Acquire()
	script, _ = lease.Manager().QueryScript(point.Logging, "main.p", struct{}{})
	snapshot, err = lease.JITModuleSnapshot(point.Logging, script)
	lease.Release()
	if err != nil {
		t.Fatal(err)
	}
	cfg.JIT.Allow[0].BundleSHA256 = fmt.Sprintf("%x", snapshot.BundleIdentity())
	if err := initJIT(cfg, ""); err != nil {
		t.Fatal(err)
	}
	run(true)
	cfg.JIT.Mode = "auto"
	cfg.JIT.Allow = nil
	if err := initJIT(cfg, ""); err == nil {
		t.Fatal("policy-only migration from Go shared state into native accepted")
	}
	run(true) // Rejected policy retained the last complete policy.
}
