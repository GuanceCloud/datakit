package config

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"text/template"
	"time"

	"github.com/influxdata/toml"
	"github.com/influxdata/toml/ast"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

var (
	l                             = logger.DefaultSLogger("config")
	Cfg                   *Config = nil
	EnabledTelegrafInputs         = map[string]interface{}{}
)

type Config struct {
	MainCfg          *MainConfig
	TelegrafAgentCfg *TelegrafAgentConfig

	Inputs map[string][]inputs.Input

	InputFilters []string

	withinDocker bool
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
	GinLog   string `toml:"gin_log"`

	ConfigDir string `toml:"config_dir"` // XXX: not used: to compatible parsing with forethought datakit.conf

	//验证dk存活
	MaxPostInterval datakit.Duration `toml:"max_post_interval"`

	//DataCleanTemplate string

	GlobalTags map[string]string `toml:"global_tags"`

	RoundInterval bool
	Interval      datakit.Duration `toml:"interval"`
	flushInterval datakit.Duration
	flushJitter   datakit.Duration

	OutputFile string `toml:"output_file,omitempty"`

	OmitHostname bool // Deprecated

	Hostname string `toml:"hostname"`
	cfgPath  string
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
			GlobalTags:      map[string]string{},
			flushInterval:   datakit.Duration{time.Second * 10},
			Interval:        datakit.Duration{time.Second * 10},
			MaxPostInterval: datakit.Duration{time.Second * 15}, // add 5s plus for network latency

			HTTPServerAddr: "0.0.0.0:9529",

			LogLevel: "info",
			Log:      filepath.Join(datakit.InstallDir, "datakit.log"),
			GinLog:   filepath.Join(datakit.InstallDir, "gin.log"),

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

	if err := Cfg.loadEnvs(); err != nil {
		return err
	}

	// loading from
	if err := Cfg.LoadMainConfig(); err != nil {
		return err
	}

	// set global log root
	logger.SetGlobalRootLogger(Cfg.MainCfg.Log, Cfg.MainCfg.LogLevel, logger.OPT_DEFAULT)
	l = logger.SLogger("config")

	l.Infof("set log to %s", Cfg.MainCfg.Log)
	l.Infof("main cfg: %+#v", Cfg.MainCfg)

	if Cfg.MainCfg.MaxPostInterval.Duration > 0 {
		datakit.MaxLifeCheckInterval = Cfg.MainCfg.MaxPostInterval.Duration
	}

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

	if val, ok := tbl.Fields["agent"]; ok {
		subTable, ok := val.(*ast.Table)
		if !ok {
			return fmt.Errorf("invalid agent configuration")
		}

		if _, ok := subTable.Fields["debug"]; ok {
			bAgentSetLogLevel = true
		}

		if err = toml.UnmarshalTable(subTable, c.TelegrafAgentCfg); err != nil {
			return fmt.Errorf("invalid telegraf configuration, %s", err)
		}

		delete(tbl.Fields, "agent")
	}

	if err := toml.UnmarshalTable(tbl, c.MainCfg); err != nil {
		l.Errorf("UnmarshalTable failed: " + err.Error())
		return err
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

	c.MainCfg.DataWayRequestURL = fmt.Sprintf("%s://%s%s?token=%s",
		c.MainCfg.DataWay.Scheme, c.MainCfg.DataWay.Host, c.MainCfg.DataWay.DefaultPath, c.MainCfg.DataWay.Token)

	// reset global tags
	for k, v := range c.MainCfg.GlobalTags {
		switch strings.ToLower(v) {
		case `$datakit_hostname`:
			c.MainCfg.GlobalTags[k] = c.MainCfg.Hostname
			l.Debugf("set global tag %s: %s", k, c.MainCfg.Hostname)

		case `$datakit_ip`:
			ip := "unavailable"
			ip, err = datakit.LocalIP()
			if err != nil {
				l.Errorf("get local ip failed: %s", err.Error())
			}
			l.Debugf("set global tag %s: %s", k, ip)
			c.MainCfg.GlobalTags[k] = ip

		case `$datakit_uuid`, `$datakit_id`:
			c.MainCfg.GlobalTags[k] = c.MainCfg.UUID
			l.Debugf("set global tag %s: %s", k, c.MainCfg.UUID)
		default:
			// pass
		}
	}

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
				return fmt.Errorf("no input configured")
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

func DumpInputsOutputs() {
	names := []string{}

	for name := range Cfg.Inputs {
		names = append(names, name)
	}

	for k, i := range TelegrafInputs {
		if i.enabled {
			names = append(names, k)
			EnabledTelegrafInputs[k] = nil
		}
	}

	l.Infof("available inputs: %s", strings.Join(names, ","))
}

func buildMainCfgFile() error {
	var err error
	t := template.New("")
	t, err = t.Parse(MainConfigTemplate)
	if err != nil {
		return fmt.Errorf("Error creating %s: %s", Cfg.MainCfg.cfgPath, err)
	}

	buf := bytes.NewBuffer([]byte{})
	if err = t.Execute(buf, Cfg.MainCfg); err != nil {
		return fmt.Errorf("Error creating %s: %s", Cfg.MainCfg.cfgPath, err)
	}

	if err := ioutil.WriteFile(Cfg.MainCfg.cfgPath, []byte(buf.Bytes()), 0664); err != nil {
		return fmt.Errorf("error creating %s: %s", Cfg.MainCfg.cfgPath, err)
	}

	return nil
}

func InitCfg() error {
	if err := buildMainCfgFile(); err != nil {
		return err
	}

	// clean all old dirs
	os.RemoveAll(datakit.ConfdDir)
	os.RemoveAll(datakit.DataDir)
	os.RemoveAll(datakit.LuaDir)
	return nil
}

/*
func initMainCfg(dwcfg *DataWayCfg, tags map[string]string) error {

	Cfg.MainCfg.UUID = cliutils.XID("dkid_")
	Cfg.MainCfg.DataWay = dwcfg
	Cfg.MainCfg.GlobalTags = tags

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
		return fmt.Errorf("error creating %s: %s", Cfg.MainCfg.cfgPath, err)
	}

	return nil
} */

func parseCfgFile(f string) (*ast.Table, error) {
	data, err := ioutil.ReadFile(f)
	if err != nil {
		l.Error(err)
		return nil, fmt.Errorf("read config %s failed: %s", f, err.Error())
	}

	tbl, err := toml.Parse(data)
	if err != nil {
		l.Errorf("parse toml %s failed", string(data))
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
					l.Errorf("parse duration(%s) from %s failed: %s", str.Value, name, err.Error())
					return err
				}
			}
		}
	}

	l.Debugf("try set MaxLifeCheckInterval to %v from %s...", dur, name)
	if datakit.MaxLifeCheckInterval+5*time.Second < dur { // use the max interval from all inputs
		datakit.MaxLifeCheckInterval = dur
		l.Debugf("set MaxLifeCheckInterval to %v from %s", dur, name)
	}

	c.Inputs[name] = append(c.Inputs[name], input)

	return nil
}

func ParseDataway(dw string) (*DataWayCfg, error) {

	dwcfg := &DataWayCfg{}

	if u, err := url.Parse(dw); err == nil {
		dwcfg.Scheme = u.Scheme
		dwcfg.Token = u.Query().Get("token")
		dwcfg.Host = u.Host
		dwcfg.DefaultPath = u.Path

		if dwcfg.Scheme == "https" {
			dwcfg.Host = dwcfg.Host + ":443"
		}

		l.Debugf("dataway: %+#v", dwcfg)

	} else {
		l.Errorf("parse url %s failed: %s", dw, err.Error())
		return nil, err
	}

	conn, err := net.DialTimeout("tcp", dwcfg.Host, time.Second*5)
	if err != nil {
		l.Errorf("TCP dial host `%s' failed: %s", dwcfg.Host, err.Error())
		return nil, err
	}

	if err := conn.Close(); err != nil {
		l.Errorf("close failed: %s", err.Error())
		return nil, err
	}

	return dwcfg, nil
}
