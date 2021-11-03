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
	URL               string   `toml:"url"`
	Output            string   `toml:"output"`
	MaxFileSize       int64    `toml:"max_file_size"`
	MeasurementPrefix string   `toml:"measurement_prefix"`
	MeasurementName   string   `toml:"measurement_name"`
	CacertFile        string   `toml:"tls_ca"`
	CertFile          string   `toml:"tls_cert"`
	KeyFile           string   `toml:"tls_key"`

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

	if opt.URL == "" {
		return nil, fmt.Errorf("invalid URL, cannot be empty")
	}

	p := Prom{opt: opt}
	p.SetClient(&http.Client{Timeout: httpTimeout})

	if opt.TLSOpen {
		tc := &net.TLSClientConfig{
			CaCerts:            []string{opt.CacertFile},
			Cert:               opt.CertFile,
			CertKey:            opt.KeyFile,
			InsecureSkipVerify: defaultInsecureSkipVerify,
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

func (p *Prom) Collect() ([]*io.Point, error) {
	resp, err := p.client.Get(p.opt.URL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close() //nolint:errcheck

	return Text2Metrics(resp.Body, p.opt, p.opt.Tags)
}

// CollectFromFile collects metrics from local file.
// If both Output and URL is configured as local file path,
// preference is given to p.opt.Output other than p.opt.URL.
func (p *Prom) CollectFromFile() ([]*io.Point, error) {
	var f *os.File
	if p.opt.Output != "" {
		f, _ = os.OpenFile(p.opt.Output, os.O_RDONLY, 0600)
	} else {
		fileName := p.opt.URL
		f, _ = os.OpenFile(fileName, os.O_RDONLY, 0600)
	}
	defer f.Close()
	return Text2Metrics(f, p.opt, p.opt.Tags)
}

// WriteFile collects data from p.opt.URL then writes it to p.opt.Output.
// WriteFile will only be called when Output is configured.
func (p *Prom) WriteFile() error {
	// If url is configured as local path file, prom does not collect from it.
	Url, _ := url.Parse(p.opt.URL)
	if Url.Scheme != "http" && Url.Scheme != "https" {
		return fmt.Errorf("url is neither http nor https")
	}

	resp, err := p.client.Get(p.opt.URL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.ContentLength > p.opt.MaxFileSize {
		return fmt.Errorf("content length is too large to handle")
	}
	fp := p.opt.Output
	if !path.IsAbs(fp) {
		fp = filepath.Join(datakit.InstallDir, fp)
	}
	// truncate if file exists
	f, err := os.Create(fp)
	if err != nil {
		return err
	}
	defer f.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if int64(len(data)) > p.opt.MaxFileSize {
		return fmt.Errorf("content length is too large to handle")
	}
	if err != nil {
		return err
	}
	if _, err := f.Write(data); err != nil {
		return err
	}
	return nil
}
