// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package cmds

import (
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"github.com/influxdata/influxdb1-client/models"
	cp "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/colorprint"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/prom"
	pr "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/promremote"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/promremote/prompb"
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
			cp.Printf("config is not found in current dir, using %s instead\n", configPath)
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
	x, err := config.LoadSingleConfFile(file, inputs.AllInputs, true)
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
	inputList, err := config.LoadSingleConfFile(configPath, inputs.AllInputs, true)
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

		input, ok := arr[0].Input.(*pr.Input)
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
	data, err := io.ReadAll(file)
	if err != nil {
		return err
	}

	var req prompb.WriteRequest
	if err := req.Unmarshal(data); err != nil {
		return fmt.Errorf("unable to unmarshal request body: %w", err)
	}

	measurements, err := input.Parse(req.Timeseries, input, nil)
	if err != nil {
		return err
	}

	return printResult(measurements)
}

func getPromInput(configPath string) (*prom.Input, error) {
	res, err := config.LoadSingleConfFile(configPath, inputs.AllInputs, true)
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

		input, ok := arr[0].Input.(*prom.Input)
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

	var clipts []*point.Point
	if input.Output != "" {
		// If input.Output is configured, raw metric text is written to file.
		// In this case, read the file and perform Text2Metric.
		clipts, err = input.CollectFromFile(input.Output)
	} else {
		// Collect from all URLs.
		if len(input.URLs) > 0 {
			clipts, err = input.CollectFromHTTP(input.URLs[0])
		} else {
			err = fmt.Errorf("error urls")
		}
	}
	if err != nil {
		return err
	}

	return printResult(clipts)
}

func printResult(points []*point.Point) error {
	cp.Printf("\n================= Line Protocol Points ==================\n\n")
	// measurements collected
	measurements := make(map[string]string)
	timeSeries := make(map[string]string)
	for _, pt := range points {
		lp := pt.LineProto()
		cp.Printf(" %s\n", lp)

		influxPoint, err := models.ParsePointsWithPrecision([]byte(lp), time.Now(), "n")
		if len(influxPoint) != 1 {
			return fmt.Errorf("parse point error")
		}
		if err != nil {
			return err
		}
		for _, v := range pt.TimeSeriesHash() {
			timeSeries[v] = trueString
		}
		name := pt.Name()
		measurements[name] = trueString
	}
	mKeys := make([]string, len(measurements))
	i := 0
	for name := range measurements {
		mKeys[i] = name
		i++
	}
	cp.Printf("\n================= Summary ==================\n\n")
	cp.Printf("Total time series: %v\n", len(timeSeries))
	cp.Printf("Total line protocol points: %v\n", len(points))
	cp.Printf("Total measurements: %v (%s)\n\n", len(measurements), strings.Join(mKeys, ", "))

	return nil
}
