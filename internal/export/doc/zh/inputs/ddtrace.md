---
title     : 'DDTrace'
summary   : '接收 DDTrace 的 APM 数据'
__int_icon: 'icon/ddtrace'
tags      :
  - 'DDTRACE'
  - '链路追踪'
dashboard :
  - desc  : '暂无'
    path  : '-'
monitor   :
  - desc  : '暂无'
    path  : '-'
---


{{.AvailableArchs}}

---

DataKit 的 `ddtrace` 采集器是一个 **DataDog Trace 协议接收端**：应用内的 DDTrace SDK 或 Java Agent 将 trace 通过 HTTP 发送给 DataKit，DataKit 再解析、处理并上报这些数据。它不会替应用安装 SDK，也不会替代应用侧的插桩。

数据链路如下：

应用代码 → DDTrace SDK / Java Agent → DataKit HTTP 接收端 → <<<custom_key.brand_name>>>

链路、Profiling 数据和运行时/JMX 指标使用的接收端不同：

- Trace 由本采集器接收，默认使用 DataKit HTTP 端口 `9529`；
- Profiling 需要单独开启 [Profiling 采集器](profile.md)；
- JMX、runtime metrics 等 DogStatsD 指标需要单独开启 [StatsD 采集器](statsd.md)，通常使用 `8125` 端口。

## 使用前请确认 {#overview}

1. DataKit 已安装并已启用本采集器；应用侧也已按所用语言完成 SDK 或 Agent 的插桩。
1. 应用与 DataKit 的网络连通。DataKit 默认仅监听 `localhost:9529`；应用不在同一主机或同一 Pod 时，需调整 [HTTP 服务监听地址](../datakit/datakit-conf.md#config-http-server) 并通过防火墙、Kubernetes Service 或 NetworkPolicy 限制访问范围。
1. 显式将应用侧的 trace 目标设置为 DataKit 地址和端口，例如 `DD_AGENT_HOST=<datakit-host>`、`DD_TRACE_AGENT_PORT=9529`。上游 Datadog Agent 的常见默认端口是 `8126`，不要依赖默认值，也不要把 StatsD 的 `8125` 当作 trace 端口。
1. 为服务设置稳定的 `DD_SERVICE`、`DD_ENV` 和 `DD_VERSION`；这三个维度决定服务、环境和版本在链路视图中的归属。

## 按语言接入 {#doc-example}

<!-- markdownlint-disable MD046 MD032 MD030 -->
<div class="grid cards" markdown>
-   :fontawesome-brands-python: **Python**

    ---

    [SDK :material-download:](https://github.com/DataDog/dd-trace-py){:target="_blank"} ·
    [:octicons-book-16: 文档](https://docs.datadoghq.com/tracing/setup_overview/setup/python?tab=containers){:target="_blank"} ·
    [:octicons-arrow-right-24: 示例](ddtrace-python.md)

-   :material-language-java: **Java**

    ---

    [SDK :material-download:](https://static.<<<custom_key.brand_main_domain>>>/dd-image/dd-java-agent.jar){:target="_blank"} ·
    [:octicons-book-16: 文档](https://docs.datadoghq.com/tracing/setup_overview/setup/java?tab=containers){:target="_blank"} ·
    [:octicons-arrow-right-24: 示例](ddtrace-java.md)

-   :material-language-ruby: **Ruby**

    ---

    [SDK :material-download:](https://github.com/DataDog/dd-trace-rb){:target="_blank"} ·
    [:octicons-book-16: 文档](https://docs.datadoghq.com/tracing/setup_overview/setup/ruby){:target="_blank"} ·
    [:octicons-arrow-right-24: 示例](ddtrace-ruby.md)

-   :fontawesome-brands-golang: **Golang**

    ---

    [SDK :material-download:](https://github.com/DataDog/dd-trace-go){:target="_blank"} ·
    [:octicons-book-16: 文档](https://docs.datadoghq.com/tracing/setup_overview/setup/go?tab=containers){:target="_blank"} ·
    [:octicons-arrow-right-24: 示例](ddtrace-golang.md)

-   :material-language-php: **PHP**

    ---

    [SDK :material-download:](https://github.com/DataDog/dd-trace-php){:target="_blank"} ·
    [:octicons-book-16: 文档](https://docs.datadoghq.com/tracing/setup_overview/setup/php?tab=containers){:target="_blank"} ·
    [:octicons-arrow-right-24: 示例](ddtrace-php.md)

-   :fontawesome-brands-node-js: **NodeJS**

    ---

    [SDK :material-download:](https://github.com/DataDog/dd-trace-js){:target="_blank"} ·
    [:octicons-book-16: 文档](https://docs.datadoghq.com/tracing/setup_overview/setup/nodejs?tab=containers){:target="_blank"} ·
    [:octicons-arrow-right-24: 示例](ddtrace-nodejs.md)

-   :material-language-cpp: **C++（兼容模式）**

    ---

    [SDK :material-download:](https://github.com/opentracing/opentracing-cpp){:target="_blank"} ·
    [:octicons-book-16: 文档](https://docs.datadoghq.com/tracing/setup_overview/setup/cpp?tab=containers){:target="_blank"} ·
    [:octicons-arrow-right-24: 示例](ddtrace-cpp.md)

-   :material-dot-net:

    ---

    [SDK :material-download:](https://github.com/DataDog/dd-trace-dotnet){:target="_blank"} ·
    [:octicons-book-16: 文档](https://docs.datadoghq.com/tracing/trace_collection/automatic_instrumentation/dd_libraries/dotnet-framework?tab=windows){:target="_blank"} ·
    [:octicons-book-16: .Net Core 文档](https://docs.datadoghq.com/tracing/trace_collection/automatic_instrumentation/dd_libraries/dotnet-core?tab=windows){:target="_blank"}
</div>

???+ info

    <<<custom_key.brand_name>>> 提供了一个扩展版 Java Agent，用于补充部分框架和细粒度采集能力。该扩展只在使用<<<custom_key.brand_name>>>版 JAR 时生效；配置和数据暴露风险请先阅读 [Java 扩展说明](ddtrace-ext-java.md) 与[更新日志](ddtrace-ext-changelog.md)。

## 配置 {#config}

本节配置的是 **DataKit 接收端**。应用侧 SDK 的地址、服务名、采样等配置仍需在应用的启动参数或环境变量中设置，具体见对应语言文档。

=== "主机安装"

    进入 DataKit 安装目录下的 `conf.d/samples` 目录，复制 `{{.InputName}}.conf.sample` 并命名为 `{{.InputName}}.conf`。示例如下：

    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```

    配置好后，[重启 DataKit](../datakit/datakit-service-how-to.md#manage-service) 即可。

=== "Kubernetes"

    可通过 [ConfigMap 方式注入采集器配置](../datakit/datakit-daemonset-deploy.md#configmap-setting) 或 [配置 ENV_DATAKIT_INPUTS](../datakit/datakit-daemonset-deploy.md#env-setting) 开启采集器。

    也支持以环境变量的方式修改配置参数（需要在 ENV_DEFAULT_ENABLED_INPUTS 中加为默认采集器）：

{{ CodeBlock .InputENVSampleZh 4 }}

`customer_tags` 用于将指定的 span `meta` 和 `metrics` 字段提升到一级。`meta` 字符串写入标签，`metrics` 数值写入字段；未匹配的内容继续保留在 `message` 中。普通字段名中的 `.` 会自动转为 `_`，例如 `http.route` 会变为 `http_route`。正则必须以 `reg:` 开头，并使用 Go 正则表达式；例如 `reg:^key_.*$` 匹配所有以 `key_` 开头的字段。请先在测试环境验证正则，非法表达式会使采集器无法正常初始化。

`sampling_priority_drop_excludes` 用于配置不在 DDTrace 早期判断处删除的上游拒绝采样优先级，支持 `-3`、`-1` 和 `0`。默认为空数组，保持对这三个值的现有删除行为；例如配置 `[0]` 后，priority=0 的 trace 将进入 DataKit 后续过滤和采样流程，而不是强制保留。该配置不能恢复上游未发送的 span，并且可能明显增加上报数据量。

DataKit 自身指标 `datakit_input_ddtrace_sampling_priority_trace_total` 和 `datakit_input_ddtrace_sampling_priority_span_total` 通过 `priority`、`action` (`drop`/`bypass`) 和 `service` 标签记录这一决策。它们能证明 DataKit 收到并删除或放行了多少数据，但不能证明 priority=0 的值一定来自 W3C `traceparent`。

### 多线路工具串联注意事项 {#trace_propagator}

DDTrace 传统 Trace ID 为 64 位整数；W3C `tracecontext` 使用 128 位、32 个十六进制字符的 Trace ID。DDTrace 载荷会把高 64 位放在 `_dd.p.tid` 中，DataKit 需要据此重建完整的 128 位 ID。

请让调用链上的 SDK 使用一致的透传协议，并按实际协议配置 DataKit：

- 使用 `tracecontext` 与 OpenTelemetry 串联时，启用 `compatible_otel=true`，以十六进制格式输出 span 和父 span ID；`trace_128_bit_id` 默认已启用，用于拼接 `_dd.p.tid` 与低 64 位 Trace ID。
- 使用 `b3multi` 且上游发送 64 位十六进制 Trace ID 时，启用 `trace_id_64_bit_hex=true`。
- 变更协议后，用一个跨服务请求验证 Trace ID 与 parent ID 是否连续；协议不一致通常表现为服务拓扑断开，而不是 DataKit 收不到数据。

更多协议组合见[多链路串联](tracing-propagator.md){:target="_blank"}。

???+ info

    - `compatible_otel`：将 `span_id` 和 `parent_id` 输出为十六进制字符串。
    - `trace_128_bit_id`：将 `meta` 中的 `_dd.p.tid` 与低 64 位 `trace_id` 拼接为 32 位十六进制字符串；默认值为 `true`。
    - `trace_id_64_bit_hex`：将上游 64 位十六进制 `trace_id` 按十六进制解析。

### 注入 Pod 和 Node 信息 {#add-pod-node-info}

当应用部署在 Kubernetes 等容器环境时，可通过 Downward API 将 Pod、Namespace 和 Node 信息写入 `DD_TAGS`。这些字段会随应用侧 span 发送；如需在链路列表中作为一级标签筛选，再在 DataKit 的 `customer_tags` 中显式添加相应字段。

```yaml hl_lines="21-30"
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-app
spec:
  selector:
    matchLabels:
      app: my-app
  replicas: 3
  template:
    metadata:
      labels:
        app: my-app
        service: my-service
    spec:
      containers:
        - name: my-app
          image: my-app:v0.0.1
          env:
            - name: POD_NAME
              valueFrom:
                fieldRef:
                  fieldPath: metadata.name
            - name: NODE_NAME
              valueFrom:
                fieldRef:
                  fieldPath: spec.nodeName
            - name: POD_NAMESPACE
              valueFrom:
                fieldRef:
                  fieldPath: metadata.namespace
            - name: DD_TAGS
              value: pod_name:$(POD_NAME),pod_namespace:$(POD_NAMESPACE),host:$(NODE_NAME)
            - name: DD_SERVICE
              valueFrom:
                fieldRef:
                  fieldPath: metadata.labels['service']
```

变量替换只能引用同一容器中**先定义**的环境变量，因此应先定义 `POD_NAME`、`POD_NAMESPACE` 和 `NODE_NAME`，再定义 `DD_TAGS`。

应用启动后，进入对应的 Pod，我们可以验证 ENV 是否生效：

```shell
$ env | grep DD_
...
```

一旦注入成功，在最终的 Span 数据中，我们就能看到该 Span 所处的 Pod 以及 Node 名称。

---

???+ warning

    - 不要修改这里的 `endpoints` 列表（除非明确知道配置逻辑和效果）。

    ```toml
    endpoints = ["/v0.3/traces", "/v0.4/traces", "/v0.5/traces"]
    ```

    - DataKit 的 `[inputs.ddtrace.sampler]` 是**接收端采样**，与 SDK 的 `DD_TRACE_SAMPLE_RATE` 相互独立。要让接收端不再采样，请删除或完整注释该表；若保留该表，必须显式设置采样率：

    ``` toml
    # [inputs.{{.InputName}}.sampler]
    # sampling_rate = 1.0
    ```

    `sampling_rate = 1.0` 表示保留全部 trace。不要只注释 `sampling_rate` 而保留表头，否则该表会按零值处理并丢弃所有 trace。默认情况下错误 trace 会绕过接收端采样；如配置 `omit_err_status`，对应 HTTP 错误状态才可能被过滤。

<!-- markdownlint-enable MD046 MD032 MD030 -->

### HTTP 设置 {#http}

DataKit 默认监听 `localhost:9529`。若 trace 来自远端主机或其他 Pod，需要在 `datakit.conf` 的 `[http_api]` 中设置应用可达的监听地址，例如：

```toml
[http_api]
  listen = "0.0.0.0:9529"
```

仅在受控网络中使用该配置：trace 端点不应直接暴露到公网，生产环境应使用防火墙、安全组、Kubernetes Service/NetworkPolicy 或 TLS 进行访问控制。DataKit 默认接收 `/v0.3/traces`、`/v0.4/traces` 和 `/v0.5/traces`，除非明确了解兼容性影响，否则不要修改 `endpoints`。

如果有 DDTrace 数据发送给 DataKit，那么在 [DataKit 的 monitor](../datakit/datakit-monitor.md) 上能看到：

<figure markdown>
  ![input-ddtrace-monitor](https://static.<<<custom_key.brand_main_domain>>>/images/datakit/input-ddtrace-monitor.png){ width="800" }
  <figcaption> DDtrace 将数据发送给了 /v0.4/traces 接口</figcaption>
</figure>

### 开启磁盘缓存 {#disk-cache}

磁盘缓存用于在瞬时流量较高时延后处理 HTTP 请求体，降低内存和处理峰值；它不是长期归档，也不能替代应用侧采样。目录必须可写、具备足够空间，并应位于容器重建后仍可保留的卷上（如确有恢复需求）。`capacity` 单位为 MiB。

``` toml
[inputs.{{.InputName}}.storage]
  path = "/path/to/ddtrace-disk-storage"
  capacity = 5120
```

### DDtrace SDK 配置 {#sdk}

配置完采集器后，再配置 SDK。不同语言对变量的支持和优先级略有差异；下列变量是通用概念，最终以所用语言 SDK 的版本文档为准。若 SDK 支持 `DD_TRACE_AGENT_URL`，该 URL 通常优先于主机和端口配置，避免同时设置相互冲突的值。

### 环境变量设置 {#sdk-envs}

| 变量 | 用途 | 使用建议 |
| --- | --- | --- |
| `DD_AGENT_HOST`、`DD_TRACE_AGENT_PORT` | Trace 接收端地址和端口 | 指向 DataKit，例如 `datakit-service:9529`；不要误用 StatsD 端口 `8125`。 |
| `DD_SERVICE`、`DD_ENV`、`DD_VERSION` | 服务、环境和版本标识 | 在所有服务中使用稳定、可检索的值。 |
| `DD_TAGS` | 应用侧全局标签 | 使用 `key:value` 对；避免写入令牌、请求体、个人信息等敏感数据。 |
| `DD_TRACE_SAMPLE_RATE` | SDK 侧采样率 | `0.0` 到 `1.0`；优先在源头控制高流量。 |
| `DD_TRACE_ENABLED` | 是否启用插桩/trace 发送 | 具体行为因语言而异，排障时确认没有被设为 `false`。 |
| `DD_TRACE_STARTUP_LOGS`、`DD_TRACE_DEBUG` | SDK 启动与调试日志 | 仅在排障期间临时开启，避免增加日志量或暴露配置细节。 |

除了在应用初始化时设置项目名，环境名以及版本号外，还可通过如下两种方式设置：

- 通过命令行注入环境变量

```shell
DD_TAGS="project:your_project_name,env=test,version=v1" ddtrace-run python app.py
```

- 在 *ddtrace.conf* 中直接配置接收端标签。这种方式会影响所有进入该 DataKit 的 DDTrace 数据，适合注入统一的部署标签；不要用它表达某个应用独有的服务信息：

```toml
# tags is ddtrace configed key value pairs
[inputs.{{.InputName}}.tags]
  some_tag = "some_value"
  more_tag = "some_other_value"
```

### APMTelemetry {#apm_telemetry}

[:octicons-tag-24: Version-1.35.0](../datakit/changelog.md#cl-1.35.0) · [:octicons-beaker-24: Experimental](../datakit/index.md#experimental)

Java Agent 可以通过 `/telemetry/proxy/api/v2/apmtelemetry` 上报启动配置、心跳、依赖与已加载集成等元数据。DataKit 默认接收该路由；如不需要，可关闭 `apmtelemetry_route_enable`。数据可在 <<<custom_key.brand_name>>> 基础设施的资源目录中查看，适合排查启动命令、依赖版本和探针加载情况。

语言不同和版本不同数据可能会有很大的差异，以实际收到的数据为准。

### 固定提取 tag {#add-tags}

从 DataKit 版本 [1.21.0](../datakit/changelog.md#cl-1.21.0) 开始，不再把 `Span.Meta` 的所有字段提升为一级标签，而是仅提取下列常用字段，以控制标签基数和索引成本。

以下是可能会提取出的标签列表：

| 原始 Meta 字段          | 提取出来的字段名            | 说明                     |
|:--------------------|:--------------------|:-----------------------|
| `http.url`          | `http_url`          | HTTP 请求完整路径            |
| `http.hostname`     | `http_hostname`     | hostname               |
| `http.route`        | `http_route`        | 路由                     |
| `http.status_code`  | `http_status_code`  | 状态码                    |
| `http.method`       | `http_method`       | 请求方法                   |
| `http.client_ip`    | `http_client_ip`    | 客户端 IP                 |
| `sampling.priority` | `sampling_priority` | 采样                     |
| `span.kind`         | `span_kind`         | span 类型                |
| `error`             | `error`             | 是否错误                   |
| `runtime.name`      | `runtime_name`      | 运行时名称                  |
| `dd.version`        | `dd_version`        | agent 版本               |
| `error.message`     | `error_message`     | 错误信息                   |
| `error.stack`       | `error_stack`       | 堆栈信息                   |
| `error.type`        | `error_type`        | 错误类型                   |
| `system.pid`        | `pid`               | pid                    |
| `error.msg`         | `error_message`     | 错误信息                   |
| `project`           | `project`           | project                |
| `version`           | `version`           | 版本                     |
| `env`               | `env`               | 环境                     |
| `host`              | `host`              | tag 中的主机名              |
| `pod_name`          | `pod_name`          | tag 中的 pod 名称          |
| `pod_namespace`     | `pod_namespace`     | tag 中的 pod 名称          |
| `_dd.base_service`  | `_dd_base_service`  | 上级服务                   |
| `peer.hostname`     | `db_host`           | 可能是 IP 或者域名，这取决于配置     |
| `db.type`           | `db_system`         | 数据库类型： mysql oracle 等等 |
| `db.instance`       | `db_name`           | 数据库名称                  |
| `out.host`          | `out_host`          | 链接中间件的 Host            |
| `dd_ext_version`    | `sdk_version`       | SDK 扩展版本号              |
| `language`          | `sdk_language`      | SDK 语言                 |

未在列表中的字段仍保留在 span 的 `meta` 中，可在链路详情中查看；是否能作为一级标签筛选取决于界面与索引配置。

从 DataKit 版本 [1.22.0](../datakit/changelog.md#cl-1.22.0) 起，可通过 `customer_tags` 增加白名单字段。提取后字段名中的 `.` 会变成 `_`；请只添加稳定且低基数的字段，避免把用户 ID、请求 ID 等高基数字段提升为标签。

### 常见排障路径 {#troubleshooting}

| 现象 | 优先检查项 |
| --- | --- |
| 完全没有 trace | DataKit 的 `ddtrace` 是否启用；应用是否真的加载 SDK/Agent；`DD_AGENT_HOST`、`DD_TRACE_AGENT_PORT`、Service、NetworkPolicy 和防火墙是否正确。 |
| 只有部分服务或调用链断开 | 上下游的透传协议是否一致；`tracecontext`/B3 对应的 DataKit ID 兼容开关是否正确。 |
| JVM 指标或 Profiling 缺失 | 它们不由本采集器接收；分别检查 `statsd` 或 `profile` 采集器和对应端口。 |
| 数据量或资源占用过高 | 先缩小应用侧采样，再评估接收端采样、`trace_max_spans`、`max_trace_body_mb` 与磁盘缓存。 |

## 数据采集字段说明 {#collected-data}

### 链路 {#tracing}

<!-- markdownlint-disable MD024 -->
{{range $i, $m := .Measurements}}

{{if eq $m.Type "tracing"}}

#### `{{$m.Name}}`

{{$m.DescZh}}

{{$m.MarkdownTable}}
{{end}}

{{end}}

### 指标 {#metric}

{{range $i, $m := .Measurements}}

{{if eq $m.Type "metric"}}

#### `{{$m.Name}}`

{{$m.DescZh}}

{{$m.MarkdownTable}}
{{end}}

{{end}}

### 资源对象 {#custom-object}

扩展版 Java Agent 启动后可上报自身配置、集成列表、依赖关系和服务元数据。当前该类资源对象仅适用于 Java Agent，常见事件如下：

- `app_client_configuration_change` 其中包含 Agent 的配置信息
- `app_dependencies_loaded` 依赖列表，包括包名和版本信息
- `app_integrations_change` 集成列表，包括包名和是否开启探针
- 其他主机信息和服务等信息

{{range $i, $m := .Measurements}}

{{if eq $m.Type "custom_object"}}

#### `{{$m.Name}}`

{{$m.DescZh}}

{{$m.MarkdownTable}}
{{end}}

{{end}}

<!-- markdownlint-enable MD024 -->

## 延伸阅读 {#more-reading}

- [DataKit Tracing 字段定义](datakit-tracing-struct.md)
- [DataKit 通用 Tracing 数据采集说明](datakit-tracing.md)
- [正确使用正则表达式来配置](../datakit/datakit-input-conf.md#debug-regex)
- [多链路串联](tracing-propagator.md)
- [Java 接入与异常说明](ddtrace-java.md)
- [DDTrace 采样策略和多链路工具串联注意事项](tracing-sample.md)
