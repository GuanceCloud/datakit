package apachelog

import (
	"bytes"
	"io/ioutil"
	"path/filepath"
	"text/template"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/toml"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

const (
	apacheLogSample = `
#[[access_log]]
#  file='/var/log/apache/access.log'
#  measurement="apache_access"
`

	apacheLogTelegrafTemplate = `
[[inputs.logparser]]
    files = ['{{.LogFile}}']
    [inputs.logparser.grok]
      patterns = ['{{.Pattern}}']
      measurement = '{{.Measurement}}'
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
	AccessLogs []*ApacheAccessLog `toml:"access_log"`
	ErrorLogs  []*ApacheErrorLog  `toml:"error_log"`
}

func (_ *ApacheLogConfig) SampleConfig() string {
	return apacheLogSample
}

func (_ *ApacheLogConfig) Description() string {
	return ""
}

func (c *ApacheLogConfig) Gather(telegraf.Accumulator) error {
	return nil
}

func (c *ApacheLogConfig) FilePath(cfgdir string) string {
	return filepath.Join(cfgdir, "apache", "apachelog.conf")
}

func (c *ApacheLogConfig) Load(f string) error {
	cfgdata, err := ioutil.ReadFile(f)
	if err != nil {
		return err
	}

	return toml.Unmarshal(cfgdata, c)
}

func (c *ApacheLogConfig) ToTelegraf(f string) (string, error) {
	if len(c.AccessLogs) == 0 && len(c.ErrorLogs) == 0 {
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

func init() {
	inputs.Add("apachelog", func() telegraf.Input {
		return &ApacheLogConfig{}
	})
}
