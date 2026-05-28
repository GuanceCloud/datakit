// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package promscrape

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"github.com/stretchr/testify/require"
)

func TestBearerTokenFileRefreshesAfterInterval(t *testing.T) {
	tokenFile := filepath.Join(t.TempDir(), "token")
	require.NoError(t, os.WriteFile(tokenFile, []byte("token-1\n"), 0o600))

	var (
		mu    sync.Mutex
		auths []string
	)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		auths = append(auths, r.Header.Get("Authorization"))
		mu.Unlock()

		_, _ = fmt.Fprintln(w, "sample_metric 1")
	}))
	defer server.Close()

	scraper, err := NewPromScraper(
		withBearerTokenFile(tokenFile, 50*time.Millisecond),
		WithCallback(func([]*point.Point) error { return nil }),
	)
	require.NoError(t, err)

	require.NoError(t, scraper.ScrapeURL(server.URL))

	require.NoError(t, os.WriteFile(tokenFile, []byte("token-2\n"), 0o600))
	require.NoError(t, scraper.ScrapeURL(server.URL))

	time.Sleep(60 * time.Millisecond)
	require.NoError(t, scraper.ScrapeURL(server.URL))

	mu.Lock()
	defer mu.Unlock()
	require.Equal(t, []string{"Bearer token-1", "Bearer token-1", "Bearer token-2"}, auths)
}

func TestBearerTokenFileResetsAfterUnauthorized(t *testing.T) {
	tokenFile := filepath.Join(t.TempDir(), "token")
	require.NoError(t, os.WriteFile(tokenFile, []byte("token-1\n"), 0o600))

	var (
		mu       sync.Mutex
		expected = "Bearer token-1"
		auths    []string
	)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")

		mu.Lock()
		auths = append(auths, auth)
		want := expected
		mu.Unlock()

		if auth != want {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		_, _ = fmt.Fprintln(w, "sample_metric 1")
	}))
	defer server.Close()

	scraper, err := NewPromScraper(
		withBearerTokenFile(tokenFile, time.Hour),
		WithCallback(func([]*point.Point) error { return nil }),
	)
	require.NoError(t, err)

	require.NoError(t, scraper.ScrapeURL(server.URL))

	require.NoError(t, os.WriteFile(tokenFile, []byte("token-2\n"), 0o600))
	mu.Lock()
	expected = "Bearer token-2"
	mu.Unlock()

	err = scraper.ScrapeURL(server.URL)
	require.Error(t, err)

	var scrapeErr *ScrapeError
	require.True(t, errors.As(err, &scrapeErr))
	require.Equal(t, http.StatusUnauthorized, scrapeErr.StatusCode)
	require.True(t, IsRetryableScrapeError(err))

	require.NoError(t, scraper.ScrapeURL(server.URL))

	mu.Lock()
	defer mu.Unlock()
	require.Equal(t, []string{"Bearer token-1", "Bearer token-1", "Bearer token-2"}, auths)
}

func TestBearerTokenFileDoesNotOverrideAuthorizationHeader(t *testing.T) {
	tokenFile := filepath.Join(t.TempDir(), "missing-token")

	var (
		mu   sync.Mutex
		auth string
	)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		auth = r.Header.Get("Authorization")
		mu.Unlock()

		_, _ = fmt.Fprintln(w, "sample_metric 1")
	}))
	defer server.Close()

	scraper, err := NewPromScraper(
		WithHTTPHeader(map[string]string{"Authorization": "Bearer static-token"}),
		WithBearerTokenFile(tokenFile),
		WithCallback(func([]*point.Point) error { return nil }),
	)
	require.NoError(t, err)
	require.NoError(t, scraper.ScrapeURL(server.URL))

	mu.Lock()
	defer mu.Unlock()
	require.Equal(t, "Bearer static-token", auth)
}
