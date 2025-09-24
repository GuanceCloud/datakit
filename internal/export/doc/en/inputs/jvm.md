---
title     : 'JVM'
summary   : 'Collect the JVM metrics'
tags:
  - 'JAVA'
__int_icon      : 'icon/jvm'
dashboard :
  - desc  : 'JVM'
    path  : 'dashboard/en/jvm'
monitor   :
  - desc  : 'N/A'
    path  : '-'
---


{{.AvailableArchs}}

---

Here, we provide two kinds of JVM metrics collection methods, one is Jolokia (deprecated) and the other is ddtrace. How to choose the way, we have the following suggestions:

- It is recommended to use DDTrace to collect JVM metrics, and Jolokia is also acceptable as it is more cumbersome to use, so it is not recommended.
- If we collect the JVM metrics of our own Java application, we recommend ddtrace scheme, which can collect the JVM metrics as well as link tracing (APM) data.

## Config {#config}

## Collect JVM Metrics Through Ddtrace {#jvm-ddtrace}

DataKit has a built-in [statsd collector](statsd.md) for receiving statsd protocol data sent over the network. Here we use ddtrace to collect metrics from the JVM and send them to the DataKit via statsd protocol.

### Collector Configuration {#input-config}

<!-- markdownlint-disable MD046 -->
=== "Host Installation"

    The following statsd configuration is recommended for collecting ddtrace JVM metrics. Copy it to the `conf.d/statsd` directory and name it `ddtrace-jvm-statsd.conf`:
    
    ```toml
    [[inputs.statsd]]
      protocol = "udp"
    
      ## Address and port to host UDP listener on
      service_address = ":8125"
    
      ## separator to use between elements of a statsd metric
      metric_separator = "_"
    
      drop_tags = ["runtime-id"]
      metric_mapping = [
        "jvm_:jvm",
        "datadog_tracer_:ddtrace",
      ]
    
      # There is no need to pay attention to the following configurations...
    
      delete_gauges = true
      delete_counters = true
      delete_sets = true
      delete_timings = true
    
      ## Percentiles to calculate for timing & histogram stats
      percentiles = [50.0, 90.0, 99.0, 99.9, 99.95, 100.0]
    
      ## Parses tags in the datadog statsd format
      ## http://docs.datadoghq.com/guides/dogstatsd/
      parse_data_dog_tags = true
    
      ## Parses datadog extensions to the statsd format
      datadog_extensions = true
    
      ## Parses distributions metric as specified in the datadog statsd format
      ## https://docs.datadoghq.com/developers/metrics/types/?tab=distribution#definition
      datadog_distributions = true
    
      ## Number of UDP messages allowed to queue up, once filled,
      ## the statsd server will start dropping packets
      allowed_pending_messages = 10000
    
      ## Number of timing/histogram values to track per-measurement in the
      ## calculation of percentiles. Raising this limit increases the accuracy
      ## of percentiles but also increases the memory usage and cpu time.
      percentile_limit = 1000
    
      ## Max duration (TTL) for each metric to stay cached/reported without being updated.
      #max_ttl = "1000h"
    
      [inputs.statsd.tags]
      # some_tag = "your-tag-value"
      # some_other_tag = "your-other-tag-value"
    ```

=== "Kubernetes"

    The collector can now be turned on by [ConfigMap injection collector configuration](../datakit/datakit-daemonset-deploy.md#configmap-setting).
<!-- markdownlint-enable -->

---

For configuration instructions here:

- `service_address` set here to `:8125`, which is the destination address where ddtrace sends out jvm metrics.
- `drop_tags` here discards `runtime-id` here because it could cause the timeline to explode. If you really need this field, just remove it from `drop_tags`.
- `metric_mapping`: In the original data sent by ddtrace, there are two types of metrics, their metrics names begin with `jvm_` and `datadog_tracer_` respectively, so we unify them into two types of metrics, one is `jvm` and the other is `ddtrace` self-running metrics.

### Start Java Application {#start-app}

A feasible JVM deployment method is as follows:

```shell
java -javaagent:dd-java-agent.jar \
    -Ddd.profiling.enabled=true \
    -Ddd.logs.injection=true \
    -Ddd.trace.sample.rate=1 \
    -Ddd.service.name=my-app \
    -Ddd.env=staging \
    -Ddd.agent.host=localhost \
    -Ddd.agent.port=9529 \
    -Ddd.jmxfetch.enabled=true \
    -Ddd.jmxfetch.check-period=1000 \
    -Ddd.jmxfetch.statsd.host=127.0.0.1  \
    -Ddd.jmxfetch.statsd.port=8125 \
    -Ddd.version=1.0 \
    -jar your-app.jar
```

Note:

- For the download of the `dd-java-agent.jar` package, see [here](ddtrace.md)
- It is recommended to name the following fields:
    - `service.name` is used to indicate which application the JVM data comes from
    - `env` is used to indicate which environment of an application the JVM data comes from (e.g. `prod/test/preprod`, etc.)

- The meaning of several options here:
    - `-Ddd.jmxfetch.check-period` denotes the collection frequency, in milliseconds
    - `-Ddd.jmxfetch.statsd.host=127.0.0.1` indicates the connection address of the statsd collector on the DataKit
    - `-Ddd.jmxfetch.statsd.port=8125` indicates the UDP connection port for the statsd collector on the DataKit, which defaults to 8125
    - `-Ddd.trace.health.xxx` ddtrace own metrics data collection and sending settings
    - If you want to turn on link tracing (APM), you can append the following two parameters (DataKit HTTP address)
        - `-Ddd.agent.host=localhost`
        - `-Ddd.agent.port=9529`

When turned on, you can collect jvm metrics exposed by DDTrace.

<!-- markdownlint-disable MD046 -->
???+ info

    The actual collected indicators are based on [DataDog's doc](https://docs.datadoghq.com/tracing/metrics/runtime_metrics/java/#data-collected){:target="_blank"}.
<!-- markdownlint-enable -->

### Metric {#metric}

- Tags

Each metric has the following tags (the actual tags are affected by Java startup parameters and statsd configuration).

| Tag Name        | Description          |
| ----          | --------      |
| `env`         | corresponding `DD_ENV` |
| `host`        | hostname        |
| `instance`    | example          |
| `jmx_domain`  |               |
| `metric_type` |               |
| `name`        |               |
| `service`     |               |
| `type`        |               |
| `version`     |               |

- Metrics

| Metrics                        | Description                                                                                                                          | Data Type | Unit   |
| ----                        | ----                                                                                                                          | :---:    | :----: |
| `heap_memory`               | The total Java heap memory used                                                                                               | int      | B      |
| `heap_memory_committed`     | The total Java heap memory committed to be used                                                                               | int      | B      |
| `heap_memory_init`          | The initial Java heap memory allocated                                                                                        | int      | B      |
| `heap_memory_max`           | The maximum Java heap memory available                                                                                        | int      | B      |
| `non_heap_memory`           | The total Java non-heap memory used. Non-heap memory is calculated as follows: `Metaspace + CompressedClassSpace + CodeCache` | int      | B      |
| `non_heap_memory_committed` | The total Java non-heap memory committed to be used                                                                           | int      | B      |
| `non_heap_memory_init`      | The initial Java non-heap memory allocated                                                                                    | int      | B      |
| `non_heap_memory_max`       | The maximum Java non-heap memory available                                                                                    | int      | B      |
| `thread_count`              | The number of live threads                                                                                                    | int      | count  |
| `gc_cms_count`              | The total number of garbage collections that have occurred                                                                    | int      | count  |
| `gc_major_collection_count` | The number of major garbage collections that have occurred. Set `new_gc_metrics: true` to receive this metric                 | int      | count  |
| `gc_minor_collection_count` | The number of minor garbage collections that have occurred. Set `new_gc_metrics: true` to receive this metric                 | int      | count  |
| `gc_parnew_time`            | The approximate accumulated garbage collection time elapsed                                                                   | int      | ms     |
| `gc_major_collection_time`  | The approximate major garbage collection time elapsed. Set `new_gc_metrics: true` to receive this metric                      | int      | ms     |
| `gc_minor_collection_time`  | The approximate minor garbage collection time elapsed. Set `new_gc_metrics: true` to receive this metric                      | int      | ms     |

Focus on explaining the following indicators: `gc_major_collection_count` `gc_minor_collection_count` `gc_major_collection_time` `gc_minor_collection_time`:

The indicator type is composed of three components of `counter`. During the collection process, each time the indicator is collected, it will be subtracted from the previous result and divided by time.
These indicators are the rate of change per second, which is not actually the case. The value in the `MBean` in the `JVM`.

## Collect JVM Metrics Through Jolokia  {#jvm-jolokia}

JVM collector can take many metrics through JMX, and collect metrics into <<<custom_key.brand_name>>> to help analyze Java operation.

### Jolokia Config {#jolokia-config}

### Preconditions {#jolokia-requirements}

Install or download  [Jolokia](https://search.maven.org/remotecontent?filepath=org/jolokia/jolokia-jvm/1.6.2/jolokia-jvm-1.6.2-agent.jar){:target="_blank"}. The downloaded Jolokia jar package is already available in the `data` directory under the DataKit installation directory. Open the Java application by:

```shell
java -javaagent:/path/to/jolokia-jvm-agent.jar=port=8080,host=localhost -jar your_app.jar
```

Already tested version:

- [x] JDK 20
- [x] JDK 17
- [x] JDK 11
- [x] JDK 8

Go to the `conf.d/samples` directory under the DataKit installation directory, copy `{{.InputName}}.conf.sample` and name it `{{.InputName}}.conf`. Examples are as follows:

```toml
{{.InputSample}}
```

After configuration, restart DataKit.

### Jolokia Metric {#jolokia-metric}

For all the following data collections, a global tag named `host` is appended by default (the tag value is the host name of the DataKit), or other tags can be specified in the configuration by `[inputs.{{.InputName}}.tags]`:

``` toml
 [inputs.{{.InputName}}.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
  # ...
```

{{ range $i, $m := .Measurements }}

#### `{{$m.Name}}`

{{$m.MarkdownTable}}

{{ end }}

## More Readings {#more-readings}

- [DDTrace Java example](ddtrace-java.md)
- [SkyWalking](skywalking.md)
- [OpenTelemetry Java example](opentelemetry-java.md)
