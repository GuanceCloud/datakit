{{.CSS}}

- 版本：{{.Version}}
- 发布日期：{{.ReleaseDate}}
- 操作系统支持：全平台

# Java 应用链路数据采集

Java 应用链路数据采集需经过如下步骤：

- Datakit 开启链路数据采集，并启动 DataKit
- 下载 Ddtrace Java JAR 包
- 开启 Java 应用

## 下载 Ddtrace JAR 包

通过如下命令下载Ddtrace Java JAR 包 `dd-java-agent.jar` ：

```shell
wget -O dd-java-agent.jar https://dtdg.co/latest-java-tracer
```

## 开启 Java 应用

通过如下命令开启 Java 应用：

```shell
java -javaagent:/path/to/dd-java-agent.jar \
-Ddd.logs.injection=true \
-Ddd.trace.sample.rate=0.1 \
-Ddd.service=my-app \
-Ddd.env=staging \
-Ddd.version=1.0 \
-Ddd.agent.host=127.0.0.1 \
-Ddd.agent.port=9529 \
-jar path/to/your/app.jar
```

其中各个参数意义是：

- `Ddd.logs.injection`: 是否开启 Java 应用日志注入，让日志与链路数据进行关联，默认为false
- `Ddd.trace.sample.rate`: 设置链路数据采样率，默认为1
- `Ddd.service`: 设置服务名
- `Ddd.env`: 设置环境名
- `Ddd.version`: 设置版本号
- `Ddd.agent.host`: 设置 Datakit 主机地址
- `Ddd.agent.port`: 设置 Datakit 端口
