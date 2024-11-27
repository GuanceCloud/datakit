// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package upgrader is for Datakit upgrade
package upgrader

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	bstoml "github.com/BurntSushi/toml"
	"github.com/GuanceCloud/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
)

var Cfg = defaultMainConfig()

type LoggerCfg struct {
	*config.LoggerCfg
	GinErrLog string `toml:"gin_err_log"`
}

type MainConfig struct {
	Listen string `toml:"listen"`
	Proxy  string `toml:"proxy"`

	DatakitAPIHTTPS  bool   `toml:"datakit_api_https"`
	DatakitAPIListen string `toml:"datakit_api_listen"`

	IPWhiteList      []string   `toml:"ip_whitelist"`
	InstallerBaseURL string     `toml:"install_base_url"`
	Logging          *LoggerCfg `toml:"logging"`
	InstallDir       string     `toml:"install_dir"`
	Username         string     `toml:"username"`
	DCAWebsocketURL  string     `toml:"dca_websocket_url"`

	upgradeUpgraderService,
	dkUpgrade,
	installOnly bool
}

type UpgraderOpt func(*MainConfig)

func WithProxy(proxy string) UpgraderOpt {
	return func(mc *MainConfig) {
		mc.Proxy = proxy
	}
}

func WithUpgradeService(on bool) UpgraderOpt {
	return func(mc *MainConfig) {
		mc.upgradeUpgraderService = on
	}
}

func WithDKUpgrade(on bool) UpgraderOpt {
	return func(mc *MainConfig) {
		mc.dkUpgrade = on
	}
}

func WithInstallOnly(on bool) UpgraderOpt {
	return func(mc *MainConfig) {
		mc.installOnly = on
	}
}

func WithInstallUserName(name string) UpgraderOpt {
	return func(mc *MainConfig) {
		mc.Username = name
	}
}

func WithListen(listen string) UpgraderOpt {
	return func(mc *MainConfig) {
		mc.Listen = listen
	}
}

func WithDatakitAPIHTTPS(on bool) UpgraderOpt {
	return func(mc *MainConfig) {
		mc.DatakitAPIHTTPS = on
	}
}

func WithDatakitAPIListen(listen string) UpgraderOpt {
	return func(mc *MainConfig) {
		mc.DatakitAPIListen = listen
	}
}

func WithIPWhiteList(list []string) UpgraderOpt {
	return func(mc *MainConfig) {
		mc.IPWhiteList = list
	}
}

func WithInstallBaseURL(bu string) UpgraderOpt {
	return func(mc *MainConfig) {
		mc.InstallerBaseURL = bu
	}
}

func WithLoggingCfg(lc *LoggerCfg) UpgraderOpt {
	return func(mc *MainConfig) {
		mc.Logging = lc
	}
}

func (c *MainConfig) LoadMainTOML(f string) error {
	if _, err := bstoml.DecodeFile(f, c); err != nil {
		return fmt.Errorf("bstoml.Decode: %w", err)
	}

	return nil
}

func (c *MainConfig) SetLogging() {
	// set global log root
	lopt := &logger.Option{
		Level: c.Logging.Level,
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

		lopt.Path = c.Logging.Log
	}

	if err := logger.InitRoot(lopt); err != nil {
		l.Errorf("set root log(options: %+#v) failed: %s", lopt, err.Error())
		return
	}

	l.Infof("set root logger(options: %+#v)ok", lopt)
}

func (c *MainConfig) createCfgFile() error {
	f, err := os.OpenFile(MainConfigFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, datakit.ConfPerm)
	if err != nil {
		return fmt.Errorf("unable to open main config file[%s]: %w", MainConfigFile, err)
	}

	if err := bstoml.NewEncoder(f).Encode(c); err != nil {
		return fmt.Errorf("unable to encode toml: %w", err)
	}

	return nil
}

func defaultMainConfig() *MainConfig {
	conf := &MainConfig{
		Listen:           "0.0.0.0:9542",
		IPWhiteList:      []string{},
		InstallerBaseURL: "",
		Logging: &LoggerCfg{
			LoggerCfg: &config.LoggerCfg{
				Log:           "/var/log/dk_upgrader/log",
				GinLog:        "/var/log/dk_upgrader/gin.log",
				Level:         "info",
				DisableColor:  false,
				Rotate:        logger.MaxSize,
				RotateBackups: logger.MaxBackups,
			},
			GinErrLog: "/var/log/dk_upgrader/gin_err.log",
		},

		InstallDir: InstallDir,
	}

	if runtime.GOOS == datakit.OSWindows {
		conf.Logging.LoggerCfg.Log = filepath.Join(conf.InstallDir, "log")
		conf.Logging.LoggerCfg.GinLog = filepath.Join(conf.InstallDir, "gin.log")
		conf.Logging.GinErrLog = filepath.Join(conf.InstallDir, "gin_err.log")
	}

	return conf
}
