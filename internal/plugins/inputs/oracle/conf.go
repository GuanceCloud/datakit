// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package oracle

const (
	configSample = `
[[inputs.oracle]]
  # host name
  host = "localhost"

  ## port
  port = 1521

  ## user name
  user = "datakit"

  ## password
  password = "<PASS>"

  ## service
  service = "XE"

  ## interval
  interval = "10s"

  ## connection timeout
  # connect_timeout = "10s"

  ## slow query time threshold defined. If larger than this, the executed sql will be reported.
  slow_query_time = "0s"

  ## Set true to enable election
  election = true

  ## Run a custom SQL query and collect corresponding metrics.
  # [[inputs.oracle.custom_queries]]
  #   sql = '''
  #     SELECT
  #       GROUP_ID, METRIC_NAME, VALUE
  #     FROM GV$SYSMETRIC
  #   '''
  #   metric = "oracle_custom"
  #   tags = ["GROUP_ID", "METRIC_NAME"]
  #   fields = ["VALUE"]

  [inputs.oracle.tags]
    # some_tag = "some_value"
    # more_tag = "some_other_value"
`
)
