// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package promsd

import "time"

var (
	inputName            = "promsd"
	maxScrapersPerWorker = 100
	workerCount          = 5

	maxRefreshInterval = time.Minute * 10
	minRefreshInterval = time.Second * 10
)

const (
	sampleConfig = `
[[inputs.promsd]]
  ## Collector alias identifier
  source = "prom_sd"

  # ============================================================================
  # Service Discovery Configuration
  # ============================================================================
  # Choose one of the following service discovery methods:

  ## File-based Service Discovery
  # [inputs.promsd.file_sd_config]
  #   ## Files patterns to read target groups from (supports glob patterns)
  #   files = ["/etc/prometheus/targets/*.json", "/path/to/targets.yaml"]
  #   ## Refresh interval for re-reading files
  #   refresh_interval = "5m"

  ## HTTP-based Service Discovery
  # [inputs.promsd.http_sd_config]
  #   ## Service discovery endpoint URL
  #   service_url = "http://your-sd-service:8080/prometheus/targets"
  #   ## Refresh interval for fetching targets from HTTP endpoint
  #   refresh_interval = "5m"
  #
  #   ## Optional: Custom HTTP headers
  #   [inputs.promsd.http_sd_config.http_headers]
  #     # X-Custom-Header = "value"
  #
  #   ## Optional: Authentication and TLS configuration
  #   [inputs.promsd.http_sd_config.auth]
  #     ## Bearer token file for authentication (automatically adds Authorization header)
  #     bearer_token_file = "/path/to/token"
  #
  #     ## TLS Configuration
  #     # insecure_skip_verify = false
  #     # ca_certs = ["/opt/tls/ca.crt"]
  #     # cert     = "/opt/tls/client.crt"
  #     # cert_key = "/opt/tls/client.key"

  ## Consul Service Discovery
  # [inputs.promsd.consul_sd_config]
  #   ## Consul server address (format: host:port)
  #   server = "localhost:8500"
  #   ## API path prefix when Consul is behind reverse proxy
  #   path_prefix = ""
  #   ## ACL token for authentication (use environment variables for security)
  #   token = ""
  #   ## Datacenter to query (empty for default)
  #   datacenter = ""
  #   ## Namespace for tenant isolation
  #   namespace = "default"
  #   ## Administrative partition
  #   partition = ""
  #   ## Protocol scheme (http or https)
  #   scheme = "http"
  #   ## Services to monitor (empty array monitors all services)
  #   services = []
  #   ## Native Consul filter expression (deprecated tags/node_meta replacement)
  #   # Example: 'Service.Tags contains "metrics" and Node.Meta.rack == "a1"'
  #   filter = ""
  #   ## Enable stale results to reduce Consul cluster load
  #   allow_stale = true
  #   ## Refresh interval for service discovery targets
  #   refresh_interval = "5m"
  #
  #   ## Optional: Authentication and TLS configuration
  #   [inputs.promsd.consul_sd_config.auth]
  #     ## Bearer token file for authentication
  #     bearer_token_file = "/path/to/token"
  #
  #     ## TLS Configuration
  #     # insecure_skip_verify = false
  #     # ca_certs = ["/opt/tls/ca.crt"]
  #     # cert     = "/opt/tls/client.crt"
  #     # cert_key = "/opt/tls/client.key"

  # ============================================================================
  # Scrape Configuration
  # ============================================================================
  [inputs.promsd.scrape]
    ## Protocol scheme for target connections
    scheme = "http"

    ## Metrics endpoint path
    metrics_path = "/metrics"

    ## Query parameters in URL-encoded format
    ## Format: "key1=value1&key2=value2"
    ## Example: "debug=true&module=http"
    params = ""

    ## Scraping interval
    interval = "30s"

    ## Optional: Custom HTTP headers
    [inputs.promsd.scrape.http_headers]
      # Authorization = "Bearer <token>"
      # X-Custom-Header = "value"

    ## Optional: Authentication and TLS configuration
    [inputs.promsd.scrape.auth]
      ## Bearer token file for authentication (automatically adds Authorization header)
      bearer_token_file = ""

      ## TLS Configuration
      # insecure_skip_verify = false
      # ca_certs = ["/opt/tls/ca.crt"]
      # cert     = "/opt/tls/client.crt"
      # cert_key = "/opt/tls/client.key"

  # ============================================================================
  # Additional Tags
  # ============================================================================
  [inputs.promsd.tags]
    # cluster = "production"
    # region  = "us-east-1"
    # team    = "platform"
`
)
