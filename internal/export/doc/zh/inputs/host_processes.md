---
title     : '进程'
summary   : '采集进程的指标和对象数据'
tags:
  - '主机'
__int_icon      : 'icon/process'
dashboard :
  - desc  : '进程'
    path  : 'dashboard/zh/process'
monitor   :
  - desc  : '暂无'
    path  : '-'
---


{{.AvailableArchs}}

---

进程采集器可以对系统中各种运行的进程进行实施监控， 获取、分析进程运行时各项指标，包括内存使用率、占用 CPU 时间、进程当前状态、进程监听的端口等，并根据进程运行时的各项指标信息，用户可以在<<<custom_key.brand_name>>>中配置相关告警，使用户了解进程的状态，在进程发生故障时，可以及时对发生故障的进程进行维护。

<!-- markdownlint-disable MD046 -->

???+ attention

    进程采集器（不管是对象还是指标），在 macOS 上可能消耗比较大，导致 CPU 飙升，可以手动将其关闭。目前默认采集器仍然开启进程对象采集器（默认 5min 运行一次）。

<!-- markdownlint-enable MD046 -->

## 配置 {#config}

### 前置条件 {#requirements}

- 进程采集器默认不采集进程指标数据，如需采集指标相关数据，可在 `{{.InputName}}.conf` 中 将 `open_metric` 设置为 `true`。比如：

```toml
[[inputs.host_processes]]
    ...
    open_metric = true
```

### 采集器配置 {#input-config}

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

<!-- markdownlint-disable MD024 -->

{{ range $i, $m := .Measurements }}

{{if eq $m.Type "metric"}}

### `{{$m.Name}}`

{{$m.Desc}}

- 标签

{{$m.TagsMarkdownTable}}

- 字段列表

{{$m.FieldsMarkdownTable}}
{{end}}

{{ end }}

## 对象 {#object}

{{ range $i, $m := .Measurements }}

{{if eq $m.Type "object"}}

### `{{$m.Name}}`

{{$m.Desc}}

- 标签

{{$m.TagsMarkdownTable}}

- 字段列表

{{$m.FieldsMarkdownTable}}
{{end}}

{{ end }}

<!-- markdownlint-enable -->
