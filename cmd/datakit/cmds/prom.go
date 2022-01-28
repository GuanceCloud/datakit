package cmds

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/influxdata/influxdb1-client/models"
	tomlAst "github.com/influxdata/toml/ast"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
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
	table, err := config.ParseCfgFile(file)
	if err != nil {
		return "", err
	}
	it := table.Fields["inputs"]
	tbl, ok := it.(*tomlAst.Table)
	if !ok {
		return "", fmt.Errorf("expect to be *tomlAst.Table")
	}

	for k := range tbl.Fields {
		return k, nil
	}
	return "", fmt.Errorf("collector name not found in config file")
}

// getPromRemoteWriteInput constructs a prom_remote_write.Input by given config file.
func getPromRemoteWriteInput(configPath string) (*pr.Input, error) {
	inputList, err := config.LoadInputConfigFile(configPath, func() inputs.Input {
		return pr.NewInput()
	})
	if err != nil {
		return nil, err
	}
	if len(inputList) != 1 {
		return nil, fmt.Errorf("should test only one prom_remote_write config, now get %v", len(inputList))
	}

	input, ok := inputList[0].(*pr.Input)

	if !ok {
		return nil, fmt.Errorf("invalid prom_remote_write instance")
	}

	return input, nil
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
	var points []*io.Point
	for _, m := range measurements {
		mm, ok := m.(*pr.Measurement)
		if !ok {
			return fmt.Errorf("expect to be *prom_remote_write.Measurement")
		}

		input.AddAndIgnoreTags(mm)
		p, err := mm.LineProto()
		if err != nil {
			return err
		}
		points = append(points, p)
	}
	return printResult(points)
}

func getPromInput(configPath string) (*prom.Input, error) {
	inputList, err := config.LoadInputConfigFile(configPath, func() inputs.Input {
		return prom.NewProm()
	})
	if err != nil {
		return nil, err
	}
	if len(inputList) != 1 {
		return nil, fmt.Errorf("should test only one prom config, now get %v", len(inputList))
	}

	input, ok := inputList[0].(*prom.Input)

	if !ok {
		return nil, fmt.Errorf("invalid prom instance")
	}

	return input, nil
}

func showPromInput(input *prom.Input) error {
	// init client
	err := input.Init()
	if err != nil {
		return err
	}

	var points []*io.Point
	if input.Output != "" {
		points, err = input.CollectFromFile(input.Output)
	} else {
		points, err = input.Collect()
	}
	if err != nil {
		return err
	}

	return printResult(points)
}

func printResult(points []*io.Point) error {
	fmt.Printf("\n================= Line Protocol Points ==================\n\n")
	// measurements collected
	measurements := make(map[string]string)
	timeSeries := make(map[string]string)
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
