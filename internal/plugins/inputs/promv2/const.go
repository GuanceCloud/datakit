// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package promv2

const sampleConfig = `
[[inputs.promv2]]
  ## Collector alias.
  source = "prom"

  urls = ["http://127.0.0.1:9100/metrics", "http://127.0.0.1:9200/metrics"]

  ## (Optional) Collect interval: (defaults to "30s").
  interval = "30s"

  ## Measurement name.
  ## If measurement_name is empty, split metric name by '_', the first field after split as measurement set name, the rest as current metric name.
  ## If measurement_name is not empty, using this as measurement set name.
  measurement_name = ""

  ## Keep Exist Metric Name
  ## If the keep_exist_metric_name is true, keep the raw value for field names.
  keep_exist_metric_name = false

  ## TLS config
  # insecure_skip_verify = true
  ## Following ca_certs/cert/cert_key are optional, if insecure_skip_verify = true.
  # ca_certs = ["/opt/tls/ca.crt"]
  # cert     = "/opt/tls/client.root.crt"
  # cert_key = "/opt/tls/client.root.key"

  ## Set to 'true' to enable election.
  election = true

  ## Add HTTP headers to data pulling (Example basic authentication).
  # [inputs.promv2.http_headers]
  # Authorization = ""

  [inputs.promv2.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
`
