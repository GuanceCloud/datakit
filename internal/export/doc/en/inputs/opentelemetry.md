
# OpenTelemetry
---

:fontawesome-brands-linux: :fontawesome-brands-windows: :fontawesome-brands-apple: :material-kubernetes: :material-docker:

---

OpenTelemetry (hereinafter referred to as OTEL) is an observability project of CNCF, which aims to provide a standardization scheme in the field of observability and solve the standardization problems of data model, collection, processing and export of observation data.

OTEL is a collection of standards and tools for managing observational data, such as trace, metrics, logs, etc. (new observational data types may appear in the future).

OTEL provides vendor-independent implementations that export observation class data to different backends, such as open source Prometheus, Jaeger, Datakit, or cloud vendor services, depending on the user's needs.

The purpose of this article is to introduce how to configure and enable OTEL data access on Datakit, and the best practices of Java and Go.

***Version Notes***: Datakit currently only accesses OTEL v1 version of otlp data.

<!-- markdownlint-disable MD046 -->
## Configuration {#config}

=== "Host Installation"

    Go to the `conf.d/opentelemetry` directory under the DataKit installation directory, copy `opentelemetry.conf.sample` and name it `opentelemetry.conf`. Examples are as follows:

    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```

    Once configured, [Restart DataKit](../datakit/datakit-service-how-to.md#manage-service).

=== "Kubernetes"

    The collector can now be turned on by [ConfigMap Injection Collector Configuration](../datakit/datakit-daemonset-deploy.md#configmap-setting).

    Multiple environment variables supported that can be used in Kubernetes showing below:

    | Envrionment Variable Name           | Type        | Example                                                                                                  |
    | ----------------------------------- | ----------- | -------------------------------------------------------------------------------------------------------- |
    | `ENV_INPUT_OTEL_IGNORE_TAGS`        | JSON string | `["block1", "block2"]`                                                                                   |
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

### Notes {#attentions}

1. It is recommended to use grpc protocol, which has the advantages of high compression ratio, fast serialization and higher efficiency.
2. The route of the http protocol is configurable and the default request path is trace: `/otel/v1/trace`, metric:`/otel/v1/metric`
3. When data of type `float` `double` is involved, a maximum of two decimal places are reserved.
4. Both http and grpc support the gzip compression format. You can configure the environment variable in exporter to turn it on: `OTEL_EXPORTER_OTLP_COMPRESSION = gzip`; gzip is not turned on by default.
5. The http protocol request format supports both json and protobuf serialization formats. But grpc only supports protobuf.

Pay attention to the configuration of environment variables when using OTEL HTTP exporter. Since the default configuration of datakit is `/otel/v1/trace` and `/otel/v1/metric`,
if you want to use the HTTP protocol, you need to configure `trace` and `trace` separately `metric`,

The default request routes of otlp are `v1/traces` and `v1/metrics`, which need to be configured separately for these two. If you modify the routing in the configuration file, just replace the routing address below.

## General SDK Configuration {#sdk-configuration}

| Command                       | doc                                                     | default                 | note                                                                                                         |
|:------------------------------|:--------------------------------------------------------|:------------------------|:-------------------------------------------------------------------------------------------------------------|
| `OTEL_SDK_DISABLED`           | Disable the SDK for all signals                         | false                   | Boolean value. If “true”, a no-op SDK implementation will be used for all telemetry signals                  |
| `OTEL_RESOURCE_ATTRIBUTES`    | Key-value pairs to be used as resource attributes       |                         |                                                                                                              |
| `OTEL_SERVICE_NAME`           | Sets the value of the `service.name` resource attribute |                         | If `service.name` is also provided in `OTEL_RESOURCE_ATTRIBUTES`, then `OTEL_SERVICE_NAME` takes precedence. |
| `OTEL_LOG_LEVEL`              | Log level used by the SDK logger                        | `info`                  |                                                                                                              |
| `OTEL_PROPAGATORS`            | Propagators to be used as a comma-separated list        | `tracecontext,baggage`  | Values MUST be deduplicated in order to register a `Propagator` only once.                                   |
| `OTEL_TRACES_SAMPLER`         | Sampler to be used for traces                           | `parentbased_always_on` |                                                                                                              |
| `OTEL_TRACES_SAMPLER_ARG`     | String value to be used as the sampler argument         | 1.0                     | 0 - 1.0                                                                                                      |
| `OTEL_EXPORTER_OTLP_PROTOCOL` | `grpc`,`http/protobuf`,`http/json`                      | gRPC                    |                                                                                                              |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | OTLP Addr                                               | <http://localhost:4317> | <http://datakit-endpoint:9529/otel/v1/trace>                                                                 |
| `OTEL_TRACES_EXPORTER`        | Trace Exporter                                          | `otlp`                  |                                                                                                              |

> You can pass the 'otel.javaagent.debug=true' parameter to the agent to view debugging logs. Please note that these logs are quite lengthy and should be used with caution in production environments.

## Tracing {#tracing}

Datakit only accepts OTLP data. OTLP has clear data types: `gRPC`, `http/protobuf` and `http/json`. For specific configuration, please refer to:

```shell
# OpenTelemetry Agent default is gRPC
-Dotel.exporter=otlp \
-Dotel.exporter.otlp.protocol=grpc \
-Dotel.exporter.otlp.endpoint=http://datakit-endpoint:4317

# use http/protobuf
-Dotel.exporter=otlp \
-Dotel.exporter.otlp.protocol=http/protobuf \
-Dotel.exporter.otlp.traces.endpoint=http://datakit-endpoint:9529/otel/v1/trace \
-Dotel.exporter.otlp.metrics.endpoint=http://datakit-endpoint:9529/otel/v1/metric 

# use http/json
-Dotel.exporter=otlp \
-Dotel.exporter.otlp.protocol=http/json \
-Dotel.exporter.otlp.traces.endpoint=http://datakit-endpoint:9529/otel/v1/trace \
-Dotel.exporter.otlp.metrics.endpoint=http://datakit-endpoint:9529/otel/v1/metric
```

### Best Practices {#bp}

Datakit currently provides [Go language](opentelemetry-go.md)、[Java](opentelemetry-java.md) languages, with other languages available later.

## Metric {#metric}

The OpenTelemetry Java Agent obtains the MBean's indicator information from the application through the JMX protocol, and the Java Agent reports the selected JMX indicator through the internal SDK, which means that all indicators are configurable.

You can enable and disable JMX metrics collection by command `otel.jmx.enabled=true/false`, which is enabled by default.

To control the time interval between MBean detection attempts, one can use the otel.jmx.discovery.delay property, which defines the number of milliseconds to elapse between the first and the next detection cycle.

In addition, the acquisition configuration of some third-party software built in the Agent. For details, please refer to: [JMX Metric Insight](https://github.com/open-telemetry/opentelemetry-java-instrumentation/blob/main/instrumentation/jmx-metrics/javaagent/README.md){:target="_blank"}

{{ range $i, $m := .Measurements }}

### `{{$m.Name}}`

{{$m.Desc}}

- tag

{{$m.TagsMarkdownTable}}

- metric list

{{$m.FieldsMarkdownTable}}

{{ end }}

## More Docs {#more-readings}
- Go open source address [opentelemetry-go](https://github.com/open-telemetry/opentelemetry-go){:target="_blank"}
- Official user manual: [opentelemetry-io-docs](https://opentelemetry.io/docs/){:target="_blank"}
- Environment variable configuration: [sdk-extensions](https://github.com/open-telemetry/opentelemetry-java/blob/main/sdk-extensions/autoconfigure/README.md#otlp-exporter-both-span-and-metric-exporters){:target="_blank"}
- GitHub GuanceCloud version [opentelemetry-java-instrumentation](https://github.com/GuanceCloud/opentelemetry-java-instrumentation){:target="_blank"}
