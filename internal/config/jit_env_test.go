// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2026-present Guance, Inc.

package config

import (
	"strings"
	"testing"
)

func TestPipelineJITEnvironmentValidation(t *testing.T) {
	for _, key := range []string{"ENV_PIPELINE_JIT_ENABLED", "ENV_PIPELINE_JIT_MODE", "ENV_PIPELINE_JIT_ON_INIT_ERROR", "ENV_PIPELINE_JIT_RUNTIME_SHA256", "ENV_PIPELINE_JIT_MAX_CACHED_PROGRAMS"} {
		t.Setenv(key, "")
	}
	for _, tc := range []struct{ key, value string }{
		{"ENV_PIPELINE_JIT_ENABLED", "tru"}, {"ENV_PIPELINE_JIT_MODE", "random"},
		{"ENV_PIPELINE_JIT_ON_INIT_ERROR", "replay"}, {"ENV_PIPELINE_JIT_RUNTIME_SHA256", "bad"},
		{"ENV_PIPELINE_JIT_MAX_CACHED_PROGRAMS", "0"}, {"ENV_PIPELINE_JIT_MAX_CACHED_PROGRAMS", "-1"},
	} {
		t.Run(tc.key, func(t *testing.T) {
			t.Setenv(tc.key, tc.value)
			if err := DefaultConfig().loadPipelineEnvs(); err == nil {
				t.Fatal("invalid setting accepted")
			}
		})
	}
	t.Setenv("ENV_PIPELINE_JIT_ENABLED", "true")
	t.Setenv("ENV_PIPELINE_JIT_MODE", "allowlist")
	cfg := DefaultConfig()
	if err := cfg.loadPipelineEnvs(); err == nil {
		t.Fatal("untrusted allowlist accepted")
	}
	t.Setenv("ENV_PIPELINE_JIT_RUNTIME_SHA256", strings.Repeat("a", 64))
	t.Setenv("ENV_PIPELINE_JIT_ON_INIT_ERROR", "go")
	cfg = DefaultConfig()
	if err := cfg.loadPipelineEnvs(); err != nil {
		t.Fatal(err)
	}
	if !cfg.Pipeline.JIT.Enabled || cfg.Pipeline.JIT.Mode != "allowlist" || cfg.Pipeline.JIT.OnInitError != "go" {
		t.Fatal("environment overrides not applied")
	}
	t.Setenv("ENV_PIPELINE_JIT_ENABLED", "false")
	t.Setenv("ENV_PIPELINE_JIT_RUNTIME_SHA256", "")
	if err := DefaultConfig().loadPipelineEnvs(); err != nil {
		t.Fatal("kill switch requires loading identity", err)
	}
}
