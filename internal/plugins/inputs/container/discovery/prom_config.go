// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package discovery

import (
	"fmt"
	"net/url"
	"strconv"
	"time"

	bstoml "github.com/BurntSushi/toml"
	iprom "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/prom"
)

var defaultPrometheusioInterval = time.Second * 60

type promConfig struct {
	Source   string        `toml:"source" json:"source"`
	Interval time.Duration `toml:"interval"`
	Timeout  time.Duration `toml:"timeout"`

	URL  string   `toml:"url" json:"url"` // deprecated
	URLs []string `toml:"urls" json:"urls"`

	IgnoreReqErr           bool         `toml:"ignore_req_err" json:"ignore_req_err"`
	MetricTypes            []string     `toml:"metric_types" json:"metric_types"`
	MetricNameFilter       []string     `toml:"metric_name_filter" json:"metric_name_filter"`
	MetricNameFilterIgnore []string     `toml:"metric_name_filter_ignore" json:"metric_name_filter_ignore"`
	MeasurementPrefix      string       `toml:"measurement_prefix" json:"measurement_prefix"`
	MeasurementName        string       `toml:"measurement_name" json:"measurement_name"`
	Measurements           []iprom.Rule `toml:"measurements" json:"measurements"`

	TLSOpen    bool   `toml:"tls_open" json:"tls_open"`
	UDSPath    string `toml:"uds_path" json:"uds_path"`
	CacertFile string `toml:"tls_ca" json:"tls_ca"`
	CertFile   string `toml:"tls_cert" json:"tls_cert"`
	KeyFile    string `toml:"tls_key" json:"tls_key"`

	TagsIgnore  []string            `toml:"tags_ignore" json:"tags_ignore"`
	TagsRename  *iprom.RenameTags   `toml:"tags_rename" json:"tags_rename"`
	AsLogging   *iprom.AsLogging    `toml:"as_logging" json:"as_logging"`
	IgnoreTagKV map[string][]string `toml:"ignore_tag_kv_match" json:"ignore_tag_kv_match"`
	HTTPHeaders map[string]string   `toml:"http_headers" json:"http_headers"`

	Tags           map[string]string
	DisableInfoTag bool `toml:"disable_info_tag" json:"disable_info_tag"`

	Auth map[string]string `toml:"auth" json:"auth"`
}

type promOption func(c *promConfig)

func newPromConfig(opts ...promOption) *promConfig {
	c := &promConfig{Interval: defaultPrometheusioInterval}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

func withSource(source string) promOption { return func(c *promConfig) { c.Source = source } }
func withURLs(urls []string) promOption {
	return func(c *promConfig) { c.URLs = append(c.URLs, urls...) }
}

func withMeasurementName(measurementName string) promOption {
	return func(c *promConfig) { c.MeasurementName = measurementName }
}

func withInterval(interval string) promOption {
	return func(c *promConfig) {
		if val, err := time.ParseDuration(interval); err == nil {
			c.Interval = val
		}
	}
}

func withTagIfNotEmpty(key, value string) promOption {
	return func(c *promConfig) {
		if key == "" {
			return
		}
		withTag(key, value)(c)
	}
}

func withTag(key, value string) promOption {
	return func(c *promConfig) {
		if c.Tags == nil {
			c.Tags = make(map[string]string)
		}
		if _, ok := c.Tags[key]; !ok {
			c.Tags[key] = value
		}
	}
}

func withTags(tags map[string]string) promOption {
	return func(c *promConfig) {
		for k, v := range tags {
			withTag(k, v)(c)
		}
	}
}

func withCustomerTags(m map[string]string, keys []string) promOption {
	return func(c *promConfig) {
		if len(keys) == 0 || len(m) == 0 {
			return
		}
		for _, key := range keys {
			if v, ok := m[key]; ok {
				withTag(key, v)(c)
			}
		}
	}
}

type wrapPromConfig struct {
	Inputs struct {
		Prom []*promConfig `toml:"prom"`
	} `toml:"inputs"`
}

func parseURLHost(cfg *promConfig) (map[string]string, error) {
	res := make(map[string]string)
	for _, urlstr := range cfg.URLs {
		u, err := url.Parse(urlstr)
		if err != nil {
			return nil, fmt.Errorf("invalid url %s, err: %w", urlstr, err)
		}
		res[urlstr] = u.Host
	}
	return res, nil
}

func parsePromConfigs(str string) ([]*promConfig, error) {
	c := wrapPromConfig{}
	if err := bstoml.Unmarshal([]byte(str), &c); err != nil {
		return nil, fmt.Errorf("unable to parse toml: %w", err)
	}
	for _, cfg := range c.Inputs.Prom {
		if cfg.URL != "" {
			cfg.URLs = append(cfg.URLs, cfg.URL)
		}
	}
	return c.Inputs.Prom, nil
}

func joinPromURL(host, port, scheme, path, rawQuery string) (string, error) {
	if _, err := strconv.Atoi(port); err != nil {
		return "", fmt.Errorf("invalid port %s", port)
	}
	u := &url.URL{
		Scheme:   defaultPromScheme,
		Path:     defaultPromPath,
		Host:     fmt.Sprintf("%s:%s", host, port),
		RawQuery: rawQuery,
	}
	if scheme == "https" {
		u.Scheme = scheme
	}
	if path != "" {
		u.Path = path
	}
	return u.String(), nil
}
