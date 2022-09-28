// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package cmds

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/lineproto"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/prom"
	pr "gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/promremote"
)

const trueString = "true"

func promDebugger(configFile string) error {
	configPath := configFile
	if !path.IsAbs(configFile) {
		currentDir, _ := os.Getwd()
		configPath = filepath.Join(currentDir, configFile)
		_, err := os.Stat(configPath)
		if err != nil {
			configPath = filepath.Join(datakit.ConfdDir, "prom", filepath.Base(configFile))
			fmt.Printf("config is not found in current dir, using %s instead\n", configPath)
		}
	}
	name, err := collectorName(configPath)
	if err != nil {
		return err
	}
	switch name {
	case "prom":
		input, err := getPromInput(configPath)
		if err != nil {
			return err
		}
		err = showPromInput(input)
		if err != nil {
			return err
		}

	case "prom_remote_write":
		input, err := getPromRemoteWriteInput(configPath)
		if err != nil {
			return err
		}
		err = showPromRemoteWriteInput(input)
		if err != nil {
			return err
		}
	}
	return nil
}

// collectorName parses given config file and returns collector name.
func collectorName(file string) (string, error) {
	x, err := config.LoadSingleConfFile(file, inputs.Inputs)
	if err != nil {
		return "", err
	}

	for k := range x {
		return k, nil
	}

	return "", fmt.Errorf("collector name not found in config file")
}

// getPromRemoteWriteInput constructs a prom_remote_write.Input by given config file.
func getPromRemoteWriteInput(configPath string) (*pr.Input, error) {
	inputList, err := config.LoadSingleConfFile(configPath, inputs.Inputs)
	if err != nil {
		return nil, err
	}

	if len(inputList) != 1 {
		return nil, fmt.Errorf("should test only one prom_remote_write config, now get %v", len(inputList))
	}

	for _, arr := range inputList {
		if len(arr) != 1 {
			return nil, fmt.Errorf("should test only one prom_remote_write config, now get %v", len(inputList))
		}

		input, ok := arr[0].(*pr.Input)
		if !ok {
			return nil, fmt.Errorf("invalid prom_remote_write instance")
		}

		return input, nil
	}

	return nil, fmt.Errorf("invalid prom_remote_write instance")
}

// showPromRemoteWriteInput reads raw data file specified by prom_remote_write.Input.Output,
// performs metric filtering and prefixing, and adds/ignores tags based on configuration.
// parsed metrics are at last passed to printResult.
func showPromRemoteWriteInput(input *pr.Input) error {
	fp := input.Output
	if !path.IsAbs(fp) {
		dir := datakit.InstallDir
		fp = filepath.Join(dir, fp)
	}
	file, err := os.Open(filepath.Clean(fp))
	if err != nil {
		return err
	}
	data, err := ioutil.ReadAll(file)
	if err != nil {
		return err
	}
	measurements, err := input.Parse(data)
	if err != nil {
		return err
	}
	var points []*point.Point
	for _, m := range measurements {
		mm, ok := m.(*pr.Measurement)
		if !ok {
			return fmt.Errorf("expect to be *prom_remote_write.Measurement")
		}

		input.SetTags(mm)
		p, err := mm.LineProto()
		if err != nil {
			return err
		}
		points = append(points, p)
	}
	return printResult(points)
}

func getPromInput(configPath string) (*prom.Input, error) {
	res, err := config.LoadSingleConfFile(configPath, inputs.Inputs)
	if err != nil {
		return nil, err
	}

	if len(res) != 1 {
		return nil, fmt.Errorf("should test only one prom config, now get %v", len(res))
	}

	for _, arr := range res {
		if len(arr) != 1 {
			return nil, fmt.Errorf("should test only one prom config, now get %v", len(arr))
		}

		input, ok := arr[0].(*prom.Input)
		if !ok {
			return nil, fmt.Errorf("invalid prom instance")
		}

		// use the first 1
		return input, nil
	}

	return nil, fmt.Errorf("invalid prom instance")
}

func showPromInput(input *prom.Input) error {
	err := input.Init()
	if err != nil {
		return err
	}

	var points []*point.Point
	if input.Output != "" {
		// If input.Output is configured, raw metric text is written to file.
		// In this case, read the file and perform Text2Metric.
		points, err = input.CollectFromFile(input.Output)
	} else {
		// Collect from all URLs.
		points, err = input.Collect()
	}
	if err != nil {
		return err
	}

	return printResult(points)
}

func printResult(points []*point.Point) error {
	fmt.Printf("\n================= Line Protocol Points ==================\n\n")
	// measurements collected
	measurements := make(map[string]string)
	timeSeries := make(map[string]string)

	encoder := lineproto.NewLineEncoder(lineproto.WithPrecisionV2(lineproto.Nanosecond))
	for _, pt := range points {
		if err := encoder.AppendPoint(pt.Point); err != nil {
			return fmt.Errorf("apend point err: %w", err)
		}

		influxPoint, err := pt.ToInfluxdbPoint(time.Now())
		if err != nil {
			return err
		}
		timeSeries[fmt.Sprint(influxPoint.HashID())] = trueString
		name := pt.Name
		measurements[name] = trueString
	}

	lines, _ := encoder.UnsafeString()
	fmt.Println(lines)

	mKeys := make([]string, len(measurements))
	i := 0
	for name := range measurements {
		mKeys[i] = name
		i++
	}
	fmt.Printf("\n================= Summary ==================\n\n")
	fmt.Printf("Total time series: %v\n", len(timeSeries))
	fmt.Printf("Total line protocol points: %v\n", len(points))
	fmt.Printf("Total measurements: %v (%s)\n\n", len(measurements), strings.Join(mKeys, ", "))

	return nil
}
