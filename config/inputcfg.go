package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/influxdata/toml"
	"github.com/influxdata/toml/ast"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

// load all inputs under @InstallDir/conf.d
func LoadInputsConfig(c *Config) error {

	availableInputCfgs := map[string]*ast.Table{}

	if err := filepath.Walk(datakit.ConfdDir, func(fp string, f os.FileInfo, err error) error {
		if err != nil {
			l.Error(err)
		}

		if f.IsDir() {
			l.Debugf("ignore dir %s", fp)
			return nil
		}

		if f.Name() == "datakit.conf" {
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

		deprecates := checkDepercatedInputs(tbl, deprecatedInputs)
		if len(deprecates) > 0 {
			for k, v := range deprecates {
				l.Warnf("input `%s' removed, please use %s instead", k, v)
			}
		}

		if len(tbl.Fields) == 0 {
			l.Debugf("no conf available on %s", fp)
			return nil
		}

		l.Debugf("parse %s ok", fp)
		availableInputCfgs[fp] = tbl
		return nil
	}); err != nil {
		l.Error(err)
		return err
	}

	// reset inputs(for reloading)
	l.Debug("reset inputs")
	inputs.ResetInputs()

	for name, creator := range inputs.Inputs {
		if !datakit.Enabled(name) {
			l.Debugf("LoadInputsConfig: ignore unchecked input %s", name)
			continue
		}

		if err := doLoadInputConf(c, name, creator, availableInputCfgs); err != nil {
			l.Errorf("load %s config failed: %v, ignored", name, err)
			return err
		}
	}

	inputs.AddSelf()

	return nil
}

func doLoadInputConf(c *Config, name string, creator inputs.Creator, inputcfgs map[string]*ast.Table) error {

	l.Debugf("search input cfg for %s", name)
	searchDatakitInputCfg(c, inputcfgs, name, creator)

	return nil
}

func searchDatakitInputCfg(c *Config,
	inputcfgs map[string]*ast.Table,
	name string,
	creator inputs.Creator) {
	var err error

	for fp, tbl := range inputcfgs {
		for field, node := range tbl.Fields {
			inputlist := []inputs.Input{}

			switch field {
			case "inputs": //nolint:goconst
				stbl, ok := node.(*ast.Table)
				if !ok {
					l.Warnf("ignore bad toml node for %s within %s", name, fp)
				} else {
					for inputName, v := range stbl.Fields {
						if inputName != name {
							continue
						}
						inputlist, err = TryUnmarshal(v, inputName, creator)
						if err != nil {
							l.Warnf("unmarshal input %s failed within %s: %s", inputName, fp, err.Error())
							continue
						}

						l.Infof("load input %s from %s ok", inputName, fp)
					}
				}

			default: // compatible with old version: no [[inputs.xxx]] header
				inputlist, err = TryUnmarshal(node, name, creator)
				if err != nil {
					l.Warnf("unmarshal input %s failed within %s: %s", name, fp, err.Error())
				}
			}

			for _, i := range inputlist {
				if err := inputs.AddInput(name, i, fp); err != nil {
					l.Error("add %s failed: %v", name, err)
					continue
				}

				l.Infof("add input %s(%s) ok", name, fp)
			}
		}
	}
}

func isDisabled(wlists, blists []*InputHostList, hostname, name string) bool {

	for _, bl := range blists {
		if bl.MatchHost(hostname) && bl.MatchInput(name) {
			return true // 一旦上榜，无脑屏蔽
		}
	}

	// 如果采集器在白名单中，但对应的 host 不在白名单，则屏蔽掉
	// 如果采集器在白名单中，对应的 host 在白名单，放行
	// 如果采集器不在白名单中，不管 host 情况，一律放行
	if len(wlists) > 0 {
		for _, wl := range wlists {
			if wl.MatchInput(name) { // 说明@name有白名单限制
				if wl.MatchHost(hostname) {
					return false
				} else { // 不在白名单中的 host，屏蔽掉
					return true
				}
			}
		}
	}
	return false
}

func TryUnmarshal(tbl interface{}, name string, creator inputs.Creator) (inputList []inputs.Input, err error) {

	tbls := []*ast.Table{}

	switch t := tbl.(type) {
	case []*ast.Table:
		tbls = tbl.([]*ast.Table)
	case *ast.Table:
		tbls = append(tbls, tbl.(*ast.Table))
	default:
		err = fmt.Errorf("invalid toml format on %s: %v", name, t)
		return
	}

	for _, t := range tbls {
		input := creator()
		err = toml.UnmarshalTable(t, input)
		if err != nil {
			l.Errorf("toml unmarshal %s failed: %v", name, err)
			continue
		}

		inputList = append(inputList, input)
	}

	return
}

func initDatakitConfSample(name string, c inputs.Creator) error {
	if name == "self" { //nolint:goconst
		return nil
	}

	input := c()
	catalog := input.Catalog()

	cfgpath := filepath.Join(datakit.ConfdDir, catalog, name+".conf.sample")
	l.Debugf("create datakit conf path %s", filepath.Join(datakit.ConfdDir, catalog))
	if err := os.MkdirAll(filepath.Join(datakit.ConfdDir, catalog), os.ModePerm); err != nil {
		l.Errorf("create catalog dir %s failed: %s", catalog, err.Error())
		return err
	}

	sample := input.SampleConfig()
	if sample == "" {
		return fmt.Errorf("no sample available on collector %s", name)
	}

	if err := ioutil.WriteFile(cfgpath, []byte(sample), 0600); err != nil {
		l.Errorf("failed to create sample configure for collector %s: %s", name, err.Error())
		return err
	}

	return nil
}

// Creata datakit input plugin's configures if not exists
func initPluginSamples() {
	for name, create := range inputs.Inputs {

		if !datakit.Enabled(name) {
			l.Debugf("initPluginSamples: ignore unchecked input %s", name)
			continue
		}

		if err := initDatakitConfSample(name, create); err != nil {
			l.Fatal(err)
		}
	}
}

func initDefaultEnabledPlugins(c *Config) {

	if len(c.DefaultEnabledInputs) == 0 {
		return
	}

	for _, name := range c.DefaultEnabledInputs {
		var fpath, sample string

		if c, ok := inputs.Inputs[name]; ok {
			i := c()
			sample = i.SampleConfig()

			fpath = filepath.Join(datakit.ConfdDir, i.Catalog(), name+".conf")
		} else {
			l.Warnf("input %s not found, ignored", name)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			l.Errorf("mkdir failed: %s, ignored", err.Error())
			continue
		}

		//check exist
		if _, err := os.Stat(fpath); err == nil {
			continue
		}

		if err := ioutil.WriteFile(fpath, []byte(sample), os.ModePerm); err != nil {
			l.Errorf("write input %s config failed: %s, ignored", name, err.Error())
			continue
		}

		l.Infof("enable input %s ok", name)
	}
}

func LoadInputConfigFile(f string, creator inputs.Creator) ([]inputs.Input, error) {

	tbl, err := parseCfgFile(f)
	if err != nil {
		return nil, fmt.Errorf("[error] parse conf failed: %s", err)
	}

	inputlist := []inputs.Input{}

	for field, node := range tbl.Fields {

		switch field {
		case "inputs": //nolint:goconst
			stbl, ok := node.(*ast.Table)
			if ok {
				for inputName, v := range stbl.Fields {
					inputlist, err = TryUnmarshal(v, inputName, creator)
					if err != nil {
						return nil, fmt.Errorf("unmarshal input failed, %s", err.Error())
					}
				}
			}

		default: // compatible with old version: no [[inputs.xxx]] header
			inputlist, err = TryUnmarshal(node, "", creator)
			if err != nil {
				return nil, fmt.Errorf("unmarshal input failed: %s", err.Error())
			}
		}
	}

	return inputlist, nil
}

var deprecatedInputs = map[string]string{
	"dockerlog":         "docker",
	"docker_containers": "docker",
}

func checkDepercatedInputs(tbl *ast.Table, entries map[string]string) (res map[string]string) {

	res = map[string]string{}

	for _, node := range tbl.Fields {
		stbl, ok := node.(*ast.Table)
		if !ok {
			continue
		}
		for inputName := range stbl.Fields {
			if x, ok := entries[inputName]; ok {
				res[inputName] = x
			}
		}
	}
	return
}
