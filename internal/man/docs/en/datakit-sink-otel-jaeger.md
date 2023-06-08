# OpenTelemetry/Jaeger
---

OpenTelemetry (OTEL) provides a variety of Export to send link data to multiple collection terminals, such as Jaeger, otlp, zipkin and prometheus.

This article describes how to use sink to send link data to the otel-collector and Jaeger.

## Specify Sink Type Through Configuration File {#config}

### Send Link Data to Otel-collector {#apm-otel}

1. Modify the configuration datakit configuration file

``` shell 
vim /usr/local/datakit/conf/datakit.conf
```

2. Modify the sink configuration. Note that if sink correlation has never been configured, you can add a configuration item.

Otel has two types of export: http and grpc. You can only choose one of them.

http configuration:

``` toml
[sinks]
  [[sinks.sink]]
    scheme = "http"
    host = "localhost"
    port = "8889"
    path = "/api/traces"
    categories = ["T"] # only Trace is supported for mow
    target = "otel"
```


grpc configuration:

``` toml
[sinks]
  [[sinks.sink]]
    scheme = "grpc"
    host = "localhost"
    port = "4317"
    # When using grpc protocol, path is not required
    path = ""
    # Only Trace types are currently supported
    categories = ["T"] 
    target = "otel"
```

### Send Link Data to Jaeger {#apm-jaeger}

Sink Jaeger supports sending link data to `jaeger.colletcor` and `jaeger.agent`. Both `HTTP` and `gRPC` protocols are supported.

Colletcor port finishing

- 14267 tcp agent sends jaeger.thrift format data
- 14250 tcp agent sends proto format data (behind gRPC)
- 14268 HTTP accepts client data directly (datakit sends to collector using HTTP)
- 14269 http health check

Agent port collation

- 5775 UDP protocol for receiving zipkin-compatible protocol data
- 6831 UDP protocol that receives jaeger-compliant protocols (datakit uses gRPC to send link data)
- 6832 UDP protocol, binary protocol for receiving jaeger
- 5778 HTTP protocol, not recommended for large data volume

Sample datakit config file configuration:

Open configuration file 

``` shell 
vim /usr/local/datakit/conf/datakit.conf
```

HTTP configuration

``` toml
[sinks]
  [[sinks.sink]]
    scheme = "http"
    host = "localhost"
    port = "14268"
    path = "/api/traces"
    # Currently, only Trace type is supported, so "T" is used
    categories = ["T"] 
    target = "jaeger"
```

grpc configuration

``` toml
[sinks]
  [[sinks.sink]]
    scheme = "grpc"
    host = "localhost"
    port = "6831"
    # When using grpc protocol, path is not required
    path = ""
    # Only Trace types are currently supported前仅支持 Trace 类型  
    categories = ["T"] 
    target = "jaeger"
```

Restart datakit after configuration is complete.

---

## Specifying Sink as an Environment Variable during Installation Phase {#install}

```shell
# jaeger-collector
DK_SINK_T="jaeger://localhost?scheme=http&port=14268" \
DK_DATAWAY="https://openway.guance.com?token=<YOUR-TOKEN>" \
bash -c "$(curl -L https://static.guance.com/datakit/community/install.sh)"
```

Datakit installed through environment variables automatically generates the corresponding configuration in the configuration file, which will prevail when the service restarts later.
