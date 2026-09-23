---
title     : 'LoongSuite'
summary   : '接收 LoongSuite 通过 OpenTelemetry 上报的链路数据'
tags      :
  - 'GOLANG'
  - 'OTEL'
  - '链路追踪'
__int_icon: 'icon/opentelemetry'
---


LoongSuite 开源版本支持将 Go 应用链路数据通过 OpenTelemetry OTLP 协议发送到采集器。DataKit 可使用 [OpenTelemetry](opentelemetry.md) 采集器接收 LoongSuite 上报的链路数据，无需额外开启专用采集器。

在接入 LoongSuite 前，请先确认 DataKit 已开启 OpenTelemetry 采集器，并根据使用的 OTLP 协议选择 gRPC 或 `HTTP/protobuf` 接入方式。

## 配置 {#config}

### DataKit 配置 {#datakit-config}

将示例配置复制到 `conf.d` 根目录，确认已开启 gRPC 或 HTTP 接收配置。默认安装路径下的命令如下：

```shell
cp /usr/local/datakit/conf.d/samples/opentelemetry.conf.sample /usr/local/datakit/conf.d/opentelemetry.conf
```

OpenTelemetry 采集器默认接收地址：

| 协议 | 地址 |
| --- | --- |
| `gRPC` | `http://<datakit-host>:4317` |
| `HTTP/protobuf` | `http://<datakit-host>:9529/otel/v1/traces` |

配置完成后，[重启 DataKit](../datakit/datakit-service-how-to.md#manage-service)。

### LoongSuite 配置 {#loongsuite-config}

推荐使用 gRPC 协议发送链路数据：

```shell
export OTEL_TRACES_EXPORTER=otlp
export OTEL_EXPORTER_OTLP_PROTOCOL=grpc
export OTEL_EXPORTER_OTLP_ENDPOINT=http://<datakit-host>:4317
export OTEL_SERVICE_NAME=<service-name>
export OTEL_RESOURCE_ATTRIBUTES=deployment.environment=prod,service.version=1.0.0
```

如果使用 `HTTP/protobuf` 协议，需要显式配置 Trace endpoint 为 DataKit OpenTelemetry 采集器的默认路径：

```shell
export OTEL_TRACES_EXPORTER=otlp
export OTEL_EXPORTER_OTLP_PROTOCOL=http/protobuf
export OTEL_EXPORTER_OTLP_TRACES_ENDPOINT=http://<datakit-host>:9529/otel/v1/traces
export OTEL_SERVICE_NAME=<service-name>
export OTEL_RESOURCE_ATTRIBUTES=deployment.environment=prod,service.version=1.0.0
```

<!-- markdownlint-disable MD046 -->
???+ warning

    DataKit OpenTelemetry 采集器的 HTTP Trace 默认路径为 `/otel/v1/traces`。如果 LoongSuite 或 OpenTelemetry SDK 使用默认 OTLP HTTP 路径 `/v1/traces`，需要在 LoongSuite 侧显式调整 endpoint。
<!-- markdownlint-enable MD046 -->

## 字段映射 {#fields}

LoongSuite 上报的 OTLP span attributes 会由 OpenTelemetry 采集器按固定标签列表提取。常见字段映射如下：

| 原始字段                        | DataKit 字段                 |
|-----------------------------|----------------------------|
| `db.system`                 | `db_system`                |
| `db.system.name`            | `db_system`                |
| `db.operation`              | `db_operation`             |
| `db.operation.name`         | `db_operation`             |
| `db.query.text`             | `db_statement`             |
| `db.namespace`              | `db_name`                  |
| `db.collection.name`        | `db_collection`            |
| `http.request.method`       | `http_method`              |
| `http.response.status_code` | `http_status_code`         |
| `network.protocol.name`     | `net_protocol_name`        |
| `network.protocol.version`  | `net_protocol_version`     |
| `network.local.address`     | `network_local_address`    |
| `network.local.port`        | `network_local_port`       |
| `user_agent.original`       | `user_agent_original`      |
| `messaging.system`          | `messaging_system`         |
| `messaging.operation.name`  | `messaging_operation`      |
| `messaging.operation.type`  | `messaging_operation_type` |
| `messaging.message.id`      | `messaging_message_id`     |
| `rpc.system.name`           | `rpc_system`               |
| `rpc.method`                | `rpc_method`               |
| `rpc.grpc.status_code`      | `rpc_grpc_status_code`     |

如果需要提取 LoongSuite 上报的自定义 attributes，可在 `opentelemetry.conf` 中配置 `customer_tags`。`customer_tags` 支持正则表达式，命中的字段会被提升为标签，字段名中的 `.` 会替换为 `_`。

```toml
[[inputs.opentelemetry]]
  customer_tags = [
    "reg:^db\\.query\\.parameter\\.",
    "reg:^kratos\\.service\\.meta\\.",
    "reg:^gen_ai\\.other_input\\.",
    "reg:^gen_ai\\.other_output\\.",
  ]
```

## 查看数据 {#view}

完成配置后，启动接入 LoongSuite 的 Go 应用并产生请求。链路数据会进入 DataKit OpenTelemetry 采集器，并在<<<custom_key.brand_name>>>平台的链路查看器中按 `service.name` 展示。

如果链路未上报，可优先检查：

- DataKit OpenTelemetry 采集器是否已开启并重启生效；
- LoongSuite OTLP exporter 的协议和 endpoint 是否与 DataKit 监听配置一致；
- 使用 `HTTP/protobuf` 时，Trace endpoint 是否为 `/otel/v1/traces`；
- `OTEL_SERVICE_NAME` 或 `service.name` 是否已配置。
