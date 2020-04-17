package config

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"text/template"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/serializers"

	"github.com/influxdata/toml"
	"github.com/influxdata/toml/ast"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/git"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/models"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/outputs/file"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/outputs/http"
)

var (
	ServiceName = "datakit"

	MainCfgPath   string
	ExecutableDir string

	ConvertedCfg []string

	AgentLogFile string

	DKConfig *Config

	MaxLifeCheckInterval time.Duration
)

const (
	mainConfigTemplate = `uuid='{{.UUID}}'
ftdataway='{{.FtGateway}}'
log='{{.Log}}'
log_level='{{.LogLevel}}'
config_dir='{{.ConfigDir}}'

## Override default hostname, if empty use os.Hostname()
hostname = ""
## If set to true, do no set the "host" tag.
omit_hostname = false

# ##tell dataway the interval to check datakit alive
#max_post_interval = '1m'

## Global tags can be specified here in key="value" format.
#[global_tags]
# name = 'admin'

# Configuration for agent
#[agent]
#  ## Default data collection interval for all inputs
#  interval = "10s"
#  ## Rounds collection interval to 'interval'
#  ## ie, if interval="10s" then always collect on :00, :10, :20, etc.
#  round_interval = true

#  ## Telegraf will send metrics to outputs in batches of at most
#  ## metric_batch_size metrics.
#  ## This controls the size of writes that Telegraf sends to output plugins.
#  metric_batch_size = 1000

#  ## Maximum number of unwritten metrics per output.
#  metric_buffer_limit = 100000

#  ## Collection jitter is used to jitter the collection by a random amount.
#  ## Each plugin will sleep for a random time within jitter before collecting.
#  ## This can be used to avoid many plugins querying things like sysfs at the
#  ## same time, which can have a measurable effect on the system.
#  collection_jitter = "0s"

#  ## Default flushing interval for all outputs. Maximum flush_interval will be
#  ## flush_interval + flush_jitter
#  flush_interval = "10s"
#  ## Jitter the flush interval by a random amount. This is primarily to avoid
#  ## large write spikes for users running a large number of telegraf instances.
#  ## ie, a jitter of 5s and interval 10s means flushes will happen every 10-15s
#  flush_jitter = "0s"

#  ## By default or when set to "0s", precision will be set to the same
#  ## timestamp order as the collection interval, with the maximum being 1s.
#  ##   ie, when interval = "10s", precision will be "1s"
#  ##       when interval = "250ms", precision will be "1ms"
#  ## Precision will NOT be used for service inputs. It is up to each individual
#  ## service input to set the timestamp at the appropriate precision.
#  ## Valid time units are "ns", "us" (or "µs"), "ms", "s".
#  precision = "ns"

#  ## Log at debug level.
#  debug = true
#  ## Log only error level messages.
#  # quiet = false

#  ## Log file name, the empty string means to log to stderr.
#  logfile = "/var/log/telegraf/telegraf.log"

#  ## The logfile will be rotated after the time interval specified.  When set
#  ## to 0 no time based rotation is performed.  Logs are rotated only when
#  ## written to, if there is no log activity rotation may be delayed.
#  # logfile_rotation_interval = "0d"

#  ## The logfile will be rotated when it becomes larger than the specified
#  ## size.  When set to 0 no size based rotation is performed.
#  # logfile_rotation_max_size = "0MB"

#  ## Maximum number of rotated archives to keep, any older logs are deleted.
#  ## If set to -1, no archives are removed.
#  # logfile_rotation_max_archives = 5

#  ## Override default hostname, if empty use os.Hostname()
#  hostname = ""
#  ## If set to true, do no set the "host" tag in the telegraf agent.
#  omit_hostname = false

`
)

type Config struct {
	MainCfg          *MainConfig
	TelegrafAgentCfg *TelegrafAgentConfig
	Inputs           []*models.RunningInput
	//Outputs          []*models.RunningOutput

	InputFilters []string

	enableDataclean bool
	datacleanBind   string
}

func NewConfig() *Config {
	c := &Config{
		TelegrafAgentCfg: defaultTelegrafAgentCfg(),
		MainCfg: &MainConfig{
			GlobalTags:    map[string]string{},
			FlushInterval: internal.Duration{Duration: 10 * time.Second},
			Interval:      internal.Duration{Duration: 10 * time.Second},
			LogLevel:      "info",
			RoundInterval: false,
		},
		Inputs: make([]*models.RunningInput, 0),
		//Outputs: make([]*models.RunningOutput, 0),
	}
	return c
}

type MainConfig struct {
	UUID      string `toml:"uuid"`
	FtGateway string `toml:"ftdataway"`

	Log      string `toml:"log"`
	LogLevel string `toml:"log_level"`

	ConfigDir string `toml:"config_dir,omitempty"`

	//验证dk存活
	MaxPostInterval internal.Duration `toml:"max_post_interval"`

	DataCleanTemplate string

	GlobalTags map[string]string `toml:"global_tags"`

	Interval      internal.Duration `toml:"interval"`
	RoundInterval bool
	FlushInterval internal.Duration
	FlushJitter   internal.Duration

	OutputsFile string `toml:"output_file,omitempty"`

	Hostname     string
	OmitHostname bool
}

type ConvertTelegrafConfig interface {
	Load(f string) error
	ToTelegraf(f string) (string, error)
	FilePath(cfgdir string) string
}

type DatacleanConfig interface {
	CheckRoute(route string) bool
	Bindaddr() string
}

func UserAgent() string {
	return fmt.Sprintf("datakit(%s), %s-%s", git.Version, runtime.GOOS, runtime.GOARCH)
}

func (c *Config) LoadMainConfig(ctx context.Context, maincfg string) error {
	data, err := ioutil.ReadFile(maincfg)
	if err != nil {
		return err
	}

	if tbl, err := parseConfig(data); err != nil {
		return err
	} else {

		//telegraf的相应配置
		bAgentSetLogLevel := false
		bAgentSetOmitHost := false
		bAgentSetHostname := false
		if val, ok := tbl.Fields["agent"]; ok {
			subTable, ok := val.(*ast.Table)
			if !ok {
				return fmt.Errorf("invalid agent configuration")
			}

			if _, ok := subTable.Fields["debug"]; ok {
				bAgentSetLogLevel = true
			}

			if _, ok := subTable.Fields["omit_hostname"]; ok {
				bAgentSetOmitHost = true
			}

			if _, ok := subTable.Fields["hostname"]; ok {
				bAgentSetHostname = true
			}

			if err = toml.UnmarshalTable(subTable, c.TelegrafAgentCfg); err != nil {
				return fmt.Errorf("invalid agent configuration, %s", err)
			}

			delete(tbl.Fields, "agent")
		}

		if err := toml.UnmarshalTable(tbl, c.MainCfg); err != nil {
			return err
		}

		if !c.MainCfg.OmitHostname {
			if c.MainCfg.Hostname == "" {
				hostname, err := os.Hostname()
				if err != nil {
					return err
				}

				c.MainCfg.Hostname = hostname
			}

			c.MainCfg.GlobalTags["host"] = c.MainCfg.Hostname
		}

		if c.TelegrafAgentCfg.LogTarget == "file" && c.TelegrafAgentCfg.Logfile == "" {
			c.TelegrafAgentCfg.Logfile = filepath.Join(ExecutableDir, "embed", "agent.log")
		}

		if AgentLogFile != "" {
			c.TelegrafAgentCfg.Logfile = AgentLogFile
		}

		//如果telegraf的agent相关配置没有，则默认使用datakit的同名配置
		if !bAgentSetLogLevel {
			c.TelegrafAgentCfg.Debug = (strings.ToLower(c.MainCfg.LogLevel) == "debug")
		}

		if !bAgentSetOmitHost {
			c.TelegrafAgentCfg.OmitHostname = c.MainCfg.OmitHostname
		}

		if !bAgentSetHostname {
			c.TelegrafAgentCfg.Hostname = c.MainCfg.Hostname
		}
	}

	return nil
}

func CheckConfd(cfgdir string) error {
	dir, err := ioutil.ReadDir(cfgdir)
	if err != nil {
		return err
	}

	configed := []string{}
	invalids := []string{}

	checkSubDir := func(path string) error {

		dir, err := ioutil.ReadDir(path)
		if err != nil {
			return err
		}

		for _, item := range dir {
			if item.IsDir() {
				continue
			}

			filename := item.Name()

			if filename == "." || filename == ".." {
				continue
			}

			if filepath.Ext(filename) != ".conf" {
				continue
			}

			var data []byte
			data, err = ioutil.ReadFile(filepath.Join(path, filename))
			if err != nil {
				return err
			}

			if len(data) == 0 {
				return ErrConfigNotFound
			}

			if tbl, err := toml.Parse(data); err != nil {
				invalids = append(invalids, filename)
				return err
			} else {
				if len(tbl.Fields) > 0 {
					configed = append(configed, filename)
				}
			}

		}

		return nil
	}

	for _, item := range dir {
		if !item.IsDir() {
			continue
		}

		if item.Name() == "." || item.Name() == ".." {
			continue
		}

		checkSubDir(filepath.Join(cfgdir, item.Name()))
	}

	log.Printf("inputs: %s", strings.Join(configed, ","))
	log.Printf("error configuration: %s", strings.Join(invalids, ","))

	return nil
}

//LoadConfig 加载conf.d下的所有配置文件
func (c *Config) LoadConfig(ctx context.Context) error {

	for name, creator := range inputs.Inputs {

		select {
		case <-ctx.Done():
			return context.Canceled
		default:
		}

		if len(c.InputFilters) > 0 {
			if !sliceContains(name, c.InputFilters) {
				continue
			}
		}

		//apachelog和nginxlog和telegraf的nginx和apache共享一个目录
		//这些采集器将转化为telegraf的采集器
		if name == "apachelog" || name == "nginxlog" {
			if p, ok := creator().(ConvertTelegrafConfig); ok {
				path := p.FilePath(c.MainCfg.ConfigDir)
				if err := p.Load(path); err != nil {
					if err == ErrConfigNotFound {
						continue
					} else {
						return fmt.Errorf("Error loading config file %s, %s", path, err)
					}
				}
				if telegrafCfg, err := p.ToTelegraf(path); err == nil {
					ConvertedCfg = append(ConvertedCfg, telegrafCfg)
				} else {
					return fmt.Errorf("convert %s failed, %s", path, err)
				}
			}
			continue
		}

		input := creator()

		var data []byte

		if internalData, ok := inputs.InternalInputsData[name]; ok {
			data = internalData
		} else {
			path := filepath.Join(c.MainCfg.ConfigDir, name, fmt.Sprintf("%s.conf", name))

			_, err := os.Stat(path)
			if err != nil && os.IsNotExist(err) {
				continue
			}

			data, err = ioutil.ReadFile(path)
			if err != nil {
				return fmt.Errorf("Error loading config file %s, %s", path, err)
			}
		}

		tbl, err := parseConfig(data)
		if err != nil {
			return fmt.Errorf("Error parse config %s, %s", name, err)
		}

		if len(tbl.Fields) == 0 {
			continue
		}

		if err := c.addInput(name, input, tbl); err != nil {
			return err
		}

		if name == "dataclean" {
			if p, ok := input.(DatacleanConfig); ok {
				if p.CheckRoute(c.MainCfg.DataCleanTemplate) {
					c.enableDataclean = true
					c.datacleanBind = p.Bindaddr()
				}
			}
		}
	}

	if c.MainCfg.MaxPostInterval.Duration == 0 {
		//默认使用最大周期
		var maxInterval time.Duration
		bHaveIntervalInput := false
		for _, ri := range c.Inputs {
			if _, ok := ri.Input.(telegraf.ServiceInput); ok {
				continue
			}
			bHaveIntervalInput = true
			if ri.Config.Interval > maxInterval {
				maxInterval = ri.Config.Interval
			}
		}

		if bHaveIntervalInput {
			if maxInterval == 0 {
				maxInterval = c.MainCfg.Interval.Duration
			}
			maxInterval += 10 * time.Second
		}

		MaxLifeCheckInterval = maxInterval
	} else {
		MaxLifeCheckInterval = c.MainCfg.MaxPostInterval.Duration
	}

	return LoadTelegrafConfigs(ctx, c.MainCfg.ConfigDir, c.InputFilters)
}

func (c *Config) LoadOutputs() ([]*models.RunningOutput, error) {

	var outputs []*models.RunningOutput

	if c.enableDataclean {

		log.Printf("enable self data clean")

		httpOutput := http.NewHttpOutput()
		if httpOutput.Headers == nil {
			httpOutput.Headers = map[string]string{}
		}
		httpOutput.Headers[`X-Datakit-UUID`] = c.MainCfg.UUID
		httpOutput.Headers[`X-Version`] = git.Version
		httpOutput.Headers[`X-TraceId`] = `self_` + c.MainCfg.UUID
		httpOutput.Headers[`User-Agent`] = UserAgent()
		if MaxLifeCheckInterval > 0 {
			httpOutput.Headers[`X-Max-POST-Interval`] = internal.IntervalString(MaxLifeCheckInterval)
		}
		httpOutput.ContentEncoding = "gzip"
		// datawayUrl := ""
		// u, err := url.Parse(c.MainCfg.FtGateway)
		// if err == nil {
		// 	u.Scheme = "http"
		// 	u.Host = c.datacleanBind
		// 	u.Path = `/v1/write/metrics`
		// 	datawayUrl = u.String()
		// } else {
		// 	datawayUrl = fmt.Sprintf(`http://%s/v1/write/metrics?template=%s`, c.datacleanBind, c.MainCfg.DataCleanTemplate)
		// }
		// datawayUrl += fmt.Sprintf("?template=%s", c.MainCfg.DataCleanTemplate)
		httpOutput.URL = fmt.Sprintf(`http://%s/v1/write/metrics?template=%s`, c.datacleanBind, c.MainCfg.DataCleanTemplate)
		log.Printf("D! self dataway url: %s", httpOutput.URL)
		if ro, err := c.newRunningOutputDirectly("dataclean", httpOutput); err != nil {
			return nil, err
		} else {
			outputs = append(outputs, ro)
		}

	} else {
		if c.MainCfg.OutputsFile != "" {
			fileOutput := file.NewFileOutput()
			fileOutput.Files = []string{c.MainCfg.OutputsFile}
			if ro, err := c.newRunningOutputDirectly("file", fileOutput); err != nil {
				return nil, err
			} else {
				outputs = append(outputs, ro)
			}
		}

		if c.MainCfg.FtGateway != "" {
			httpOutput := http.NewHttpOutput()
			if httpOutput.Headers == nil {
				httpOutput.Headers = map[string]string{}
			}
			httpOutput.Headers[`X-Datakit-UUID`] = c.MainCfg.UUID
			httpOutput.Headers[`X-Version`] = git.Version
			httpOutput.Headers[`X-Version`] = git.Version
			httpOutput.Headers[`User-Agent`] = UserAgent()
			if MaxLifeCheckInterval > 0 {
				httpOutput.Headers[`X-Max-POST-Interval`] = internal.IntervalString(MaxLifeCheckInterval)
			}
			httpOutput.ContentEncoding = "gzip"
			httpOutput.URL = c.MainCfg.FtGateway
			if ro, err := c.newRunningOutputDirectly("http", httpOutput); err != nil {
				return nil, err
			} else {
				outputs = append(outputs, ro)
			}
		}
	}

	return outputs, nil
}

func (c *Config) DumpInputsOutputs() {
	names := []string{}

	for _, p := range c.Inputs {
		names = append(names, p.Config.Name)
	}

	for idx, b := range MetricsEnablesFlags {
		if b {
			names = append(names, SupportsTelegrafMetraicNames[idx])
		}
	}

	log.Printf("avariable inputs: %s", strings.Join(names, ","))

	// names = names[:0]
	// for _, p := range c.Outputs {
	// 	names = append(names, p.Config.Name)
	// }

	// log.Printf("avariable outputs: %s", strings.Join(names, ","))
}

func InitMainCfg(cfg *MainConfig, path string) error {

	var err error
	tm := template.New("")
	tm, err = tm.Parse(mainConfigTemplate)
	if err != nil {
		return fmt.Errorf("Error creating %s, %s", path, err)
	}

	buf := bytes.NewBuffer([]byte{})
	if err = tm.Execute(buf, cfg); err != nil {
		return fmt.Errorf("Error creating %s, %s", path, err)
	}

	if err := ioutil.WriteFile(path, []byte(buf.Bytes()), 0664); err != nil {
		return fmt.Errorf("Error creating %s, %s", path, err)
	}
	return nil
}

func CreateDataDir() error {
	dataDir := filepath.Join(ExecutableDir, "data")
	os.MkdirAll(dataDir, 0755)
	//datakit定义的插件的配置文件
	for name, _ := range inputs.Inputs {
		if name == "zabbix" {
			pluginDataDir := filepath.Join(dataDir, "zabbix")
			if err := os.MkdirAll(pluginDataDir, 0775); err != nil {
				return fmt.Errorf("Error create %s, %s", pluginDataDir, err)
			}
		}
	}
	return nil
}

func CreatePluginConfigs(cfgdir string, upgrade bool) error {

	//datakit定义的插件的配置文件
	for name, creator := range inputs.Inputs {

		if name == "self" {
			continue
		}

		plugindir := ""
		if name == "apachelog" {
			plugindir = filepath.Join(cfgdir, "apache")
		} else if name == "nginxlog" {
			plugindir = filepath.Join(cfgdir, "nginx")
		} else {
			plugindir = filepath.Join(cfgdir, name)
		}
		cfgpath := filepath.Join(plugindir, fmt.Sprintf(`%s.conf`, name))

		if upgrade {
			//更新时，不动它
			_, err := os.Stat(cfgpath)
			if err == nil {
				continue
			}
		}
		if err := os.MkdirAll(plugindir, 0775); err != nil {
			return fmt.Errorf("Error create %s, %s", plugindir, err)
		}
		input := creator()
		if err := ioutil.WriteFile(cfgpath, []byte(input.SampleConfig()), 0666); err != nil {
			return fmt.Errorf("Error create %s, %s", cfgpath, err)
		}
	}

	//创建各个telegraf插件的配置文件
	for _, name := range SupportsTelegrafMetraicNames {

		plugindir := filepath.Join(cfgdir, name)
		cfgpath := filepath.Join(plugindir, fmt.Sprintf(`%s.conf`, name))
		if upgrade {
			//更新时，不动它
			_, err := os.Stat(cfgpath)
			if err == nil {
				continue
			}
		}

		if err := os.MkdirAll(plugindir, 0775); err != nil {
			return fmt.Errorf("Error create %s, %s", plugindir, err)
		}
		if samp, ok := telegrafCfgSamples[name]; ok {

			if err := ioutil.WriteFile(cfgpath, []byte(samp), 0664); err != nil {
				return fmt.Errorf("Error create %s, %s", cfgpath, err)
			}
		}
	}

	return nil
}

func parseConfig(contents []byte) (*ast.Table, error) {
	// contents = trimBOM(contents)

	// parameters := envVarRe.FindAllSubmatch(contents, -1)
	// for _, parameter := range parameters {
	// 	if len(parameter) != 3 {
	// 		continue
	// 	}

	// 	var env_var []byte
	// 	if parameter[1] != nil {
	// 		env_var = parameter[1]
	// 	} else if parameter[2] != nil {
	// 		env_var = parameter[2]
	// 	} else {
	// 		continue
	// 	}

	// 	env_val, ok := os.LookupEnv(strings.TrimPrefix(string(env_var), "$"))
	// 	if ok {
	// 		env_val = escapeEnv(env_val)
	// 		contents = bytes.Replace(contents, parameter[0], []byte(env_val), 1)
	// 	}
	// }

	return toml.Parse(contents)
}

func sliceContains(name string, list []string) bool {
	for _, b := range list {
		if b == name {
			return true
		}
	}
	return false
}

func (c *Config) addInput(name string, input telegraf.Input, table *ast.Table) error {

	if len(c.InputFilters) > 0 && !sliceContains(name, c.InputFilters) {
		return nil
	}

	pluginConfig, err := buildInput(name, table)
	if err != nil {
		return err
	}

	if err := toml.UnmarshalTable(table, input); err != nil {
		return err
	}

	rp := models.NewRunningInput(input, pluginConfig)
	rp.SetDefaultTags(c.MainCfg.GlobalTags)
	c.Inputs = append(c.Inputs, rp)
	return nil
}

func (c *Config) newRunningOutputDirectly(name string, output telegraf.Output) (*models.RunningOutput, error) {

	switch t := output.(type) {
	case serializers.SerializerOutput:
		serializer, err := buildSerializer(name, nil)
		if err != nil {
			return nil, err
		}
		t.SetSerializer(serializer)
	}

	outputConfig, err := buildOutput(name, nil)
	if err != nil {
		return nil, err
	}

	ro := models.NewRunningOutput(name, output, outputConfig, 0, 0)
	return ro, nil
}

// func (c *Config) addOutputDirectly(name string, output telegraf.Output) error {

// 	switch t := output.(type) {
// 	case serializers.SerializerOutput:
// 		serializer, err := buildSerializer(name, nil)
// 		if err != nil {
// 			return err
// 		}
// 		t.SetSerializer(serializer)
// 	}

// 	outputConfig, err := buildOutput(name, nil)
// 	if err != nil {
// 		return err
// 	}

// 	ro := models.NewRunningOutput(name, output, outputConfig, 0, 0)
// 	c.Outputs = append(c.Outputs, ro)
// 	return nil
// }

// func (c *Config) addOutput(name string, output telegraf.Output, table *ast.Table) error {

// 	switch t := output.(type) {
// 	case serializers.SerializerOutput:
// 		serializer, err := buildSerializer(name, table)
// 		if err != nil {
// 			return err
// 		}
// 		t.SetSerializer(serializer)
// 	}

// 	outputConfig, err := buildOutput(name, table)
// 	if err != nil {
// 		return err
// 	}

// 	if err := toml.UnmarshalTable(table, output); err != nil {
// 		return err
// 	}

// 	ro := models.NewRunningOutput(name, output, outputConfig, 0, 0)
// 	c.Outputs = append(c.Outputs, ro)
// 	return nil
// }

func buildSerializer(name string, tbl *ast.Table) (serializers.Serializer, error) {
	c := &serializers.Config{TimestampUnits: time.Duration(1 * time.Second)}

	if tbl != nil {
		if node, ok := tbl.Fields["data_format"]; ok {
			if kv, ok := node.(*ast.KeyValue); ok {
				if str, ok := kv.Value.(*ast.String); ok {
					c.DataFormat = str.Value
				}
			}
		}
	}

	if c.DataFormat == "" {
		c.DataFormat = "influx"
	}

	if tbl != nil {
		delete(tbl.Fields, "influx_max_line_bytes")
		delete(tbl.Fields, "influx_sort_fields")
		delete(tbl.Fields, "influx_uint_support")
		delete(tbl.Fields, "graphite_tag_support")
		delete(tbl.Fields, "data_format")
		delete(tbl.Fields, "prefix")
		delete(tbl.Fields, "template")
		delete(tbl.Fields, "json_timestamp_units")
		delete(tbl.Fields, "splunkmetric_hec_routing")
		delete(tbl.Fields, "wavefront_source_override")
		delete(tbl.Fields, "wavefront_use_strict")
	}

	return serializers.NewSerializer(c)
}

func buildOutput(name string, tbl *ast.Table) (*models.OutputConfig, error) {
	var filter models.Filter
	var err error

	if tbl != nil {
		filter, err = buildFilter(tbl)
		if err != nil {
			return nil, err
		}
	}

	oc := &models.OutputConfig{
		Name:   name,
		Filter: filter,
	}

	// TODO
	// Outputs don't support FieldDrop/FieldPass, so set to NameDrop/NamePass
	if len(oc.Filter.FieldDrop) > 0 {
		oc.Filter.NameDrop = oc.Filter.FieldDrop
	}
	if len(oc.Filter.FieldPass) > 0 {
		oc.Filter.NamePass = oc.Filter.FieldPass
	}

	if tbl != nil {
		if node, ok := tbl.Fields["flush_interval"]; ok {
			if kv, ok := node.(*ast.KeyValue); ok {
				if str, ok := kv.Value.(*ast.String); ok {
					dur, err := time.ParseDuration(str.Value)
					if err != nil {
						return nil, err
					}

					oc.FlushInterval = dur
				}
			}
		}

		if node, ok := tbl.Fields["metric_buffer_limit"]; ok {
			if kv, ok := node.(*ast.KeyValue); ok {
				if integer, ok := kv.Value.(*ast.Integer); ok {
					v, err := integer.Int()
					if err != nil {
						return nil, err
					}
					oc.MetricBufferLimit = int(v)
				}
			}
		}

		if node, ok := tbl.Fields["metric_batch_size"]; ok {
			if kv, ok := node.(*ast.KeyValue); ok {
				if integer, ok := kv.Value.(*ast.Integer); ok {
					v, err := integer.Int()
					if err != nil {
						return nil, err
					}
					oc.MetricBatchSize = int(v)
				}
			}
		}

		if node, ok := tbl.Fields["alias"]; ok {
			if kv, ok := node.(*ast.KeyValue); ok {
				if str, ok := kv.Value.(*ast.String); ok {
					oc.Alias = str.Value
				}
			}
		}

		delete(tbl.Fields, "flush_interval")
		delete(tbl.Fields, "metric_buffer_limit")
		delete(tbl.Fields, "metric_batch_size")
		delete(tbl.Fields, "alias")
	}

	return oc, nil
}

func buildInput(name string, tbl *ast.Table) (*models.InputConfig, error) {
	cp := &models.InputConfig{Name: name}
	if node, ok := tbl.Fields["interval"]; ok {
		if kv, ok := node.(*ast.KeyValue); ok {
			if str, ok := kv.Value.(*ast.String); ok {
				dur, err := time.ParseDuration(str.Value)
				if err != nil {
					return nil, err
				}

				cp.Interval = dur
			}
		}
	}

	if node, ok := tbl.Fields["name_prefix"]; ok {
		if kv, ok := node.(*ast.KeyValue); ok {
			if str, ok := kv.Value.(*ast.String); ok {
				cp.MeasurementPrefix = str.Value
			}
		}
	}

	if node, ok := tbl.Fields["name_suffix"]; ok {
		if kv, ok := node.(*ast.KeyValue); ok {
			if str, ok := kv.Value.(*ast.String); ok {
				cp.MeasurementSuffix = str.Value
			}
		}
	}

	if node, ok := tbl.Fields["name_override"]; ok {
		if kv, ok := node.(*ast.KeyValue); ok {
			if str, ok := kv.Value.(*ast.String); ok {
				cp.NameOverride = str.Value
			}
		}
	}

	if node, ok := tbl.Fields["alias"]; ok {
		if kv, ok := node.(*ast.KeyValue); ok {
			if str, ok := kv.Value.(*ast.String); ok {
				cp.Alias = str.Value
			}
		}
	}

	cp.Tags = make(map[string]string)
	if node, ok := tbl.Fields["tags"]; ok {
		if subtbl, ok := node.(*ast.Table); ok {
			if err := toml.UnmarshalTable(subtbl, cp.Tags); err != nil {
				log.Printf("E! Could not parse tags for input %s\n", name)
			}
		}
	}

	delete(tbl.Fields, "name_prefix")
	delete(tbl.Fields, "name_suffix")
	delete(tbl.Fields, "name_override")
	delete(tbl.Fields, "alias")
	delete(tbl.Fields, "interval")
	delete(tbl.Fields, "tags")
	var err error
	cp.Filter, err = buildFilter(tbl)
	if err != nil {
		return cp, err
	}
	return cp, nil
}

func buildFilter(tbl *ast.Table) (models.Filter, error) {
	f := models.Filter{}

	if node, ok := tbl.Fields["namepass"]; ok {
		if kv, ok := node.(*ast.KeyValue); ok {
			if ary, ok := kv.Value.(*ast.Array); ok {
				for _, elem := range ary.Value {
					if str, ok := elem.(*ast.String); ok {
						f.NamePass = append(f.NamePass, str.Value)
					}
				}
			}
		}
	}

	if node, ok := tbl.Fields["namedrop"]; ok {
		if kv, ok := node.(*ast.KeyValue); ok {
			if ary, ok := kv.Value.(*ast.Array); ok {
				for _, elem := range ary.Value {
					if str, ok := elem.(*ast.String); ok {
						f.NameDrop = append(f.NameDrop, str.Value)
					}
				}
			}
		}
	}

	fields := []string{"pass", "fieldpass"}
	for _, field := range fields {
		if node, ok := tbl.Fields[field]; ok {
			if kv, ok := node.(*ast.KeyValue); ok {
				if ary, ok := kv.Value.(*ast.Array); ok {
					for _, elem := range ary.Value {
						if str, ok := elem.(*ast.String); ok {
							f.FieldPass = append(f.FieldPass, str.Value)
						}
					}
				}
			}
		}
	}

	fields = []string{"drop", "fielddrop"}
	for _, field := range fields {
		if node, ok := tbl.Fields[field]; ok {
			if kv, ok := node.(*ast.KeyValue); ok {
				if ary, ok := kv.Value.(*ast.Array); ok {
					for _, elem := range ary.Value {
						if str, ok := elem.(*ast.String); ok {
							f.FieldDrop = append(f.FieldDrop, str.Value)
						}
					}
				}
			}
		}
	}

	if node, ok := tbl.Fields["tagpass"]; ok {
		if subtbl, ok := node.(*ast.Table); ok {
			for name, val := range subtbl.Fields {
				if kv, ok := val.(*ast.KeyValue); ok {
					tagfilter := &models.TagFilter{Name: name}
					if ary, ok := kv.Value.(*ast.Array); ok {
						for _, elem := range ary.Value {
							if str, ok := elem.(*ast.String); ok {
								tagfilter.Filter = append(tagfilter.Filter, str.Value)
							}
						}
					}
					f.TagPass = append(f.TagPass, *tagfilter)
				}
			}
		}
	}

	if node, ok := tbl.Fields["tagdrop"]; ok {
		if subtbl, ok := node.(*ast.Table); ok {
			for name, val := range subtbl.Fields {
				if kv, ok := val.(*ast.KeyValue); ok {
					tagfilter := &models.TagFilter{Name: name}
					if ary, ok := kv.Value.(*ast.Array); ok {
						for _, elem := range ary.Value {
							if str, ok := elem.(*ast.String); ok {
								tagfilter.Filter = append(tagfilter.Filter, str.Value)
							}
						}
					}
					f.TagDrop = append(f.TagDrop, *tagfilter)
				}
			}
		}
	}

	if node, ok := tbl.Fields["tagexclude"]; ok {
		if kv, ok := node.(*ast.KeyValue); ok {
			if ary, ok := kv.Value.(*ast.Array); ok {
				for _, elem := range ary.Value {
					if str, ok := elem.(*ast.String); ok {
						f.TagExclude = append(f.TagExclude, str.Value)
					}
				}
			}
		}
	}

	if node, ok := tbl.Fields["taginclude"]; ok {
		if kv, ok := node.(*ast.KeyValue); ok {
			if ary, ok := kv.Value.(*ast.Array); ok {
				for _, elem := range ary.Value {
					if str, ok := elem.(*ast.String); ok {
						f.TagInclude = append(f.TagInclude, str.Value)
					}
				}
			}
		}
	}
	if err := f.Compile(); err != nil {
		return f, err
	}

	delete(tbl.Fields, "namedrop")
	delete(tbl.Fields, "namepass")
	delete(tbl.Fields, "fielddrop")
	delete(tbl.Fields, "fieldpass")
	delete(tbl.Fields, "drop")
	delete(tbl.Fields, "pass")
	delete(tbl.Fields, "tagdrop")
	delete(tbl.Fields, "tagpass")
	delete(tbl.Fields, "tagexclude")
	delete(tbl.Fields, "taginclude")
	return f, nil
}
