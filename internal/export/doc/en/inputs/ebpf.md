---
title     : 'eBPF'
summary   : 'Collect Linux network data through eBPF'
__int_icon      : 'icon/ebpf'
dashboard :
  - desc  : 'N/A'
    path  : '-'
monitor   :
  - desc  : 'N/A'
    path  : '-'
---

<!-- markdownlint-disable MD025 -->
# eBPF
<!-- markdownlint-enable -->
---

{{.AvailableArchs}}

---

eBPF collector, collecting host network TCP, UDP connection information, Bash execution log, etc. This collector mainly includes `ebpf-net`, `ebpf-conntrack` and `ebpf-bash` three plugins:

* `ebpf-net`:
    * Data category: Network
    * It is composed of netflow, httpflow and dnsflow, which are used to collect host TCP/UDP connection statistics and host DNS resolution information respectively;

* `ebpf-bash`:

    * Data category: Logging
    * Collect Bash execution log, including Bash process number, user name, executed command and time, etc.;

* `ebpf-conntrack`: [:octicons-tag-24: Version-1.8.0](../datakit/changelog.md#cl-1.8.0) · [:octicons-beaker-24: Experimental](../datakit/index.md#experimental)
    * Add two tags `dst_nat_ip` and `dst_nat_port` to the network flow data.


* `ebpf-trace`:
    * Application call relationship tracking.

* `bpf-netlog`:
    * Data category: `Logging`, `Network`
    * This plugin implements `ebpf-net`’s `netflow/httpflow`

## Configuration {#config}

### Preconditions {#requirements}

For DataKit before v1.5.6, you need to execute the installation command to install:

* v1.2.13 ~ v1.2.18
    * Install time [specify environment variable](datakit-install.md#extra-envs)：`DK_INSTALL_EXTERNALS="datakit-ebpf"`
    * After the DataKit is installed, manually install the eBPF collector: `datakit install --datakit-ebpf`
* v1.2.19+
    * [specify environment variable](datakit-install.md#extra-envs)：`DK_INSTALL_EXTERNALS="ebpf"` when installing
    * After the DataKit is installed, manually install the eBPF collector: `datakit install --ebpf`
* v1.5.6+
    * No manual installation required

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

### HTTPS Support {#https}

[:octicons-tag-24: Version-1.4.6](../datakit/changelog.md#cl-1.4.6) ·
[:octicons-beaker-24: Experimental](../datakit/index.md#experimental)

If eBPF-net is required to start https request data collection support for processes in the container, you need to mount the overlay directory to the container.

`datakit.yaml` reference changes:

<!-- markdownlint-disable MD046 -->
=== "Docker"

    ```yaml
    ...
            volumeMounts:
            - mountPath: /var/lib/docker/overlay2/
              name: vol-docker-overlay
              readOnly: true
    ...
          volumes:
          - hostPath:
              path: /var/lib/docker/overlay2/
              type: ""
            name: vol-docker-overlay
    ```

=== "Containerd"

    ```yaml
            volumeMounts:
            - mountPath: /run/containerd/io.containerd.runtime.v2.task/
              name: vol-containerd-overlay
              readOnly: true
    ...
          volumes:
          - hostPath:
              path: /run/containerd/io.containerd.runtime.v2.task/
              type: ""
            name: vol-containerd-overlay
    ```
<!-- markdownlint-enable -->
You can view the overlay mount point through `cat /proc/mounts`


### Collector Configuration {#input-config}

<!-- markdownlint-disable MD046 -->
=== "Host Installation"

    Go to the `conf.d/{{.Catalog}}` directory under the DataKit installation directory, copy `{{.InputName}}.conf.sample` and name it `{{.InputName}}.conf`. Examples are as follows:
    
    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```
    
    The default configuration does not turn on eBPF-bash. If you need to turn on, add `eBPF-bash` in the `enabled_plugins` configuration item;
    
    After configuration, restart DataKit.

=== "Kubernetes"

    In Kubernetes, collection can be started by ConfigMap or directly enabling eBPF collector by default:
    
    1. Refer to the generic [Installation Sample](../datakit/datakit-daemonset-deploy.md#configmap-setting) for the ConfigMap mode.
    2. Append `eBPF` to the environment variable `ENV_ENABLE_INPUTS` in `datakit.yaml`, using the default configuration, which only turns on eBPF-net network data collection.
    
    ```yaml
    - name: ENV_ENABLE_INPUTS
           value: cpu,disk,diskio,mem,swap,system,hostobject,net,host_processes,container,eBPF
    ```

### Environment variables configuration {#input-cfg-field-env}

    The eBPF collection configuration in Kubernetes can be adjusted by the following environment variables:
    
    | Environment variable name | Corresponding configuration parameter item | Parameter example | Description |
    | :------------------------ | ------------------------------------------ |------------------ | ----------- |
    | `ENV_INPUT_EBPF_ENABLED_PLUGINS` | `enabled_plugins` | `ebpf-net,ebpf-trace` | Used to configure the built-in plug-in of the collector |
    | `ENV_INPUT_EBPF_L7NET_ENABLED` | `l7net_enabled` | `httpflow` | Enable http protocol data collection |
    | `ENV_INPUT_EBPF_IPV6_DISABLED` | `ipv6_disabled` | `false` | Whether the system does not support IPv6 |
    | `ENV_INPUT_EBPF_EPHEMERAL_PORT` | `ephemeral_port` | `32768` | The starting position of the ephemeral port |
    | `ENV_INPUT_EBPF_INTERVAL` | `interval` | `60s` | Data aggregation period |
    | `ENV_INPUT_EBPF_TRACE_SERVER` | `trace_server` | `<datakit ip>:<datakit port>` | The address of DataKit, you need to enable DataKit `ebpftrace` collector to receive eBPF link data |
    | `ENV_INPUT_EBPF_TRACE_ALL_PROCESS` | `trace_all_process` | `false` | Trace all processes in the system |
    | `ENV_INPUT_EBPF_TRACE_NAME_BLACKLIST` | `trace_name_blacklist` | `datakit,datakit-ebpf` | The process with the specified process name will be prohibited from collecting link data. The process in the example has been hard-coded to prohibit collection |
    | `ENV_INPUT_EBPF_TRACE_ENV_BLACKLIST` | `trace_env_blacklist` | `datakit,datakit-ebpf` | Processes containing any specified environment variable name will be prohibited from collecting link data |
    | `ENV_INPUT_EBPF_TRACE_ENV_LIST` | `trace_env_list` | `DK_BPFTRACE_SERVICE,DD_SERVICE,OTEL_SERVICE_NAME` | The link data of the process containing any specified environment variables will be tracked and reported |
    | `ENV_INPUT_EBPF_TRACE_NAME_LIST` | `trace_name_list` | `chrome,firefox` | Processes with process names in the specified set will be tracked and reported |
    | `ENV_INPUT_EBPF_CONV_TO_DDTRACE` | `conv_to_ddtrace` | `false` | Convert all application `trace_id` to decimal strings |
    | `ENV_NETLOG_BLACKLIST` | `netlog_blacklist` | `ip_saddr=='127.0.0.1' \|\| ip_daddr=='127.0.0.1'` | Used to filter packets after packet capture |
    | `ENV_NETLOG_METRIC_ONLY` | `netlog_metric_only` | `false` | In addition to network metrics, also enable the network logs |
    | `ENV_INPUT_EBPF_CPU_LIMIT` | `cpu_limit` | `"2.0"` | Maximum number of CPU cores used per unit time limit |
    | `ENV_INPUT_EBPF_MEM_LIMIT` | `mem_limit` | `"4GiB"` | Memory size usage limit |
    | `ENV_INPUT_EBPF_NET_LIMIT` | `net_limit` | `"100MiB/s"` | Network bandwidth (any network card) limit |

<!-- markdownlint-enable -->

### The blacklist function of the `netlog` plug-in

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
| 2        | `!`   | Logical NOT, unary operator | Right             |
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

## Metric {#metric}

For all of the following data collections, a global tag named `host` is appended by default (the tag value is the host name of the DataKit), or other tags can be specified in the configuration by `[inputs.{{.InputName}}.tags]`:

``` toml
 [inputs.{{.InputName}}.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
  # ...
```

{{ range $i, $m := .Measurements }}

### `{{$m.Name}}`

* tag

{{$m.TagsMarkdownTable}}

* metric list

{{$m.FieldsMarkdownTable}}

{{ end }}
