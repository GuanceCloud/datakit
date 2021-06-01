package cmds

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline"
)

var (
	l = logger.DefaultSLogger("cmds")
)

func PipelineDebugger(plname, txt string) {
	if txt == "" {
		l.Fatal("-txt required")
	}

	if err := pipeline.Init(datakit.DataDir); err != nil {
		l.Fatalf("pipeline init failed: %s", err.Error())
	}

	start := time.Now()
	pl, err := pipeline.NewPipelineFromFile(filepath.Join(datakit.PipelineDir, plname))
	if err != nil {
		l.Fatalf("new pipeline failed: %s", err.Error())
	}

	res, err := pl.Run(txt).Result()
	if err != nil {
		l.Fatalf("run pipeline failed: %s", err.Error())
	}

	if len(res) == 0 {
		fmt.Println("No data extracted from pipeline")
		return
	}

	if j, err := json.MarshalIndent(res, "", "    "); err != nil {
		l.Fatal(err)
	} else {
		fmt.Printf("Extracted data(cost: %v):\n", time.Since(start))
		fmt.Printf("%s\n", string(j))
	}
}
