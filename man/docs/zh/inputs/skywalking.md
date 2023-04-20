{{.CSS}}
# SkyWalking
---

{{.AvailableArchs}}

---

Datakit 内嵌的 SkyWalking Agent 用于接收，运算，分析 Skywalking Tracing 协议数据。

## SkyWalking 文档 {#doc}

> 最新的 Datakit Skywalking 实现支持所有 8.x.x 的 Skywalking APM Agent

- [Quickstart](https://skywalking.apache.org/docs/skywalking-showcase/latest/readme/){:target="_blank"}
- [Docs](https://skywalking.apache.org/docs/){:target="_blank"}
- [Clients Download](https://skywalking.apache.org/downloads/){:target="_blank"}
- [Souce Code](https://github.com/apache/skywalking){:target="_blank"}

## 配置 SkyWalking Client {#client-config}

打开文件 /path_to_skywalking_agent/config/agent.config 进行配置

```conf
# The service name in UI
agent.service_name=${SW_AGENT_NAME:your-service-name}
# Backend service addresses.
collector.backend_service=${SW_AGENT_COLLECTOR_BACKEND_SERVICES:<datakit-ip:skywalking-agent-port>}
```

## 配置 SkyWalking Agent {#agent-config}

=== "主机安装"

    进入 DataKit 安装目录下的 `conf.d/{{.Catalog}}` 目录，复制 `{{.InputName}}.conf.sample` 并命名为 `{{.InputName}}.conf`。示例如下：

    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```

    Datakit Skywalking Agent 目前支持 HTTP 协议和 GRPC 协议两种网络传输方式。

    /v3/profiling 接口目前只作为兼容性接口使用, profiling 数据并不上报数据中心。

    通过 HTTP 协议传输
    ```toml
      ## Skywalking HTTP endpoints for tracing, metric, logging and profiling.
      ## NOTE: DO NOT EDIT.
      endpoints = ["/v3/trace", "/v3/metric", "/v3/logging", "/v3/profiling"]
    ```

    通过 GRPC 协议传输
    ```toml
      ## Skywalking GRPC server listening on address.
      address = "localhost:11800"
    ```

    以下所有数据采集，默认会追加名为 `host` 的全局 tag（tag 值为 DataKit 所在主机名），也可以在配置中通过 `[inputs.{{.InputName}}.tags]` 指定其它标签：

    ```toml
      [inputs.{{.InputName}}.tags]
      # some_tag = "some_value"
      # more_tag = "some_other_value"
      # ...
    ```

=== "Kubernetes 内安装"

    目前可以通过 [ConfigMap 方式注入采集器配置](datakit-daemonset-deploy.md#configmap-setting)来开启采集器。

## 启动 Java Client {#start-java}

```command
  java -javaagent:/path/to/skywalking/agent -jar /path/to/your/service.jar
```

## 将日志发送到 Datakit {#logging}
- log4j2

toolkit 依赖包添加到 maven 或者 gradle 中。
```xml
	<dependency>
    <groupId>org.apache.skywalking</groupId>
    <artifactId>apm-toolkit-log4j-2.x</artifactId>
    <version>{project.release.version}</version>
	</dependency>
```

通过 grpc 协议发送出去：
```xml
  <GRPCLogClientAppender name="grpc-log">
    <PatternLayout pattern="%d{HH:mm:ss.SSS} %-5level %logger{36} - %msg%n"/>
  </GRPCLogClientAppender>
```

其他：

- [log4j-1.x](https://github.com/apache/skywalking-java/blob/main/docs/en/setup/service-agent/java-agent/Application-toolkit-log4j-1.x.md){:target="_blank"}
- [logback-1.x](https://github.com/apache/skywalking-java/blob/main/docs/en/setup/service-agent/java-agent/Application-toolkit-logback-1.x.md){:target="_blank"}


## SkyWalking JVM 指标集 {#jvm-measurements}

{{ range $i, $m := .Measurements }}

{{$m.Desc}}

- 标签

{{$m.TagsMarkdownTable}}

- 指标列表

{{$m.FieldsMarkdownTable}}

{{ end }}


=== "Unordered list"

    * Sed sagittis eleifend rutrum
    * Donec vitae suscipit est
    * Nulla tempor lobortis orci

=== "Ordered list"

    1. Sed sagittis eleifend rutrum
    2. Donec vitae suscipit est
    3. Nulla tempor lobortis orci
