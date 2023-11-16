// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package statsd

const sampleConfig = `
[[inputs.statsd]]
  ## Collector alias.
  # source = "statsd/-/-"

  ## Collect interval, default is 10 seconds. (optional)
  # interval = '10s'

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
  # "stats_:stats", "jvm_:jvm", "tomcat_:tomcat",
  metric_mapping = [ ]

  ## Number of UDP messages allowed to queue up, once filled,
  ## the statsd server will start dropping packets, default is 128.
  # allowed_pending_messages = 128

  ## Number of timing/histogram values to track per-measurement in the
  ## calculation of percentiles. Raising this limit increases the accuracy
  ## of percentiles but also increases the memory usage and cpu time.
  percentile_limit = 1000

  ## Max duration (TTL) for each metric to stay cached/reported without being updated.
  #max_ttl = "1000h"

  [inputs.statsd.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"`
