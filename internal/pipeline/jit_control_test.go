// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2026-present Guance, Inc.

package pipeline

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"github.com/GuanceCloud/pipeline-go/constants"
	pljit "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline/jit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline/plval"
)

func TestRolloutExternalPatternsDoNotReachNativeCompiler(t *testing.T) {
	snapshot, err := pljit.NewScopedModuleSnapshot("default", "logging", "main.p", map[string]string{"main.p": "a=1"})
	if err != nil {
		t.Fatal(err)
	}
	snapshot = snapshot.WithExternalRules(map[string]string{"EXTERNAL": "[a-z]+"})
	policy, err := pljit.NewRolloutPolicy("allowlist", []pljit.AllowRule{{Category: "logging", Namespace: "default", Script: "main.p", BundleSHA256: fmt.Sprintf("%x", snapshot.BundleIdentity())}}, nil)
	if err != nil {
		t.Fatal(err)
	}
	control := &jitControl{policy: policy}
	control.active.Store(true)
	if admission := control.admission(nil, snapshot); admission.Reason != "pre_route" {
		t.Fatal("unbound external patterns reached native admission")
	}
}

type rolloutTestRunner struct {
	generationTestRunner
	calls int
	err   error
}

func (r *rolloutTestRunner) Process(string, []byte) (pljit.Batch, error) {
	r.calls++
	return pljit.Batch{}, r.err
}

func installRolloutTest(t *testing.T, admission *pljit.Admission, runner jitProcessor) *plval.ScriptManager {
	t.Helper()
	old, _ := plval.GetManager()
	manager := plval.NewScriptManager(nil, nil)
	if err := manager.LoadScriptWithCatChecked(point.Logging, constants.NSDefault, map[string]string{"main.p": "add_key(engine, \"go\")"}, nil); err != nil {
		t.Fatal(err)
	}
	manager.UpdateDefaultScript(map[point.Category]string{point.Logging: "main.p"})
	manager.SetJITCheck(func(string) bool { return true })
	manager.SetJITAdmission(func(*pljit.ModuleSnapshot) *pljit.Admission { return admission })
	plval.SetManager(manager)
	if previous := jitRunners.replace(runner); previous != nil {
		t.Fatal("unexpected existing runner")
	}
	t.Cleanup(func() {
		if err := closeJITGeneration(jitRunners.replace(nil)); err != nil {
			t.Error(err)
		}
		plval.SetManager(old)
	})
	return manager
}

func TestRolloutQuarantineChangesOnlyFutureExecution(t *testing.T) {
	for _, stateless := range []bool{true, false} {
		t.Run(map[bool]string{true: "stateless", false: "stateful"}[stateless], func(t *testing.T) {
			var health pljit.HealthRegistry
			admission := health.Admit([32]byte{1}, [32]byte{2}, stateless)
			runner := &rolloutTestRunner{err: &pljit.ProtocolError{Cause: errors.New("broken result")}}
			installRolloutTest(t, admission, runner)
			first := point.NewPoint("test", point.NewKVs(map[string]any{"message": "first"}))
			if _, err := RunPlContext(context.Background(), point.Logging, []*point.Point{first}, nil); err != nil {
				t.Fatal(err)
			}
			if first.Get("engine") != nil || first.Get(plStatus) != sFailed || runner.calls != 1 {
				t.Fatal("failed native record was replayed", first.KVMap(), runner.calls)
			}
			second := point.NewPoint("test", point.NewKVs(map[string]any{"message": "second"}))
			if _, err := RunPlContext(context.Background(), point.Logging, []*point.Point{second}, nil); err != nil {
				t.Fatal(err)
			}
			if runner.calls != 1 {
				t.Fatal("quarantined program executed again")
			}
			if stateless && second.Get("engine") != "go" {
				t.Fatal("future stateless record did not use Go")
			}
			if !stateless && (second.Get("engine") != nil || second.Get(plStatus) != sFailed) {
				t.Fatal("shared state silently migrated")
			}
		})
	}
}

func TestRolloutDisableDoesNotWaitForManagerOrRunnerLease(t *testing.T) {
	var health pljit.HealthRegistry
	runner := &rolloutTestRunner{}
	manager := installRolloutTest(t, health.Admit([32]byte{1}, [32]byte{2}, true), runner)
	managerLease, _ := manager.Acquire()
	runnerLease := jitRunners.acquire()
	defer managerLease.Release()
	defer runnerLease.release()
	done := make(chan error, 1)
	go func() { done <- initJIT(&plval.PipelineCfg{JIT: &plval.JITCfg{Enabled: false}}, "") }()
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(time.Second):
		t.Fatal("disable blocked on an old lease")
	}
	if runner.closed.Load() {
		t.Fatal("in-use runner closed")
	}
	if next := jitRunners.acquire(); next != nil {
		next.release()
		t.Fatal("disabled runner still published")
	}
	managerLease.Release()
	runnerLease.release()
	deadline := time.After(time.Second)
	for !runner.closed.Load() {
		select {
		case <-deadline:
			t.Fatal("retired runner not closed")
		case <-time.After(time.Millisecond):
		}
	}
}

func TestRolloutInitialFailureSelectsGo(t *testing.T) {
	if os.Getenv("JIT_ROLLOUT_INIT_CHILD") != "1" {
		for _, policy := range []string{"", "go", "error", "invalid"} {
			cmd := exec.Command(os.Args[0], "-test.run=^TestRolloutInitialFailureSelectsGo$", "-test.count=1")
			cmd.Env = append(os.Environ(), "JIT_ROLLOUT_INIT_CHILD=1", "JIT_ROLLOUT_INIT_POLICY="+policy)
			if output, err := cmd.CombinedOutput(); err != nil {
				t.Fatalf("policy %q: %v\n%s", policy, err, output)
			}
		}
		return
	}
	plval.SetManager(nil)
	cfg := &plval.PipelineCfg{JIT: &plval.JITCfg{Enabled: true, OnInitError: os.Getenv("JIT_ROLLOUT_INIT_POLICY"), RuntimePath: filepath.Join(t.TempDir(), "missing.so")}}
	if cfg.JIT.OnInitError == "error" || cfg.JIT.OnInitError == "invalid" {
		if err := InitPipeline(cfg, nil, nil, t.TempDir()); err == nil {
			t.Fatal("strict or invalid policy must fail initialization")
		}
		if manager, ok := plval.GetManager(); ok && manager != nil {
			t.Fatal("failed initialization published a manager")
		}
		return
	}
	if err := InitPipeline(cfg, nil, nil, t.TempDir()); err != nil {
		t.Fatal(err)
	}
	if !cfg.JIT.Enabled {
		t.Fatal("caller config mutated")
	}
	if lease := jitRunners.acquire(); lease != nil {
		lease.release()
		t.Fatal("native runner published after failure")
	}
	manager, ok := plval.GetManager()
	if !ok || manager == nil {
		t.Fatal("Go manager not initialized")
	}
	if err := manager.LoadScriptWithCatChecked(point.Logging, constants.NSDefault, map[string]string{"main.p": "add_key(engine, \"go\")"}, nil); err != nil {
		t.Fatal(err)
	}
	manager.UpdateDefaultScript(map[point.Category]string{point.Logging: "main.p"})
	pt := point.NewPoint("test", point.NewKVs(map[string]any{"message": "x"}))
	if _, err := RunPlContext(context.Background(), point.Logging, []*point.Point{pt}, nil); err != nil {
		t.Fatal(err)
	}
	if pt.Get("engine") != "go" {
		t.Fatal("initial fallback is not usable")
	}
	if err := InitPipeline(cfg, nil, nil, t.TempDir()); err == nil {
		t.Fatal("reload failure silently replaced an active Go generation")
	}
	current, _ := plval.GetManager()
	if current != manager {
		t.Fatal("failed reload replaced existing manager")
	}
}
