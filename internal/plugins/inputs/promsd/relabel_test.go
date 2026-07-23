// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package promsd

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"testing"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/GuanceCloud/cliutils/logger"
	"github.com/GuanceCloud/cliutils/point"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/promscrape"
)

func TestRelabelConfigDecode(t *testing.T) {
	t.Parallel()

	var cfg struct {
		Inputs struct {
			PromSD []*Input `toml:"promsd"`
		} `toml:"inputs"`
	}

	_, err := toml.Decode(`
[[inputs.promsd]]
  source = "app-cn-dev-services"

  [inputs.promsd.scrape]
    scheme = "http"
    metrics_path = "/actuator/prometheus"
    interval = "30s"

  [[inputs.promsd.scrape.relabel_configs]]
    source_labels = ["context_path"]
    regex = "(.+)"
    target_label = "__metrics_path__"
    replacement = "$1/actuator/prometheus"
    action = "replace"
`, &cfg)
	require.NoError(t, err)
	require.Len(t, cfg.Inputs.PromSD, 1)
	require.NotNil(t, cfg.Inputs.PromSD[0].Scrape)
	require.Len(t, cfg.Inputs.PromSD[0].Scrape.RelabelConfigs, 1)

	rule := cfg.Inputs.PromSD[0].Scrape.RelabelConfigs[0]
	assert.Equal(t, []string{"context_path"}, rule.SourceLabels)
	assert.Equal(t, "(.+)", valueOrEmpty(rule.Regex))
	assert.Equal(t, "__metrics_path__", rule.TargetLabel)
	assert.Equal(t, "$1/actuator/prometheus", valueOrEmpty(rule.Replacement))
	assert.Equal(t, "replace", valueOrEmpty(rule.Action))
}

func TestPrepareTargetCustomerContextPath(t *testing.T) {
	t.Parallel()

	cfg := &ScrapeConfig{
		Scheme:      "http",
		MetricsPath: "/actuator/prometheus",
		RelabelConfigs: []*RelabelConfig{
			{
				SourceLabels: []string{"context_path"},
				Regex:        strptr("(.+)"),
				TargetLabel:  "__metrics_path__",
				Replacement:  strptr("$1/actuator/prometheus"),
				Action:       strptr("replace"),
			},
			{
				SourceLabels: []string{"__metrics_path__"},
				Regex:        strptr("^/*(.+)"),
				TargetLabel:  "__metrics_path__",
				Replacement:  strptr("/$1"),
				Action:       strptr("replace"),
			},
			{
				SourceLabels: []string{"context_path"},
				TargetLabel:  "app_context_path",
				Action:       strptr("replace"),
			},
		},
	}
	require.NoError(t, cfg.setupRelabelConfigs())

	tests := []struct {
		name       string
		address    string
		labels     map[string]string
		wantURL    string
		wantAppCtx string
	}{
		{
			name:    "default metrics path",
			address: "10.63.54.187:8201",
			labels: map[string]string{
				"management_endpoints_web_base_path": "/actuator",
			},
			wantURL: "http://10.63.54.187:8201/actuator/prometheus",
		},
		{
			name:    "context path from service discovery",
			address: "10.63.36.176:8413",
			labels: map[string]string{
				"context_path": "/ai-dm",
			},
			wantURL:    "http://10.63.36.176:8413/ai-dm/actuator/prometheus",
			wantAppCtx: "/ai-dm",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, keep, err := prepareTarget(cfg, tc.address, tc.labels, mustParseParams(t, cfg.Params))
			require.NoError(t, err)
			require.True(t, keep)
			assert.Equal(t, tc.wantURL, got.URL)
			assert.Equal(t, tc.wantAppCtx, got.Tags["app_context_path"])
			for key := range got.Tags {
				assert.NotRegexp(t, `^__`, key)
			}
		})
	}
}

func TestPrepareTargetKeepDrop(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		action string
		value  string
		keep   bool
	}{
		{name: "keep matching target", action: "keep", value: "prod", keep: true},
		{name: "drop non-matching keep target", action: "keep", value: "test", keep: false},
		{name: "drop matching target", action: "drop", value: "prod", keep: false},
		{name: "keep non-matching drop target", action: "drop", value: "test", keep: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			cfg := &ScrapeConfig{
				Scheme:      "http",
				MetricsPath: "/metrics",
				RelabelConfigs: []*RelabelConfig{
					{
						SourceLabels: []string{"env"},
						Regex:        strptr("prod"),
						Action:       strptr(tc.action),
					},
				},
			}
			require.NoError(t, cfg.setupRelabelConfigs())

			_, keep, err := prepareTarget(cfg, "127.0.0.1:9100", map[string]string{"env": tc.value}, nil)
			require.NoError(t, err)
			assert.Equal(t, tc.keep, keep)
		})
	}
}

func TestPrepareTargetRewritesAddressAndParams(t *testing.T) {
	t.Parallel()

	cfg := &ScrapeConfig{
		Scheme:      "http",
		MetricsPath: "/metrics",
		Params:      "module=first&module=second&debug=true",
		RelabelConfigs: []*RelabelConfig{
			{
				SourceLabels: []string{"__address__"},
				Regex:        strptr("(.+):8080"),
				TargetLabel:  "__address__",
				Replacement:  strptr("$1:9100"),
			},
			{
				Replacement: strptr("false"),
				TargetLabel: "__param_debug",
			},
		},
	}
	require.NoError(t, cfg.setupRelabelConfigs())

	target, keep, err := prepareTarget(cfg, "127.0.0.1:8080", nil, mustParseParams(t, cfg.Params))
	require.NoError(t, err)
	require.True(t, keep)
	assert.Equal(t, "http://127.0.0.1:9100/metrics?debug=false&module=first&module=second", target.URL)
}

func TestPrepareTargetPreservesURLParameterNames(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		params     string
		wantValues []string
	}{
		{
			name:       "single match parameter",
			params:     "match[]=up",
			wantValues: []string{"up"},
		},
		{
			name:       "repeated match parameter",
			params:     `match[]=up&match[]=%7Bjob%3D%22api%22%7D`,
			wantValues: []string{"up", `{job="api"}`},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			cfg := &ScrapeConfig{
				Scheme:      "http",
				MetricsPath: "/federate",
				Params:      tc.params,
				RelabelConfigs: []*RelabelConfig{
					{
						SourceLabels: []string{"__address__"},
						Regex:        strptr(".+"),
						Action:       strptr("keep"),
					},
				},
			}
			require.NoError(t, cfg.setupRelabelConfigs())

			target, keep, err := prepareTarget(cfg, "127.0.0.1:9090", nil, mustParseParams(t, cfg.Params))
			require.NoError(t, err)
			require.True(t, keep)

			targetURL, err := url.Parse(target.URL)
			require.NoError(t, err)
			assert.Equal(t, "/federate", targetURL.Path)
			assert.Equal(t, tc.wantValues, targetURL.Query()["match[]"])
		})
	}
}

func TestPrepareTargetUsesRelabeledInternalLabels(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		rule    *RelabelConfig
		wantURL string
	}{
		{
			name: "deleted metrics path is not restored",
			rule: &RelabelConfig{
				SourceLabels: []string{"__metrics_path__"},
				TargetLabel:  "__metrics_path__",
				Replacement:  strptr(""),
				Action:       strptr("replace"),
			},
			wantURL: "http://127.0.0.1:9100",
		},
		{
			name: "non-http scheme is not replaced by default",
			rule: &RelabelConfig{
				TargetLabel: "__scheme__",
				Replacement: strptr("ftp"),
				Action:      strptr("replace"),
			},
			wantURL: "ftp://127.0.0.1:9100/metrics",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			cfg := &ScrapeConfig{
				Scheme:         "http",
				MetricsPath:    "/metrics",
				RelabelConfigs: []*RelabelConfig{tc.rule},
			}
			require.NoError(t, cfg.setupRelabelConfigs())

			target, keep, err := prepareTarget(cfg, "127.0.0.1:9100", nil, nil)
			require.NoError(t, err)
			require.True(t, keep)
			assert.Equal(t, tc.wantURL, target.URL)
		})
	}
}

func TestBuildScrapersWithoutRelabelPreservesLegacyBehavior(t *testing.T) {
	t.Parallel()

	type observedRequest struct {
		path  string
		query url.Values
	}
	requestCh := make(chan observedRequest, 1)
	server := httptest.NewServer(http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		requestCh <- observedRequest{path: req.URL.Path, query: req.URL.Query()}
		if _, err := io.WriteString(resp, "legacy_metric 1\n"); err != nil {
			t.Errorf("write metrics response: %s", err)
		}
	}))
	defer server.Close()

	serverURL, err := url.Parse(server.URL)
	require.NoError(t, err)

	var points []*point.Point
	opts := []promscrape.Option{
		promscrape.WithCallback(func(got []*point.Point) error {
			points = got
			return nil
		}),
	}
	cfg := &ScrapeConfig{
		Scheme:      "https",
		MetricsPath: "/configured",
		Params:      "module=from-config-1&module=from-config-2&debug=true",
	}
	group := TargetGroup{
		Targets: []string{serverURL.Host},
		Labels: map[string]string{
			"__meta_clusterName": "DEFAULT",
			"__metrics_path__":   "/legacy",
			"__param_module":     "from-label",
			"__scheme__":         serverURL.Scheme,
			"app":                "legacy-app",
			"instance":           "from-discovery",
		},
	}

	scrapers, err := buildScrapersFromGroup(cfg, opts, group, nil)
	require.NoError(t, err)
	require.Len(t, scrapers, 1)
	assert.Equal(t,
		server.URL+"/legacy?debug=true&module=from-label&module=from-config-1&module=from-config-2",
		scrapers[0].targetURL())

	require.NoError(t, scrapers[0].scrape(time.Now().UnixNano()))
	request := <-requestCh
	assert.Equal(t, "/legacy", request.path)
	assert.Equal(t, []string{"from-label", "from-config-1", "from-config-2"}, request.query["module"])
	assert.Equal(t, "true", request.query.Get("debug"))

	require.Len(t, points, 1)
	assert.Equal(t, "DEFAULT", points[0].GetTag("__meta_clusterName"))
	assert.Equal(t, "legacy-app", points[0].GetTag("app"))
	assert.Equal(t, serverURL.Host, points[0].GetTag("instance"))
}

func TestBuildRelabeledScrapersSkipsInvalidTarget(t *testing.T) {
	t.Parallel()

	cfg := &ScrapeConfig{
		Scheme:      "http",
		MetricsPath: "/metrics",
		RelabelConfigs: []*RelabelConfig{
			{
				SourceLabels: []string{"__address__"},
				Regex:        strptr("drop-me"),
				TargetLabel:  "__address__",
				Replacement:  strptr(""),
				Action:       strptr("replace"),
			},
		},
	}
	require.NoError(t, cfg.setupRelabelConfigs())

	scrapers, err := buildRelabeledScrapersFromGroup(
		cfg,
		nil,
		TargetGroup{Targets: []string{"drop-me", "127.0.0.1:9090"}},
		logger.SLogger("promsd-relabel-test"),
	)
	require.NoError(t, err)
	require.Len(t, scrapers, 1)
	assert.Equal(t, "http://127.0.0.1:9090/metrics", scrapers[0].targetURL())
}

func TestBuildRelabeledScrapersReturnsSharedScraperError(t *testing.T) {
	t.Parallel()

	cfg := &ScrapeConfig{
		Scheme:      "https",
		MetricsPath: "/metrics",
		RelabelConfigs: []*RelabelConfig{
			{
				SourceLabels: []string{"__address__"},
				Regex:        strptr(".+"),
				Action:       strptr("keep"),
			},
		},
	}
	require.NoError(t, cfg.setupRelabelConfigs())

	opts := []promscrape.Option{
		promscrape.WithTLSOpen(true),
		promscrape.WithCacertFiles([]string{filepath.Join(t.TempDir(), "missing-ca.pem")}),
	}
	scrapers, err := buildRelabeledScrapersFromGroup(
		cfg,
		opts,
		TargetGroup{Targets: []string{"127.0.0.1:9090"}},
		logger.SLogger("promsd-relabel-test"),
	)
	require.ErrorContains(t, err, "create scraper for target")
	assert.Nil(t, scrapers)
}

func TestRelabelConfigValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		rule    *RelabelConfig
		wantErr string
	}{
		{
			name:    "empty rule",
			wantErr: "rule is empty",
		},
		{
			name:    "invalid regex",
			rule:    &RelabelConfig{Regex: strptr("["), Action: strptr("drop")},
			wantErr: "invalid regex",
		},
		{
			name:    "replace without target label",
			rule:    &RelabelConfig{Action: strptr("replace")},
			wantErr: "requires target_label",
		},
		{
			name:    "hashmod without modulus",
			rule:    &RelabelConfig{TargetLabel: "shard", Action: strptr("hashmod")},
			wantErr: "requires a non-zero modulus",
		},
		{
			name:    "unknown action",
			rule:    &RelabelConfig{Action: strptr("unknown")},
			wantErr: "unknown action",
		},
		{
			name:    "invalid source label",
			rule:    &RelabelConfig{SourceLabels: []string{"invalid-label"}, Action: strptr("drop")},
			wantErr: "invalid source_label",
		},
		{
			name:    "invalid replace target label",
			rule:    &RelabelConfig{TargetLabel: "invalid-label", Action: strptr("replace")},
			wantErr: "invalid target_label",
		},
		{
			name:    "invalid hashmod target label",
			rule:    &RelabelConfig{Modulus: 2, TargetLabel: "invalid-label", Action: strptr("hashmod")},
			wantErr: "invalid target_label",
		},
		{
			name: "labeldrop with source labels",
			rule: &RelabelConfig{
				SourceLabels: []string{"env"}, Regex: strptr("debug"), Action: strptr("labeldrop"),
			},
			wantErr: "only supports regex",
		},
		{
			name: "labelkeep with target label",
			rule: &RelabelConfig{
				Regex: strptr("env"), TargetLabel: "job", Action: strptr("labelkeep"),
			},
			wantErr: "only supports regex",
		},
		{
			name: "keepequal with modulus",
			rule: &RelabelConfig{
				SourceLabels: []string{"service"}, Modulus: 2, TargetLabel: "job", Action: strptr("keepequal"),
			},
			wantErr: "only supports source_labels and target_label",
		},
		{
			name: "dropequal with separator",
			rule: &RelabelConfig{
				SourceLabels: []string{"service"}, Separator: strptr(","), TargetLabel: "job", Action: strptr("dropequal"),
			},
			wantErr: "only supports source_labels and target_label",
		},
		{
			name: "lowercase with replacement",
			rule: &RelabelConfig{
				SourceLabels: []string{"service"}, TargetLabel: "job", Replacement: strptr("prefix-$1"), Action: strptr("lowercase"),
			},
			wantErr: "replacement cannot be set",
		},
		{
			name: "uppercase with replacement",
			rule: &RelabelConfig{
				SourceLabels: []string{"service"}, TargetLabel: "job", Replacement: strptr("prefix-$1"), Action: strptr("uppercase"),
			},
			wantErr: "replacement cannot be set",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			cfg := &ScrapeConfig{RelabelConfigs: []*RelabelConfig{tc.rule}}
			require.ErrorContains(t, cfg.setupRelabelConfigs(), tc.wantErr)
		})
	}
}

func TestRelabelConfigPrometheusCompatibility(t *testing.T) {
	t.Parallel()

	for _, action := range []string{"keep", "drop"} {
		t.Run(action+" ignores unused target_label", func(t *testing.T) {
			t.Parallel()

			_, err := compileRelabelConfig(&RelabelConfig{
				Action:      strptr(action),
				TargetLabel: "invalid-label",
			})
			require.NoError(t, err)
		})
	}

	for _, action := range []string{"keepequal", "dropequal"} {
		t.Run(action+" rejects explicitly configured regex", func(t *testing.T) {
			t.Parallel()

			_, err := compileRelabelConfig(&RelabelConfig{
				SourceLabels: []string{"service"},
				Regex:        strptr(defaultRelabelRegex),
				TargetLabel:  "job",
				Action:       strptr(action),
			})
			require.ErrorContains(t, err, "only supports source_labels and target_label")
		})
	}
}

func TestRelabelActions(t *testing.T) { //nolint:funlen
	t.Parallel()

	tests := []struct {
		name       string
		labels     map[string]string
		rule       *RelabelConfig
		wantKeep   bool
		wantLabels map[string]string
	}{
		{
			name:   "replace",
			labels: map[string]string{"service": "API"},
			rule: &RelabelConfig{
				SourceLabels: []string{"service"}, TargetLabel: "job", Action: strptr("replace"),
			},
			wantKeep:   true,
			wantLabels: map[string]string{"service": "API", "job": "API"},
		},
		{
			name:   "keepequal",
			labels: map[string]string{"service": "api", "job": "api"},
			rule: &RelabelConfig{
				SourceLabels: []string{"service"}, TargetLabel: "job", Action: strptr("keepequal"),
			},
			wantKeep:   true,
			wantLabels: map[string]string{"service": "api", "job": "api"},
		},
		{
			name:   "dropequal",
			labels: map[string]string{"service": "api", "job": "api"},
			rule: &RelabelConfig{
				SourceLabels: []string{"service"}, TargetLabel: "job", Action: strptr("dropequal"),
			},
			wantKeep:   false,
			wantLabels: map[string]string{"service": "api", "job": "api"},
		},
		{
			name:   "lowercase",
			labels: map[string]string{"service": "API"},
			rule: &RelabelConfig{
				SourceLabels: []string{"service"}, TargetLabel: "job", Action: strptr("lowercase"),
			},
			wantKeep:   true,
			wantLabels: map[string]string{"service": "API", "job": "api"},
		},
		{
			name:   "uppercase",
			labels: map[string]string{"service": "api"},
			rule: &RelabelConfig{
				SourceLabels: []string{"service"}, TargetLabel: "job", Action: strptr("uppercase"),
			},
			wantKeep:   true,
			wantLabels: map[string]string{"service": "api", "job": "API"},
		},
		{
			name:   "hashmod",
			labels: map[string]string{"instance": "10.0.0.1:9100"},
			rule: &RelabelConfig{
				SourceLabels: []string{"instance"}, TargetLabel: "shard", Modulus: 2, Action: strptr("hashmod"),
			},
			wantKeep:   true,
			wantLabels: map[string]string{"instance": "10.0.0.1:9100", "shard": "1"},
		},
		{
			name:   "labelmap",
			labels: map[string]string{"__meta_service_env": "prod"},
			rule: &RelabelConfig{
				Regex: strptr("__meta_service_(.+)"), Replacement: strptr("$1"), Action: strptr("labelmap"),
			},
			wantKeep: true,
			wantLabels: map[string]string{
				"__meta_service_env": "prod", "env": "prod",
			},
		},
		{
			name:   "labeldrop",
			labels: map[string]string{"env": "prod", "debug": "true"},
			rule: &RelabelConfig{
				Regex: strptr("debug"), Action: strptr("labeldrop"),
			},
			wantKeep:   true,
			wantLabels: map[string]string{"env": "prod"},
		},
		{
			name:   "labelkeep",
			labels: map[string]string{"env": "prod", "debug": "true"},
			rule: &RelabelConfig{
				Regex: strptr("env"), Action: strptr("labelkeep"),
			},
			wantKeep:   true,
			wantLabels: map[string]string{"env": "prod"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			compiled, err := compileRelabelConfig(tc.rule)
			require.NoError(t, err)
			assert.Equal(t, tc.wantKeep, applyRelabelConfigs(tc.labels, []*compiledRelabelConfig{compiled}))
			assert.Equal(t, tc.wantLabels, tc.labels)
		})
	}
}

func TestLabelMapUsesStableOrder(t *testing.T) {
	t.Parallel()

	compiled, err := compileRelabelConfig(&RelabelConfig{
		Regex:       strptr(`source_.+`),
		Replacement: strptr("target"),
		Action:      strptr("labelmap"),
	})
	require.NoError(t, err)

	for i := 0; i < 100; i++ {
		labels := map[string]string{
			"source_a": "first",
			"source_b": "second",
		}
		require.True(t, applyRelabelConfig(labels, compiled))
		assert.Equal(t, "second", labels["target"])
	}
}

func strptr(value string) *string { return &value }

func valueOrEmpty(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func mustParseParams(t *testing.T, raw string) url.Values {
	t.Helper()

	params, err := url.ParseQuery(raw)
	require.NoError(t, err)
	return params
}
