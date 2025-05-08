---
title     : 'JVM'
summary   : '采集 JVM 的指标数据'
tags:
  - 'JAVA'
__int_icon      : 'icon/jvm'
dashboard :
  - desc  : 'JVM'
    path  : 'dashboard/zh/jvm'
monitor   :
  - desc  : '暂无'
    path  : '-'
---


{{.AvailableArchs}}

---

这里我们提供俩类 JVM 指标采集方式，一种是 DDTrace，另一种方案是 Jolokia（已弃用），选择方式的建议如下：

- 推荐使用 DDTrace 进行采集 JVM 指标，Jolokia 也是可以的，用起来比较麻烦所以不推荐使用。
- 如果采集自己开发的 Java 应用 JVM 指标，推荐 DDTrace 方案，除了能采集 JVM 指标外，还能实现链路追踪（APM）数据采集。

## 配置 {#config}

### 通过 DDTrace 采集 JVM 指标 {#jvm-ddtrace}

DataKit 内置了 [StatsD 采集器](statsd.md)，用于接收网络上发送过来的 StatsD 协议的数据。此处我们利用 DDTrace 来采集 JVM 的指标数据，并通过 StatsD 协议发送给 Datakit。

<!-- markdownlint-disable MD046 -->
=== "主机安装"

    这里推荐使用如下的 StatsD 配置来采集 DDTrace JVM 指标。将其拷贝到 `conf.d/statsd` 目录下，并命名为 `ddtrace-jvm-statsd.conf`：

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
    
      # 以下配置无需关注
    
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

    目前可以通过 [ConfigMap 方式注入采集器配置](../datakit/datakit-daemonset-deploy.md#configmap-setting)来开启采集器。
<!-- markdownlint-enable -->

---

关于这里的配置说明：

- `service_address` 此处设置成 `:8125`，指 DDTrace 将 jvm 指标发送出来的目标地址
- `drop_tags` 此处我们将 `runtime-id` 丢弃，因为其可能导致时间线爆炸。如确实需要该字段，将其从 `drop_tags` 中移除即可
- `metric_mapping` 在 ddtrace 发送出来的原始数据中，有俩类指标，它们的指标名称分别以 `jvm_` 和 `datadog_tracer_` 开头，故我们将它们统一规约到俩类指标集中，一个是 `jvm`，一个是 `ddtrace` 自身运行指标

### 启动 Java 应用 {#start-app}

一种可行的 JVM 部署方式如下：

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

注意：

- 关于 `dd-java-agent.jar` 包的下载，参见 [这里](ddtrace.md)
- 建议给如下几个字段命名：
    - `service.name` 用于表示该 JVM 数据来自哪个应用
    - `env` 用于表示该 JVM 数据来自某个应用的哪个环境（如 `prod/testing/preprod` 等）

- 此处几个选项的意义：
    - `-Ddd.jmxfetch.check-period` 表示采集频率，单位为毫秒
    - `-Ddd.jmxfetch.statsd.host=127.0.0.1` 表示 Datakit 上 StatsD 采集器的连接地址
    - `-Ddd.jmxfetch.statsd.port=8125` 表示 DataKit 上 StatsD 采集器的 UDP 连接端口，默认为 8125
    - `-Ddd.trace.health.xxx` DDTrace 自身指标数据采集和发送设置
    - 如果要开启链路追踪（APM）可追加如下两个参数（DataKit HTTP 地址）
        - `-Ddd.agent.host=localhost`
        - `-Ddd.agent.port=9529`

开启后，就能采集到 DDTrace 暴露出来的 jvm  指标。

<!-- markdownlint-disable MD046 -->
???+ attention

    实际采集到的指标，以 [DataDog 的文档](https://docs.datadoghq.com/tracing/metrics/runtime_metrics/java/#data-collected){:target="_blank"} 为准。
<!-- markdownlint-enable -->

## 指标 {#metric}

- 标签

其中每个指标有如下 tags （实际 tags 受 Java 启动参数以及 StatsD 配置影响）

| 标签名        | 描述          |
| ----          | --------      |
| `env`         | 对应 `DD_ENV` |
| `host`        | 主机名        |
| `instance`    | 实例          |
| `jmx_domain`  |               |
| `metric_type` |               |
| `name`        |               |
| `service`     |               |
| `type`        |               |
| `version`     |               |

- 指标列表

| 指标                        | 描述                                                                                                                          | 数据类型 | 单位   |
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


重点解释一下以下几个指标：`gc_major_collection_count` `gc_minor_collection_count` `gc_major_collection_time` `gc_minor_collection_time`:

指标类型是 `counter` 也就是计数器，在采集过程中每次采集到指标后会和上一次的结果相减，并除以时间，也就是说 这些指标就是每秒的变化速率，并不是实际 `JVM` 中 `MBean` 中的值。

### 通过 Jolokia 采集 JVM 指标 {#jvm-jolokia}

JVM 采集器可以通过 JMX 来采取很多指标，并将指标采集到<<<custom_key.brand_name>>>，帮助分析 Java 运行情况。

### 配置 {#jolokia-config}

### 前置条件 {#jolokia-requirements}

安装或下载 [Jolokia](https://search.maven.org/remotecontent?filepath=org/jolokia/jolokia-jvm/1.6.2/jolokia-jvm-1.6.2-agent.jar){:target="_blank"}。DataKit 安装目录下的 `data` 目录中已经有下载好的 Jolokia jar 包。通过如下方式开启 Java 应用：

```shell
java -javaagent:/path/to/jolokia-jvm-agent.jar=port=8080,host=localhost -jar your_app.jar
```

已测试的版本：

- [x] JDK 20
- [x] JDK 17
- [x] JDK 11
- [x] JDK 8

进入 DataKit 安装目录下的 `conf.d/{{.Catalog}}` 目录，复制 `{{.InputName}}.conf.sample` 并命名为 `{{.InputName}}.conf`。示例如下：

```toml
{{.InputSample}}
```

配置好后，重启 DataKit 即可。

### Jolokia 指标 {#jolokia-metric}

以下所有数据采集，默认会追加名为 `host` 的全局 tag（tag 值为 DataKit 所在主机名），也可以在配置中通过 `[inputs.{{.InputName}}.tags]` 指定其它标签：

``` toml
 [inputs.{{.InputName}}.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
  # ...
```

{{ range $i, $m := .Measurements }}

#### `{{$m.Name}}`

- 标签

{{$m.TagsMarkdownTable}}

- 指标列表

{{$m.FieldsMarkdownTable}}

{{ end }}

## 延伸阅读 {#more-readings}

- [DDTrace Java 示例](ddtrace-java.md)
- [SkyWalking](skywalking.md)
- [OpenTelemetry Java 示例](opentelemetry-java.md)
