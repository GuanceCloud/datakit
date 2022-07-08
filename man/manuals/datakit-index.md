---
icon: zy/datakit
---

# DataKit
---


## 概述

DataKit 是一款开源、一体式的数据采集 Agent，它提供全平台操作系统（Linux/Windows/macOS）支持，拥有全面数据采集能力，涵盖主机、容器、中间件、Tracing、日志以及安全巡检等各种场景。

## 主要功能

## 主要功能点

- 支持主机、中间件、日志、APM 等领域的指标、日志以及 Tracing 几大类数据采集
- 完整支持 Kubernetes 云原生生态
- [Pipeline](pipeline.md)：简便的结构化数据提取
- 支持接入其它第三方数据采集
    - [Telegraf](../integrations/telegraf.md)
    - [Prometheus](../integrations/prom.md)
    - [Statsd](../integrations/statsd.md)
    - [Fluentd](../integrations/logstreaming.md)
    - [Filebeats](../integrations/beats_output.md)
    - [Function](../dataflux-func/write-data-via-datakit.md)
    - Tracing 相关
        - [OpenTelemetry](../integrations/opentelemetry.md)
        - [DDTrace](../integrations/ddtrace.md)
        - [Zipkin](../integrations/zipkin.md)
        - [Jaeger](../integrations/jaeger.md)
        - [Skywalking](../integrations/skywalking.md)）

## 说明

### 实验性功能 {#experimental}

DataKit 发布的时候，会带上一些实验性功能，这些功能往往是初次发布的新功能，这些功能的实现，可能会有一些欠缺考虑或不严谨的地方，甚至对于一些配置方面，我们也不保证其兼容性。对于这部分功能，请大家慎重使用。

对于实验性功能，相关问题可以提交到 issue 中：

- [Gitlab](https://gitlab.jiagouyun.com/cloudcare-tools/datakit/-/issues/new?issue%5Bmilestone_id%5D=){:target="_blank"} 
- [Github](https://github.com/GuanceCloud/datakit/issues/new){:target="_blank"}
- [极狐](https://jihulab.com/guance-cloud/datakit/-/issues/new){:target="_blank"}
