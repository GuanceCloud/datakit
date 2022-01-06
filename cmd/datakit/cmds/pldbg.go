package cmds

import (
	"encoding/json"
	"fmt"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline"
)

func pipelineDebugger(plname, txt string) error {
	if txt == "" {
		l.Fatal("-txt required")
	}

	if err := pipeline.Init(datakit.DataDir); err != nil {
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

	if res == nil || (len(res.Data) == 0 && len(res.Tags) == 0) {
		fmt.Println("No data extracted from pipeline")
		return nil
	}

	result := map[string]interface{}{}

	for k, v := range res.Data {
		result[k] = v
	}
	for k, v := range res.Tags {
		result[k+"#"] = v
	}
	j, err := json.MarshalIndent(result, "", defaultJSONIndent)
	if err != nil {
		return err
	}

	fmt.Printf("Extracted data(drop: %v, cost: %v):\n", res.Dropped, time.Since(start))
	fmt.Printf("%s\n", string(j))
	return nil
}
