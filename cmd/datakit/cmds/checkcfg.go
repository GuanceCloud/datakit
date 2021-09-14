package cmds

import (
	"time"

	"github.com/influxdata/toml"
	"github.com/influxdata/toml/ast"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

var (
	failed  = 0
	unknown = 0
	passed  = 0
	ignored = 0
)

func checkInputCfg(tpl *ast.Table, fp string) {

	var err error

	if len(tpl.Fields) == 0 {
		warnf("[E] no content in %s\n", fp)
		failed++
		return
	}

	for field, node := range tpl.Fields {

		switch field {
		default:
			infof("[I] ignore config %s\n", fp)
			ignored++
			return

		case "inputs": //nolint:goconst
			stbl, ok := node.(*ast.Table)
			if !ok {
				l.Warnf("ignore bad toml node within %s", fp)
			} else {
				for inputName, v := range stbl.Fields {
					if c, ok := inputs.Inputs[inputName]; !ok {
						warnf("[W] unknown input `%s' found in %s\n", inputName, fp)
						unknown++
					} else {
						if _, err = config.TryUnmarshal(v, inputName, c); err != nil {
							errorf("[E] failed to init input %s from %s:\n%s\n", inputName, fp, err.Error())
							failed++
						} else {
							if FlagVVV {
								infof("[OK] %s/%s\n", inputName, fp)
							}
							passed++
						}
					}
				}
			}
		}
	}
}

// check samples of every inputs
func checkSample() {
	start := time.Now()

	for k, c := range inputs.Inputs {
		i := c()

		if k == "self" {
			continue
		}

		tpl, err := toml.Parse([]byte(i.SampleConfig()))
		if err != nil {
			errorf("[E] failed to parse %s: %s", k, err.Error())
			failed++
		} else {
			checkInputCfg(tpl, k)
		}
	}

	infof("\n------------------------\n")
	infof("checked %d sample, %d ignored, %d passed, %d failed, %d unknown, ",
		len(inputs.Inputs), ignored, passed, failed, unknown)

	infof("cost %v\n", time.Since(start))
}

func checkConfig() {
	start := time.Now()
	fps := config.SearchDir(datakit.ConfdDir, ".conf")

	for _, fp := range fps {
		tpl, err := config.ParseCfgFile(fp)
		if err != nil {
			errorf("[E] failed to parse %s: %s", fp, err.Error())
			failed++
		} else {
			checkInputCfg(tpl, fp)
		}
	}

	infof("\n------------------------\n")
	infof("checked %d conf, %d ignored, %d passed, %d failed, %d unknown, ",
		len(fps), ignored, passed, failed, unknown)

	infof("cost %v\n", time.Since(start))
}
