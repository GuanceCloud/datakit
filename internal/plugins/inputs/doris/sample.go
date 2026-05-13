// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package doris

const sampleCfg = `
[[inputs.doris]]
  ## Doris FE SQL endpoint.
  host = "127.0.0.1"
  port = 9030

  ## Doris user and password.
  user = "datakit"
  password = "<PASS>"

  ## Gathering interval.
  interval = "10s"

  ## Connection timeout for Doris SQL and metrics endpoints.
  connect_timeout = "10s"

  ## Set true to enable election.
  election = true

  ## FE metrics URL. Used to collect qps and avg_query_time for database object.
  fe_metric_url = "http://127.0.0.1:8030/metrics"

  ## TLS Config for FE metrics endpoint. Used when fe_metric_url is HTTPS.
  # [inputs.doris.metric_tls]
    # tls_ca = "/etc/doris/metric-ca.pem"
    # tls_cert = "/etc/doris/metric-cert.pem"
    # tls_key = "/etc/doris/metric-key.pem"

    ## Use TLS but skip chain & host verification
    # insecure_skip_verify = true

    ## by default, support TLS 1.2 and above.
    ## set to true if server side uses TLS 1.0 or TLS 1.1
    # allow_tls10 = false

  ## collect Doris object
  [inputs.doris.object]
    ## Set true to enable Doris object collection.
    enabled = true

    ## Interval to collect Doris object which will be greater than collection interval.
    interval = "600s"

  ## TLS Config
  # [inputs.doris.tls]
    # tls_ca = "/etc/doris/ca.pem"
    # tls_cert = "/etc/doris/cert.pem"
    # tls_key = "/etc/doris/key.pem"

    ## Use TLS but skip chain & host verification
    # insecure_skip_verify = true

    ## by default, support TLS 1.2 and above.
    ## set to true if server side uses TLS 1.0 or TLS 1.1
    # allow_tls10 = false

  ## Run a custom SQL query and collect corresponding metrics.
  # [[inputs.doris.custom_queries]]
  #   sql = '''
  #     SELECT
  #       TABLE_SCHEMA AS table_schema,
  #       COUNT(*) AS table_count
  #     FROM information_schema.tables
  #     GROUP BY TABLE_SCHEMA
  #   '''
  #   metric = "doris_custom"
  #   tags = ["table_schema"]
  #   fields = ["table_count"]
  #   interval = "10s"

  [inputs.doris.tags]
    # some_tag = "some_value"
    # more_tag = "some_other_value"

## The following Prometheus config collects full Doris FE/BE metrics.
## Keep it when you still want the existing doris_fe/doris_be/doris_common/doris_jvm metrics.
[[inputs.prom]]
  ## Collector alias.
  source = "doris"

  ## (Optional) Collect interval: (defaults to "30s").
  # interval = "15s"

  ## Exporter URLs.
  urls = ["http://127.0.0.1:8030/metrics","http://127.0.0.1:8040/metrics"]

  ## Stream Size. 
  ## The source stream segmentation size.
  ## Default 1, source stream undivided. 
  stream_size = 0

  ## TLS configuration.
  tls_open = false
  # tls_ca = "/tmp/ca.crt"
  # tls_cert = "/tmp/peer.crt"
  # tls_key = "/tmp/peer.key"

  ## Set to 'true' to enable election.
  election = true

  ## disable setting host tag for this input
  disable_host_tag = false

  ## disable setting instance tag for this input
  disable_instance_tag = false

  ## Measurement name.
  ## If measurement_name is empty, split metric name by '_', the first field after split as measurement set name, the rest as current metric name.
  ## If measurement_name is not empty, using this as measurement set name.
  ## Always add 'measurement_prefix' prefix at last.
  measurement_name = "doris_common"

## Customize measurement set name.
## Treat those metrics with prefix as one set.
## Prioritier over 'measurement_name' configuration.
[[inputs.prom.measurements]]
  prefix = "doris_fe_"
  name = "doris_fe"

[[inputs.prom.measurements]]
  prefix = "doris_be_"
  name = "doris_be"

[[inputs.prom.measurements]]
  prefix = "jvm_"
  name = "doris_jvm"

## Customize tags.
# [inputs.prom.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
`
