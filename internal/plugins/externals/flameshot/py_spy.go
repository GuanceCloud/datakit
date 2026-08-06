// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package flameshot

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
	defaultPySpyPath       = "py-spy"
	defaultPySpyRate       = 100
	pySpyAttachmentName    = "prof"
	pySpyOutputFormat      = "raw"
	pySpyUploadEventFormat = "collapse"
)

var runPySpyCommand = func(ctx context.Context, pySpyPath string, args []string) (string, string, error) {
	cmd := exec.CommandContext(ctx, pySpyPath, args...) //nolint:gosec

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	return stdout.String(), stderr.String(), err
}

func runPythonPySpy(ctx context.Context, stats *triggerStats) error {
	if stats == nil {
		return fmt.Errorf("trigger stats is nil")
	}
	if stats.PID <= 0 {
		return fmt.Errorf("PID (%d) must be > 0", stats.PID)
	}

	outputPath := resolvePySpyOutputPath(stats)
	if outputPath == "" {
		return fmt.Errorf("py-spy output path is empty")
	}
	if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
		return fmt.Errorf("create py-spy output dir: %w", err)
	}
	if err := os.Remove(outputPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove stale py-spy output file: %w", err)
	}
	cleanupOutput := true
	defer func() {
		if !cleanupOutput {
			return
		}
		if err := os.Remove(outputPath); err != nil && !os.IsNotExist(err) {
			log.Errorf("remove failed py-spy output file %s: %v", outputPath, err)
		}
	}()

	stats.startTime = time.Now().Format(time.RFC3339Nano)
	args := buildPySpyArgs(stats, outputPath)
	pySpyPath := strings.TrimSpace(stats.PySpyPath)
	if pySpyPath == "" {
		pySpyPath = defaultPySpyPath
	}

	log.Infof("exec command: %s %v", pySpyPath, args)
	stdout, stderr, err := runPySpyCommand(ctx, pySpyPath, args)
	stats.endTime = time.Now().Format(time.RFC3339Nano)
	if err != nil {
		return fmt.Errorf("py-spy exec err: %w, stdout is %s, stderr is %s", err, stdout, stderr)
	}
	log.Infof("ok, py-spy output file is %s", outputPath)
	log.Infof("stdout: %s", stdout)

	data, err := os.ReadFile(outputPath) //nolint:gosec
	if err != nil {
		return fmt.Errorf("read py-spy output file: %w", err)
	}
	if len(data) == 0 {
		return fmt.Errorf("py-spy output file is empty: %s", outputPath)
	}

	stats.OutputFile = outputPath
	stats.OutputFormat = pySpyUploadEventFormat
	stats.Attachments = []*profileAttachment{
		{
			FieldName: pySpyAttachmentName,
			FileName:  pySpyAttachmentName,
			Data:      data,
		},
	}
	cleanupOutput = false

	return nil
}

func buildPySpyArgs(stats *triggerStats, outputPath string) []string {
	duration := DefaultDuration
	pid := int32(0)
	if stats != nil && stats.Duration > 0 {
		duration = stats.Duration
	}
	if stats != nil {
		pid = stats.PID
	}

	args := []string{
		"record",
		"--format", pySpyOutputFormat,
		"--output", outputPath,
		"--duration", strconv.Itoa(duration),
		"--pid", strconv.Itoa(int(pid)),
	}

	if stats != nil && stats.PySpyRate > 0 {
		args = append(args, "--rate", strconv.Itoa(stats.PySpyRate))
	}
	if stats != nil && stats.PySpySubproc {
		args = append(args, "--subprocesses")
	}
	if stats != nil && stats.PySpyIdle {
		args = append(args, "--idle")
	}

	return args
}

func resolvePySpyOutputPath(stats *triggerStats) string {
	outputPath := ""
	if stats != nil {
		outputPath = strings.TrimSpace(stats.PySpyOutput)
	}
	if outputPath == "" {
		pid := int32(0)
		if stats != nil {
			pid = stats.PID
		}
		outputPath = fmt.Sprintf("pyspy_%d_%s.raw", pid, time.Now().Format("20060102_150405_000000000"))
	}
	if filepath.IsAbs(outputPath) {
		return filepath.Clean(outputPath)
	}
	return filepath.Join(DefaultOutput, outputPath)
}
