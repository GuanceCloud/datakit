package host_process

import (
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
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

type Input struct {
	ProcessName    []string         `toml:"process_name,omitempty"`
	ObjectInterval datakit.Duration `toml:"object_interval,omitempty"`
	RunTime        datakit.Duration `toml:"min_run_time,omitempty"`
	OpenMetric     bool             `toml:"open_metric,omitempty"`
	MetricInterval datakit.Duration `toml:"metric_interval,omitempty"`
	Pipeline       string           `toml:"pipeline,omitempty"`

	re     string
	isTest bool
}

type ProcessMetric struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
	ts     time.Time
}

func (m *ProcessMetric) LineProto() (*io.Point, error) {
	return io.MakePoint(m.name, m.tags, m.fields, m.ts)
}

func (m *ProcessMetric) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   "process",
		Fields: map[string]interface{}{},
		Tags:   map[string]interface{}{},
	}
}

type ProcessObject struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
	ts     time.Time
}

func (m *ProcessObject) LineProto() (*io.Point, error) {
	return io.MakePoint(m.name, m.tags, m.fields, m.ts)
}

func (m *ProcessObject) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   "process",
		Fields: map[string]interface{}{},
		Tags:   map[string]interface{}{},
	}
}
