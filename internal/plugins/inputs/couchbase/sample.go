// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package couchbase

const sampleCfg = `
[[inputs.couchbase]]
  ## Collect interval, default is 30 seconds. (optional)
  # interval = "30s"

  ## Timeout: (defaults to "5s"). (optional)
  # timeout = "5s"

  ## Scheme, "http" or "https".
  scheme = "http"

  ## Host url or ip.
  host = "127.0.0.1"

  ## Host port. If "https" will be 18091.
  port = 8091

  ## Additional host port for index metric. If "https" will be 19102.
  additional_port = 9102

  ## Host user name.
  user = "Administrator"

  ## Host password.
  password = "123456"

  ## TLS configuration.
  tls_open = false
  # tls_ca = ""
  # tls_cert = "/var/cb/clientcertfiles/travel-sample.pem"
  # tls_key = "/var/cb/clientcertfiles/travel-sample.key"

  ## Disable setting host tag for this input
  disable_host_tag = false

  ## Disable setting instance tag for this input
  disable_instance_tag = false

  ## Set to 'true' to enable election.
  election = true

# [inputs.couchbase.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
`
