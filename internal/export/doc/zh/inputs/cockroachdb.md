---
title     : 'CockroachDB'
summary   : '采集 CockroachDB 的指标数据'
__int_icon      : 'icon/cockroachdb'
tags:
  - '数据库'
dashboard :
  - desc  : 'CockroachDB'
    path  : 'dashboard/zh/cockroachdb'
monitor   :
  - desc  : 'CockroachDB'
    path  : 'monitor/zh/cockroachdb'
---


{{.AvailableArchs}}

---

CockroachDB 采集器用于采集 CockroachDB 相关的指标数据，目前只支持 Prometheus 格式的数据

已测试的版本：

- [x] CockroachDB 19.2
- [x] CockroachDB 20.2
- [x] CockroachDB 21.2
- [x] CockroachDB 22.2
- [x] CockroachDB 23.2.4

## 配置 {#config}

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

<!-- markdownlint-enable -->

## 指标 {#metric}

{{ range $i, $m := .Measurements }}

### `{{$m.Name}}`

{{$m.MarkdownTable}}

{{ end }}
