// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package flameshot

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"

	pprofile "github.com/google/pprof/profile"
)

const (
	goPProfTypeCPU       = "cpu"
	goPProfTypeGoroutine = "goroutine"
	goPProfTypeHeap      = "heap"
	goPProfTypeMutex     = "mutex"
	goPProfTypeBlock     = "block"

	defaultGoPProfType       = goPProfTypeCPU
	defaultGoPProfMaxBytes   = int64(128 << 20)
	defaultGoPProfSlack      = 15 * time.Second
	defaultGoPProfGoroutines = "goroutines"
)

var allGoPProfTypes = []string{
	goPProfTypeCPU,
	goPProfTypeGoroutine,
	goPProfTypeHeap,
	goPProfTypeMutex,
	goPProfTypeBlock,
}

type goPProfValueType struct {
	Type string
	Unit string
}

type goPProfItem struct {
	path        string
	fileName    string
	query       url.Values
	deltaValues []goPProfValueType
}

var goPProfItems = map[string]goPProfItem{
	goPProfTypeCPU: {
		path:     "/debug/pprof/profile",
		fileName: "cpu.pprof",
	},
	goPProfTypeGoroutine: {
		path:     "/debug/pprof/goroutine",
		fileName: "goroutines.pprof",
		query:    url.Values{"debug": []string{"0"}},
	},
	goPProfTypeHeap: {
		path:     "/debug/pprof/heap",
		fileName: "delta-heap.pprof",
		deltaValues: []goPProfValueType{
			{Type: "alloc_objects", Unit: "count"},
			{Type: "alloc_space", Unit: "bytes"},
		},
	},
	goPProfTypeMutex: {
		path:     "/debug/pprof/mutex",
		fileName: "delta-mutex.pprof",
		deltaValues: []goPProfValueType{
			{Type: "contentions", Unit: "count"},
			{Type: "delay", Unit: "nanoseconds"},
		},
	},
	goPProfTypeBlock: {
		path:     "/debug/pprof/block",
		fileName: "delta-block.pprof",
		deltaValues: []goPProfValueType{
			{Type: "contentions", Unit: "count"},
			{Type: "delay", Unit: "nanoseconds"},
		},
	},
}

type goPProfDeltaStore struct {
	mu       sync.Mutex
	profiles map[string]*pprofile.Profile
}

func newGoPProfDeltaStore() *goPProfDeltaStore {
	return &goPProfDeltaStore{
		profiles: make(map[string]*pprofile.Profile),
	}
}

func runGoPProf(ctx context.Context, stats *triggerStats) error {
	if stats == nil {
		return fmt.Errorf("trigger stats is nil")
	}
	if stats.PProfURL == "" {
		return fmt.Errorf("go pprof_url is required")
	}

	baseURL, err := url.Parse(stats.PProfURL)
	if err != nil {
		return fmt.Errorf("parse go pprof_url: %w", err)
	}
	if baseURL.Scheme == "" || baseURL.Host == "" {
		return fmt.Errorf("go pprof_url must include scheme and host")
	}

	types := resolveGoPProfTypes(stats)
	timeout := goPProfRequestTimeout(stats)
	client := &http.Client{Timeout: timeout}
	stats.startTime = time.Now().Format(time.RFC3339Nano)

	attachments := make([]*profileAttachment, 0, len(types))
	errs := make([]string, 0)
	for _, typ := range types {
		item, ok := goPProfItems[typ]
		if !ok {
			continue
		}

		attachment, err := pullGoPProfItem(ctx, client, baseURL, stats, typ, item)
		if err != nil {
			errs = append(errs, fmt.Sprintf("%s: %v", typ, err))
			log.Warnf("pull go pprof %s failed: %v", typ, err)
			continue
		}
		if attachment != nil {
			attachments = append(attachments, attachment)
		}
	}

	stats.endTime = time.Now().Format(time.RFC3339Nano)
	stats.Attachments = attachments
	stats.OutputFile = ""
	if stats.Event == "" {
		stats.Event = strings.Join(types, ",")
	}

	if len(attachments) == 0 {
		if len(errs) > 0 {
			return fmt.Errorf("no go pprof data collected: %s", strings.Join(errs, "; "))
		}
		return fmt.Errorf("no go pprof data collected")
	}

	return nil
}

func pullGoPProfItem(
	ctx context.Context,
	client *http.Client,
	baseURL *url.URL,
	stats *triggerStats,
	profileType string,
	item goPProfItem,
) (*profileAttachment, error) {
	data, err := pullGoPProfData(ctx, client, buildGoPProfURL(baseURL, item, stats.Duration))
	if err != nil {
		return nil, err
	}

	if len(item.deltaValues) == 0 {
		return &profileAttachment{
			FieldName: item.fileName,
			FileName:  item.fileName,
			Data:      data,
		}, nil
	}

	curProf, err := pprofile.ParseData(data)
	if err != nil {
		return nil, fmt.Errorf("parse go pprof %s: %w", profileType, err)
	}

	store := stats.goPProfDeltas
	if store == nil {
		store = newGoPProfDeltaStore()
	}

	deltaProfile, ok, err := store.delta(goPProfDeltaKey(stats, profileType), curProf, item.deltaValues)
	if err != nil {
		return nil, err
	}
	if !ok {
		log.Infof("go pprof %s first sample stored, skip delta upload", profileType)
		return nil, nil
	}

	buf := &bytes.Buffer{}
	if err := deltaProfile.Write(buf); err != nil {
		return nil, fmt.Errorf("write go pprof %s delta: %w", profileType, err)
	}

	return &profileAttachment{
		FieldName: item.fileName,
		FileName:  item.fileName,
		Data:      buf.Bytes(),
	}, nil
}

func pullGoPProfData(ctx context.Context, client *http.Client, u string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close() //nolint:errcheck

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4<<10))
		return nil, fmt.Errorf("invalid response status: %s, body: %s", resp.Status, string(body))
	}

	data, err := io.ReadAll(io.LimitReader(resp.Body, defaultGoPProfMaxBytes+1))
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > defaultGoPProfMaxBytes {
		return nil, fmt.Errorf("go pprof response exceeds %d bytes", defaultGoPProfMaxBytes)
	}
	return data, nil
}

func buildGoPProfURL(baseURL *url.URL, item goPProfItem, duration int) string {
	u := *baseURL
	u.Path = joinGoPProfPath(baseURL.Path, item.path)
	q := cloneURLValues(baseURL.Query())
	for key, vals := range item.query {
		for _, val := range vals {
			q.Add(key, val)
		}
	}
	if item.path == goPProfItems[goPProfTypeCPU].path {
		if duration <= 0 {
			duration = DefaultDuration
		}
		q.Set("seconds", strconv.Itoa(duration))
	}
	u.RawQuery = q.Encode()
	return u.String()
}

func joinGoPProfPath(basePath, itemPath string) string {
	if basePath == "" || basePath == "/" {
		return itemPath
	}
	return "/" + path.Join(strings.TrimPrefix(basePath, "/"), strings.TrimPrefix(itemPath, "/"))
}

func cloneURLValues(v url.Values) url.Values {
	out := make(url.Values, len(v))
	for key, vals := range v {
		out[key] = append([]string(nil), vals...)
	}
	return out
}

func resolveGoPProfTypes(stats *triggerStats) []string {
	var raw []string
	if stats != nil && len(stats.PProfTypes) > 0 {
		raw = stats.PProfTypes
	} else if stats != nil && stats.Event != "" {
		raw = strings.Split(stats.Event, ",")
	}

	seen := make(map[string]struct{}, len(raw))
	resolved := make([]string, 0, len(raw))
	for _, typ := range raw {
		typ = normalizeGoPProfType(typ)
		if typ == "all" || typ == "--all" {
			return append([]string(nil), allGoPProfTypes...)
		}
		if typ == "" {
			continue
		}
		if _, ok := goPProfItems[typ]; !ok {
			log.Warnf("unsupported go pprof type: %s", typ)
			continue
		}
		if _, ok := seen[typ]; ok {
			continue
		}
		seen[typ] = struct{}{}
		resolved = append(resolved, typ)
	}
	if len(resolved) == 0 {
		return []string{defaultGoPProfType}
	}
	return resolved
}

func normalizeGoPProfType(typ string) string {
	typ = strings.ToLower(strings.TrimSpace(typ))
	switch typ {
	case defaultGoPProfGoroutines:
		return goPProfTypeGoroutine
	default:
		return typ
	}
}

func goPProfRequestTimeout(stats *triggerStats) time.Duration {
	duration := DefaultDuration
	if stats != nil && stats.Duration > 0 {
		duration = stats.Duration
	}
	timeout := time.Duration(duration)*time.Second + defaultGoPProfSlack
	if stats == nil || stats.PProfTimeout == "" {
		return timeout
	}

	configured, err := time.ParseDuration(stats.PProfTimeout)
	if err != nil || configured <= 0 {
		log.Warnf("invalid go pprof timeout %q, use %s", stats.PProfTimeout, timeout)
		return timeout
	}
	if configured < timeout {
		log.Warnf("go pprof timeout %s is shorter than duration plus slack %s, use %s", configured, timeout, timeout)
		return timeout
	}
	return configured
}

func goPProfDeltaKey(stats *triggerStats, profileType string) string {
	if stats == nil {
		return profileType
	}
	return fmt.Sprintf("%s:%d:%s:%s", stats.Service, stats.PID, stats.PProfURL, profileType)
}

func (s *goPProfDeltaStore) delta(
	key string,
	cur *pprofile.Profile,
	sampleTypes []goPProfValueType,
) (*pprofile.Profile, bool, error) {
	if s == nil {
		return nil, false, nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	prev, ok := s.profiles[key]
	s.profiles[key] = cur
	if !ok {
		return nil, false, nil
	}

	deltaProfile, err := getGoDeltaProfile(prev, cur, sampleTypes)
	if err != nil {
		return nil, false, err
	}
	deltaProfile.TimeNanos = cur.TimeNanos
	deltaProfile.DurationNanos = cur.TimeNanos - prev.TimeNanos
	return deltaProfile, true, nil
}

func getGoDeltaProfile(a, b *pprofile.Profile, sampleTypes []goPProfValueType) (*pprofile.Profile, error) {
	if a == nil || b == nil {
		return nil, errors.New("profile expected not to be nil")
	}

	ratios := make([]float64, len(a.SampleType))
	found := 0
	for i, st := range a.SampleType {
		for _, deltaSt := range sampleTypes {
			if deltaSt.Type == st.Type && deltaSt.Unit == st.Unit {
				ratios[i] = -1
				found++
			}
		}
	}
	if found != len(sampleTypes) {
		return nil, errors.New("one or more sample type(s) was not found in the profile")
	}

	if err := a.ScaleN(ratios); err != nil {
		return nil, fmt.Errorf("failed scaling profile a: %w", err)
	}

	delta, err := pprofile.Merge([]*pprofile.Profile{a, b})
	if err != nil {
		return nil, err
	}
	return delta, delta.CheckValid()
}
