---
title: 'DataKit 自身指标采集'
summary: '采集 DataKit 自身运行指标'
tags:
  - '主机'
__int_icon: 'icon/dk'
dashboard:
  - desc: 'DataKit 内置视图'
    path: 'dashboard/zh/dk'
  - desc: 'DataKit 拨测内置视图'
    path: 'dashboard/zh/dialtesting'

monitor:
  - desc: '暂无'
    path: '-'
---


{{.AvailableArchs}} · [:octicons-tag-24: Version-1.11.0](../datakit/changelog.md#cl-1.11.0)

---

DataKit 采集器用于自身运行指标采集，包括运行环境信息、CPU、内存占用、各个核心模块指标等。采集到的数据可用于 DataKit 内置视图、Bug Report 问题排查和运行状态归档。

## 配置 {#config}

DataKit 启动后会暴露 [Prometheus 指标](../datakit/datakit-metrics.md)。`dk` 采集器默认随 DataKit 启动，替代了之前的 `self` 采集器。

基础功能：

- 采集 DataKit 自身 CPU、内存、Goroutine、HTTP API、数据上传、Pipeline、过滤器、磁盘缓存等运行指标。
- 通过 `interval` 调整采集间隔，通过 `metric_types` 限制采集的指标类型。
- 支持通过 `[inputs.dk.tags]` 追加自定义标签。

[:octicons-tag-24: Version-2.2.0](../datakit/changelog-2026.md#cl-2.2.0) 起，`dk` 默认采集除内部屏蔽指标外的全部自身指标，不再提供 `metric_name_filter` 配置；如需仅保留部分指标，建议在 Pipeline 中按指标名过滤。

如需关闭自身指标采集，可在 `dk.conf` 中设置：

```toml
[[inputs.dk]]
  enabled = false
```

<!-- markdownlint-disable MD046 -->
=== "主机部署"

    如需调整采集间隔、指标类型、自身 Profile 采集等配置，进入 DataKit 安装目录下的 `conf.d/samples` 目录，复制 `{{.InputName}}.conf.sample` 并命名为 `{{.InputName}}.conf`。示例如下：

    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```

    配置好后，重启 DataKit 即可。

=== "Kubernetes"

    可通过 [ConfigMap 方式注入采集器配置](../datakit/datakit-daemonset-deploy.md#configmap-setting) 或 [配置 ENV_DATAKIT_INPUTS](../datakit/datakit-daemonset-deploy.md#env-setting) 开启采集器。

    也支持以环境变量的方式修改配置参数：

{{ CodeBlock .InputENVSampleZh 4 }}

<!-- markdownlint-enable MD046 -->

## 指标 {#metric}

DataKit 自身指标主要是一些 Prometheus 指标，其文档参见[这里](../datakit/datakit-metrics.md)

## 自身 Profile 采集 {#self-profiling}

[:octicons-tag-24: Version-2.2.0](../datakit/changelog-2026.md#cl-2.2.0) 起，`dk` 支持按资源阈值触发 DataKit 自身 Profile 采集。该功能默认关闭，可用于在 DataKit CPU 或内存达到阈值时采集 CPU、heap、goroutine 等 Profile。

主机部署可在 `dk.conf` 中开启：

```toml
[inputs.dk.self_profiling]
  enabled = true
```

Kubernetes 部署可通过环境变量开启：

```shell
ENV_INPUT_DK_ENABLE_SELF_PROFILING=true
```

更多阈值、采集类型、缓存和上报配置参见上方示例配置中的 `[inputs.dk.self_profiling]`。
