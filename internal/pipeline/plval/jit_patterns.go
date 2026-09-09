// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2026-present Guance, Inc.

package plval

import (
	"sync/atomic"

	"github.com/GuanceCloud/pipeline-go/manager"
	pljit "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline/jit"
)

type jitPatternRules struct {
	rules    map[string]string
	external bool
}

var loadedJITPatternRules atomic.Pointer[jitPatternRules]

// Pattern files are loaded during pipeline initialization. Publish a separate
// immutable snapshot so admission never reads the mutable runtime global map.
func publishJITPatternRules(rules map[string]string) {
	defaults := manager.CopyDefalutPatterns()
	next := &jitPatternRules{rules: make(map[string]string, len(rules)), external: len(rules) != len(defaults)}
	for name, value := range rules {
		next.rules[name] = value
		if original, found := defaults[name]; !found || original != value {
			next.external = true
		}
	}
	loadedJITPatternRules.Store(next)
}

func bindJITPatternRules(snapshot *pljit.ModuleSnapshot) *pljit.ModuleSnapshot {
	if rules := loadedJITPatternRules.Load(); rules != nil {
		if rules.external {
			return snapshot.WithExternalRules(rules.rules)
		}
		return snapshot.WithRules(rules.rules)
	}
	return snapshot.WithRules(manager.CopyDefalutPatterns())
}
