{{.CSS}}

- DataKit 版本：{{.Version}}
- 文档发布日期：{{.ReleaseDate}}
- 操作系统支持：`{{.AvailableArchs}}`

# Flink

Flink 采集器可以从 Flink 实例中采取很多指标，比如 Flink 服务器状态和网络的状态等多种指标，并将指标采集到 DataFlux ，帮助你监控分析 Flink 各种异常情况。

下面以版本 1.14 为例。

## 前置条件

目前 Flink 官方提供两种 metrics 上报方式: [Prometheus](https://nightlies.apache.org/flink/flink-docs-release-1.14/docs/deployment/metric_reporters/#prometheus) 和 [PrometheusPushGateway](https://nightlies.apache.org/flink/flink-docs-release-1.14/docs/deployment/metric_reporters/#prometheuspushgateway)。

它们主要的区别是:
- PrometheusPushGateway 方式是把集群所有的 metrics 统一汇报给 pushgateway，所以需要额外安装 pushgateway
- Prometheus 方式需要集群每个节点暴露一个唯一端口，不需要额外安装其它软件，但需要 N 个可用端口，配置略微复杂

### PrometheusPushGateway 方式（推荐）

#### pushgateway 下载与启动

可以在 [Prometheuse 官方页面](https://prometheus.io/download/#pushgateway) 进行下载。

启动 pushgateway: `nohup ./pushgateway &`（此命令仅供参考，具体命令根据实际环境可能有所不同）

#### 配置 `flink-conf.yaml` 把 metrics 统一汇报给 pushgateway

配置 Flink 的配置文件 `conf/flink-conf.yaml` 示例:

```yml
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

启动 Flink: `./bin/start-cluster.sh`（此命令仅供参考，具体命令根据实际环境可能有所不同）

### Prometheus 方式

#### 配置 `flink-conf.yaml` 暴露各个节点的 metrics

配置 Flink 的配置文件 `conf/flink-conf.yaml` 示例:

```yml
metrics.reporter.prom.class: org.apache.flink.metrics.prometheus.PrometheusReporter
metrics.reporter.prom.port: 9250-9260 # 各个节点的端口区间，根据节点数量有所不同，一个端口对应一个节点
```

启动 Flink: `./bin/start-cluster.sh`（此命令仅供参考，具体命令根据实际环境可能有所不同）

## 配置

进入 DataKit 安装目录下的 `conf.d/flink` 目录，复制如下示例并命名为 `flink.conf`。示例如下：

```toml
{{.InputSample}} 
```

配置好后，重启 DataKit 即可。

## 指标集

{{ range $i, $m := .Measurements }}

### `{{$m.Name}}`

-  标签

{{$m.TagsMarkdownTable}}

- 指标列表

{{$m.FieldsMarkdownTable}}

{{ end }}
