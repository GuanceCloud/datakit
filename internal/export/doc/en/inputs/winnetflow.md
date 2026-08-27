---
title     : 'WinNetFlow'
summary   : 'Collect per-process L4 network flow metrics on Windows via ETW'
tags:
  - 'Network'
__int_icon      : 'icon/windows'
dashboard :
  - desc  : 'N/A'
    path  : '-'
monitor   :
  - desc  : 'N/A'
    path  : '-'
---


{{.AvailableArchs}}

---

WinNetFlow collects per-process L4 network flow metrics on Windows by consuming
TCP/UDP events from the `Microsoft-Windows-TCPIP` ETW provider. The measurement
core field and endpoint-role schema is compatible with the Linux
`ebpf-net/netflow` collector, so host-level network dashboards can reuse the
same `src/dst`, `client/server`, `conn_side` and traffic fields.

It also consumes `Microsoft-Windows-HttpService` events to collect HTTP request
metrics (`httpflow`, enabled by default) for HTTP.sys based web services such as
IIS, HttpListener and ASP.NET Core, with a schema compatible with the Linux
`ebpf-net/httpflow` collector.

## Configuration {#config}

### Requirements {#requirements}

- OS: Windows 10 / Windows Server 2016 or later (64-bit)
- The collector must run with administrator privileges (required to create a
  real-time ETW session)
- Independent from the eBPF collector; only one should be enabled per host

<!-- markdownlint-disable MD046 -->
=== "Host installation"

    Copy `{{.InputName}}.conf.sample` from the `conf.d/samples` directory of
    your DataKit installation and rename it to `{{.InputName}}.conf`. Example:

    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```

    After configuration, [restart DataKit](../datakit/datakit-service-how-to.md#manage-service).

=== "Kubernetes"

    You can enable the collector via a
    [ConfigMap](../datakit/datakit-daemonset-deploy.md#configmap-setting) or
    [ENV_DATAKIT_INPUTS](../datakit/datakit-daemonset-deploy.md#env-setting).

    Parameters can also be adjusted via environment variables (the collector
    must be added to `ENV_DEFAULT_ENABLED_INPUTS`):

{{ CodeBlock .InputENVSample 4 }}

<!-- markdownlint-enable MD046 -->

## Measurements {#metric}

All collected data is tagged with the `host` tag by default (the DataKit host
name). Additional tags can be set via `[inputs.{{.InputName}}.tags]`:

``` toml
 [inputs.{{.InputName}}.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
  # ...
```

{{ range $i, $m := .Measurements }}

{{$m.MarkdownTable}}

{{ end }}

## Known limitations {#limitations}

- TCP send events carry no packet count; `packets_written` is only populated for
  UDP (message count used as an approximation of packets);
- UDP has no connection semantics; direction is inferred heuristically from
  bound sockets on non-ephemeral ports (`incoming` when the local port matches,
  `outgoing` otherwise);
- TCP connections that already exist when collection starts are attributed as
  incoming when their local endpoint matches a TCP listener snapshot, and as
  outgoing otherwise. The listener snapshot refreshes every 30 seconds.
- Windows ETW does not expose network namespaces, Kubernetes endpoint metadata,
  DNS domains or NAT translation. `dst_nat_ip` and `dst_nat_port` are emitted as
  `N/A`; Kubernetes and namespace tags must come from separately configured
  global tags where applicable.
- Under extreme load a real-time ETW session may drop events; the collector
  logs session statistics (decoded/dropped/parse errors/lost counters) roughly
  every 10 minutes and warns when anomalies are detected.
- The number of flows tracked per interval is capped (default 65536, configurable
  via `max_flows`); new flows beyond the cap are dropped and counted as
  `flows_skipped`.
- `httpflow` only covers HTTP traffic that goes through HTTP.sys (IIS,
  HttpListener, ASP.NET Core, ...); self-hosted socket HTTP servers (e.g. some
  Go/Node services) are not captured.
- HTTP.sys events do not carry the HTTP version or request body size;
  `http_version` is always empty and `bytes_read` is always 0.
  `bytes_written` is only available for cache-served responses (event 16).
- HTTP process attribution is best effort. The connection event PID belongs to
  the client and is intentionally ignored; the collector resolves the server
  PID as soon as HTTP.sys reports a response. `process_name` can be `unknown` if
  the server exits before Windows allows its name to be resolved.
- In-flight HTTP request tracking is capped (default 65536, configurable via
  `max_http_requests`); excess requests are dropped and counted in the periodic
  summary. URL query strings are excluded to avoid collecting secrets and
  unbounded metric cardinality. Request paths longer than
  `httpflow_path_limit` (default 256) are truncated and flagged with
  `truncated`.
