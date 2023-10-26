// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package consul

//nolint:lll
const (
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
