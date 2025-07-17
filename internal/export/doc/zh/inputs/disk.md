---
title     : '磁盘'
summary   : '采集磁盘的指标数据'
tags:
  - '主机'
__int_icon      : 'icon/disk'
dashboard :
  - desc  : '磁盘'
    path  : 'dashboard/zh/disk'
monitor   :
  - desc  : '主机检测库'
    path  : 'monitor/zh/host'
---

{{.AvailableArchs}}

---

磁盘采集器用于主机磁盘信息采集，如磁盘存储空间、Inode 使用情况等。

## 配置 {#config}

成功安装 DataKit 并启动后，会默认开启 Disk 采集器，无需手动开启。

<!-- markdownlint-disable MD046 -->

=== "主机安装"

    进入 DataKit 安装目录下的 `conf.d/{{.Catalog}}` 目录，复制 `{{.InputName}}.conf.sample` 并命名为 `{{.InputName}}.conf`。示例如下：

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

以下所有数据采集，默认会追加名为 `host` 的全局 tag（tag 值为 DataKit 所在主机名），也可以在配置中通过 `[inputs.{{.InputName}}.tags]` 指定其它标签：

```toml
 [inputs.{{.InputName}}.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
  # ...
```

<!-- markdownlint-disable MD046 -->
???+ info "磁盘指标来源"
    在 Linux 中，指标是通过获取 */proc/1/mountinfo* 其中的挂载信息，然后再逐个获取对应挂载点的磁盘指标（`statfs()`）。对 Windows 而言，则通过一系列 Windows API，诸如 `GetLogicalDriveStringsW()` 系统调用获取挂载点，然后再通过 `GetDiskFreeSpaceExW()` 获取磁盘用量信息。

    在 [:octicons-tag-24: Version-1.66.0](../datakit/changelog-2025.md#cl-1.66.0) 版本中，优化了磁盘信息采集，但是相同设备的挂载点仍然会合并成一个，且只取第一个出现的挂载点为准。如果要采集所有的挂载点，需关闭特定的 flag（`merge_on_device/ENV_INPUT_DISK_MERGE_ON_DEVICE`），关闭该合并功能后，磁盘指标集中可能会额外多出非常多的时间线。
<!-- markdownlint-enable -->

{{ range $i, $m := .Measurements }}

### `{{$m.Name}}`

{{$m.MarkdownTable}}

{{ end }}
