# Java 示例

Pinpoint Java Agent [下载地址](https://github.com/pinpoint-apm/pinpoint/releases){:target="_blank"}

---

## 配置 Datakit Agent {#config-datakit-agent}

参考[配置 Datakit 中的 Pinpoint Agent](pinpoint.md#agent-config)

## 配置 Pinpoint Java Agent {#config-pinpoint-java-agent}

运行一下命令启动 Pinpoint Agent

```shell
java -javaagent：/path-to-pinpoint-agent-path/pinpoint-bootstrap.jar \
     -Dpinpoint.agentId=agent-id \
     -Dpinpoint.applicationName=service-name \
     -Dpinpoint.config=/path-to-pinpoint-agent-config-path/pinpoint-root.config \
     -jar /path-to-java-app
```

基本参数说明：

- pinpoint.profiler.profiles.active               : Pinpoint profiler 工作模式(release/local) 与日志输出有关
- pinpoint.applicationName                        : 服务名
- pinpoint.agentId                                : Agent ID
- pinpoint.agentName                              : Agent name
- profiler.transport.module                       : 传输协议（gRPC/Thrift）
- profiler.transport.grpc.collector.ip            : Collector IP 地址（即启动 Datakit 的主机地址）
- profiler.transport.grpc.agent.collector.port    : Agent collector port（即 Pinpoint Agent 在 Datakit 中的监听端口）
- profiler.transport.grpc.metadata.collector.port : Metadata collector port（即 Pinpoint Agent 在 Datakit 中的监听端口）
- profiler.transport.grpc.stat.collector.port     : stat collector port（即 Pinpoint Agent 在 Datakit 中的监听端口）
- profiler.transport.grpc.span.collector.port     : span collector port（即 Pinpoint Agent 在 Datakit 中的监听端口）
- profiler.sampling.enable                        : 是否启动采样
- profiler.sampling.type                          : 采样算法
- profiler.sampling.counting.sampling-rate        : 采样率
- profiler.sampling.percent.sampling-rat          : 采样率

## 支持的模块 {#supported-modules}

- JDK 8+
- Tomcat, Jetty, JBoss EAP, Resin, Websphere, Vertx, Weblogic, Undertow, Akka HTTP
- Spring, Spring Boot (Embedded Tomcat, Jetty, Undertow, Reactor Netty), Spring WebFlux
- Apache HttpClient 3 / 4 / 5, JDK HttpConnector, GoogleHttpClient, OkHttpClient, NingAsyncHttpClient
- Thrift, DUBBO, GRPC, Apache CXF
- ActiveMQ, RabbitMQ, Kafka, RocketMQ, Paho MQTT
- MySQL, Oracle, MSSQL, JTDS, CUBRID, POSTGRESQL, MARIA, Informix, Spring Data R2DBC
- Arcus, Memcached, Redis(Jedis, Lettuce, Redisson), CASSANDRA, MongoDB, Hbase, Elasticsearch
- iBATIS, MyBatis
- DBCP, DBCP2, HIKARICP, DRUID
- Gson, Jackson, JSON Lib, Fastjson
- log4j, Logback, log4j2
- OpenWhisk, Kotlin Coroutines

## 兼容性 {#compatibility}

当前 Datakit 使用的 Pinpoint Agent 为 [pinpoint-go-agent](https://github.com/pinpoint-apm/pinpoint-go-agent){:target="_blank"}-v1.3.2

当前完成测试的 Pinpoint Agent 版本包括：

- pinpoint-agent-2.2.1
- pinpoint-agent-2.3.1
- pinpoint-agent-2.4.1
- pinpoint-agent-2.5.1
