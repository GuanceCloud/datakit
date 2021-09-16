package consul

import (
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

const (
	inputName = "consul"
	configSample = `
[[inputs.mysql]]
	pipeline = "consul.p"
`
	pipelineCfg = `
add_pattern("_clog_date", "%{YEAR}-%{MONTHNUM}-%{MONTHDAY}T%{HOUR}:%{MINUTE}:%{SECOND}%{INT}")
add_pattern("_clog_level", "(DEBUG|INFO|WARN|ERROR|FATAL)")
add_pattern("_clog_character", "%{NOTSPACE}")
add_pattern("_clog_message", "%{GREEDYDATA}")
grok(_, '%{SYSLOGTIMESTAMP}%{SPACE}%{SYSLOGHOST}%{SPACE}consul\\[%{POSINT}\\]:%{SPACE}%{_clog_date:date}%{SPACE}\\[%{_clog_level:level}\\]%{SPACE}%{_clog_character:character}:%{SPACE}%{_clog_message:msg}')
drop_origin_data()
`
)

type Input struct {

}

func (i *Input) Catalog() string {
	return "consul"
}

func (i *Input) SampleConfig() string {
	return configSample
}

func (i *Input) Run() {

}

func (i *Input) AvailableArchs() []string {
	return datakit.AllArch
}

func (i *Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{}
}

func (i *Input) PipelineConfig() map[string]string {
	pipelineMap := map[string]string{
		"consul": pipelineCfg,
	}
	return pipelineMap
}

func (i *Input) RunPipeline() {

}

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return &Input{}
	})
}

