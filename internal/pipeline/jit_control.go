// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2026-present Guance, Inc.

package pipeline

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"sync/atomic"
	"time"

	pljit "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline/jit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline/plval"
)

var jitHealth pljit.HealthRegistry

var jitPolicyReasons = [...]string{"policy_disabled", "forced_go", "not_allowed", "bundle_changed", "stateful_or_unknown", "health_capacity", "quarantined"}

func checkJITRetirementCapacity() error {
	jitRunners.mu.Lock()
	defer jitRunners.mu.Unlock()
	if jitRunners.pending >= 8 {
		return fmt.Errorf("pipeline JIT has %d undrained generations; disable remains available", jitRunners.pending)
	}
	return nil
}

type jitControl struct {
	active            atomic.Bool
	policy            *pljit.RolloutPolicy
	runtimeSHA        [32]byte
	settingsSHA       [32]byte
	owner             *plval.ScriptManager
	aggregatePrepared bool
}

func prepareJITControl(cfg *plval.PipelineCfg, path string) (*jitControl, error) {
	policy, err := pljit.NewRolloutPolicy(cfg.JIT.Mode, cfg.JIT.Allow, cfg.JIT.ForceGo)
	if err != nil {
		return nil, err
	}
	digest, err := pljit.VerifyRuntime(path, cfg.JIT.RuntimeSHA256, policy.RequiresStateless())
	if err != nil {
		return nil, fmt.Errorf("verify pipeline JIT runtime before loading: %w", err)
	}
	control := &jitControl{policy: policy, runtimeSHA: digest}
	settings, err := json.Marshal(struct {
		Path     string
		Max      int
		Services pljit.ServiceConfig
	}{path, cfg.JIT.MaxCachedPrograms, jitServiceConfig(cfg)})
	if err != nil {
		return nil, err
	}
	control.settingsSHA = sha256.Sum256(settings)
	control.active.Store(true)
	return control, nil
}

func reuseJITForPolicy(control *jitControl) (bool, error) {
	jitRunners.mu.Lock()
	current := jitRunners.current
	jitRunners.mu.Unlock()
	if current == nil || current.control == nil || current.control.runtimeSHA != control.runtimeSHA ||
		current.control.settingsSHA != control.settingsSHA {
		return false, nil
	}
	runner, ok := current.runner.(*pljit.Runner)
	if !ok {
		return false, nil
	}
	manager, ok := plval.GetManager()
	if !ok || manager == nil || current.control.owner != manager {
		return false, nil
	}
	if err := manager.ReplaceJITAdmission(func(snapshot *pljit.ModuleSnapshot) *pljit.Admission {
		return control.admission(runner, snapshot)
	}); err != nil {
		return true, err
	}
	control.owner = manager
	// Existing source hooks reference the retired control and now reject cold
	// checks. The new admission hook owns checking against the unchanged runner.
	current.control.active.Store(false)
	jitRunners.mu.Lock()
	current.control = control
	jitRunners.mu.Unlock()
	return true, nil
}

func (control *jitControl) admission(runner *pljit.Runner, snapshot *pljit.ModuleSnapshot) *pljit.Admission {
	bundle := snapshot.BundleIdentity()
	if !control.active.Load() {
		return &pljit.Admission{Bundle: bundle, Reason: "policy_disabled"}
	}
	if reason := control.policy.Check(snapshot.ScriptIdentity(), bundle); reason != "" {
		return &pljit.Admission{Bundle: bundle, Reason: reason}
	}
	if control.policy.RequiresStateless() && snapshot.RequiresExternalRules() {
		return &pljit.Admission{Bundle: bundle, Reason: "pre_route"}
	}
	var result pljit.CheckResult
	started := time.Now()
	if snapshot.HasDependencies() {
		result = runner.CheckModules(snapshot)
	} else {
		result = runner.Check(snapshot.EntrySource())
	}
	observeJITCheckDuration(result, time.Since(started))
	if result.Route != pljit.RouteJITNative && result.Route != pljit.RouteJITWithHost {
		return &pljit.Admission{Bundle: bundle, Reason: "pre_route"}
	}
	stateless := result.Capabilities.Stateless != nil && *result.Capabilities.Stateless
	if control.policy.RequiresStateless() && !stateless {
		return &pljit.Admission{Bundle: bundle, Reason: "stateful_or_unknown"}
	}
	return jitHealth.Admit(control.runtimeSHA, bundle, stateless)
}

func hasPublishedPipeline() bool {
	if manager, ok := plval.GetManager(); ok && manager != nil {
		return true
	}
	jitRunners.mu.Lock()
	defer jitRunners.mu.Unlock()
	return jitRunners.current != nil
}

func jitCanInitiallyDegrade(cfg *plval.PipelineCfg) bool {
	return cfg != nil && cfg.JIT != nil && (cfg.JIT.OnInitError == "" || cfg.JIT.OnInitError == "go") && !hasPublishedPipeline()
}

func jitDisabledConfig(cfg *plval.PipelineCfg) *plval.PipelineCfg {
	next := *cfg
	settings := *cfg.JIT
	settings.Enabled = false
	next.JIT = &settings
	return &next
}
