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
	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	iprom "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/prom"
)

var (
	defaultPrometheusioInterval         = time.Second * 60
	defaultPrometheusioConnectKeepAlive = time.Second * 20
	defaultStreamSize                   = 10
)

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

type promRunner struct {
	conf     *promConfig
	pm       *iprom.Prom
	feeder   io.Feeder
	lastTime time.Time
}

func newPromRunnerWithURLParams(source, measurementName, host, port, scheme, path string) (*promRunner, error) {
	u, err := getPromURL(host, port, scheme, path)
	if err != nil {
		return nil, err
	}

	return newPromRunner(source, measurementName, []string{u.String()}, "")
}

func newPromRunner(source, measurementName string, urls []string, interval string) (*promRunner, error) {
	c := &promConfig{
		Source:          source,
		URLs:            urls,
		MeasurementName: measurementName,
		Tags:            make(map[string]string),
	}

	if val, err := time.ParseDuration(interval); err != nil {
		c.Interval = defaultPrometheusioInterval
	} else {
		c.Interval = val
	}

	return newPromRunnerWithConfig(c)
}

type wrapPromConfig struct {
	Inputs struct {
		Prom []*promConfig `toml:"prom"`
	} `toml:"inputs"`
}

func newPromRunnerWithTomlConfig(str string) ([]*promRunner, error) {
	c := wrapPromConfig{}
	if err := bstoml.Unmarshal([]byte(str), &c); err != nil {
		return nil, fmt.Errorf("unable to parse toml: %w", err)
	}

	var res []*promRunner

	for _, promCfg := range c.Inputs.Prom {
		p, err := newPromRunnerWithConfig(promCfg)
		if err != nil {
			return nil, err
		}
		res = append(res, p)
	}

	return res, nil
}

func newPromRunnerWithConfig(c *promConfig) (*promRunner, error) {
	if c.Tags == nil {
		c.Tags = make(map[string]string)
	}
	if c.URL != "" {
		c.URLs = append(c.URLs, c.URL)
	}

	for _, u := range c.URLs {
		uu, err := url.Parse(u)
		if err != nil {
			return nil, fmt.Errorf("invalid url %s, err: %w", u, err)
		}
		if _, ok := c.Tags["instance"]; !ok {
			c.Tags["instance"] = uu.Host
		}
	}

	p := &promRunner{
		conf:     c,
		feeder:   io.DefaultFeeder(),
		lastTime: time.Now(),
	}

	callbackFunc := func(pts []*point.Point) error {
		if p.conf.AsLogging != nil && p.conf.AsLogging.Enable {
			for _, pt := range pts {
				err := p.feeder.Feed(string(pt.Name()), point.Logging, []*point.Point{pt},
					&io.Option{CollectCost: time.Since(p.lastTime)},
				)
				if err != nil {
					klog.Warnf("failed to feed prom logging: %s, ignored", err)
				}
			}
		} else {
			err := p.feeder.Feed(p.conf.Source, point.Metric, pts,
				&io.Option{CollectCost: time.Since(p.lastTime)},
			)
			if err != nil {
				klog.Warnf("failed to feed prom metrics: %s, ignored", err)
			}
		}
		return nil
	}

	opts := []iprom.PromOption{
		iprom.WithLogger(klog), // WithLogger must in the first
		iprom.WithSource(c.Source),
		iprom.WithTimeout(c.Timeout),
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
		iprom.WithMaxBatchCallback(defaultStreamSize, callbackFunc),
	}

	pm, err := iprom.NewProm(opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create prom: %w", err)
	}

	p.pm = pm
	return p, nil
}

func (p *promRunner) setCustomerTags(m map[string]string, keys []string) {
	if len(keys) == 0 || len(m) == 0 {
		return
	}
	for _, key := range keys {
		if v, ok := m[key]; ok {
			p.setTag(key, v)
		}
	}
}

func (p *promRunner) setTags(tags map[string]string) {
	for k, v := range tags {
		p.setTag(k, v)
	}
}

func (p *promRunner) setTag(k, v string) {
	if p.conf == nil {
		return
	}
	if p.conf.Tags == nil {
		p.conf.Tags = make(map[string]string)
	}
	if _, ok := p.conf.Tags[k]; ok {
		return
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

	klog.Debugf("running collect from source %s", p.conf.Source)
	p.lastTime = time.Now()

	for _, u := range p.conf.URLs {
		// use callback processor, not return pts
		_, err := p.pm.CollectFromHTTPV2(u)
		if err != nil {
			klog.Warnf("failed to collect prom: %s", err)
			return
		}
	}
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
