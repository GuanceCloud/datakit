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
	"strings"
	"testing"

	"github.com/GuanceCloud/cliutils/point"
	"github.com/coreos/go-systemd/v22/sdjournal"
)

func TestShouldExcludeField(t *testing.T) {
	tests := []struct {
		name          string
		field         string
		excludeFields []string
		expected      bool
	}{
		{name: "field not in exclude list", field: "MESSAGE", excludeFields: []string{"_PID"}, expected: false},
		{name: "field in exclude list", field: "_PID", excludeFields: []string{"_PID", "SYSLOG_PID"}, expected: true},
		{name: "empty exclude list", field: "MESSAGE", excludeFields: []string{}, expected: false},
		{name: "case sensitive", field: "message", excludeFields: []string{"MESSAGE"}, expected: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ipt := &Input{config: &config{ExcludeFields: tt.excludeFields}}
			if got := ipt.shouldExcludeField(tt.field); got != tt.expected {
				t.Errorf("shouldExcludeField(%q) = %v, expected %v", tt.field, got, tt.expected)
			}
		})
	}
}

func TestResolvePaths(t *testing.T) {
	tmpDir1 := t.TempDir()
	tmpDir2 := t.TempDir()

	tests := []struct {
		name          string
		paths         []string
		expectedCount int
	}{
		{name: "single existing path", paths: []string{tmpDir1}, expectedCount: 1},
		{name: "multiple existing paths", paths: []string{tmpDir1, tmpDir2}, expectedCount: 2},
		{name: "mix of existing and non-existing", paths: []string{tmpDir1, "/nonexistent", tmpDir2}, expectedCount: 2},
		{name: "all non-existing", paths: []string{"/nonexistent1", "/nonexistent2"}, expectedCount: 0},
		{name: "empty paths", paths: []string{}, expectedCount: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ipt := &Input{config: &config{Paths: tt.paths}}
			resolved := ipt.resolvePaths()
			if len(resolved) != tt.expectedCount {
				t.Errorf("resolvePaths() returned %d paths, expected %d", len(resolved), tt.expectedCount)
			}
		})
	}
}

func TestLoadCursor(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name           string
		cursorContent  string
		cursorFile     string
		expectedCursor string
	}{
		{name: "load valid cursor", cursorContent: "s=abc123;i=100", cursorFile: "cursor.txt", expectedCursor: "s=abc123;i=100"},
		{name: "load cursor with whitespace", cursorContent: "  s=abc123  \n", cursorFile: "cursor.txt", expectedCursor: "s=abc123"},
		{name: "no cursor file", cursorContent: "", cursorFile: "nonexistent.txt", expectedCursor: ""},
		{name: "empty cursor file path", cursorContent: "", cursorFile: "", expectedCursor: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cursorFile := ""
			if tt.cursorFile != "" {
				cursorFile = filepath.Join(tmpDir, tt.cursorFile)
				if tt.cursorContent != "" {
					os.WriteFile(cursorFile, []byte(tt.cursorContent), 0o600)
				}
			}

			ipt := &Input{config: &config{CursorFile: cursorFile, SaveCursor: true}, cursor: ""}
			ipt.loadCursor()

			if ipt.cursor != tt.expectedCursor {
				t.Errorf("loadCursor() cursor = %q, expected %q", ipt.cursor, tt.expectedCursor)
			}
		})
	}
}

func TestSaveCursor(t *testing.T) {
	tests := []struct {
		name       string
		cursor     string
		cursorFile string
		expectFile bool
	}{
		{name: "save valid cursor", cursor: "s=abc123", cursorFile: "cursor.txt", expectFile: true},
		{name: "empty cursor does not save", cursor: "", cursorFile: "cursor2.txt", expectFile: false},
		{name: "empty cursor file path", cursor: "s=abc123", cursorFile: "", expectFile: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			cursorFile := ""
			if tt.cursorFile != "" {
				cursorFile = filepath.Join(tmpDir, tt.cursorFile)
			}

			ipt := &Input{config: &config{CursorFile: cursorFile, SaveCursor: true}, cursor: tt.cursor}
			ipt.saveCursor()

			if cursorFile != "" {
				_, err := os.Stat(cursorFile)
				fileExists := err == nil
				if fileExists != tt.expectFile {
					t.Errorf("cursor file exists = %v, expected %v", fileExists, tt.expectFile)
				}
			}
		})
	}
}

func TestTerminate(t *testing.T) {
	t.Run("terminate saves cursor when enabled", func(t *testing.T) {
		tmpDir := t.TempDir()
		cursorFile := filepath.Join(tmpDir, "cursor.txt")

		ipt := &Input{
			config:  &config{CursorFile: cursorFile, SaveCursor: true},
			cursor:  "test-cursor-123",
			done:    make(chan bool),
			journal: nil,
		}

		ipt.Terminate()

		content, err := os.ReadFile(cursorFile)
		if err != nil {
			t.Errorf("failed to read cursor file: %v", err)
		} else if string(content) != "test-cursor-123" {
			t.Errorf("cursor content = %q, expected %q", string(content), "test-cursor-123")
		}

		select {
		case <-ipt.done:
		default:
			t.Error("done channel should be closed")
		}
	})

	t.Run("terminate without cursor saving", func(t *testing.T) {
		tmpDir := t.TempDir()
		cursorFile := filepath.Join(tmpDir, "cursor.txt")

		ipt := &Input{
			config:  &config{CursorFile: cursorFile, SaveCursor: false},
			cursor:  "test-cursor",
			done:    make(chan bool),
			journal: nil,
		}

		ipt.Terminate()

		if _, err := os.Stat(cursorFile); err == nil {
			t.Error("cursor file should not exist when SaveCursor is false")
		}
	})
}

func TestEntryToPoints_Basic(t *testing.T) {
	entry := &sdjournal.JournalEntry{
		Fields: map[string]string{
			"MESSAGE":   "test message",
			"_HOSTNAME": "test-host",
			"_PID":      "1234",
			"PRIORITY":  "6",
		},
		RealtimeTimestamp: 1234567890,
	}

	ipt := &Input{
		config: &config{},
		tags:   map[string]string{"source": "journald"},
	}

	pt := ipt.entryToPoints(entry)

	if pt == nil {
		t.Fatal("entryToPoints returned nil")
	}
	if pt.Name() != "journald" {
		t.Errorf("point name = %q, expected %q", pt.Name(), "journald")
	}

	lp := pt.LineProto()
	if !strings.Contains(lp, `message="test message"`) {
		t.Errorf("message field not found: %s", lp)
	}
	if !strings.Contains(lp, `status="info"`) {
		t.Errorf("status=info not found: %s", lp)
	}
	if !strings.Contains(lp, "priority=6i") {
		t.Errorf("priority field not found: %s", lp)
	}
	if !strings.Contains(lp, "host=test-host") {
		t.Errorf("host tag not found: %s", lp)
	}
}

func TestEntryToPoints_ExcludedFields(t *testing.T) {
	entry := &sdjournal.JournalEntry{
		Fields: map[string]string{
			"MESSAGE":     "test",
			"_BOOT_ID":    "boot-123",
			"_MACHINE_ID": "machine-456",
		},
		RealtimeTimestamp: 1234567890,
	}

	ipt := &Input{
		config: &config{ExcludeFields: []string{"_BOOT_ID", "_MACHINE_ID"}},
	}

	pt := ipt.entryToPoints(entry)
	lp := pt.LineProto()

	if strings.Contains(lp, "_BOOT_ID") {
		t.Errorf("excluded field _BOOT_ID found: %s", lp)
	}
	if strings.Contains(lp, "_MACHINE_ID") {
		t.Errorf("excluded field _MACHINE_ID found: %s", lp)
	}
}

func TestEntryToPoints_ServiceTag(t *testing.T) {
	tests := []struct {
		name            string
		fields          map[string]string
		expectedService string
	}{
		{name: "SYSLOG_IDENTIFIER", fields: map[string]string{"MESSAGE": "test", "SYSLOG_IDENTIFIER": "nginx"}, expectedService: "nginx"},
		{name: "_SYSTEMD_UNIT", fields: map[string]string{"MESSAGE": "test", "_SYSTEMD_UNIT": "docker.service"}, expectedService: "docker.service"},
		{name: "_COMM", fields: map[string]string{"MESSAGE": "test", "_COMM": "bash"}, expectedService: "bash"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry := &sdjournal.JournalEntry{Fields: tt.fields, RealtimeTimestamp: 1234567890}
			ipt := &Input{config: &config{}}
			pt := ipt.entryToPoints(entry)
			lp := pt.LineProto()
			if !strings.Contains(lp, "service="+tt.expectedService) {
				t.Errorf("service=%s not found: %s", tt.expectedService, lp)
			}
		})
	}
}

func TestEntryToPoints_TimestampOverride(t *testing.T) {
	entry := &sdjournal.JournalEntry{
		Fields: map[string]string{
			"MESSAGE":                    "test",
			"_SOURCE_REALTIME_TIMESTAMP": "9876543210",
		},
		RealtimeTimestamp: 1234567890,
	}

	ipt := &Input{config: &config{}}
	pt := ipt.entryToPoints(entry)
	lp := pt.LineProto()

	if !strings.Contains(lp, "987654321") {
		t.Errorf("_SOURCE_REALTIME_TIMESTAMP not used: %s", lp)
	}
}

func TestEntryToPoints_PIDHandling(t *testing.T) {
	t.Run("uses _PID when both present", func(t *testing.T) {
		entry := &sdjournal.JournalEntry{
			Fields:            map[string]string{"MESSAGE": "test", "_PID": "1234", "SYSLOG_PID": "5678"},
			RealtimeTimestamp: 1234567890,
		}
		ipt := &Input{config: &config{}}
		pt := ipt.entryToPoints(entry)
		lp := pt.LineProto()
		if !strings.Contains(lp, "pid=1234i") || strings.Contains(lp, "pid=5678i") {
			t.Errorf("should use _PID not SYSLOG_PID: %s", lp)
		}
	})

	t.Run("uses SYSLOG_PID when _PID not present", func(t *testing.T) {
		entry := &sdjournal.JournalEntry{
			Fields:            map[string]string{"MESSAGE": "test", "SYSLOG_PID": "5678"},
			RealtimeTimestamp: 1234567890,
		}
		ipt := &Input{config: &config{}}
		pt := ipt.entryToPoints(entry)
		lp := pt.LineProto()
		if !strings.Contains(lp, "pid=5678i") {
			t.Errorf("SYSLOG_PID not used: %s", lp)
		}
	})
}

func TestEntryToPoints_PriorityMappings(t *testing.T) {
	tests := []struct {
		priority       string
		expectedStatus string
	}{
		{"0", "error"},
		{"1", "warn"},
		{"2", "critical"},
		{"3", "error"},
		{"4", "warn"},
		{"5", "notice"},
		{"6", "info"},
		{"7", "debug"},
		{"8", "unknown"},
	}

	for _, tt := range tests {
		t.Run("priority_"+tt.priority, func(t *testing.T) {
			entry := &sdjournal.JournalEntry{
				Fields:            map[string]string{"MESSAGE": "test", "PRIORITY": tt.priority},
				RealtimeTimestamp: 1234567890,
			}
			ipt := &Input{config: &config{}}
			pt := ipt.entryToPoints(entry)
			lp := pt.LineProto()
			if tt.expectedStatus != "unknown" && !strings.Contains(lp, `status="`+tt.expectedStatus+`"`) {
				t.Errorf("status=%s not found for priority %s: %s", tt.expectedStatus, tt.priority, lp)
			}
		})
	}
}

func TestEntryToPoints_NumericFields(t *testing.T) {
	entry := &sdjournal.JournalEntry{
		Fields: map[string]string{
			"MESSAGE": "test",
			"_UID":    "1000",
			"_GID":    "1000",
			"FLOAT":   "3.14",
		},
		RealtimeTimestamp: 1234567890,
	}

	ipt := &Input{config: &config{}}
	pt := ipt.entryToPoints(entry)
	lp := pt.LineProto()

	if !strings.Contains(lp, "_UID=1000i") {
		t.Errorf("_UID field not found: %s", lp)
	}
	if !strings.Contains(lp, "_GID=1000i") {
		t.Errorf("_GID field not found: %s", lp)
	}
	if !strings.Contains(lp, "FLOAT=3.14") {
		t.Errorf("FLOAT field not found: %s", lp)
	}
}

func TestFeedPoint_Empty(t *testing.T) {
	ipt := &Input{dkURLPath: "http://test:9529/v1/write/logging"}

	if err := ipt.feedPoint(nil); err != nil {
		t.Errorf("feedPoint(nil) returned error: %v", err)
	}
	if err := ipt.feedPoint([]*point.Point{}); err != nil {
		t.Errorf("feedPoint(empty) returned error: %v", err)
	}
}

func TestCheckSystemd(t *testing.T) {
	err := checkSystemd()
	t.Logf("checkSystemd() returned: %v", err)
}
