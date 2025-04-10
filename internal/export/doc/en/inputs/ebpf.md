---
title     : 'eBPF'
summary   : 'Collect Linux network data through eBPF'
tags:
  - 'EBPF'
  - 'NETWORK'
__int_icon      : 'icon/ebpf'
dashboard :
  - desc  : 'N/A'
    path  : '-'
monitor   :
  - desc  : 'N/A'
    path  : '-'
---


{{.AvailableArchs}}

---

eBPF collector, collecting host network TCP, UDP connection information, Bash execution log, etc. This collector mainly includes `ebpf-net`, `ebpf-conntrack` and `ebpf-bash` three plugins:

- `ebpf-net`:
    - Data category: Network
    - It is composed of netflow, httpflow and dnsflow, which are used to collect host TCP/UDP connection statistics and host DNS resolution information respectively;

- `ebpf-bash`:

    - Data category: Logging
    - Collect Bash execution log, including Bash process number, user name, executed command and time, etc.;

- `ebpf-conntrack`: [:octicons-tag-24: Version-1.8.0](../datakit/changelog.md#cl-1.8.0) · [:octicons-beaker-24: Experimental](../datakit/index.md#experimental)
    - Add two tags `dst_nat_ip` and `dst_nat_port` to the network flow data.


- `ebpf-trace`:
    - Application call relationship tracking.

- `bpf-netlog`:
    - Data categories: `Logging`, `Network`
    - This plugin implements the collection of network logs `bpf_net_l4_log/bpf_net_l7_log`, and can also replace `ebpf-net`'s `netflow/httpflow` data collection when the kernel does not support eBPF;

## Configuration {#config}

### Preconditions {#requirements}

For DataKit before v1.5.6, you need to execute the installation command to install:

- v1.2.13 ~ v1.2.18
    - Install time [specify environment variable](../datakit/datakit-install.md#extra-envs)：`DK_INSTALL_EXTERNALS="datakit-ebpf"`
    - After the DataKit is installed, manually install the eBPF collector: `datakit install --datakit-ebpf`
- v1.2.19+
    - [specify environment variable](../datakit/datakit-install.md#extra-envs)：`DK_INSTALL_EXTERNALS="ebpf"` when installing
    - After the DataKit is installed, manually install the eBPF collector: `datakit install --ebpf`
- v1.5.6+
    - No manual installation required

When deploying in Kubernetes environment, you must mount the host's' `/sys/kernel/debug` directory into pod, refer to the latest `datakit.yaml`;

### Linux Kernel Version Requirement {#kernel}

In addition to CentOS 7.6+ and Ubuntu 16.04, other distributions recommend that the Linux kernel version is higher than 4.9, otherwise the eBPF collector may not start.

If you want to enable the  *eBPF-conntrack*  plugin, usually requires a higher kernel version, such as v5.4.0 etc., please confirm whether the symbols in the kernel contain `nf_ct_delete` and `__nf_conntrack_hash_insert`, you can execute the following command to view:

```sh
cat /proc/kallsyms | awk '{print $3}' | grep "^nf_ct_delete$\|^__nf_conntrack_hash_insert$"
```
<!-- markdownlint-disable MD046 -->
???+ warning "kernel restrictions"

    When the DataKit version is lower than **v1.5.2**, the httpflow data collection in the eBPF-net category cannot be enabled for CentOS 7.6+, because its Linux 3.10.x kernel does not support the BPF_PROG_TYPE_SOCKET_FILTER type in the eBPF program;

    When the DataKit version is lower than **v1.5.2**, because BPF_FUNC_skb_load_bytes does not exist in Linux Kernel <= 4.4, if you want to enable httpflow, you need Linux Kernel >= 4.5, and this problem will be further optimized;
<!-- markdownlint-enable -->

### SELinux-enabled System {#selinux}

For SELinux-enabled systems, you need to shut them down (pending subsequent optimization), and execute the following command to shut them down:

```sh
setenforce 0
```

### Collector Configuration {#input-config}

<!-- markdownlint-disable MD046 -->
=== "Host Installation"

    Go to the `conf.d/{{.Catalog}}` directory under the DataKit installation directory, copy `{{.InputName}}.conf.sample` and name it `{{.InputName}}.conf`. The example is as follows:
    
    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```
    
    After configuration, restart DataKit.

=== "Kubernetes"

    In Kubernetes, you can enable collection through ConfigMap or directly enable the eBPF collector by default:

    1. For the ConfigMap method, refer to the general [Installation Example](../datakit/datakit-daemonset-deploy.md#configmap-setting).
    2. Add `ebpf` to the environment variable `ENV_ENABLE_INPUTS` in *datakit.yaml*. In this case, the default configuration is used, that is, only `ebpf-net` network data collection is enabled.
    
    ```yaml
    - name: ENV_ENABLE_INPUTS
           value: cpu,disk,diskio,mem,swap,system,hostobject,net,host_processes,container,ebpf
    ```

### Environment variables and configuration items {#input-cfg-field-env}

The following environment variables can be used to adjust the eBPF collection configuration in Kubernetes:

Configuration items:

- `enabled_plugins`:
    - Description: Used to configure the built-in plugins for the collector
    - Environment variable: `ENV_INPUT_EBPF_ENABLED_PLUGINS`
    - Example: `ebpf-net,ebpf-trace`

- `l7net_enabled`
    - Description: Enable http protocol data collection
    - Environment variable: `ENV_INPUT_EBPF_L7NET_ENABLED`
    - Example: `httpflow`

- `interval`
    - Description: Set the sampling time interval
    - Environment variable: `ENV_INPUT_EBPF_INTERVAL`
    - Example: `1m30s`

- `ipv6_disabled`
    - Description: Whether the system does not support IPv6
    - Environment variable: `ENV_INPUT_EBPF_IPV6_DISABLED`
    - Example: `false`

- `ephemeral_port`
    - Description: Ephemeral port start location
    - Environment variable: `ENV_INPUT_EBPF_EPHEMERAL_PORT`
    - Example: `32768`

- `pprof_host`
    - Description: pprof host
    - Environment variable: `ENV_INPUT_EBPF_PPROF_HOST`
    - Example: `127.0.0.1`

- `pprof_port`
    - Description: pprof port
    - Environment variable: `ENV_INPUT_EBPF_PPROF_PORT`
    - Example: `6061`

<!-- - `interval`
    - Description: Data aggregation period
    - Environment variable: `ENV_INPUT_EBPF_INTERVAL`
    - Example: `60s` -->

- `trace_server`
    - Description: The address of DataKit ELinker/Datakit to enable the `ebpftrace` collector
    - Environment variable: `ENV_INPUT_EBPF_TRACE_SERVER`
    - Example: `<ip>:<port>`

- `trace_all_process`
    - Description: Trace all processes in the system
    - Environment variable: `ENV_INPUT_EBPF_TRACE_ALL_PROCESS`
    - Example: `false`

- `trace_name_blacklist`
    - Description: The process with the specified process name will be disabled from collecting trace data
    - Environment variable: `ENV_INPUT_EBPF_TRACE_NAME_BLACKLIST`
    - Example:

- `trace_env_blacklist`
    - Description: Any process containing any of the specified environment variable names will be disabled from collecting trace data
    - Environment variable: `ENV_INPUT_EBPF_TRACE_ENV_BLACKLIST`
    - Example: `DKE_DISABLE_ETRACE`

- `trace_env_list`
    - Description: Link data for processes with any specified environment variables will be traced and reported
    - Environment variable: `ENV_INPUT_EBPF_TRACE_ENV_LIST`
    - Example: `DK_BPFTRACE_SERVICE,DD_SERVICE,OTEL_SERVICE_NAME`

- `trace_name_list`
    - Description: Processes whose names are in the specified set will be traced and reported
    - Environment variable: `ENV_INPUT_EBPF_TRACE_NAME_LIST`
    - Example: `chrome,firefox`

- `conv_to_ddtrace`
    - Description: Convert all application side link IDs to decimal strings for compatibility purposes, not used unless necessary
    - Environment variable: `ENV_INPUT_EBPF_CONV_TO_DDTRACE`
    - Example: `false`

- `netlog_blacklist`
    - Description: Used to filter packets after packet capture
    - Environment variable: `ENV_INPUT_EBPF_NETLOG_BLACKLIST`
    - Example: `ip_saddr=='127.0.0.1' \|\| ip_daddr=='127.0.0.1'`

- `netlog_metric`
    - Description: Collect network metrics from network packet analysis
    - Environment variable: `ENV_INPUT_EBPF_NETLOG_METRIC`
    - Example: `true`

- `netlog_log`
    - Description: Collect network logs from network packet analysis
    - Environment variable: `ENV_INPUT_EBPF_NETLOG_LOG`
    - Example: `false`

- `cpu_limit`
    - Description: The maximum number of CPU cores used per unit time. When the upper limit is reached, the collector exits.
    - Environment variable: `ENV_INPUT_EBPF_CPU_LIMIT`
    - Example: "2.0"`

- `mem_limit`
    - Description: Memory size usage limit
    - Environment variable: `ENV_INPUT_EBPF_MEM_LIMIT`
    - Example: `"4GiB"`

- `net_limit`
    - Description: Network bandwidth (any network card) limit
    - Environment variable: `ENV_INPUT_EBPF_NET_LIMIT`
    - Example: `"100MiB/s"`

- `sampling_rate`
    - Description: The sampling rate when the eBPF collector reports data, ranging from `0.01 to 1.00`; Mutually exclusive with the `samping_rate_pts_per_min` setting
    - Environment variable: `ENV_INPUT_EBPF_SAMPLING_RATE`
    - Example: `0.50`

- `sampling_rate_pts_per_min`
    - Description: Set the data volume threshold per minute when the eBPF collector reports data, and dynamically adjust the sampling rate
    - Environment variable: `ENV_INPUT_EBPF_SAMPLING_RATE_PTSPERMIN`
    - Example: `1500`

- `workload_labels`
    - Description: Set all specified labels of the K8s workload to be added to the data
    - Environment variable: `ENV_INPUT_EBPF_WORKLOAD_LABELS`
    - Example: `app,project_id`

- `workload_label_prefix`
    - Description: Add a prefix to the k8s workload label
    - Environment variable: `ENV_INPUT_EBPF_WORKLOAD_LABEL_PREFIX`
    - Example: `k8s_workload_label_`

<!-- markdownlint-enable -->

## eBPF Tracing function {#ebpf-tracing}

`ebpf-trace` collects and analyzes the network data read and written by the process on the host, and tracks the kernel-level threads/user-level threads (such as golang goroutine) of the process to generate link eBPF Span. This data needs to be collected by `ebpftrace` for further processing.

When using, you need to deploy the eBPF collector with link data collection enabled on multiple nodes, then you need to send all eBPF Span data to the same DataKit ELinker/DataKit with the [`ebpftrace`](./ebpftrace.md#ebpftrace-config) collector plug-in enabled. For more configuration details, see the [eBPF link document](./ebpftrace.md#ebpf-config)

<!-- markdownlint-disable MD013 -->
### The blacklist function of the `bpf-netlog` plug-in {#blacklist}
<!-- markdownlint-enable -->

Filter rule example:

Single rule:

The following rules filter network data with ip `1.1.1.1` and port 80. (Line breaks allowed after operator)

```py
(ip_saddr == "1.1.1.1" || ip_saddr == "1.1.1.1") &&
      (src_port == 80 || dst_port == 80)
```

Multiple rules:

Use `;` or `\n` to separate the rules. If any rule is met, the data will be filtered.

```py
udp
ip_saddr == "1.1.1.1" && (src_port == 80 || dst_port == 80);
ip_saddr == "10.10.0.1" && (src_port == 80 || dst_port == 80)

ipnet_contains("127.0.0.0/8", ip_saddr); ipv6
```

Data available for filtering:

This filter is used to filter network data. Comparable data is as follows:

| key name      | type | description                                                                             |
| ------------- | ---- | --------------------------------------------------------------------------------------- |
| `tcp`         | bool | Whether it is `TCP` protocol                                                            |
| `udp`         | bool | Whether it is `UDP` protocol                                                            |
| `ipv4`        | bool | Whether it is `IPv4` protocol                                                           |
| `ipv6`        | bool | Whether it is `IPv6` protocol                                                           |
| `src_port`    | int  | Source port (based on the observed network card/host/container as the reference system) |
| `dst_port`    | int  | target port                                                                             |
| `ip_saddr`    | str  | Source `IPv4` network address                                                           |
| `ip_saddr`    | str  | Target `IPv4` network address                                                           |
| `ip6_saddr`   | str  | Source `IPv6` network address                                                           |
| `ip6_daddr`   | str  | Destination `IPv6` network address                                                      |
| `k8s_src_pod` | str  | source `pod` name                                                                       |
| `k8s_dst_pod` | str  | target `pod` name                                                                       |

Operator:

Operators from highest to lowest:

| Priority | Op     | Name                        | Binding Direction |
| -------- | ------ | --------------------------- | ----------------- |
| 1        | `()`   | parentheses                 | left              |
| 2        | `!`    | Logical NOT, unary operator | Right             |
| 3        | `!=`   | Not equal to                | Left              |
| 3        | `>=`   | Greater than or equal to    | Left              |
| 3        | `>`    | greater than                | left              |
| 3        | `==`   | equal to                    | left              |
| 3        | `<=`   | Less than or equal to       | Left              |
| 3        | `<`    | less than                   | left              |
| 4        | `&&`   | Logical AND                 | Left              |
| 4        | `\|\|` | Logical OR                  | Left              |

function:

1. **ipnet_contains**

    Function signature: `fn ipnet_contains(ipnet: str, ipaddr: str) bool`

    Description: Determine whether the address is within the specified network segment

     Example:

    ```py
    ipnet_contains("127.0.0.0/8", ip_saddr)
    ```

    If the `ip_saddr` value is "127.0.0.1", then this rule returns `true` and the TCP connection packet/UDP packet will be filtered.

2. **has_prefix**

    Function signature: `fn has_prefix(s: str, prefix: str) bool`

    Description: Specifies whether the field contains a certain prefix

    Example:

    ```py
    has_prefix(k8s_src_pod, "datakit-") || has_prefix(k8s_dst_pod, "datakit-")
    ```

    This rule returns `true` if the pod name is `datakit-kfez321`.

## Network aggregation data {#network}

For all of the following data collections, a global tag named `host` is appended by default (the tag value is the host name of the DataKit), or other tags can be specified in the configuration by `[inputs.{{.InputName}}.tags]`:

``` toml
 [inputs.{{.InputName}}.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
  # ...
```

{{ range $i, $m := .Measurements }}

{{if eq $m.Type "network"}}

### `{{$m.Name}}`

- Tags

{{$m.TagsMarkdownTable}}

- Metrics

{{$m.FieldsMarkdownTable}}

{{ end }}

{{ end }}

## Logging {#logging}

{{ range $i, $m := .Measurements }}

{{if eq $m.Type "logging"}}

### `{{$m.Name}}`

- Tags

{{$m.TagsMarkdownTable}}

- Metrics

{{$m.FieldsMarkdownTable}}

{{ end }}

{{ end }}

## Tracing {#tracing}

{{ range $i, $m := .Measurements }}

{{if eq $m.Type "tracing"}}

### `{{$m.Name}}`

- Tags

{{$m.TagsMarkdownTable}}

- Metrics

{{$m.FieldsMarkdownTable}}
{{end}}

{{ end }}
