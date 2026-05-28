// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package httpapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	dkfilter "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/filter"
)

func TestAPIDatakitPullDefaultFilters(t *testing.T) {
	setupDatakitPullTestDataDir(t)

	router := setupDatakitPullTestRouter(t)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/datakit/pull?filters=true", nil)
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var got datakitPullResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &got))
	assert.Equal(t, "30m", got.PullInterval)
	assert.Equal(t, dkfilter.FilterConditions{}, got.Filters[datakit.CategoryRUM])
	assert.Equal(t, dkfilter.FilterConditions{}, got.Filters[datakit.CategoryLogging])
	assert.NotContains(t, rec.Body.String(), "content")
}

func TestAPIDatakitPullFromCache(t *testing.T) {
	dir := setupDatakitPullTestDataDir(t)

	body := fmt.Sprintf(`{
		"filters": {
			"rum": ["{ source = 'resource' and app_id = 'appid_xxx' }"],
			"logging": ["{ source = 'browser_log' and message match ['timeout.*'] }"],
			"metric": ["{ measurement = 'cpu' }"]
		},
		"pull_interval": %d
	}`, 30*time.Minute)
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".pull"), []byte(body), 0o600))

	router := setupDatakitPullTestRouter(t)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/datakit/pull?filters=true", nil)
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var got datakitPullResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &got))
	assert.Equal(t, "30m", got.PullInterval)
	assert.Equal(t, dkfilter.FilterConditions{`{ source = 'resource' and app_id = 'appid_xxx' }`}, got.Filters[datakit.CategoryRUM])
	assert.Equal(t, dkfilter.FilterConditions{`{ source = 'browser_log' and message match ['timeout.*'] }`}, got.Filters[datakit.CategoryLogging])
	assert.NotContains(t, got.Filters, datakit.CategoryMetric)
}

func TestAPIDatakitPullInvalidRequest(t *testing.T) {
	setupDatakitPullTestDataDir(t)

	router := setupDatakitPullTestRouter(t)

	for _, req := range []*http.Request{
		httptest.NewRequest(http.MethodGet, "/v1/datakit/pull", nil),
		httptest.NewRequest(http.MethodGet, "/v1/datakit/pull?filters=false", nil),
		httptest.NewRequest(http.MethodGet, "/v1/datakit/pull?filters=true", bytes.NewBufferString("body")),
	} {
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	}
}

func TestFormatDatakitPullInterval(t *testing.T) {
	assert.Equal(t, "1h", formatDatakitPullInterval(time.Hour))
	assert.Equal(t, "30m", formatDatakitPullInterval(30*time.Minute))
	assert.Equal(t, "1m20s", formatDatakitPullInterval(80*time.Second))
	assert.Equal(t, "45s", formatDatakitPullInterval(45*time.Second))
}

func TestDatakitPullCache(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".pull")

	body := fmt.Sprintf(`{
		"filters": {
			"rum": ["{ app_id = 'app_1' }"]
		},
		"pull_interval": %d
	}`, time.Minute)
	require.NoError(t, os.WriteFile(path, []byte(body), 0o600))

	cache := &datakitPullCache{}
	got := cache.load(path)
	require.Equal(t, dkfilter.FilterConditions{`{ app_id = 'app_1' }`}, got.Filters[datakit.CategoryRUM])

	body = fmt.Sprintf(`{
		"filters": {
			"rum": ["{ app_id = 'app_2' }"],
			"metric": ["{ measurement = 'cpu' }"]
		},
		"pull_interval": %d
	}`, 2*time.Minute)
	require.NoError(t, os.WriteFile(path, []byte(body), 0o600))

	got = cache.load(path)
	require.Equal(t, dkfilter.FilterConditions{`{ app_id = 'app_1' }`}, got.Filters[datakit.CategoryRUM])
	require.NotContains(t, got.Filters, datakit.CategoryMetric)

	cache.checkedAt = time.Now().Add(-datakitPullCacheTTL - time.Second)

	got = cache.load(path)
	require.Equal(t, dkfilter.FilterConditions{`{ app_id = 'app_2' }`}, got.Filters[datakit.CategoryRUM])
	require.NotContains(t, got.Filters, datakit.CategoryMetric)
	require.Equal(t, "2m", got.PullInterval)
}

func setupDatakitPullTestDataDir(t *testing.T) string {
	t.Helper()

	oldDataDir := datakit.DataDir
	dir := t.TempDir()
	datakit.DataDir = dir

	t.Cleanup(func() {
		datakit.DataDir = oldDataDir
	})

	return dir
}

func TestAPIDatakitPullAutoAddedToAPIWhitelist(t *testing.T) {
	setupDatakitPullTestDataDir(t)

	router := setupDatakitPullTestRouter(t)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/datakit/pull?filters=true", nil)
	req.RemoteAddr = "192.168.1.100:12345"
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var got datakitPullResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &got))
	assert.Equal(t, "30m", got.PullInterval)
}

func TestAPIDatakitPullNotRegisteredByDefault(t *testing.T) {
	setupDatakitPullTestDataDir(t)
	CleanHTTPHandler()
	t.Cleanup(CleanHTTPHandler)

	hs := defaultHTTPServerConf()
	hs.apiConfig.DisableWhitelist = true
	router := setupRouter(hs)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/datakit/pull?filters=true", nil)
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusNotFound, rec.Code)
}

func setupDatakitPullTestRouter(t *testing.T) http.Handler {
	t.Helper()

	CleanHTTPHandler()
	RegDatakitPullHTTPRoute()
	t.Cleanup(CleanHTTPHandler)

	hs := defaultHTTPServerConf()
	hs.apiConfig.DisableWhitelist = false
	hs.apiConfig.PublicAPIs = nil
	return setupRouter(hs)
}
