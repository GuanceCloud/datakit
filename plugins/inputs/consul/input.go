// Package consul collect consul metrics by using input prom
//nolint:lll
package consul

import (
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/tailer"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

const (
	inputName    = "consul"
	configSample = `
[[inputs.prom]]
  ## Exporter 地址
  url = "http://127.0.0.1:9107/metrics"

  ## 采集器别名
  source = "consul"

  ## 指标类型过滤, 可选值为 counter, gauge, histogram, summary
  # 默认只采集 counter 和 gauge 类型的指标
  # 如果为空，则不进行过滤
  metric_types = ["counter", "gauge"]

  ## 指标名称过滤
  # 支持正则，可以配置多个，即满足其中之一即可
  # 如果为空，则不进行过滤
    metric_name_filter = ["consul_raft_leader", "consul_raft_peers", "consul_serf_lan_members", "consul_catalog_service", "consul_catalog_service_node_healthy", "consul_health_node_status", "consul_serf_lan_member_status"]


  ## 指标集名称前缀
  # 配置此项，可以给指标集名称添加前缀
  measurement_prefix = ""

  ## 过滤tags, 可配置多个tag
  # 匹配的tag将被忽略
  tags_ignore = ["check"]

  ## 采集间隔 "ns", "us" (or "µs"), "ms", "s", "m", "h"
  interval = "10s"

  [[inputs.prom.log]]
     pipeline = "consul.p"
     files = ["", ""]
     source = "consul"
     service = "consul"

  ## 自定义指标集名称
  # 可以将包含前缀prefix的指标归为一类指标集
  # 自定义指标集名称配置优先measurement_name配置项
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

func (*Input) Catalog() string {
	return inputName
}

func (*Input) SampleConfig() string {
	return configSample
}

func (*Input) Run() {
}

func (*Input) AvailableArchs() []string {
	return datakit.AllArch
}

func (*Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{
		&HostMeasurement{},
		&ServiceMeasurement{},
		&HealthMeasurement{},
		&MemberMeasurement{},
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
			Source:   inputName,
			Service:  inputName,
			Pipeline: ipt.Log.Pipeline,
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
