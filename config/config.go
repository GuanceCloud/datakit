package config

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"text/template"
	"time"

	"github.com/influxdata/toml"
	"github.com/influxdata/toml/ast"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

var (
	ConvertedCfg []string

	l   *logger.Logger
	Cfg *Config = nil
)

type Config struct {
	MainCfg          *MainConfig
	TelegrafAgentCfg *TelegrafAgentConfig

	Inputs map[string][]inputs.Input

	InputFilters []string
}

type DataWayCfg struct {
	Host        string `toml:"host"`
	Scheme      string `toml:"scheme"`
	Token       string `toml:"token"`
	DefaultPath string `toml:"default_path"`
}

type MainConfig struct {
	UUID string `toml:"uuid"`

	DataWay           *DataWayCfg `toml:"dataway"`
	DataWayRequestURL string      `toml:"-"`

	HTTPServerAddr string `toml:"http_server_addr"`

	FtGateway string `toml:"ftdataway"` // XXX: deprecated

	Log      string `toml:"log"`
	LogLevel string `toml:"log_level"`

	ConfigDir string `toml:"config_dir"` // XXX: not used: to compatible parsing with forethought datakit.conf

	//验证dk存活
	MaxPostInterval internal.Duration `toml:"max_post_interval"`

	//DataCleanTemplate string

	GlobalTags map[string]string `toml:"global_tags"`

	Interval      internal.Duration `toml:"interval"`
	RoundInterval bool
	FlushInterval internal.Duration
	FlushJitter   internal.Duration

	OutputFile string `toml:"output_file,omitempty"`

	Hostname     string
	OmitHostname bool

	cfgPath string
}

func init() {
	osarch := runtime.GOOS + "/" + runtime.GOARCH

	switch osarch {
	case "windows/amd64":
		datakit.InstallDir = `C:\Program Files\dataflux\` + datakit.ServiceName

	case "windows/386":
		datakit.InstallDir = `C:\Program Files (x86)\dataflux\` + datakit.ServiceName

	case "linux/amd64", "linux/386", "linux/arm", "linux/arm64",
		"darwin/amd64", "darwin/386",
		"freebsd/amd64", "freebsd/386":
		datakit.InstallDir = `/usr/local/cloudcare/dataflux/` + datakit.ServiceName
	default:
		panic("unsupported os/arch: %s" + osarch)
	}

	datakit.AgentLogFile = filepath.Join(datakit.InstallDir, "embed", "agent.log")

	datakit.TelegrafDir = filepath.Join(datakit.InstallDir, "embed")
	datakit.DataDir = filepath.Join(datakit.InstallDir, "data")
	datakit.LuaDir = filepath.Join(datakit.InstallDir, "lua")
	datakit.ConfdDir = filepath.Join(datakit.InstallDir, "conf.d")
	datakit.GRPCDomainSock = filepath.Join(datakit.InstallDir, "datakit.sock")

	datakit.Exit = cliutils.NewSem()
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

			HTTPServerAddr: "0.0.0.0:9529",

			LogLevel: "info",
			Log:      filepath.Join(datakit.InstallDir, "datakit.log"),

			RoundInterval: false,
			cfgPath:       filepath.Join(datakit.InstallDir, "datakit.conf"),
		},
		Inputs: map[string][]inputs.Input{},
	}
}

func InitDirs() {
	if err := os.MkdirAll(filepath.Join(datakit.InstallDir, "embed"), os.ModePerm); err != nil {
		panic("[error] mkdir embed failed: " + err.Error())
	}

	for _, dir := range []string{datakit.TelegrafDir, datakit.DataDir, datakit.LuaDir, datakit.ConfdDir} {
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			panic(fmt.Sprintf("create %s failed: %s", dir, err))
		}
	}
}

func LoadCfg() error {

	InitDirs()

	if err := Cfg.LoadMainConfig(); err != nil {
		return err
	}

	// set global log root
	logger.SetGlobalRootLogger(Cfg.MainCfg.Log,
		Cfg.MainCfg.LogLevel,
		logger.OPT_ENC_CONSOLE|logger.OPT_SHORT_CALLER)
	l = logger.SLogger("config")

	l.Infof("set log to %s", Cfg.MainCfg.Log)

	datakit.Init()

	initPluginCfgs()

	if err := Cfg.LoadConfig(); err != nil {
		l.Error(err)
		return err
	}

	DumpInputsOutputs()

	return nil
}

func (c *Config) LoadMainConfig() error {

	tbl, err := parseCfgFile(c.MainCfg.cfgPath)
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
			return fmt.Errorf("invalid telegraf configuration, %s", err)
		}

		delete(tbl.Fields, "agent")
	}

	if err := toml.UnmarshalTable(tbl, c.MainCfg); err != nil {
		panic("UnmarshalTable failed: " + err.Error())
	}

	if !c.MainCfg.OmitHostname { // get default host-name
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
		c.TelegrafAgentCfg.Logfile = filepath.Join(datakit.InstallDir, "embed", "agent.log")
	}

	if datakit.AgentLogFile != "" {
		c.TelegrafAgentCfg.Logfile = datakit.AgentLogFile
	}

	if c.MainCfg.OutputFile != "" {
		datakit.OutputFile = c.MainCfg.OutputFile
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

	c.MainCfg.DataWayRequestURL = fmt.Sprintf("%s://%s%s?token=%s",
		c.MainCfg.DataWay.Scheme, c.MainCfg.DataWay.Host, c.MainCfg.DataWay.DefaultPath, c.MainCfg.DataWay.Token)

	return nil
}

func CheckConfd() error {
	dir, err := ioutil.ReadDir(datakit.ConfdDir)
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

		checkSubDir(filepath.Join(datakit.ConfdDir, item.Name()))
	}

	fmt.Printf("inputs: %s\n", strings.Join(configed, ","))
	fmt.Printf("error configuration: %s\n", strings.Join(invalids, ","))

	return nil
}

func (c *Config) searchInput(files []os.FileInfo, name, dir string, creator inputs.Creator) error {
	for _, f := range files {
		fname := filepath.Join(dir, f.Name())
		if f.IsDir() {
			l.Debugf("ignore dir %s", fname)
			continue
		}

		if strings.HasSuffix(f.Name(), ".sample") {
			l.Debugf("ignore sample %s", fname)
			continue
		}

		// parse any text files
		tbl, err := parseCfgFile(fname)
		if err != nil {
			l.Warnf("[error] parse conf %s failed on [%s]: %s, ignored", fname, name, err)
		}

		if len(tbl.Fields) == 0 {
			l.Debugf("no conf available on %s", name)
			continue
		}

		for f, val := range tbl.Fields {
			switch f {
			case "inputs":
				tbl_, ok := val.(*ast.Table)
				if !ok {
					l.Warnf("ignore bad toml node")
				} else {
					for inputName, v := range tbl_.Fields {
						if inputName != name {
							l.Debugf("input %s ignore input %s", name, inputName)
							continue
						}

						if err := c.tryUnmarshal(v, name, creator); err != nil {
							l.Error(err)
							return err
						}

						l.Infof("load input %s from %s ok", name, fname)
					}
				}

			default:
				if err := c.tryUnmarshal(val, name, creator); err != nil {
					l.Errorf("unmarshal %s failed: %s", fname, err)
					return err
				}
				l.Infof("load input %s from %s ok", name, fname)
			}
		}
	}

	return nil
}

// search all inputs.@name under catalog dir
func (c *Config) doLoadInputConf(name string, creator inputs.Creator) error {
	if len(c.InputFilters) > 0 {
		if !sliceContains(name, c.InputFilters) {
			return nil
		}
	}

	if name == "self" {
		c.Inputs[name] = append(c.Inputs[name], creator())
		return nil
	}

	dummyInput := creator()
	catalogdir := filepath.Join(datakit.ConfdDir, dummyInput.Catalog())
	allcfgs, err := ioutil.ReadDir(catalogdir)
	if err != nil {
		l.Errorf("ReadDir(%s) failed: %s", name, err.Error())
		return err
	}

	if err := c.searchInput(allcfgs, name, catalogdir, creator); err != nil {
		return err
	}

	return nil
}

func (c *Config) tryUnmarshal(tbl interface{}, name string, creator inputs.Creator) error {

	tbls := []*ast.Table{}

	switch tbl.(type) {
	case []*ast.Table:
		tbls = tbl.([]*ast.Table)
	case *ast.Table:
		tbls = append(tbls, tbl.(*ast.Table))
	default:
		return fmt.Errorf("invalid toml format on %s: %v", name, reflect.TypeOf(tbl))
	}

	for _, t := range tbls {
		input := creator()

		if err := toml.UnmarshalTable(t, input); err != nil {
			l.Errorf("toml unmarshal %s failed: %v", name, err)
			return err
		}

		if err := c.addInput(name, input, t); err != nil {
			l.Error("add %s failed: %v", name, err)
			return err
		}
	}

	return nil
}

// load all inputs under @InstallDir/conf.d
func (c *Config) LoadConfig() error {

	for name, creator := range inputs.Inputs {
		if err := c.doLoadInputConf(name, creator); err != nil {
			l.Errorf("load %s config failed: %v, ignored", name, err)
			return err
		}
	}

	return LoadTelegrafConfigs(datakit.ConfdDir, c.InputFilters)
}

func DumpInputsOutputs() {
	names := []string{}

	for name := range Cfg.Inputs {
		names = append(names, name)
	}

	for k, i := range SupportsTelegrafMetricNames {
		if i.enabled {
			names = append(names, k)
		}
	}

	l.Infof("avariable inputs: %s", strings.Join(names, ","))
}

func InitCfg(dwcfg *DataWayCfg) error {
	if err := initMainCfg(dwcfg); err != nil {
		return err
	}

	// clean all old dirs
	os.RemoveAll(datakit.ConfdDir)
	os.RemoveAll(datakit.DataDir)
	os.RemoveAll(datakit.LuaDir)
	return nil
}

func initMainCfg(dwcfg *DataWayCfg) error {

	Cfg.MainCfg.UUID = cliutils.XID("dkid_")
	Cfg.MainCfg.DataWay = dwcfg

	var err error
	tm := template.New("")
	tm, err = tm.Parse(MainConfigTemplate)
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

// Creata datakit input plugin's configures if not exists
func initPluginCfgs() {
	for name, create := range inputs.Inputs {
		if name == "self" {
			continue
		}

		input := create()
		catalog := input.Catalog()

		cfgpath := filepath.Join(datakit.ConfdDir, catalog, name+".conf.sample")
		old := filepath.Join(datakit.ConfdDir, catalog, name+".conf")

		if _, err := os.Stat(old); err == nil {
			tbl, err := parseCfgFile(old)
			if err != nil {
				l.Warnf("[error] parse conf %s failed on [%s]: %s, ignored", old, name, err)
			} else {
				if len(tbl.Fields) == 0 { // old config not used
					os.Remove(old)
				}
			}
		}

		// overwrite old config sample
		l.Debugf("create datakit conf path %s", filepath.Join(datakit.ConfdDir, catalog))
		if err := os.MkdirAll(filepath.Join(datakit.ConfdDir, catalog), os.ModePerm); err != nil {
			l.Fatalf("create catalog dir %s failed: %s", catalog, err.Error())
		}

		sample := input.SampleConfig()
		if sample == "" {
			l.Fatalf("no sample available on collector %s", name)
		}

		if err := ioutil.WriteFile(cfgpath, []byte(sample), 0644); err != nil {
			l.Fatalf("failed to create sample configure for collector %s: %s", name, err.Error())
		}
	}

	// create telegraf input plugin's configures
	for name, input := range SupportsTelegrafMetricNames {

		cfgpath := filepath.Join(datakit.ConfdDir, input.Catalog, name+".conf.sample")
		old := filepath.Join(datakit.ConfdDir, input.Catalog, name+".conf")

		if _, err := os.Stat(old); err == nil {
			tbl, err := parseCfgFile(old)
			if err != nil {
				l.Warnf("[error] parse conf %s failed on [%s]: %s, ignored", old, name, err)
			} else {
				if len(tbl.Fields) == 0 { // old config not used
					os.Remove(old)
				}
			}
		}

		// overwrite old telegraf config sample
		l.Debugf("create telegraf conf path %s", filepath.Join(datakit.ConfdDir, input.Catalog))
		if err := os.MkdirAll(filepath.Join(datakit.ConfdDir, input.Catalog), os.ModePerm); err != nil {
			l.Fatalf("create catalog dir %s failed: %s", input.Catalog, err.Error())
		}

		if sample, ok := TelegrafCfgSamples[name]; ok {
			if err := ioutil.WriteFile(cfgpath, []byte(sample), 0644); err != nil {
				l.Fatalf("failed to create sample configure for collector %s: %s", name, err.Error())
			}
		}
	}
}

func parseCfgFile(f string) (*ast.Table, error) {
	data, err := ioutil.ReadFile(f)
	if err != nil {
		return nil, fmt.Errorf("read config %s failed: %s", f, err.Error())
	}

	tbl, err := toml.Parse(data)
	if err != nil {
		return nil, err
	}

	return tbl, nil
}

func sliceContains(name string, list []string) bool {
	for _, b := range list {
		if b == name {
			return true
		}
	}
	return false
}

func (c *Config) addInput(name string, input inputs.Input, table *ast.Table) error {

	var dur time.Duration
	var err error
	if node, ok := table.Fields["interval"]; ok {
		if kv, ok := node.(*ast.KeyValue); ok {
			if str, ok := kv.Value.(*ast.String); ok {
				dur, err = time.ParseDuration(str.Value)
				if err != nil {
					return err
				}
			}
		}
	}

	if c.MainCfg.MaxPostInterval.Duration != 0 && datakit.MaxLifeCheckInterval < dur {
		datakit.MaxLifeCheckInterval = dur
	}

	c.Inputs[name] = append(c.Inputs[name], input)

	return nil
}
