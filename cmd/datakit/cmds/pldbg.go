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

	start := time.Now()
	plPath, err := config.GetPipelinePath(plname)
	if err != nil {
		return fmt.Errorf("get pipeline failed: %w", err)
	}
	pl, err := pipeline.NewPipelineFromFile(plPath)
	if err != nil {
		return fmt.Errorf("new pipeline failed: %w", err)
	}

	res, err := pl.Run(txt).Result()
	if err != nil {
		return fmt.Errorf("run pipeline failed: %w", err)
	}

	if len(res) == 0 {
		fmt.Println("No data extracted from pipeline")
		return nil
	}

	j, err := json.MarshalIndent(res, "", defaultJSONIndent)
	if err != nil {
		return err
	}

	fmt.Printf("Extracted data(cost: %v):\n", time.Since(start))
	fmt.Printf("%s\n", string(j))
	return nil
}
