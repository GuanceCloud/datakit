
# OpenTelemetry
---

:fontawesome-brands-linux: :fontawesome-brands-windows: :fontawesome-brands-apple: :material-kubernetes: :material-docker:

---

OpenTelemetry (hereinafter referred to as OTEL) is an observability project of CNCF, which aims to provide a standardization scheme in the field of observability and solve the standardization problems of data model, collection, processing and export of observation data.

OTEL is a collection of standards and tools for managing observational data, such as trace, metrics, logs, etc. (new observational data types may appear in the future).

OTEL provides vendor-independent implementations that export observation class data to different backends, such as open source Prometheus, Jaeger, Datakit, or cloud vendor services, depending on the user's needs.

The purpose of this article is to introduce how to configure and enable OTEL data access on Datakit, and the best practices of Java and Go.

***Version Notes***: Datakit currently only accesses OTEL v1 version of otlp data.

## Configuration {#config}

=== "Host Installation"

    Go to the `conf.d/opentelemetry` directory under the DataKit installation directory, copy `opentelemetry.conf.sample` and name it `opentelemetry.conf`. Examples are as follows:
    
    ```toml
        
    [[inputs.opentelemetry]]
      ## When you create 'trace', 'Span', 'resource', you add a lot of tags that will eventually appear in 'Span'
      ## If you don't want too many of these tags to cause unnecessary traffic loss on the network, you can choose to ignore these tags
      ## Support regular expression, note: replace all '.' with '_'
      ## When creating 'trace', 'span' and 'resource', many labels will be added, and these labels will eventually appear in all 'spans'
      ## When you don't want too many labels to cause unnecessary traffic loss on the network, you can choose to ignore these labels
      ## Support regular expression. Note!!!: all '.' Replace with '_'
      # ignore_attribute_keys = ["os_*","process_*"]
    
      ## Keep rare tracing resources list switch.
      ## If some resources are rare enough(not presend in 1 hour), those resource will always send
      ## to data center and do not consider samplers and filters.
      # keep_rare_resource = false
    
      ## By default every error presents in span will be send to data center and omit any filters or
      ## sampler. If you want to get rid of some error status, you can set the error status list here.
      # omit_err_status = ["404"]
    
      ## Ignore tracing resources map like service:[resources...].
      ## The service name is the full service name in current application.
      ## The resource list is regular expressions uses to block resource names.
      ## If you want to block some resources universally under all services, you can set the
      ## service name as "*". Note: double quotes "" cannot be omitted.
      # [inputs.opentelemetry.close_resource]
        # service1 = ["resource1", "resource2", ...]
        # service2 = ["resource1", "resource2", ...]
        # "*" = ["close_resource_under_all_services"]
        # ...
    
      ## Sampler config uses to set global sampling strategy.
      ## sampling_rate used to set global sampling rate.
      # [inputs.opentelemetry.sampler]
        # sampling_rate = 1.0
    
      # [inputs.opentelemetry.tags]
        # key1 = "value1"
        # key2 = "value2"
        # ...
    
      ## Threads config controls how many goroutines an agent cloud start to handle HTTP request.
      ## buffer is the size of jobs' buffering of worker channel.
      ## threads is the total number fo goroutines at running time.
      # [inputs.opentelemetry.threads]
        # buffer = 100
        # threads = 8
    
      ## Storage config a local storage space in hard dirver to cache trace data.
      ## path is the local file path used to cache data.
      ## capacity is total space size(MB) used to store data.
      # [inputs.opentelemetry.storage]
        # path = "./otel_storage"
        # capacity = 5120
    
      ## OTEL agent HTTP config for trace and metrics
      ## If enable set to be true, trace and metrics will be received on path respectively, by default is:
      ## trace : /otel/v1/trace
      ## metric: /otel/v1/metric
      ## and the client side should be configured properly with Datakit listening port(default: 9529)
      ## or custom HTTP request path
      ## for example http://127.0.0.1:9529/otel/v1/trace
      ## The acceptable http_status_ok values will be 200 or 202.
      [inputs.opentelemetry.http]
        enable = true
        http_status_ok = 200
        trace_api = "/otel/v1/trace"
        metric_api = "/otel/v1/metric"
      
      ## OTEL agent GRPC config for trace and metrics.
      ## GRPC services for trace and metrics can be enabled respectively as setting either to be true.
      ## add is the listening on address for GRPC server.
      [inputs.opentelemetry.grpc]
        trace_enable = true
        metric_enable = true
        addr = "127.0.0.1:4317"
      
      ## If 'expectedHeaders' is well configed, then the obligation of sending certain wanted HTTP headers is on the client side,
      ## otherwise HTTP status code 400(bad request) will be provoked.
      ## Note: expectedHeaders will be effected on both trace and metrics if setted up.
      # [inputs.opentelemetry.expectedHeaders]
      #   ex_version = "1.2.3"
      #   ex_name = "env_resource_name"
      # ...

    
    ```

    Once configured, [restart DataKit](datakit-service-how-to.md#manage-service) 即可。

=== "Kubernetes"

    The collector can now be turned on by [ConfigMap Injection Collector Configuration](datakit-daemonset-deploy.md#configmap-setting).

### Notes {#attentions}

1. It is recommended to use grpc protocol, which has the advantages of high compression ratio, fast serialization and higher efficiency.

1. Since datakit version v1.10.0, The route of the http protocol is configurable and the request path is trace: `/otel/v1/trace`, metric:`/otel/v1/metric`

1. When data of type `float` `double` is involved, a maximum of two decimal places are reserved.

1. Both http and grpc support the gzip compression format. You can configure the environment variable in exporter to turn it on: `OTEL_EXPORTER_OTLP_COMPRESSION = gzip`; gzip is not turned on by default.
    
1. The http protocol request format supports both json and protobuf serialization formats. But grpc only supports protobuf.

1. The configuration field `ignore_attribute_keys` is to filter out some unwanted keys. But in OTEL, `attributes` are separated by `.` in most tags. For example, in the source code of resource:

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

## More Docs {#more-readings}
- Go open source address [opentelemetry-go](https://github.com/open-telemetry/opentelemetry-go){:target="_blank"}
- Official user manual: [opentelemetry-io-docs](https://opentelemetry.io/docs/){:target="_blank"}
- Environment variable configuration: [sdk-extensions](https://github.com/open-telemetry/opentelemetry-java/blob/main/sdk-extensions/autoconfigure/README.md#otlp-exporter-both-span-and-metric-exporters){:target="_blank"}
- GitHub GuanceCloud version [opentelemetry-java-instrumentation](https://github.com/GuanceCloud/opentelemetry-java-instrumentation){:target="_blank"}
