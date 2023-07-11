// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package consul collect consul metrics by using input prom
//
//nolint:lll
package consul

import (
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/tailer"
)

const (
	inputName    = "consul"
	configSample = `
[[inputs.prom]]
  url = "http://127.0.0.1:9107/metrics"
  source = "consul"
  metric_types = ["counter", "gauge"]
  metric_name_filter = ["consul_raft_leader", "consul_raft_peers", "consul_serf_lan_members", "consul_catalog_service", "consul_catalog_service_node_healthy", "consul_health_node_status", "consul_serf_lan_member_status"]
  measurement_prefix = ""
  tags_ignore = ["check"]
  interval = "10s"

[[inputs.prom.measurements]]
  prefix = "consul_"
  name = "consul"
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

type Input struct { // keep compatible with old version's conf
	Log *inputs.XLog `toml:"log"`

	TokenDeprecated      string `toml:"token,omitempty"`
	AddressDeprecated    string `toml:"address,omitempty"`
	SchemeDeprecated     string `toml:"scheme,omitempty"`
	UsernameDeprecated   string `toml:"username,omitempty"`
	PasswordDeprecated   string `toml:"password,omitempty"`
	DatacenterDeprecated string `toml:"datacenter,omitempty"`
}

var _ inputs.InputV2 = (*Input)(nil)

func (*Input) Terminate() {
	// do nothing
}

func (ipt *Input) LogExamples() map[string]map[string]string {
	return map[string]map[string]string{
		inputName: {
			"Consul log": `Sep 18 19:30:23 derrick-ThinkPad-X230 consul[11803]: 2021-09-18T19:30:23.522+0800 [INFO]  agent.server.connect: initialized primary datacenter CA with provider: provider=consul`,
		},
	}
}

func (*Input) Catalog() string {
	return inputName
}

func (*Input) SampleConfig() string {
	return configSample
}

func (*Input) Run() {
}

func (*Input) AvailableArchs() []string {
	return datakit.AllOSWithElection
}

func (*Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{
		&ConsulMeasurement{},
	}
}

func (*Input) PipelineConfig() map[string]string {
	pipelineMap := map[string]string{
		inputName: pipelineCfg,
	}
	return pipelineMap
}

func (ipt *Input) GetPipeline() []*tailer.Option {
	return []*tailer.Option{
		{
			Source:  inputName,
			Service: inputName,
			Pipeline: func() string {
				if ipt.Log != nil {
					return ipt.Log.Pipeline
				}
				return ""
			}(),
		},
	}
}

func (*Input) RunPipeline() {
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return &Input{}
	})
}
