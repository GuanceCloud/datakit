{{.CSS}}
# Java 示例

- DataKit 版本：{{.Version}}
- 文档发布日期：{{.ReleaseDate}}
- 操作系统支持：全平台

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
- dd.logs.injection: 是否开启 Java 应用日志注入，让日志与链路数据进行关联，默认为 false，对应环境变量 DD_LOGS_INJECTION。
- dd.tags: 为每个 Span 添加默认 Tags，对应环境变量 DD_TAGS。
- dd.agent.host: Datakit 监听的地址名，默认 localhost，对应环境变量 DD_AGENT_HOST。
- dd.trace.agent.port: Datakit 监听的端口号，默认 9529，对应环境变量 DD_TRACE_AGENT_PORT。
