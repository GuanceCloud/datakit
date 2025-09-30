---
title     : 'CPU'
summary   : '采集 CPU 指标数据'
tags:
  - 'HOST'
__int_icon      : 'icon/cpu'
dashboard :
  - desc  : 'CPU'
    path  : 'dashboard/zh/cpu'
monitor   :
  - desc  : '主机检测库'
    path  : 'monitor/zh/host'
---

{{.AvailableArchs}}

---

CPU 采集器用于系统 CPU 使用率等指标的采集。

## 配置  {#config}

### 采集器配置 {#input-config}

成功安装 DataKit 并启动后，会默认开启 CPU 采集器，无需手动开启。

<!-- markdownlint-disable MD046 -->

=== "主机部署"

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

---

## 指标 {#metric}

以下所有数据采集，默认会追加名为 `host` 的全局 tag（tag 值为 DataKit 所在主机名），也可以在配置中通过 `[inputs.{{.InputName}}.tags]` 指定其它标签：

```toml
 [inputs.{{.InputName}}.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
  # ...
```

{{ range $i, $m := .Measurements }}

### `{{$m.Name}}`

{{$m.Desc}}

{{$m.MarkdownTable}}

{{ end }}
