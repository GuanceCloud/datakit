// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package gitlab

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/GuanceCloud/cliutils/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/promscrape"
)

func TestPrometheusScraping(t *testing.T) {
	// Create a test server that serves the prometheus.data content
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Serve the prometheus.data content
		data := `# HELP gitlab_cache_misses_total Multiprocess metric
# TYPE gitlab_cache_misses_total counter
gitlab_cache_misses_total{controller="HelpController",action="index",feature_category="not_owned"} 268
gitlab_cache_misses_total{controller="MetricsController",action="index",feature_category=""} 4
gitlab_cache_misses_total{controller="RootController",action="index",feature_category="projects"} 0

# HELP gitlab_cache_operations_total Multiprocess metric
# TYPE gitlab_cache_operations_total counter
gitlab_cache_operations_total{controller="HelpController",action="index",feature_category="not_owned",operation="read"} 622
gitlab_cache_operations_total{controller="HelpController",action="index",feature_category="not_owned",operation="write"} 235
gitlab_cache_operations_total{controller="MetricsController",action="index",feature_category="",operation="read"} 8
gitlab_cache_operations_total{controller="MetricsController",action="index",feature_category="",operation="write"} 4
gitlab_cache_operations_total{controller="RootController",action="index",feature_category="projects",operation="read"} 0
gitlab_cache_operations_total{controller="RootController",action="index",feature_category="projects",operation="write"} 0

# HELP gitlab_database_connection_pool_busy Multiprocess metric
# TYPE gitlab_database_connection_pool_busy gauge
gitlab_database_connection_pool_busy{host="/var/opt/gitlab/postgresql",port="5432",class="ActiveRecord::Base",pid="puma_0"} 0
gitlab_database_connection_pool_busy{host="/var/opt/gitlab/postgresql",port="5432",class="ActiveRecord::Base",pid="puma_1"} 0
gitlab_database_connection_pool_busy{host="/var/opt/gitlab/postgresql",port="5432",class="ActiveRecord::Base",pid="puma_2"} 0`

		w.Header().Set("Content-Type", "text/plain; version=0.0.4")
		w.Write([]byte(data))
	}))
	defer ts.Close()

	// Create input with test server URL
	ipt := defaultInput()
	ipt.URL = ts.URL
	ipt.EnableCollect = true

	// We can't easily test the full scraping without mocking the scraper,
	// but we can verify that the input can be created and configured
	assert.NotNil(t, ipt)
	assert.True(t, ipt.EnableCollect)
	assert.Equal(t, ts.URL, ipt.URL)

	// Test that the sample config is valid
	sample := ipt.SampleConfig()
	assert.Contains(t, sample, "[[inputs.gitlab]]")
	assert.Contains(t, sample, "prometheus_url")
	assert.Contains(t, sample, "enable_collect")
}

func TestInputConfig(t *testing.T) {
	// Test TLS configuration
	ipt := &Input{
		EnableCollect:      true,
		URL:                "https://gitlab.example.com/-/metrics",
		Interval:           "30s",
		TLSCert:            "/path/to/cert.pem",
		TLSKey:             "/path/to/key.pem",
		TLSCA:              "/path/to/ca.pem",
		InsecureSkipVerify: true,
		Tags: map[string]string{
			"environment": "production",
			"team":        "platform",
		},
	}

	assert.True(t, ipt.hasTLSConfig())
	assert.Equal(t, "30s", ipt.Interval)
	assert.Equal(t, "production", ipt.Tags["environment"])
	assert.Equal(t, "platform", ipt.Tags["team"])
}

func TestBuildScraperOptions(t *testing.T) {
	ipt := defaultInput()
	ipt.URL = "http://localhost:8080/-/metrics"
	ipt.Tags = map[string]string{"env": "test"}
	ipt.HTTPHeaders = map[string]string{"X-Custom-Header": "value"}

	// Initialize logger
	ipt.logger = logger.SLogger(inputName)

	// Parse URL to set endpoint
	err := ipt.parseURL()
	assert.NoError(t, err)

	tags := ipt.buildTags()
	opts := ipt.buildScraperOptions(tags)

	assert.NotNil(t, opts)
	// Should have at least WithSource, WithMeasurement, WithHTTPHeader, WithExtraTags, etc.
	assert.Greater(t, len(opts), 3)
}

func TestScrapeMetricNameConversion(t *testing.T) {
	data := `# HELP gitlab_cache_misses_total Multiprocess metric
# TYPE gitlab_cache_misses_total counter
gitlab_cache_misses_total{controller="RootController",action="index",feature_category="projects"} 0`

	feeder := NewMockedFeederEmpty()
	ipt := defaultInput()
	ipt.feeder = feeder
	ipt.logger = logger.SLogger(inputName)
	ipt.URL = "http://127.0.0.1:8080/-/metrics"

	err := ipt.parseURL()
	require.NoError(t, err)

	scraper, err := promscrape.NewPromScraper(ipt.buildScraperOptions(ipt.buildTags())...)
	require.NoError(t, err)
	ipt.scraper = scraper
	ipt.lastStart = time.Now()

	err = ipt.scraper.ParserStream(bytes.NewBufferString(data))
	require.NoError(t, err)

	select {
	case pts := <-feeder.ch:
		require.Len(t, pts, 1)

		pt := pts[0]
		assert.Equal(t, "gitlab", pt.Name())
		assert.Equal(t, "RootController", pt.GetTag("controller"))
		assert.Equal(t, "index", pt.GetTag("action"))
		assert.Equal(t, "projects", pt.GetTag("feature_category"))
		assert.Equal(t, float64(0), pt.Get("cache_misses_total"))
		assert.Nil(t, pt.Get("gitlab_cache_misses_total"))

	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for scraped points")
	}
}

func TestAuthSetup(t *testing.T) {
	ipt := defaultInput()
	ipt.BearerTokenFile = "/tmp/test-token.txt"

	// This will fail because file doesn't exist, but we can test the error handling
	err := ipt.setupAuth()
	// Expect error because file doesn't exist
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "read bearer token file failed")

	// Test without bearer token file (should succeed)
	ipt.BearerTokenFile = ""
	err = ipt.setupAuth()
	assert.NoError(t, err)
}
