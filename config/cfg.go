package config

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	bstoml "github.com/BurntSushi/toml"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/ddtrace/tracer"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	dkhttp "gitlab.jiagouyun.com/cloudcare-tools/datakit/http"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/dataway"
	dktracer "gitlab.jiagouyun.com/cloudcare-tools/datakit/tracer"
)

var (
	IntervalDuration = 10 * time.Second

	Cfg = DefaultConfig()

	l = logger.DefaultSLogger("config")
)

func SetLog() {
	l = logger.SLogger("config")
}

func DefaultConfig() *Config {
	c := &Config{ //nolint:dupl
		GlobalTags: map[string]string{
			"project": "",
			"cluster": "",
			"site":    "",
		},

		Environments: map[string]string{
			"ENV_HOSTNAME": "", // not set
		}, // default nothing

		IOConf: &IOConfig{
			FeedChanSize:              1024,
			HighFreqFeedChanSize:      2048,
			MaxCacheCount:             1024,
			CacheDumpThreshold:        512,
			MaxDynamicCacheCount:      1024,
			DynamicCacheDumpThreshold: 512,
			FlushInterval:             "10s",
		},

		DataWay: &dataway.DataWayCfg{},

		ProtectMode: true,

		HTTPAPI: &dkhttp.APIConfig{
			RUMOriginIPHeader: "X-Forwarded-For",
			Listen:            "localhost:9529",
		},

		Logging: &LoggerCfg{
			Level:  "info",
			Rotate: 32,
			Log:    filepath.Join("/var/log/datakit", "log"),
			GinLog: filepath.Join("/var/log/datakit", "gin.log"),
		},

		BlackList: []*inputHostList{
			&inputHostList{Hosts: []string{}, Inputs: []string{}},
		},
		WhiteList: []*inputHostList{
			&inputHostList{Hosts: []string{}, Inputs: []string{}},
		},
		Cgroup: &Cgroup{Enable: false, CPUMax: 30.0, CPUMin: 5.0},
	}

	// windows 下，日志继续跟 datakit 放在一起
	if runtime.GOOS == "windows" {
		c.Logging.Log = filepath.Join(datakit.InstallDir, "log")
		c.Logging.GinLog = filepath.Join(datakit.InstallDir, "gin.log")
	}

	return c
}

type Cgroup struct {
	Enable bool    `toml:"enable"`
	CPUMax float64 `toml:"cpu_max"`
	CPUMin float64 `toml:"cpu_min"`
}

type IOConfig struct {
	FeedChanSize              int    `toml:"feed_chan_size"`
	HighFreqFeedChanSize      int    `toml:"high_frequency_feed_chan_size"`
	MaxCacheCount             int64  `toml:"max_cache_count"`
	CacheDumpThreshold        int64  `toml:"cache_dump_threshold"`
	MaxDynamicCacheCount      int64  `toml:"max_dynamic_cache_count"`
	DynamicCacheDumpThreshold int64  `toml:"dynamic_cache_dump_threshold"`
	FlushInterval             string `toml:"flush_interval"`
	OutputFile                string `toml:"output_file"`
}

type LoggerCfg struct {
	Log          string `toml:"log"`
	GinLog       string `toml:"gin_log"`
	Level        string `toml:"level"`
	DisableColor bool   `toml:"disable_color"`
	Rotate       int    `toml:"rotate,omitzero"`
}

type Config struct {
	UUID           string `toml:"-"`
	UUIDDeprecated string `toml:"uuid,omitempty"` // deprecated

	Name      string `toml:"name,omitempty"`
	Hostname  string `toml:"-"`
	Namespace string `toml:"namespace"`

	IOConf *IOConfig `toml:"io"`

	DataWay *dataway.DataWayCfg `toml:"dataway,omitempty"`

	// http config: TODO: merge into APIConfig
	HTTPBindDeprecated       string            `toml:"http_server_addr,omitempty"`
	HTTPListenDeprecated     string            `toml:"http_listen,omitempty"`
	Disable404PageDeprecated bool              `toml:"disable_404page,omitempty"`
	HTTPAPI                  *dkhttp.APIConfig `toml:"http_api"`

	// logging config
	LogDeprecated       string     `toml:"log,omitempty"`
	LogLevelDeprecated  string     `toml:"log_level,omitempty"`
	LogRotateDeprecated int        `toml:"log_rotate,omitzero"`
	GinLogDeprecated    string     `toml:"gin_log,omitempty"`
	Logging             *LoggerCfg `toml:"logging"`

	GlobalTags   map[string]string `toml:"global_tags"`
	Environments map[string]string `toml:"environments"`

	OutputFileDeprecated string `toml:"output_file,omitempty"`

	EnablePProf bool `toml:"enable_pprof,omitempty"`
	ProtectMode bool `toml:"protect_mode"`

	IntervalDeprecated string `toml:"interval,omitempty"`

	DefaultEnabledInputs []string  `toml:"default_enabled_inputs,omitempty"`
	InstallDate          time.Time `toml:"install_date,omitempty"`

	BlackList []*inputHostList `toml:"black_lists,omitempty"`
	WhiteList []*inputHostList `toml:"white_lists,omitempty"`
	Cgroup    *Cgroup          `toml:"cgroup"`

	EnableElection         bool           `toml:"enable_election"`
	IOCacheCountDeprecated int64          `toml:"io_cache_count,omitzero"`
	Tracer                 *tracer.Tracer `toml:"tracer,omitempty"`

	// 是否已开启自动更新，通过 dk-install --ota 来开启
	AutoUpdate bool `toml:"auto_update,omitempty"`

	EnableUncheckedInputs bool `toml:"enable_unchecked_inputs,omitempty"`
}

func (c *Config) String() string {
	buf := new(bytes.Buffer)
	if err := bstoml.NewEncoder(buf).Encode(c); err != nil {
		return ""
	}

	return buf.String()
}

func (c *Config) SetUUID() error {
	if c.Hostname == "" {
		hn, err := os.Hostname()
		if err != nil {
			l.Errorf("get hostname failed: %s", err.Error())
			return err
		}

		c.UUID = hn
	} else {
		c.UUID = c.Hostname
	}
	return nil
}

func (c *Config) LoadMainTOML(p string) error {
	cfgdata, err := ioutil.ReadFile(p)
	if err != nil {
		l.Errorf("read main cfg %s failed: %s", p, err.Error())
		return err
	}

	_, err = bstoml.Decode(string(cfgdata), c)
	if err != nil {
		l.Errorf("unmarshal main cfg failed %s", err.Error())
		return err
	}

	_ = c.SetUUID()

	return nil
}

type inputHostList struct {
	Hosts  []string `toml:"hosts"`
	Inputs []string `toml:"inputs"`
}

func (i *inputHostList) MatchHost(host string) bool {
	for _, hostname := range i.Hosts {
		if hostname == host {
			return true
		}
	}

	return false
}

func (i *inputHostList) MatchInput(input string) bool {
	for _, name := range i.Inputs {
		if name == input {
			return true
		}
	}

	return false
}

func InitDirs() {
	for _, dir := range []string{
		datakit.DataDir,
		datakit.ConfdDir,
		datakit.PipelineDir,
		datakit.PipelinePatternDir} {
		if err := os.MkdirAll(dir, datakit.ConfPerm); err != nil {
			l.Fatalf("create %s failed: %s", dir, err)
		}
	}
}

func (c *Config) InitCfg(p string) error {

	if c.Hostname == "" {
		c.setHostname()
	}

	if mcdata, err := datakit.TomlMarshal(c); err != nil {
		l.Errorf("TomlMarshal(): %s", err.Error())
		return err
	} else {
		if err := ioutil.WriteFile(p, mcdata, datakit.ConfPerm); err != nil {
			l.Errorf("error creating %s: %s", p, err)
			return err
		}
	}

	return nil
}

func (c *Config) setupDataway() error {
	// 如果 env 已传入了 dataway 配置, 则不再追加老的 dataway 配置,
	// 避免俩边配置了同样的 dataway, 造成数据混乱
	if c.DataWay.DeprecatedURL != "" && len(c.DataWay.URLs) == 0 {
		c.DataWay.URLs = []string{c.DataWay.DeprecatedURL}
	}

	if len(c.DataWay.URLs) == 0 {
		return fmt.Errorf("dataway not set")
	}

	dataway.ExtraHeaders = map[string]string{
		"X-Datakit-Info": fmt.Sprintf("%s; %s", c.Hostname, datakit.Version),
	}

	c.DataWay.Hostname = c.Hostname

	// setup dataway
	return c.DataWay.Apply()
}

func (c *Config) setupGlobalTags() error {
	if c.GlobalTags == nil {
		c.GlobalTags = map[string]string{}
	}

	delete(c.GlobalTags, "host") // delete host tag if configured

	// setup global tags
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

			if ipaddr, err := datakit.LocalIP(); err != nil {
				l.Errorf("get local ip failed: %s", err.Error())
			} else {
				l.Infof("set global tag %s: %s", k, ipaddr)
				c.GlobalTags[k] = ipaddr
			}

		case `__datakit_uuid`, `__datakit_id`, `$datakit_uuid`, `$datakit_id`:
			c.GlobalTags[k] = c.UUID
			l.Debugf("set global tag %s: %s", k, c.UUID)

		default:
			// pass
		}
	}

	return nil
}

func (c *Config) setLogging() {

	// set global log root
	if c.Logging.Log != "stdout" || c.Logging.Log != "" { // set log to disk file

		l.Infof("set log to %s", c.Logging.Log)

		if c.Logging.Rotate > 0 {
			logger.MaxSize = c.Logging.Rotate
		}

		if err := logger.InitRoot(&logger.Option{
			Path:  c.Logging.Log,
			Level: c.Logging.Level,
			Flags: logger.OPT_DEFAULT}); err != nil {
			l.Errorf("set root log faile: %s", err.Error())
		}
	} else {

		l.Info("set log to stdout, rotate disabled")

		optflags := (logger.OPT_DEFAULT | logger.OPT_STDOUT)
		if !c.Logging.DisableColor {
			optflags |= logger.OPT_COLOR
		}

		if err := logger.InitRoot(
			&logger.Option{
				Level: c.Logging.Level,
				Flags: optflags}); err != nil {
			l.Errorf("set root log faile: %s", err.Error())
		}
	}
}

func (c *Config) ApplyMainConfig() error {

	c.setLogging()

	l = logger.SLogger("config")

	if c.EnableUncheckedInputs {
		datakit.EnableUncheckInputs = true
	}

	if c.Hostname == "" {
		if err := c.setHostname(); err != nil {
			return err
		}
	}

	if err := c.setupDataway(); err != nil {
		return err
	}

	// initialize global tracer
	if c.Tracer != nil {
		dktracer.GlobalTracer = c.Tracer
	}

	datakit.AutoUpdate = c.AutoUpdate

	// config default io
	if c.IOConf != nil {
		if c.IOConf.MaxCacheCount == 0 && c.IOCacheCountDeprecated != 0 {
			c.IOConf.MaxCacheCount = c.IOCacheCountDeprecated
		}
		if c.IOConf.OutputFile == "" && c.OutputFileDeprecated != "" {
			c.IOConf.OutputFile = c.OutputFileDeprecated
		}
		dkio.ConfigDefaultIO(dkio.SetFeedChanSize(c.IOConf.FeedChanSize), dkio.SetHighFreqFeedChanSize(c.IOConf.HighFreqFeedChanSize), dkio.SetMaxCacheCount(c.IOConf.MaxCacheCount), dkio.SetCacheDumpThreshold(c.IOConf.CacheDumpThreshold), dkio.SetMaxDynamicCacheCount(c.IOConf.MaxDynamicCacheCount), dkio.SetDynamicCacheDumpThreshold(c.IOConf.DynamicCacheDumpThreshold), dkio.SetFlushInterval(c.IOConf.FlushInterval), dkio.SetOutputFile(c.IOConf.OutputFile), dkio.SetDataway(c.DataWay))
	}

	if err := c.setupGlobalTags(); err != nil {
		return err
	}

	for k, v := range c.GlobalTags {
		dkio.SetExtraTags(k, v)
	}

	// 此处不将 host 计入 c.GlobalTags，因为 c.GlobalTags 是读取的用户配置，而 host
	// 是不允许修改的, 故单独添加这个 tag 到 io 模块
	dkio.SetExtraTags("host", c.Hostname)

	if c.IntervalDeprecated != "" {
		du, err := time.ParseDuration(c.IntervalDeprecated)
		if err != nil {
			l.Warnf("parse %s failed: %s, set default to 10s", c.IntervalDeprecated)
			du = time.Second * 10
		}
		IntervalDuration = du
	}

	// remove deprecated UUID field in main configure
	if c.UUIDDeprecated != "" {
		c.UUIDDeprecated = "" // clear deprecated UUID field
		buf := new(bytes.Buffer)
		if err := bstoml.NewEncoder(buf).Encode(c); err != nil {
			l.Fatalf("encode main configure failed: %s", err.Error())
		}
		if err := ioutil.WriteFile(datakit.MainConfPath, buf.Bytes(), datakit.ConfPerm); err != nil {
			l.Fatalf("refresh main configure failed: %s", err.Error())
		}

		l.Info("refresh main configure ok")
	}

	return nil
}

func (c *Config) setHostname() error {

	// try get hostname from configure
	if v, ok := c.Environments["ENV_HOSTNAME"]; ok && v != "" {
		c.Hostname = v
		l.Infof("set hostname to %s from config ENV_HOSTNAME", v)
		return nil
	}

	// try get hostname from $env
	if v := datakit.GetEnv("ENV_HOSTNAME"); v != "" {
		c.Hostname = v
		l.Infof("set hostname to %s from env ENV_HOSTNAME", v)
		return nil
	}

	// get real hostname
	hn, err := os.Hostname()
	if err != nil {
		l.Errorf("get hostname failed: %s", err.Error())
		return err
	}

	c.Hostname = hn
	l.Infof("set hostname to %s", hn)
	return nil
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

func (c *Config) LoadEnvs() error {
	if c.IOConf == nil {
		c.IOConf = &IOConfig{}
	}
	for _, envkey := range []string{"ENV_MAX_CACHE_COUNT", "ENV_CACHE_DUMP_THRESHOLD", "ENV_MAX_DYNAMIC_CACHE_COUNT", "ENV_DYNAMIC_CACHE_DUMP_THRESHOLD"} {
		if v := datakit.GetEnv(envkey); v != "" {
			value, err := strconv.ParseInt(v, 10, 64)
			if err != nil {
				l.Errorf("invalid env key value pair [%s:%s]", envkey, v)
			} else {
				switch envkey {
				case "ENV_MAX_CACHE_COUNT":
					c.IOConf.MaxCacheCount = value
				case "ENV_CACHE_DUMP_THRESHOLD":
					c.IOConf.CacheDumpThreshold = value
				case "ENV_MAX_DYNAMIC_CACHE_COUNT":
					c.IOConf.MaxDynamicCacheCount = value
				case "ENV_DYNAMIC_CACHE_DUMP_THRESHOLD":
					c.IOConf.DynamicCacheDumpThreshold = value
				}
			}
		}
	}

	if v := datakit.GetEnv("ENV_NAMESPACE"); v != "" {
		c.Namespace = v
	}

	if v := datakit.GetEnv("ENV_DISABLE_404PAGE"); v != "" {
		c.HTTPAPI.Disable404Page = true
	}

	if v := datakit.GetEnv("ENV_GLOBAL_TAGS"); v != "" {
		c.GlobalTags = ParseGlobalTags(v)
	}

	if v := datakit.GetEnv("ENV_LOG_LEVEL"); v != "" {
		c.Logging.Level = v
	}

	if v := datakit.GetEnv("ENV_LOG"); v != "" {
		c.Logging.Log = v
	}

	// 多个 dataway 支持 ',' 分割
	if v := datakit.GetEnv("ENV_DATAWAY"); v != "" {

		if c.DataWay == nil {
			c.DataWay = &dataway.DataWayCfg{}
		}
		c.DataWay.URLs = strings.Split(v, ",")
	}

	if v := datakit.GetEnv("ENV_HOSTNAME"); v != "" {
		c.Hostname = v
	}

	if v := datakit.GetEnv("ENV_NAME"); v != "" {
		c.Name = v
	}

	if v := datakit.GetEnv("ENV_HTTP_LISTEN"); v != "" {
		c.HTTPAPI.Listen = v
	}

	if v := datakit.GetEnv("ENV_RUM_ORIGIN_IP_HEADER"); v != "" {
		c.HTTPAPI = &dkhttp.APIConfig{RUMOriginIPHeader: v}
	}

	if v := datakit.GetEnv("ENV_ENABLE_PPROF"); v != "" {
		c.EnablePProf = true
	}

	if v := datakit.GetEnv("ENV_DISABLE_PROTECT_MODE"); v != "" {
		c.ProtectMode = false
	}

	if v := datakit.GetEnv("ENV_DEFAULT_ENABLED_INPUTS"); v != "" {
		c.DefaultEnabledInputs = strings.Split(v, ",")
	} else {
		if v := datakit.GetEnv("ENV_ENABLE_INPUTS"); v != "" { // deprecated
			c.DefaultEnabledInputs = strings.Split(v, ",")
		}
	}

	if v := datakit.GetEnv("ENV_ENABLE_ELECTION"); v != "" {
		c.EnableElection = true
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
	return ioutil.WriteFile(f, []byte(uuid), datakit.ConfPerm)
}

func LoadUUID(f string) (string, error) {
	if data, err := ioutil.ReadFile(f); err != nil {
		return "", err
	} else {
		return string(data), nil
	}
}

func emptyDir(fp string) bool {
	fd, err := os.Open(fp)
	if err != nil {
		l.Error(err)
		return false
	}

	defer fd.Close()

	_, err = fd.ReadDir(1)
	switch err {
	case io.EOF:
		return true
	default:
		return false
	}
}

// remove all xxx.conf.sample
func removeSamples() {

	l.Debugf("searching samples under %s", datakit.ConfdDir)

	fps := SearchDir(datakit.ConfdDir, ".conf.sample")

	l.Debugf("searched %d samples", len(fps))

	for _, fp := range fps {
		if err := os.Remove(fp); err != nil {
			l.Error(err)
			continue
		}

		l.Debugf("remove sample %s", fp)

		// check if directory empty
		pwd := filepath.Dir(fp)
		if emptyDir(pwd) {
			if err := os.RemoveAll(pwd); err != nil {
				l.Error(err)
			}
		}

		l.Debugf("remove dir %s", pwd)
	}
}

func MoveDeprecatedCfg() {

	if _, err := os.Stat(datakit.MainConfPathDeprecated); err == nil {
		if err := os.Rename(datakit.MainConfPathDeprecated, datakit.MainConfPath); err != nil {
			l.Fatal("move deprecated main configure failed: %s", err.Error())
		}
		l.Infof("move %s to %s", datakit.MainConfPathDeprecated, datakit.MainConfPath)
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

func CreateSymlinks() error {

	x := [][2]string{}

	if runtime.GOOS == "windows" {
		x = [][2]string{
			[2]string{
				filepath.Join(datakit.InstallDir, "datakit.exe"),
				`C:\WINDOWS\system32\datakit.exe`,
			},
		}
	} else {
		x = [][2]string{
			[2]string{
				filepath.Join(datakit.InstallDir, "datakit"),
				"/usr/local/bin/datakit",
			},

			[2]string{
				filepath.Join(datakit.InstallDir, "datakit"),
				"/usr/local/sbin/datakit",
			},

			[2]string{
				filepath.Join(datakit.InstallDir, "datakit"),
				"/sbin/datakit",
			},

			[2]string{
				filepath.Join(datakit.InstallDir, "datakit"),
				"/usr/sbin/datakit",
			},

			[2]string{
				filepath.Join(datakit.InstallDir, "datakit"),
				"/usr/bin/datakit",
			},
		}
	}

	for _, item := range x {
		if err := symlink(item[0], item[1]); err != nil {
			l.Warnf("create datakit symlink: %s -> %s: %s, ignored", item[1], item[0], err.Error())
		}
	}

	return nil
}

func symlink(src, dst string) error {

	l.Debugf("remove link %s...", dst)
	if err := os.Remove(dst); err != nil {
		l.Warnf("%s, ignored", err)
	}

	return os.Symlink(src, dst)
}
