//nolint:gocyclo
package config

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/influxdata/toml"
	"github.com/influxdata/toml/ast"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
)

var (
	l = logger.DefaultSLogger("config")
)

func LoadCfg(c *datakit.Config) error {

	datakit.InitDirs()

	if err := c.LoadEnvs(); err != nil {
		return err
	}

	if err := c.LoadMainConfig(); err != nil {
		return err
	}

	l.Infof("set log to %s", c.MainCfg.Log)

	// set global log root
	logger.MaxSize = c.MainCfg.LogRotate
	logger.SetGlobalRootLogger(c.MainCfg.Log, c.MainCfg.LogLevel, logger.OPT_DEFAULT)
	l = logger.SLogger("config")

	l.Infof("main cfg: %+#v", c.MainCfg)

	initPluginSamples()
	initDefaultEnabledPlugins(c)

	if err := LoadInputsConfig(c); err != nil {
		l.Error(err)
		return err
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
