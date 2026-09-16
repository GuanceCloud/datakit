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

For Ruby Continuous Profiling setup, see [Profiling Ruby](profile-ruby.md).

## Install Dependencies {#dependence}

<!-- cspell:ignore datadog -->
The current Ruby SDK uses the `datadog` gem, while older applications may still use `ddtrace`. Their loading methods and configuration precedence can differ, so first identify the Gemfile and SDK version in use. See the [Datadog Ruby setup guide](https://docs.datadoghq.com/tracing/trace_collection/dd_libraries/ruby/){:target="_blank"} for compatibility and complete setup.

<!-- markdownlint-disable MD046 -->
=== "datadog gem (recommended)"

    Add the auto-instrumentation entry point to `Gemfile`, then run `bundle install`:

    ```ruby
    gem "datadog", require: "datadog/auto_instrument"
    ```

=== "ddtrace gem (legacy)"

    For applications still on the 1.x line, keep the legacy dependency and load it according to that version's documentation:

    ```ruby
    gem "ddtrace", require: "ddtrace/auto_instrument"
    ```
<!-- markdownlint-enable -->

## Configuration {#config}

Ruby applications usually configure tracing through startup environment variables or a `Datadog.configure` block. Environment variables work well for containers and platforms; code configuration is useful when settings are managed by the application. Do not set conflicting destinations or service identity in both places. See the [Datadog Ruby configuration guide](https://docs.datadoghq.com/tracing/trace_collection/library_config/ruby/){:target="_blank"} for complete settings.

When DataKit is the trace receiver, explicitly replace the upstream default `127.0.0.1:8126` with the DataKit address and port `9529`. For example:

```ruby
Datadog.configure do |c|
  c.agent.host = '127.0.0.1'
  c.agent.port = 9529
  c.service = 'my-ruby-service'
  c.env = 'production'
end
```

You can instead set `DD_AGENT_HOST`, `DD_TRACE_AGENT_PORT`, `DD_SERVICE`, `DD_ENV`, and `DD_VERSION` before the process starts. `DD_TRACE_AGENT_URL`, when set, normally takes precedence over host/port; use only one destination mechanism.

If you do not need SDK telemetry, disable it only after confirming that the installed SDK supports the option, to reduce additional diagnostic reporting:

```shell
export DD_INSTRUMENTATION_TELEMETRY_ENABLED=false
```

Start the application, call an instrumented route, and confirm a trace request in the DataKit monitor. Enable detailed SDK logging only temporarily during troubleshooting.

## Environment Variable Support {#envs}

The following are common Ruby APM parameters. Set them before the Ruby process starts; see the [Datadog Ruby configuration guide](https://docs.datadoghq.com/tracing/trace_collection/library_config/ruby/){:target="_blank"} for the complete list and precedence.

- **`DD_AGENT_HOST`**

    **Default**: `127.0.0.1`

    The trace receiver host. For DataKit, use the DataKit host or Kubernetes Service.

- **`DD_TRACE_AGENT_PORT`**

    **Upstream default**: `8126`

    The trace delivery port. When using DataKit, explicitly set it to `9529`.

- **`DD_ENV`**

    **Default**: `nil`

    Sets the environment for the application, such as `production` or `staging`.

- **`DD_SERVICE`**

    **Default**: Determined by SDK and application startup mode

    Sets the application service name. Set it explicitly in production so it does not vary with an entry file or framework.

- **`DD_TAGS`**

    **Default**: `nil`

    Sets custom tags on all traces, for example `team:core,layer:api`. Do not send tokens, personal data, or high-cardinality request identifiers.

- **`DD_VERSION`**

    **Default**: `nil`

    Sets the application version.

- **`DD_TRACE_ENABLED`**

    **Default**: `true`

    Enables or disables trace delivery. Behavior after disabling can vary by SDK version; during troubleshooting, make sure it is not `false`.

- **`DD_LOGS_INJECTION`**

    **Default**: `true`

    Injects trace-correlation information into supported logs. The log-collection path must retain those fields for log-to-trace correlation.

- **`DD_TRACE_SAMPLE_RATE`**

    **Default**: `nil`

    Sets the trace sampling rate from `0.0` (0%) to `1.0` (100%).

- **`DD_TRACE_RATE_LIMIT`**

    **Default**: `100`

    Sets the maximum number of sampled traces per second. It applies when SDK-side sampling rules or a sample rate are active.

- **`DD_INSTRUMENTATION_TELEMETRY_ENABLED`**

    **Default**: `true`

    Enables or disables telemetry data sent by the tracer.
