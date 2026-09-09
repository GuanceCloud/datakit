// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2026-present Guance, Inc.

package jit

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
)

func TestRolloutPolicyIdentityPrecedenceAndEmpty(t *testing.T) {
	id := ScriptIdentity{"logging", "default", "main.p"}
	bundle := [32]byte{7}
	rule := AllowRule{id.Category, id.Namespace, id.Script, fmt.Sprintf("%x", bundle)}
	for _, tc := range []struct {
		mode   string
		allow  []AllowRule
		deny   []ScriptIdentity
		id     ScriptIdentity
		bundle [32]byte
		want   string
	}{
		{"", nil, nil, id, bundle, ""},
		{"allowlist", nil, nil, id, bundle, "not_allowed"},
		{"allowlist", []AllowRule{rule}, nil, id, bundle, ""},
		{"allowlist", []AllowRule{rule}, []ScriptIdentity{id}, id, bundle, "forced_go"},
		{"auto", nil, []ScriptIdentity{id}, id, bundle, "forced_go"},
		{"allowlist", []AllowRule{rule}, nil, id, [32]byte{8}, "bundle_changed"},
		{"allowlist", []AllowRule{rule}, nil, ScriptIdentity{"logging", "remote", "main.p"}, bundle, "not_allowed"},
	} {
		p, err := NewRolloutPolicy(tc.mode, tc.allow, tc.deny)
		if err != nil {
			t.Fatal(err)
		}
		if got := p.Check(tc.id, tc.bundle); got != tc.want {
			t.Fatalf("%+v: got %s", tc, got)
		}
	}
	for _, tc := range []struct {
		mode  string
		allow []AllowRule
		deny  []ScriptIdentity
	}{
		{"typo", nil, nil}, {"allowlist", []AllowRule{rule, rule}, nil},
		{"auto", nil, []ScriptIdentity{id, id}},
		{"allowlist", []AllowRule{{Category: "logging", Namespace: "default", Script: "*.p", BundleSHA256: rule.BundleSHA256}}, nil},
		{"allowlist", []AllowRule{{Category: "logging", Namespace: "default", Script: "main.p", BundleSHA256: "bad"}}, nil},
	} {
		if _, err := NewRolloutPolicy(tc.mode, tc.allow, tc.deny); err == nil {
			t.Fatalf("accepted invalid policy: %+v", tc)
		}
	}
}

func TestRolloutHealthPersistsWithoutPoisoningOtherVersions(t *testing.T) {
	var registry HealthRegistry
	runtime, bundle := [32]byte{1}, [32]byte{2}
	a := registry.Admit(runtime, bundle, true)
	var wg sync.WaitGroup
	for range 16 {
		wg.Add(1)
		go func() { defer wg.Done(); a.Quarantine(false); a.Decision() }()
	}
	wg.Wait()
	for _, current := range []*Admission{a, registry.Admit(runtime, bundle, true)} {
		if reason, blocked := current.Decision(); reason != "quarantined" || blocked {
			t.Fatal(reason, blocked)
		}
	}
	stateful := registry.Admit(runtime, bundle, false)
	if _, blocked := stateful.Decision(); !blocked {
		t.Fatal("shared state silently migrated")
	}
	next := registry.Admit([32]byte{3}, bundle, true)
	other := registry.Admit(runtime, [32]byte{4}, true)
	if r, _ := next.Decision(); r != "" {
		t.Fatal("new runtime poisoned")
	}
	if r, _ := other.Decision(); r != "" {
		t.Fatal("unrelated bundle poisoned")
	}
	a.Quarantine(true)
	if r, _ := other.Decision(); r != "quarantined" {
		t.Fatal("runtime fault not propagated")
	}
	if r, _ := next.Decision(); r != "" {
		t.Fatal("late old completion poisoned replacement runtime")
	}
	if allocs := testing.AllocsPerRun(100, func() { a.Decision() }); allocs != 0 {
		t.Fatal("admission allocates", allocs)
	}
}

func TestRolloutHealthCapacityDoesNotForgetIsolation(t *testing.T) {
	var registry HealthRegistry
	first := registry.Admit([32]byte{1}, [32]byte{}, true)
	first.Quarantine(false)
	for n := 1; n < MaxRolloutEntries; n++ {
		var digest [32]byte
		digest[0] = byte(n)
		digest[1] = byte(n >> 8)
		registry.Admit([32]byte{1}, digest, true)
	}
	if a := registry.Admit([32]byte{2}, [32]byte{}, true); a.Reason != "health_capacity" {
		t.Fatal("capacity exceeded")
	}
	if r, _ := registry.Admit([32]byte{1}, [32]byte{}, true).Decision(); r != "quarantined" {
		t.Fatal("quarantine evicted")
	}
}

func TestRolloutFaultClassification(t *testing.T) {
	for _, tc := range []struct {
		err         error
		fault, wide bool
	}{
		{errors.New("JIT panic fake user text"), false, false}, {context.Canceled, false, false},
		{context.DeadlineExceeded, false, false}, {&NativeCallError{Status: 3}, false, false},
		{&NativeCallError{Status: 4}, true, true}, {&NativeCallError{Status: 2}, true, true},
		{&ProtocolError{Cause: errors.New("invalid result")}, true, false},
	} {
		if f, w := EngineFault(tc.err); f != tc.fault || w != tc.wide {
			t.Fatalf("%v: %v/%v", tc.err, f, w)
		}
	}
}

func TestRolloutBundleBindsRulesAndDependencies(t *testing.T) {
	a, _ := NewScopedModuleSnapshot("default", "logging", "main.p", map[string]string{"main.p": "use(\"child.p\")", "child.p": "a=1"})
	b, _ := NewScopedModuleSnapshot("default", "logging", "main.p", map[string]string{"child.p": "a=1", "main.p": "use(\"child.p\")"})
	a = a.WithRules(map[string]string{"X": "a", "Y": "b"})
	b = b.WithRules(map[string]string{"Y": "b", "X": "a"})
	if a.BundleIdentity() != b.BundleIdentity() {
		t.Fatal("map iteration changed digest")
	}
	if a.BundleIdentity() == a.WithRules(map[string]string{"X": "changed", "Y": "b"}).BundleIdentity() {
		t.Fatal("rules not bound")
	}
	c, _ := NewScopedModuleSnapshot("default", "logging", "main.p", map[string]string{"main.p": "use(\"child.p\")", "child.p": "a=2"})
	if a.BundleIdentity() == c.WithRules(map[string]string{"X": "a", "Y": "b"}).BundleIdentity() {
		t.Fatal("dependency not bound")
	}
}
