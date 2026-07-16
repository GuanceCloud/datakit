---
title     : 'Network Path'
summary   : 'Collect network paths and end-to-end quality for TCP, UDP, and ICMP targets'
tags:
  - 'Network'
__int_icon      : 'icon/net'
dashboard :
  - desc  : 'N/A'
    path  : '-'
monitor   :
  - desc  : 'N/A'
    path  : '-'
---


{{.AvailableArchs}}

---

The NetPath collector actively probes network paths from the DataKit node. It supports TCP, UDP, and ICMP targets. Targets can be configured statically or discovered dynamically by local traffic sources such as `datakit-ebpf`. NetPath currently supports Linux and macOS; Windows is not supported.

Each execution produces one `netpath` record in the Network (`N`) category, including source and destination context, end-to-end latency, ICMP loss metrics, and per-hop results from multiple traceroute runs.

## Configuration {#config}

<!-- markdownlint-disable MD046 -->
=== "Host Installation"

    Go to the `conf.d/samples` directory under the DataKit installation directory, copy `netpath.conf.sample`, and rename it to `netpath.conf`:

    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```

    After configuration, [restart DataKit](../datakit/datakit-service-how-to.md#manage-service).

=== "Kubernetes"

    Enable the collector with a [ConfigMap](../datakit/datakit-daemonset-deploy.md#configmap-setting), or add `netpath` to `ENV_DEFAULT_ENABLED_INPUTS` and configure it with environment variables:

{{ CodeBlock .InputENVSample 4 }}
<!-- markdownlint-enable -->

### Static Targets {#static-targets}

Configure static targets with `[[inputs.netpath.targets]]`:

```toml
[[inputs.netpath.targets]]
  name = "api-gateway"
  target = "api.example.com"
  port = 443
  protocol = "tcp"
  interval = "60s"
  timeout = "1s"
  max_ttl = 30
  traceroute_queries = 3
  e2e_queries = 10

  [inputs.netpath.targets.tags]
    service = "api"
```

With the `auto` protocol, targets with a port use TCP and targets without a port use ICMP. UDP is currently supported only on Linux and requires permission to receive raw ICMP responses. The effective `max_ttl` limit is `60` for TCP/ICMP traceroute and `255` for Linux UDP traceroute.

`traceroute_queries` is the number of complete traceroute executions. Each execution independently probes from TTL 1 until it reaches the destination or `max_ttl`, and produces one `message.runs[]` element; it is not a retry count for the same TTL. For a domain target, every run performs its own DNS lookup. The actual target is stored in `runs[].destination.ip_address`, and one test is not guaranteed to cover every IP of the domain.

`e2e_queries` is the number of independent end-to-end probes and defaults to `10`. E2E does not inspect intermediate hops and runs concurrently with traceroute: TCP uses connection responses, ICMP uses Echo Reply, and Linux UDP uses destination ICMP responses. E2E packets use a normal end-to-end IP TTL rather than the traceroute `max_ttl`. A domain is resolved independently for E2E, and the selected address is stored in `e2e_dest_ip`; it may differ from `message.runs[].destination.ip_address`. UDP application silence is counted in `e2e_unknown` rather than being treated directly as loss.

### Dynamic Targets {#dynamic-targets}

The dynamic target endpoint is `POST /v1/netpath/candidates`. `dynamic.enabled` is enabled by default. If the DataKit HTTP listener is reachable outside the host, configure a non-empty `dynamic.token`; clients must send the same token in the `X-Datakit-Netpath-Token` header.

Dynamic candidates with a hostname are deduplicated and probed by hostname; otherwise the IP is used. Different observed IPs for the same non-NAT hostname share one scheduled task, which runs at `interval` while it remains alive for `ttl`; translated destinations additionally retain the original destination tuple in their identity. Dynamic context, queue, rate, and worker limits do not consume or drop configured static targets. Keep `monitor_ip_without_domain = false` where possible, and configure `dynamic.filters` before increasing traffic volume to exclude unwanted namespaces, networks, or ports.

For hostname candidates, destination host and CIDR filters are re-evaluated against every DNS result before traceroute or E2E sends a probe packet. Candidate identity fields are limited to 1024 bytes each and 4096 bytes in total per test, including request-level defaults. Request-level and test-level tags are limited to 64 combined tags per candidate. Tag keys are limited to 128 bytes and values to 1024 bytes. Custom tags cannot override fixed contract fields such as `traceroute_status`, `traceroute_fail_type`, or endpoint tuple tags.

Dynamic discovery from `datakit-ebpf` also requires:

```toml
[inputs.ebpf]
  network_path_enabled = true
  network_path_api = "http://127.0.0.1:9529/v1/netpath/candidates"
  network_path_token = ""
```

### Reverse DNS {#reverse-dns}

Reverse DNS is disabled by default. When enabled, DataKit performs PTR lookups for the destination and responsive hop IPs and uses a TTL cache to limit repeated queries:

```toml
[inputs.netpath.reverse_dns]
  enabled = true
  timeout = "500ms"
  cache_ttl = "10m"
  cache_size = 4096
```

## Data Model {#data}

Result tags identify the task, source, destination, and path. The main tags are:

- task: `task_name`, `task_source`, `origin`, `run_type`, `protocol`;
- flow tuple: `src_ip`, `src_port`, `dst_ip`, and `dst_port`, plus `dst_nat_ip` and `dst_nat_port` when DNAT occurs;
- endpoint context: `dst_domain`, `source_host`, `source_service`, `source_process`, `source_container_id`, `src_cloud_provider`, and `dst_cloud_provider`;
- actual probe route: `probe_source_ip`, `probe_gateway_ip`, `probe_interface`, `probe_netns`;
- status: `traceroute_protocol`, `traceroute_status`, `traceroute_fail_type`, and `e2e_status`.

Source and destination tags do not depend on per-hop results. Their tuple names align with NetFlow: `dst_*` is the original destination and `dst_nat_*` is the translated probe destination. Unknown ports use `"*"`; a missing source IP falls back to `probe_source_ip`. The optional endpoint `*_cloud_provider` tags provide cloud-provider context for searching.

NetPath does not upload `branch_key` or `path_key`. Query history with the structured task, source, and destination tags above; derive actual route branches and their changes from `message.runs[].hops[]`.

Use `traceroute_status` (`reached`, `partial`, or `failed`) for path completion. It describes whether traceroute reached the destination, not end-to-end quality.

Use the independent `e2e_*` fields as the canonical end-to-end quality fields:

- `e2e_dest_ip`: the independently resolved IPv4 address used by E2E probes;
- `e2e_packets_sent`, `e2e_packets_received`, and `e2e_unknown`: sent, responded, and ambiguous probe counts;
- `e2e_probe_loss_percent`: no-response percentage among determinate probes; unknown outcomes are excluded;
- `e2e_rtt_avg`, `e2e_rtt_min`, and `e2e_rtt_max`: end-to-end RTT in microseconds;
- `e2e_rtt_variation_avg` and `e2e_rtt_variation_max`: absolute RTT differences between consecutive successful probes in send order, in microseconds.

The frontend should use `e2e_rtt_avg` for whole-path latency and `e2e_probe_loss_percent` for the end-to-end probe no-response percentage. `e2e_status` is independent of `traceroute_status`. TCP connection refused/RST proves destination reachability and is counted as received. `e2e_probe_loss_percent` is not the actual packet-loss rate of an intermediate device.

### Per-hop Path {#message}

The complete path is stored as standard JSON in the `message` field:

```json
{
  "runs": [
    {
      "run_id": "1",
      "destination": {
        "ip_address": "8.8.8.8",
        "port": 443,
        "reverse_dns": ["dns.google"]
      },
      "hops": [
        {
          "ttl": 1,
          "ip_address": "10.0.0.1",
          "reverse_dns": ["gateway.local"],
          "rtt": 0.8315,
          "reachable": true
        },
        {
          "ttl": 2,
          "reachable": false
        },
        {
          "ttl": 3,
          "ip_address": "8.8.8.8",
          "rtt": 12.45,
          "reachable": true,
          "asn": 15169,
          "as_name": "GOOGLE",
          "as_prefix": "8.8.8.0/24",
          "cloud_provider": "gcp"
        }
      ]
    }
  ],
  "hop_count": {
    "avg": 3,
    "min": 3,
    "max": 3
  }
}
```

| Field | Type | Description |
| --- | --- | --- |
| `runs[].run_id` | string | Sequential traceroute run ID inside `message`. |
| `runs[].destination.ip_address` | string | Actual destination IPv4 used by this run. |
| `runs[].destination.port` | uint16 | TCP or UDP destination port. |
| `runs[].destination.reverse_dns` | string[] | Hostname used for a domain target. |
| `runs[].hops[].ttl` | int | Hop number. |
| `runs[].hops[].reachable` | bool | Whether this TTL probe received a response. |
| `runs[].hops[].ip_address` | string | Hop IP. Omitted for non-responsive hops; `"*"` is not used. |
| `runs[].hops[].reverse_dns` | string[] | Optional reverse DNS names. |
| `runs[].hops[].rtt` | number | Round-trip time of this TTL probe in milliseconds. It is not the latency between adjacent hops. |
| `runs[].hops[].asn` | uint64 | ASN added for public IPs by Kodo using a local offline database. |
| `runs[].hops[].as_name` | string | Optional autonomous-system organization name. |
| `runs[].hops[].as_prefix` | string | Optional autonomous-system network prefix. |
| `runs[].hops[].cloud_provider` | string | Optional cloud provider added by Kodo using a local IP ownership database. |
| `hop_count.avg/min/max` | number | Hop-count statistics across traceroute runs. |

ASN and cloud-provider fields are optional server-side enrichment. They are absent for private, CGNAT, and non-responsive hops, when the database is not deployed, or when no record matches. Enrichment does not send hop IPs to a third-party service and does not infer a cloud provider from `as_name`.

If a probe fails before producing a path, `message` can contain error text instead of JSON. Handle it together with `traceroute_status=failed`, `traceroute_fail_type`, and `traceroute_fail_reason`.

<!-- markdownlint-disable MD046 -->
???+ warning

    `reachable=false` only means that a response was not received for this TTL probe. Device policy or ICMP rate limiting can cause this, so it must not be interpreted directly as actual packet loss at that hop. Per-hop `probe_count`, `response_count`, and `timeout_count` are not uploaded.
<!-- markdownlint-enable -->

## Logging {#logging}

{{ range $i, $m := .Measurements }}

### `{{$m.Name}}`

{{$m.MarkdownTable}}

{{ end }}

## Security and Privacy {#security}

- DataKit uploads hop IPs observed by traceroute because they are required for path visualization.
- DataKit does not send hop IPs to an external ASN lookup service. Kodo performs ASN enrichment with an offline database.
- Reverse DNS causes DNS queries from the DataKit network. Keep it disabled if these queries may disclose sensitive information.
- If raw hop IPs must not leave the customer network, assess whether NetPath can be enabled. Moving ASN enrichment to Kodo does not hide raw IPs from the observability platform.

## Troubleshooting {#troubleshooting}

| Symptom | Check |
| --- | --- |
| Candidate API returns HTTP 401 | The request token must match `dynamic.token`. |
| Candidate API returns HTTP 404 | Ensure `dynamic.enabled` is not disabled. |
| `ip_without_domain` | Provide a hostname or explicitly enable `monitor_ip_without_domain`. |
| `traceroute_fail_type=permission` | Provide the raw socket permissions required by ICMP/UDP traceroute. |
| Only part of the path is visible | Check for `traceroute_status=partial`; intermediate devices may not respond to TTL probes. |
| ASN is absent | Ensure an ASN/ISP MMDB is deployed in Kodo; private and non-responsive hops do not have ASN fields. |
