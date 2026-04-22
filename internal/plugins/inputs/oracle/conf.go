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

  ## Interval (waiting event, locked session metrics).
  interval = "10s"

  ## connection timeout
  connect_timeout = "30s"

  ## slow query time threshold defined. If larger than this, the executed sql will be reported.
  slow_query_time = "0s"

  ## Metric name in metric_exclude_list will not be collected.
  metric_exclude_list = [""]

  ## Set true to enable election
  election = true

  ## v2+ override all metric measurements to "oracle", default: v2
  ## If you want to use the old metric set, you can change it to "v1"
  measurement_version = "v2"

  ## collect object
  [inputs.oracle.object]
    # Set true to enable collecting objects
    enabled = true

    # interval to collect oracle object which will be greater than collection interval
    interval = "600s"

  ## tablespace collection
  [inputs.oracle.tablespace]
    # Set true to enable collecting tablespace metrics (default: true)
    enabled = true
    # Collection interval for tablespace metrics (default: 600s)
    interval = "600s"

  ## slow query collection
  [inputs.oracle.slow_query]
    # Set true to enable collecting slow query metrics (default: true)
    enabled = true
    # Collection interval for slow query metrics (default: 60s)
    interval = "60s"

  ## process collection
  [inputs.oracle.process]
    # Set true to enable collecting process metrics (default: true)
    enabled = true
    # Collection interval for process metrics (default: 60s)
    interval = "60s"

  ## system metrics collection
  [inputs.oracle.system]
    # Set true to enable collecting system metrics (default: true)
    enabled = true
    # Collection interval for system metrics (default: 60s)
    interval = "60s"

  ## Database Monitoring (DBM) configuration
  ## DBM provides deep visibility into database performance by collecting query metrics, activity, and execution plans
  [inputs.oracle.dbm]
    # Set true to enable DBM metrics collection
    enabled = false

  ## Config DBM metric (query metrics)
  ## Collects cumulative execution statistics of SQL queries aggregated by query signature, plan hash, and PDB
  [inputs.oracle.dbm.metric]
    # Set true to enable collecting query metrics
    enabled = true
    # Collection interval for query metrics (default: 60s)
    collection_interval = "60s"
    # Maximum number of rows to collect from V$SQLSTATS (default: 10000)
    # This limits the initial query result size before aggregation
    db_rows_limit = 10000
    # Maximum number of queries to report per collection interval (default: 500)
    # Only the top N queries (sorted by derivative elapsed time) will be reported as metrics
    max_queries = 500
    # Lookback window in seconds for filtering queries (default: 300)
    # Only queries that executed within this time window will be collected
    lookback_window = 300
    # Enable plan collection (default: true)
    plan_enabled = true
    # Plan object cache TTL (default: 1h)
    plan_cache_ttl = "1h"
    # Maximum runtime in seconds for statement metrics collection (default: 30)
    # If collection takes longer than this, plan collection will be skipped
    max_run_time = 30
    # Disable last active time filter (default: false)
    # If true, queries will be selected randomly instead of by last active time
    disable_last_active = false

  ## Config DBM activity (current active queries)
  ## Collects information about currently executing queries and active sessions
  [inputs.oracle.dbm.activity]
    # Set true to enable collecting active query information
    enabled = true
    # Collection interval for activity metrics (default: 10s)
    collection_interval = "10s"
    # Maximum number of rows to collect from V$SESSION (default: 1000)
    db_rows_limit = 1000

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
  #   interval = "10s"

  [inputs.oracle.tags]
    # some_tag = "some_value"
    # more_tag = "some_other_value"
`
)
