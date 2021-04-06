//nolint:gocyclo
package config

import (
	"fmt"
	"io/ioutil"

	"github.com/influxdata/toml"
	"github.com/influxdata/toml/ast"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
)

var (
	l = logger.DefaultSLogger("config")
)

func LoadCfg(c *datakit.Config, mcp string) error {

	datakit.InitDirs()

	if err := c.LoadEnvs(mcp); err != nil {
		return err
	}

	if err := c.LoadMainConfig(mcp); err != nil {
		return err
	}

	// set global log root
	l.Infof("set log to %s", c.MainCfg.Log)
	logger.MaxSize = c.MainCfg.LogRotate
	logger.SetGlobalRootLogger(c.MainCfg.Log, c.MainCfg.LogLevel, logger.OPT_DEFAULT)
	l = logger.SLogger("config")

	l.Infof("main cfg: %+#v", c.MainCfg)

	initPluginSamples()
	if err := initPluginPipeline(); err != nil {
		l.Fatal(err)
	}

	initDefaultEnabledPlugins(c)

	if err := LoadInputsConfig(c); err != nil {
		l.Error(err)
		return err
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
