---
title     : 'DDTrace Java'
summary   : '使用 DDTrace Java Agent 采集 Java 应用链路'
tags      :
  - 'DDTRACE'
  - 'JAVA'
  - '链路追踪'
__int_icon: 'icon/ddtrace'
---

DDTrace Java Agent 利用 JVM 的 `-javaagent` 机制在类加载时自动插桩，因此大多数常见框架无需修改业务代码。本页说明如何将 Java Agent 的 trace 发送到 DataKit；它不包含日志采集、Profiling 或 JMX 指标的全部配置。

## 前置条件 {#requirements}

1. 已安装 DataKit，并启用 [DDTrace 采集器](ddtrace.md)。Java Agent 的 trace 目标应显式设置为 DataKit 的 HTTP 地址与端口 `9529`。
1. 使用 JDK 8 或更高版本，并选择与应用运行时、操作系统和 CPU 架构兼容的 Agent 版本。升级 JDK 或 Agent 前应在预发布环境验证。
1. Agent JAR 对运行用户可读，且 JVM 启动命令能访问该绝对路径。
1. Profiling 与 JMX 指标是独立数据通路：前者需要 [Profiling 采集器](profile.md)，后者需要 [StatsD 采集器](statsd.md)。

<!-- markdownlint-disable MD046 -->
???+ warning

    同一 JVM 可以加载多个 Java Agent，但它们的类转换顺序和兼容性会影响启动与数据质量。不要把两个 DDTrace Java Agent 同时加入启动参数；与其他 APM、安全或字节码 Agent 组合时，务必先压测并验证关键业务路径。
<!-- markdownlint-enable MD046 -->

## 获取 Agent {#dependence}

<!-- markdownlint-disable MD046 -->
=== "<<<custom_key.brand_name>>>扩展版"

    扩展版在上游 Java Agent 的基础上增加部分框架和细粒度采集能力。下载后请固定具体版本并阅读 [Java 扩展说明](ddtrace-ext-java.md) 与[更新日志](ddtrace-ext-changelog.md)：

    ```shell
    wget -O /opt/dd-java-agent.jar \
      'https://static.<<<custom_key.brand_main_domain>>>/dd-image/dd-java-agent.jar'
    ```

=== "Datadog 原生版"

    ```shell
    wget -O /opt/dd-java-agent.jar 'https://dtdg.co/latest-java-tracer'
    ```
<!-- markdownlint-enable MD046 -->

扩展版和原生版的默认值、支持的配置项可能不同。无论使用哪一种，都建议在启动参数中明确写出 DataKit 主机和端口，避免把上游常见的默认端口 `8126` 误用于 DataKit。

## 运行应用 {#instrument}

### 主机应用 {#host-application}

将所有 `-D` 参数放在 `-jar` 之前。以下命令展示最小、可辨识的服务配置：

```shell linenums="1"
java \
  -javaagent:/opt/dd-java-agent.jar \
  -Ddd.service=my-java-service \
  -Ddd.env=production \
  -Ddd.version=1.0.0 \
  -Ddd.agent.host=127.0.0.1 \
  -Ddd.trace.agent.port=9529 \
  -Ddd.logs.injection=true \
  -jar /opt/my-app.jar
```

`dd.service` 的环境变量为 `DD_SERVICE`；不要把它写成 `dd.service.name`。也可以在进程启动环境中设置 `DD_AGENT_HOST`、`DD_TRACE_AGENT_PORT`、`DD_SERVICE`、`DD_ENV` 和 `DD_VERSION`。若设置了 `DD_TRACE_AGENT_URL` / `dd.trace.agent.url`，该 URL 通常优先于 host 与 port；请只保留一种目标配置。

### Kubernetes {#kubernetes}

在 Kubernetes 中推荐使用 [DataKit Operator 注入 DDTrace](../operator-ddtrace.md)，由 Operator 负责向新建 Pod 注入 JAR、卷挂载和语言相关启动配置。修改 Operator 配置或 Pod 注解后，需要重新创建 Pod（例如执行 rollout restart）才会触发 webhook。

也可以由应用镜像自行携带 Agent。以下片段仅在 JAR 已存在于镜像的 `/opt/dd-java-agent.jar` 时有效；**仅设置环境变量而没有加载 JAR 不会产生自动插桩**：

```yaml
containers:
  - name: app
    image: registry.example/my-java-app:1.0.0
    env:
      - name: JAVA_TOOL_OPTIONS
        value: "-javaagent:/opt/dd-java-agent.jar"
      - name: DD_SERVICE
        value: "my-java-service"
      - name: DD_ENV
        value: "production"
      - name: DD_VERSION
        value: "1.0.0"
      - name: DD_AGENT_HOST
        value: "datakit-service.datakit.svc.cluster.local"
      - name: DD_TRACE_AGENT_PORT
        value: "9529"
```

如镜像中没有 JAR，请使用 Operator，或通过受控的 initContainer 与共享卷挂载；不要把未校验的 JAR 下载逻辑放入业务容器启动脚本。

### 开启 Profiling {#instrument-profiling}

Profiling 数据发送到 DataKit 的 Profiling 接收端，而不是 DDTrace trace 接收端。先开启 [Profiling 采集器](profile.md)，再开启 Agent 功能：

```shell
DD_PROFILING_ENABLED=true \
java -javaagent:/opt/dd-java-agent.jar \
  -Ddd.agent.host=127.0.0.1 \
  -Ddd.trace.agent.port=9529 \
  -jar /opt/my-app.jar
```

在生产环境启用前评估性能、存储与合规影响；不要将 Profiling 的缺失误判为 trace 接收端故障。

### 设置采样率 {#instrument-sampling}

应用侧采样率用于在数据产生源头降低流量：

```shell
DD_TRACE_SAMPLE_RATE=0.8 \
java -javaagent:/opt/dd-java-agent.jar \
  -Ddd.agent.host=127.0.0.1 \
  -Ddd.trace.agent.port=9529 \
  -jar /opt/my-app.jar
```

`0.8` 表示 SDK 侧按 80% 采样。它与 DataKit `ddtrace` 采集器的接收端采样独立；先明确在哪一层控制流量，避免两个规则叠加后难以解释数据量。

### 开启 JVM 指标采集 {#instrument-jvm-metrics}

JMXFetch 默认可采集一部分 JVM 指标，但数据通过 DogStatsD 发送。需开启 [StatsD 采集器](statsd.md)，并使用**独立的** StatsD 地址和端口，通常是 UDP `8125`：

```shell
DD_JMXFETCH_ENABLED=true \
DD_JMXFETCH_STATSD_HOST=<YOUR-DATAKIT-HOST> \
DD_JMXFETCH_STATSD_PORT=8125 \
java -javaagent:/opt/dd-java-agent.jar \
  -Ddd.agent.host=<YOUR-DATAKIT-HOST> \
  -Ddd.trace.agent.port=9529 \
  -jar /opt/my-app.jar
```

不要将 `DD_JMXFETCH_STATSD_PORT` 设置为 `9529`。自定义 MBean 指标见 [DDTrace JMX](ddtrace-jmxfetch.md)。

## 参数解释 {#start-options}

下表列出与 DataKit 接入最相关的参数。完整参数、版本要求和默认值以 [Datadog Java 配置文档](https://docs.datadoghq.com/tracing/trace_collection/library_config/java/){:target="_blank"} 为准。

| JVM 属性 | 环境变量 | 用途与 DataKit 接入建议 |
| --- | --- | --- |
| `dd.service` | `DD_SERVICE` | 服务名。显式设置，避免由启动类推断出的名称变化。 |
| `dd.env` | `DD_ENV` | 部署环境，如 `production`、`staging`。 |
| `dd.version` | `DD_VERSION` | 应用版本，用于发布与回归比较。 |
| `dd.trace.agent.url` | `DD_TRACE_AGENT_URL` | 完整 trace 接收 URL，通常优先于 host/port。 |
| `dd.agent.host` | `DD_AGENT_HOST` | DataKit 主机或 Kubernetes Service。 |
| `dd.trace.agent.port` | `DD_TRACE_AGENT_PORT` | Trace 接收端端口。上游常见默认值为 `8126`；DataKit 使用 `9529`。 |
| `dd.trace.sample.rate` | `DD_TRACE_SAMPLE_RATE` | SDK 侧根 trace 采样率，范围 `0.0` 到 `1.0`。 |
| `dd.trace.enabled` | `DD_TRACE_ENABLED` | 控制自动插桩和 trace 生成；排障时确认其不是 `false`。 |
| `dd.logs.injection` | `DD_LOGS_INJECTION` | 向支持的日志框架注入 trace/span 关联字段；仍需单独配置日志采集。 |
| `dd.profiling.enabled` | `DD_PROFILING_ENABLED` | 开启 Continuous Profiling；需要 DataKit `profile` 采集器。 |
| `dd.jmxfetch.enabled` | `DD_JMXFETCH_ENABLED` | 开启 JMXFetch；指标应发送到 DataKit StatsD `8125`。 |
| `dd.trace.startup.logs`、`dd.trace.debug` | `DD_TRACE_STARTUP_LOGS`、`DD_TRACE_DEBUG` | 输出启动配置或调试日志。仅在排障期间临时开启。 |

## 链路错误情况说明 {#error}

自动插桩是否将 span 标记为错误取决于集成和 SDK 配置，不能简单理解为“所有异常或所有 4xx/5xx 都是错误”。通常应注意：

- 未被应用处理并传播到受支持框架边界的异常，通常会写入错误信息与堆栈；
- HTTP 服务端和客户端的错误状态范围可分别配置。上游常见默认行为是服务端 `5xx`、客户端 `4xx` 视为错误；扩展版 Agent 还可能提供额外开关；
- 在业务代码中捕获并消化异常后，自动插桩不一定会将当前 span 标为错误。如该错误对业务仍有意义，应使用所用 SDK 的 API 显式记录错误；
- DataKit 通常将 `error.type`、`error.message` 和 `error.stack` 提取为可查看字段。堆栈和异常消息可能含敏感信息，应按组织的数据治理要求控制访问与保留周期。

排查时先复现一个最小请求，再同时查看应用日志、Agent 启动日志和 trace 详情；不要仅凭 HTTP 状态码推断插桩是否工作。

## 验证与排障 {#verify}

1. 临时设置 `DD_TRACE_STARTUP_LOGS=true`，确认 JVM 日志中 Agent 已加载且目标地址为 DataKit。
1. 调用一个已知会经过 HTTP、数据库或 RPC 的接口，在 DataKit monitor 中确认 `/v0.3/traces`、`/v0.4/traces` 或 `/v0.5/traces` 有请求。
1. Kubernetes 场景同时确认 Pod 模板中的 `JAVA_TOOL_OPTIONS`/启动命令、Agent 文件挂载、`DD_AGENT_HOST` 和 `DD_TRACE_AGENT_PORT`。
1. 若 trace 存在但跨服务断开，检查各服务的透传协议和 [多链路串联](tracing-propagator.md) 配置。

## 更多 {#more-reading}

- [DDTrace 接收端配置、采样与标签](ddtrace.md)
- [<<<custom_key.brand_name>>> Java Agent 扩展](ddtrace-ext-java.md)
- [Java Agent 扩展更新日志](ddtrace-ext-changelog.md)
- [JMXFetch 与自定义 JVM 指标](ddtrace-jmxfetch.md)
- [Profiling 采集器](profile.md)
- [多链路串联](tracing-propagator.md)
- [DDTrace 采样策略](tracing-sample.md)
