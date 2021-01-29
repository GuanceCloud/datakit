package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

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

	availableInput := map[string]map[string]*ast.Table{}
	availableTgiInput := map[string]map[string]*ast.Table{}

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

		fileName := strings.TrimSuffix(filepath.Base(fp), path.Ext(fp))

		if _, ok := inputs.Inputs[fileName]; ok {
			availableInput[fileName] = map[string]*ast.Table{fp: tbl}
			return nil
		}
		if _, ok := tgi.TelegrafInputs[fileName]; ok {
			availableTgiInput[fileName] = map[string]*ast.Table{fp: tbl}
			return nil
		}
		l.Errorf("config:%s must name by input.conf example :disk.conf", fp)
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

	for name, available := range availableInput {
		creator, _ := inputs.Inputs[name]

		if isDisabled(c.MainCfg.WhiteList, c.MainCfg.BlackList, c.MainCfg.Hostname, name) {
			l.Warnf("input `%s' banned by white/black list on `%s'", name, c.MainCfg.Hostname)
			continue
		}

		if err := doLoadInputConf(c, name, creator, available); err != nil {
			l.Errorf("load %s config failed: %v, ignored", name, err)
			return err
		}
	}

	tgiInput := map[string]*ast.Table{}

	for _, v := range availableTgiInput {
		for fp, tbl := range v {
			tgiInput[fp] = tbl
		}
	}

	telegrafRawCfg, err := loadTelegrafInputsConfigs(c, tgiInput, c.InputFilters)
	if err != nil {
		return err
	}
	if telegrafRawCfg != "" {
		if err := ioutil.WriteFile(filepath.Join(datakit.TelegrafDir, "agent.conf"), []byte(telegrafRawCfg), os.ModePerm); err != nil {
			l.Errorf("create telegraf conf failed: %s", err.Error())
			return err
		}
	}

	inputs.AddSelf()
	if len(inputs.InputsInfo["telegraf_http"]) == 0 && inputs.HaveTelegrafInputs() {
		inputs.AddTelegrafHTTP()
	}

	return nil
}

func doLoadInputConf(c *datakit.Config, name string, creator inputs.Creator, inputcfgs map[string]*ast.Table) error {
	if len(c.InputFilters) > 0 {
		if !sliceContains(name, c.InputFilters) {
			return nil
		}
	}

	l.Debugf("search input cfg for %s", name)
	searchDatakitInputCfg(c, inputcfgs, name, creator)

	return nil
}

func searchDatakitInputCfg(c *datakit.Config,
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
						//if inputName != name {
						//	continue
						//}
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

func isDisabled(wlists, blists []*datakit.InputHostList, hostname, name string) bool {

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
			return
		}

		inputList = append(inputList, input)
	}

	return
}

func migrateOldCfg(name string, c inputs.Creator) error {
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
		if err := migrateOldCfg(name, create); err != nil {
			l.Fatal(err)
		}
	}

	// create telegraf input plugin's configures
	for name, input := range tgi.TelegrafInputs {

		cfgpath := filepath.Join(datakit.ConfdDir, input.Catalog, name+".conf.sample")
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

func initDefaultEnabledPlugins(c *datakit.Config) {

	if len(c.MainCfg.DefaultEnabledInputs) == 0 {
		return
	}

	for _, name := range c.MainCfg.DefaultEnabledInputs {
		var fpath, sample string

		if i, ok := tgi.TelegrafInputs[name]; ok {
			fpath = filepath.Join(datakit.ConfdDir, i.Catalog, name+".conf")
			sample = i.SampleConfig()
		} else if c, ok := inputs.Inputs[name]; ok {
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

		if err := ioutil.WriteFile(fpath, []byte(sample), os.ModePerm); err != nil {
			l.Errorf("write input %s config failed: %s, ignored", name, err.Error())
			continue
		}

		l.Infof("enable input %s ok", name)
	}
}

func LoadInputConfig(data []byte, creator inputs.Creator) ([]inputs.Input, error) {

	tbl, err := toml.Parse(data)
	if err != nil {
		return nil, err
	}

	var result []inputs.Input

	for field, node := range tbl.Fields {
		inputlist := []inputs.Input{}

		switch field {
		case "inputs": //nolint:goconst
			stbl, ok := node.(*ast.Table)
			if !ok {
				return nil, fmt.Errorf("ignore bad toml node")
			}
			for inputName, v := range stbl.Fields {
				//if inputName != name {
				//	continue
				//}
				inputlist, err = TryUnmarshal(v, inputName, creator)
				if err != nil {
					return nil, fmt.Errorf("unmarshal input %s failed: %s", inputName, err.Error())
				}
			}

		default: // compatible with old version: no [[inputs.xxx]] header
			inputlist, err = TryUnmarshal(node, "", creator)
			if err != nil {
				return nil, fmt.Errorf("unmarshal input failed: %s", err.Error())
			}
		}

		for _, i := range inputlist {

			result = append(result, i)
		}
	}

	return result, nil
}
