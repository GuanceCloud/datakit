---
title     : 'KafkaMQ'
summary   : 'Collect metrics and log data via Kafka'
__int_icon      : 'icon/kafka'
dashboard :
  - desc  : 'N/A'
    path  : '-'
monitor   :
  - desc  : 'N/A'
    path  : '-'
---

<!-- markdownlint-disable MD025 -->
# Subscribe to Data in Kafka
<!-- markdownlint-enable -->

---

{{.AvailableArchs}}

---

Datakit supports subscribing messages from kafka to gather link, metric, and log information. Currently, only `SkyWalking`,`Jaeger` and `custom topic` are supported.

## Config {#config}

### Collector Configuration {#input-config}

<!-- markdownlint-disable MD046 -->
=== "Host deployment"

    Go to the `conf.d/{{.Catalog}}` directory under the DataKit installation directory, copy `{{.InputName}}.conf.sample` and name it `{{.InputName}}.conf`. Examples are as follows:
    
    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```

    After configuring, [restart DataKit](../datakit/datakit-service-how-to.md#manage-service).

=== "Kubernetes"

    The collector can now be turned on by [ConfigMap Injection Collector Configuration](../datakit/datakit-daemonset-deploy.md#configmap-setting).   
<!-- markdownlint-enable -->

---

> Note: Starting from v1.6.0, sampling and rate limiting are fully supported for all messages. Previous versions only supported custom messages.

Notes on configuration files:

1. `kafka_version`: The version length is 3, such as 1.0.0, 1.2.1, and so on.
2. `offsets`: note: Newest or Oldest.
3. `SASL`: If security authentication is enabled, please configure the user and password correctly.
4. Starting from v1.23.0, it supports multi-threaded mode.

### Consumer group mode {#consumer_group}

The Kafka consumer group mode is a pattern in which multiple consumers collectively consume messages.
In this mode, multiple consumers can form a consumer group, and each consumer can independently consume messages, but the same message will only be processed by one consumer within the group.
This mode helps to achieve message load balancing, improve message processing throughput, and reliability.
When a consumer within the group fails or goes offline, Kafka will automatically reassign the messages from that consumer to other consumers for processing, thereby achieving fault tolerance and high availability.

### SkyWalking {#KafkaMQ-SkyWalking}
The kafka plugin will send `traces`, `JVM metrics`, `logging`, `Instance Properties`, and `profiled snapshots` to the kafka cluster by default.

This feature is disabled by default. Need to put `kafka-reporter-plugin-x.y.z.jar` from `agent/optional-reporter-plugins` into `agent/plugins` to take effect.

config:

```toml
 ## skywalking custom
 [inputs.kafkamq.skywalking]
  
   ## Required！send to datakit skywalking input.
   dk_endpoint="http://localhost:9529"

   topics = [
     "skywalking-metrics",
     "skywalking-profilings",
     "skywalking-segments",
     "skywalking-managements",
     "skywalking-meters",
     "skywalking-logging",
   ]
   namespace = ""
```

Open the comment to start the subscription. The subscribed topic is in the SkyWalking agent configuration file `config/agent.config`.

> Note: This collector just forwards the subscribed data to the Datakit SkyWalking collector, please open the [SkyWalking](skywalking.md) collector and open the dk_endpoint comment!

### Jaeger {#jaeger}

Configuration：

```toml
## Jaeger from kafka. Please make sure your Datakit Jaeger collector is open ！！！
[inputs.kafkamq.jaeger]
    ## Required！ ipv6 is "[::1]:9529"
    dk_endpoint="http://localhost:9529"

    ## Required！ topics 
    topics=["jaeger-spans","jaeger-my-spans"]
```

> Note: This collector just forwards the subscribed data to the Datakit Jaeger collector, please open the [jaeger](jaeger.md) collector and open the `dk_endpoint` comment!

### Custom Topic {#kafka-custom}

Sometimes users don't use common tools in the market, and some tripartite libraries are not open source, and the data structure is not public. This requires manual processing according to the collected data structure, which reflects the power of Pipeline, and users can subscribe and consume messages through custom configuration.

In many cases, existing systems have already been sending data to Kafka, and as developers and operations personnel iterate, making modifications to the output becomes complex and challenging to achieve. In such situations, using a custom schema is a good approach.

Configuration:

```toml
 ...
  ## user custom message with PL script.
  [inputs.kafkamq.custom]
    #spilt_json_body = true
    [inputs.kafkamq.custom.spilt_topic_map]
      "log_topic"=true
      "log01"=false
    [inputs.kafkamq.custom.log_topic_map]
      "log_topic"="log.p"
      "log"="rum_apm.p"
    [inputs.kafkamq.custom.metric_topic_map]
      "metric_topic"="rum_apm.p"
      
    [inputs.kafkamq.custom.rum_topic_map]
      "rum"="rum.p"
      
```

> Note: The pl script of metric should be placed in the `pipeline/metric/` directory, and the pl script of rum should be placed in the `pipeline/rum/` directory.

Theoretically, each message body should be a log or a metric. If your message is multiple logs, you can use `spilt_json_body` to enable the global function of splitting arrays, you can also use `spilt_topic_map` to enable the function of splitting arrays for a single topic: When the data is an array and conforms to the JSON format, it can be set to true, and PL can be used to Arrays are sliced into individual log or metric data.

### Consumer OpenTelemetry Data {#otel}

Configuration:

```toml
## Receive and consume OTEL data from kafka.
[inputs.kafkamq.otel]
    dk_endpoint="http://localhost:9529"
    trace_api="/otel/v1/trace"
    metric_api="/otel/v1/metric"
    trace_topics=["trace1","trace2"]
    metric_topics=["otel-metric","otel-metric1"]
```

The `dk_endpoint` `trace_api` `metric_api` corresponds to the address of the DataKit and the API address of the OpenTelemetry collector.

> Note: The messages subscribed to from Kafka are not directly parsed, but are directly sent to the `OpenTelemetry` collector, so it is necessary to open the [OpenTelemetry collector](openTelemetry. md). Currently, only the `x-protobuf` data stream format is supported.

### Example {#example}

Take a simple metric as an example to show you how to subscribe to messages using custom configuration.

When you don't know what format the data structure sent to kafka is. You can change the logging level of Datakit to debug first. Open the subscription, and there will be output in the Datakit log. Suppose you get the following data:

```shell
# After opening the debug log level, look at the log, and datakit prints out the message information.
tailf /var/log/datakit/log | grep "kafka_message"
```

Suppose you get a JSON-formatted plain text string of metric:

```json
{"time": 1666492218, "dimensions": {"bk_biz_id": 225,"ip": "10.200.64.45" },  "metrics": { "cpu_usage_pct": 0.01}, "exemplar": null}
```

With the data format, you can write Pipeline scripts by hand. Log in to Guance Cloud-> Management-> Text Processing (Pipeline) to write scripts. Such as:

```python
data=load_json(message)
drop_origin_data()

hostip=data["dimensions"]["ip"]
bkzid=data["bk_biz_id"]
cast(bkzid,"sttr")

set_tag(hostip,hostip)
set_tag(bk_biz_id,bkzid)

add_key(cpu_usage_pct,data["metrics"]["cpu_usage_pct"])

# Note that this is the line protocol default, and the message_len can be deleted after the pl script is passed.
drop_key(message_len)
```

Place the file in the directory `/usr/local/datakit/pipeline/metric/`.

> Note: The pl script for metrics data is placed under `metric/` and the pl script for logging data is placed under `pipeline/`

Configure the PL script and restart Datakit.

## Handle {#handle}

Configuration:

```toml
  [inputs.kafkamq.remote_handle]
    ## Required！
    endpoint="http://localhost:8080"
    ## Required！ topics
    topics=["spans","my-spans"]
    send_message_count = 100
    debug = false
    is_response_point = true
    # header_check = false 
```

KafkaMQ provides a plug-in mechanism: send data ([]byte) to an external handle via HTTP, and then return the formatted data of the LineProtocol through response after processing. Realize customized data processing.

Config instructions:

- `endpoint` Handle addr
- `send_message_count` The number of messages sent at one time.
- `topics`  Topic array
- `debug` bool value, When the debug function is enabled, `message_points` will be invalid. If the debug mode is enabled, the data in the original message body will be sent without message merging.
- `is_response_point` Whether to send row protocol data back
- `header_check` Special head detection (bfy customized, not generic)


After receiving the messages, KafkaMQ merges them into a package containing `send_message_count` messages and sends them to the specified handle address. The data structure is as follows:

```txt
[
  {"topic": "bfySpan", "value": "dmFsdWUx"},
  {"topic": "bfySpan", "value": "dmFsdWUx"},
  {"topic": "bfySpan", "value": "dmFsdWUx"},
  {"topic": "bfySpan", "value": "dmFsdWUx"},
  ...
]
```

The returned data should comply with `v1/write/tracing` interface regulations, [Interface Documentation](apis.md#api-v1-write)

The returned header should also indicate the type of the data: the default is `tracing`

```txt
X-category=tracing  
```

[DataKit api category](apis/#category)

As long as the data is received, it means that KafkaMQ has successfully sent the data. Regardless of the parsing, it should return 200 and wait for the next request.

If the parsing fails, it is recommended to set `debug=true` in the KafkaMQ configuration. At this time, JSON assembly and serialization will not be performed. Instead, the `body` of the request is the message itself.

---

External plugins have some constraints:

- KafkaMQ receives data but is not responsible for serialization after parsing, because this is a custom development and cannot be used by all users.
- The data parsed by the external plug-in can be sent to [dk apis](apis.md#api-v1-write) , or returned to KafkaMQ and then sent to Observation Cloud.
- The response returned to KafkaMQ must be in the ***line protocol format***, if it is `JSON` format, the header information needs to be included: `Content-Type:application/json` In addition, the returned header information should also include Type: `X-category:tracing` indicates this link information.
- External plug-ins should return 200 regardless of parsing failure or not when data is received.
- If KafkaMQ sends data to the external plug-in, if there is a timeout, the port does not exist and other problems. Will try to reconnect. Messages from Kafka are no longer consumed.

## benchmark {#benchmark}

The consumption capability of messages is limited by the network and bandwidth, so the benchmark test only tests the consumption capability of DK rather than the IO capability. The machine configuration for this test is 4 cores, 8 threads, and 16G memory. CPU peaked at 60% to 70% during the test and memory increased by 10%.

| message num | time  | num/Second |
|-------------|-------|------------|
| 100k        | 5s~7s | 16k        |
| 1000k       | 1m30s | 11k        |

Additionally, reducing log output, turning off cgroup limits, and increasing intranet and public network bandwidth can increase spending power.

### load balancing {#load-balancing}

When the amount of messages is large and the consumption capacity of one Datakit is insufficient, multiple Datakit can be added for consumption. Here are three points to note:

1. Make sure that the topic partition is not one (at least 2), which can be viewed through the tool [kafka-map](https://github.com/dushixiang/kafka-map/releases){:target="_blank"}
1. Make sure that the configuration of the KafkaMQ collector is `assignor = "roundrobin"` (a type of load balancing strategy), `group_id="datakit"` (group names must be consistent, otherwise consumption will be repeated)
1. Make sure that the producer of the message sends the message to multiple partitions. The method is different for different languages. The code is not listed here, and you can google it yourself.

## FAQ {#faq}

<!-- markdownlint-disable MD013 -->
### :material-chat-question: how to test Pipeline script {#test_Pipeline}
<!-- markdownlint-enable -->

Script test command to see if cutting is correct:

```shell
datakit pipeline -P metric.p -T '{"time": 1666492218,"dimensions":{"bk_biz_id": 225,"ip": "172.253.64.45"},"metrics": {"cpu_usage_pct": 0.01}, "exemplar": null}'
```

We can setup [recorder](../datakit/datakit-tools-how-to.md#enable-recorder) in *datakit.conf* to check if data ok.

Connection failure may be a version problem: Please fill in the kafka version correctly in the configuration file.

<!-- markdownlint-disable MD013 -->
### :material-chat-question: How to handle the situation of message backlog in Kafka {#message_backlog}
<!-- markdownlint-enable -->

1. Enable multi threaded mode to increase consumption capacity.
2. If performance reaches a bottleneck, then expand physical memory and CPU.
3. Increase the write capacity of the backend.
4. Remove any network bandwidth restrictions.
5. Increase the number of collectors and expand the number of message partitions to allow more consumers to consume.
6. If the above solutions still cannot solve the problem, you can use [bug-report](../datakit/why-no-data/#bug-report) to collect runtime metrics for analysis.


Other issues:

View through the `datakit monitor` command, or through `datakit monitor -V`.
