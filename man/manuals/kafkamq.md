{{.CSS}}
# 订阅 Kafka 中的数据
---

{{.AvailableArchs}}

***作者：宋龙奇***

---

Datakit 支持从 kafka 中订阅消息采集链路、指标和日志信息。目前仅支持 `SkyWalking`.

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
  # kafka 地址
  addr = "localhost:9092"  
  # topic groupID
  group_id = "datakit-group"

  [inputs.kafkamq.skywalking]  
    topics = [
      "skywalking-metrics",
      "skywalking-profilings",
      "skywalking-segments",
      "skywalking-managements",
      "skywalking-meters",
      "skywalking-logging",
    ]
    # 如果 skywalking-agent 已经配置了 namespace ，这里必填。
    namespace = ""

  # [inputs.kafkamq.threads]
    # buffer = 100
    # threads = 8

  plugins = ["db.type"]

  # customer_tags = ["key1", "key2", ...]

  ## Keep rare tracing resources list switch.
  ## If some resources are rare enough(not presend in 1 hour), those resource will always send
  ## to data center and do not consider samplers and filters.
  # keep_rare_resource = false

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

  # 本地缓存
  # [inputs.kafkamq.storage]
    # path = "./skywalking_storage"
    # capacity = 5120
 
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

