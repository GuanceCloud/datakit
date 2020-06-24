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
	"github.com/influxdata/toml"
	"github.com/influxdata/toml/ast"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/git"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/models"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

var (
	Exit *cliutils.Sem

	DKUserAgent = fmt.Sprintf("datakit(%s), %s-%s", git.Version, runtime.GOOS, runtime.GOARCH)

	ServiceName = "datakit"

	ConvertedCfg []string
	AgentLogFile string

	MaxLifeCheckInterval time.Duration

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
	Inputs           []*models.RunningInput
	//Outputs          []*models.RunningOutput

	InputFilters []string
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
		log.Fatalf("[fatal] invalid os/arch: %s", osarch)
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
		Inputs: []*models.RunningInput{},
	}
}

func InitDirs() {
	if err := os.MkdirAll(filepath.Join(InstallDir, "embed"), os.ModePerm); err != nil {
		log.Fatalf("[error] mkdir embed failed: %s", err)
	}

	for _, dir := range []string{TelegrafDir, DataDir, LuaDir, ConfdDir} {
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			log.Fatalf("create %s failed: %s", dir, err)
		}
	}
}

func LoadCfg() error {

	if err := Cfg.LoadMainConfig(); err != nil {
		return err
	}

	// some old log file under:
	//  /usr/local/cloudcare/forethought/...
	// we force set to
	//  /usr/local/cloudcare/DataFlux/...
	//Cfg.MainCfg.Log = filepath.Join(InstallDir, "datakit.log")

	logConfig := logger.LogConfig{
		Debug:     (strings.ToLower(Cfg.MainCfg.LogLevel) == "debug"),
		Quiet:     false,
		LogTarget: logger.LogTargetFile,
		Logfile:   Cfg.MainCfg.Log,
	}

	logConfig.RotationMaxSize.Size = (20 << 10 << 10)
	logger.SetupLogging(logConfig)

	log.SetFlags(log.Llongfile)

	log.Printf("D! set log to %s", logConfig.Logfile)

	createPluginCfgsIfNotExists()

	if err := Cfg.LoadConfig(); err != nil {
		return err
	}

	Cfg.DumpInputsOutputs()

	return nil
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

	FtGateway string `toml:"ftdataway"` // deprecated

	Log      string `toml:"log"`
	LogLevel string `toml:"log_level"`

	ConfigDir string `toml:"config_dir"` // not used: to compatible parsing with forethought datakit.conf

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

		input := creator()

		var data []byte

		if internalData, ok := inputs.InternalInputsData[name]; ok {
			data = internalData
		} else {
			oldPath := filepath.Join(ConfdDir, name, fmt.Sprintf("%s.conf", name))
			newPath := filepath.Join(ConfdDir, input.Catalog(), fmt.Sprintf("%s.conf", name))

			path := newPath
			_, err := os.Stat(path)
			if err != nil && os.IsNotExist(err) {
				if _, err = os.Stat(oldPath); err == nil {
					path = oldPath
				}
			}

			data, err = ioutil.ReadFile(path)
			if err != nil {
				log.Printf("[error] load %s failed: %s", path, err)
				return err
			}
		}

		tbl, err := parseConfig(data)
		if err != nil {
			log.Printf("[error] parse failed: %s", err)
			return err
		}

		if len(tbl.Fields) == 0 {
			continue
		}

		if err := c.addInput(name, input, tbl); err != nil {
			return err
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

	return LoadTelegrafConfigs(ConfdDir, c.InputFilters)
}

func (c *Config) DumpInputsOutputs() {
	names := []string{}

	for _, p := range c.Inputs {
		log.Printf("input %s enabled", p.Config.Name)
		names = append(names, p.Config.Name)
	}

	for k, i := range SupportsTelegrafMetricNames {
		if i.enabled {
			log.Printf("telegraf input %s enabled", k)
			names = append(names, k)
		}
	}

	log.Printf("avariable inputs: %s", strings.Join(names, ","))
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

		log.Printf("I! check datakit input conf %s: %s, %s", name, oldCfgPath, cfgpath)

		if _, err := os.Stat(oldCfgPath); err == nil {
			if oldCfgPath == cfgpath {
				continue // do nothing
			}

			log.Printf("I! migrate %s: %s -> %s", name, oldCfgPath, cfgpath)

			if err := os.MkdirAll(filepath.Dir(cfgpath), os.ModePerm); err != nil {
				log.Fatalf("E! create dir %s failed: %s", filepath.Dir(cfgpath), err.Error())
			}

			if err := os.Rename(oldCfgPath, cfgpath); err != nil {
				log.Fatalf("E! move %s -> %s failed: %s", oldCfgPath, cfgpath, err.Error())
			}

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
	for name, input := range SupportsTelegrafMetricNames {

		cfgpath := filepath.Join(ConfdDir, input.Catalog, name+".conf")
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

			log.Printf("D! create telegraf conf path %s", filepath.Join(ConfdDir, input.Catalog))
			if err := os.MkdirAll(filepath.Join(ConfdDir, input.Catalog), os.ModePerm); err != nil {
				log.Fatalf("create catalog dir %s failed: %s", input.Catalog, err.Error())
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
			delete(tbl.Fields, "interval")
		}
	}

	if node, ok := tbl.Fields["dataway_path"]; ok {
		if kv, ok := node.(*ast.KeyValue); ok {
			if str, ok := kv.Value.(*ast.String); ok {

				log.Printf("D! dataway_path: %s", str.Value)

				cp.DataWayRequestURL = fmt.Sprintf("%s://%s%s?token=%s",
					Cfg.MainCfg.DataWay.Scheme, Cfg.MainCfg.DataWay.Host, str.Value, Cfg.MainCfg.DataWay.Token)
			}
		}
	}

	if node, ok := tbl.Fields["output_file"]; ok {
		if kv, ok := node.(*ast.KeyValue); ok {
			if str, ok := kv.Value.(*ast.String); ok {
				cp.OutputFile = str.Value
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

	delete(tbl.Fields, "dataway_path")
	delete(tbl.Fields, "output_file")
	delete(tbl.Fields, "tags")

	return cp, nil
}
