package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/influxdata/toml"
	"github.com/influxdata/toml/ast"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

// load all inputs under @InstallDir/conf.d
func (c *Config) LoadConfig() error {

	// detect same-name input name between datakit and telegraf
	for k, _ := range TelegrafCfgSamples {
		if _, ok := inputs.Inputs[k]; ok {
			panic(fmt.Sprintf("same name input %s within datakit and telegraf", k))
		}
	}

	availableInputCfgs := map[string]*ast.Table{}

	if err := filepath.Walk(datakit.ConfdDir, func(fp string, f os.FileInfo, err error) error {
		if err != nil {
			l.Error(err)
		}

		if f.IsDir() {
			l.Debugf("ignore dir %s", fp)
			return nil
		}

		if !strings.HasSuffix(f.Name(), ".conf") {
			l.Debugf("ignore non-conf %s", fp)
			return nil
		}

		tbl, err := parseCfgFile(fp)
		if err != nil {
			l.Warnf("[error] parse conf %s failed: %s, ignored", fp, err)
			return nil
		}

		if len(tbl.Fields) == 0 {
			l.Debugf("no conf available on %s", fp)
			return nil
		}

		availableInputCfgs[fp] = tbl
		return nil
	}); err != nil {
		l.Error(err)
		return err
	}

	for name, creator := range inputs.Inputs {
		if err := c.doLoadInputConf(name, creator, availableInputCfgs); err != nil {
			l.Errorf("load %s config failed: %v, ignored", name, err)
			return err
		}
	}

	telegrafRawCfg, err := c.loadTelegrafConfigs(availableInputCfgs, c.InputFilters)
	if err != nil {
		return err
	}

	if telegrafRawCfg != "" {
		if err := ioutil.WriteFile(filepath.Join(datakit.TelegrafDir, "agent.conf"), []byte(telegrafRawCfg), os.ModePerm); err != nil {
			l.Errorf("create telegraf conf failed: %s", err.Error())
			return err
		}
	}

	return nil
}

func (c *Config) doLoadInputConf(name string, creator inputs.Creator, inputcfgs map[string]*ast.Table) error {
	if len(c.InputFilters) > 0 {
		if !sliceContains(name, c.InputFilters) {
			return nil
		}
	}

	if name == "self" {
		c.Inputs[name] = append(c.Inputs[name], creator())
		return nil
	}

	l.Debugf("search input cfg for %s", name)
	c.searchDatakitInputCfg(inputcfgs, name, creator)

	return nil
}

func (c *Config) searchDatakitInputCfg(inputcfgs map[string]*ast.Table, name string, creator inputs.Creator) {
	for fp, tbl := range inputcfgs {

		for field, node := range tbl.Fields {
			switch field {
			case "inputs":
				tbl_, ok := node.(*ast.Table)
				if !ok {
					l.Warnf("ignore bad toml node for %s within %s", name, fp)
				} else {
					for inputName, v := range tbl_.Fields {
						if inputName != name {
							l.Debugf("input %s ignore input %s", name, inputName)
							continue
						}

						if err := c.tryUnmarshal(v, name, creator); err != nil {
							l.Warnf("unmarshal input %s failed within %s: %s", name, fp, err.Error())
							continue
						}

						l.Infof("load input %s from %s ok", name, fp)
					}
				}

			default:
				if err := c.tryUnmarshal(node, name, creator); err != nil {
					l.Warnf("unmarshal input %s failed within %s: %s", name, fp, err.Error())
				} else {
					l.Infof("load input %s from %s ok", name, fp)
				}
			}
		}
	}
}

func (c *Config) tryUnmarshal(tbl interface{}, name string, creator inputs.Creator) error {

	tbls := []*ast.Table{}

	switch tbl.(type) {
	case []*ast.Table:
		tbls = tbl.([]*ast.Table)
	case *ast.Table:
		tbls = append(tbls, tbl.(*ast.Table))
	default:
		return fmt.Errorf("invalid toml format on %s: %v", name, reflect.TypeOf(tbl))
	}

	for _, t := range tbls {
		input := creator()

		if err := toml.UnmarshalTable(t, input); err != nil {
			l.Errorf("toml unmarshal %s failed: %v", name, err)
			return err
		}

		if err := c.addInput(name, input, t); err != nil {
			l.Error("add %s failed: %v", name, err)
			return err
		}
	}

	return nil
}

// Creata datakit input plugin's configures if not exists
func initPluginCfgs() {
	for name, create := range inputs.Inputs {
		if name == "self" {
			continue
		}

		input := create()
		catalog := input.Catalog()

		cfgpath := filepath.Join(datakit.ConfdDir, catalog, name+".conf.sample")
		old := filepath.Join(datakit.ConfdDir, catalog, name+".conf")

		if _, err := os.Stat(old); err == nil {
			tbl, err := parseCfgFile(old)
			if err != nil {
				l.Warnf("[error] parse conf %s failed on [%s]: %s, ignored", old, name, err)
			} else {
				if len(tbl.Fields) == 0 { // old config not used
					os.Remove(old)
				}
			}
		}

		// overwrite old config sample
		l.Debugf("create datakit conf path %s", filepath.Join(datakit.ConfdDir, catalog))
		if err := os.MkdirAll(filepath.Join(datakit.ConfdDir, catalog), os.ModePerm); err != nil {
			l.Fatalf("create catalog dir %s failed: %s", catalog, err.Error())
		}

		sample := input.SampleConfig()
		if sample == "" {
			l.Fatalf("no sample available on collector %s", name)
		}

		if err := ioutil.WriteFile(cfgpath, []byte(sample), 0644); err != nil {
			l.Fatalf("failed to create sample configure for collector %s: %s", name, err.Error())
		}
	}

	// create telegraf input plugin's configures
	for name, input := range TelegrafInputs {

		cfgpath := filepath.Join(datakit.ConfdDir, input.Catalog, name+".conf.sample")
		old := filepath.Join(datakit.ConfdDir, input.Catalog, name+".conf")

		if _, err := os.Stat(old); err == nil {
			tbl, err := parseCfgFile(old)
			if err != nil {
				l.Warnf("[error] parse conf %s failed on [%s]: %s, ignored", old, name, err)
			} else {
				if len(tbl.Fields) == 0 { // old config not used
					os.Remove(old)
				}
			}
		}

		// overwrite old telegraf config sample
		l.Debugf("create telegraf conf path %s", filepath.Join(datakit.ConfdDir, input.Catalog))
		if err := os.MkdirAll(filepath.Join(datakit.ConfdDir, input.Catalog), os.ModePerm); err != nil {
			l.Fatalf("create catalog dir %s failed: %s", input.Catalog, err.Error())
		}

		if sample, ok := TelegrafCfgSamples[name]; ok {
			if err := ioutil.WriteFile(cfgpath, []byte(sample), 0644); err != nil {
				l.Fatalf("failed to create sample configure for collector %s: %s", name, err.Error())
			}
		}
	}
}
