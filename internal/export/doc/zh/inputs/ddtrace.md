---
title     : 'DDTrace'
summary   : '接收 DDTrace 的 APM 数据'
__int_icon: 'icon/ddtrace'
dashboard :
  - desc  : '暂无'
    path  : '-'
monitor   :
  - desc  : '暂无'
    path  : '-'
---

<!-- markdownlint-disable MD025 -->
# DDTrace
<!-- markdownlint-enable -->

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

    [SDK :material-download:](https://static.guance.com/dd-image/dd-java-agent.jar){:target="_blank"} ·
    [:octicons-book-16: 文档](https://docs.datadoghq.com/tracing/setup_overview/setup/java?tab=containers){:target="_blank"} ·
    [:octicons-arrow-right-24: 示例](ddtrace-java.md)

-   :material-language-ruby: __Ruby__

    ---

    [SDK :material-download:](https://github.com/DataDog/dd-trace-rb){:target="_blank"} ·
    [:octicons-book-16: 文档](https://docs.datadoghq.com/tracing/setup_overview/setup/ruby){:target="_blank"} ·
    [:octicons-arrow-right-24: 示例](ddtrace-java.md)

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
    [:octicons-book-16: 文档](https://docs.datadoghq.com/tracing/setup_overview/setup/dotnet-framework?tab=windows){:target="_blank"} ·
    [:octicons-book-16: .Net Core 文档](https://docs.datadoghq.com/tracing/setup_overview/setup/dotnet-framework?tab=windows){:target="_blank"}
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

    目前可以通过 [ConfigMap 方式注入采集器配置](../datakit/datakit-daemonset-deploy.md#configmap-setting)来开启采集器。

    在 Kubernetes 中支持的环境变量如下表：

    | 环境变量名                             | 类型        | 示例                                                                             |
    | -------------------------------------- | ----------- | -------------------------------------------------------------------------------- |
    | `ENV_INPUT_DDTRACE_ENDPOINTS`          | JSON string | `["/v0.3/traces", "/v0.4/traces", "/v0.5/traces"]`                               |
    | `ENV_INPUT_DDTRACE_IGNORE_TAGS`        | JSON string | `["block1", "block2"]`                                                           |
    | `ENV_INPUT_DDTRACE_KEEP_RARE_RESOURCE` | bool        | true                                                                             |
    | `ENV_INPUT_DDTRACE_COMPATIBLE_OTEL`    | bool        | true                                                                             |
    | `ENV_INPUT_DDTRACE_DEL_MESSAGE`        | bool        | true                                                                             |
    | `ENV_INPUT_DDTRACE_OMIT_ERR_STATUS`    | JSON string | `["404", "403", "400"]`                                                          |
    | `ENV_INPUT_DDTRACE_CLOSE_RESOURCE`     | JSON string | `{"service1":["resource1"], "service2":["resource2"], "service3":["resource3"]}` |
    | `ENV_INPUT_DDTRACE_SAMPLER`            | float       | 0.3                                                                              |
    | `ENV_INPUT_DDTRACE_TAGS`               | JSON string | `{"k1":"v1", "k2":"v2", "k3":"v3"}`                                              |
    | `ENV_INPUT_DDTRACE_THREADS`            | JSON string | `{"buffer":1000, "threads":100}`                                                 |
    | `ENV_INPUT_DDTRACE_STORAGE`            | JSON string | `{"storage":"./ddtrace_storage", "capacity": 5120}`                              |

???+ attention

    - 不要修改这里的 `endpoints` 列表（除非明确知道配置逻辑和效果）。

    ```toml
    endpoints = ["/v0.3/traces", "/v0.4/traces", "/v0.5/traces"]
    ```

    - 如果要关闭采样（即采集所有数据），采样率字段需做如下设置：

    ``` toml
    # [inputs.ddtrace.sampler]
    # sampling_rate = 1.0
    ```

    不要只注释 `sampling_rate = 1.0` 这一行，必须连同 `[inputs.ddtrace.sampler]` 也一并注释掉，否则采集器会认为 `sampling_rate` 被置为 0.0，从而导致所有数据都被丢弃。

<!-- markdownlint-enable -->

### HTTP 设置 {#http}

如果 Trace 数据是跨机器发送过来的，那么需要设置 [DataKit 的 HTTP 设置](../datakit/datakit-conf.md#config-http-server)。

如果有 DDTrace 数据发送给 Datakit，那么在 [DataKit 的 monitor](../datakit/datakit-monitor.md) 上能看到：

<figure markdown>
  ![](https://static.guance.com/images/datakit/input-ddtrace-monitor.png){ width="800" }
  <figcaption> DDtrace 将数据发送给了 /v0.4/traces 接口</figcaption>
</figure>

### 开启磁盘缓存 {#disk-cache}

如果 Trace 数据量很大，为避免给主机造成大量的资源开销，可以将 Trace 数据临时缓存到磁盘中，延迟处理：

``` toml
[inputs.ddtrace.storage]
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
[inputs.ddtrace.tags]
  some_tag = "some_value"
  more_tag = "some_other_value"
```

### 在代码中添加业务 tag {#add-tags}

在应用代码中，可通过诸如 `span.SetTag(some-tag-key, some-tag-value)`（不同语言方式不同） 这样的方式来设置业务自定义 tag。对于这些业务自定义 tag，可通过在 ddtrace.conf 中配置 `customer_tags` 来识别并提取：

```toml
customer_tags = [
  "order_id",
  "task_id",
  "some.key",  # 被重命名为 some_key
]
```

注意，这些 tag-key 中不能包含英文字符 '.'，带 `.` 的 tag-key 会替换为 `_`。

<!-- markdownlint-disable MD046 -->
???+ attention "应用代码中添加业务 tag 注意事项"

    - 在应用代码中添加了对应的 Tag 后，必须在 *ddtrace.conf* 的 `customer_tags` 中也同步添加对应的 Tag-Key 列表，否则 Datakit 不会对这些业务 Tag 进行提取
    - 在开启了采样的情况下，部分添加了 Tag 的 Span 有可能被舍弃
<!-- markdownlint-enable -->

## 链路字段 {#tracing}

{{range $i, $m := .Measurements}}

{{if eq $m.Type "tracing"}}

### `{{$m.Name}}`

{{$m.Desc}}

- 标签

{{$m.TagsMarkdownTable}}

- 指标列表

{{$m.FieldsMarkdownTable}}
{{end}}

{{end}}

## 延伸阅读 {#more-reading}

- [DataKit Tracing 字段定义](datakit-tracing-struct.md)
- [DataKit 通用 Tracing 数据采集说明](datakit-tracing.md)
- [正确使用正则表达式来配置](../datakit/datakit-input-conf.md#debug-regex)
- [多链路串联](tracing-propagator.md)
