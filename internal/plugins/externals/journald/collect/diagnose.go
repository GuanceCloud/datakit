// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build linux
// +build linux

package collect

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/coreos/go-systemd/v22/sdjournal"
)

type probeReason string

const (
	reasonOK                probeReason = "ok"
	reasonUnsupportedFormat probeReason = "unsupported-format"
	reasonPermissionDenied  probeReason = "permission-denied"
	reasonUnexpectedOpen    probeReason = "unexpected-open-error"
	reasonNoJournalFiles    probeReason = "no-journal-files"
)

type probeResult struct {
	reason        probeReason
	target        string
	message       string
	readerVersion string
}

var runJournalctlVersion = defaultRunJournalctlVersion

func selectProbeFailure(results []probeResult) probeResult {
	if len(results) == 0 {
		return probeResult{reason: reasonNoJournalFiles}
	}

	best := results[0]
	bestSeverity := probeReasonSeverity(best.reason)
	for _, result := range results[1:] {
		if severity := probeReasonSeverity(result.reason); severity > bestSeverity {
			best = result
			bestSeverity = severity
		}
	}

	return best
}

func probeReasonSeverity(reason probeReason) int {
	switch reason {
	case reasonOK:
		return 0
	case reasonUnsupportedFormat:
		return 4
	case reasonPermissionDenied:
		return 3
	case reasonUnexpectedOpen:
		return 2
	case reasonNoJournalFiles:
		return 1
	default:
		return 0
	}
}

func detectReaderVersion() string {
	out, err := runJournalctlVersion()
	if err != nil {
		return ""
	}

	fields := strings.Fields(string(out))
	if len(fields) < 2 || fields[0] != "systemd" {
		return ""
	}

	version := strings.TrimSpace(fields[1])
	for _, ch := range version {
		if ch < '0' || ch > '9' {
			return ""
		}
	}

	return version
}

func defaultRunJournalctlVersion() ([]byte, error) {
	return exec.Command("journalctl", "--version").Output() //nolint:gosec
}

func probeJournalTarget(target string) probeResult {
	info, err := os.Stat(target)
	if err != nil {
		return classifyProbeError(target, err)
	}

	var journal *sdjournal.Journal
	if info.IsDir() {
		journal, err = sdjournal.NewJournalFromDir(target)
	} else {
		journal, err = sdjournal.NewJournalFromFiles(target)
	}
	if err != nil {
		return classifyProbeError(target, err)
	}
	defer func() { _ = journal.Close() }()

	if err := journal.SeekHead(); err != nil {
		return classifyProbeError(target, err)
	}

	if _, err := journal.Next(); err != nil {
		return classifyProbeError(target, err)
	}

	return probeResult{reason: reasonOK, target: target}
}

func resolveProbeTargets(paths []string) []string {
	seen := make(map[string]struct{})
	var targets []string

	addTarget := func(target string) {
		if _, ok := seen[target]; ok {
			return
		}
		seen[target] = struct{}{}
		targets = append(targets, target)
	}

	for _, path := range paths {
		info, err := os.Stat(path)
		if err != nil {
			continue
		}

		if !info.IsDir() {
			if strings.HasSuffix(path, ".journal") {
				addTarget(path)
			}
			continue
		}

		if directoryHasJournalFiles(path) {
			addTarget(path)
			continue
		}

		entries, err := os.ReadDir(path)
		if err != nil {
			continue
		}

		var expanded []string
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}

			target := filepath.Join(path, entry.Name())
			if directoryHasJournalFiles(target) {
				expanded = append(expanded, target)
			}
		}

		sort.Strings(expanded)
		for _, target := range expanded {
			addTarget(target)
		}
	}

	return targets
}

func directoryHasJournalFiles(path string) bool {
	entries, err := os.ReadDir(path)
	if err != nil {
		return false
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if strings.HasSuffix(entry.Name(), ".journal") {
			return true
		}
	}

	return false
}

func aggregateProbeResults(results map[string]probeResult) probeResult {
	if len(results) == 0 {
		return probeResult{reason: reasonNoJournalFiles}
	}

	keys := make([]string, 0, len(results))
	for key := range results {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	failures := make([]probeResult, 0, len(keys))
	for _, key := range keys {
		result := results[key]
		if result.reason == reasonOK {
			return result
		}
		failures = append(failures, result)
	}

	return selectProbeFailure(failures)
}

func (ipt *Input) preflightCompatibility() probeResult {
	readerVersion := detectReaderVersion()
	targets := resolveProbeTargets(ipt.config.Paths)
	if len(targets) == 0 {
		targets = append(targets, ipt.config.Paths...)
	}

	if len(targets) == 0 {
		return probeResult{
			reason:        reasonNoJournalFiles,
			message:       "no journal files found in configured paths",
			readerVersion: readerVersion,
		}
	}

	results := make(map[string]probeResult, len(targets))
	for _, target := range targets {
		result := probeJournalTargetFn(target)
		if result.target == "" {
			result.target = target
		}
		result.readerVersion = readerVersion
		if result.reason == reasonOK {
			return result
		}
		results[target] = result
	}

	result := aggregateProbeResults(results)
	result.readerVersion = readerVersion
	return result
}

func classifyProbeError(target string, err error) probeResult {
	message := err.Error()
	lowerMessage := strings.ToLower(message)

	switch {
	case strings.Contains(lowerMessage, "unsupported feature"):
		return probeResult{reason: reasonUnsupportedFormat, target: target, message: message}
	case errors.Is(err, os.ErrPermission), strings.Contains(lowerMessage, "permission denied"):
		return probeResult{reason: reasonPermissionDenied, target: target, message: message}
	case errors.Is(err, os.ErrNotExist), strings.Contains(lowerMessage, "no such file or directory"):
		return probeResult{reason: reasonNoJournalFiles, target: target, message: message}
	default:
		return probeResult{reason: reasonUnexpectedOpen, target: target, message: message}
	}
}

func compatibilityAdvice(reason probeReason) string {
	switch reason {
	case reasonOK:
		return "no compatibility issue detected"
	case reasonUnsupportedFormat:
		return "host journal requires newer libsystemd; copy compatible systemd libraries from the host " +
			"into a dedicated directory, mount them into the DataKit container, and configure the " +
			"collector to prefer that library path"
	case reasonPermissionDenied:
		return "check journal file permissions and mounts before restarting the collector"
	case reasonUnexpectedOpen:
		return "check journal target accessibility and bundled libsystemd compatibility"
	case reasonNoJournalFiles:
		return "verify the configured journald paths and host journal mounts"
	default:
		return "check journal target accessibility and bundled libsystemd compatibility"
	}
}

func (ipt *Input) logCompatibilityFailure(result probeResult) {
	l.Warnf(
		"journald compatibility check failed: configured_paths=%q target=%s reason=%s reader_version=%s message=%q; collector will stay inactive; %s",
		strings.Join(ipt.config.Paths, ","),
		result.target,
		result.reason,
		result.readerVersion,
		result.message,
		compatibilityAdvice(result.reason),
	)
}
