package telegrafwrap

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"time"

	"github.com/influxdata/toml"

	"github.com/alecthomas/template"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/git"
)

const (
	httpOutputTemplate = `
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

func marshalAgentCfg(cfg *config.TelegrafAgentConfig) (string, error) {

	type dummyAgentCfg struct {
		Interval                   time.Duration
		RoundInterval              bool
		Precision                  time.Duration
		CollectionJitter           time.Duration
		FlushInterval              time.Duration
		FlushJitter                time.Duration
		MetricBatchSize            int
		MetricBufferLimit          int
		FlushBufferWhenFull        bool
		UTC                        bool          `toml:"utc"`
		Debug                      bool          `toml:"debug"`
		Quiet                      bool          `toml:"quiet"`
		LogTarget                  string        `toml:"logtarget"`
		Logfile                    string        `toml:"logfile"`
		LogfileRotationInterval    time.Duration `toml:"logfile_rotation_interval"`
		LogfileRotationMaxSize     int64         `toml:"logfile_rotation_max_size"`
		LogfileRotationMaxArchives int           `toml:"logfile_rotation_max_archives"`
		Hostname                   string
		OmitHostname               bool
	}

	c := &dummyAgentCfg{
		Interval:                   cfg.Interval.Duration / time.Second,
		RoundInterval:              cfg.RoundInterval,
		Precision:                  cfg.Precision.Duration / time.Second,
		CollectionJitter:           cfg.CollectionJitter.Duration / time.Second,
		FlushInterval:              cfg.FlushInterval.Duration / time.Second,
		FlushJitter:                cfg.FlushJitter.Duration / time.Second,
		MetricBatchSize:            cfg.MetricBatchSize,
		MetricBufferLimit:          cfg.MetricBufferLimit,
		FlushBufferWhenFull:        cfg.FlushBufferWhenFull,
		UTC:                        cfg.UTC,
		Debug:                      cfg.Debug,
		Quiet:                      cfg.Quiet,
		LogTarget:                  cfg.LogTarget,
		Logfile:                    cfg.Logfile,
		LogfileRotationInterval:    cfg.LogfileRotationInterval.Duration / time.Second,
		LogfileRotationMaxSize:     cfg.LogfileRotationMaxSize.Size,
		LogfileRotationMaxArchives: cfg.LogfileRotationMaxArchives,
		Hostname:                   cfg.Hostname,
		OmitHostname:               cfg.OmitHostname,
	}

	agdata, err := toml.Marshal(c)
	if err != nil {
		return "", err
	}
	return string(agdata), nil
}

func (s *TelegrafSvr) GenerateTelegrafConfig() (string, error) {

	agentcfg, err := marshalAgentCfg(s.Cfg.TelegrafAgentCfg)
	if err != nil {
		return "", nil
	}
	agentcfg = "\n[agent]\n" + agentcfg
	agentcfg += "\n"

	globalTags := "[global_tags]\n"
	for k, v := range s.Cfg.MainCfg.GlobalTags {
		tag := fmt.Sprintf("%s='%s'\n", k, v)
		globalTags += tag
	}

	type httpoutCfg struct {
		FtGateway   string
		DKUUID      string
		DKVERSION   string
		DKUserAgent string
	}

	cfg := httpoutCfg{
		FtGateway:   s.Cfg.MainCfg.FtGateway,
		DKUUID:      s.Cfg.MainCfg.UUID,
		DKVERSION:   git.Version,
		DKUserAgent: config.UserAgent(),
	}

	tpl := template.New("")
	tpl, err = tpl.Parse(httpOutputTemplate)
	if err != nil {
		return "", err
	}

	buf := bytes.NewBuffer([]byte{})
	if err = tpl.Execute(buf, &cfg); err != nil {
		return "", err
	}

	tlegrafConfig := globalTags + agentcfg + string(buf.Bytes())

	pluginCfgs := ""
	for index, n := range config.SupportsTelegrafMetraicNames {
		if !config.MetricsEnablesFlags[index] {
			continue
		}
		cfgpath := filepath.Join(s.Cfg.MainCfg.ConfigDir, n, fmt.Sprintf(`%s.conf`, n))
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
