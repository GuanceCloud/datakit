// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package coredns

const configSample = `
[[inputs.prom]]
url = "http://127.0.0.1:9153/metrics"
source = "coredns"
metric_types = ["counter", "gauge"]

## filter metrics by names
metric_name_filter = ["^coredns_(acl|cache|dnssec|forward|grpc|hosts|template|dns)_([a-z_]+)$"]

# measurement_prefix = ""
# measurement_name = "prom"

interval = "10s"

# tags_ignore = [""]

## TLS config
tls_open = false
# tls_ca = "/tmp/ca.crt"
# tls_cert = "/tmp/peer.crt"
# tls_key = "/tmp/peer.key"

## customize metrics
[[inputs.prom.measurements]]
  prefix = "coredns_acl_"
  name = "coredns_acl"

[[inputs.prom.measurements]]
  prefix = "coredns_cache_"
  name = "coredns_cache"

[[inputs.prom.measurements]]
  prefix = "coredns_dnssec_"
  name = "coredns_dnssec"

[[inputs.prom.measurements]]
  prefix = "coredns_forward_"
  name = "coredns_forward"

[[inputs.prom.measurements]]
  prefix = "coredns_grpc_"
  name = "coredns_grpc"

[[inputs.prom.measurements]]
  prefix = "coredns_hosts_"
  name = "coredns_hosts"

[[inputs.prom.measurements]]
  prefix = "coredns_template_"
  name = "coredns_template"

[[inputs.prom.measurements]]
  prefix = "coredns_dns_"
  name = "coredns"`
