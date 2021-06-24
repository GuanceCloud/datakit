package cmds

import (
	"fmt"
	"time"

	"github.com/influxdata/toml/ast"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

func CheckConfig() {
	start := time.Now()
	fps := config.SearchDir(datakit.ConfdDir, ".conf")

	failed := 0
	for _, fp := range fps {
		tpl, err := config.ParseCfgFile(fp)
		if err != nil {
			fmt.Printf("[E] failed to parse %s:\n%s", fp, err.Error())
			failed++
			continue
		}

		for field, node := range tpl.Fields {

			switch field {
			case "inputs": //nolint:goconst
				stbl, ok := node.(*ast.Table)
				if !ok {
					l.Warnf("ignore bad toml node within %s", fp)
				} else {
					for inputName, v := range stbl.Fields {
						if c, ok := inputs.Inputs[inputName]; !ok {
							fmt.Printf("[W] unknown input %s found in %s\n", inputName, fp)
						} else {
							if _, err = config.TryUnmarshal(v, inputName, c); err != nil {
								fmt.Printf("[E] failed to init input %s from %s:\n%s\n", inputName, fp, err.Error())
								failed++
							}
						}
					}
				}
			}
		}
	}

	fmt.Println("------------------------")
	fmt.Printf("checked %d conf, ", len(fps))

	if failed > 0 {
		fmt.Printf("%d failed, ", failed)
	} else {
		fmt.Printf("all passing, ")
	}

	fmt.Printf("cost %v\n", time.Since(start))
}
