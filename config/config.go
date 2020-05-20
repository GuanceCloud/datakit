package config

import (
	"bytes"
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
	"github.com/influxdata/telegraf/logger"
	"github.com/influxdata/telegraf/plugins/serializers"
	"github.com/influxdata/toml"
	"github.com/influxdata/toml/ast"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/git"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/models"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/outputs/file"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/outputs/http"
)

var (
	userAgent = fmt.Sprintf("datakit(%s), %s-%s", git.Version, runtime.GOOS, runtime.GOARCH)

	ServiceName = "datakit"

	ConvertedCfg []string
	AgentLogFile string

	MaxLifeCheckInterval time.Duration

	Cfg         *Config = nil
	InstallDir          = ""
	TelegrafDir         = ""
	DataDir             = ""
	LuaDir              = ""
	ConfdDir            = ""
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

func init() {
	osarch := runtime.GOOS + "/" + runtime.GOARCH

	switch osarch {
	case "windows/amd64":
		InstallDir = `C:\Program Files (x86)\DataFlux\` + ServiceName

	case "windows/386":
		InstallDir = `C:\Program Files\DataFlux\` + ServiceName

	case "linux/amd64", "linux/386", "linux/arm", "linux/arm64",
		"darwin/amd64", "darwin/386",
		"freebsd/amd64", "freebsd/386":
		InstallDir = `/usr/local/cloudcare/DataFlux/` + ServiceName
	default:
		log.Fatal("[fatal] invalid os/arch: %s", osarch)
	}

	if err := os.MkdirAll(filepath.Join(InstallDir, "embed"), os.ModePerm); err != nil {
		log.Fatalf("[error] mkdir embed  failed: %s", err)
	}

	AgentLogFile = filepath.Join(InstallDir, "embed", "agent.log")

	TelegrafDir = filepath.Join(InstallDir, "embed")
	DataDir = filepath.Join(InstallDir, "data")
	LuaDir = filepath.Join(InstallDir, "lua")
	ConfdDir = filepath.Join(InstallDir, "conf.d")
	for _, dir := range []string{TelegrafDir, DataDir, LuaDir, ConfdDir} {
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			log.Fatalf("create %s failed: %s", dir, err)
		}
	}

	Cfg = newDefaultCfg()

	initTelegrafSamples()
}

func newDefaultCfg() *Config {

	return &Config{
		TelegrafAgentCfg: defaultTelegrafAgentCfg(),
		MainCfg: &MainConfig{
			GlobalTags:    map[string]string{},
			FlushInterval: internal.Duration{Duration: 10 * time.Second},
			Interval:      internal.Duration{Duration: 10 * time.Second},

			LogLevel: "info",
			Log:      filepath.Join(InstallDir, "datakit.log"),

			RoundInterval: false,
			cfgPath:       filepath.Join(InstallDir, "datakit.conf"),
		},
		Inputs: []*models.RunningInput{},
	}
}

func LoadCfg() error {
	if err := Cfg.LoadMainConfig(); err != nil {
		return err
	}

	logConfig := logger.LogConfig{
		Debug:     (strings.ToLower(Cfg.MainCfg.LogLevel) == "debug"),
		Quiet:     false,
		LogTarget: logger.LogTargetFile,
		Logfile:   Cfg.MainCfg.Log,
	}

	logConfig.RotationMaxSize.Size = (20 << 10 << 10)
	logger.SetupLogging(logConfig)

	log.Printf("D! set log to %s", logConfig.Logfile)

	createPluginCfgsIfNotExists()

	if err := Cfg.LoadConfig(); err != nil {
		return err
	}

	Cfg.DumpInputsOutputs()

	return nil
}

type MainConfig struct {
	UUID string `toml:"uuid"`

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

	cfgPath string
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

func (c *Config) LoadMainConfig() error {

	data, err := ioutil.ReadFile(c.MainCfg.cfgPath)
	if err != nil {
		return fmt.Errorf("main config error, %s", err.Error())
	}

	tbl, err := parseConfig(data)
	if err != nil {
		return err
	}

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
		c.TelegrafAgentCfg.Logfile = filepath.Join(InstallDir, "embed", "agent.log")
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

	return nil
}

func CheckConfd() error {
	dir, err := ioutil.ReadDir(ConfdDir)
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

		checkSubDir(filepath.Join(ConfdDir, item.Name()))
	}

	log.Printf("inputs: %s", strings.Join(configed, ","))
	log.Printf("error configuration: %s", strings.Join(invalids, ","))

	return nil
}

//LoadConfig 加载conf.d下的所有配置文件
func (c *Config) LoadConfig() error {

	for name, creator := range inputs.Inputs {

		if len(c.InputFilters) > 0 {
			if !sliceContains(name, c.InputFilters) {
				continue
			}
		}

		//apachelog和nginxlog和telegraf的nginx和apache共享一个目录
		//这些采集器将转化为telegraf的采集器
		/*
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
			} */

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

	return LoadTelegrafConfigs(c.MainCfg.ConfigDir, c.InputFilters)
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
		httpOutput.Headers[`User-Agent`] = userAgent
		if MaxLifeCheckInterval > 0 {
			httpOutput.Headers[`X-Max-POST-Interval`] = internal.IntervalString(MaxLifeCheckInterval)
		}
		httpOutput.ContentEncoding = "gzip"
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
			httpOutput.Headers[`User-Agent`] = userAgent
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

	for k, i := range SupportsTelegrafMetraicNames {
		if i.enabled {
			names = append(names, k)
		}
	}

	log.Printf("avariable inputs: %s", strings.Join(names, ","))
}

func InitCfg(dw string) error {
	if err := initMainCfg(dw); err != nil {
		return err
	}

	// clean all old plugin configs
	os.RemoveAll(ConfdDir)
	return nil
}

func initMainCfg(dw string) error {

	Cfg.MainCfg.UUID = cliutils.XID("dkid_")
	Cfg.MainCfg.ConfigDir = ConfdDir
	Cfg.MainCfg.FtGateway = dw

	var err error
	tm := template.New("")
	tm, err = tm.Parse(mainConfigTemplate)
	if err != nil {
		return fmt.Errorf("Error creating %s: %s", Cfg.MainCfg.cfgPath, err)
	}

	buf := bytes.NewBuffer([]byte{})
	if err = tm.Execute(buf, Cfg.MainCfg); err != nil {
		return fmt.Errorf("Error creating %s: %s", Cfg.MainCfg.cfgPath, err)
	}

	if err := ioutil.WriteFile(Cfg.MainCfg.cfgPath, []byte(buf.Bytes()), 0664); err != nil {
		return fmt.Errorf("Error creating %s: %s", Cfg.MainCfg.cfgPath, err)
	}

	return nil
}

func createPluginCfgsIfNotExists() {
	// creata datakit input plugin's configures
	for name, create := range inputs.Inputs {
		if name == "self" {
			continue
		}

		input := create()
		catalog := input.Catalog()

		// migrate old config to new catalog path
		oldCfgPath := filepath.Join(ConfdDir, name, name+".conf")
		cfgpath := filepath.Join(ConfdDir, catalog, name+".conf")

		log.Printf("check datakit input conf %s...", name)

		if _, err := os.Stat(oldCfgPath); err == nil {
			if oldCfgPath == cfgpath {
				continue // do nothing
			}

			os.Rename(oldCfgPath, cfgpath)
			os.RemoveAll(filepath.Dir(oldCfgPath))
			continue
		}

		if _, err := os.Stat(cfgpath); err != nil { // file not exists

			log.Printf("D! %s not exists, create it...", cfgpath)

			log.Printf("D! create datakit conf path %s", filepath.Join(ConfdDir, catalog))
			if err := os.MkdirAll(filepath.Join(ConfdDir, catalog), os.ModePerm); err != nil {
				log.Fatalf("create catalog dir %s failed: %s", catalog, err.Error())
			}

			sample := input.SampleConfig()
			if sample == "" {
				log.Fatalf("no sample available on collector %s", name)
			}

			if err := ioutil.WriteFile(cfgpath, []byte(sample), 0644); err != nil {
				log.Fatalf("failed to create sample configure for collector %s: %s", name, err.Error())
			}
		}
	}

	// create telegraf input plugin's configures
	for name, input := range SupportsTelegrafMetraicNames {

		cfgpath := filepath.Join(ConfdDir, input.catalog, name+".conf")
		oldCfgPath := filepath.Join(ConfdDir, name, name+".conf")

		log.Printf("check telegraf input conf %s...", name)

		if _, err := os.Stat(oldCfgPath); err == nil {

			if oldCfgPath == cfgpath {
				log.Printf("D! %s exists, skip", oldCfgPath)
				continue // do nothing
			}

			log.Printf("D! %s exists, migrate to %s", oldCfgPath, cfgpath)
			os.Rename(oldCfgPath, cfgpath)
			os.RemoveAll(filepath.Dir(oldCfgPath))
			continue
		}

		if _, err := os.Stat(cfgpath); err != nil {

			log.Printf("D! %s not exists, create it...", cfgpath)

			log.Printf("D! create telegraf conf path %s", filepath.Join(ConfdDir, input.catalog))
			if err := os.MkdirAll(filepath.Join(ConfdDir, input.catalog), os.ModePerm); err != nil {
				log.Fatalf("create catalog dir %s failed: %s", input.catalog, err.Error())
			}

			if sample, ok := telegrafCfgSamples[name]; ok {
				if err := ioutil.WriteFile(cfgpath, []byte(sample), 0644); err != nil {
					log.Fatalf("failed to create sample configure for collector %s: %s", name, err.Error())
				}
			}
		}
	}
}

func parseConfig(contents []byte) (*ast.Table, error) {
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

	pluginConfig, err := buildInput(name, table, input)
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

func buildInput(name string, tbl *ast.Table, input telegraf.Input) (*models.InputConfig, error) {
	cp := &models.InputConfig{Name: name}

	if _, bsvrInput := input.(telegraf.ServiceInput); !bsvrInput {
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
	}

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
