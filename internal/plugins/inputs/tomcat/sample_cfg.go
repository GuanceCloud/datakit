// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package tomcat

const (
	//nolint:lll
	tomcatSampleCfg = `
[[inputs.statsd]]
  protocol = "udp"

  ## Address and port to host UDP listener on
  service_address = ":8125"

  ## Tag request metric. Used for distinguish feed metric name.
  ## eg, DD_TAGS=source_key:tomcat,host_key:cn-shanghai-sq5ei
  ## eg, -Ddd.tags=source_key:tomcat,host_key:cn-shanghai-sq5ei
  # statsd_source_key = "source_key"
  # statsd_host_key   = "host_key"
  ## Indicate whether report tag statsd_source_key and statsd_host_key.
  # save_above_key    = false

  delete_gauges = true
  delete_counters = true
  delete_sets = true
  delete_timings = true

  ## Percentiles to calculate for timing & histogram stats
  percentiles = [50.0, 90.0, 99.0, 99.9, 99.95, 100.0]

  ## separator to use between elements of a statsd metric
  metric_separator = "_"

  ## Parses tags in the datadog statsd format
  ## http://docs.datadoghq.com/guides/dogstatsd/
  parse_data_dog_tags = true

  ## Parses datadog extensions to the statsd format
  datadog_extensions = true

  ## Parses distributions metric as specified in the datadog statsd format
  ## https://docs.datadoghq.com/developers/metrics/types/?tab=distribution#definition
  datadog_distributions = true

  ## We do not need following tags(they may create tremendous of time-series under influxdb's logic)
  # Examples:
  # "runtime-id", "metric-type"
  drop_tags = [ ]

  # All metric-name prefixed with 'jvm_' are set to influxdb's measurement 'jvm'
  # All metric-name prefixed with 'stats_' are set to influxdb's measurement 'stats'
  # Examples:
  # "stats_:stats", "jvm_:jvm"
  metric_mapping = [ ]

  ## Number of UDP messages allowed to queue up, once filled,
  ## the statsd server will start dropping packets
  allowed_pending_messages = 10000

  ## Number of timing/histogram values to track per-measurement in the
  ## calculation of percentiles. Raising this limit increases the accuracy
  ## of percentiles but also increases the memory usage and cpu time.
  percentile_limit = 1000

  ## Max duration (TTL) for each metric to stay cached/reported without being updated.
  #max_ttl = "1000h"

  [inputs.statsd.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"`

	//nolint:lll
	pipelineCfg = `
# juli OneLineFormatter format
# catiline / host-manager / localhost / manager log
add_pattern("olf_time", "%{MONTHDAY}-%{MONTH}-%{YEAR} %{TIME}")
grok(_, "%{olf_time:time} %{LOGLEVEL:status} \\[%{NOTSPACE:thread_name}\\] %{NOTSPACE:report_source} %{GREEDYDATA:msg}")

# localhost_access_log log
grok(_, "%{NOTSPACE:client_ip} %{NOTSPACE:http_ident} %{NOTSPACE:http_auth} \\[%{HTTPDATE:time}\\] \"%{DATA:http_method} %{GREEDYDATA:http_url} HTTP/%{NUMBER:http_version}\" %{INT:status_code} %{INT:bytes}")

cast(status_code, "int")
cast(bytes, "int")
group_between(status_code, [200,299], "OK", status)
group_between(status_code, [300,399], "notice", status)
group_between(status_code, [400,499], "warning", status)
group_between(status_code, [500,599], "error", status)

nullif(http_ident, "-")
nullif(http_auth, "-")

default_time(time)
`
)
