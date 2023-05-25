
# eBPF
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

* `ebpf-conntrack`: [:octicons-tag-24: Version-1.8.0](changelog.md#cl-1.8.0) · [:octicons-beaker-24: Experimental](index.md#experimental)
    * Add two tags `dst_nat_ip` and `dst_nat_port` to the network flow data.

## Preconditions {#requirements}

For DataKit before v1.5.6, you need to execute the installation command to install:

- v1.2.13 ~ v1.2.18
  - Install time [specify environment variable](datakit-install.md#extra-envs)：`DK_INSTALL_EXTERNALS="datakit-ebpf"`
  - After the DataKit is installed, manually install the eBPF collector: `datakit install --datakit-ebpf`
- v1.2.19+
  - [specify environment variable](datakit-install.md#extra-envs)：`DK_INSTALL_EXTERNALS="ebpf"` when installing
  - After the DataKit is installed, manually install the eBPF collector: `datakit install --ebpf`
- v1.5.6+
  - No manual installation required

When deploying in Kubernetes environment, you must mount the host's' `/sys/kernel/debug` directory into pod, refer to the latest datakit.yaml;

### HTTPS Support {#https}

[:octicons-tag-24: Version-1.4.6](changelog.md#cl-1.4.6) ·
[:octicons-beaker-24: Experimental](index.md#experimental)

If ebpf-net is required to start https request data collection support for processes in the container, you need to mount the overlay directory to the container.

datakit.yaml reference changes:

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

You can view the overlay mount point through `cat /proc/mounts`

### Linux Kernel Version Requirement {#kernel}

In addition to CentOS 7.6+ and Ubuntu 16.04, other distributions recommend that the Linux kernel version is higher than 4.9, otherwise the ebpf collector may not start.

???+ warning "kernel restrictions"

    When the DataKit version is lower than **v1.5.2**, the httpflow data collection in the ebpf-net category cannot be enabled for CentOS 7.6+, because its Linux 3.10.x kernel does not support the BPF_PROG_TYPE_SOCKET_FILTER type in the eBPF program;

    When the DataKit version is lower than **v1.5.2**, because BPF_FUNC_skb_load_bytes does not exist in Linux Kernel <= 4.4, if you want to enable httpflow, you need Linux Kernel >= 4.5, and this problem will be further optimized;

### SELinux-enabled System {#selinux}

For SELinux-enabled systems, you need to shut them down (pending subsequent optimization), and execute the following command to shut them down:

```sh
setenforce 0
```

## Configuation {#config}

=== "Host Installation"

    Go to the `conf.d/{{.Catalog}}` directory under the DataKit installation directory, copy `{{.InputName}}.conf.sample` and name it `{{.InputName}}.conf`. Examples are as follows:
    
    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```
    
    The default configuration does not turn on ebpf-bash. If you need to turn on, add `ebpf-bash` in the `enabled_plugins` configuration item;
    
    After configuration, restart DataKit.

=== "Kubernetes"

    In Kubernetes, collection can be started by ConfigMap or directly enabling ebpf collector by default:
    
    1. Refer to the generic [Installation Sample](datakit-daemonset-deploy.md#configmap-setting) for the ConfigMap mode.
    2. Append `ebpf` to the environment variable `ENV_ENABLE_INPUTS` in datakit.yaml, using the default configuration, which only turns on ebpf-net network data collection.
    
    ```yaml
    - name: ENV_ENABLE_INPUTS
           value: cpu,disk,diskio,mem,swap,system,hostobject,net,host_processes,container,ebpf
    ```
    
    The ebpf collection configuration in Kubernetes can be adjusted by the following environment variables:
    
    | Environment Variable Name                                    | Corresponding Configuration Parameter Item                 | Parameter Example                    |
    | :---                                        | ---                           | ---                        |
    | `ENV_INPUT_EBPF_ENABLED_PLUGINS`            | `enabled_plugins`             | `ebpf-net,ebpf-bash,ebpf-conntrack`       |
    | `ENV_INPUT_EBPF_L7NET_ENABLED`              | `l7net_enabled`               | `httpflow,httpflow-tls`    |
    | `ENV_INPUT_EBPF_IPV6_DISABLED`              | `ipv6_disabled`               | `false/true`               |
    | `ENV_INPUT_EBPF_EPHEMERAL_PORT`             | `ephemeral_port`              | `32768`                    |
    | `ENV_INPUT_EBPF_INTERVAL`                   | `interval`                    | `60s`                      |


## Measurements {#measurements}

For all of the following data collections, a global tag named `host` is appended by default (the tag value is the host name of the DataKit), or other tags can be specified in the configuration by `[inputs.ebpf.tags]`:

``` toml
 [inputs.ebpf.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
  # ...
```

{{ range $i, $m := .Measurements }}

### `{{$m.Name}}`

- tag

{{$m.TagsMarkdownTable}}

- metric list

{{$m.FieldsMarkdownTable}}

{{ end }}
