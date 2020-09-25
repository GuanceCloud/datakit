package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/influxdata/toml"
	"github.com/influxdata/toml/ast"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/tailf"
	tgi "gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/telegraf_inputs"
)

func addTailfInputs(mc *datakit.MainConfig) {

	src := mc.Name
	if src == "" {
		src = mc.UUID
	}

	tailfDatakitLog := &tailf.Tailf{
		LogFiles: []string{
			mc.Log,
			datakit.AgentLogFile,
		},
		Source: src,
	}

	l.Infof("add input %+#v", tailfDatakitLog)
	_ = inputs.AddInput("tailf", tailfDatakitLog, "no-config-available")
}

// load all inputs under @InstallDir/conf.d
func LoadInputsConfig(c *datakit.Config) error {

	// detect same-name input name between datakit and telegraf
	for k := range tgi.TelegrafInputs {
		if _, ok := inputs.Inputs[k]; ok {
			l.Fatalf(fmt.Sprintf("same name input %s within datakit and telegraf", k))
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
	if c.MainCfg.LogUpload { // re-add tailf on datakit log
		addTailfInputs(c.MainCfg)
	}

	for name, creator := range inputs.Inputs {
		if err := doLoadInputConf(c, name, creator, availableInputCfgs); err != nil {
			l.Errorf("load %s config failed: %v, ignored", name, err)
			return err
		}
	}

	telegrafRawCfg, err := loadTelegrafInputsConfigs(c, availableInputCfgs, c.InputFilters)
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

func doLoadInputConf(c *datakit.Config, name string, creator inputs.Creator, inputcfgs map[string]*ast.Table) error {
	if len(c.InputFilters) > 0 {
		if !sliceContains(name, c.InputFilters) {
			return nil
		}
	}

	if name == "self" { //nolint:goconst
		inputs.AddSelf(creator())
		return nil
	}

	l.Debugf("search input cfg for %s", name)
	searchDatakitInputCfg(c, inputcfgs, name, creator)

	return nil
}

func searchDatakitInputCfg(c *datakit.Config, inputcfgs map[string]*ast.Table, name string, creator inputs.Creator) {
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

						inputlist, err = tryUnmarshal(v, name, creator)
						if err != nil {
							l.Warnf("unmarshal input %s failed within %s: %s", name, fp, err.Error())
							continue
						}

						l.Infof("load input %s from %s ok", name, fp)
					}
				}

			default: // compatible with old version: no [[inputs.xxx]] header
				inputlist, err = tryUnmarshal(node, name, creator)
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

func tryUnmarshal(tbl interface{}, name string, creator inputs.Creator) (inputList []inputs.Input, err error) {

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
			return
		}

		l.Debugf("try set MaxLifeCheckInterval from ", name)
		trySetMaxPostInterval(t)

		inputList = append(inputList, input)
	}

	return
}

func trySetMaxPostInterval(t *ast.Table) {
	var dur time.Duration
	var err error
	node, ok := t.Fields["interval"]
	if !ok {
		return
	}

	if kv, ok := node.(*ast.KeyValue); ok {
		if str, ok := kv.Value.(*ast.String); ok {
			dur, err = time.ParseDuration(str.Value)
			if err != nil {
				l.Errorf("parse duration(%s) from %+#v failed: %s, ignored", str.Value, t, err.Error())
				return
			}

			if datakit.MaxLifeCheckInterval+5*time.Second < dur { // use the max interval from all inputs
				datakit.MaxLifeCheckInterval = dur
				l.Debugf("set MaxLifeCheckInterval to %v ok", dur)
			}
		}
	}
}

func migrateOldCfg(name string, c inputs.Creator) error {
	if name == "self" { //nolint:goconst
		return nil
	}

	input := c()
	catalog := input.Catalog()

	cfgpath := filepath.Join(datakit.ConfdDir, catalog, name+".conf.sample")
	old := filepath.Join(datakit.ConfdDir, catalog, name+".conf")

	if _, err := os.Stat(old); err == nil {
		tbl, err := parseCfgFile(old)
		if err != nil {
			l.Warnf("[error] parse conf %s failed on [%s]: %s, ignored", old, name, err)
		} else if len(tbl.Fields) == 0 { // old config not used
			if err := os.Remove(old); err != nil {
				l.Errorf("Remove: %s, ignored", err.Error())
			}
		}
	}

	// overwrite old config sample
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
		if err := migrateOldCfg(name, create); err != nil {
			l.Fatal(err)
		}
	}

	// create telegraf input plugin's configures
	for name, input := range tgi.TelegrafInputs {

		cfgpath := filepath.Join(datakit.ConfdDir, input.Catalog, name+".conf.sample")
		old := filepath.Join(datakit.ConfdDir, input.Catalog, name+".conf")

		if _, err := os.Stat(old); err == nil {
			tbl, err := parseCfgFile(old)
			if err != nil {
				l.Warnf("[error] parse conf %s failed on [%s]: %s, ignored", old, name, err)
			} else if len(tbl.Fields) == 0 { // old config not used
				os.Remove(old)
			}
		}

		// overwrite old telegraf config sample
		l.Debugf("create telegraf conf path %s", filepath.Join(datakit.ConfdDir, input.Catalog))
		if err := os.MkdirAll(filepath.Join(datakit.ConfdDir, input.Catalog), os.ModePerm); err != nil {
			l.Fatalf("create catalog dir %s failed: %s", input.Catalog, err.Error())
		}

		if input, ok := tgi.TelegrafInputs[name]; ok {
			if err := ioutil.WriteFile(cfgpath, []byte(input.SampleConfig()), 0600); err != nil {
				l.Fatalf("failed to create sample configure for collector %s: %s", name, err.Error())
			}
		}
	}
}

func initDefaultEnabledPlugins(c *datakit.Config) error {

	if len(c.MainCfg.DefaultEnabledInputs) == 0 {
		return nil
	}

	fdir := "default_enabled"

	if err := os.MkdirAll(filepath.Join(datakit.ConfdDir, fdir), os.ModePerm); err != nil {
		l.Error("mkdir failed: %s, ignored", err.Error())
		return err
	}

	for _, name := range c.MainCfg.DefaultEnabledInputs {

		fpath := filepath.Join(datakit.ConfdDir, fdir, name+".conf")
		sample, err := inputs.GetSample(name)
		if err != nil {
			l.Error("failed to get %s sample, ignored", name)
			continue
		}

		if err := ioutil.WriteFile(fpath, []byte(sample), os.ModePerm); err != nil {
			l.Error("write input %s config failed: %s, ignored", name, err.Error())
			continue
		}

		l.Debugf("enable input %s ok", name)
	}

	return nil
}

func EnableInputs(inputlist string) {
	elems := strings.Split(inputlist, ",")
	if len(elems) == 0 {
		return
	}

	for _, name := range elems {
		fpath, sample, err := doEnableInput(name)
		if err != nil {
			l.Debug("enable input %s failed, ignored", name)
		}

		if err := os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			l.Error("mkdir failed: %s, ignored", err.Error())
			continue
		}

		if err := ioutil.WriteFile(fpath, []byte(sample), os.ModePerm); err != nil {
			l.Error("write input %s config failed: %s, ignored", name, err.Error())
			continue
		}
		l.Debugf("enable input %s ok", name)
	}
}

func doEnableInput(name string) (fpath, sample string, err error) {
	if i, ok := tgi.TelegrafInputs[name]; ok {
		fpath = filepath.Join(datakit.ConfdDir, i.Catalog, name+".conf")
		sample = i.SampleConfig()
		return
	}

	if c, ok := inputs.Inputs[name]; ok {
		i := c()
		sample = i.SampleConfig()

		fpath = filepath.Join(datakit.ConfdDir, i.Catalog(), name+".conf")
		return
	}

	err = fmt.Errorf("input %s not found, ignored", name)
	return
}
