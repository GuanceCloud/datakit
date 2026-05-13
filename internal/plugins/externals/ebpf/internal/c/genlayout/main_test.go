//go:build linux
// +build linux

package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestGeneratedNetflowLayoutIsCurrent(t *testing.T) {
	assertGeneratedLayoutIsCurrent(t, generateNetflow,
		"internal/plugins/externals/ebpf/internal/netflow/conn_stats_layout_gen.go",
		"conn_stats_layout_gen.go")
}

func TestGeneratedBashHistoryLayoutIsCurrent(t *testing.T) {
	assertGeneratedLayoutIsCurrent(t, generateBashHistory,
		"internal/plugins/externals/ebpf/internal/bashhistory/bash_history_layout_gen.go",
		"bash_history_layout_gen.go")
}

func TestGeneratedOffsetLayoutIsCurrent(t *testing.T) {
	assertGeneratedLayoutIsCurrent(t, generateOffset,
		"internal/plugins/externals/ebpf/internal/offset/offset_layout_gen.go",
		"offset_layout_gen.go")
}

func TestGeneratedProcwatchLayoutIsCurrent(t *testing.T) {
	assertGeneratedLayoutIsCurrent(t, generateProcwatch,
		"internal/plugins/externals/ebpf/internal/procwatch/runtime_layout_gen.go",
		"runtime_layout_gen.go")
}

func TestGeneratedL7FlowLayoutIsCurrent(t *testing.T) {
	assertGeneratedLayoutIsCurrent(t, generateL7Flow,
		"internal/plugins/externals/ebpf/internal/l7flow/l7_layout_gen.go",
		"l7_layout_gen.go")
}

func assertGeneratedLayoutIsCurrent(t *testing.T, generate func(string) error, repoPath string, fileName string) {
	t.Helper()

	root, err := findRepoRoot()
	if err != nil {
		t.Fatal(err)
	}

	tmpDir := t.TempDir()
	out := filepath.Join(tmpDir, fileName)
	if err := generate(out); err != nil {
		t.Fatal(err)
	}

	got, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	wantPath := filepath.Join(root, repoPath)
	want, err := os.ReadFile(wantPath)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(got, want) {
		t.Fatalf("generated layout %s is stale; run go generate ./internal/plugins/externals/ebpf/internal/...", repoPath)
	}
}
