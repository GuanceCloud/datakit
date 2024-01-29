# Metric exposure and custom metrics

---

## JMXFetch {#ddtrace-jmxfetch}

When DDTrace is run in agent form, the user does not need to specifically open the jmx port. If the port is not opened, the agent will randomly open a local port.

JMXFetch collects metrics from the JMX server and sends them out in the form of a statsD data structure. Itself integrated in *dd-java-agent*.

By default, JVM information will be collected: JVM CPU, Mem, Thread, Class, etc. Specific [metric set list](jvm.md#metric)

By default, the collected indicator information is sent to `localhost:8125`. Make sure [turn on statsd collector](statsd.md) is enabled.

If it is a k8s environment, you need to configure StatsD host and port:

```shell
DD_JMXFETCH_STATSD_HOST=datakit_url
DD_JMXFETCH_STATSD_PORT=8125
```

You can use `dd.jmxfetch.<INTEGRATION_NAME>.enabled=true` to enable the specified collector.

Before filling in `INTEGRATION_NAME`, you can check [Default supported third-party software](https://docs.datadoghq.com/integrations/){:target="_blank"}

You can use `dd.jmxfetch.<INTEGRATION_NAME>.enabled=true` to enable the specified collector.

Before filling in `INTEGRATION_NAME`, you can check [Default supported third-party software](https://docs.datadoghq.com/integrations/){:target="_blank"}

For example tomcat:

## custom metrics {#custom_metric}

```shell
-Ddd.jmxfetch.tomcat.enabled=true
```

<!-- markdownlint-disable MD013 -->
## How to collect metrics through custom configuration {#custom-metric}
<!-- markdownlint-enable -->

How to collect metrics through custom configuration.

- `jvm.total_thread_count`
- `jvm.peak_thread_count`
- `jvm.daemon_thread_count`

> `dd-java-agent` has built-in these three indicators starting from v1.17.3-guance, and no additional configuration is required. However, other MBean indicators can still be configured in this customized way.

Custom indicators need to add configuration files:

1. mkdir `/usr/local/ddtrace/conf.d`, Other directories can be used.
2. Create a configuration file under the folder `guance.d/conf.yaml`.
3. `conf.yaml` at end of doc.

My service name is `tmall.jar` and the merged startup parameters are:

```shell
java -javaagent:/usr/local/dd-java-agent.jar \
  -Dcom.sun.management.jmxremote.host=127.0.0.1 \
  -Dcom.sun.manaagement.jmxremote.port=9012 \
  -Dcom.sun.management.jmxremote.ssl=false \
  -Dcom.sun.management.jmxremote.authenticate=false \
  -Ddd.jmxfetch.config.dir="/usr/local/ddtrace/conf.d/" \
  -Ddd.jmxfetch.config="guance.d/conf.yaml" \
  -jar tmall.jar
```

The conf.yaml configuration file is as follows:

```yaml
init_config:
  is_jmx: true
  collect_default_metrics: true

instances:
  - jvm_direct: true
    host: localhost
    port: 9012
    conf: 
      - include:
          domain: java.lang
          type: Threading
          attribute:
            TotalStartedThreadCount:
              alias: jvm.total_thread_count
              metric_type: gauge
            PeakThreadCount:
              alias: jvm.peak_thread_count
              metric_type: gauge
            DaemonThreadCount:
              alias: jvm.daemon_thread_count
              metric_type: gauge
```
