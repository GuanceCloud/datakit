{{.CSS}}

- DataKit 版本：{{.Version}}
- 文档发布日期：{{.ReleaseDate}}
- 操作系统支持：`{{.AvailableArchs}}`

# Flink

Flink 采集器可以从 Flink 实例中采取很多指标，比如 Flink 服务器状态和网络的状态等多种指标，并将指标采集到 DataFlux ，帮助你监控分析 Flink 各种异常情况。

## 前置条件

### pushgateway

可以在 [Prometheuse 官方页面](https://prometheus.io/download/#pushgateway) 进行下载。

启动 pushgateway: `nohup ./pushgateway &`

### 配置 Flink 对接 pushgateway

下面以版本 1.14 为例。

摘自官方文档: [PrometheusPushGateway](https://nightlies.apache.org/flink/flink-docs-release-1.14/docs/deployment/metric_reporters/#prometheuspushgateway)

#### 放置库文件

将 Flink 安装目录里面的 plugins 文件下的 `metrics-prometheus` 文件夹下面的 `flink-metrics-prometheus-1.14.0.jar` 拷贝到 Flink 的 lib 目录下。

```shell
cp plugins/metrics-prometheus/flink-metrics-prometheus-1.14.0.jar lib/
```

#### 配置 `flink-conf.yaml`

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

启动 Flink。

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
