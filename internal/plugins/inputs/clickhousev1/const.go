// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package clickhousev1

const sampleCfg = `
[[inputs.clickhousev1]]
  ## Exporter URLs.
  urls = ["http://127.0.0.1:9363/metrics"]

  ## Unix Domain Socket URL. Using socket to request data when not empty.
  uds_path = ""

  ## Ignore URL request errors.
  ignore_req_err = false

  ## Collector alias.
  source = "clickhouse"

  ## Collect data output.
  ## Fill this when want to collect the data to local file nor center.
  ## After filling, could use 'datakit debug --prom-conf /path/to/this/conf' to debug local storage measurement set.
  ## Using '--prom-conf' when priority debugging data in 'output' path.
  # output = "/abs/path/to/file"

  ## Collect data upper limit as bytes.
  ## Only available when set output to local file.
  ## If collect data exceeded the limit, the data would be dropped.
  ## Default is 32MB.
  # max_file_size = 0

  ## Metrics type whitelist. Optional: counter, gauge, histogram, summary
  ## Example: metric_types = ["counter", "gauge"], only collect 'counter' and 'gauge'.
  ## Default collect all.
  # metric_types = []

  ## Metrics name whitelist.
  ## Regex supported. Multi supported, conditions met when one matched.
  ## Collect all if empty.
  # metric_name_filter = ["cpu"]

  ## Metrics name blacklist.
  ## If a word both in blacklist and whitelist, blacklist priority.
  ## Regex supported. Multi supported, conditions met when one matched.
  ## Collect all if empty.
  # metric_name_filter_ignore = ["foo","bar"]

  ## Measurement prefix.
  ## Add prefix to measurement set name.
  measurement_prefix = ""

  ## Measurement name.
  ## If measurement_name is empty, split metric name by '_', the first field after split as measurement set name, the rest as current metric name.
  ## If measurement_name is not empty, using this as measurement set name.
  ## Always add 'measurement_prefix' prefix at last.
  # measurement_name = "clickhouse"

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

  ## disable info tag for this input
  disable_info_tag = false

  ## Ignore tags. Multi supported.
  ## The matched tags would be dropped, but the item would still be sent.
  # tags_ignore = ["xxxx"]

  ## Customize authentification. For now support Bearer Token only.
  ## Filling in 'token' or 'token_file' is acceptable.
  # [inputs.clickhousev1.auth]
    # type = "bearer_token"
    # token = "xxxxxxxx"
    # token_file = "/tmp/token"

  ## Customize measurement set name.
  ## Treat those metrics with prefix as one set.
  ## Prioritier over 'measurement_name' configuration.
  [[inputs.clickhousev1.measurements]]
    prefix = "ClickHouseProfileEvents_"
    name = "ClickHouseProfileEvents"

  [[inputs.clickhousev1.measurements]]
    prefix = "ClickHouseMetrics_"
    name = "ClickHouseMetrics"

  [[inputs.clickhousev1.measurements]]
    prefix = "ClickHouseAsyncMetrics_"
    name = "ClickHouseAsyncMetrics"

  [[inputs.clickhousev1.measurements]]
    prefix = "ClickHouseStatusInfo_"
    name = "ClickHouseStatusInfo"

  ## Not collecting those data when tag matched.
  [inputs.clickhousev1.ignore_tag_kv_match]
    # key1 = [ "val1.*", "val2.*"]
    # key2 = [ "val1.*", "val2.*"]

  ## Add HTTP headers to data pulling.
  [inputs.clickhousev1.http_headers]
    # Root = "passwd"
    # Michael = "1234"

  ## Rename tag key in clickhouse data.
  [inputs.clickhousev1.tags_rename]
    overwrite_exist_tags = false
  [inputs.clickhousev1.tags_rename.mapping]
    # tag1 = "new-name-1"
    # tag2 = "new-name-2"
    # tag3 = "new-name-3"

  ## Send collected metrics to center as log.
  ## When 'service' field is empty, using 'service tag' as measurement set name.
  [inputs.clickhousev1.as_logging]
    enable = false
    service = "service_name"

  ## Customize tags.
  [inputs.clickhousev1.tags]
    # some_tag = "some_value"
    # more_tag = "some_other_value"
  
  ## (Optional) Collect interval: (defaults to "30s").
  # interval = "30s"

  ## (Optional) Timeout: (defaults to "30s").
  # timeout = "30s"
  `
