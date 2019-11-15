package apache

import (
	"bytes"
	"io/ioutil"
	"path/filepath"
	"text/template"

	"github.com/influxdata/toml"
)

const (
	apacheLogSample = `
disable = true
[[access_log]]
  file="/var/log/apache/access.log"
  measurement="apache_access"
`

	apacheLogTelegrafTemplate = `
[[inputs.logparser]]
    files = ["{{.LogFile}}"]
    [inputs.logparser.grok]
      patterns = ["{{.Pattern}}"]
      measurement = "{{.Measurement}}"
`
	accessLogPattern = `%{COMMON_LOG_FORMAT}`
	errorLogPattern  = `%{HTTPD_ERRORLOG}`
)

type ApacheAccessLog struct {
	LogFile     string `toml:"file"`
	Measurement string `toml:"measurement"`
	Pattern     string `toml:"-"`
}

type ApacheErrorLog struct {
	LogFile     string `toml:"file"`
	Measurement string `toml:"measurement"`
	//Level       string
	Pattern string `toml:"-"`
}

type ApacheLogConfig struct {
	//Logs []*NginxAccessLog `yaml:"logs"`
	Disable    bool               `toml:"disable"`
	AccessLogs []*ApacheAccessLog `toml:"access_log"`
	ErrorLogs  []*ApacheErrorLog  `toml:"error_log"`
}

func (c *ApacheLogConfig) SampleConfig() string {
	return apacheLogSample
}

func (c *ApacheLogConfig) ToTelegraf() (string, error) {
	if c.Disable {
		return "", nil
	}
	cfg := ""
	t := template.New("")
	var err error
	buf := bytes.NewBuffer([]byte{})

	t, err = t.Parse(apacheLogTelegrafTemplate)
	if err != nil {
		return "", err
	}

	for _, l := range c.AccessLogs {
		l.Pattern = accessLogPattern

		if err = t.Execute(buf, l); err != nil {
			return "", err
		}
		cfg += string(buf.Bytes())
		buf.Reset()
	}

	for _, l := range c.ErrorLogs {
		l.Pattern = errorLogPattern

		if err = t.Execute(buf, l); err != nil {
			return "", err
		}
		cfg += string(buf.Bytes())
		buf.Reset()
	}

	return cfg, err
}

func (c *ApacheLogConfig) FilePath(root string) string {
	d := filepath.Join(root, "apache")
	return filepath.Join(d, "apachelog.toml")
}

func (c *ApacheLogConfig) Load(f string) error {
	cfgdata, err := ioutil.ReadFile(f)
	if err != nil {
		return err
	}

	return toml.Unmarshal(cfgdata, c)
}
