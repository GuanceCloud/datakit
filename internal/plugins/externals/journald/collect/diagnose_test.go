// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build linux
// +build linux

package collect

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/GuanceCloud/cliutils/logger"
	"github.com/GuanceCloud/cliutils/point"
	"github.com/stretchr/testify/require"
)

func TestProbeSeverityOrder(t *testing.T) {
	results := []probeResult{
		{reason: reasonNoJournalFiles, target: "/a"},
		{reason: reasonUnsupportedFormat, target: "/b"},
	}

	got := selectProbeFailure(results)
	if got.reason != reasonUnsupportedFormat {
		t.Fatalf("reason = %s, want %s", got.reason, reasonUnsupportedFormat)
	}
}

func TestDetectReaderVersion_ParsesJournalctlVersion(t *testing.T) {
	runJournalctlVersion = func() ([]byte, error) {
		return []byte("systemd 249 (249.11-0ubuntu3.19)\n"), nil
	}
	t.Cleanup(func() { runJournalctlVersion = defaultRunJournalctlVersion })

	got := detectReaderVersion()
	if got != "249" {
		t.Fatalf("version = %q, want %q", got, "249")
	}
}

func TestResolveProbeTargets_ExpandsJournalRoot(t *testing.T) {
	root := t.TempDir()
	machineDir := filepath.Join(root, "machine-id-1")
	require.NoError(t, os.MkdirAll(machineDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(machineDir, "system.journal"), []byte("x"), 0o644))

	got := resolveProbeTargets([]string{root})
	require.Equal(t, []string{machineDir}, got)
}

func TestResolveProbeTargets_PreservesDirectJournalFile(t *testing.T) {
	root := t.TempDir()
	file := filepath.Join(root, "system.journal")
	require.NoError(t, os.WriteFile(file, []byte("x"), 0o644))

	got := resolveProbeTargets([]string{file})
	require.Equal(t, []string{file}, got)
}

func TestResolveProbeTargets_PreservesDirectMachineIDDirectory(t *testing.T) {
	root := t.TempDir()
	machineDir := filepath.Join(root, "ec2f02bf505ce9dd2cc7dce0561ccd18")
	require.NoError(t, os.MkdirAll(machineDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(machineDir, "system.journal"), []byte("x"), 0o644))

	got := resolveProbeTargets([]string{machineDir})
	require.Equal(t, []string{machineDir}, got)
}

func TestRun_SkipsCollectionWhenProbeFails(t *testing.T) {
	systemdCheckFn = func() error { return nil }
	probeJournalTargetFn = func(target string) probeResult {
		return probeResult{reason: reasonUnsupportedFormat, target: target, message: "unsupported feature"}
	}
	t.Cleanup(restoreProbeSeams)

	ipt := &Input{config: &config{Paths: []string{"/rootfs/var/log/journal"}}, done: make(chan bool)}
	got := ipt.preflightCompatibility()
	require.Equal(t, reasonUnsupportedFormat, got.reason)
}

func TestRun_ExitsBeforeInitJournalWhenProbeFails(t *testing.T) {
	systemdCheckFn = func() error { return nil }
	probeJournalTargetFn = func(target string) probeResult {
		return probeResult{reason: reasonUnsupportedFormat, target: target, message: "unsupported feature"}
	}
	initJournalFn = func(*Input) error {
		t.Fatal("initJournal should not run when compatibility probe fails")
		return nil
	}
	doCollectFn = func(*Input) []*point.Point {
		t.Fatal("doCollect should not run when compatibility probe fails")
		return nil
	}
	t.Cleanup(restoreProbeSeams)

	ipt := &Input{config: &config{Paths: []string{"/rootfs/var/log/journal"}}, done: make(chan bool)}
	ipt.Run()
}

func TestSelectProbeFailure_TieBreaksByProbeOrder(t *testing.T) {
	results := []probeResult{
		{reason: reasonPermissionDenied, target: "/a"},
		{reason: reasonPermissionDenied, target: "/b"},
	}

	got := selectProbeFailure(results)
	require.Equal(t, "/a", got.target)
}

func TestPreflightCompatibility_ClassifiesFailuresAndSuccess(t *testing.T) {
	tests := []struct {
		name    string
		results map[string]probeResult
		want    probeReason
	}{
		{name: "permission denied", results: map[string]probeResult{"/a": {reason: reasonPermissionDenied, target: "/a"}}, want: reasonPermissionDenied},
		{name: "no journal files", results: map[string]probeResult{"/a": {reason: reasonNoJournalFiles, target: "/a"}}, want: reasonNoJournalFiles},
		{name: "unexpected open error", results: map[string]probeResult{"/a": {reason: reasonUnexpectedOpen, target: "/a"}}, want: reasonUnexpectedOpen},
		{name: "one target succeeds", results: map[string]probeResult{"/a": {reason: reasonUnsupportedFormat, target: "/a"}, "/b": {reason: reasonOK, target: "/b"}}, want: reasonOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := aggregateProbeResults(tt.results)
			require.Equal(t, tt.want, got.reason)
		})
	}
}

func TestRun_LogsCompatibilityWarningDetails(t *testing.T) {
	logger.Reset()
	logPath := filepath.Join(t.TempDir(), "journald.log")
	require.NoError(t, logger.InitRoot(&logger.Option{Level: "debug", Path: logPath, Flags: logger.OPT_DEFAULT}))
	l = logger.DefaultSLogger("cjournal")

	systemdCheckFn = func() error { return nil }
	probeJournalTargetFn = func(target string) probeResult {
		return probeResult{
			reason:  reasonUnsupportedFormat,
			target:  "/rootfs/var/log/journal/machine/system.journal",
			message: "unsupported feature",
		}
	}
	runJournalctlVersion = func() ([]byte, error) {
		return []byte("systemd 249 (249.11-0ubuntu3.19)\n"), nil
	}
	t.Cleanup(func() {
		restoreProbeSeams()
		logger.Reset()
		l = logger.DefaultSLogger("cjournal")
	})

	ipt := &Input{config: &config{Paths: []string{"/rootfs/var/log/journal"}}, done: make(chan bool)}
	ipt.Run()

	out, err := os.ReadFile(logPath)
	require.NoError(t, err)
	require.Contains(t, string(out), "reason=unsupported-format")
	require.Contains(t, string(out), "target=/rootfs/var/log/journal/machine/system.journal")
	require.Contains(t, string(out), "reader_version=249")
	require.Contains(t, string(out), "collector will stay inactive")
	require.Contains(t, string(out), "newer libsystemd")
}

func restoreProbeSeams() {
	systemdCheckFn = checkSystemd
	probeJournalTargetFn = probeJournalTarget
	initJournalFn = func(ipt *Input) error { return ipt.initJournal() }
	doCollectFn = func(ipt *Input) []*point.Point { return ipt.doCollect() }
	runJournalctlVersion = defaultRunJournalctlVersion
}
