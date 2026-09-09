// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2026-present Guance, Inc.

package plval

import (
	"testing"

	"github.com/GuanceCloud/pipeline-go/manager"
	pljit "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline/jit"
)

func TestJITApprovalIncludesExternalPatternRules(t *testing.T) {
	previous := loadedJITPatternRules.Load()
	t.Cleanup(func() { loadedJITPatternRules.Store(previous) })
	snapshot, err := pljit.NewScopedModuleSnapshot("default", "logging", "main.p", map[string]string{"main.p": "a=1"})
	if err != nil {
		t.Fatal(err)
	}
	rules := manager.CopyDefalutPatterns()
	publishJITPatternRules(rules)
	plain := bindJITPatternRules(snapshot)
	if plain.RequiresExternalRules() {
		t.Fatal("builtin rules rejected")
	}
	rules["ROLLOUT_EXTERNAL"] = "[a-z]+"
	publishJITPatternRules(rules)
	external := bindJITPatternRules(snapshot)
	if !external.RequiresExternalRules() || plain.BundleIdentity() == external.BundleIdentity() {
		t.Fatal("external rules omitted from admission")
	}
	rules["ROLLOUT_EXTERNAL"] = "[0-9]+"
	if bindJITPatternRules(snapshot).BundleIdentity() != external.BundleIdentity() {
		t.Fatal("caller mutated published rules")
	}
	publishJITPatternRules(rules)
	if bindJITPatternRules(snapshot).BundleIdentity() == external.BundleIdentity() {
		t.Fatal("rule update did not invalidate approval")
	}
}
