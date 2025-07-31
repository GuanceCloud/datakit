// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package promsd

var (
	inputName            = "promsd"
	maxScrapersPerWorker = 100
	workerNumber         = 5
)

const (
	sampleConfig = `
[[inputs.promsd]]
  ## Collector alias.
  source = "prom_sd"

  # [inputs.promsd.http_sd_config]
  #   ## Service_url of HTTP service discovery endpoint
  #   service_url = "http://<your-http-sd-service>:8080/prometheus/targets"

  #   ## Refresh interval, defaults to "3m".
  #   refresh_interval = "3m"

  #   [inputs.promsd.http_sd_config.auth]
  #     ## --- TLS Configuration ---
  #     # insecure_skip_verify = false
  #     # ca_certs = ["/opt/tls/ca.crt"]
  #     # cert     = "/opt/tls/client.crt"
  #     # cert_key = "/opt/tls/client.key"

  [inputs.promsd.scrape]
    ## Protocol for target connections (http or https)
    scheme = "http"
    ## Scraping interval, defaults to "30s"
    interval = "30s"

    [inputs.promsd.scrape.auth]
      ## Bearer token file path for authentication (auto adds Authorization header)
      # bearer_token_file = ""
      ## --- TLS Configuration ---
      # insecure_skip_verify = false
      # ca_certs = ["/opt/tls/ca.crt"]
      # cert     = "/opt/tls/client.crt"
      # cert_key = "/opt/tls/client.key"

    ## Add HTTP headers to data pulling (Example basic authentication).
    # [inputs.promsd.scrape.http_headers]
    # Authorization = ""

  [inputs.promsd.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
`
)
