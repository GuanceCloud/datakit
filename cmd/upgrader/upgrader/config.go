// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package upgrader is for Datakit upgrade
package upgrader

import (
	"fmt"
	"os"

	bstoml "github.com/BurntSushi/toml"
	"github.com/GuanceCloud/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
)

var sampleConfig = `
listen = "0.0.0.0:9542"
ip_whitelist = []

[logging]
  log = "/var/log/dk_upgrader/log"
  gin_log = "/var/log/dk_upgrader/gin.log"
  gin_err_log = "/var/log/dk_upgrader/gin_err.log"
  level = "info"
  disable_color = false
  rotate = 32
`

var Cfg = DefaultMainConfig()

type LoggerCfg struct {
	*config.LoggerCfg
	GinErrLog string `toml:"gin_err_log"`
}

type MainConfig struct {
	Listen      string     `toml:"listen"`
	IPWhiteList []string   `toml:"ip_whitelist"`
	Logging     *LoggerCfg `toml:"logging"`
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
		L().Info("set log to stdout, rotate disabled")
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
		L().Errorf("set root log(options: %+#v) failed: %s", lopt, err.Error())
		return
	}

	L().Infof("set root logger(options: %+#v)ok", lopt)
}

func (c *MainConfig) CreateCfgFile() error {
	f, err := os.OpenFile(MainConfigFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, datakit.ConfPerm)
	if err != nil {
		return fmt.Errorf("unable to open main config file[%s]: %w", MainConfigFile, err)
	}

	if err := bstoml.NewEncoder(f).Encode(c); err != nil {
		return fmt.Errorf("unable to encode toml: %w", err)
	}

	return nil
}

func DefaultMainConfig() *MainConfig {
	var cfg MainConfig
	_, err := bstoml.Decode(sampleConfig, &cfg)
	if err == nil {
		return &cfg
	}

	L().Warnf("unable to decode config sample: %s", err)

	return &MainConfig{
		Listen:      "0.0.0.0:9542",
		IPWhiteList: []string{},
		Logging: &LoggerCfg{
			LoggerCfg: &config.LoggerCfg{
				Log:          "/var/log/dk_upgrader/log",
				GinLog:       "/var/log/dk_upgrader/gin.log",
				Level:        "info",
				DisableColor: false,
				Rotate:       32,
			},
			GinErrLog: "/var/log/dk_upgrader/gin_err.log",
		},
	}
}
