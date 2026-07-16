---
title     : 'DDTrace JMX'
summary   : 'Collect JVM JMX metrics through the DDTrace Java Agent'
tags      :
  - 'DDTRACE'
  - 'JAVA'
  - 'JMX'
__int_icon: 'icon/ddtrace'
---

## JMXFetch {#ddtrace-jmxfetch}

JMXFetch is embedded in `dd-java-agent`. It reads JVM MBeans and sends metrics through DogStatsD. It is independent of trace ingestion:

`Java Agent JMXFetch` → `DogStatsD` → `DataKit StatsD collector` → `<<<custom_key.brand_name>>>`

The default JVM configuration uses `jvm_direct: true` to read MBeans **inside the current JVM process**, so JMXFetch normally does not require opening `com.sun.management.jmxremote.port`. Remote JMX is a separate operations capability. If the application genuinely needs it, configure authentication, TLS, and network access independently; never disable authentication and expose a port in production just for JMXFetch.

### Prerequisites {#prerequisites}

1. The Java application has loaded the Agent through [DDTrace Java](ddtrace-java.md).
1. DataKit has the [StatsD collector](statsd.md) enabled. Its default listener is UDP `:8125`; across hosts or Pods, verify Service, NetworkPolicy, and UDP reachability.
1. `DD_JMXFETCH_STATSD_HOST` and `DD_JMXFETCH_STATSD_PORT` point to the DataKit StatsD receiver, not trace port `9529`.

### Default Metrics and Built-In Integrations {#default-metrics}

JMXFetch commonly collects JVM memory, threads, class loading, and GC metrics. The actual set varies by Java Agent version, JVM, and configuration; use the observed [JVM metric set](jvm.md#metric) as the source of truth.

Some Java Agent versions also include JMX configurations for third-party products. Enable one only when needed, for example:

```shell
-Ddd.jmxfetch.tomcat.enabled=true
```

Here, `tomcat` is the integration name. Confirm that the installed Agent contains the integration before enabling it, and avoid unnecessary checks that create extra queries or metric cardinality.

## Configure the DataKit DogStatsD Target {#statsd-target}

For local deployment, defaults may be sufficient. Across containers or Kubernetes, configure the target explicitly:

```shell
DD_JMXFETCH_ENABLED=true \
DD_JMXFETCH_STATSD_HOST=datakit-service.datakit.svc.cluster.local \
DD_JMXFETCH_STATSD_PORT=8125 \
java -javaagent:/opt/dd-java-agent.jar \
  -Ddd.agent.host=datakit-service.datakit.svc.cluster.local \
  -Ddd.trace.agent.port=9529 \
  -jar /opt/my-app.jar
```

`DD_AGENT_HOST` / `DD_TRACE_AGENT_PORT` configure the trace destination; `DD_JMXFETCH_STATSD_HOST` / `DD_JMXFETCH_STATSD_PORT` configure the metric destination. They are not interchangeable.

## Collect Metrics with a Custom Configuration {#custom-metric}

When built-in metrics are insufficient, provide an additional JMXFetch YAML file to the Java Agent. Supply it through a ConfigMap, version-controlled image file, or managed volume, and make sure the JVM runtime user can read it.

### 1. Create the Configuration Directory {#create-config-dir}

```shell
mkdir -p /opt/ddtrace/conf.d/ext.d
```

Save the following content as `/opt/ddtrace/conf.d/ext.d/conf.yaml`.

### 2. Write the YAML {#write-yaml}

This example reads the current JVM's `java.lang:type=Threading` MBean. With `jvm_direct: true`, no remote-JMX host, port, or disabled authentication is needed:

```yaml title="conf.yaml"
init_config:
  is_jmx: true
  collect_default_metrics: true

instances:
  - jvm_direct: true
    name: application-jvm
    conf:
      - include:
          domain: java.lang
          type: Threading
          attribute:
            ThreadCount:
              alias: jvm.thread.count
              metric_type: gauge
            PeakThreadCount:
              alias: jvm.thread.peak_count
              metric_type: gauge
            DaemonThreadCount:
              alias: jvm.thread.daemon_count
              metric_type: gauge
```

`alias` becomes the metric name in DataKit. Use stable, readable names that do not grow indefinitely with dynamic values such as instance IDs. Assess cardinality and collection overhead before adding many MBeans or tags.

### 3. Load the Configuration in the Agent {#load-config}

Specify the directory and relative file name:

```shell
java -javaagent:/opt/dd-java-agent.jar \
  -Ddd.jmxfetch.config.dir=/opt/ddtrace/conf.d \
  -Ddd.jmxfetch.config=ext.d/conf.yaml \
  -Ddd.jmxfetch.statsd.host=datakit-service.datakit.svc.cluster.local \
  -Ddd.jmxfetch.statsd.port=8125 \
  -jar /opt/my-app.jar
```

The corresponding environment variables are `DD_JMXFETCH_CONFIG_DIR`, `DD_JMXFETCH_CONFIG`, `DD_JMXFETCH_STATSD_HOST`, and `DD_JMXFETCH_STATSD_PORT`. Do not set conflicting JVM properties and environment variables for the same setting.

### 4. Verify {#verify}

1. Temporarily enable Agent startup or JMXFetch logging and confirm that the YAML file loaded and the DogStatsD target is correct.
1. Confirm that the DataKit StatsD collector is running on the expected protocol and port.
1. Wait for one collection cycle, then find `jvm.thread.count`, `jvm.thread.peak_count`, and `jvm.thread.daemon_count` in Metric Explorer.
1. If data is missing, check the mount path and permissions, UDP connectivity, whether JMXFetch is enabled, and YAML indentation and MBean names in that order.

## Related Configuration {#configuration}

| JVM property | Environment variable | Description |
| --- | --- | --- |
| `dd.jmxfetch.enabled` | `DD_JMXFETCH_ENABLED` | Enables or disables JMXFetch. |
| `dd.jmxfetch.check-period` | `DD_JMXFETCH_CHECK_PERIOD` | Metric send interval in milliseconds. |
| `dd.jmxfetch.refresh-beans-period` | `DD_JMXFETCH_REFRESH_BEANS_PERIOD` | MBean-list refresh interval in seconds. |
| `dd.jmxfetch.config.dir` | `DD_JMXFETCH_CONFIG_DIR` | Additional configuration directory. |
| `dd.jmxfetch.config` | `DD_JMXFETCH_CONFIG` | Additional YAML configuration file. |
| `dd.jmxfetch.statsd.host` | `DD_JMXFETCH_STATSD_HOST` | DataKit StatsD address. |
| `dd.jmxfetch.statsd.port` | `DD_JMXFETCH_STATSD_PORT` | DataKit StatsD port, normally `8125`. |

For complete settings and built-in integrations, consult the [Datadog Java configuration guide](https://docs.datadoghq.com/tracing/trace_collection/library_config/java/){:target="_blank"} and the installed Agent version.
