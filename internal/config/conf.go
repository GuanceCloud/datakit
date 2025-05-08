// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package config manage datakit's configurations, include all inputs TOML configure.
package config

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"

	bstoml "github.com/BurntSushi/toml"
	"github.com/GuanceCloud/cliutils/logger"
	"github.com/GuanceCloud/cliutils/point"
	gctoml "github.com/GuanceCloud/toml"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/operator"
)

var (
	Cfg           = DefaultConfig()
	l             = logger.DefaultSLogger("config")
	versionRegExp = regexp.MustCompile(`install_version\s+=\s+"[^"]+"`)
)

func SetLog() {
	l = logger.SLogger("config")
}

type LoggerCfg struct {
	Log           string `toml:"log"`
	GinLog        string `toml:"gin_log"`
	Level         string `toml:"level"`
	DisableColor  bool   `toml:"disable_color"`
	Rotate        int    `toml:"rotate,omitzero"`
	RotateBackups int    `toml:"rotate_backups"`
}

type DKUpgraderCfg struct {
	Host string `toml:"host"`
	Port int    `toml:"port"`
}

type GitRepository struct {
	Enable                bool   `toml:"enable"`
	URL                   string `toml:"url"`
	SSHPrivateKeyPath     string `toml:"ssh_private_key_path"`
	SSHPrivateKeyPassword string `toml:"ssh_private_key_password"`
	Branch                string `toml:"branch"`
}

type GitRepost struct {
	PullInterval string           `toml:"pull_interval"`
	Repos        []*GitRepository `toml:"repo"`
}

type configCrpto struct {
	AESKey     string `toml:"aes_key"`
	AESKeyFile string `toml:"aes_key_file"`
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
	cfgdata, err := os.ReadFile(filepath.Clean(p))
	if err != nil {
		return fmt.Errorf("os.ReadFile: %w", err)
	}

	_, err = bstoml.Decode(string(cfgdata), c)
	if err != nil {
		return fmt.Errorf("bstoml.Decode: %w", err)
	}

	_ = c.SetUUID()

	return nil
}

func (c *Config) loadMainTOMLString(cfgData string) (gctoml.MetaData, error) {
	meta, err := gctoml.Decode(cfgData, c)
	if err != nil {
		return meta, fmt.Errorf("bstoml.Decode: %w", err)
	}

	_ = c.SetUUID()

	return meta, nil
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

func (c *Config) TryUpgradeCfg(p string) error {
	oldData, err := os.ReadFile(filepath.Clean(p))
	if err != nil {
		l.Warnf("unable to open old configuration file: %s", err)
		return c.InitCfg(p)
	}

	// replace the install_version
	replacedText := string(oldData)
	matches := versionRegExp.FindAllString(replacedText, -1)
	for _, match := range matches {
		replacedText = strings.Replace(replacedText, match, fmt.Sprintf(`install_version = "%s"`, c.InstallVer), 1)
	}

	newTomlText, err := c.InitCfgOutput()
	if err != nil {
		return fmt.Errorf("unable to encode toml: %w", err)
	}

	// check if the toml file has been modified
	if ok, err := ifTOMLEqual(replacedText, string(newTomlText)); err == nil && ok {
		if c.Hostname == "" {
			if err := c.SetHostname(); err != nil {
				return err
			}
		}

		l.Infof("dump datakit.conf...")
		if err := os.WriteFile(p, []byte(replacedText), datakit.ConfPerm); err == nil {
			return nil
		} else {
			l.Warnf("unable to write the toml file: %s", err)
		}
	}

	oldCfg := DefaultConfig()
	// load comments from old toml data
	meta, err := oldCfg.loadMainTOMLString(string(oldData))
	if err != nil {
		return fmt.Errorf("unable to load toml by gctoml: %w", err)
	}
	if err := c.InitCfgWithComments(p, meta); err == nil {
		return nil
	} else {
		l.Warnf("unable to generate toml with comments: %s", err)
	}

	l.Infof("Datakit main configuration file will be backup")
	// Todo: keep comments when update toml instead
	cp := p + ".old." + time.Now().Format("20060102150405")
	if err := os.WriteFile(cp, oldData, datakit.ConfPerm); err != nil {
		l.Warnf("unable to backup old configuration file: %s", err)
		return c.InitCfg(p)
	}
	return c.InitCfg(p)
}

func (c *Config) InitCfgOutput() ([]byte, error) {
	if c.Hostname == "" {
		if err := c.SetHostname(); err != nil {
			return nil, err
		}
	}

	tomlText, err := datakit.TomlMarshal(c)
	if err != nil {
		l.Errorf("TomlMarshal(): %s", err.Error())
		return nil, err
	}
	return tomlText, nil
}

func (c *Config) InitCfg(p string) error {
	tomlText, err := c.InitCfgOutput()
	if err != nil {
		return err
	}

	if err := os.WriteFile(p, tomlText, datakit.ConfPerm); err != nil {
		l.Errorf("error creating %s: %s", p, err)
		return err
	}

	return nil
}

func ifTOMLEqual(toml1, toml2 string) (bool, error) {
	c1, c2 := DefaultConfig(), DefaultConfig()

	if _, err := bstoml.Decode(toml1, &c1); err != nil {
		return false, fmt.Errorf("unable to decode toml: %w", err)
	}

	if _, err := bstoml.Decode(toml2, &c2); err != nil {
		return false, fmt.Errorf("unable to decode toml: %w", err)
	}

	return c1.String() == c2.String(), nil
}

func (c *Config) InitCfgWithComments(path string, meta gctoml.MetaData) error {
	if c.Hostname == "" {
		if err := c.SetHostname(); err != nil {
			return err
		}
	}

	buffer := &bytes.Buffer{}
	if err := gctoml.NewEncoder(buffer).EncodeWithComments(c, meta); err != nil {
		return fmt.Errorf("unable to encode by gctoml: %w", err)
	}

	// check the correctness of toml generated by GuanceCloud/toml
	tomlWithNoComments, err := datakit.TomlMarshal(c)
	if err != nil {
		return fmt.Errorf("unable to generate toml data by bstoml: %w", err)
	}

	ok, err := ifTOMLEqual(buffer.String(), string(tomlWithNoComments))
	if err != nil {
		return fmt.Errorf("unable to compare two toml text: %w", err)
	}
	if !ok {
		return fmt.Errorf("the toml generated by gctoml is incorrect, gctoml: %s, bstoml: %s", buffer.String(), string(tomlWithNoComments))
	}

	if err := os.WriteFile(path, buffer.Bytes(), datakit.ConfPerm); err != nil {
		l.Errorf("error creating %s: %s", path, err)
		return err
	}

	return nil
}

func initCfgSample(p string) error {
	if err := os.WriteFile(p, []byte(datakit.MainConfSample(datakit.BrandDomain)), datakit.ConfPerm); err != nil {
		l.Errorf("error creating %s: %s", p, err)
		return err
	}
	l.Debugf("create datakit sample conf ok, %s!", p)
	return nil
}

func (c *Config) parseGlobalHostTags() {
	if c.GlobalHostTags == nil {
		c.GlobalHostTags = map[string]string{}
	}

	// Parse placeholder of global tags.
	//
	// Accept `__` and `$` as tag-key prefix, to keep compatible with old prefix `$`
	// by using `__` as prefix, avoid escaping `$` in Powershell and shell.
	for k, v := range c.GlobalHostTags {
		switch strings.ToLower(v) {
		case `__datakit_hostname`, `$datakit_hostname`:
			hostName := ""
			if c.Hostname == "" {
				if err := c.SetHostname(); err != nil {
					l.Warnf("setHostname: %s, ignored", err)
				} else {
					hostName = c.Hostname
				}
			} else {
				if hn, err := c.getHostName(); err != nil {
					l.Errorf("get hostname failed: %s", err.Error())
				} else {
					hostName = hn
				}
			}

			c.GlobalHostTags[k] = hostName
			l.Debugf("set global tag %s: %s", k, hostName)

		case `__datakit_ip`, `$datakit_ip`:
			c.GlobalHostTags[k] = getLocalIPUntilOK()

		case `__datakit_uuid`, `__datakit_id`, `$datakit_uuid`, `$datakit_id`:
			c.GlobalHostTags[k] = c.UUID
			l.Debugf("set global tag %s: %s", k, c.UUID)

		default:
			// pass
		}
	}
}

func getLocalIPUntilOK() string {
	retry := 0
	for {
		if ip, err := datakit.LocalIP(); err != nil {
			l.Warnf("[%d] LocalIP: %s, retry...", retry, err)
			time.Sleep(time.Second)
		} else {
			l.Infof("get local IP: %q", ip)
			return ip
		}
		// NOTE: should we check datakit.Exit()?
		retry++
	}
}

func (c *Config) setLogging() {
	// Under debug mode(datakit debug ...), do not setup log here, we should set log
	// via sub-command's --log command.
	if c.cmdlineMode {
		l.Infof("skip log settings under cmdline mode")
		return
	}

	// set global log root
	lopt := &logger.Option{
		Level: strings.ToLower(c.Logging.Level),
		Flags: logger.OPT_DEFAULT,
	}

	switch c.Logging.Log {
	case "stdout", "":
		l.Info("set log to stdout, rotate disabled")
		lopt.Flags |= logger.OPT_STDOUT

		if !c.Logging.DisableColor {
			lopt.Flags |= logger.OPT_COLOR
		}

	default:

		if c.Logging.Rotate > 0 {
			logger.MaxSize = c.Logging.Rotate
		}
		if c.Logging.RotateBackups > 0 {
			logger.MaxBackups = c.Logging.RotateBackups
		}

		lopt.Path = c.Logging.Log
	}

	if err := logger.InitRoot(lopt); err != nil {
		l.Errorf("set root log(options: %+#v) failed: %s", lopt, err.Error())
	} else {
		l.Infof("set root logger(options: %+#v)ok", lopt)
	}
}

// setup global host/election tags.
func (c *Config) setupGlobalTags() {
	c.parseGlobalHostTags()

	if len(c.GlobalTagsDeprecated) != 0 { // c.GlobalTags deprecated, move them to GlobalHostTags
		for k, v := range c.GlobalTagsDeprecated {
			c.GlobalHostTags[k] = v
		}
	}

	if _, ok := c.GlobalHostTags["host"]; !ok {
		c.GlobalHostTags["host"] = c.Hostname
	}

	for k, v := range c.GlobalHostTags {
		datakit.SetGlobalHostTags(k, v)
	}

	// 此处不将 host 计入 c.GlobalHostTags，因为 c.GlobalHostTags 是读取
	// 的用户配置，而 host 是不允许修改的, 故单独添加这个 tag 到 io 模块
	datakit.SetGlobalHostTags("host", c.Hostname)

	if c.Election.Enable {
		// 开启选举且开启开关的情况下，将选举的命名空间塞到 global-election-tags 中
		if c.Election.EnableNamespaceTag {
			c.Election.Tags["election_namespace"] = c.Election.Namespace
		}

		for k, v := range c.Election.Tags {
			datakit.SetGlobalElectionTags(k, v)
		}
	} else {
		// If not election, assigning the global host tags to election.
		datakit.ClearGlobalElectionTags()
		globalHostTags := datakit.GlobalHostTags()
		datakit.SetGlobalElectionTagsByMap(globalHostTags)

		// also copy global host tags to global election tags
		for k, v := range c.GlobalHostTags {
			c.Election.Tags[k] = v
		}
	}
}

func (c *Config) setupPublicWriteAPIs() {
	if len(c.HTTPAPI.PublicAPIs) == 0 {
		// default enable all data upload APIs.
		for _, cat := range point.AllCategories() {
			c.HTTPAPI.PublicAPIs = append(c.HTTPAPI.PublicAPIs, cat.URL())
		}
	}
}

// SetCommandLineMode set configure behavior in debug mode, undeer
// this mode, some actions are different with normal mode.
func (c *Config) SetCommandLineMode(on bool) {
	c.cmdlineMode = on
}

func (c *Config) ApplyMainConfig() error {
	c.setLogging()

	l = logger.SLogger("config")

	c.setupUlimit()

	if c.Hostname == "" {
		if err := c.SetHostname(); err != nil {
			return err
		}
	}

	c.setupGlobalTags()
	c.setupPublicWriteAPIs()

	if c.Dataway != nil && len(c.Dataway.URLs) > 0 {
		if err := c.setupDataway(); err != nil {
			return err
		}
	} else {
		l.Warnf("dataway not set")
	}

	datakit.AutoUpdate = c.AutoUpdate

	// config default io
	if c.IO != nil {
		if c.IO.MaxCacheCount < 1000 {
			l.Warnf("io max cache count < 1000, got %d", c.IO.MaxCacheCount)
		}
	}

	// remove deprecated UUID field in main configure
	if c.UUIDDeprecated != "" {
		c.UUIDDeprecated = "" // clear deprecated UUID field
		buf := new(bytes.Buffer)
		if err := bstoml.NewEncoder(buf).Encode(c); err != nil {
			l.Fatalf("encode main configure failed: %s", err.Error())
		}
		if err := os.WriteFile(datakit.MainConfPath, buf.Bytes(), datakit.ConfPerm); err != nil {
			l.Fatalf("refresh main configure failed: %s", err.Error())
		}

		l.Info("refresh main configure ok")
	}

	InitGitreposDir()
	// Operator 使用 ENV 初始化
	c.Operator = operator.NewOperatorClientFromEnv()
	// 初始化 ENC 密码加密功能。
	initCrypto(c)
	//
	fillPipelineConfig(c)
	return nil
}

func (c *Config) setupUlimit() {
	// Set up ulimit.
	ulimitStatus := "ok"
	ulimitVal := int64(c.Ulimit)

	if err := setUlimit(uint64(ulimitVal)); err != nil {
		l.Warnf("fail to set ulimit to %d: %v", c.Ulimit, err)
		ulimitStatus = "fail"
		ulimitVal = -1
	}

	if runtime.GOOS == "windows" {
		ulimitStatus = "N/A"
	}

	setUlimitVec.WithLabelValues(ulimitStatus).Set(float64(ulimitVal))

	soft, hard, err := getUlimit()
	if err != nil {
		l.Warnf("fail to get ulimit: %v", err)
	} else {
		l.Infof("ulimit set to softLimit = %d, hardLimit = %d", soft, hard)
	}
}

func (c *Config) SetHostname() error {
	if hn, err := c.getHostName(); err != nil {
		return err
	} else {
		c.Hostname = hn
		datakit.DatakitHostName = c.Hostname
		return nil
	}
}

func (c *Config) getHostName() (string, error) {
	// try get hostname from configure
	if v, ok := c.Environments["ENV_HOSTNAME"]; ok && v != "" {
		return v, nil
	}

	// get real hostname
	hn, err := os.Hostname()
	if err != nil {
		return "", err
	} else {
		return hn, nil
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
		if name == "-" {
			continue
		}

		if _, ok := inputsUnique[name]; !ok {
			inputsUnique[name] = true
			inputs = append(inputs, name)
		}
	}

	if len(inputs) == 0 {
		l.Warnf("no default inputs enabled!")
	}

	c.DefaultEnabledInputs = inputs
}

func ParseGlobalTags(s string) map[string]string {
	if s == "" {
		return map[string]string{}
	}

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
	return os.WriteFile(f, []byte(uuid), datakit.ConfPerm)
}

func LoadUUID(f string) (string, error) {
	if data, err := os.ReadFile(filepath.Clean(f)); err != nil {
		return "", err
	} else {
		return string(data), nil
	}
}

func emptyDir(fp string) bool {
	fd, err := os.Open(filepath.Clean(fp))
	if err != nil {
		l.Error(err)
		return false
	}

	defer fd.Close() //nolint:errcheck,gosec

	_, err = fd.ReadDir(1)
	return errors.Is(err, io.EOF)
}

// remove all xxx.conf.sample.
func removeSamples() {
	l.Debugf("searching samples under %s", datakit.ConfdDir)

	fps := SearchDir(datakit.ConfdDir, ".conf.sample", ".git")

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

func CreateSymlinks() error {
	var x [][2]string

	if runtime.GOOS == datakit.OSWindows {
		x = [][2]string{
			{
				filepath.Join(datakit.InstallDir, "datakit.exe"),
				`C:\WINDOWS\system32\datakit.exe`,
			},
		}
	} else {
		x = [][2]string{
			{
				filepath.Join(datakit.InstallDir, "datakit"),
				"/usr/local/bin/datakit",
			},

			{
				filepath.Join(datakit.InstallDir, "datakit"),
				"/usr/local/sbin/datakit",
			},

			{
				filepath.Join(datakit.InstallDir, "datakit"),
				"/sbin/datakit",
			},

			{
				filepath.Join(datakit.InstallDir, "datakit"),
				"/usr/sbin/datakit",
			},

			{
				filepath.Join(datakit.InstallDir, "datakit"),
				"/usr/bin/datakit",
			},
		}
	}

	ok := 0
	for _, item := range x {
		if err := os.MkdirAll(filepath.Dir(item[1]), os.ModePerm); err != nil {
			l.Warnf("create dir %s failed: %s, ignored", err.Error())
			continue
		}

		if err := symlink(item[0], item[1]); err != nil {
			l.Warnf("create datakit symlink: %s -> %s: %s, ignored", item[1], item[0], err.Error())
			continue
		}
		ok++
	}

	if ok == 0 {
		return fmt.Errorf("create symlink failed")
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

func GetToken() string {
	if Cfg.Dataway == nil {
		return ""
	}

	tokens := Cfg.Dataway.GetTokens()

	if len(tokens) > 0 {
		return tokens[0]
	}

	return ""
}

func GitHasEnabled() bool {
	return len(datakit.GitReposRepoName) > 0 && len(datakit.GitReposRepoFullPath) > 0
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
