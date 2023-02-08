
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
    
      [inputs.opentelemetry.expectedHeaders]
      # If header is configured, the request must carry the otherwise return status code 500
      ## Be used as a security check and must be all lowercase
      # ex_version = xxx
      # ex_name = xxx
      # ...
    
      ## grpc
      [inputs.opentelemetry.grpc]
      ## trace for grpc
      trace_enable = true
    
      ## metric for grpc
      metric_enable = true
    
      ## grpc listen addr
      addr = "127.0.0.1:4317"
    
      ## http
      [inputs.opentelemetry.http]
      ## if enable=true
      ## http path (do not edit):
      ##	trace : /otel/v1/trace
      ##	metric: /otel/v1/metric
      ## use as : http://127.0.0.1:9529/otel/v1/trace . Method = POST
      enable = true
      ## return to client status_ok_code :200/202
      http_status_ok = 200
    
    ```

    Once configured, [restart DataKit](datakit-service-how-to.md#manage-service) 即可。

=== "Kubernetes"

    The collector can now be turned on by [ConfigMap Injection Collector Configuration](datakit-daemonset-deploy.md#configmap-setting).

### Notes {#attentions}

1. It is recommended to use grpc protocol, which has the advantages of high compression ratio, fast serialization and higher efficiency.

1. The route of the http protocol is not configurable and the request path is trace: `/otel/v1/trace`, metric:`/otel/v1/metric`

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

### Best Practices {#bp}

Datakit currently provides [Go language](opentelemetry-go.md)、[Java](opentelemetry-java.md) languages, with other languages available later.

## More Docs {#more-readings}
- Go open source address [opentelemetry-go](https://github.com/open-telemetry/opentelemetry-go){:target="_blank"}
- Official user manual: [opentelemetry-io-docs](https://opentelemetry.io/docs/){:target="_blank"}
- Environment variable configuration: [sdk-extensions](https://github.com/open-telemetry/opentelemetry-java/blob/main/sdk-extensions/autoconfigure/README.md#otlp-exporter-both-span-and-metric-exporters){:target="_blank"}
