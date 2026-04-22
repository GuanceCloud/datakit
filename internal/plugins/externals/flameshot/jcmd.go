// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package flameshot

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"golang.org/x/net/context/ctxhttp"
)

const (
	defaultJcmdTimeout          = 10 * time.Second
	jcmdSummaryMeasurement      = "flameshot_jcmd_snapshot"
	maxJcmdSummaryPreviewLength = 2048
)

type jcmdSnapshotRequest struct {
	Service     string
	PID         int32
	ProcessName string
	DetectedAt  time.Time
	MemPercent  float64
	Tags        []string
}

type jcmdSnapshotSummary struct {
	Service         string
	PID             int32
	ProcessName     string
	SnapshotType    string
	OutputPath      string
	OutputSizeBytes int64
	Preview         string
	DetectedAt      time.Time
	MemPercent      float64
}

var findJcmdBinary = func() string {
	if path, err := exec.LookPath("jcmd"); err == nil {
		return path
	}

	candidates := []string{
		"/opt/java/openjdk/bin/jcmd",
		"/usr/local/openjdk/bin/jcmd",
		"/usr/bin/jcmd",
	}
	for _, candidate := range candidates {
		if pathExists(candidate) {
			return candidate
		}
	}
	return ""
}

var runJcmdCommand = func(ctx context.Context, binary string, pid int32, command string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, binary, fmt.Sprintf("%d", pid), command) //nolint:gosec
	return cmd.CombinedOutput()
}

func (m *monitor) getJcmdTimeout() time.Duration {
	if m == nil || m.config == nil || m.config.JCmdTimeout == "" {
		return defaultJcmdTimeout
	}

	d, err := time.ParseDuration(m.config.JCmdTimeout)
	if err != nil || d <= 0 {
		log.Warnf("invalid jcmd timeout %q, fallback to %s", m.config.JCmdTimeout, defaultJcmdTimeout)
		return defaultJcmdTimeout
	}
	return d
}

func buildJcmdOutputPath(root string, pid int32, snapshotType string, now time.Time) string {
	filename := fmt.Sprintf("jcmd_%s_%d_%s.txt", snapshotType, pid, now.Format("20060102_150405"))
	return filepath.Join(root, filename)
}

func shortenText(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max]
}

func summarizeJcmdOutput(snapshotType string, output []byte) string {
	switch snapshotType {
	case "gc_class_histogram":
		return summarizeClassHistogram(output)
	case "thread_print":
		return summarizeThreadPrint(output)
	default:
		return shortenText(strings.TrimSpace(string(output)), maxJcmdSummaryPreviewLength)
	}
}

func summarizeClassHistogram(output []byte) string {
	scanner := bufio.NewScanner(bytes.NewReader(output))
	lines := make([]string, 0, 6)
	started := false
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "num") || strings.HasPrefix(line, "1:") {
			started = true
		}
		if !started {
			continue
		}
		lines = append(lines, line)
		if len(lines) >= 6 {
			break
		}
	}
	if len(lines) == 0 {
		return shortenText(strings.TrimSpace(string(output)), maxJcmdSummaryPreviewLength)
	}
	return strings.Join(lines, "\n")
}

func summarizeThreadPrint(output []byte) string {
	scanner := bufio.NewScanner(bytes.NewReader(output))
	lines := make([]string, 0, 12)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "\"") || strings.Contains(line, "java.lang.Thread.State") {
			lines = append(lines, line)
		}
		if len(lines) >= 12 {
			break
		}
	}
	if len(lines) == 0 {
		return shortenText(strings.TrimSpace(string(output)), maxJcmdSummaryPreviewLength)
	}
	return strings.Join(lines, "\n")
}

func writeSnapshotOutput(path string, data []byte) (int64, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		return 0, err
	}
	if err := os.WriteFile(path, data, 0o644); err != nil { //nolint
		return 0, err
	}
	info, err := os.Stat(path)
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}

func uploadJcmdSummaryLog(summary *jcmdSnapshotSummary, datakitAddr string, tags []string) error {
	if summary == nil {
		return nil
	}

	logURL, err := deriveLoggingURL(datakitAddr)
	if err != nil {
		return err
	}

	kvs := point.KVs{
		point.NewKV("message", fmt.Sprintf("captured %s snapshot for pid=%d at %s", summary.SnapshotType, summary.PID, summary.OutputPath)),
		point.NewKV("status", "info"),
		point.NewKV("snapshot_type", summary.SnapshotType),
		point.NewKV("output_path", summary.OutputPath),
		point.NewKV("output_size_bytes", summary.OutputSizeBytes),
		point.NewKV("preview", summary.Preview),
		point.NewKV("process_name", summary.ProcessName),
		point.NewKV("cgroup_mem_percent", summary.MemPercent),
	}

	for _, tag := range tags {
		if k, v, ok := strings.Cut(tag, ":"); ok && k != "" {
			kvs = kvs.AddTag(k, v)
		}
	}
	if summary.Service != "" {
		kvs = kvs.AddTag("service", summary.Service)
	}

	pt := point.NewPoint(jcmdSummaryMeasurement, kvs, point.DefaultLoggingOptions()...)
	pt.SetTime(summary.DetectedAt)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, logURL, bytes.NewBuffer([]byte(pt.LineProto()+"\n")))
	if err != nil {
		return err
	}

	resp, err := ctxhttp.Do(ctx, http.DefaultClient, req)
	if err != nil {
		return err
	}
	defer resp.Body.Close() //nolint:errcheck

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("upload jcmd summary failed, status=%d", resp.StatusCode)
	}
	return nil
}

func (m *monitor) handleJcmdSnapshot(req *jcmdSnapshotRequest) {
	if m == nil || m.config == nil || req == nil || m.config.ProfilingPath == "" {
		return
	}

	jcmdBinary := findJcmdBinary()
	if jcmdBinary == "" {
		log.Warnf("jcmd binary not found, skip lightweight snapshot for pid=%d", req.PID)
		return
	}

	snapshots := []struct {
		snapshotType string
		command      string
	}{
		{snapshotType: "gc_class_histogram", command: "GC.class_histogram"},
		{snapshotType: "thread_print", command: "Thread.print"},
	}

	for _, snapshot := range snapshots {
		ctx, cancel := context.WithTimeout(context.Background(), m.getJcmdTimeout())
		output, err := runJcmdCommand(ctx, jcmdBinary, req.PID, snapshot.command)
		cancel()
		if err != nil {
			log.Warnf("run jcmd %s failed for pid=%d: %v, output=%s", snapshot.command, req.PID, err, shortenText(string(output), 512))
			continue
		}

		outputPath := buildJcmdOutputPath(m.config.ProfilingPath, req.PID, snapshot.snapshotType, req.DetectedAt)
		size, err := writeSnapshotOutput(outputPath, output)
		if err != nil {
			log.Errorf("write jcmd snapshot output failed: %v", err)
			continue
		}

		summary := &jcmdSnapshotSummary{
			Service:         req.Service,
			PID:             req.PID,
			ProcessName:     req.ProcessName,
			SnapshotType:    snapshot.snapshotType,
			OutputPath:      outputPath,
			OutputSizeBytes: size,
			Preview:         summarizeJcmdOutput(snapshot.snapshotType, output),
			DetectedAt:      req.DetectedAt,
			MemPercent:      req.MemPercent,
		}
		if err := uploadJcmdSummaryLog(summary, m.config.DataKitAddr, req.Tags); err != nil {
			log.Errorf("upload jcmd summary failed: %v", err)
		}
	}
}
