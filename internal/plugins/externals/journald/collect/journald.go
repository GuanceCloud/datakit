// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build linux
// +build linux

package collect

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/GuanceCloud/cliutils/logger"
	"github.com/GuanceCloud/cliutils/point"
	"github.com/coreos/go-systemd/v22/sdjournal"
)

var (
	logRate = 1.0
	l       = logger.DefaultSLogger("cjournal")
	ptsOpts = point.DefaultLoggingOptions()
)

// checkSystemd verifies that libsystemd is available before attempting to use it.
// This prevents silent failures when deploying to systems without systemd.
func checkSystemd() error {
	// Attempt to create a journal reader - this will fail if libsystemd is not available
	_, err := sdjournal.NewJournal()
	if err != nil {
		// Check for common dlopen errors
		if errors.Is(err, os.ErrNotExist) ||
			err.Error() == "no such file or directory" ||
			err.Error() == "libsystemd.so.0: cannot open shared object file: No such file or directory" ||
			err.Error() == "libsystemd.so.0: cannot open shared object file: Not a directory" {
			return fmt.Errorf("systemd library not found (libsystemd.so.0): %w", err)
		}
		return fmt.Errorf("failed to access systemd journal: %w", err)
	}
	return nil
}

type Input struct {
	opt     *Option
	config  *config
	journal *sdjournal.Journal
	cursor  string
	done    chan bool
	tags    map[string]string

	dkURLPath      string
	totalCollected int64
}

type config struct {
	Paths              []string `long:"paths" description:"Journal paths"`
	Units              []string `long:"units" description:"Filter by systemd units"`
	Priorities         []string `long:"priorities" description:"Filter by priority levels"`
	ExcludeFields      []string `long:"exclude-fields" description:"Fields to exclude"`
	TailOnly           bool     `long:"tail-only" description:"Only read new entries"`
	MaxEntriesPerBatch int      `long:"max-entries" description:"Maximum entries per batch"`
	SaveCursor         bool     `long:"save-cursor" description:"Save cursor position"`
	CursorFile         string   `long:"cursor-file" description:"Cursor file path"`
}

// NewInput create new journald collector.
func NewInput(opt *Option) IInput {
	l = logger.SLogger("cjournal", logger.WithRateLimiter(logRate, ""))

	// Parse configuration from command line arguments
	config := &config{
		TailOnly:           opt.TailOnly,
		MaxEntriesPerBatch: 1000,
		SaveCursor:         opt.SaveCursor,
		CursorFile:         opt.CursorFile,
	}

	// Parse paths
	if opt.Paths != "" {
		config.Paths = strings.Split(opt.Paths, ",")
	} else {
		config.Paths = []string{
			"/var/log/journal",
			"/run/log/journal",
		}
	}

	// Parse units
	if opt.Units != "" {
		config.Units = strings.Split(opt.Units, ",")
	}

	// Parse priorities
	if opt.Priorities != "" {
		config.Priorities = strings.Split(opt.Priorities, ",")
	}

	// Parse exclude fields
	if opt.ExcludeFields != "" {
		config.ExcludeFields = strings.Split(opt.ExcludeFields, ",")
	}

	// Set other options from command line
	if opt.MaxEntries > 0 {
		config.MaxEntriesPerBatch = opt.MaxEntries
	}

	l.Infof("journald input created with %d paths", len(config.Paths))

	// Send to DataKit using configured endpoint

	var dkURLPath string
	if opt.DatakitHTTPEndpoint != "" {
		dkURLPath = opt.DatakitHTTPEndpoint + "/v1/write/logging" + "?input=journald"
	}

	return &Input{
		opt:       opt,
		config:    config,
		done:      make(chan bool),
		tags:      opt.parseTags(), // Add command line options as fields
		dkURLPath: dkURLPath,
	}
}

func (ipt *Input) Run() {
	l.Info("journald input starting")

	// Pre-flight check: verify libsystemd is available
	if err := checkSystemd(); err != nil {
		l.Errorf("systemd pre-flight check failed: %v", err)
		l.Error(" journald collector requires systemd (libsystemd.so.0) to be installed")
		l.Error(" On custom Linux distributions, install systemd or use alternative log collection")
		return
	}

	// Load cursor if exists (only if not in tail-only mode)
	if ipt.config.SaveCursor && !ipt.config.TailOnly {
		ipt.loadCursor()
	}

	// Initialize journal reader
	if err := ipt.initJournal(); err != nil {
		l.Errorf("failed to initialize journal: %v", err)
		return
	}

	for {
		pts := ipt.doCollect()
		if len(pts) > 0 {
			if ipt.totalCollected == 0 {
				l.Infof("first collected log is %s", pts[0].Pretty())
			}

			ipt.totalCollected += int64(len(pts))
			l.RLDebugf(logRate, "collected %d logs, total %d", len(pts), ipt.totalCollected)

			// Send points to DataKit
			if err := ipt.feedPoint(pts); err != nil {
				l.Errorf("failed to write points: %v, ignored", err)
			}
		}

		select {
		case <-ipt.done:
			l.Info("journald input stopping")
			return
		default:
			// Wait for new entries before collecting again
			ipt.journal.Wait(1 * time.Second)
		}
	}
}

func (ipt *Input) initJournal() error {
	resolvedPaths := ipt.resolvePaths()
	if len(resolvedPaths) == 0 {
		return fmt.Errorf("no accessible journal paths found")
	}

	l.Infof("resolved journal paths: %v", resolvedPaths)

	// Open journal
	var err error
	if len(resolvedPaths) == 1 {
		ipt.journal, err = sdjournal.NewJournalFromDir(resolvedPaths[0])
		l.Infof("opening journal from single directory: %s", resolvedPaths[0])
	} else {
		// For multiple directories, use the default system journal
		ipt.journal, err = sdjournal.NewJournal()
		l.Info("opening default system journal")
	}

	if err != nil {
		return fmt.Errorf("failed to open journal: %w", err)
	}

	l.Info("journal opened successfully")

	// Apply filters
	if err := ipt.applyFilters(); err != nil {
		return fmt.Errorf("failed to apply filters: %w", err)
	}
	l.Info("filters applied")

	// Seek to cursor or tail
	if ipt.cursor != "" {
		l.Infof("attempting to seek to cursor: %s", ipt.cursor)
		if err := ipt.journal.SeekCursor(ipt.cursor); err != nil {
			l.Warnf("failed to seek to cursor: %v, starting from tail", err)
			if err := ipt.journal.SeekTail(); err != nil {
				return fmt.Errorf("failed to seek to tail: %w", err)
			}
			l.Info("seek to tail after cursor failure")
		} else {
			l.Info("successfully seeked to cursor")
		}
	}

	if ipt.config.TailOnly {
		l.Info("seeking to tail (tail-only mode)")
		if err := ipt.journal.SeekTail(); err != nil {
			return fmt.Errorf("failed to seek to tail: %w", err)
		}
		// Move back one entry so Next() will get new entries
		if _, err := ipt.journal.Previous(); err != nil {
			l.Warnf("failed to call Previous(): %v", err)
		}
		l.Info("successfully seeked to tail, positioned for new entries")
	} else {
		l.Info("starting from beginning (not tail-only mode)")
	}

	return nil
}

func (ipt *Input) resolvePaths() []string {
	var resolved []string
	for _, path := range ipt.config.Paths {
		if _, err := os.Stat(path); err == nil {
			resolved = append(resolved, path)
		}
	}
	return resolved
}

func (ipt *Input) applyFilters() error {
	// Filter by units
	if len(ipt.config.Units) > 0 {
		// Add OR disjunction for units
		for i, unit := range ipt.config.Units {
			match := sdjournal.Match{
				Field: "_SYSTEMD_UNIT",
				Value: unit,
			}
			if err := ipt.journal.AddMatch(match.String()); err != nil {
				return fmt.Errorf("failed to add unit filter: %w", err)
			}
			// Add disjunction between unit filters (OR)
			if i < len(ipt.config.Units)-1 {
				if err := ipt.journal.AddDisjunction(); err != nil {
					return fmt.Errorf("failed to add disjunction: %w", err)
				}
			}
		}
	}

	// Filter by priorities
	if len(ipt.config.Priorities) > 0 {
		// Add OR disjunction for priorities
		for i, priority := range ipt.config.Priorities {
			match := sdjournal.Match{
				Field: "PRIORITY",
				Value: priority,
			}
			if err := ipt.journal.AddMatch(match.String()); err != nil {
				return fmt.Errorf("failed to add priority filter: %w", err)
			}
			// Add disjunction between priority filters (OR)
			if i < len(ipt.config.Priorities)-1 {
				if err := ipt.journal.AddDisjunction(); err != nil {
					return fmt.Errorf("failed to add disjunction: %w", err)
				}
			}
		}
	}

	return nil
}

func (ipt *Input) doCollect() []*point.Point {
	var (
		count = 0
		pts   []*point.Point
	)

	// Read all available entries up to MaxEntriesPerBatch
	for count < ipt.config.MaxEntriesPerBatch {
		select {
		case <-ipt.done:
			l.Info("journald collect interrupted")
			goto terminated
		default: // pass
		}

		n, err := ipt.journal.Next()
		if err != nil {
			l.Errorf("failed to read journal entry: %v", err)
			break
		}

		if n == 0 {
			// No more entries available
			break
		}

		// Read entry
		entry, err := ipt.journal.GetEntry()
		if err != nil {
			l.Errorf("failed to get journal entry: %v", err)
			continue
		}

		// Convert entry to point
		pts = append(pts, ipt.entryToPoints(entry))

		count++
	}

terminated:
	// Save cursor to file if we collected entries
	if count > 0 {
		// Save cursor
		if ipt.config.SaveCursor {
			cursor, err := ipt.journal.GetCursor()
			if err == nil {
				ipt.cursor = cursor
			}
		}

		if ipt.config.SaveCursor && ipt.cursor != "" {
			ipt.saveCursor()
		}
	}

	return pts
}

func (ipt *Input) entryToPoints(entry *sdjournal.JournalEntry) *point.Point {
	var (
		// default use journald received time.
		// entry.RealtimeTimestamp is microseconds since epoch
		ts = int64(entry.RealtimeTimestamp) * int64(time.Microsecond)

		// Add service tag based on available identifiers
		// Priority: SYSLOG_IDENTIFIER -> _SYSTEMD_UNIT -> _COMM
		service = ""

		kvs point.KVs
	)

	for k, v := range ipt.tags {
		kvs = kvs.AddTag(k, v)
	}

	kvs = kvs.Add("journald_timestamp", ts)

	// Resolve PID before iterating the map so _PID consistently wins over
	// SYSLOG_PID regardless of Go's randomized map iteration order.
	if pidValue, ok := entry.Fields["_PID"]; ok {
		if pid, err := strconv.ParseInt(pidValue, 10, 64); err == nil {
			kvs = kvs.Add("pid", pid)
		}
	} else if pidValue, ok := entry.Fields["SYSLOG_PID"]; ok {
		if pid, err := strconv.ParseInt(pidValue, 10, 64); err == nil {
			kvs = kvs.Add("pid", pid)
		}
	}

	// Add journal-specific fields
	for key, value := range entry.Fields {
		if ipt.shouldExcludeField(key) {
			continue
		}

		switch key {
		case "_HOSTNAME":
			kvs = kvs.AddTag("host", value)

		// But if source timestamp exists, USE IT INSTEAD
		case "_SOURCE_REALTIME_TIMESTAMP":
			if x, err := strconv.ParseInt(value, 0, 64); err == nil {
				ts = x * int64(time.Microsecond)
			}

		case "SYSLOG_PID":
		case "_PID":
			// PID is already handled before the loop to keep the precedence stable.
		case "PRIORITY":
			if prio, err := strconv.Atoi(value); err == nil {
				// Map journald priority to Guance status field
				status := priorityToStatus(prio)
				kvs = kvs.Add("status", status)
				// Also keep numeric priority as field
				kvs = kvs.Add("priority", prio)
			}

		case "MESSAGE":
			kvs = kvs.Add("message", value)

		case "SYSLOG_IDENTIFIER",
			"_SYSTEMD_UNIT",
			"_COMM": // for service identify
			if service == "" {
				service = value
			}

		default:
			// Try to parse numeric values
			if intVal, err := strconv.ParseInt(value, 10, 64); err == nil {
				kvs = kvs.Add(key, intVal)
			} else if floatVal, err := strconv.ParseFloat(value, 64); err == nil {
				kvs = kvs.Add(key, floatVal)
			} else {
				kvs = kvs.Add(key, value)
			}
		}
	}

	// Add service tag if identified
	if service != "" {
		kvs = kvs.AddTag("service", service)
	}

	// Create point (timestamp already set via point.WithTime)
	pt := point.NewPoint("journald", kvs, ptsOpts...)
	pt.SetTime(time.Unix(0, ts))
	return pt
}

func (ipt *Input) shouldExcludeField(field string) bool {
	for _, exclude := range ipt.config.ExcludeFields {
		if field == exclude {
			return true
		}
	}
	return false
}

func priorityToStatus(priority int) string {
	// Map journald priority (0-7) to Guance status values
	// Based on typical Guance logging status mapping
	switch priority {
	case 0, 3: // emerg,err
		return "error"
	case 2: // crit
		return "critical"
	case 1, 4: // alert,warning
		return "warn"
	case 5: // notice
		return "notice"
	case 6: // info
		return "info"
	case 7: // debug
		return "debug"
	default:
		return "unknown"
	}
}

func (ipt *Input) loadCursor() {
	if ipt.config.CursorFile == "" {
		return
	}

	data, err := os.ReadFile(ipt.config.CursorFile)
	if err != nil {
		l.Debugf("no cursor file found: %v", err)
		return
	}

	ipt.cursor = strings.TrimSpace(string(data))
	l.Infof("loaded cursor: %s", ipt.cursor)
}

func (ipt *Input) saveCursor() {
	if ipt.config.CursorFile == "" || ipt.cursor == "" {
		return
	}

	// Ensure directory exists
	dir := filepath.Dir(ipt.config.CursorFile)
	if err := os.MkdirAll(dir, 0o750); err != nil {
		l.Errorf("failed to create cursor directory: %v", err)
		return
	}

	if err := os.WriteFile(ipt.config.CursorFile, []byte(ipt.cursor), 0o600); err != nil {
		l.Errorf("failed to save cursor: %v", err)
	}
}

func (ipt *Input) Terminate() {
	l.Info("journald input terminating")

	// Save final cursor
	if ipt.config.SaveCursor && ipt.cursor != "" {
		ipt.saveCursor()
	}

	// Close journal
	if ipt.journal != nil {
		if err := ipt.journal.Close(); err != nil {
			l.Warnf("journal.Close: %s, ignored", err)
		}
	}

	close(ipt.done)
}
