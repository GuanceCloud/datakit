---
title     : 'DDTrace Ruby'
summary   : 'Tracing Ruby applications with DDTrace'
tags      :
  - 'DDTRACE'
  - 'RUBY'
  - 'APM'
  - 'TRACING'
__int_icon: 'icon/ddtrace'
---


## Install Dependencies {#dependence}

For Ruby APM installation and auto-instrumentation, refer to the [Datadog Ruby Integration Documentation](https://docs.datadoghq.com/tracing/trace_collection/dd_libraries/ruby/){:target="_blank"}.

## Configuration {#config}

Ruby applications usually configure DDTrace either through environment variables or in a `Datadog.configure` block. For complete configuration details, refer to the [Datadog Ruby Tracing Documentation](https://docs.datadoghq.com/tracing/trace_collection/dd_libraries/ruby/){:target="_blank"}.

When using DataKit as the trace receiver, make sure the trace target points to DataKit instead of the default Datadog Agent endpoint. For example:

```ruby
Datadog.configure do |c|
  c.agent.host = '127.0.0.1'
  c.agent.port = 9529
end
```

If you do not need tracer telemetry data, you can disable it to reduce extra diagnostic reporting:

```shell
export DD_INSTRUMENTATION_TELEMETRY_ENABLED=false
```

## Environment Variable Support {#envs}

Below are common Ruby APM parameters. For a complete list, refer to the [Datadog Ruby configuration documentation](https://docs.datadoghq.com/tracing/trace_collection/dd_libraries/ruby/){:target="_blank"}.

- **`DD_AGENT_HOST`**

    **Default**: `127.0.0.1`

    The host address where DataKit is listening.

- **`DD_TRACE_AGENT_PORT`**

    **Default**: `8126`

    The port number where trace data is sent. When using DataKit, set it to `9529`.

- **`DD_ENV`**

    **Default**: `nil`

    Sets the environment for the application, such as `production` or `staging`.

- **`DD_SERVICE`**

    **Default**: Ruby filename

    Sets the default service name for the application.

- **`DD_TAGS`**

    **Default**: `nil`

    Sets custom tags on all traces, for example: `team:core,layer:api`.

- **`DD_VERSION`**

    **Default**: `nil`

    Sets the application version.

- **`DD_TRACE_ENABLED`**

    **Default**: `true`

    Enables or disables trace sending. When set to `false`, instrumentation still runs, but traces are not sent.

- **`DD_LOGS_INJECTION`**

    **Default**: `true`

    Injects trace correlation information into supported logs. In Rails, this is enabled by default for common loggers.

- **`DD_TRACE_SAMPLE_RATE`**

    **Default**: `nil`

    Sets the trace sampling rate from `0.0` (0%) to `1.0` (100%).

- **`DD_TRACE_RATE_LIMIT`**

    **Default**: `100`

    Sets the maximum number of sampled traces per second.

- **`DD_INSTRUMENTATION_TELEMETRY_ENABLED`**

    **Default**: `true`

    Enables or disables telemetry data sent by the tracer.
