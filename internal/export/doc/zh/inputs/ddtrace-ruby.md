---
title     : 'DDTrace Ruby'
summary   : 'DDTrace Ruby 集成'
tags      :
  - 'DDTRACE'
  - 'RUBY'
  - '链路追踪'
__int_icon: 'icon/ddtrace'
---


## 安装依赖 {#dependence}

Ruby APM 安装及自动注入方式，参见 [Datadog Ruby 接入文档](https://docs.datadoghq.com/tracing/trace_collection/dd_libraries/ruby/){:target="_blank"}。

## 配置 {#config}

Ruby 应用通常通过环境变量或 `Datadog.configure` 代码块配置 DDTrace。完整配置项说明，参见 [Datadog Ruby 链路追踪文档](https://docs.datadoghq.com/tracing/trace_collection/dd_libraries/ruby/){:target="_blank"}。

当使用 DataKit 作为 trace 接收端时，需要将 trace 上报目标从默认 Datadog Agent 地址改为 DataKit。例如：

```ruby
Datadog.configure do |c|
  c.agent.host = '127.0.0.1'
  c.agent.port = 9529
end
```

如果不需要 tracer 的遥测数据，可关闭该能力以减少额外诊断信息上报：

```shell
export DD_INSTRUMENTATION_TELEMETRY_ENABLED=false
```

## 环境变量支持 {#envs}

下面是常用的 Ruby APM 参数配置。完整参数列表，参见 [Datadog Ruby 配置文档](https://docs.datadoghq.com/tracing/trace_collection/dd_libraries/ruby/){:target="_blank"}。

- **`DD_AGENT_HOST`**

    **默认值**：`127.0.0.1`

    DataKit 监听的主机地址。

- **`DD_TRACE_AGENT_PORT`**

    **默认值**：`8126`

    Trace 数据发送端口。接入 DataKit 时需设置为 `9529`。

- **`DD_ENV`**

    **默认值**：`nil`

    设置应用运行环境，比如 `production`、`staging`。

- **`DD_SERVICE`**

    **默认值**：Ruby 文件名

    设置应用默认服务名。

- **`DD_TAGS`**

    **默认值**：`nil`

    为所有 trace 设置自定义标签，例如：`team:core,layer:api`。

- **`DD_VERSION`**

    **默认值**：`nil`

    设置应用版本号。

- **`DD_TRACE_ENABLED`**

    **默认值**：`true`

    启用或禁用 trace 发送。设置为 `false` 后，埋点仍会执行，但不会发送 trace 数据。

- **`DD_LOGS_INJECTION`**

    **默认值**：`true`

    为支持的日志输出注入 trace 关联信息。对于 Rails 常见 logger，默认启用。

- **`DD_TRACE_SAMPLE_RATE`**

    **默认值**：`nil`

    设置 trace 采样率，范围为 `0.0`（0%）到 `1.0`（100%）。

- **`DD_TRACE_RATE_LIMIT`**

    **默认值**：`100`

    设置每秒最多采样多少条 trace。

- **`DD_INSTRUMENTATION_TELEMETRY_ENABLED`**

    **默认值**：`true`

    启用或禁用 tracer 发送的遥测数据。
