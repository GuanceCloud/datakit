// Package prom used to parsing promemetheuse exportor metrics.
package prom

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/net"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
)

type Rule struct {
	Pattern string `toml:"pattern"`
	Prefix  string `toml:"prefix"`
	Name    string `toml:"name"`
}

type Option struct {
	MetricTypes       []string `toml:"metric_types"`
	MetricNameFilter  []string `toml:"metric_name_filter"`
	Measurements      []Rule   `json:"measurements"`
	TagsIgnore        []string `toml:"tags_ignore"`
	Source            string   `toml:"source"`
	Interval          string   `toml:"interval"`
	URL               string   `toml:"url,omitempty"` // Deprecated
	URLs              []string `toml:"urls"`
	IgnoreReqErr      bool     `toml:"ignore_req_err"`
	Output            string   `toml:"output"`
	MaxFileSize       int64    `toml:"max_file_size"`
	MeasurementPrefix string   `toml:"measurement_prefix"`
	MeasurementName   string   `toml:"measurement_name"`
	CacertFile        string   `toml:"tls_ca"`
	CertFile          string   `toml:"tls_cert"`
	KeyFile           string   `toml:"tls_key"`

	Auth     map[string]string `toml:"auth"`
	Tags     map[string]string `toml:"tags"`
	interval time.Duration

	TLSOpen bool `toml:"tls_open"`
	Disabel bool `toml:"disble"`
}

const defaultInterval = time.Second * 10

func (opt *Option) IsDisable() bool {
	return opt.Disabel
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
	httpTimeout               = time.Second * 10
	defaultInsecureSkipVerify = false
)

type Prom struct {
	opt    *Option
	client *http.Client
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

	p := Prom{opt: opt}
	p.SetClient(&http.Client{Timeout: httpTimeout})

	if opt.TLSOpen {
		caCerts := []string{}
		insecureSkipVerify := defaultInsecureSkipVerify
		if len(opt.CacertFile) != 0 {
			caCerts = append(caCerts, opt.CacertFile)
		} else {
			insecureSkipVerify = true
		}
		tc := &net.TLSClientConfig{
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
		}
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

func (p *Prom) Collect() ([]*io.Point, error) {
	var allPts []*io.Point
	for _, u := range p.opt.URLs {
		resp, err := p.Request(u)
		if err != nil {
			if p.opt.IgnoreReqErr {
				continue
			} else {
				return nil, err
			}
		}
		defer resp.Body.Close() //nolint:errcheck
		pts, err := p.Text2Metrics(resp.Body)
		if err != nil {
			return nil, err
		}
		allPts = append(allPts, pts...)
	}
	return allPts, nil
}

// CollectFromFile collects metrics from local file.
// If both Output and URL is configured as local file path,
// preference is given to p.opt.Output other than p.opt.URL.
func (p *Prom) CollectFromFile() ([]*io.Point, error) {
	var f *os.File
	if p.opt.Output != "" {
		f, _ = os.OpenFile(p.opt.Output, os.O_RDONLY, 0o600)
	} else {
		fileName := p.opt.URL
		f, _ = os.OpenFile(fileName, os.O_RDONLY, 0o600) //nolint:gosec
	}
	defer f.Close() //nolint:errcheck,gosec
	return p.Text2Metrics(f)
}

// WriteFile scrapes metrics from p.opt.URLs then writes them directly to p.opt.Output.
// WriteFile will only be called when field Output is configured.
func (p *Prom) WriteFile() error {
	fp := p.opt.Output
	if !path.IsAbs(fp) {
		fp = filepath.Join(datakit.InstallDir, fp)
	}
	// truncate if file exists
	f, err := os.Create(fp)
	if err != nil {
		return err
	}
	defer f.Close() //nolint:errcheck,gosec

	for _, u := range p.opt.URLs {
		uu, err := url.Parse(u)
		// If url is configured as local path file, prom does not collect from it.
		if err != nil {
			return fmt.Errorf("url parse error, %w", err)
		}

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
		if int64(len(data)) > p.opt.MaxFileSize {
			return fmt.Errorf("content length is too large to handle, max: %d, got: %d", p.opt.MaxFileSize, len(data))
		}
		if err != nil {
			return err
		}
		if _, err := f.Write(data); err != nil {
			return err
		}
	}

	return nil
}
