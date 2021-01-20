package process

import "gitlab.jiagouyun.com/cloudcare-tools/datakit"

const (
	inputName = "processes"

	sampleConfig = `
[[inputs.processes]]
 ## process name support regexp
 process_name = [".*datakit.*"]
 ## period 5m
 interval     = "5m"  
 ## process run time default 10m,Collection  the process of running more than ten minutes
 run_time     = "10m"  
 ## open collection metric	
 open_metric = false
 ## pipeline path 
 # pipeline = ""
`

	pipelineSample = ``
)

type Processes struct {
	ProcessName []string         `toml:"process_name"`
	Interval    datakit.Duration `toml:"interval"`
	RunTime     datakit.Duration `toml:"run_time"`
	OpenMetric  bool             `toml:"open_metric"`
	Pipeline    string           `toml:"pipeline"`

	re       string
	username string
	state    string
}
