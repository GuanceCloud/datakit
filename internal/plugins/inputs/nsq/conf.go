// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package nsq

const sampleCfg = `
[[inputs.nsq]]
  ## NSQ Lookupd HTTP API endpoint
  lookupd = "http://localhost:4161"

  ## NSQD HTTP API endpoint
  ## example:
  ##   ["http://localhost:4151"]
  nsqd = []
  
  ## time units are "ms", "s", "m", "h"
  interval = "10s"

  ## Set true to enable election
  election = true
  
  ## Optional TLS Config
  # tls_ca = "/etc/telegraf/ca.pem"
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"
  ## Use TLS but skip chain & host verification
  # insecure_skip_verify = false
  
  [inputs.nsq.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
`
