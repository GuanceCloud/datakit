package cmds

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"runtime/debug"
	"strings"
	"time"

	"github.com/influxdata/influxdb1-client/models"
	"github.com/influxdata/toml/ast"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

func inputDebugger(configFile string) error {
	var err error
	defer func() {
		if err != nil {
			debug.PrintStack()
		}
	}()

	configPath := configFile
	if !path.IsAbs(configFile) {
		currentDir, _ := os.Getwd()
		configPath = filepath.Join(currentDir, configFile)
		if _, err = os.Stat(configPath); err != nil {
			fmt.Printf("stat failed: %v\n", err)
			return err
		}
	}

	fmt.Printf("config path: %s\n", configPath)

	tpl, err := config.ParseCfgFile(configPath)
	if err != nil {
		fmt.Printf("parse failed: %v\n", err)
		return err
	}

	for field, node := range tpl.Fields {
		switch field {
		case "inputs": //nolint:goconst
			{
				stbl, ok := node.(*ast.Table)
				if !ok {
					fmt.Println("not node!")
					continue
				}
				for inputName, tbl := range stbl.Fields {
					creator, ok := inputs.Inputs[inputName]
					if !ok {
						fmt.Printf("%s not found!\n", inputName)
						continue
					}

					ctor := creator()
					switch ctor.(type) {
					case inputs.InputV2:
					default:
						fmt.Printf("%s is not input!\n", inputName)
						continue
					}

					var inputList []inputs.Input
					if inputList, err = config.TryUnmarshal(tbl, inputName, creator); err != nil {
						fmt.Printf("%s unmarshal failed!\n", inputName)
						continue
					}

					for _, input := range inputList {
						ipt, ok := input.(inputs.InputOnceRunnable)
						if !ok {
							fmt.Printf("%s not implement for now.\n", inputName)
							continue
						}
						mpts, e := ipt.Collect()
						if e != nil {
							err = e
							fmt.Printf("%s Collect failed: %s\n", inputName, e.Error())
							return err
						}
						if err = printResultEx(mpts); err != nil {
							fmt.Printf("%s print failed: %s\n", inputName, e.Error())
							return err
						}

						if len(mpts) > 0 {
							fmt.Println("check succeeded!")
						} else {
							fmt.Println("Collect empty!")
							return fmt.Errorf("collect_empty")
						}
					}
				}
			}
		default:
		} // field
	}
	return nil
}

func printResultEx(mpts map[string][]*io.Point) error {
	fmt.Printf("\n================= Line Protocol Points ==================\n\n")
	// measurements collected
	measurements := make(map[string]string)
	timeSeries := make(map[string]string)

	ptsLen := 0

	for category, points := range mpts {
		category = filepath.Base(category)
		fmt.Printf("%s: ", category)
		ptsLen += len(points)
		for _, point := range points {
			lp := point.String()
			fmt.Printf(" %s\n", lp)

			influxPoint, err := models.ParsePointsWithPrecision([]byte(lp), time.Now(), "n")
			if len(influxPoint) != 1 {
				return fmt.Errorf("parse point error")
			}
			if err != nil {
				return err
			}
			timeSeries[fmt.Sprint(influxPoint[0].HashID())] = trueString
			name := point.Name()
			measurements[name] = trueString
		}
	}

	mKeys := make([]string, len(measurements))
	i := 0
	for name := range measurements {
		mKeys[i] = name
		i++
	}
	fmt.Printf("\n================= Summary ==================\n\n")
	fmt.Printf("Total time series: %v\n", len(timeSeries))
	fmt.Printf("Total line protocol points: %v\n", ptsLen)
	fmt.Printf("Total measurements: %v (%s)\n\n", len(measurements), strings.Join(mKeys, ", "))

	return nil
}
