{{.CSS}}
# 订阅 Kafka 中的数据
---

{{.AvailableArchs}}

***作者：宋龙奇***

---

Datakit 支持从 kafka 中订阅消息采集链路、指标和日志信息。目前仅支持 `SkyWalking` 以及自定义 topic.

## SkyWalking {#kafkamq-SkyWalking}

### java agent启动配置 {#agent}

kafka 插件默认会将 `traces`, `JVM metrics`, `logging`,  `Instance Properties`, and `profiled snapshots` 发送到 kafka 集群中。 该功能默认是关闭的。需要将 `kafka-reporter-plugin-x.y.z.jar`, 从 `agent/optional-reporter-plugins` 放到 `agent/plugins` 才会生效.


修改配置文件 agent/config/agent.config
```txt
# 服务名称：最终会在 UI 中展示，确保唯一
agent.service_name=${SW_AGENT_NAME:myApp}

# kafka 地址
plugin.kafka.bootstrap_servers=${SW_KAFKA_BOOTSTRAP_SERVERS:<ip>:<port>}

```

> 在启动之前请先确保 kafka 已经启动。

或者 通过环境变量方式
```shell
-Dskywalking.agent.service_name=myApp 
-Dskywalking.plugin.kafka.bootstrap_servers=10.200.14.114:9092
```


启动java项目（jar包形式启动）

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


### 配置 datakit {#datakit-config}
复制配置文件并修改

```txt
cd /usr/local/datakit/conf/kafkamq
cp kafkamq.conf.sample kafka.conf

```

配置文件说明
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
  #offsets=-2

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
  ## 目前仅支持 log 和 metrics， topic 和 pl 是必填
  [inputs.kafkamq.custom]
  #group_id="datakit"
  #log_topics=["apm"]
  #log_pl="log.p"
  #metric_topic=["metric1"]
  #metric_pl="kafka_metric.p"
  ## rate limit. 限速：速率/秒
  #limit_sec = 100
  ## sample 采样率
  # sampling_rate = 1.0

  ## todo: add other input-mq
 
```

重启 datakit

## 将日志发送到 kafka {#log-to-kafka}

- log4j2

toolkit 依赖包添加到 maven 或者 gradle 中。
```xml
	<dependency>
      	<groupId>org.apache.skywalking</groupId>
      	<artifactId>apm-toolkit-log4j-2.x</artifactId>
      	<version>{project.release.version}</version>
	</dependency>
```

在日志中打印 trace ID

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

将日志发送到kafka
```xml
<GRPCLogClientAppender name="grpc-log">
        <PatternLayout pattern="%d{HH:mm:ss.SSS} [%t] %-5level %logger{36} - %msg%n"/>
    </GRPCLogClientAppender>
```

整体配置：
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

至此 agent 会将日志发送到 kafka 中。

更多日志如何配置：

- [log4j-1.x](https://github.com/apache/skywalking-java/blob/main/docs/en/setup/service-agent/java-agent/Application-toolkit-log4j-1.x.md){:target="_blank"}
- [logback-1.x](https://github.com/apache/skywalking-java/blob/main/docs/en/setup/service-agent/java-agent/Application-toolkit-logback-1.x.md){:target="_blank"}

从 kafka 中采集的日志不需要通过 pipeline 处理。已经全部切割好。

## 自定义Topic {#kafka-custom}

有些时候用户使用的并不是市面上常用的工具，有些的三方库并不是开源的，数据结构也不是公开的。这样就需要根据采集到的数据结构手动进行处理，这时候就体现到 pipeline  的强大之处，用户可通过自定义配置进行订阅并消费消息。

配置文件：
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

### 示例 {#example}

以一个简单的metric为例，介绍如何使用自定义配置订阅消息。

当不知道发送到 kafka 上的数据结构时什么格式时。可以先将 datakit 的日志级别改为 debug。 将订阅打开，在 datakit 日志中会有输出。假设拿到的如下数据：
```shell
# 打开 debug 日志级别之后,查看日志, datakit 会将消息信息打印出来.
tailf /var/log/datakit/log | grep "kafka_message"
```

假设拿到的这是一个 metric 的 json 格式纯文本字符串：

```json
{"time": 1666492218, "dimensions": {"bk_biz_id": 225,"ip": "10.200.64.45" },  "metrics": { "cpu_usage_pct": 0.01}, "exemplar": null}
```


有了数据格式，就可以手写 pipeline 脚本。登录 观测云 -> 管理 -> 文本处理(pipeline) 编写脚本。 如：

```toml
data=load_json(message)
drop_origin_data()

hostip=data["dimensions"]["ip"]
bkzid=data["bk_biz_id"]
cast(bkzid,"sttr")

set_tag(hostip,hostip)
set_tag(bk_biz_id,bkzid)

add_key(cpu_usage_pct,data["metrics"]["cpu_usage_pct"])
# 注意 此处为行协议缺省值，pl脚本通过之后 这个 message_len 就可以删掉了。
drop_key(message_len)
```

将文件放到 `/usr/local/datakit/pipeline/metric/` 目录下。

> 注意：指标数据的pl脚本放到`metric/`下，logging数据的pl脚本放到 `pipeline/`

配置好 PL 脚本，重启 datakit。

### 问题排查 {#some_problems}

脚本测试命令 查看切割是否正确：
```shell
datakit pipeline metric.p -T '{"time": 1666492218,"dimensions":{"bk_biz_id": 225,"ip": "172.253.64.45"},"metrics": {"cpu_usage_pct": 0.01}, "exemplar": null}'
```

将 outputfile 设置为本地， 查看行协议格式是否正确：
```shell
vim conf/datakit.conf
# 设置为本地文件，就不会输出到io，测试结束之后赋值为空即可。
output_file = "/usr/local/datakit/out.pts"
# 查看文件 out.pts 是否正确
```

连接失败可能是版本问题：请在配置文件中正确填写 kafka 版本。

其他问题：

通过 `datakit monitor` 命令查看。 或者 `datakit monitor -V` 查看。



