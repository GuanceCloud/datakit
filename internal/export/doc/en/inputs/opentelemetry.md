---
title     : 'OpenTelemetry'
summary   : 'Collect OpenTelemetry metric, log and APM data'
tags      :
  - 'OTEL'
  - 'APM'
  - 'TRACING'
__int_icon      : 'icon/opentelemetry'
dashboard :
  - desc  : 'Opentelemetry JVM Monitoring View'
    path  : 'dashboard/en/opentelemetry'
monitor   :
  - desc  : 'N/A'
    path  : '-'
---

{{.AvailableArchs}}

---

:fontawesome-brands-linux: :fontawesome-brands-windows: :fontawesome-brands-apple: :material-kubernetes: :material-docker:

---

OpenTelemetry (hereinafter referred to as OTEL) is an observability project of CNCF, which aims to provide a standardization scheme in the field of observability and solve the standardization problems of data model, collection, processing and export of observation data.

OTEL is a collection of standards and tools for managing observational data, such as trace, metrics, logs, etc. (new observational data types may appear in the future).

OTEL provides vendor-independent implementations that export observation class data to different backends, such as open source Prometheus, Jaeger, DataKit, or cloud vendor services, depending on the user's needs.

The purpose of this article is to introduce how to configure and enable OTEL data access on DataKit, and the best practices of Java and Go.

<!-- markdownlint-disable MD046 -->
## Configuration {#config}

### Collector Configuration {#input-config}

=== "Host Installation"

    Go to the `conf.d/opentelemetry` directory under the DataKit installation directory, copy `opentelemetry.conf.sample` and name it `opentelemetry.conf`. Examples are as follows:

    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```

    Once configured, [Restart DataKit](../datakit/datakit-service-how-to.md#manage-service).

=== "Kubernetes"

    Can be turned on by [ConfigMap Injection Collector Configuration](../datakit/datakit-daemonset-deploy.md#configmap-setting) or [Config ENV_DATAKIT_INPUTS](../datakit/datakit-daemonset-deploy.md#env-setting) .

    Can also be turned on by environment variables, (needs to be added as the default collector in ENV_DEFAULT_ENABLED_INPUTS):
    
{{ CodeBlock .InputENVSample 4 }}

<!-- markdownlint-enable -->

### Notes {#attentions}

1. It is recommended to use grpc protocol, which has the advantages of high compression ratio, fast serialization and higher efficiency.
2. The route of the http protocol is configurable and the default request path is trace: `/otel/v1/traces`, metric:`/otel/v1/metrics`,logs:`/otel/v1/logs`
3. When data of type `float` `double` is involved, a maximum of two decimal places are reserved.
4. Both http and grpc support the gzip compression format. You can configure the environment variable in exporter to turn it on: `OTEL_EXPORTER_OTLP_COMPRESSION = gzip`; gzip is not turned on by default.
5. The http protocol request format supports both JSON and Protobuf serialization formats. But grpc only supports Protobuf.

<!-- markdownlint-disable MD046 -->
???+ warning

    - The service name in the DDTrace is named based on the service name or the referenced third-party library, while the service name of the OTEL collector is defined according to `otel.service.name`.
    - To display service names separately, a field configuration has been added: `spilt_service_name = true`.
    - The service name is extracted from the label of the link data. For example, if the label of the DB type is `db.system=mysql`, then the service name is `mysql`. If it is the MQ type: `messaging.system=kafka`, then the service name is `kafka`.
    - By default, the following three tags are extracted: "db.system/rpc.system/messaging.system`.
<!-- markdownlint-enable -->

Pay attention to the configuration of environment variables when using OTEL HTTP exporter. Since the default configuration of DataKit is `/otel/v1/traces` and `/otel/v1/metrics`,
if you want to use the HTTP protocol, you need to configure `trace` and `trace` separately `metric`,

The default request routes of OTLP are `/otel/v1/logs` `v1/traces` and `v1/metrics`, which need to be configured separately for these two. If you modify the routing in the configuration file, just replace the routing address below.

## Agent V2 version {#v2}

The default OTLP protocol has been changed from `grpc` to `http/protobuf` in order to align with the specification.
You can switch to the `grpc` protocol using `OTEL_EXPORTER_OTLP_PROTOCOL=grpc` or `-Dotel.exporter.otlp.protocol=grpc`.

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

Use gPRC:

```shell
java -javaagent:/usr/local/ddtrace/opentelemetry-javaagent-2.5.0.jar \
  -Dotel.exporter=otlp \
  -Dotel.exporter.otlp.protocol=grpc \
  -Dotel.exporter.otlp.endpoint=http://localhost:4317
  -Dotel.service.name=app \
  -jar app.jar
```

The default log is enabled. If you want to turn off log collection, the exporter configuration can be empty: `-Dotel.logs.exporter=none`

For more major changes in the V2 version, please check the official documentation or [GitHub GuanCe Cloud](https://github.com/GuanceCloud/opentelemetry-java-instrumentation/releases/tag/v2.11.0-guance){:target="_blank"} version notes


## General SDK Configuration {#sdk-configuration}

| ENV                           | Command                       | doc                                                     | default                 | note                                                                                                         |
|:------------------------------|:------------------------------|:--------------------------------------------------------|:------------------------|:-------------------------------------------------------------------------------------------------------------|
| `OTEL_SDK_DISABLED`           | `otel.sdk.disabled`           | Disable the SDK for all signals                         | false                   | Boolean value. If “true”, a no-op SDK implementation will be used for all telemetry signals                  |
| `OTEL_RESOURCE_ATTRIBUTES`    | `otel.resource.attributes`    | Key-value pairs to be used as resource attributes       |                         |                                                                                                              |
| `OTEL_SERVICE_NAME`           | `otel.service.name`           | Sets the value of the `service.name` resource attribute |                         | If `service.name` is also provided in `OTEL_RESOURCE_ATTRIBUTES`, then `OTEL_SERVICE_NAME` takes precedence. |
| `OTEL_LOG_LEVEL`              | `otel.log.level`              | Log level used by the SDK logger                        | `info`                  |                                                                                                              |
| `OTEL_PROPAGATORS`            | `otel.propagators`            | Propagators to be used as a comma-separated list        | `tracecontext,baggage`  | Values MUST be deduplicated in order to register a `Propagator` only once.                                   |
| `OTEL_TRACES_SAMPLER`         | `otel.traces.sampler`         | Sampler to be used for traces                           | `parentbased_always_on` |                                                                                                              |
| `OTEL_TRACES_SAMPLER_ARG`     | `otel.traces.sampler.arg`     | String value to be used as the sampler argument         | 1.0                     | 0 - 1.0                                                                                                      |
| `OTEL_EXPORTER_OTLP_PROTOCOL` | `otel.exporter.otlp.protocol` | `grpc`,`http/protobuf`,`http/json`                      | gRPC                    |                                                                                                              |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | `otel.exporter.otlp.endpoint` | OTLP Addr                                               | <http://localhost:4317> | <http://datakit-endpoint:9529/otel/v1/traces>                                                                |
| `OTEL_TRACES_EXPORTER`        | `otel.traces.exporter`        | Trace Exporter                                          | `otlp`                  |                                                                                                              |
| `OTEL_LOGS_EXPORTER`          | `otel.logs.exporter`          | Logging Exporter                                        | `otlp`                  | default disable                                                                                              |

> You can pass the 'otel.javaagent.debug=true' parameter to the agent to view debugging logs. Please note that these logs are quite lengthy and should be used with caution in production environments.

## Tracing {#tracing}

DataKit only accepts OTLP data. OTLP has clear data types: `gRPC`, `http/protobuf` and `http/json`. For specific configuration, please refer to:

```shell
# OpenTelemetry Agent default is gRPC
-Dotel.exporter=otlp \
-Dotel.exporter.otlp.protocol=grpc \
-Dotel.exporter.otlp.endpoint=http://datakit-endpoint:4317

# use http/protobuf
-Dotel.exporter=otlp \
-Dotel.exporter.otlp.protocol=http/protobuf \
-Dotel.exporter.otlp.traces.endpoint=http://datakit-endpoint:9529/otel/v1/traces \
-Dotel.exporter.otlp.metrics.endpoint=http://datakit-endpoint:9529/otel/v1/metrics 

# use http/json
-Dotel.exporter=otlp \
-Dotel.exporter.otlp.protocol=http/json \
-Dotel.exporter.otlp.traces.endpoint=http://datakit-endpoint:9529/otel/v1/traces \
-Dotel.exporter.otlp.metrics.endpoint=http://datakit-endpoint:9529/otel/v1/metrics
```

### Tag {#tag}

Starting from DataKit version [1.22.0](../datakit/changelog.md#cl-1.22.0) ,`ignore_tags` is deprecated.
Add a fixed tags, only those in this list will be extracted into the tag. The following is the fixed list:

| Attributes            | tag                   |
|:----------------------|:----------------------|
| http.url              | http_url              |
| http.hostname         | http_hostname         |
| http.route            | http_route            |
| http.status_code      | http_status_code      |
| http.request.method   | http_request_method   |
| http.method           | http_method           |
| http.client_ip        | http_client_ip        |
| http.scheme           | http_scheme           |
| url.full              | url_full              |
| url.scheme            | url_scheme            |
| url.path              | url_path              |
| url.query             | url_query             |
| span_kind             | span_kind             |
| db.system             | db_system             |
| db.operation          | db_operation          |
| db.name               | db_name               |
| db.statement          | db_statement          |
| server.address        | server_address        |
| net.host.name         | net_host_name         |
| server.port           | server_port           |
| net.host.port         | net_host_port         |
| network.peer.address  | network_peer_address  |
| network.peer.port     | network_peer_port     |
| network.transport     | network_transport     |
| messaging.system      | messaging_system      |
| messaging.operation   | messaging_operation   |
| messaging.message     | messaging_message     |
| messaging.destination | messaging_destination |
| rpc.service           | rpc_service           |
| rpc.system            | rpc_system            |
| error                 | error                 |
| error.message         | error_message         |
| error.stack           | error_stack           |
| error.type            | error_type            |
| error.msg             | error_message         |
| project               | project               |
| version               | version               |
| env                   | env                   |
| host                  | host                  |
| pod_name              | pod_name              |
| pod_namespace         | pod_namespace         |

If you want to add custom labels, you can use environment variables:

```shell
-Dotel.resource.attributes=username=myName,env=1.1.0
```

And modify the whitelist in the configuration file so that a custom label can appear in the first level label of the <<<custom_key.brand_name>>> link details.

```toml
customer_tags = ["sink_project", "username","env"]
```

### Kind {#kind}

All `Span` has `span_kind` tag,

- `unspecified`: unspecified.
- `internal`: internal span.
- `server`:  WEB server or RPC server.
- `client`:  HTTP client or RPC client.
- `producer`:  message producer.
- `consumer`:  message consumer.


### Best Practices {#bp}

DataKit currently provides [Go language](opentelemetry-go.md)、[Java](opentelemetry-java.md) languages, with other languages available later.

## Metric {#metric}

The OpenTelemetry Java Agent obtains the MBean's indicator information from the application through the JMX protocol, and the Java Agent reports the selected JMX indicator through the internal SDK, which means that all indicators are configurable.

You can enable and disable JMX metrics collection by command `otel.jmx.enabled=true/false`, which is enabled by default.

To control the time interval between MBean detection attempts, one can use the OTEL.jmx.discovery.delay property, which defines the number of milliseconds to elapse between the first and the next detection cycle.

In addition, the acquisition configuration of some third-party software built in the Agent. For details, please refer to: [JMX Metric Insight](https://github.com/open-telemetry/opentelemetry-java-instrumentation/blob/main/instrumentation/jmx-metrics/javaagent/README.md){:target="_blank"}

<!-- markdownlint-disable MD046 -->
???+ warning "metric"

    Starting from [DataKit 1.68.0](../datakit/changelog-2025.md#cl-1.68.0), the indicator set name has been changed:
    All indicators sent to the observation cloud have a unified indicator set name: `otel_service`.
    If you already have a dashboard, export the existing dashboard and change `otel-serivce` to `otel_service` and then import it.

<!-- markdownlint-enable -->

When exporting **Histogram** metrics to Observability Cloud, some metrics undergo special processing:

- OpenTelemetry histogram buckets are directly mapped to Prometheus histogram buckets.
- The count for each bucket is converted to Prometheus' cumulative count format.
- For example, OpenTelemetry buckets `[0, 10)`, `[10, 50)`, and `[50, 100)` are converted into Prometheus `_bucket` metrics with an `le` label:

```text
  my_histogram_bucket{le="10"} 100
  my_histogram_bucket{le="50"} 200
  my_histogram_bucket{le="100"} 250
```

- The total number of observations in the OpenTelemetry histogram is converted into the Prometheus `_count` metric.
- The sum of the OpenTelemetry histogram is converted into the Prometheus `_sum` metric, and `_max` and `_min` are also added.

```text
  my_histogram_count 250
  my_histogram_max 100
  my_histogram_min 50
  my_histogram_sum 12345.67
```

Any metric ending with `_bucket` is histogram data, and it will always have corresponding metrics ending with `_max`, `_min`, `_count`, and `_sum`.

In histogram data, the `le` (less or equal) label can be used for classification, and filtering can be performed based on labels. You can refer to [OpenTelemetry Metrics](https://opentelemetry.io/docs/specs/semconv/){:target="_blank"} for all metrics and labels.

This conversion enables seamless integration of OpenTelemetry-collected histogram data into Prometheus, leveraging Prometheus' powerful querying and visualization capabilities for analysis.

## Delete Metric Tags {#del-metric}

There are many useless tags in the indicators reported by OTEL. These are all of **String** type. They have been deleted because they occupy too much memory and bandwidth and will not be uploaded to the GuanCe cloud center.

These tags include:

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


{{ range $i, $m := .Measurements }}

### `{{$m.Name}}`

{{$m.Desc}}

- Tags

{{$m.TagsMarkdownTable}}

- Metrics

{{$m.FieldsMarkdownTable}}

{{ end }}

## Logging {#logging}

[:octicons-tag-24: Version-1.33.0](../datakit/changelog.md#cl-1.33.0)

“Standard output” LogRecord Exporter is a LogRecord Exporter which outputs the logs to stdout/console.

If a language provides a mechanism to automatically configure a LogRecordProcessor to pair with the associated exporter (e.g., using the `OTEL_LOGS_EXPORTER` environment variable),
by default the standard output exporter SHOULD be paired with a simple processor.

The `source` of the logs collected through OTEL is the `service.name`, and it can also be customized by adding tags such as `log.source`,
for example: `-Dotel.resource.attributes="log.source=sourcename"`.

You can [View logging documents](https://opentelemetry.io/docs/specs/otel/logs/sdk_exporters/stdout/){:target="_blank"}

By default, the maximum length of log content is 500kb, and the excess length will be divided into multiple logs.
The maximum length of the log label is 32KB. This field is not configurable, and the excess length will be cut off.

> Note: If the app is running in a container environment (such as k8s), [DataKit will automatically collect logs](container-log.md#logging-stdout){:target="_blank"}. If `otel` collects logs again, there will be a problem of duplicate collection.
> It is recommended to manually [turn off DataKit's autonomous log](container-log.md#logging-with-image-config){:target="_blank"} collection behavior before enabling `otel` to collect logs.

## More Docs {#more-readings}

- Go open source address [OpenTelemetry-go](https://github.com/open-telemetry/opentelemetry-go){:target="_blank"}
- Official user manual: [opentelemetry-io-docs](https://opentelemetry.io/docs/){:target="_blank"}
- Environment variable configuration: [sdk-extensions](https://github.com/open-telemetry/opentelemetry-java/blob/main/sdk-extensions/autoconfigure/README.md#otlp-exporter-both-span-and-metric-exporters){:target="_blank"}
- GitHub GuanceCloud version [OpenTelemetry-Java-instrumentation](https://github.com/GuanceCloud/opentelemetry-java-instrumentation){:target="_blank"}
