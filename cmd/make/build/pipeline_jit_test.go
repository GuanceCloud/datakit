// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2026-present Guance, Inc.

package build

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

const testPipelineJITRevision = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

func alignTestELF(value, alignment int) int {
	return (value + alignment - 1) &^ (alignment - 1)
}

// testPipelineJITELF builds a structurally valid ELF64 ET_DYN with a real
// .dynstr/.dynsym pair. It is intentionally not executable; staging validation
// verifies artifact identity and exported ABI while target containers perform
// the loader smoke test.
func testPipelineJITELF(goarch string, symbols []string) []byte {
	dynstr := []byte{0}
	nameOffsets := make([]uint32, len(symbols))
	for index, symbol := range symbols {
		nameOffsets[index] = uint32(len(dynstr))
		dynstr = append(dynstr, symbol...)
		dynstr = append(dynstr, 0)
	}
	dynsym := make([]byte, (len(symbols)+1)*24)
	for index, nameOffset := range nameOffsets {
		entry := dynsym[(index+1)*24:]
		binary.LittleEndian.PutUint32(entry[0:4], nameOffset)
		entry[4] = 0x12 // STB_GLOBAL | STT_FUNC
		binary.LittleEndian.PutUint16(entry[6:8], 1)
	}

	const programHeaderBytes = 2 * 56
	dynstrOffset := 64 + programHeaderBytes
	dynsymOffset := alignTestELF(dynstrOffset+len(dynstr), 8)
	sectionOffset := alignTestELF(dynsymOffset+len(dynsym), 8)
	payload := make([]byte, sectionOffset+3*64)
	copy(payload[0:4], "\x7fELF")
	payload[4], payload[5], payload[6] = 2, 1, 1
	binary.LittleEndian.PutUint16(payload[16:18], 3) // ET_DYN
	machine := uint16(62)                            // EM_X86_64
	if goarch == "arm64" {
		machine = 183 // EM_AARCH64
	}
	binary.LittleEndian.PutUint16(payload[18:20], machine)
	binary.LittleEndian.PutUint32(payload[20:24], 1)
	binary.LittleEndian.PutUint64(payload[32:40], 64)
	binary.LittleEndian.PutUint64(payload[40:48], uint64(sectionOffset))
	binary.LittleEndian.PutUint16(payload[52:54], 64)
	binary.LittleEndian.PutUint16(payload[54:56], 56)
	binary.LittleEndian.PutUint16(payload[56:58], 2)
	binary.LittleEndian.PutUint16(payload[58:60], 64)
	binary.LittleEndian.PutUint16(payload[60:62], 3)
	loadHeader := payload[64:120]
	binary.LittleEndian.PutUint32(loadHeader[0:4], 1) // PT_LOAD
	binary.LittleEndian.PutUint32(loadHeader[4:8], 4) // PF_R
	binary.LittleEndian.PutUint64(loadHeader[32:40], uint64(len(payload)))
	binary.LittleEndian.PutUint64(loadHeader[40:48], uint64(len(payload)))
	binary.LittleEndian.PutUint64(loadHeader[48:56], 0x1000)
	dynamicHeader := payload[120:176]
	binary.LittleEndian.PutUint32(dynamicHeader[0:4], 2) // PT_DYNAMIC
	binary.LittleEndian.PutUint32(dynamicHeader[4:8], 4) // PF_R
	binary.LittleEndian.PutUint64(dynamicHeader[8:16], uint64(dynstrOffset))
	binary.LittleEndian.PutUint64(dynamicHeader[32:40], 16)
	binary.LittleEndian.PutUint64(dynamicHeader[40:48], 16)
	binary.LittleEndian.PutUint64(dynamicHeader[48:56], 8)
	copy(payload[dynstrOffset:], dynstr)
	copy(payload[dynsymOffset:], dynsym)

	dynstrHeader := payload[sectionOffset+64 : sectionOffset+128]
	binary.LittleEndian.PutUint32(dynstrHeader[4:8], 3) // SHT_STRTAB
	binary.LittleEndian.PutUint64(dynstrHeader[24:32], uint64(dynstrOffset))
	binary.LittleEndian.PutUint64(dynstrHeader[32:40], uint64(len(dynstr)))
	binary.LittleEndian.PutUint64(dynstrHeader[48:56], 1)
	dynsymHeader := payload[sectionOffset+128 : sectionOffset+192]
	binary.LittleEndian.PutUint32(dynsymHeader[4:8], 11) // SHT_DYNSYM
	binary.LittleEndian.PutUint64(dynsymHeader[24:32], uint64(dynsymOffset))
	binary.LittleEndian.PutUint64(dynsymHeader[32:40], uint64(len(dynsym)))
	binary.LittleEndian.PutUint32(dynsymHeader[40:44], 1)
	binary.LittleEndian.PutUint32(dynsymHeader[44:48], 1)
	binary.LittleEndian.PutUint64(dynsymHeader[48:56], 8)
	binary.LittleEndian.PutUint64(dynsymHeader[56:64], 24)
	return payload
}

func writeTestPipelineJITRuntime(t *testing.T, root, goarch string, symbols []string) []byte {
	t.Helper()
	directory := filepath.Join(root, "linux-"+goarch)
	if err := os.MkdirAll(directory, 0o755); err != nil {
		t.Fatal(err)
	}
	payload := testPipelineJITELF(goarch, symbols)
	path := filepath.Join(directory, pipelineJITLibraryName)
	if err := os.WriteFile(path, payload, 0o644); err != nil {
		t.Fatal(err)
	}
	digest := sha256.Sum256(payload)
	if err := os.WriteFile(path+".sha256", []byte(fmt.Sprintf("%x  %s\n", digest, pipelineJITLibraryName)), 0o644); err != nil {
		t.Fatal(err)
	}
	target := "x86_64-unknown-linux-gnu"
	if goarch == "arm64" {
		target = "aarch64-unknown-linux-gnu"
	}
	manifest, err := json.Marshal(pipelineJITManifest{
		SchemaVersion: 1,
		Features:      []string{"scalar-slots"}, RustToolchain: "1.97.1",
		SourceRevision: testPipelineJITRevision,
		Target:         target,
		Artifact:       pipelineJITLibraryName,
		SHA256:         fmt.Sprintf("%x", digest),
		RuntimeABI:     pipelineJITRuntimeABI,
		Profile:        pipelineJITProfile,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(directory, pipelineJITManifestName), manifest, 0o644); err != nil {
		t.Fatal(err)
	}
	return payload
}

func enableTestPipelineJIT(t *testing.T, artifacts string) {
	t.Helper()
	t.Setenv(pipelineJITEnabledEnv, "1")
	t.Setenv(pipelineJITRuntimeDirEnv, artifacts)
	t.Setenv(pipelineJITExpectedRevisionEnv, testPipelineJITRevision)
}

func TestStagePipelineJITRuntime(t *testing.T) {
	artifacts := t.TempDir()
	dist := t.TempDir()
	want := writeTestPipelineJITRuntime(t, artifacts, "amd64", pipelineJITRequiredSymbols)
	enableTestPipelineJIT(t, artifacts)
	if err := stagePipelineJITRuntime(dist, "linux", "amd64"); err != nil {
		t.Fatalf("stage Pipeline JIT runtime: %v", err)
	}
	got, err := os.ReadFile(filepath.Join(dist, "lib", pipelineJITLibraryName))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(want) {
		t.Fatal("staged runtime differs from source")
	}
	for _, name := range []string{pipelineJITLibraryName + ".sha256", pipelineJITStagedManifestName} {
		if _, err := os.Stat(filepath.Join(dist, "lib", name)); err != nil {
			t.Fatalf("staged audit artifact %s: %v", name, err)
		}
	}
}

func TestStagePipelineJITRuntimeRejectsStringOnlyFake(t *testing.T) {
	artifacts := t.TempDir()
	sourceDir := filepath.Join(artifacts, "linux-amd64")
	if err := os.MkdirAll(sourceDir, 0o755); err != nil {
		t.Fatal(err)
	}
	fake := []byte("\x7fELF")
	for _, symbol := range pipelineJITRequiredSymbols {
		fake = append(fake, symbol...)
		fake = append(fake, 0)
	}
	if err := os.WriteFile(filepath.Join(sourceDir, pipelineJITLibraryName), fake, 0o644); err != nil {
		t.Fatal(err)
	}
	digest := sha256.Sum256(fake)
	if err := os.WriteFile(filepath.Join(sourceDir, pipelineJITLibraryName+".sha256"), []byte(fmt.Sprintf("%x  %s\n", digest, pipelineJITLibraryName)), 0o644); err != nil {
		t.Fatal(err)
	}
	manifest, err := json.Marshal(pipelineJITManifest{
		SchemaVersion: 1,
		Features:      []string{"scalar-slots"}, RustToolchain: "1.97.1",
		SourceRevision: testPipelineJITRevision,
		Target:         "x86_64-unknown-linux-gnu",
		Artifact:       pipelineJITLibraryName,
		SHA256:         fmt.Sprintf("%x", digest),
		RuntimeABI:     pipelineJITRuntimeABI,
		Profile:        pipelineJITProfile,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sourceDir, pipelineJITManifestName), manifest, 0o644); err != nil {
		t.Fatal(err)
	}
	enableTestPipelineJIT(t, artifacts)
	if err := stagePipelineJITRuntime(t.TempDir(), "linux", "amd64"); err == nil {
		t.Fatal("string-only fake Pipeline JIT runtime was accepted")
	}
}

func TestStagePipelineJITRuntimeRejectsMissingDynamicSymbol(t *testing.T) {
	artifacts := t.TempDir()
	writeTestPipelineJITRuntime(t, artifacts, "amd64", pipelineJITRequiredSymbols[:len(pipelineJITRequiredSymbols)-1])
	enableTestPipelineJIT(t, artifacts)
	if err := stagePipelineJITRuntime(t.TempDir(), "linux", "amd64"); err == nil {
		t.Fatal("Pipeline JIT runtime with missing dynamic symbol was accepted")
	}
}

func TestStagePipelineJITRuntimeRequiresConfiguredArtifact(t *testing.T) {
	enableTestPipelineJIT(t, t.TempDir())
	if err := stagePipelineJITRuntime(t.TempDir(), "linux", "arm64"); err == nil {
		t.Fatal("missing configured Pipeline JIT runtime was accepted")
	}
}

func TestStagePipelineJITRuntimeRequiresExpectedRevision(t *testing.T) {
	artifacts := t.TempDir()
	writeTestPipelineJITRuntime(t, artifacts, "amd64", pipelineJITRequiredSymbols)
	t.Setenv(pipelineJITEnabledEnv, "1")
	t.Setenv(pipelineJITRuntimeDirEnv, artifacts)
	t.Setenv(pipelineJITExpectedRevisionEnv, "")
	if err := stagePipelineJITRuntime(t.TempDir(), "linux", "amd64"); err == nil {
		t.Fatal("Pipeline JIT runtime without an expected revision was accepted")
	}
}

func TestStagePipelineJITRuntimeRejectsStaleRevision(t *testing.T) {
	artifacts := t.TempDir()
	writeTestPipelineJITRuntime(t, artifacts, "amd64", pipelineJITRequiredSymbols)
	enableTestPipelineJIT(t, artifacts)
	t.Setenv(pipelineJITExpectedRevisionEnv, "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb")
	if err := stagePipelineJITRuntime(t.TempDir(), "linux", "amd64"); err == nil {
		t.Fatal("Pipeline JIT runtime from an unexpected revision was accepted")
	}
}

func TestStagePipelineJITRuntimeSkipsDisabledAndUnsupportedTargets(t *testing.T) {
	t.Setenv(pipelineJITEnabledEnv, "0")
	if err := stagePipelineJITRuntime(t.TempDir(), "linux", "amd64"); err != nil {
		t.Fatalf("disabled Pipeline JIT should be skipped: %v", err)
	}
	t.Setenv(pipelineJITEnabledEnv, "1")
	if err := stagePipelineJITRuntime(t.TempDir(), "windows", "amd64"); err != nil {
		t.Fatalf("unsupported target should be skipped: %v", err)
	}
}

func TestPipelineJITBuildTags(t *testing.T) {
	t.Setenv(pipelineJITEnabledEnv, "1")
	tags, err := pipelineJITBuildTags("linux", "amd64", "with_inputs")
	if err != nil {
		t.Fatal(err)
	}
	if tags != "with_inputs pipeline_jit" {
		t.Fatalf("build tags = %q", tags)
	}
	tags, err = pipelineJITBuildTags("linux", "386", "with_inputs")
	if err != nil {
		t.Fatal(err)
	}
	if tags != "with_inputs" {
		t.Fatalf("unsupported build tags = %q", tags)
	}
}

func TestConfiguredPipelineJITRuntimeArtifacts(t *testing.T) {
	artifactRoot := os.Getenv(pipelineJITRuntimeDirEnv)
	expectedRevision := os.Getenv(pipelineJITExpectedRevisionEnv)
	if artifactRoot == "" || expectedRevision == "" {
		t.Skipf("set %s and %s to validate release artifacts", pipelineJITRuntimeDirEnv, pipelineJITExpectedRevisionEnv)
	}
	t.Setenv(pipelineJITEnabledEnv, "1")
	for _, goarch := range []string{"amd64", "arm64"} {
		t.Run(goarch, func(t *testing.T) {
			directory := filepath.Join(artifactRoot, "linux-"+goarch)
			path := filepath.Join(directory, pipelineJITLibraryName)
			if _, err := validatePipelineJITRuntime(path, filepath.Join(directory, pipelineJITManifestName), goarch, expectedRevision); err != nil {
				t.Fatalf("validate configured Pipeline JIT runtime: %v", err)
			}
			dist := t.TempDir()
			if err := stagePipelineJITRuntime(dist, "linux", goarch); err != nil {
				t.Fatalf("stage configured Pipeline JIT runtime: %v", err)
			}
			for _, name := range []string{pipelineJITLibraryName, pipelineJITLibraryName + ".sha256", pipelineJITStagedManifestName} {
				if _, err := os.Stat(filepath.Join(dist, "lib", name)); err != nil {
					t.Fatalf("stat staged configured artifact %s: %v", name, err)
				}
			}
		})
	}
}
