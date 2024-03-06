---
title: 'DataKit 自身指标采集'
summary: '采集 Datakit 自身运行指标'
__int_icon: 'icon/dk'
dashboard:
  - desc: 'Datakit 内置视图'
    path: 'dashboard/zh/dk'
  - desc: 'Datakit 拨测内置视图'
    path: 'dashboard/zh/dialtesting'

monitor:
  - desc: '暂无'
    path: '-'
---

<!-- markdownlint-disable MD025 -->
# DataKit 自身指标
<!-- markdownlint-enable -->

---

{{.AvailableArchs}} · [:octicons-tag-24: Version-1.11.0](../datakit/changelog.md#cl-1.11.0)

---

Datakit 采集器用于自身基本信息的采集，包括运行环境信息、CPU、内存占用、各个核心模块指标等。

## 配置 {#config}

Datakit 启动后。默认会暴露一些 [Prometheus 指标](../datakit/datakit-metrics.md)，没有额外的操作需要执行，本采集器也是默认启动的，替代了之前的 `self` 采集器。

<!-- markdownlint-disable MD046 -->
=== "主机部署"

    进入 DataKit 安装目录下的 `conf.d/{{.Catalog}}` 目录，复制 `{{.InputName}}.conf.sample` 并命名为 `{{.InputName}}.conf`。示例如下：

    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```

    配置好后，重启 DataKit 即可。

=== "Kubernetes"

    可通过 [ConfigMap 方式注入采集器配置](../datakit/datakit-daemonset-deploy.md#configmap-setting) 或 [配置 ENV_DATAKIT_INPUTS](../datakit/datakit-daemonset-deploy.md#env-setting) 开启采集器。

    也支持以环境变量的方式修改配置参数（需要在 ENV_DEFAULT_ENABLED_INPUTS 中加为默认采集器）：

{{ CodeBlock .InputENVSampleZh 4 }}

<!-- markdownlint-enable -->

## 指标 {#metric}

Datakit 自身指标主要是一些 Prometheus 指标，其文档参见[这里](../datakit/datakit-metrics.md)
