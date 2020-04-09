package config

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
	"text/template"
	"time"

	"github.com/influxdata/toml"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/git"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal"
)

//用于支持在datakit.conf中加入telegraf的agent配置
type TelegrafAgentConfig struct {
	// Interval at which to gather information
	Interval internal.Duration

	// RoundInterval rounds collection interval to 'interval'.
	//     ie, if Interval=10s then always collect on :00, :10, :20, etc.
	RoundInterval bool

	// By default or when set to "0s", precision will be set to the same
	// timestamp order as the collection interval, with the maximum being 1s.
	//   ie, when interval = "10s", precision will be "1s"
	//       when interval = "250ms", precision will be "1ms"
	// Precision will NOT be used for service inputs. It is up to each individual
	// service input to set the timestamp at the appropriate precision.
	Precision internal.Duration

	// CollectionJitter is used to jitter the collection by a random amount.
	// Each plugin will sleep for a random time within jitter before collecting.
	// This can be used to avoid many plugins querying things like sysfs at the
	// same time, which can have a measurable effect on the system.
	CollectionJitter internal.Duration

	// FlushInterval is the Interval at which to flush data
	FlushInterval internal.Duration

	// FlushJitter Jitters the flush interval by a random amount.
	// This is primarily to avoid large write spikes for users running a large
	// number of telegraf instances.
	// ie, a jitter of 5s and interval 10s means flushes will happen every 10-15s
	FlushJitter internal.Duration

	// MetricBatchSize is the maximum number of metrics that is wrote to an
	// output plugin in one call.
	MetricBatchSize int

	// MetricBufferLimit is the max number of metrics that each output plugin
	// will cache. The buffer is cleared when a successful write occurs. When
	// full, the oldest metrics will be overwritten. This number should be a
	// multiple of MetricBatchSize. Due to current implementation, this could
	// not be less than 2 times MetricBatchSize.
	MetricBufferLimit int

	// FlushBufferWhenFull tells Telegraf to flush the metric buffer whenever
	// it fills up, regardless of FlushInterval. Setting this option to true
	// does _not_ deactivate FlushInterval.
	FlushBufferWhenFull bool

	// TODO(cam): Remove UTC and parameter, they are no longer
	// valid for the agent config. Leaving them here for now for backwards-
	// compatibility
	UTC bool `toml:"utc"`

	// Debug is the option for running in debug mode
	Debug bool `toml:"debug"`

	// Quiet is the option for running in quiet mode
	Quiet bool `toml:"quiet"`

	// Log target controls the destination for logs and can be one of "file",
	// "stderr" or, on Windows, "eventlog".  When set to "file", the output file
	// is determined by the "logfile" setting.
	LogTarget string `toml:"logtarget"`

	// Name of the file to be logged to when using the "file" logtarget.  If set to
	// the empty string then logs are written to stderr.
	Logfile string `toml:"logfile"`

	// The file will be rotated after the time interval specified.  When set
	// to 0 no time based rotation is performed.
	LogfileRotationInterval internal.Duration `toml:"logfile_rotation_interval"`

	// The logfile will be rotated when it becomes larger than the specified
	// size.  When set to 0 no size based rotation is performed.
	LogfileRotationMaxSize internal.Size `toml:"logfile_rotation_max_size"`

	// Maximum number of rotated archives to keep, any older logs are deleted.
	// If set to -1, no archives are removed.
	LogfileRotationMaxArchives int `toml:"logfile_rotation_max_archives"`

	Hostname     string
	OmitHostname bool
}

func defaultTelegrafAgentCfg() *TelegrafAgentConfig {
	c := &TelegrafAgentConfig{
		Interval: internal.Duration{
			Duration: time.Second * 10,
		},
		RoundInterval:     true,
		MetricBatchSize:   1000,
		MetricBufferLimit: 100000,
		CollectionJitter: internal.Duration{
			Duration: 0,
		},
		FlushInterval: internal.Duration{
			Duration: time.Second * 10,
		},
		FlushJitter: internal.Duration{
			Duration: 0,
		},
		Precision: internal.Duration{
			Duration: time.Nanosecond,
		},
		Debug:                      false,
		Quiet:                      false,
		LogTarget:                  "file",
		LogfileRotationMaxArchives: 5,
		OmitHostname:               false,
	}
	return c
}

//LoadTelegrafConfigs 加载conf.d下telegraf的配置文件
func LoadTelegrafConfigs(ctx context.Context, cfgdir string, inputFilters []string) error {

	for index, name := range SupportsTelegrafMetraicNames {

		select {
		case <-ctx.Done():
			return context.Canceled
		default:
		}

		if len(inputFilters) > 0 {
			if !sliceContains(name, inputFilters) {
				continue
			}
		}

		cfgpath := filepath.Join(cfgdir, name, fmt.Sprintf(`%s.conf`, name))
		err := VerifyToml(cfgpath, true)

		if err == nil {
			MetricsEnablesFlags[index] = true
		} else {
			MetricsEnablesFlags[index] = false
			if err == ErrConfigNotFound {
				//ignore
			} else if err == ErrEmptyInput {
				log.Printf("W! %s, %s", cfgpath, err.Error())
			} else {
				return fmt.Errorf("Error loading config file %s, %s", cfgpath, err)
			}
		}

	}
	return nil
}

const (
	fileOutputTemplate = `
[[outputs.file]]
## Files to write to, "stdout" is a specially handled file.
files = ['{{.OutputFiles}}']
`

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

func marshalAgentCfg(cfg *TelegrafAgentConfig) (string, error) {

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
		LogTarget                  string        `toml:"-"`
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

func GenerateTelegrafConfig(cfg *Config) (string, error) {

	agentcfg, err := marshalAgentCfg(cfg.TelegrafAgentCfg)
	if err != nil {
		return "", err
	}
	agentcfg = "\n[agent]\n" + agentcfg
	agentcfg += "\n"

	globalTags := "[global_tags]\n"
	for k, v := range cfg.MainCfg.GlobalTags {
		tag := fmt.Sprintf("%s='%s'\n", k, v)
		globalTags += tag
	}

	type fileoutCfg struct {
		OutputFiles string
	}

	type httpoutCfg struct {
		FtGateway   string
		DKUUID      string
		DKVERSION   string
		DKUserAgent string
	}

	fileoutstr := ""
	httpoutstr := ""

	if cfg.MainCfg.OutputsFile != "" {
		fileCfg := fileoutCfg{
			OutputFiles: cfg.MainCfg.OutputsFile,
		}

		tpl := template.New("")
		tpl, err = tpl.Parse(fileOutputTemplate)
		if err != nil {
			return "", err
		}

		buf := bytes.NewBuffer([]byte{})
		if err = tpl.Execute(buf, &fileCfg); err != nil {
			return "", err
		}
		fileoutstr = string(buf.Bytes())
	}

	if cfg.MainCfg.FtGateway != "" {
		httpCfg := httpoutCfg{
			FtGateway:   cfg.MainCfg.FtGateway,
			DKUUID:      cfg.MainCfg.UUID,
			DKVERSION:   git.Version,
			DKUserAgent: UserAgent(),
		}

		tpl := template.New("")
		tpl, err = tpl.Parse(httpOutputTemplate)
		if err != nil {
			return "", err
		}

		buf := bytes.NewBuffer([]byte{})
		if err = tpl.Execute(buf, &httpCfg); err != nil {
			return "", err
		}

		httpoutstr = string(buf.Bytes())
	}

	tlegrafConfig := globalTags + agentcfg + fileoutstr + httpoutstr

	pluginCfgs := ""
	for index, n := range SupportsTelegrafMetraicNames {
		if !MetricsEnablesFlags[index] {
			continue
		}
		cfgpath := filepath.Join(cfg.MainCfg.ConfigDir, n, fmt.Sprintf(`%s.conf`, n))
		d, err := ioutil.ReadFile(cfgpath)
		if err != nil {
			return "", err
		}

		pluginCfgs += string(d) + "\n"
	}

	if len(ConvertedCfg) > 0 {
		for _, c := range ConvertedCfg {
			pluginCfgs += c + "\n"
		}
	}

	if pluginCfgs == "" {
		return "", ErrConfigNotFound
	}

	tlegrafConfig += pluginCfgs

	return tlegrafConfig, err
}
