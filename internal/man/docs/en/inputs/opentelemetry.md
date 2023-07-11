
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

    Once configured, [Restart DataKit](datakit-service-how-to.md#manage-service).

=== "Kubernetes"

    The collector can now be turned on by [ConfigMap Injection Collector Configuration](datakit-daemonset-deploy.md#configmap-setting).

    Multiple environment variables supported that can be used in Kubernetes showing below:

    | Envrionment Variable Name              | Type        | Example                                                                                                  |
    | -------------------------------------- | ----------- | -------------------------------------------------------------------------------------------------------- |
    | `ENV_INPUT_OTEL_IGNORE_ATTRIBUTE_KEYS` | JSON string | `["os_*", "process_*"]`                                                                                  |
    | `ENV_INPUT_OTEL_KEEP_RARE_RESOURCE`    | bool        | true                                                                                                     |
    | `ENV_INPUT_OTEL_OMIT_ERR_STATUS`       | JSON string | `["404", "403", "400"]`                                                                                  |
    | `ENV_INPUT_OTEL_CLOSE_RESOURCE`        | JSON string | `{"service1":["resource1"], "service2":["resource2"], "service3":["resource3"]}`                         |
    | `ENV_INPUT_OTEL_SAMPLER`               | float       | 0.3                                                                                                      |
    | `ENV_INPUT_OTEL_TAGS`                  | JSON string | `{"k1":"v1", "k2":"v2", "k3":"v3"}`                                                                      |
    | `ENV_INPUT_OTEL_THREADS`               | JSON string | `{"buffer":1000, "threads":100}`                                                                         |
    | `ENV_INPUT_OTEL_STORAGE`               | JSON string | `{"storage":"./otel_storage", "capacity": 5120}`                                                         |
    | `ENV_INPUT_OTEL_HTTP`                  | JSON string | `{"enable":true, "http_status_ok": 200, "trace_api": "/otel/v1/trace", "metric_api": "/otel/v1/metric"}` |
    | `ENV_INPUT_OTEL_GRPC`                  | JSON string | `{"trace_enable": true, "metric_enable": true, "addr": "127.0.0.1:4317"}`                                |
    | `ENV_INPUT_OTEL_EXPECTED_HEADERS`      | JSON string | `{"ex_version": "1.2.3", "ex_name": "env_resource_name"}`                                                |

<!-- markdownlint-enable -->

### Notes {#attentions}

1. It is recommended to use grpc protocol, which has the advantages of high compression ratio, fast serialization and higher efficiency.
2. The route of the http protocol is configurable and the default request path is trace: `/otel/v1/trace`, metric:`/otel/v1/metric`
3. When data of type `float` `double` is involved, a maximum of two decimal places are reserved.
4. Both http and grpc support the gzip compression format. You can configure the environment variable in exporter to turn it on: `OTEL_EXPORTER_OTLP_COMPRESSION = gzip`; gzip is not turned on by default.
5. The http protocol request format supports both json and protobuf serialization formats. But grpc only supports protobuf.
6. The configuration field `ignore_attribute_keys` is to filter out some unwanted keys. But in OTEL, `attributes` are separated by `.` in most tags. For example, in the source code of resource:

```golang
ServiceNameKey = attribute.Key("service.name")
ServiceNamespaceKey = attribute.Key("service.namespace")
TelemetrySDKNameKey = attribute.Key("telemetry.sdk.name")
TelemetrySDKLanguageKey = attribute.Key("telemetry.sdk.language")
OSTypeKey = attribute.Key("os.type")
OSDescriptionKey = attribute.Key("os.description")
...
```

Therefore, if you want to filter all subtype tags under `teletemetry.sdk` and `os`, you should configure this:

``` toml
# When you create trace, Span, Resource, you add a lot of tags, and these tags will eventually appear in the Span
# If you don't want too many tags to cause unnecessary traffic loss on the network, you can choose to ignore these tags
# Support regular expression
# Note: Replace all '.' with '_'
ignore_attribute_keys = ["os_*","teletemetry_sdk*"]
```

Pay attention to the configuration of environment variables when using OTEL HTTP exporter. Since the default configuration of datakit is `/otel/v1/trace` and `/otel/v1/metric`,
if you want to use the HTTP protocol, you need to configure `trace` and `trace` separately `metric`,

The default request routes of otlp are `v1/traces` and `v1/metrics`, which need to be configured separately for these two. If you modify the routing in the configuration file, just replace the routing address below.

example:

```shell
java -javaagent:/usr/local/opentelemetry-javaagent-1.26.1-guance.jar \
 -Dotel.exporter=otlp \
 -Dotel.exporter.otlp.protocol=http/protobuf \
 -Dotel.exporter.otlp.traces.endpoint=http://localhost:9529/otel/v1/trace \
 -Dotel.exporter.otlp.metrics.endpoint=http://localhost:9529/otel/v1/metric \
 -jar tmall.jar

# If the default routes in the configuration file are changed to `v1/traces` and `v1/metrics`,
# then the above command can be written as follows:
java -javaagent:/usr/local/opentelemetry-javaagent-1.26.1-guance.jar \
 -Dotel.exporter=otlp \
 -Dotel.exporter.otlp.protocol=http/protobuf \
 -Dotel.exporter.otlp.endpoint=http://localhost:9529/ \
 -jar tmall.jar
```

### Best Practices {#bp}

Datakit currently provides [Go language](opentelemetry-go.md)、[Java](opentelemetry-java.md) languages, with other languages available later.

## Measurements {#measurements}

{{range $i, $m := .Measurements}}

{{if eq $m.Type "tracing"}}

### `{{$m.Name}}`

{{$m.Desc}}

- tag

{{$m.TagsMarkdownTable}}

- metric list

{{$m.FieldsMarkdownTable}}
{{end}}

{{end}}

## More Docs {#more-readings}
- Go open source address [opentelemetry-go](https://github.com/open-telemetry/opentelemetry-go){:target="_blank"}
- Official user manual: [opentelemetry-io-docs](https://opentelemetry.io/docs/){:target="_blank"}
- Environment variable configuration: [sdk-extensions](https://github.com/open-telemetry/opentelemetry-java/blob/main/sdk-extensions/autoconfigure/README.md#otlp-exporter-both-span-and-metric-exporters){:target="_blank"}
- GitHub GuanceCloud version [opentelemetry-java-instrumentation](https://github.com/GuanceCloud/opentelemetry-java-instrumentation){:target="_blank"}
