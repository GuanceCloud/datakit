// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package flameshot

import (
	"bytes"
	"context"
	"fmt"
	"io/fs"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"golang.org/x/net/context/ctxhttp"
)

const (
	defaultOOMHProfMatchWindow = 2 * time.Minute
	defaultOOMHProfInputName   = "flameshot"
	oomSummaryMeasurement      = "flameshot_oom_hprof"
)

type OOMEvent struct {
	DetectedAt   time.Time
	OOMKillDelta uint64
	Candidates   []*oomProcessCandidate
}

type oomProcessCandidate struct {
	Service     string
	PID         int32
	ProcessName string
	HProfPath   string
	Tags        []string
}

type hprofFile struct {
	Path      string
	SizeBytes int64
	ModTime   time.Time
}

type hprofSummary struct {
	Service        string
	PID            int32
	ProcessName    string
	DetectedAt     time.Time
	HProfPath      string
	HProfSizeBytes int64
	HProfModTime   time.Time
	MatchedBy      string
	SummaryText    string
	OOMKillDelta   uint64
}

type processedHProfStore struct {
	mu sync.Mutex
	m  map[string]struct{}
}

func newProcessedHProfStore() *processedHProfStore {
	return &processedHProfStore{m: make(map[string]struct{})}
}

func (s *processedHProfStore) seen(key string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, ok := s.m[key]
	return ok
}

func (s *processedHProfStore) mark(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.m[key] = struct{}{}
}

func getOOMHProfMatchWindow(cfg *Config) time.Duration {
	if cfg == nil || cfg.OOMHProfMatchWindow == "" {
		return defaultOOMHProfMatchWindow
	}

	d, err := time.ParseDuration(cfg.OOMHProfMatchWindow)
	if err != nil || d <= 0 {
		return defaultOOMHProfMatchWindow
	}

	return d
}

func deriveLoggingURL(datakitAddr string) (string, error) {
	if datakitAddr == "" {
		return "", fmt.Errorf("datakit addr is empty")
	}

	u, err := url.Parse(datakitAddr)
	if err != nil {
		return "", err
	}

	base := &url.URL{
		Scheme: u.Scheme,
		Host:   u.Host,
	}

	logURL := base.String() + "/v1/write/logging?input=" + defaultOOMHProfInputName
	return logURL, nil
}

func findMatchingHProf(root string, detectedAt time.Time, window time.Duration) (*hprofFile, error) {
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
	if !info.IsDir() {
		if !strings.HasSuffix(strings.ToLower(root), ".hprof") {
			return nil, nil
		}
		modTime := info.ModTime()
		if modTime.Before(detectedAt.Add(-window)) || modTime.After(detectedAt.Add(window)) {
			return nil, nil
		}
		return &hprofFile{
			Path:      root,
			SizeBytes: info.Size(),
			ModTime:   modTime,
		}, nil
	}

	var candidates []*hprofFile
	err = filepath.WalkDir(root, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(strings.ToLower(d.Name()), ".hprof") {
			return nil
		}
		fi, err := d.Info()
		if err != nil {
			return err
		}
		modTime := fi.ModTime()
		if modTime.Before(detectedAt.Add(-window)) || modTime.After(detectedAt.Add(window)) {
			return nil
		}
		candidates = append(candidates, &hprofFile{
			Path:      path,
			SizeBytes: fi.Size(),
			ModTime:   modTime,
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

func buildHProfSummary(candidate *oomProcessCandidate, event *OOMEvent, file *hprofFile) *hprofSummary {
	if candidate == nil || event == nil || file == nil {
		return nil
	}

	return &hprofSummary{
		Service:        candidate.Service,
		PID:            candidate.PID,
		ProcessName:    candidate.ProcessName,
		DetectedAt:     event.DetectedAt,
		HProfPath:      file.Path,
		HProfSizeBytes: file.SizeBytes,
		HProfModTime:   file.ModTime,
		MatchedBy:      "path+time_window",
		SummaryText: fmt.Sprintf(
			"OOM detected via cgroup oom_kill, heap dump found at %s (%d bytes)",
			file.Path, file.SizeBytes,
		),
		OOMKillDelta: event.OOMKillDelta,
	}
}

func processedHProfKey(file *hprofFile) string {
	if file == nil {
		return ""
	}
	return fmt.Sprintf("%s:%d:%d", file.Path, file.SizeBytes, file.ModTime.UnixNano())
}

var (
	heapDumpPathRe  = regexp.MustCompile(`(?:^|\s)-XX:HeapDumpPath=([^\s]+)`)
	heapDumpOnOOMRe = regexp.MustCompile(`(?:^|\s)-XX:\+HeapDumpOnOutOfMemoryError(?:\s|$)`)
)

func detectOOMHProfPath(pm *processM) string {
	if pm == nil || pm.Cmdline == "" {
		return ""
	}
	if !heapDumpOnOOMRe.MatchString(pm.Cmdline) {
		return ""
	}
	matches := heapDumpPathRe.FindStringSubmatch(pm.Cmdline)
	if len(matches) != 2 {
		return ""
	}
	return strings.Trim(matches[1], `"'`)
}

func uploadOOMSummaryLog(summary *hprofSummary, datakitAddr string, tags []string) error {
	if summary == nil {
		return nil
	}

	logURL, err := deriveLoggingURL(datakitAddr)
	if err != nil {
		return err
	}

	kvs := point.KVs{
		point.NewKV("message", summary.SummaryText),
		point.NewKV("status", "error"),
		point.NewKV("hprof_path", summary.HProfPath),
		point.NewKV("hprof_size_bytes", summary.HProfSizeBytes),
		point.NewKV("oom_kill_delta", int64(summary.OOMKillDelta)),
		point.NewKV("matched_by", summary.MatchedBy),
		point.NewKV("process_name", summary.ProcessName),
	}
	opts := point.DefaultLoggingOptions()
	for _, tag := range tags {
		if k, v, ok := strings.Cut(tag, ":"); ok && k != "" {
			kvs = kvs.AddTag(k, v)
		}
	}
	if summary.Service != "" {
		kvs = kvs.AddTag("service", summary.Service)
	}

	pt := point.NewPoint(oomSummaryMeasurement, kvs, opts...)
	pt.SetTime(summary.DetectedAt)
	data := []byte(pt.LineProto() + "\n")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, logURL, bytes.NewBuffer(data))
	if err != nil {
		return err
	}

	resp, err := ctxhttp.Do(ctx, http.DefaultClient, req)
	if err != nil {
		return err
	}
	defer resp.Body.Close() //nolint:errcheck

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("upload oom summary failed, status=%d", resp.StatusCode)
	}

	return nil
}

func (m *monitor) handleOOMEvent(event *OOMEvent) {
	if m == nil || m.config == nil || event == nil {
		return
	}

	match, candidate, err := m.findOOMHProfMatch(event)
	if err != nil {
		log.Errorf("find matching hprof failed: %v", err)
		return
	}
	if match == nil || candidate == nil {
		log.Infof("oom event detected at %s but no matching hprof found", event.DetectedAt.Format(time.RFC3339))
		return
	}

	key := processedHProfKey(match)
	if m.processedHProf.seen(key) {
		return
	}

	summary := buildHProfSummary(candidate, event, match)
	if summary == nil {
		return
	}

	if err := uploadOOMSummaryLog(summary, m.config.DataKitAddr, candidate.Tags); err != nil {
		log.Errorf("upload oom hprof summary failed: %v", err)
		return
	}

	m.processedHProf.mark(key)
	log.Infof("uploaded oom hprof summary for service=%s pid=%d hprof=%s", candidate.Service, candidate.PID, match.Path)
}

func (m *monitor) findOOMHProfMatch(event *OOMEvent) (*hprofFile, *oomProcessCandidate, error) {
	if event == nil {
		return nil, nil, nil
	}

	window := getOOMHProfMatchWindow(m.config)
	var (
		bestFile      *hprofFile
		bestCandidate *oomProcessCandidate
	)

	for _, candidate := range event.Candidates {
		if candidate == nil || candidate.HProfPath == "" {
			continue
		}

		file, err := findMatchingHProf(candidate.HProfPath, event.DetectedAt, window)
		if err != nil {
			return nil, nil, err
		}
		if file == nil {
			log.Infof("oom event for service=%s pid=%d but no hprof found in %s", candidate.Service, candidate.PID, candidate.HProfPath)
			continue
		}
		if bestFile == nil || file.ModTime.After(bestFile.ModTime) {
			bestFile = file
			bestCandidate = candidate
		}
	}

	return bestFile, bestCandidate, nil
}
