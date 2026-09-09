// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2026-present Guance, Inc.

package jit

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestReleaseArtifactChecksBeforeLoading(t *testing.T) {
	if runtime.GOOS != "linux" || (runtime.GOARCH != "amd64" && runtime.GOARCH != "arm64") {
		t.Skip("release targets Linux amd64/arm64")
	}
	dir := t.TempDir()
	path := filepath.Join(dir, "libplatypus_jit.so")
	data := []byte("not executable: verification test never calls dlopen")
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
	digest := fmt.Sprintf("%x", sha256.Sum256(data))
	manifest := map[string]any{"schema_version": 1, "source_revision": "0123456789012345678901234567890123456789",
		"sha256": digest, "target": map[string]string{"amd64": "x86_64-unknown-linux-gnu", "arm64": "aarch64-unknown-linux-gnu"}[runtime.GOARCH],
		"artifact": "libplatypus_jit.so", "profile": "pipeline-go-1.4.3-datakit", "runtime_abi": "pp-jit-flat-v1+ppr3+pps2+modules-v2+services-v1+cancel-v1+cache-v1+aggregate-ppf1-v2",
		"features": []string{"scalar-slots"}, "rust_toolchain": "1.97.1"}
	write := func() {
		t.Helper()
		payload, err := json.Marshal(manifest)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "manifest.json"), payload, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write()
	if _, err := VerifyRuntime(path, digest, true); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		key   string
		value any
	}{
		{"features", []string{}}, {"features", []string{"scalar-slots", "phase-profile"}},
		{"rust_toolchain", "unknown"}, {"target", "other"}, {"source_revision", "garbage"},
		{"profile", "other"}, {"sha256", fmt.Sprintf("%064x", 0)},
	} {
		old := manifest[tc.key]
		manifest[tc.key] = tc.value
		write()
		if _, err := VerifyRuntime(path, digest, true); err == nil {
			t.Fatalf("accepted %s=%v", tc.key, tc.value)
		}
		manifest[tc.key] = old
	}
	write()
	if _, err := VerifyRuntime(path, "", true); err == nil {
		t.Fatal("untrusted adjacent checksum accepted")
	}
	if err := os.WriteFile(path, []byte("tampered"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := VerifyRuntime(path, digest, true); err == nil {
		t.Fatal("tampering accepted")
	}
}
