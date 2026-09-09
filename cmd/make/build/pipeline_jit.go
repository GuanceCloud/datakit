// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2026-present Guance, Inc.

package build

import (
	"bytes"
	"crypto/sha256"
	"debug/elf"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	pljit "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline/jit"
)

const (
	pipelineJITEnabledEnv          = "PIPELINE_JIT_ENABLED"
	pipelineJITRuntimeDirEnv       = "PLATYPUS_JIT_RUNTIME_DIR"
	pipelineJITExpectedRevisionEnv = "PLATYPUS_JIT_EXPECTED_REVISION"
	pipelineJITManifestName        = "manifest.json"
	pipelineJITLibraryName         = "libplatypus_jit.so"
	pipelineJITStagedManifestName  = "platypus-jit-manifest.json"
	pipelineJITRuntimeABI          = "pp-jit-flat-v1+ppr3+pps2+modules-v2+services-v1+cancel-v1+cache-v1+aggregate-ppf1-v2"
	pipelineJITProfile             = "pipeline-go-1.4.3-datakit"
)

var pipelineJITRequiredSymbols = []string{
	"pp_jit_compile_flat_v1",
	"pp_jit_compile_configured_flat_v2",
	"pp_jit_compile_modules_flat_v1",
	"pp_jit_input_projection_flat_v1",
	"pp_jit_program_capabilities_flat_v1",
	"pp_jit_process_raw_batch_flat_v3",
	"pp_jit_static_output_schema_flat_v1",
	"pp_jit_process_static_batch_flat_v1",
	"pp_jit_destroy_v1",
	"pp_jit_buffer_free_flat_v1",
	"pp_jit_services_create_flat_v1",
	"pp_jit_services_destroy_v1",
	"pp_jit_compile_modules_with_services_flat_v2",
	"pp_jit_cache_ticks_enable_v1",
	"pp_jit_cache_tick_v1",
	"pp_jit_poll_aggregate_events_flat_v2",
	"pp_jit_cancellation_create_v1",
	"pp_jit_cancellation_cancel_v1",
	"pp_jit_cancellation_destroy_v1",
	"pp_jit_process_cancellable_flat_v1",
}

type pipelineJITManifest struct {
	SchemaVersion  int      `json:"schema_version"`
	SourceRevision string   `json:"source_revision"`
	Target         string   `json:"target"`
	Artifact       string   `json:"artifact"`
	SHA256         string   `json:"sha256"`
	RuntimeABI     string   `json:"runtime_abi"`
	Profile        string   `json:"profile"`
	Features       []string `json:"features"`
	RustToolchain  string   `json:"rust_toolchain"`
}

func pipelineJITEnabled() (bool, error) {
	switch strings.ToLower(strings.TrimSpace(os.Getenv(pipelineJITEnabledEnv))) {
	case "", "0", "false", "off", "no":
		return false, nil
	case "1", "true", "on", "yes":
		return true, nil
	default:
		return false, fmt.Errorf("%s must be one of 0, 1, false, true, off, or on", pipelineJITEnabledEnv)
	}
}

func pipelineJITTargetSupported(goos, goarch string) bool {
	return goos == "linux" && (goarch == "amd64" || goarch == "arm64")
}

func pipelineJITBuildTags(goos, goarch, tags string) (string, error) {
	enabled, err := pipelineJITEnabled()
	if err != nil {
		return "", err
	}
	if !enabled || !pipelineJITTargetSupported(goos, goarch) {
		return tags, nil
	}
	if tags == "" {
		return "pipeline_jit", nil
	}
	return tags + " pipeline_jit", nil
}

// stagePipelineJITRuntime copies an independently built platplusplus runtime
// into an explicitly enabled Pipeline JIT DataKit package. The artifact root
// uses linux-<arch> directories so CI can stage both release targets without
// relying on a sibling checkout. An enabled build fails closed when its source
// revision, manifest, checksum, ELF identity, or dynamic ABI is missing.
func stagePipelineJITRuntime(distDir, goos, goarch string) error {
	enabled, err := pipelineJITEnabled()
	if err != nil {
		return err
	}
	if !enabled || !pipelineJITTargetSupported(goos, goarch) {
		return nil
	}

	artifactRoot := strings.TrimSpace(os.Getenv(pipelineJITRuntimeDirEnv))
	if artifactRoot == "" {
		return fmt.Errorf("%s=1 requires %s", pipelineJITEnabledEnv, pipelineJITRuntimeDirEnv)
	}
	expectedRevision := strings.TrimSpace(os.Getenv(pipelineJITExpectedRevisionEnv))
	if expectedRevision == "" {
		return fmt.Errorf("%s=1 requires %s", pipelineJITEnabledEnv, pipelineJITExpectedRevisionEnv)
	}

	sourceDir := filepath.Join(artifactRoot, goos+"-"+goarch)
	source := filepath.Join(sourceDir, pipelineJITLibraryName)
	manifestPath := filepath.Join(sourceDir, pipelineJITManifestName)
	if _, err := validatePipelineJITRuntime(source, manifestPath, goarch, expectedRevision); err != nil {
		return err
	}

	libDir := filepath.Join(distDir, "lib")
	if err := os.MkdirAll(libDir, 0o755); err != nil {
		return fmt.Errorf("create Pipeline JIT library directory: %w", err)
	}
	for _, pair := range [][2]string{
		{source, filepath.Join(libDir, pipelineJITLibraryName)},
		{source + ".sha256", filepath.Join(libDir, pipelineJITLibraryName+".sha256")},
		{manifestPath, filepath.Join(libDir, pipelineJITStagedManifestName)},
	} {
		if err := copyPipelineJITFile(pair[0], pair[1]); err != nil {
			return err
		}
	}
	return nil
}

func copyPipelineJITFile(source, destination string) error {
	input, err := os.Open(source) //nolint:gosec
	if err != nil {
		return fmt.Errorf("open Pipeline JIT artifact %q: %w", source, err)
	}
	defer input.Close() //nolint:errcheck
	info, err := input.Stat()
	if err != nil {
		return fmt.Errorf("stat Pipeline JIT artifact %q: %w", source, err)
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("Pipeline JIT artifact %q is not a regular file", source)
	}

	output, err := os.OpenFile(destination, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644) //nolint:gosec
	if err != nil {
		return fmt.Errorf("create Pipeline JIT artifact %q: %w", destination, err)
	}
	if _, err := io.Copy(output, input); err != nil {
		_ = output.Close()
		return fmt.Errorf("copy Pipeline JIT artifact %q: %w", source, err)
	}
	if err := output.Close(); err != nil {
		return fmt.Errorf("close Pipeline JIT artifact %q: %w", destination, err)
	}
	return nil
}

func validatePipelineJITRuntime(path, manifestPath, goarch, expectedRevision string) (pipelineJITManifest, error) {
	manifest, err := readPipelineJITManifest(manifestPath, goarch, expectedRevision)
	if err != nil {
		return pipelineJITManifest{}, err
	}
	if manifest.Artifact != filepath.Base(path) {
		return pipelineJITManifest{}, fmt.Errorf("Pipeline JIT manifest artifact %q does not match %q", manifest.Artifact, filepath.Base(path))
	}

	payload, err := os.ReadFile(path) //nolint:gosec
	if err != nil {
		return pipelineJITManifest{}, fmt.Errorf("read Pipeline JIT runtime %q: %w", path, err)
	}
	digest := sha256.Sum256(payload)
	gotDigest := hex.EncodeToString(digest[:])
	if !strings.EqualFold(gotDigest, manifest.SHA256) {
		return pipelineJITManifest{}, fmt.Errorf("Pipeline JIT manifest checksum mismatch for %q", path)
	}
	if err := validatePipelineJITChecksumFile(path+".sha256", digest); err != nil {
		return pipelineJITManifest{}, err
	}
	if err := validatePipelineJITELF(path, goarch); err != nil {
		return pipelineJITManifest{}, err
	}
	return manifest, nil
}

func readPipelineJITManifest(path, goarch, expectedRevision string) (pipelineJITManifest, error) {
	payload, err := os.ReadFile(path) //nolint:gosec
	if err != nil {
		return pipelineJITManifest{}, fmt.Errorf("read Pipeline JIT manifest %q: %w", path, err)
	}
	var manifest pipelineJITManifest
	if err := json.Unmarshal(payload, &manifest); err != nil {
		return pipelineJITManifest{}, fmt.Errorf("decode Pipeline JIT manifest %q: %w", path, err)
	}
	if manifest.SchemaVersion != 1 {
		return pipelineJITManifest{}, fmt.Errorf("Pipeline JIT manifest %q has unsupported schema version %d", path, manifest.SchemaVersion)
	}
	if !validPipelineJITRevision(manifest.SourceRevision) {
		return pipelineJITManifest{}, fmt.Errorf("Pipeline JIT manifest %q has invalid source revision", path)
	}
	if !validPipelineJITRevision(expectedRevision) {
		return pipelineJITManifest{}, fmt.Errorf("%s must be a full 40- or 64-character hexadecimal revision", pipelineJITExpectedRevisionEnv)
	}
	if !strings.EqualFold(manifest.SourceRevision, expectedRevision) {
		return pipelineJITManifest{}, fmt.Errorf("Pipeline JIT source revision %s does not match expected %s", manifest.SourceRevision, expectedRevision)
	}
	wantTarget := map[string]string{
		"amd64": "x86_64-unknown-linux-gnu",
		"arm64": "aarch64-unknown-linux-gnu",
	}[goarch]
	if manifest.Target != wantTarget {
		return pipelineJITManifest{}, fmt.Errorf("Pipeline JIT manifest target %q does not match %s", manifest.Target, wantTarget)
	}
	if manifest.Artifact != pipelineJITLibraryName {
		return pipelineJITManifest{}, fmt.Errorf("Pipeline JIT manifest artifact %q is invalid", manifest.Artifact)
	}
	if digest, err := hex.DecodeString(manifest.SHA256); err != nil || len(digest) != sha256.Size {
		return pipelineJITManifest{}, fmt.Errorf("Pipeline JIT manifest %q has invalid SHA256", path)
	}
	if manifest.RuntimeABI != pipelineJITRuntimeABI {
		return pipelineJITManifest{}, fmt.Errorf("Pipeline JIT manifest runtime ABI %q does not match %q", manifest.RuntimeABI, pipelineJITRuntimeABI)
	}
	if manifest.Profile != pipelineJITProfile {
		return pipelineJITManifest{}, fmt.Errorf("Pipeline JIT manifest profile %q does not match %q", manifest.Profile, pipelineJITProfile)
	}
	if err := pljit.ValidateReleaseFeatures(manifest.Features, manifest.RustToolchain); err != nil {
		return pipelineJITManifest{}, err
	}
	return manifest, nil
}

func validPipelineJITRevision(revision string) bool {
	if len(revision) != 40 && len(revision) != 64 {
		return false
	}
	_, err := hex.DecodeString(revision)
	return err == nil
}

func validatePipelineJITChecksumFile(path string, got [sha256.Size]byte) error {
	payload, err := os.ReadFile(path) //nolint:gosec
	if err != nil {
		return fmt.Errorf("read Pipeline JIT checksum %q: %w", path, err)
	}
	fields := strings.Fields(string(payload))
	if len(fields) < 1 {
		return fmt.Errorf("Pipeline JIT checksum %q is empty", path)
	}
	want, err := hex.DecodeString(fields[0])
	if err != nil || len(want) != sha256.Size {
		return fmt.Errorf("Pipeline JIT checksum %q is invalid", path)
	}
	if !bytes.Equal(got[:], want) {
		return fmt.Errorf("Pipeline JIT checksum mismatch for %q", strings.TrimSuffix(path, ".sha256"))
	}
	return nil
}

func validatePipelineJITELF(path, goarch string) error {
	file, err := elf.Open(path)
	if err != nil {
		return fmt.Errorf("open Pipeline JIT ELF %q: %w", path, err)
	}
	defer file.Close() //nolint:errcheck
	if file.Class != elf.ELFCLASS64 || file.Data != elf.ELFDATA2LSB || file.Type != elf.ET_DYN {
		return fmt.Errorf("Pipeline JIT runtime %q must be a 64-bit little-endian ELF shared object", path)
	}
	wantMachine := elf.EM_X86_64
	if goarch == "arm64" {
		wantMachine = elf.EM_AARCH64
	}
	if file.Machine != wantMachine {
		return fmt.Errorf("Pipeline JIT runtime %q has ELF machine %s, expected %s for %s", path, file.Machine, wantMachine, goarch)
	}
	hasLoad, hasDynamic := false, false
	for _, program := range file.Progs {
		switch program.Type { //nolint:exhaustive // Only load and dynamic headers are required.
		case elf.PT_LOAD:
			hasLoad = true
		case elf.PT_DYNAMIC:
			hasDynamic = true
		}
	}
	if !hasLoad || !hasDynamic {
		return fmt.Errorf("Pipeline JIT runtime %q is missing ELF load or dynamic program headers", path)
	}
	symbols, err := file.DynamicSymbols()
	if err != nil {
		return fmt.Errorf("read Pipeline JIT dynamic symbols from %q: %w", path, err)
	}
	exported := make(map[string]struct{}, len(symbols))
	for _, symbol := range symbols {
		binding := elf.ST_BIND(symbol.Info)
		if symbol.Section != elf.SHN_UNDEF && elf.ST_TYPE(symbol.Info) == elf.STT_FUNC &&
			(binding == elf.STB_GLOBAL || binding == elf.STB_WEAK) {
			exported[symbol.Name] = struct{}{}
		}
	}
	for _, symbol := range pipelineJITRequiredSymbols {
		if _, ok := exported[symbol]; !ok {
			return fmt.Errorf("Pipeline JIT runtime %q is missing exported ABI symbol %q", path, symbol)
		}
	}
	return nil
}
