package cmds

import (
	"fmt"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/influxdata/toml"
	"github.com/influxdata/toml/ast"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

var (
	failed  = []string{}
	unknown = []string{}
	ignored = []string{}
	passed  = 0
	checked = 0
)

func checkInputCfg(tpl *ast.Table, fp string) {
	var err error

	if len(tpl.Fields) == 0 {
		warnf("[W] no content in %s\n", fp)
		ignored = append(ignored, fp)
		return
	}

	for field, node := range tpl.Fields {
		switch field {
		default:
			infof("[I] ignore config %s\n", fp)
			ignored = append(ignored, fp)
			return

		case "inputs": //nolint:goconst
			stbl, ok := node.(*ast.Table)
			if !ok {
				l.Warnf("ignore bad toml node within %s", fp)
			} else {
				for inputName, v := range stbl.Fields {
					if c, ok := inputs.Inputs[inputName]; !ok {
						warnf("[W] unknown input `%s' found in %s\n", inputName, fp)
						unknown = append(unknown, fp)
					} else {
						if _, err = config.TryUnmarshal(v, inputName, c); err != nil {
							errorf("[E] failed to init input %s from %s:\n%s\n", inputName, fp, err.Error())
							failed = append(failed, fp+": "+err.Error())
						} else {
							if FlagVVV {
								output("[OK] %s/%s\n", inputName, fp)
							}
							passed++
						}
					}
				}
			}
		}
	}
}

func showCheckResult() {
	infof("\n------------------------\n")
	infof("checked %d samples, %d ignored, %d passed, %d failed, %d unknown, ",
		checked, len(ignored), passed, len(failed), len(unknown))

	if len(ignored) > 0 {
		infof("ignored:\n")
		for _, x := range ignored {
			infof("\t%s\n", x)
		}
	}

	if len(unknown) > 0 {
		infof("unknown:\n")
		for _, x := range unknown {
			warnf("\t%s\n", x)
		}
	}

	if len(failed) > 0 {
		infof("failed:\n")
		for _, x := range failed {
			errorf("\t%s\n", x)
		}
	}
}

// check samples of every inputs.
func checkSample() error {
	failed = []string{}
	unknown = []string{}
	passed = 0
	checked = len(inputs.Inputs)
	ignored = []string{}

	for k, c := range inputs.Inputs {
		i := c()

		if k == datakit.DatakitInputName {
			warnf("[W] ignore self input\n")
			ignored = append(ignored, k)
			continue
		}

		tpl, err := toml.Parse([]byte(i.SampleConfig()))
		if err != nil {
			errorf("[E] failed to parse %s: %s", k, err.Error())
			failed = append(failed, k+": "+err.Error())
		} else {
			checkInputCfg(tpl, k)
		}
	}

	showCheckResult()

	if len(failed) > 0 {
		return fmt.Errorf("load %v sample failed", failed)
	}
	return nil
}

func checkConfig(dir, suffix string) error {
	fps := config.SearchDir(dir, suffix)

	failed = []string{}
	unknown = []string{}
	passed = 0
	checked = 0
	ignored = []string{}

	for _, fp := range fps {
		// Skip hidden files.
		if strings.HasPrefix(filepath.Base(fp), ".") {
			continue
		}
		tpl, err := config.ParseCfgFile(fp)
		if err != nil {
			errorf("[E] failed to parse %s: %s, %s", fp, err.Error(), reflect.TypeOf(err))
			failed = append(failed, fp+": "+err.Error())
		} else {
			checkInputCfg(tpl, fp)
		}
		checked++
	}

	showCheckResult()

	if len(failed) > 0 {
		return fmt.Errorf("load %d conf failed", len(failed))
	}

	return nil
}
