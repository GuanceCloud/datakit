---
title     : 'DDTrace PHP'
summary   : 'Tracing PHP applications with DDTrace'
tags      :
  - 'DDTRACE'
  - 'PHP'
  - 'TRACING'
  - 'APM'
__int_icon: 'icon/ddtrace'
---


## Install Dependencies {#dependence}

The PHP tracer is an extension that loads before user code runs. Install an extension compatible with the current PHP version and execution mode (CLI, PHP-FPM, or Apache) by following the [Datadog PHP setup guide](https://docs.datadoghq.com/tracing/trace_collection/dd_libraries/php/){:target="_blank"}. Do not rely on setting configuration dynamically in application PHP code: automatic instrumentation has normally initialized by then.

## Configuration {#config}

The configuration location depends on the PHP execution model (CLI, PHP-FPM, or Apache module). Set environment variables at the container, FPM-pool, or web-server process level; set INI values in `php.ini` or the matching execution-mode configuration. Restart the corresponding FPM/Apache/container process after changes. See the [Datadog PHP tracer configuration guide](https://docs.datadoghq.com/tracing/trace_collection/library_config/php/){:target="_blank"} for complete settings.

When using DataKit, explicitly override the upstream default trace port `8126`. This environment-variable example is suitable for a container or process startup:

```shell
DD_SERVICE=my-php-service \
DD_ENV=production \
DD_VERSION=1.0.0 \
DD_AGENT_HOST=datakit-service \
DD_TRACE_AGENT_PORT=9529 \
php -S 0.0.0.0:8080 -t public
```

You can instead use the `datadog.agent_host` and `datadog.trace.agent_port` INI settings. `DD_TRACE_AGENT_URL` / `datadog.trace.agent_url`, when set, takes precedence over host and port; use only one destination mechanism.

High-traffic Web applications create PHP request contexts frequently. If the installed tracer supports the telemetry option and you do not need SDK telemetry, disable it in the process startup environment to reduce additional diagnostic reporting:

```shell
export DD_INSTRUMENTATION_TELEMETRY_ENABLED=false
```

After startup, request an instrumented page and confirm trace traffic in the DataKit monitor. In PHP-FPM, tracer logs normally go to the effective `error_log`; do not use only `php -i` output to diagnose runtime configuration.

## Environment Variable Support {#envs}

The following are common PHP APM parameters. See the [Datadog PHP configuration guide](https://docs.datadoghq.com/tracing/trace_collection/library_config/php/){:target="_blank"} for the full set and version-specific behavior.

- **`DD_AGENT_HOST`**

    **INI**: `datadog.agent_host`

    **Default**: `localhost`

    The trace receiver host. For DataKit, use the DataKit host or Kubernetes Service.

- **`DD_TRACE_AGENT_PORT`**

    **INI**: `datadog.trace.agent_port`

    **Upstream default**: `8126`

    The trace receiver port. Explicitly set it to `9529` for DataKit.

- **`DD_ENV`**

    **INI**: `datadog.env`

    **Default**: `null`

    Sets the deployment environment, for example `production` or `staging`.

- **`DD_SERVICE`**

    **INI**: `datadog.service`

    **Default**: `null`

    Sets the application service name. Set it explicitly in production.

- **`DD_SERVICE_MAPPING`**

    **INI**: `datadog.service_mapping`

    **Default**: `null`

    Renames APM service names, for example: `DD_SERVICE_MAPPING=pdo:payments-db,mysqli:orders-db`.

- **`DD_TRACE_AGENT_CONNECT_TIMEOUT`**

    **INI**: `datadog.trace.agent_connect_timeout`

    **Default**: `100`

    The trace receiver connection timeout in milliseconds. For remote deployments, deployments through a proxy, or cross-cluster deployments, evaluate it against actual latency rather than increasing it blindly.

- **`DD_TAGS`**

    **INI**: `datadog.tags`

    **Default**: `null`

    Sets tags appended to every span, for example `key1:value1,key2:value2`. Avoid user IDs, session IDs, tokens, and request bodies because they are high-cardinality or sensitive.

- **`DD_VERSION`**

    **INI**: `datadog.version`

    Sets the service version.

- **`DD_TRACE_SAMPLE_RATE`**

    **INI**: `datadog.trace.sample_rate`

    **Default**: `-1`

    Sets the SDK-side sampling rate from `0.0` (0%) to `1.0` (100%). It is independent of DataKit receiver-side sampling.
