// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package promscrape

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"
)

// Keep this low enough to pick up rotated token files before most token
// lifetimes end, while avoiding per-scrape file reads.
const defaultBearerTokenRefreshInterval = 15 * time.Minute

type cachedFileBearerTokenSource struct {
	path            string
	refreshInterval time.Duration

	mu    sync.RWMutex
	token string
	t     time.Time

	now func() time.Time
}

func newCachedFileBearerTokenSource(path string, refreshInterval time.Duration) *cachedFileBearerTokenSource {
	if refreshInterval <= 0 {
		refreshInterval = defaultBearerTokenRefreshInterval
	}

	return &cachedFileBearerTokenSource{
		path:            path,
		refreshInterval: refreshInterval,
		now:             time.Now,
	}
}

func (ts *cachedFileBearerTokenSource) Token() (string, error) {
	now := ts.now()

	ts.mu.RLock()
	token := ts.token
	readAt := ts.t
	ts.mu.RUnlock()

	if token != "" && readAt.Add(ts.refreshInterval).After(now) {
		return token, nil
	}

	ts.mu.Lock()
	defer ts.mu.Unlock()

	if ts.token != "" && ts.t.Add(ts.refreshInterval).After(now) {
		return ts.token, nil
	}

	token, err := readBearerTokenFile(ts.path)
	if err != nil {
		if ts.token == "" {
			return "", err
		}
		return ts.token, nil
	}

	ts.token = token
	ts.t = now

	return token, nil
}

func (ts *cachedFileBearerTokenSource) ResetToken() {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	ts.token = ""
	ts.t = time.Time{}
}

func readBearerTokenFile(path string) (string, error) {
	token, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read bearer token file %q: %w", path, err)
	}

	tokenText := strings.TrimSpace(string(token))
	if tokenText == "" {
		return "", fmt.Errorf("read empty bearer token from file %q", path)
	}

	return tokenText, nil
}
