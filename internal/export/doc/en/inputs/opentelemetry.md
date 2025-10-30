---
title     : 'OpenTelemetry'
summary   : 'Receive OpenTelemetry Metrics, Logs, and APM Data'
__int_icon: 'icon/opentelemetry'
tags      :
  - 'OTEL'
  - 'Distributed Tracing'
dashboard :
  - desc  : 'OpenTelemetry JVM Monitoring View'
    path  : 'dashboard/en/opentelemetry'
monitor   :
  - desc  : 'None'
    path  : '-'
---


{{.AvailableArchs}}

---

OpenTelemetry (hereinafter referred to as OTEL) is an observability project under CNCF (Cloud Native Computing Foundation). It aims to provide a standardized solution in the field of observability, addressing standardization issues related to the data model, collection, processing, and export of observability data.

OTEL is a collection of standards and tools designed to manage observability data such as traces, metrics, and logs. This document describes how to configure and enable OTEL data ingestion on DataKit, as well as best practices for Java and Go.

## Configuration {#config}

<!-- markdownlint-disable MD046 -->
=== "Host Installation"

    Navigate to the `conf.d/{{.Catalog}}` directory under the DataKit installation directory, copy `{{.InputName}}.conf.sample` and rename it to `{{.InputName}}.conf`. An example is as follows:

    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```

    After configuration, [restart DataKit](../datakit/datakit-service-how-to.md#manage-service) to take effect.

=== "Kubernetes"

    You can enable the collector by [injecting collector configuration via ConfigMap](../datakit/datakit-daemonset-deploy.md#configmap-setting) or [configuring ENV_DATAKIT_INPUTS](../datakit/datakit-daemonset-deploy.md#env-setting).

    You can also modify configuration parameters via environment variables (you need to add the collector to ENV_DEFAULT_ENABLED_INPUTS as a default collector):

{{ CodeBlock .InputENVSample 4 }}

<!-- markdownlint-enable -->

The `customer_tags` parameter supports regular expressions but requires a fixed prefix format `reg:`. For example, `reg:key_*` matches all keys starting with `key_`.

### Notes {#attentions}

1. It is recommended to use the gRPC protocol, as gRPC offers advantages such as high compression rate, fast serialization, and higher efficiency.
1. Starting from DataKit version [1.10.0](../datakit/changelog.md#cl-1.10.0), the routes for the HTTP protocol are configurable. The default request paths (for Trace/Metric) are `/otel/v1/traces`, `/otel/v1/logs`, and `/otel/v1/metrics` respectively.
1. For `float/double` type data, a maximum of two decimal places will be retained.
1. Both HTTP and gRPC support the gzip compression format. You can configure an environment variable in the exporter to enable it: `OTEL_EXPORTER_OTLP_COMPRESSION = gzip`; gzip is disabled by default.
1. The HTTP protocol request format supports both JSON and Protobuf serialization formats. However, gRPC only supports the Protobuf format.

<!-- markdownlint-disable MD046 -->
???+ warning

    - The service name in DDTrace trace data is named based on the service name or referenced third-party libraries, while the service name of the OTEL collector is defined by `otel.service.name`.
    - To display service names separately, an additional field configuration is added: `spilt_service_name = true`.
    - The service name is extracted from the tags in the trace data. For example, if the DB-type tag is `db.system=mysql`, the service name will be `mysql`. For message queue types (e.g., `messaging.system=kafka`), the service name will be `kafka`.
    - By default, the service name is extracted from these three tags: `db.system/rpc.system/messaging.system`.
<!-- markdownlint-enable -->


Note the environment variable configuration when using the OTEL HTTP exporter. Since the default configuration of DataKit uses `/otel/v1/traces`, `/otel/v1/logs`, and `/otel/v1/metrics`, you need to configure `trace` and `metric` separately if you want to use the HTTP protocol.

### Agent V2 Version {#v2}

The V2 version uses `otlp exporter` by default, changing the previous `grpc` to `http/protobuf`. You can set it via the command `-Dotel.exporter.otlp.protocol=grpc`, or use the default `http/protobuf`.

If using HTTP, the path for each exporter needs to be explicitly configured. For example:

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

If using the gRPC protocol, explicit configuration is required; otherwise, the default HTTP protocol will be used:

```shell
java -javaagent:/usr/local/ddtrace/opentelemetry-javaagent-2.5.0.jar \
  -Dotel.exporter=otlp \
  -Dotel.exporter.otlp.protocol=grpc \
  -Dotel.exporter.otlp.endpoint=http://localhost:4317 \
  -Dotel.service.name=app \
  -jar app.jar
```

Logging is enabled by default. To disable log collection, set the exporter configuration to empty: `-Dotel.logs.exporter=none`.

For more major changes in the V2 version, refer to the official documentation or GitHub release notes: [Github-v2.0.0](https://github.com/open-telemetry/opentelemetry-java-instrumentation/releases/tag/v2.0.0){:target="_blank"}

### Common Commands {#sdk-configuration}

The following configurations are commonly used when starting an application:

| ENV (Corresponding Command)                                | Description                                                                                                 |
| ---:                                                       | ---                                                                                                          |
| `OTEL_SDK_DISABLED(otel.sdk.disabled)`                     | Disable the SDK; default is `false`. No trace metrics will be generated after disabling.                    |
| `OTEL_RESOURCE_ATTRIBUTES(otel.resource.attributes)`       | Add [global custom tags](https://opentelemetry.io/docs/languages/sdk-configuration/general/#otel_resource_attributes){:target="_blank"}. These custom tags will be included in each span. Example: `service.name=App,project=app-a` |
| `OTEL_SERVICE_NAME(otel.service.name)`                     | Set the service name; it has higher priority than custom tags.                                              |
| `OTEL_LOG_LEVEL(otel.log.level)`                           | Log level; default is `info`.                                                                                |
| `OTEL_PROPAGATORS(otel.propagators)`                       | Set the [propagation protocol](https://opentelemetry.io/docs/languages/sdk-configuration/general/#otel_propagators){:target="_blank"}; default is `tracecontext,baggage`. |
| `OTEL_TRACES_SAMPLER(otel.traces.sampler)`                 | Set the [sampler type](https://opentelemetry.io/docs/languages/sdk-configuration/general/#otel_traces_sampler){:target="_blank"}. |
| `OTEL_TRACES_SAMPLER_ARG(otel.traces.sampler.arg)`         | Used with the above sampler parameter; value range is *0~1.0*; default is `1.0`.                             |
| `OTEL_EXPORTER_OTLP_PROTOCOL(otel.exporter.otlp.protocol)` | Set the transmission protocol; default is `grpc`; optional values are `grpc,http/protobuf,http/json`.         |
| `OTEL_EXPORTER_OTLP_ENDPOINT(otel.exporter.otlp.endpoint)` | Set the Trace upload address; it should be set to the DataKit address: `http://datakit-endpoint:9529/otel/v1/traces`. |
| `OTEL_TRACES_EXPORTER(otel.traces.exporter)`               | Trace exporter; default is `otlp`.                                                                           |
| `OTEL_LOGS_EXPORTER(otel.logs.exporter)`                   | Log exporter; default is `otlp`. Note: Explicit configuration is required for OTEL V1 version; otherwise, it is disabled by default. |

> You can pass the `otel.javaagent.debug=true` parameter to the Agent to view debug logs. Note that these logs are quite verbose; use them with caution in production environments.

### Trace Sampling {#sample}

You can use head-based sampling or tail-based sampling. For details, refer to the two best practice documents:

- Tail-based sampling with collector: [OpenTelemetry Sampling Best Practices](../best-practices/cloud-native/opentelemetry-simpling.md)
- Head-based sampling on the Agent side: [OpenTelemetry Java Agent Sampling Strategy](../best-practices/cloud-native/otel-agent-sampling.md)

#### Tag Extraction {#tags}

Starting from DataKit version [1.22.0](../datakit/changelog.md#cl-1.22.0), the blacklist function is deprecated. A fixed tag list is added, and only tags in this list will be extracted into top-level tags. The fixed list is as follows:

| Attributes              | Tags                    | Description                         |
| ----------------------: | :---------------------- | :---------------------------------- |
| `http.url`              | `http_url`              | Full HTTP request path              |
| `http.hostname`         | `http_hostname`         | Hostname                            |
| `http.route`            | `http_route`            | Route                               |
| `http.status_code`      | `http_status_code`      | Status code                         |
| `http.request.method`   | `http_request_method`   | Request method                      |
| `http.method`           | `http_method`           | Same as above                       |
| `http.client_ip`        | `http_client_ip`        | Client IP                           |
| `http.scheme`           | `http_scheme`           | Request protocol                    |
| `url.full`              | `url_full`              | Full request URL                    |
| `url.scheme`            | `url_scheme`            | Request protocol                    |
| `url.path`              | `url_path`              | Request path                        |
| `url.query`             | `url_query`             | Request parameters                  |
| `span_kind`             | `span_kind`             | Span type                           |
| `db.system`             | `db_system`             | Span type                           |
| `db.operation`          | `db_operation`          | DB action                           |
| `db.name`               | `db_name`               | Database name                       |
| `db.statement`          | `db_statement`          | Detailed information                |
| `server.address`        | `server_address`        | Service address                     |
| `net.host.name`         | `net_host_name`         | Requested host                      |
| `server.port`           | `server_port`           | Service port number                 |
| `net.host.port`         | `net_host_port`         | Same as above                       |
| `network.peer.address`  | `network_peer_address`  | Network address                     |
| `network.peer.port`     | `network_peer_port`     | Network port                        |
| `network.transport`     | `network_transport`     | Protocol                            |
| `messaging.system`      | `messaging_system`      | Message queue name                  |
| `messaging.operation`   | `messaging_operation`   | Message action                      |
| `messaging.message`     | `messaging_message`     | Message                             |
| `messaging.destination` | `messaging_destination` | Message details                     |
| `rpc.service`           | `rpc_service`           | RPC service address                 |
| `rpc.system`            | `rpc_system`            | RPC service name                    |
| `error`                 | `error`                 | Whether an error occurred           |
| `error.message`         | `error_message`         | Error message                       |
| `error.stack`           | `error_stack`           | Stack trace information             |
| `error.type`            | `error_type`            | Error type                          |
| `error.msg`             | `error_message`         | Error message                       |
| `project`               | `project`               | Project                             |
| `version`               | `version`               | Version                             |
| `env`                   | `env`                   | Environment                         |
| `host`                  | `host`                  | Host tag in Attributes              |
| `pod_name`              | `pod_name`              | `pod_name` tag in Attributes        |
| `pod_namespace`         | `pod_namespace`         | `pod_namespace` tag in Attributes   |

To add custom tags, use the following environment variable:

```shell
# Add custom tags via startup parameters
-Dotel.resource.attributes=username=myName,env=1.1.0
```

##### Span kind {#kind}

All spans have the `span_kind` tag, which has 6 attributes:

- `unspecified`: Not set.
- `internal`: Internal span or child span type.
- `server`: WEB service, RPC service, etc.
- `client`: Client type.
- `producer`: Message producer.
- `consumer`: Message consumer.

### Metric Collection {#metric}

The OpenTelemetry Java Agent obtains MBean metric information from applications via the JMX protocol. The Java Agent reports selected JMX metrics through the internal SDK, which means all metrics are configurable.

You can enable or disable JMX metric reporting using the command `otel.jmx.enabled=true/false` (enabled by default). To control the time interval between MBean detection attempts, use the `otel.jmx.discovery.delay` command. This attribute defines the interval in milliseconds between the first and subsequent detection cycles.

In addition, the Agent has built-in collection configurations for some third-party software. For details, refer to: [GitHub OTEL JMX Metric](https://github.com/open-telemetry/opentelemetry-java-instrumentation/blob/main/instrumentation/jmx-metrics/javaagent/README.md){:target="_blank"}

We have implemented special handling for **Histogram** metrics:

- OpenTelemetry histogram buckets are directly mapped to Prometheus histogram buckets.

- The count of each bucket is converted to the Prometheus cumulative count format. For example, OpenTelemetry buckets `[0, 10)`, `[10, 50)`, `[50, 100)` are converted to Prometheus `_bucket` metrics with the `le` tag:

```text
  my_histogram_bucket{le="10"} 100
  my_histogram_bucket{le="50"} 200
  my_histogram_bucket{le="100"} 250
```

- The total number of observations in the OpenTelemetry histogram is converted to the Prometheus `_count` metric.

- The sum of the OpenTelemetry histogram is converted to the Prometheus `_sum` metric, and `_max` and `_min` are also added.

```text
  my_histogram_count 250
  my_histogram_max 100
  my_histogram_min 50
  my_histogram_sum 12345.67
```

All metrics ending with `_bucket` are histogram data, and there must be corresponding metrics ending with `_max`, `_min`, `_count`, and `sum`.

You can use the `le (less than or equal)` tag to categorize histogram data and filter based on tags. For all metrics and tags, refer to [OpenTelemetry Metrics](https://opentelemetry.io/docs/specs/semconv/){:target="_blank"}.

This conversion enables seamless integration of histogram data collected by OpenTelemetry into Prometheus, allowing you to leverage Prometheus' powerful query and visualization capabilities for analysis.

### Log Collection {#logging}

[:octicons-tag-24: Version-1.33.0](../datakit/changelog.md#cl-1.33.0)

Currently, the JAVA Agent supports collecting `stdout` logs and sending them to DataKit via the `otlp` protocol using the [Standard output](https://opentelemetry.io/docs/specs/otel/logs/sdk_exporters/stdout/){:target="_blank"} method.

By default, log collection is disabled for OTEL Agent V1. Explicit commands are required to enable it. The enabling methods are as follows:

```shell hl_lines='2 8'
# env
export OTEL_LOGS_EXPORTER=OTLP
export OTEL_EXPORTER_OTLP.ENDPOINT=http://<DataKit Addr>:4317
java -jar app.jar

# command
java -javaagent:/path/to/agnet.jar \
  -otel.logs.exporter=otlp \
  -Dotel.exporter.otlp.endpoint=http://<DataKit Addr>:4317 \
  -jar app.jar
```

By default, the maximum length of log content is 500KB. Content exceeding this limit will be split into multiple logs. The maximum length of log tags is 32KB (this field is not configurable), and content exceeding this limit will be truncated.

The `source` of logs collected via OTEL is the service name. You can also customize it by adding a tag: `log.source`. For example: `-Dotel.resource.attributes="log.source=source_name"`.

> Note: If the app runs in a container environment (e.g., k8s), DataKit will [automatically collect logs](container-log.md#logging-stdout){:target="_blank"} by default. Enabling log collection again will result in duplicate collection. It is recommended to [manually disable DataKit's independent log collection](container-log.md#logging-with-image-config){:target="_blank"} before enabling OTEL log collection.

For more languages, refer to the [official documentation](https://opentelemetry.io/docs/specs/otel/logs/){:target="_blank"}.

## Collection Field Description {#fields}

### Tracing {#tracing}

{{ range $i, $m := .Measurements }}

{{if eq $m.Type "tracing"}}

#### `{{$m.Name}}`

{{$m.Desc}}

{{$m.MarkdownTable}}

{{ end }}
{{ end }}

### Metrics {#metrics}

{{ range $i, $m := .Measurements }}

{{if eq $m.Type "metric"}}

#### `{{$m.Name}}`

{{$m.Desc}}

{{$m.MarkdownTable}}

{{ end }}
{{ end }}

##### Deleted Tags in Metrics {#del-metric}

In the `otel_service` metric set, there are many useless tags in the originally reported metrics. These tags are of String type and are discarded due to high memory and bandwidth consumption. The discarded tags are as follows:

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

## Examples {#examples}

DataKit currently provides best practices for the following two languages:

- [Golang](opentelemetry-go.md)
- [Java](opentelemetry-java.md)


## More Documents {#more-readings}

- [Golang SDK](https://github.com/open-telemetry/opentelemetry-go){:target="_blank"}
- [Official User Guide](https://opentelemetry.io/docs/){:target="_blank"}
- [Environment Variable Configuration](https://github.com/open-telemetry/opentelemetry-java/blob/main/sdk-extensions/autoconfigure/README.md#otlp-exporter-both-span-and-metric-exporters){:target="_blank"}
- [Sampling Strategy Notes for DDTrace and OpenTelemetry Integration](tracing-sample.md)
