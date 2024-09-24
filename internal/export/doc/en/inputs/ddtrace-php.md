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

For the installation of the PHP APM plugin, refer to the [Datadog PHP Integration Documentation](https://docs.datadoghq.com/tracing/trace_collection/automatic_instrumentation/dd_libraries/php/#install-the-extension){:target="_blank"}.

## Configuration {#config}

Depending on the PHP runtime environment (Apache/NGINX), there are some differences in the configuration. See the [Datadog PHP Trace SDK Configuration Documentation](https://docs.datadoghq.com/tracing/trace_collection/library_config/php/){:target="_blank"}.

## Environment Variable Support {#envs}

Below are common PHP APM parameter configurations. For a complete list of parameters, refer to the [Datadog Documentation](https://docs.datadoghq.com/tracing/trace_collection/library_config/php/){:target="_blank"}.

- **`DD_AGENT_HOST`**

    **INI**: `datadog.agent_host`

    **Default**: `localhost`

    The host address where Datakit is listening.

- **`DD_TRACE_AGENT_PORT`**

    **INI**: `datadog.trace.agent_port`

    **Default**: `8126`

    The port number where Datakit is listening, which should be manually set to 9529.

- **`DD_ENV`**

    **INI**: `datadog.env`

    **Default**: `null`

    Sets the environment information for the program, such as `prod/pre-prod`.

- **`DD_SERVICE`**

    **INI**: `datadog.service`

    **Default**: `null`

    Sets the APP service name.

- **`DD_SERVICE_MAPPING`**

    **INI**: `datadog.service_mapping`

    **Default**: `null`

    Renames APM service names, for example: `DD_SERVICE_MAPPING=pdo:payments-db,mysqli:orders-db`.

- **`DD_TRACE_AGENT_CONNECT_TIMEOUT`**

    **INI**: `datadog.trace.agent_connect_timeout`

    **Default**: `100`

    Agent connection timeout configuration to Datakit (unit ms), default is 100.

- **`DD_TAGS`**

    **INI**: `datadog.tags`

    **Default**: `null`

    Sets a list of tags that will be appended to each span by default, for example: `key1:value1,key2:value2`.

- **`DD_VERSION`**

    **INI**: `datadog.version`

    Sets the service version.

- **`DD_TRACE_SAMPLE_RATE`**

    **INI**: `datadog.trace.sample_rate`

    **Default**: `-1`

    Sets the sampling rate from 0.0 (0%) to 1.0 (100%).
