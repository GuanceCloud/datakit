---
title     : 'eBPF'
summary   : '通过 eBPF 采集 Linux 网络数据'
icon      : 'icon/ebpf'
dashboard :
  - desc  : '暂无'
    path  : '-'
monitor   :
  - desc  : '暂无'
    path  : '-'
---

<!-- markdownlint-disable MD025 -->
# eBPF
<!-- markdownlint-enable -->

---

{{.AvailableArchs}}

---

eBPF 采集器，采集主机网络 TCP、UDP 连接信息，Bash 执行日志等。本采集器主要包含 `ebpf-net`、`ebpf-conntrack` 及 `ebpf-bash` 三个插件：

- `ebpf-net`:
    - 数据类别：Network
    - 由 `netflow/httpflow/dnsflow` 构成，分别用于采集主机 TCP/UDP 连接统计信息和主机 DNS 解析信息；

- `ebpf-bash`:

    - 数据类别：Logging
    - 采集 Bash 的执行日志，包含 Bash 进程号、用户名、执行的命令和时间等；

- `ebpf-conntrack`: [:octicons-tag-24: Version-1.8.0](changelog.md#cl-1.8.0) · [:octicons-beaker-24: Experimental](index.md#experimental)
    - 往网络流数据上添加两个标签 `dst_nat_ip` 和 `dst_nat_port`；

## 配置 {#config}

### 前置条件 {#requirements}

对于 v1.5.6 之前的 DataKit，需执行安装命令进行安装：

- v1.2.13 ~ v1.2.18
    - 安装时[指定环境变量](datakit-install.md#extra-envs)：`DK_INSTALL_EXTERNALS="datakit-ebpf"`
    - DataKit 安装完后，再手动安装 eBPF 采集器：`datakit install --datakit-ebpf`
- v1.2.19+
    - 安装时[指定环境变量](datakit-install.md#extra-envs)：`DK_INSTALL_EXTERNALS="ebpf"`
    - DataKit 安装完后，再手动安装 eBPF 采集器：`datakit install --ebpf`
- v1.5.6+
    - 无需手动安装

在 Kubernetes 环境下部署时，必须挂载主机的 `/sys/kernel/debug` 目录到 Pod 内，可参考最新的 *datakit.yaml*；

### HTTPS 支持 {#https}

[:octicons-tag-24: Version-1.4.6](changelog.md#cl-1.4.6) ·
[:octicons-beaker-24: Experimental](index.md#experimental)

若需要 `ebpf-net` 开启对容器内的进程采集 HTTPS 请求数据采集支持，则需要挂载 overlay 目录到容器

*datakit.yaml* 参考修改：

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

可通过 `cat /proc/mounts` 查看 overlay 挂载点

### Linux 内核版本要求 {#kernel}

目前 Linux 3.10 内核的项目生命周期已经结束，建议您升级至 Linux 4.9 及以上 LTS 版内核。

除 CentOS 7.6+ 和 Ubuntu 16.04 以外，其他发行版本推荐 Linux 内核版本高于 4.9，否则可能无法启动 eBPF 采集器

<!-- markdownlint-disable MD046 -->
???+ warning "内核限制"

    Datakit 版本低于 v1.5.2 时，对于 CentOS 7.6+ 不能开启 `ebpf-net` 类别中的 `httpflow` 数据采集，由于其 Linux 3.10.x 内核不支持 eBPF 程序中的 BPF_PROG_TYPE_SOCKET_FILTER 类型；

    Datakit 版本低于 **v1.5.2** 时，由于 `BPF_FUNC_skb_load_bytes` 不存在于 Linux Kernel <= 4.4，若需开启 `httpflow`，需要 Linux Kernel >= 4.5，此问题待后续优化；
<!-- markdownlint-enable -->

### 已启用 SELinux 的系统 {#selinux}

对于启用了 SELinux 的系统，需要关闭其，执行以下命令进行关闭：

```shell
setenforce 0
```

### 采集器配置 {#input-config}

<!-- markdownlint-disable MD046 -->
=== "主机安装"

    进入 DataKit 安装目录下的 `conf.d/{{.Catalog}}` 目录，复制 `{{.InputName}}.conf.sample` 并命名为 `{{.InputName}}.conf`。示例如下：
    
    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```
    
    默认配置不开启 `ebpf-bash`，若需开启在 `enabled_plugins` 配置项中添加 `ebpf-bash`；
    
    配置好后，重启 DataKit 即可。

=== "Kubernetes"

    Kubernetes 中可以通过 ConfigMap 或者直接默认启用 eBPF 采集器两种方式来开启采集：

    1. ConfigMap 方式参照通用的[安装示例](datakit-daemonset-deploy.md#configmap-setting)。
    2. 在 *datakit.yaml* 中的环境变量 `ENV_ENABLE_INPUTS` 中追加 `ebpf`，此时使用默认配置，即仅开启 `ebpf-net` 网络数据采集
    
    ```yaml
    - name: ENV_ENABLE_INPUTS
           value: cpu,disk,diskio,mem,swap,system,hostobject,net,host_processes,container,ebpf
    ```

    通过以下环境变量可以调整 Kubernetes 中 eBPF 采集配置：
    
    | 环境变量名                                    | 对应的配置参数项                 | 参数示例                    |
    | :---                                        | ---                           | ---                        |
    | `ENV_INPUT_EBPF_ENABLED_PLUGINS`            | `enabled_plugins`             | `ebpf-net,ebpf-bash,ebpf-conntrack`       |
    | `ENV_INPUT_EBPF_L7NET_ENABLED`              | `l7net_enabled`               | `httpflow,httpflow-tls`    |
    | `ENV_INPUT_EBPF_IPV6_DISABLED`              | `ipv6_disabled`               | `false/true`               |
    | `ENV_INPUT_EBPF_EPHEMERAL_PORT`             | `ephemeral_port`              | `32768`                    |
    | `ENV_INPUT_EBPF_INTERVAL`                   | `interval`                    | `60s`                      |
<!-- markdownlint-enable -->

## 指标 {#metric}

以下所有数据采集，默认会追加名为 `host` 的全局 tag（tag 值为 DataKit 所在主机名），也可以在配置中通过 `[inputs.{{.InputName}}.tags]` 指定其它标签：

``` toml
 [inputs.{{.InputName}}.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
  # ...
```

{{ range $i, $m := .Measurements }}

### `{{$m.Name}}`

- 标签

{{$m.TagsMarkdownTable}}

- 指标列表

{{$m.FieldsMarkdownTable}}

{{ end }}
