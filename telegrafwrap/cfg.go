package telegrafwrap

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/alecthomas/template"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/git"
)

const (
	telegrafConfTemplate = `
[agent]
  interval = "10s"
  round_interval = true

  metric_batch_size = 1000
  metric_buffer_limit = 10000
  collection_jitter = "0s"
  flush_interval = "10s"
  flush_jitter = "0s"
  precision = ""
  logfile='{{.LogFile}}'
  debug = {{.DebugMode}}
  quiet = false
  hostname = ""
  omit_hostname = false

[[outputs.http]]
  url = "{{.FtGateway}}"
  method = "POST"
  data_format = "influx"
  content_encoding = "gzip"

  ## Additional HTTP headers
  [outputs.http.headers]
    ## Should be set manually to "application/json" for json data_format
	X-Datakit-UUID = "{{.DKUUID}}"
	X-Version = "{{.DKVERSION}}"
	User-Agent = '{{.DKUserAgent}}'
`
)

func (s *TelegrafSvr) GenerateTelegrafConfig() (string, error) {

	globalTags := "[global_tags]\n"
	for k, v := range s.MainCfg.GlobalTags {
		tag := fmt.Sprintf("%s='%s'\n", k, v)
		globalTags += tag
	}

	type telegrafCfg struct {
		LogFile     string
		FtGateway   string
		DKUUID      string
		DKVERSION   string
		DKUserAgent string
		DebugMode   bool
	}

	cfg := telegrafCfg{
		LogFile:     filepath.Join(config.ExecutableDir, "agent.log"),
		FtGateway:   s.MainCfg.FtGateway,
		DKUUID:      s.MainCfg.UUID,
		DKVERSION:   git.Version,
		DKUserAgent: config.UserAgent(),
		DebugMode:   s.MainCfg.LogLevel == "debug",
	}

	var err error
	tpl := template.New("")
	tpl, err = tpl.Parse(telegrafConfTemplate)
	if err != nil {
		return "", err
	}

	buf := bytes.NewBuffer([]byte{})
	if err = tpl.Execute(buf, &cfg); err != nil {
		return "", err
	}

	tlegrafConfig := globalTags + string(buf.Bytes())

	pluginCfgs := ""
	for index, n := range config.SupportsTelegrafMetraicNames {
		if !config.MetricsEnablesFlags[index] {
			continue
		}
		cfgpath := filepath.Join(s.MainCfg.ConfigDir, n, fmt.Sprintf(`%s.conf`, n))
		d, err := ioutil.ReadFile(cfgpath)
		if err != nil {
			return "", err
		}

		pluginCfgs += string(d)
	}

	if len(config.ConvertedCfg) > 0 {
		for _, c := range config.ConvertedCfg {
			pluginCfgs += c + "\n"
		}
	}

	if pluginCfgs == "" {
		return "", config.ErrNoTelegrafConf
	}

	tlegrafConfig += pluginCfgs

	return tlegrafConfig, err
}
