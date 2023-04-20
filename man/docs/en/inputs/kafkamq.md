
# Subscribe to Data in Kafka
---

{{.AvailableArchs}}

---

Datakit supports subscribing messages from kafka to gather link, metric, and log information. Currently, only `SkyWalking` and custom topic are supported.

## SkyWalking {#kafkamq-SkyWalking}

### java Agent Startup Configuration {#agent}

By default, the kafka plug-in sends `traces`, `JVM metrics`, `logging`, `Instance Properties`, and `profiled snapshots` to the kafka cluster. This feature is turned off by default. You need to put `kafka-reporter-plugin-x.y.z.jar` from `agent/optional-reporter-plugins` to `agent/plugins` to take effect.


Modify the configuration file agent/config/agent.config
```txt
# Service name: Eventually displayed in the UI, making sure it is unique
agent.service_name=${SW_AGENT_NAME:myApp}

# kafka address
plugin.kafka.bootstrap_servers=${SW_KAFKA_BOOTSTRAP_SERVERS:<ip>:<port>}

```

> Make sure kafka is started before starting.

Or through environment variables
```shell
-Dskywalking.agent.service_name=myApp 
-Dskywalking.plugin.kafka.bootstrap_servers=10.200.14.114:9092
```


Start a java project (start as a jar package)

- Linux Tomcat 7, Tomcat 8, Tomcat 9  
  Change the first line of `tomcat/bin/catalina.sh`.

```shell
CATALINA_OPTS="$CATALINA_OPTS -javaagent:/path/to/skywalking-agent/skywalking-agent.jar"; export CATALINA_OPTS
```

- Windows Tomcat 7, Tomcat 8, Tomcat 9  
  Change the first line of `tomcat/bin/catalina.bat`.

```shell
set "CATALINA_OPTS=-javaagent:/path/to/skywalking-agent/skywalking-agent.jar"
```

- JAR file  
  Add `-javaagent` argument to command line in which you start your app. eg:

 ```shell
 java -javaagent:/path/to/skywalking-agent/skywalking-agent.jar -jar yourApp.jar
 ```

- Jetty  
  Modify `jetty.sh`, add `-javaagent` argument to command line in which you start your app. eg:

```shell
export JAVA_OPTIONS="${JAVA_OPTIONS} -javaagent:/path/to/skywalking-agent/skywalking-agent.jar"
```


### Configure datakit {#datakit-config}
Copy configuration files and modify

```txt
cd /usr/local/datakit/conf.d/kafkamq
cp kafkamq.conf.sample kafkamq.conf

```

Profile description
```toml
[[inputs.kafkamq]]
  addrs = ["localhost:9092"]
  # your kafka version:0.8.2.0 ~ 2.8.0
  kafka_version = "2.8.0"
  group_id = "datakit-group"
  plugins = ["db.type"]
  # Consumer group partition assignment strategy (range, roundrobin, sticky)
  assignor = "roundrobin"

  ## kafka tls config
  # tls_enable = true
  # tls_security_protocol = "text"
  # tls_sasl_mechanism = "mechanism"
  # tls_sasl_plain_username = "user"
  # tls_sasl_plain_password = "pw"

  ## -1:Offset Newest, -2:Offset Oldest
  offsets=-1

  # customer_tags = ["key1", "key2", ...]

  ## Keep rare tracing resources list switch.
  ## If some resources are rare enough(not presend in 1 hour), those resource will always send
  ## to data center and do not consider samplers and filters.
  # keep_rare_resource = false

  [inputs.kafkamq.skywalking]
    topics = [
      "skywalking-metrics",
      "skywalking-profilings",
      "skywalking-segments",
      "skywalking-managements",
      "skywalking-meters",
      "skywalking-logging",
    ]
    namespace = ""

  ## Ignore tracing resources map like service:[resources...].
  ## The service name is the full service name in current application.
  ## The resource list is regular expressions uses to block resource names.
  ## If you want to block some resources universally under all services, you can set the
  ## service name as "*". Note: double quotes "" cannot be omitted.
  # [inputs.kafkamq.close_resource]
    # service1 = ["resource1", "resource2", ...]
    # service2 = ["resource1", "resource2", ...]
    # "*" = ["close_resource_under_all_services"]
    # ...

  ## Sampler config uses to set global sampling strategy.
  ## sampling_rate used to set global sampling rate.
  # [inputs.kafkamq.sampler]
    # sampling_rate = 1.0

  # [inputs.kafkamq.tags]
    # key1 = "value1"
    # key2 = "value2"
    # ...

  ## Storage config a local storage space in hard dirver to cache trace data.
  ## path is the local file path used to cache data.
  ## capacity is total space size(MB) used to store data.
  # [inputs.kafkamq.storage]
    # path = "./skywalking_storage"
    # capacity = 5120

  ## user custom message with PL script.
  ## Currently only log and metrics are supported, topic and pl are required
  [inputs.kafkamq.custom]
  #group_id="datakit"
  #log_topics=["apm"]
  #log_pl="log.p"
  #metric_topic=["metric1"]
  #metric_pl="kafka_metric.p"
  ## rate limit. Speed limit: rate/sec
  #limit_sec = 100
  ## sample rate
  # sampling_rate = 1.0
  #spilt_json_body = true

  ## todo: add other input-mq
 
```

Notes on configuration files:
1. `kafka_version`: The version length is 3, such as 1.0.0, 1.2.1, and so on.
2. `offsets`: note: Newest or Oldest.
3. `spilt_json_body`: When the data is an array and conforms to JSON format, it can be set to true and will be automatically cut into multiple rows of logs.

Restart datakit

## Send log to kafka {#log-to-kafka}

- log4j2

The toolkit dependency package is added to the maven or gradle.
```xml
	<dependency>
      	<groupId>org.apache.skywalking</groupId>
      	<artifactId>apm-toolkit-log4j-2.x</artifactId>
      	<version>{project.release.version}</version>
	</dependency>
```

Print trace ID in log

```xml
    <Configuration>
        <Appenders>
            <Console name="Console" target="SYSTEM_OUT">
                <PatternLayout pattern="%d [%traceId] %-5p %c{1}:%L - %m%n"/>
            </Console>
        </Appenders>
        <Loggers>
            <AsyncRoot level="INFO">
                <AppenderRef ref="Console"/>
            </AsyncRoot>
        </Loggers>
    </Configuration>
```

Send log to kafka
```xml
<GRPCLogClientAppender name="grpc-log">
        <PatternLayout pattern="%d{HH:mm:ss.SSS} [%t] %-5level %logger{36} - %msg%n"/>
    </GRPCLogClientAppender>
```

Overall configuration:
```xml
<Configuration status="WARN">
    <Appenders>
        <Console name="Console" target="SYSTEM_OUT">
            <PatternLayout pattern="%d{HH:mm:ss.SSS} [%traceId] [%t] %-5level %logger{36} %msg%n"/>
        </Console>
        <GRPCLogClientAppender name="grpc-log">
            <PatternLayout pattern="%d{yyyy-MM-dd HH:mm:ss.SSS} [%traceId] [%t] %-5level %logger{36} %msg%n"/>
        </GRPCLogClientAppender>
        <RandomAccessFile name="fileAppender" fileName="${sys:user.home}/tmp/skywalking-logs/log4j2/e2e-service-provider.log" immediateFlush="true" append="true">
            <PatternLayout>
                <Pattern>[%sw_ctx] [%p] %d{yyyy-MM-dd HH:mm:ss.SSS} [%t] %c:%L - %m%n</Pattern>
            </PatternLayout>
        </RandomAccessFile>
    </Appenders>

    <Loggers>
        <Root level="info">
            <AppenderRef ref="Console"/>
            <AppenderRef ref="grpc-log"/>
        </Root>
        <Logger name="fileLogger" level="info" additivity="false">
            <AppenderRef ref="fileAppender"/>
        </Logger>
    </Loggers>
</Configuration>
```

At this point, the agent will send the log to kafka.

How to configure more logs:

- [log4j-1.x](https://github.com/apache/skywalking-java/blob/main/docs/en/setup/service-agent/java-agent/Application-toolkit-log4j-1.x.md){:target="_blank"}
- [logback-1.x](https://github.com/apache/skywalking-java/blob/main/docs/en/setup/service-agent/java-agent/Application-toolkit-logback-1.x.md){:target="_blank"}

Logs collected from kafka do not need to be processed by pipeline. It's all cut.

## Custom Topic {#kafka-custom}

Sometimes users don't use common tools in the market, and some tripartite libraries are not open source, and the data structure is not public. This requires manual processing according to the collected data structure, which reflects the power of pipeline, and users can subscribe and consume messages through custom configuration.

Profile:
```toml
 ...
  ## 《复制时注意中文！》 
  ## user custom message with PL script.
  ## 目前仅支持 log 和 metrics， topic 和 pl 是必填
  [inputs.kafkamq.custom]
  group_id="datakit"
  log_topics=["apm"]
  log_pl="log.p"
  metric_topic=["metric1"]
  metric_pl="kafka_metric.p"
  ## rate limit. 限速：速率/秒
  limit_sec = 100
  ## sample 采样率
  sampling_rate = 1.0

 ...

```

### Example {#example}

Take a simple metric as an example to show you how to subscribe to messages using custom configuration.

When you don't know what format the data structure sent to kafka is. You can change the logging level of datakit to debug first. Open the subscription, and there will be output in the datakit log. Suppose you get the following data:
```shell
# After opening the debug log level, look at the log, and datakit prints out the message information.
tailf /var/log/datakit/log | grep "kafka_message"
```

Suppose you get a json-formatted plain text string of metric:

```json
{"time": 1666492218, "dimensions": {"bk_biz_id": 225,"ip": "10.200.64.45" },  "metrics": { "cpu_usage_pct": 0.01}, "exemplar": null}
```


With the data format, you can write pipeline scripts by hand. Log in to Guance Cloud-> Management-> Text Processing (Pipeline) to write scripts. Such as:

```toml
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

Configure the PL script and restart datakit.

### Troubleshooting {#some_problems}

Script test command to see if cutting is correct:

```shell
datakit pipeline -P metric.p -T '{"time": 1666492218,"dimensions":{"bk_biz_id": 225,"ip": "172.253.64.45"},"metrics": {"cpu_usage_pct": 0.01}, "exemplar": null}'
```

Set outputfile to local to see if the line protocol format is correct:
```shell
vim conf/datakit.conf
# If it is set to a local file, it will not be output to io, and it can be assigned to null after the test.
output_file = "/usr/local/datakit/out.pts"
# Check to see if the file out.pts is correct
```

Connection failure may be a version problem: Please fill in the kafka version correctly in the configuration file.

Other issues:

View through the `datakit monitor` command, or through `datakit monitor -V`.



