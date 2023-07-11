---
title     : 'Flink'
summary   : '采集 Flink 的指标数据'
icon      : 'icon/flink'
dashboard :
  - desc  : 'Flink'
    path  : 'dashboard/zh/flink'
monitor   :
  - desc  : 'Flink'
    path  : 'monitor/zh/flink'
---

<!-- markdownlint-disable MD025 -->
# Flink
<!-- markdownlint-enable -->

---

{{.AvailableArchs}}

---

Flink 采集器可以从 Flink 实例中采取很多指标，比如 Flink 服务器状态和网络的状态等多种指标，并将指标采集到观测云 ，帮助你监控分析 Flink 各种异常情况。

## 配置 {#config}

### 前置条件 {#requirements}

> 说明：示例 Flink 版本为 Flink 1.14.2 (CentOS)，各个不同版本指标可能存在差异。

目前 Flink 官方提供两种 metrics 上报方式：[Prometheus](https://nightlies.apache.org/flink/flink-docs-release-1.14/docs/deployment/metric_reporters/#prometheus){:target="_blank"} 和 [Prometheus PushGateway](https://nightlies.apache.org/flink/flink-docs-release-1.14/docs/deployment/metric_reporters/#prometheuspushgateway){:target="_blank"}。它们主要的区别是：

- Prometheus PushGateway 方式是把集群所有的 metrics 统一汇报给 PushGateway，所以需要额外安装 PushGateway。
- Prometheus 方式需要集群每个节点暴露一个唯一端口，不需要额外安装其它软件，但需要 N 个可用端口，配置略微复杂。

### PrometheusPushGateway 方式（推荐） {#push-gateway}

- 下载和安装：PushGateWay 可以在 [Prometheus 官方页面](https://prometheus.io/download/#pushgateway){:target="_blank"} 进行下载。

启动 Push Gateway：（此命令仅供参考，具体命令根据实际环境可能有所不同）

```shell
nohup ./pushgateway &
```

- 配置 `flink-conf.yaml` 把 metrics 统一汇报给 PushGateway

配置 Flink 的配置文件 `conf/flink-conf.yaml` 示例：

```bash
metrics.reporter.promgateway.class: org.apache.flink.metrics.prometheus.PrometheusPushGatewayReporter # 固定这个值，不能改
metrics.reporter.promgateway.host: localhost # promgateway 的 IP 地址
metrics.reporter.promgateway.port: 9091 # promgateway 的监听端口
metrics.reporter.promgateway.interval: 15 SECONDS # 采集间隔
metrics.reporter.promgateway.groupingKey: k1=v1;k2=v2

# 以下是可选参数
# metrics.reporter.promgateway.jobName: myJob
# metrics.reporter.promgateway.randomJobNameSuffix: true
# metrics.reporter.promgateway.deleteOnShutdown: false
```

启动 Flink：`./bin/start-cluster.sh`（此命令仅供参考，具体命令根据实际环境可能有所不同）

### Prometheus 方式 {#prometheus}

- 配置 `flink-conf.yaml` 暴露各个节点的 metrics。配置 Flink 的配置文件 `conf/flink-conf.yaml` 示例：

```bash
metrics.reporter.prom.class: org.apache.flink.metrics.prometheus.PrometheusReporter
metrics.reporter.prom.port: 9250-9260 # 各个节点的端口区间，根据节点数量有所不同，一个端口对应一个节点
```

- 启动 Flink: `./bin/start-cluster.sh`（此命令仅供参考，具体命令根据实际环境可能有所不同）
- 可以访问外网的主机<[安装 Datakit](https://www.yuque.com/dataflux/datakit/datakit-install){:target="_blank"}>
- 更改 Flink 配置添加如下内容，开启 Prometheus 采集

```bash
metrics.reporter.prom.class: org.apache.flink.metrics.prometheus.PrometheusReporter
metrics.reporter.prom.port: 9250-9260
```

> 注意：`metrics.reporter.prom.port` 设置请参考集群 `jobmanager` 和 `taskmanager` 数量而定

- 重启 Flink 集群应用配置
- `curl http://{Flink iP}:9250-9260` 返回结果正常即可开始采集

## 指标 {#metric}

默认情况下，Flink 会收集多个指标，这些[指标](https://nightlies.apache.org/flink/flink-docs-release-1.14/docs/ops/metrics/#system-metrics){:target="_blank"}可提供对当前状态的深入洞察。

{{ range $i, $m := .Measurements }}

### `{{$m.Name}}`

- 标签

{{$m.TagsMarkdownTable}}

- 指标列表

{{$m.FieldsMarkdownTable}}

{{ end }}
