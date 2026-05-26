---
title     : '数据采集可用性检测'
summary   : '通过自动生成和查询探测数据来检测数据采集的可用性和延迟'
tags:
  - '可用性'
  - '延迟'
  - '监控'
__int_icon      : 'icon/ingestion_canary'
dashboard :
  - desc  : '暂无'
    path  : '-'
monitor   :
  - desc  : '暂无'
    path  : '-'
---

{{.AvailableArchs}}

---

该采集器用于检测数据采集的可用性和延迟。它会自动生成探测数据（指标、日志、链路追踪），然后通过 DQL 查询来验证数据是否成功采集，并测量从数据发送到可查询的延迟时间。

## 前置条件 {#requirements}

- 需要配置 DataWay，用于数据上报和 DQL 查询
- 如果配置了 `result_workspace`，需要确保该工作空间 URL 可访问

## 配置 {#config}

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

## 指标说明 {#measurements}

该采集器会生成两类数据：

1. **测试数据**：用于测试数据采集可用性的探测数据点（指标、日志、链路追踪）
2. **结果指标**：测试结果的指标数据，包含延迟时间和测试状态

探测数据点不携带全局标签，只包含数据点自身的字段和标签，以及配置中指定的 `tags`。

### 测试数据 {#probe-data}

{{ range $i, $m := .Measurements }}
{{ if and (ne $m.Name "ingestion_canary_result") (eq (printf "%v" $m.Cat) "metric") }}
#### `{{ $m.Name }}` (指标)

{{ if $m.Desc }}{{ $m.Desc }}{{ end }}

{{ $m.MarkdownTable }}

{{ end }}
{{ end }}

{{ range $i, $m := .Measurements }}
{{ if and (ne $m.Name "ingestion_canary_result") (eq (printf "%v" $m.Cat) "logging") }}
#### `{{ $m.Name }}` (日志)

{{ if $m.Desc }}{{ $m.Desc }}{{ end }}

{{ $m.MarkdownTable }}

{{ end }}
{{ end }}

{{ range $i, $m := .Measurements }}
{{ if and (ne $m.Name "ingestion_canary_result") (eq (printf "%v" $m.Cat) "tracing") }}
#### `{{ $m.Name }}` (链路追踪)

{{ if $m.Desc }}{{ $m.Desc }}{{ end }}

{{ $m.MarkdownTable }}

{{ end }}
{{ end }}

### 结果指标 {#result-metric}

{{ range $i, $m := .Measurements }}
{{ if eq $m.Name "ingestion_canary_result" }}
#### `{{ $m.Name }}`

{{ if $m.Desc }}{{ $m.Desc }}{{ end }}

{{ $m.MarkdownTable }}

{{ end }}
{{ end }}

## 命令行工具 {#cli-tool}

除了采集器模式，还提供了命令行工具用于一次性测试：

```bash
# 使用默认配置
datakit tool --ingestion-canary

# 指定日志数据的 storage index
datakit tool --ingestion-canary --ingestion-canary-index my_index
```

**参数说明：**

- `--ingestion-canary`: 启用 ingestion canary 测试工具
- `--ingestion-canary-index`: 指定日志数据的 storage index，默认为 "default"（仅对日志数据有效）

**功能说明：**

该工具会生成一轮探测数据（指标、日志、链路追踪），发送到 DataWay 后持续查询直到找到数据或用户中断，并输出各数据类型的延迟时间。工具会持续运行，每轮测试间隔 10 秒。
