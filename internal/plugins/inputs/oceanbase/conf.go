// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package oceanbase

const (
	configSample = `
[[inputs.oceanbase]]
  # host name
  host = "localhost"

  ## port
  port = 2883

  ## tenant name
  tenant = "sys"

  ## cluster name
  cluster = "obcluster"

  ## user name
  user = "datakit"

  ## password
  password = "<PASS>"

  ## database name
  database = "oceanbase"

  ## mode. mysql only.
  mode = "mysql"

  ## @param connect_timeout - number - optional - default: 10s
  # connect_timeout = "10s"

  interval = "10s"

  ## OceanBase slow query time threshold defined. If larger than this, the executed sql will be reported.
  slow_query_time = "0s"

  ## Set true to enable election
  election = true

  ## Run a custom SQL query and collect corresponding metrics.
  # [[inputs.oceanbase.custom_queries]]
    # sql = '''
    #   select
    #     CON_ID tenant_id,
    #     STAT_ID,
    #     replace(name, " ", "_") metric_name,
    #     VALUE
    #   from
    #     v$sysstat;
    # '''
    # metric = "oceanbase_custom"
    # tags = ["metric_name", "tenant_id"]
    # fields = ["VALUE"]

  [inputs.oceanbase.tags]
    # some_tag = "some_value"
    # more_tag = "some_other_value"
`
)
