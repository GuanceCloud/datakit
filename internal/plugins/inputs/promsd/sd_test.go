// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package promsd

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/GuanceCloud/cliutils/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testTargetGroupsJSON = `[
  {
    "targets": ["127.0.0.1:19090"],
    "labels": {"context_path": "/ai-dm"}
  }
]`

func TestHTTPSDDiscoverTargetGroups(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(resp http.ResponseWriter, _ *http.Request) {
		resp.Header().Set("Content-Type", "application/json")
		if _, err := resp.Write([]byte(testTargetGroupsJSON)); err != nil {
			t.Errorf("write HTTP SD response: %s", err)
		}
	}))
	defer server.Close()

	sd := &HTTPSD{ServiceURL: server.URL, logger: logger.SLogger("promsd-test-http")}
	groups, err := sd.discoverTargetGroups()
	require.NoError(t, err)
	require.Len(t, groups, 1)
	assert.Equal(t, []string{"127.0.0.1:19090"}, groups[0].Targets)
	assert.Equal(t, "/ai-dm", groups[0].Labels["context_path"])
	assertDiscoveredTargetURL(t, relabelMetricsPathConfig(t, "context_path"), groups[0].Targets[0], groups[0].Labels,
		"http://127.0.0.1:19090/ai-dm/actuator/prometheus")
}

func TestFileSDReadTargetGroups(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "targets.json")
	require.NoError(t, os.WriteFile(path, []byte(testTargetGroupsJSON), 0o600))

	groups, err := readTargetGroups([]string{path})
	require.NoError(t, err)
	require.Len(t, groups, 1)
	assert.Equal(t, []string{"127.0.0.1:19090"}, groups[0].Targets)
	assert.Equal(t, "/ai-dm", groups[0].Labels["context_path"])
	assertDiscoveredTargetURL(t, relabelMetricsPathConfig(t, "context_path"), groups[0].Targets[0], groups[0].Labels,
		"http://127.0.0.1:19090/ai-dm/actuator/prometheus")
}

func relabelMetricsPathConfig(t *testing.T, sourceLabel string) *ScrapeConfig {
	t.Helper()

	cfg := &ScrapeConfig{
		Scheme:      "http",
		MetricsPath: "/actuator/prometheus",
		RelabelConfigs: []*RelabelConfig{
			{
				SourceLabels: []string{sourceLabel},
				Regex:        strptr("(.+)"),
				TargetLabel:  "__metrics_path__",
				Replacement:  strptr("$1/actuator/prometheus"),
			},
		},
	}
	require.NoError(t, cfg.setupRelabelConfigs())
	return cfg
}

func assertDiscoveredTargetURL(
	t *testing.T,
	cfg *ScrapeConfig,
	address string,
	labels map[string]string,
	wantURL string,
) {
	t.Helper()

	prepared, keep, err := prepareTarget(cfg, address, labels, mustParseParams(t, cfg.Params))
	require.NoError(t, err)
	require.True(t, keep)
	assert.Equal(t, wantURL, prepared.URL)
}
