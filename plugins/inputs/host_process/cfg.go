package host_process

import "gitlab.jiagouyun.com/cloudcare-tools/datakit"

const (
	inputName = "host_processes"
	category  = "host"

	sampleConfig = `
[[inputs.host_processes]]
 ## process name support regexp
 # process_name = [".*datakit.*"]
 ## write object interval
 object_interval     = "5m"  
 ## process run time default 10m,Collection  the process of running more than ten minutes
 run_time     = "10m"  
 ## open collection metric	
 open_metric = false
 ## write metric interval 
 # metric_interval = "10s"
 ## pipeline path 
 # pipeline = ""
`

	pipelineSample = ``
)

type Processes struct {
	ProcessName    []string         `toml:"process_name,omitempty"`
	ObjectInterval datakit.Duration `toml:"object_interval,omitempty"`
	RunTime        datakit.Duration `toml:"run_time,omitempty"`
	OpenMetric     bool             `toml:"open_metric,omitempty"`
	MetricInterval datakit.Duration `toml:"metric_interval,omitempty"`
	Pipeline       string           `toml:"pipeline,omitempty"`

	re string
}
