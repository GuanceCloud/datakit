package datakit

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	bstoml "github.com/BurntSushi/toml"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils"
)

var (
	IntervalDuration = 10 * time.Second

	Cfg = DefaultConfig()
)

func DefaultConfig() *Config {
	return &Config{ //nolint:dupl
		MainCfg: &MainConfig{
			GlobalTags: map[string]string{
				"project": "",
				"cluster": "",
				"site":    "",
			},

			DataWay: &DataWayCfg{},

			flushInterval: Duration{Duration: time.Second * 10},
			Interval:      "10s",
			ProtectMode:   true,

			HTTPListen: "localhost:9529",

			LogLevel:  "info",
			Log:       filepath.Join(InstallDir, "log"),
			LogRotate: 32,
			GinLog:    filepath.Join(InstallDir, "gin.log"),

			BlackList: []*InputHostList{
				&InputHostList{Hosts: []string{}, Inputs: []string{}},
			},
			WhiteList: []*InputHostList{
				&InputHostList{Hosts: []string{}, Inputs: []string{}},
			},

			TelegrafAgentCfg: &TelegrafCfg{
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
				Logfile:                    filepath.Join(TelegrafDir, "agent.log"),
				LogfileRotationMaxArchives: 5,
				LogfileRotationMaxSize:     "32MB",
				OmitHostname:               true, // do not append host tag
			},
		},
	}
}

//用于支持在datakit.conf中加入telegraf的agent配置
type TelegrafCfg struct {
	Interval                   string `toml:"interval"`
	RoundInterval              bool   `toml:"round_interval"`
	Precision                  string `toml:"precision"`
	CollectionJitter           string `toml:"collection_jitter"`
	FlushInterval              string `toml:"flush_interval"`
	FlushJitter                string `toml:"flush_jitter"`
	MetricBatchSize            int    `toml:"metric_batch_size"`
	MetricBufferLimit          int    `toml:"metric_buffer_limit"`
	FlushBufferWhenFull        bool   `toml:"-"`
	UTC                        bool   `toml:"utc"`
	Debug                      bool   `toml:"debug"`
	Quiet                      bool   `toml:"quiet"`
	LogTarget                  string `toml:"logtarget"`
	Logfile                    string `toml:"logfile"`
	LogfileRotationInterval    string `toml:"logfile_rotation_interval"`
	LogfileRotationMaxSize     string `toml:"logfile_rotation_max_size"`
	LogfileRotationMaxArchives int    `toml:"logfile_rotation_max_archives"`
	OmitHostname               bool   `toml:"omit_hostname"`
}

type Config struct {
	MainCfg *MainConfig
}

type MainConfig struct {
	UUID           string `toml:"-"`
	UUIDDeprecated string `toml:"uuid,omitempty"` // deprecated

	Name    string      `toml:"name,omitempty"`
	DataWay *DataWayCfg `toml:"dataway,omitempty"`

	HTTPBindDeprecated string `toml:"http_server_addr,omitempty"`
	HTTPListen         string `toml:"http_listen,omitempty"`

	Log       string `toml:"log"`
	LogLevel  string `toml:"log_level"`
	LogRotate int    `toml:"log_rotate,omitempty"`

	GinLog     string            `toml:"gin_log"`
	GlobalTags map[string]string `toml:"global_tags"`

	EnablePProf bool `toml:"enable_pprof,omitempty"`
	ProtectMode bool `toml:"protect_mode,omitempty"`

	Interval             string `toml:"interval"`
	flushInterval        Duration
	OutputFile           string       `toml:"output_file"`
	Hostname             string       `toml:"hostname,omitempty"`
	DefaultEnabledInputs []string     `toml:"default_enabled_inputs,omitempty"`
	InstallDate          time.Time    `toml:"install_date,omitempty"`
	TelegrafAgentCfg     *TelegrafCfg `toml:"agent"`

	BlackList []*InputHostList `toml:"black_lists,omitempty"`
	WhiteList []*InputHostList `toml:"white_lists,omitempty"`

	EnableUncheckedInputs bool `toml:"enable_unchecked_inputs,omitempty"`
}

type InputHostList struct {
	Hosts  []string `toml:"hosts"`
	Inputs []string `toml:"inputs"`
}

func (i *InputHostList) MatchHost(host string) bool {
	for _, hostname := range i.Hosts {
		if hostname == host {
			return true
		}
	}

	return false
}

func (i *InputHostList) MatchInput(input string) bool {
	for _, name := range i.Inputs {
		if name == input {
			return true
		}
	}

	return false
}

func InitDirs() {
	for _, dir := range []string{TelegrafDir,
		DataDir,
		LuaDir,
		ConfdDir,
		PipelineDir,
		PipelinePatternDir} {
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			l.Fatalf("create %s failed: %s", dir, err)
		}
	}
}

func (c *Config) LoadMainConfig(p string) error {
	cfgdata, err := ioutil.ReadFile(p)
	if err != nil {
		l.Errorf("read main cfg %s failed: %s", p, err.Error())
		return err
	}

	return c.DoLoadMainConfig(cfgdata)
}

func (c *Config) InitCfg(p string) error {

	if c.MainCfg.Hostname == "" {
		c.setHostname()
	}

	if mcdata, err := TomlMarshal(c.MainCfg); err != nil {
		l.Errorf("TomlMarshal(): %s", err.Error())
		return err
	} else {

		if err := ioutil.WriteFile(p, mcdata, 0600); err != nil {
			l.Errorf("error creating %s: %s", p, err)
			return err
		}
	}

	return nil
}

func (c *Config) DoLoadMainConfig(cfgdata []byte) error {
	_, err := bstoml.Decode(string(cfgdata), c.MainCfg)
	if err != nil {
		l.Errorf("unmarshal main cfg failed %s", err.Error())
		return err
	}

	if c.MainCfg.EnableUncheckedInputs {
		EnableUncheckInputs = true
	}

	// load datakit UUID
	if c.MainCfg.UUIDDeprecated != "" {
		// dump UUIDDeprecated to .id file
		if err := CreateUUIDFile(Cfg.MainCfg.UUIDDeprecated); err != nil {
			l.Fatalf("create datakit id failed: %s", err.Error())
		}
		c.MainCfg.UUID = c.MainCfg.UUIDDeprecated
	} else {
		c.MainCfg.UUID, err = LoadUUID()
		if err != nil {
			l.Fatalf("load datakit id failed: %s", err.Error())
		}
	}

	if c.MainCfg.TelegrafAgentCfg.LogTarget == "file" && c.MainCfg.TelegrafAgentCfg.Logfile == "" {
		c.MainCfg.TelegrafAgentCfg.Logfile = filepath.Join(InstallDir, "embed", "agent.log")
	}

	if c.MainCfg.OutputFile != "" {
		OutputFile = c.MainCfg.OutputFile
	}

	if c.MainCfg.Hostname == "" {
		c.setHostname()
	}
	if c.MainCfg.GlobalTags == nil {
		c.MainCfg.GlobalTags = map[string]string{}
	}

	// add global tag implicitly
	if _, ok := c.MainCfg.GlobalTags["host"]; !ok {
		c.MainCfg.GlobalTags["host"] = c.MainCfg.Hostname
	}

	if c.MainCfg.DataWay.DeprecatedURL != "" {
		c.MainCfg.DataWay.Urls = append(c.MainCfg.DataWay.Urls, c.MainCfg.DataWay.DeprecatedURL)
	}

	if len(c.MainCfg.DataWay.Urls) == 0 {
		l.Fatal("dataway URL not set")
	}

	dw, err := ParseDataway(c.MainCfg.DataWay.Urls)
	if err != nil {
		return err
	}

	c.MainCfg.DataWay = dw

	if c.MainCfg.Interval != "" {
		du, err := time.ParseDuration(c.MainCfg.Interval)
		if err != nil {
			l.Warnf("parse %s failed: %s, set default to 10s", c.MainCfg.Interval)
			du = time.Second * 10
		}
		IntervalDuration = du
	}

	c.MainCfg.TelegrafAgentCfg.Debug = strings.EqualFold(strings.ToLower(c.MainCfg.LogLevel), "debug")

	// reset global tags
	for k, v := range c.MainCfg.GlobalTags {

		// NOTE: accept `__` and `$` as tag-key prefix, to keep compatible with old prefix `$`
		// by using `__` as prefix, avoid escaping `$` in Powershell and shell

		switch strings.ToLower(v) {
		case `__datakit_hostname`, `$datakit_hostname`:
			if c.MainCfg.Hostname == "" {
				c.setHostname()
			}

			c.MainCfg.GlobalTags[k] = c.MainCfg.Hostname
			l.Debugf("set global tag %s: %s", k, c.MainCfg.Hostname)

		case `__datakit_ip`, `$datakit_ip`:
			c.MainCfg.GlobalTags[k] = "unavailable"

			if ipaddr, err := LocalIP(); err != nil {
				l.Errorf("get local ip failed: %s", err.Error())
			} else {
				l.Debugf("set global tag %s: %s", k, ipaddr)
				c.MainCfg.GlobalTags[k] = ipaddr
			}

		case `__datakit_uuid`, `__datakit_id`, `$datakit_uuid`, `$datakit_id`:
			c.MainCfg.GlobalTags[k] = c.MainCfg.UUID
			l.Debugf("set global tag %s: %s", k, c.MainCfg.UUID)

		default:
			// pass
		}
	}

	// remove deprecated UUID field in main configure
	if c.MainCfg.UUIDDeprecated != "" {
		c.MainCfg.UUIDDeprecated = "" // clear deprecated UUID field
		buf := new(bytes.Buffer)
		if err := bstoml.NewEncoder(buf).Encode(c.MainCfg); err != nil {
			l.Fatalf("encode main configure failed: %s", err.Error())
		}
		if err := ioutil.WriteFile(MainConfPath, buf.Bytes(), os.ModePerm); err != nil {
			l.Fatalf("refresh main configure failed: %s", err.Error())
		}

		l.Info("refresh main configure ok")
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

func (c *Config) EnableDefaultsInputs(inputlist string) {
	elems := strings.Split(inputlist, ",")
	if len(elems) == 0 {
		return
	}

	for _, name := range elems {
		c.MainCfg.DefaultEnabledInputs = append(c.MainCfg.DefaultEnabledInputs, name)
	}
}

func (c *Config) LoadEnvs(mcp string) error {
	if !Docker { // only accept configs from ENV within docker
		return nil
	}

	enableInputs := os.Getenv("ENV_ENABLE_INPUTS")
	if enableInputs != "" {
		c.EnableDefaultsInputs(enableInputs)
	}

	globalTags := os.Getenv("ENV_GLOBAL_TAGS")
	if globalTags != "" {
		c.MainCfg.GlobalTags = ParseGlobalTags(globalTags)
	}

	loglvl := os.Getenv("ENV_LOG_LEVEL")
	if loglvl != "" {
		c.MainCfg.LogLevel = loglvl
	}

	dwURL := os.Getenv("ENV_DATAWAY")
	dwURLs := []string{dwURL}
	if len(dwURL) != 0 {
		dw, err := ParseDataway(dwURLs)
		if err != nil {
			return err
		}

		if err := dw.Test(); err != nil {
			return err
		}

		c.MainCfg.DataWay = dw
	}

	dkhost := os.Getenv("ENV_HOSTNAME")
	if dkhost != "" {
		l.Debugf("set hostname to %s from ENV", dkhost)
		c.MainCfg.Hostname = dkhost
	} else {
		c.setHostname()
	}

	c.MainCfg.Name = os.Getenv("ENV_NAME")

	if fi, err := os.Stat(mcp); err != nil || fi.Size() == 0 { // create the main config
		if c.MainCfg.UUID == "" { // datakit.conf not exit: we have to create new datakit with new UUID
			c.MainCfg.UUID = cliutils.XID("dkid_")
		}

		c.MainCfg.InstallDate = time.Now()

		cfgdata, err := TomlMarshal(c.MainCfg)
		if err != nil {
			l.Errorf("failed to build main cfg %s", err)
			return err
		}

		l.Debugf("generating datakit.conf...")
		if err := ioutil.WriteFile(mcp, cfgdata, os.ModePerm); err != nil {
			l.Error(err)
			return err
		}
	}

	dkid := os.Getenv("ENV_UUID")
	if dkid == "" {
		return fmt.Errorf("ENV_UUID not set")
	}

	if err := CreateUUIDFile(dkid); err != nil {
		l.Errorf("create id file: %s", err.Error())
		return err
	}

	return nil
}

func ParseGlobalTags(s string) map[string]string {
	tags := map[string]string{}

	parts := strings.Split(s, ",")
	for _, p := range parts {
		arr := strings.Split(p, "=")
		if len(arr) != 2 {
			l.Warnf("invalid global tag: %s, ignored", p)
			continue
		}

		tags[arr[0]] = arr[1]
	}

	return tags
}

func CreateUUIDFile(uuid string) error {
	return ioutil.WriteFile(UUIDFile, []byte(uuid), os.ModePerm)
}

func LoadUUID() (string, error) {
	if data, err := ioutil.ReadFile(UUIDFile); err != nil {
		return "", err
	} else {
		return string(data), nil
	}
}

func MoveDeprecatedMainCfg() {
	if _, err := os.Stat(MainConfPathDeprecated); err == nil {
		if err := os.Rename(MainConfPathDeprecated, MainConfPath); err != nil {
			l.Fatal("move deprecated main configure failed: %s", err.Error())
		}
		l.Infof("move %s to %s", MainConfPathDeprecated, MainConfPath)
	}
}

func ProtectedInterval(min, max, cur time.Duration) time.Duration {
	if Cfg.MainCfg.ProtectMode {
		if cur >= max {
			return max
		}

		if cur <= min {
			return min
		}
	}

	return cur
}
