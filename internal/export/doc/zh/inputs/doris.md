---
title     : 'Doris'
summary   : '采集 Doris 的指标数据'
__int_icon      : 'icon/doris'
dashboard :
  - desc  : 'Doris'
    path  : 'dashboard/zh/doris'
monitor   :
  - desc  : 'Doris'
    path  : 'monitor/zh/doris'
---

<!-- markdownlint-disable MD025 -->
# Doris
<!-- markdownlint-enable -->
---

{{.AvailableArchs}}

---

Doris 采集器用于采集 Doris 相关的指标数据，目前只支持 Prometheus 格式的数据

## 配置 {#config}

已测试的版本：

- [x] 2.0.0

### 前置条件 {#requirements}

Doris 默认开启 Prometheus 端口

验证前端：curl ip:8030/metrics

验证后端：curl ip:8040/metrics

### 采集器配置 {input-config}

<!-- markdownlint-disable MD046 -->
=== "主机安装"

    进入 DataKit 安装目录下的 `conf.d/{{.Catalog}}` 目录，复制 `{{.InputName}}.conf.sample` 并命名为 `{{.InputName}}.conf`。示例如下：
    
    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```

    配置好后，[重启 DataKit](../datakit/datakit-service-how-to.md#manage-service) 即可。

=== "Kubernetes"

    可通过 [ConfigMap 方式注入采集器配置](../datakit/datakit-daemonset-deploy.md#configmap-setting) 或 [配置 ENV_DATAKIT_INPUTS](../datakit/datakit-daemonset-deploy.md#env-setting) 开启采集器。

<!-- markdownlint-enable -->

## 指标 {#metric}

{{ range $i, $m := .Measurements }}

### `{{$m.Name}}`

- 标签

{{$m.TagsMarkdownTable}}

- 指标列表

{{$m.FieldsMarkdownTable}}

{{ end }}
