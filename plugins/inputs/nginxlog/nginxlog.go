package nginxlog

import (
	"bytes"
	"io/ioutil"
	"path/filepath"
	"text/template"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/toml"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

const (
	nginxLogSample = `
#[[access_log]]
#  file='/var/log/nginx/access.log'
#  measurement='nginx_access'
`

	nginxLogTelegrafTemplate = `
[[inputs.logparser]]
    files = ['{{.LogFile}}']
    [inputs.logparser.grok]
      patterns = ['{{.Pattern}}']
      measurement = '{{.Measurement}}'
`
	accessLogPattern = `%{COMMON_LOG_FORMAT}`
	errorLogPattern  = `%{HTTPD_ERRORLOG}`
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

func (_ *NginxLogConfig) SampleConfig() string {
	return nginxLogSample
}

func (_ *NginxLogConfig) Description() string {
	return ""
}

func (c *NginxLogConfig) Gather(telegraf.Accumulator) error {
	return nil
}

func (c *NginxLogConfig) FilePath(cfgdir string) string {
	return filepath.Join(cfgdir, "nginx", "nginxlog.conf")
}

func (c *NginxLogConfig) Load(f string) error {
	cfgdata, err := ioutil.ReadFile(f)
	if err != nil {
		return err
	}

	if err = toml.Unmarshal(cfgdata, c); err != nil {
		return err
	}

	if len(c.AccessLogs) == 0 && len(c.ErrorLogs) == 0 {
		return config.ErrNoTelegrafConf
	}

	return nil
}

func (c *NginxLogConfig) ToTelegraf(f string) (string, error) {
	if len(c.AccessLogs) == 0 && len(c.ErrorLogs) == 0 {
		return "", nil
	}
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

func init() {
	inputs.Add("nginxlog", func() telegraf.Input {
		return &NginxLogConfig{}
	})
}
