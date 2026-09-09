// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2026-present Guance, Inc.

package jit

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// ValidateReleaseFeatures binds the production package to the measured build.
// New features require an explicit review here; diagnostic timers never ship.
func ValidateReleaseFeatures(features []string, toolchain string) error {
	if len(features) != 1 || features[0] != "scalar-slots" {
		return fmt.Errorf("JIT release requires exactly features=[scalar-slots], got %v", features)
	}
	if toolchain != "1.97.1" {
		return fmt.Errorf("JIT release requires Rust toolchain 1.97.1, got %q", toolchain)
	}
	return nil
}

// VerifyRuntime runs before dlopen in the production adapter. In allowlist
// mode the trusted deployment config pins the SO digest; the adjacent manifest
// describes that artifact but is not itself a trust anchor. Install immutable
// version directories: verification does not protect files writable by an attacker.
func VerifyRuntime(path, expectedSHA string, requireRelease bool) ([32]byte, error) {
	var zero [32]byte
	f, err := os.Open(path)
	if err != nil {
		return zero, err
	}
	defer f.Close() //nolint:errcheck
	info, err := f.Stat()
	if err != nil {
		return zero, err
	}
	if !info.Mode().IsRegular() {
		return zero, fmt.Errorf("JIT runtime must be a regular file")
	}
	const maxRuntimeBytes = 256 << 20
	if info.Size() <= 0 || info.Size() > maxRuntimeBytes {
		return zero, fmt.Errorf("JIT runtime must be between 1 byte and 256 MiB")
	}
	hash := sha256.New()
	if n, err := io.Copy(hash, io.LimitReader(f, maxRuntimeBytes+1)); err != nil {
		return zero, err
	} else if n > maxRuntimeBytes {
		return zero, fmt.Errorf("JIT runtime grew beyond 256 MiB")
	}
	var digest [32]byte
	copy(digest[:], hash.Sum(nil))
	if expectedSHA != "" {
		want, err := ParseSHA256(expectedSHA)
		if err != nil {
			return zero, err
		}
		if digest != want {
			return zero, fmt.Errorf("JIT runtime checksum does not match trusted runtime_sha256")
		}
	} else if requireRelease {
		return zero, fmt.Errorf("JIT release requires a trusted runtime_sha256")
	}
	if !requireRelease {
		return digest, nil
	}
	if !filepath.IsAbs(path) {
		return zero, fmt.Errorf("JIT release runtime_path must be absolute")
	}
	if info.Mode().Perm()&0o022 != 0 {
		return zero, fmt.Errorf("JIT release runtime must not be group/world writable")
	}
	var manifest struct {
		SchemaVersion  int      `json:"schema_version"`
		SourceRevision string   `json:"source_revision"`
		SHA256         string   `json:"sha256"`
		Target         string   `json:"target"`
		Artifact       string   `json:"artifact"`
		Profile        string   `json:"profile"`
		RuntimeABI     string   `json:"runtime_abi"`
		Features       []string `json:"features"`
		Toolchain      string   `json:"rust_toolchain"`
	}
	manifestPath := filepath.Join(filepath.Dir(path), "platypus-jit-manifest.json")
	metadata, err := os.Open(manifestPath)
	if os.IsNotExist(err) {
		metadata, err = os.Open(filepath.Join(filepath.Dir(path), "manifest.json"))
	}
	if err != nil {
		return zero, err
	}
	defer metadata.Close() //nolint:errcheck
	data, err := io.ReadAll(io.LimitReader(metadata, 65537))
	if err != nil {
		return zero, err
	}
	if len(data) > 65536 {
		return zero, fmt.Errorf("JIT release manifest exceeds 64 KiB")
	}
	if err := json.Unmarshal(data, &manifest); err != nil {
		return zero, err
	}
	target := map[string]string{"amd64": "x86_64-unknown-linux-gnu", "arm64": "aarch64-unknown-linux-gnu"}[runtime.GOARCH]
	if runtime.GOOS != "linux" || target == "" || manifest.Target != target || manifest.SchemaVersion != 1 ||
		manifest.Artifact != filepath.Base(path) || manifest.Profile != "pipeline-go-1.4.3-datakit" ||
		manifest.RuntimeABI != "pp-jit-flat-v1+ppr3+pps2+modules-v2+services-v1+cancel-v1+cache-v1+aggregate-ppf1-v2" {
		return zero, fmt.Errorf("JIT release manifest identity, platform, ABI or profile mismatch")
	}
	if len(manifest.SourceRevision) != 40 && len(manifest.SourceRevision) != 64 {
		return zero, fmt.Errorf("invalid JIT release source revision")
	}
	for _, c := range strings.ToLower(manifest.SourceRevision) {
		if !strings.ContainsRune("0123456789abcdef", c) {
			return zero, fmt.Errorf("invalid JIT release source revision")
		}
	}
	if !strings.EqualFold(manifest.SHA256, fmt.Sprintf("%x", digest)) {
		return zero, fmt.Errorf("JIT release manifest checksum mismatch")
	}
	if err := ValidateReleaseFeatures(manifest.Features, manifest.Toolchain); err != nil {
		return zero, err
	}
	return digest, nil
}
