// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package journald collects systemd journal logs.
package journald

import (
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/GuanceCloud/cliutils"
	"github.com/GuanceCloud/cliutils/logger"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/external"
)

const (
	inputName = "journald"
)

var (
	l                  = logger.DefaultSLogger(inputName)
	_ inputs.Singleton = (*Input)(nil)

	defaultBinaryName   = "journald"
	journaldBinaryPaths = []string{
		"/usr/local/datakit/externals/journald",
		"./externals/journald",
		"journald",
	}
)

type Input struct {
	external.Input

	HTTPEndpoint string `toml:"http_endpoint"`
	LogPath      string `toml:"log_path"`
	LogLevel     string `toml:"log_level"`

	// Journal paths
	Paths []string `toml:"paths"`

	// Filter by systemd units
	Units []string `toml:"units"`

	// Filter by priority levels
	Priorities []string `toml:"priorities"`

	// Fields to exclude
	ExcludeFields []string `toml:"exclude_fields"`

	// Collection behavior
	TailOnly           bool `toml:"tail_only"`
	MaxEntriesPerBatch int  `toml:"max_entries_per_batch"`

	// Cursor management
	SaveCursor bool   `toml:"save_cursor"`
	CursorFile string `toml:"cursor_file"`

	semStop *cliutils.Sem
}

func (ipt *Input) Singleton() {}

func (ipt *Input) Run() {
	l = logger.SLogger(inputName)

	l.Info("journald input starting")

	// Runtime guard: journald only works on Linux
	// This allows code to compile on macOS/Windows for documentation export,
	// but actual data collection only runs on Linux
	if runtime.GOOS != "linux" {
		l.Warnf("journald input is only supported on Linux (current OS: %s), skipping data collection", runtime.GOOS)
		return
	}

	// Find journald binary
	execFile := ipt.findJournaldBinary()
	if execFile == "" {
		l.Errorf("journald binary not found, tried paths: %v", journaldBinaryPaths)
		return
	}

	// Update command to use found binary
	ipt.Input.Cmd = execFile

	// Build arguments for external binary
	ipt.buildArgs()

	// Run external binary via external.Input
	ipt.Input.Run()
}

func (ipt *Input) findJournaldBinary() string {
	// Try configured binary path first
	if ipt.Input.Cmd != "" {
		if _, err := os.Stat(ipt.Input.Cmd); err == nil {
			return ipt.Input.Cmd
		}
	}

	// Try default paths
	for _, path := range journaldBinaryPaths {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	return ""
}

func (ipt *Input) buildArgs() {
	args := []string{
		"--datakit-http-endpoint", ipt.HTTPEndpoint,
		"--log-path", ipt.LogPath,
		"--log-level", ipt.LogLevel,
	}

	// Add journal paths
	if len(ipt.Paths) > 0 {
		args = append(args, "--paths")
		args = append(args, strings.Join(ipt.Paths, ","))
	}

	// Add unit filters
	if len(ipt.Units) > 0 {
		args = append(args, "--units")
		args = append(args, strings.Join(ipt.Units, ","))
	}

	// Add priority filters
	if len(ipt.Priorities) > 0 {
		args = append(args, "--priorities")
		args = append(args, strings.Join(ipt.Priorities, ","))
	}

	// Add exclude fields
	if len(ipt.ExcludeFields) > 0 {
		args = append(args, "--exclude-fields")
		args = append(args, strings.Join(ipt.ExcludeFields, ","))
	}

	// Add tail only flag
	if ipt.TailOnly {
		args = append(args, "--tail-only")
	}

	// Add max entries per batch
	args = append(args, "--max-entries", strconv.Itoa(ipt.MaxEntriesPerBatch))

	// Add cursor options
	if ipt.SaveCursor && ipt.CursorFile != "" {
		args = append(args, "--save-cursor", "--cursor-file", ipt.CursorFile)
	}

	ipt.Input.Args = args
}

func (ipt *Input) Terminate() {
	if ipt.semStop != nil {
		ipt.semStop.Close()
	}
	ipt.Input.Terminate()
}

func (*Input) Catalog() string { return "logging" }

func (*Input) SampleConfig() string { return sampleConfig }

func (*Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{
		&journalMeasurement{},
	}
}

func (*Input) AvailableArchs() []string {
	return []string{datakit.OSLabelLinux, datakit.LabelK8s, datakit.LabelDocker}
}

func defaultInput() *Input {
	extInput := external.NewInput()

	extInput.Name = inputName
	extInput.Election = false // journald used to collect local log, do not need election
	extInput.Daemon = true
	extInput.Interval = "10s"
	extInput.Cmd = defaultBinaryName

	return &Input{
		Input:        *extInput,
		HTTPEndpoint: "http://localhost:9529",
		Paths: []string{
			"/var/log/journal",
			"/run/log/journal",
		},
		TailOnly:           true,
		MaxEntriesPerBatch: 1000,
		SaveCursor:         true,
		CursorFile:         filepath.Join(datakit.DataDir, "cache", "journald.cursor"),
		semStop:            cliutils.NewSem(),
	}
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return defaultInput()
	})
}
