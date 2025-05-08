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
- [Pipeline](../pipeline/use-pipeline/index.md)：简便的结构化数据提取
- 支持接入其它第三方数据采集
    - [Telegraf](../integrations/telegraf.md)
    - [Prometheus](../integrations/prom.md)
    - [Statsd](../integrations/statsd.md)
    - [Fluentd](../integrations/logstreaming.md)
    - [Filebeat](../integrations/beats_output.md)
    - [Function](https://func.<<<custom_key.brand_main_domain>>>/doc/practice-write-data-via-datakit/){:target="_blank"}
    - Tracing 相关
        - [OpenTelemetry](../integrations/opentelemetry.md)
        - [DDTrace](../integrations/ddtrace.md)
        - [Zipkin](../integrations/zipkin.md)
        - [Jaeger](../integrations/jaeger.md)
        - [SkyWalking](../integrations/skywalking.md)
        - [Pinpoint](../integrations/pinpoint.md)

## 说明 {#spec}

### 实验性功能 {#experimental}

DataKit 发布的时候，会带上一些实验性功能，这些功能往往是初次发布的新功能，这些功能的实现，可能会有一些欠缺考虑或不严谨的地方，故使用实验性功能的时候，需考虑如下一些可能的情况：

- 功能不太稳定
- 于一些功能配置，在后续的迭代过程中，不保证其兼容性
- 由于其局限性，功能可能会被移除，但会有对应的其它措施来满足对应的需求

对于这部分功能，请大家慎重使用。

在使用实验性功能的过程中，相关问题可以提交到 issue 中：

- [GitLab](https://gitlab.jiagouyun.com/cloudcare-tools/datakit/-/issues/new?issue%5Bmilestone_id%5D=){:target="_blank"}
- [GitHub](https://github.com/GuanceCloud/datakit/issues/new){:target="_blank"}
- [极狐](https://jihulab.com/guance-cloud/datakit/-/issues/new){:target="_blank"}

### 图例说明 {#legends}

| 图例                                                                                                                       | 说明                                                            |
| ---                                                                                                                        | ---                                                             |
| :fontawesome-solid-flag-checkered:                                                                                         | 表示该采集器支持选举                                            |
| :fontawesome-brands-linux: :fontawesome-brands-windows: :fontawesome-brands-apple: :material-kubernetes: :material-docker: | 例分别用来表示 Linux、Windows、macOS、 Kubernetes 以及 Docker   |
| :octicons-beaker-24:                                                                                                       | 表示实验性功能（参见[实验性功能的描述](index.md#experimental)） |

## 注意事项 {#disclaimer}

在使用 Datakit 过程中，对已有的系统可能会有如下一些影响：

1. 日志采集会导致的磁盘高速读取，日志量越大，读取的 iops 越高
1. 如果在 Web/App 应用中加入了 RUM SDK，那么会有持续的 RUM 相关的数据上传，如果上传的带宽有相关限制，可能会导致 Web/App 的页面卡顿
1. eBPF 开启后，由于采集的数据量比较大，会占用一定量的内存和 CPU。其中 bpf-netlog 开启后，会根据主机和容器网卡的所有 TCP 数据包，产生大量的日志
1. 在 Datakit 繁忙的时候（接入了大量的日志/Trace 以及外部数据导入等），其会占用相当量的 CPU 和内存资源，建议设置合理的 cgroup 来加以控制
1. 当 Datakit 部署在 Kubernetes 中时，对 API server 会有一定的请求压力
