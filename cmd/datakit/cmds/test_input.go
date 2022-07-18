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

	"github.com/influxdata/influxdb1-client/models"
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
			warnf("[W] stat failed: %v\n", err)
			return err
		}
	}

	infof("[I] config path: %s\n", configPath)

	inputsInstance, err := config.LoadSingleConfFile(configPath, inputs.Inputs)
	if err != nil {
		errorf("[E] parse failed: %v\n", err)
		return err
	}

	for k, arr := range inputsInstance {
		for _, x := range arr {
			if i, ok := x.(inputs.InputOnceRunnableCollect); !ok {
				warnf("[W] %s not implement for now.\n", k)
				continue
			} else {
				mpts, e := i.Collect()
				if e != nil {
					err = e
					warnf("[W] %s Collect failed: %s\n", k, e.Error())
					return err
				}
				if err = printResultEx(mpts); err != nil {
					warnf("[W] %s print failed: %s\n", k, e.Error())
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
