// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package promsd

import "time"

var (
	inputName            = "promsd"
	maxScrapersPerWorker = 100
	workerNumber         = 5

	maxInterval = time.Minute * 10
	minInterval = time.Second * 10
)

const (
	sampleConfig = `
[[inputs.promsd]]
  ## Collector alias.
  source = "prom_sd"

  # [inputs.promsd.file_sd_config]
  #   # Patterns for files from which target groups are extracted.
  #   files = ["<filename_pattern>"]
  #   # Refresh interval to re-read the files.
  #   refresh_interval = "5m"

  # [inputs.promsd.http_sd_config]
  #   # Service_url of HTTP service discovery endpoint
  #   service_url = "http://<your-http-sd-service>:8080/prometheus/targets"
  #   # Interval for refreshing service discovery targets
  #   refresh_interval = "5m"
  #
  #   # Advanced HTTP configuration (TLS, proxies, etc.)
  #   # Uncomment and configure as needed:
  #   [inputs.promsd.http_sd_config.auth]
  #     # Bearer token file path for authentication (auto adds Authorization header)
  #     bearer_token_file = ""
  #     # --- TLS Configuration ---
  #     # insecure_skip_verify = false
  #     # ca_certs = ["/opt/tls/ca.crt"]
  #     # cert     = "/opt/tls/client.crt"
  #     # cert_key = "/opt/tls/client.key"

  # [inputs.promsd.consul_sd_config]
  #   # Address of the Consul server (format: host:port)
  #   server = "localhost:8500"
  #   # API path prefix when Consul is behind a reverse proxy/API gateway
  #   path_prefix = ""
  #   # ACL token for authentication (consider using environment variables for security)
  #   token = ""
  #   # Specific datacenter to query (empty = default datacenter)
  #   datacenter = ""
  #   # Namespace for tenant isolation
  #   namespace = "default"
  #   # Administrative partition
  #   partition = ""
  #   # Protocol scheme to use (http or https)
  #   scheme = "http"
  #   # List of services to monitor (empty array = all services)
  #   services = [ ]
  #   # Native Consul filter expression (replaces deprecated tags/node_meta)
  #   # Example: 'Service.Tags contains "metrics" and Node.Meta.rack == "a1"'
  #   filter = ""
  #   # Allow stale results to reduce load on Consul cluster
  #   allow_stale = true
  #   # Interval for refreshing service discovery targets
  #   refresh_interval = "5m"
  #
  #   # Advanced HTTP configuration (TLS, proxies, etc.)
  #   # Uncomment and configure as needed:
  #   [inputs.promsd.consul_sd_config.auth]
  #     ## --- TLS Configuration ---
  #     # insecure_skip_verify = false
  #     # ca_certs = ["/opt/tls/ca.crt"]
  #     # cert     = "/opt/tls/client.crt"
  #     # cert_key = "/opt/tls/client.key"

  [inputs.promsd.scrape]
    ## Protocol scheme for target connections (http or https)
    scheme = "http"
    ## Path to metrics endpoint (default is /metrics)
    metrics_path = "/metrics"
    ## Query parameters in URL-encoded format
    ## Format: "key1=value1&key2=value2&key3=value3"
    ## Example: "debug=true&module=http"
    params = ""

    ## Scraping interval (default 30s)
    interval = "30s"

    ## Additional HTTP headers (optional)
    # [inputs.promsd.scrape.http_headers]
    #   X-Scrape-Token = "secret-value"
    #   Cache-Control = "no-cache"

    ## TLS configuration (optional)
    [inputs.promsd.scrape.auth]
      ## Bearer token file path for authentication (auto adds Authorization header)
      # bearer_token_file = ""
      ## --- TLS Configuration ---
      # insecure_skip_verify = false
      # ca_certs = ["/opt/tls/ca.crt"]
      # cert     = "/opt/tls/client.crt"
      # cert_key = "/opt/tls/client.key"

  [inputs.promsd.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
`
)
