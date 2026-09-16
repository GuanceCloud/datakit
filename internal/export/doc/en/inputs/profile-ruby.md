---
title     : 'Profiling Ruby'
summary   : 'Profile Ruby applications'
tags:
  - 'RUBY'
  - 'PROFILE'
__int_icon: 'icon/profiling'
---

DataKit can receive profiles from the [Datadog Ruby Continuous Profiler](https://docs.datadoghq.com/profiler/enabling/?tab=ruby){:target="_blank"} and forward them to <<<custom_key.brand_name>>>. The Ruby Profiler collects CPU time, wall time, object allocations, and other runtime data.

## Requirements {#requirements}

- Install [DataKit](https://www.<<<custom_key.brand_main_domain>>>){:target="_blank"} and enable the [Profile input](profile.md#config).
- Use CRuby 2.5 or later. CRuby 3.2.3 or later is recommended. JRuby and TruffleRuby are not supported.
- Run the application on a supported Linux x86-64 or arm64 environment, including glibc- and musl-based distributions. The Ruby Profiler does not support Serverless environments.
- Use the `datadog` gem. Version `~> 2.30` is recommended. Versions earlier than 2.30 also require `pkg-config` or `pkgconf` to build the native extension.

The compatibility range can change with SDK releases. Also check the [Ruby Profiler supported versions](https://docs.datadoghq.com/profiler/enabling/supported_versions/?tab=ruby){:target="_blank"}.

## Enable the DataKit Profile Input {#datakit}

In the `conf.d/profile` directory under the DataKit installation directory, copy `profile.conf.sample` to `profile.conf`. The default configuration already contains the endpoint used by the Ruby SDK:

```toml
[[inputs.profile]]
  endpoints = ["/profiling/v1/input"]
  body_size_limit_mb = 32
```

After you [restart DataKit](../datakit/datakit-service-how-to.md#manage-service), the receiver is available at:

```text
http://<DataKit-host>:9529/profiling/v1/input
```

Configure the Agent base URL as `http://<DataKit-host>:9529` in the Ruby SDK. Do not append `/profiling/v1/input` to the configured URL.

## Install the Ruby SDK {#install}

Add the following dependency to the application's `Gemfile`:

```ruby
gem "datadog", "~> 2.30"
```

Then install the dependency:

```shell
bundle install
```

To enable Ruby APM auto-instrumentation as well, see [DDTrace Ruby](ddtrace-ruby.md) for the gem loading entry point.

## Configure and Start the Profiler {#run}

### Environment Variables {#environment-variables}

The following example sends profiles to a local DataKit:

```shell
DD_PROFILING_ENABLED=true \
DD_TRACE_AGENT_URL=http://127.0.0.1:9529 \
DD_ENV=production \
DD_SERVICE=my-ruby-service \
DD_VERSION=1.0.0 \
DD_TAGS=team:apm,region:cn \
bundle exec ddprofrb exec ruby app.rb
```

Start a Rails application with the same settings:

```shell
DD_PROFILING_ENABLED=true \
DD_TRACE_AGENT_URL=http://127.0.0.1:9529 \
DD_ENV=production \
DD_SERVICE=my-rails-service \
DD_VERSION=1.0.0 \
bundle exec ddprofrb exec bin/rails server
```

Alternatively, configure the destination with `DD_AGENT_HOST` and `DD_TRACE_AGENT_PORT`:

```shell
export DD_AGENT_HOST=127.0.0.1
export DD_TRACE_AGENT_PORT=9529
```

`DD_TRACE_AGENT_URL` takes precedence over the host and port settings. Do not configure conflicting destinations. In a container or Kubernetes environment, replace `127.0.0.1` with a DataKit address reachable from the application when they are not in the same container.

### Code Configuration {#code-configuration}

You can also configure the Profiler during application startup. For example, add the following code to a Rails initializer:

```ruby
require "datadog"

Datadog.configure do |c|
  c.agent.host = "127.0.0.1"
  c.agent.port = 9529
  c.profiling.enabled = true
  c.env = "production"
  c.service = "my-rails-service"
  c.version = "1.0.0"
  c.tags = { "team" => "apm", "region" => "cn" }
end
```

Even with code configuration, starting the application with `ddprofrb exec` is recommended so the Profiler loads as early as possible. If the launcher cannot be used, load the Profiler at the beginning of the application entry point:

```ruby
require "datadog/profiling/preload"
```

Then start the application with its original command.

## Common Settings {#configuration}

| Environment variable | Default | Description |
| --- | --- | --- |
| `DD_PROFILING_ENABLED` | `false` | Enables the Continuous Profiler. Set it to `true` for this integration. |
| `DD_PROFILING_ALLOCATION_ENABLED` | `false` | Collects object allocation data. It adds runtime overhead, so evaluate it in a pre-production environment first. |
| `DD_PROFILING_MAX_FRAMES` | `400` | Sets the maximum number of frames collected for each stack. |
| `DD_PROFILING_EXPERIMENTAL_HEAP_ENABLED` | `false` | Enables experimental heap profiling. Allocation profiling must also be enabled. |
| `DD_ENV` | None | Sets the deployment environment, such as `production` or `staging`. |
| `DD_SERVICE` | Inferred by the SDK | Sets the service name. Set it explicitly in production. |
| `DD_VERSION` | None | Sets the application version. |
| `DD_TAGS` | None | Adds comma-separated tags in `key:value` format. |

The support range and overhead of experimental features can change with the SDK version. See the [Ruby Profiler configuration](https://docs.datadoghq.com/profiler/enabling/?tab=ruby#configuration){:target="_blank"} before enabling them.

## View Profiles {#view}

After the application starts, the Ruby Profiler periodically uploads data to DataKit. Wait one or two minutes, then open [APM -> Profile](https://console.<<<custom_key.brand_main_domain>>>/tracing/profile){:target="_blank"} in your <<<custom_key.brand_name>>> workspace and filter by `service`, `env`, and `version`.

When the application also uses DDTrace Ruby for tracing, compatible SDK versions automatically include the information used to connect traces and profiles. See [DDTrace Ruby](ddtrace-ruby.md) for tracing setup.

## DataKit Metric Generation {#metrics}

DataKit recognizes `language: ruby` in the SDK payload and forwards the original profile files and metadata. Currently, `generate_metrics` extracts `profiling_metrics` only from Java, Go, and Python profiles. Ruby profiles therefore do not generate additional `profiling_metrics`, even when this option is `true`. This does not affect flame graphs or profile details.

## Troubleshooting {#troubleshooting}

- **No profile data**: Confirm that `profile.conf` is enabled, `DD_PROFILING_ENABLED=true` is set, and at least one upload interval has elapsed.
- **Connection refused**: Confirm that the application can reach `<DataKit-host>:9529`. In a container, `127.0.0.1` refers only to that container.
- **Destination settings do not take effect**: Check whether `DD_TRACE_AGENT_URL` and `DD_AGENT_HOST`/`DD_TRACE_AGENT_PORT` are both set. Keep only one destination mechanism.
- **Native extension fails to load**: Confirm that CRuby and a supported Linux architecture are in use. For older gem versions, also confirm that `pkg-config` or `pkgconf` is installed, and inspect the gem installation output and `mkmf.log`.
- **Request body is too large**: If the DataKit log reports that the request exceeds the limit, increase `body_size_limit_mb` as needed and restart DataKit.
- **Sampling signal conflict**: The Ruby Profiler uses `SIGPROF`. If the application or another library also uses this signal, see [Troubleshooting the Ruby Profiler](https://docs.datadoghq.com/profiler/profiler_troubleshooting/ruby/){:target="_blank"}.
