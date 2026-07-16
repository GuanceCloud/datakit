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

OpenTelemetry (OTEL) is an observability project in CNCF.
This document explains how to collect OTEL data on DataKit and what you should configure for Java and Go scenarios.

## Configuration {#config}

<!-- markdownlint-disable MD046 -->
=== "Host Installation"

    Go to `conf.d/{{.Catalog}}` in the DataKit installation directory, copy `{{.InputName}}.conf.sample` to `{{.InputName}}.conf`, and edit it.

    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```

    Restart DataKit after configuration:
    [restart DataKit](../datakit/datakit-service-how-to.md#manage-service)

=== "Kubernetes"

    Configure the collector with ConfigMap or `ENV_DATAKIT_INPUTS`:
    [ConfigMap injection](../datakit/datakit-daemonset-deploy.md#configmap-setting)
    or
    [ENV configuration](../datakit/datakit-daemonset-deploy.md#env-setting).

    Or set environment variables. Add the collector to `ENV_DEFAULT_ENABLED_INPUTS` first:

{{ CodeBlock .InputENVSample 4 }}
<!-- markdownlint-enable MD046 -->

The `customer_tags` field supports regular-expression matching. Use a fixed prefix `reg:` for regex values, for example `reg:key_*`.

### Notes {#attentions}

1. We recommend gRPC for OTEL ingestion by default because it has better compression and lower overhead.
2. Since DataKit [1.10.0](../datakit/changelog.md#cl-1.10.0), OTEL HTTP routes are configurable. Defaults:
   - traces: `/otel/v1/traces`
   - metrics: `/otel/v1/metrics`
   - logs: `/otel/v1/logs`
3. For `float`/`double` values, precision is kept to two decimal places.
4. Both HTTP and gRPC support gzip compression. Enable it with exporter-side config, for example `OTEL_EXPORTER_OTLP_COMPRESSION=gzip`.
5. HTTP requests support both JSON and Protobuf in clients, but DataKit HTTP handlers parse Protobuf only.

<!-- markdownlint-disable MD046 -->
???+ warning

    - DDTrace service names are based on DDTrace service tags or referenced libraries.
    - OTEL service names are determined by `otel.service.name`.
    - To split service names by resource type (for example `db.system`, `rpc.system`, `messaging.system`), enable:

      `split_service_name = true`
    - The default priority is to derive service name from `db.system`, then `rpc.system`, then `messaging.system` when the split mode is enabled.

<!-- markdownlint-enable MD046 -->

If you are using OTEL HTTP exporter, configure the endpoint paths explicitly in DataKit:
traces `/otel/v1/traces`, metrics `/otel/v1/metrics`, logs `/otel/v1/logs` (default port is 9529).

### Java Agent V2 protocol behavior {#v2}

In OTEL Java Agent V2, default OTLP protocol is `http/protobuf`.
To keep compatibility, you can still switch back to gRPC:

```shell
java -javaagent:/usr/local/ddtrace/opentelemetry-javaagent-2.5.0.jar \
  -Dotel.exporter=otlp \
  -Dotel.exporter.otlp.protocol=grpc \
  -Dotel.exporter.otlp.endpoint=http://localhost:4317 \
  -Dotel.service.name=app \
  -jar app.jar
```

For HTTP mode (DataKit default path), configure each exporter endpoint:

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

Disable OTEL logs when not needed:

`-Dotel.logs.exporter=none`

For V2 release notes, see: [GitHub-v2.0.0](https://github.com/open-telemetry/opentelemetry-java-instrumentation/releases/tag/v2.0.0){:target="_blank"}

### Common commands {#sdk-configuration}

The table below is a practical subset of configuration items used for DataKit ingestion.

| ENV (System Property) | Description |
| --- | --- |
| `OTEL_SDK_DISABLED(otel.sdk.disabled)` | Disable entire SDK (`false` by default). |
| `OTEL_RESOURCE_ATTRIBUTES(otel.resource.attributes)` | Add global resource tags, e.g. `service.name=app,project=app-a`. |
| `OTEL_SERVICE_NAME(otel.service.name)` | Override service name, priority higher than resource tags. |
| `OTEL_LOG_LEVEL(otel.log.level)` | SDK log level (`info` by default). |
| `OTEL_PROPAGATORS(otel.propagators)` | Propagation format (`tracecontext,baggage` by default). |
| `OTEL_TRACES_SAMPLER(otel.traces.sampler)` | Sampling strategy. |
| `OTEL_TRACES_SAMPLER_ARG(otel.traces.sampler.arg)` | Sampler arguments, default `1.0` (`0~1.0`). |
| `OTEL_EXPORTER_OTLP_PROTOCOL(otel.exporter.otlp.protocol)` | Transport protocol, default `grpc`; supported `grpc` and `http/protobuf`. |
| `OTEL_EXPORTER_OTLP_ENDPOINT(otel.exporter.otlp.endpoint)` | General OTLP endpoint, e.g. `http://datakit-host:4317` (gRPC) or `http://datakit-host:9529` (HTTP base). |
| `OTEL_EXPORTER_OTLP_TRACES_ENDPOINT(otel.exporter.otlp.traces.endpoint)` | Trace endpoint for HTTP mode, e.g. `http://datakit-host:9529/otel/v1/traces`. |
| `OTEL_EXPORTER_OTLP_METRICS_ENDPOINT(otel.exporter.otlp.metrics.endpoint)` | Metrics endpoint for HTTP mode, e.g. `http://datakit-host:9529/otel/v1/metrics`. |
| `OTEL_EXPORTER_OTLP_LOGS_ENDPOINT(otel.exporter.otlp.logs.endpoint)` | Logs endpoint for HTTP mode, e.g. `http://datakit-host:9529/otel/v1/logs`. |
| `OTEL_TRACES_EXPORTER(otel.traces.exporter)` | Trace exporter type; `otlp` by default. |
| `OTEL_LOGS_EXPORTER(otel.logs.exporter)` | Log exporter type; set to `otlp` explicitly when logs are needed. |
| `OTEL_METRICS_EXPORTER(otel.metrics.exporter)` | Metric exporter type; set to `otlp` explicitly for OTEL metric ingestion. |

Starting from DataKit [1.85.0](../datakit/changelog.md#cl-1.85.0), `http/json` is removed from supported DataKit HTTP content types. Use `http/protobuf` only.

Pass `otel.javaagent.debug=true` to print Java Agent debug logs (use with caution in production).

### Trace sampling {#sample}

Choose head-based or tail-based sampling.

- Tail-based sampling (collector side): [OpenTelemetry Sampling Best Practices](../best-practices/cloud-native/opentelemetry-simpling.md)
- Head-based sampling (agent side): [OpenTelemetry Java Agent Sampling Strategy](../best-practices/cloud-native/otel-agent-sampling.md)

#### Tag extraction {#tags}

From DataKit [1.22.0](../datakit/changelog.md#cl-1.22.0), fixed tags are extracted by default (blacklist mode was removed).

| Attributes | Tags | Description |
|:---|:---|:---|
| `http.url` | `http_url` | Full request URL |
| `http.hostname` | `http_hostname` | Request hostname |
| `http.route` | `http_route` | HTTP route |
| `http.status_code` | `http_status_code` | HTTP status code |
| `http.request.method` | `http_request_method` | HTTP request method |
| `http.method` | `http_method` | Same as above |
| `http.client_ip` | `http_client_ip` | Client IP |
| `http.scheme` | `http_scheme` | Request protocol |
| `url.full` | `url_full` | Full request URL |
| `url.scheme` | `url_scheme` | URL scheme |
| `url.path` | `url_path` | Request path |
| `url.query` | `url_query` | Query string |
| `span_kind` | `span_kind` | Span role |
| `db.system` | `db_system` | DB system |
| `db.operation` | `db_operation` | DB operation |
| `db.name` | `db_name` | Database name |
| `db.statement` | `db_statement` | SQL statement |
| `server.address` | `server_address` | Service host |
| `net.host.name` | `net_host_name` | Host name |
| `server.port` | `server_port` | Service port |
| `net.host.port` | `net_host_port` | Host port |
| `network.peer.address` | `network_peer_address` | Peer host |
| `network.peer.port` | `network_peer_port` | Peer port |
| `network.transport` | `network_transport` | Network protocol |
| `messaging.system` | `messaging_system` | Message queue type |
| `messaging.operation` | `messaging_operation` | Message operation |
| `messaging.message` | `messaging_message` | Message details |
| `messaging.destination` | `messaging_destination` | Message destination |
| `rpc.service` | `rpc_service` | RPC service name |
| `rpc.system` | `rpc_system` | RPC system |
| `error` | `error` | Whether error |
| `error.message` | `error_message` | Error message |
| `error.stack` | `error_stack` | Error stack |
| `error.type` | `error_type` | Error type |
| `project` | `project` | Project tag |
| `version` | `version` | Version |
| `env` | `env` | Environment |
| `host` | `host` | Host tag |
| `pod_name` | `pod_name` | Pod name |
| `pod_namespace` | `pod_namespace` | Pod namespace |
| `telemetry.sdk.language` | `sdk_language` | SDK language |
| `telemetry.sdk.name` | `sdk_name` | SDK name |
| `telemetry.sdk.version` | `sdk_version` | SDK version |

To add custom resource tags:

```shell
-Dotel.resource.attributes=service.name=app,version=1.1.0,env=prod
```

##### Span kind {#kind}

- `unspecified`: not set
- `internal`: internal span
- `server`: server side
- `client`: client side
- `producer`: message producer
- `consumer`: message consumer

### Metrics {#metric}

Java Instrumentation can report JVM and JMX metrics through its built-in metrics Pipeline.

- Control JMX collection with `otel.jmx.enabled=true/false` (default is enabled).
- Adjust scan interval by `otel.jmx.discovery.delay` (milliseconds).

For built-in JMX extension details, see:
[GitHub OTEL JMX Metrics](https://github.com/open-telemetry/opentelemetry-java-instrumentation/blob/main/instrumentation/jmx-metrics/javaagent/README.md){:target="_blank"}

### Histogram conversion {#histogram-conversion}

OTEL histogram metrics are converted to Prometheus style metrics:

Input buckets:

```text
[0, 10), [10, 50), [50, 100)
```

Converted metrics:

```text
my_histogram_bucket{le="10"} 100
my_histogram_bucket{le="50"} 200
my_histogram_bucket{le="100"} 250
```

And DataKit additionally generates:

```text
my_histogram_count 250
my_histogram_max 100
my_histogram_min 50
my_histogram_sum 12345.67
```

Metrics ending with `_bucket` correspond to histogram buckets and should also include
`_count`, `_sum`, `_min`, `_max`.

### Log collection {#logging}

[:octicons-tag-24: Version-1.33.0](../datakit/changelog.md#cl-1.33.0)

DataKit receives OTEL logs through OTLP. Default is disabled in OTEL V1 and needs to be enabled explicitly.

```shell hl_lines='1 4 5 8'
# environment variable form
export OTEL_LOGS_EXPORTER=otlp
export OTEL_EXPORTER_OTLP_ENDPOINT=http://<DataKit Addr>:4317
export OTEL_EXPORTER_OTLP_PROTOCOL=grpc
java -jar app.jar

# command line form
java -javaagent:/path/to/agent.jar \
  -Dotel.logs.exporter=otlp \
  -Dotel.exporter.otlp.endpoint=http://<DataKit Addr>:4317 \
  -Dotel.exporter.otlp.protocol=grpc \
  -jar app.jar
```

For HTTP protocol in OTEL V2, use `-Dotel.exporter.otlp.protocol=http/protobuf` and set
`-Dotel.exporter.otlp.logs.endpoint=http://<DataKit Addr>:9529/otel/v1/logs`.

Default maximum `message` field size is 500KB. Can not parts above the limit are truncated.
