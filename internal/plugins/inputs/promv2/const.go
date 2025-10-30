// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package promv2

const sampleConfig = `
[[inputs.promv2]]
  ## Collector alias for identification.
  source = "prom"

  ## Prometheus metrics endpoint URL.
  url = "http://127.0.0.1:9100/metrics"

  ## Scraping interval. Supports "ns", "us", "ms", "s", "m", "h" units.
  interval = "30s"

  ## Measurement name mapping strategy.
  ## If empty, metric name will be split by '_' character:
  ##   - First segment becomes the measurement name
  ##   - Remaining segments become the metric name
  ## If not empty, this value will be used as the measurement name.
  measurement_name = ""

  ## Preserve original metric name as field name.
  ## When true, keeps the raw metric name; when false, uses the parsed name.
  keep_exist_metric_name = true

  ## Use timestamps from Prometheus metrics.
  ## Set to false to use the current scrape time instead.
  honor_timestamps = true

  ## Enable leader election for distributed deployments.
  ## Only one instance will scrape when multiple datakits are running.
  election = true

  ## Bearer token file path for authentication.
  ## Automatically adds 'Authorization: Bearer <token>' header.
  # bearer_token_file = "/path/to/token"

  ## --- TLS Configuration ---

  ## Skip TLS certificate verification (INSECURE - for testing only!).
  ## Disables verification of server certificates.
  # insecure_skip_verify = false

  ## Root CA certificates for server verification.
  ## List of PEM-encoded certificate files.
  # ca_certs = ["/opt/tls/ca.crt"]

  ## Client certificate for mutual TLS authentication.
  ## PEM-encoded certificate file.
  # cert = "/opt/tls/client.crt"

  ## Client private key for mutual TLS authentication.
  ## PEM-encoded private key file.
  # cert_key = "/opt/tls/client.key"

  ## Custom HTTP headers to add to requests.
  ## Example: Basic authentication or custom authentication tokens.
  # [inputs.promv2.http_headers]
  # Authorization = "Basic dXNlcm5hbWU6cGFzc3dvcmQ="
  # X-Custom-Header = "custom_value"

  ## Additional tags to attach to all metrics.
  ## These tags help identify and filter data points.
  [inputs.promv2.tags]
  # environment = "production"
  # service = "backend"
  # team = "platform"
`
