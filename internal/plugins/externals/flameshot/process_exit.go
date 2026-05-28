// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package flameshot

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"golang.org/x/net/context/ctxhttp"
)

const (
	processExitMeasurement           = "flameshot_process_exit"
	defaultProcessExitArtifactWindow = 2 * time.Minute
	defaultExitSamplesToReport       = 15
)

type localArtifact struct {
	Path      string
	SizeBytes int64
	ModTime   time.Time
}

type processExitEvent struct {
	Service              string
	PID                  int32
	ProcessName          string
	DetectedAt           time.Time
	LastSeenAt           time.Time
	LastProfileTime      time.Time
	LastMemoryPressureAt time.Time
	RecentMemPeakPercent float64
	RecentMemPeakMB      float64
	RecentCPUPeak        float64
	RecentSamples        []resourceSample
	LastProfileArtifact  *profileArtifactSummary
	RecentJFRArtifact    *localArtifact
	RecentHProfArtifact  *localArtifact
	SuspectedReason      string
	ExitError            string
	Tags                 []string
}

func (m *monitor) buildProcessExitEvent(pm *processM, detectedAt time.Time, reason error) *processExitEvent {
	if pm == nil || pm.configProcess == nil {
		return nil
	}

	lastSeenAt, lastProfileTime, lastMemoryPressureAt, artifact := pm.snapshotState()
	recentMemPeakPercent, recentMemPeakMB, recentCPUPeak := pm.recentPeaks(pm.SampleMaxSize)
	samples := pm.snapshotRecentSamples(defaultExitSamplesToReport)
	configTags := make([]string, 0)
	profilingPath := ""
	if m != nil && m.config != nil {
		configTags = m.config.Tags
		profilingPath = m.config.ProfilingPath
	}

	tags := make([]string, 0, len(configTags)+len(pm.configProcess.Tags)+1)
	tags = append(tags, configTags...)
	tags = append(tags, pm.configProcess.Tags...)
	tags = append(tags, fmt.Sprintf("service:%s", pm.configProcess.Service))

	hprofRoot := detectOOMHProfPath(pm)
	if hprofRoot == "" {
		hprofRoot = profilingPath
	}
	hprofArtifact, _ := findMatchingHProf(hprofRoot, detectedAt, defaultProcessExitArtifactWindow)
	jfrArtifact, _ := findLatestArtifact(profilingPath, detectedAt, defaultProcessExitArtifactWindow, ".jfr", "profiler_")

	return &processExitEvent{
		Service:              pm.configProcess.Service,
		PID:                  pm.Pid,
		ProcessName:          pm.Name,
		DetectedAt:           detectedAt,
		LastSeenAt:           lastSeenAt,
		LastProfileTime:      lastProfileTime,
		LastMemoryPressureAt: lastMemoryPressureAt,
		RecentMemPeakPercent: recentMemPeakPercent,
		RecentMemPeakMB:      recentMemPeakMB,
		RecentCPUPeak:        recentCPUPeak,
		RecentSamples:        samples,
		LastProfileArtifact:  artifact,
		RecentJFRArtifact:    jfrArtifact,
		RecentHProfArtifact:  toLocalArtifact(hprofArtifact),
		SuspectedReason:      inferExitReason(detectedAt, lastMemoryPressureAt, hprofArtifact != nil),
		ExitError:            errorString(reason),
		Tags:                 tags,
	}
}

func inferExitReason(detectedAt, lastMemoryPressureAt time.Time, hasHProf bool) string {
	if hasHProf {
		return "process_exit_with_heap_dump"
	}
	if !lastMemoryPressureAt.IsZero() && detectedAt.Sub(lastMemoryPressureAt) <= defaultProcessExitArtifactWindow {
		return "process_exit_after_memory_pressure"
	}
	return "process_exit"
}

func errorString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

func toLocalArtifact(file *hprofFile) *localArtifact {
	if file == nil {
		return nil
	}
	return &localArtifact{
		Path:      file.Path,
		SizeBytes: file.SizeBytes,
		ModTime:   file.ModTime,
	}
}

func findLatestArtifact(root string, detectedAt time.Time, window time.Duration, suffix, prefix string) (*localArtifact, error) {
	if root == "" {
		return nil, nil
	}

	info, err := os.Stat(root)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var candidates []*localArtifact
	if !info.IsDir() {
		if !matchesArtifact(filepath.Base(root), suffix, prefix) {
			return nil, nil
		}
		if info.ModTime().Before(detectedAt.Add(-window)) || info.ModTime().After(detectedAt.Add(window)) {
			return nil, nil
		}
		return &localArtifact{
			Path:      root,
			SizeBytes: info.Size(),
			ModTime:   info.ModTime(),
		}, nil
	}

	err = filepath.WalkDir(root, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() || !matchesArtifact(d.Name(), suffix, prefix) {
			return nil
		}
		fi, err := d.Info()
		if err != nil {
			return err
		}
		if fi.ModTime().Before(detectedAt.Add(-window)) || fi.ModTime().After(detectedAt.Add(window)) {
			return nil
		}
		candidates = append(candidates, &localArtifact{
			Path:      path,
			SizeBytes: fi.Size(),
			ModTime:   fi.ModTime(),
		})
		return nil
	})
	if err != nil {
		return nil, err
	}
	if len(candidates) == 0 {
		return nil, nil
	}

	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].ModTime.After(candidates[j].ModTime)
	})
	return candidates[0], nil
}

func matchesArtifact(name, suffix, prefix string) bool {
	if suffix != "" && !strings.HasSuffix(strings.ToLower(name), strings.ToLower(suffix)) {
		return false
	}
	if prefix != "" && !strings.HasPrefix(name, prefix) {
		return false
	}
	return true
}

func (m *monitor) handleProcessExit(event *processExitEvent) {
	if m == nil || m.config == nil || event == nil {
		return
	}

	if err := uploadProcessExitLog(event, m.config.DataKitAddr); err != nil {
		log.Errorf("upload process exit event failed: %v", err)
		return
	}

	log.Infof("uploaded process exit event for service=%s pid=%d reason=%s", event.Service, event.PID, event.SuspectedReason)
}

func uploadProcessExitLog(event *processExitEvent, datakitAddr string) error {
	if event == nil {
		return nil
	}

	logURL, err := deriveLoggingURL(datakitAddr)
	if err != nil {
		return err
	}

	recentSamples, err := json.Marshal(event.RecentSamples)
	if err != nil {
		return err
	}

	message := fmt.Sprintf("process exit detected for pid=%d service=%s reason=%s", event.PID, event.Service, event.SuspectedReason)
	kvs := point.KVs{
		point.NewKV("message", message),
		point.NewKV("status", "warning"),
		point.NewKV("process_name", event.ProcessName),
		point.NewKV("suspected_reason", event.SuspectedReason),
		point.NewKV("exit_error", event.ExitError),
		point.NewKV("last_seen_at", formatLogTime(event.LastSeenAt)),
		point.NewKV("last_profile_time", formatLogTime(event.LastProfileTime)),
		point.NewKV("last_memory_pressure_at", formatLogTime(event.LastMemoryPressureAt)),
		point.NewKV("recent_mem_peak_percent", event.RecentMemPeakPercent),
		point.NewKV("recent_mem_peak_mb", event.RecentMemPeakMB),
		point.NewKV("recent_cpu_peak", event.RecentCPUPeak),
		point.NewKV("recent_samples", string(recentSamples)),
	}

	if event.RecentJFRArtifact != nil {
		kvs = kvs.Add("recent_jfr_path", event.RecentJFRArtifact.Path)
		kvs = kvs.Add("recent_jfr_size_bytes", event.RecentJFRArtifact.SizeBytes)
		kvs = kvs.Add("recent_jfr_mod_time", formatLogTime(event.RecentJFRArtifact.ModTime))
	}
	if event.RecentHProfArtifact != nil {
		kvs = kvs.Add("recent_hprof_path", event.RecentHProfArtifact.Path)
		kvs = kvs.Add("recent_hprof_size_bytes", event.RecentHProfArtifact.SizeBytes)
		kvs = kvs.Add("recent_hprof_mod_time", formatLogTime(event.RecentHProfArtifact.ModTime))
	}
	if event.LastProfileArtifact != nil {
		kvs = kvs.Add("last_profile_output_path", event.LastProfileArtifact.OutputPath)
		kvs = kvs.Add("last_profile_uploaded_at", formatLogTime(event.LastProfileArtifact.UploadedAt))
		kvs = kvs.Add("last_profile_start_time", event.LastProfileArtifact.StartTime)
		kvs = kvs.Add("last_profile_end_time", event.LastProfileArtifact.EndTime)
		kvs = kvs.Add("last_profile_event", event.LastProfileArtifact.Event)
		kvs = kvs.Add("last_profile_duration_sec", int64(event.LastProfileArtifact.DurationSec))
	}

	for _, tag := range event.Tags {
		if k, v, ok := strings.Cut(tag, ":"); ok && k != "" {
			kvs = kvs.AddTag(k, v)
		}
	}
	if event.Service != "" {
		kvs = kvs.AddTag("service", event.Service)
	}

	pt := point.NewPoint(processExitMeasurement, kvs, point.DefaultLoggingOptions()...)
	pt.SetTime(event.DetectedAt)

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
		return fmt.Errorf("upload process exit log failed, status=%d", resp.StatusCode)
	}

	return nil
}

func formatLogTime(ts time.Time) string {
	if ts.IsZero() {
		return ""
	}
	return ts.Format(time.RFC3339Nano)
}
