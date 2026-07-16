// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package netpath

const sampleConfig = `
[[inputs.netpath]]
  ## Default protocol for static targets: tcp/udp/icmp/auto.
  protocol = "tcp"

  ## Default static target probe interval.
  interval = "60s"

  ## Per-probe timeout.
  timeout = "1s"

  ## Independent end-to-end probes. These do not inspect intermediate hops.
  e2e_queries = 10

  ## Maximum traceroute TTL and number of complete traceroute runs. The
  ## effective TTL limit is 60 for TCP/ICMP and 255 for Linux UDP.
  max_ttl = 30
  traceroute_queries = 3

  ## Static network path targets.
  # [[inputs.netpath.targets]]
  #   name = "api-gateway"
  #   target = "api.example.com"
  #   port = 443
  #   protocol = "tcp"
  #   interval = "60s"
  #   timeout = "1s"
  #   max_ttl = 30
  #   traceroute_queries = 3
  #   e2e_queries = 10
  #   [inputs.netpath.targets.tags]
  #     service = "api"

  ## Dynamic targets discovered from local traffic sources such as datakit-ebpf.
  [inputs.netpath.dynamic]
    enabled = true
    ## Optional only when both client and accepted server address are loopback.
    ## All other requests require a token.
    ## Configure the same token for datakit-ebpf, which sends it in this header:
    ## X-Datakit-Netpath-Token: <token>
    token = ""

    ## auto/tcp/udp/icmp. auto uses candidate protocol and falls back to tcp for
    ## candidates with a port, icmp for address-only candidates. Traceroute
    ## requires raw socket permission; UDP traceroute is supported on Linux.
    protocol = "auto"

    ## Dynamic candidates are deduplicated and kept for ttl. The scheduler runs
    ## each candidate at interval while it is alive.
    contexts_limit = 5000
    ## Total estimated bytes retained by stored and currently running contexts.
    contexts_bytes_limit = 67108864
    ttl = "50m"
    interval = "20m"
    flush_interval = "10s"
    max_per_minute = 150
    workers = 4
    timeout = "1s"
    ## The effective TTL limit is 60 for TCP/ICMP and 255 for Linux UDP.
    max_ttl = 30
    traceroute_queries = 3
    e2e_queries = 10
    ## Maximum concurrent candidate admissions waiting for the store lock.
    input_queue = 1000
    process_queue = 1000
    max_tests_per_request = 1000
    max_body_bytes = 1048576

    ## Set true to allow address-only candidates that do not carry a domain or
    ## hostname. Keeping this false reduces noisy high-cardinality path tests.
    monitor_ip_without_domain = false

    ## Exclude candidate rules. Conditions inside one rule are ANDed; values in
    ## the same condition are ORed. A candidate matching any rule is dropped
    ## before entering the scheduler. Hostname destinations are checked again
    ## against destination host/CIDR rules after every DNS lookup and before
    ## any probe packet is sent.
    # [[inputs.netpath.dynamic.filters]]
    #   name = "ignore-kube-system"
    #   namespaces = ["kube-system"]
    #
    # [[inputs.netpath.dynamic.filters]]
    #   name = "ignore-private-db"
    #   dest_cidrs = ["10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16"]
    #   ports = [5432, 6379]

  ## Optional reverse DNS enrichment for destination and hop IPs. Disabled by
  ## default to avoid adding DNS lookup latency to every traceroute result.
  [inputs.netpath.reverse_dns]
    enabled = false
    timeout = "500ms"
    cache_ttl = "10m"
    cache_size = 4096

  [inputs.netpath.tags]
    # some_tag = "some_value"
`
