---
title     : 'OpenTelemetry'
summary   : '接收 OpenTelemetry 指标、日志、APM 数据'
__int_icon: 'icon/opentelemetry'
tags      :
  - 'OTEL'
  - '链路追踪'
dashboard :
  - desc  : 'Opentelemetry JVM 监控视图'
    path  : 'dashboard/zh/opentelemetry'
monitor   :
  - desc  : '暂无'
    path  : '-'
---


{{.AvailableArchs}}

---

OpenTelemetry （以下简称 OTEL）是 CNCF 的一个可观测性项目，旨在提供可观测性领域的标准化方案，解决观测数据的数据模型、采集、处理、导出等的标准化问题。

OTEL 是一组标准和工具的集合，旨在管理观测类数据，如 trace、metrics、logs 。

本篇旨在介绍如何在 DataKit 上配置并开启 OTEL 的数据接入，以及 Java、Go 的最佳实践。


## 配置 {#config}

<!-- markdownlint-disable MD046 -->
=== "主机安装"

    进入 DataKit 安装目录下的 `conf.d/{{.Catalog}}` 目录，复制 `{{.InputName}}.conf.sample` 并命名为 `{{.InputName}}.conf`。示例如下：

    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```

    配置好后，[重启 DataKit](../datakit/datakit-service-how-to.md#manage-service) 即可。

=== "Kubernetes"

    可通过 [ConfigMap 方式注入采集器配置](../datakit/datakit-daemonset-deploy.md#configmap-setting) 或 [配置 ENV_DATAKIT_INPUTS](../datakit/datakit-daemonset-deploy.md#env-setting) 开启采集器。

    也支持以环境变量的方式修改配置参数（需要在 ENV_DEFAULT_ENABLED_INPUTS 中加为默认采集器）：

{{ CodeBlock .InputENVSampleZh 4 }}

<!-- markdownlint-enable -->

### 注意事项 {#attentions}

1. 建议使用 gRPC 协议，gRPC 具有压缩率高、序列化快、效率更高等优点
2. 自 [DataKit 1.10.0](../datakit/changelog.md#cl-1.10.0) 版本开始，http 协议的路由是可配置的，默认请求路径（Trace/Metric）分别为 `/otel/v1/traces` `/otel/v1/logs` 以及 `/otel/v1/metrics`
3. 在涉及到 `float/double` 类型数据时，会最多保留两位小数
4. HTTP 和 gRPC 都支持 gzip 压缩格式。在 exporter 中可配置环境变量来开启：`OTEL_EXPORTER_OTLP_COMPRESSION = gzip`, 默认是不会开启 gzip。
5. HTTP 协议请求格式同时支持 JSON 和 Protobuf 两种序列化格式。但 gRPC 仅支持 Protobuf 一种。

<!-- markdownlint-disable MD046 -->
???+ warning

    - DDTrace 链路数据中的服务名是根据服务名或者引用的三方库命名的，而 OTEL 采集器的服务名是按照 `otel.service.name` 定义的
    - 为了分开显示服务名，增加了一个字段配置：`spilt_service_name = true`
    - 服务名从链路数据的标签中取出，比如 DB 类型的标签 `db.system=mysql` 那么服务名就是 mysql。如果是消息队列类型，如 `messaging.system=kafka`，那么服务名就是 `kafka`
    - 默认从这三个标签中取出：`db.system/rpc.system/messaging.system`
<!-- markdownlint-enable -->


使用 OTEL HTTP exporter 时注意环境变量的配置，由于 DataKit 的默认配置是 `/otel/v1/traces` `/otel/v1/logs` 和 `/otel/v1/metrics`，所以想要使用 HTTP 协议的话，需要单独配置 `trace` 和 `metric`，

## Agent V2 版本 {#v2}

V2 版本默认使用 `otlp exporter` 将之前的 `grpc` 改为 `http/protobuf` ， 可以通过命令 `-Dotel.exporter.otlp.protocol=grpc` 设置，或者使用默认的 `http/protobuf`

使用 http 的话，每个 exporter 路径需要显性配置 如：

```shell
java -javaagent:/usr/local/ddtrace/opentelemetry-javaagent-2.5.0.jar \
  -Dotel.exporter=otlp \
  -Dotel.exporter.otlp.protocol=http/protobuf \
  -Dotel.exporter.otlp.logs.endpoint=http://localhost:9529/otel/v1/logs \
  -Dotel.exporter.otlp.traces.endpoint=http://localhost:9529/otel/v1/traces \
  -Dotel.exporter.otlp.metrics.endpoint=http://localhost:9529/otel/v1/metrics \
  -Dotel.service.name=app \
  -jar app.jar
```

使用 gRPC 协议的话，必须是显式配置，否则就是默认的 http 协议：

```shell
java -javaagent:/usr/local/ddtrace/opentelemetry-javaagent-2.5.0.jar \
  -Dotel.exporter=otlp \
  -Dotel.exporter.otlp.protocol=grpc \
  -Dotel.exporter.otlp.endpoint=http://localhost:4317
  -Dotel.service.name=app \
  -jar app.jar
```

默认日志是开启的，要关闭日志采集的话，exporter 配置为空即可：`-Dotel.logs.exporter=none`

更多关于 V2 版本的重大修改请查看官方文档或者 GitHub <<<custom_key.brand_name>>>版本说明： [Github-v2.0.0](https://github.com/open-telemetry/opentelemetry-java-instrumentation/releases/tag/v2.0.0){:target="_blank"}

## 常规命令 {#sdk-configuration}

| ENV                           | Command                       | 说明                                       | 默认                    | 注意                                            |
|:------------------------------|:------------------------------|:-----------------------------------------|:------------------------|:----------------------------------------------|
| `OTEL_SDK_DISABLED`           | `otel.sdk.disabled`           | 关闭 SDK                                   | false                   | 关闭后将不会产生任何链路指标信息                              |
| `OTEL_RESOURCE_ATTRIBUTES`    | `otel.resource.attributes`    | "service.name=App,username=liu"          |                         | 每一个 span 中都会有该 tag 信息                         |
| `OTEL_SERVICE_NAME`           | `otel.service.name`           | 服务名，等效于上面 "service.name=App"             |                                  | 优先级高于上面                                       |
| `OTEL_LOG_LEVEL`              | `otel.log.level`              | 日志级别                                     | `info`                          |                                               |
| `OTEL_PROPAGATORS`            | `otel.propagators`            | 透传协议                                     | `tracecontext,baggage`          |                                               |
| `OTEL_TRACES_SAMPLER`         | `otel.traces.sampler`         | 采样                                       | `parentbased_always_on`         |                                               |
| `OTEL_TRACES_SAMPLER_ARG`     | `otel.traces.sampler.arg`     | 配合上面采样 参数                                | 1.0                             | 0 - 1.0                                       |
| `OTEL_EXPORTER_OTLP_PROTOCOL` | `otel.exporter.otlp.protocol` | 协议包括： `grpc`,`http/protobuf`,`http/json` | gRPC                            |                                               |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | `otel.exporter.otlp.endpoint` | OTLP 地址                                  | <http://localhost:4317>                  | <http://datakit-endpoint:9529/otel/v1/traces> |
| `OTEL_TRACES_EXPORTER`        | `otel.traces.exporter`        | 链路导出器                                    | `otlp`                                   |                                               |
| `OTEL_LOGS_EXPORTER`          | `otel.logs.exporter`          | 日志导出器                                    | `otlp`                                   | OTEL V1 版本需要显式配置，否则默认不开启                      |


> 您可以将 `otel.javaagent.debug=true` 参数传递给 Agent 以查看调试日志。请注意，这些日志内容相当冗长，生产环境下谨慎使用。

## 链路 {#tracing}

Trace（链路）是由多个 span 组成的一条链路信息。
无论是单个服务还是一个服务集群，链路信息提供了一个请求发生到结束所经过的所有服务之间完整路径的集合。

DataKit 只接收 OTLP 的数据，OTLP 有三种数据类型： `gRPC` ， `http/protobuf` 和 `http/json` ，具体配置可以参考：

```shell
# OpenTelemetry 默认采用 gPRC 协议发送到 DataKit
-Dotel.exporter=otlp \
-Dotel.exporter.otlp.protocol=grpc \
-Dotel.exporter.otlp.endpoint=http://datakit-endpoint:4317

# 使用 http/protobuf 方式
-Dotel.exporter=otlp \
-Dotel.exporter.otlp.protocol=http/protobuf \
-Dotel.exporter.otlp.traces.endpoint=http://datakit-endpoint:9529/otel/v1/traces \
-Dotel.exporter.otlp.metrics.endpoint=http://datakit-endpoint:9529/otel/v1/metrics 

# 使用 http/json 方式
-Dotel.exporter=otlp \
-Dotel.exporter.otlp.protocol=http/json \
-Dotel.exporter.otlp.traces.endpoint=http://datakit-endpoint:9529/otel/v1/traces \
-Dotel.exporter.otlp.metrics.endpoint=http://datakit-endpoint:9529/otel/v1/metrics
```

### 链路采样 {#sample}

可以采用头部采样或者尾部采样，具体可以查看两篇最佳实践：

- 需要配合 collector 的尾部采样： [OpenTelemetry 采样最佳实践](../best-practices/cloud-native/opentelemetry-simpling.md)
- Agent 端的头部采样： [OpenTelemetry Java Agent 端采样策略](../best-practices/cloud-native/otel-agent-sampling.md)

### Tag {#tags}

从 DataKit 版本 [1.22.0](../datakit/changelog.md#cl-1.22.0) 开始，黑名单功能废弃。增加固定标签列表，只有在此列表中的才会提取到一级标签中，以下是固定列表：

| Attributes            | tag                   | 说明                             |
|:----------------------|:----------------------|:-------------------------------|
| http.url              | http_url              | HTTP 请求完整路径                    |
| http.hostname         | http_hostname         | hostname                       |
| http.route            | http_route            | 路由                             |
| http.status_code      | http_status_code      | 状态码                            |
| http.request.method   | http_request_method   | 请求方法                           |
| http.method           | http_method           | 同上                             |
| http.client_ip        | http_client_ip        | 客户端 IP                         |
| http.scheme           | http_scheme           | 请求协议                           |
| url.full              | url_full              | 请求全路径                          |
| url.scheme            | url_scheme            | 请求协议                           |
| url.path              | url_path              | 请求路径                           |
| url.query             | url_query             | 请求参数                           |
| span_kind             | span_kind             | span 类型                        |
| db.system             | db_system             | span 类型                        |
| db.operation          | db_operation          | DB 动作                          |
| db.name               | db_name               | 数据库名称                          |
| db.statement          | db_statement          | 详细信息                           |
| server.address        | server_address        | 服务地址                           |
| net.host.name         | net_host_name         | 请求的 host                       |
| server.port           | server_port           | 服务端口号                          |
| net.host.port         | net_host_port         | 同上                             |
| network.peer.address  | network_peer_address  | 网络地址                           |
| network.peer.port     | network_peer_port     | 网络端口                           |
| network.transport     | network_transport     | 协议                             |
| messaging.system      | messaging_system      | 消息队列名称                         |
| messaging.operation   | messaging_operation   | 消息动作                           |
| messaging.message     | messaging_message     | 消息                             |
| messaging.destination | messaging_destination | 消息详情                           |
| rpc.service           | rpc_service           | RPC 服务地址                       |
| rpc.system            | rpc_system            | RPC 服务名称                       |
| error                 | error                 | 是否错误                           |
| error.message         | error_message         | 错误信息                           |
| error.stack           | error_stack           | 堆栈信息                           |
| error.type            | error_type            | 错误类型                           |
| error.msg             | error_message         | 错误信息                           |
| project               | project               | project                        |
| version               | version               | 版本                             |
| env                   | env                   | 环境                             |
| host                  | host                  | Attributes 中的 host 标签          |
| pod_name              | pod_name              | Attributes 中的 pod_name 标签      |
| pod_namespace         | pod_namespace         | Attributes 中的 pod_namespace 标签 |

如果想要增加自定义标签，可使用环境变量：

```shell
# 通过启动参数添加自定义标签
-Dotel.resource.attributes=username=myName,env=1.1.0
```

并修改配置文件中的白名单，这样就可以在<<<custom_key.brand_name>>>的链路详情的一级标签出现自定义的标签。

```toml
customer_tags = ["sink_project", "username","env"]
```

### Kind {#kind}

所有的 `Span` 都有 `span_kind` 标签，共有 6 中属性：

- `unspecified`:  未设置。
- `internal`:  内部 span 或子 span 类型。
- `server`:  WEB 服务、RPC 服务 等等。
- `client`:  客户端类型。
- `producer`:  消息的生产者。
- `consumer`:  消息的消费者。


## 指标 {#metric}

OpenTelemetry Java Agent 从应用程序中通过 JMX 协议获取 MBean 的指标信息，Java Agent 通过内部 SDK 报告选定的 JMX 指标，这意味着所有的指标都是可以配置的。

可以通过命令 `otel.jmx.enabled=true/false` 开启和关闭 JMX 指标采集，默认是开启的。

为了控制 MBean 检测尝试之间的时间间隔，可以使用 `otel.jmx.discovery.delay` 命令，该属性定义了在第一个和下一个检测周期之间通过的毫秒数。

另外 Agent 内置的一些三方软件的采集配置。具体可以参考： [GitHub OTEL JMX Metric](https://github.com/open-telemetry/opentelemetry-java-instrumentation/blob/main/instrumentation/jmx-metrics/javaagent/README.md){:target="_blank"}

<!-- markdownlint-disable MD046 -->
???+ warning

    从版本 [DataKit 1.68.0](../datakit/changelog-2025.md#cl-1.68.0) 开始指标集名称做了改动：
    所有发送到<<<custom_key.brand_name>>>的指标有一个统一的指标集的名字： `otel_service` 
    如果已经有了仪表板，将已有的仪表板导出后统一将 `otel-serivce` 改为 `otel_service` 再导入即可。

<!-- markdownlint-enable -->

在将 **Histogram** 指标转到<<<custom_key.brand_name>>>的时候有些指标做了特殊处理：

- OpenTelemetry 的直方图桶会被直接映射到 Prometheus 的直方图桶。
- 每个桶的计数会被转换为 Prometheus 的累积计数格式。
- 例如，OpenTelemetry 的桶 `[0, 10)`、`[10, 50)`、`[50, 100)` 会被转换为 Prometheus 的 `_bucket` 指标，并附带 `le` 标签：

```text
  my_histogram_bucket{le="10"} 100
  my_histogram_bucket{le="50"} 200
  my_histogram_bucket{le="100"} 250
```

- OpenTelemetry 直方图的总观测值数量会被转换为 Prometheus 的 `_count` 指标。
- OpenTelemetry 直方图的总和会被转换为 Prometheus 的 `_sum` 指标，还会添加 `_max` `_min`。

```text
  my_histogram_count 250
  my_histogram_max 100
  my_histogram_min 50
  my_histogram_sum 12345.67
```

凡是以 `_bucket` 结尾的指标都是直方图数据，并且一定有 `_max` `_min` `_count` `sum` 结尾的指标。

在直方图数据中可以使用 `le(less or equal)` 标签进行分类，并且可以根据标签进行筛选，可以查看 [OpenTelemetry Metrics](https://opentelemetry.io/docs/specs/semconv/){:target="_blank"} 所有的指标和标签。

这种转换使得 OpenTelemetry 收集的直方图数据能够无缝集成到 Prometheus 中，并利用 Prometheus 的强大查询和可视化功能进行分析。

## 数据字段说明 {#fields}

{{ range $i, $m := .Measurements }}

### `{{$m.Name}}`

{{$m.Desc}}

- Tags

{{$m.TagsMarkdownTable}}

- Metrics

{{$m.FieldsMarkdownTable}}

{{ end }}


## 指标中删除的标签 {#del-metric}

OTEL 上报的指标中有很多无用的标签，这些都是 String 类型，由于太占用内存和带宽就做了删除，不会上传到<<<custom_key.brand_name>>>。

这些标签包括：

```text
process.command_line
process.executable.path
process.runtime.description
process.runtime.name
process.runtime.version
telemetry.distro.name
telemetry.distro.version
telemetry.sdk.language
telemetry.sdk.name
telemetry.sdk.version
```

## 日志 {#logging}

[:octicons-tag-24: Version-1.33.0](../datakit/changelog.md#cl-1.33.0)

目前 JAVA Agent 支持采集 `stdout` 日志。并使用 [Standard output](https://opentelemetry.io/docs/specs/otel/logs/sdk_exporters/stdout/){:target="_blank"} 方式通过 `otlp` 协议发送到 DataKit 中。

`OTEL Agent` 默认情况下**不开启** log 采集，必须需要通过显式命令： `otel.logs.exporter` 开启方式为：

```shell
# env
export OTEL_LOGS_EXPORTER=OTLP
export OTEL_EXPORTER_OTLP.ENDPOINT=http://<DataKit Addr>:4317
# other env
java -jar app.jar

# command
java -javaagent:/path/to/agnet.jar \
  -otel.logs.exporter=otlp \
  -Dotel.exporter.otlp.endpoint=http://<DataKit Addr>:4317 \
  -jar app.jar
```

默认情况下，日志内容的最大长度为 500KB ，超过的部分会分成多条日志。日志的标签最大长度为 32KB ，该字段不可配置，超过的部分会切割掉。

通过 OTEL 采集的日志的 `source` 为服务名，也可以通过添加标签的方式自定义：`log.source` ，比如：`-Dotel.resource.attributes="log.source=source_name"`。

> 注意：如果 app 是运行在容器环境（比如 k8s），DataKit 本来就会[自动采集日志](container-log.md#logging-stdout){:target="_blank"}（默认行为），如果再采集一次，会有重复采集的问题。建议在开启采集日志之前，[手动关闭 DataKit 自主的日志采集行为](container-log.md#logging-with-image-config){:target="_blank"}

更多语言可以[查看官方文档](https://opentelemetry.io/docs/specs/otel/logs/){:target="_blank"}

## 示例 {#examples}

DataKit 目前提供了如下两种语言的最佳实践：

- [Golang](opentelemetry-go.md)
- [Java](opentelemetry-java.md)


## 更多文档 {#more-readings}

- [Golang SDK](https://github.com/open-telemetry/opentelemetry-go){:target="_blank"}
- [官方使用手册](https://opentelemetry.io/docs/){:target="_blank"}
- [环境变量配置](https://github.com/open-telemetry/opentelemetry-java/blob/main/sdk-extensions/autoconfigure/README.md#otlp-exporter-both-span-and-metric-exporters){:target="_blank"}
- [DDTrace 与 OpenTelemetry 串联时采样策略注意事项](tracing-sample.md)
