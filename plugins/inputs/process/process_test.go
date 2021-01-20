package process

import (
	"encoding/json"
	"fmt"
	pr "github.com/shirou/gopsutil/v3/process"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"testing"
	"time"
)

func TestParseField(t *testing.T) {
	p, _ := pr.NewProcess(1)
	field, _ := parseField(p)
	v, _ := json.Marshal(field)
	fmt.Println(string(v))
}

func TestProcessesRun(t *testing.T) {
	var obj = Processes{
		ProcessName: []string{"zsh"},
		Interval:    datakit.Duration{Duration: 5 * time.Minute},
		RunTime:     datakit.Duration{Duration: 10 * time.Minute},
		OpenMetric:  false,
		re:          ".*Google",
	}
	obj.run()

}
