---
title     : 'Kafka'
summary   : 'Collect metrics of Kafka'
tags:
  - 'MIDDLEWARE'
  - 'MESSAGE QUEUES'
__int_icon      : 'icon/kafka'
dashboard :
  - desc  : 'Kafka'
    path  : 'dashboard/en/kafka'
monitor   :
  - desc  : 'Kafka'
    path  : 'monitor/en/kafka'
---


{{.AvailableArchs}}

---

Collect Kafka indicators and logs and report them to <<<custom_key.brand_name>>> to help you monitor and analyze various abnormal situations of Kafka.

## Configuration {#config}

### Requirements {#requirements}

Install or download [Jolokia](https://search.maven.org/remotecontent?filepath=org/jolokia/jolokia-jvm/1.6.2/jolokia-jvm-1.6.2-agent.jar){:target="_blank"}. The downloaded Jolokia jar package is already available in the `data` directory under the DataKit installation directory.

Jolokia is a Java agent of Kafka, which provides an external interface using JSON as data format based on HTTP protocol for DataKit to use. When Kafka starts, first configure the `KAFKA_OPTS` environment variable: (port can be modified to be available according to the actual situation)

```shell
export KAFKA_OPTS="$KAFKA_OPTS -javaagent:/usr/local/datakit/data/jolokia-jvm-agent.jar=host=*,port=8080"
```

Alternatively, you can start Jolokia separately and point it to the Kafka process PID:

```shell
java -jar </path/to/jolokia-jvm-agent.jar> --host 127.0.0.1 --port=8080 start <Kafka-PID>
```

<!-- markdownlint-disable MD046 -->

???+ info

    - Jolokia not allows change port number in the running state. If found command with `--port` can't change the port, this indicates Jolokia is still in running. If want to change Jolokia port, you must exit Jolokia first and restart it.
    - Exit Jolokia command: `java -jar </path/to/jolokia-jvm-agent.jar> --quiet stop <Kafka-PID>`.

    For more Jolokia command information can refer to [here](https://jolokia.org/reference/html/agents.html#jvm-agent){:target="_blank"}.

<!-- markdownlint-enable -->

### Collector Configuration {#input-config}

<!-- markdownlint-disable MD046 -->
=== "Host Installation"

    Go to the `conf.d/samples` directory under the DataKit installation directory, copy `{{.InputName}}.conf.sample` and name it `{{.InputName}}.conf`. Examples are as follows:
    
    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```
    
    Once configured, [restart DataKit](../datakit/datakit-service-how-to.md#manage-service).

=== "Kubernetes"

    The collector can now be turned on by [ConfigMap Injection Collector Configuration](../datakit/datakit-daemonset-deploy.md#configmap-setting).
<!-- markdownlint-enable -->

### Collection Mode {#collection-mode}

The Kafka collector supports two collection modes:

1. **Auto Collect Mode** (default): Automatically discovers and collects all Kafka-related MBean metrics (``kafka.*:*``) without manual configuration.
2. **Manual Mode**: Requires manual configuration of MBean metrics to collect via `[[inputs.kafka.metric]]` configuration items.

#### Auto Collect Mode (Default) {#enable-auto-collect-mode}

By default, the Kafka collector uses auto collect mode, which automatically discovers and collects all Kafka MBean metrics. No additional configuration is required:

```toml
[[inputs.kafka]]
  urls = ["http://localhost:8080/jolokia"]
  enable_auto_collect = true  # Enabled by default
```

> **Important**: When auto collect mode is enabled, all metrics will be unified into the `kafka` measurement, and field names will follow the `domain.type.name.attr` format. If you need to use the previous measurements (such as `kafka_controller`, `kafka_replica_manager`, `kafka_topic`, etc.) and field names, please set `enable_auto_collect = false` and use manual mode for configuration.

##### MBean Blacklist {#mbean-blacklist}

In auto collect mode, you can exclude unwanted MBeans using the `mbean_blacklist` configuration item (supports wildcards `*` and `?`):

```toml
[[inputs.kafka]]
  mbean_blacklist = [
    "kafka.log:*",  # Exclude all MBeans under kafka.log domain
    "kafka.server:name=*,topic=*,type=BrokerTopicMetrics",  # Exclude all topic-level BrokerTopicMetrics MBeans
  ]
```

##### Field Naming Rules {#field-naming-rules}

In auto collect mode, field names follow this format: `domain.type.name.attr`

- `domain`: MBean domain name (e.g., `kafka.server`, `kafka.controller`)
- `type`: MBean type property value (e.g., `BrokerTopicMetrics`, `ReplicaManager`)
- `name`: MBean name property value (e.g., `MessagesInPerSec`, `BytesInPerSec`)
- `attr`: MBean attribute name (e.g., `Count`, `Value`, `MeanRate`, `OneMinuteRate`, `FiveMinuteRate`, `FifteenMinuteRate`)

Common MBean attributes include:

- `Count`: Cumulative value since startup
- `Value`: Current value (for Gauge type metrics)
- `MeanRate`: Average rate since startup
- `OneMinuteRate`: Average rate over the last minute
- `FiveMinuteRate`: Average rate over the last five minutes
- `FifteenMinuteRate`: Average rate over the last fifteen minutes

Example field names:

- `kafka.controller.KafkaController.GlobalTopicCount.Value`: Global topic count
- `kafka.server.ReplicaManager.PartitionCount.Value`: Number of partitions

##### Tag Extraction Rules {#tag-extraction-rules}

Auto collect mode extracts tags from MBean properties. Common tags include:

- `request`: Request type
- `topic`: Kafka topic name
- `delayedOperation`: Delayed operation type
- `error`: Error type
- `version`: Version information
- `partition`: Kafka partition ID

> Note: `domain`, `type`, and `name` will not be used as tags by default. These are included as part of the field name (format: `domain.type.name.attr`).

???+ tip "MBean Reference Documentation"

For detailed information about Kafka MBeans and their meanings, please refer to the [Kafka official documentation](https://kafka.apache.org/documentation/#monitoring){:target="_blank"} monitoring metrics section.

#### Using Manual Mode {#manual-mode}

If you need to use manual mode, set `enable_auto_collect = false` and configure `[[inputs.kafka.metric]]` items:

```toml
[[inputs.kafka]]
  urls = ["http://localhost:8080/jolokia"]
  enable_auto_collect = false

  [[inputs.kafka.metric]]
    name         = "kafka_controller"
    mbean        = "kafka.controller:name=*,type=*"
    field_prefix = "#1."
```

## Metric {#metric}

For all of the following data collections, the global election tags will added automatically, we can add extra tags in `[inputs.{{.InputName}}.tags]` if needed:

```toml
[inputs.{{.InputName}}.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
  # ...
```

{{ range $i, $m := .Measurements }}

### `{{$m.Name}}`

{{$m.Desc}}

{{$m.MarkdownTable}}

{{ end }}

## Log Collection {#logging}

To collect kafka's log, open `files` in {{.InputName}}.conf and write to the absolute path of the kafka log file. For example:

```toml
[[inputs.{{.InputName}}]]
  ...
  [inputs.{{.InputName}}.log]
    files = ["/usr/local/var/log/kafka/error.log","/usr/local/var/log/kafka/kafka.log"]
```

When log collection is turned on, a log with a log `source` of `kafka` is generated by default.

>Note: DataKit must be installed on Kafka's host to collect Kafka logs.

Example of cutting logs:

```log
[2020-07-07 15:04:29,333] DEBUG Progress event: HTTP_REQUEST_COMPLETED_EVENT, bytes: 0 (io.confluent.connect.s3.storage.S3OutputStream:286)
```

The list of cut fields is as follows:

| Field Name | Field Value                                                 |
| ------ | ------------------------------------------------------ |
| msg    | Progress event: HTTP_REQUEST_COMPLETED_EVENT, bytes: 0 |
| name   | io.confluent.connect.s3.storage.S3OutputStream:286     |
| status | DEBUG                                                  |
| time   | 1594105469333000000                                    |

## FAQ {#faq}

<!-- markdownlint-disable MD013 -->
### Why can't see `kafka_producer/kafka_consumer/kafka_connect` measurements? {#faq-no-data}

After Kafka service is started, if you need to collect Producer/Consumer/Connector indicators, you need to configure Jolokia for them respectively.

Referring to [Kafka Quick Start](https://kafka.apache.org/quickstart){:target="_blank"}, configure the `KAFKA_OPTS` environment variable for the example of Producer, as follows:

```shell
export KAFKA_OPTS="-javaagent:/usr/local/datakit/data/jolokia-jvm-agent.jar=host=127.0.0.1,port=8090"
```

Go into the Kafka directory and start a Producer:

```shell
bin/kafka-console-producer.sh --topic quickstart-events --bootstrap-server localhost:9092
```

Copy a Kafka.conf to open multiple Kafka collectors and configure the url:

```toml
  urls = ["http://localhost:8090/jolokia"]
```

And remove comments from the fields in the collect producer metrics section:

```toml
  # The following metrics are available on producer instances.  
  [[inputs.{{.InputName}}.metric]]
    name       = "kafka_producer"
    mbean      = "kafka.producer:type=*,client-id=*"
    tag_keys   = ["client-id", "type"]
```

Restart DataKit, which then collects metrics for the Producer instance.

<!-- markdownlint-enable -->
