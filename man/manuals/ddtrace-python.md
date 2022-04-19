{{.CSS}}

- DataKit 版本：{{.Version}}
- 文档发布日期：{{.ReleaseDate}}
- 操作系统支持：全平台

# [Tracing Python Applications](https://docs.datadoghq.com/tracing/setup_overview/setup/python)

## [Compatibility requirements](#compatibility-requirements)

The Python library supports CPython versions 2.7 and 3.5-3.10 on Linux, MacOS and Windows. For more information about Datadog’s Python version support, see the [Compatibility Requirements](https://docs.datadoghq.com/tracing/compatibility_requirements/python) page.

## [Installation and getting started](#installation-and-getting-started)

### [Follow the in-app documentation (recommended)](#follow-the-in-app-documentation-recommended)

Follow the [Quickstart instructions](https://app.datadoghq.com/apm/docs) within the Datadog app for the best experience, including:

- Step-by-step instructions scoped to your deployment configuration (hosts, Docker, Kubernetes, or Amazon ECS).
- Dynamically set `service`, `env`, and `version` tags.
- Enable the Continuous Profiler, ingesting 100% of traces, and Trace ID injection into logs during setup.

Otherwise, to begin tracing applications written in Python, install the Datadog Tracing library, `ddtrace`, using pip:

    pip install ddtrace

**Note:** This command requires pip version `18.0.0` or greater. For Ubuntu, Debian, or another package manager, update your pip version with the following command:

    pip install --upgrade pip

Then to instrument your Python application use the included `ddtrace-run` command. To use it, prefix your Python entry-point command with `ddtrace-run`.

For example, if your application is started with `python app.py` then:

    ddtrace-run python app.py

### [Upgrading to v1](#upgrading-to-v1)

If you are upgrading to ddtrace v1, review the [upgrade guide](https://ddtrace.readthedocs.io/en/stable/upgrading.html#upgrade-0-x) and the [release notes](https://ddtrace.readthedocs.io/en/stable/release_notes.html#v1-0-0) in the library documentation for full details.

### [Configure the Datadog Agent for APM](#configure-the-datadog-agent-for-apm)

Install and configure the Datadog Agent to receive traces from your now instrumented application. By default the Datadog Agent is enabled in your `datadog.yaml` file under `apm_config` with `enabled: true` and listens for trace data by default at `http://localhost:8126`. For containerized environments, follow the links below to enable trace collection within the Datadog Agent.

- [Containers](#)
- [AWS Lambda](#)
- [Other Environments](#)

- [Containers](#)

  [Containers](#)[AWS Lambda](#)[Other Environments](#)

1.  Set `apm_non_local_traffic: true` in the `apm_config` section of your main [`datadog.yaml` configuration file](https://docs.datadoghq.com/agent/guide/agent-configuration-files/#agent-main-configuration-file).

2.  See the specific setup instructions to ensure that the Agent is configured to receive traces in a containerized environment:

    [](https://docs.datadoghq.com/agent/docker/apm/?tab=java)[

](https://docs.datadoghq.com/agent/kubernetes/apm/?tab=helm)[

](https://docs.datadoghq.com/agent/amazon_ecs/apm/?tab=python)[

](https://docs.datadoghq.com/integrations/ecs_fargate/#trace-collection)

3.  After the application is instrumented, the trace client attempts to send traces to the Unix domain socket `/var/run/datadog/apm.socket` by default. If the socket does not exist, traces are sent to `http://localhost:8126`.

    If a different socket, host, or port is required, use the `DD_TRACE_AGENT_URL` environment variable. Some examples:

        DD_TRACE_AGENT_URL=http://custom-hostname:1234
        DD_TRACE_AGENT_URL=unix:///var/run/datadog/apm.socket

    The connection for traces can also be configured in code:

        from ddtrace import tracer

        # Network sockets
        tracer.configure(
            https=False,
            hostname="custom-hostname",
            port="1234",
        )

        # Unix domain socket configuration
        tracer.configure(
            uds_path="/var/run/datadog/apm.socket",
        )

    Similarly, the trace client attempts to send stats to the `/var/run/datadog/dsd.socket` Unix domain socket. If the socket does not exist then stats are sent to `http://localhost:8125`.

    If a different configuration is required, the `DD_DOGSTATSD_URL` environment variable can be used. Some examples:

        DD_DOGSTATSD_URL=http://custom-hostname:1234
        DD_DOGSTATSD_URL=unix:///var/run/datadog/dsd.socket

    The connection for stats can also be configured in code:

        from ddtrace import tracer

        # Network socket
        tracer.configure(
          dogstatsd_url="http://localhost:8125",
        )

        # Unix domain socket configuration
        tracer.configure(
          dogstatsd_url="unix:///var/run/datadog/dsd.socket",
        )

4.  Set `DD_SITE` in the Datadog Agent to `datadoghq.com` to ensure the Agent sends data to the right Datadog location.

To set up Datadog APM in AWS Lambda, see the [Tracing Serverless Functions](https://docs.datadoghq.com/tracing/serverless_functions/) documentation.

Tracing is available for a number of other environments, such as [Heroku](https://docs.datadoghq.com/agent/basic_agent_usage/heroku/#installation), [Cloud Foundry](https://docs.datadoghq.com/integrations/cloud_foundry/#trace-collection), [AWS Elastic Beanstalk](https://docs.datadoghq.com/integrations/amazon_elasticbeanstalk/), and [Azure App Service](https://docs.datadoghq.com/infrastructure/serverless/azure_app_services/#overview).

For other environments, please refer to the [Integrations](https://docs.datadoghq.com/integrations/) documentation for that environment and [contact support](https://docs.datadoghq.com/help/) if you are encountering any setup issues.

Once you’ve finished setup and are running the tracer with your application, you can run `ddtrace-run --info` to check that configurations are working as expected. Note that the output from this command does not reflect configuration changes made during runtime in code.

For more advanced usage, configuration, and fine-grain control, see Datadog’s [API documentation](https://app.datadoghq.com/apm/docs).

## [Configuration](#configuration)

When using **ddtrace-run**, the following [environment variable options](https://ddtrace.readthedocs.io/en/stable/advanced_usage.html#ddtracerun) can be used:

`DD_TRACE_DEBUG`

**Default**: `false`
Enable debug logging in the tracer.

`DD_PATCH_MODULES`

Override the modules patched for this application execution. Follow the format: `DD_PATCH_MODULES=module:patch,module:patch...`

It is recommended to use `DD_ENV`, `DD_SERVICE`, and `DD_VERSION` to set `env`, `service`, and `version` for your services. Refer to the [Unified Service Tagging](https://docs.datadoghq.com/getting_started/tagging/unified_service_tagging) documentation for recommendations on how to configure these environment variables.

`DD_ENV`

Set the application’s environment, for example: `prod`, `pre-prod`, `staging`. Learn more about [how to setup your environment](https://docs.datadoghq.com/tracing/guide/setting_primary_tags_to_scope/). Available in version 0.38+.

`DD_SERVICE`

The service name to be used for this application. The value is passed through when setting up middleware for web framework integrations like Pylons, Flask, or Django. For tracing without a web integration, it is recommended that you set the service name in code ([for example, see these Django docs](https://ddtrace.readthedocs.io/en/stable/integrations.html#django)). Available in version 0.38+.

`DD_SERVICE_MAPPING`

Define service name mappings to allow renaming services in traces, for example: `postgres:postgresql,defaultdb:postgresql`. Available in version 0.47+.

`DD_VERSION`

Set the application’s version, for example: `1.2.3`, `6c44da20`, `2020.02.13`. Available in version 0.38+.

`DD_TRACE_SAMPLE_RATE`

Enable trace volume control

`DD_TRACE_RATE_LIMIT`

Maximum number of spans to sample per-second, per-Python process. Defaults to `100` when `DD_TRACE_SAMPLE_RATE` is set. Otherwise, delegates rate limiting to the Datadog Agent.

`DD_TAGS`

A list of default tags to be added to every span and profile, for example: `layer:api,team:intake`. Available in version 0.38+.

`DD_TRACE_ENABLED`

**Default**: `true`
Enable web framework and library instrumentation. When `false`, the application code doesn’t generate any traces.

`DD_AGENT_HOST`

**Default**: `localhost`
Override the address of the trace Agent host that the default tracer attempts to submit traces to.

`DD_AGENT_PORT`

**Default**: `8126`
Override the port that the default tracer submit traces to.

`DD_TRACE_AGENT_URL`

The URL of the Trace Agent that the tracer submits to. If set, this takes priority over hostname and port. Supports Unix Domain Sockets (UDS) in combination with the `apm_config.receiver_socket` configuration in your `datadog.yaml` file or the `DD_APM_RECEIVER_SOCKET` environment variable set on the Datadog Agent. For example, `DD_TRACE_AGENT_URL=http://localhost:8126` for HTTP URL and `DD_TRACE_AGENT_URL=unix:///var/run/datadog/apm.socket` for UDS.

`DD_DOGSTATSD_URL`

The URL used to connect to the Datadog Agent for DogStatsD metrics. If set, this takes priority over hostname and port. Supports Unix Domain Sockets (UDS) in combination with the `dogstatsd_socket` configuration in your `datadog.yaml` file or the `DD_DOGSTATSD_SOCKET` environment variable set on the Datadog Agent. For example, `DD_DOGSTATSD_URL=udp://localhost:8126` for UDP URL and `DD_DOGSTATSD_URL=unix:///var/run/datadog/dsd.socket` for UDS.

`DD_DOGSTATSD_HOST`

**Default**: `localhost`
Override the address of the trace Agent host that the default tracer attempts to submit DogStatsD metrics to. Use `DD_AGENT_HOST` to override `DD_DOGSTATSD_HOST`.

`DD_DOGSTATSD_PORT`

**Default**: `8126`
Override the port that the default tracer submits DogStatsD metrics to.

`DD_LOGS_INJECTION`

**Default**: `false`
Enable [connecting logs and trace injection](https://docs.datadoghq.com/tracing/connect_logs_and_traces/python/).
