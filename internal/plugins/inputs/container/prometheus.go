// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package container

import (
	"fmt"
	"net/url"
	"strconv"
	"time"

	bstoml "github.com/BurntSushi/toml"
	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	iprom "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/prom"
	"golang.org/x/exp/slices"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	defaultPrometheusioInterval         = time.Second * 60
	defaultPrometheusioConnectKeepAlive = time.Second * 20
)

type promConfig struct {
	Source   string        `toml:"source" json:"source"`
	Interval time.Duration `toml:"interval"`
	URLs     []string      `toml:"urls" json:"urls"`

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

type promRunner struct {
	conf     *promConfig
	pm       *iprom.Prom
	feeder   io.Feeder
	lastTime time.Time
}

func newPromRunnerWithURLParams(source, host, port, scheme, path string, tags map[string]string) (*promRunner, error) {
	u, err := getPromURL(host, port, scheme, path)
	if err != nil {
		return nil, err
	}

	return newPromRunner(source, []string{u.String()}, "", tags)
}

func newPromRunner(source string, urls []string, interval string, tags map[string]string) (*promRunner, error) {
	tagsTemp := make(map[string]string, len(tags))
	for k, v := range tags {
		tagsTemp[k] = v
	}

	c := &promConfig{
		Source: source,
		URLs:   urls,
		Tags:   tagsTemp,
	}

	if val, err := time.ParseDuration(interval); err != nil {
		c.Interval = defaultPrometheusioInterval
	} else {
		c.Interval = val
	}

	return newPromRunnerWithConfig(c)
}

func newPromRunnerWithTomlConfig(str string) (*promRunner, error) {
	c := promConfig{}
	if err := bstoml.Unmarshal([]byte(str), &c); err != nil {
		return nil, fmt.Errorf("unable to parse toml: %w", err)
	}
	return newPromRunnerWithConfig(&c)
}

func newPromRunnerWithConfig(c *promConfig) (*promRunner, error) {
	opts := []iprom.PromOption{
		iprom.WithLogger(l), // WithLogger must in the first
		iprom.WithSource(c.Source),
		iprom.WithKeepAlive(defaultPrometheusioConnectKeepAlive),
		iprom.WithIgnoreReqErr(c.IgnoreReqErr),
		iprom.WithMetricTypes(c.MetricTypes),
		iprom.WithMetricNameFilter(c.MetricNameFilter),
		iprom.WithMetricNameFilterIgnore(c.MetricNameFilterIgnore),
		iprom.WithMeasurementPrefix(c.MeasurementPrefix),
		iprom.WithMeasurementName(c.MeasurementName),
		iprom.WithMeasurements(c.Measurements),
		iprom.WithTLSOpen(c.TLSOpen),
		iprom.WithUDSPath(c.UDSPath),
		iprom.WithCacertFile(c.CacertFile),
		iprom.WithCertFile(c.CertFile),
		iprom.WithKeyFile(c.KeyFile),
		iprom.WithTagsIgnore(c.TagsIgnore),
		iprom.WithTagsRename(c.TagsRename),
		iprom.WithAsLogging(c.AsLogging),
		iprom.WithIgnoreTagKV(c.IgnoreTagKV),
		iprom.WithHTTPHeaders(c.HTTPHeaders),
		iprom.WithTags(c.Tags),
		iprom.WithDisableInfoTag(c.DisableInfoTag),
		iprom.WithAuth(c.Auth),
	}

	pm, err := iprom.NewProm(opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create prom: %w", err)
	}

	return &promRunner{
		conf:     c,
		pm:       pm,
		feeder:   io.DefaultFeeder(),
		lastTime: time.Now(),
	}, nil
}

func (p *promRunner) addTags(tags map[string]string) {
	for k, v := range tags {
		p.addSingleTag(k, v)
	}
}

func (p *promRunner) addSingleTag(k, v string) {
	if p.conf == nil {
		return
	}
	if p.conf.Tags == nil {
		p.conf.Tags = make(map[string]string)
	}

	p.conf.Tags[k] = v
}

func (p *promRunner) runOnce() {
	if p.conf == nil {
		return
	}
	if time.Since(p.lastTime) < p.conf.Interval {
		return
	}

	l.Debugf("running collect from source %s", p.conf.Source)
	p.lastTime = time.Now()

	start := time.Now()

	pts, err := p.collect()
	if err != nil {
		l.Warnf("failed to collect prom: %w", err)
		return
	}
	if len(pts) == 0 {
		l.Warnf("points got nil from collect")
		return
	}

	if p.conf.AsLogging != nil && p.conf.AsLogging.Enable {
		// Feed measurement as logging.
		for _, pt := range pts {
			// We need to feed each point separately because
			// each point might have different measurement name.
			err := p.feeder.Feed(
				string(pt.Name()),
				point.Logging,
				[]*point.Point{pt},
				&io.Option{CollectCost: time.Since(start)},
			)
			if err != nil {
				l.Warnf("failed to feed prom logging: %s, ignored", err)
			}
		}
	} else {
		err := p.feeder.Feed(
			p.conf.Source,
			point.Metric,
			pts,
			&io.Option{CollectCost: time.Since(start)},
		)
		if err != nil {
			l.Warnf("failed to feed prom metrics: %s, ignored", err)
		}
	}
}

func (p *promRunner) collect() ([]*point.Point, error) {
	if p.pm == nil {
		return nil, fmt.Errorf("unreachable, invalid prom")
	}

	var points []*point.Point

	for _, u := range p.conf.URLs {
		pts, err := p.pm.CollectFromHTTPV2(u)
		if err != nil {
			return nil, err
		}
		points = append(points, pts...)
	}

	return points, nil
}

func getPromURL(host, port, scheme, path string) (*url.URL, error) {
	if _, err := strconv.Atoi(port); err != nil {
		return nil, fmt.Errorf("invalid port %s", port)
	}
	u := &url.URL{
		Scheme: defaultPromScheme,
		Path:   defaultPromPath,
		Host:   fmt.Sprintf("%s:%s", host, port),
	}
	if scheme == "https" {
		u.Scheme = scheme
	}
	if path != "" {
		u.Path = path
	}
	return u, nil
}

type prometheusMonitoringExtraConfig struct {
	Matches []struct {
		NamespaceSelector struct {
			Any             bool     `json:"any,omitempty"`
			MatchNamespaces []string `json:"matchNamespaces,omitempty"`
		} `json:"namespaceSelector,omitempty"`

		Selector metav1.LabelSelector `json:"selector"`

		PromConfig *promConfig `json:"promConfig"`
	} `json:"matches"`
}

func (p *prometheusMonitoringExtraConfig) matchPromConfig(targetLabels map[string]string, namespace string) *promConfig {
	if len(p.Matches) == 0 {
		return nil
	}

	for _, match := range p.Matches {
		if !match.NamespaceSelector.Any {
			if len(match.NamespaceSelector.MatchNamespaces) != 0 &&
				slices.Index(match.NamespaceSelector.MatchNamespaces, namespace) == -1 {
				continue
			}
		}
		if !newLabelSelector(match.Selector.MatchLabels, match.Selector.MatchExpressions).Matches(targetLabels) {
			continue
		}
		return match.PromConfig
	}

	return nil
}

func mergePromConfig(c1 *promConfig, c2 *promConfig) *promConfig {
	c3 := &promConfig{
		Source:   c1.Source,
		Interval: c1.Interval,
		URLs:     c1.URLs,
		Tags:     c1.Tags,
	}

	c3.IgnoreReqErr = c2.IgnoreReqErr
	c3.MetricTypes = c2.MetricTypes
	c3.MetricNameFilter = c2.MetricNameFilter
	c3.MetricNameFilterIgnore = c2.MetricNameFilterIgnore
	c3.MeasurementPrefix = c2.MeasurementPrefix
	c3.MeasurementName = c2.MeasurementName
	c3.Measurements = c2.Measurements

	c3.TLSOpen = c2.TLSOpen
	c3.UDSPath = c2.UDSPath
	c3.CacertFile = c2.CacertFile
	c3.CertFile = c2.CertFile
	c3.KeyFile = c2.KeyFile

	c3.TagsIgnore = c2.TagsIgnore
	c3.TagsRename = c2.TagsRename
	c3.AsLogging = c2.AsLogging
	c3.IgnoreTagKV = c2.IgnoreTagKV

	c3.HTTPHeaders = c2.HTTPHeaders
	c3.DisableInfoTag = c2.DisableInfoTag

	for k, v := range c2.Tags {
		if _, ok := c3.Tags[k]; !ok {
			c3.Tags[k] = v
		}
	}

	return c3
}
