---
title     : 'Prometheus Exporter (New)'
summary   : '采集 Prometheus Exporter 暴露的指标数据（v2）'
tags:
  - '外部数据接入'
  - 'PROMETHEUS'
__int_icon: 'icon/prometheus'
dashboard :
  - desc  : '暂无'
    path  : '-'
monitor   :
  - desc  : '暂无'
    path  : '-'
---

{{.AvailableArchs}}

---

PromV2 采集器是 Prom 采集器的升级版，简化了配置方式，提高了采集性能。

## 配置 {#config}

<!-- markdownlint-disable MD046 -->
???+ attention

    PromV2 缺少很多数据修改的配置项，只能通过 Pipeline 来对采集的数据进行调整。
<!-- markdownlint-enable -->

<!-- markdownlint-disable MD046 -->

=== "主机安装"

    进入 DataKit 安装目录下的 *conf.d/samples* 目录，复制 *{{.InputName}}.conf.sample* 并命名为 *{{.InputName}}.conf*。示例如下：
    
    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```
    
    配置好后，[重启 DataKit](../datakit/datakit-service-how-to.md#manage-service) 即可。

=== "Kubernetes"

    目前可以通过 [ConfigMap 方式注入采集器配置](../datakit/datakit-daemonset-deploy.md#configmap-setting)来开启采集器。

<!-- markdownlint-enable -->
