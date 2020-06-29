package config

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"text/template"
	"time"

	"github.com/influxdata/toml"
	"github.com/influxdata/toml/ast"
	"go.uber.org/zap"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/git"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

var (
	Exit *cliutils.Sem
	WG   sync.WaitGroup = sync.WaitGroup{}

	DKUserAgent = fmt.Sprintf("datakit(%s), %s-%s", git.Version, runtime.GOOS, runtime.GOARCH)

	ServiceName = "datakit"

	ConvertedCfg []string
	AgentLogFile string

	MaxLifeCheckInterval time.Duration

	l *zap.SugaredLogger

	Cfg            *Config = nil
	InstallDir             = ""
	TelegrafDir            = ""
	DataDir                = ""
	LuaDir                 = ""
	ConfdDir               = ""
	GRPCDomainSock         = ""
)

type Config struct {
	MainCfg          *MainConfig
	TelegrafAgentCfg *TelegrafAgentConfig

	Inputs []*inputs.RunningInput

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

	OutputsFile string `toml:"output_file,omitempty"`

	Hostname     string
	OmitHostname bool

	cfgPath string
}

func init() {
	osarch := runtime.GOOS + "/" + runtime.GOARCH

	switch osarch {
	case "windows/amd64":
		InstallDir = `C:\Program Files\DataFlux\` + ServiceName

	case "windows/386":
		InstallDir = `C:\Program Files (x86)\DataFlux\` + ServiceName

	case "linux/amd64", "linux/386", "linux/arm", "linux/arm64",
		"darwin/amd64", "darwin/386",
		"freebsd/amd64", "freebsd/386":
		InstallDir = `/usr/local/cloudcare/DataFlux/` + ServiceName
	default:
		panic("unsupported os/arch: %s" + osarch)
	}

	AgentLogFile = filepath.Join(InstallDir, "embed", "agent.log")

	TelegrafDir = filepath.Join(InstallDir, "embed")
	DataDir = filepath.Join(InstallDir, "data")
	LuaDir = filepath.Join(InstallDir, "lua")
	ConfdDir = filepath.Join(InstallDir, "conf.d")
	GRPCDomainSock = filepath.Join(InstallDir, "datakit.sock")

	Exit = cliutils.NewSem()

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
		Inputs: []*inputs.RunningInput{},
	}
}

func InitDirs() {
	if err := os.MkdirAll(filepath.Join(InstallDir, "embed"), os.ModePerm); err != nil {
		panic("[error] mkdir embed failed: " + err.Error())
	}

	for _, dir := range []string{TelegrafDir, DataDir, LuaDir, ConfdDir} {
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			panic(fmt.Sprintf("create %s failed: %s", dir, err))
		}
	}
}

func LoadCfg() error {

	if err := Cfg.LoadMainConfig(); err != nil {
		return err
	}

	// set global log root
	logger.SetGlobalRootLogger(Cfg.MainCfg.Log,
		Cfg.MainCfg.LogLevel,
		logger.OPT_ENC_CONSOLE|logger.OPT_SHORT_CALLER)
	l = logger.SLogger("config")

	l.Debugf("set log to %s", Cfg.MainCfg.Log)

	initPluginCfgs()

	if err := Cfg.LoadConfig(); err != nil {
		l.Error(err)
		return err
	}

	DumpInputsOutputs()

	return nil
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
			return fmt.Errorf("invalid telegraf configuration, %s", err)
		}

		delete(tbl.Fields, "agent")
	}

	if err := toml.UnmarshalTable(tbl, c.MainCfg); err != nil {
		l.Error("UnmarshalTable failed: %s", err)
		return err
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

	c.MainCfg.DataWayRequestURL = fmt.Sprintf("%s://%s%s?token=%s",
		c.MainCfg.DataWay.Scheme, c.MainCfg.DataWay.Host, c.MainCfg.DataWay.DefaultPath, c.MainCfg.DataWay.Token)

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

	l.Infof("inputs: %s", strings.Join(configed, ","))
	l.Infof("error configuration: %s", strings.Join(invalids, ","))

	return nil
}

func (c *Config) doLoadInputConf(name string, creator inputs.Creator) error {
	if len(c.InputFilters) > 0 {
		if !sliceContains(name, c.InputFilters) {
			return nil
		}
	}

	input := creator()

	var data []byte

	if internalData, ok := inputs.InternalInputsData[name]; ok {
		data = internalData
	} else {
		// migrate old configures into new place
		oldPath := filepath.Join(ConfdDir, name, fmt.Sprintf("%s.conf", name))
		newPath := filepath.Join(ConfdDir, input.Catalog(), fmt.Sprintf("%s.conf", name))

		path := newPath
		_, err := os.Stat(path)
		if err != nil && os.IsNotExist(err) {
			if _, err = os.Stat(oldPath); err == nil {
				l.Infof("migrate %s to %s", oldPath, path)
				path = oldPath
			}
		}

		data, err = ioutil.ReadFile(path)
		if err != nil {
			l.Errorf("load %s failed: %s", path, err)
			return err
		}
	}

	tbl, err := parseConfig(data)
	if err != nil {
		l.Errorf("[error] parse conf failed on %s: %s", name, err)
		return err
	}

	if len(tbl.Fields) == 0 {
		l.Debugf("no conf available on %s", name)
		return nil
	}

	if err := c.addInput(name, input, tbl); err != nil {
		l.Errorf("add input %s failed: %s", name, err.Error())
		return err
	}

	l.Debugf("add input %s ok", name)

	return nil
}

func (c *Config) setMaxPostInterval() {
	MaxLifeCheckInterval = c.MainCfg.MaxPostInterval.Duration
	if MaxLifeCheckInterval != 0 {
		return
	}

	for _, ri := range c.Inputs { // find the max interval
		if ri.Config.Interval > MaxLifeCheckInterval {
			MaxLifeCheckInterval = ri.Config.Interval
		}
	}

	MaxLifeCheckInterval += time.Second * 10 // add extra 10 seconds
}

// load all inputs under @InstallDir/conf.d
func (c *Config) LoadConfig() error {

	for name, creator := range inputs.Inputs {
		if err := c.doLoadInputConf(name, creator); err != nil {
			return err
		}
	}

	c.setMaxPostInterval()

	return LoadTelegrafConfigs(ConfdDir, c.InputFilters)
}

func DumpInputsOutputs() {
	names := []string{}

	for _, p := range Cfg.Inputs {
		l.Debugf("input %s enabled", p.Config.Name)
		names = append(names, p.Config.Name)
	}

	for k, i := range SupportsTelegrafMetricNames {
		if i.enabled {
			l.Debugf("telegraf input %s enabled", k)
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
	os.RemoveAll(ConfdDir)
	os.RemoveAll(DataDir)
	os.RemoveAll(LuaDir)
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

		// migrate old config to new catalog path
		oldCfgPath := filepath.Join(ConfdDir, name, name+".conf")
		cfgpath := filepath.Join(ConfdDir, catalog, name+".conf")

		l.Infof("check datakit input conf %s: %s, %s", name, oldCfgPath, cfgpath)

		if _, err := os.Stat(oldCfgPath); err == nil {
			if oldCfgPath == cfgpath {
				continue // do nothing
			}

			l.Infof("migrate %s: %s -> %s", name, oldCfgPath, cfgpath)

			if err := os.MkdirAll(filepath.Dir(cfgpath), os.ModePerm); err != nil {
				l.Fatalf("create dir %s failed: %s", filepath.Dir(cfgpath), err.Error())
			}

			if err := os.Rename(oldCfgPath, cfgpath); err != nil {
				l.Fatalf("move %s -> %s failed: %s", oldCfgPath, cfgpath, err.Error())
			}

			os.RemoveAll(filepath.Dir(oldCfgPath))
			continue
		}

		if _, err := os.Stat(cfgpath); err != nil { // file not exists

			l.Debugf("%s not exists, create it...", cfgpath)

			l.Debugf("create datakit conf path %s", filepath.Join(ConfdDir, catalog))
			if err := os.MkdirAll(filepath.Join(ConfdDir, catalog), os.ModePerm); err != nil {
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
	}

	// create telegraf input plugin's configures
	for name, input := range SupportsTelegrafMetricNames {

		cfgpath := filepath.Join(ConfdDir, input.Catalog, name+".conf")
		oldCfgPath := filepath.Join(ConfdDir, name, name+".conf")

		l.Debugf("check telegraf input conf %s...", name)

		if _, err := os.Stat(oldCfgPath); err == nil {

			if oldCfgPath == cfgpath {
				l.Debugf("%s exists, skip", oldCfgPath)
				continue // do nothing
			}

			l.Debugf("%s exists, migrate to %s", oldCfgPath, cfgpath)
			os.Rename(oldCfgPath, cfgpath)
			os.RemoveAll(filepath.Dir(oldCfgPath))
			continue
		}

		if _, err := os.Stat(cfgpath); err != nil {

			l.Debugf("%s not exists, create it...", cfgpath)

			l.Debugf("create telegraf conf path %s", filepath.Join(ConfdDir, input.Catalog))
			if err := os.MkdirAll(filepath.Join(ConfdDir, input.Catalog), os.ModePerm); err != nil {
				l.Fatalf("create catalog dir %s failed: %s", input.Catalog, err.Error())
			}

			if sample, ok := telegrafCfgSamples[name]; ok {
				if err := ioutil.WriteFile(cfgpath, []byte(sample), 0644); err != nil {
					l.Fatalf("failed to create sample configure for collector %s: %s", name, err.Error())
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

func (c *Config) addInput(name string, input inputs.Input, table *ast.Table) error {

	if len(c.InputFilters) > 0 && !sliceContains(name, c.InputFilters) {
		return nil
	}

	pluginConfig, err := buildInput(name, table, input)
	if err != nil {
		return err
	}

	if err := toml.UnmarshalTable(table, input); err != nil {
		l.Errorf("toml UnmarshalTable failed on %s: %s", name, err.Error())
		return err
	}

	l.Debugf("configured input: %+#v", input)

	ri := inputs.NewRunningInput(input, pluginConfig)
	ri.SetDefaultTags(c.MainCfg.GlobalTags)

	c.Inputs = append(c.Inputs, ri)
	return nil
}

func buildInput(name string, tbl *ast.Table, input inputs.Input) (*inputs.InputConfig, error) {
	ic := &inputs.InputConfig{Name: name}

	if node, ok := tbl.Fields["interval"]; ok {
		if kv, ok := node.(*ast.KeyValue); ok {
			if str, ok := kv.Value.(*ast.String); ok {
				dur, err := time.ParseDuration(str.Value)
				if err != nil {
					return nil, err
				}

				ic.Interval = dur
			}
		}
		delete(tbl.Fields, "interval")
	}

	if node, ok := tbl.Fields["output_file"]; ok {
		if kv, ok := node.(*ast.KeyValue); ok {
			if str, ok := kv.Value.(*ast.String); ok {
				ic.OutputFile = str.Value
			}
		}
	}

	ic.Tags = make(map[string]string)
	if node, ok := tbl.Fields["tags"]; ok {
		if subtbl, ok := node.(*ast.Table); ok {
			if err := toml.UnmarshalTable(subtbl, ic.Tags); err != nil {
				l.Errorf("could not parse tags for input %s", name)
			}
		}
	}

	delete(tbl.Fields, "dataway_path")
	delete(tbl.Fields, "output_file")
	delete(tbl.Fields, "tags")

	return ic, nil
}
