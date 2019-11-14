package nginx

import (
	"bytes"
	"io/ioutil"
	"path/filepath"
	"text/template"

	"github.com/influxdata/toml"
)

const (
	nginxLogSample = `
[[access_log]]
  file="/var/log/nginx/access.log"
  measurement="nginx_access"
`

	nginxLogTelegrafTemplate = `
[[inputs.logparser]]
    files = ["{{.LogFile}}"]
    [inputs.logparser.grok]
      patterns = ["{{.Pattern}}"]
      measurement = "{{.Measurement}}"
`
	accessLogPattern = `COMMON_LOG_FORMAT`
	errorLogPattern  = `HTTPD_ERRORLOG`
)

type NginxAccessLog struct {
	LogFile     string `toml:"file"`
	Measurement string `toml:"measurement"`
	Pattern     string `toml:"-"`
}

type NginxErrorLog struct {
	LogFile     string `toml:"file"`
	Measurement string `toml:"measurement"`
	//Level       string
	Pattern string `toml:"-"`
}

type NginxLogConfig struct {
	//Logs []*NginxAccessLog `yaml:"logs"`
	AccessLogs []*NginxAccessLog `toml:"access_log"`
	ErrorLogs  []*NginxErrorLog  `toml:"error_log"`
}

func (c *NginxLogConfig) SampleConfig() string {
	return nginxLogSample
}

func (c *NginxLogConfig) ToTelegraf() (string, error) {
	cfg := ""
	t := template.New("")
	var err error
	buf := bytes.NewBuffer([]byte{})

	t, err = t.Parse(nginxLogTelegrafTemplate)
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

func (c *NginxLogConfig) FilePath(dir string) string {
	nginxdir := filepath.Join(dir, "nginx")
	return filepath.Join(nginxdir, "nginxlog.toml")
}

func (c *NginxLogConfig) Load(f string) error {
	cfgdata, err := ioutil.ReadFile(f)
	if err != nil {
		return err
	}

	return toml.Unmarshal(cfgdata, c)
}
