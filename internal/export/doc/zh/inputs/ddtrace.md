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

DDTrace 是 DataDog 开源的 APM 产品，Datakit 内嵌的 DDTrace Agent 用于接收，运算，分析 DataDog Tracing 协议数据。

## DDTrace 文档和示例 {#doc-example}

<!-- markdownlint-disable MD046 MD032 MD030 -->
<div class="grid cards" markdown>
-   :fontawesome-brands-python: __Python__

    ---

    [SDK :material-download:](https://github.com/DataDog/dd-trace-py){:target="_blank"} ·
    [:octicons-book-16: 文档](https://docs.datadoghq.com/tracing/setup_overview/setup/python?tab=containers){:target="_blank"} ·
    [:octicons-arrow-right-24: 示例](ddtrace-python.md)

-   :material-language-java: __Java__

    ---

    [SDK :material-download:](https://static.<<<custom_key.brand_main_domain>>>/dd-image/dd-java-agent.jar){:target="_blank"} ·
    [:octicons-book-16: 文档](https://docs.datadoghq.com/tracing/setup_overview/setup/java?tab=containers){:target="_blank"} ·
    [:octicons-arrow-right-24: 示例](ddtrace-java.md)

-   :material-language-ruby: __Ruby__

    ---

    [SDK :material-download:](https://github.com/DataDog/dd-trace-rb){:target="_blank"} ·
    [:octicons-book-16: 文档](https://docs.datadoghq.com/tracing/setup_overview/setup/ruby){:target="_blank"} ·
    [:octicons-arrow-right-24: 示例](ddtrace-ruby.md)

-   :fontawesome-brands-golang: __Golang__

    ---

    [SDK :material-download:](https://github.com/DataDog/dd-trace-go){:target="_blank"} ·
    [:octicons-book-16: 文档](https://docs.datadoghq.com/tracing/setup_overview/setup/go?tab=containers){:target="_blank"} ·
    [:octicons-arrow-right-24: 示例](ddtrace-golang.md)

-   :material-language-php: __PHP__

    ---

    [SDK :material-download:](https://github.com/DataDog/dd-trace-php){:target="_blank"} ·
    [:octicons-book-16: 文档](https://docs.datadoghq.com/tracing/setup_overview/setup/php?tab=containers){:target="_blank"} ·
    [:octicons-arrow-right-24: 示例](ddtrace-php.md)

-   :fontawesome-brands-node-js: __NodeJS__

    ---

    [SDK :material-download:](https://github.com/DataDog/dd-trace-js){:target="_blank"} ·
    [:octicons-book-16: 文档](https://docs.datadoghq.com/tracing/setup_overview/setup/nodejs?tab=containers){:target="_blank"} ·
    [:octicons-arrow-right-24: 示例](ddtrace-nodejs.md)

-   :material-language-cpp:

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

???+ tip

    我们对 DDTrace 做了一些[功能扩展](ddtrace-ext-changelog.md)，便于支持更多的主流框架和更细粒度的数据追踪。

## 配置 {#config}

=== "主机安装"

    进入 DataKit 安装目录下的 `conf.d/{{.Catalog}}` 目录，复制 `{{.InputName}}.conf.sample` 并命名为 `{{.InputName}}.conf`。示例如下：

    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```

    配置好后，[重启 DataKit](../datakit/datakit-service-how-to.md#manage-service) 即可。

=== "Kubernetes"

    可通过 [ConfigMap 方式注入采集器配置](../datakit/datakit-daemonset-deploy.md#configmap-setting) 或 [配置 ENV_DATAKIT_INPUTS](../datakit/datakit-daemonset-deploy.md#env-setting) 开启采集器。

    也支持以环境变量的方式修改配置参数（需要在 ENV_DEFAULT_ENABLED_INPUTS 中加为默认采集器）：

{{ CodeBlock .InputENVSampleZh 4 }}

### 多线路工具串联注意事项 {#trace_propagator}

DDTrace 数据结构中 TraceID 是 uint64 类型，在使用透传协议 `tracecontext` 时，DDTrace 链路详情内部会增加一个 `_dd.p.tid:67c573cf00000000` 原因是因为 `tracecontext` 协议
中的 `trace_id` 是 128 位 16 进制编码的字符串，为了兼容只能增加了一个高位的 tag 。

DDTrace 目前支持的透传协议有：`datadog/b3multi/tracecontext` ，有两种情况需要注意：

- 当使用 `tracecontext` 时，由于链路 ID 为 128 位需要将配置中的 `compatible_otel=true` 和 `trace_128_bit_id` 开关打开。
- 当使用 `b3multi` 时，需要注意 `trace_id` 的长度，如果为 64 位的 hex 编码，需要将配置文件中的 `trace_id_64_bit_hex=true` 打开。
- 更多的透传协议及工具使用请查看： [多链路串联](tracing-propagator.md){:target="_blank"}


???+ tip

    compatible_otel 作用：将 span_id 和 parent_id 转成 16 进制的字符串。
    trace_128_bit_id 作用：将 meta 中的 "_dd.p.tid" 加上 trace_id 组合成一个长度为 32 的 16 进制编码的字符串。
    trace_id_64_bit_hex 作用：将 64 位的 trace_id 转成 16 进制编码的字符串。

### 注入 Pod 和 Node 信息 {#add-pod-node-info}

当应用在 Kubernetes 等容器环境部署时，我们可以在在最终的 Span 数据上追加 Pod/Node 信息，通过修改应用的 Yaml 即可，下面是一个 Kubernetes Deployment 的 yaml 示例：

```yaml
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
            - name: POD_NAME    # <------
              valueFrom:
                fieldRef:
                  fieldPath: metadata.name
            - name: NODE_NAME
              valueFrom:
                fieldRef:
                  fieldPath: spec.nodeName
            - name: DD_SERVICE
              valueFrom:
                fieldRef:
                  fieldPath: metadata.labels['service']
            - name: DD_TAGS
              value: pod_name:$(POD_NAME),host:$(NODE_NAME)
```

注意，此处要先定义 `POD_NAME` 和 `NODE_NAME`，然后再将它们嵌入到到 DDTrace 专用的环境变量中。

应用启动后，进入对应的 Pod，我们可以验证 ENV 是否生效：

```shell
$ env | grep DD_
...
```

一旦注入成功，在最终的 Span 数据中，我们就能看到该 Span 所处的 Pod 以及 Node 名称。

---

???+ attention

    - 不要修改这里的 `endpoints` 列表（除非明确知道配置逻辑和效果）。

    ```toml
    endpoints = ["/v0.3/traces", "/v0.4/traces", "/v0.5/traces"]
    ```

    - 如果要关闭采样（即采集所有数据），采样率字段需做如下设置：

    ``` toml
    # [inputs.{{.InputName}}.sampler]
    # sampling_rate = 1.0
    ```

    不要只注释 `sampling_rate = 1.0` 这一行，必须连同 `[inputs.{{.InputName}}.sampler]` 也一并注释掉，否则采集器会认为 `sampling_rate` 被置为 0.0，从而导致所有数据都被丢弃。

<!-- markdownlint-enable -->

### HTTP 设置 {#http}

如果 Trace 数据是跨机器发送过来的，那么需要设置 [DataKit 的 HTTP 设置](../datakit/datakit-conf.md#config-http-server)。

如果有 DDTrace 数据发送给 Datakit，那么在 [DataKit 的 monitor](../datakit/datakit-monitor.md) 上能看到：

<figure markdown>
  ![input-ddtrace-monitor](https://static.<<<custom_key.brand_main_domain>>>/images/datakit/input-ddtrace-monitor.png){ width="800" }
  <figcaption> DDtrace 将数据发送给了 /v0.4/traces 接口</figcaption>
</figure>

### 开启磁盘缓存 {#disk-cache}

如果 Trace 数据量很大，为避免给主机造成大量的资源开销，可以将 Trace 数据临时缓存到磁盘中，延迟处理：

``` toml
[inputs.{{.InputName}}.storage]
  path = "/path/to/ddtrace-disk-storage"
  capacity = 5120
```

### DDtrace SDK 配置 {#sdk}

配置完采集器之后，还可以对 DDtrace SDK 端做一些配置。

### 环境变量设置 {#sdk-envs}

- `DD_TRACE_ENABLED`: Enable global tracer (部分语言平台支持)
- `DD_AGENT_HOST`: DDtrace agent host address
- `DD_TRACE_AGENT_PORT`: DDtrace agent host port
- `DD_SERVICE`: Service name
- `DD_TRACE_SAMPLE_RATE`: Set sampling rate
- `DD_VERSION`: Application version (optional)
- `DD_TRACE_STARTUP_LOGS`: DDtrace logger
- `DD_TRACE_DEBUG`: DDtrace debug mode
- `DD_ENV`: Application env values
- `DD_TAGS`: Application

除了在应用初始化时设置项目名，环境名以及版本号外，还可通过如下两种方式设置：

- 通过命令行注入环境变量

```shell
DD_TAGS="project:your_project_name,env=test,version=v1" ddtrace-run python app.py
```

- 在 _ddtrace.conf_ 中直接配置自定义标签。这种方式会影响所有发送给 Datakit tracing 服务的数据，需慎重考虑：

```toml
# tags is ddtrace configed key value pairs
[inputs.{{.InputName}}.tags]
  some_tag = "some_value"
  more_tag = "some_other_value"
```

## APMTelemetry {#apm_telemetry}

[:octicons-tag-24: Version-1.35.0](../datakit/changelog.md#cl-1.35.0) · [:octicons-beaker-24: Experimental](../datakit/index.md#experimental)

DDTrace 探针启动后，会不断通额外的接口上报服务有关的信息，比如启动配置、心跳、加载的探针列表等信息。可在<<<custom_key.brand_name>>> 基础设施 -> 资源目录 中查看。展示的数据对于排查启动命令和引用的三方库版本问题有帮助。其中还包括主机信息、服务信息、产生的 Span 数信息等。

语言不同和版本不同数据可能会有很大的差异，以实际收到的数据为准。


{{range $i, $m := .Measurements}}

{{if eq $m.Type "object"}}

### `{{$m.Name}}`

{{$m.Desc}}

- 标签（String 类型）

{{$m.TagsMarkdownTable}}

- 指标列表（非 String 类型，或者长 String 类型）

{{$m.FieldsMarkdownTable}}
{{end}}

{{end}}

### 固定提取 tag {#add-tags}

从 DataKit 版本 [1.21.0](../datakit/changelog.md#cl-1.21.0) 开始，黑名单功能废弃，并且不在将 Span.Mate 中全部都提前到一级标签中，而是选择性提取。

以下是可能会提取出的标签列表：

| 原始 Meta 字段      | 提取出来的字段名    | 说明              |
| :------------------ | :------------------ | :---------------  |
| `http.url`          | `http_url`          | HTTP 请求完整路径 |
| `http.hostname`     | `http_hostname`     | hostname          |
| `http.route`        | `http_route`        | 路由              |
| `http.status_code`  | `http_status_code`  | 状态码            |
| `http.method`       | `http_method`       | 请求方法          |
| `http.client_ip`    | `http_client_ip`    | 客户端 IP         |
| `sampling.priority` | `sampling_priority` | 采样              |
| `span.kind`         | `span_kind`         | span 类型         |
| `error`             | `error`             | 是否错误          |
| `dd.version`        | `dd_version`        | agent 版本        |
| `error.message`     | `error_message`     | 错误信息          |
| `error.stack`       | `error_stack`       | 堆栈信息          |
| `error.type`        | `error_type`        | 错误类型          |
| `system.pid`        | `pid`               | pid               |
| `error.msg`         | `error_message`     | 错误信息          |
| `project`           | `project`           | project           |
| `version`           | `version`           | 版本              |
| `env`               | `env`               | 环境              |
| `host`              | `host`              | tag 中的主机名    |
| `pod_name`          | `pod_name`          | tag 中的 pod 名称 |
| `_dd.base_service`  | `_dd_base_service`  | 上级服务          |

在<<<custom_key.brand_name>>>中的链路界面，不在列表中的标签也可以进行筛选。

从 DataKit 版本 [1.22.0](../datakit/changelog.md#cl-1.22.0) 恢复白名单功能，如果有必须要提取到一级标签列表中的标签，可以在 `customer_tags` 中配置。
配置的白名单标签如果是原生的 `message.meta` 中，会使用 `.` 作为分隔符，采集器会进行转换将 `.` 替换成 `_` 。

## 链路字段说明 {#tracing}

{{range $i, $m := .Measurements}}

{{if eq $m.Type "tracing"}}

### `{{$m.Name}}`

{{$m.Desc}}

- 标签（String 类型）

{{$m.TagsMarkdownTable}}

- 指标列表（非 String 类型，或者长 String 类型）

{{$m.FieldsMarkdownTable}}
{{end}}

{{end}}

## 延伸阅读 {#more-reading}

- [DataKit Tracing 字段定义](datakit-tracing-struct.md)
- [DataKit 通用 Tracing 数据采集说明](datakit-tracing.md)
- [正确使用正则表达式来配置](../datakit/datakit-input-conf.md#debug-regex)
- [多链路串联](tracing-propagator.md)
- [Java 接入与异常说明](ddtrace-java.md)
