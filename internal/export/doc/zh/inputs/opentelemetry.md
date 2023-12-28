---
title     : 'OpenTelemetry'
summary   : '接收 OpenTelemetry 指标、日志、APM 数据'
__int_icon      : 'icon/opentelemetry'
dashboard :
  - desc  : 'Opentelemetry JVM 监控视图'
    path  : 'dashboard/zh/opentelemetry'
monitor   :
  - desc  : '暂无'
    path  : '-'
---

<!-- markdownlint-disable MD025 -->
# OpenTelemetry
<!-- markdownlint-enable -->

---

{{.AvailableArchs}}

---

OpenTelemetry （以下简称 OTEL）是 CNCF 的一个可观测性项目，旨在提供可观测性领域的标准化方案，解决观测数据的数据模型、采集、处理、导出等的标准化问题。

OTEL 是一组标准和工具的集合，旨在管理观测类数据，如 trace、metrics、logs 等（未来可能有新的观测类数据类型出现）。

OTEL 提供与 vendor 无关的实现，根据用户的需要将观测类数据导出到不同的后端，如开源的 Prometheus、Jaeger、Datakit 或云厂商的服务中。

本篇旨在介绍如何在 Datakit 上配置并开启 OTEL 的数据接入，以及 Java、Go 的最佳实践。

> 版本说明：Datakit 目前只接入 OTEL-v1 版本的 `otlp` 数据。

## 配置 {#config}

<!-- markdownlint-disable MD046 -->
=== "主机安装"

    进入 DataKit 安装目录下的 `conf.d/{{.Catalog}}` 目录，复制 `{{.InputName}}.conf.sample` 并命名为 `{{.InputName}}.conf`。示例如下：

    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```

    配置好后，[重启 DataKit](../datakit/datakit-service-how-to.md#manage-service) 即可。

=== "Kubernetes"

    目前可以通过 [ConfigMap 方式注入采集器配置](../datakit/datakit-daemonset-deploy.md#configmap-setting)来开启采集器。

    在 Kubernetes 中支持的环境变量如下表：

    | 环境变量名                          | 类型        | 示例                                                                                                     |
    | ----------------------------------- | ----------- | -------------------------------------------------------------------------------------------------------- |
    | `ENV_INPUT_OTEL_CUSTOMER_TAGS`      | JSON string | `["sink_project", "custom.tag"]`                                                                         |
    | `ENV_INPUT_OTEL_KEEP_RARE_RESOURCE` | bool        | true                                                                                                     |
    | `ENV_INPUT_OTEL_DEL_MESSAGE`        | bool        | true                                                                                                     |
    | `ENV_INPUT_OTEL_OMIT_ERR_STATUS`    | JSON string | `["404", "403", "400"]`                                                                                  |
    | `ENV_INPUT_OTEL_CLOSE_RESOURCE`     | JSON string | `{"service1":["resource1"], "service2":["resource2"], "service3":["resource3"]}`                         |
    | `ENV_INPUT_OTEL_SAMPLER`            | float       | 0.3                                                                                                      |
    | `ENV_INPUT_OTEL_TAGS`               | JSON string | `{"k1":"v1", "k2":"v2", "k3":"v3"}`                                                                      |
    | `ENV_INPUT_OTEL_THREADS`            | JSON string | `{"buffer":1000, "threads":100}`                                                                         |
    | `ENV_INPUT_OTEL_STORAGE`            | JSON string | `{"storage":"./otel_storage", "capacity": 5120}`                                                         |
    | `ENV_INPUT_OTEL_HTTP`               | JSON string | `{"enable":true, "http_status_ok": 200, "trace_api": "/otel/v1/trace", "metric_api": "/otel/v1/metric"}` |
    | `ENV_INPUT_OTEL_GRPC`               | JSON string | `{"trace_enable": true, "metric_enable": true, "addr": "127.0.0.1:4317"}`                                |
    | `ENV_INPUT_OTEL_EXPECTED_HEADERS`   | JSON string | `{"ex_version": "1.2.3", "ex_name": "env_resource_name"}`                                                |

<!-- markdownlint-enable -->

### 注意事项 {#attentions}

1. 建议使用 gRPC 协议，gRPC 具有压缩率高、序列化快、效率更高等优点
2. 自 [Datakit 1.10.0](../datakit/changelog.md#cl-1.10.0) 版本开始，http 协议的路由是可配置的，默认请求路径（Trace/Metric）分别为 `/otel/v1/trace` 和 `/otel/v1/metric`
3. 在涉及到 `float/double` 类型数据时，会最多保留两位小数
4. HTTP 和 gRPC 都支持 gzip 压缩格式。在 exporter 中可配置环境变量来开启：`OTEL_EXPORTER_OTLP_COMPRESSION = gzip`, 默认是不会开启 gzip。
5. HTTP 协议请求格式同时支持 JSON 和 Protobuf 两种序列化格式。但 gRPC 仅支持 Protobuf 一种。

使用 OTEL HTTP exporter 时注意环境变量的配置，由于 Datakit 的默认配置是 `/otel/v1/trace` 和 `/otel/v1/metric`，所以想要使用 HTTP 协议的话，需要单独配置 `trace` 和 `metric`，

## SDK 常规配置 {#sdk-configuration}

| 命令                           | 说明                                         | 默认                    | 注意                                          |
|:------------------------------|:--------------------------------------------|:------------------------|:---------------------------------------------|
| `OTEL_SDK_DISABLED`           | 关闭 SDK                                     | false                   | 关闭后将不会产生任何链路指标信息                   |
| `OTEL_RESOURCE_ATTRIBUTES`    | "service.name=App,username=liu"             |                         | 每一个 span 中都会有该 tag 信息                  |
| `OTEL_SERVICE_NAME`           | 服务名，等效于上面 "service.name=App"           |                         | 优先级高于上面                                 |
| `OTEL_LOG_LEVEL`              | 日志级别                                      | `info`                  |                                              |
| `OTEL_PROPAGATORS`            | 透传协议                                      | `tracecontext,baggage`  |                                              |
| `OTEL_TRACES_SAMPLER`         | 采样                                         | `parentbased_always_on` |                                              |
| `OTEL_TRACES_SAMPLER_ARG`     | 配合上面采样 参数                              | 1.0                     | 0 - 1.0                                      |
| `OTEL_EXPORTER_OTLP_PROTOCOL` | 协议包括： `grpc`,`http/protobuf`,`http/json` | gRPC                    |                                              |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | OTLP 地址                                    | <http://localhost:4317> | <http://datakit-endpoint:9529/otel/v1/trace> |
| `OTEL_TRACES_EXPORTER`        | 链路导出器                                    | `otlp`                  |                                              |


> 您可以将 `otel.javaagent.debug=true` 参数传递给 Agent 以查看调试日志。请注意，这些日志内容相当冗长，生产环境下谨慎使用。

## 链路 {#tracing}

Trace（链路）是由多个 span 组成的一条链路信息。
无论是单个服务还是一个服务集群，链路信息提供了一个请求发生到结束所经过的所有服务之间完整路径的集合。

Datakit 只接收 OTLP 的数据，OTLP 有三种数据类型： `gRPC` ， `http/protobuf` 和 `http/json` ，具体配置可以参考：

```shell
# OpenTelemetry 默认采用 gPRC 协议发送到 Datakit
-Dotel.exporter=otlp \
-Dotel.exporter.otlp.protocol=grpc \
-Dotel.exporter.otlp.endpoint=http://datakit-endpoint:4317

# 使用 http/protobuf 方式
-Dotel.exporter=otlp \
-Dotel.exporter.otlp.protocol=http/protobuf \
-Dotel.exporter.otlp.traces.endpoint=http://datakit-endpoint:9529/otel/v1/trace \
-Dotel.exporter.otlp.metrics.endpoint=http://datakit-endpoint:9529/otel/v1/metric 

# 使用 http/json 方式
-Dotel.exporter=otlp \
-Dotel.exporter.otlp.protocol=http/json \
-Dotel.exporter.otlp.traces.endpoint=http://datakit-endpoint:9529/otel/v1/trace \
-Dotel.exporter.otlp.metrics.endpoint=http://datakit-endpoint:9529/otel/v1/metric
```

### 链路采样 {#sample}

可以采用头部采样或者尾部采样，具体可以查看两篇最佳实践：

- 需要配合 collector 的尾部采样： [OpenTelemetry 采样最佳实践](../best-practices/cloud-native/opentelemetry-simpling.md)
- Agent 端的头部采样： [OpenTelemetry Java Agent 端采样策略](../best-practices/cloud-native/otel-agent-sampling.md)

### Tag {#tags}

从 DataKit 版本 [1.22.0](../datakit/changelog.md#cl-1.22.0) 开始，黑名单功能废弃。增加固定标签列表，只有在此列表中的才会提取到一级标签中，以下是固定列表：

| Attributes                 | tag                   | 说明                        |
|:---------------------------|:----------------------|:--------------------------|
| http.url                   | http_url              | HTTP 请求完整路径               |
| http.hostname              | http_hostname         | hostname                  |
| http.route                 | http_route            | 路由                        |
| http.status_code           | http_status_code      | 状态码                       |
| http.request.method        | http_request_method   | 请求方法                      |
| http.method                | http_method           | 同上                        |
| http.client_ip             | http_client_ip        | 客户端 IP                    |
| http.scheme                | http_scheme           | 请求协议                      |
| url.full                   | url_full              | 请求全路径                     |
| url.scheme                 | url_scheme            | 请求协议                      |
| url.path                   | url_path              | 请求路径                      |
| url.query                  | url_query             | 请求参数                      |
| span_kind                  | span_kind             | span 类型                   |
| db.system                  | db_system             | span 类型                   |
| db.operation               | db_operation          | DB 动作                     |
| db.name                    | db_name               | 数据库名称                     |
| db.statement               | db_statement          | 详细信息                      |
| server.address             | server_address        | 服务地址                      |
| net.host.name              | net_host_name         | 请求的 host                  |
| server.port                | server_port           | 服务端口号                     |
| net.host.port              | net_host_port         | 同上                        |
| network.peer.address       | network_peer_address  | 网络地址                      |
| network.peer.port          | network_peer_port     | 网络端口                      |
| network.transport          | network_transport     | 协议                        |
| messaging.system           | messaging_system      | 消息队列名称                    |
| messaging.operation        | messaging_operation   | 消息动作                      |
| messaging.message          | messaging_message     | 消息                        |
| messaging.destination      | messaging_destination | 消息详情                      |
| rpc.service                | rpc_service           | RPC 服务地址                  |
| rpc.system                 | rpc_system            | RPC 服务名称                  |
| error                      | error                 | 是否错误                      |
| error.message              | error_message         | 错误信息                      |
| error.stack                | error_stack           | 堆栈信息                      |
| error.type                 | error_type            | 错误类型                      |
| error.msg                  | error_message         | 错误信息                      |
| project                    | project               | project                   |
| version                    | version               | 版本                        |
| env                        | env                   | 环境                        |
| host                       | host                  | Attributes 中的 host 标签     |
| pod_name                   | pod_name              | Attributes 中的 pod_name 标签 |

如果想要增加自定义标签，可使用环境变量：

```shell
# 通过启动参数添加自定义标签
-Dotel.resource.attributes=username=myName,env=1.1.0
```

并修改配置文件中的白名单，这样就可以在观测云的链路详情的一级标签出现自定义的标签。

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

### 示例 {#examples}

Datakit 目前提供了如下两种语言的最佳实践：

- [Golang](opentelemetry-go.md)
- [Java](opentelemetry-java.md)

## 指标 {#metric}

OpenTelemetry Java Agent 从应用程序中通过 JMX 协议获取 MBean 的指标信息，Java Agent 通过内部 SDK 报告选定的 JMX 指标，这意味着所有的指标都是可以配置的。

可以通过命令 `otel.jmx.enabled=true/false` 开启和关闭 JMX 指标采集，默认是开启的。

为了控制 MBean 检测尝试之间的时间间隔，可以使用 `otel.jmx.discovery.delay` 命令，该属性定义了在第一个和下一个检测周期之间通过的毫秒数。

另外 Agent 内置的一些三方软件的采集配置。具体可以参考： [GitHub OTEL JMX Metric](https://github.com/open-telemetry/opentelemetry-java-instrumentation/blob/main/instrumentation/jmx-metrics/javaagent/README.md){:target="_blank"}

{{ range $i, $m := .Measurements }}

### `{{$m.Name}}`

{{$m.Desc}}

- 标签

{{$m.TagsMarkdownTable}}

- 指标列表

{{$m.FieldsMarkdownTable}}

{{ end }}

## 更多文档 {#more-readings}

- [Golang SDK](https://github.com/open-telemetry/opentelemetry-go){:target="_blank"}
- [官方使用手册](https://opentelemetry.io/docs/){:target="_blank"}
- [环境变量配置](https://github.com/open-telemetry/opentelemetry-java/blob/main/sdk-extensions/autoconfigure/README.md#otlp-exporter-both-span-and-metric-exporters){:target="_blank"}
- [观测云二次开发版本](https://github.com/GuanceCloud/opentelemetry-java-instrumentation){:target="_blank"}
