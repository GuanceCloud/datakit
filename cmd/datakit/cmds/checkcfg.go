package cmds

import (
	"fmt"
	"time"

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

	for field, node := range tpl.Fields {

		switch field {
		default:
			fmt.Printf("[I] ignore config %s\n", fp)
			ignored++
			return

		case "inputs": //nolint:goconst
			stbl, ok := node.(*ast.Table)
			if !ok {
				l.Warnf("ignore bad toml node within %s", fp)
			} else {
				for inputName, v := range stbl.Fields {
					if c, ok := inputs.Inputs[inputName]; !ok {
						fmt.Printf("[W] unknown input `%s' found in %s\n", inputName, fp)
						unknown++
					} else {
						if _, err = config.TryUnmarshal(v, inputName, c); err != nil {
							fmt.Printf("[E] failed to init input %s from %s:\n%s\n", inputName, fp, err.Error())
							failed++
						} else {
							passed++
						}
					}
				}
			}
		}
	}
}

func checkConfig() {
	start := time.Now()
	fps := config.SearchDir(datakit.ConfdDir, ".conf")

	for _, fp := range fps {
		tpl, err := config.ParseCfgFile(fp)
		if err != nil {
			fmt.Printf("[E] failed to parse %s: %s", fp, err.Error())
			failed++
		} else {
			checkInputCfg(tpl, fp)
		}
	}

	fmt.Printf("\n------------------------\n")
	fmt.Printf("checked %d conf, %d ignored, %d passed, %d failed, %d unknown, ",
		len(fps), ignored, passed, failed, unknown)

	fmt.Printf("cost %v\n", time.Since(start))
}
