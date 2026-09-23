---
title     : 'LoongSuite'
summary   : 'Receive LoongSuite trace data through OpenTelemetry'
tags      :
  - 'GOLANG'
  - 'OTEL'
  - 'Distributed Tracing'
__int_icon: 'icon/opentelemetry'
---


The open-source version of LoongSuite can export Go application trace data through the OpenTelemetry OTLP protocol. DataKit can receive LoongSuite trace data through the [OpenTelemetry](opentelemetry.md) input. No dedicated LoongSuite input is required.

Before connecting LoongSuite to DataKit, make sure the DataKit OpenTelemetry input is enabled, then choose either gRPC or `HTTP/protobuf` according to the OTLP protocol used by LoongSuite.

## Configuration {#config}

### DataKit Configuration {#datakit-config}

Copy the sample configuration to the `conf.d` root and make sure the gRPC or HTTP receiver is enabled. With the default installation path, run:

```shell
cp /usr/local/datakit/conf.d/samples/opentelemetry.conf.sample /usr/local/datakit/conf.d/opentelemetry.conf
```

Default endpoints of the OpenTelemetry input:

| Protocol | Endpoint |
| --- | --- |
| `gRPC` | `http://<datakit-host>:4317` |
| `HTTP/protobuf` | `http://<datakit-host>:9529/otel/v1/traces` |

After configuration, [restart DataKit](../datakit/datakit-service-how-to.md#manage-service).

### LoongSuite Configuration {#loongsuite-config}

The gRPC protocol is recommended for trace export:

```shell
export OTEL_TRACES_EXPORTER=otlp
export OTEL_EXPORTER_OTLP_PROTOCOL=grpc
export OTEL_EXPORTER_OTLP_ENDPOINT=http://<datakit-host>:4317
export OTEL_SERVICE_NAME=<service-name>
export OTEL_RESOURCE_ATTRIBUTES=deployment.environment=prod,service.version=1.0.0
```

If `HTTP/protobuf` is used, explicitly set the Trace endpoint to the default path of the DataKit OpenTelemetry input:

```shell
export OTEL_TRACES_EXPORTER=otlp
export OTEL_EXPORTER_OTLP_PROTOCOL=http/protobuf
export OTEL_EXPORTER_OTLP_TRACES_ENDPOINT=http://<datakit-host>:9529/otel/v1/traces
export OTEL_SERVICE_NAME=<service-name>
export OTEL_RESOURCE_ATTRIBUTES=deployment.environment=prod,service.version=1.0.0
```

<!-- markdownlint-disable MD046 -->
???+ warning

    The default HTTP Trace path of the DataKit OpenTelemetry input is `/otel/v1/traces`. If LoongSuite or the OpenTelemetry SDK uses the default OTLP HTTP path `/v1/traces`, explicitly adjust the endpoint on the LoongSuite side.
<!-- markdownlint-enable MD046 -->

## Field Mapping {#fields}

OTLP span attributes reported by LoongSuite are extracted by the OpenTelemetry input according to the fixed tag list. Common mappings are as follows:

| Original Field | DataKit Field |
| --- | --- |
| `db.system` | `db_system` |
| `db.system.name` | `db_system` |
| `db.operation` | `db_operation` |
| `db.operation.name` | `db_operation` |
| `db.query.text` | `db_statement` |
| `db.namespace` | `db_name` |
| `db.collection.name` | `db_collection` |
| `http.request.method` | `http_method` |
| `http.response.status_code` | `http_status_code` |
| `network.protocol.name` | `net_protocol_name` |
| `network.protocol.version` | `net_protocol_version` |
| `network.local.address` | `network_local_address` |
| `network.local.port` | `network_local_port` |
| `user_agent.original` | `user_agent_original` |
| `messaging.system` | `messaging_system` |
| `messaging.operation.name` | `messaging_operation` |
| `messaging.operation.type` | `messaging_operation_type` |
| `messaging.message.id` | `messaging_message_id` |
| `rpc.system.name` | `rpc_system` |
| `rpc.method` | `rpc_method` |
| `rpc.grpc.status_code` | `rpc_grpc_status_code` |

To extract custom attributes reported by LoongSuite, configure `customer_tags` in `opentelemetry.conf`. The `customer_tags` setting supports regular expressions. Matched fields are promoted to tags, and `.` in field names is replaced with `_`.

```toml
[[inputs.opentelemetry]]
  customer_tags = [
    "reg:^db\\.query\\.parameter\\.",
    "reg:^kratos\\.service\\.meta\\.",
    "reg:^gen_ai\\.other_input\\.",
    "reg:^gen_ai\\.other_output\\.",
  ]
```

## View Data {#view}

After the configuration is complete, start the Go application instrumented with LoongSuite and generate requests. Trace data will be received by the DataKit OpenTelemetry input and displayed by `service.name` in the trace viewer on the <<<custom_key.brand_name>>> platform.

If no trace data is reported, check the following items first:

- Whether the DataKit OpenTelemetry input is enabled and DataKit has been restarted;
- Whether the LoongSuite OTLP exporter protocol and endpoint match the DataKit listener configuration;
- When `HTTP/protobuf` is used, whether the Trace endpoint is `/otel/v1/traces`;
- Whether `OTEL_SERVICE_NAME` or `service.name` is configured.
