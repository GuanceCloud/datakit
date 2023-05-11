// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package cmds

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"runtime/debug"
	"strings"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"github.com/influxdata/influxdb1-client/models"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	cp "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/colorprint"
	dkpt "gitlab.jiagouyun.com/cloudcare-tools/datakit/io/point"
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
			cp.Warnf("[W] stat failed: %v\n", err)
			return err
		}
	}

	cp.Infof("[I] config path: %s\n", configPath)

	inputsInstance, err := config.LoadSingleConfFile(configPath, inputs.Inputs, false)
	if err != nil {
		cp.Errorf("[E] parse failed: %v\n", err)
		return err
	}

	for k, arr := range inputsInstance {
		for _, x := range arr {
			switch ipt := x.(type) {
			case inputs.InputOnceRunnableCollect: // DEPRECATED.
				mpts, e := ipt.Collect()
				if e != nil {
					err = e
					cp.Warnf("[W] %s Collect failed: %s\n", k, e.Error())
					return err
				}
				if err = printResultEx(mpts); err != nil {
					cp.Warnf("[W] %s print failed: %s\n", k, e.Error())
					return err
				}

				if len(mpts) > 0 {
					fmt.Println("check succeeded!")
				} else {
					fmt.Println("Collect empty!")
					return fmt.Errorf("collect_empty")
				}

			case inputs.InputOnceRunnableCollectV2:
				mpts, e := ipt.Collect()
				if e != nil {
					err = e
					cp.Warnf("[W] %s Collect failed: %s\n", k, e.Error())
					return err
				}
				if err = printResultV2(mpts); err != nil {
					cp.Warnf("[W] %s print failed: %s\n", k, e.Error())
					return err
				}

				if len(mpts) > 0 {
					fmt.Println("check succeeded!")
				} else {
					fmt.Println("Collect empty!")
					return fmt.Errorf("collect_empty")
				}

			default:
				cp.Warnf("[W] %s not implement for now.\n", k)
				continue
			}
		}
	}

	return nil
}

func printResultEx(mpts map[string][]*dkpt.Point) error {
	fmt.Printf("\n================= Line Protocol Points ==================\n\n")
	// measurements collected
	measurements := make(map[string]string)
	timeSeries := make(map[string]string)

	ptsLen := 0

	for category, points := range mpts {
		category = filepath.Base(category)
		fmt.Printf("%s: ", category)
		ptsLen += len(points)

		for _, pt := range points {
			lp := pt.String()
			fmt.Println(lp)

			influxPoint, err := models.ParsePointsWithPrecision([]byte(lp), time.Now(), "n")
			if len(influxPoint) != 1 {
				return fmt.Errorf("parse point error")
			}

			if err != nil {
				return err
			}
			timeSeries[fmt.Sprint(influxPoint[0].HashID())] = trueString
			name := pt.Name()
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

func printResultV2(mpts map[point.Category][]*point.Point) error {
	fmt.Printf("\n================= Line Protocol Points ==================\n\n")
	// measurements collected
	measurements := make(map[string]string)
	timeSeries := make(map[string]string)

	ptsLen := 0

	for cate, points := range mpts {
		category := cate.String()
		fmt.Printf("%s: ", category)
		ptsLen += len(points)

		for _, pt := range points {
			lp := pt.LineProto()
			fmt.Println(lp)

			influxPoint, err := models.ParsePointsWithPrecision([]byte(lp), time.Now(), "n")
			if len(influxPoint) != 1 {
				return fmt.Errorf("parse point error")
			}

			if err != nil {
				return err
			}
			timeSeries[fmt.Sprint(influxPoint[0].HashID())] = trueString
			name := string(pt.Name())
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
