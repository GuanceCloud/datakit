// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package phpfpm

const sampleCfg = `
[[inputs.phpfpm]]
  ##(optional) collect interval, default is 10 seconds
  interval = '10s'

  ## URL to fetch PHP-FPM pool metrics. Defaults to HTTP (e.g., http://localhost/status).
  ## For TCP (e.g., tcp://127.0.0.1:9000/status) or Unix socket (e.g., unix:///run/php/php-fpm.sock;/status),
  ## set use_fastcgi to true.
  status_url = "http://localhost/status"

  ## (optional) use fastcgi, default is false (use http)
  use_fastcgi = false

  ## Set true to enable election
  election = true

  [inputs.phpfpm.tags]
    # some_tag = "some_value"
    # more_tag = "some_other_value"
`
