// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package logging

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const (
	DefaultEndpoint = "https://logging.googleapis.com"
	MaxSeenEntries  = 100000
)

type Config struct {
	ProjectID   string
	ClusterName string
	Location    string
	Endpoint    string
}

type State struct {
	Watermark time.Time        `json:"watermark"`
	Seen      map[string]int64 `json:"seen"`
}

type ListRequest struct {
	ResourceNames []string `json:"resourceNames"`
	Filter        string   `json:"filter"`
	OrderBy       string   `json:"orderBy"`
	PageSize      int      `json:"pageSize"`
	PageToken     string   `json:"pageToken,omitempty"`
}

type listResponse struct {
	Entries       []Entry `json:"entries"`
	NextPageToken string  `json:"nextPageToken"`
}

type Entry struct {
	LogName          string `json:"logName"`
	Timestamp        string `json:"timestamp"`
	ReceiveTimestamp string `json:"receiveTimestamp"`
	InsertID         string `json:"insertId"`
	Severity         string `json:"severity"`

	TextPayload  string          `json:"textPayload"`
	JSONPayload  json.RawMessage `json:"jsonPayload"`
	ProtoPayload json.RawMessage `json:"protoPayload"`

	Resource struct {
		Type   string            `json:"type"`
		Labels map[string]string `json:"labels"`
	} `json:"resource"`
}

func ListEntries(ctx context.Context, client *http.Client, cfg Config, start, end time.Time) ([]Entry, error) {
	if cfg.ProjectID == "" {
		return nil, fmt.Errorf("gcp project id is required")
	}
	if cfg.ClusterName == "" {
		return nil, fmt.Errorf("gke cluster name is required")
	}
	if cfg.Location == "" {
		return nil, fmt.Errorf("gke cluster location is required")
	}
	if client == nil {
		return nil, fmt.Errorf("http client is required")
	}
	if cfg.Endpoint == "" {
		cfg.Endpoint = DefaultEndpoint
	}

	filterText := fmt.Sprintf(
		`resource.type="k8s_container" AND resource.labels.cluster_name=%q AND resource.labels.location=%q `+
			`AND timestamp>=%q AND timestamp<%q AND (`+
			`logName=%q OR logName=%q)`,
		cfg.ClusterName,
		cfg.Location,
		start.UTC().Format(time.RFC3339Nano),
		end.UTC().Format(time.RFC3339Nano),
		fmt.Sprintf("projects/%s/logs/stdout", cfg.ProjectID),
		fmt.Sprintf("projects/%s/logs/stderr", cfg.ProjectID),
	)

	var result []Entry
	pageToken := ""
	for {
		body, err := json.Marshal(ListRequest{
			ResourceNames: []string{"projects/" + cfg.ProjectID},
			Filter:        filterText,
			OrderBy:       "timestamp desc",
			PageSize:      1000,
			PageToken:     pageToken,
		})
		if err != nil {
			return nil, fmt.Errorf("encode logging request: %w", err)
		}

		req, err := http.NewRequestWithContext(
			ctx,
			http.MethodPost,
			strings.TrimRight(cfg.Endpoint, "/")+"/v2/entries:list",
			bytes.NewReader(body),
		)
		if err != nil {
			return nil, fmt.Errorf("create logging request: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("send logging request: %w", err)
		}
		var page listResponse
		if err := decodeResponse(resp, &page); err != nil {
			return nil, err
		}
		result = append(result, page.Entries...)
		if page.NextPageToken == "" {
			break
		}
		pageToken = page.NextPageToken
	}
	return result, nil
}

func LoadState(path string) (State, error) {
	state := State{Seen: map[string]int64{}}
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return state, nil
	}
	if err != nil {
		return state, err
	}
	if err := json.Unmarshal(data, &state); err != nil {
		return state, err
	}
	if state.Seen == nil {
		state.Seen = map[string]int64{}
	}
	return state, nil
}

func SaveState(path string, state State) error {
	data, err := json.Marshal(state)
	if err != nil {
		return err
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

func AdvanceWatermark(state *State, now time.Time) {
	if state != nil && now.After(state.Watermark) {
		state.Watermark = now
	}
}

func PruneSeen(state *State, now time.Time, overlap time.Duration, limit int) {
	if state == nil || state.Seen == nil {
		return
	}
	cutoff := now.Add(-overlap * 2).Unix()
	for key, seenAt := range state.Seen {
		if seenAt < cutoff {
			delete(state.Seen, key)
		}
	}
	if limit <= 0 || len(state.Seen) <= limit {
		return
	}

	type seenEntry struct {
		key    string
		seenAt int64
	}
	entries := make([]seenEntry, 0, len(state.Seen))
	for key, seenAt := range state.Seen {
		entries = append(entries, seenEntry{key: key, seenAt: seenAt})
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].seenAt < entries[j].seenAt })
	for i := 0; i < len(entries)-limit; i++ {
		delete(state.Seen, entries[i].key)
	}
}

func Timestamp(entry *Entry) (time.Time, bool) {
	if entry == nil {
		return time.Time{}, false
	}
	for _, value := range []string{entry.Timestamp, entry.ReceiveTimestamp} {
		if value == "" {
			continue
		}
		timestamp, err := time.Parse(time.RFC3339Nano, value)
		if err == nil {
			return timestamp, true
		}
	}
	return time.Time{}, false
}

func EntryKey(entry *Entry) string {
	if entry == nil {
		return ""
	}
	if entry.InsertID != "" && entry.Timestamp != "" {
		return entry.Timestamp + "\x00" + entry.InsertID
	}
	data, _ := json.Marshal(entry)
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func Message(entry *Entry) string {
	if entry == nil {
		return ""
	}
	if entry.TextPayload != "" {
		return entry.TextPayload
	}
	for _, payload := range []json.RawMessage{entry.JSONPayload, entry.ProtoPayload} {
		if len(payload) == 0 || bytes.Equal(payload, []byte("null")) {
			continue
		}
		var compact bytes.Buffer
		if err := json.Compact(&compact, payload); err == nil {
			return compact.String()
		}
		return string(payload)
	}
	return ""
}

func decodeResponse(resp *http.Response, dst any) error {
	defer func() {
		_ = resp.Body.Close()
	}()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 64*1024))
		return fmt.Errorf("gcp api returned %s: %s", resp.Status, strings.TrimSpace(string(body)))
	}
	if err := json.NewDecoder(resp.Body).Decode(dst); err != nil {
		return fmt.Errorf("decode gcp api response: %w", err)
	}
	return nil
}
