---
title     : 'DDTrace Java'
summary   : 'Trace Java applications with the DDTrace Java Agent'
tags      :
  - 'DDTRACE'
  - 'JAVA'
  - 'APM'
  - 'TRACING'
__int_icon: 'icon/ddtrace'
---

The DDTrace Java Agent uses the JVM `-javaagent` mechanism to instrument classes as they load, so common frameworks usually need no business-code changes. This page explains how to send Java Agent traces to DataKit; it is not a complete guide to log collection, profiling, or JMX metrics.

## Preconditions {#requirements}

1. Install DataKit and enable the [DDTrace collector](ddtrace.md). Explicitly configure the Java Agent to send traces to the DataKit HTTP address and port `9529`.
1. Use JDK 8 or later and choose an Agent version compatible with the application runtime, operating system, and CPU architecture. Validate JDK or Agent upgrades before production.
1. Ensure that the JVM runtime user can read the Agent JAR at its absolute path.
1. Profiling and JMX metrics use separate data paths: profiling requires the [Profiling collector](profile.md), and JMX metrics require the [StatsD collector](statsd.md).

<!-- markdownlint-disable MD046 -->
???+ warning

    A JVM can load multiple Java Agents, but their class-transformation order and compatibility affect startup and data quality. Do not add two DDTrace Java Agents to one process. When combining with another APM, security, or bytecode Agent, load-test and validate critical application paths first.
<!-- markdownlint-enable MD046 -->

## Get the Agent {#dependence}

<!-- markdownlint-disable MD046 -->
=== "<<<custom_key.brand_name>>> Extended Version"

    The extended version adds selected framework integrations and finer-grained collection on top of the upstream Java Agent. Pin an exact version after downloading it, and read the [Java extension guide](ddtrace-ext-java.md) and [changelog](ddtrace-ext-changelog.md):

    ```shell
    wget -O /opt/dd-java-agent.jar \
      'https://static.<<<custom_key.brand_main_domain>>>/dd-image/dd-java-agent.jar'
    ```

=== "Datadog Upstream Version"

    ```shell
    wget -O /opt/dd-java-agent.jar 'https://dtdg.co/latest-java-tracer'
    ```
<!-- markdownlint-enable MD046 -->

The extended and upstream versions can have different defaults and supported settings. Whichever JAR you use, explicitly set the DataKit host and port in startup configuration to avoid sending to the common upstream default port `8126` by mistake.

## Running the Application {#instrument}

### Host Application {#host-application}

Place every `-D` option before `-jar`. The following command provides a minimal, identifiable service configuration:

```shell linenums="1"
java \
  -javaagent:/opt/dd-java-agent.jar \
  -Ddd.service=my-java-service \
  -Ddd.env=production \
  -Ddd.version=1.0.0 \
  -Ddd.agent.host=127.0.0.1 \
  -Ddd.trace.agent.port=9529 \
  -Ddd.logs.injection=true \
  -jar /opt/my-app.jar
```

The environment-variable equivalent of `dd.service` is `DD_SERVICE`; do not write the property as `dd.service.name`. You can also set `DD_AGENT_HOST`, `DD_TRACE_AGENT_PORT`, `DD_SERVICE`, `DD_ENV`, and `DD_VERSION` before process startup. `DD_TRACE_AGENT_URL` / `dd.trace.agent.url`, when set, normally takes precedence over host and port, so use only one destination mechanism.

### Kubernetes {#kubernetes}

For Kubernetes, use [DataKit Operator DDTrace injection](../operator-ddtrace.md) where possible. The Operator injects the JAR, volumes, and language-specific startup configuration into newly created Pods. Changing Operator configuration or a Pod annotation requires Pod recreation (for example, a rollout restart) before the webhook runs again.

An application image can also carry its own Agent. The following fragment works only when the JAR already exists at `/opt/dd-java-agent.jar` in the image. **Environment variables alone do not load an Agent and cannot enable automatic instrumentation.**

```yaml
containers:
  - name: app
    image: registry.example/my-java-app:1.0.0
    env:
      - name: JAVA_TOOL_OPTIONS
        value: "-javaagent:/opt/dd-java-agent.jar"
      - name: DD_SERVICE
        value: "my-java-service"
      - name: DD_ENV
        value: "production"
      - name: DD_VERSION
        value: "1.0.0"
      - name: DD_AGENT_HOST
        value: "datakit-service.datakit.svc.cluster.local"
      - name: DD_TRACE_AGENT_PORT
        value: "9529"
```

If the image does not contain the JAR, use the Operator or a controlled initContainer plus shared volume. Do not put an unverified JAR download into a business-container startup script.

### Enable Profiling {#instrument-profiling}

Profiling data goes to the DataKit Profiling receiver, not the DDTrace trace receiver. Enable the [Profiling collector](profile.md) first, then enable the Agent feature:

```shell
DD_PROFILING_ENABLED=true \
java -javaagent:/opt/dd-java-agent.jar \
  -Ddd.agent.host=127.0.0.1 \
  -Ddd.trace.agent.port=9529 \
  -jar /opt/my-app.jar
```

Assess performance, storage, and compliance impact before enabling it in production. Do not diagnose missing profiling data as a trace-receiver problem.

### Set a Sampling Rate {#instrument-sampling}

Application-side sampling reduces traffic at the source:

```shell
DD_TRACE_SAMPLE_RATE=0.8 \
java -javaagent:/opt/dd-java-agent.jar \
  -Ddd.agent.host=127.0.0.1 \
  -Ddd.trace.agent.port=9529 \
  -jar /opt/my-app.jar
```

`0.8` means 80% SDK-side sampling. It is independent of receiver-side sampling in the DataKit `ddtrace` collector. Decide which layer controls traffic so the combined result remains explainable.

### Enable JVM Metrics Collection {#instrument-jvm-metrics}

JMXFetch can collect JVM metrics by default, but it sends them through DogStatsD. Enable the [StatsD collector](statsd.md) and use a **separate** StatsD address and port, normally UDP `8125`:

```shell
DD_JMXFETCH_ENABLED=true \
DD_JMXFETCH_STATSD_HOST=<YOUR-DATAKIT-HOST> \
DD_JMXFETCH_STATSD_PORT=8125 \
java -javaagent:/opt/dd-java-agent.jar \
  -Ddd.agent.host=<YOUR-DATAKIT-HOST> \
  -Ddd.trace.agent.port=9529 \
  -jar /opt/my-app.jar
```

Do not set `DD_JMXFETCH_STATSD_PORT` to `9529`. See [DDTrace JMX](ddtrace-jmxfetch.md) for custom MBean metrics.

## Parameter Explanation {#start-options}

The table lists settings most relevant to DataKit. For the complete set, version requirements, and defaults, use the [Datadog Java configuration guide](https://docs.datadoghq.com/tracing/trace_collection/library_config/java/){:target="_blank"}.

| JVM property | Environment variable | Purpose and DataKit guidance |
| --- | --- | --- |
| `dd.service` | `DD_SERVICE` | Service name. Set it explicitly so it does not change with auto-detection. |
| `dd.env` | `DD_ENV` | Deployment environment, such as `production` or `staging`. |
| `dd.version` | `DD_VERSION` | Application version, useful for release and regression comparison. |
| `dd.trace.agent.url` | `DD_TRACE_AGENT_URL` | Full trace-receiver URL. It normally overrides host/port. |
| `dd.agent.host` | `DD_AGENT_HOST` | DataKit host or Kubernetes Service. |
| `dd.trace.agent.port` | `DD_TRACE_AGENT_PORT` | Trace receiver port. The common upstream default is `8126`; DataKit uses `9529`. |
| `dd.trace.sample.rate` | `DD_TRACE_SAMPLE_RATE` | SDK-side root-trace sampling rate from `0.0` to `1.0`. |
| `dd.trace.enabled` | `DD_TRACE_ENABLED` | Controls automatic instrumentation and trace generation; during troubleshooting, make sure it is not `false`. |
| `dd.logs.injection` | `DD_LOGS_INJECTION` | Adds trace/span correlation fields to supported log frameworks; log collection remains a separate setup. |
| `dd.profiling.enabled` | `DD_PROFILING_ENABLED` | Enables Continuous Profiling and requires DataKit `profile`. |
| `dd.jmxfetch.enabled` | `DD_JMXFETCH_ENABLED` | Enables JMXFetch; metrics must go to DataKit StatsD on `8125`. |
| `dd.trace.startup.logs`, `dd.trace.debug` | `DD_TRACE_STARTUP_LOGS`, `DD_TRACE_DEBUG` | Startup configuration and debug logs. Enable only temporarily for troubleshooting. |

## Trace Error Behavior {#error}

Whether automatic instrumentation marks a span as an error depends on the integration and SDK configuration. Do not assume that every exception or every 4xx/5xx response is an error. In general:

- An exception that is not handled by the application and reaches a supported framework boundary is normally recorded with error information and a stack trace.
- HTTP server and client error-status ranges are configurable independently. Common upstream defaults treat server `5xx` and client `4xx` responses as errors; the extended Agent can add further behavior.
- When application code catches and handles an exception, automatic instrumentation may not mark the current span as an error. If the failure is still meaningful to the business, use the installed SDK API to record it explicitly.
- DataKit commonly exposes `error.type`, `error.message`, and `error.stack`. Stack traces and messages can contain sensitive data, so control access and retention according to your data-governance policy.

To investigate, reproduce a minimal request and inspect the application log, Agent startup log, and trace detail together. Do not infer instrumentation success solely from an HTTP status code.

## Verification and Troubleshooting {#verify}

1. Temporarily set `DD_TRACE_STARTUP_LOGS=true` and confirm in JVM logs that the Agent loaded and points to DataKit.
1. Call an endpoint that passes through HTTP, a database, or RPC, then confirm requests to `/v0.3/traces`, `/v0.4/traces`, or `/v0.5/traces` in the DataKit monitor.
1. In Kubernetes, also check `JAVA_TOOL_OPTIONS`/the command line, the Agent-file mount, `DD_AGENT_HOST`, and `DD_TRACE_AGENT_PORT` in the Pod.
1. If traces arrive but cross-service topology is broken, check propagation protocols and [Multi-Tracing Propagation](tracing-propagator.md).

## More Reading {#more-reading}

- [DDTrace receiver configuration, sampling, and tags](ddtrace.md)
- [<<<custom_key.brand_name>>> Java Agent extensions](ddtrace-ext-java.md)
- [Java Agent extension changelog](ddtrace-ext-changelog.md)
- [JMXFetch and custom JVM metrics](ddtrace-jmxfetch.md)
- [Profiling collector](profile.md)
- [Multi-Tracing Propagation](tracing-propagator.md)
- [DDTrace sampling strategies](tracing-sample.md)
