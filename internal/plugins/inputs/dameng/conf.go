// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package dameng

const sampleCfg = `
[[inputs.dameng]]
  # host name
  host = "localhost"

  ## port
  port = 5236

  ## user name
  user = "SYSDBA"

  ## password
  password = "datakit"

  ## database name
  database = "DMTEST"

  ## Slow query threshold in milliseconds, default 1000
  slow_query_threshold = 1000

  ## @param connect_timeout - number - optional - default: 10s
  # connect_timeout = "10s"

  interval = "10s"

  ## Set true to enable election
  election = true

  ## Metric name in metric_exclude_list will not be collected.
  #
  metric_exclude_list = [""]

  ## Run a custom SQL query and collect corresponding metrics.
  #
  # [[inputs.dameng.custom_queries]]
  #   sql = 'SELECT name AS "name", stat_val AS "stat_val" FROM sys.v$sysstat;'
  #   metric = "dameng_custom_query"
  #   tags = ["name"]
  #   fields = ["stat_val"]
  #   interval = "30s"

  ## Log collection
  #
  [inputs.dameng.log]
    # files = []
    # pipeline = "dameng.p"
    ## The pattern should be a regexp. Note the use of '''this regexp'''
    ## regexp link: https://golang.org/pkg/regexp/syntax/#hdr-Syntax
    multiline_match = '''^\\d{4}-\\d{2}-\\d{2}\\s+\\d{2}:\\d{2}:\\d{2}\\s+\\[.*?\\]'''

  [inputs.dameng.tags]
    # some_tag = "some_value"
    # more_tag = "some_other_value"
`

const pipelineCfg = `
add_pattern("log_date", "%{YEAR}-%{MONTHNUM}-%{MONTHDAY}%{SPACE}%{HOUR}:%{MINUTE}:%{SECOND}")
add_pattern("status", "(ERROR|FATAL|WARNING|INFO)")
# default
grok(_, "%{log_date:time}%{SPACE}\\[%{status:status}\\]\\s+%{GREEDYDATA:msg}")
default_time(time)
`
