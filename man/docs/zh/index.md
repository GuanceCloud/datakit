---
icon: zy/datakit
---

# DataKit
---


## 概述 {#intro}

DataKit 是一款开源、一体式的数据采集 Agent，它提供全平台操作系统（Linux/Windows/macOS）支持，拥有全面数据采集能力，涵盖主机、容器、中间件、Tracing、日志以及安全巡检等各种场景。

## 主要功能 {#features}

- 支持主机、中间件、日志、APM 等领域的指标、日志以及 Tracing 几大类数据采集
- 完整支持 Kubernetes 云原生生态
- [Pipeline](../developers/pipeline.md)：简便的结构化数据提取
- 支持接入其它第三方数据采集
    - [Telegraf](telegraf.md)
    - [Prometheus](prom.md)
    - [Statsd](statsd.md)
    - [Fluentd](logstreaming.md)
    - [Filebeats](beats_output.md)
    - [Function](../dataflux-func/write-data-via-datakit.md)
    - Tracing 相关
        - [OpenTelemetry](opentelemetry.md)
        - [DDTrace](ddtrace.md)
        - [Zipkin](zipkin.md)
        - [Jaeger](jaeger.md)
        - [Skywalking](skywalking.md)

## 说明 {#spec}

### 实验性功能 {#experimental}

DataKit 发布的时候，会带上一些实验性功能，这些功能往往是初次发布的新功能，这些功能的实现，可能会有一些欠缺考虑或不严谨的地方，故使用实验性功能的时候，需考虑如下一些可能的情况：

- 功能不太稳定
- 于一些功能配置，在后续的迭代过程中，不保证其兼容性
- 由于其局限性，功能可能会被移除，但会有对应的其它措施来满足对应的需求

对于这部分功能，请大家慎重使用。

在使用实验性功能的过程中，相关问题可以提交到 issue 中：

- [Gitlab](https://gitlab.jiagouyun.com/cloudcare-tools/datakit/-/issues/new?issue%5Bmilestone_id%5D=){:target="_blank"} 
- [Github](https://github.com/GuanceCloud/datakit/issues/new){:target="_blank"}
- [极狐](https://jihulab.com/guance-cloud/datakit/-/issues/new){:target="_blank"}

### 图例说明 {#legends}

| 图例                                                                                                                       | 说明                                                          |
| ---                                                                                                                        | ---                                                           |
| :fontawesome-solid-flag-checkered:                                                                                         | 表示该采集器支持选举                                          |
| :fontawesome-brands-linux: :fontawesome-brands-windows: :fontawesome-brands-apple: :material-kubernetes: :material-docker: | 例分别用来表示 Linux、Windows、macOS、 Kubernetes 以及 Docker |
