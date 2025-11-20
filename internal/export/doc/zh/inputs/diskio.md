---
title     : '磁盘 IO'
summary   : '采集磁盘 IO 指标数据'
tags:
  - '主机'
__int_icon      : 'icon/diskio'
dashboard :
  - desc  : '磁盘 IO'
    path  : 'dashboard/zh/diskio'
monitor   :
  - desc  : '主机检测库'
    path  : 'monitor/zh/host'
---


{{.AvailableArchs}}

---

磁盘 IO 采集器用于磁盘流量和时间的指标的采集。

## 配置 {#config}

成功安装 DataKit 并启动后，会默认开启 DiskIO 采集器，无需手动开启。

### 前置条件 {#requirement}

对于部分旧版本 Windows 操作系统，如若遇到 DataKit 报错： **"The system cannot find the file specified."**

请以管理员身份运行 PowerShell，并执行：

```powershell
$ diskperf -Y
...
```

在执行成功后需要重启 DataKit 服务。

### 采集器配置 {#input-config}

<!-- markdownlint-disable MD046 -->

=== "主机安装"

    进入 DataKit 安装目录下的 `conf.d/samples` 目录，复制 `{{.InputName}}.conf.sample` 并命名为 `{{.InputName}}.conf`。示例如下：

    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```

    配置好后，[重启 DataKit](../datakit/datakit-service-how-to.md#manage-service) 即可。

=== "Kubernetes"

    可通过 [ConfigMap 方式注入采集器配置](../datakit/datakit-daemonset-deploy.md#configmap-setting) 或 [配置 ENV_DATAKIT_INPUTS](../datakit/datakit-daemonset-deploy.md#env-setting) 开启采集器。

    也支持以环境变量的方式修改配置参数（需要在 ENV_DEFAULT_ENABLED_INPUTS 中加为默认采集器）：

{{ CodeBlock .InputENVSampleZh 4 }}

<!-- markdownlint-enable -->

## 指标 {#metric}

以下所有数据采集，默认会追加名为 `host` 的全局 tag（tag 值为 DataKit 所在主机名），也可以在配置中通过 `[[inputs.{{.InputName}}.tags]]` 另择 host 来命名。

{{ range $i, $m := .Measurements }}

### `{{$m.Name}}`

{{$m.MarkdownTable}}

{{ end }}

## FAQ {#faq}

### `diskio` 指标在 Linux 主机上的数据来源是什么 {#linux-diskio}

在 Linux 主机上，指标从 */proc/diskstats* 文件获取并通过解析和计算得出；其中每一列的解释可参考[文档](https://www.kernel.org/doc/Documentation/ABI/testing/procfs-diskstats){:target="_blank"}；

部分数据来源列和指标的对应关系为：

| `diskstats` 字段                           | `diskio` 指标                                                                                                  |
| ---                                        | ---                                                                                                            |
| col04: reads completed successfully        | `reads`                                                                                                        |
| col05: reads merged                        | `merged_reads`                                                                                                 |
| col06: sectors read                        | `read_bytes = col06 * sector_size`; `read_bytes/sec = (read_bytes - last(read_bytes))/(time - last(time))`     |
| col07: time spent reading (ms)             | `read_time`                                                                                                    |
| col08: writes completed                    | `writes`                                                                                                       |
| col09: writes merged                       | `merged_writes`                                                                                                |
| col10: sectors written                     | `write_bytes = col10 * sector_size`; `write_bytes/sec = (write_bytes - last(write_bytes))/(time - last(time))` |
| col11: time spent writing (ms)             | `write_time`                                                                                                   |
| col12: I/Os currently in progress          | `iops_in_progress`                                                                                             |
| col13: time spent doing I/Os (ms)          | `io_time`                                                                                                      |
| col14: weighted time spent doing I/Os (ms) | `weighted_io_time`                                                                                             |

注意：

1. 扇区大小（`sector_size`）为 512 字节；
1. 除 `read_bytes/sec` 和 `write_bytes/sec` 外均为递增值。
