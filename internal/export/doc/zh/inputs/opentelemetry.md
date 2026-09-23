---
title     : 'OpenTelemetry'
summary   : '接收 OpenTelemetry 指标、日志、APM 数据'
__int_icon: 'icon/opentelemetry'
tags      :
  - 'OTEL'
  - '链路追踪'
dashboard :
  - desc  : 'OpenTelemetry JVM 监控视图'
    path  : 'dashboard/zh/opentelemetry'
monitor   :
  - desc  : '暂无'
    path  : '-'
---


{{.AvailableArchs}}

---

OpenTelemetry（简称 OTEL）是 CNCF 的可观测性标准体系。DataKit 的 `opentelemetry` 输入用于接收 OTEL 的 traces、metrics、logs。

## 配置 {#config}

<!-- markdownlint-disable MD046 -->
=== "主机安装"

    将示例配置复制到 `conf.d` 根目录并修改。默认安装路径下的命令如下：

    ```shell
    cp /usr/local/datakit/conf.d/samples/opentelemetry.conf.sample /usr/local/datakit/conf.d/opentelemetry.conf
    ```

    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```

    配置完成后重启 DataKit 生效：
    [重启 DataKit](../datakit/datakit-service-how-to.md#manage-service)

=== "Kubernetes"

    通过 ConfigMap 或环境变量方式开启采集器：
    [ConfigMap 注入方式](../datakit/datakit-daemonset-deploy.md#configmap-setting) 或
    [ENV_DATAKIT_INPUTS 方式](../datakit/datakit-daemonset-deploy.md#env-setting)。

    也可以使用环境变量直接配置（需将该采集器加入 `ENV_DEFAULT_ENABLED_INPUTS`）：

{{ CodeBlock .InputENVSampleZh 4 }}
<!-- markdownlint-enable MD046 -->

`customer_tags` 支持正则匹配，使用时必须以 `reg:` 为前缀，比如 `reg:key_*`。

### 注意事项 {#attentions}

1. 推荐使用 gRPC，压缩率高、序列化快，资源开销相对更低。
2. 自 DataKit [1.10.0](../datakit/changelog.md#cl-1.10.0) 起，HTTP 路由可配置。默认值：
   - traces: `/otel/v1/traces`
   - metrics: `/otel/v1/metrics`
   - logs: `/otel/v1/logs`
3. `float/double` 类型在 DataKit 侧会保留最多两位小数。
4. HTTP 与 gRPC 都支持 gzip。可通过 exporter 配置开启，例如 `OTEL_EXPORTER_OTLP_COMPRESSION=gzip`。
5. HTTP 支持 JSON 与 Protobuf 两种序列化提交格式；但 DataKit HTTP 采集仅支持 `application/x-protobuf`。

<!-- markdownlint-disable MD046 -->
???+ warning

    - DDTrace 链路的服务名通常来自 DDTrace 或框架库的 `service.name`。
    - OTEL 链路的服务名由 `otel.service.name` 决定。
    - 如果你希望按 `db.system`、`rpc.system`、`messaging.system` 将服务名拆分展示，可开启：

      `split_service_name = true`
    - 开启后优先级为 `db.system`，其次 `rpc.system`，最后 `messaging.system`。

<!-- markdownlint-enable MD046 -->

使用 OTEL HTTP exporter 时，请按 DataKit 实际地址分别配置 OTEL 路由：
traces `/otel/v1/traces`、metrics `/otel/v1/metrics`、logs `/otel/v1/logs`（默认监听端口 9529）。

### Java Agent V2 协议行为 {#v2}

OTEL Java Agent V2 默认使用 `http/protobuf` 作为 OTLP 协议，若需切回 gRPC：

```shell
java -javaagent:/opt/opentelemetry/opentelemetry-javaagent.jar \
  -Dotel.traces.exporter=otlp \
  -Dotel.exporter.otlp.protocol=grpc \
  -Dotel.exporter.otlp.endpoint=http://localhost:4317 \
  -Dotel.service.name=app \
  -jar app.jar
```

如使用 HTTP 方式，请为每类数据显式配置 endpoint：

```shell
java -javaagent:/opt/opentelemetry/opentelemetry-javaagent.jar \
  -Dotel.traces.exporter=otlp \
  -Dotel.exporter.otlp.protocol=http/protobuf \
  -Dotel.exporter.otlp.logs.endpoint=http://localhost:9529/otel/v1/logs \
  -Dotel.exporter.otlp.traces.endpoint=http://localhost:9529/otel/v1/traces \
  -Dotel.exporter.otlp.metrics.endpoint=http://localhost:9529/otel/v1/metrics \
  -Dotel.service.name=app \
  -jar app.jar
```

需要关闭日志采集时设置：

`-Dotel.logs.exporter=none`

更多 V2 变更请参见： [GitHub-v2.0.0](https://github.com/open-telemetry/opentelemetry-java-instrumentation/releases/tag/v2.0.0){:target="_blank"}

### 常用配置 {#sdk-configuration}

以下为接入 DataKit 常用的 OTEL 配置项（节选）：

| 配置项（环境变量/系统属性） | 说明 |
| --- | --- |
| `OTEL_SDK_DISABLED(otel.sdk.disabled)` | 是否关闭 SDK，默认 `false`。 |
| `OTEL_RESOURCE_ATTRIBUTES(otel.resource.attributes)` | 全局资源标签，例如 `service.name=app,project=app-a`。 |
| `OTEL_SERVICE_NAME(otel.service.name)` | 服务名，优先级高于资源标签。 |
| `OTEL_LOG_LEVEL(otel.log.level)` | SDK 日志级别，默认 `info`。 |
| `OTEL_PROPAGATORS(otel.propagators)` | 透传协议，默认 `tracecontext,baggage`。 |
| `OTEL_TRACES_SAMPLER(otel.traces.sampler)` | 采样器类型。 |
| `OTEL_TRACES_SAMPLER_ARG(otel.traces.sampler.arg)` | 与采样器配合的参数，范围 `0~1.0`，默认 `1.0`。 |
| `OTEL_EXPORTER_OTLP_PROTOCOL(otel.exporter.otlp.protocol)` | 传输协议，支持 `grpc`、`http/protobuf`；默认值取决于 SDK 或发行版，Java Agent 2.x 默认使用 `http/protobuf`。 |
| `OTEL_EXPORTER_OTLP_ENDPOINT(otel.exporter.otlp.endpoint)` | 统一 OTLP 地址，例如 gRPC 模式使用 `http://datakit-host:4317`，HTTP 模式使用服务地址 `http://datakit-host:9529`。 |
| `OTEL_EXPORTER_OTLP_TRACES_ENDPOINT(otel.exporter.otlp.traces.endpoint)` | HTTP traces 端点，例如 `http://datakit-host:9529/otel/v1/traces`。 |
| `OTEL_EXPORTER_OTLP_METRICS_ENDPOINT(otel.exporter.otlp.metrics.endpoint)` | HTTP metrics 端点，例如 `http://datakit-host:9529/otel/v1/metrics`。 |
| `OTEL_EXPORTER_OTLP_LOGS_ENDPOINT(otel.exporter.otlp.logs.endpoint)` | HTTP logs 端点，例如 `http://datakit-host:9529/otel/v1/logs`。 |
| `OTEL_TRACES_EXPORTER(otel.traces.exporter)` | 链路导出器，默认 `otlp`。 |
| `OTEL_LOGS_EXPORTER(otel.logs.exporter)` | 日志导出器，启用日志时需设置为 `otlp`。 |
| `OTEL_METRICS_EXPORTER(otel.metrics.exporter)` | 指标导出器，启用指标时需设置为 `otlp`。 |

从 DataKit [1.85.0](../datakit/changelog.md#cl-1.85.0) 起，HTTP `http/json` 已不再支持；使用 `http/protobuf`。

可设置 `otel.javaagent.debug=true` 打开 Java Agent 调试日志，请勿在生产持续开启。

### 链路采样 {#sample}

可选 head-based 或 tail-based 方案：

- 尾部采样（collector）：[OpenTelemetry 采样最佳实践](../best-practices/cloud-native/opentelemetry-simpling.md)
- 头部采样（Agent）：[OpenTelemetry Java Agent 采样策略](../best-practices/cloud-native/otel-agent-sampling.md)

#### Tag 提取 {#tags}

从 DataKit [1.22.0](../datakit/changelog.md#cl-1.22.0) 起，`tags` 提取从黑名单改为白名单，以下为固定映射清单：

| Attributes | Tags | 说明 |
|:---|:---|:---|
| `http.url` | `http_url` | 请求完整 URL |
| `http.hostname` | `http_hostname` | Hostname |
| `http.route` | `http_route` | 路由 |
| `http.status_code` | `http_status_code` | 状态码 |
| `http.request.method` | `http_request_method` | 请求方法 |
| `http.method` | `http_method` | 同上 |
| `http.client_ip` | `http_client_ip` | 客户端 IP |
| `http.scheme` | `http_scheme` | 请求协议 |
| `url.full` | `url_full` | 完整请求 URL |
| `url.scheme` | `url_scheme` | URL 协议 |
| `url.path` | `url_path` | 请求路径 |
| `url.query` | `url_query` | 请求参数 |
| `span_kind` | `span_kind` | Span 类型 |
| `db.system` | `db_system` | 数据库系统 |
| `db.operation` | `db_operation` | DB 操作 |
| `db.name` | `db_name` | 数据库名 |
| `db.statement` | `db_statement` | SQL 文本 |
| `server.address` | `server_address` | 服务地址 |
| `net.host.name` | `net_host_name` | Host 名 |
| `server.port` | `server_port` | 服务端口 |
| `net.host.port` | `net_host_port` | 主机端口 |
| `network.peer.address` | `network_peer_address` | 对端地址 |
| `network.peer.port` | `network_peer_port` | 对端端口 |
| `network.transport` | `network_transport` | 网络协议 |
| `messaging.system` | `messaging_system` | 消息系统 |
| `messaging.operation` | `messaging_operation` | 消息动作 |
| `messaging.message` | `messaging_message` | 消息 |
| `messaging.destination` | `messaging_destination` | 消息目标 |
| `rpc.service` | `rpc_service` | RPC 服务名 |
| `rpc.system` | `rpc_system` | RPC 系统 |
| `error` | `error` | 是否错误 |
| `error.message` | `error_message` | 错误信息 |
| `error.stack` | `error_stack` | 堆栈 |
| `error.type` | `error_type` | 错误类型 |
| `project` | `project` | 项目 |
| `version` | `version` | 版本 |
| `env` | `env` | 环境 |
| `host` | `host` | 主机 |
| `pod_name` | `pod_name` | Pod 名 |
| `pod_namespace` | `pod_namespace` | Pod 命名空间 |
| `telemetry.sdk.language` | `sdk_language` | SDK 语言 |
| `telemetry.sdk.name` | `sdk_name` | SDK 名 |
| `telemetry.sdk.version` | `sdk_version` | SDK 版本 |

添加自定义资源标签：

```shell
-Dotel.resource.attributes=service.name=app,version=1.1.0,env=prod
```

##### Span Kind {#kind}

- `unspecified`：未指定
- `internal`：内部 span
- `server`：服务端 span
- `client`：客户端 span
- `producer`：消息生产者
- `consumer`：消息消费者

### 指标采集 {#metric}

Java Agent 通过内置 JMX 支持上报 JVM 和相关中间件指标。可在应用中通过：

- `otel.jmx.enabled=true/false`（默认开启）控制是否上报 JMX；
- `otel.jmx.discovery.delay` 调整探测间隔（单位：毫秒）。

更多 JMX 扩展支持见： [GitHub OTEL JMX Metric](https://github.com/open-telemetry/opentelemetry-java-instrumentation/blob/main/instrumentation/jmx-metrics/javaagent/README.md){:target="_blank"}

### 直方图指标转换 {#histogram-conversion}

OTEL 直方图会转换为 Prometheus 直方图语义：

输入区间：

```text
[0, 10), [10, 50), [50, 100)
```

转换后的指标：

```text
my_histogram_bucket{le="10"} 100
my_histogram_bucket{le="50"} 200
my_histogram_bucket{le="100"} 250
```

并补齐：

```text
my_histogram_count 250
my_histogram_max 100
my_histogram_min 50
my_histogram_sum 12345.67
```

凡 `_bucket` 结尾的指标应对应同一维度的 `_count`、`_sum`、`_min`、`_max`。

### 日志采集 {#logging}

[:octicons-tag-24: Version-1.33.0](../datakit/changelog.md#cl-1.33.0)

OTEL logs 通过 OTLP 上报到 DataKit。OTEL V1 默认不采集日志，需显式开启：

```shell hl_lines='1 4 5 8'
# 环境变量
export OTEL_LOGS_EXPORTER=otlp
export OTEL_EXPORTER_OTLP_ENDPOINT=http://<DataKit Addr>:4317
export OTEL_EXPORTER_OTLP_PROTOCOL=grpc
java -jar app.jar

# 命令行
java -javaagent:/path/to/agent.jar \
  -Dotel.logs.exporter=otlp \
  -Dotel.exporter.otlp.endpoint=http://<DataKit Addr>:4317 \
  -Dotel.exporter.otlp.protocol=grpc \
  -jar app.jar
```

如果你使用 V2 的 HTTP/Protobuf，需要设置：
`-Dotel.exporter.otlp.protocol=http/protobuf` 和
`-Dotel.exporter.otlp.logs.endpoint=http://<DataKit Addr>:9529/otel/v1/logs`。

默认 message 字段最大 500KB，超出部分会截断，tag 最大 32KB。
