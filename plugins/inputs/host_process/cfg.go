package host_process

import (
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
)

const (
	inputName = "host_processes"
	category  = "host"

	sampleConfig = `
[[inputs.host_processes]]
 ## process name support regexp
 # process_name = [".*datakit.*"]
 ## process min run time default 10m,Collection  the process of running more than ten minutes
 min_run_time     = "10m"
 ## open collection metric
 open_metric = false
 ## pipeline path
 # pipeline = ""
`

	pipelineSample = ``
)

type Processes struct {
	ProcessName    []string         `toml:"process_name,omitempty"`
	ObjectInterval datakit.Duration `toml:"object_interval,omitempty"`
	RunTime        datakit.Duration `toml:"min_run_time,omitempty"`
	OpenMetric     bool             `toml:"open_metric,omitempty"`
	MetricInterval datakit.Duration `toml:"metric_interval,omitempty"`
	Pipeline       string           `toml:"pipeline,omitempty"`

	re     string
	isTest bool
}
