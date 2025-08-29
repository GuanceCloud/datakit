// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package prom

const sampleCfg = `
[[inputs.prom]]
  ## Exporter URLs.
  urls = ["http://127.0.0.1:9100/metrics", "http://127.0.0.1:9200/metrics"]

  ## (Optional) Collect interval: (defaults to "30s").
  # interval = "30s"

  ## (Optional) Timeout: (defaults to "30s").
  # timeout = "30s"

  ## Stream Size. 
  ## The source stream segmentation size, (defaults to 1).
  ## 0 source stream undivided. 
  # stream_size = 1

  ## Unix Domain Socket URL. Using socket to request data when not empty.
  uds_path = ""

  ## Ignore URL request errors.
  ignore_req_err = false

  ## Collector alias.
  source = "prom"

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
  # measurement_name = "prom"

  ## Keep Exist Metric Name
  ## If the keep_exist_metric_name is true, keep the raw value for field names.
  keep_exist_metric_name = false

  ## Use the timestamps provided by the target. Set to 'false' to use the scrape time.
  honor_timestamps = true

  ## TLS config
  # insecure_skip_verify = true
  ## Following ca_certs/cert/cert_key are optional, if insecure_skip_verify = true.
  # ca_certs = ["/opt/tls/ca.crt"]
  # cert = "/opt/tls/client.root.crt"
  # cert_key = "/opt/tls/client.root.key"
  ## we can encode these file content in base64 format:
  # ca_certs_base64 = ["LONG_BASE64_STRING......"]
  # cert_base64 = "LONG_BASE64_STRING......"
  # cert_key_base64 = "LONG_BASE64_STRING......"
  # server_name = "your-SNI-name"

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
  # [inputs.prom.auth]
    # type = "bearer_token"
    # token = "xxxxxxxx"
    # token_file = "/tmp/token"

  ## Customize measurement set name.
  ## Treat those metrics with prefix as one set.
  ## Prioritier over 'measurement_name' configuration.
  [[inputs.prom.measurements]]
    prefix = "etcd_network_"
    name = "etcd_network"
    
  [[inputs.prom.measurements]]
    prefix = "etcd_server_"
    name = "etcd_server"

  ## Not collecting those data when tag matched.
  # [inputs.prom.ignore_tag_kv_match]
    # key1 = [ "val1.*", "val2.*"]
    # key2 = [ "val1.*", "val2.*"]

  ## Add HTTP headers to data pulling (Example basic authentication).
  # [inputs.prom.http_headers]
    # Authorization = "Basic bXl0b21jYXQ="

  ## Rename tag key in prom data.
  [inputs.prom.tags_rename]
    overwrite_exist_tags = false

  # [inputs.prom.tags_rename.mapping]
    # tag1 = "new-name-1"
    # tag2 = "new-name-2"
    # tag3 = "new-name-3"

  ## Send collected metrics to center as log.
  ## When 'service' field is empty, using 'service tag' as measurement set name.
  [inputs.prom.as_logging]
    enable = false
    service = "service_name"

  ## Customize tags.
  # [inputs.prom.tags]
    # some_tag = "some_value"
    # more_tag = "some_other_value"
`
