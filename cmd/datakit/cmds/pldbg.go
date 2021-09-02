package cmds

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
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
	pl, err := pipeline.NewPipelineFromFile(filepath.Join(datakit.PipelineDir, plname))
	if err != nil {
		return fmt.Errorf("new pipeline failed: %s", err.Error())
	}

	res, err := pl.Run(txt).Result()
	if err != nil {
		return fmt.Errorf("run pipeline failed: %s", err.Error())
	}

	if len(res) == 0 {
		fmt.Println("No data extracted from pipeline")
		return nil
	}

	j, err := json.MarshalIndent(res, "", "    ")
	if err != nil {
		return err
	}

	fmt.Printf("Extracted data(cost: %v):\n", time.Since(start))
	fmt.Printf("%s\n", string(j))
	return nil
}
