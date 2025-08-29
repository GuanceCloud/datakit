// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package promv2

const sampleConfig = `
[[inputs.promv2]]
  ## Collector alias.
  source = "prom"

  url = "http://127.0.0.1:9100/metrics"

  ## Scraping interval, defaults to "30s"
  interval = "30s"

  ## Measurement name.
  ## If measurement_name is empty, split metric name by '_', the first field after split as measurement set name, the rest as current metric name.
  ## If measurement_name is not empty, using this as measurement set name.
  measurement_name = ""

  ## Keep Exist Metric Name
  ## If the keep_exist_metric_name is true, keep the raw value for field names.
  keep_exist_metric_name = true

  ## Use the timestamps provided by the target. Set to 'false' to use the scrape time.
  honor_timestamps = true

  ## Bearer token file path for authentication (auto adds Authorization header)
  # bearer_token_file = ""

  ## --- TLS Configuration ---
  ## [DANGER] Skip TLS verification (INSECURE, testing only!)
  # insecure_skip_verify = false
  ## Root CA certificates for server verification (PEM files)
  # ca_certs = ["/opt/tls/ca.crt"]
  ## Client certificate for mTLS (PEM format)
  # cert     = "/opt/tls/client.crt"
  ## Client private key for mTLS (PEM format)
  # cert_key = "/opt/tls/client.key"

  ## Set to 'true' to enable election.
  election = true

  ## Add HTTP headers to data pulling (Example basic authentication).
  # [inputs.promv2.http_headers]
  # Authorization = ""

  [inputs.promv2.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
`
