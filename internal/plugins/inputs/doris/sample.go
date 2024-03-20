// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package doris

const sampleCfg = `
[[inputs.prom]]
  ## Collector alias.
  source = "doris"

  ## (Optional) Collect interval: (defaults to "30s").
  # interval = "15s"

  ## Exporter URLs.
  urls = ["http://127.0.0.1:8030/metrics","http://127.0.0.1:8040/metrics"]

  ## Stream Size. 
  ## The source stream segmentation size.
  ## Default 1, source stream undivided. 
  stream_size = 0

  ## TLS configuration.
  tls_open = false
  # tls_ca = "/tmp/ca.crt"
  # tls_cert = "/tmp/peer.crt"
  # tls_key = "/tmp/peer.key"

  ## Set to 'true' to enable election.
  election = true

  ## disable setting host tag for this input
  disable_host_tag = false

  ## disable setting instance tag for this input
  disable_instance_tag = false

  ## Measurement name.
  ## If measurement_name is empty, split metric name by '_', the first field after split as measurement set name, the rest as current metric name.
  ## If measurement_name is not empty, using this as measurement set name.
  ## Always add 'measurement_prefix' prefix at last.
  measurement_name = "doris_common"

## Customize measurement set name.
## Treat those metrics with prefix as one set.
## Prioritier over 'measurement_name' configuration.
[[inputs.prom.measurements]]
  prefix = "doris_fe_"
  name = "doris_fe"

[[inputs.prom.measurements]]
  prefix = "doris_be_"
  name = "doris_be"

[[inputs.prom.measurements]]
  prefix = "jvm_"
  name = "doris_jvm"

## Customize tags.
# [inputs.prom.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
`
