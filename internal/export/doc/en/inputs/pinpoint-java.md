# Java example

Pinpoint Java Agent [Download address](https://github.com/pinpoint-apm/pinpoint/releases){:target="_blank"}

---

## Configure Datakit Agent {#config-datakit-agent}

Refer to [Configuring Pinpoint Agent in Datakit](pinpoint.md#config)

## Configure Pinpoint Java Agent {#config-pinpoint-java-agent}

Run the following command to start Pinpoint Agent

```shell
java -javaagent：/path-to-pinpoint-agent-path/pinpoint-bootstrap.jar \
     -Dpinpoint.agentId=agent-id \
     -Dpinpoint.applicationName=service-name \
     -Dpinpoint.config=/path-to-pinpoint-agent-config-path/pinpoint-root.config \
     -jar /path-to-java-app
```

Basic parameter description:

- pinpoint.profiler.profiles.active               : Pinpoint profiler working mode (release/local) related to log output
- pinpoint.applicationName                        : service name
- pinpoint.agentId                                : Agent ID
- pinpoint.agentName                              : Agent name
- profiler.transport.module                       : Transport protocol (gRPC/Thrift)
- profiler.transport.grpc.collector.ip            : Collector IP address (that is, the host address where Datakit is started)
- profiler.transport.grpc.agent.collector.port    : Agent collector port（ (that is, the listening port of Pinpoint Agent in Datakit)
- profiler.transport.grpc.metadata.collector.port : Metadata collector port (that is, the listening port of Pinpoint Agent in Datakit)
- profiler.transport.grpc.stat.collector.port     : stat collector port (that is, the listening port of Pinpoint Agent in Datakit)
- profiler.transport.grpc.span.collector.port     : span collector port (that is, the listening port of Pinpoint Agent in Datakit)
- profiler.sampling.enable                        : whether to start sampling
- profiler.sampling.type                          : sampling algorithm
- profiler.sampling.counting.sampling-rate        : sampling rate
- profiler.sampling.percent.sampling-rat          : sampling rate

## Supported modules {#supported-modules}

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

## Compatibility {#compatibility}

The Pinpoint Agent currently used by Datakit is [pinpoint-go-agent](https://github.com/pinpoint-apm/pinpoint-go-agent){:target="_blank"}-v1.3.2

Pinpoint Agent versions currently undergoing testing include:

- pinpoint-agent-2.2.1
- pinpoint-agent-2.3.1
- pinpoint-agent-2.4.1
- pinpoint-agent-2.5.1
