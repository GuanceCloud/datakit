{{.CSS}}
# eBPF
---

- DataKit 版本：{{.Version}}
- 操作系统支持：`{{.AvailableArchs}}`

eBPF 采集器，采集主机网络 TCP、UDP 连接信息，Bash 执行日志等。本采集器主要包含 `ebpf-net` 及 `ebpf-bash` 俩类:

  * `ebpf-net`:
    * 数据类别: Network
    * 由 netflow、httpflow 和 dnsflow 构成，分别用于采集主机 TCP/UDP 连接统计信息和主机 DNS 解析信息；

  * `ebpf-bash`:
    * 数据类别: Logging
    * 采集 Bash 的执行日志，包含 Bash 进程号、用户名、执行的命令和时间等;

## 前置条件

由于该采集器的可执行文件体积较大，自 v1.2.13 起不再打包在 DataKit 中，但 DataKit 容器镜像默认包含该采集器；对于新装 DataKit，需执行安装命令进行安装，有以下两种方法：

- v1.2.13 ~ v1.2.18
  - 安装时[指定环境变量](datakit-install#f9858758)：`DK_INSTALL_EXTERNALS="datakit-ebpf"`
  - DataKit 安装完后，再手动安装 eBPF 采集器：`datakit install --datakit-ebpf`
- v1.2.19+
  - 安装时[指定环境变量](datakit-install#f9858758)：`DK_INSTALL_EXTERNALS="ebpf"`
  - DataKit 安装完后，再手动安装 eBPF 采集器：`datakit install --ebpf`

### Linux 内核版本要求

```txt
* 目前 Linux 3.10 内核的项目生命周期已经结束，建议您升级至 Linux 4.9 及以上 LTS 版内核
```

除 CentOS 7.6+ 和 Ubuntu 16.04 以外，其他发行版本需要 Linux 内核版本高于 4.0.0, 可使用命令 `uname -r` 查看，如下：

```sh
$ uname -r 
5.11.0-25-generic
```

对于 CentOS 7.6+ 和 Ubuntu 16.04 不能开启 ebpf-net 类别中的 httpflow 数据采集，由于其 Linux 3.10.x 内核不支持 eBPF 程序中的 BPF_PROG_TYPE_SOCKET_FILTER 类型

### 已启用 SELinux 的系统

对于启用了 SELinux 的系统，需要关闭其(待后续优化)，执行以下命令进行关闭:

```sh
setenforce 0
```

## 配置

进入 DataKit 安装目录下的 `conf.d/{{.Catalog}}` 目录，复制 `{{.InputName}}.conf.sample` 并命名为 `{{.InputName}}.conf`。示例如下：

```toml
{{.InputSample}}
```

默认配置不开启 ebpf-bash，若需开启在 `enabled_plugins` 配置项中添加 `ebpf-bash`；

配置好后，重启 DataKit 即可。

### Kubernetes 安装

1. 参照通用的 [ConfigMap 安装示例](datakit-daemonset-deploy#fb919c14)。
2. 在 datakit.yaml 中的环境变量 `ENV_ENABLE_INPUTS` 中追加 `ebpf`，此时使用默认配置，即仅开启 ebpf-net 网络数据采集

```yaml
   - name: ENV_ENABLE_INPUTS
          value: cpu,disk,diskio,mem,swap,system,hostobject,net,host_processes,container,ebpf

```

## 指标集

以下所有数据采集，默认会追加名为 `host` 的全局 tag（tag 值为 DataKit 所在主机名），也可以在配置中通过 `[inputs.{{.InputName}}.tags]` 指定其它标签：

``` toml
 [inputs.{{.InputName}}.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
  # ...
```

{{ range $i, $m := .Measurements }}

### `{{$m.Name}}`

-  标签

{{$m.TagsMarkdownTable}}

- 指标列表

{{$m.FieldsMarkdownTable}}

{{ end }}
