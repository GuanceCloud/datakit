package datakit

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	bstoml "github.com/BurntSushi/toml"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
)

var (
	IntervalDuration = 10 * time.Second

	Cfg = DefaultConfig()
)

func DefaultConfig() *Config {
	c := &Config{ //nolint:dupl
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
		HTTPAPI: &apiConfig{
			RUMOriginIPHeader: "X-Forward-For",
		},

		LogLevel:  "info",
		LogRotate: 32,
		Log:       filepath.Join("/var/log/datakit", "log"),
		GinLog:    filepath.Join("/var/log/datakit", "gin.log"),

		BlackList: []*InputHostList{
			&InputHostList{Hosts: []string{}, Inputs: []string{}},
		},
		WhiteList: []*InputHostList{
			&InputHostList{Hosts: []string{}, Inputs: []string{}},
		},
	}

	// windows 下，日志继续跟 datakit 放在一起
	if runtime.GOOS == OSWindows {
		c.Log = filepath.Join(InstallDir, "log")
		c.GinLog = filepath.Join(InstallDir, "gin.log")
	}

	return c
}

type apiConfig struct {
	RUMOriginIPHeader string `toml:"rum_origin_ip_header"`
}

type Config struct {
	UUID           string `toml:"-"`
	UUIDDeprecated string `toml:"uuid,omitempty"` // deprecated

	Name    string      `toml:"name,omitempty"`
	DataWay *DataWayCfg `toml:"dataway,omitempty"`

	HTTPBindDeprecated string `toml:"http_server_addr,omitempty"`
	HTTPListen         string `toml:"http_listen,omitempty"`

	HTTPAPI *apiConfig `toml:"http_api"`

	Log       string `toml:"log"`
	LogLevel  string `toml:"log_level"`
	LogRotate int    `toml:"log_rotate,omitempty"`

	GinLog     string            `toml:"gin_log"`
	GlobalTags map[string]string `toml:"global_tags"`

	EnablePProf bool `toml:"enable_pprof,omitempty"`
	ProtectMode bool `toml:"protect_mode,omitempty"`

	Interval             string `toml:"interval"`
	flushInterval        Duration
	OutputFile           string    `toml:"output_file"`
	Hostname             string    `toml:"hostname,omitempty"`
	DefaultEnabledInputs []string  `toml:"default_enabled_inputs,omitempty"`
	InstallDate          time.Time `toml:"install_date,omitempty"`

	BlackList []*InputHostList `toml:"black_lists,omitempty"`
	WhiteList []*InputHostList `toml:"white_lists,omitempty"`

	EnableElection bool `toml:"enable_election"`

	EnableUncheckedInputs bool `toml:"enable_unchecked_inputs,omitempty"`
}

func (c *Config) String() string {
	buf := new(bytes.Buffer)
	if err := bstoml.NewEncoder(buf).Encode(c); err != nil {
		return ""
	}

	return buf.String()
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
	for _, dir := range []string{
		DataDir,
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

	if c.Hostname == "" {
		c.setHostname()
	}

	if mcdata, err := TomlMarshal(c); err != nil {
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
	_, err := bstoml.Decode(string(cfgdata), c)
	if err != nil {
		l.Errorf("unmarshal main cfg failed %s", err.Error())
		return err
	}

	if c.EnableUncheckedInputs {
		EnableUncheckInputs = true
	}

	// load datakit UUID
	if c.UUIDDeprecated != "" {
		// dump UUIDDeprecated to .id file
		if err := CreateUUIDFile(UUIDFile, Cfg.UUIDDeprecated); err != nil {
			l.Fatalf("create datakit id failed: %s", err.Error())
		}
		c.UUID = c.UUIDDeprecated
	} else {
		c.UUID, err = LoadUUID()
		if err != nil {
			l.Fatalf("load datakit id failed: %s", err.Error())
		}
	}

	if c.OutputFile != "" {
		OutputFile = c.OutputFile
	}

	if c.Hostname == "" {
		c.setHostname()
	}
	if c.GlobalTags == nil {
		c.GlobalTags = map[string]string{}
	}

	// add global tag implicitly
	if _, ok := c.GlobalTags["host"]; !ok {
		c.GlobalTags["host"] = c.Hostname
	}

	if c.DataWay.DeprecatedURL != "" {
		c.DataWay.URLs = append(c.DataWay.URLs, c.DataWay.DeprecatedURL)
	}

	if len(c.DataWay.URLs) == 0 {
		l.Fatal("dataway URL not set")
	}

	// set global log root
	l.Infof("set log to %s", c.Log)
	logger.MaxSize = c.LogRotate
	logger.SetGlobalRootLogger(c.Log, c.LogLevel, logger.OPT_DEFAULT)

	l = logger.SLogger("datakit")

	dw, err := ParseDataway(c.DataWay.URLs)
	if err != nil {
		return err
	}

	c.DataWay = dw

	if c.Interval != "" {
		du, err := time.ParseDuration(c.Interval)
		if err != nil {
			l.Warnf("parse %s failed: %s, set default to 10s", c.Interval)
			du = time.Second * 10
		}
		IntervalDuration = du
	}

	// reset global tags
	for k, v := range c.GlobalTags {

		// NOTE: accept `__` and `$` as tag-key prefix, to keep compatible with old prefix `$`
		// by using `__` as prefix, avoid escaping `$` in Powershell and shell

		switch strings.ToLower(v) {
		case `__datakit_hostname`, `$datakit_hostname`:
			if c.Hostname == "" {
				c.setHostname()
			}

			c.GlobalTags[k] = c.Hostname
			l.Debugf("set global tag %s: %s", k, c.Hostname)

		case `__datakit_ip`, `$datakit_ip`:
			c.GlobalTags[k] = "unavailable"

			if ipaddr, err := LocalIP(); err != nil {
				l.Errorf("get local ip failed: %s", err.Error())
			} else {
				l.Debugf("set global tag %s: %s", k, ipaddr)
				c.GlobalTags[k] = ipaddr
			}

		case `__datakit_uuid`, `__datakit_id`, `$datakit_uuid`, `$datakit_id`:
			c.GlobalTags[k] = c.UUID
			l.Debugf("set global tag %s: %s", k, c.UUID)

		default:
			// pass
		}
	}

	// remove deprecated UUID field in main configure
	if c.UUIDDeprecated != "" {
		c.UUIDDeprecated = "" // clear deprecated UUID field
		buf := new(bytes.Buffer)
		if err := bstoml.NewEncoder(buf).Encode(c); err != nil {
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
		c.Hostname = hn
		l.Infof("set hostname to %s", hn)
	}
}

func (c *Config) EnableDefaultsInputs(inputlist string) {
	inputs := []string{}
	inputsUnique := make(map[string]bool)

	for _, name := range c.DefaultEnabledInputs {
		if _, ok := inputsUnique[name]; !ok {
			inputsUnique[name] = true
			inputs = append(inputs, name)
		}
	}

	elems := strings.Split(inputlist, ",")
	for _, name := range elems {
		if _, ok := inputsUnique[name]; !ok {
			inputsUnique[name] = true
			inputs = append(inputs, name)
		}
	}

	c.DefaultEnabledInputs = inputs
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
		c.GlobalTags = ParseGlobalTags(globalTags)
	}

	loglvl := os.Getenv("ENV_LOG_LEVEL")
	if loglvl != "" {
		c.LogLevel = loglvl
	}

	dwURL := os.Getenv("ENV_DATAWAY")
	if dwURL != "" {
		dwURLs := []string{dwURL}
		dw, err := ParseDataway(dwURLs)
		if err != nil {
			return err
		}

		if err := dw.Test(); err != nil {
			return err
		}

		c.DataWay = dw
		c.DataWay.URLs = dwURLs
	}

	dkhost := os.Getenv("ENV_HOSTNAME")
	if dkhost != "" {
		l.Debugf("set hostname to %s from ENV", dkhost)
		c.Hostname = dkhost
	} else {
		c.setHostname()
	}

	c.Name = os.Getenv("ENV_NAME")

	if fi, err := os.Stat(mcp); err != nil || fi.Size() == 0 { // create the main config
		if c.UUID == "" { // datakit.conf not exit: we have to create new datakit with new UUID
			c.UUID = cliutils.XID("dkid_")
		}

		c.InstallDate = time.Now()

		cfgdata, err := TomlMarshal(c)
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

	if err := CreateUUIDFile(UUIDFile, dkid); err != nil {
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

func CreateUUIDFile(f, uuid string) error {
	return ioutil.WriteFile(f, []byte(uuid), os.ModePerm)
}

func LoadUUID() (string, error) {
	if data, err := ioutil.ReadFile(UUIDFile); err != nil {
		return "", err
	} else {
		return string(data), nil
	}
}

func MoveDeprecatedCfg() {
	if _, err := os.Stat(MainConfPathDeprecated); err == nil {
		if err := os.Rename(MainConfPathDeprecated, MainConfPath); err != nil {
			l.Fatal("move deprecated main configure failed: %s", err.Error())
		}
		l.Infof("move %s to %s", MainConfPathDeprecated, MainConfPath)
	}
}

func ProtectedInterval(min, max, cur time.Duration) time.Duration {
	if Cfg.ProtectMode {
		if cur >= max {
			return max
		}

		if cur <= min {
			return min
		}
	}

	return cur
}
