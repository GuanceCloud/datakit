// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package prom used to parsing promemetheuse exportor metrics.
package prom

import (
	"context"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"github.com/prometheus/common/expfmt"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	dnet "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/net"
	inpt "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/point"
	dkpt "gitlab.jiagouyun.com/cloudcare-tools/datakit/io/point"
)

type Rule struct {
	Pattern string `toml:"pattern" json:"pattern"`
	Prefix  string `toml:"prefix" json:"prefix"`
	Name    string `toml:"name" json:"name"`
}

type RenameTags struct {
	OverwriteExistTags bool              `toml:"overwrite_exist_tags" json:"overwrite_exist_tags"`
	Mapping            map[string]string `toml:"mapping" json:"mapping"`
}

type AsLogging struct {
	Enable  bool   `toml:"enable" json:"enable"`
	Service string `toml:"service" json:"service"`
}

type IgnoreTagKeyValMatch map[string][]*regexp.Regexp

type Option struct {
	MetricTypes      []string `toml:"metric_types"`
	MetricNameFilter []string `toml:"metric_name_filter"`
	Measurements     []Rule   `json:"measurements"`
	Source           string   `toml:"source"`
	Interval         string   `toml:"interval"`
	Timeout          string   `toml:"timeout"`

	URL  string   `toml:"url,omitempty"` // Deprecated
	URLs []string `toml:"urls"`

	IgnoreReqErr bool `toml:"ignore_req_err"`

	Output string `toml:"output"`

	MaxFileSize int64 `toml:"max_file_size"`

	MeasurementPrefix string `toml:"measurement_prefix"`
	MeasurementName   string `toml:"measurement_name"`

	CacertFile string `toml:"tls_ca"`
	CertFile   string `toml:"tls_cert"`
	KeyFile    string `toml:"tls_key"`

	Auth        map[string]string `toml:"auth"`
	HTTPHeaders map[string]string `toml:"http_headers"`
	interval    time.Duration

	Tags       map[string]string `toml:"tags"`
	RenameTags *RenameTags       `toml:"rename_tags"`
	AsLogging  *AsLogging        `toml:"as_logging"`

	// do not keep these tags in scraped prom data
	TagsIgnore []string `toml:"tags_ignore"`

	// drop scraped prom data if tag key's value matched
	IgnoreTagKV IgnoreTagKeyValMatch

	DisableHostTag     bool
	DisableInstanceTag bool

	Election bool
	pointOpt *dkpt.PointOption

	TLSOpen bool   `toml:"tls_open"`
	UDSPath string `toml:"uds_path"`
	Disable bool   `toml:"disble"`
}

const defaultInterval = 30 * time.Second

func (opt *Option) IsDisable() bool {
	return opt.Disable
}

func (opt *Option) GetSource(defaultSource ...string) string {
	if opt.Source != "" {
		return opt.Source
	}
	if len(defaultSource) > 0 {
		return defaultSource[0]
	}
	return "prom" //nolint:goconst
}

func (opt *Option) GetIntervalDuration() time.Duration {
	if opt.interval > 0 {
		return opt.interval
	}

	t, err := time.ParseDuration(opt.Interval)
	if err != nil {
		t = defaultInterval
	}

	opt.interval = t
	return t
}

const (
	httpTimeout               = time.Second * 3
	defaultInsecureSkipVerify = false
)

type Prom struct {
	opt    *Option
	client *http.Client
	parser expfmt.TextParser
}

func NewProm(opt *Option) (*Prom, error) {
	if opt == nil {
		return nil, fmt.Errorf("invalid option")
	}

	if opt.URL == "" && len(opt.URLs) == 0 {
		return nil, fmt.Errorf("invalid URL, cannot be empty")
	}

	// double check opt.URL is placed in opt.URLs
	if opt.URL != "" {
		placed := false
		for _, u := range opt.URLs {
			if u == opt.URL {
				placed = true
				break
			}
		}
		if !placed {
			opt.URLs = append(opt.URLs, opt.URL)
		}
	}

	timeout, err := time.ParseDuration(opt.Timeout)
	if err != nil || timeout < httpTimeout {
		timeout = httpTimeout
	}

	p := Prom{opt: opt}

	var dialContext func(_ context.Context, _ string, _ string) (net.Conn, error)
	if p.opt.UDSPath != "" {
		dialContext = func(_ context.Context, _, _ string) (net.Conn, error) {
			return net.Dial("unix", p.opt.UDSPath)
		}
	}
	p.SetClient(&http.Client{Timeout: timeout, Transport: &http.Transport{
		DialContext: dialContext,
	}})

	if opt.TLSOpen {
		caCerts := []string{}
		insecureSkipVerify := defaultInsecureSkipVerify
		if len(opt.CacertFile) != 0 {
			caCerts = append(caCerts, opt.CacertFile)
		} else {
			insecureSkipVerify = true
		}
		tc := &dnet.TLSClientConfig{
			CaCerts:            caCerts,
			Cert:               opt.CertFile,
			CertKey:            opt.KeyFile,
			InsecureSkipVerify: insecureSkipVerify,
		}

		tlsconfig, err := tc.TLSConfig()
		if err != nil {
			return nil, err
		}
		p.client.Transport = &http.Transport{
			TLSClientConfig: tlsconfig,
			DialContext:     dialContext,
		}
	}

	if p.opt.AsLogging != nil && p.opt.AsLogging.Enable {
		p.opt.pointOpt = dkpt.LOptElectionV2(p.opt.Election)
	} else {
		p.opt.pointOpt = dkpt.MOptElectionV2(p.opt.Election)
	}

	return &p, nil
}

func (p *Prom) Option() *Option {
	return p.opt
}

func (p *Prom) SetClient(cli *http.Client) {
	p.client = cli
}

func (p *Prom) GetReq(url string) (*http.Request, error) {
	var (
		req *http.Request
		err error
	)

	if len(p.opt.Auth) > 0 {
		if authType, ok := p.opt.Auth["type"]; ok {
			if authFunc, ok := AuthMaps[authType]; ok {
				req, err = authFunc(p.opt.Auth, url)
			} else {
				req, err = http.NewRequest("GET", url, nil)
			}
		}
	} else {
		req, err = http.NewRequest("GET", url, nil)
	}
	for k, v := range p.opt.HTTPHeaders {
		req.Header.Set(k, v)
	}
	return req, err
}

func (p *Prom) Request(url string) (*http.Response, error) {
	req, err := p.GetReq(url)
	if err != nil {
		return nil, err
	}

	r, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}

	return r, nil
}

// CollectFromHTTPV2 convert point from old format to new format.
func (p *Prom) CollectFromHTTPV2(u string) ([]*point.Point, error) {
	// Here got old format point.
	dkPts, err := p.CollectFromHTTP(u)
	if err != nil {
		return nil, err
	}

	// Convert
	pts := inpt.Dkpt2point(dkPts...)

	return pts, nil
}

// CollectFromHTTP Deprecated: use CollectFromHTTPV2.
func (p *Prom) CollectFromHTTP(u string) ([]*dkpt.Point, error) {
	resp, err := p.Request(u)
	if err != nil {
		if p.opt.IgnoreReqErr {
			return []*dkpt.Point{}, nil
		} else {
			return nil, fmt.Errorf("collect from %s: %w", u, err)
		}
	}
	defer resp.Body.Close() //nolint:errcheck
	pts, err := p.text2Metrics(resp.Body, u)
	if err != nil {
		return nil, err
	}
	return pts, nil
}

// CollectFromFileV2 convert point from old format to new format.
func (p *Prom) CollectFromFileV2(filepath string) ([]*point.Point, error) {
	// Here got old format point.
	dkPts, err := p.CollectFromFile(filepath)
	if err != nil {
		return nil, err
	}

	// Convert
	pts := inpt.Dkpt2point(dkPts...)

	return pts, nil
}

// CollectFromFile Deprecated: use CollectFromFileV2.
func (p *Prom) CollectFromFile(filepath string) ([]*dkpt.Point, error) {
	f, err := os.OpenFile(filepath, os.O_RDONLY, 0o600) //nolint:gosec
	if err != nil {
		return nil, err
	}
	defer f.Close() //nolint:errcheck,gosec
	return p.text2Metrics(f, "")
}

// WriteMetricText2File scrapes raw prometheus metric text from u
// then appends them directly to file p.opt.Output.
func (p *Prom) WriteMetricText2File(u string) error {
	fp := p.opt.Output
	if !path.IsAbs(fp) {
		fp = filepath.Join(datakit.InstallDir, fp)
	}
	// Append to file if already exist.
	f, err := os.OpenFile(fp, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0o660) //nolint:gosec
	if err != nil {
		return err
	}
	defer f.Close() //nolint:errcheck,gosec

	uu, err := url.Parse(u)
	if err != nil {
		return fmt.Errorf("url parse error, %w", err)
	}
	// If url is configured as local path file, prom does not collect from it.
	if uu.Scheme != "http" && uu.Scheme != "https" {
		return fmt.Errorf("url is neither http nor https")
	}

	resp, err := p.client.Get(u)
	if err != nil {
		return err
	}
	defer resp.Body.Close() //nolint:errcheck

	if resp.ContentLength > p.opt.MaxFileSize {
		return fmt.Errorf("content length is too large to handle, max: %d, got: %d", p.opt.MaxFileSize, resp.ContentLength)
	}
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if int64(len(data)) > p.opt.MaxFileSize {
		return fmt.Errorf("content length is too large to handle, max: %d, got: %d", p.opt.MaxFileSize, len(data))
	}
	if _, err := f.Write(data); err != nil {
		return err
	}
	return nil
}
