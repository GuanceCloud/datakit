---
title     : 'DDTrace Ruby'
summary   : 'DDTrace Ruby 集成'
tags      :
  - 'DDTRACE'
  - 'RUBY'
  - '链路追踪'
__int_icon: 'icon/ddtrace'
---

如需接入 Ruby Continuous Profiling，参见 [Profiling Ruby](profile-ruby.md)。

## 安装依赖 {#dependence}

<!-- cspell:ignore datadog -->
Ruby 的新一代 SDK 使用 `datadog` gem；遗留应用可能仍使用 `ddtrace` gem。两者的加载方式与配置优先级可能不同，因此请先确认当前 Gemfile 和 SDK 版本。完整兼容性和接入说明见 [Datadog Ruby 接入文档](https://docs.datadoghq.com/tracing/trace_collection/dd_libraries/ruby/){:target="_blank"}。

<!-- markdownlint-disable MD046 -->
=== "datadog gem（推荐）"

    在 `Gemfile` 中加入自动插桩入口后执行 `bundle install`：

    ```ruby
    gem "datadog", require: "datadog/auto_instrument"
    ```

=== "ddtrace gem（遗留）"

    对仍在使用 1.x 的应用，保留原有依赖并按该版本的官方文档加载：

    ```ruby
    gem "ddtrace", require: "ddtrace/auto_instrument"
    ```
<!-- markdownlint-enable -->

## 配置 {#config}

Ruby 应用通常通过进程启动环境变量或 `Datadog.configure` 代码块配置。环境变量适合容器和平台部署；代码配置适合需要按运行环境集中管理的应用。不要在两处设置互相冲突的地址或服务标识。完整配置项见 [Datadog Ruby 配置文档](https://docs.datadoghq.com/tracing/trace_collection/library_config/ruby/){:target="_blank"}。

使用 DataKit 作为 trace 接收端时，显式覆盖上游默认的 `127.0.0.1:8126`，改为 DataKit 的地址和 `9529` 端口。例如：

```ruby
Datadog.configure do |c|
  c.agent.host = '127.0.0.1'
  c.agent.port = 9529
  c.service = 'my-ruby-service'
  c.env = 'production'
end
```

也可在进程启动前设置 `DD_AGENT_HOST`、`DD_TRACE_AGENT_PORT`、`DD_SERVICE`、`DD_ENV` 和 `DD_VERSION`。若设置了 `DD_TRACE_AGENT_URL`，它通常优先于 host/port；请只保留一种目标配置。

如不需要 SDK 遥测数据，可在已验证所用 SDK 支持该选项后关闭它，以减少额外诊断信息上报：

```shell
export DD_INSTRUMENTATION_TELEMETRY_ENABLED=false
```

启动应用后访问一个已插桩的路由，并在 DataKit monitor 中确认 trace 请求。排障期间可临时开启 SDK 调试日志；不要长期在生产环境开启详细调试。

## 环境变量支持 {#envs}

下面列出常用 Ruby APM 参数。它们应在 Ruby 进程启动前设置；完整参数与优先级见 [Datadog Ruby 配置文档](https://docs.datadoghq.com/tracing/trace_collection/library_config/ruby/){:target="_blank"}。

- **`DD_AGENT_HOST`**

    **默认值**：`127.0.0.1`

    Trace 接收端主机地址。接入 DataKit 时填写 DataKit 主机或 Kubernetes Service。

- **`DD_TRACE_AGENT_PORT`**

    **上游默认值**：`8126`

    Trace 数据发送端口。接入 DataKit 时必须显式设置为 `9529`。

- **`DD_ENV`**

    **默认值**：`nil`

    设置应用运行环境，比如 `production`、`staging`。

- **`DD_SERVICE`**

    **默认值**：由 SDK 和应用启动方式决定

    设置应用服务名。生产环境请显式设置，避免服务名随入口文件或框架变化。

- **`DD_TAGS`**

    **默认值**：`nil`

    为所有 trace 设置自定义标签，例如 `team:core,layer:api`。不要传递令牌、个人信息或高基数请求标识。

- **`DD_VERSION`**

    **默认值**：`nil`

    设置应用版本号。

- **`DD_TRACE_ENABLED`**

    **默认值**：`true`

    启用或禁用 trace 发送。关闭后的具体插桩行为取决于 SDK 版本；排障时确认其不是 `false`。

- **`DD_LOGS_INJECTION`**

    **默认值**：`true`

    向支持的日志输出注入 trace 关联信息。需同时让日志采集链路保留相应字段，才能在日志与 trace 之间关联。

- **`DD_TRACE_SAMPLE_RATE`**

    **默认值**：`nil`

    设置 trace 采样率，范围为 `0.0`（0%）到 `1.0`（100%）。

- **`DD_TRACE_RATE_LIMIT`**

    **默认值**：`100`

    设置每秒最多采样多少条 trace；它只在 SDK 侧采样规则或采样率生效时起作用。

- **`DD_INSTRUMENTATION_TELEMETRY_ENABLED`**

    **默认值**：`true`

    启用或禁用 tracer 发送的遥测数据。
