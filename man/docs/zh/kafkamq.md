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
  在配置文件 `tomcat/bin/catalina.sh` 中的第一行添加环境变量.

```shell
CATALINA_OPTS="$CATALINA_OPTS -javaagent:/path/to/skywalking-agent/skywalking-agent.jar"; export CATALINA_OPTS
```

- Windows Tomcat 7, Tomcat 8, Tomcat 9  
  在配置文件 `tomcat/bin/catalina.bat` 中的第一行添加环境变量.

```shell
set "CATALINA_OPTS=-javaagent:/path/to/skywalking-agent/skywalking-agent.jar"
```

- JAR 包形式启动  
  在启动 java 项目时候添加 `-javaagent` 参数:

 ```shell
 java -javaagent:/path/to/skywalking-agent/skywalking-agent.jar -jar yourApp.jar
 ```

- Jetty  
  修改 `jetty.sh`, 并添加启动参数 `-javaagent` :

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

更多的情况往往是现有的系统已经将数据发送到 kafka，而随着开发运维人员迭代，进行修改输出变的复杂难以实现，这时候使用自定义模式便是很好的方式。

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

### 基准测试 {#benchmark}

消息的消费能力受限于网络和带宽的限制，所以基准测试只是测试 DK 的消费能力而不是 IO 能力。本次测试的机器配置是4核8线程、16G内存。
测试过程中 CPU 峰值 60%~70%，内存增加10%。

| 消息数量  | 用时    | 每秒消费能力（条） |
|-------|-------|-----------|
| 100k  | 5s~7s | 16k       |
| 1000k | 1m30s | 11k       |

另外减少日志输出、关闭 cgroup 限制、增加内网和公网带宽等，可以增加消费能力。

### 多台 datakit 负载均衡 {#datakit-assignor}

当消息量很大，一台 datakit 消费能力不足时可以增加多台 datakit 进行消费，这里有三点需要注意：
1. 确保topic分区不是一个（至少2个），这个可以通过工具查看 [kafka-map](https://github.com/dushixiang/kafka-map/releases)
1. 确保 kafkamq 采集器的配置是 assignor = "roundrobin"(负载均衡策略的一种) ,  group_id="datakit"（组名称必须一致，否则会重复消费）
1. 确保消息的生产者将消息发送多分区，语言不同方法不同 这里不列出代码了，自行 google


### 问题排查 {#some_problems}

当写好 pipeline 脚本之后不确定是否能切割正确，可以使用测试命令：
```shell
datakit pipeline metric.p -T '{"time": 1666492218,"dimensions":{"bk_biz_id": 225,"ip": "172.253.64.45"},"metrics": {"cpu_usage_pct": 0.01}, "exemplar": null}'
```

切割正确之后，可以查看行协议数据是否正确，暂时将 output_file 设置为本地文件：
```shell
vim conf/datakit.conf
# 设置为本地文件，就不会输出到io，测试结束之后赋值为空即可。
output_file = "/usr/local/datakit/out.pts"
# 查看文件 out.pts 是否正确
```

连接失败可能是版本问题：请在配置文件中正确填写 kafka 版本。目前支持的版本列表：

| Kafka版本  | 对应的配置字段  |
|----------|----------|
| 0.8.2    | 0_8_2_0  |  
| 0.8.2.1  | 0_8_2_1  |   
| 0.9.0.0  | 0_9_0_0  |  
| 0.9.0.1  | 0_9_0_1  |  
| 0.10.0   | 0_10_0_0 |  
| 0.10.0.1 | 0_10_0_1 |  
| 0.10.1.0 | 0_10_1_0 |  
| 0.11.0   | 0_11_0_0 |  
| 0.11.0.1 | 0_11_0_1 |  
| 0.11.0.2 | 0_11_0_2 |  
| 1.0.0.0  | 1_0_0    |  
| 1.1.0.0  | 1_1_0    |  
| 1.1.1.0  | 1_1_1    |  
| 2.0.0.0  | 2_0_0    |  
| 2.0.1.0  | 2_0_1    |  
| 2.1.0.0  | 2_1_0    |  
| 2.2.0.0  | 2_2_0    |  
| 2.3.0.0  | 2_3_0    |  
| 2.4.0.0  | 2_4_0    |  
| 2.5.0.0  | 2_5_0    |  
| 2.6.0.0  | 2_6_0    |  
| 2.7.0.0  | 2_7_0    |  
| 2.8.0.0  | 2_8_0    |  

> 版本是以`0`开头，版本的长度应该是4位，如：0_8_2_0 ，如果不是`0`开头，版本的长度应该是3位 如：2_8_0


其他问题：

通过 `datakit monitor` 命令查看。 或者 `datakit monitor -V` 查看。



