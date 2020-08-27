//nolint:gocyclo
package config

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/influxdata/toml"
	"github.com/influxdata/toml/ast"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
)

var (
	l = logger.DefaultSLogger("config")

	Cfg = &Config{ //nolint:dupl
		MainCfg: &MainConfig{
			GlobalTags:       map[string]string{},
			flushInterval:    datakit.Duration{Duration: time.Second * 10},
			Interval:         "10s",
			MaxPostInterval:  "15s", // add 5s plus for network latency
			IntervalDuration: 10 * time.Second,

			HTTPBind: "0.0.0.0:9529",

			LogLevel: "info",
			Log:      filepath.Join(datakit.InstallDir, "datakit.log"),
			GinLog:   filepath.Join(datakit.InstallDir, "gin.log"),

			RoundInterval: false,
			cfgPath:       filepath.Join(datakit.InstallDir, "datakit.conf"),
			TelegrafAgentCfg: &agent{
				Interval:                   "10s",
				RoundInterval:              true,
				MetricBatchSize:            1000,
				MetricBufferLimit:          100000,
				CollectionJitter:           "0s",
				FlushInterval:              "10s",
				FlushJitter:                "0s",
				Precision:                  "ns",
				Debug:                      false,
				Quiet:                      false,
				LogTarget:                  "file",
				Logfile:                    filepath.Join(datakit.TelegrafDir, "agent.log"),
				LogfileRotationMaxArchives: 5,
				LogfileRotationMaxSize:     "32MB",
				OmitHostname:               true, // do not append host tag
			},
		},
	}
)

type Config struct {
	MainCfg      *MainConfig
	InputFilters []string
}

type DataWayCfg struct {
	Host        string `toml:"host"`
	Scheme      string `toml:"scheme"`
	Token       string `toml:"token"`
	Timeout     string `toml:"timeout"`
	DefaultPath string `toml:"default_path"`
}

type MainConfig struct {
	UUID string `toml:"uuid"`
	Name string `toml:"name"`

	DataWay           *DataWayCfg `toml:"dataway"`
	DataWayRequestURL string      `toml:"-"`

	HTTPBind string `toml:"http_server_addr"`

	FtGateway string `toml:"ftdataway"` // XXX: deprecated

	Log      string `toml:"log"`
	LogLevel string `toml:"log_level"`
	GinLog   string `toml:"gin_log"`

	ConfigDir string `toml:"config_dir"` // XXX: not used: to compatible parsing with forethought datakit.conf

	//验证dk存活
	MaxPostInterval string        `toml:"max_post_interval"`
	maxPostInterval time.Duration `toml:"-"`

	GlobalTags map[string]string `toml:"global_tags"`

	RoundInterval    bool
	Interval         string `toml:"interval"`
	IntervalDuration time.Duration

	flushInterval datakit.Duration

	OutputFile string `toml:"output_file"`

	OmitHostname bool // Deprecated

	Hostname string `toml:"hostname"`
	cfgPath  string

	TelegrafAgentCfg *agent `toml:"agent"`
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

	if err := Cfg.LoadMainConfig(); err != nil {
		return err
	}

	l.Infof("set log to %s", Cfg.MainCfg.Log)

	// set global log root
	logger.SetGlobalRootLogger(Cfg.MainCfg.Log, Cfg.MainCfg.LogLevel, logger.OPT_DEFAULT)
	l = logger.SLogger("config")

	l.Infof("main cfg: %+#v", Cfg.MainCfg)

	if Cfg.MainCfg.maxPostInterval > 0 {
		datakit.MaxLifeCheckInterval = Cfg.MainCfg.maxPostInterval
	}

	initPluginSamples()

	if err := Cfg.LoadConfig(); err != nil {
		l.Error(err)
		return err
	}

	return nil
}

func (c *Config) setHostname() {
	hn, err := os.Hostname()
	if err != nil {
		l.Errorf("get hostname failed: %s", err.Error())
	} else {
		c.MainCfg.Hostname = hn
		l.Infof("set hostname to %s", hn)
	}
}

func (c *Config) doLoadMainConfig(cfgdata []byte) error {
	if err := toml.Unmarshal(cfgdata, c.MainCfg); err != nil {
		l.Errorf("unmarshal main cfg failed %s", err.Error())
		return err
	}

	if c.MainCfg.TelegrafAgentCfg.LogTarget == "file" && c.MainCfg.TelegrafAgentCfg.Logfile == "" {
		c.MainCfg.TelegrafAgentCfg.Logfile = filepath.Join(datakit.InstallDir, "embed", "agent.log")
	}

	if c.MainCfg.OutputFile != "" {
		datakit.OutputFile = c.MainCfg.OutputFile
	}

	if c.MainCfg.Hostname == "" {
		c.setHostname()
	}

	if c.MainCfg.MaxPostInterval != "" {
		du, err := time.ParseDuration(c.MainCfg.MaxPostInterval)
		if err != nil {
			l.Warnf("parse %s failed: %s, set default to 15s", c.MainCfg.MaxPostInterval)
			du = time.Second * 15
		}
		c.MainCfg.maxPostInterval = du
	}

	if c.MainCfg.Interval != "" {
		du, err := time.ParseDuration(c.MainCfg.Interval)
		if err != nil {
			l.Warnf("parse %s failed: %s, set default to 10s", c.MainCfg.Interval)
			du = time.Second * 10
		}
		c.MainCfg.IntervalDuration = du
	}

	c.MainCfg.TelegrafAgentCfg.Debug = strings.EqualFold(strings.ToLower(c.MainCfg.LogLevel), "debug")

	c.MainCfg.DataWayRequestURL = fmt.Sprintf("%s://%s%s?token=%s",
		c.MainCfg.DataWay.Scheme, c.MainCfg.DataWay.Host, c.MainCfg.DataWay.DefaultPath, c.MainCfg.DataWay.Token)

	// reset global tags
	for k, v := range c.MainCfg.GlobalTags {
		switch strings.ToLower(v) {
		case `$datakit_hostname`:
			if c.MainCfg.Hostname == "" {
				c.setHostname()
			}

			c.MainCfg.GlobalTags[k] = c.MainCfg.Hostname
			l.Debugf("set global tag %s: %s", k, c.MainCfg.Hostname)

		case `$datakit_ip`:
			c.MainCfg.GlobalTags[k] = "unavailable"

			if ipaddr, err := datakit.LocalIP(); err != nil {
				l.Errorf("get local ip failed: %s", err.Error())
			} else {
				l.Debugf("set global tag %s: %s", k, ipaddr)
				c.MainCfg.GlobalTags[k] = ipaddr
			}

		case `$datakit_uuid`, `$datakit_id`:
			c.MainCfg.GlobalTags[k] = c.MainCfg.UUID
			l.Debugf("set global tag %s: %s", k, c.MainCfg.UUID)
		default:
			// pass
		}
	}

	return nil
}

func (c *Config) LoadMainConfig() error {
	cfgdata, err := ioutil.ReadFile(c.MainCfg.cfgPath)
	if err != nil {
		l.Errorf("reaed main cfg %s failed: %s", c.MainCfg.cfgPath, err.Error())
		return err
	}

	return c.doLoadMainConfig(cfgdata)
}

func CheckConfd() error {
	dir, err := ioutil.ReadDir(datakit.ConfdDir)
	if err != nil {
		return err
	}

	configed := []string{}
	invalids := []string{}

	checkSubDir := func(path string) error {

		ent, err := ioutil.ReadDir(path)
		if err != nil {
			return err
		}

		for _, item := range ent {
			if item.IsDir() {
				continue
			}

			filename := item.Name()

			if filename == "." || filename == ".." { //nolint:goconst
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
			} else if len(tbl.Fields) > 0 {
				configed = append(configed, filename)
			}
		}

		return nil
	}

	for _, item := range dir {
		if !item.IsDir() {
			continue
		}

		if item.Name() == "." || item.Name() == ".." { //nolint:goconst
			continue
		}

		if err := checkSubDir(filepath.Join(datakit.ConfdDir, item.Name())); err != nil {
			l.Error("checkSubDir: %s", err.Error())
		}
	}

	fmt.Printf("inputs: %s\n", strings.Join(configed, ","))
	fmt.Printf("error configuration: %s\n", strings.Join(invalids, ","))

	return nil
}

func buildMainCfg(mc *MainConfig) ([]byte, error) {
	data, err := toml.Marshal(mc)
	return data, err
}

func InitCfg() error {
	data, err := buildMainCfg(Cfg.MainCfg)
	if err != nil {
		return err
	}

	if Cfg.MainCfg.Hostname == "" {
		Cfg.setHostname()
	}

	if err := ioutil.WriteFile(Cfg.MainCfg.cfgPath, data, 0600); err != nil {
		return fmt.Errorf("error creating %s: %s", Cfg.MainCfg.cfgPath, err)
	}

	return nil
}

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

func ParseDataway(dw string) (*DataWayCfg, error) {

	dwcfg := &DataWayCfg{
		Timeout: "30s",
	}

	if u, err := url.Parse(dw); err == nil {
		dwcfg.Scheme = u.Scheme
		dwcfg.Token = u.Query().Get("token")
		dwcfg.Host = u.Host
		dwcfg.DefaultPath = u.Path

		if dwcfg.Scheme == "https" {
			dwcfg.Host += ":443"
		}

		l.Debugf("dataway: %+#v", dwcfg)
	} else {
		l.Errorf("parse url %s failed: %s", dw, err.Error())
		return nil, err
	}

	conn, err := net.DialTimeout("tcp", dwcfg.Host, time.Minute)
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
