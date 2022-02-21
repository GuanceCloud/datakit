package cmds

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"reflect"
	"sort"
	"strings"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline"
)

func runPLFlags() error {
	var txt string
	if *flagPLTxtFile != "" {
		txtBytes, err := ioutil.ReadFile(*flagPLTxtFile)
		if err != nil {
			return fmt.Errorf("ioutil.ReadFile: %w", err)
		}
		txt = string(txtBytes)
		txt = strings.TrimSuffix(txt, "\n")
	}

	if txt == "" {
		if *flagPLTxtData != "" {
			txt = *flagPLTxtData
		}
	}

	if txt == "" {
		return fmt.Errorf("empty txt")
	}

	if strings.HasSuffix(txt, "\n") {
		warnf("[E] txt has suffix EOL\n")
	}

	return pipelineDebugger(debugPipelineName, txt)
}

func pipelineDebugger(plname, txt string) error {
	if err := pipeline.Init(config.Cfg.Pipeline); err != nil {
		return err
	}

	plPath, err := config.GetPipelinePath(plname)
	if err != nil {
		return fmt.Errorf("get pipeline failed: %w", err)
	}
	pl, err := pipeline.NewPipelineFromFile(plPath, true)
	if err != nil {
		return fmt.Errorf("new pipeline failed: %w", err)
	}

	start := time.Now()
	res, err := pl.Run(txt).Result()
	if err != nil {
		return fmt.Errorf("run pipeline failed: %w", err)
	}
	cost := time.Since(start)

	if res == nil || (len(res.Data) == 0 && len(res.Tags) == 0) {
		errorf("[E] No data extracted from pipeline\n")
		return nil
	}

	result := map[string]interface{}{}
	maxWidth := 0

	for k, v := range res.Data {
		if len(k) > maxWidth {
			maxWidth = len(k)
		}

		switch k {
		case "time":
			switch x := v.(type) {
			case int64:
				if *flagPLDate {
					date := time.Unix(0, x)
					result[k] = fmt.Sprintf("%d(%s)", x, date.String())
				} else {
					result[k] = v
				}
			default:
				warnf("`time' should be int64, but got %s\n", reflect.TypeOf(v).String())
			}
		default:
			result[k] = v
		}
	}

	for k, v := range res.Tags {
		result[k+"#"] = v
		if len(k)+1 > maxWidth {
			maxWidth = len(k) + 1
		}
	}

	if *flagPLTable {
		fmtStr := fmt.Sprintf("%% %ds: %%v", maxWidth)
		lines := []string{}
		for k, v := range result {
			lines = append(lines, fmt.Sprintf(fmtStr, k, v))
		}

		sort.Strings(lines)
		for _, l := range lines {
			fmt.Println(l)
		}
	} else {
		j, err := json.MarshalIndent(result, "", defaultJSONIndent)
		if err != nil {
			return err
		}
		fmt.Printf("%s\n", string(j))
	}

	infof("---------------\n")
	infof("Extracted %d fields, %d tags; drop: %v, cost: %v\n", len(res.Data), len(res.Tags), res.Dropped, cost)

	return nil
}
