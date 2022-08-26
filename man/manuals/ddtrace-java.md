{{.CSS}}

# Java 示例

---

- 操作系统支持：:fontawesome-brands-linux: :fontawesome-brands-windows: :fontawesome-brands-apple:

## Install Libarary & Dependence

下载最新的 ddtrace agent dd-java-agent.jar。

```shell
wget -O dd-java-agent.jar https://dtdg.co/latest-java-tracer
```

## Run Java Code With DDTrace

可以通过多种途径运行你的 Java Code，如 IDE，Maven，Gradle 或直接通过 java -jar 命令，以下通过 java 命令启动应用：

```shell
java -javaagent:/path/to/dd-java-agent.jar \
-Ddd.logs.injection=true \
-Ddd.service=my-app \
-Ddd.env=staging \
-Ddd.version=1.0.0 \
-Ddd.agent.host=localhost \
-Ddd.trace.agent.port=9529 \
-jar path/to/your/app.jar
```

## Start Parameters For Tracing Java Code

- dd.env: 为服务设置环境变量，对应环境变量 DD_ENV。
- dd.version: APP 版本号，对应环境变量 DD_VERSION。
- dd.service: 设置服务名，对应环境变量 DD_SERVICE。
- dd.trace.agent.timeout: 客户端网络发送超时默认 10s，对应环境变量 DD_TRACE_AGENT_TIMEOUT。
- dd.logs.injection: 是否开启 Java 应用日志注入，让日志与链路数据进行关联，默认为 false，对应环境变量 DD_LOGS_INJECTION。
- dd.tags: 为每个 Span 添加默认 Tags，对应环境变量 DD_TAGS。
- dd.agent.host: Datakit 监听的地址名，默认 localhost，对应环境变量 DD_AGENT_HOST。
- dd.trace.agent.port: Datakit 监听的端口号，默认 9529，对应环境变量 DD_TRACE_AGENT_PORT。
- dd.trace.sample.rate: 设置采样率从 0.0(0%) ~ 1.0(100%)。
- dd.jmxfetch.enabled: 开启 JMX metrics 采集，默认值 true， 对应环境变量 DD_JMXFETCH_ENABLED
- dd.jmxfetch.config.dir: 额外的 JMX metrics 采集配置目录。Java Agent 将会在 yaml 配置文件中的 instance section 寻找 jvm_direct:true 来修改配置，对应环境变量 DD_JMXFETCH_CONFIG_DIR
- dd.jmxfetch.config: 额外的 JMX metrics 采集配置文件。JAVA agent 将会在 yaml 配置文件中的 instance section 寻找 jvm_direct:true 来修改配置对应环境变量，DD_JMXFETCH_CONFIG
- dd.jmxfetch.check-period: JMX metrics 发送频率(ms)，默认值 1500，对应环境变量 DD_JMXFETCH_CHECK_PERIOD。
- dd.jmxfetch.refresh-beans-period: 刷新 JMX beans 频率(s)，默认值 600，对应环境变量 DD_JMXFETCH_REFRESH_BEANS_PERIOD。
- dd.jmxfetch.statsd.host: Statsd 主机地址用来接收 JMX metrics，如果使用 Unix Domain Socket 请使用形如 `unix://PATH_TO_UDS_SOCKET` 的主机地址。默认值同 agent.host ，对应环境变量 DD_JMXFETCH_STATSD_HOST
- dd.jmxfetch.statsd.port: StatsD 端口号用来接收 JMX metrics ，如果使用 Unix Domain Socket 请使填写 0。默认值同 agent.port 对应环境变量 DD_JMXFETCH_STATSD_PORT

## Connect OpenTelemetry Traces and Logs
